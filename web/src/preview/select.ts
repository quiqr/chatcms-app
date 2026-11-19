export type Words = {
  chip: string;
  chipLabel: string;
  card: string;
  placeholder: string;
  send: string;
  quick: string[];
  goesTo: string;
  keys: string;
};

export type Asked = {
  text: string;
  prompt: string;
  before: string;
  after: string;
  path: string;
};

export type ToApp =
  | { chatcms: "ready" }
  | ({ chatcms: "send" } & Asked)
  | ({ chatcms: "draft" } & Asked);

export type FromApp =
  | { chatcms: "setup"; words: Words; on: boolean }
  | { chatcms: "highlight"; before: string; after: string };

const reach = 200;
const most = 2000;
const highlightFor = 4000;
const changedName = "chatcms-changed";
const selectedName = "chatcms-selected";
const gap = 12;
const margin = 12;
const cardWidth = 380;

export const waits = { pointer: 150, keyboard: 400, touch: 500 };

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

export type Box = { top: number; bottom: number; left: number; right: number };

export function placeCard(first: Box, last: Box, view: { width: number; height: number }, card: { width: number; height: number }) {
  const width = Math.min(card.width, view.width - 2 * margin);
  const anchor = last.right;
  const left = Math.min(Math.max(anchor - 32, margin), Math.max(margin, view.width - width - margin));
  const arrow = Math.min(Math.max(anchor - left, 16), width - 16);
  const below = last.bottom + gap;
  const fitsBelow = below + card.height + margin <= view.height;
  const above = first.top - gap - card.height;
  if (!fitsBelow && above >= margin) return { top: above, left, width, arrow, above: true };
  return { top: Math.max(margin, Math.min(below, view.height - card.height - margin)), left, width, arrow, above: false };
}

export function fieldOf(node: Node | null): string | null {
  const element = node instanceof Element ? node : node?.parentElement;
  return element?.closest("[data-quiqr-field]")?.getAttribute("data-quiqr-field") ?? null;
}

type Highlights = { set(name: string, value: unknown): void; delete(name: string): void };

const styles = `
:host { all: initial; }
* { box-sizing: border-box; }
.layer { position: fixed; inset: 0; pointer-events: none; z-index: 2147483647;
  font-family: Geist, ui-sans-serif, system-ui, -apple-system, "Segoe UI", sans-serif; color: #18181B; }
.chip, .card { pointer-events: auto; position: fixed; }
.chip { display: none; align-items: center; gap: 6px; height: 30px; padding: 0 12px 0 6px; border: 0;
  border-radius: 999px; background: #18181B; color: #FFFFFF; font: inherit; font-size: 13px; font-weight: 500;
  box-shadow: 0 6px 16px rgba(16,24,40,.25); cursor: pointer; }
.chip:focus-visible, .card button:focus-visible { outline: 3px solid rgba(11,107,170,.45); outline-offset: 2px; }
.q { width: 18px; height: 18px; border-radius: 5px; background: #FFFFFF; color: #18181B; font-size: 11px;
  font-weight: 700; display: inline-flex; align-items: center; justify-content: center; }
.card { display: none; flex-direction: column; gap: 8px; padding: 10px; background: #FFFFFF;
  border: 1px solid rgba(24,24,27,.14); border-radius: 12px; box-shadow: 0 16px 40px rgba(16,24,40,.22); }
.arrow { position: absolute; width: 12px; height: 12px; background: #FFFFFF; transform: rotate(45deg); }
.arrow.below { top: -7px; border-left: 1px solid rgba(24,24,27,.14); border-top: 1px solid rgba(24,24,27,.14); }
.arrow.above { bottom: -7px; border-right: 1px solid rgba(24,24,27,.14); border-bottom: 1px solid rgba(24,24,27,.14); }
.row { display: flex; align-items: flex-end; gap: 8px; }
textarea { flex: 1; min-width: 0; border: 0; outline: 0; resize: none; font: inherit; font-size: 14px;
  line-height: 1.45; color: #18181B; padding: 6px 4px; background: transparent; }
.send { width: 36px; height: 36px; flex: none; border: 0; border-radius: 50%; background: #0B6BAA; color: #FFFFFF;
  display: flex; align-items: center; justify-content: center; cursor: pointer; }
.send:disabled { opacity: .45; cursor: default; }
.quick { display: flex; flex-wrap: wrap; gap: 6px; }
.quick button { height: 28px; padding: 0 10px; border: 1px solid #E5E5E2; border-radius: 999px; background: #FFFFFF;
  font: inherit; font-size: 12px; color: #3F3F46; cursor: pointer; }
.quick button:hover { background: #F1F1EE; }
.foot { display: flex; align-items: center; gap: 6px; padding: 6px 4px 0; border-top: 1px solid #F1F1EE;
  font-size: 11px; color: #6B6B73; }
.foot .q { width: 16px; height: 16px; background: #18181B; color: #FFFFFF; font-size: 10px; border-radius: 4px; }
.foot .where { flex: 1; }
`;

export function install(win: Window, app: string, where: () => string = () => win.location.pathname) {
  const doc = win.document;
  let words: Words | null = null;
  let on = false;
  let chosen: { range: Range; text: string; before: string; after: string; path: string } | null = null;
  let open = false;
  let held = false;
  let touching = false;
  let shifting = false;
  let timer: ReturnType<typeof setTimeout> | undefined;

  const host = doc.createElement("div");
  host.setAttribute("data-chatcms", "");
  const shadow = host.attachShadow({ mode: "open" });
  const sheet = doc.createElement("style");
  sheet.textContent = styles;
  const layer = doc.createElement("div");
  layer.className = "layer";
  shadow.append(sheet, layer);

  const chip = doc.createElement("button");
  chip.type = "button";
  chip.className = "chip";
  const q = doc.createElement("span");
  q.className = "q";
  q.textContent = "Q";
  q.setAttribute("aria-hidden", "true");
  const chipText = doc.createElement("span");
  chip.append(q, chipText);

  const card = doc.createElement("div");
  card.className = "card";
  card.setAttribute("role", "dialog");
  const arrow = doc.createElement("span");
  arrow.className = "arrow below";
  arrow.setAttribute("aria-hidden", "true");
  const row = doc.createElement("div");
  row.className = "row";
  const ask = doc.createElement("textarea");
  ask.rows = 2;
  const send = doc.createElement("button");
  send.type = "button";
  send.className = "send";
  send.innerHTML = '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M12 19V5M6 11l6-6 6 6"/></svg>';
  row.append(ask, send);
  const quick = doc.createElement("div");
  quick.className = "quick";
  const foot = doc.createElement("div");
  foot.className = "foot";
  const footQ = q.cloneNode(true);
  const whereText = doc.createElement("span");
  whereText.className = "where";
  const keysText = doc.createElement("span");
  foot.append(footQ, whereText, keysText);
  card.append(arrow, row, quick, foot);
  layer.append(chip, card);

  const changedStyle = doc.createElement("style");
  changedStyle.textContent = `::highlight(${changedName}){background-color:#fde68a;color:inherit}::highlight(${selectedName}){background-color:rgba(11,107,170,.2);color:inherit}`;

  const highlights = () => (win as unknown as { CSS?: { highlights?: Highlights } }).CSS?.highlights;
  const Made = () => (win as unknown as { Highlight?: new (range: Range) => unknown }).Highlight;

  function applyWords() {
    if (!words) return;
    chipText.textContent = words.chip;
    chip.setAttribute("aria-label", words.chipLabel);
    card.setAttribute("aria-label", words.card);
    ask.placeholder = words.placeholder;
    ask.setAttribute("aria-label", words.placeholder);
    send.setAttribute("aria-label", words.send);
    quick.replaceChildren(
      ...words.quick.map((label) => {
        const button = doc.createElement("button");
        button.type = "button";
        button.textContent = label;
        button.addEventListener("click", () => submit(label));
        return button;
      }),
    );
    whereText.textContent = words.goesTo;
    keysText.textContent = words.keys;
  }

  function lines(range: Range): { first: Box; last: Box } {
    const rects = Array.from(range.getClientRects?.() ?? []).filter((rect) => rect.width > 0 || rect.height > 0);
    const whole = range.getBoundingClientRect();
    return { first: rects[0] ?? whole, last: rects[rects.length - 1] ?? whole };
  }

  function inView(box: Box) {
    return box.bottom > 0 && box.top < win.innerHeight;
  }

  function hideChip() {
    chip.style.display = "none";
  }

  function placeChip() {
    if (!chosen) return;
    const { last } = lines(chosen.range);
    if (!inView(last)) return hideChip();
    chip.style.display = "inline-flex";
    const width = chip.offsetWidth || 80;
    chip.style.top = `${Math.min(last.bottom + 6, win.innerHeight - 36)}px`;
    chip.style.left = `${Math.min(Math.max(last.right - width / 2, 8), win.innerWidth - width - 8)}px`;
  }

  function placeTheCard() {
    if (!chosen) return;
    const { first, last } = lines(chosen.range);
    if (!inView(first) && !inView(last)) return close();
    const placed = placeCard(first, last, { width: win.innerWidth, height: win.innerHeight }, { width: cardWidth, height: card.offsetHeight || 150 });
    card.style.top = `${placed.top}px`;
    card.style.left = `${placed.left}px`;
    card.style.width = `${placed.width}px`;
    arrow.className = placed.above ? "arrow above" : "arrow below";
    arrow.style.left = `${placed.arrow - 6}px`;
    card.dataset.above = placed.above ? "true" : "false";
  }

  function look() {
    if (open) return;
    if (!on || !words) {
      chosen = null;
      return hideChip();
    }
    const selection = win.getSelection();
    if (!selection || selection.rangeCount === 0 || selection.isCollapsed) {
      chosen = null;
      return hideChip();
    }
    const range = selection.getRangeAt(0);
    if (!doc.body.contains(range.commonAncestorContainer) || host.contains(range.commonAncestorContainer)) return hideChip();
    const text = selection.toString().trim();
    if (text.length < 2 || text.length > most) {
      chosen = null;
      return hideChip();
    }
    const start = fieldOf(range.startContainer);
    const end = fieldOf(range.endContainer);
    if (start !== end) {
      chosen = null;
      return hideChip();
    }
    chosen = { range: range.cloneRange(), text, ...textAround(doc, range), path: where() };
    placeChip();
  }

  function later(wait: number) {
    clearTimeout(timer);
    hideChip();
    timer = setTimeout(look, wait);
  }

  function mark(range: Range | null) {
    const all = highlights();
    const Highlight = Made();
    if (!all || !Highlight) return;
    if (range) all.set(selectedName, new Highlight(range));
    else all.delete(selectedName);
  }

  function openCard() {
    if (!chosen || !on) return;
    open = true;
    hideChip();
    mark(chosen.range);
    card.style.display = "flex";
    placeTheCard();
    send.disabled = ask.value.trim() === "";
    ask.focus();
  }

  function asked(prompt: string): Asked | null {
    if (!chosen) return null;
    return { text: chosen.text, prompt, before: chosen.before, after: chosen.after, path: chosen.path };
  }

  function close() {
    if (!open) return;
    const typed = ask.value.trim();
    const about = asked(typed);
    if (typed && about) win.parent.postMessage({ chatcms: "draft", ...about } satisfies ToApp, app);
    reset();
  }

  function reset() {
    open = false;
    ask.value = "";
    card.style.display = "none";
    mark(null);
    chosen = null;
    hideChip();
  }

  function submit(prompt: string) {
    const text = prompt.trim();
    const about = asked(text);
    if (!text || !about) return;
    win.parent.postMessage({ chatcms: "send", ...about } satisfies ToApp, app);
    reset();
  }

  chip.addEventListener("pointerdown", (event) => event.preventDefault());
  chip.addEventListener("click", openCard);
  send.addEventListener("click", () => submit(ask.value));
  ask.addEventListener("input", () => {
    send.disabled = ask.value.trim() === "";
  });
  ask.addEventListener("keydown", (event) => {
    if (event.key === "Enter" && !event.shiftKey) {
      event.preventDefault();
      submit(ask.value);
    }
  });

  const pressed = (event: PointerEvent) => {
    if (event.composedPath().includes(host)) return;
    if (open) return close();
    held = true;
    touching = event.pointerType === "touch";
    hideChip();
  };
  const released = (event: PointerEvent) => {
    if (event.composedPath().includes(host)) return;
    held = false;
    if (!touching) later(waits.pointer);
  };
  const keys = (event: KeyboardEvent) => {
    if (event.key === "Escape") {
      if (open) return close();
      chosen = null;
      return hideChip();
    }
    if (event.key.toLowerCase() === "k" && event.shiftKey && (event.ctrlKey || event.metaKey)) {
      event.preventDefault();
      clearTimeout(timer);
      look();
      openCard();
      return;
    }
    shifting = event.shiftKey;
  };
  const changed = () => {
    if (open || held) return;
    later(touching ? waits.touch : shifting ? waits.keyboard : waits.pointer);
  };
  const moved = () => {
    if (open) placeTheCard();
    else if (chip.style.display !== "none") placeChip();
  };

  const heard = (event: MessageEvent) => {
    if (event.origin !== app || event.source !== win.parent) return;
    const said = event.data as FromApp;
    if (said?.chatcms === "setup" && said.words) {
      words = said.words;
      on = Boolean(said.on);
      applyWords();
      if (!on) reset();
    }
    if (said?.chatcms === "highlight") highlight(String(said.before ?? ""), String(said.after ?? ""));
  };

  function highlight(before: string, after: string) {
    const range = findBetween(doc, before, after);
    const all = highlights();
    const Highlight = Made();
    if (!range || !all || !Highlight) return;
    all.set(changedName, new Highlight(range));
    const box = range.getBoundingClientRect();
    if (box.top < 0 || box.bottom > win.innerHeight) win.scrollBy({ top: box.top - win.innerHeight / 3 });
    setTimeout(() => all.delete(changedName), highlightFor);
  }

  doc.addEventListener("pointerdown", pressed, true);
  doc.addEventListener("pointerup", released, true);
  doc.addEventListener("keydown", keys, true);
  doc.addEventListener("selectionchange", changed);
  win.addEventListener("scroll", moved, true);
  win.addEventListener("resize", moved);
  win.addEventListener("message", heard);
  doc.head.appendChild(changedStyle);
  doc.body.appendChild(host);
  win.parent.postMessage({ chatcms: "ready" } satisfies ToApp, app);

  const stop = () => {
    clearTimeout(timer);
    doc.removeEventListener("pointerdown", pressed, true);
    doc.removeEventListener("pointerup", released, true);
    doc.removeEventListener("keydown", keys, true);
    doc.removeEventListener("selectionchange", changed);
    win.removeEventListener("scroll", moved, true);
    win.removeEventListener("resize", moved);
    win.removeEventListener("message", heard);
    host.remove();
    changedStyle.remove();
  };
  return { look, highlight, stop, chip, card, ask, openCard };
}

const tag = typeof document !== "undefined"
  ? document.querySelector<HTMLScriptElement>('script[data-app][src$="__chatcms/select.js"]')
  : null;
if (tag && window.parent !== window) install(window, tag.dataset.app || window.location.origin);
