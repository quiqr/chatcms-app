import { GlobalWorkerOptions, TextLayer, getDocument } from "pdfjs-dist";
import worker from "pdfjs-dist/build/pdf.worker.min.mjs?url";

import { install } from "../select";
import { fileOf, fitsTheWidth, remember, remembered, scaleFor } from "./where";

function sessionStore(): Storage | null {
  try {
    return window.sessionStorage;
  } catch {
    return null;
  }
}

async function show() {
  const file = fileOf(window.location);
  const holder = document.getElementById("pages")!;
  if (!file) return;
  const storage = sessionStore();
  let current = remembered(storage, file, window.location.hash);

  const said = document.querySelector<HTMLMetaElement>('meta[name="chatcms-app"]');
  const app = said ? said.content || window.location.origin : "";
  if (app && window.parent !== window) {
    install(window, app, () => `${file.pathname}#page=${current}`);
  }

  GlobalWorkerOptions.workerSrc = worker;
  const pdf = await getDocument({ url: file.href }).promise;
  const pages: HTMLElement[] = [];
  const fit = fitsTheWidth(window.location);

  for (let number = 1; number <= pdf.numPages; number++) {
    const page = await pdf.getPage(number);
    const scale = scaleFor(page.getViewport({ scale: 1 }).width, document.documentElement.clientWidth, fit);
    const viewport = page.getViewport({ scale });
    const ratio = window.devicePixelRatio || 1;

    const box = document.createElement("section");
    box.className = "page";
    box.dataset.page = String(number);
    box.style.width = `${viewport.width}px`;
    box.style.height = `${viewport.height}px`;
    box.style.setProperty("--scale-factor", String(scale));
    box.style.setProperty("--total-scale-factor", String(scale));

    const canvas = document.createElement("canvas");
    canvas.width = Math.floor(viewport.width * ratio);
    canvas.height = Math.floor(viewport.height * ratio);
    canvas.style.width = `${viewport.width}px`;
    canvas.style.height = `${viewport.height}px`;
    const text = document.createElement("div");
    text.className = "textLayer";
    box.append(canvas, text);
    holder.append(box);
    pages.push(box);

    await page.render({ canvas, viewport, transform: ratio === 1 ? undefined : [ratio, 0, 0, ratio, 0, 0] }).promise;
    await new TextLayer({ textContentSource: page.streamTextContent(), container: text, viewport }).render();
    if (number === current) box.scrollIntoView();
  }

  const watcher = new IntersectionObserver(
    (entries) => {
      const seen = entries.filter((entry) => entry.isIntersecting).sort((a, b) => b.intersectionRatio - a.intersectionRatio)[0];
      if (!seen) return;
      current = Number((seen.target as HTMLElement).dataset.page) || 1;
      remember(storage, file, current);
    },
    { threshold: [0.25, 0.5, 0.75] },
  );
  pages.forEach((one) => watcher.observe(one));
}

if (typeof document !== "undefined" && document.getElementById("pages")) void show();
