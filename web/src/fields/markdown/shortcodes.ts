import type { Node } from "@milkdown/kit/prose/model";
import type { EditorView, NodeView } from "@milkdown/kit/prose/view";

import { say } from "../../lib/say";
import type { ShortcodeDeclaration } from "../../lib/tenancy";
import { build, rewrite, valuesOf, type Declared } from "./hugoTag";

export type { ShortcodeDeclaration };

export type ShortcodeEdit = {
  declaration: ShortcodeDeclaration;
  values: Record<string, string>;
  inner: string;
  done: (values: Record<string, string>, inner: string) => void;
};

export type ShortcodeOptions = {
  dialect?: string;
  declarations: () => ShortcodeDeclaration[];
  edit?: (request: ShortcodeEdit) => void;
};

export function declaredOf(declaration: ShortcodeDeclaration): Declared {
  return {
    key: declaration.key,
    inner: declaration.inner,
    order: declaration.fields.map((field) => field.path),
    positions: declaration.positions ?? {},
  };
}

export function newShortcode(declaration: ShortcodeDeclaration, values: Record<string, string>, inner: string): string {
  return build(declaredOf(declaration), values, inner);
}

function summaryOf(declaration: ShortcodeDeclaration, values: Record<string, string>) {
  const text = declaration.fields.find((field) => ["string", "markdown", "easymde", "select"].includes(field.type) && values[field.path]);
  const picture = declaration.fields.find((field) => field.type === "image-select" && values[field.path]);
  const said = text ? values[text.path]! : "";
  const label = text?.options?.find((option) => option.value === said)?.label;
  return { text: label ?? said, picture: picture ? values[picture.path] : "" };
}

export function shortcodeView(tag: "div" | "span", options: ShortcodeOptions) {
  return (first: Node, view: EditorView, getPos: () => number | undefined): NodeView => {
    let node = first;
    const dom = document.createElement(tag);
    dom.setAttribute("contenteditable", "false");

    function draw() {
      dom.className = tag === "div" ? "shortcode-block" : "shortcode-inline";
      dom.dataset.shortcode = node.attrs.name as string;
      dom.dataset.raw = node.attrs.raw as string;
      dom.replaceChildren();
      const declaration = options.declarations().find((one) => one.key === node.attrs.name);
      const read = declaration ? valuesOf(node.attrs.raw as string, declaredOf(declaration)) : null;
      if (!declaration || !read || !options.edit) {
        delete dom.dataset.declared;
        dom.textContent = node.attrs.name as string;
        return;
      }
      dom.dataset.declared = "true";
      const button = document.createElement("button");
      button.type = "button";
      button.className = "shortcode-edit";
      button.setAttribute("aria-label", say("body-editor.edit-shortcode", { title: declaration.title }));
      const summary = summaryOf(declaration, read.values);
      if (summary.picture) {
        const picture = document.createElement("img");
        picture.src = summary.picture;
        picture.alt = "";
        picture.addEventListener("error", () => picture.remove());
        button.append(picture);
      }
      const title = document.createElement("strong");
      title.textContent = declaration.title;
      button.append(title);
      if (summary.text) {
        const text = document.createElement("span");
        text.textContent = summary.text;
        button.append(text);
      }
      button.addEventListener("mousedown", (event) => event.preventDefault());
      button.addEventListener("click", () => {
        options.edit?.({
          declaration,
          values: read.values,
          inner: read.inner,
          done: (values, inner) => {
            const raw = rewrite(node.attrs.raw as string, declaredOf(declaration), values, inner);
            const at = getPos();
            if (raw === null || at === undefined || raw === node.attrs.raw) return;
            view.dispatch(view.state.tr.setNodeMarkup(at, undefined, { ...node.attrs, raw }));
          },
        });
      });
      dom.append(button);
    }

    draw();
    return {
      dom,
      update(next) {
        if (next.type !== node.type) return false;
        node = next;
        draw();
        return true;
      },
      stopEvent: (event) => event.target instanceof HTMLElement && event.target.closest("button") !== null,
      ignoreMutation: () => true,
    };
  };
}
