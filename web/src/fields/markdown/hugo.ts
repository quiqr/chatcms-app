import { $nodeSchema, $remark } from "@milkdown/kit/utils";

type Point = { offset?: number };
type MdNode = {
  type: string;
  value?: string;
  name?: string;
  children?: MdNode[];
  position?: { start: Point; end: Point };
};

const opening = /^\{\{([<%])\s*([\w./-]+)(?:(?!\{\{)[^])*?(?:>|%)\}\}/;
const anywhere = /\{\{[<%]\s*\/?\s*[\w./-]+(?:(?!\{\{)[^])*?[>%]\}\}/g;

function closingFor(name: string, delimiter: string) {
  const d = delimiter === "<" ? "<" : "%";
  const e = delimiter === "<" ? ">" : "%";
  return new RegExp(`\\{\\{${d}\\s*/\\s*${name.replace(/[.*+?^${}()|[\]\\/]/g, "\\$&")}\\s*${e}\\}\\}\\s*$`);
}

function sourceOf(node: MdNode, text: string): string {
  const from = node.position?.start.offset;
  const to = node.position?.end.offset;
  return from === undefined || to === undefined ? "" : text.slice(from, to);
}

export function shortcodeName(raw: string): string {
  return raw.match(/^\{\{[<%]\s*\/?\s*([\w./-]+)/)?.[1] ?? "";
}

function foldBlocks(children: MdNode[], text: string): MdNode[] {
  const out: MdNode[] = [];
  for (let i = 0; i < children.length; i++) {
    const node = children[i]!;
    const own = sourceOf(node, text).trim();
    const head = own.match(opening);
    if ((node.type !== "paragraph" && node.type !== "html") || !head) {
      out.push(node);
      continue;
    }
    const [whole, delimiter, name] = head;
    if (whole.length === own.length && !whole.includes("{{" + delimiter + " /") && !closesLater(children, i, text, name!, delimiter!)) {
      out.push(shortcode(node, node, text, name!));
      continue;
    }
    const closing = closingFor(name!, delimiter!);
    let end = -1;
    for (let k = i; k < children.length; k++) {
      if (closing.test(sourceOf(children[k]!, text).trim())) {
        end = k;
        break;
      }
    }
    if (end === -1) {
      out.push(node);
      continue;
    }
    out.push(shortcode(node, children[end]!, text, name!));
    i = end;
  }
  return out;
}

function closesLater(children: MdNode[], at: number, text: string, name: string, delimiter: string): boolean {
  const closing = closingFor(name, delimiter);
  for (let k = at + 1; k < children.length; k++) {
    if (closing.test(sourceOf(children[k]!, text).trim())) return true;
  }
  return false;
}

function shortcode(first: MdNode, last: MdNode, text: string, name: string): MdNode {
  const from = first.position!.start.offset!;
  const to = last.position!.end.offset!;
  return {
    type: "shortcodeBlock",
    value: text.slice(from, to),
    name,
    position: { start: first.position!.start, end: last.position!.end },
  };
}

function splitInline(node: MdNode) {
  if (!node.children) return;
  const next: MdNode[] = [];
  for (const child of node.children) {
    if (child.type !== "text" || !child.value || !child.value.includes("{{")) {
      splitInline(child);
      next.push(child);
      continue;
    }
    let last = 0;
    for (const found of child.value.matchAll(anywhere)) {
      const at = found.index ?? 0;
      if (at > last) next.push({ type: "text", value: child.value.slice(last, at) });
      next.push({ type: "shortcodeInline", value: found[0], name: shortcodeName(found[0]) });
      last = at + found[0].length;
    }
    if (last < child.value.length) next.push({ type: "text", value: child.value.slice(last) });
  }
  node.children = next;
}

function foldEverywhere(node: MdNode, text: string) {
  if (!node.children) return;
  node.children = foldBlocks(node.children, text);
  for (const child of node.children) {
    if (child.type !== "shortcodeBlock") foldEverywhere(child, text);
  }
}

export function hugoShortcodes() {
  return (tree: MdNode, file: { value?: unknown }) => {
    const text = String(file.value ?? "");
    foldEverywhere(tree, text);
    splitInline(tree);
  };
}

export const hugoRemark = $remark("hugoShortcodes", () => hugoShortcodes as never);

export const shortcodeBlock = $nodeSchema("shortcode_block", () => ({
  group: "block",
  atom: true,
  selectable: true,
  draggable: true,
  attrs: { raw: { default: "" }, name: { default: "" } },
  toDOM: (node) => [
    "div",
    { "data-shortcode": node.attrs.name, "data-raw": node.attrs.raw, contenteditable: "false", class: "shortcode-block" },
    node.attrs.name,
  ],
  parseDOM: [
    {
      tag: "div[data-shortcode]",
      getAttrs: (dom) => ({ raw: (dom as HTMLElement).dataset.raw ?? "", name: (dom as HTMLElement).dataset.shortcode ?? "" }),
    },
  ],
  parseMarkdown: {
    match: (node) => node.type === "shortcodeBlock",
    runner: (state, node, type) => {
      state.addNode(type, { raw: node.value as string, name: node.name as string });
    },
  },
  toMarkdown: {
    match: (node) => node.type.name === "shortcode_block",
    runner: (state, node) => {
      state.addNode("html", undefined, node.attrs.raw as string);
    },
  },
}));

export const shortcodeInline = $nodeSchema("shortcode_inline", () => ({
  group: "inline",
  inline: true,
  atom: true,
  selectable: true,
  attrs: { raw: { default: "" }, name: { default: "" } },
  toDOM: (node) => [
    "span",
    { "data-shortcode": node.attrs.name, "data-raw": node.attrs.raw, contenteditable: "false", class: "shortcode-inline" },
    node.attrs.name,
  ],
  parseDOM: [
    {
      tag: "span[data-shortcode]",
      getAttrs: (dom) => ({ raw: (dom as HTMLElement).dataset.raw ?? "", name: (dom as HTMLElement).dataset.shortcode ?? "" }),
    },
  ],
  parseMarkdown: {
    match: (node) => node.type === "shortcodeInline",
    runner: (state, node, type) => {
      state.addNode(type, { raw: node.value as string, name: node.name as string });
    },
  },
  toMarkdown: {
    match: (node) => node.type.name === "shortcode_inline",
    runner: (state, node) => {
      state.addNode("html", undefined, node.attrs.raw as string);
    },
  },
}));

export const hugo = [hugoRemark, shortcodeBlock, shortcodeInline].flat();
