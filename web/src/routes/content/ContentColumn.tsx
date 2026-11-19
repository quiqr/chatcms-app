import { ChevronDown, ChevronRight, FileText, Layers, PanelLeftClose, Search } from "lucide-react";
import { useState } from "react";
import { Link, useLocation } from "react-router";

import { Button } from "../../components/ui/button";
import { cn } from "../../lib/cn";
import { recall, remember } from "../../lib/remember";
import { plural, say } from "../../lib/say";
import { readContentMenu, type ContentMenu, type MenuEntry, type MenuGroup } from "../../lib/tenancy";
import { useAsync } from "../../lib/useAsync";
import { useHere } from "../../lib/whoami";
import { openQuickFind } from "./QuickFind";

export function contentAddress(organisation: string, project: string, entry: Pick<MenuEntry, "key" | "kind">) {
  return entry.kind === "collection"
    ? `/o/${organisation}/p/${project}/content/c/${entry.key}`
    : `/o/${organisation}/p/${project}/content/${entry.key}`;
}

export function groupTitle(menu: ContentMenu, group: MenuGroup) {
  if (!menu.fallback) return group.title;
  return group.key === "collections" ? say("content.collections") : say("content.pages");
}

const showing = "chatcms:show-content-menu";

export function showContentMenu() {
  window.dispatchEvent(new Event(showing));
}

export function whenAskedToShow(then: () => void) {
  window.addEventListener(showing, then);
  return () => window.removeEventListener(showing, then);
}

export function currentEntry(menu: ContentMenu, pathname: string, base: string) {
  for (const group of menu.groups) {
    for (const entry of group.items) {
      if (isCurrent(pathname, base, entry)) return { group, entry };
    }
  }
  return null;
}

export function foldedKey(person: string, project: string) {
  return `chatcms.content.groups.${person}.${project}`;
}

export function isCurrent(pathname: string, base: string, entry: MenuEntry) {
  if (entry.kind === "collection") {
    const collection = `${base}/c/${entry.key}`;
    const itemOf = `${base}/${entry.key}/`;
    return pathname === collection || pathname.startsWith(collection + "/") || pathname.startsWith(itemOf);
  }
  return pathname === `${base}/${entry.key}`;
}

export function ContentColumn({
  project,
  onChoose,
  onHide,
  className,
}: {
  project: string;
  onChoose?: () => void;
  onHide?: () => void;
  className?: string;
}) {
  const { organisation, who } = useHere();
  const { pathname } = useLocation();
  const [menu] = useAsync(() => readContentMenu(organisation, project), [organisation, project], `menu:${organisation}/${project}`);
  const key = foldedKey(who.person.id, project);
  const [folded, setFolded] = useState<string[]>(() => recall<string[]>(key, []));
  const base = `/o/${organisation}/p/${project}/content`;

  function toggle(group: string) {
    const next = folded.includes(group) ? folded.filter((one) => one !== group) : [...folded, group];
    setFolded(next);
    remember(key, next);
  }

  return (
    <nav aria-label={say("content-menu.content")} className={cn("flex min-w-0 flex-col gap-3", className)}>
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0">
          <h2 className="text-base font-semibold">{say("content-menu.content")}</h2>
          {menu.state === "ready" && (
            <p className="text-[13px] text-subtle-foreground">
              {plural(menu.value.totals.singles, "content-menu.pages-one", "content-menu.pages-many")}
              {" · "}
              {plural(menu.value.totals.collections, "content-menu.collections-one", "content-menu.collections-many")}
            </p>
          )}
        </div>
        {onHide && (
          <Button variant="ghost" size="icon" aria-label={say("content-menu.hide")} onClick={onHide}>
            <PanelLeftClose aria-hidden="true" className="size-4" />
          </Button>
        )}
      </div>

      <button
        type="button"
        onClick={openQuickFind}
        className="flex min-h-11 items-center gap-2 rounded-md border border-input bg-card px-3 text-left text-[15px] text-subtle-foreground hover:bg-muted"
      >
        <Search aria-hidden="true" className="size-4" />
        <span className="flex-1">{say("quick-find.open")}</span>
        <kbd className="rounded border border-border px-1.5 text-[11px]">{say("quick-find.keys")}</kbd>
      </button>

      {menu.state === "loading" && <p className="text-[15px] text-subtle-foreground">{say("common.looking")}</p>}
      {menu.state === "failed" && <p className="text-[15px]">{menu.message}</p>}
      {menu.state === "ready" && menu.value.groups.length === 0 && (
        <p className="text-[15px] text-subtle-foreground">{say("content-menu.nothing-here")}</p>
      )}

      {menu.state === "ready" &&
        menu.value.groups.map((group) => {
          const isFolded = folded.includes(group.key);
          const title = groupTitle(menu.value, group);
          return (
            <section key={group.key} aria-label={title} className="flex flex-col gap-0.5">
              <button
                type="button"
                aria-expanded={!isFolded}
                onClick={() => toggle(group.key)}
                className="flex min-h-11 items-center gap-1.5 rounded-md px-1 text-left text-xs font-semibold uppercase tracking-wide text-subtle-foreground hover:bg-muted"
              >
                {isFolded ? (
                  <ChevronRight aria-hidden="true" className="size-4" />
                ) : (
                  <ChevronDown aria-hidden="true" className="size-4" />
                )}
                <span className="flex-1 truncate">{title}</span>
                {isFolded && <span className="font-normal normal-case">{group.items.length}</span>}
              </button>

              {!isFolded &&
                group.items.map((entry) => {
                  const current = isCurrent(pathname, base, entry);
                  const Icon = entry.kind === "collection" ? Layers : FileText;
                  return (
                    <Link
                      key={entry.kind + entry.key}
                      to={contentAddress(organisation, project, entry)}
                      title={entry.title}
                      aria-current={current ? "page" : undefined}
                      onClick={onChoose}
                      className={cn(
                        "flex min-h-11 items-center gap-2 rounded-lg px-2.5 text-[15px]",
                        current ? "bg-primary-soft font-semibold text-primary-soft-foreground" : "hover:bg-muted",
                      )}
                    >
                      <Icon aria-hidden="true" className="size-4 shrink-0 text-subtle-foreground" />
                      <span className="min-w-0 flex-1 truncate">{entry.title}</span>
                      {entry.hasChanges && (
                        <span className="size-2 shrink-0 rounded-full bg-warning" aria-label={say("content-menu.waiting")} />
                      )}
                      {entry.kind === "collection" && (
                        <span className="shrink-0 text-[13px] text-subtle-foreground">{entry.count ?? 0}</span>
                      )}
                    </Link>
                  );
                })}
            </section>
          );
        })}
    </nav>
  );
}
