import { MessageSquare } from "lucide-react";
import type { ReactNode } from "react";
import { NavLink } from "react-router";

import { Button } from "../components/ui/button";
import { useHere } from "../lib/whoami";
import { named } from "./roles";
import { SitePanel } from "./SitePanel";
import { SyncIndicator, type SyncState } from "./SyncIndicator";
import { TopBar } from "./TopBar";
import type { Project } from "../lib/tenancy";
import { say } from "../lib/say";
import { QuickFind } from "../routes/content/QuickFind";

export type ProjectShellProps = {
  project: Project;
  viewer: { name: string; role: string };
  title: ReactNode;
  actions?: ReactNode;
  syncState?: SyncState;
  beside?: ReactNode;
  children: ReactNode;
};

export function ProjectShell({
  project,
  viewer,
  title,
  actions,
  syncState = "saved",
  beside,
  children,
}: ProjectShellProps) {
  const { organisation } = useHere();
  return (
    <div data-testid="project-shell" className="hidden h-dvh overflow-hidden sm:flex">
      <SitePanel project={project} role={viewer.role} />
      {beside}

      <div className="flex min-h-0 min-w-0 shrink grow-[999] basis-[560px] flex-col">
        <TopBar viewer={{ name: viewer.name, role: named(viewer.role) }} className="shrink-0 px-8">
          <h1 className="flex-1 text-lg font-semibold">{title}</h1>
          <SyncIndicator state={syncState} />
          {actions}
          <Button asChild count={project.unpublished}>
            <NavLink to={`/o/${organisation}/p/${project.slug}/publish`}>{say("common.publish")}</NavLink>
          </Button>
        </TopBar>

        <main data-scrolls="main" className="min-h-0 min-w-0 flex-1 overflow-y-auto p-8">{children}</main>
        <QuickFind project={project.slug} />
      </div>
    </div>
  );
}

export function AssistantSlot({ open, onToggle }: { open?: boolean; onToggle?: () => void }) {
  if (!onToggle) {
    return (
      <Button variant="ghost" size="icon" aria-label={say("project-shell.assistant-coming-later")} disabled>
        <MessageSquare aria-hidden="true" className="size-5" />
      </Button>
    );
  }
  return (
    <Button
      variant={open ? "outline" : "ghost"}
      size="icon"
      aria-label={say("assistant.open")}
      aria-pressed={open}
      onClick={onToggle}
    >
      <MessageSquare aria-hidden="true" className="size-5" />
    </Button>
  );
}
