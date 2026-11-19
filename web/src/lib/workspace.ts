import { useEffect, useState } from "react";

import type { Key } from "./say";
import { openWorkspace, type WorkspaceStep } from "./tenancy";

export function useStartWorkspace(organisation: string, project: string) {
  useEffect(() => {
    if (!organisation || !project) return;
    openWorkspace(organisation, project).catch(() => undefined);
  }, [organisation, project]);
}

export const stepNames: Record<WorkspaceStep["step"], Key> = {
  "starting-workspace": "preview.step-starting-workspace",
  "fetching-site": "preview.step-fetching-site",
  "building-preview": "preview.step-building-preview",
};

export const stepStates: Record<WorkspaceStep["state"], Key> = {
  waiting: "preview.step-waiting",
  running: "preview.step-running",
  done: "preview.step-done",
  failed: "preview.step-failed",
};

const restartedEvent = "chatcms:workspace-restarted";

export function announceRestart() {
  window.dispatchEvent(new Event(restartedEvent));
}

export function useRestarted() {
  const [times, setTimes] = useState(0);
  useEffect(() => {
    const count = () => setTimes((n) => n + 1);
    window.addEventListener(restartedEvent, count);
    return () => window.removeEventListener(restartedEvent, count);
  }, []);
  return times;
}
