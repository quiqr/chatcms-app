import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router";

import { fill } from "./fill";
import { Inside } from "./here";import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from "vitest";

import { PageEditor } from "../PageEditor";


const aPage = {
  key: "home",
  title: "Home page",
  section: "Pages",
  hasBody: true,
  body: "The body of the page.\n",
  fields: [
    { path: "title", type: "string", title: "Page title", value: "Welcome" },
    {
      path: "intro",
      type: "string",
      title: "Introduction",
      value: "We sell apples.",
      config: { multiLine: true },
    },
    { path: "published", type: "boolean", title: "Published", value: true },
    {
      path: "colour",
      type: "select",
      title: "Colour",
      value: "red",
      options: [
        { value: "red", label: "Red" },
        { value: "blue", label: "Blue" },
      ],
    },
    { path: "when", type: "date", title: "Date", value: "2026-03-01T09:00:00Z" },
    { path: "tags", type: "leaf-array", title: "Tags", value: ["bread", "cake"] },
    {
      path: "links",
      type: "accordion",
      title: "Links",
      value: [{ label: "Shop", url: "/shop/" }],
      fields: [
        { path: "label", type: "string", title: "Label" },
        { path: "url", type: "string", title: "Address" },
      ],
    },
    {
      path: "seo",
      type: "section",
      title: "Search",
      fields: [{ path: "seo.description", type: "string", title: "Description", value: "A bakery." }],
    },
    { path: "mystery", type: "eisenhouwer", title: "Mystery", value: "something", unknown: true },
  ],
};

function server(sent?: (body: unknown) => void, answer: unknown = aPage) {
  return vi.fn().mockImplementation((_path: string, options?: RequestInit) => {
    if (options?.body) sent?.(JSON.parse(String(options.body)));
    return Promise.resolve({ ok: true, status: 200, json: async () => answer });
  });
}

function renderEditor() {
  return render(
    <Inside><MemoryRouter initialEntries={["/o/technative/p/korenaar/content/home"]}>
      <Routes>
        <Route path="/o/:organisation/p/:project/content/*" element={<PageEditor organisation="technative" />} />
      </Routes>
    </MemoryRouter></Inside>,
  );
}

beforeEach(() => {
  document.cookie = "chatcms_request_token=a-token; path=/";
  vi.useRealTimers();
});

afterEach(() => {
  vi.unstubAllGlobals();
});

beforeAll(async () => {
  await import("../../fields/markdown/MarkdownEditor");
}, 30_000);

describe("the page editor", () => {
  it("draws a control for each kind of field the model declares", async () => {
    vi.stubGlobal("fetch", server());
    renderEditor();

    expect(await screen.findByLabelText("Page title")).toHaveValue("Welcome");
    expect(screen.getByLabelText("Introduction")).toHaveValue("We sell apples.");
    expect(screen.getByLabelText("Published")).toBeChecked();
    expect(screen.getByLabelText("Colour")).toHaveValue("red");
    expect(screen.getByLabelText("Date")).toHaveValue("2026-03-01");
    expect(screen.getByLabelText("Tags 1")).toHaveValue("bread");
    expect(await screen.findByRole("textbox", { name: "Page text" })).toHaveTextContent("The body of the page.");
  });

  it("shows a group with the fields inside it", async () => {
    vi.stubGlobal("fetch", server());
    renderEditor();

    await screen.findByLabelText("Page title");
    expect(screen.getByRole("button", { name: "Search" })).toBeInTheDocument();
    expect(screen.getByLabelText("Description")).toHaveValue("A bakery.");
  });

  it("shows a field it cannot draw as plain text, and says so", async () => {
    vi.stubGlobal("fetch", server());
    renderEditor();

    expect(await screen.findByLabelText("Mystery")).toHaveValue("something");
    expect(screen.getByText(/shown as plain text for now/)).toBeInTheDocument();
  });

  it("saves what was typed, as one save rather than one per letter", async () => {
    const sent: unknown[] = [];
    vi.stubGlobal("fetch", server((body) => sent.push(body)));
    renderEditor();

    const title = await screen.findByLabelText("Page title");
    await userEvent.clear(title);
    await userEvent.type(title, "Welcome home");

    await waitFor(() => expect(sent.at(-1)).toEqual({ values: { title: "Welcome home" } }), { timeout: 8000 });
    expect(sent.length).toBeLessThanOrEqual(2);
  });

  it("says when it is saving and when it is done", async () => {
    vi.stubGlobal("fetch", server());
    renderEditor();

    const title = await screen.findByLabelText("Page title");
    expect(screen.getByText("Saved")).toBeInTheDocument();

    await userEvent.type(title, "!");
    await waitFor(() => expect(screen.getByText("Saved")).toBeInTheDocument(), {
      timeout: 4000,
    });
  });

  it("says so when a save does not work", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockImplementation((_path: string, options?: RequestInit) => {
        if (options?.method === "PUT") {
          return Promise.resolve({
            ok: false,
            status: 409,
            json: async () => ({ error: { code: "refused", message: "That could not be saved." } }),
          });
        }
        return Promise.resolve({ ok: true, status: 200, json: async () => aPage });
      }),
    );
    renderEditor();

    await userEvent.type(await screen.findByLabelText("Page title"), "!");

    expect(await screen.findByText(/keep trying/, {}, { timeout: 4000 })).toBeInTheDocument();
  });

  it("adds and removes an item in a list of values", async () => {
    const sent: unknown[] = [];
    vi.stubGlobal("fetch", server((body) => sent.push(body)));
    renderEditor();

    await screen.findByLabelText("Tags 1");
    await userEvent.click(screen.getByRole("button", { name: /Add tags/i }));
    expect(screen.getByLabelText("Tags 3")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "Remove Tags 1" }));
    await waitFor(() => expect(screen.getByLabelText("Tags 1")).toHaveValue("cake"));
  });

  it("opens an item in a list of items and edits a field inside it", async () => {
    const sent: unknown[] = [];
    vi.stubGlobal("fetch", server((body) => sent.push(body)));
    renderEditor();

    await screen.findByLabelText("Page title");
    expect(screen.getByLabelText("Label")).toHaveValue("Shop");

    await userEvent.clear(screen.getByLabelText("Label"));
    await fill(screen.getByLabelText("Label"), "Our shop");

    await waitFor(() => expect(sent.at(-1)).toEqual({ values: { links: [{ label: "Our shop", url: "/shop/" }] } }), { timeout: 4000 });
  });

  it("shows the text formatted, and as Markdown exactly as it is saved", async () => {
    vi.stubGlobal("fetch", server());
    renderEditor();

    const formatted = await screen.findByRole("textbox", { name: "Page text" });
    expect(formatted).toHaveAttribute("contenteditable", "true");
    await userEvent.click(screen.getByRole("radio", { name: "Markdown" }));
    expect(screen.getByRole("textbox", { name: "Page text" })).toHaveValue("The body of the page.\n");
    await userEvent.click(screen.getByRole("radio", { name: "Formatted" }));
    expect(await screen.findByRole("textbox", { name: "Page text" })).toHaveTextContent("The body of the page.");
  });

  it("can show the Markdown source instead", async () => {
    vi.stubGlobal("fetch", server());
    renderEditor();

    await screen.findByLabelText("Page text");
    expect(screen.getByRole("toolbar", { name: "Text formatting" })).toBeInTheDocument();

    await userEvent.click(screen.getByRole("radio", { name: "Markdown" }));
    expect(screen.queryByRole("toolbar", { name: "Text formatting" })).not.toBeInTheDocument();
  });

  it("shows a page's text but does not edit it when the model gives it no field", async () => {
    vi.stubGlobal("fetch", server(undefined, { ...aPage, hasBody: false }));
    renderEditor();

    await screen.findByLabelText("Page title");
    expect(screen.queryByLabelText("Page text")).not.toBeInTheDocument();
    expect(screen.getByText(/not set up to be edited here/)).toBeInTheDocument();
    expect(screen.getByText(/The body of the page/)).toBeInTheDocument();
  });

  it("has no Save button anywhere", async () => {
    vi.stubGlobal("fetch", server());
    renderEditor();

    await screen.findByLabelText("Page title");
    expect(screen.queryByRole("button", { name: /^Save/i })).not.toBeInTheDocument();
  });

  it("uses no git word anywhere an editor reads", async () => {
    vi.stubGlobal("fetch", server());
    renderEditor();

    await screen.findByLabelText("Page title");
    const words = (document.body.textContent ?? "").toLowerCase();
    for (const word of ["commit", "branch", "push", "merge", "repository", "git", "sha", "head"]) {
      expect(words).not.toContain(word);
    }
  });
});

describe("undo", () => {
  function living(sent: unknown[] = []) {
    const held = structuredClone(aPage);
    return vi.fn().mockImplementation((_path: string, options?: RequestInit) => {
      if (options?.method === "PUT" && options.body) {
        const edit = JSON.parse(String(options.body)) as { values: Record<string, unknown>; body?: string };
        sent.push(edit);
        for (const [path, value] of Object.entries(edit.values)) {
          const control = held.fields.find((one) => one.path === path);
          if (control) control.value = value as never;
        }
        if (edit.body !== undefined) held.body = edit.body;
      }
      return Promise.resolve({ ok: true, status: 200, json: async () => structuredClone(held) });
    });
  }

  const saved = (sent: unknown[], count: number) =>
    waitFor(() => expect(sent).toHaveLength(count), { timeout: 4000 });

  beforeEach(() => {
    window.sessionStorage.clear();
  });

  it("has nothing to undo or redo on opening", async () => {
    vi.stubGlobal("fetch", living());
    renderEditor();
    await screen.findByLabelText("Page title");
    expect(screen.getByRole("button", { name: "Undo" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Redo" })).toBeDisabled();
  });

  it("makes one step of two quick edits, named after both fields", async () => {
    const sent: unknown[] = [];
    vi.stubGlobal("fetch", living(sent));
    renderEditor();
    const title = (await screen.findByLabelText("Page title")) as HTMLInputElement;
    const intro = screen.getByLabelText("Introduction") as HTMLInputElement;
    act(() => {
      fireEvent.change(title, { target: { value: title.value + "!" } });
      fireEvent.change(intro, { target: { value: intro.value + "!" } });
    });
    await saved(sent, 1);
    expect(await screen.findByRole("button", { name: "Undo Page title, Introduction" })).toBeEnabled();
  });

  it("puts a cleared field back as a save, and does it again on redo", async () => {
    const sent: unknown[] = [];
    vi.stubGlobal("fetch", living(sent));
    renderEditor();
    const intro = await screen.findByLabelText("Introduction");
    await userEvent.clear(intro);
    await saved(sent, 1);

    await userEvent.click(await screen.findByRole("button", { name: "Undo Introduction" }));
    await saved(sent, 2);
    expect(sent[1]).toEqual({ values: { intro: "We sell apples." } });
    expect(screen.getByLabelText("Introduction")).toHaveValue("We sell apples.");

    await userEvent.click(await screen.findByRole("button", { name: "Redo Introduction" }));
    await saved(sent, 3);
    expect(sent[2]).toEqual({ values: { intro: "" } });
    expect(screen.getByLabelText("Introduction")).toHaveValue("");
  });

  it("saves what is waiting before undoing it", async () => {
    const sent: unknown[] = [];
    vi.stubGlobal("fetch", living(sent));
    renderEditor();
    await fill(await screen.findByLabelText("Page title"), " home");
    await userEvent.click(screen.getByRole("button", { name: "Undo Page title" }));
    await saved(sent, 2);
    expect(sent).toEqual([{ values: { title: "Welcome home" } }, { values: { title: "Welcome" } }]);
    expect(screen.getByLabelText("Page title")).toHaveValue("Welcome");
  });

  it("makes no step of a refused save", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockImplementation((_path: string, options?: RequestInit) =>
        options?.method === "PUT"
          ? Promise.resolve({
              ok: false,
              status: 403,
              json: async () => ({ error: { code: "refused", message: "No." } }),
            })
          : Promise.resolve({ ok: true, status: 200, json: async () => aPage }),
      ),
    );
    renderEditor();
    await fill(await screen.findByLabelText("Page title"), "!");
    await screen.findByText(/keep trying/, {}, { timeout: 4000 });
    expect(screen.getByRole("button", { name: "Undo" })).toBeDisabled();
  });

  it("goes back to the opened page in one save from the undo history", async () => {
    const sent: unknown[] = [];
    vi.stubGlobal("fetch", living(sent));
    renderEditor();
    await fill(await screen.findByLabelText("Page title"), "1");
    await saved(sent, 1);
    await fill(screen.getByLabelText("Introduction"), "2");
    await saved(sent, 2);
    await fill(screen.getByLabelText("Page text"), "3");
    await saved(sent, 3);

    await userEvent.click(screen.getByRole("button", { name: "Undo history" }));
    const rows = await screen.findAllByRole("menuitem");
    expect(rows.map((row) => row.textContent?.replace(/just now|\d+ min ago/, ""))).toEqual([
      "Page text",
      "Introduction",
      "Page title",
      "Opened",
    ]);
    expect(rows[0]).toHaveAttribute("aria-current", "step");
    await userEvent.click(screen.getByRole("menuitem", { name: /Opened/ }));
    await saved(sent, 4);
    expect(sent[3]).toEqual({
      values: { title: "Welcome", intro: "We sell apples." },
      body: "The body of the page.\n",
    });

    await userEvent.click(screen.getByRole("button", { name: "Undo history" }));
    expect(await screen.findByRole("menuitem", { name: /Opened/ })).toHaveAttribute("aria-current", "step");
    expect(screen.getAllByText("can be done again")).toHaveLength(3);
  });

  it("drops what could be redone when something new is typed", async () => {
    const sent: unknown[] = [];
    vi.stubGlobal("fetch", living(sent));
    renderEditor();
    await fill(await screen.findByLabelText("Page title"), "!");
    await saved(sent, 1);
    await userEvent.click(screen.getByRole("button", { name: "Undo Page title" }));
    await saved(sent, 2);
    expect(screen.getByRole("button", { name: "Redo Page title" })).toBeEnabled();

    await fill(screen.getByLabelText("Introduction"), "!");
    await saved(sent, 3);
    await waitFor(() => expect(screen.getByRole("button", { name: "Redo" })).toBeDisabled());
  });

  it("undoes and redoes with the keys outside a text field only", async () => {
    const sent: unknown[] = [];
    vi.stubGlobal("fetch", living(sent));
    renderEditor();
    const body = await screen.findByRole("textbox", { name: "Page text" });
    await fill(screen.getByLabelText("Page title"), "!");
    await saved(sent, 1);

    fireEvent.keyDown(body, { key: "z", ctrlKey: true });
    await new Promise((resolve) => setTimeout(resolve, 100));
    expect(sent).toHaveLength(1);

    fireEvent.keyDown(document.body, { key: "z", metaKey: true });
    await saved(sent, 2);
    expect(sent[1]).toEqual({ values: { title: "Welcome" } });

    fireEvent.keyDown(document.body, { key: "Z", ctrlKey: true, shiftKey: true });
    await saved(sent, 3);
    expect(sent[2]).toEqual({ values: { title: "Welcome!" } });

    fireEvent.keyDown(document.body, { key: "z", ctrlKey: true });
    await saved(sent, 4);
    fireEvent.keyDown(document.body, { key: "y", ctrlKey: true });
    await saved(sent, 5);
    expect(sent[4]).toEqual({ values: { title: "Welcome!" } });
  });

  it("keeps the steps through a reload of the tab", async () => {
    const sent: unknown[] = [];
    const fetcher = living(sent);
    vi.stubGlobal("fetch", fetcher);
    const first = renderEditor();
    await fill(await screen.findByLabelText("Page title"), "!");
    await saved(sent, 1);
    await screen.findByRole("button", { name: "Undo Page title" });
    first.unmount();

    renderEditor();
    expect(await screen.findByRole("button", { name: "Undo Page title" })).toBeEnabled();
  });

  it("starts afresh when the page was changed elsewhere", async () => {
    const sent: unknown[] = [];
    vi.stubGlobal("fetch", living(sent));
    const first = renderEditor();
    await fill(await screen.findByLabelText("Page title"), "!");
    await saved(sent, 1);
    await screen.findByRole("button", { name: "Undo Page title" });
    first.unmount();

    vi.stubGlobal("fetch", server());
    renderEditor();
    await screen.findByLabelText("Page title");
    expect(screen.getByRole("button", { name: "Undo" })).toBeDisabled();
  });
});
