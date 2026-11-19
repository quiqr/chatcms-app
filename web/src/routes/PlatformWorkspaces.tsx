import type { ReactNode } from "react";

import { readWorkspaceStatistics, type WorkspaceCounts } from "../lib/administration";
import { aNumber } from "../lib/numbers";
import { say } from "../lib/say";
import { useAsync } from "../lib/useAsync";
import { sizeOf } from "./ArtifactsPage";

function Section({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section aria-label={title} className="flex flex-col gap-3 rounded-xl border border-border bg-card p-5">
      <h2 className="text-lg font-semibold">{title}</h2>
      {children}
    </section>
  );
}

function Figure({ label, value, note }: { label: string; value: string; note?: string }) {
  return (
    <div className="flex min-w-[8rem] flex-1 flex-col gap-0.5 rounded-lg bg-muted px-4 py-3">
      <span className="text-[13px] text-subtle-foreground">{label}</span>
      <span className="text-[24px] font-semibold tabular-nums">{value}</span>
      {note && <span className="text-[13px] text-subtle-foreground">{note}</span>}
    </div>
  );
}

export function seconds(millis: number) {
  return say("workspaces-admin.seconds", { value: aNumber(Math.round(millis / 100) / 10) });
}

function CountsLine({ counts }: { counts: WorkspaceCounts }) {
  return (
    <span className="text-[13px] tabular-nums text-subtle-foreground">
      {say("workspaces-admin.counts", { starting: counts.starting, warm: counts.warm, sleeping: counts.sleeping })}
    </span>
  );
}

export function PlatformWorkspaces() {
  const [loaded] = useAsync(readWorkspaceStatistics, []);

  return (
    <div className="flex max-w-[960px] flex-col gap-6">
      <h1 className="text-[28px] font-semibold tracking-tight">{say("workspaces-admin.title")}</h1>

      {loaded.state === "loading" && <p className="text-[15px] text-subtle-foreground">{say("common.looking")}</p>}
      {loaded.state === "failed" && (
        <p role="alert" className="text-[15px]">
          {loaded.message}
        </p>
      )}

      {loaded.state === "ready" && (
        <>
          <Section title={say("workspaces-admin.now")}>
            <div className="flex flex-wrap gap-3">
              <Figure
                label={say("workspaces-admin.running")}
                value={say("workspaces-admin.of-limit", { running: loaded.value.now.starting + loaded.value.now.warm, limit: loaded.value.limit })}
              />
              <Figure label={say("workspaces-admin.warm")} value={aNumber(loaded.value.now.warm)} />
              <Figure label={say("workspaces-admin.starting")} value={aNumber(loaded.value.now.starting)} />
              <Figure label={say("workspaces-admin.sleeping")} value={aNumber(loaded.value.now.sleeping)} />
            </div>
            <p className="text-[15px]">
              {loaded.value.backends.length > 0
                ? say("workspaces-admin.backends", { backends: loaded.value.backends.join(", ") })
                : say("workspaces-admin.none-running")}
            </p>
          </Section>

          <Section title={say("workspaces-admin.past-days", { days: loaded.value.days })}>
            <div className="flex flex-wrap gap-3">
              <Figure label={say("workspaces-admin.starts")} value={aNumber(loaded.value.starts)} />
              <Figure label={say("workspaces-admin.cold-start")} value={seconds(loaded.value.medianMillis)} note={say("workspaces-admin.slow", { value: seconds(loaded.value.slowMillis) })} />
              <Figure label={say("workspaces-admin.refused")} value={aNumber(loaded.value.refused)} />
            </div>
            <h3 className="text-[15px] font-semibold">{say("workspaces-admin.failed")}</h3>
            {loaded.value.failed.length === 0 ? (
              <p className="text-[15px] text-subtle-foreground">{say("workspaces-admin.no-failures")}</p>
            ) : (
              <ul aria-label={say("workspaces-admin.failed")} className="flex flex-col divide-y divide-border rounded-lg border border-border">
                {loaded.value.failed.map((one) => (
                  <li key={one.reason} className="flex items-start gap-3 px-3 py-2 text-[15px]">
                    <span className="min-w-0 flex-1 break-words">{one.reason}</span>
                    <span className="shrink-0 tabular-nums">{aNumber(one.times)}</span>
                  </li>
                ))}
              </ul>
            )}
            <h3 className="text-[15px] font-semibold">{say("workspaces-admin.peaks")}</h3>
            {loaded.value.peaks.length === 0 ? (
              <p className="text-[15px] text-subtle-foreground">{say("workspaces-admin.no-peaks")}</p>
            ) : (
              <ul aria-label={say("workspaces-admin.peaks")} className="grid grid-cols-[repeat(auto-fill,minmax(7.5rem,1fr))] gap-2">
                {loaded.value.peaks.map((peak) => (
                  <li key={peak.day} className="flex flex-col rounded-lg bg-muted px-3 py-2">
                    <span className="text-[13px] text-subtle-foreground">{peak.day}</span>
                    <span className="text-[15px] font-semibold tabular-nums">{say("workspaces-admin.warm-at-once", { warm: peak.warm })}</span>
                  </li>
                ))}
              </ul>
            )}
            <p className="text-[13px] text-subtle-foreground">{say("workspaces-admin.kept", { days: loaded.value.historyDays })}</p>
          </Section>

          <Section title={say("workspaces-admin.per-organisation")}>
            {loaded.value.organisations.length === 0 ? (
              <p className="text-[15px] text-subtle-foreground">{say("workspaces-admin.none-yet")}</p>
            ) : (
              <ul aria-label={say("workspaces-admin.per-organisation")} className="flex flex-col divide-y divide-border">
                {loaded.value.organisations.map((one) => (
                  <li key={one.slug} className="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1 py-2">
                    <span className="text-[15px] font-medium">{one.name}</span>
                    <CountsLine counts={one.counts} />
                  </li>
                ))}
              </ul>
            )}
          </Section>

          <Section title={say("workspaces-admin.per-project")}>
            {loaded.value.projects.length === 0 ? (
              <p className="text-[15px] text-subtle-foreground">{say("workspaces-admin.none-yet")}</p>
            ) : (
              <ul aria-label={say("workspaces-admin.per-project")} className="flex flex-col divide-y divide-border">
                {loaded.value.projects.map((one) => (
                  <li key={`${one.organisation}/${one.slug}`} className="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1 py-2">
                    <span className="flex min-w-0 flex-col">
                      <span className="text-[15px] font-medium">{one.name}</span>
                      <span className="text-[13px] text-subtle-foreground">{one.organisation}</span>
                    </span>
                    <span className="flex flex-col items-end">
                      <CountsLine counts={one.counts} />
                      <span className="text-[13px] tabular-nums">
                        {one.diskKnown ? say("workspaces-admin.disk", { size: sizeOf(one.diskBytes) }) : say("workspaces-admin.disk-unknown")}
                      </span>
                    </span>
                  </li>
                ))}
              </ul>
            )}
          </Section>
        </>
      )}
    </div>
  );
}
