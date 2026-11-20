import { Box, RotateCw } from "lucide-react";
import { useEffect, useState } from "react";
import { Link } from "react-router";

import { Button } from "../components/ui/button";
import { Popover, PopoverContent, PopoverTrigger } from "../components/ui/popover";
import { cn } from "../lib/cn";
import { maybe, say, type Key } from "../lib/say";
import { logAddress, readAssistant, readWorkspace, restartWorkspace, type Workspace } from "../lib/tenancy";
import { when } from "../lib/when";
import { announceRestart, stepNames, stepStates, useRestarted } from "../lib/workspace";

type Shown = "ready" | "starting" | "sleeping" | "did-not-start";

const looks: Record<Shown, { label: Key; tone: string }> = {
  ready: { label: "workspace-indicator.ready", tone: "text-success" },
  starting: { label: "workspace-indicator.starting", tone: "text-info animate-pulse" },
  sleeping: { label: "workspace-indicator.sleeping", tone: "text-subtle-foreground" },
  "did-not-start": { label: "workspace-indicator.did-not-start", tone: "text-destructive" },
};

export function shownOf(workspace: Workspace | null): Shown {
  if (!workspace) return "sleeping";
  if (workspace.state === "warm") return "ready";
  if (workspace.state === "starting") return "starting";
  return workspace.reason ? "did-not-start" : "sleeping";
}

const everyWhileStarting = 1_000;
const everyOtherwise = 10_000;

function useStanding(organisation: string, project: string) {
  const [workspace, setWorkspace] = useState<Workspace | null>(null);
  const restarted = useRestarted();

  useEffect(() => {
    let live = true;
    let timer: ReturnType<typeof setTimeout> | undefined;
    const look = () => {
      if (document.visibilityState === "hidden") {
        timer = setTimeout(look, everyOtherwise);
        return;
      }
      readWorkspace(organisation, project).then(
        (next) => {
          if (!live) return;
          setWorkspace(next);
          timer = setTimeout(look, next.state === "starting" ? everyWhileStarting : everyOtherwise);
        },
        () => {
          if (live) timer = setTimeout(look, everyOtherwise);
        },
      );
    };
    look();
    return () => {
      live = false;
      clearTimeout(timer);
    };
  }, [organisation, project, restarted]);

  return [workspace, setWorkspace] as const;
}

export function WorkspaceIndicator({ organisation, project }: { organisation: string; project: string }) {
  const [workspace, setWorkspace] = useStanding(organisation, project);
  const [assistant, setAssistant] = useState<boolean | null>(null);
  const [restarting, setRestarting] = useState(false);
  const shown = shownOf(workspace);
  const { label, tone } = looks[shown];

  const opened = (open: boolean) => {
    if (!open) return;
    readAssistant(organisation, project).then(
      (answer) => setAssistant(answer.available),
      () => setAssistant(false),
    );
  };

  const startAgain = () => {
    setRestarting(true);
    restartWorkspace(organisation, project)
      .then((next) => {
        setWorkspace(next);
        announceRestart();
      })
      .catch(() => undefined)
      .finally(() => setRestarting(false));
  };

  return (
    <Popover onOpenChange={opened}>
      <PopoverTrigger asChild>
        <Button
          variant="ghost"
          size="iconCompact"
          aria-label={say("workspace-indicator.your-preview", { state: say(label) })}
          data-state-shown={shown}
          className="size-11 sm:size-9"
        >
          <Box aria-hidden="true" className={cn("size-[18px]", tone)} />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="flex w-72 flex-col gap-2 p-4 text-[13px]" aria-label={say(label)}>
        <p className="text-[15px] font-semibold">{say(label)}</p>
        {shown === "did-not-start" && workspace?.reason && <p>{maybe(workspace.reason) ?? ""}</p>}
        {workspace?.since && shown !== "sleeping" && (
          <p className="text-muted-foreground">{say("workspace-indicator.started", { when: when(workspace.since) })}</p>
        )}
        {shown === "sleeping" && <p className="text-muted-foreground">{say("workspace-indicator.wakes-when-opened")}</p>}
        {assistant !== null && (
          <p>{say(assistant ? "workspace-indicator.assistant-available" : "workspace-indicator.assistant-unavailable")}</p>
        )}
        {workspace?.backend && (
          <div className="mt-2 flex flex-col gap-2 border-t border-border pt-2">
            <p className="text-muted-foreground">
              {say("workspace-indicator.runs-on", {
                backend: maybe(`workspace-indicator.backend-${workspace.backend}`) ?? workspace.backend,
              })}
            </p>
            {workspace.steps && (
              <ul className="flex flex-col gap-1">
                {workspace.steps.map((step) => (
                  <li key={step.step} className="flex justify-between gap-4">
                    <span>{say(stepNames[step.step])}</span>
                    <span className="text-subtle-foreground">{say(stepStates[step.state])}</span>
                  </li>
                ))}
              </ul>
            )}
            {workspace.entry && (
              <Link to={logAddress(organisation, project, workspace.entry)} className="text-primary">
                {say("log.open-the-log")}
              </Link>
            )}
            <Button variant="outline" size="secondary" onClick={startAgain} disabled={restarting} type="button">
              <RotateCw aria-hidden="true" className={cn("size-4", restarting && "animate-spin")} />
              {say("workspace-indicator.start-again")}
            </Button>
          </div>
        )}
      </PopoverContent>
    </Popover>
  );
}
