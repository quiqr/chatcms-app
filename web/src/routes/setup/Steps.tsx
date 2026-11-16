import { Check } from "lucide-react";

import { say } from "../../lib/say";

export type Step = { key: string; label: string; done: boolean };

export function Steps({ steps, current }: { steps: Step[]; current: string }) {
  return (
    <ol className="flex flex-col gap-1">
      {steps.map((step) => (
        <li
          key={step.key}
          aria-current={step.key === current ? "step" : undefined}
          className={
            "flex min-h-11 items-center gap-2.5 rounded-lg px-3 text-[15px] " +
            (step.key === current ? "bg-muted font-medium" : "text-muted-foreground")
          }
        >
          <span
            aria-hidden="true"
            className={
              "inline-flex size-5 items-center justify-center rounded-full text-xs " +
              (step.done ? "bg-success-soft text-success-foreground" : "bg-muted")
            }
          >
            {step.done ? <Check className="size-3" /> : ""}
          </span>
          {step.label}
          {step.done && <span className="sr-only">{say("onboarding.done")}</span>}
        </li>
      ))}
    </ol>
  );
}
