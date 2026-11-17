import { useCallback, useEffect, useRef, useState } from "react";
import { useParams } from "react-router";

import { Button } from "../components/ui/button";
import { BodyEditor } from "../fields/BodyEditor";
import { Draw, withPicking } from "../fields/Fields";
import { MediaPage } from "./MediaPage";
import { readPage, savePage, type Control, type Page } from "../lib/tenancy";
import { say } from "../lib/say";
import { whatWentWrong } from "../lib/api";
import { toast } from "../components/ui/toast";
import { typing } from "../lib/typing";
import { now } from "../lib/when";
import { useHere } from "../lib/whoami";
import {
  bodyPath,
  fits,
  keep,
  none,
  planRedo,
  planTo,
  planUndo,
  recallSteps,
  record,
  same,
  undoKey,
  type Current,
  type Plan,
  type Step,
  type Steps,
} from "./undo";
import { stepTitle, UndoControls } from "./UndoControls";

type Saving = "saved" | "saving" | "trouble";

const quietFor = 800;

/**
 * withValue replaces one control's value wherever it sits, so that what is
 * drawn always comes from one place rather than from a copy that can drift.
 */
function withValue(controls: Control[], path: string, value: unknown): Control[] {
  return controls.map((control) => {
    if (control.path === path) {
      return { ...control, value };
    }
    if (control.fields && control.type !== "accordion") {
      return { ...control, fields: withValue(control.fields, path, value) };
    }
    return control;
  });
}

function find(controls: Control[], path: string): Control | undefined {
  for (const control of controls) {
    if (control.path === path) return control;
    if (control.fields && control.type !== "accordion") {
      const found = find(control.fields, path);
      if (found) return found;
    }
  }
  return undefined;
}

function currentOf(page: Page): Current {
  return { value: (path) => find(page.fields, path)?.value, body: page.body };
}

function Status({ state }: { state: Saving }) {
  const said = {
    saved: say("sync-indicator.all-changes-saved"),
    saving: "Saving…",
    trouble: say("page-editor.not-saved-yet-we-will-keep-trying"),
  }[state];

  // Named, because the top bar carries a status too and two live regions
  // saying the same words are two things a screen reader cannot tell apart.
  return (
    <p
      aria-live="polite"
      aria-label={say("page-editor.whether-this-page-is-saved")}
      className="text-sm text-subtle-foreground"
    >
      {said}
    </p>
  );
}

export function PageEditor({
  organisation,
  onSaved,
  onOpened,
  headed = true,
}: {
  organisation: string;
  /** onSaved is how the preview is told to show the site again. */
  onSaved?: () => void;
  /** onOpened hands the page's name to whoever draws the bar above it. */
  onOpened?: (opened: { title: string; section: string }) => void;
  /** headed is false when the bar above already says which page this is. */
  headed?: boolean;
}) {
  const { project = "", "*": page = "" } = useParams();
  const [loaded, setLoaded] = useState<Page | null>(null);
  const [trouble, setTrouble] = useState("");
  const [state, setState] = useState<Saving>("saved");
  // choosing holds what to do with the picture somebody picks, which is how
  // a field reaches the library without the library knowing about fields.
  const [choosing, setChoosing] = useState<((path: string) => void) | null>(null);

  // waiting holds what has changed but has not been written yet, so that
  // several quick edits become one save rather than one each.
  const waiting = useRef<{ values: Record<string, unknown>; body?: string }>({ values: {} });
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const befores = useRef<{ values: Record<string, unknown>; body?: string }>({ values: {} });
  const shown = useRef<Page | null>(null);
  const inFlight = useRef<Promise<unknown>>(Promise.resolve());
  const { who } = useHere();
  const key = undoKey(who.person.id, project, page);
  const [steps, setSteps] = useState<Steps>(none);
  const stepsRef = useRef<Steps>(none);

  const [waitingFor, setWaitingFor] = useState<Step | null>(null);

  function show(next: Page) {
    shown.current = next;
    setLoaded(next);
  }

  function noteWaiting() {
    const values = Object.fromEntries(Object.keys(waiting.current.values).map((path) => [path, { before: null, after: null }]));
    const body = waiting.current.body === undefined ? undefined : { before: null, after: null };
    setWaitingFor({ at: "", values, body });
  }

  useEffect(() => {
    let cancelled = false;
    readPage(organisation, project, page)
      .then((answer) => {
        if (cancelled) return;
        const stored = recallSteps(key);
        const start = fits(stored, currentOf(answer)) ? stored : none;
        stepsRef.current = start;
        setSteps(start);
        keep(key, start);
        show(answer);
        onOpened?.({ title: answer.title, section: answer.section });
      })
      .catch((error: unknown) => {
        if (!cancelled) setTrouble(whatWentWrong(error, "page-editor.we-could-not-open-this-page"));
      });
    return () => {
      cancelled = true;
    };
  }, [organisation, project, page]);

  const save = useCallback(async () => {
    const pending = waiting.current;
    const before = befores.current;
    waiting.current = { values: {} };
    befores.current = { values: {} };
    setWaitingFor(null);
    if (Object.keys(pending.values).length === 0 && pending.body === undefined) {
      return;
    }

    setState("saving");
    try {
      await savePage(organisation, project, page, {
        values: pending.values,
        body: pending.body,
      });
      setState("saved");
      onSaved?.();
      const step: Step = { at: now().toISOString(), values: {} };
      for (const [path, after] of Object.entries(pending.values)) {
        if (!same(before.values[path], after)) step.values[path] = { before: before.values[path], after };
      }
      if (pending.body !== undefined && !same(before.body, pending.body)) {
        step.body = { before: before.body ?? "", after: pending.body };
      }
      settle(record(stepsRef.current, step));
    } catch (error: unknown) {
      setState("trouble");
      setTrouble(whatWentWrong(error, "page-editor.we-could-not-save-that"));
    }
  }, [organisation, project, page, onSaved]);

  const write = useCallback(() => {
    const run = inFlight.current.then(save);
    inFlight.current = run.catch(() => undefined);
    return run;
  }, [save]);

  const later = useCallback(() => {
    if (timer.current) clearTimeout(timer.current);
    timer.current = setTimeout(() => void write(), quietFor);
  }, [write]);

  useEffect(() => {
    return () => {
      if (timer.current) clearTimeout(timer.current);
      void write();
    };
  }, [write]);

  function settle(next: Steps) {
    stepsRef.current = next;
    setSteps(next);
    keep(key, next);
  }

  function title(path: string) {
    if (path === bodyPath) return say("page-editor.page-text");
    return (shown.current && find(shown.current.fields, path)?.title) || path;
  }

  function tell(leftAlone: string[]) {
    const [only] = leftAlone;
    if (leftAlone.length === 1 && only !== undefined) toast(say("undo.left-alone-one", { field: title(only) }));
    if (leftAlone.length > 1) toast(say("undo.left-alone-many", { fields: leftAlone.map(title).join(", ") }));
  }

  async function carryOut(plan: Plan) {
    const nothing = Object.keys(plan.values).length === 0 && plan.body === undefined;
    if (!nothing) {
      setState("saving");
      try {
        await savePage(organisation, project, page, { values: plan.values, body: plan.body });
      } catch (error: unknown) {
        setState("trouble");
        setTrouble(whatWentWrong(error, "page-editor.we-could-not-save-that"));
        return;
      }
      const held = shown.current;
      if (held) {
        show({
          ...held,
          fields: Object.entries(plan.values).reduce((fields, [path, value]) => withValue(fields, path, value), held.fields),
          body: plan.body ?? held.body,
        });
      }
      setState("saved");
      onSaved?.();
    }
    settle(plan.steps);
    tell(plan.leftAlone);
  }

  function go(planner: (steps: Steps, current: Current) => Plan | null) {
    if (timer.current) clearTimeout(timer.current);
    const run = inFlight.current.then(save).then(async () => {
      const held = shown.current;
      if (!held) return;
      const plan = planner(stepsRef.current, currentOf(held));
      if (plan) await carryOut(plan);
    });
    inFlight.current = run.catch(() => undefined);
    return run;
  }

  const undo = () => go(planUndo);
  const redo = () => go(planRedo);
  const goTo = (point: number) => go((steps, current) => planTo(steps, point, current));

  useEffect(() => {
    function keys(event: KeyboardEvent) {
      if (!(event.ctrlKey || event.metaKey) || event.altKey || typing(event.target)) return;
      const letter = event.key.toLowerCase();
      if (letter === "z" && !event.shiftKey) {
        event.preventDefault();
        void undo();
      } else if ((letter === "z" && event.shiftKey) || (letter === "y" && !event.shiftKey)) {
        event.preventDefault();
        void redo();
      }
    }
    window.addEventListener("keydown", keys);
    return () => window.removeEventListener("keydown", keys);
  });

  function change(path: string, value: unknown) {
    if (!(path in befores.current.values) && !(path in waiting.current.values) && shown.current) {
      befores.current.values[path] = find(shown.current.fields, path)?.value;
    }
    const held = shown.current;
    if (held) show({ ...held, fields: withValue(held.fields, path, value) });
    waiting.current.values[path] = value;
    noteWaiting();
    later();
  }

  function changeBody(body: string) {
    if (befores.current.body === undefined && waiting.current.body === undefined && shown.current) {
      befores.current.body = shown.current.body;
    }
    const held = shown.current;
    if (held) show({ ...held, body });
    waiting.current.body = body;
    noteWaiting();
    later();
  }

  if (trouble && !loaded) {
    return <p className="text-[15px]">{trouble}</p>;
  }
  if (!loaded) {
    return <p className="text-[15px] text-subtle-foreground">{say("page-editor.opening-this-page")}</p>;
  }

  withPicking((choose) => setChoosing(() => choose));
  const draft = loaded.fields.find((control) => control.path === "draft" && control.type === "boolean");

  if (choosing) {
    return (
      <section className="flex flex-col gap-4">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <h2 className="text-lg font-semibold">{say("common.choose-a-picture")}</h2>
          <Button variant="ghost" onClick={() => setChoosing(null)}>
            {say("page-editor.back-to-the-page")}
          </Button>
        </div>
        <MediaPage
          picking
          onUse={(image) => {
            choosing("/" + image.path.replace(/^static\//, ""));
            setChoosing(null);
          }}
        />
      </section>
    );
  }

  return (
    <section className="flex max-w-[760px] flex-col gap-5">
      <header className="flex flex-wrap items-baseline justify-between gap-2">
        {headed ? (
          <div className="flex flex-col gap-0.5">
            <p className="text-sm text-subtle-foreground">{loaded.section}</p>
            <h2 className="text-lg font-semibold">{loaded.title}</h2>
          </div>
        ) : (
          <span />
        )}
        <div className="flex items-center gap-2">
          <UndoControls
            steps={steps}
            waiting={waitingFor}
            named={(step) => stepTitle(step, title)}
            onUndo={() => void undo()}
            onRedo={() => void redo()}
            onTo={(point) => void goTo(point)}
          />
          <Status state={state} />
        </div>
      </header>

      {draft && (
        <label className="flex items-start justify-between gap-4 rounded-lg border border-border bg-card p-3">
          <span className="flex flex-col gap-0.5">
            <span className="text-[15px] font-medium">{say("editor.show-on-website")}</span>
            <span className="text-[13px] text-subtle-foreground">{say("editor.show-on-website-hint")}</span>
          </span>
          <input
            type="checkbox"
            role="switch"
            checked={draft.value !== true}
            onChange={(event) => change(draft.path, !event.target.checked)}
            className="mt-1 size-5 shrink-0"
          />
        </label>
      )}

      {loaded.fields
        .filter((control) => control !== draft)
        .map((control) => (
          <Draw key={control.path} control={control} onChange={change} />
        ))}

      {loaded.hasBody ? (
        <BodyEditor value={loaded.body} onChange={changeBody} />
      ) : (
        loaded.body.trim() !== "" && (
          <div className="flex flex-col gap-1.5">
            <span className="text-sm font-medium">{say("page-editor.page-text")}</span>
            <p className="whitespace-pre-wrap rounded-md bg-muted p-3 text-[15px] text-subtle-foreground">
              {loaded.body}
            </p>
            <p className="text-[13px] text-subtle-foreground">
              {say("page-editor.this-page-s-text-is-not")}
            </p>
          </div>
        )
      )}
    </section>
  );
}
