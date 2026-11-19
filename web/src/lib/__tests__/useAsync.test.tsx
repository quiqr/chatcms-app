import { act, renderHook, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";

import { forgetKept, useAsync } from "../useAsync";

afterEach(() => forgetKept());

function later<T>() {
  let settle!: (value: T) => void;
  const promise = new Promise<T>((resolve) => {
    settle = resolve;
  });
  return { promise, settle };
}

describe("useAsync", () => {
  it("starts loading when nothing is kept", async () => {
    const { result } = renderHook(() => useAsync(() => Promise.resolve(1), [], "one"));
    expect(result.current[0].state).toBe("loading");
    await waitFor(() => expect(result.current[0]).toEqual({ state: "ready", value: 1 }));
  });

  it("starts with what it kept, and shows the fresh answer when it comes", async () => {
    const first = renderHook(() => useAsync(() => Promise.resolve("before"), [], "kept"));
    await waitFor(() => expect(first.result.current[0].state).toBe("ready"));
    first.unmount();

    const fresh = later<string>();
    const second = renderHook(() => useAsync(() => fresh.promise, [], "kept"));
    expect(second.result.current[0]).toEqual({ state: "ready", value: "before" });
    await act(async () => fresh.settle("after"));
    expect(second.result.current[0]).toEqual({ state: "ready", value: "after" });
  });

  it("keeps nothing without a name, and nothing after it is told to forget", async () => {
    const named = renderHook(() => useAsync(() => Promise.resolve(2), [], "two"));
    await waitFor(() => expect(named.result.current[0].state).toBe("ready"));
    named.unmount();
    forgetKept();
    expect(renderHook(() => useAsync(() => Promise.resolve(2), [], "two")).result.current[0].state).toBe("loading");
    expect(renderHook(() => useAsync(() => Promise.resolve(3), [])).result.current[0].state).toBe("loading");
  });

  it("uses what it kept for the new name when the keys change", async () => {
    const a = renderHook(() => useAsync(() => Promise.resolve("a"), [], "a"));
    await waitFor(() => expect(a.result.current[0].state).toBe("ready"));
    a.unmount();

    const pending = later<string>();
    const { result, rerender } = renderHook(({ name }) => useAsync(() => (name === "a" ? pending.promise : new Promise<string>(() => {})), [name], name), {
      initialProps: { name: "b" },
    });
    expect(result.current[0].state).toBe("loading");
    rerender({ name: "a" });
    expect(result.current[0]).toEqual({ state: "ready", value: "a" });
  });
});
