import { editorViewCtx, type Editor } from "@milkdown/kit/core";
import type { Node } from "@milkdown/kit/prose/model";
import { readdirSync, readFileSync, statSync } from "node:fs";
import { join, relative } from "node:path";
import { afterEach, describe, expect, it } from "vitest";

import { TextSelection } from "@milkdown/kit/prose/state";

import { format, openMarkdown } from "../MarkdownEditor";

const fixtures = process.env.CHATCMS_MARKDOWN_FIXTURES ?? join(__dirname, "../../../../../test/fixtures");

function markdownFiles(dir: string): string[] {
  return readdirSync(dir).flatMap((name) => {
    const path = join(dir, name);
    if (statSync(path).isDirectory()) return name === "node_modules" ? [] : markdownFiles(path);
    return name.endsWith(".md") ? [path] : [];
  });
}

function bodyOf(text: string) {
  const found = text.match(/^---\n[\s\S]*?\n---\n([\s\S]*)$/);
  return found ? found[1]! : text;
}

const open: Editor[] = [];

afterEach(async () => {
  while (open.length > 0) await open.pop()!.destroy();
});

async function opened(source: string) {
  const edits: string[] = [];
  const root = document.createElement("div");
  document.body.append(root);
  const made = await openMarkdown(root, source, "Page text", (next) => edits.push(next));
  if (!made) throw new Error("the text did not line up");
  open.push(made.editor);
  const act = (change: (view: ReturnType<typeof viewOf>) => void) => made.editor.action((ctx) => change(ctx.get(editorViewCtx)));
  const viewOf = () => made.editor.action((ctx) => ctx.get(editorViewCtx));
  return { made, edits, act, doc: () => viewOf().state.doc, root };
}

function blockStarting(doc: Node, text: string) {
  let found: { node: Node; pos: number } | null = null;
  doc.forEach((node, pos) => {
    if (!found && node.textContent.startsWith(text)) found = { node, pos };
  });
  if (!found) throw new Error(`no block starts with ${text}`);
  return found as { node: Node; pos: number };
}

function onlyDifference(before: string, after: string) {
  const a = before.split("\n");
  const b = after.split("\n");
  let start = 0;
  while (start < a.length && a[start] === b[start]) start++;
  let endA = a.length - 1;
  let endB = b.length - 1;
  while (endA >= start && endB >= start && a[endA] === b[endB]) {
    endA--;
    endB--;
  }
  return { removed: a.slice(start, endA + 1), added: b.slice(start, endB + 1) };
}

const tricky = readFileSync(join(fixtures, "markdown/tricky.md"), "utf8");

describe("the text keeps its own writing", () => {
  const files = markdownFiles(fixtures);

  it.each(files.map((path) => [relative(fixtures, path), path]))("%s opens without a save and lines up", async (_, path) => {
    const body = bodyOf(readFileSync(path, "utf8"));
    const { edits } = await opened(body);
    expect(edits).toEqual([]);
  });

  it("writes back only the paragraph that changed", async () => {
    const { act, doc, edits } = await opened(tricky);
    const { pos } = blockStarting(doc(), "Escapes");
    act((view) => view.dispatch(view.state.tr.insertText("Changed. ", pos + 1)));
    expect(onlyDifference(tricky, edits.at(-1)!)).toEqual({
      removed: ["Escapes: \\* not a list, 3 \\< 4, a_b_c, 100%."],
      added: [expect.stringMatching(/^Changed\. Escapes: \\\* not a list, 3 \\?< 4, a_b_c, 100%\.$/)],
    });
  });

  it("adds a paragraph with one blank line, and removes one cleanly", async () => {
    const { act, doc, edits } = await opened(tricky);
    const { pos, node } = blockStarting(doc(), "Escapes");
    act((view) => view.dispatch(view.state.tr.insert(pos, view.state.schema.nodes.paragraph!.create(null, view.state.schema.text("A new paragraph.")))));
    expect(onlyDifference(tricky, edits.at(-1)!)).toEqual({ removed: [], added: ["A new paragraph.", ""] });

    const second = await opened(tricky);
    second.act((view) => view.dispatch(view.state.tr.delete(pos, pos + node.nodeSize)));
    expect(onlyDifference(tricky, second.edits.at(-1)!)).toEqual({
      removed: ["Escapes: \\* not a list, 3 \\< 4, a_b_c, 100%.", ""],
      added: [],
    });
  });

  it("writes an empty text's first words without inventing the rest", async () => {
    const { act, edits } = await opened("\n");
    act((view) => view.dispatch(view.state.tr.insertText("Hello.", 1)));
    expect(edits.at(-1)).toBe("Hello.\n");
  });
});

describe("shortcodes", () => {
  it("are locked blocks and chips named after the shortcode", async () => {
    const { doc } = await opened(tricky);
    const names: Array<[string, string]> = [];
    doc().descendants((node) => {
      if (node.type.name.startsWith("shortcode")) names.push([node.type.name, node.attrs.name as string]);
    });
    expect(names).toEqual([
      ["shortcode_block", "figure"],
      ["shortcode_block", "notice"],
      ["shortcode_inline", "icon"],
    ]);
  });

  it("fold a paired shortcode across blank lines into one block", async () => {
    const source = "Before.\n\n{{% notice warning %}}\n\nInside.\n\n{{% /notice %}}\n\nAfter.\n";
    const { doc, act, edits } = await opened(source);
    expect(doc().childCount).toBe(3);
    expect(doc().child(1).attrs.raw).toBe("{{% notice warning %}}\n\nInside.\n\n{{% /notice %}}");
    act((view) => view.dispatch(view.state.tr.insertText("Just ", 1)));
    expect(edits.at(-1)).toBe("Just Before.\n\n{{% notice warning %}}\n\nInside.\n\n{{% /notice %}}\n\nAfter.\n");
  });

  it("keep their text when the line around an inline one is edited", async () => {
    const { act, doc, edits } = await opened(tricky);
    const { pos } = blockStarting(doc(), "Inline ");
    act((view) => view.dispatch(view.state.tr.insertText("An ", pos + 1)));
    expect(onlyDifference(tricky, edits.at(-1)!)).toEqual({
      removed: ['Inline {{< icon "bread" >}} shortcode in text.'],
      added: ['An Inline {{< icon "bread" >}} shortcode in text.'],
    });
  });

  it("are drawn as locked, named elements", async () => {
    const { root } = await opened('{{< figure src="/a.jpg" >}}\n');
    const block = root.querySelector("[data-shortcode='figure']");
    expect(block).not.toBeNull();
    expect(block?.getAttribute("contenteditable")).toBe("false");
  });
});

describe("the editable area", () => {
  it("is a labelled, multi-line text box", async () => {
    const { root } = await opened("Text.\n");
    const box = root.querySelector("[role='textbox']");
    expect(box?.getAttribute("aria-label")).toBe("Page text");
    expect(box?.getAttribute("aria-multiline")).toBe("true");
  });
});

describe("the toolbar", () => {
  it("makes the chosen word bold, and the text says so in Markdown", async () => {
    const source = "Some plain words.\n";
    const { made, act, edits } = await opened(source);
    act((view) => view.dispatch(view.state.tr.setSelection(TextSelection.create(view.state.doc, 6, 11))));
    format(made.editor, { kind: "bold" });
    expect(edits.at(-1)).toBe("Some **plain** words.\n");
  });

  it("links the chosen words to an address", async () => {
    const { made, act, edits } = await opened("Visit the shop.\n");
    act((view) => view.dispatch(view.state.tr.setSelection(TextSelection.create(view.state.doc, 11, 15))));
    format(made.editor, { kind: "link", href: "/shop/" });
    expect(edits.at(-1)).toBe("Visit the [shop](/shop/).\n");
  });

  it("turns a paragraph into a heading, leaving the rest alone", async () => {
    const { made, edits } = await opened("Title\n\nBody text.\n");
    format(made.editor, { kind: "heading" });
    expect(edits.at(-1)).toBe("## Title\n\nBody text.\n");
  });
});
