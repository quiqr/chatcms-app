import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { findBetween, install } from "../select";

const app = "https://app.example";

function page(html: string) {
  document.head.innerHTML = "";
  document.body.innerHTML = html;
}

function select(node: Node, from: number, to: number) {
  const range = document.createRange();
  range.setStart(node, from);
  range.setEnd(node, to);
  const selection = window.getSelection()!;
  selection.removeAllRanges();
  selection.addRange(range);
}

function fromApp(data: unknown, origin = app, source: MessageEventSource | null = window) {
  window.dispatchEvent(new MessageEvent("message", { data, origin, source }));
}

function theButton() {
  const host = document.querySelector<HTMLElement>("[data-chatcms]");
  return host?.shadowRoot?.querySelector("button") ?? null;
}

let posted: ReturnType<typeof vi.spyOn>;
let installed: { stop: () => void }[] = [];

function start() {
  const script = install(window, app);
  installed.push(script);
  return script;
}

beforeEach(() => {
  Range.prototype.getBoundingClientRect = () =>
    ({ top: 10, bottom: 30, left: 10, right: 100, width: 90, height: 20, x: 10, y: 10, toJSON() {} }) as DOMRect;
  posted = vi.spyOn(window, "postMessage").mockImplementation(() => {});
});

afterEach(() => {
  installed.forEach((script) => script.stop());
  installed = [];
  vi.restoreAllMocks();
  window.getSelection()?.removeAllRanges();
});

describe("the selection script", () => {
  it("says it is ready to the app", () => {
    page("<p>Fresh bread</p>");
    start();
    expect(posted).toHaveBeenCalledWith({ chatcms: "ready" }, app);
  });

  it("sends a selection with its context and address to the app only", () => {
    page("<h1>Korenaar</h1><p>Fresh bread every morning, baked on the Oude Gracht.</p>");
    const script = start();
    fromApp({ chatcms: "setup", label: "Ask the assistant about this" });

    select(document.querySelector("p")!.firstChild!, 6, 25);
    script.look();
    const button = theButton()!;
    expect(button.style.display).toBe("block");
    expect(button.getAttribute("aria-label")).toBe("Ask the assistant about this");

    button.click();
    expect(posted).toHaveBeenLastCalledWith(
      {
        chatcms: "selection",
        text: "bread every morning",
        before: "KorenaarFresh ",
        after: ", baked on the Oude Gracht.",
        path: window.location.pathname,
      },
      app,
    );
    expect(button.style.display).toBe("none");
  });

  it("shows nothing before the app has said hello, or for nothing selected", () => {
    page("<p>Fresh bread</p>");
    const script = start();
    select(document.querySelector("p")!.firstChild!, 0, 5);
    script.look();
    expect(theButton()!.style.display).toBe("none");

    fromApp({ chatcms: "setup", label: "Ask" });
    window.getSelection()!.removeAllRanges();
    script.look();
    expect(theButton()!.style.display).toBe("none");
  });

  it("listens to the app only", () => {
    page("<p>Fresh bread</p>");
    const script = start();
    fromApp({ chatcms: "setup", label: "Ask" }, "https://elsewhere.example");
    fromApp({ chatcms: "setup", label: "Ask" }, app, null);
    select(document.querySelector("p")!.firstChild!, 0, 5);
    script.look();
    expect(theButton()!.style.display).toBe("none");
  });

  it("highlights the changed text when the app asks", () => {
    page("<p>Come in and smell the new rye. Open daily.</p>");
    const set = vi.fn();
    const remove = vi.fn();
    Object.assign(window, { Highlight: class { constructor(readonly range: Range) {} } });
    Object.defineProperty(window, "CSS", { value: { highlights: { set, delete: remove } }, configurable: true });
    vi.useFakeTimers();
    start();

    fromApp({ chatcms: "highlight", before: "Come in and smell ", after: " Open daily." });
    expect(set).toHaveBeenCalledTimes(1);
    expect((set.mock.calls[0]![1] as { range: Range }).range.toString()).toBe("the new rye.");
    vi.advanceTimersByTime(4000);
    expect(remove).toHaveBeenCalledWith("chatcms-changed");
    vi.useRealTimers();
  });
});

describe("finding the changed text", () => {
  it("finds what now stands between the old context, across elements", () => {
    page("<p>Come in and <em>smell</em> the new rye loaf today. Open daily.</p>");
    expect(findBetween(document, "Come in and smell ", " Open daily.")?.toString()).toBe("the new rye loaf today.");
  });

  it("finds nothing when the context is gone", () => {
    page("<p>Something else entirely</p>");
    expect(findBetween(document, "Come in and ", " Open daily.")).toBeNull();
  });

  it("takes the whole page when there is no context", () => {
    page("<p>Only this</p>");
    expect(findBetween(document, "", "")?.toString()).toBe("Only this");
  });
});
