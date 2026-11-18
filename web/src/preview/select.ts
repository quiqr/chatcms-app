export type Selection = {
  chatcms: "selection";
  text: string;
  before: string;
  after: string;
  path: string;
};

export type FromApp =
  | { chatcms: "setup"; label: string }
  | { chatcms: "highlight"; before: string; after: string };

const reach = 200;
const most = 2000;
const highlightFor = 4000;
const name = "chatcms-changed";

function textAround(doc: Document, range: Range) {
  const before = doc.createRange();
  before.setStart(doc.body, 0);
  before.setEnd(range.startContainer, range.startOffset);
  const after = doc.createRange();
  after.setStart(range.endContainer, range.endOffset);
  after.setEnd(doc.body, doc.body.childNodes.length);
  return { before: before.toString().slice(-reach), after: after.toString().slice(0, reach) };
}

function textNodes(doc: Document) {
  const walker = doc.createTreeWalker(doc.body, NodeFilter.SHOW_TEXT);
  const nodes: { node: Text; start: number }[] = [];
  let text = "";
  for (let node = walker.nextNode(); node; node = walker.nextNode()) {
    nodes.push({ node: node as Text, start: text.length });
    text += (node as Text).data;
  }
  return { nodes, text };
}

function point(nodes: { node: Text; start: number }[], at: number): [Text, number] | null {
  for (let i = nodes.length - 1; i >= 0; i--) {
    const entry = nodes[i]!;
    if (entry.start <= at) return [entry.node, Math.min(at - entry.start, entry.node.data.length)];
  }
  return null;
}

export function findBetween(doc: Document, before: string, after: string): Range | null {
  const { nodes, text } = textNodes(doc);
  const head = before.slice(-40);
  const tail = after.slice(0, 40);
  let start = 0;
  if (head) {
    const found = text.lastIndexOf(head);
    if (found < 0) return null;
    start = found + head.length;
  }
  let end = text.length;
  if (tail) {
    const found = text.indexOf(tail, start);
    if (found < 0) return null;
    end = found;
  }
  if (end <= start) return null;
  const from = point(nodes, start);
  const to = point(nodes, end);
  if (!from || !to) return null;
  const range = doc.createRange();
  range.setStart(from[0], from[1]);
  range.setEnd(to[0], to[1]);
  return range;
}

type Highlights = { set(name: string, value: unknown): void; delete(name: string): void };

export function install(win: Window, app: string, where: () => string = () => win.location.pathname) {
  const doc = win.document;
  let label = "";
  let chosen: Omit<Selection, "chatcms"> | null = null;

  const host = doc.createElement("div");
  host.setAttribute("data-chatcms", "");
  const shadow = host.attachShadow({ mode: "open" });
  const button = doc.createElement("button");
  button.type = "button";
  button.textContent = "✨";
  button.setAttribute("style", [
    "position:fixed", "z-index:2147483647", "display:none", "min-width:44px", "min-height:44px",
    "border:0", "border-radius:22px", "background:#1f6f5c", "color:#fff", "font:16px system-ui",
    "box-shadow:0 2px 8px rgba(0,0,0,.25)", "cursor:pointer",
  ].join(";"));
  shadow.appendChild(button);

  const style = doc.createElement("style");
  style.textContent = `::highlight(${name}){background-color:#fde68a;color:inherit}`;

  const hide = () => {
    button.style.display = "none";
    chosen = null;
  };

  const look = () => {
    const selection = win.getSelection();
    if (!selection || selection.rangeCount === 0 || selection.isCollapsed) return hide();
    const range = selection.getRangeAt(0);
    if (!doc.body.contains(range.commonAncestorContainer) || host.contains(range.commonAncestorContainer)) return hide();
    const text = selection.toString().trim();
    if (!text || text.length > most || !label) return hide();
    chosen = { text, ...textAround(doc, range), path: where() };
    const box = range.getBoundingClientRect();
    const top = Math.min(box.bottom + 8, win.innerHeight - 52);
    const left = Math.min(Math.max(box.right - 22, 8), win.innerWidth - 52);
    button.style.top = `${Math.max(top, 8)}px`;
    button.style.left = `${left}px`;
    button.style.display = "block";
  };

  button.addEventListener("pointerdown", (event) => event.preventDefault());
  button.addEventListener("click", () => {
    if (!chosen) return;
    const message: Selection = { chatcms: "selection", ...chosen };
    win.parent.postMessage(message, app);
    hide();
  });

  let timer: ReturnType<typeof setTimeout> | undefined;
  const changed = () => {
    clearTimeout(timer);
    timer = setTimeout(look, 150);
  };
  doc.addEventListener("selectionchange", changed);

  const heard = (event: MessageEvent) => {
    if (event.origin !== app || event.source !== win.parent) return;
    const said = event.data as FromApp;
    if (said?.chatcms === "setup" && typeof said.label === "string") {
      label = said.label;
      button.setAttribute("aria-label", label);
      button.title = label;
    }
    if (said?.chatcms === "highlight") highlight(String(said.before ?? ""), String(said.after ?? ""));
  };
  win.addEventListener("message", heard);

  function highlight(before: string, after: string) {
    const range = findBetween(doc, before, after);
    const highlights = (win as unknown as { CSS?: { highlights?: Highlights } }).CSS?.highlights;
    const Made = (win as unknown as { Highlight?: new (range: Range) => unknown }).Highlight;
    if (!range || !highlights || !Made) return;
    highlights.set(name, new Made(range));
    const box = range.getBoundingClientRect();
    if (box.top < 0 || box.bottom > win.innerHeight) win.scrollBy({ top: box.top - win.innerHeight / 3 });
    setTimeout(() => highlights.delete(name), highlightFor);
  }

  doc.head.appendChild(style);
  doc.body.appendChild(host);
  win.parent.postMessage({ chatcms: "ready" }, app);
  const stop = () => {
    clearTimeout(timer);
    doc.removeEventListener("selectionchange", changed);
    win.removeEventListener("message", heard);
    host.remove();
    style.remove();
  };
  return { look, highlight, stop };
}

const tag = typeof document !== "undefined"
  ? document.querySelector<HTMLScriptElement>('script[data-app][src$="__chatcms/select.js"]')
  : null;
if (tag?.dataset.app && window.parent !== window) install(window, tag.dataset.app);
