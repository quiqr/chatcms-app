import { useEffect, useState } from "react";

import { Button } from "../components/ui/button";
import { Field } from "../components/ui/field";
import { readAIUse, setAILimits, type AIUse } from "../lib/administration";
import { whatWentWrong } from "../lib/api";
import { aNumber } from "../lib/numbers";
import { say } from "../lib/say";

/**
 * chatcms-audience: implementer
 */
export function AIUseSection({ organisation }: { organisation: string }) {
  const [use, setUse] = useState<AIUse | null>(null);
  const [requests, setRequests] = useState<number | null>(null);
  const [tokens, setTokens] = useState<number | null>(null);
  const [busy, setBusy] = useState(false);
  const [saved, setSaved] = useState(false);
  const [trouble, setTrouble] = useState("");

  useEffect(() => {
    let live = true;
    readAIUse(organisation).then(
      (found) => {
        if (!live) return;
        setUse(found);
        setRequests(found.requestsPerMinute);
        setTokens(found.tokensPerMonth);
      },
      () => undefined,
    );
    return () => {
      live = false;
    };
  }, [organisation]);

  if (!use) return null;

  function save() {
    setBusy(true);
    setTrouble("");
    setSaved(false);
    const atLeastOne = (limit: number | null) => (limit === null ? null : Math.max(1, limit));
    setAILimits(organisation, atLeastOne(requests), atLeastOne(tokens))
      .then((found) => {
        setUse(found);
        setSaved(true);
      })
      .catch((error: unknown) => setTrouble(whatWentWrong(error)))
      .finally(() => setBusy(false));
  }

  return (
    <section aria-labelledby="ai-use" className="flex flex-col gap-4">
      <h2 id="ai-use" className="text-[15px] font-semibold">
        {say("platform.ai-use")}
      </h2>
      <p className="text-[15px] text-muted-foreground">
        {say("platform.ai-this-month", {
          requests: aNumber(use.requestsThisMonth),
          tokens: aNumber(use.tokensThisMonth),
        })}
      </p>

      {trouble && (
        <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          {trouble}
        </p>
      )}

      <Limit label={say("platform.ai-limit-requests")} value={requests} fallback={30} onChange={setRequests} />
      <Limit label={say("platform.ai-limit-tokens")} value={tokens} fallback={1_000_000} onChange={setTokens} />

      <p className="text-[13px] text-subtle-foreground">{say("platform.ai-limits-explain")}</p>

      <div className="flex flex-wrap items-center gap-3">
        <Button onClick={save} disabled={busy}>
          {say("platform.save-ai-limits")}
        </Button>
        {saved && (
          <p role="status" className="text-[15px] text-muted-foreground">
            {say("common.saved")}
          </p>
        )}
      </div>
    </section>
  );
}

function Limit({
  label,
  value,
  fallback,
  onChange,
}: {
  label: string;
  value: number | null;
  fallback: number;
  onChange: (next: number | null) => void;
}) {
  return (
    <div className="flex flex-col gap-2">
      <label className="flex min-h-11 items-center gap-2 text-[15px]">
        {/* chatcms-target: the label around this one carries the hit area */}
        <input
          type="checkbox"
          checked={value !== null}
          onChange={(event) => onChange(event.target.checked ? fallback : null)}
          className="size-4"
        />
        {label}
      </label>
      {value !== null && (
        <Field
          label={label}
          type="number"
          min={1}
          value={value}
          onChange={(event) => onChange(Number(event.target.value) || 0)}
          className="max-w-48 tabular-nums"
        />
      )}
    </div>
  );
}
