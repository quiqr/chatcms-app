import { Navigate, useParams } from "react-router";

import { say } from "../../lib/say";
import { readContentMenu, readProject } from "../../lib/tenancy";
import { useAsync } from "../../lib/useAsync";
import { useHere } from "../../lib/whoami";
import { useIsPhone } from "../../shells/useIsPhone";
import { CollectionView } from "./CollectionView";
import { ContentColumn } from "./ContentColumn";
import { ContentPanel } from "./ContentPanel";

function FirstView({ project }: { project: string }) {
  const { organisation } = useHere();
  const [menu] = useAsync(() => readContentMenu(organisation, project), [organisation, project], `menu:${organisation}/${project}`);

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
  const [loaded] = useAsync(() => readProject(organisation, project), [organisation, project], `project:${organisation}/${project}`);
  const mayEdit = loaded.state === "ready" && loaded.value.youMay.includes("edit-content");

  if (onAPhone) {
    if (!collection) return <ContentColumn project={project} />;
    return <CollectionView project={project} collection={collection} mayEdit={mayEdit} onAPhone />;
  }

  if (collection) {
    return <CollectionView key={collection} project={project} collection={collection} mayEdit={mayEdit} />;
  }
  return <FirstView project={project} />;
}

export function ContentBeside() {
  const { project = "" } = useParams();
  return <ContentPanel project={project} />;
}
