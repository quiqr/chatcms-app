import { editorViewCtx, type Editor } from "@milkdown/kit/core";
import type { Node } from "@milkdown/kit/prose/model";
import { readdirSync, readFileSync, statSync } from "node:fs";
import { join, relative } from "node:path";
import { afterEach, describe, expect, it } from "vitest";

import { TextSelection } from "@milkdown/kit/prose/state";

import { format, insertShortcode, openMarkdown } from "../MarkdownEditor";
import type { ShortcodeDeclaration, ShortcodeEdit, ShortcodeOptions } from "../shortcodes";

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

async function opened(source: string, options: ShortcodeOptions = { dialect: "hugo", declarations: () => [] }) {
  const edits: string[] = [];
  const root = document.createElement("div");
  document.body.append(root);
  const made = await openMarkdown(root, source, "Page text", (next) => edits.push(next), options);
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

describe("the dialect", () => {
  it("locks Quarto's shortcodes, with a link inside, and writes them back as they were", async () => {
    const source = 'Watch {{< video https://example.com/a.mp4 >}} now.\n\n{{% not a quarto shortcode %}}\n';
    const { doc, act, edits } = await opened(source, { dialect: "quarto", declarations: () => [] });
    const names: string[] = [];
    doc().descendants((node) => {
      if (node.type.name.startsWith("shortcode")) names.push(node.attrs.name as string);
    });
    expect(names).toEqual(["video"]);
    act((view) => view.dispatch(view.state.tr.insertText("Do ", 1)));
    expect(edits.at(-1)).toBe('Do Watch {{< video https://example.com/a.mp4 >}} now.\n\n{{% not a quarto shortcode %}}\n');
  });

  it("knows no shortcodes where the site has no dialect", async () => {
    const { doc } = await opened("Watch {{< video a.mp4 >}} now.\n", { declarations: () => [] });
    let found = false;
    doc().descendants((node) => {
      if (node.type.name.startsWith("shortcode")) found = true;
    });
    expect(found).toBe(false);
  });

  it("keeps a Hugo shortcode with a link inside whole", async () => {
    const source = 'See {{< ref "https://example.com/x" >}} and {{< icon bread >}}.\n';
    const { act, edits } = await opened(source);
    act((view) => view.dispatch(view.state.tr.insertText("Do ", 1)));
    expect(edits.at(-1)).toBe('Do See {{< ref "https://example.com/x" >}} and {{< icon bread >}}.\n');
  });
});

const figure: ShortcodeDeclaration = {
  key: "figure",
  title: "Picture",
  fields: [
    { path: "src", type: "image-select", title: "Picture" },
    { path: "caption", type: "string", title: "Caption" },
  ],
};
const notice: ShortcodeDeclaration = {
  key: "notice",
  title: "Notice",
  inner: "markdown",
  fields: [{ path: "type", type: "select", title: "Type", options: [{ value: "info", label: "Information" }, { value: "warning", label: "Warning" }] }],
  positions: { type: 0 },
};

describe("declared shortcodes", () => {
  function declaring(asked: ShortcodeEdit[]): ShortcodeOptions {
    return { dialect: "hugo", declarations: () => [figure, notice], edit: (request) => asked.push(request) };
  }

  it("show their title and first text, and open their form", async () => {
    const asked: ShortcodeEdit[] = [];
    const { root } = await opened('{{< figure src="/a.jpg" class="wide" caption="Our loaf" >}}\n', declaring(asked));
    const button = root.querySelector<HTMLButtonElement>("[data-shortcode='figure'] button")!;
    expect(button.getAttribute("aria-label")).toBe("Edit Picture");
    expect(button.textContent).toBe("PictureOur loaf");
    expect(button.querySelector("img")?.getAttribute("src")).toBe("/a.jpg");
    button.click();
    expect(asked[0]).toMatchObject({ values: { src: "/a.jpg", caption: "Our loaf" }, inner: "" });
  });

  it("write back only the value that changed, as one edit", async () => {
    const asked: ShortcodeEdit[] = [];
    const source = 'Before.\n\n{{< figure src="/a.jpg" class="wide" caption="Our loaf" >}}\n\nAfter.\n';
    const { root, edits } = await opened(source, declaring(asked));
    root.querySelector<HTMLButtonElement>("[data-shortcode='figure'] button")!.click();
    asked[0]!.done({ src: "/a.jpg", caption: "Fresh loaf" }, "");
    expect(edits).toEqual(['Before.\n\n{{< figure src="/a.jpg" class="wide" caption="Fresh loaf" >}}\n\nAfter.\n']);
    expect(root.querySelector("[data-shortcode='figure'] button")?.textContent).toBe("PictureFresh loaf");
  });

  it("change a positional value and the text between", async () => {
    const asked: ShortcodeEdit[] = [];
    const { root, edits } = await opened("{{% notice info %}}\nOld.\n{{% /notice %}}\n", declaring(asked));
    root.querySelector<HTMLButtonElement>("[data-shortcode='notice'] button")!.click();
    expect(asked[0]).toMatchObject({ values: { type: "info" }, inner: "Old." });
    expect(root.querySelector("[data-shortcode='notice'] button")?.textContent).toBe("NoticeInformation");
    asked[0]!.done({ type: "warning" }, "New **text**.");
    expect(edits.at(-1)).toBe("{{% notice warning %}}\nNew **text**.\n{{% /notice %}}\n");
  });

  it("stay locked when the model does not declare them", async () => {
    const { root } = await opened("{{< gallery dir=/a >}}\n", declaring([]));
    expect(root.querySelector("[data-shortcode='gallery'] button")).toBeNull();
    expect(root.querySelector("[data-shortcode='gallery']")?.textContent).toBe("gallery");
  });

  it("are inserted after the block the cursor is in", async () => {
    const { made, edits } = await opened("First.\n\nSecond.\n", declaring([]));
    insertShortcode(made.editor, figure, { src: "/b.jpg", caption: "New" }, "");
    expect(edits.at(-1)).toBe('First.\n\n{{< figure src="/b.jpg" caption="New" >}}\n\nSecond.\n');
  });
});
