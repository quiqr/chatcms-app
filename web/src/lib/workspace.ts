import { useEffect } from "react";

import { openWorkspace } from "./tenancy";

export function useStartWorkspace(organisation: string, project: string) {
  useEffect(() => {
    if (!organisation || !project) return;
    openWorkspace(organisation, project).catch(() => undefined);
  }, [organisation, project]);
}
