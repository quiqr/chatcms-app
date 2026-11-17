import { ExternalLink } from "lucide-react";
import type { ReactNode } from "react";

import { Disclosure } from "../components/ui/disclosure";
import { readAbout, type About } from "../lib/administration";
import { say } from "../lib/say";
import { useAsync } from "../lib/useAsync";

function Section({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section aria-label={title} className="flex flex-col gap-3 rounded-xl border border-border bg-card p-5">
      <h2 className="text-lg font-semibold">{title}</h2>
      {children}
    </section>
  );
}

function Outside({ to, children }: { to: string; children: ReactNode }) {
  return (
    <a
      href={to}
      target="_blank"
      rel="noreferrer"
      className="inline-flex min-h-11 items-center gap-1.5 text-[15px] text-primary underline-offset-4 hover:underline"
    >
      {children}
      <ExternalLink aria-hidden="true" className="size-4" />
    </a>
  );
}

export function versionLine(about: Pick<About, "version" | "source" | "changed">) {
  if (!about.source) return say("about.development-build", { version: about.version });
  if (about.changed) return say("about.version-changed", { version: about.version, source: about.source });
  return say("about.version-at", { version: about.version, source: about.source });
}

export function AboutScreen() {
  const [about] = useAsync(readAbout, []);

  return (
    <div className="flex max-w-[760px] flex-col gap-6">
      <h1 className="text-[28px] font-semibold tracking-tight">{say("about.title")}</h1>

      {about.state === "loading" && <p className="text-[15px] text-subtle-foreground">{say("common.looking")}</p>}
      {about.state === "failed" && <p role="alert" className="text-[15px]">{about.message}</p>}

      {about.state === "ready" && (
        <>
          <Section title={say("about.version")}>
            <p className="text-[15px]">{versionLine(about.value)}</p>
          </Section>

          <Section title={say("about.runs-on")}>
            <dl className="grid grid-cols-[auto_1fr] gap-x-6 gap-y-2 text-[15px]">
              {(
                [
                  ["Hugo", about.value.runsOn.hugo],
                  ["git", about.value.runsOn.git],
                  ["PostgreSQL", about.value.runsOn.postgres],
                  ["Go", about.value.runsOn.go],
                ] as const
              ).map(([name, value]) => (
                <div key={name} className="contents">
                  <dt className="text-subtle-foreground">{name}</dt>
                  <dd>{value || say("about.unknown")}</dd>
                </div>
              ))}
            </dl>
          </Section>

          <Section title={say("about.license")}>
            <p className="text-[15px]">{say("about.license-is", { name: about.value.license.name })}</p>
            <Disclosure summary={say("about.read-the-license")}>
              <pre className="max-h-96 overflow-auto whitespace-pre-wrap rounded-md bg-muted p-3 text-[13px]">
                {about.value.license.text}
              </pre>
            </Disclosure>
          </Section>

          <Section title={say("about.source-code")}>
            <div className="flex flex-col">
              <Outside to={about.value.links.source}>{say("about.on-github")}</Outside>
              <Outside to={about.value.links.changelog}>{say("about.changelog")}</Outside>
              <Outside to={about.value.links.issues}>{say("about.report-a-problem")}</Outside>
            </div>
          </Section>

          <Section title={say("about.credits")}>
            <p className="text-[15px]">{say("about.made-by")}</p>
            <Outside to={about.value.links.contributors}>{say("about.everyone-who-contributed")}</Outside>
          </Section>
        </>
      )}
    </div>
  );
}
