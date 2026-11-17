export type Change = { before: unknown; after: unknown };
export type Step = { at: string; values: Record<string, Change>; body?: Change };
export type Steps = { done: Step[]; undone: Step[] };
export type Current = { value: (path: string) => unknown; body: string };
export type Plan = { values: Record<string, unknown>; body?: string; leftAlone: string[]; steps: Steps };

export const bodyPath = "";
export const kept = 100;
export const none: Steps = { done: [], undone: [] };

export function same(a: unknown, b: unknown): boolean {
  if (a === b) return true;
  if (a == null && b == null) return true;
  if (a == null || b == null || typeof a !== "object" || typeof b !== "object") return false;
  if (Array.isArray(a) !== Array.isArray(b)) return false;
  const left = a as Record<string, unknown>;
  const right = b as Record<string, unknown>;
  const keys = new Set([...Object.keys(left), ...Object.keys(right)]);
  for (const key of keys) {
    if (!same(left[key], right[key])) return false;
  }
  return true;
}

export function isEmpty(step: Step) {
  return Object.keys(step.values).length === 0 && step.body === undefined;
}

export function record(steps: Steps, step: Step): Steps {
  if (isEmpty(step)) return steps;
  return { done: [...steps.done, step].slice(-kept), undone: [] };
}

export function touched(step: Step): string[] {
  const paths = Object.keys(step.values);
  return step.body ? [...paths, bodyPath] : paths;
}

function overlay(current: Current, plan: Plan): Current {
  return {
    value: (path) => (path in plan.values ? plan.values[path] : current.value(path)),
    body: plan.body ?? current.body,
  };
}

function apply(step: Step, current: Current, plan: Plan, from: "after" | "before") {
  const to = from === "after" ? "before" : "after";
  for (const [path, change] of Object.entries(step.values)) {
    if (same(current.value(path), change[from])) plan.values[path] = change[to];
    else if (!plan.leftAlone.includes(path)) plan.leftAlone.push(path);
  }
  if (step.body) {
    if (same(current.body, step.body[from])) plan.body = step.body[to] as string;
    else if (!plan.leftAlone.includes(bodyPath)) plan.leftAlone.push(bodyPath);
  }
}

export function planUndo(steps: Steps, current: Current): Plan | null {
  return planTo(steps, steps.done.length - 1, current);
}

export function planRedo(steps: Steps, current: Current): Plan | null {
  return planTo(steps, steps.done.length + 1, current);
}

export function planTo(steps: Steps, point: number, current: Current): Plan | null {
  const total = steps.done.length + steps.undone.length;
  if (point < 0 || point > total || point === steps.done.length) return null;

  const plan: Plan = { values: {}, leftAlone: [], steps };
  let done = [...steps.done];
  let undone = [...steps.undone];

  while (done.length > point) {
    const step = done[done.length - 1];
    if (!step) break;
    apply(step, overlay(current, plan), plan, "after");
    done = done.slice(0, -1);
    undone = [...undone, step];
  }
  while (done.length < point) {
    const step = undone[undone.length - 1];
    if (!step) break;
    apply(step, overlay(current, plan), plan, "before");
    undone = undone.slice(0, -1);
    done = [...done, step];
  }

  plan.steps = { done, undone };
  return plan;
}

function standing(steps: Steps): Map<string, unknown> {
  const held = new Map<string, unknown>();
  const note = (path: string, value: unknown) => {
    if (!held.has(path)) held.set(path, value);
  };
  for (const step of [...steps.done].reverse()) {
    for (const [path, change] of Object.entries(step.values)) note(path, change.after);
    if (step.body) note(bodyPath, step.body.after);
  }
  for (const step of [...steps.undone].reverse()) {
    for (const [path, change] of Object.entries(step.values)) note(path, change.before);
    if (step.body) note(bodyPath, step.body.before);
  }
  return held;
}

export function fits(steps: Steps, current: Current): boolean {
  for (const [path, value] of standing(steps)) {
    const now = path === bodyPath ? current.body : current.value(path);
    if (!same(now, value)) return false;
  }
  return true;
}

export function undoKey(person: string, project: string, page: string) {
  return `chatcms.undo.${person}.${project}.${page}`;
}

export function keep(key: string, steps: Steps) {
  try {
    window.sessionStorage.setItem(key, JSON.stringify(steps));
  } catch {
    return;
  }
}

export function recallSteps(key: string): Steps {
  try {
    const held = window.sessionStorage.getItem(key);
    if (!held) return none;
    const steps = JSON.parse(held) as Steps;
    return Array.isArray(steps.done) && Array.isArray(steps.undone) ? steps : none;
  } catch {
    return none;
  }
}
