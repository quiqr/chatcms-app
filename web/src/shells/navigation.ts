import {
  FileDown,
  FileText,
  FolderTree,
  History,
  Image,
  ScrollText,
  SlidersHorizontal,
  Upload,
  Users,
  Workflow,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";
import { say } from "../lib/say";

export type NavEntry = {
  key: string;
  label: string;
  to: string;
  icon: LucideIcon;
  soon?: boolean;
};

export function siteNavigation(organisation: string, projectSlug: string, hasArtifacts = false): NavEntry[] {
  const artifacts: NavEntry[] = hasArtifacts
    ? [{ key: "artifacts", label: say("app.artifacts"), to: `/o/${organisation}/p/${projectSlug}/artifacts`, icon: FileDown }]
    : [];
  return [
    { key: "content", label: say("app.content"), to: `/o/${organisation}/p/${projectSlug}/content`, icon: FileText },
    { key: "media", label: say("app.media"), to: `/o/${organisation}/p/${projectSlug}/media`, icon: Image },
    { key: "publish", label: say("common.publish"), to: `/o/${organisation}/p/${projectSlug}/publish`, icon: Upload },
    { key: "history", label: say("app.history"), to: `/o/${organisation}/p/${projectSlug}/history`, icon: History },
    ...artifacts,
  ];
}

export function setupNavigation(organisation: string, projectSlug: string): NavEntry[] {
  return [
    { key: "people", label: say("people.people-roles"), to: `/o/${organisation}/p/${projectSlug}/people`, icon: Users },
    { key: "settings", label: say("app.project-settings"), to: `/o/${organisation}/p/${projectSlug}/settings`, icon: SlidersHorizontal },
    { key: "files", label: say("app.files"), to: `/o/${organisation}/p/${projectSlug}/files`, icon: FolderTree },
    { key: "log", label: say("app.log"), to: `/o/${organisation}/p/${projectSlug}/log`, icon: ScrollText },
    { key: "workflows", label: say("app.workflows"), to: `/o/${organisation}/p/${projectSlug}/workflows`, icon: Workflow, soon: true },
  ];
}
