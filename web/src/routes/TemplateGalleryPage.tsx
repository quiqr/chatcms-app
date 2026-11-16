/*
 * chatcms-audience: implementer
 */
import { ExternalLink, Search } from "lucide-react";
import { useMemo, useState } from "react";
import { NavLink, useNavigate, useParams } from "react-router";

import { Button } from "../components/ui/button";
import { Segmented } from "../components/ui/segmented";
import { plural, say } from "../lib/say";
import { listTemplates, matches, type Template } from "../lib/templates";
import { useAsync } from "../lib/useAsync";
import { useHere } from "../lib/whoami";
import { OrganisationShell } from "../shells/OrganisationShell";

const COMMUNITY_CATALOGUE = "https://github.com/quiqr/quiqr-community-templates";

export function Screenshot({ template, className = "" }: { template: Template; className?: string }) {
  const [broken, setBroken] = useState(false);

  if (!template.screenshotUrl || broken) {
    return (
      <span
        className={
          "flex items-center justify-center bg-muted px-4 text-center text-lg font-semibold text-muted-foreground " +
          className
        }
      >
        {template.name}
      </span>
    );
  }
  return (
    <img
      src={template.screenshotUrl}
      alt={say("onboarding.screenshot-of", { name: template.name })}
      loading="lazy"
      onError={() => setBroken(true)}
      className={"object-cover object-top " + className}
    />
  );
}

function byline(template: Template) {
  return [template.author, template.license].filter(Boolean).join(" · ");
}

function TemplateCard({ organisation, template }: { organisation: string; template: Template }) {
  const use = `/o/${organisation}/new/template/${encodeURIComponent(template.id)}`;

  return (
    <article
      aria-label={template.name}
      className="flex flex-col overflow-hidden rounded-xl border border-border bg-card"
    >
      <NavLink to={use} tabIndex={-1} aria-hidden="true" className="block">
        <Screenshot template={template} className="h-[150px] w-full sm:h-auto sm:aspect-[16/10]" />
      </NavLink>

      <div className="flex flex-1 flex-wrap items-center gap-3 p-4 sm:flex-col sm:flex-nowrap sm:items-stretch">
        <div className="flex min-w-0 flex-1 flex-col gap-1">
          <h3 className="text-base font-semibold">{template.name}</h3>
          <p className="line-clamp-1 text-sm text-muted-foreground sm:line-clamp-3">{template.description}</p>
          {byline(template) && (
            <p className="hidden text-[13px] text-subtle-foreground sm:block">{byline(template)}</p>
          )}
        </div>

        <div className="flex gap-2 sm:mt-auto">
          {template.demoUrl && (
            <Button asChild variant="outline" size="compact" className="hidden sm:inline-flex">
              <a href={template.demoUrl} target="_blank" rel="noreferrer">
                {say("onboarding.live-demo")}
                <ExternalLink aria-hidden="true" className="size-4" />
              </a>
            </Button>
          )}
          <Button asChild size="compact">
            <NavLink to={use} aria-label={`${say("onboarding.use-template")}: ${template.name}`}>
              <span className="sm:hidden">{say("onboarding.use")}</span>
              <span className="hidden sm:inline">{say("onboarding.use-template")}</span>
            </NavLink>
          </Button>
        </div>
      </div>
    </article>
  );
}

function Skeletons() {
  return (
    <div
      aria-busy="true"
      className="grid gap-[18px] [grid-template-columns:repeat(auto-fill,minmax(min(270px,100%),1fr))]"
    >
      <span className="sr-only">{say("onboarding.loading-the-templates")}</span>
      {Array.from({ length: 6 }, (_, index) => (
        <div key={index} aria-hidden="true" className="overflow-hidden rounded-xl border border-border bg-card">
          <div className="h-[150px] animate-pulse bg-muted sm:h-auto sm:aspect-[16/10]" />
          <div className="flex flex-col gap-2 p-4">
            <div className="h-4 w-1/2 animate-pulse rounded bg-muted" />
            <div className="h-3 w-full animate-pulse rounded bg-muted" />
            <div className="h-3 w-3/4 animate-pulse rounded bg-muted" />
          </div>
        </div>
      ))}
    </div>
  );
}

function Unavailable({ organisation, onRetry }: { organisation: string; onRetry: () => void }) {
  return (
    <div
      role="alert"
      className="flex flex-wrap items-center gap-4 rounded-xl border border-destructive/40 bg-card p-5"
    >
      <div className="min-w-[220px] flex-1">
        <p className="text-[15px] font-semibold">{say("onboarding.template-list-could-not-load")}</p>
        <p className="text-sm text-muted-foreground">{say("onboarding.template-list-could-not-load-why")}</p>
      </div>
      <Button variant="outline" size="compact" onClick={onRetry}>
        {say("onboarding.try-again")}
      </Button>
      <NavLink
        to={`/o/${organisation}/projects?new=repository`}
        className="inline-flex min-h-11 items-center text-[15px] text-primary"
      >
        {say("onboarding.import-a-git-repository-instead")}
      </NavLink>
    </div>
  );
}

export function TemplateGalleryPage() {
  const { organisation: fromAddress = "" } = useParams();
  const { who } = useHere();
  const goTo = useNavigate();
  const [search, setSearch] = useState("");
  const [gallery, again] = useAsync(() => listTemplates(fromAddress), [fromAddress]);

  const shown = useMemo(
    () => (gallery.state === "ready" ? gallery.value.templates.filter((one) => matches(one, search)) : []),
    [gallery, search],
  );

  const organisationName = who.organisations.find((one) => one.slug === fromAddress)?.name ?? fromAddress;
  const total = gallery.state === "ready" ? gallery.value.templates.length : 0;

  return (
    <OrganisationShell
      organisation={{ slug: fromAddress, name: organisationName }}
      organisations={who.organisations}
      viewer={{ name: who.person.displayName || who.person.email }}
    >
      <nav aria-label={say("onboarding.new-project")} className="mb-4 text-[13px] text-subtle-foreground">
        {/* chatcms-not-a-string: a trail of names already said elsewhere */}
        {organisationName} / {say("onboarding.new-project")} / {say("onboarding.choose-a-template")}
      </nav>

      <div className="max-w-full overflow-x-auto">
        <Segmented
          label={say("onboarding.where-the-site-comes-from")}
          value="template"
          onValueChange={(value) => {
            if (value !== "template") goTo(`/o/${fromAddress}/projects?new=${value}`);
          }}
          options={[
            { value: "repository", label: say("onboarding.a-site-already-in-git") },
            { value: "export", label: say("onboarding.from-quiqr-desktop") },
            { value: "template", label: say("onboarding.start-from-a-template") },
          ]}
        />
      </div>

      <div className="mt-6 flex flex-wrap items-end gap-4">
        <div className="min-w-[260px] flex-1">
          <h1 className="text-[28px] font-semibold tracking-tight">{say("onboarding.start-from-a-template")}</h1>
          <p className="mt-1 text-[15px] text-muted-foreground">{say("onboarding.ready-made-hugo-sites")}</p>
        </div>

        <label className="flex h-11 w-full items-center gap-2 rounded-md border border-input bg-card px-3 sm:w-[326px]">
          <Search aria-hidden="true" className="size-4 text-subtle-foreground" />
          <span className="sr-only">{say("onboarding.search-templates")}</span>
          <input
            // chatcms-target: the label around this one carries the hit area
            type="search"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder={say("onboarding.search-templates")}
            className="h-full min-w-0 flex-1 bg-transparent text-[15px] outline-hidden"
          />
        </label>
      </div>

      <div className="mt-6">
        {gallery.state === "loading" && <Skeletons />}

        {gallery.state === "failed" && (
          <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
            {gallery.message}
          </p>
        )}

        {gallery.state === "ready" && !gallery.value.offered && (
          <p role="status" className="text-[15px] text-muted-foreground">
            {say("onboarding.no-templates-here")}
          </p>
        )}

        {gallery.state === "ready" && gallery.value.offered && !gallery.value.available && (
          <Unavailable organisation={fromAddress} onRetry={again} />
        )}

        {gallery.state === "ready" && gallery.value.available && (
          <>
            <p role="status" className="mb-3 text-[13px] text-subtle-foreground">
              {search.trim()
                ? say("onboarding.shown-of-templates", { shown: shown.length, count: total })
                : plural(total, "onboarding.templates-one", "onboarding.templates-many")}
            </p>

            {shown.length === 0 ? (
              <p className="text-[15px] text-muted-foreground">
                {say("onboarding.no-templates-match", { search: search.trim() })}
              </p>
            ) : (
              <div className="grid gap-[18px] [grid-template-columns:repeat(auto-fill,minmax(min(270px,100%),1fr))]">
                {shown.map((template) => (
                  <TemplateCard key={template.id} organisation={fromAddress} template={template} />
                ))}
              </div>
            )}

            <p className="mt-6 text-[13px] text-subtle-foreground">
              {say("onboarding.templates-come-from")}{" "}
              <a
                className="inline-flex min-h-11 items-center text-primary underline"
                href={COMMUNITY_CATALOGUE}
                target="_blank"
                rel="noreferrer"
              >
                {say("onboarding.quiqr-community-catalogue")}
              </a>
              {/* chatcms-not-a-string: the full stop that ends the sentence */}
              . {say("onboarding.want-yours-listed")}
            </p>
          </>
        )}
      </div>
    </OrganisationShell>
  );
}
