/*
 * chatcms-audience: implementer
 */
import { ChevronLeft, ExternalLink } from "lucide-react";
import { useState } from "react";
import { NavLink, useNavigate, useParams } from "react-router";

import { Button } from "../components/ui/button";
import { Field } from "../components/ui/field";
import { whyThereIsNoRoom, type Usage } from "../components/ui/usage";
import { whatWentWrong } from "../lib/api";
import { plural, say } from "../lib/say";
import { createProject, listProjects } from "../lib/tenancy";
import {
  hugoLabel,
  listTemplates,
  sameHugo,
  slugFor,
  templateSetupAddress,
  type Hugo,
  type Template,
} from "../lib/templates";
import { useAsync } from "../lib/useAsync";
import { useHere } from "../lib/whoami";
import { OrganisationShell } from "../shells/OrganisationShell";
import { Steps } from "./setup/Steps";
import { Screenshot } from "./TemplateGalleryPage";

function Detail({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <>
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="min-w-0 break-words">{children}</dd>
    </>
  );
}

function Link({ href, children }: { href: string; children: React.ReactNode }) {
  if (!href) return <>{children}</>;
  return (
    <a href={href} target="_blank" rel="noreferrer" className="inline-flex min-h-11 items-center text-primary underline">
      {children}
    </a>
  );
}

function forms(count: number) {
  if (count === 0) return say("onboarding.forms-none");
  return plural(count, "onboarding.forms-one", "onboarding.forms-many");
}

function TemplateSummary({ organisation, template }: { organisation: string; template: Template }) {
  const hugo = hugoLabel({ version: template.hugoVersion, extended: template.extended });

  return (
    <section
      aria-label={template.name}
      className="flex flex-col overflow-hidden rounded-xl border border-border bg-card"
    >
      <Screenshot template={template} className="h-[150px] w-full sm:h-auto sm:aspect-[16/10]" />
      <div className="flex flex-col gap-3 p-5">
        <h2 className="text-lg font-semibold">{template.name}</h2>
        <p className="text-[15px] text-muted-foreground">{template.description}</p>

        <dl className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-2 text-sm">
          {template.author && (
            <Detail label={say("onboarding.author")}>
              <Link href={template.authorUrl}>{template.author}</Link>
            </Detail>
          )}
          {template.license && (
            <Detail label={say("onboarding.licence")}>
              <Link href={template.licenseUrl}>{template.license}</Link>
            </Detail>
          )}
          {template.model.length > 0 && (
            <Detail label={say("onboarding.content-model")}>
              <span className="flex flex-wrap gap-1">
                {template.model.map((part) => (
                  <span key={part} className="rounded bg-muted px-1.5 py-0.5 text-[13px]">
                    {part}
                  </span>
                ))}
              </span>
            </Detail>
          )}
          <Detail label={say("onboarding.forms")}>{forms(template.forms)}</Detail>
          {(hugo || template.hugoTheme) && (
            <Detail label={say("onboarding.hugo")}>
              {[hugo, template.hugoTheme && say("onboarding.theme", { theme: template.hugoTheme })]
                .filter(Boolean)
                .join(" · ")}
            </Detail>
          )}
          <Detail label={say("onboarding.source")}>
            <Link href={template.sourceUrl}>{template.sourceUrl.replace(/^https?:\/\//, "")}</Link>
          </Detail>
        </dl>

        <div className="flex flex-wrap items-center gap-3">
          {template.demoUrl && (
            <Button asChild variant="outline" size="compact">
              <a href={template.demoUrl} target="_blank" rel="noreferrer">
                {say("onboarding.live-demo")}
                <ExternalLink aria-hidden="true" className="size-4" />
              </a>
            </Button>
          )}
          <NavLink
            to={`/o/${organisation}/new/template`}
            className="inline-flex min-h-11 items-center text-[15px] text-primary"
          >
            {say("onboarding.choose-another-template")}
          </NavLink>
        </div>
      </div>
    </section>
  );
}

function HugoNote({ template, installed }: { template: Template; installed: Hugo }) {
  const wanted = { version: template.hugoVersion, extended: template.extended };
  if (!wanted.version) return null;

  const differs = installed.version !== "" && !sameHugo(wanted, installed);
  return (
    <p
      role="note"
      className={
        "rounded-md px-3 py-2 text-[13px] " +
        (differs ? "bg-warning-soft text-warning-foreground" : "bg-muted text-muted-foreground")
      }
    >
      {differs
        ? say("onboarding.hugo-differs", { wanted: hugoLabel(wanted), installed: hugoLabel(installed) })
        : say("onboarding.builds-with-hugo", { version: hugoLabel(wanted) })}
    </p>
  );
}

function lastOneLeft(usage: Usage | undefined) {
  const projects = usage?.projects;
  return projects?.allowed !== undefined && projects.allowed - projects.held === 1;
}

export function UseTemplatePage() {
  const { organisation: fromAddress = "", template: id = "" } = useParams();
  const { who } = useHere();
  const goTo = useNavigate();

  const places = who.organisations.filter((one) => one.youAdminister || who.administers);
  const [target, setTarget] = useState(fromAddress);
  const [name, setName] = useState<string | null>(null);
  const [remote, setRemote] = useState("");
  const [trouble, setTrouble] = useState("");
  const [busy, setBusy] = useState(false);

  const [gallery] = useAsync(() => listTemplates(fromAddress), [fromAddress]);
  const [projects] = useAsync(() => listProjects(target), [target]);

  const template = gallery.state === "ready" ? gallery.value.templates.find((one) => one.id === id) : undefined;
  const usage = projects.state === "ready" ? projects.value.usage : undefined;
  const noRoom = usage ? whyThereIsNoRoom(usage.projects, "sites") : null;
  const targetName = places.find((one) => one.slug === target)?.name ?? target;
  const shownName = name ?? template?.name ?? "";

  async function proceed(event: React.FormEvent) {
    event.preventDefault();
    if (!template) return;
    if (slugFor(shownName) === "") {
      setTrouble(say("onboarding.project-needs-a-name"));
      return;
    }
    setBusy(true);
    setTrouble("");
    try {
      const made = await createProject(target, { slug: slugFor(shownName), name: shownName.trim(), domain: "" });
      goTo(templateSetupAddress(target, made.slug, template.id, remote.trim()));
    } catch (error: unknown) {
      setTrouble(whatWentWrong(error));
      setBusy(false);
    }
  }

  const organisationName = who.organisations.find((one) => one.slug === fromAddress)?.name ?? fromAddress;

  return (
    <OrganisationShell
      organisation={{ slug: fromAddress, name: organisationName }}
      organisations={who.organisations}
      viewer={{ name: who.person.displayName || who.person.email }}
    >
      <div className="mb-6 flex items-center gap-3">
        <Button asChild variant="outline" size="icon">
          <NavLink to={`/o/${fromAddress}/new/template`} aria-label={say("onboarding.choose-another-template")}>
            <ChevronLeft aria-hidden="true" className="size-4" />
          </NavLink>
        </Button>
        <div className="min-w-0">
          <p className="text-[13px] text-subtle-foreground">
            {/* chatcms-not-a-string: a trail of names already said elsewhere */}
            {organisationName} · {say("onboarding.new-project")}
          </p>
          <h1 className="truncate text-xl font-semibold">
            {say("onboarding.start-from-name", { name: template?.name ?? id })}
          </h1>
        </div>
      </div>

      {gallery.state === "loading" && (
        <p className="text-[15px] text-muted-foreground">{say("onboarding.loading-the-templates")}</p>
      )}
      {gallery.state === "failed" && (
        <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          {gallery.message}
        </p>
      )}
      {gallery.state === "ready" && !template && (
        <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          {say("onboarding.template-not-in-the-list")}
        </p>
      )}

      {gallery.state === "ready" && template && (
        <div className="flex flex-wrap items-start gap-6">
          <nav aria-label={say("onboarding.setting-up-this-project")} className="min-w-0 grow basis-[200px]">
            <Steps
              current="name"
              steps={[
                { key: "template", label: say("onboarding.template"), done: true },
                { key: "name", label: say("onboarding.name-and-repository"), done: false },
                { key: "key", label: say("onboarding.key-for-the-git-host"), done: false },
                { key: "publishing", label: say("project-settings.how-it-is-published"), done: false },
                { key: "people", label: say("onboarding.invite-people"), done: false },
              ]}
            />
          </nav>

          <div className="min-w-0 grow basis-[300px]">
            <TemplateSummary organisation={fromAddress} template={template} />
          </div>

          <form
            onSubmit={proceed}
            aria-label={say("onboarding.name-and-repository")}
            className="flex min-w-0 grow basis-[340px] flex-col gap-5 rounded-xl border-[1.5px] border-primary bg-card p-5 ring-[3px] ring-primary-soft"
          >
            <div>
              <h2 className="text-lg font-semibold">{say("onboarding.name-and-repository")}</h2>
              <p className="mt-1 text-sm text-muted-foreground">
                {say("onboarding.copies-into-a-repository", { name: template.name })}
              </p>
            </div>

            {trouble && (
              <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
                {trouble}
              </p>
            )}

            <Field
              label={say("onboarding.project-name")}
              required
              value={shownName}
              onChange={(event) => setName(event.target.value)}
            />

            <div className="flex flex-col gap-1.5">
              <label htmlFor="template-organisation" className="text-sm font-medium">
                {say("onboarding.organisation")}
              </label>
              <select
                id="template-organisation"
                value={target}
                onChange={(event) => setTarget(event.target.value)}
                className="h-11 rounded-md border border-input bg-card px-3 text-base sm:text-[15px]"
              >
                {(places.length > 0 ? places : [{ slug: fromAddress, name: organisationName }]).map((one) => (
                  <option key={one.slug} value={one.slug}>
                    {one.name}
                  </option>
                ))}
              </select>
              {usage?.projects.allowed !== undefined && (
                <p className="text-[13px] text-subtle-foreground">
                  {say("limits.sites-used", { held: usage.projects.held, allowed: usage.projects.allowed })}
                </p>
              )}
              {noRoom ? (
                <p role="alert" className="text-[13px] text-destructive-foreground">
                  {noRoom}
                </p>
              ) : (
                lastOneLeft(usage) && (
                  <p className="text-[13px] text-warning-foreground">
                    {say("onboarding.uses-the-last-free-project", { organisation: targetName })}
                  </p>
                )
              )}
            </div>

            <Field
              label={say("onboarding.empty-repository-for-your-copy")}
              hint={say("onboarding.empty-repository-hint")}
              className="font-mono text-sm"
              required
              value={remote}
              onChange={(event) => setRemote(event.target.value)}
            />

            {gallery.state === "ready" && <HugoNote template={template} installed={gallery.value.hugo} />}

            <div className="flex flex-wrap justify-end gap-3 border-t border-border pt-4">
              <Button asChild variant="outline">
                <NavLink to={`/o/${fromAddress}/new/template`}>{say("onboarding.back")}</NavLink>
              </Button>
              <Button type="submit" disabled={busy || noRoom !== null}>
                {say("onboarding.continue-to-deploy-key")}
              </Button>
            </div>
          </form>
        </div>
      )}
    </OrganisationShell>
  );
}
