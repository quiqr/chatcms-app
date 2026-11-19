import { Check, Eye, Quote, Send, X } from "lucide-react";
import { useEffect, useImperativeHandle, useRef, useState, type Ref } from "react";
import { useNavigate } from "react-router";

import { Button } from "../components/ui/button";
import { whatWentWrong } from "../lib/api";
import { say } from "../lib/say";
import {
  askAssistant,
  assistantEvents,
  goBackTo,
  readAssistant,
  stopAssistant,
  type AssistantChange,
  type AssistantEvent,
  type PreviewSelection,
} from "../lib/tenancy";

const listenWithin = 120_000;

type Asked = PreviewSelection & { page?: string };

export type RequestState = { id: number; state: "working" | "done" | "question"; before: string; after: string; text: string };

export type ChatControl = {
  undo: (id: number) => void;
  stop: (id: number) => void;
  seen: (id: number) => void;
  focus: () => void;
};

type Said = {
  who: "person" | "assistant";
  id?: number;
  text: string;
  about?: Asked;
  stopped?: boolean;
  changes: (AssistantChange & { undone?: boolean })[];
  working?: boolean;
  trouble?: string;
};

export function ChatPanel({
  organisation,
  project,
  onChanged,
  onClose,
  quote = null,
  onQuoteDropped,
  onRunning,
  onShowPlace,
  pageTitle,
  request = null,
  drafted = null,
  onRequests,
  control,
}: {
  organisation: string;
  project: string;
  onChanged: (about?: PreviewSelection) => void;
  onClose?: () => void;
  quote?: PreviewSelection | null;
  onQuoteDropped?: () => void;
  onRunning?: (count: number) => void;
  onShowPlace?: (about: PreviewSelection) => void;
  pageTitle?: string;
  request?: { about: PreviewSelection; text: string; id: number } | null;
  drafted?: { text: string; id: number } | null;
  onRequests?: (requests: RequestState[]) => void;
  control?: Ref<ChatControl>;
}) {
  const [available, setAvailable] = useState<boolean | null>(null);
  const [said, setSaid] = useState<Said[]>([]);
  const [draft, setDraft] = useState("");
  const [busy, setBusy] = useState(false);
  const listening = useRef(false);
  const waiting = useRef<(() => void)[]>([]);
  const [trouble, setTrouble] = useState("");
  const changed = useRef(onChanged);
  changed.current = onChanged;
  const asking = useRef<PreviewSelection | null>(null);
  const message = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    if (quote) message.current?.focus();
  }, [quote]);

  const [seen, setSeen] = useState<number[]>([]);
  const [waitingTurn, setWaitingTurn] = useState<{ id: number; text: string; selection: PreviewSelection }[]>([]);
  const requests = requestStates(said, waitingTurn, seen);
  const running = requests.filter((one) => one.state !== "done").length + said.filter((entry) => entry.who === "assistant" && entry.working && !entry.about).length;
  const reportRunning = useRef(onRunning);
  reportRunning.current = onRunning;
  useEffect(() => {
    reportRunning.current?.(running);
  }, [running]);
  const reportRequests = useRef(onRequests);
  reportRequests.current = onRequests;
  const requestsKey = JSON.stringify(requests);
  useEffect(() => {
    reportRequests.current?.(JSON.parse(requestsKey) as RequestState[]);
  }, [requestsKey]);

  useEffect(() => {
    let live = true;
    readAssistant(organisation, project).then(
      (found) => live && setAvailable(found.available),
      () => live && setAvailable(false),
    );
    return () => {
      live = false;
    };
  }, [organisation, project]);

  useEffect(() => {
    if (!available || typeof EventSource === "undefined") return;
    const source = new EventSource(assistantEvents(organisation, project));
    source.onmessage = (message) => {
      const event = JSON.parse(message.data) as AssistantEvent;
      if (event.type === "ready") {
        listening.current = true;
        waiting.current.splice(0).forEach((go) => go());
        return;
      }
      setSaid((before) => {
        const last = before[before.length - 1];
        const replying = last && last.who === "assistant" ? last : null;
        const rest = replying ? before.slice(0, -1) : before;
        const current: Said = replying ?? { who: "assistant", text: "", changes: [], working: true, about: asking.current ?? undefined };
        switch (event.type) {
          case "text":
            return [...rest, { ...current, text: current.text + event.text }];
          case "change":
            return [...rest, { ...current, changes: [...current.changes, event.change] }];
          case "done":
            return replying ? [...rest, { ...current, working: false }] : before;
          case "error":
            return [...rest, { ...current, working: false, trouble: event.message || say("assistant.did-not-answer") }];
          default:
            return before;
        }
      });
      if (event.type === "change") changed.current(asking.current ?? undefined);
      if (event.type === "done" || event.type === "error") {
        working.current = false;
        current.current = null;
        setBusy(false);
        asking.current = null;
      }
    };
    return () => {
      source.close();
      listening.current = false;
    };
  }, [available, organisation, project]);

  const working = useRef(false);
  const queued = useRef<{ id: number; text: string; selection: PreviewSelection }[]>([]);
  const current = useRef<number | null>(null);
  const title = useRef(pageTitle);
  title.current = pageTitle;

  useEffect(() => {
    if (!request) return;
    void ask(request.text, request.about, request.id);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [request?.id]);

  useImperativeHandle(control, () => ({
    undo: (id) => {
      const index = said.findIndex((entry) => entry.who === "assistant" && entry.id === id);
      if (index >= 0) void undoAll(index);
    },
    stop: (id) => void stop(id),
    seen: (id) => setSeen((now) => (now.includes(id) ? now : [...now, id])),
    focus: () => message.current?.focus(),
  }));

  useEffect(() => {
    if (!drafted) return;
    setDraft(drafted.text);
    message.current?.focus();
  }, [drafted?.id]);

  useEffect(() => {
    if (busy || queued.current.length === 0) return;
    const next = queued.current.shift()!;
    setWaitingTurn([...queued.current]);
    void ask(next.text, next.selection, next.id);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [busy]);

  async function send() {
    const text = draft.trim();
    if (!text || busy) return;
    setDraft("");
    onQuoteDropped?.();
    await ask(text, quote, Date.now());
  }

  async function ask(text: string, selection: PreviewSelection | null, id: number) {
    if (working.current) {
      if (selection) {
        queued.current.push({ id, text, selection });
        setWaitingTurn([...queued.current]);
      }
      return;
    }
    working.current = true;
    current.current = id;
    const about: Asked | null = selection ? { ...selection, page: title.current } : null;
    setTrouble("");
    setBusy(true);
    asking.current = selection;
    setSaid((before) => [
      ...before,
      { who: "person", id, text, about: about ?? undefined, changes: [] },
      { who: "assistant", id, text: "", changes: [], working: true, about: about ?? undefined },
    ]);
    try {
      await untilListening();
      await askAssistant(organisation, project, text, selection);
    } catch (error: unknown) {
      setTrouble(whatWentWrong(error, "assistant.did-not-answer"));
      working.current = false;
      setBusy(false);
      asking.current = null;
      setSaid((before) => before.slice(0, -1));
    }
  }

  function untilListening() {
    if (listening.current) return Promise.resolve();
    return new Promise<void>((resolve, reject) => {
      const timer = setTimeout(() => {
        waiting.current = waiting.current.filter((one) => one !== go);
        reject(new Error("not listening"));
      }, listenWithin);
      const go = () => {
        clearTimeout(timer);
        resolve();
      };
      waiting.current.push(go);
    });
  }

  async function stop(id: number) {
    if (queued.current.some((one) => one.id === id)) {
      queued.current = queued.current.filter((one) => one.id !== id);
      setWaitingTurn([...queued.current]);
      return;
    }
    if (current.current !== id) return;
    try {
      await stopAssistant(organisation, project);
      setSaid((before) => before.map((entry) => (entry.who === "assistant" && entry.id === id ? { ...entry, working: false, stopped: true } : entry)));
      working.current = false;
      current.current = null;
      setBusy(false);
      asking.current = null;
    } catch (error: unknown) {
      setTrouble(whatWentWrong(error));
    }
  }

  async function undoAll(index: number) {
    const entry = said[index];
    const first = entry?.changes.find((change) => !change.undone && change.before);
    if (!entry || !first) return;
    try {
      await goBackTo(organisation, project, first.before, false);
      setSaid((before) =>
        before.map((one, at) => (at === index ? { ...one, changes: one.changes.map((change) => ({ ...change, undone: true })) } : one)),
      );
      changed.current();
    } catch (error: unknown) {
      setTrouble(whatWentWrong(error));
    }
  }

  async function undo(index: number, change: AssistantChange) {
    try {
      await goBackTo(organisation, project, change.before, false);
      setSaid((before) =>
        before.map((entry, at) =>
          at === index
            ? { ...entry, changes: entry.changes.map((one) => (one.after === change.after ? { ...one, undone: true } : one)) }
            : entry,
        ),
      );
      changed.current();
    } catch (error: unknown) {
      setTrouble(whatWentWrong(error));
    }
  }

  return (
    <section aria-label={say("assistant.assistant")} className="flex h-full min-h-0 flex-col gap-3 p-4">
      <div className={onClose ? "flex items-center justify-between gap-2" : "sr-only"}>
        <h2 className="text-[15px] font-semibold">{say("assistant.assistant")}</h2>
        {onClose && (
          <Button variant="ghost" size="iconCompact" aria-label={say("assistant.close")} onClick={onClose}>
            <X aria-hidden="true" className="size-4" />
          </Button>
        )}
      </div>

      {available === false && (
        <p role="status" className="text-[15px] text-muted-foreground">
          {say("assistant.not-available")}
        </p>
      )}

      {available && (
        <>
          <ol aria-live="polite" className="flex min-h-0 flex-1 flex-col gap-3 overflow-y-auto">
            {said.length === 0 && <li className="text-[15px] text-muted-foreground">{say("assistant.ask-for-a-change")}</li>}
            {said.map((entry, index) => (
              <li
                key={index}
                className={
                  "flex flex-col gap-2 text-[15px] " +
                  (entry.who === "person" ? "ml-8 self-end rounded-xl bg-muted px-3 py-2" : "")
                }
              >
                {entry.who === "person" && entry.about && (
                  <RequestCard about={entry.about} onShow={() => entry.about && onShowPlace?.(entry.about)} />
                )}
                {entry.text && <p className="whitespace-pre-wrap">{entry.text}</p>}
                {entry.who === "person" && entry.about && (
                  <p className="flex items-center gap-1.5 text-[12px] text-subtle-foreground">
                    <Eye aria-hidden="true" className="size-3.5" />
                    {say("assistant.from-the-preview", { page: entry.about.page ?? "" })}
                  </p>
                )}
                {entry.who === "assistant" && entry.working && !entry.text && (
                  <p className="text-subtle-foreground">{say("assistant.working")}</p>
                )}
                {entry.stopped && <p className="text-subtle-foreground">{say("assistant.stopped")}</p>}
                {entry.trouble && (
                  <p role="alert" className="text-destructive-foreground">
                    {entry.trouble}
                  </p>
                )}
                {entry.changes.map((change) => (
                  <Chip
                    key={change.after}
                    change={change}
                    organisation={organisation}
                    project={project}
                    onUndo={() => undo(index, change)}
                    onShow={() => entry.about && onShowPlace?.(entry.about)}
                  />
                ))}
              </li>
            ))}
          </ol>

          {trouble && (
            <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
              {trouble}
            </p>
          )}

          {quote && (
            <figure className="flex items-start gap-2 rounded-md border border-border bg-muted px-3 py-2">
              <Quote aria-hidden="true" className="mt-0.5 size-4 shrink-0 text-subtle-foreground" />
              <figcaption className="sr-only">{say("assistant.about-this")}</figcaption>
              <blockquote aria-label={say("assistant.about-this")} className="flex-1 text-[13px] line-clamp-4">
                {quote.text}
              </blockquote>
              <Button variant="ghost" size="iconCompact" aria-label={say("assistant.leave-out-the-quote")} onClick={onQuoteDropped}>
                <X aria-hidden="true" className="size-4" />
              </Button>
            </figure>
          )}

          <form
            className="flex items-end gap-2"
            onSubmit={(event) => {
              event.preventDefault();
              void send();
            }}
          >
            <label className="flex flex-1 flex-col gap-1.5">
              <span className="sr-only">{say("assistant.your-message")}</span>
              <textarea
                ref={message}
                value={draft}
                onChange={(event) => setDraft(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key === "Enter" && !event.shiftKey) {
                    event.preventDefault();
                    void send();
                  }
                }}
                rows={2}
                placeholder={quote ? say("assistant.what-should-change") : say("assistant.ask-or-select")}
                className="min-h-11 resize-none rounded-md border border-input bg-card px-3 py-2 text-base sm:text-[15px] focus-visible:border-primary focus-visible:ring-[3px] focus-visible:ring-primary-soft"
              />
            </label>
            <Button type="submit" size="icon" aria-label={say("assistant.send")} disabled={busy || !draft.trim()}>
              <Send aria-hidden="true" className="size-4" />
            </Button>
          </form>
        </>
      )}
    </section>
  );
}

function RequestCard({ about, onShow }: { about: Asked; onShow: () => void }) {
  const before = about.before.slice(-60);
  const after = about.after.slice(0, 60);
  return (
    <button
      type="button"
      onClick={onShow}
      aria-label={say("assistant.show-the-place", { text: about.text.slice(0, 80) })}
      className="rounded-lg border border-border bg-card px-3 py-2 text-left font-serif text-[12px] leading-relaxed text-muted-foreground line-clamp-3 hover:border-primary"
    >
      {before.length < about.before.length && "…"}
      {before}
      <mark className="rounded-sm bg-primary-soft text-foreground">{about.text}</mark>
      {after}
      {after.length < about.after.length && "…"}
    </button>
  );
}

function Chip({
  change,
  organisation,
  project,
  onUndo,
  onShow,
}: {
  change: AssistantChange & { undone?: boolean };
  organisation: string;
  project: string;
  onUndo: () => void;
  onShow: () => void;
}) {
  const navigate = useNavigate();
  const action = "min-h-11 rounded-md px-1 text-[13px] font-medium hover:underline sm:min-h-9";
  return (
    <div className="flex flex-col gap-1 rounded-lg border border-border bg-card px-3 py-2">
      <p className="flex items-center gap-2 text-[13px] font-medium">
        <Check aria-hidden="true" className="size-4 shrink-0 text-success-foreground" />
        {change.undone ? say("assistant.change-undone", { title: change.title }) : say("assistant.changed", { title: change.title })}
      </p>
      <div className="flex flex-wrap items-center gap-3 pl-6">
        <button
          type="button"
          className={action + " text-primary"}
          onClick={() => {
            navigate(`/o/${organisation}/p/${project}/content/${change.page}`);
            onShow();
          }}
        >
          {say("assistant.show-changes")}
        </button>
        {!change.undone && change.before && (
          <button type="button" className={action + " text-muted-foreground"} onClick={onUndo}>
            {say("assistant.undo")}
          </button>
        )}
      </div>
    </div>
  );
}

function requestStates(said: Said[], waiting: { id: number; selection: PreviewSelection }[], seen: number[]): RequestState[] {
  const states: RequestState[] = [];
  said.forEach((entry, index) => {
    if (entry.who !== "assistant" || entry.id === undefined || !entry.about || entry.stopped) return;
    const anchor = { id: entry.id, before: entry.about.before, after: entry.about.after, text: entry.about.text };
    if (entry.working) return states.push({ ...anchor, state: "working" });
    if (entry.trouble) return;
    const live = entry.changes.filter((change) => !change.undone);
    if (live.length > 0) {
      if (!seen.includes(entry.id)) states.push({ ...anchor, state: "done" });
      return;
    }
    const answeredLater = said.slice(index + 1).some((later) => later.who === "person");
    if (entry.changes.length === 0 && entry.text.trim().endsWith("?") && !answeredLater) states.push({ ...anchor, state: "question" });
  });
  for (const one of waiting) {
    states.push({ id: one.id, before: one.selection.before, after: one.selection.after, text: one.selection.text, state: "working" });
  }
  return states;
}
