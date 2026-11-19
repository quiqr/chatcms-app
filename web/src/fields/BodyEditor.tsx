import { Bold, Image, Italic, Link2, List, Type, type LucideIcon } from "lucide-react";
import { Suspense, lazy, useRef, useState, type FormEvent } from "react";

import { Button } from "../components/ui/button";
import { Field } from "../components/ui/field";
import { Popover, PopoverContent, PopoverTrigger } from "../components/ui/popover";
import { Segmented } from "../components/ui/segmented";
import { say, type Key } from "../lib/say";
import type { Formatting, MarkdownHandle } from "./markdown/MarkdownEditor";

const MarkdownEditor = lazy(() => import("./markdown/MarkdownEditor"));

type Plain = { kind: "heading" | "bold" | "italic" | "list"; label: Key; icon: LucideIcon };

const plain: Plain[] = [
  { kind: "heading", label: "body-editor.heading", icon: Type },
  { kind: "bold", label: "body-editor.bold", icon: Bold },
  { kind: "italic", label: "body-editor.italic", icon: Italic },
];

const toolClass = "flex size-11 items-center justify-center rounded-md text-subtle-foreground hover:bg-muted";

export function BodyEditor({
  value,
  onChange,
  label = say("page-editor.page-text"),
}: {
  value: string;
  onChange: (value: string) => void;
  label?: string;
}) {
  const [showing, setShowing] = useState<"formatted" | "markdown">("formatted");
  const [markdownOnly, setMarkdownOnly] = useState(false);
  const editor = useRef<MarkdownHandle>(null);
  const formatted = showing === "formatted" && !markdownOnly;

  function apply(what: Formatting) {
    editor.current?.format(what);
  }

  return (
    <div className="flex flex-col gap-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span className="text-sm font-medium">{label}</span>
        {!markdownOnly && (
          <Segmented
            label={say("body-editor.how-to-edit-the-text")}
            value={showing}
            options={[
              { value: "formatted", label: say("body-editor.formatted") },
              { value: "markdown", label: say("body-editor.markdown") },
            ]}
            onValueChange={(next) => setShowing(next as "formatted" | "markdown")}
          />
        )}
      </div>

      {markdownOnly && <p className="text-[13px] text-subtle-foreground">{say("body-editor.markdown-only")}</p>}

      {formatted && (
        <div className="flex flex-wrap gap-1" role="toolbar" aria-label={say("body-editor.text-formatting")}>
          {plain.map((tool) => (
            <button key={tool.kind} type="button" aria-label={say(tool.label)} title={say(tool.label)} onMouseDown={(event) => event.preventDefault()} onClick={() => apply({ kind: tool.kind })} className={toolClass}>
              <tool.icon aria-hidden="true" className="size-4" />
            </button>
          ))}
          <AskAddress
            icon={Link2}
            label="body-editor.link"
            fields={[{ name: "href", label: "body-editor.link-address", placeholder: "https://" }]}
            onDone={(said) => apply({ kind: "link", href: said.href ?? "" })}
          />
          <button type="button" aria-label={say("body-editor.list")} title={say("body-editor.list")} onMouseDown={(event) => event.preventDefault()} onClick={() => apply({ kind: "list" })} className={toolClass}>
            <List aria-hidden="true" className="size-4" />
          </button>
          <AskAddress
            icon={Image}
            label="body-editor.image"
            fields={[
              { name: "src", label: "body-editor.image-address", placeholder: "/images/" },
              { name: "alt", label: "body-editor.image-description" },
            ]}
            onDone={(said) => apply({ kind: "image", src: said.src ?? "", alt: said.alt ?? "" })}
          />
        </div>
      )}

      {formatted ? (
        <Suspense fallback={<p className="min-h-64 whitespace-pre-wrap rounded-md border border-input bg-card p-3 text-[15px] text-subtle-foreground">{value}</p>}>
          <div className="min-h-64 rounded-md border border-input bg-card p-3 text-base focus-within:ring-[3px] focus-within:ring-primary-soft sm:text-[15px]">
            <MarkdownEditor
              handle={editor}
              value={value}
              label={label}
              onChange={onChange}
              onCannot={() => {
                setMarkdownOnly(true);
                setShowing("markdown");
              }}
            />
          </div>
        </Suspense>
      ) : (
        <textarea
          aria-label={label}
          value={value}
          rows={14}
          spellCheck={false}
          onChange={(event) => onChange(event.target.value)}
          className="min-h-64 rounded-md border border-input bg-card p-3 font-mono text-[13px]"
        />
      )}
    </div>
  );
}

type Asked = { name: string; label: Key; placeholder?: string };

function AskAddress({
  icon: Icon,
  label,
  fields,
  onDone,
}: {
  icon: LucideIcon;
  label: Key;
  fields: Asked[];
  onDone: (said: Record<string, string>) => void;
}) {
  const [open, setOpen] = useState(false);
  const [said, setSaid] = useState<Record<string, string>>({});

  function done(event: FormEvent) {
    event.preventDefault();
    if (!(said[fields[0]!.name] ?? "").trim()) return;
    setOpen(false);
    onDone(said);
    setSaid({});
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger type="button" aria-label={say(label)} title={say(label)} className={toolClass}>
        <Icon aria-hidden="true" className="size-4" />
      </PopoverTrigger>
      <PopoverContent className="w-72 p-3" onCloseAutoFocus={(event) => event.preventDefault()}>
        <form onSubmit={done} className="flex flex-col gap-3">
          {fields.map((field, at) => (
            <Field
              key={field.name}
              label={say(field.label)}
              autoFocus={at === 0}
              placeholder={field.placeholder}
              value={said[field.name] ?? ""}
              onChange={(event) => setSaid((now) => ({ ...now, [field.name]: event.target.value }))}
            />
          ))}
          <Button type="submit" size="compact" className="self-end">
            {say("body-editor.add")}
          </Button>
        </form>
      </PopoverContent>
    </Popover>
  );
}
