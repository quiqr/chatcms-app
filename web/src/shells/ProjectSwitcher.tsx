import { Check } from "lucide-react";
import { useState, type ReactNode } from "react";
import { NavLink } from "react-router";

import { Tile } from "../components/ui/avatar";
import { Menu, MenuContent, MenuItem, MenuLabel, MenuSeparator, MenuTrigger } from "../components/ui/menu";
import { projectPath, projectsPath } from "../lib/paths";
import { say } from "../lib/say";
import { listProjects, type Project } from "../lib/tenancy";
import { useHere } from "../lib/whoami";

export function ProjectSwitcher({ project, children }: { project: Project; children: ReactNode }) {
  const { organisation } = useHere();
  const [projects, setProjects] = useState<Project[] | null>(null);
  const [unreadable, setUnreadable] = useState(false);

  function read(open: boolean) {
    if (!open) return;
    setUnreadable(false);
    listProjects(organisation).then(
      (answer) => setProjects(answer.projects.filter((one) => !one.archived)),
      () => setUnreadable(true),
    );
  }

  return (
    <Menu onOpenChange={read}>
      <MenuTrigger asChild aria-label={say("project-switcher.switch-from", { name: project.name })}>
        {children}
      </MenuTrigger>

      <MenuContent align="start" className="min-w-60">
        {unreadable ? (
          <MenuLabel>{say("project-switcher.could-not-read")}</MenuLabel>
        ) : projects === null ? (
          <MenuLabel>{say("project-switcher.looking")}</MenuLabel>
        ) : (
          projects.map((one) => (
            <MenuItem key={one.slug} asChild>
              <NavLink
                to={projectPath(organisation, one.slug)}
                aria-current={one.slug === project.slug ? "true" : undefined}
              >
                <Tile name={one.name} className="size-7 text-xs" />
                <span className="min-w-0 flex-1 truncate">{one.name}</span>
                {one.slug === project.slug && (
                  <Check aria-label={say("project-switcher.this-one")} className="size-4" />
                )}
              </NavLink>
            </MenuItem>
          ))
        )}

        <MenuSeparator />

        <MenuItem asChild>
          <NavLink to={projectsPath(organisation)}>{say("project-switcher.all-projects")}</NavLink>
        </MenuItem>
      </MenuContent>
    </Menu>
  );
}
