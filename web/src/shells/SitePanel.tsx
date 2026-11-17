import { ChevronsUpDown, PanelLeftClose, PanelLeftOpen } from "lucide-react";
import { NavLink } from "react-router";

import { Tile } from "../components/ui/avatar";
import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Tooltip } from "../components/ui/tooltip";
import { cn } from "../lib/cn";
import { say } from "../lib/say";
import type { Project } from "../lib/tenancy";
import { useFold } from "../lib/useFold";
import { useHere } from "../lib/whoami";
import { setupNavigation, siteNavigation, type NavEntry } from "./navigation";
import { ProjectSwitcher } from "./ProjectSwitcher";
import { canSeeSetup } from "./roles";

function NavGroup({ label, entries, unpublished }: { label: string; entries: NavEntry[]; unpublished: number }) {
  return (
    <div role="group" aria-label={label} className="flex flex-col gap-0.5">
      {entries.map((entry) => (
        <NavLink
          key={entry.key}
          to={entry.to}
          className={({ isActive }) =>
            cn(
              "flex h-11 items-center gap-2.5 rounded-lg px-3 text-[15px] text-muted-foreground",
              isActive && "bg-muted font-medium text-foreground",
            )
          }
        >
          <entry.icon aria-hidden="true" className="size-[18px]" />
          <span className="flex-1">{entry.label}</span>
          {entry.key === "publish" && unpublished > 0 && <Badge tone="changed">{unpublished}</Badge>}
          {entry.soon && <Badge tone="neutral">{say("project-shell.soon")}</Badge>}
        </NavLink>
      ))}
    </div>
  );
}

function RailGroup({ label, entries, unpublished }: { label: string; entries: NavEntry[]; unpublished: number }) {
  return (
    <div role="group" aria-label={label} className="flex flex-col items-center gap-1">
      {entries.map((entry) => (
        <Tooltip key={entry.key} side="right" content={entry.label}>
          <NavLink
            to={entry.to}
            aria-label={entry.label}
            className={({ isActive }) =>
              cn(
                "relative flex size-11 items-center justify-center rounded-md text-muted-foreground hover:bg-muted",
                isActive && "bg-muted text-foreground",
              )
            }
          >
            <entry.icon aria-hidden="true" className="size-5" />
            {entry.key === "publish" && unpublished > 0 && (
              <span className="absolute right-1 top-1 size-2 rounded-full bg-warning" aria-hidden="true" />
            )}
          </NavLink>
        </Tooltip>
      ))}
    </div>
  );
}

export function SitePanel({ project, role }: { project: Project; role: string }) {
  const { organisation, who } = useHere();
  const [folded, setFolded] = useFold("site", who.person.id);
  const site = siteNavigation(organisation, project.slug);
  const setup = canSeeSetup(role) ? setupNavigation(organisation, project.slug) : [];

  if (folded) {
    return (
      <aside
        aria-label={say("project-shell.project-navigation")}
        data-panel="site"
        data-folded="true"
        className="flex shrink-0 grow-0 basis-16 flex-col items-center gap-3 border-r border-border bg-card py-3"
      >
        <Tooltip side="right" content={say("site-panel.show")}>
          <Button variant="ghost" size="icon" aria-label={say("site-panel.show")} onClick={() => setFolded(false)}>
            <PanelLeftOpen aria-hidden="true" className="size-5" />
          </Button>
        </Tooltip>
        <ProjectSwitcher project={project}>
          <button type="button" className="rounded-lg p-1 hover:bg-muted">
            <Tile name={project.name} />
          </button>
        </ProjectSwitcher>
        <RailGroup label={say("project-shell.site")} entries={site} unpublished={project.unpublished} />
        {setup.length > 0 && (
          <>
            <span className="h-px w-8 bg-border" aria-hidden="true" />
            <RailGroup label={say("project-shell.setup")} entries={setup} unpublished={0} />
          </>
        )}
      </aside>
    );
  }

  return (
    <aside
      aria-label={say("project-shell.project-navigation")}
      data-panel="site"
      data-folded="false"
      className="flex shrink grow basis-60 flex-col gap-4 border-r border-border bg-card p-3"
    >
      <div className="flex items-center gap-2 pl-2 pt-1">
        <img src="/quiqr-logo.svg" alt={say("common.quiqr")} className="h-6 w-auto" />
        <span className="flex-1 text-[15px] text-muted-foreground">{say("common.chatcms")}</span>
        <Button variant="ghost" size="icon" aria-label={say("site-panel.hide")} onClick={() => setFolded(true)}>
          <PanelLeftClose aria-hidden="true" className="size-4" />
        </Button>
      </div>

      <ProjectSwitcher project={project}>
        <button
          type="button"
          className="flex min-h-14 items-center gap-2.5 rounded-xl border border-border bg-card px-2.5 py-2 text-left hover:bg-muted"
        >
          <Tile name={project.name} />
          <span className="flex min-w-0 flex-1 flex-col">
            <span className="truncate text-[15px] font-semibold">{project.name}</span>
            <span className="truncate text-xs text-subtle-foreground">{project.domain}</span>
          </span>
          <ChevronsUpDown aria-hidden="true" className="size-4 shrink-0 text-subtle-foreground" />
        </button>
      </ProjectSwitcher>

      <NavGroup label={say("project-shell.site")} entries={site} unpublished={project.unpublished} />
      {setup.length > 0 && <NavGroup label={say("project-shell.setup")} entries={setup} unpublished={0} />}
    </aside>
  );
}
