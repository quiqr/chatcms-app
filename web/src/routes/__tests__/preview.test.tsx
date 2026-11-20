import { act, render as draw, screen, waitFor } from "@testing-library/react";
import { renderHook } from "@testing-library/react";
import type { ReactElement } from "react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { useStartWorkspace } from "../../lib/workspace";
import { Preview } from "../Preview";
import { setWidth } from "./width";

function render(screenUnderTest: ReactElement) {
  return draw(<MemoryRouter>{screenUnderTest}</MemoryRouter>);
}

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

  it("loads the page again after an assistant's change, once the site has had time to rebuild", async () => {
    vi.stubGlobal("fetch", server([{ body: { state: "warm" } }]));
    const view = render(<Preview organisation="technative" project="korenaar" settle={0} />);
    const first = await screen.findByTitle("Your site");

    vi.useFakeTimers();
    try {
      view.rerender(<MemoryRouter><Preview organisation="technative" project="korenaar" settle={1} /></MemoryRouter>);
      expect(screen.getByTitle("Your site")).toBe(first);
      await act(() => vi.advanceTimersByTimeAsync(2_500));
      const second = screen.getByTitle("Your site");
      expect(second).not.toBe(first);
      await act(() => vi.advanceTimersByTimeAsync(5_500));
      expect(screen.getByTitle("Your site")).not.toBe(second);
    } finally {
      vi.useRealTimers();
    }
  });

  it("explains an item with no page of its own, and shows the home page when asked", async () => {
    vi.stubGlobal("fetch", server([{ body: { state: "warm" } }]));
    render(<Preview organisation="technative" project="korenaar" page="dataServices" shows={{ kind: "none" }} />);

    expect(await screen.findByText("No page of its own")).toBeInTheDocument();
    expect(screen.queryByTitle("Your site")).toBeNull();
    await userEvent.click(screen.getByRole("button", { name: "Show the home page" }));
    expect(await screen.findByTitle("Your site")).toHaveAttribute("src", "/preview/technative/korenaar/");
  });

  it("shows the address the model declares, and nothing where it hides the preview", async () => {
    vi.stubGlobal("fetch", server([{ body: { state: "warm" } }]));
    const view = render(<Preview organisation="technative" project="korenaar" page="dataSidebar" shows={{ kind: "declared", address: "/about/" }} />);
    expect(await screen.findByTitle("Your site")).toHaveAttribute("src", "/preview/technative/korenaar/about/");
    view.unmount();

    render(<Preview organisation="technative" project="korenaar" page="secret" shows={{ kind: "hidden" }} />);
    expect(await screen.findByText("This site shows no preview for this.")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Show the home page" })).toBeNull();
  });

  it("gives an implementer a link to this start in the log", async () => {
    vi.stubGlobal("fetch", server([{ body: { state: "starting", steps, entry: "start:7" } }]));

    render(<Preview organisation="technative" project="korenaar" />);

    expect(await screen.findByRole("link", { name: "Show the log" })).toHaveAttribute(
      "href",
      "/o/technative/p/korenaar/log?entry=start%3A7",
    );
  });

  it("sends an implementer to what went wrong when the start failed", async () => {
    vi.stubGlobal("fetch", server([{ body: { state: "sleeping", reason: "refused.preview-did-not-start-site", entry: "start:8" } }]));

    render(<Preview organisation="technative" project="korenaar" />);

    expect(await screen.findByRole("link", { name: "See what went wrong" })).toHaveAttribute(
      "href",
      "/o/technative/p/korenaar/log?entry=start%3A8",
    );
  });

  it("gives a content creator nothing to unfold", async () => {
    vi.stubGlobal("fetch", server([{ body: { state: "starting" } }]));

    render(<Preview organisation="technative" project="korenaar" />);

    await screen.findByRole("status");
    expect(screen.queryByRole("link", { name: /log|went wrong/ })).toBeNull();
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

    expect(await screen.findByTitle("Your site")).toHaveAttribute("src", "https://preview.example/v/k1/__chatcms/pdf/?file=report.pdf&lang=en");
  });

  it("asks a PDF to fit the width of a phone", async () => {
    setWidth(390);
    vi.stubGlobal("fetch", server([{ body: { state: "warm", address: "https://preview.example/v/k1/", type: "pdf", open: "/report.pdf" } }]));

    render(<Preview organisation="technative" project="korenaar" page="posts/one" />);

    expect(await screen.findByTitle("Your site")).toHaveAttribute("src", "https://preview.example/v/k1/__chatcms/pdf/?file=report.pdf&fit=width&lang=en");
    setWidth(1440);
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

  async function framed(onAsk = vi.fn(), highlight: { before: string; after: string; id: number } | null = null, chatOn = true) {
    vi.stubGlobal("fetch", server([{ body: { state: "warm", address, type: "web", open: "/" } }]));
    const view = render(<Preview organisation="technative" project="korenaar" page="posts/one" onAsk={onAsk} chatOn={chatOn} highlight={highlight} />);
    const frame = (await screen.findByTitle("Your site")) as HTMLIFrameElement;
    const sent = vi.fn();
    Object.defineProperty(frame, "contentWindow", { value: { postMessage: sent }, configurable: true });
    return { frame, sent, onAsk, view };
  }

  function fromFrame(frame: HTMLIFrameElement, data: unknown, origin = "https://preview.example") {
    window.dispatchEvent(new MessageEvent("message", { data, origin, source: frame.contentWindow as MessageEventSource }));
  }

  const asked = { text: "Fresh bread", before: "Korenaar ", after: ".", path: "/v/k1/posts/one/", prompt: "Make it shorter" };

  it("answers the frame's hello with the overlay's words and whether chat is on", async () => {
    const { frame, sent } = await framed();
    fromFrame(frame, { chatcms: "ready" });
    const setup = sent.mock.calls.map(([data]) => data).filter((data) => data.chatcms === "setup").at(-1);
    expect(setup).toMatchObject({ chatcms: "setup", on: true, words: { chip: "Chat", chipLabel: "Chat about the selected text", placeholder: "What should change?" } });
    expect(setup.words.quick).toEqual(["Make it shorter", "Simpler words", "Fix spelling"]);
  });

  it("loads an empty page again, a few times at most", async () => {
    const { frame } = await framed();
    vi.useFakeTimers();
    try {
      fromFrame(frame, { chatcms: "ready", blank: true });
      await act(() => vi.advanceTimersByTimeAsync(1_500));
      const second = screen.getByTitle("Your site") as HTMLIFrameElement;
      expect(second).not.toBe(frame);
      Object.defineProperty(second, "contentWindow", { value: { postMessage: vi.fn() }, configurable: true });
      for (let i = 0; i < 5; i++) {
        const current = screen.getByTitle("Your site") as HTMLIFrameElement;
        Object.defineProperty(current, "contentWindow", { value: { postMessage: vi.fn() }, configurable: true });
        fromFrame(current, { chatcms: "ready", blank: true });
        await act(() => vi.advanceTimersByTimeAsync(10_000));
      }
      const last = screen.getByTitle("Your site");
      await act(() => vi.advanceTimersByTimeAsync(30_000));
      expect(screen.getByTitle("Your site")).toBe(last);
    } finally {
      vi.useRealTimers();
    }
  });

  it("leaves a page with something on it alone", async () => {
    const { frame } = await framed();
    vi.useFakeTimers();
    try {
      fromFrame(frame, { chatcms: "ready", blank: false });
      await act(() => vi.advanceTimersByTimeAsync(10_000));
      expect(screen.getByTitle("Your site")).toBe(frame);
    } finally {
      vi.useRealTimers();
    }
  });

  it("tells the overlay again when chat is switched off", async () => {
    const { sent, view } = await framed();
    view.rerender(<MemoryRouter><Preview organisation="technative" project="korenaar" page="posts/one" onAsk={vi.fn()} chatOn={false} /></MemoryRouter>);
    const setup = sent.mock.calls.map(([data]) => data).filter((data) => data.chatcms === "setup").at(-1);
    expect(setup).toMatchObject({ on: false });
  });

  it("passes a request and a draft on, with the address inside the preview", async () => {
    const { frame, onAsk } = await framed();
    fromFrame(frame, { chatcms: "send", ...asked });
    fromFrame(frame, { chatcms: "draft", ...asked });
    const request = { text: "Fresh bread", before: "Korenaar ", after: ".", path: "/posts/one/", prompt: "Make it shorter" };
    expect(onAsk.mock.calls).toEqual([
      [request, "send"],
      [request, "draft"],
    ]);
  });

  it("ignores a message from another origin, another window, or while chat is off", async () => {
    const { frame, onAsk } = await framed();
    fromFrame(frame, { chatcms: "send", ...asked }, "https://elsewhere.example");
    window.dispatchEvent(new MessageEvent("message", { data: { chatcms: "send", ...asked }, origin: "https://preview.example", source: window }));
    fromFrame(frame, { chatcms: "send", ...asked, text: 3 });
    expect(onAsk).not.toHaveBeenCalled();

    const off = await framed(vi.fn(), null, false);
    fromFrame(off.frame, { chatcms: "send", ...asked });
    expect(off.onAsk).not.toHaveBeenCalled();
  });

  it("gives the frame the pills on every hello, and passes their buttons on", async () => {
    vi.stubGlobal("fetch", server([{ body: { state: "warm", address, type: "web", open: "/" } }]));
    const pill = { id: 4, state: "working" as const, before: "Korenaar ", after: ".", text: "Fresh bread" };
    const pressed = vi.fn();
    render(<Preview organisation="technative" project="korenaar" page="posts/one" onAsk={vi.fn()} chatOn pills={[pill]} onPill={pressed} />);
    const frame = (await screen.findByTitle("Your site")) as HTMLIFrameElement;
    const sent = vi.fn();
    Object.defineProperty(frame, "contentWindow", { value: { postMessage: sent }, configurable: true });
    fromFrame(frame, { chatcms: "ready" });
    expect(sent).toHaveBeenCalledWith({ chatcms: "pills", pills: [pill] }, "https://preview.example");

    fromFrame(frame, { chatcms: "undo", id: 4 });
    fromFrame(frame, { chatcms: "open-chat", id: 4 });
    fromFrame(frame, { chatcms: "seen", id: "4" });
    expect(pressed.mock.calls).toEqual([
      ["undo", 4],
      ["open-chat", 4],
    ]);
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

describe("a preview with pages of its own", () => {
  it("opens its own start page rather than the page being edited", async () => {
    vi.stubGlobal("fetch", server([{ body: { state: "warm", type: "web", open: "/", ownPages: true } }]));
    render(<Preview organisation="technative" project="korenaar" page="preface" />);
    expect(await screen.findByTitle("Your site")).toHaveAttribute("src", "/preview/technative/korenaar/");
  });
});

describe("a preview on the app's own address", () => {
  it("is framed by its path, so the app's session goes with it", async () => {
    vi.stubGlobal("fetch", server([{ body: { state: "warm", address: "/preview/technative/korenaar/p/site/", type: "web", open: "/", ownPages: true } }]));
    render(<Preview organisation="technative" project="korenaar" />);
    expect(await screen.findByTitle("Your site")).toHaveAttribute("src", "/preview/technative/korenaar/p/site/");
  });
});

describe("a PDF preview", () => {
  it("has no width to choose, since a PDF's size is its own", async () => {
    vi.stubGlobal("fetch", server([{ body: { state: "warm", address: "/preview/technative/korenaar/", type: "pdf", open: "/book.pdf" } }]));
    render(<Preview organisation="technative" project="korenaar" />);
    const frame = await screen.findByTitle("Your site");
    expect(screen.queryByRole("radiogroup", { name: "How wide to show the site" })).toBeNull();
    expect(frame.className).toContain("w-full");
  });
});
