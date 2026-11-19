import { PanelLeftOpen } from "lucide-react";
import { useEffect, useState } from "react";
import { useLocation } from "react-router";

import { Tooltip } from "../../components/ui/tooltip";
import { say } from "../../lib/say";
import { readContentMenu } from "../../lib/tenancy";
import { useAsync } from "../../lib/useAsync";
import { useFold } from "../../lib/useFold";
import { useHere } from "../../lib/whoami";
import { useIsWide } from "../../shells/useIsWide";
import { ContentColumn, currentEntry, whenAskedToShow } from "./ContentColumn";

export function ContentDrawer({ project, onClose }: { project: string; onClose: () => void }) {
  return (
    <div className="fixed inset-0 z-40 flex" role="dialog" aria-modal="true" aria-label={say("content-menu.open")}>
      <div className="w-[min(20rem,85vw)] overflow-y-auto border-r border-border bg-card p-4 shadow-lg">
        <ContentColumn project={project} onChoose={onClose} onHide={onClose} />
      </div>
      <button
        type="button"
        aria-label={say("content-menu.hide")}
        className="flex-1 bg-foreground/30"
        onClick={onClose}
      />
    </div>
  );
}

export function oldFoldKey(person: string) {
  return `chatcms.content.folded.${person}`;
}

export function ContentPanel({ project, className }: { project: string; className?: string }) {
  const { who } = useHere();
  const wide = useIsWide();
  const [folded, setFolded] = useFold("content", who.person.id, oldFoldKey(who.person.id));
  const [drawer, setDrawer] = useState(false);
  const asStrip = folded || !wide;

  function unfold() {
    if (wide) setFolded(false);
    else setDrawer(true);
  }

  useEffect(() => whenAskedToShow(unfold));

  return (
    <>
      <aside
        data-panel="content"
        data-folded={asStrip ? "true" : "false"}
        className={
          asStrip
            ? "flex w-12 shrink-0 basis-12 border-r border-border bg-card"
            : `shrink-0 basis-64 overflow-y-auto border-r border-border bg-card p-4 ${className ?? ""}`
        }
      >
        {asStrip ? (
          <Strip project={project} onUnfold={unfold} />
        ) : (
          <ContentColumn project={project} onHide={() => setFolded(true)} />
        )}
      </aside>
      {drawer && <ContentDrawer project={project} onClose={() => setDrawer(false)} />}
    </>
  );
}

function Strip({ project, onUnfold }: { project: string; onUnfold: () => void }) {
  const { organisation } = useHere();
  const { pathname } = useLocation();
  const [menu] = useAsync(() => readContentMenu(organisation, project), [organisation, project], `menu:${organisation}/${project}`);
  const ready = menu.state === "ready" ? menu.value : null;
  const here = ready ? currentEntry(ready, pathname, `/o/${organisation}/p/${project}/content`) : null;
  const waiting = ready ? ready.groups.some((group) => group.items.some((entry) => entry.hasChanges)) : false;
  const label = here ? say("content-menu.show-named", { item: here.entry.title }) : say("content-menu.show");

  return (
    <Tooltip side="right" content={here ? here.entry.title : say("content-menu.show")}>
      <button
        type="button"
        data-strip="content"
        aria-label={label}
        onClick={onUnfold}
        className="flex h-full w-full flex-col items-center gap-4 pt-3.5 hover:bg-muted focus-visible:outline-none focus-visible:ring-[3px] focus-visible:ring-inset focus-visible:ring-primary-soft"
      >
        <span className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-muted">
          <PanelLeftOpen aria-hidden="true" className="size-5" />
        </span>
        <span
          aria-hidden="true"
          className="flex min-h-0 items-center gap-2.5 [writing-mode:vertical-rl]"
          style={{ maxHeight: "calc(100% - 120px)" }}
        >
          <span className="shrink-0 text-xs font-semibold uppercase tracking-[0.08em] text-muted-foreground">
            {say("content-menu.content")}
          </span>
          {here && (
            <>
              <span className="shrink-0 text-subtle-foreground">·</span>
              <span data-strip-title className="min-h-0 truncate text-[13px] font-semibold text-primary-soft-foreground">
                {here.entry.title}
              </span>
              {here.entry.kind === "collection" && (
                <span className="shrink-0 text-[13px] text-subtle-foreground">{here.entry.count ?? 0}</span>
              )}
            </>
          )}
        </span>
        {waiting && (
          <span
            data-strip-dot
            className="size-[7px] shrink-0 rounded-full bg-warning"
            role="img"
            aria-label={say("content-menu.waiting")}
          />
        )}
      </button>
    </Tooltip>
  );
}
