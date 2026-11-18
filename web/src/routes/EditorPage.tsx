import { ChevronLeft } from "lucide-react";
import { useState } from "react";
import { NavLink, useParams } from "react-router";

import { Button } from "../components/ui/button";
import { Segmented } from "../components/ui/segmented";
import { readProject, type PreviewSelection } from "../lib/tenancy";
import { useAsync } from "../lib/useAsync";
import { EditorShell, type EditorLayout } from "../shells/EditorShell";
import { PhoneShell } from "../shells/PhoneShell";
import { AssistantSlot } from "../shells/ProjectShell";
import { useHere } from "../lib/whoami";
import { useIsPhone } from "../shells/useIsPhone";
import { useStartWorkspace } from "../lib/workspace";
import { PageEditor } from "./PageEditor";
import { Preview } from "./Preview";
import { ChatPanel } from "./ChatPanel";
import { say } from "../lib/say";
import { Stepper, usePlace } from "./content/usePlace";

export function EditorPage() {
  const { organisation, who } = useHere();
  const { project: slug, "*": page = "" } = useParams();
  const [loaded] = useAsync(() => readProject(organisation, slug ?? ""), [organisation, slug]);
  const [layout, setLayout] = useState<EditorLayout>("split");
  const [saves, setSaves] = useState(0);
  const [changes, setChanges] = useState(0);
  const [talking, setTalking] = useState(false);
  const [chatted, setChatted] = useState(false);
  const [showing, setShowing] = useState<"edit" | "preview">("edit");
  const [opened, setOpened] = useState<{ title: string; section: string } | null>(null);
  const [quote, setQuote] = useState<PreviewSelection | null>(null);
  const [highlight, setHighlight] = useState<{ before: string; after: string; id: number } | null>(null);
  const [asked, setAsked] = useState(0);
  const onAPhone = useIsPhone();
  const place = usePlace(organisation, slug ?? "", page);
  useStartWorkspace(organisation, slug ?? "");

  if (loaded.state !== "ready") {
    return <p className="p-6 text-[15px] text-muted-foreground">{say("common.looking")}</p>;
  }

  const project = loaded.value;

  const changed = (about?: PreviewSelection) => {
    setChanges((count) => count + 1);
    if (about) setHighlight({ before: about.before, after: about.after, id: Date.now() });
  };

  const select = (selection: PreviewSelection) => {
    setQuote(selection);
    setAsked((count) => count + 1);
    setChatted(true);
    setTalking(true);
  };

  if (!onAPhone) {
    return (
      <div>
        <EditorShell
          project={project}
          viewer={{ name: who.person.displayName || who.person.email }}
          title={opened ? opened.title : page}
          context={place.context ?? opened?.section}
          backTo={place.backTo}
          backLabel={place.backLabel}
          stepper={<Stepper place={place} />}
          layout={layout}
          onLayoutChange={setLayout}
          openAssistant={asked}
          form={
            <PageEditor
              key={changes}
              organisation={organisation}
              headed={false}
              onOpened={setOpened}
              onSaved={() => setSaves((count) => count + 1)}
            />
          }
          preview={<Preview organisation={organisation} project={slug ?? ""} page={page} showAgain={saves + changes} onSelect={select} highlight={highlight} />}
          assistant={(close) => (
            <ChatPanel
              organisation={organisation}
              project={slug ?? ""}
              onChanged={changed}
              quote={quote}
              onQuoteDropped={() => setQuote(null)}
              onClose={close}
            />
          )}
        />
      </div>
    );
  }

  return (
    <PhoneShell
      project={project}
      viewer={{ name: who.person.displayName || who.person.email, role: project.yourRole }}
      bottomBar={
        <>
          <AssistantSlot
            open={talking}
            onToggle={() => {
              setChatted(true);
              setTalking((now) => !now);
            }}
          />
          <Button className="flex-1" count={project.unpublished}>
            {say("common.publish")}
          </Button>
        </>
      }
    >
      <div className="flex flex-col gap-4">
        <div className="flex items-center gap-2">
          <NavLink
            to={place.backTo ?? `/o/${organisation}/p/${slug}/content`}
            aria-label={place.backLabel ?? say("editor-shell.back-to-content")}
            className="flex size-11 shrink-0 items-center justify-center rounded-md hover:bg-muted"
          >
            <ChevronLeft aria-hidden="true" className="size-5" />
          </NavLink>
          <div className="flex min-w-0 flex-1 flex-col">
            {place.context && <p className="truncate text-[13px] text-subtle-foreground">{place.context}</p>}
            <p className="truncate text-[15px] font-semibold">{opened ? opened.title : page}</p>
          </div>
          <Stepper place={place} />
        </div>
        <Segmented
          label={say("editor.what-to-show")}
          className="grid grid-cols-2"
          value={showing}
          options={[
            { value: "edit", label: "Edit" },
            { value: "preview", label: "Preview" },
          ]}
          onValueChange={(next) => setShowing(next as "edit" | "preview")}
        />
        {chatted && (
          <div className="h-[70vh]" hidden={!talking}>
            <ChatPanel
              organisation={organisation}
              project={slug ?? ""}
              onChanged={changed}
              quote={quote}
              onQuoteDropped={() => setQuote(null)}
              onClose={() => setTalking(false)}
            />
          </div>
        )}
        {talking ? null : showing === "edit" ? (
          <PageEditor
            key={changes}
            organisation={organisation}
            headed={false}
            onOpened={setOpened}
            onSaved={() => setSaves((count) => count + 1)}
          />
        ) : (
          <div className="h-[70vh]">
            <Preview
              organisation={organisation}
              project={slug ?? ""}
              page={page}
              showAgain={saves + changes}
              onSelect={select}
              highlight={highlight}
            />
          </div>
        )}
      </div>
    </PhoneShell>
  );
}
