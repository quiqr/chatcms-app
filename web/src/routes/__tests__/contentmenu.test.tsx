import { act, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes, useLocation } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { fill } from "./fill";
import { Inside } from "./here";
import { setWidth } from "./width";
import { CollectionView } from "../content/CollectionView";
import { ContentColumn } from "../content/ContentColumn";
import { ContentBeside, ContentScreen } from "../content/ContentScreen";
import { QuickFind } from "../content/QuickFind";
import { Stepper, usePlace } from "../content/usePlace";
import { PageEditor } from "../PageEditor";
import { EditorShell } from "../../shells/EditorShell";

const base = "/api/organisations/technative/projects/journal";

type Call = { path: string; method: string; body?: unknown };
type Answer = { status?: number; ok?: boolean; body: unknown };

function serve(byPath: Record<string, Answer | ((url: string) => Answer)>, calls: Call[] = []) {
  vi.stubGlobal(
    "fetch",
    vi.fn().mockImplementation((url: string, options?: { method?: string; body?: string }) => {
      const method = options?.method ?? "GET";
      calls.push({ path: url, method, body: options?.body ? JSON.parse(options.body) : undefined });
      const path = url.split("?")[0] ?? url;
      const found = byPath[`${method} ${path}`] ?? byPath[path];
      const answer = typeof found === "function" ? found(url) : found;
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


function Where() {
  const location = useLocation();
  return <p data-testid="where">{location.pathname + location.search}</p>;
}

function at(path: string, routes: React.ReactNode) {
  return render(
    <Inside>
      <MemoryRouter initialEntries={[path]}>
        <Routes>{routes}</Routes>
        <Where />
      </MemoryRouter>
    </Inside>,
  );
}

const theMenu = {
  groups: [
    {
      key: "journal",
      title: "Journal",
      items: [
        { key: "home", kind: "single", title: "Home", hasChanges: false },
        { key: "posts", kind: "collection", title: "Posts", itemName: "post", count: 48, hasChanges: true },
      ],
    },
    { key: "people", title: "People", items: [{ key: "team", kind: "collection", title: "Team", itemName: "person", count: 2, hasChanges: false }] },
  ],
  totals: { singles: 1, collections: 2 },
  fallback: false,
};

function posts(overrides: Record<string, unknown> = {}) {
  return {
    key: "posts",
    title: "Posts",
    itemName: "post",
    total: 2,
    all: 2,
    sort: "newest",
    folders: [],
    showsPaths: false,
    hasImages: true,
    items: [
      { key: "posts/spring-sale", title: "Spring sale", status: "changed", date: "2024-03-01T00:00:00Z", author: "Anna", thumbnail: "loaf.png" },
      { key: "posts/autumn", title: "Autumn walk", status: "published", date: "2023-10-01T00:00:00Z" },
    ],
    ...overrides,
  };
}

const project = { slug: "journal", name: "Journal", domain: "", yourRole: "editor", youMay: ["read-content", "edit-content"], unpublished: 1 };

beforeEach(() => {
  document.cookie = "chatcms_request_token=a-token; path=/";
  window.localStorage.clear();
  setWidth(1440);
});

afterEach(() => {
  vi.unstubAllGlobals();
  vi.useRealTimers();
});

describe("the content column", () => {
  it("shows the menu's groups and rows, with counts and what waits", async () => {
    serve({ [`${base}/content-menu`]: { body: theMenu } });
    at("/o/technative/p/journal/content/c/posts", <Route path="*" element={<ContentColumn project="journal" />} />);

    const groups = await screen.findAllByRole("region");
    expect(groups.map((group) => group.getAttribute("aria-label"))).toEqual(["Journal", "People"]);
    expect(screen.getByText("1 page · 2 collections")).toBeInTheDocument();
    const postsRow = screen.getByRole("link", { name: /Posts/ });
    expect(postsRow).toHaveTextContent("48");
    expect(postsRow).toHaveAttribute("aria-current", "page");
    expect(within(postsRow).getByLabelText("Waiting to be published")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Home" })).not.toHaveAttribute("aria-current");
  });

  it("keeps a collection current while one of its items is open", async () => {
    serve({ [`${base}/content-menu`]: { body: theMenu } });
    at("/o/technative/p/journal/content/posts/spring-sale", <Route path="*" element={<ContentColumn project="journal" />} />);

    expect(await screen.findByRole("link", { name: /Posts/ })).toHaveAttribute("aria-current", "page");
  });

  it("folds a group and remembers it", async () => {
    serve({ [`${base}/content-menu`]: { body: theMenu } });
    const user = userEvent.setup();
    const first = at("/o/technative/p/journal/content", <Route path="*" element={<ContentColumn project="journal" />} />);

    await user.click(await screen.findByRole("button", { name: /People/ }));
    expect(screen.queryByRole("link", { name: /Team/ })).toBeNull();
    first.unmount();

    at("/o/technative/p/journal/content", <Route path="*" element={<ContentColumn project="journal" />} />);
    const people = await screen.findByRole("button", { name: /People/ });
    expect(people).toHaveAttribute("aria-expanded", "false");
    expect(people).toHaveTextContent("1");
  });
});

describe("the content routes", () => {
  const routes = (
    <>
      <Route path="/o/:organisation/p/:project/content" element={<ContentScreen />} />
      <Route path="/o/:organisation/p/:project/content/c/:collection" element={<ContentScreen />} />
      <Route path="/o/:organisation/p/:project/content/*" element={<p>the editor</p>} />
    </>
  );

  it("opens the first collection from the content address", async () => {
    serve({
      [`${base}/content-menu`]: { body: theMenu },
      [base]: { body: project },
      [`${base}/collections/posts/items`]: { body: posts() },
    });
    at("/o/technative/p/journal/content", routes);

    await waitFor(() => expect(screen.getByTestId("where")).toHaveTextContent("/content/c/posts"));
    expect(await screen.findByRole("heading", { name: "Posts", level: 2 })).toBeInTheDocument();
  });

  it("still opens an editor link sent before", async () => {
    serve({});
    at("/o/technative/p/journal/content/posts/spring-sale", routes);

    expect(await screen.findByText("the editor")).toBeInTheDocument();
  });
});

describe("a collection", () => {
  const route = (
    <>
      <Route path="/o/:organisation/p/:project/content/c/:collection" element={<CollectionView project="journal" collection="posts" mayEdit />} />
      <Route path="/o/:organisation/p/:project/content/*" element={<p>the editor</p>} />
    </>
  );

  it("lists its items with what the model has", async () => {
    serve({ [`${base}/collections/posts/items`]: { body: posts() } });
    at("/o/technative/p/journal/content/c/posts", route);

    expect(await screen.findByText("2 posts · Newest first")).toBeInTheDocument();
    const row = screen.getByRole("link", { name: "Spring sale" }).closest("li") as HTMLElement;
    expect(within(row).getByText(/Anna/)).toBeInTheDocument();
    expect(within(row).getByText("Changed")).toBeInTheDocument();
    expect(row.querySelector("img")).toHaveAttribute("src", `${base}/images/loaf.png?small=true`);
    expect(screen.getByText("Showing 2 of 2")).toBeInTheDocument();
  });

  it("offers twenty more at a time", async () => {
    const calls = serve({ [`${base}/collections/posts/items`]: { body: posts({ total: 48, all: 48 }) } });
    const user = userEvent.setup();
    at("/o/technative/p/journal/content/c/posts", route);

    await user.click(await screen.findByRole("button", { name: "Show 20 more" }));
    await waitFor(() => expect(calls.some((call) => call.path.includes("pageSize=40"))).toBe(true));
  });

  it("searches after a pause, asking the server once", async () => {
    const calls = serve({ [`${base}/collections/posts/items`]: { body: posts() } });
    at("/o/technative/p/journal/content/c/posts", route);
    const field = await screen.findByLabelText("Search in Posts");

    fireEvent.change(field, { target: { value: "s" } });
    fireEvent.change(field, { target: { value: "sp" } });
    fireEvent.change(field, { target: { value: "spring" } });
    await waitFor(() => expect(screen.getByTestId("where")).toHaveTextContent("q=spring"));
    await waitFor(() => expect(calls.filter((call) => call.path.includes("q=")).length).toBe(1));
    expect(calls.filter((call) => call.path.includes("q=")).at(0)?.path).toContain("q=spring");
  });

  it("says when nothing matches, and clears the search", async () => {
    serve({ [`${base}/collections/posts/items`]: (url) => ({ body: url.includes("q=") ? posts({ items: [], total: 0 }) : posts() }) });
    const user = userEvent.setup();
    at("/o/technative/p/journal/content/c/posts?q=zebra", route);

    expect(await screen.findByText('No posts match "zebra"')).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Clear search" }));
    expect(await screen.findByRole("link", { name: "Spring sale" })).toBeInTheDocument();
  });

  it("says when there are none yet", async () => {
    serve({ [`${base}/collections/posts/items`]: { body: posts({ items: [], total: 0, all: 0 }) } });
    at("/o/technative/p/journal/content/c/posts", route);

    expect(await screen.findByText("No posts yet")).toBeInTheDocument();
  });

  it("offers folders as tabs, a list beyond five, and nothing without", async () => {
    serve({ [`${base}/collections/posts/items`]: { body: posts({ folders: [{ name: "nl", label: "nl", count: 1 }, { name: "en", label: "en", count: 1 }] }) } });
    const tabs = at("/o/technative/p/journal/content/c/posts", route);
    expect(await screen.findByRole("radio", { name: "nl · 1" })).toBeInTheDocument();
    tabs.unmount();

    const many = ["a", "b", "c", "d", "e", "f"].map((name) => ({ name, label: name, count: 1 }));
    serve({ [`${base}/collections/posts/items`]: { body: posts({ folders: many }) } });
    const list = at("/o/technative/p/journal/content/c/posts", route);
    expect(await screen.findByRole("combobox", { name: "Folder" })).toBeInTheDocument();
    list.unmount();

    serve({ [`${base}/collections/posts/items`]: { body: posts() } });
    at("/o/technative/p/journal/content/c/posts", route);
    await screen.findByRole("link", { name: "Spring sale" });
    expect(screen.queryByRole("radiogroup", { name: "Folder" })).toBeNull();
    expect(screen.queryByRole("combobox", { name: "Folder" })).toBeNull();
  });

  it("offers file names only to those who see technical detail", async () => {
    serve({ [`${base}/collections/posts/items`]: { body: posts() } });
    const editor = at("/o/technative/p/journal/content/c/posts", route);
    await screen.findByRole("link", { name: "Spring sale" });
    expect(screen.queryByRole("switch", { name: "File names" })).toBeNull();
    editor.unmount();

    const withPaths = posts({ showsPaths: true });
    const first = (withPaths.items as { path?: string }[])[0];
    if (first) first.path = "content/post/spring-sale/index.md";
    serve({ [`${base}/collections/posts/items`]: { body: withPaths } });
    const user = userEvent.setup();
    at("/o/technative/p/journal/content/c/posts", route);
    await user.click(await screen.findByRole("switch", { name: "File names" }));
    expect(screen.getByText("content/post/spring-sale/index.md")).toBeInTheDocument();
  });

  it("makes an item and opens it", async () => {
    const calls = serve({
      [`${base}/collections/posts/items`]: { body: posts() },
      [`${base}/collections/posts/address`]: { body: { slug: "summer", address: "/post/summer/", free: true } },
      [`POST ${base}/collections/posts/items`]: { status: 201, body: { key: "posts/summer", title: "Summer" } },
    });
    const user = userEvent.setup();
    at("/o/technative/p/journal/content/c/posts", route);

    await user.click(await screen.findByRole("button", { name: "New post" }));
    await fill(screen.getByLabelText("Title"), "Summer");
    const create = screen.getByRole("button", { name: "Create and edit" });
    await waitFor(() => expect(create).toBeEnabled());
    await user.click(create);

    expect(await screen.findByText("the editor")).toBeInTheDocument();
    expect(calls.find((call) => call.method === "POST")?.body).toEqual({ title: "Summer", folder: "", slug: "" });
  });

  it("keeps Create and edit unavailable while the address is taken", async () => {
    serve({
      [`${base}/collections/posts/items`]: { body: posts() },
      [`${base}/collections/posts/address`]: { body: { slug: "spring-sale", address: "/post/spring-sale/", free: false } },
    });
    const user = userEvent.setup();
    at("/o/technative/p/journal/content/c/posts", route);

    await user.click(await screen.findByRole("button", { name: "New post" }));
    await fill(screen.getByLabelText("Title"), "Spring sale");
    expect(await screen.findByText("There's already a post at this address. Change the title or the address.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Create and edit" })).toBeDisabled();
  });

  it("gives a viewer nothing to change", async () => {
    serve({ [`${base}/collections/posts/items`]: { body: posts() } });
    const user = userEvent.setup();
    at(
      "/o/technative/p/journal/content/c/posts",
      <Route path="*" element={<CollectionView project="journal" collection="posts" mayEdit={false} />} />,
    );

    await screen.findByRole("link", { name: "Spring sale" });
    expect(screen.queryByRole("button", { name: "New post" })).toBeNull();
    await user.click(screen.getByRole("button", { name: "What to do with Spring sale" }));
    expect(screen.getByRole("button", { name: "Open" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Rename" })).toBeNull();
  });
});

function PlaceOf({ page }: { page: string }) {
  const place = usePlace("technative", "journal", page);
  return (
    <div>
      <p data-testid="context">{place.context}</p>
      <p data-testid="back">{place.backTo}</p>
      <Stepper place={place} />
    </div>
  );
}

describe("an open item's place", () => {
  it("says where it is and steps through the list as it was", async () => {
    const calls = serve({
      [`${base}/collections/posts/items`]: {
        body: posts({ total: 3, around: { index: 2, previous: "posts/autumn", next: "posts/winter" }, items: [] }),
      },
    });
    at("/o/technative/p/journal/content/posts/spring-sale?q=s&sort=title", <Route path="/o/:organisation/p/:project/content/*" element={<PlaceOf page="posts/spring-sale" />} />);

    expect(await screen.findByText("Posts · 2 of 3")).toBeInTheDocument();
    expect(calls[0]?.path).toContain("around=posts%2Fspring-sale");
    expect(calls[0]?.path).toContain("q=s");
    expect(screen.getByTestId("back")).toHaveTextContent("/content/c/posts?q=s&sort=title");

    fireEvent.keyDown(window, { key: "ArrowDown", altKey: true });
    await waitFor(() => expect(screen.getByTestId("where")).toHaveTextContent("/content/posts/winter?q=s&sort=title"));
  });

  it("cannot step past the ends", async () => {
    serve({ [`${base}/collections/posts/items`]: { body: posts({ around: { index: 1, next: "posts/autumn" }, items: [] }) } });
    at("/o/technative/p/journal/content/posts/spring-sale", <Route path="*" element={<PlaceOf page="posts/spring-sale" />} />);

    expect(await screen.findByText("Posts · 1 of 2")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Previous" })).toBeDisabled();
    expect(screen.getByRole("link", { name: "Next" })).toBeInTheDocument();
  });

  it("names a single's group", async () => {
    serve({ [`${base}/content-menu`]: { body: theMenu } });
    at("/o/technative/p/journal/content/home", <Route path="*" element={<PlaceOf page="home" />} />);

    await waitFor(() => expect(screen.getByTestId("context")).toHaveTextContent("Journal"));
    expect(screen.queryByRole("button", { name: "Previous" })).toBeNull();
  });
});

describe("the content column beside the editor", () => {
  function editor() {
    return at(
      "/o/technative/p/journal/content/posts/spring-sale",
      <Route
        path="*"
        element={
          <EditorShell
            project={project}
            viewer={{ name: "Joris" }}
            title="Spring sale"
            layout="form"
            onLayoutChange={() => {}}
            form={<p>Fields</p>}
            preview={null}
          />
        }
      />,
    );
  }

  const panel = () => document.querySelector("[data-panel='content']")!;

  it("folds to a strip, stays folded, and comes back from the strip", async () => {
    serve({ [`${base}/content-menu`]: { body: theMenu } });
    const user = userEvent.setup();
    const first = editor();

    expect(panel()).toHaveAttribute("data-folded", "false");
    await user.click(await screen.findByRole("button", { name: "Hide the content menu" }));
    expect(panel()).toHaveAttribute("data-folded", "true");
    expect(screen.queryByRole("navigation", { name: "Content" })).toBeNull();
    first.unmount();

    editor();
    expect(panel()).toHaveAttribute("data-folded", "true");
    await user.click(screen.getByRole("button", { name: "Show the content menu" }));
    expect(panel()).toHaveAttribute("data-folded", "false");
    expect(await screen.findByRole("navigation", { name: "Content" })).toBeInTheDocument();
  });

  it("stays folded in the editor when it was folded on the collection view", async () => {
    serve({ [`${base}/content-menu`]: { body: theMenu } });
    const user = userEvent.setup();
    const list = at("/o/technative/p/journal/content/c/posts", <Route path="*" element={<ContentBeside />} />);
    await user.click(await screen.findByRole("button", { name: "Hide the content menu" }));
    list.unmount();

    editor();
    expect(panel()).toHaveAttribute("data-folded", "true");
  });

  it("keeps a fold made before folding worked on every screen", () => {
    window.localStorage.setItem("chatcms.content.folded.person-1", "true");
    serve({ [`${base}/content-menu`]: { body: theMenu } });
    at("/o/technative/p/journal/content/c/posts", <Route path="*" element={<ContentBeside />} />);
    expect(panel()).toHaveAttribute("data-folded", "true");
  });
});

describe("the Show on website switch", () => {
  const page = (fields: unknown[]) => ({ key: "posts/spring-sale", title: "Spring sale", section: "Posts", fields, body: "", hasBody: false });

  it("is the draft field, turned around", async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    const calls = serve({
      [`${base}/pages/posts/spring-sale`]: { body: page([{ path: "title", type: "string", title: "Title", value: "Spring sale" }, { path: "draft", type: "boolean", title: "Draft", value: false }]) },
      [`PUT ${base}/pages/posts/spring-sale`]: { body: page([]) },
    });
    at("/o/technative/p/journal/content/posts/spring-sale", <Route path="/o/:organisation/p/:project/content/*" element={<PageEditor organisation="technative" />} />);

    const toggle = await screen.findByRole("switch", { name: /Show on website/ });
    expect(toggle).toBeChecked();
    expect(screen.queryByText("Draft")).toBeNull();
    fireEvent.click(toggle);
    await act(async () => {
      vi.advanceTimersByTime(1000);
    });
    await waitFor(() => expect(calls.find((call) => call.method === "PUT")?.body).toEqual({ values: { draft: true } }));
  });

  it("is absent without the field", async () => {
    serve({ [`${base}/pages/posts/spring-sale`]: { body: page([{ path: "title", type: "string", title: "Title", value: "x" }]) } });
    at("/o/technative/p/journal/content/posts/spring-sale", <Route path="/o/:organisation/p/:project/content/*" element={<PageEditor organisation="technative" />} />);

    await screen.findByLabelText("Title");
    expect(screen.queryByRole("switch", { name: /Show on website/ })).toBeNull();
  });
});

describe("quick find", () => {
  const found = {
    groups: [
      { title: "Posts", items: [{ key: "posts/spring-sale", title: "Spring sale", kind: "item" }] },
      { title: "Journal", items: [{ key: "home", title: "Spring home", kind: "single" }] },
    ],
  };

  it("opens on Ctrl+K, groups what it finds, and Enter opens the first", async () => {
    serve({ [`${base}/search`]: (url) => ({ body: url.includes("q=nothing") ? { groups: [] } : found }) });
    const user = userEvent.setup();
    at("/o/technative/p/journal/content", <Route path="*" element={<QuickFind project="journal" />} />);

    fireEvent.keyDown(window, { key: "k", ctrlKey: true });
    const field = await screen.findByLabelText("Title");
    await user.type(field, "spring");
    expect(await screen.findByRole("region", { name: "Posts" })).toBeInTheDocument();
    expect(screen.getByRole("region", { name: "Journal" })).toBeInTheDocument();
    await user.keyboard("{Enter}");
    await waitFor(() => expect(screen.getByTestId("where")).toHaveTextContent("/content/posts/spring-sale"));
  });

  it("says when nothing is found", async () => {
    serve({ [`${base}/search`]: { body: { groups: [] } } });
    const user = userEvent.setup();
    at("/o/technative/p/journal/content", <Route path="*" element={<QuickFind project="journal" />} />);

    fireEvent.keyDown(window, { key: "k", metaKey: true });
    await user.type(await screen.findByLabelText("Title"), "nothing");
    expect(await screen.findByText("Nothing has that in its title.")).toBeInTheDocument();
  });
});

describe("narrower screens", () => {
  const screenAndPanel = (
    <>
      <ContentBeside />
      <ContentScreen />
    </>
  );
  const routes = (
    <>
      <Route path="/o/:organisation/p/:project/content" element={screenAndPanel} />
      <Route path="/o/:organisation/p/:project/content/c/:collection" element={screenAndPanel} />
    </>
  );

  it("opens the content menu as a drawer that closes when something is chosen", async () => {
    setWidth(800);
    serve({
      [`${base}/content-menu`]: { body: theMenu },
      [base]: { body: project },
      [`${base}/collections/posts/items`]: { body: posts() },
      [`${base}/collections/team/items`]: { body: posts({ key: "team", title: "Team" }) },
    });
    const user = userEvent.setup();
    at("/o/technative/p/journal/content/c/posts", routes);

    expect(document.querySelector("[data-panel='content']")).toHaveAttribute("data-folded", "true");
    await user.click(await screen.findByRole("button", { name: "Show the content menu" }));
    const drawer = await screen.findByRole("dialog", { name: "Content menu" });
    await user.click(await within(drawer).findByRole("link", { name: /Team/ }));
    expect(screen.queryByRole("dialog", { name: "Content menu" })).toBeNull();
    await waitFor(() => expect(screen.getByTestId("where")).toHaveTextContent("/content/c/team"));
  });

  it("is a menu screen, then a collection screen, on a phone", async () => {
    setWidth(390);
    serve({
      [`${base}/content-menu`]: { body: theMenu },
      [base]: { body: project },
      [`${base}/collections/posts/items`]: { body: posts() },
    });
    const user = userEvent.setup();
    at("/o/technative/p/journal/content", routes);

    await user.click(await screen.findByRole("link", { name: /Posts/ }));
    expect(await screen.findByRole("heading", { name: "Posts", level: 2 })).toBeInTheDocument();
    expect(screen.queryByText("Published")).toBeNull();
    expect(screen.getByRole("link", { name: "Back to the content menu" })).toBeInTheDocument();
  });
});
