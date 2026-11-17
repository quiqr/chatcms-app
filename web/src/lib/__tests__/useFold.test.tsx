import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";

import { foldKey, useFold } from "../useFold";

beforeEach(() => {
  window.localStorage.clear();
});

describe("useFold", () => {
  it("starts unfolded", () => {
    const { result } = renderHook(() => useFold("site", "person-1"));
    expect(result.current[0]).toBe(false);
  });

  it("remembers a fold for the next mount, per person", () => {
    const first = renderHook(() => useFold("site", "person-1"));
    act(() => first.result.current[1](true));
    expect(first.result.current[0]).toBe(true);
    first.unmount();

    expect(renderHook(() => useFold("site", "person-1")).result.current[0]).toBe(true);
    expect(renderHook(() => useFold("site", "person-2")).result.current[0]).toBe(false);
    expect(window.localStorage.getItem(foldKey("site", "person-1"))).toBe("true");
  });

  it("starts from the fallback key when its own is absent", () => {
    window.localStorage.setItem("old", "true");
    expect(renderHook(() => useFold("content", "person-1", "old")).result.current[0]).toBe(true);
  });

  it("prefers its own key over the fallback", () => {
    window.localStorage.setItem("old", "true");
    window.localStorage.setItem(foldKey("content", "person-1"), "false");
    expect(renderHook(() => useFold("content", "person-1", "old")).result.current[0]).toBe(false);
  });
});
