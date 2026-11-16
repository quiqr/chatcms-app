import { ArrowLeft, FileText, History, Image, Upload } from "lucide-react";
import { useState, type ReactNode } from "react";
import { NavLink } from "react-router";

import { recall, remember } from "../lib/remember";
import { ContentColumn } from "../routes/content/ContentColumn";
import { ContentDrawer } from "../routes/content/ContentScreen";
import { QuickFind } from "../routes/content/QuickFind";
import { useIsWide } from "./useIsWide";

import { Button } from "../components/ui/button";
import { Segmented } from "../components/ui/segmented";
import { AssistantSlot } from "./ProjectShell";
import { TopBar } from "./TopBar";
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
};

const railEntries = [
  { key: "media", label: "Media", icon: Image },
  { key: "publish", label: "Publish", icon: Upload },
  { key: "history", label: "History", icon: History },
];

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
}: EditorShellProps) {
  const { organisation, who } = useHere();
  const wide = useIsWide();
  const foldKey = `chatcms.content.folded.${who.person.id}`;
  const [folded, setFolded] = useState(() => recall(foldKey, false));
  const [drawer, setDrawer] = useState(false);

  function fold(next: boolean) {
    setFolded(next);
    remember(foldKey, next);
  }

  function showColumn() {
    if (wide) fold(false);
    else setDrawer(true);
  }

  return (
    <div data-testid="editor-shell" className="flex min-h-dvh flex-wrap">
      <nav
        aria-label={say("editor-shell.project-sections")}
        data-testid="editor-rail"
        className="hidden w-16 shrink-0 flex-col items-center gap-2 border-r border-border bg-card py-3 sm:flex"
      >
        <button
          type="button"
          aria-label={say("content-menu.show")}
          aria-pressed={wide && !folded}
          onClick={showColumn}
          className="flex size-11 items-center justify-center rounded-md text-muted-foreground hover:bg-muted"
        >
          <FileText aria-hidden="true" className="size-5" />
        </button>
        {railEntries.map((entry) => (
          <NavLink
            key={entry.key}
            to={`/o/${organisation}/p/${project.slug}/${entry.key}`}
            aria-label={entry.label}
            className="flex size-11 items-center justify-center rounded-md text-muted-foreground hover:bg-muted"
          >
            <entry.icon aria-hidden="true" className="size-5" />
          </NavLink>
        ))}
      </nav>

      {wide && !folded && (
        <aside
          data-testid="editor-content-column"
          className="w-64 shrink-0 overflow-y-auto border-r border-border bg-card p-4"
        >
          <ContentColumn project={project.slug} onHide={() => fold(true)} />
        </aside>
      )}
      {drawer && <ContentDrawer project={project.slug} onClose={() => setDrawer(false)} />}
      <QuickFind project={project.slug} />

      <div className="flex min-w-0 shrink grow-[999] basis-[560px] flex-col">
        <TopBar viewer={viewer}>
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
            value={layout}
            onValueChange={(next) => onLayoutChange(next as EditorLayout)}
            options={[
              { value: "form", label: "Form" },
              { value: "split", label: say("editor-shell.side-by-side") },
              { value: "preview", label: "Preview" },
            ]}
            className="hidden md:inline-flex"
          />
          <AssistantSlot />
          <Button asChild count={project.unpublished}>
            <NavLink to={`/o/${organisation}/p/${project.slug}/publish`}>{say("common.publish")}</NavLink>
          </Button>
        </TopBar>

        <main data-testid="editor-slots" className="flex min-w-0 flex-1 flex-wrap items-stretch">
          {layout !== "preview" && (
            <section
              data-testid="editor-slot-form"
              aria-label={say("editor-shell.page-fields")}
              className="min-w-0 shrink grow basis-[420px] p-6"
            >
              <div className="mx-auto w-full max-w-[760px]">{form}</div>
            </section>
          )}

          {layout !== "form" && (
            <section
              data-testid="editor-slot-preview"
              aria-label={say("editor-shell.preview")}
              className="min-w-0 shrink grow basis-[420px] border-l border-border bg-background p-6"
            >
              {preview}
            </section>
          )}

          <aside
            data-testid="editor-slot-assistant"
            aria-label={say("editor-shell.assistant")}
            hidden
            className="w-[340px] shrink-0 border-l border-border bg-card"
          />
        </main>
      </div>
    </div>
  );
}
