import { Bold, Image, Italic, Link2, List, Plus, Type, type LucideIcon } from "lucide-react";
import { Suspense, lazy, useRef, useState, type FormEvent } from "react";

import { Button } from "../components/ui/button";
import { Dialog, DialogContent } from "../components/ui/dialog";
import { Field } from "../components/ui/field";
import { Popover, PopoverContent, PopoverTrigger } from "../components/ui/popover";
import { Segmented } from "../components/ui/segmented";
import { say, type Key } from "../lib/say";
import type { Control, ShortcodeDeclaration } from "../lib/tenancy";
import { MediaPage, pictureAddress } from "../routes/MediaPage";
import { Draw } from "./Fields";
import type { Formatting, MarkdownHandle } from "./markdown/MarkdownEditor";
import type { ShortcodeEdit } from "./markdown/shortcodes";

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
  dialect,
  shortcodes = [],
}: {
  value: string;
  onChange: (value: string) => void;
  label?: string;
  dialect?: string;
  shortcodes?: ShortcodeDeclaration[];
}) {
  const [showing, setShowing] = useState<"formatted" | "markdown">("formatted");
  const [markdownOnly, setMarkdownOnly] = useState(false);
  const editor = useRef<MarkdownHandle>(null);
  const [editing, setEditing] = useState<ShortcodeEdit | null>(null);
  const declared = useRef(shortcodes);
  declared.current = shortcodes;
  const offered = dialect === "hugo" ? shortcodes : [];
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
          {offered.length > 0 && (
            <InsertMenu
              shortcodes={offered}
              onChoose={(declaration) =>
                setEditing({
                  declaration,
                  values: {},
                  inner: "",
                  done: (values, inner) => editor.current?.insert(declaration, values, inner),
                })
              }
            />
          )}
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
              dialect={dialect}
              shortcodes={{ declarations: () => declared.current, edit: setEditing }}
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

      {editing && <ShortcodeForm key={editing.declaration.key} editing={editing} onClose={() => setEditing(null)} />}
    </div>
  );
}

function InsertMenu({ shortcodes, onChoose }: { shortcodes: ShortcodeDeclaration[]; onChoose: (declaration: ShortcodeDeclaration) => void }) {
  const [open, setOpen] = useState(false);
  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger type="button" aria-label={say("body-editor.insert")} title={say("body-editor.insert")} className={toolClass}>
        <Plus aria-hidden="true" className="size-4" />
      </PopoverTrigger>
      <PopoverContent className="w-60 p-1" onCloseAutoFocus={(event) => event.preventDefault()}>
        <div role="menu" aria-label={say("body-editor.insert")} className="flex flex-col">
          {shortcodes.map((declaration) => (
            <button
              key={declaration.key}
              type="button"
              role="menuitem"
              onClick={() => {
                setOpen(false);
                onChoose(declaration);
              }}
              className="min-h-11 rounded-md px-3 text-left text-[15px] hover:bg-muted"
            >
              {declaration.title}
            </button>
          ))}
        </div>
      </PopoverContent>
    </Popover>
  );
}

function typed(control: Control, value: string | undefined): unknown {
  if (value === undefined || value === "") return undefined;
  if (control.type === "boolean") return value === "true";
  if ((control.type === "number" || control.type === "slider") && !Number.isNaN(Number(value))) return Number(value);
  return value;
}

function asParameter(value: unknown): string {
  if (value === undefined || value === null) return "";
  return String(value);
}

function ShortcodeForm({ editing, onClose }: { editing: ShortcodeEdit; onClose: () => void }) {
  const { declaration } = editing;
  const [values, setValues] = useState<Record<string, unknown>>(() =>
    Object.fromEntries(declaration.fields.map((control) => [control.path, typed(control, editing.values[control.path])])),
  );
  const [inner, setInner] = useState(editing.inner);
  const [choosing, setChoosing] = useState<((path: string) => void) | null>(null);

  function done(event: FormEvent) {
    event.preventDefault();
    const said = Object.fromEntries(Object.entries(values).map(([key, value]) => [key, asParameter(value)]));
    for (const key of Object.keys(editing.values)) {
      if (!(key in said)) said[key] = editing.values[key]!;
    }
    editing.done(said, inner);
    onClose();
  }

  return (
    <Dialog open onOpenChange={(open) => !open && onClose()}>
      <DialogContent
        title={choosing ? say("common.choose-a-picture") : declaration.title}
        className={choosing ? "max-h-[85vh] w-[min(56rem,calc(100vw-2rem))] overflow-y-auto" : undefined}
      >
        {choosing && (
          <div className="flex flex-col gap-4">
            <Button variant="ghost" className="self-start" onClick={() => setChoosing(null)}>
              {say("body-editor.back-to-the-form")}
            </Button>
            <MediaPage
              picking
              onUse={(image) => {
                choosing(pictureAddress(image));
                setChoosing(null);
              }}
            />
          </div>
        )}
        <form onSubmit={done} className={choosing ? "hidden" : "flex flex-col gap-4"}>
          {declaration.fields.map((control) => (
            <Draw
              key={control.path}
              control={{ ...control, value: values[control.path] }}
              onChange={(path, value) => setValues((now) => ({ ...now, [path]: value }))}
              picking={(choose) => setChoosing(() => choose)}
            />
          ))}
          {declaration.inner && (
            <label className="flex flex-col gap-1.5">
              <span className="text-sm font-medium">{say("body-editor.between")}</span>
              <textarea
                value={inner}
                rows={6}
                onChange={(event) => setInner(event.target.value)}
                className="rounded-md border border-input bg-card p-3 font-mono text-[13px]"
              />
            </label>
          )}
          <div className="flex justify-end gap-2">
            <Button type="button" variant="ghost" onClick={onClose}>
              {say("body-editor.cancel")}
            </Button>
            <Button type="submit">{say("body-editor.done")}</Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
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
