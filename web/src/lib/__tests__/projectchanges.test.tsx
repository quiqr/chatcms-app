import { act, render } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { post, tellProjectChanged } from "../api";
import { useProjectChanges } from "../projectChanges";

function Watching({ again }: { again: () => void }) {
  useProjectChanges("technative", "korenaar", again);
  return null;
}

afterEach(() => {
  vi.useRealTimers();
  vi.unstubAllGlobals();
});

describe("a project that changed is read again", () => {
  it("at once after a write, and again once the save has settled", () => {
    vi.useFakeTimers();
    const again = vi.fn();
    render(<Watching again={again} />);

    act(() => tellProjectChanged("/api/organisations/technative/projects/korenaar/pages/home"));
    expect(again).toHaveBeenCalledTimes(1);

    act(() => vi.advanceTimersByTime(2_000));
    expect(again).toHaveBeenCalledTimes(2);
  });

  it("not for another project, nor for opening the preview or the chat", () => {
    const again = vi.fn();
    render(<Watching again={again} />);

    act(() => {
      tellProjectChanged("/api/organisations/technative/projects/other/pages/home");
      tellProjectChanged("/api/organisations/technative/projects/korenaar/workspace");
      tellProjectChanged("/api/organisations/technative/projects/korenaar/chat");
      tellProjectChanged("/api/administration/organisations");
    });
    expect(again).not.toHaveBeenCalled();
  });

  it("when a save to the project succeeds", async () => {
    document.cookie = "chatcms_request_token=t; path=/";
    vi.stubGlobal("fetch", vi.fn(() => Promise.resolve({ ok: true, status: 200, json: async () => ({}) })));
    const again = vi.fn();
    render(<Watching again={again} />);

    await act(() => post("/api/organisations/technative/projects/korenaar/pages/home"));
    expect(again).toHaveBeenCalledTimes(1);
    document.cookie = "chatcms_request_token=; max-age=0; path=/";
  });
});
