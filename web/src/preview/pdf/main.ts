import { GlobalWorkerOptions, TextLayer, getDocument, type PDFDocumentProxy, type PDFPageProxy } from "pdfjs-dist";
import worker from "pdfjs-dist/build/pdf.worker.min.mjs?url";

import { sayIn, type Key } from "../../lib/say";
import { install } from "../select";
import {
  fileOf,
  fitsTheWidth,
  languageOf,
  nextZoom,
  ownSize,
  pageWithin,
  percentOf,
  remember,
  remembered,
  scaleFor,
} from "./where";

type Page = { number: number; page: PDFPageProxy; box: HTMLElement; canvas: HTMLCanvasElement; text: HTMLElement };
type Outline = { title: string; dest: string | unknown[] | null; items: Outline[] };

const icons = {
  previous: '<path d="m15 18-6-6 6-6"/>',
  next: '<path d="m9 18 6-6-6-6"/>',
  out: '<path d="M5 12h14"/>',
  in: '<path d="M5 12h14"/><path d="M12 5v14"/>',
  contents: '<path d="M3 12h.01"/><path d="M3 18h.01"/><path d="M3 6h.01"/><path d="M8 12h13"/><path d="M8 18h13"/><path d="M8 6h13"/>',
  download: '<path d="M12 15V3"/><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><path d="m7 10 5 5 5-5"/>',
};

function sessionStore(): Storage | null {
  try {
    return window.sessionStorage;
  } catch {
    return null;
  }
}

function element<K extends keyof HTMLElementTagNameMap>(tag: K, className?: string): HTMLElementTagNameMap[K] {
  const made = document.createElement(tag);
  if (className) made.className = className;
  return made;
}

function iconButton(icon: keyof typeof icons, label: string): HTMLButtonElement {
  const button = element("button", "tool");
  button.type = "button";
  button.setAttribute("aria-label", label);
  button.title = label;
  button.innerHTML = `<svg viewBox="0 0 24 24" aria-hidden="true">${icons[icon]}</svg>`;
  return button;
}

async function show() {
  const file = fileOf(window.location);
  const holder = document.getElementById("pages")!;
  if (!file) return;
  const language = languageOf(window.location);
  const say = (key: Key, fill?: Record<string, string | number>) => sayIn(language, key, fill);
  document.documentElement.lang = language;
  const storage = sessionStore();
  let current = remembered(storage, file, window.location.hash);

  const said = document.querySelector<HTMLMetaElement>('meta[name="chatcms-app"]');
  const app = said ? said.content || window.location.origin : "";
  if (app && window.parent !== window) {
    install(window, app, () => `${file.pathname}#page=${current}`);
  }

  GlobalWorkerOptions.workerSrc = worker;
  const pdf = await getDocument({ url: file.href }).promise;
  const pages: Page[] = [];
  for (let number = 1; number <= pdf.numPages; number++) {
    const box = element("section", "page");
    box.dataset.page = String(number);
    const canvas = element("canvas");
    const text = element("div", "textLayer");
    box.append(canvas, text);
    holder.append(box);
    pages.push({ number, page: await pdf.getPage(number), box, canvas, text });
  }
  current = pageWithin(current, pages.length);

  const fit = fitsTheWidth(window.location);
  const first = pages[0]?.page.getViewport({ scale: 1 }).width ?? 0;
  let scale = scaleFor(first, document.documentElement.clientWidth, fit);
  let drawing = 0;

  const tools = buildTools(say, pdf.numPages, file);
  document.body.prepend(tools.bar);

  function size() {
    for (const one of pages) {
      const viewport = one.page.getViewport({ scale });
      one.box.style.width = `${viewport.width}px`;
      one.box.style.height = `${viewport.height}px`;
    }
  }

  async function paint(one: Page, at: number, token: number) {
    const viewport = one.page.getViewport({ scale: at });
    const ratio = window.devicePixelRatio || 1;
    const canvas = element("canvas");
    canvas.width = Math.floor(viewport.width * ratio);
    canvas.height = Math.floor(viewport.height * ratio);
    canvas.style.width = `${viewport.width}px`;
    canvas.style.height = `${viewport.height}px`;
    await one.page.render({ canvas, viewport, transform: ratio === 1 ? undefined : [ratio, 0, 0, ratio, 0, 0] }).promise;
    if (token !== drawing) return;
    const text = element("div", "textLayer");
    await new TextLayer({ textContentSource: one.page.streamTextContent(), container: text, viewport }).render();
    if (token !== drawing) return;
    one.box.style.setProperty("--scale-factor", String(at));
    one.box.style.setProperty("--total-scale-factor", String(at));
    one.canvas.replaceWith(canvas);
    one.text.replaceWith(text);
    one.canvas = canvas;
    one.text = text;
  }

  async function draw() {
    const token = ++drawing;
    const at = scale;
    size();
    const order = [...pages].sort((a, b) => Math.abs(a.number - current) - Math.abs(b.number - current));
    for (const one of order) {
      if (token !== drawing) return;
      await paint(one, at, token);
    }
  }

  function goTo(number: number) {
    const target = pages[pageWithin(number, pages.length) - 1];
    if (!target) return;
    current = target.number;
    target.box.scrollIntoView();
    tools.showPage(current);
  }

  function zoomTo(next: number) {
    if (Math.abs(next - scale) < 0.0001) return;
    const keep = current;
    scale = next;
    tools.showZoom(percentOf(scale));
    size();
    goTo(keep);
    void draw();
  }

  tools.onPage(goTo);
  tools.onZoom((direction) => zoomTo(direction === 0 ? ownSize : nextZoom(scale, direction)));
  tools.showZoom(percentOf(scale));
  tools.showPage(current);

  size();
  pages[current - 1]?.box.scrollIntoView();
  await draw();

  const outline = (await pdf.getOutline().catch(() => null)) as Outline[] | null;
  if (outline && outline.length > 0) {
    tools.offerContents(outline, async (dest) => {
      const number = await pageOf(pdf, dest);
      if (number) goTo(number);
    });
  }

  const watcher = new IntersectionObserver(
    (entries) => {
      const seen = entries.filter((entry) => entry.isIntersecting).sort((a, b) => b.intersectionRatio - a.intersectionRatio)[0];
      if (!seen) return;
      current = Number((seen.target as HTMLElement).dataset.page) || 1;
      remember(storage, file, current);
      tools.showPage(current);
    },
    { threshold: [0.25, 0.5, 0.75] },
  );
  pages.forEach((one) => watcher.observe(one.box));
}

async function pageOf(pdf: PDFDocumentProxy, dest: Outline["dest"]): Promise<number | null> {
  try {
    const explicit = typeof dest === "string" ? await pdf.getDestination(dest) : dest;
    const target = explicit?.[0];
    if (target == null) return null;
    const index = typeof target === "number" ? target : await pdf.getPageIndex(target as Parameters<PDFDocumentProxy["getPageIndex"]>[0]);
    return index + 1;
  } catch {
    return null;
  }
}

function buildTools(say: (key: Key, fill?: Record<string, string | number>) => string, total: number, file: URL) {
  const bar = element("header", "tools");
  bar.setAttribute("role", "toolbar");
  bar.setAttribute("aria-label", say("pdf.tools"));

  const contents = iconButton("contents", say("pdf.contents"));
  contents.hidden = true;
  contents.setAttribute("aria-expanded", "false");

  const previous = iconButton("previous", say("pdf.previous"));
  const field = element("input", "page-field");
  field.type = "text";
  field.inputMode = "numeric";
  field.setAttribute("aria-label", say("pdf.page"));
  const of = element("span", "of");
  of.textContent = say("pdf.total", { total });
  const next = iconButton("next", say("pdf.next"));

  const out = iconButton("out", say("pdf.zoom-out"));
  const percent = element("button", "tool percent");
  percent.type = "button";
  percent.setAttribute("aria-label", say("pdf.own-size"));
  percent.title = say("pdf.own-size");
  const into = iconButton("in", say("pdf.zoom-in"));

  const download = element("a", "tool");
  download.href = file.href;
  download.download = decodeURIComponent(file.pathname.split("/").pop() ?? "document.pdf");
  download.setAttribute("aria-label", say("pdf.download"));
  download.title = say("pdf.download");
  download.innerHTML = `<svg viewBox="0 0 24 24" aria-hidden="true">${icons.download}</svg>`;

  const paging = element("div", "group");
  paging.append(previous, field, of, next);
  const zooming = element("div", "group");
  zooming.append(out, percent, into);
  const spacer = element("div", "spacer");
  bar.append(contents, paging, spacer, zooming, download);

  const panel = element("nav", "contents");
  panel.setAttribute("aria-label", say("pdf.contents"));
  panel.hidden = true;
  document.body.append(panel);

  let shown = 1;
  function showPage(number: number) {
    shown = number;
    if (document.activeElement !== field) field.value = String(number);
    previous.disabled = number <= 1;
    next.disabled = number >= total;
  }

  function onPage(go: (number: number) => void) {
    previous.addEventListener("click", () => go(shown - 1));
    next.addEventListener("click", () => go(shown + 1));
    field.addEventListener("keydown", (event) => {
      if (event.key !== "Enter") return;
      event.preventDefault();
      go(Number(field.value));
      field.value = String(shown);
    });
    field.addEventListener("blur", () => {
      field.value = String(shown);
    });
  }

  function showZoom(value: number) {
    percent.textContent = say("pdf.zoom", { percent: value });
  }

  function onZoom(zoom: (direction: 1 | -1 | 0) => void) {
    out.addEventListener("click", () => zoom(-1));
    into.addEventListener("click", () => zoom(1));
    percent.addEventListener("click", () => zoom(0));
  }

  function setOpen(open: boolean) {
    panel.hidden = !open;
    contents.setAttribute("aria-expanded", String(open));
    contents.setAttribute("aria-label", open ? say("pdf.contents-close") : say("pdf.contents"));
    document.body.classList.toggle("with-contents", open);
  }

  function offerContents(outline: Outline[], choose: (dest: Outline["dest"]) => void) {
    const narrow = () => document.documentElement.clientWidth < 640;
    function list(items: Outline[]): HTMLUListElement {
      const ul = element("ul");
      for (const item of items) {
        const li = element("li");
        const button = element("button", "entry");
        button.type = "button";
        button.textContent = item.title;
        button.addEventListener("click", () => {
          choose(item.dest);
          if (narrow()) setOpen(false);
        });
        li.append(button);
        if (item.items.length > 0) li.append(list(item.items));
        ul.append(li);
      }
      return ul;
    }
    panel.append(list(outline));
    contents.hidden = false;
    contents.addEventListener("click", () => setOpen(panel.hidden));
  }

  return { bar, showPage, onPage, showZoom, onZoom, offerContents };
}

if (typeof document !== "undefined" && document.getElementById("pages")) void show();
