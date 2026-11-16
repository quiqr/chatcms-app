import { useState } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router";

import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Field } from "../components/ui/field";
import { Segmented } from "../components/ui/segmented";
import { useAsync } from "../lib/useAsync";
import { KeyPanel, PublishingPanel } from "./setup/Panels";
import { Steps, type Step } from "./setup/Steps";
import { useHere } from "../lib/whoami";
import { say } from "../lib/say";
import {
  importSite,
  readKey,
  readPublishing,
  readSite,
  type SiteReport,
} from "../lib/tenancy";
import { whatWentWrong } from "../lib/api";

function Section({
  title,
  active,
  children,
}: {
  title: string;
  active?: boolean;
  children: React.ReactNode;
}) {
  return (
    <section
      aria-label={title}
      className={
        "rounded-xl border bg-card p-5 " +
        (active ? "border-[1.5px] border-primary ring-[3px] ring-primary-soft" : "border-border")
      }
    >
      <h3 className="text-[15px] font-semibold">{title}</h3>
      <div className="mt-3">{children}</div>
    </section>
  );
}

function siteSummary(report: SiteReport) {
  if (report.kind === "none") return say("onboarding.nothing-has-been-brought-in-yet");
  if (report.ready) {
    return `Hugo site, Quiqr model found: ${report.pages} pages and ${report.collections} collections.`;
  }
  return report.problem ?? "Something is missing.";
}

function FromTemplate({
  template,
  remote,
  hasKey,
  busy,
  onCopy,
}: {
  template: string;
  remote: string;
  hasKey: boolean;
  busy: boolean;
  onCopy: () => void;
}) {
  return (
    <div className="flex flex-col gap-3">
      <p className="text-[15px]">
        {say("onboarding.starting-from-template", { name: template, remote })}
      </p>
      <div>
        <Button variant="ink" onClick={onCopy} disabled={busy || !hasKey}>
          {busy ? say("onboarding.copying-the-template") : say("onboarding.copy-the-template-in")}
        </Button>
      </div>
    </div>
  );
}

export function OnboardingPage() {
  const { organisation } = useHere();
  const { project = "" } = useParams();
  const goTo = useNavigate();
  const [asked] = useSearchParams();
  const fromTemplate = asked.get("template") ?? "";
  const templateRemote = asked.get("remote") ?? "";

  const [site, reloadSite] = useAsync(() => readSite(organisation, project), [organisation, project]);
  // Read here only to say which step is done; the panels load their own.
  const [key, reloadKey] = useAsync(() => readKey(organisation, project), [organisation, project]);
  const [publishing] = useAsync(
    () => readPublishing(organisation, project),
    [organisation, project],
  );

  const [source, setSource] = useState("repository");
  const [address, setAddress] = useState("");
  const [problem, setProblem] = useState("");
  const [busy, setBusy] = useState(false);

  async function run(what: () => Promise<unknown>) {
    setProblem("");
    setBusy(true);
    try {
      await what();
    } catch (error) {
      setProblem(whatWentWrong(error, "common.that-did-not-work-try-again"));
    } finally {
      setBusy(false);
    }
  }

  const report = site.state === "ready" ? site.value : undefined;
  const hasKey = key.state === "ready" && key.value.publicKey !== "";
  const arrived = Boolean(report && report.kind !== "none");
  const templateMode = fromTemplate !== "" && !arrived;
  const current = templateMode
    ? hasKey
      ? "site"
      : "key"
    : !arrived
      ? "site"
      : !hasKey
        ? "key"
        : "publishing";

  const steps: Step[] = fromTemplate
    ? [
        { key: "template", label: say("onboarding.template"), done: true },
        { key: "name", label: say("onboarding.name-and-repository"), done: true },
        { key: "key", label: say("onboarding.key-for-the-git-host"), done: hasKey && arrived },
        { key: "publishing", label: say("project-settings.how-it-is-published"), done: arrived && publishing.state === "ready" },
        { key: "people", label: say("onboarding.invite-people"), done: false },
      ]
    : [
        { key: "site", label: say("onboarding.where-the-site-comes-from"), done: arrived },
        { key: "key", label: say("onboarding.key-for-the-git-host"), done: hasKey },
        { key: "publishing", label: say("project-settings.how-it-is-published"), done: publishing.state === "ready" },
        { key: "people", label: say("onboarding.invite-people"), done: false },
      ];

  return (
    <div className="flex flex-wrap items-start gap-6">
      <nav aria-label={say("onboarding.setting-up-this-project")} className="min-w-0 shrink grow basis-[220px]">
        <Steps steps={steps} current={current} />
      </nav>

      <div className="flex min-w-0 shrink grow-[999] basis-[520px] flex-col gap-4">
        {problem && (
          <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
            {problem}
          </p>
        )}

        <Section title={say("onboarding.where-the-site-comes-from")} active={current === "site"}>
          {templateMode ? (
            <FromTemplate
              template={fromTemplate}
              remote={templateRemote}
              hasKey={hasKey}
              busy={busy}
              onCopy={() =>
                void run(async () => {
                  await importSite(organisation, project, {
                    kind: "template",
                    address: fromTemplate,
                    remote: templateRemote,
                  });
                  reloadSite();
                })
              }
            />
          ) : (
          <>
          <Segmented
            label={say("onboarding.where-the-site-comes-from")}
            value={source}
            onValueChange={(value) => {
              if (value === "template") {
                goTo(`/o/${organisation}/new/template`);
                return;
              }
              setSource(value);
            }}
            options={[
              { value: "repository", label: say("onboarding.a-site-already-in-git") },
              { value: "export", label: say("onboarding.from-quiqr-desktop") },
              { value: "template", label: say("onboarding.start-from-a-template") },
            ]}
          />

          <form
            className="mt-4 flex flex-wrap items-end gap-3"
            onSubmit={(event) => {
              event.preventDefault();
              void run(async () => {
                await importSite(organisation, project, { kind: source, address });
                reloadSite();
              });
            }}
          >
            <div className="min-w-[240px] flex-1">
              <Field
                label={
                  source === "repository"
                    ? say("onboarding.address-of-the-site-in-git")
                    : say("onboarding.folder-holding-the-exported-site")
                }
                className="font-mono text-sm"
                required
                value={address}
                onChange={(event) => setAddress(event.target.value)}
              />
            </div>
            <Button type="submit" variant="ink" disabled={busy}>
              {say("onboarding.bring-it-in")}
            </Button>
          </form>
          </>
          )}

          <p className="mt-3 text-[13px] text-subtle-foreground">
            {site.state === "loading" ? say("common.looking") : report ? siteSummary(report) : ""}
          </p>

          {report?.ready && <Badge tone="live">{say("onboarding.ready-to-edit")}</Badge>}
        </Section>

        <Section title={say("onboarding.key-for-the-git-host")} active={current === "key"}>
          <KeyPanel organisation={organisation} project={project} onTrouble={setProblem} onKey={reloadKey} />
        </Section>

        <Section title={say("onboarding.how-it-is-published")} active={current === "publishing"}>
          <PublishingPanel organisation={organisation} project={project} onTrouble={setProblem} />
        </Section>

      </div>
    </div>
  );
}
