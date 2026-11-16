import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router";

import { fill } from "./fill";
import { Inside } from "./here";import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { ContentPage } from "../ContentPage";

const base = "/api/organisations/technative/projects/korenaar";

const theMenu = {
  groups: [
    {
      key: "pages",
      title: "Pages",
      items: [
        { key: "home", kind: "single", title: "Home page", hasChanges: false },
        { key: "news", kind: "collection", title: "News", itemName: "article", count: 2, hasChanges: true },
      ],
    },
  ],
  totals: { singles: 1, collections: 1 },
  fallback: true,
};

const theNews = {
  key: "news",
  title: "News",
  itemName: "article",
  total: 2,
  folders: [],
  items: [
    { key: "news/one", title: "One", status: "changed", by: "Anna" },
    { key: "news/two", title: "Two", status: "published" },
  ],
};

const theSections = {
  [`${base}/content-menu`]: theMenu,
  [`${base}/collections/news/items`]: theNews,
};

function server(answers: Record<string, unknown>, sent?: (path: string, method: string, body: unknown) => void) {
  return vi.fn().mockImplementation((path: string, options?: RequestInit) => {
    const method = options?.method ?? "GET";
    if (method !== "GET") {
      sent?.(path, method, options?.body ? JSON.parse(String(options.body)) : null);
    }
    const key = method === "GET" ? path : `${method} ${path}`;
    return Promise.resolve({
      ok: true,
      status: method === "DELETE" ? 204 : 200,
      json: async () => answers[key] ?? answers[path] ?? {},
    });
  });
}

function renderPage() {
  return render(
    <Inside><MemoryRouter initialEntries={["/o/technative/p/korenaar/content"]}>
      <Routes>
        <Route path="/o/:organisation/p/:project/content" element={<ContentPage />} />
        <Route path="/o/:organisation/p/:project/content/*" element={<p>the editor</p>} />
        <Route path="/o/:organisation/p/:project/setup" element={<p>the setup screen</p>} />
      </Routes>
    </MemoryRouter></Inside>,
  );
}

beforeEach(() => {
  document.cookie = "chatcms_request_token=a-token; path=/";
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("the content browser", () => {
  it("lists the pages and the collections with their counts", async () => {
    vi.stubGlobal("fetch", server(theSections));
    renderPage();

    expect(await screen.findByRole("heading", { name: /Pages/ })).toBeInTheDocument();
    expect(screen.getByRole("heading", { name: /News/ })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Home page" })).toBeInTheDocument();
    expect(await screen.findByRole("button", { name: "One" })).toBeInTheDocument();
  });

  it("marks a page that has not been published", async () => {
    vi.stubGlobal("fetch", server(theSections));
    renderPage();

    expect(await screen.findByText("Not published")).toBeInTheDocument();
    expect(screen.getByText(/Anna/)).toBeInTheDocument();
  });

  it("names the new button after what the collection holds", async () => {
    vi.stubGlobal("fetch", server(theSections));
    renderPage();

    expect(await screen.findByRole("button", { name: /New article/ })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /New page/ })).not.toBeInTheDocument();
  });

  it("makes a page and opens it", async () => {
    const sent: Array<[string, string, unknown]> = [];
    vi.stubGlobal(
      "fetch",
      server(
        {
          ...theSections,
          [`POST ${base}/content`]: { key: "news/spring-prices", title: "Spring prices", fields: [], body: "", section: "News", hasBody: true },
        },
        (path, method, body) => sent.push([path, method, body]),
      ),
    );
    renderPage();

    await userEvent.click(await screen.findByRole("button", { name: /New article/ }));
    await fill(screen.getByLabelText("Name"), "Spring prices");
    await userEvent.click(screen.getByRole("button", { name: "Make it" }));

    expect(sent).toContainEqual([`${base}/content`, "POST", { section: "news", title: "Spring prices" }]);
    expect(await screen.findByText("the editor")).toBeInTheDocument();
  });

  it("renames a page", async () => {
    const sent: Array<[string, string, unknown]> = [];
    vi.stubGlobal(
      "fetch",
      server(
        { ...theSections, [`POST ${base}/rename/news/one`]: { key: "news/the-first" } },
        (path, method, body) => sent.push([path, method, body]),
      ),
    );
    renderPage();

    await userEvent.click(await screen.findByRole("button", { name: "What to do with One" }));
    await userEvent.click(screen.getByRole("button", { name: "Rename" }));
    await userEvent.clear(screen.getByLabelText("Name"));
    await fill(screen.getByLabelText("Name"), "The first");
    await userEvent.click(screen.getByRole("button", { name: "Save the name" }));

    expect(sent).toContainEqual([`${base}/rename/news/one`, "POST", { title: "The first", copy: false }]);
  });

  it("makes a copy of a page", async () => {
    const sent: Array<[string, string, unknown]> = [];
    vi.stubGlobal(
      "fetch",
      server(
        { ...theSections, [`POST ${base}/rename/news/one`]: { key: "news/one-again" } },
        (path, method, body) => sent.push([path, method, body]),
      ),
    );
    renderPage();

    await userEvent.click(await screen.findByRole("button", { name: "What to do with One" }));
    await userEvent.click(screen.getByRole("button", { name: "Make a copy" }));
    await userEvent.clear(screen.getByLabelText("Name"));
    await fill(screen.getByLabelText("Name"), "One again");
    await userEvent.click(screen.getByRole("button", { name: "Save the name" }));

    expect(sent).toContainEqual([`${base}/rename/news/one`, "POST", { title: "One again", copy: true }]);
  });

  it("asks before removing, and says what will happen", async () => {
    const sent: Array<[string, string, unknown]> = [];
    vi.stubGlobal(
      "fetch",
      server(theSections, (path, method, body) => sent.push([path, method, body])),
    );
    renderPage();

    await userEvent.click(await screen.findByRole("button", { name: "What to do with One" }));
    await userEvent.click(screen.getByRole("button", { name: "Remove…" }));

    const dialog = await screen.findByRole("dialog");
    expect(dialog).toHaveTextContent("the next time you publish");
    expect(dialog).toHaveTextContent("bring it back");

    await userEvent.click(screen.getByRole("button", { name: "Yes, remove it" }));
    await waitFor(() => expect(sent).toContainEqual([`${base}/pages/news/one`, "DELETE", null]));
  });

  it("can be told to keep a page after all", async () => {
    const sent: Array<[string, string, unknown]> = [];
    vi.stubGlobal(
      "fetch",
      server(theSections, (path, method, body) => sent.push([path, method, body])),
    );
    renderPage();

    await userEvent.click(await screen.findByRole("button", { name: "What to do with One" }));
    await userEvent.click(screen.getByRole("button", { name: "Remove…" }));
    await userEvent.click(screen.getByRole("button", { name: "Keep it" }));

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(sent).toHaveLength(0);
  });

  it("offers nothing to change on a page the model names", async () => {
    vi.stubGlobal("fetch", server(theSections));
    renderPage();

    await screen.findByRole("button", { name: "Home page" });
    expect(screen.queryByRole("button", { name: "What to do with Home page" })).not.toBeInTheDocument();
  });

  it("narrows the list as you search", async () => {
    vi.stubGlobal("fetch", server(theSections));
    renderPage();

    await userEvent.type(await screen.findByLabelText("Search"), "two");

    expect(screen.getByRole("button", { name: "Two" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "One" })).not.toBeInTheDocument();
  });

  it("uses no git word anywhere an editor reads", async () => {
    vi.stubGlobal("fetch", server(theSections));
    renderPage();

    await screen.findByRole("button", { name: "Home page" });
    const words = (document.body.textContent ?? "").toLowerCase();
    for (const word of ["commit", "branch", "push", "merge", "repository", "git", "sha", "head"]) {
      expect(words).not.toContain(word);
    }
  });

  it("shows the groups of the model's menu in its order", async () => {
    vi.stubGlobal(
      "fetch",
      server({
        [`${base}/content-menu`]: {
          groups: [
            { key: "site", title: "Our site", items: [{ key: "home", kind: "single", title: "Home page", hasChanges: false }] },
            { key: "blog", title: "Writing", items: [{ key: "posts", kind: "collection", title: "Posts", itemName: "post", count: 1, hasChanges: false }] },
          ],
          totals: { singles: 1, collections: 1 },
          fallback: false,
        },
        [`${base}/collections/posts/items`]: {
          key: "posts", title: "Posts", itemName: "post", total: 1, folders: [{ name: "2024", label: "2024", count: 1 }],
          items: [{ key: "posts/2024/spring-sale", title: "Spring sale", folder: "2024", status: "published" }],
        },
      }),
    );
    renderPage();

    const groups = await screen.findAllByRole("region");
    expect(groups.map((group) => group.getAttribute("aria-label"))).toEqual(["Our site", "Writing"]);
    expect(await screen.findByRole("button", { name: "Spring sale" })).toBeInTheDocument();
    expect(screen.queryByText("Pages")).not.toBeInTheDocument();
  });

  it("opens an item of a page bundle by its key", async () => {
    vi.stubGlobal(
      "fetch",
      server({
        [`${base}/content-menu`]: {
          groups: [{ key: "blog", title: "Writing", items: [{ key: "posts", kind: "collection", title: "Posts", itemName: "post", count: 1, hasChanges: false }] }],
          totals: { singles: 0, collections: 1 },
          fallback: false,
        },
        [`${base}/collections/posts/items`]: {
          key: "posts", title: "Posts", itemName: "post", total: 1, folders: [],
          items: [{ key: "posts/2024/spring-sale", title: "Spring sale", status: "published" }],
        },
      }),
    );
    renderPage();

    await userEvent.click(await screen.findByRole("button", { name: "Spring sale" }));
    expect(await screen.findByText("the editor")).toBeInTheDocument();
  });

  it("names the fallback groups in the reader's language", async () => {
    vi.stubGlobal("fetch", server(theSections));
    renderPage();

    expect(await screen.findByRole("region", { name: "Pages" })).toBeInTheDocument();
  });
});
