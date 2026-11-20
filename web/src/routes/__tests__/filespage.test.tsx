import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { Inside } from "./here";
import { setWidth } from "./width";
import { FilesPage } from "../FilesPage";

vi.mock("../files/CodeEditor", () => ({
  default: ({ text, label, onChange, onSave }: { text: string; label: string; onChange: (text: string) => void; onSave: () => void }) => (
    <textarea aria-label={label} defaultValue={text} onChange={(event) => onChange(event.target.value)} onKeyDown={(event) => event.key === "s" && event.ctrlKey && onSave()} />
  ),
}));

const base = "/api/organisations/technative/projects/korenaar/files";

type Answer = { status?: number; body: unknown };

function serve(answers: Record<string, Answer | (() => Answer)>) {
  const asked: { method: string; url: string; body?: string }[] = [];
  vi.stubGlobal(
    "fetch",
    vi.fn((url: string, init?: RequestInit) => {
      const method = init?.method ?? "GET";
      asked.push({ method, url, body: init?.body as string | undefined });
      const found = answers[`${method} ${url}`];
      const answer = typeof found === "function" ? found() : found ?? { status: 404, body: {} };
      const status = answer.status ?? 200;
      return Promise.resolve({ ok: status < 300, status, json: async () => answer.body });
    }),
  );
  return asked;
}

const tree = { entries: [{ path: "layouts", dir: true }, { path: "layouts/index.html" }, { path: "static", dir: true }, { path: "static/loaf.png" }] };

function at(address: string) {
  return render(
    <Inside>
      <MemoryRouter initialEntries={[address]}>
        <Routes>
          <Route path="/o/:organisation/p/:project/files" element={<FilesPage />} />
        </Routes>
      </MemoryRouter>
    </Inside>,
  );
}

beforeEach(() => {
  setWidth(1440);
  document.cookie = "chatcms_request_token=t; path=/";
});
afterEach(() => {
  vi.unstubAllGlobals();
  document.cookie = "chatcms_request_token=; max-age=0; path=/";
});

describe("the Files screen", () => {
  it("opens a file from the tree, and saves it with the sum it was read with", async () => {
    const asked = serve({
      [`GET ${base}`]: { body: tree },
      [`GET ${base}/layouts/index.html`]: { body: { path: "layouts/index.html", text: "<h1>{{ .Title }}</h1>", sum: "s1" } },
      [`PUT ${base}/layouts/index.html`]: { body: { path: "layouts/index.html", text: "<h2>{{ .Title }}</h2>", sum: "s2" } },
    });
    at("/o/technative/p/korenaar/files");

    expect(await screen.findByText("Choose a file to open it.")).toBeInTheDocument();
    await userEvent.click(screen.getByRole("treeitem", { name: "layouts" }));
    await userEvent.click(screen.getByRole("treeitem", { name: "index.html" }));

    const editor = await screen.findByRole("textbox", { name: "Text of layouts/index.html" });
    expect(screen.getByRole("button", { name: "Save" })).toBeDisabled();
    await userEvent.clear(editor);
    await userEvent.type(editor, "<h2>{{{{ .Title }}</h2>");
    expect(screen.getByRole("status")).toHaveTextContent("Not saved yet");
    await userEvent.click(screen.getByRole("button", { name: "Save" }));

    await waitFor(() => expect(screen.getByRole("status")).toHaveTextContent("Saved"));
    const put = asked.find((one) => one.method === "PUT");
    expect(JSON.parse(put?.body ?? "{}")).toEqual({ text: "<h2>{{ .Title }}</h2>", sum: "s1" });
  });

  it("says when the file changed meanwhile, and opens it again", async () => {
    let reads = 0;
    serve({
      [`GET ${base}`]: { body: tree },
      [`GET ${base}/layouts/index.html`]: () => ({ body: { path: "layouts/index.html", text: reads++ === 0 ? "old" : "newer", sum: `s${reads}` } }),
      [`PUT ${base}/layouts/index.html`]: { status: 409, body: { error: { code: "refused", key: "refused.file-changed", message: "This file changed since it was opened." } } },
    });
    at("/o/technative/p/korenaar/files?file=layouts%2Findex.html");

    const editor = await screen.findByRole("textbox", { name: "Text of layouts/index.html" });
    await userEvent.type(editor, "!");
    await userEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("This file changed since you opened it.");
    await userEvent.click(screen.getByRole("button", { name: "Open it again" }));
    expect(await screen.findByDisplayValue("newer")).toBeInTheDocument();
  });

  it("says a picture cannot be opened here", async () => {
    serve({
      [`GET ${base}`]: { body: tree },
      [`GET ${base}/static/loaf.png`]: { body: { path: "static/loaf.png", text: "", unusable: "binary" } },
    });
    at("/o/technative/p/korenaar/files?file=static%2Floaf.png");
    expect(await screen.findByText("This file cannot be opened here: it is not text.")).toBeInTheDocument();
  });

  it("fills a phone's screen with an open file, and goes back to the tree", async () => {
    setWidth(390);
    serve({
      [`GET ${base}`]: { body: tree },
      [`GET ${base}/layouts/index.html`]: { body: { path: "layouts/index.html", text: "x", sum: "s1" } },
    });
    at("/o/technative/p/korenaar/files?file=layouts%2Findex.html");
    expect(await screen.findByRole("textbox", { name: "Text of layouts/index.html" })).toBeInTheDocument();
    expect(screen.queryByRole("tree")).toBeNull();
    await userEvent.click(screen.getByRole("button", { name: "Back to the files" }));
    expect(await screen.findByRole("tree", { name: "Files of the site" })).toBeInTheDocument();
  });
});
