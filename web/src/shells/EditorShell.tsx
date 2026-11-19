import { ArrowLeft } from "lucide-react";
import { useEffect, useState, type ReactNode } from "react";
import { NavLink } from "react-router";

import { ContentPanel } from "../routes/content/ContentPanel";
import { QuickFind } from "../routes/content/QuickFind";
import { SitePanel } from "./SitePanel";

import { Button } from "../components/ui/button";
import { Segmented } from "../components/ui/segmented";
import { AssistantSlot } from "./ProjectShell";
import { TopBar } from "./TopBar";
import { arrange, useEditorRoom } from "./useEditorRoom";
import { useHere } from "../lib/whoami";
import { SyncIndicator } from "./SyncIndicator";
import type { Project } from "../lib/tenancy";
import { say } from "../lib/say";

export type EditorLayout = "form" | "split" | "preview";

export type EditorShellProps = {
  project: Project;
  viewer: { name: string };
  title: string;
  context?: ReactNode;
  backTo?: string;
  backLabel?: string;
  stepper?: ReactNode;
  layout: EditorLayout;
  onLayoutChange: (layout: EditorLayout) => void;
  form: ReactNode;
  preview: ReactNode;
  assistant?: (close: () => void) => ReactNode;
  openAssistant?: number;
};

export function EditorShell({
  project,
  viewer,
  title,
  context,
  backTo,
  backLabel,
  stepper,
  layout,
  onLayoutChange,
  form,
  preview,
  assistant,
  openAssistant = 0,
}: EditorShellProps) {
  const { organisation } = useHere();
  const [talking, setTalking] = useState(false);
  const measured = useEditorRoom();
  const [opened, setOpened] = useState(false);
  useEffect(() => {
    if (openAssistant > 0) {
      setOpened(true);
      setTalking(true);
    }
  }, [openAssistant]);
  const room = { ...arrange(measured.width, layout, talking), ref: measured.ref };
  return (
    <div data-testid="editor-shell" className="flex h-dvh overflow-hidden">
      <SitePanel project={project} role={project.yourRole} />
      <ContentPanel project={project.slug} />
      <QuickFind project={project.slug} />

      <div className="flex min-h-0 min-w-0 shrink grow-[999] basis-[560px] flex-col">
        <TopBar viewer={viewer} className="shrink-0">
          <Button
            variant="ghost"
            size="iconCompact"
            asChild
            aria-label={backLabel ?? say("editor-shell.back-to-content")}
          >
            <NavLink to={backTo ?? `/o/${organisation}/p/${project.slug}/content`}>
              <ArrowLeft aria-hidden="true" className="size-4" />
            </NavLink>
          </Button>
          <div className="flex min-w-0 flex-1 flex-col">
            {context && <p className="truncate text-[13px] text-subtle-foreground">{context}</p>}
            <h1 className="truncate text-lg font-semibold">{title}</h1>
          </div>
          {stepper}
          <SyncIndicator />
          <Segmented
            label={say("editor-shell.how-to-show-this-page")}
            value={room.shown}
            onValueChange={(next) => onLayoutChange(next as EditorLayout)}
            options={[
              { value: "form", label: "Form" },
              ...(room.splitFits ? [{ value: "split", label: say("editor-shell.side-by-side") }] : []),
              { value: "preview", label: "Preview" },
            ]}
          />
          <AssistantSlot
            open={talking}
            onToggle={
              assistant
                ? () => {
                    setOpened(true);
                    setTalking((now) => !now);
                  }
                : undefined
            }
          />
          <Button asChild count={project.unpublished}>
            <NavLink to={`/o/${organisation}/p/${project.slug}/publish`}>{say("common.publish")}</NavLink>
          </Button>
        </TopBar>

        <main ref={room.ref} data-testid="editor-slots" className="flex min-h-0 min-w-0 flex-1">
          {room.shown !== "preview" && room.assistant !== "alone" && (
            <section
              data-testid="editor-slot-form"
              data-scrolls="form"
              aria-label={say("editor-shell.page-fields")}
              className="min-h-0 min-w-0 flex-1 basis-0 overflow-y-auto overscroll-contain p-6"
            >
              <div className="mx-auto w-full max-w-[760px]">{form}</div>
            </section>
          )}

          {room.shown !== "form" && room.assistant !== "alone" && (
            <section
              data-testid="editor-slot-preview"
              aria-label={say("editor-shell.preview")}
              className="flex min-h-0 min-w-0 flex-1 basis-0 flex-col border-l border-border bg-background p-6"
            >
              {preview}
            </section>
          )}

          <aside
            data-testid="editor-slot-assistant"
            data-room={room.assistant}
            aria-label={say("editor-shell.assistant")}
            hidden={!talking}
            className={
              "min-h-0 border-l border-border bg-card " +
              (room.assistant === "column" ? "w-[340px] shrink-0" : "min-w-0 flex-1 basis-0")
            }
          >
            {opened && assistant?.(() => setTalking(false))}
          </aside>
        </main>
      </div>
    </div>
  );
}
