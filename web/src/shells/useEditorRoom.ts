import { useCallback, useEffect, useState } from "react";

import type { EditorLayout } from "./EditorShell";

export const slotWidth = 420;
export const chatWidth = 400;

export type Room = {
  shown: EditorLayout;
  chat: "column" | "alone" | "closed";
  splitFits: boolean;
};

export function arrange(width: number, chosen: EditorLayout): Room {
  const splitFits = width >= 2 * slotWidth;
  const shown: EditorLayout = chosen === "split" && !splitFits ? "form" : chosen;
  if (shown !== "chat") return { shown, chat: "closed", splitFits };
  return { shown, chat: width >= chatWidth + slotWidth ? "column" : "alone", splitFits };
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
