import { Download } from "lucide-react";
import { useParams } from "react-router";

import { aNumber } from "../lib/numbers";
import { say } from "../lib/say";
import { artifactAddress, readArtifacts, type Artifact } from "../lib/tenancy";
import { useAsync } from "../lib/useAsync";
import { when } from "../lib/when";
import { useHere } from "../lib/whoami";

export function sizeOf(bytes: number) {
  if (bytes < 1024) return say("artifacts.size-bytes", { size: aNumber(bytes) });
  if (bytes < 1024 * 1024) return say("artifacts.size-kb", { size: aNumber(Math.round(bytes / 1024)) });
  return say("artifacts.size-mb", { size: aNumber(Math.round((bytes / (1024 * 1024)) * 10) / 10) });
}

export function ArtifactList({ organisation, project, artifacts }: { organisation: string; project: string; artifacts: Artifact[] }) {
  return (
    <ul className="flex flex-col divide-y divide-border rounded-lg border border-border bg-card">
      {artifacts.map((artifact) => (
        <li key={artifact.id} className="flex flex-wrap items-center gap-3 px-4 py-3">
          <div className="flex min-w-0 flex-1 flex-col">
            <span className="truncate text-[15px] font-medium">{artifact.name}</span>
            <span className="text-[13px] text-subtle-foreground">
              {sizeOf(artifact.size)}
              {artifact.createdAt ? ` · ${when(artifact.createdAt)}` : ""}
            </span>
          </div>
          <a
            href={artifactAddress(organisation, project, artifact.id)}
            download
            aria-label={say("artifacts.download", { name: artifact.name })}
            className="flex min-h-11 items-center gap-2 rounded-md px-3 text-sm font-medium text-primary hover:bg-muted"
          >
            <Download aria-hidden="true" className="size-4" />
            {say("artifacts.download-short")}
          </a>
        </li>
      ))}
    </ul>
  );
}

export function ArtifactsPage() {
  const { organisation } = useHere();
  const { project = "" } = useParams();
  const [loaded] = useAsync(() => readArtifacts(organisation, project), [organisation, project]);

  if (loaded.state === "loading") {
    return <p className="text-[15px] text-muted-foreground">{say("common.looking")}</p>;
  }
  if (loaded.state === "failed") {
    return (
      <p role="alert" className="text-[15px] text-destructive-foreground">
        {say("artifacts.could-not-be-read")}
      </p>
    );
  }
  if (loaded.value.length === 0) {
    return <p className="text-[15px] text-muted-foreground">{say("artifacts.none-yet")}</p>;
  }
  return (
    <div className="flex max-w-[760px] flex-col gap-4">
      <p className="text-[15px] text-muted-foreground">{say("artifacts.what-these-are")}</p>
      <ArtifactList organisation={organisation} project={project} artifacts={loaded.value} />
    </div>
  );
}
