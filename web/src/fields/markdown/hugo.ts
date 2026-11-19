import { $nodeSchema, $remark } from "@milkdown/kit/utils";

type Point = { offset?: number };
type MdNode = {
  type: string;
  value?: string;
  name?: string;
  children?: MdNode[];
  position?: { start: Point; end: Point };
};

type Syntax = { opening: RegExp; anywhere: RegExp };

const hugoSyntax: Syntax = {
  opening: /^\{\{([<%])\s*([\w./-]+)(?:(?!\{\{)[^])*?(?:>|%)\}\}/,
  anywhere: /\{\{[<%]\s*\/?\s*[\w./-]+(?:(?!\{\{)[^])*?[>%]\}\}/g,
};

const quartoSyntax: Syntax = {
  opening: /^\{\{(<)\s*([\w./-]+)(?:(?!\{\{)[^])*?>\}\}/,
  anywhere: /\{\{<\s*[\w./-]+(?:(?!\{\{)[^])*?>\}\}/g,
};

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

function foldBlocks(children: MdNode[], text: string, syntax: Syntax): MdNode[] {
  const out: MdNode[] = [];
  for (let i = 0; i < children.length; i++) {
    const node = children[i]!;
    const own = sourceOf(node, text).trim();
    const head = own.match(syntax.opening);
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

const inlineTypes = new Set([
  "text", "emphasis", "strong", "delete", "link", "linkReference", "inlineCode",
  "html", "break", "image", "imageReference", "footnoteReference",
]);

function phrasing(node: MdNode) {
  return node.children !== undefined && node.children.length > 0 && node.children.every((child) => inlineTypes.has(child.type));
}

function startOf(node: MdNode) {
  return node.position?.start.offset;
}

function endOf(node: MdNode) {
  return node.position?.end.offset;
}

function textPiece(value: string): MdNode {
  return { type: "text", value };
}

function splitInline(node: MdNode, text: string, syntax: Syntax) {
  if (!node.children) return;
  if (!phrasing(node)) {
    for (const child of node.children) if (child.type !== "shortcodeBlock") splitInline(child, text, syntax);
    return;
  }
  const from = startOf(node.children[0]!);
  const to = endOf(node.children[node.children.length - 1]!);
  if (from === undefined || to === undefined) return;
  const source = text.slice(from, to);
  if (!source.includes("{{")) return;

  let children = node.children;
  const found = [...source.matchAll(syntax.anywhere)].reverse();
  for (const match of found) {
    const start = from + (match.index ?? 0);
    const end = start + match[0].length;
    const first = children.findIndex((child) => (endOf(child) ?? 0) > start);
    let last = -1;
    for (let k = children.length - 1; k >= 0; k--) {
      if ((startOf(children[k]!) ?? Infinity) < end) {
        last = k;
        break;
      }
    }
    if (first === -1 || last < first) continue;
    const head = children[first]!;
    const tail = children[last]!;
    const headFrom = startOf(head);
    const tailTo = endOf(tail);
    if (headFrom === undefined || tailTo === undefined) continue;
    const before = text.slice(headFrom, start);
    const after = text.slice(end, tailTo);
    if (before !== "" && (head.type !== "text" || !head.value?.startsWith(before))) continue;
    if (after !== "" && (tail.type !== "text" || !tail.value?.endsWith(after))) continue;
    const pieces: MdNode[] = [];
    if (before !== "") pieces.push({ ...textPiece(before), position: { start: { offset: headFrom }, end: { offset: start } } });
    pieces.push({
      type: "shortcodeInline",
      value: match[0],
      name: shortcodeName(match[0]),
      position: { start: { offset: start }, end: { offset: end } },
    });
    if (after !== "") pieces.push({ ...textPiece(after), position: { start: { offset: end }, end: { offset: tailTo } } });
    children = [...children.slice(0, first), ...pieces, ...children.slice(last + 1)];
  }
  node.children = children;
}

function foldEverywhere(node: MdNode, text: string, syntax: Syntax) {
  if (!node.children) return;
  node.children = foldBlocks(node.children, text, syntax);
  for (const child of node.children) {
    if (child.type !== "shortcodeBlock") foldEverywhere(child, text, syntax);
  }
}

function shortcodesIn(syntax: Syntax) {
  return () => (tree: MdNode, file: { value?: unknown }) => {
    const text = String(file.value ?? "");
    foldEverywhere(tree, text, syntax);
    splitInline(tree, text, syntax);
  };
}

export const hugoShortcodes = shortcodesIn(hugoSyntax);
export const quartoShortcodes = shortcodesIn(quartoSyntax);

export const hugoRemark = $remark("hugoShortcodes", () => hugoShortcodes as never);
export const quartoRemark = $remark("quartoShortcodes", () => quartoShortcodes as never);

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
export const quarto = [quartoRemark, shortcodeBlock, shortcodeInline].flat();
