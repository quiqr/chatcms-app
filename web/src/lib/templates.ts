import { get, put } from "./api";
import { say } from "./say";

export type Hugo = { version: string; extended: boolean };

export type Template = {
  id: string;
  name: string;
  description: string;
  author: string;
  authorUrl: string;
  license: string;
  licenseUrl: string;
  homepage: string;
  demoUrl: string;
  hugoTheme: string;
  hugoVersion: string;
  extended: boolean;
  model: string[];
  forms: number;
  sourceUrl: string;
  screenshotUrl: string;
};

export type Gallery = {
  offered: boolean;
  available: boolean;
  templates: Template[];
  hugo: Hugo;
};

export type ProjectHugo = Hugo & { installed: Hugo };

export function listTemplates(organisation: string) {
  return get<Gallery>(`/api/organisations/${organisation}/templates`);
}

export function readHugo(organisation: string, project: string) {
  return get<ProjectHugo>(`/api/organisations/${organisation}/projects/${project}/hugo`);
}

export function setHugo(organisation: string, project: string, hugo: Hugo) {
  return put<ProjectHugo>(`/api/organisations/${organisation}/projects/${project}/hugo`, hugo);
}

export function hugoLabel(hugo: Hugo): string {
  if (!hugo.version) return "";
  return hugo.extended ? say("onboarding.extended", { version: hugo.version }) : hugo.version;
}

export function sameHugo(a: Hugo, b: Hugo): boolean {
  return a.version === b.version && (!a.extended || b.extended);
}

export function matches(template: Template, search: string): boolean {
  const needle = search.trim().toLowerCase();
  if (!needle) return true;
  return [template.name, template.description, template.author].some((part) =>
    part.toLowerCase().includes(needle),
  );
}

export function slugFor(name: string): string {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

export function templateSetupAddress(organisation: string, slug: string, template: string, remote: string) {
  const query = new URLSearchParams({ template, remote });
  return `/o/${organisation}/p/${slug}/setup?${query.toString()}`;
}
