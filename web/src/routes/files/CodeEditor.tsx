import { defaultKeymap, history, historyKeymap, indentWithTab } from "@codemirror/commands";
import { bracketMatching, defaultHighlightStyle, indentOnInput, syntaxHighlighting } from "@codemirror/language";
import { highlightSelectionMatches, searchKeymap } from "@codemirror/search";
import { EditorState } from "@codemirror/state";
import { drawSelection, EditorView, highlightActiveLine, highlightActiveLineGutter, keymap, lineNumbers } from "@codemirror/view";
import { useEffect, useRef } from "react";

import { languageFor } from "./languages";

const theme = EditorView.theme({
  "&": { height: "100%", backgroundColor: "var(--card)", color: "var(--foreground)", fontSize: "13px" },
  ".cm-scroller": { fontFamily: "var(--font-mono)", lineHeight: "1.6" },
  ".cm-content": { caretColor: "var(--foreground)" },
  ".cm-gutters": { backgroundColor: "var(--muted)", color: "var(--subtle-foreground)", border: "none" },
  ".cm-activeLine": { backgroundColor: "var(--primary-tint)" },
  ".cm-activeLineGutter": { backgroundColor: "var(--primary-soft)" },
  "&.cm-focused": { outline: "2px solid var(--ring)" },
  ".cm-selectionBackground, &.cm-focused .cm-selectionBackground": { backgroundColor: "var(--primary-soft)" },
  ".cm-go-template": { color: "var(--primary)", backgroundColor: "var(--primary-tint)", borderRadius: "3px" },
  ".cm-panels": { backgroundColor: "var(--popover)", color: "var(--popover-foreground)" },
});

export default function CodeEditor({
  path,
  text,
  label,
  onChange,
  onSave,
}: {
  path: string;
  text: string;
  label: string;
  onChange: (text: string) => void;
  onSave: () => void;
}) {
  const host = useRef<HTMLDivElement>(null);
  const view = useRef<EditorView | null>(null);
  const changed = useRef(onChange);
  const save = useRef(onSave);
  changed.current = onChange;
  save.current = onSave;

  useEffect(() => {
    if (!host.current) return;
    const made = new EditorView({
      parent: host.current,
      state: EditorState.create({
        doc: text,
        extensions: [
          lineNumbers(),
          highlightActiveLineGutter(),
          history(),
          drawSelection(),
          indentOnInput(),
          bracketMatching(),
          highlightActiveLine(),
          highlightSelectionMatches(),
          syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
          keymap.of([
            { key: "Mod-s", preventDefault: true, run: () => (save.current(), true) },
            ...defaultKeymap,
            ...historyKeymap,
            ...searchKeymap,
            indentWithTab,
          ]),
          EditorView.contentAttributes.of({ "aria-label": label }),
          EditorView.updateListener.of((update) => {
            if (update.docChanged) changed.current(update.state.doc.toString());
          }),
          theme,
          ...languageFor(path),
        ],
      }),
    });
    view.current = made;
    return () => {
      made.destroy();
      view.current = null;
    };
  }, [path, text, label]);

  return <div ref={host} data-testid="code-editor" className="h-full min-h-0 overflow-hidden" />;
}
