import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { MemoryRouter } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { Project } from "../../lib/tenancy";
import { EditorShell } from "../EditorShell";
import { OrganisationAdminShell } from "../OrganisationAdminShell";
import { OrganisationShell } from "../OrganisationShell";
import { PhoneShell } from "../PhoneShell";
import { PlatformAdminShell } from "../PlatformAdminShell";
import { ProjectShell } from "../ProjectShell";
import { Inside } from "../../routes/__tests__/here";
import { canSeeSetup, canSeeTechnicalDetail } from "../roles";
import { phoneWidth, useIsPhone } from "../useIsPhone";

// The shells are given a project the way the server describes one.
const theViewer = { name: "Joris Bakker", role: "owner" };

const project: Project = {
  slug: "korenaar",
  name: "Korenaar Bakery",
  domain: "korenaar.example",
  yourRole: "owner",
  youMay: ["read-content", "edit-content", "publish", "change-settings"],
  unpublished: 4,
};

function at(path: string, element: ReactNode) {
  return render(
    <Inside>
      <MemoryRouter initialEntries={[path]}>{element}</MemoryRouter>
    </Inside>,
  );
}

const sitePanel = () => document.querySelector("[data-panel='site']")!;

beforeEach(() => {
  window.localStorage.clear();
});

describe("the site panel", () => {
  const shell = (role: string) =>
    at(
      "/o/technative/p/korenaar/content",
      <ProjectShell project={project} viewer={{ name: "Joris", role }} title="Content">
        body
      </ProjectShell>,
    );

  it("folds to a rail of the same entries and unfolds again", async () => {
    const user = userEvent.setup();
    shell("implementer");

    await user.click(screen.getByRole("button", { name: "Hide the navigation" }));
    const rail = screen.getByRole("complementary", { name: "Project navigation" });
    expect(rail).toHaveAttribute("data-folded", "true");
    for (const label of ["Content", "Media", "Publish", "History", "People & roles", "Project settings"]) {
      expect(within(rail).getByRole("link", { name: label })).toBeInTheDocument();
    }
    expect(within(rail).getByRole("button", { name: /Korenaar Bakery/ })).toBeInTheDocument();

    await user.click(within(rail).getByRole("button", { name: "Show the navigation" }));
    expect(sitePanel()).toHaveAttribute("data-folded", "false");
    expect(screen.getByText("Korenaar Bakery")).toBeInTheDocument();
  });

  it("leaves setup out of the rail for an editor", async () => {
    const user = userEvent.setup();
    shell("editor");
    await user.click(screen.getByRole("button", { name: "Hide the navigation" }));
    expect(screen.queryByRole("link", { name: "People & roles" })).toBeNull();
    expect(screen.queryByRole("group", { name: "Setup" })).toBeNull();
  });

  it("stays folded after a remount", async () => {
    const user = userEvent.setup();
    const first = shell("editor");
    await user.click(screen.getByRole("button", { name: "Hide the navigation" }));
    first.unmount();
    shell("editor");
    expect(sitePanel()).toHaveAttribute("data-folded", "true");
  });
});

describe("ProjectShell", () => {
  it("shows the site navigation", () => {
    at("/o/technative/p/korenaar/content", <ProjectShell project={project} viewer={theViewer} title="Content">body</ProjectShell>);

    const navigation = screen.getByRole("complementary", { name: "Project navigation" });
    for (const label of ["Content", "Media", "Publish", "History"]) {
      expect(within(navigation).getByRole("link", { name: new RegExp(label) })).toBeInTheDocument();
    }
  });

  it("shows how much is waiting to be published", () => {
    at("/o/technative/p/korenaar/content", <ProjectShell project={project} viewer={theViewer} title="Content">body</ProjectShell>);

    const publish = screen.getAllByRole("link", { name: /Publish/ });
    expect(publish.some((link) => link.textContent?.includes(String(project.unpublished)))).toBe(true);
  });

  it("leaves setup out for an editor rather than showing it greyed out", () => {
    at(
      "/o/technative/p/korenaar/content",
      <ProjectShell project={project} viewer={{ name: "Marieke", role: "editor" }} title="Content">
        body
      </ProjectShell>,
    );

    expect(screen.queryByRole("link", { name: /People & roles/ })).toBeNull();
    expect(screen.queryByRole("link", { name: /Project settings/ })).toBeNull();
    expect(screen.queryByText(/Setup/)).toBeNull();
  });

  it("gives the content the room and lets the navigation yield", () => {
    const { container } = at(
      "/o/technative/p/korenaar/content",
      <ProjectShell project={project} viewer={theViewer} title="Content">body</ProjectShell>,
    );

    const shell = container.querySelector("[data-testid='project-shell']")!;
    expect(shell.className).toContain("flex-wrap");

    const content = shell.querySelector("[class*='grow-[999]']");
    expect(content).not.toBeNull();
    expect(content!.className).toContain("min-w-0");
  });

  it("is hidden at phone widths", () => {
    const { container } = at(
      "/o/technative/p/korenaar/content",
      <ProjectShell project={project} viewer={theViewer} title="Content">body</ProjectShell>,
    );

    const shell = container.querySelector("[data-testid='project-shell']")!;
    expect(shell.className).toContain("hidden");
    expect(shell.className).toContain("sm:flex");
  });
});

describe("PhoneShell", () => {
  it("has a tab bar for the site sections", () => {
    at("/o/technative/p/korenaar/content", <PhoneShell project={project} viewer={theViewer}>body</PhoneShell>);

    const tabs = screen.getByRole("navigation", { name: "Site sections" });
    for (const label of ["Content", "Media", "Publish", "History"]) {
      expect(within(tabs).getByRole("link", { name: new RegExp(label) })).toBeInTheDocument();
    }
  });

  it("keeps its bottom bar clear of the device", () => {
    at("/o/technative/p/korenaar/content", <PhoneShell project={project} viewer={theViewer}>body</PhoneShell>);

    expect(screen.getByTestId("phone-tab-bar").className).toContain("pb-safe");
  });

  it("gives way to the editing actions while a page is being edited", () => {
    at(
      "/o/technative/p/korenaar/content/about",
      <PhoneShell project={project} viewer={theViewer} bottomBar={<button type="button">Publish</button>}>
        body
      </PhoneShell>,
    );

    expect(screen.queryByTestId("phone-tab-bar")).toBeNull();
    expect(screen.getByTestId("phone-action-bar").className).toContain("pb-safe-2");
  });

  it("is hidden above phone widths", () => {
    const { container } = at("/o/technative/p/korenaar/content", <PhoneShell project={project} viewer={theViewer}>body</PhoneShell>);

    expect(container.querySelector("[data-testid='phone-shell']")!.className).toContain("sm:hidden");
  });
});

describe("OrganisationShell", () => {
  it("uses a header rather than the project navigation", () => {
    at(
      "/projects",
      <OrganisationShell organisation={{ name: "TechNative" }} viewer={theViewer}>
        body
      </OrganisationShell>,
    );

    expect(screen.getByRole("button", { name: /TechNative/ })).toBeInTheDocument();
    expect(screen.queryByRole("complementary", { name: "Project navigation" })).toBeNull();
  });

  it("gives every organisation in the switcher somewhere to go", async () => {
    // It listed them as text once, so the switcher said where somebody
    // could go and took them nowhere.
    const user = userEvent.setup();
    at(
      "/o/technative/projects",
      <OrganisationShell
        organisation={{ slug: "technative", name: "TechNative" }}
        organisations={[
          { slug: "technative", name: "TechNative" },
          { slug: "bakkerij-zuid", name: "Bakkerij Zuid" },
        ]}
        viewer={theViewer}
      >
        body
      </OrganisationShell>,
    );

    await user.click(screen.getByRole("button", { name: /TechNative/ }));

    const elsewhere = await screen.findByRole("menuitem", { name: /Bakkerij Zuid/ });
    expect(elsewhere).toHaveAttribute("href", "/o/bakkerij-zuid/projects");

    const here = screen.getByRole("menuitem", { name: /TechNative/ });
    expect(here).toHaveAttribute("href", "/o/technative/projects");
    expect(here).toHaveAttribute("aria-current", "true");
  });

  it("says which organisation it is when there is only the one", async () => {
    const user = userEvent.setup();
    at(
      "/projects",
      <OrganisationShell organisation={{ slug: "technative", name: "TechNative" }} viewer={theViewer}>
        body
      </OrganisationShell>,
    );

    await user.click(screen.getByRole("button", { name: /TechNative/ }));

    expect(await screen.findByRole("menuitem", { name: "TechNative" })).toBeInTheDocument();
  });
});

describe("EditorShell", () => {
  function editor(layout: "form" | "split" | "preview" = "split") {
    return at(
      "/o/technative/p/korenaar/content/about",
      <EditorShell
        project={project}
        viewer={theViewer}
        title="Home page"
        layout={layout}
        onLayoutChange={() => {}}
        form={<p>Fields</p>}
        preview={<p>Preview</p>}
      />,
    );
  }

  it("has a slot for the form, the preview, and the one that comes later", () => {
    editor();

    expect(screen.getByTestId("editor-slot-form")).toBeInTheDocument();
    expect(screen.getByTestId("editor-slot-preview")).toBeInTheDocument();

    const assistant = screen.getByTestId("editor-slot-assistant");
    expect(assistant).toBeInTheDocument();
    expect(assistant).not.toBeVisible();
  });

  it("shows the navigation as it was left on a project screen", async () => {
    const user = userEvent.setup();
    const list = at("/o/technative/p/korenaar/content", <ProjectShell project={project} viewer={theViewer} title="Content">body</ProjectShell>);
    expect(sitePanel()).toHaveAttribute("data-folded", "false");
    list.unmount();

    const unfolded = editor();
    expect(sitePanel()).toHaveAttribute("data-folded", "false");
    await user.click(screen.getByRole("button", { name: "Hide the navigation" }));
    unfolded.unmount();

    at("/o/technative/p/korenaar/content", <ProjectShell project={project} viewer={theViewer} title="Content">body</ProjectShell>);
    expect(sitePanel()).toHaveAttribute("data-folded", "true");
  });

  it("shows one column when asked for one", () => {
    const form = editor("form");
    expect(screen.queryByTestId("editor-slot-preview")).toBeNull();
    form.unmount();

    editor("preview");
    expect(screen.queryByTestId("editor-slot-form")).toBeNull();
  });

  it("keeps the assistant reachable but says it is not ready", () => {
    editor();
    expect(screen.getByRole("button", { name: /Assistant/ })).toBeDisabled();
  });
});

describe("roles", () => {
  it("decides who sees setup and technical detail", () => {
    // There is no owner of a project; somebody who administers the
    // organisation reaches it as its implementer.
    const expected: Record<string, boolean> = {
      implementer: true,
      editor: false,
      viewer: false,
    };

    for (const [role, allowed] of Object.entries(expected)) {
      expect(canSeeSetup(role)).toBe(allowed);
      expect(canSeeTechnicalDetail(role)).toBe(allowed);
    }
  });
});

describe("useIsPhone", () => {
  it("reads the same width the stylesheet uses", () => {
    expect(phoneWidth).toBe(640);
  });

  it("says no when the window is wide", () => {
    const { result } = renderHook(() => useIsPhone());
    expect(result).toBe(false);
  });

  it("says yes when the window is narrow", () => {
    const original = window.matchMedia;
    window.matchMedia = ((query: string) =>
      ({
        matches: true,
        media: query,
        addEventListener: () => {},
        removeEventListener: () => {},
      }) as unknown as MediaQueryList) as typeof window.matchMedia;

    const { result } = renderHook(() => useIsPhone());
    expect(result).toBe(true);

    window.matchMedia = original;
  });
});

function renderHook<T>(use: () => T): { result: T } {
  const box = { result: undefined as unknown as T };

  function Probe() {
    box.result = use();
    return null;
  }

  render(<Probe />);
  return box;
}

describe("the phone frame and sideways scrolling", () => {
  it("clips sideways rather than letting the page move", () => {
    at("/o/technative/p/korenaar/content", <PhoneShell project={project} viewer={theViewer}>body</PhoneShell>);

    const shell = screen.getByTestId("phone-shell");
    expect(shell.className).toContain("overflow-x-clip");

    const main = shell.querySelector("main")!;
    expect(main.className).toContain("overflow-x-clip");
  });
});

describe("the way out", () => {
  // This defect was one shell having the link and three having a picture
  // that did nothing, so the test walks all four rather than the one that
  // happened to be wrong.
  const everyShell: [string, ReactNode][] = [
    ["ProjectShell", <ProjectShell project={project} viewer={theViewer} title="Content">{null}</ProjectShell>],
    ["PhoneShell", <PhoneShell project={project} viewer={theViewer}>{null}</PhoneShell>],
    [
      "EditorShell",
      <EditorShell
        project={project}
        viewer={theViewer}
        title="Home page"
        layout="split"
        onLayoutChange={() => {}}
        form={null}
        preview={null}
      />,
    ],
    [
      "OrganisationShell",
      <OrganisationShell organisation={{ name: "TechNative" }} viewer={theViewer}>{null}</OrganisationShell>,
    ],
    [
      "OrganisationAdminShell",
      <OrganisationAdminShell
        organisation="technative"
        name="TechNative"
        usage={{ members: { held: 2, full: false }, projects: { held: 1, full: false } }}
        viewer={theViewer}
      >
        {null}
      </OrganisationAdminShell>,
    ],
    [
      "PlatformAdminShell",
      <PlatformAdminShell organisations={1} users={2} viewer={theViewer}>
        {null}
      </PlatformAdminShell>,
    ],
  ];

  for (const [name, shell] of everyShell) {
    it(`${name} offers the way to yourself and the way out`, async () => {
      const { unmount } = at("/o/technative/p/korenaar/content", shell);

      const person = screen.getByTestId("person-menu");
      expect(person, `${name} has no person menu`).toBeInTheDocument();

      await userEvent.click(person);

      expect(await screen.findByText("Your account"), `${name} has no way to your account`).toBeInTheDocument();
      expect(screen.getByText("Sign out"), `${name} has no way out`).toBeInTheDocument();
      expect(screen.queryByText("Sign out everywhere")).toBeNull();

      unmount();
    });

    it(`${name} keeps the person menu at the end of its top bar`, () => {
      const { unmount } = at("/o/technative/p/korenaar/content", shell);

      const bar = screen.getByRole("banner");
      const person = screen.getByTestId("person-menu");
      expect(bar, `${name} has no person menu in its top bar`).toContainElement(person);
      expect(bar.lastElementChild).toBe(person);

      unmount();
    });
  }

  it("says the role on the project", async () => {
    at(
      "/o/technative/p/korenaar/content",
      <ProjectShell project={project} viewer={{ name: "Marieke", role: "editor" }} title="Content">
        {null}
      </ProjectShell>,
    );

    expect(screen.getByRole("complementary", { name: "Project navigation" })).not.toHaveTextContent("Marieke");

    await userEvent.click(screen.getByTestId("person-menu"));

    const menu = await screen.findByRole("menu");
    expect(menu).toHaveTextContent("Marieke");
    expect(menu).toHaveTextContent("Editor");
  });

  it("says you are in the platform admin area", async () => {
    at(
      "/administration/organisations",
      <PlatformAdminShell organisations={1} users={2} viewer={theViewer}>
        {null}
      </PlatformAdminShell>,
    );

    expect(within(screen.getByRole("banner")).getByRole("link", { name: "Back to my projects" })).toBeInTheDocument();

    await userEvent.click(screen.getByTestId("person-menu"));

    expect(await screen.findByRole("menu")).toHaveTextContent("Platform admin");
  });
});

describe("the project switcher", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  const theirProjects = {
    projects: [
      project,
      { ...project, slug: "studio-noord", name: "Studio Noord", unpublished: 0 },
      { ...project, slug: "oud", name: "Oud", archived: true },
    ],
    usage: { members: { held: 1, full: false }, projects: { held: 3, full: false } },
  };

  function answering(status: number, body: unknown) {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: status < 400, status, json: async () => body }));
  }

  it("offers the other projects, the current one marked, and the way to all of them", async () => {
    answering(200, theirProjects);
    at("/o/technative/p/korenaar/content", <ProjectShell project={project} viewer={theViewer} title="Content">{null}</ProjectShell>);

    await userEvent.click(screen.getByRole("button", { name: "Korenaar Bakery, switch project" }));

    const menu = await screen.findByRole("menu");
    const other = await within(menu).findByRole("menuitem", { name: /Studio Noord/ });
    expect(other).toHaveAttribute("href", "/o/technative/p/studio-noord/content");

    const current = within(menu).getByRole("menuitem", { name: /Korenaar Bakery/ });
    expect(current).toHaveAttribute("aria-current", "true");
    expect(within(current).getByLabelText("You are in this one")).toBeInTheDocument();

    expect(within(menu).queryByRole("menuitem", { name: /Oud/ })).not.toBeInTheDocument();
    expect(within(menu).getByRole("menuitem", { name: "All projects" })).toHaveAttribute("href", "/o/technative/projects");
  });

  it("says so when the projects cannot be read, and still leads to all of them", async () => {
    answering(500, { code: "internal", message: "no" });
    at("/o/technative/p/korenaar/content", <ProjectShell project={project} viewer={theViewer} title="Content">{null}</ProjectShell>);

    await userEvent.click(screen.getByRole("button", { name: "Korenaar Bakery, switch project" }));

    const menu = await screen.findByRole("menu");
    expect(await within(menu).findByText("Your projects could not be read just now")).toBeInTheDocument();
    expect(within(menu).getByRole("menuitem", { name: "All projects" })).toBeInTheDocument();
  });

  it("opens from the project in the phone header", async () => {
    answering(200, theirProjects);
    at("/o/technative/p/korenaar/content", <PhoneShell project={project} viewer={theViewer}>{null}</PhoneShell>);

    await userEvent.click(screen.getByRole("button", { name: "Korenaar Bakery, switch project" }));

    const menu = await screen.findByRole("menu");
    expect(await within(menu).findByRole("menuitem", { name: /Studio Noord/ })).toHaveAttribute(
      "href",
      "/o/technative/p/studio-noord/content",
    );
  });
});
