import { Send, Undo2, X } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router";

import { Button } from "../components/ui/button";
import { whatWentWrong } from "../lib/api";
import { say } from "../lib/say";
import {
  askAssistant,
  assistantEvents,
  goBackTo,
  readAssistant,
  type AssistantChange,
  type AssistantEvent,
} from "../lib/tenancy";

type Said = {
  who: "person" | "assistant";
  text: string;
  changes: (AssistantChange & { undone?: boolean })[];
  working?: boolean;
  trouble?: string;
};

export function ChatPanel({
  organisation,
  project,
  onChanged,
  onClose,
}: {
  organisation: string;
  project: string;
  onChanged: () => void;
  onClose?: () => void;
}) {
  const [available, setAvailable] = useState<boolean | null>(null);
  const [said, setSaid] = useState<Said[]>([]);
  const [draft, setDraft] = useState("");
  const [busy, setBusy] = useState(false);
  const [trouble, setTrouble] = useState("");
  const changed = useRef(onChanged);
  changed.current = onChanged;

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
      setSaid((before) => {
        const last = before[before.length - 1];
        const replying = last && last.who === "assistant" ? last : null;
        const rest = replying ? before.slice(0, -1) : before;
        const current: Said = replying ?? { who: "assistant", text: "", changes: [], working: true };
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
      if (event.type === "change") changed.current();
      if (event.type === "done" || event.type === "error") setBusy(false);
    };
    return () => source.close();
  }, [available, organisation, project]);

  async function send() {
    const text = draft.trim();
    if (!text || busy) return;
    setDraft("");
    setTrouble("");
    setBusy(true);
    setSaid((before) => [
      ...before,
      { who: "person", text, changes: [] },
      { who: "assistant", text: "", changes: [], working: true },
    ]);
    try {
      await askAssistant(organisation, project, text);
    } catch (error: unknown) {
      setTrouble(whatWentWrong(error));
      setBusy(false);
      setSaid((before) => before.slice(0, -1));
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
      <div className="flex items-center justify-between gap-2">
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
                  "flex flex-col gap-2 rounded-lg px-3 py-2 text-[15px] " +
                  (entry.who === "person" ? "self-end bg-primary-soft text-primary-soft-foreground" : "bg-muted")
                }
              >
                {entry.text && <p className="whitespace-pre-wrap">{entry.text}</p>}
                {entry.who === "assistant" && entry.working && !entry.text && (
                  <p className="text-subtle-foreground">{say("assistant.working")}</p>
                )}
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
                value={draft}
                onChange={(event) => setDraft(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key === "Enter" && !event.shiftKey) {
                    event.preventDefault();
                    void send();
                  }
                }}
                rows={2}
                placeholder={say("assistant.ask-for-a-change")}
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

function Chip({
  change,
  organisation,
  project,
  onUndo,
}: {
  change: AssistantChange & { undone?: boolean };
  organisation: string;
  project: string;
  onUndo: () => void;
}) {
  const navigate = useNavigate();
  return (
    <div className="flex flex-wrap items-center gap-2 rounded-md border border-border bg-card px-2 py-1">
      <span className="flex-1 text-[13px]">
        {change.undone ? say("assistant.change-undone", { title: change.title }) : say("assistant.changed", { title: change.title })}
      </span>
      <Button variant="ghost" onClick={() => navigate(`/o/${organisation}/p/${project}/content/${change.page}`)}>
        {say("assistant.show")}
      </Button>
      {!change.undone && change.before && (
        <Button variant="ghost" onClick={onUndo}>
          <Undo2 aria-hidden="true" className="size-4" />
          {say("assistant.undo")}
        </Button>
      )}
    </div>
  );
}
