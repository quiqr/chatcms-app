import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { findBetween, install, placeCard } from "../select";

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

const words = {
  chip: "Chat",
  chipLabel: "Chat about the selected text",
  card: "Chat about the selected text",
  placeholder: "What should change?",
  send: "Send to the assistant",
  quick: ["Make it shorter", "Simpler words", "Fix spelling"],
  goesTo: "Chat · goes to your chat",
  keys: "Enter to send · Esc to close",
};

let posted: ReturnType<typeof vi.spyOn>;
let installed: { stop: () => void }[] = [];

function start() {
  const script = install(window, app);
  installed.push(script);
  return script;
}

function ready(on = true) {
  const script = start();
  fromApp({ chatcms: "setup", words, on });
  return script;
}

function pointer(type: string, target: EventTarget = document.body) {
  target.dispatchEvent(new MouseEvent(type, { bubbles: true, composed: true }));
}

function selectionChanged() {
  document.dispatchEvent(new Event("selectionchange"));
}

const shown = (element: HTMLElement) => element.style.display !== "none" && element.style.display !== "";

beforeEach(() => {
  Range.prototype.getBoundingClientRect = () =>
    ({ top: 10, bottom: 30, left: 10, right: 100, width: 90, height: 20, x: 10, y: 10, toJSON() {} }) as DOMRect;
  posted = vi.spyOn(window, "postMessage").mockImplementation(() => {});
  vi.useFakeTimers();
});

afterEach(() => {
  installed.forEach((script) => script.stop());
  installed = [];
  vi.useRealTimers();
  vi.restoreAllMocks();
  window.getSelection()?.removeAllRanges();
});

describe("the selection overlay", () => {
  const text = "<h1>Korenaar</h1><p>Fresh bread every morning, baked on the Oude Gracht.</p>";

  it("says it is ready to the app", () => {
    page("<p>Fresh bread</p>");
    start();
    expect(posted).toHaveBeenCalledWith({ chatcms: "ready" }, app);
  });

  it("shows the chip only after the selection ends, never while dragging", () => {
    page(text);
    const script = ready();
    pointer("pointerdown");
    select(document.querySelector("p")!.firstChild!, 6, 25);
    selectionChanged();
    vi.advanceTimersByTime(1000);
    expect(shown(script.chip)).toBe(false);

    pointer("pointerup");
    vi.advanceTimersByTime(149);
    expect(shown(script.chip)).toBe(false);
    vi.advanceTimersByTime(1);
    expect(shown(script.chip)).toBe(true);
    expect(script.chip.getAttribute("aria-label")).toBe("Chat about the selected text");
    expect(script.chip.textContent).toBe("QChat");
  });

  it("waits longer for a selection made with the keyboard", () => {
    page(text);
    const script = ready();
    document.dispatchEvent(new KeyboardEvent("keydown", { key: "ArrowRight", shiftKey: true }));
    select(document.querySelector("p")!.firstChild!, 6, 25);
    selectionChanged();
    vi.advanceTimersByTime(399);
    expect(shown(script.chip)).toBe(false);
    vi.advanceTimersByTime(1);
    expect(shown(script.chip)).toBe(true);
  });

  it("shows nothing for a single character, while chat is off, or across two fields", () => {
    page('<p data-quiqr-field="title">A</p><p data-quiqr-field="body">Fresh bread</p>');
    const script = ready();
    select(document.querySelectorAll("p")[1]!.firstChild!, 0, 1);
    script.look();
    expect(shown(script.chip)).toBe(false);

    const range = document.createRange();
    range.setStart(document.querySelectorAll("p")[0]!.firstChild!, 0);
    range.setEnd(document.querySelectorAll("p")[1]!.firstChild!, 5);
    window.getSelection()!.removeAllRanges();
    window.getSelection()!.addRange(range);
    script.look();
    expect(shown(script.chip)).toBe(false);

    fromApp({ chatcms: "setup", words, on: false });
    select(document.querySelectorAll("p")[1]!.firstChild!, 0, 11);
    script.look();
    expect(shown(script.chip)).toBe(false);
  });

  it("hides the chip on Esc", () => {
    page(text);
    const script = ready();
    select(document.querySelector("p")!.firstChild!, 6, 25);
    script.look();
    expect(shown(script.chip)).toBe(true);
    document.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    expect(shown(script.chip)).toBe(false);
  });

  it("opens the card from the chip and sends what is typed with the selection", () => {
    page(text);
    const script = ready();
    select(document.querySelector("p")!.firstChild!, 6, 25);
    script.look();
    script.chip.click();
    expect(shown(script.card)).toBe(true);
    expect(script.card.getAttribute("aria-label")).toBe("Chat about the selected text");
    script.ask.value = "Shorter, please";
    script.ask.dispatchEvent(new KeyboardEvent("keydown", { key: "Enter" }));
    expect(posted).toHaveBeenLastCalledWith(
      {
        chatcms: "send",
        text: "bread every morning",
        prompt: "Shorter, please",
        before: "KorenaarFresh ",
        after: ", baked on the Oude Gracht.",
        path: window.location.pathname,
      },
      app,
    );
    expect(shown(script.card)).toBe(false);
  });

  it("sends a quick action at once", () => {
    page(text);
    const script = ready();
    select(document.querySelector("p")!.firstChild!, 6, 25);
    script.look();
    script.chip.click();
    const quick = Array.from(script.card.querySelectorAll<HTMLButtonElement>(".quick button"));
    expect(quick.map((one) => one.textContent)).toEqual(["Make it shorter", "Simpler words", "Fix spelling"]);
    quick[2]!.click();
    expect(posted).toHaveBeenLastCalledWith(expect.objectContaining({ chatcms: "send", prompt: "Fix spelling" }), app);
  });

  it("keeps what was typed as a draft when closed with Esc or a click outside", () => {
    page(text);
    const script = ready();
    select(document.querySelector("p")!.firstChild!, 6, 25);
    script.look();
    script.chip.click();
    script.ask.value = "Half a thought";
    document.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    expect(posted).toHaveBeenLastCalledWith(expect.objectContaining({ chatcms: "draft", prompt: "Half a thought", text: "bread every morning" }), app);
    expect(shown(script.card)).toBe(false);

    select(document.querySelector("p")!.firstChild!, 6, 25);
    script.look();
    script.chip.click();
    script.ask.value = "Another";
    pointer("pointerdown");
    expect(posted).toHaveBeenLastCalledWith(expect.objectContaining({ chatcms: "draft", prompt: "Another" }), app);
  });

  it("opens the card at once on Ctrl+Shift+K", () => {
    page(text);
    const script = ready();
    select(document.querySelector("p")!.firstChild!, 6, 25);
    document.dispatchEvent(new KeyboardEvent("keydown", { key: "K", shiftKey: true, ctrlKey: true }));
    expect(shown(script.card)).toBe(true);
  });

  it("listens to the app only", () => {
    page("<p>Fresh bread</p>");
    const script = start();
    fromApp({ chatcms: "setup", words, on: true }, "https://elsewhere.example");
    fromApp({ chatcms: "setup", words, on: true }, app, null);
    select(document.querySelector("p")!.firstChild!, 0, 5);
    script.look();
    expect(shown(script.chip)).toBe(false);
  });

  it("highlights the changed text when the app asks", () => {
    page("<p>Come in and smell the new rye. Open daily.</p>");
    const set = vi.fn();
    const remove = vi.fn();
    Object.assign(window, { Highlight: class { constructor(readonly range: Range) {} } });
    Object.defineProperty(window, "CSS", { value: { highlights: { set, delete: remove } }, configurable: true });
    start();

    fromApp({ chatcms: "highlight", before: "Come in and smell ", after: " Open daily." });
    expect(set).toHaveBeenCalledTimes(1);
    expect((set.mock.calls[0]![1] as { range: Range }).range.toString()).toBe("the new rye.");
    vi.advanceTimersByTime(4000);
    expect(remove).toHaveBeenCalledWith("chatcms-changed");
  });
});

describe("placing the card", () => {
  const view = { width: 1000, height: 700 };
  const card = { width: 380, height: 150 };

  it("sits 12 px below the selection with its arrow at the selection's end", () => {
    const line = { top: 100, bottom: 120, left: 200, right: 400 };
    expect(placeCard(line, line, view, card)).toEqual({ top: 132, left: 368, width: 380, arrow: 32, above: false });
  });

  it("flips above when there is no room below", () => {
    const line = { top: 600, bottom: 620, left: 200, right: 400 };
    expect(placeCard(line, line, view, card)).toMatchObject({ top: 438, above: true });
  });

  it("stays 12 px inside the edges, and narrows on a narrow preview", () => {
    const line = { top: 100, bottom: 120, left: 900, right: 990 };
    expect(placeCard(line, line, view, card)).toMatchObject({ left: 608 });
    expect(placeCard(line, line, { width: 300, height: 700 }, card)).toMatchObject({ width: 276, left: 12 });
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
