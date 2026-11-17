import { Check, ListRestart, Redo2, Undo2 } from "lucide-react";

import { Button } from "../components/ui/button";
import { Menu, MenuContent, MenuItem, MenuLabel, MenuTrigger } from "../components/ui/menu";
import { cn } from "../lib/cn";
import { say } from "../lib/say";
import { when } from "../lib/when";
import { touched, type Step, type Steps } from "./undo";

export function UndoControls({
  steps,
  waiting,
  named,
  onUndo,
  onRedo,
  onTo,
}: {
  steps: Steps;
  waiting: Step | null;
  named: (step: Step) => string;
  onUndo: () => void;
  onRedo: () => void;
  onTo: (point: number) => void;
}) {
  const last = waiting ?? steps.done[steps.done.length - 1];
  const next = waiting ? undefined : steps.undone[steps.undone.length - 1];
  const current = steps.done.length;
  const rows = [
    ...[...steps.done, ...[...steps.undone].reverse()].map((step, index) => ({ point: index + 1, step })),
  ].reverse();

  return (
    <div className="flex items-center gap-0.5">
      <Button
        variant="ghost"
        size="icon"
        disabled={!last}
        aria-label={last ? say("undo.undo-what", { fields: named(last) }) : say("undo.undo")}
        title={say("undo.undo")}
        onClick={onUndo}
      >
        <Undo2 aria-hidden="true" className="size-4" />
      </Button>
      <Button
        variant="ghost"
        size="icon"
        disabled={!next}
        aria-label={next ? say("undo.redo-what", { fields: named(next) }) : say("undo.redo")}
        title={say("undo.redo")}
        onClick={onRedo}
      >
        <Redo2 aria-hidden="true" className="size-4" />
      </Button>
      <Menu>
        <MenuTrigger asChild>
          <Button
            variant="ghost"
            size="icon"
            disabled={rows.length === 0}
            aria-label={say("undo.history")}
            title={say("undo.history")}
          >
            <ListRestart aria-hidden="true" className="size-4" />
          </Button>
        </MenuTrigger>
        <MenuContent align="end" className="min-w-64">
          <MenuLabel>{say("undo.history")}</MenuLabel>
          {[...rows, { point: 0, step: null as Step | null }].map(({ point, step }) => {
            const here = point === current;
            const redoable = point > current;
            return (
              <MenuItem
                key={point}
                aria-current={here ? "step" : undefined}
                onSelect={() => {
                  if (!here) onTo(point);
                }}
                className={cn("flex items-center gap-2", redoable && "text-subtle-foreground")}
              >
                <Check aria-hidden="true" className={cn("size-4 shrink-0", !here && "invisible")} />
                <span className="min-w-0 flex-1 truncate">{step ? named(step) : say("undo.opened")}</span>
                {redoable && <span className="sr-only">{say("undo.can-redo")}</span>}
                {step && <span className="shrink-0 text-[13px] text-subtle-foreground">{when(step.at)}</span>}
              </MenuItem>
            );
          })}
        </MenuContent>
      </Menu>
    </div>
  );
}

export function stepTitle(step: Step, title: (path: string) => string) {
  return touched(step).map(title).join(", ");
}
