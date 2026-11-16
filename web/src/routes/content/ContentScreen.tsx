import { PanelLeftOpen } from "lucide-react";
import { useState } from "react";
import { Navigate, useParams } from "react-router";

import { Button } from "../../components/ui/button";
import { say } from "../../lib/say";
import { readContentMenu, readProject } from "../../lib/tenancy";
import { useAsync } from "../../lib/useAsync";
import { useHere } from "../../lib/whoami";
import { useIsPhone } from "../../shells/useIsPhone";
import { useIsWide } from "../../shells/useIsWide";
import { CollectionView } from "./CollectionView";
import { ContentColumn } from "./ContentColumn";

export function ContentDrawer({ project, onClose }: { project: string; onClose: () => void }) {
  return (
    <div className="fixed inset-0 z-40 flex" role="dialog" aria-modal="true" aria-label={say("content-menu.open")}>
      <div className="w-[min(20rem,85vw)] overflow-y-auto border-r border-border bg-card p-4 shadow-lg">
        <ContentColumn project={project} onChoose={onClose} onHide={onClose} />
      </div>
      <button
        type="button"
        aria-label={say("content-menu.hide")}
        className="flex-1 bg-foreground/30"
        onClick={onClose}
      />
    </div>
  );
}

function FirstView({ project }: { project: string }) {
  const { organisation } = useHere();
  const [menu] = useAsync(() => readContentMenu(organisation, project), [organisation, project]);

  if (menu.state === "loading") {
    return <p className="text-[15px] text-subtle-foreground">{say("common.looking")}</p>;
  }
  if (menu.state === "failed") {
    return <p className="text-[15px]">{menu.message}</p>;
  }
  for (const group of menu.value.groups) {
    const first = group.items.find((entry) => entry.kind === "collection");
    if (first) {
      return <Navigate replace to={`/o/${organisation}/p/${project}/content/c/${first.key}`} />;
    }
  }
  return <p className="text-[15px] text-muted-foreground">{say("content-menu.choose")}</p>;
}

export function ContentScreen() {
  const { organisation } = useHere();
  const { project = "", collection } = useParams();
  const onAPhone = useIsPhone();
  const wide = useIsWide();
  const [drawer, setDrawer] = useState(false);
  const [loaded] = useAsync(() => readProject(organisation, project), [organisation, project]);
  const mayEdit = loaded.state === "ready" && loaded.value.youMay.includes("edit-content");

  if (onAPhone) {
    if (!collection) return <ContentColumn project={project} />;
    return <CollectionView project={project} collection={collection} mayEdit={mayEdit} onAPhone />;
  }

  const main = collection ? (
    <CollectionView key={collection} project={project} collection={collection} mayEdit={mayEdit} />
  ) : (
    <FirstView project={project} />
  );

  return (
    <div className="flex items-start gap-8">
      {wide ? (
        <ContentColumn project={project} className="w-64 shrink-0" />
      ) : (
        drawer && <ContentDrawer project={project} onClose={() => setDrawer(false)} />
      )}
      <div className="flex min-w-0 flex-1 flex-col gap-4">
        {!wide && (
          <Button variant="outline" className="w-fit" onClick={() => setDrawer(true)}>
            <PanelLeftOpen aria-hidden="true" className="size-4" />
            {say("content-menu.open")}
          </Button>
        )}
        {main}
      </div>
    </div>
  );
}
