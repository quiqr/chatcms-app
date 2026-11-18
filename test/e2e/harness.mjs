// Starts everything an acceptance test needs and stops it again: a PostgreSQL
// of its own, a bare remote on disk, a seeded database and the server. It
// takes nothing from the machine it runs on beyond the programs Nix pins, so
// the suite says the same thing in the gate as on a developer's laptop.
import { execFileSync, spawn } from "node:child_process";
import { mkdtempSync, rmSync, writeFileSync, existsSync, mkdirSync, readFileSync, statSync } from "node:fs";
import { createServer } from "node:http";
import { tmpdir } from "node:os";
import { join, normalize } from "node:path";
import { randomBytes } from "node:crypto";

const sh = (command, args, options = {}) =>
  execFileSync(command, args, { encoding: "utf8", ...options });

function freePort() {
  return new Promise((given, failed) => {
    const probe = createServer();
    probe.once("error", failed);
    probe.listen(0, "127.0.0.1", () => {
      const { port } = probe.address();
      probe.close(() => given(port));
    });
  });
}

export class Harness {
  constructor({ assistant = false } = {}) {
    this.assistant = assistant;
    this.asked = [];
    this.root = mkdtempSync(join(tmpdir(), "chatcms-e2e-"));
    this.pgdata = join(this.root, "pgdata");
    this.pgsock = join(this.root, "sock");
    this.state = join(this.root, "state");
    this.server = null;
    this.mailpit = null;
    this.templates = null;
    this.mailLog = "";
    this.sendsMail = false;
    this.whyNoMail = "";
  }

  // theInbox is where mail this run sent can be read.
  get theInbox() {
    return `http://127.0.0.1:${this.inboxPort}`;
  }

  get url() {
    return `http://127.0.0.1:${this.port}`;
  }

  // remoteOf is the address of a project's own remote on disk, which is what
  // an implementer would paste into the setup screen.
  remoteOf(project) {
    return join(this.state, "dev-remotes", `${project}.git`);
  }

  async start() {
    mkdirSync(this.pgsock, { recursive: true });
    mkdirSync(this.state, { recursive: true });

    sh("initdb", ["--username=chatcms", "--auth=trust", "--encoding=UTF8", "--no-locale", "-D", this.pgdata]);
    sh("pg_ctl", ["-D", this.pgdata, "-o", `-k ${this.pgsock} -h ''`, "-l", join(this.root, "pg.log"), "start"]);
    sh("createdb", ["-h", this.pgsock, "-U", "chatcms", "chatcms_e2e"]);

    this.database = `postgres:///chatcms_e2e?host=${this.pgsock}&user=chatcms`;

    // Written as text, not as raw bytes: the server trims whitespace from the
    // key file, so raw bytes that happen to start or end with a space would
    // leave it a byte short, which is a failure one run in twenty.
    const key = join(this.state, "master.key");
    writeFileSync(key, randomBytes(32).toString("hex"), { mode: 0o600 });

    // The seed makes an organisation, people, two projects and one site with
    // a remote of its own, which is what the acceptance tests act on.
    sh(process.env.CHATCMS_SEED ?? "seed", [this.database], {
      env: { ...process.env, CHATCMS_STATE_DIR: this.state },
    });

    this.port = await freePort();

    // A Mailpit of this run's own, so that what was sent can be read back
    // rather than fished out of a log, and so that two runs never read each
    // other's mail.
    this.smtpPort = await freePort();
    this.inboxPort = await freePort();
    await this.startTheInbox();
    await this.startTheTemplates();
    const assistant = this.assistant ? await this.startTheModel() : {};

    this.server = spawn(process.env.CHATCMS_BIN ?? "chatcmsd", [], {
      env: {
        ...process.env,
        CHATCMS_DATABASE_URL: this.database,
        CHATCMS_SECRET_KEY_FILE: key,
        CHATCMS_STATE_DIR: this.state,
        CHATCMS_LISTEN: `127.0.0.1:${this.port}`,
        CHATCMS_BASE_URL: this.url,
        // With no Mailpit there is nowhere to send, and the server says so
        // and writes links to its log instead, which is what it does on an
        // installation with no mail server.
        CHATCMS_MAIL_HOST: this.sendsMail ? "127.0.0.1" : "",
        CHATCMS_MAIL_PORT: String(this.smtpPort),
        CHATCMS_MAIL_FROM: "chatcms@localhost",
        // Mailpit has no certificate, and asking for one would mean nothing
        // could be sent at all.
        CHATCMS_MAIL_STARTTLS: "no",
        CHATCMS_TEMPLATE_FEED: `${this.templatesUrl}/templates.json`,
        ...assistant,
      },
      stdio: ["ignore", "pipe", "pipe"],
    });

    this.log = "";
    this.server.stdout.on("data", (chunk) => (this.log += chunk));
    this.server.stderr.on("data", (chunk) => (this.log += chunk));

    await this.waitUntilAnswering();
  }

  // startTheInbox runs a Mailpit for this harness. A missing Mailpit is not
  // fatal: the server is then started with nowhere to send, so the specs
  // that need mail fail saying so and the thirty that do not still run. One
  // missing tool should break the thing that needs it, not everything.
  async startTheInbox() {
    const binary = process.env.CHATCMS_MAILPIT_BIN ?? "mailpit";

    try {
      this.mailpit = spawn(binary, [
        "--listen", `127.0.0.1:${this.inboxPort}`,
        "--smtp", `127.0.0.1:${this.smtpPort}`,
        "--database", join(this.root, "mailpit.db"),
        "--quiet",
      ], { stdio: ["ignore", "pipe", "pipe"] });

      this.mailpit.on("error", (why) => (this.mailLog += `${why.message}\n`));
      this.mailpit.stdout.on("data", () => {});
      this.mailpit.stderr.on("data", (chunk) => (this.mailLog += chunk));

      await this.waitUntilTheInboxAnswers();
      this.sendsMail = true;
    } catch (why) {
      this.sendsMail = false;
      this.whyNoMail = `${binary} could not be started: ${why.message}`;
      if (this.mailpit && this.mailpit.exitCode === null) this.mailpit.kill("SIGKILL");
      this.mailpit = null;
    }
  }

  async waitUntilTheInboxAnswers() {
    const deadline = Date.now() + 30_000;
    while (Date.now() < deadline) {
      if (this.mailpit.exitCode !== null) {
        throw new Error(`Mailpit stopped before it listened:\n${this.mailLog}`);
      }
      try {
        const answer = await fetch(`${this.theInbox}/api/v1/info`);
        if (answer.ok) return;
      } catch {
        // Not up yet.
      }
      await new Promise((wake) => setTimeout(wake, 100));
    }
    throw new Error(`Mailpit did not answer in time:\n${this.mailLog}`);
  }

  // waitForAMessage asks the inbox until one addressed to this person
  // arrives, which is how a test follows an invitation the way a person
  // would rather than reading a token out of the database.
  async waitForAMessage(to) {
    if (!this.sendsMail) {
      throw new Error(`this run has no mail server, so nothing was sent: ${this.whyNoMail}`);
    }
    const deadline = Date.now() + 30_000;
    while (Date.now() < deadline) {
      const answer = await fetch(`${this.theInbox}/api/v1/search?query=${encodeURIComponent("to:" + to)}`);
      if (answer.ok) {
        const found = await answer.json();
        if (found.messages && found.messages.length > 0) {
          const full = await fetch(`${this.theInbox}/api/v1/message/${found.messages[0].ID}`);
          return await full.json();
        }
      }
      await new Promise((wake) => setTimeout(wake, 200));
    }
    throw new Error(`no message for ${to} arrived in time`);
  }

  // Waits by asking, with a deadline, rather than by sleeping for a guess.
  async waitUntilAnswering() {
    const deadline = Date.now() + 60_000;
    while (Date.now() < deadline) {
      if (this.server.exitCode !== null) {
        throw new Error(`the server stopped before it listened:\n${this.log}`);
      }
      try {
        const answer = await fetch(`${this.url}/api/health`);
        if (answer.ok) return;
      } catch {
        // Not up yet.
      }
      await new Promise((wake) => setTimeout(wake, 100));
    }
    throw new Error(`the server did not answer in time:\n${this.log}`);
  }

  async startTheTemplates() {
    const git = process.env.CHATCMS_GIT_BIN ?? "git";
    const root = join(this.root, "templates");
    const bakery = join(root, "bakery.git");
    mkdirSync(join(root, "templates", "Bakery"), { recursive: true });

    const site = join(this.state, "dev-remotes");
    const seeded = sh("ls", [site]).trim().split("\n")[0];
    sh(git, ["clone", "--bare", join(site, seeded), bakery]);
    sh(git, ["-C", bakery, "update-server-info"]);
    writeFileSync(
      join(root, "templates", "Bakery", "screenshot.png"),
      Buffer.from(
        "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==",
        "base64",
      ),
    );

    const bundles = join(root, "bundles.git");
    this.makeTheBundleTemplate(git, join(root, "bundles-work"), bundles);
    mkdirSync(join(root, "templates", "Bundles"), { recursive: true });

    this.templates = createServer((request, response) => {
      const path = normalize(decodeURIComponent(new URL(request.url, "http://x").pathname));
      if (path === "/templates.json") {
        response.setHeader("Content-Type", "application/json");
        response.end(
          JSON.stringify([
            {
              QuiqrEtalageName: "Bakery",
              NormalizedName: "Bakery",
              QuiqrEtalageDescription: "A small bakery with a homepage and a model.",
              QuiqrEtalageAuthor: "The acceptance suite",
              QuiqrEtalageLicense: "MIT",
              HugoVersion: "extended_0.101.0",
              QuiqrModel: "base singles",
              ScreenshotImageType: "png",
              SourceLink: `${this.templatesUrl}/bakery.git`,
            },
            {
              QuiqrEtalageName: "Bundles",
              NormalizedName: "Bundles",
              QuiqrEtalageDescription: "A journal whose posts are page bundles, with a menu.",
              QuiqrEtalageAuthor: "The acceptance suite",
              QuiqrEtalageLicense: "MIT",
              HugoVersion: "0.101.0",
              QuiqrModel: "base collections menu",
              SourceLink: `${this.templatesUrl}/bundles.git`,
            },
          ]),
        );
        return;
      }
      const file = join(root, path);
      if (!file.startsWith(root) || !existsSync(file) || !statSync(file).isFile()) {
        response.statusCode = 404;
        response.end();
        return;
      }
      response.end(readFileSync(file));
    });
    await new Promise((listening) => this.templates.listen(0, "127.0.0.1", listening));
    this.templatesUrl = `http://127.0.0.1:${this.templates.address().port}`;
  }

  makeTheBundleTemplate(git, work, bare) {
    const files = {
      "hugo.toml": "baseURL = 'https://journal.example'\ntitle = 'Journal'\n",
      "quiqr/model/base.yaml": [
        "hugover: 0.101.0",
        "singles:",
        "  - key: home",
        "    title: Home",
        "    file: content/_index.md",
        "    fields:",
        "      - key: title",
        "        type: string",
        "  - key: readme",
        "    title: Developer notes",
        "    file: quiqr/home/index.md",
        "    fields:",
        "      - key: mainContent",
        "        type: markdown",
        "collections:",
        "  - key: posts",
        "    title: Posts",
        "    itemtitle: post",
        "    folder: content/post/",
        "    extension: md",
        "    dataformat: yaml",
        "    hideIndex: true",
        "    sortkey: date",
        "    fields:",
        "      - key: title",
        "        type: string",
        "      - key: date",
        "        type: date",
        "      - key: mainContent",
        "        type: markdown",
        "menu:",
        "  - key: journal",
        "    title: Journal",
        "    menuItems:",
        "      - key: home",
        "      - key: posts",
        "  - key: developer",
        "    title: Developer",
        "    matchRole: siteDeveloper",
        "    menuItems:",
        "      - key: readme",
        "",
      ].join("\n"),
      "content/_index.md": "---\ntitle: Journal\n---\n\nWelcome.\n",
      "content/post/_index.md": "---\ntitle: All posts\n---\n",
      "content/post/2024/spring-sale/index.md": "---\ntitle: Spring sale\ndate: 2024-03-01\n---\n\nPlums are in.\n",
      "content/post/2024/spring-sale/cover.svg": "<svg xmlns='http://www.w3.org/2000/svg' width='4' height='4'/>\n",
      "content/post/autumn/index.md": "---\ntitle: Autumn walk\ndate: 2023-10-01\n---\n\nLeaves.\n",
      "quiqr/home/index.md": "Notes for whoever builds this site.\n",
    };
    for (const [name, content] of Object.entries(files)) {
      mkdirSync(join(work, name, ".."), { recursive: true });
      writeFileSync(join(work, name), content);
    }
    sh(git, ["-C", work, "init", "--initial-branch=main"]);
    sh(git, ["-C", work, "add", "--all"]);
    sh(git, ["-C", work, "-c", "user.name=Suite", "-c", "user.email=suite@example.org", "commit", "--message=The journal"]);
    sh(git, ["clone", "--bare", work, bare]);
    sh(git, ["-C", bare, "update-server-info"]);
  }

  emptyRemote(name) {
    const remote = join(this.state, "dev-remotes", `${name}.git`);
    sh(process.env.CHATCMS_GIT_BIN ?? "git", ["init", "--bare", "--initial-branch=main", remote]);
    return remote;
  }

  // onTheRemote asks the customer's own remote what it holds, which is how a
  // test checks that publishing actually published.
  onTheRemote(project, ...args) {
    const remote = join(this.state, "dev-remotes", `${project}.git`);
    if (!existsSync(remote)) throw new Error(`no remote at ${remote}`);
    return sh(process.env.CHATCMS_GIT_BIN ?? "git", ["-C", remote, ...args]);
  }

  // projectId is how the seed names the site on disk, which the remote path
  // and the working copy are both named after.
  projectId() {
    const ids = sh("ls", [join(this.state, "sites")]).trim().split("\n");
    return ids[0];
  }

  // pushFromSomewhereElse is somebody editing the site outside ChatCMS, which
  // is what acceptance test 6 is about.
  pushFromSomewhereElse(project, path, content, message) {
    const working = mkdtempSync(join(this.root, "elsewhere-"));
    const remote = join(this.state, "dev-remotes", `${project}.git`);
    const git = process.env.CHATCMS_GIT_BIN ?? "git";

    sh(git, ["clone", remote, working]);
    const where = join(working, path);
    mkdirSync(join(where, ".."), { recursive: true });
    writeFileSync(where, content);
    sh(git, ["-C", working, "add", "--all"]);
    sh(git, ["-C", working, "-c", "user.name=Joris", "-c", "user.email=joris@example.org",
      "commit", "--message=" + message]);
    sh(git, ["-C", working, "push", "origin", "HEAD:refs/heads/main"]);
  }

  async startTheModel() {
    const port = await freePort();
    this.model = createServer((request, response) => {
      let body = "";
      request.on("data", (chunk) => (body += chunk));
      request.on("end", () => {
        const asked = body ? JSON.parse(body) : {};
        this.asked.push(asked);
        if (request.url.endsWith("/models")) {
          response.setHeader("Content-Type", "application/json");
          response.end(JSON.stringify({ object: "list", data: [{ id: "e2e", object: "model" }] }));
          return;
        }
        const update = (asked.tools ?? []).find((tool) => /update_page$/.test(tool.function?.name ?? ""));
        const messages = asked.messages ?? [];
        const answered = messages.some((message) => message.role === "tool");
        const said = [...messages].reverse().find((message) => message.role === "user");
        const text = typeof said?.content === "string" ? said.content : (said?.content ?? []).map((part) => part.text ?? "").join("");
        const title = (text.match(/title\b.* to (.+?)\.?$/i) ?? [])[1];

        let chunks;
        if (update && title && !answered) {
          chunks = [
            { delta: { role: "assistant", tool_calls: [{ index: 0, id: "call_e2e", type: "function", function: { name: update.function.name, arguments: "" } }] } },
            { delta: { tool_calls: [{ index: 0, function: { arguments: JSON.stringify({ key: "home", fields: { title } }) } }] } },
            { delta: {}, finish_reason: "tool_calls" },
          ];
        } else {
          const reply = update && answered ? "Done. The home page has its new title." : "Chat";
          chunks = [
            { delta: { role: "assistant", content: reply.slice(0, 5) } },
            { delta: { content: reply.slice(5) } },
            { delta: {}, finish_reason: "stop" },
          ];
        }
        response.setHeader("Content-Type", "text/event-stream");
        for (const chunk of chunks) {
          const choice = { index: 0, delta: chunk.delta, finish_reason: chunk.finish_reason ?? null };
          response.write(`data: ${JSON.stringify({ id: "e2e", object: "chat.completion.chunk", created: 1, model: "e2e", choices: [choice] })}\n\n`);
        }
        response.write(`data: ${JSON.stringify({ id: "e2e", object: "chat.completion.chunk", created: 1, model: "e2e", choices: [], usage: { prompt_tokens: 10, completion_tokens: 5, total_tokens: 15 } })}\n\n`);
        response.end("data: [DONE]\n\n");
      });
    });
    await new Promise((listening) => this.model.listen(port, "127.0.0.1", listening));
    return {
      CHATCMS_AI_PROVIDER_URL: `http://127.0.0.1:${port}/v1`,
      CHATCMS_AI_KEY: "e2e-key",
      CHATCMS_AI_MODEL: "e2e",
      CHATCMS_AI_LISTEN: `127.0.0.1:${await freePort()}`,
    };
  }

  async stop() {
    if (this.server && this.server.exitCode === null) {
      this.server.kill("SIGTERM");
      await new Promise((done) => this.server.on("exit", done));
    }
    if (this.mailpit && this.mailpit.exitCode === null) {
      this.mailpit.kill("SIGTERM");
      await new Promise((done) => this.mailpit.on("exit", done));
    }
    if (this.templates) {
      await new Promise((closed) => this.templates.close(closed));
    }
    if (this.model) {
      this.model.closeAllConnections();
      await new Promise((closed) => this.model.close(closed));
    }
    try {
      sh("pg_ctl", ["-D", this.pgdata, "stop", "-m", "immediate"]);
    } catch {
      // Already gone.
    }
    rmSync(this.root, { recursive: true, force: true });
  }
}

export const owner = { email: "owner@example.org", password: "a long enough passphrase" };

// The person who runs the service: they administer the installation and
// belong to no organisation, which is the whole point of the role.
export const administrator = {
  email: "admin@example.org",
  password: "a long enough passphrase",
};
