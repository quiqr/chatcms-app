import { beforeEach, describe, expect, it, vi } from "vitest";

import {
  bodyPath,
  fits,
  keep,
  kept,
  none,
  planRedo,
  planTo,
  planUndo,
  recallSteps,
  record,
  same,
  undoKey,
  type Current,
  type Step,
  type Steps,
} from "../undo";

function step(values: Record<string, [unknown, unknown]>, body?: [string, string]): Step {
  return {
    at: "2026-11-17T14:00:00Z",
    values: Object.fromEntries(Object.entries(values).map(([path, [before, after]]) => [path, { before, after }])),
    body: body && { before: body[0], after: body[1] },
  };
}

function page(values: Record<string, unknown>, body = ""): Current {
  return { value: (path) => values[path], body };
}

beforeEach(() => {
  window.sessionStorage.clear();
});

describe("recording", () => {
  it("adds one step per record and empties redo", () => {
    const one = record(none, step({ title: ["a", "b"] }));
    expect(one.done).toHaveLength(1);
    const undone: Steps = { done: [], undone: one.done };
    expect(record(undone, step({ summary: ["x", "y"] })).undone).toHaveLength(0);
  });

  it("keeps at most a hundred, dropping the oldest", () => {
    let steps = none;
    for (let index = 0; index < kept + 5; index++) {
      steps = record(steps, step({ title: [String(index), String(index + 1)] }));
    }
    expect(steps.done).toHaveLength(kept);
    expect(steps.done[0]?.values.title?.before).toBe("5");
  });

  it("records nothing for an empty step", () => {
    expect(record(none, step({}))).toBe(none);
  });
});

describe("undo and redo", () => {
  const steps = record(none, step({ title: ["Old", "New"], summary: ["Long", ""] }, ["text", "more text"]));

  it("writes the values from before", () => {
    const plan = planUndo(steps, page({ title: "New", summary: "" }, "more text"))!;
    expect(plan.values).toEqual({ title: "Old", summary: "Long" });
    expect(plan.body).toBe("text");
    expect(plan.leftAlone).toEqual([]);
    expect(plan.steps.done).toHaveLength(0);
    expect(plan.steps.undone).toHaveLength(1);
  });

  it("writes the values from after when redone", () => {
    const undone = planUndo(steps, page({ title: "New", summary: "" }, "more text"))!.steps;
    const plan = planRedo(undone, page({ title: "Old", summary: "Long" }, "text"))!;
    expect(plan.values).toEqual({ title: "New", summary: "" });
    expect(plan.body).toBe("more text");
    expect(plan.steps).toEqual(steps);
  });

  it("leaves a field alone that changed since, and names it", () => {
    const plan = planUndo(steps, page({ title: "New", summary: "Somebody else" }, "changed"))!;
    expect(plan.values).toEqual({ title: "Old" });
    expect(plan.body).toBeUndefined();
    expect(plan.leftAlone).toEqual(["summary", bodyPath]);
  });

  it("has nothing to plan at either end", () => {
    expect(planUndo(none, page({}))).toBeNull();
    expect(planRedo(steps, page({}))).toBeNull();
  });
});

describe("going to a point", () => {
  let steps = none;
  steps = record(steps, step({ title: ["A", "B"] }));
  steps = record(steps, step({ summary: ["x", "y"] }));
  steps = record(steps, step({ title: ["B", "C"] }));
  const now = page({ title: "C", summary: "y" });

  it("folds several undos into one write", () => {
    const plan = planTo(steps, 0, now)!;
    expect(plan.values).toEqual({ title: "A", summary: "x" });
    expect(plan.steps.done).toHaveLength(0);
    expect(plan.steps.undone).toHaveLength(3);
  });

  it("goes forward again from there", () => {
    const back = planTo(steps, 0, now)!;
    const plan = planTo(back.steps, 2, page({ title: "A", summary: "x" }))!;
    expect(plan.values).toEqual({ title: "B", summary: "y" });
    expect(plan.steps.done).toHaveLength(2);
    expect(plan.steps.undone).toHaveLength(1);
  });

  it("has nothing to do at the current point or past the ends", () => {
    expect(planTo(steps, 3, now)).toBeNull();
    expect(planTo(steps, 4, now)).toBeNull();
    expect(planTo(steps, -1, now)).toBeNull();
  });
});

describe("whether the steps still fit the page", () => {
  let steps = none;
  steps = record(steps, step({ title: ["A", "B"] }, ["one", "two"]));
  steps = record(steps, step({ title: ["B", "C"] }));

  it("fits the page where the steps left it", () => {
    expect(fits(steps, page({ title: "C", other: "anything" }, "two"))).toBe(true);
  });

  it("fits after an undo too", () => {
    const back = planTo(steps, 1, page({ title: "C" }, "two"))!.steps;
    expect(fits(back, page({ title: "B" }, "two"))).toBe(true);
  });

  it("does not fit a page changed elsewhere", () => {
    expect(fits(steps, page({ title: "D" }, "two"))).toBe(false);
    expect(fits(steps, page({ title: "C" }, "three"))).toBe(false);
  });
});

describe("keeping the steps for the tab", () => {
  const steps = record(none, step({ title: ["a", "b"] }));

  it("survives a reload, per page", () => {
    keep(undoKey("person-1", "korenaar", "home"), steps);
    expect(recallSteps(undoKey("person-1", "korenaar", "home"))).toEqual(steps);
    expect(recallSteps(undoKey("person-1", "korenaar", "about"))).toEqual(none);
  });

  it("works without storage", () => {
    const failing = vi.spyOn(Storage.prototype, "getItem").mockImplementation(() => {
      throw new Error("blocked");
    });
    vi.spyOn(Storage.prototype, "setItem").mockImplementation(() => {
      throw new Error("blocked");
    });
    expect(() => keep("k", steps)).not.toThrow();
    expect(recallSteps("k")).toEqual(none);
    failing.mockRestore();
    vi.restoreAllMocks();
  });

  it("ignores something that is not steps", () => {
    window.sessionStorage.setItem("k", "{\"nonsense\":true}");
    expect(recallSteps("k")).toEqual(none);
  });
});

describe("same", () => {
  it("compares values deeply", () => {
    expect(same([{ a: 1 }], [{ a: 1 }])).toBe(true);
    expect(same([{ a: 1 }], [{ a: 2 }])).toBe(false);
    expect(same(undefined, null)).toBe(true);
    expect(same("", undefined)).toBe(false);
    expect(same([], {})).toBe(false);
  });
});
