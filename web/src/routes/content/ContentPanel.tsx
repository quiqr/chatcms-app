import { PanelLeftOpen } from "lucide-react";
import { useState } from "react";

import { Button } from "../../components/ui/button";
import { Tooltip } from "../../components/ui/tooltip";
import { say } from "../../lib/say";
import { useFold } from "../../lib/useFold";
import { useHere } from "../../lib/whoami";
import { useIsWide } from "../../shells/useIsWide";
import { ContentColumn } from "./ContentColumn";

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

  return (
    <>
      <aside
        data-panel="content"
        data-folded={asStrip ? "true" : "false"}
        className={
          asStrip
            ? "flex shrink-0 basis-12 flex-col items-center border-r border-border bg-card py-3"
            : `shrink-0 basis-64 overflow-y-auto border-r border-border bg-card p-4 ${className ?? ""}`
        }
      >
        {asStrip ? (
          <Tooltip side="right" content={say("content-menu.show")}>
            <Button variant="ghost" size="icon" aria-label={say("content-menu.show")} onClick={unfold}>
              <PanelLeftOpen aria-hidden="true" className="size-5" />
            </Button>
          </Tooltip>
        ) : (
          <ContentColumn project={project} onHide={() => setFolded(true)} />
        )}
      </aside>
      {drawer && <ContentDrawer project={project} onClose={() => setDrawer(false)} />}
    </>
  );
}
