import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { fill } from "./fill";
import { Inside } from "./here";
import { OnboardingPage } from "../OnboardingPage";
import { ProjectSettingsPage } from "../ProjectSettingsPage";
import { TemplateGalleryPage } from "../TemplateGalleryPage";
import { UseTemplatePage } from "../UseTemplatePage";

type Answer = { ok?: boolean; status?: number; body: unknown };
type Call = { path: string; method: string; body?: unknown };

const summer = {
  id: "Summer",
  name: "Summer",
  description: "Simple fresh looking recipe site, ported from Jekyll to Hugo",
  author: "Connor Bär (Ported by Pim Snel)",
  authorUrl: "https://example.org/connor",
  license: "MIT",
  licenseUrl: "https://example.org/mit",
  homepage: "",
  demoUrl: "https://mipmip.github.io/summer-qremix/",
  hugoTheme: "summer-qremix",
  hugoVersion: "0.101.0",
  extended: true,
  model: ["base", "collections", "menu", "singles"],
  forms: 0,
  sourceUrl: "https://github.com/mipmip/summer-qremix",
  screenshotUrl: "/api/organisations/technative/templates/Summer/screenshot",
};

const kitchen = {
  ...summer,
  id: "KitchenSink",
  name: "KitchenSink",
  description: "Every Quiqr field type in one site.",
  author: "Quiqr",
  demoUrl: "",
  hugoVersion: "0.139.0",
  extended: true,
  screenshotUrl: "",
};

const paper = { ...summer, id: "Paper", name: "Paper", description: "A clean, dark blog theme.", author: "Someone" };

const installed = { version: "0.139.0", extended: true };

function gallery(overrides: Record<string, unknown> = {}) {
  return { offered: true, available: true, templates: [kitchen, paper, summer], hugo: installed, ...overrides };
}

function serve(byPath: Record<string, Answer>, calls: Call[] = []) {
  vi.stubGlobal(
    "fetch",
    vi.fn().mockImplementation((path: string, options?: { method?: string; body?: string }) => {
      const method = options?.method ?? "GET";
      calls.push({ path, method, body: options?.body ? JSON.parse(options.body) : undefined });
      const answer = byPath[`${method} ${path}`] ?? byPath[path];
      if (!answer) {
        return Promise.resolve({
          ok: false,
          status: 404,
          json: async () => ({ error: { code: "not_found", message: "There is nothing here." } }),
        });
      }
      return Promise.resolve({ ok: answer.ok ?? true, status: answer.status ?? 200, json: async () => answer.body });
    }),
  );
  return calls;
}

function at(path: string, routes: React.ReactNode) {
  return render(
    <Inside>
      <MemoryRouter initialEntries={[path]}>
        <Routes>
          {routes}
          <Route path="/o/:organisation/p/:project/setup" element={<p>the setup screen</p>} />
          <Route path="/o/:organisation/projects" element={<p>the projects screen</p>} />
        </Routes>
      </MemoryRouter>
    </Inside>,
  );
}

const galleryRoute = <Route path="/o/:organisation/new/template" element={<TemplateGalleryPage />} />;
const stepRoute = <Route path="/o/:organisation/new/template/:template" element={<UseTemplatePage />} />;
const templates = "/api/organisations/technative/templates";

beforeEach(() => {
  document.cookie = "chatcms_request_token=a-token; path=/";
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("the template gallery", () => {
  it("shows every template as a card, in the order it is given", async () => {
    serve({ [templates]: { body: gallery() } });
    at("/o/technative/new/template", galleryRoute);

    const cards = await screen.findAllByRole("article");
    expect(cards.map((card) => card.getAttribute("aria-label"))).toEqual(["KitchenSink", "Paper", "Summer"]);
    expect(screen.getByText("3 templates")).toBeInTheDocument();
    expect(screen.getByText("Connor Bär (Ported by Pim Snel) · MIT")).toBeInTheDocument();
    expect(within(cards[2] as HTMLElement).getByAltText("Screenshot of Summer")).toHaveAttribute(
      "src",
      "/api/organisations/technative/templates/Summer/screenshot",
    );
  });

  it("offers a live demo only where there is one", async () => {
    serve({ [templates]: { body: gallery() } });
    at("/o/technative/new/template", galleryRoute);

    const [kitchenCard, , summerCard] = (await screen.findAllByRole("article")) as [HTMLElement, HTMLElement, HTMLElement];
    expect(within(kitchenCard).queryByRole("link", { name: /Live demo/ })).toBeNull();
    expect(within(summerCard).getByRole("link", { name: /Live demo/ })).toHaveAttribute(
      "href",
      "https://mipmip.github.io/summer-qremix/",
    );
    expect(within(kitchenCard).getByText("KitchenSink", { selector: "span" })).toBeInTheDocument();
  });

  it("searches name, description and author, and says how many it shows", async () => {
    serve({ [templates]: { body: gallery() } });
    at("/o/technative/new/template", galleryRoute);
    await screen.findAllByRole("article");

    await fill(screen.getByPlaceholderText("Search name, description or author"), "RECIPE");
    expect(screen.getAllByRole("article").map((card) => card.getAttribute("aria-label"))).toEqual(["Summer"]);
    expect(screen.getByText("1 of 3 templates")).toBeInTheDocument();

    await userEvent.clear(screen.getByPlaceholderText("Search name, description or author"));
    await fill(screen.getByPlaceholderText("Search name, description or author"), "someone");
    expect(screen.getAllByRole("article").map((card) => card.getAttribute("aria-label"))).toEqual(["Paper"]);
  });

  it("says when nothing matches", async () => {
    serve({ [templates]: { body: gallery() } });
    at("/o/technative/new/template", galleryRoute);
    await screen.findAllByRole("article");

    await fill(screen.getByPlaceholderText("Search name, description or author"), "spaceship");
    expect(screen.queryAllByRole("article")).toHaveLength(0);
    expect(screen.getByText('No templates match "spaceship".')).toBeInTheDocument();
  });

  it("says when the list could not be loaded, and tries again", async () => {
    const calls = serve({ [templates]: { body: gallery({ available: false, templates: [] }) } });
    const user = userEvent.setup();
    at("/o/technative/new/template", galleryRoute);

    const alert = await screen.findByRole("alert");
    expect(within(alert).getByText("The template list couldn't be loaded")).toBeInTheDocument();
    expect(within(alert).getByRole("link", { name: "Bring in a site from git instead" })).toBeInTheDocument();

    const before = calls.filter((call) => call.path === templates).length;
    await user.click(within(alert).getByRole("button", { name: "Try again" }));
    await vi.waitFor(() => expect(calls.filter((call) => call.path === templates).length).toBe(before + 1));
  });

  it("says when the installation offers no templates", async () => {
    serve({ [templates]: { body: gallery({ offered: false, available: false, templates: [] }) } });
    at("/o/technative/new/template", galleryRoute);

    expect(await screen.findByText(/This installation offers no templates/)).toBeInTheDocument();
  });

  it("is refused to somebody who may not make projects", async () => {
    serve({
      [templates]: {
        ok: false,
        status: 403,
        body: { error: { code: "forbidden", message: "You are not allowed to do that." } },
      },
    });
    at("/o/technative/new/template", galleryRoute);

    expect(await screen.findByRole("alert")).toHaveTextContent("You are not allowed to do that.");
    expect(screen.queryAllByRole("article")).toHaveLength(0);
  });

  it("leads each card to its next step", async () => {
    serve({ [templates]: { body: gallery() } });
    at("/o/technative/new/template", galleryRoute);

    expect(await screen.findByRole("link", { name: "Use template: Summer" })).toHaveAttribute(
      "href",
      "/o/technative/new/template/Summer",
    );
  });
});

describe("naming a project made from a template", () => {
  const projects = "/api/organisations/technative/projects";

  function usage(held: number, allowed?: number) {
    return {
      projects: [],
      usage: {
        members: { held: 1, full: false },
        projects: { held, allowed, full: allowed !== undefined && held >= allowed },
      },
    };
  }

  it("shows the template's details and the Hugo it asks for", async () => {
    serve({ [templates]: { body: gallery() }, [projects]: { body: usage(1) } });
    at("/o/technative/new/template/Summer", stepRoute);

    expect(await screen.findByRole("heading", { name: "Start from Summer" })).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Connor Bär (Ported by Pim Snel)" })).toBeInTheDocument();
    expect(screen.getByText("menu")).toBeInTheDocument();
    expect(screen.getByLabelText("Project name")).toHaveValue("Summer");
    expect(screen.getByRole("note")).toHaveTextContent(
      "This template asks for Hugo 0.101.0 extended. This installation builds with Hugo 0.139.0 extended for now.",
    );
  });

  it("says the Hugo version simply when it is the one installed", async () => {
    serve({ [templates]: { body: gallery() }, [projects]: { body: usage(1) } });
    at("/o/technative/new/template/KitchenSink", stepRoute);

    expect(await screen.findByRole("note")).toHaveTextContent(
      "Builds with Hugo 0.139.0 extended, the version this template asks for.",
    );
  });

  it("lets a template that asks for another Hugo go ahead", async () => {
    serve({ [templates]: { body: gallery() }, [projects]: { body: usage(1) } });
    at("/o/technative/new/template/Summer", stepRoute);

    expect(await screen.findByRole("note")).toHaveTextContent("asks for Hugo 0.101.0 extended");
    expect(screen.getByRole("button", { name: "Continue to the key" })).toBeEnabled();
  });

  it("says what is missing instead of greying the button out", async () => {
    const calls = serve({ [templates]: { body: gallery() }, [projects]: { body: usage(1) } });
    const user = userEvent.setup();
    at("/o/technative/new/template/Summer", stepRoute);

    await screen.findByRole("heading", { name: "Start from Summer" });
    await user.clear(screen.getByLabelText("Project name"));
    await fill(screen.getByLabelText("Project name"), "!!!");
    await fill(screen.getByLabelText("Empty git address for your copy"), "git@github.com:technative/recipes.git");
    await user.click(screen.getByRole("button", { name: "Continue to the key" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("A project needs a name with at least one letter or digit.");
    expect(calls.some((call) => call.method === "POST")).toBe(false);
  });

  it("warns when this uses the last free project", async () => {
    serve({ [templates]: { body: gallery() }, [projects]: { body: usage(4, 5) } });
    at("/o/technative/new/template/Summer", stepRoute);

    expect(await screen.findByText("This uses the last free project in technative.")).toBeInTheDocument();
  });

  it("refuses to continue when there is no project left", async () => {
    serve({ [templates]: { body: gallery() }, [projects]: { body: usage(5, 5) } });
    at("/o/technative/new/template/Summer", stepRoute);

    await screen.findByRole("heading", { name: "Start from Summer" });
    await fill(screen.getByLabelText("Empty git address for your copy"), "git@github.com:technative/recipes.git");
    await vi.waitFor(() => expect(screen.getByRole("button", { name: "Continue to the key" })).toBeDisabled());
    expect(screen.getByText(/5/, { selector: "[role=alert]" })).toBeInTheDocument();
  });

  it("says when the name is already taken", async () => {
    serve({
      [templates]: { body: gallery() },
      [projects]: { body: usage(1) },
      [`POST ${projects}`]: {
        ok: false,
        status: 409,
        body: { error: { code: "refused", key: "refused.address-already-taken", message: "That address is already taken." } },
      },
    });
    const user = userEvent.setup();
    at("/o/technative/new/template/Summer", stepRoute);

    await screen.findByRole("heading", { name: "Start from Summer" });
    await fill(screen.getByLabelText("Empty git address for your copy"), "git@github.com:technative/recipes.git");
    await user.click(screen.getByRole("button", { name: "Continue to the key" }));

    expect(await screen.findByRole("alert")).toHaveTextContent("That address is already taken.");
  });

  it("makes the project and goes on to the key", async () => {
    const calls = serve({
      [templates]: { body: gallery() },
      [projects]: { body: usage(1) },
      [`POST ${projects}`]: {
        status: 201,
        body: { slug: "recipes", name: "Recipes", domain: "", yourRole: "implementer", youMay: [], unpublished: 0 },
      },
    });
    const user = userEvent.setup();
    at("/o/technative/new/template/Summer", stepRoute);

    await screen.findByRole("heading", { name: "Start from Summer" });
    await user.clear(screen.getByLabelText("Project name"));
    await fill(screen.getByLabelText("Project name"), "Recipes");
    await fill(screen.getByLabelText("Empty git address for your copy"), "git@github.com:technative/recipes.git");
    await user.click(screen.getByRole("button", { name: "Continue to the key" }));

    expect(await screen.findByText("the setup screen")).toBeInTheDocument();
    const made = calls.find((call) => call.method === "POST" && call.path === projects);
    expect(made?.body).toEqual({ slug: "recipes", name: "Recipes", domain: "" });
  });

  it("says when the template is not in the list", async () => {
    serve({ [templates]: { body: gallery() }, [projects]: { body: usage(1) } });
    at("/o/technative/new/template/Nobody", stepRoute);

    expect(await screen.findByText("That template is not in the list any more.")).toBeInTheDocument();
  });
});

describe("copying the template in during setup", () => {
  const base = "/api/organisations/technative/projects/recipes";
  const nothingYet = { kind: "none", isHugoSite: false, hasModel: false, ready: false, pages: 0, collections: 0 };
  const arrived = { kind: "template", isHugoSite: true, hasModel: true, ready: true, pages: 4, collections: 1 };

  function setUp(publicKey: string, calls: Call[] = []) {
    let site: unknown = nothingYet;
    serve(
      {
        [`${base}/key`]: { body: { publicKey, advice: "" } },
        [`${base}/publishing`]: { body: { to: "main-line", form: "source" } },
        get [`${base}/site`]() {
          return { body: site };
        },
        [`POST ${base}/import`]: {
          status: 201,
          get body() {
            site = arrived;
            return arrived;
          },
        },
      },
      calls,
    );
    return render(
      <Inside>
        <MemoryRouter
          initialEntries={["/o/technative/p/recipes/setup?template=Summer&remote=git%40github.com%3Atechnative%2Frecipes.git"]}
        >
          <Routes>
            <Route path="/o/:organisation/p/:project/setup" element={<OnboardingPage />} />
          </Routes>
        </MemoryRouter>
      </Inside>,
    );
  }

  it("waits for the key before copying", async () => {
    setUp("");
    expect(await screen.findByText(/Starting from Summer, copied into git@github.com:technative\/recipes.git/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Copy the template in" })).toBeDisabled();

    const steps = screen.getByRole("navigation", { name: "Setting up this project" });
    expect(within(steps).getByText("Template")).toBeInTheDocument();
    expect(within(steps).getByText("Name & place in git")).toBeInTheDocument();
  });

  it("lets the template be copied as soon as a key is made", async () => {
    let made = "";
    serve({
      get [`${base}/key`]() {
        return { body: { publicKey: made, advice: "" } };
      },
      [`POST ${base}/key`]: {
        status: 201,
        get body() {
          made = "ssh-ed25519 AAAAexample chatcms";
          return { publicKey: made, advice: "" };
        },
      },
      [`${base}/publishing`]: { body: { to: "main-line", form: "source" } },
      [`${base}/site`]: { body: nothingYet },
    });
    render(
      <Inside>
        <MemoryRouter initialEntries={["/o/technative/p/recipes/setup?template=Summer&remote=git%40example.org%3Ar.git"]}>
          <Routes>
            <Route path="/o/:organisation/p/:project/setup" element={<OnboardingPage />} />
          </Routes>
        </MemoryRouter>
      </Inside>,
    );
    const user = userEvent.setup();

    const copy = await screen.findByRole("button", { name: "Copy the template in" });
    expect(copy).toBeDisabled();
    await user.click(await screen.findByRole("button", { name: "Make a key" }));
    await vi.waitFor(() => expect(copy).toBeEnabled());
  });

  it("copies the template in once there is a key", async () => {
    const calls: Call[] = [];
    setUp("ssh-ed25519 AAAAexample chatcms", calls);
    const user = userEvent.setup();

    const copy = await screen.findByRole("button", { name: "Copy the template in" });
    await vi.waitFor(() => expect(copy).toBeEnabled());
    await user.click(copy);

    expect(await screen.findByText("Ready to edit")).toBeInTheDocument();
    const imported = calls.find((call) => call.method === "POST" && call.path === `${base}/import`);
    expect(imported?.body).toEqual({
      kind: "template",
      address: "Summer",
      remote: "git@github.com:technative/recipes.git",
    });
  });
});

describe("the Hugo version in project settings", () => {
  const base = "/api/organisations/technative/projects/recipes";

  function setUp(hugo: unknown, calls: Call[] = []) {
    serve(
      {
        [base]: {
          body: { slug: "recipes", name: "Recipes", domain: "", yourRole: "implementer", youMay: ["change-settings"], unpublished: 0 },
        },
        [`${base}/key`]: { body: { publicKey: "", advice: "" } },
        [`${base}/publishing`]: { body: { to: "main-line", form: "source" } },
        [`${base}/hugo`]: { body: hugo },
        [`PUT ${base}/hugo`]: { body: { version: "0.139.0", extended: true, installed } },
      },
      calls,
    );
    return render(
      <Inside>
        <MemoryRouter initialEntries={["/o/technative/p/recipes/settings"]}>
          <Routes>
            <Route path="/o/:organisation/p/:project/settings" element={<ProjectSettingsPage />} />
          </Routes>
        </MemoryRouter>
      </Inside>,
    );
  }

  it("says when the version differs from the one installed", async () => {
    setUp({ version: "0.101.0", extended: true, installed });

    expect(await screen.findByLabelText("Hugo version")).toHaveValue("0.101.0");
    expect(screen.getByRole("note")).toHaveTextContent(
      "This installation builds with Hugo 0.139.0 extended for now, whatever is set here.",
    );
  });

  it("says nothing when it is the one installed, and saves a change", async () => {
    const calls: Call[] = [];
    setUp({ version: "0.139.0", extended: true, installed }, calls);
    const user = userEvent.setup();

    const field = await screen.findByLabelText("Hugo version");
    expect(screen.queryByRole("note")).toBeNull();

    await user.clear(field);
    await fill(field, "0.139.0");
    const form = field.closest("form") as HTMLFormElement;
    await user.click(within(form).getByRole("button", { name: "Save Hugo version" }));

    await vi.waitFor(() =>
      expect(calls.find((call) => call.method === "PUT" && call.path === `${base}/hugo`)?.body).toEqual({
        version: "0.139.0",
        extended: true,
      }),
    );
  });
});
