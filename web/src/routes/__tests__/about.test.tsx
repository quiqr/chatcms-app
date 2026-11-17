import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { afterEach, describe, expect, it, vi } from "vitest";

import { Inside } from "./here";
import { AboutScreen, versionLine } from "../AboutScreen";
import { PlatformAdminShell } from "../../shells/PlatformAdminShell";

const theAnswer = {
  version: "0.1.0",
  source: "a3bab43",
  changed: false,
  describe: "0.1.0 · a3bab43",
  runsOn: { hugo: "0.152.2", git: "2.51.0", postgres: "", go: "go1.26.9" },
  license: { name: "Apache 2.0", text: "Apache License\nVersion 2.0, January 2004" },
  links: {
    source: "https://github.com/quiqr/chatcms-app",
    changelog: "https://github.com/quiqr/chatcms-app/blob/main/CHANGELOG.md",
    issues: "https://github.com/quiqr/chatcms-app/issues",
    contributors: "https://github.com/quiqr/chatcms-app/graphs/contributors",
  },
};

function serve(body: unknown) {
  vi.stubGlobal(
    "fetch",
    vi.fn(() => Promise.resolve({ ok: true, status: 200, json: async () => body })),
  );
}

function renderAbout() {
  return render(
    <Inside>
      <MemoryRouter>
        <AboutScreen />
      </MemoryRouter>
    </Inside>,
  );
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("the about screen", () => {
  it("says the version and the source it was built from", async () => {
    serve(theAnswer);
    renderAbout();
    const version = await screen.findByRole("region", { name: "Version" });
    expect(version).toHaveTextContent("0.1.0 · a3bab43");
  });

  it("says when the build is a development build or came from changed files", () => {
    expect(versionLine({ version: "0.1.0", source: "", changed: false })).toBe("0.1.0 · development build");
    expect(versionLine({ version: "0.1.0", source: "a3bab43", changed: true })).toBe("0.1.0 · a3bab43, changed");
  });

  it("names what it runs on, and unknown for what did not answer", async () => {
    serve(theAnswer);
    renderAbout();
    const runsOn = await screen.findByRole("region", { name: "Runs on" });
    expect(runsOn).toHaveTextContent("0.152.2");
    expect(runsOn).toHaveTextContent("2.51.0");
    expect(runsOn).toHaveTextContent("go1.26.9");
    expect(within(runsOn).getByText("PostgreSQL").nextElementSibling).toHaveTextContent("unknown");
  });

  it("shows the license text when it is opened", async () => {
    serve(theAnswer);
    renderAbout();
    const license = await screen.findByRole("region", { name: "License" });
    expect(license).toHaveTextContent("Apache 2.0");
    await userEvent.click(within(license).getByRole("button", { name: /Read the license/ }));
    expect(await within(license).findByText(/Version 2.0, January 2004/)).toBeVisible();
  });

  it("links to the source, the changelog, the issues and the contributors, outside the app", async () => {
    serve(theAnswer);
    renderAbout();
    await screen.findByRole("region", { name: "Source code" });
    const links = [
      ["Quiqr ChatCMS on GitHub", theAnswer.links.source],
      ["Changelog", theAnswer.links.changelog],
      ["Report a problem", theAnswer.links.issues],
      ["Everyone who contributed", theAnswer.links.contributors],
    ];
    for (const [name, href] of links) {
      const link = screen.getByRole("link", { name });
      expect(link).toHaveAttribute("href", href);
      expect(link).toHaveAttribute("target", "_blank");
      expect(link).toHaveAttribute("rel", "noreferrer");
    }
    expect(screen.getByText("Made by the Quiqr Team, with help from its contributors.")).toBeInTheDocument();
  });

  it("says so when it cannot be read", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(() =>
        Promise.resolve({
          ok: false,
          status: 403,
          json: async () => ({ error: { code: "forbidden", message: "You are not allowed to do that." } }),
        }),
      ),
    );
    renderAbout();
    expect(await screen.findByRole("alert")).toBeInTheDocument();
  });

  it("is an entry in the platform admin area, after History", () => {
    render(
      <Inside>
        <MemoryRouter>
          <PlatformAdminShell organisations={1} users={2} viewer={{ name: "Pim" }}>
            body
          </PlatformAdminShell>
        </MemoryRouter>
      </Inside>,
    );
    const links = screen.getAllByRole("link").map((link) => link.textContent ?? "");
    const history = links.findIndex((text) => text.includes("History"));
    const about = links.findIndex((text) => text.includes("About"));
    expect(about).toBeGreaterThan(history);
    expect(screen.getByRole("link", { name: "About" })).toHaveAttribute("href", "/administration/about");
  });
});
