import { useCallback, useEffect, useState } from "react";

import type { EditorLayout } from "./EditorShell";

export const slotWidth = 420;
export const assistantWidth = 340;

export type Room = {
  shown: EditorLayout;
  assistant: "column" | "instead-of-form" | "alone" | "closed";
  splitFits: boolean;
};

export function arrange(width: number, chosen: EditorLayout, talking: boolean): Room {
  const splitFits = width >= 2 * slotWidth;
  const shown: EditorLayout = chosen === "split" && !splitFits ? "form" : chosen;
  if (!talking) return { shown, assistant: "closed", splitFits };
  const slots = shown === "split" ? 2 : 1;
  if (width >= slots * slotWidth + assistantWidth) return { shown, assistant: "column", splitFits };
  if (width >= slotWidth + assistantWidth) {
    return { shown: "preview", assistant: "instead-of-form", splitFits };
  }
  return { shown, assistant: "alone", splitFits };
}

export function useEditorRoom() {
  const [width, setWidth] = useState(Number.POSITIVE_INFINITY);
  const [node, setNode] = useState<HTMLElement | null>(null);
  const ref = useCallback((element: HTMLElement | null) => setNode(element), []);

  useEffect(() => {
    if (!node || typeof ResizeObserver === "undefined") return;
    const watch = new ResizeObserver((entries) => {
      const seen = entries[0]?.contentRect.width;
      if (seen !== undefined && seen > 0) setWidth(seen);
    });
    watch.observe(node);
    return () => watch.disconnect();
  }, [node]);

  return { ref, width };
}
