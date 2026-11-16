import { useState } from "react";

import { Button } from "../../components/ui/button";
import { Checkbox, Field } from "../../components/ui/field";
import { whatWentWrong } from "../../lib/api";
import { say } from "../../lib/say";
import { hugoLabel, readHugo, sameHugo, setHugo } from "../../lib/templates";
import { useAsync } from "../../lib/useAsync";

export function HugoPanel({
  organisation,
  project,
  onTrouble,
}: {
  organisation: string;
  project: string;
  onTrouble: (said: string) => void;
}) {
  const [loaded, again] = useAsync(() => readHugo(organisation, project), [organisation, project]);
  const [version, setVersion] = useState<string | null>(null);
  const [extended, setExtended] = useState<boolean | null>(null);
  const [busy, setBusy] = useState(false);
  const [saved, setSaved] = useState(false);

  if (loaded.state !== "ready") {
    return <p className="text-[15px] text-muted-foreground">{say("common.looking")}</p>;
  }

  const shownVersion = version ?? loaded.value.version;
  const shownExtended = extended ?? loaded.value.extended;
  const installed = loaded.value.installed;
  const asked = { version: loaded.value.version, extended: loaded.value.extended };

  async function save(event: React.FormEvent) {
    event.preventDefault();
    setBusy(true);
    setSaved(false);
    onTrouble("");
    try {
      await setHugo(organisation, project, { version: shownVersion.trim(), extended: shownExtended });
      setVersion(null);
      setExtended(null);
      again();
      setSaved(true);
    } catch (error: unknown) {
      onTrouble(whatWentWrong(error));
    } finally {
      setBusy(false);
    }
  }

  return (
    <form onSubmit={save} className="flex flex-col gap-4">
      <Field
        label={say("project-settings.hugo-version")}
        hint={say("project-settings.hugo-version-hint")}
        className="font-mono text-sm"
        value={shownVersion}
        onChange={(event) => setVersion(event.target.value)}
      />
      <Checkbox
        label={say("project-settings.extended")}
        checked={shownExtended}
        onChange={(event) => setExtended(event.target.checked)}
      />
      {installed.version && asked.version && !sameHugo(asked, installed) && (
        <p role="note" className="rounded-md bg-warning-soft px-3 py-2 text-[13px] text-warning-foreground">
          {say("project-settings.hugo-installed", { installed: hugoLabel(installed) })}
        </p>
      )}
      {saved && (
        <p role="status" className="text-[15px] text-muted-foreground">
          {say("common.saved")}
        </p>
      )}
      <div>
        <Button type="submit" disabled={busy}>
          {busy ? say("common.saving") : say("project-settings.save-hugo-version")}
        </Button>
      </div>
    </form>
  );
}
