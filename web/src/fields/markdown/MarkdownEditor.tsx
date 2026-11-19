import {
  Editor,
  commandsCtx,
  defaultValueCtx,
  editorViewCtx,
  editorViewOptionsCtx,
  remarkCtx,
  rootCtx,
  serializerCtx,
} from "@milkdown/kit/core";
import {
  commonmark,
  insertImageCommand,
  toggleEmphasisCommand,
  toggleLinkCommand,
  toggleStrongCommand,
  wrapInBulletListCommand,
  wrapInHeadingCommand,
} from "@milkdown/kit/preset/commonmark";
import { gfm } from "@milkdown/kit/preset/gfm";
import type { Node } from "@milkdown/kit/prose/model";
import { Plugin, PluginKey } from "@milkdown/kit/prose/state";
import type { NodeViewConstructor } from "@milkdown/kit/prose/view";
import { $prose } from "@milkdown/kit/utils";
import { useEffect, useImperativeHandle, useRef, type Ref } from "react";

import { hugo, quarto } from "./hugo";
import { newShortcode, shortcodeView, type ShortcodeDeclaration, type ShortcodeOptions } from "./shortcodes";
import { blocksOf, keepSource, linesUp, type Range } from "./keep";

export type Formatting =
  | { kind: "heading" }
  | { kind: "bold" }
  | { kind: "italic" }
  | { kind: "list" }
  | { kind: "link"; href: string }
  | { kind: "image"; src: string; alt: string };

export type MarkdownHandle = {
  format: (what: Formatting) => void;
  insert: (declaration: ShortcodeDeclaration, values: Record<string, string>, inner: string) => void;
};

type Opened = { editor: Editor; source: string; blocks: Range[]; first: Node };

export type MarkdownEditorProps = {
  value: string;
  label: string;
  onChange: (value: string) => void;
  onCannot: () => void;
  handle?: Ref<MarkdownHandle>;
  dialect?: string;
  shortcodes?: ShortcodeOptions;
};

export async function openMarkdown(
  root: HTMLElement,
  source: string,
  label: string,
  edited: (next: string) => void,
  shortcodes: ShortcodeOptions = { declarations: () => [] },
): Promise<Opened | null> {
  let opened: Opened | null = null;
  const speaksHugo = shortcodes.dialect === "hugo";
  const made = Editor.make()
    .config((ctx) => {
      ctx.set(rootCtx, root);
      ctx.set(defaultValueCtx, source);
      ctx.update(editorViewOptionsCtx, (options) => ({
        ...options,
        attributes: {
          role: "textbox",
          "aria-multiline": "true",
          "aria-label": label,
          class: "markdown-text",
        },
        nodeViews: (speaksHugo
          ? { shortcode_block: shortcodeView("div", shortcodes), shortcode_inline: shortcodeView("span", shortcodes) }
          : {}) as Record<string, NodeViewConstructor>,
      }));
    })
    .use(commonmark)
    .use(gfm);
  if (speaksHugo) made.use(hugo);
  if (shortcodes.dialect === "quarto") made.use(quarto);
  const editor = await made
    .use(
      $prose(
        (ctx) =>
          new Plugin({
            key: new PluginKey("chatcms-edits"),
            view: () => ({
              update: (view, before) => {
                if (!opened || view.state.doc.eq(before.doc)) return;
                const serializer = ctx.get(serializerCtx);
                const write = (node: Node) => serializer(view.state.schema.topNodeType.create(null, [node]));
                const next = keepSource(opened.source, opened.blocks, opened.first, view.state.doc, write);
                if (next !== null) edited(next);
              },
            }),
          }),
      ),
    )
    .create();

  const lined = editor.action((ctx) => {
    const remark = ctx.get(remarkCtx);
    const tree = remark.runSync(remark.parse(source), source) as unknown as Parameters<typeof blocksOf>[0];
    const blocks = blocksOf(tree);
    const first = ctx.get(editorViewCtx).state.doc;
    return linesUp(blocks, first) ? { blocks, first } : null;
  });
  if (!lined) {
    await editor.destroy();
    return null;
  }
  opened = { editor, source, ...lined };
  return opened;
}

export function format(editor: Editor, what: Formatting) {
  editor.action((ctx) => {
    const commands = ctx.get(commandsCtx);
    switch (what.kind) {
      case "heading":
        commands.call(wrapInHeadingCommand.key, 2);
        break;
      case "bold":
        commands.call(toggleStrongCommand.key);
        break;
      case "italic":
        commands.call(toggleEmphasisCommand.key);
        break;
      case "list":
        commands.call(wrapInBulletListCommand.key);
        break;
      case "link":
        commands.call(toggleLinkCommand.key, { href: what.href });
        break;
      case "image":
        commands.call(insertImageCommand.key, { src: what.src, alt: what.alt });
        break;
    }
    ctx.get(editorViewCtx).focus();
  });
}

export function insertShortcode(editor: Editor, declaration: ShortcodeDeclaration, values: Record<string, string>, inner: string) {
  editor.action((ctx) => {
    const view = ctx.get(editorViewCtx);
    const type = view.state.schema.nodes.shortcode_block;
    if (!type) return;
    const node = type.create({ raw: newShortcode(declaration, values, inner), name: declaration.key });
    const { $from } = view.state.selection;
    const at = $from.depth > 0 ? $from.after(1) : $from.pos;
    view.dispatch(view.state.tr.insert(at, node));
    view.focus();
  });
}

export default function MarkdownEditor({ value, label, onChange, onCannot, handle, dialect, shortcodes }: MarkdownEditorProps) {
  const root = useRef<HTMLDivElement>(null);
  const current = useRef<Opened | null>(null);
  const sent = useRef<string | null>(null);
  const changed = useRef(onChange);
  const cannot = useRef(onCannot);
  const options = useRef(shortcodes);
  const spoken = useRef<string | undefined>(undefined);
  changed.current = onChange;
  cannot.current = onCannot;
  options.current = shortcodes;

  useImperativeHandle(handle, () => ({
    format: (what) => {
      if (current.current) format(current.current.editor, what);
    },
    insert: (declaration, values, inner) => {
      if (current.current) insertShortcode(current.current.editor, declaration, values, inner);
    },
  }));

  useEffect(() => {
    if (value === sent.current && dialect === spoken.current) return;
    const holder = root.current;
    if (!holder) return;
    let gone = false;
    const previous = current.current;
    current.current = null;
    void (async () => {
      await previous?.editor.destroy();
      if (gone) return;
      holder.replaceChildren();
      const opened = await openMarkdown(
        holder,
        value,
        label,
        (next) => {
          if (next === sent.current) return;
          sent.current = next;
          changed.current(next);
        },
        {
          dialect,
          declarations: () => options.current?.declarations() ?? [],
          edit: (request) => options.current?.edit?.(request),
        },
      );
      if (gone) {
        await opened?.editor.destroy();
        return;
      }
      if (!opened) {
        cannot.current();
        return;
      }
      sent.current = value;
      spoken.current = dialect;
      current.current = opened;
    })();
    return () => {
      gone = true;
    };
  }, [value, label, dialect]);

  useEffect(
    () => () => {
      void current.current?.editor.destroy();
      current.current = null;
    },
    [],
  );

  return <div ref={root} className="markdown-editor" />;
}
