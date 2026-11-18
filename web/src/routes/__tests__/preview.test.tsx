import { render, screen, waitFor } from "@testing-library/react";
import { renderHook } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { useStartWorkspace } from "../../lib/workspace";
import { Preview } from "../Preview";

const workspacePath = "/api/organisations/technative/projects/korenaar/workspace";

type Answer = { status?: number; body: unknown };

function server(answers: Answer[], asked: string[] = []) {
  let next = 0;
  return vi.fn().mockImplementation((path: string, options?: RequestInit) => {
    asked.push(`${options?.method ?? "GET"} ${path}`);
    const answer = answers[Math.min(next, answers.length - 1)]!;
    next += 1;
    const status = answer.status ?? 200;
    return Promise.resolve({ ok: status < 400, status, json: async () => answer.body });
  });
}

const steps = [
  { step: "starting-workspace", state: "done" },
  { step: "fetching-site", state: "done" },
  { step: "building-preview", state: "running" },
];

beforeEach(() => {
  document.cookie = "chatcms_request_token=a-token; path=/";
});

afterEach(() => {
  vi.unstubAllGlobals();
  vi.useRealTimers();
});

describe("preview", () => {
  it("says the preview is being prepared, then shows it once warm", async () => {
    const asked: string[] = [];
    vi.stubGlobal("fetch", server([{ body: { state: "starting" } }, { body: { state: "warm" } }], asked));

    render(<Preview organisation="technative" project="korenaar" />);

    expect(await screen.findByRole("status")).toHaveTextContent("Preparing your preview");
    expect(screen.queryByTitle("Your site")).toBeNull();

    const frame = await screen.findByTitle("Your site", {}, { timeout: 3000 });
    expect(frame).toHaveAttribute("src", "/preview/technative/korenaar/");
    expect(asked[0]).toBe(`POST ${workspacePath}`);
    expect(asked[1]).toBe(`GET ${workspacePath}`);
  });

  it("lets somebody given the steps unfold them, with the log", async () => {
    vi.stubGlobal("fetch", server([{ body: { state: "starting", steps, log: ["hugo starting"] } }]));

    render(<Preview organisation="technative" project="korenaar" />);

    const unfold = await screen.findByText("Show what is happening");
    await userEvent.click(unfold);
    expect(screen.getByText("Building the preview")).toBeInTheDocument();
    expect(screen.getByText("Busy")).toBeInTheDocument();
    expect(screen.getByText("hugo starting")).toBeInTheDocument();
  });

  it("gives a content creator nothing to unfold", async () => {
    vi.stubGlobal("fetch", server([{ body: { state: "starting" } }]));

    render(<Preview organisation="technative" project="korenaar" />);

    await screen.findByRole("status");
    expect(screen.queryByTestId("preview-steps")).toBeNull();
    expect(screen.getByRole("status").textContent).not.toMatch(/git|commit|branch|repository|build log/i);
  });

  it("says why when the preview did not start", async () => {
    vi.stubGlobal("fetch", server([{ body: { state: "sleeping", reason: "refused.preview-did-not-start-site" } }]));

    render(<Preview organisation="technative" project="korenaar" />);

    expect(await screen.findByRole("status")).toHaveTextContent("The preview did not start");
    expect(screen.queryByTitle("Your site")).toBeNull();
  });

  it("says the refusal when the server will not start one", async () => {
    vi.stubGlobal(
      "fetch",
      server([
        {
          status: 429,
          body: { error: { code: "refused", message: "too many", key: "refused.too-many-sites-being-previewed" } },
        },
      ]),
    );

    render(<Preview organisation="technative" project="korenaar" />);

    await waitFor(() => expect(screen.getByRole("status").textContent).not.toBe("Preparing your preview…"));
  });
});

describe("starting a workspace", () => {
  it("asks for the workspace when a project is opened, and not without one", async () => {
    const asked: string[] = [];
    vi.stubGlobal("fetch", server([{ body: { state: "starting" } }], asked));

    renderHook(() => useStartWorkspace("technative", "korenaar"));
    renderHook(() => useStartWorkspace("technative", ""));

    await waitFor(() => expect(asked).toEqual([`POST ${workspacePath}`]));
  });

  it("does not fail when the workspace cannot start", async () => {
    vi.stubGlobal("fetch", server([{ status: 404, body: {} }]));

    renderHook(() => useStartWorkspace("technative", "korenaar"));

    await new Promise((resolve) => setTimeout(resolve, 10));
  });
});

describe("a preview from its own origin", () => {
  it("shows the page at the address the workspace gives", async () => {
    vi.stubGlobal("fetch", server([{ body: { state: "warm", address: "https://preview.example/v/k1/", type: "web", open: "/" } }]));

    render(<Preview organisation="technative" project="korenaar" page="posts/one" />);

    const frame = await screen.findByTitle("Your site");
    expect(frame).toHaveAttribute("src", "https://preview.example/v/k1/posts/one/");
  });

  it("keeps the app's own path when the address is the app's", async () => {
    const own = `${window.location.origin}/preview/technative/korenaar/`;
    vi.stubGlobal("fetch", server([{ body: { state: "warm", address: own, type: "web" } }]));

    render(<Preview organisation="technative" project="korenaar" />);

    expect(await screen.findByTitle("Your site")).toHaveAttribute("src", "/preview/technative/korenaar/");
  });

  it("shows a PDF at the page the runtime opens", async () => {
    vi.stubGlobal("fetch", server([{ body: { state: "warm", address: "https://preview.example/v/k1/", type: "pdf", open: "/report.pdf" } }]));

    render(<Preview organisation="technative" project="korenaar" page="posts/one" />);

    expect(await screen.findByTitle("Your site")).toHaveAttribute("src", "https://preview.example/v/k1/__chatcms/pdf/?file=report.pdf");
  });

  it("offers to open what it cannot show", async () => {
    vi.stubGlobal("fetch", server([{ body: { state: "warm", address: "https://preview.example/v/k1/", type: "other", open: "/book.epub" } }]));

    render(<Preview organisation="technative" project="korenaar" />);

    const link = await screen.findByRole("link", { name: "Open the result" });
    expect(link).toHaveAttribute("href", "https://preview.example/v/k1/book.epub");
    expect(screen.queryByTitle("Your site")).toBeNull();
  });
});

describe("selecting in the preview", () => {
  const address = "https://preview.example/v/k1/";

  async function framed(onSelect = vi.fn(), highlight: { before: string; after: string; id: number } | null = null) {
    vi.stubGlobal("fetch", server([{ body: { state: "warm", address, type: "web", open: "/" } }]));
    const view = render(<Preview organisation="technative" project="korenaar" page="posts/one" onSelect={onSelect} highlight={highlight} />);
    const frame = (await screen.findByTitle("Your site")) as HTMLIFrameElement;
    const sent = vi.fn();
    Object.defineProperty(frame, "contentWindow", { value: { postMessage: sent }, configurable: true });
    return { frame, sent, onSelect, view };
  }

  function fromFrame(frame: HTMLIFrameElement, data: unknown, origin = "https://preview.example") {
    window.dispatchEvent(new MessageEvent("message", { data, origin, source: frame.contentWindow as MessageEventSource }));
  }

  it("answers the frame's hello with the button's label", async () => {
    const { frame, sent } = await framed();
    fromFrame(frame, { chatcms: "ready" });
    expect(sent).toHaveBeenCalledWith({ chatcms: "setup", label: "Ask the assistant about this" }, "https://preview.example");
  });

  it("passes a selection on, with the address inside the preview", async () => {
    const { frame, onSelect } = await framed();
    fromFrame(frame, { chatcms: "selection", text: "Fresh bread", before: "Korenaar ", after: ".", path: "/v/k1/posts/one/" });
    expect(onSelect).toHaveBeenCalledWith({ text: "Fresh bread", before: "Korenaar ", after: ".", path: "/posts/one/" });
  });

  it("ignores a message from another origin or another window", async () => {
    const { frame, onSelect } = await framed();
    fromFrame(frame, { chatcms: "selection", text: "x", before: "", after: "", path: "/" }, "https://elsewhere.example");
    window.dispatchEvent(
      new MessageEvent("message", { data: { chatcms: "selection", text: "x", before: "", after: "", path: "/" }, origin: "https://preview.example", source: window }),
    );
    fromFrame(frame, { chatcms: "selection", text: 3 });
    expect(onSelect).not.toHaveBeenCalled();
  });

  it("asks each reload shortly after the change to highlight it", async () => {
    const { frame, sent } = await framed(vi.fn(), { before: "Korenaar ", after: ".", id: Date.now() });
    fromFrame(frame, { chatcms: "ready" });
    fromFrame(frame, { chatcms: "ready" });
    const highlights = sent.mock.calls.filter(([data]) => (data as { chatcms: string }).chatcms === "highlight");
    expect(highlights).toEqual([
      [{ chatcms: "highlight", before: "Korenaar ", after: "." }, "https://preview.example"],
      [{ chatcms: "highlight", before: "Korenaar ", after: "." }, "https://preview.example"],
    ]);
  });

  it("does not highlight an old change", async () => {
    const { frame, sent } = await framed(vi.fn(), { before: "Korenaar ", after: ".", id: Date.now() - 60_000 });
    fromFrame(frame, { chatcms: "ready" });
    expect(sent.mock.calls.some(([data]) => (data as { chatcms: string }).chatcms === "highlight")).toBe(false);
  });
});
