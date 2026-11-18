import { Monitor, Smartphone } from "lucide-react";
import { useEffect, useState } from "react";

import { Segmented } from "../components/ui/segmented";
import { whatWentWrong } from "../lib/api";
import { maybe, say, type Key } from "../lib/say";
import { openWorkspace, readWorkspace, type Workspace, type WorkspaceStep } from "../lib/tenancy";

/**
 * Preview shows the site as a reader will see it. It is served through the
 * app's own address, so somebody who may not see the project cannot see its
 * unpublished work either.
 */
export function Preview({
  organisation,
  project,
  page,
  showAgain = 0,
}: {
  organisation: string;
  project: string;
  page?: string;
  /**
   * showAgain is bumped by whatever knows the site changed. Hugo rebuilds by
   * itself, but the page already in the frame cannot be told so: its live
   * reload talks to Hugo directly, and Hugo is not reachable from a browser
   * on purpose.
   */
  showAgain?: number;
}) {
  const [device, setDevice] = useState<"desktop" | "phone">("desktop");
  const [reloads, setReloads] = useState(0);

  const at = `/preview/${organisation}/${project}/${pathOf(page)}`;
  const { workspace, refused } = useWorkspace(organisation, project);
  const ready = workspace?.state === "warm" || (workspace?.state === "sleeping" && !workspace.reason);

  return (
    <div className="flex h-full flex-col gap-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span className="truncate rounded-full bg-muted px-3 py-1 text-[13px] text-subtle-foreground">
          {say("preview.preview")} {at}
        </span>
        <div className="flex items-center gap-2">
          <Segmented
            label={say("preview.how-wide-to-show-the-site")}
            value={device}
            options={[
              { value: "desktop", label: "Desktop" },
              { value: "phone", label: "Phone" },
            ]}
            onValueChange={(next) => setDevice(next as "desktop" | "phone")}
          />
          <button
            type="button"
            onClick={() => setReloads(reloads + 1)}
            className="flex min-h-11 items-center px-2 text-sm font-medium text-primary"
          >
            {say("preview.refresh")}
          </button>
        </div>
      </div>

      <div className="flex flex-1 justify-center overflow-hidden rounded-lg border border-border bg-card">
        {ready ? (
          <iframe
            key={`${reloads}-${showAgain}`}
            title={say("preview.your-site")}
            src={at}
            className={
              "h-full border-0 bg-card " + (device === "phone" ? "w-[390px]" : "w-full")
            }
          />
        ) : (
          <Preparing workspace={workspace} refused={refused} />
        )}
      </div>

      <p className="flex items-center gap-2 text-[13px] text-subtle-foreground">
        {device === "phone" ? (
          <Smartphone aria-hidden="true" className="size-4" />
        ) : (
          <Monitor aria-hidden="true" className="size-4" />
        )}
        {say("preview.this-is-your-site-as-a")}
      </p>
    </div>
  );
}

/**
 * pathOf turns a page's key into the address Hugo serves it at: a single is
 * the site's front page unless it says otherwise, and an item of a collection
 * sits under its collection.
 */
function pathOf(page?: string) {
  if (!page || page === "home") return "";
  return page.endsWith("/") ? page : page + "/";
}

const stepNames: Record<WorkspaceStep["step"], Key> = {
  "starting-workspace": "preview.step-starting-workspace",
  "fetching-site": "preview.step-fetching-site",
  "building-preview": "preview.step-building-preview",
};

const stepStates: Record<WorkspaceStep["state"], Key> = {
  waiting: "preview.step-waiting",
  running: "preview.step-running",
  done: "preview.step-done",
  failed: "preview.step-failed",
};

function useWorkspace(organisation: string, project: string) {
  const [workspace, setWorkspace] = useState<Workspace | null>(null);
  const [refused, setRefused] = useState("");

  useEffect(() => {
    let live = true;
    let timer: ReturnType<typeof setTimeout> | undefined;

    const settle = (next: Workspace) => {
      if (!live) return;
      setWorkspace(next);
      if (next.state === "starting") {
        timer = setTimeout(() => readWorkspace(organisation, project).then(settle, fail), 500);
      }
    };
    const fail = (error: unknown) => {
      if (live) setRefused(whatWentWrong(error, "preview.did-not-start"));
    };

    setWorkspace(null);
    setRefused("");
    openWorkspace(organisation, project).then(settle, fail);
    return () => {
      live = false;
      clearTimeout(timer);
    };
  }, [organisation, project]);

  return { workspace, refused };
}

function Preparing({ workspace, refused }: { workspace: Workspace | null; refused: string }) {
  const failed = refused || (workspace?.state === "sleeping" && workspace.reason
    ? maybe(workspace.reason) ?? say("preview.did-not-start")
    : "");

  return (
    <div className="flex w-full max-w-[520px] flex-col items-center justify-center gap-4 p-6 text-center">
      <p role="status" className="text-[15px] text-muted-foreground">
        {failed || say("preview.preparing")}
      </p>
      {workspace?.steps && (
        <details data-testid="preview-steps" className="w-full text-left text-[13px]">
          <summary className="cursor-pointer text-primary">{say("preview.show-what-is-happening")}</summary>
          <ul className="mt-2 flex flex-col gap-1">
            {workspace.steps.map((step) => (
              <li key={step.step} className="flex justify-between gap-4">
                <span>{say(stepNames[step.step])}</span>
                <span className="text-subtle-foreground">{say(stepStates[step.state])}</span>
              </li>
            ))}
          </ul>
          {workspace.log && workspace.log.length > 0 && (
            <pre className="mt-2 max-h-48 overflow-auto rounded bg-muted p-2 text-[12px] whitespace-pre-wrap">
              {workspace.log.join("\n")}
            </pre>
          )}
        </details>
      )}
    </div>
  );
}
