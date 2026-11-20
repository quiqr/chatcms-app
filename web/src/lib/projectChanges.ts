import { useEffect } from "react";

import { projectChangedEvent } from "./api";

const settleAfter = 2_000;

export function useProjectChanges(organisation: string, project: string, again: () => void) {
  useEffect(() => {
    if (!organisation || !project) return;
    const timers = new Set<ReturnType<typeof setTimeout>>();
    const changed = (event: Event) => {
      const { detail } = event as CustomEvent<{ organisation: string; project: string }>;
      if (detail.organisation !== organisation || detail.project !== project) return;
      again();
      const later = setTimeout(() => {
        timers.delete(later);
        again();
      }, settleAfter);
      timers.add(later);
    };
    const back = () => {
      if (document.visibilityState === "visible") again();
    };
    window.addEventListener(projectChangedEvent, changed);
    document.addEventListener("visibilitychange", back);
    return () => {
      window.removeEventListener(projectChangedEvent, changed);
      document.removeEventListener("visibilitychange", back);
      timers.forEach(clearTimeout);
    };
  }, [organisation, project, again]);
}
