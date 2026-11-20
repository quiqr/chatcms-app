import { ChevronRight, File, FileCode, FileText, Folder, FolderOpen, Image, Link2 } from "lucide-react";
import { useMemo, useRef, useState, type KeyboardEvent } from "react";

import { cn } from "../../lib/cn";
import type { FileEntry } from "../../lib/tenancy";
import { extensionOf } from "./languages";

type Node = FileEntry & { name: string; parent: string; depth: number };

const pictures = new Set(["png", "jpg", "jpeg", "gif", "webp", "svg", "ico", "avif"]);
const code = new Set(["html", "htm", "css", "scss", "js", "mjs", "ts", "tsx", "json", "toml", "yaml", "yml", "xml"]);
const prose = new Set(["md", "markdown", "qmd", "txt"]);

function iconOf(node: Node, open: boolean) {
  if (node.dir) return open ? FolderOpen : Folder;
  if (node.link) return Link2;
  const extension = extensionOf(node.path);
  if (pictures.has(extension)) return Image;
  if (code.has(extension)) return FileCode;
  if (prose.has(extension)) return FileText;
  return File;
}

export function arrange(entries: FileEntry[]) {
  const children = new Map<string, Node[]>();
  for (const entry of entries) {
    const cut = entry.path.lastIndexOf("/");
    const parent = cut < 0 ? "" : entry.path.slice(0, cut);
    const node: Node = { ...entry, name: entry.path.slice(cut + 1), parent, depth: entry.path.split("/").length };
    children.set(parent, [...(children.get(parent) ?? []), node]);
  }
  for (const list of children.values()) {
    list.sort((a, b) => Number(Boolean(b.dir)) - Number(Boolean(a.dir)) || a.name.localeCompare(b.name));
  }
  return children;
}

function visibleOf(children: Map<string, Node[]>, open: Set<string>) {
  const shown: Node[] = [];
  const walk = (parent: string) => {
    for (const node of children.get(parent) ?? []) {
      shown.push(node);
      if (node.dir && open.has(node.path)) walk(node.path);
    }
  };
  walk("");
  return shown;
}

export function FileTree({
  entries,
  label,
  chosen,
  onOpen,
}: {
  entries: FileEntry[];
  label: string;
  chosen: string;
  onOpen: (path: string) => void;
}) {
  const children = useMemo(() => arrange(entries), [entries]);
  const [open, setOpen] = useState<Set<string>>(() => {
    const parents = new Set<string>();
    let at = chosen;
    while (at.includes("/")) {
      at = at.slice(0, at.lastIndexOf("/"));
      parents.add(at);
    }
    return parents;
  });
  const shown = useMemo(() => visibleOf(children, open), [children, open]);
  const [focused, setFocused] = useState(chosen || shown[0]?.path || "");
  const items = useRef(new Map<string, HTMLLIElement>());

  const current = shown.find((node) => node.path === focused) ? focused : shown[0]?.path ?? "";

  const focus = (path: string) => {
    setFocused(path);
    items.current.get(path)?.focus();
  };
  const toggle = (path: string, to?: boolean) =>
    setOpen((before) => {
      const next = new Set(before);
      const opening = to ?? !next.has(path);
      if (opening) next.add(path);
      else next.delete(path);
      return next;
    });
  const choose = (node: Node) => {
    if (node.dir) toggle(node.path);
    else onOpen(node.path);
  };

  const keyed = (event: KeyboardEvent<HTMLUListElement>) => {
    const at = shown.findIndex((node) => node.path === current);
    const node = shown[at];
    if (!node) return;
    const move = (index: number) => {
      const target = shown[Math.max(0, Math.min(shown.length - 1, index))];
      if (target) focus(target.path);
    };
    switch (event.key) {
      case "ArrowDown":
        move(at + 1);
        break;
      case "ArrowUp":
        move(at - 1);
        break;
      case "Home":
        move(0);
        break;
      case "End":
        move(shown.length - 1);
        break;
      case "ArrowRight":
        if (node.dir && !open.has(node.path)) toggle(node.path, true);
        else if (node.dir) move(at + 1);
        break;
      case "ArrowLeft":
        if (node.dir && open.has(node.path)) toggle(node.path, false);
        else if (node.parent) focus(node.parent);
        break;
      case "Enter":
      case " ":
        choose(node);
        break;
      default:
        return;
    }
    event.preventDefault();
  };

  return (
    <ul role="tree" aria-label={label} onKeyDown={keyed} className="flex flex-col py-1 text-[13px]">
      {shown.map((node) => {
        const opened = Boolean(node.dir) && open.has(node.path);
        const Icon = iconOf(node, opened);
        const siblings = children.get(node.parent) ?? [];
        return (
          <li
            key={node.path}
            ref={(element) => {
              if (element) items.current.set(node.path, element);
              else items.current.delete(node.path);
            }}
            role="treeitem"
            aria-level={node.depth}
            aria-setsize={siblings.length}
            aria-posinset={siblings.indexOf(node) + 1}
            aria-expanded={node.dir ? opened : undefined}
            aria-selected={node.path === chosen}
            tabIndex={node.path === current ? 0 : -1}
            onClick={() => {
              setFocused(node.path);
              choose(node);
            }}
            onFocus={() => setFocused(node.path)}
            className={cn(
              "flex min-h-11 cursor-pointer items-center gap-1.5 pr-3 outline-none hover:bg-muted focus-visible:bg-primary-tint focus-visible:ring-2 focus-visible:ring-ring sm:min-h-9",
              node.path === chosen && "bg-primary-soft text-primary-soft-foreground",
            )}
            style={{ paddingLeft: `${(node.depth - 1) * 16 + 8}px` }}
          >
            <ChevronRight
              aria-hidden="true"
              className={cn("size-3.5 shrink-0 text-subtle-foreground transition-transform", !node.dir && "invisible", opened && "rotate-90")}
            />
            <Icon aria-hidden="true" className="size-4 shrink-0 text-subtle-foreground" />
            <span className="truncate">{node.name}</span>
          </li>
        );
      })}
    </ul>
  );
}
