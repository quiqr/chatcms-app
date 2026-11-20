import { css } from "@codemirror/lang-css";
import { html } from "@codemirror/lang-html";
import { javascript } from "@codemirror/lang-javascript";
import { json } from "@codemirror/lang-json";
import { markdown } from "@codemirror/lang-markdown";
import { yaml } from "@codemirror/lang-yaml";
import { StreamLanguage } from "@codemirror/language";
import { toml } from "@codemirror/legacy-modes/mode/toml";
import type { Extension } from "@codemirror/state";
import { Decoration, MatchDecorator, ViewPlugin, type DecorationSet, type EditorView, type ViewUpdate } from "@codemirror/view";

const action = Decoration.mark({ class: "cm-go-template" });

const actions = new MatchDecorator({ regexp: /\{\{-?[\s\S]*?-?\}\}/g, decoration: () => action });

export const goTemplates = ViewPlugin.fromClass(
  class {
    decorations: DecorationSet;
    constructor(view: EditorView) {
      this.decorations = actions.createDeco(view);
    }
    update(update: ViewUpdate) {
      this.decorations = actions.updateDeco(update, this.decorations);
    }
  },
  { decorations: (plugin) => plugin.decorations },
);

export function extensionOf(path: string) {
  const name = path.split("/").pop() ?? "";
  const dot = name.lastIndexOf(".");
  return dot < 0 ? "" : name.slice(dot + 1).toLowerCase();
}

export function languageFor(path: string): Extension[] {
  switch (extensionOf(path)) {
    case "html":
    case "htm":
    case "xml":
    case "svg":
      return [html(), goTemplates];
    case "css":
    case "scss":
      return [css(), goTemplates];
    case "js":
    case "mjs":
    case "cjs":
      return [javascript()];
    case "ts":
    case "tsx":
      return [javascript({ typescript: true, jsx: extensionOf(path) === "tsx" })];
    case "json":
      return [json()];
    case "yaml":
    case "yml":
      return [yaml()];
    case "toml":
      return [StreamLanguage.define(toml)];
    case "md":
    case "markdown":
    case "qmd":
      return [markdown(), goTemplates];
    default:
      return [];
  }
}
