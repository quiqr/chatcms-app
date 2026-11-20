import { ArrowLeft, Box, Hammer, Upload, type LucideIcon } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { useParams, useSearchParams } from "react-router";

import { Button } from "../components/ui/button";
import { Segmented } from "../components/ui/segmented";
import { whatWentWrong } from "../lib/api";
import { cn } from "../lib/cn";
import { maybe, say, type Key } from "../lib/say";
import { readLog, readLogEntry, type LogDetail, type LogEntry, type LogKind } from "../lib/tenancy";
import { when } from "../lib/when";
import { stepNames, stepStates } from "../lib/workspace";
import { useHere } from "../lib/whoami";
import { useIsPhone } from "../shells/useIsPhone";
import { seconds } from "./PlatformWorkspaces";

const kinds: Record<LogKind, { one: Key; many: Key; icon: LucideIcon }> = {
  start: { one: "log.one-start", many: "log.kind-start", icon: Box },
  build: { one: "log.one-build", many: "log.kind-build", icon: Hammer },
  publish: { one: "log.one-publish", many: "log.kind-publish", icon: Upload },
};

const running = new Set(["starting", "running"]);
const fine = new Set(["started", "succeeded"]);

function stateOf(state: string): string {
  if (running.has(state)) return say("log.state-running");
  if (fine.has(state)) return say("log.state-succeeded");
  return maybe(`log.state-${state}`) ?? state;
}

function toneOf(state: string) {
  if (running.has(state)) return "text-info";
  if (fine.has(state)) return "text-success";
  return "text-destructive";
}

function whatOf(entry: LogEntry) {
  if (entry.kind === "publish") return say("log.changes", { count: entry.what ?? "0" });
  return entry.what ?? "";
}

const listEvery = 10_000;
const entryEvery = 1_000;

function useEvery(every: number, look: () => void, on: boolean) {
  useEffect(() => {
    if (!on) return;
    const timer = setInterval(() => {
      if (document.visibilityState === "visible") look();
    }, every);
    return () => clearInterval(timer);
  }, [every, look, on]);
}

function Copy({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);
  return (
    <Button
      type="button"
      variant="ghost"
      size="secondary"
      onClick={() => {
        void navigator.clipboard?.writeText(text);
        setCopied(true);
      }}
    >
      {say(copied ? "log.copied" : "log.copy")}
    </Button>
  );
}

function Facts({ rows }: { rows: [string, string][] }) {
  return (
    <dl className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-[13px]">
      {rows.map(([label, value]) => (
        <div key={label} className="contents">
          <dt className="text-subtle-foreground">{label}</dt>
          <dd className="min-w-0 break-all font-mono">{value}</dd>
        </div>
      ))}
    </dl>
  );
}

function stepName(kind: LogKind, step: string) {
  if (kind === "start" && step in stepNames) return say(stepNames[step as keyof typeof stepNames]);
  return step;
}

function stepState(state: string) {
  if (state in stepStates) return say(stepStates[state as keyof typeof stepStates]);
  return stateOf(state);
}

export function Detail({ organisation, project, entry }: { organisation: string; project: string; entry: string }) {
  const [detail, setDetail] = useState<LogDetail | null>(null);
  const [trouble, setTrouble] = useState("");

  const look = useCallback(() => {
    readLogEntry(organisation, project, entry).then(setDetail, (error: unknown) => setTrouble(whatWentWrong(error)));
  }, [organisation, project, entry]);

  useEffect(() => {
    setDetail(null);
    setTrouble("");
    look();
  }, [look]);
  useEvery(entryEvery, look, detail !== null && running.has(detail.state));

  if (trouble) {
    return (
      <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
        {trouble}
      </p>
    );
  }
  if (!detail) return <p className="text-[13px] text-muted-foreground">{say("common.looking")}</p>;

  const hugo = detail.hugo;
  return (
    <div data-testid="log-detail" className="flex flex-col gap-4">
      {detail.reason && (
        <section aria-label={say("log.what-went-wrong")} className="flex flex-col gap-1">
          <h3 className="text-[13px] font-semibold">{say("log.what-went-wrong")}</h3>
          <p className="break-all font-mono text-[12px]">{detail.reason}</p>
        </section>
      )}
      {hugo && (
        <section aria-label={say("log.hugo")} className="flex flex-col gap-1">
          <h3 className="text-[13px] font-semibold">{say("log.hugo")}</h3>
          <Facts
            rows={[
              [say("log.hugo-asked"), hugo.asked || say("log.hugo-none")],
              [say("log.hugo-program"), hugo.program],
              [say("log.hugo-because"), maybe(`log.because-${hugo.because}`) ?? hugo.because],
              ...(hugo.printed ? ([[say("log.hugo-printed"), hugo.printed]] as [string, string][]) : []),
            ]}
          />
        </section>
      )}
      {detail.steps.length > 0 && (
        <section aria-label={say("log.steps")} className="flex flex-col gap-1">
          <h3 className="text-[13px] font-semibold">{say("log.steps")}</h3>
          <ul className="flex flex-col gap-1 text-[13px]">
            {detail.steps.map((step) => (
              <li key={step.step} className="flex flex-col gap-1">
                <span className="flex justify-between gap-4">
                  <span>{stepName(detail.kind, step.step)}</span>
                  <span className={toneOf(step.state === "done" ? "succeeded" : step.state === "waiting" ? "running" : step.state)}>
                    {stepState(step.state)}
                  </span>
                </span>
                {step.detail && (
                  <pre className="max-h-40 overflow-auto rounded bg-muted p-2 text-[12px] whitespace-pre-wrap">{step.detail}</pre>
                )}
              </li>
            ))}
          </ul>
        </section>
      )}
      <section aria-label={say("log.the-log")} className="flex flex-col gap-1">
        <div className="flex items-center justify-between gap-2">
          <h3 className="text-[13px] font-semibold">{say("log.the-log")}</h3>
          {detail.log && <Copy text={detail.log} />}
        </div>
        {detail.exit !== undefined && <p className="text-[13px] text-subtle-foreground">{say("log.exit", { code: detail.exit })}</p>}
        {detail.log ? (
          <pre className="max-h-[28rem] overflow-auto rounded bg-muted p-2 text-[12px] whitespace-pre-wrap">{detail.log}</pre>
        ) : (
          <p className="text-[13px] text-subtle-foreground">{say("log.empty-log")}</p>
        )}
      </section>
    </div>
  );
}

function Row({ entry, open, onToggle }: { entry: LogEntry; open: boolean; onToggle: () => void }) {
  const { one, icon: Icon } = kinds[entry.kind];
  return (
    <button
      type="button"
      aria-expanded={open}
      onClick={onToggle}
      className="flex min-h-11 w-full items-center gap-3 px-4 py-2 text-left hover:bg-muted"
    >
      <Icon aria-hidden="true" className="size-4 shrink-0 text-subtle-foreground" />
      <span className="flex min-w-0 flex-1 flex-col sm:flex-row sm:items-center sm:gap-2">
        <span className="truncate text-[15px] font-medium">{say(one)}</span>
        <span className="truncate text-[13px] text-subtle-foreground">
          {[entry.who, whatOf(entry)].filter(Boolean).join(" · ")}
        </span>
      </span>
      <span className="flex shrink-0 flex-col items-end text-[13px] sm:flex-row sm:gap-3">
        <span className="text-subtle-foreground">{when(entry.at)}</span>
        <span className={cn("tabular-nums", toneOf(entry.state))}>
          {stateOf(entry.state)}
          {entry.millis !== undefined && !running.has(entry.state) ? ` · ${seconds(entry.millis)}` : ""}
        </span>
      </span>
    </button>
  );
}

export function LogPage() {
  const { organisation } = useHere();
  const { project = "" } = useParams();
  const onAPhone = useIsPhone();
  const [asked, setAsked] = useSearchParams();
  const kind = (asked.get("kind") ?? "") as LogKind | "";
  const opened = asked.get("entry") ?? "";
  const [entries, setEntries] = useState<LogEntry[] | null>(null);
  const [more, setMore] = useState(false);
  const [trouble, setTrouble] = useState("");

  const look = useCallback(() => {
    readLog(organisation, project, kind).then(
      (answer) => {
        setEntries(answer.entries);
        setMore(answer.more);
        setTrouble("");
      },
      (error: unknown) => setTrouble(whatWentWrong(error)),
    );
  }, [organisation, project, kind]);

  useEffect(() => {
    setEntries(null);
    look();
  }, [look]);
  useEvery(listEvery, look, true);

  const olderThan = () => {
    const last = entries?.[entries.length - 1];
    if (!last) return;
    readLog(organisation, project, kind, last.at).then((answer) => {
      setEntries((before) => [...(before ?? []), ...answer.entries]);
      setMore(answer.more);
    });
  };

  const choose = (next: { kind?: string; entry?: string }) => {
    const query = new URLSearchParams(asked);
    for (const [name, value] of Object.entries(next)) {
      if (value) query.set(name, value);
      else query.delete(name);
    }
    setAsked(query, { replace: next.kind !== undefined });
  };

  if (onAPhone && opened) {
    return (
      <div className="flex flex-col gap-4 p-4">
        <Button type="button" variant="ghost" size="secondary" className="self-start" onClick={() => choose({ entry: "" })}>
          <ArrowLeft aria-hidden="true" className="size-4" />
          {say("log.back")}
        </Button>
        <Detail organisation={organisation} project={project} entry={opened} />
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4 p-4 sm:p-6">
      <div className="flex flex-col gap-1">
        <h1 className="text-[28px] font-semibold tracking-tight">{say("log.title")}</h1>
        <p className="text-[15px] text-muted-foreground">{say("log.intro")}</p>
      </div>
      <Segmented
        label={say("log.which-kind")}
        className="flex h-auto flex-wrap self-start"
        value={kind || "all"}
        onValueChange={(next) => choose({ kind: next === "all" ? "" : next })}
        options={[
          { value: "all", label: say("log.all") },
          { value: "start", label: say("log.kind-start") },
          { value: "build", label: say("log.kind-build") },
          { value: "publish", label: say("log.kind-publish") },
        ]}
      />
      {trouble && (
        <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          {trouble}
        </p>
      )}
      {entries === null && !trouble && <p className="text-[15px] text-muted-foreground">{say("common.looking")}</p>}
      {entries !== null && entries.length === 0 && <p className="text-[15px] text-muted-foreground">{say("log.nothing-yet")}</p>}
      {entries !== null && entries.length > 0 && (
        <ul aria-label={say("log.title")} className="flex flex-col divide-y divide-border overflow-hidden rounded-xl border border-border bg-card">
          {entries.map((entry) => (
            <li key={entry.entry}>
              <Row entry={entry} open={opened === entry.entry} onToggle={() => choose({ entry: opened === entry.entry ? "" : entry.entry })} />
              {!onAPhone && opened === entry.entry && (
                <div className="border-t border-border px-4 py-4">
                  <Detail organisation={organisation} project={project} entry={entry.entry} />
                </div>
              )}
            </li>
          ))}
        </ul>
      )}
      {more && (
        <Button type="button" variant="outline" size="secondary" className="self-start" onClick={olderThan}>
          {say("log.older")}
        </Button>
      )}
    </div>
  );
}
