/*
 * chatcms-audience: implementer
 */
import { useState } from "react";

import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Field } from "../../components/ui/field";
import { whatWentWrong } from "../../lib/api";
import { say } from "../../lib/say";
import { connectSite, readSite } from "../../lib/tenancy";
import { useAsync } from "../../lib/useAsync";
import { KeyPanel } from "./Panels";

export function ConnectPanel({
  organisation,
  project,
  onTrouble,
}: {
  organisation: string;
  project: string;
  onTrouble: (said: string) => void;
}) {
  const [site, reloadSite] = useAsync(() => readSite(organisation, project), [organisation, project]);
  const [remote, setRemote] = useState("");
  const [busy, setBusy] = useState(false);
  const [connected, setConnected] = useState(false);

  const kept = site.state === "ready" && site.value.hosted === true;

  async function connect(event: React.FormEvent) {
    event.preventDefault();
    setBusy(true);
    onTrouble("");
    try {
      await connectSite(organisation, project, remote.trim());
      setConnected(true);
      setRemote("");
      reloadSite();
    } catch (error: unknown) {
      onTrouble(whatWentWrong(error));
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="flex flex-col gap-4">
      {connected && (
        <p role="status" className="text-[15px] text-muted-foreground">
          {say("project-settings.connected")}
        </p>
      )}
      {kept && (
        <div className="flex flex-col gap-2">
          <div>
            <Badge tone="attention">{say("project-settings.kept-not-connected")}</Badge>
          </div>
          <p className="text-[15px] text-muted-foreground">{say("project-settings.kept-explained")}</p>
        </div>
      )}

      <KeyPanel organisation={organisation} project={project} onTrouble={onTrouble} />

      {kept && (
        <form onSubmit={connect} className="flex flex-wrap items-end gap-3">
          <div className="min-w-[240px] flex-1">
            <Field
              label={say("project-settings.address-to-connect")}
              className="font-mono text-sm"
              value={remote}
              onChange={(event) => setRemote(event.target.value)}
            />
          </div>
          <Button type="submit" variant="ink" disabled={busy}>
            {busy ? say("project-settings.connecting") : say("project-settings.connect")}
          </Button>
        </form>
      )}
    </div>
  );
}
