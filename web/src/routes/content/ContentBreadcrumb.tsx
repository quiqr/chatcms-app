import { ChevronDown, ChevronRight, FileText, Layers, Search } from "lucide-react";
import { useEffect, useRef, useState, type KeyboardEvent } from "react";
import { useLocation, useNavigate, useParams } from "react-router";

import { Popover, PopoverContent, PopoverTrigger } from "../../components/ui/popover";
import { cn } from "../../lib/cn";
import { say } from "../../lib/say";
import { readContentMenu, searchContent, type ContentMenu, type MenuGroup, type SearchResults } from "../../lib/tenancy";
import { useAsync } from "../../lib/useAsync";
import { useHere } from "../../lib/whoami";
import { contentAddress, currentEntry, groupTitle, showContentMenu } from "./ContentColumn";

const pause = 200;
const large = 6;

export function ContentBreadcrumb({ project: given, position }: { project?: string; position?: string }) {
  const { organisation } = useHere();
  const { project: routed = "" } = useParams();
  const project = given ?? routed;
  const { pathname } = useLocation();
  const [open, setOpen] = useState(false);
  const [menu] = useAsync(() => readContentMenu(organisation, project), [organisation, project], `menu:${organisation}/${project}`);
  const ready = menu.state === "ready" ? menu.value : null;
  const base = `/o/${organisation}/p/${project}/content`;
  const here = ready ? currentEntry(ready, pathname, base) : null;
  const last = here ? here.entry.title : say("content-menu.content");

  return (
    <nav aria-label={say("breadcrumb.label")} className="flex min-w-0 items-center gap-1.5 text-[13px] text-subtle-foreground">
      {here && (
        <>
          <span className="shrink-0">{say("content-menu.content")}</span>
          <span aria-hidden="true" className="shrink-0">/</span>
          {ready && groupTitle(ready, here.group) && (
            <>
              <span className="min-w-8 shrink-[10] truncate">{groupTitle(ready, here.group)}</span>
              <span aria-hidden="true" className="shrink-0">/</span>
            </>
          )}
        </>
      )}
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger
          aria-label={say("breadcrumb.choose", { item: last })}
          className="flex min-w-16 items-center gap-1 rounded-md px-1 font-semibold text-foreground hover:bg-muted focus-visible:outline-none focus-visible:ring-[3px] focus-visible:ring-primary-soft"
        >
          <span className="truncate">{last}</span>
          <ChevronDown aria-hidden="true" className="size-3.5 shrink-0" />
        </PopoverTrigger>
        <PopoverContent className="w-80 p-0" aria-label={say("breadcrumb.picker")}>
          {ready && (
            <Picker
              menu={ready}
              project={project}
              base={base}
              current={here?.entry.key}
              currentGroup={here?.group.key}
              onDone={() => setOpen(false)}
            />
          )}
        </PopoverContent>
      </Popover>
      {position && <span className="shrink-0">· {position}</span>}
    </nav>
  );
}

function Picker({
  menu,
  project,
  base,
  current,
  currentGroup,
  onDone,
}: {
  menu: ContentMenu;
  project: string;
  base: string;
  current?: string;
  currentGroup?: string;
  onDone: () => void;
}) {
  const { organisation } = useHere();
  const goTo = useNavigate();
  const list = useRef<HTMLDivElement>(null);
  const [q, setQ] = useState("");
  const [found, setFound] = useState<SearchResults | null>(null);
  const [opened, setOpened] = useState<string[]>(currentGroup ? [currentGroup] : []);
  const [menuHidden] = useState(
    () => document.querySelector("[data-panel='content']")?.getAttribute("data-folded") === "true",
  );

  useEffect(() => {
    if (q.trim() === "") {
      setFound(null);
      return;
    }
    let cancelled = false;
    const timer = setTimeout(() => {
      searchContent(organisation, project, q)
        .then((answer) => !cancelled && setFound(answer))
        .catch(() => !cancelled && setFound({ groups: [] }));
    }, pause);
    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  }, [q, organisation, project]);

  function go(to: string) {
    onDone();
    goTo(to);
  }

  function entries() {
    return Array.from(list.current?.querySelectorAll<HTMLElement>("[data-picker-entry]") ?? []);
  }

  function move(event: KeyboardEvent) {
    if (event.key !== "ArrowDown" && event.key !== "ArrowUp") return;
    const all = entries();
    if (all.length === 0) return;
    event.preventDefault();
    const at = all.indexOf(document.activeElement as HTMLElement);
    const step = event.key === "ArrowDown" ? 1 : -1;
    const next = at === -1 ? (step === 1 ? 0 : all.length - 1) : Math.min(all.length - 1, Math.max(0, at + step));
    all[next]?.focus();
  }

  function isOpen(group: MenuGroup) {
    return group.items.length <= large || opened.includes(group.key);
  }

  function toggle(group: MenuGroup) {
    setOpened((now) => (now.includes(group.key) ? now.filter((key) => key !== group.key) : [...now, group.key]));
  }

  return (
    <div ref={list} onKeyDown={move} className="flex max-h-[min(70vh,520px)] flex-col">
      <div className="flex items-center gap-2 border-b border-border px-3">
        <Search aria-hidden="true" className="size-4 shrink-0 text-subtle-foreground" />
        <input
          type="search"
          autoFocus
          value={q}
          aria-label={say("quick-find.field")}
          placeholder={say("quick-find.placeholder")}
          onChange={(event) => setQ(event.target.value)}
          className="min-h-11 min-w-0 flex-1 bg-transparent text-[15px] outline-none placeholder:text-subtle-foreground"
        />
        <kbd className="shrink-0 rounded border border-border px-1.5 text-xs text-subtle-foreground">{say("quick-find.keys")}</kbd>
      </div>

      <div className="min-h-0 flex-1 overflow-y-auto p-2">
        {found ? (
          found.groups.length === 0 ? (
            <p className="px-2 py-3 text-[15px] text-subtle-foreground">{say("quick-find.nothing")}</p>
          ) : (
            found.groups.map((group) => (
              <section key={group.title} aria-label={group.title} className="flex flex-col gap-0.5 pb-2">
                <h3 className="px-2 pt-1 text-xs font-semibold uppercase tracking-wide text-subtle-foreground">{group.title}</h3>
                {group.items.map((item) => {
                  const Icon = item.kind === "single" ? FileText : Layers;
                  return (
                    <button
                      key={item.key}
                      type="button"
                      data-picker-entry
                      onClick={() => go(`${base}/${item.key}`)}
                      className="flex min-h-11 items-center gap-2 rounded-lg px-2 text-left text-[15px] hover:bg-muted focus-visible:bg-muted focus-visible:outline-none"
                    >
                      <Icon aria-hidden="true" className="size-4 shrink-0 text-subtle-foreground" />
                      <span className="min-w-0 flex-1 truncate">{item.title}</span>
                    </button>
                  );
                })}
              </section>
            ))
          )
        ) : (
          menu.groups.map((group) => {
            const title = groupTitle(menu, group);
            const shown = isOpen(group);
            return (
              <section key={group.key} aria-label={title || undefined} className="flex flex-col gap-0.5 pb-2">
                {title &&
                  (group.items.length > large ? (
                    <button
                      type="button"
                      data-picker-entry
                      aria-expanded={shown}
                      onClick={() => toggle(group)}
                      className="flex min-h-9 items-center gap-1.5 rounded-md px-2 text-left text-xs font-semibold uppercase tracking-wide text-subtle-foreground hover:bg-muted focus-visible:bg-muted focus-visible:outline-none"
                    >
                      {shown ? (
                        <ChevronDown aria-hidden="true" className="size-3.5" />
                      ) : (
                        <ChevronRight aria-hidden="true" className="size-3.5" />
                      )}
                      <span className="flex-1">{title}</span>
                      <span>{group.items.length}</span>
                    </button>
                  ) : (
                    <h3 className="flex px-2 pt-1 text-xs font-semibold uppercase tracking-wide text-subtle-foreground">
                      <span className="flex-1">{title}</span>
                      <span>{group.items.length}</span>
                    </h3>
                  ))}
                {shown &&
                  group.items.map((entry) => {
                    const Icon = entry.kind === "single" ? FileText : Layers;
                    const isHere = entry.key === current;
                    return (
                      <button
                        key={entry.key}
                        type="button"
                        data-picker-entry
                        aria-current={isHere ? "page" : undefined}
                        onClick={() => go(contentAddress(organisation, project, entry))}
                        className={cn(
                          "flex min-h-11 items-center gap-2 rounded-lg px-2 text-left text-[15px] hover:bg-muted focus-visible:bg-muted focus-visible:outline-none",
                          isHere && "bg-primary-soft font-semibold text-primary-soft-foreground",
                        )}
                      >
                        <Icon aria-hidden="true" className="size-4 shrink-0" />
                        <span className="min-w-0 flex-1 truncate">{entry.title}</span>
                        {entry.kind === "collection" && (
                          <span className="shrink-0 text-[13px] text-subtle-foreground">{entry.count ?? 0}</span>
                        )}
                      </button>
                    );
                  })}
              </section>
            );
          })
        )}
      </div>

      {menuHidden && (
        <button
          type="button"
          data-picker-entry
          onClick={() => {
            onDone();
            showContentMenu();
          }}
          className="min-h-11 border-t border-border px-4 text-left text-[13px] font-medium text-primary hover:bg-muted focus-visible:bg-muted focus-visible:outline-none"
        >
          {say("breadcrumb.show-menu")}
        </button>
      )}
    </div>
  );
}
