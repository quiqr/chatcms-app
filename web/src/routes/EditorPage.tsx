import { ChevronLeft } from "lucide-react";
import { ContentBreadcrumb } from "./content/ContentBreadcrumb";
import { useRef, useState } from "react";
import { NavLink, useParams } from "react-router";

import { Button } from "../components/ui/button";
import { Segmented } from "../components/ui/segmented";
import { readAssistant, readProject, type PagePreview, type PreviewRequest, type PreviewSelection } from "../lib/tenancy";
import { useAsync } from "../lib/useAsync";
import { useProjectChanges } from "../lib/projectChanges";
import { EditorShell, type EditorLayout } from "../shells/EditorShell";
import { PhoneShell } from "../shells/PhoneShell";
import { AssistantSlot } from "../shells/ProjectShell";
import { recall, remember } from "../lib/remember";
import { useHere } from "../lib/whoami";
import { useIsPhone } from "../shells/useIsPhone";
import { useStartWorkspace } from "../lib/workspace";
import { PageEditor } from "./PageEditor";
import { Preview } from "./Preview";
import { ChatPanel, type ChatControl, type RequestState } from "./ChatPanel";
import { say } from "../lib/say";
import { Stepper, usePlace } from "./content/usePlace";

export function EditorPage() {
  const { organisation, who } = useHere();
  const { project: slug, "*": page = "" } = useParams();
  const [loaded, again] = useAsync(() => readProject(organisation, slug ?? ""), [organisation, slug], `project:${organisation}/${slug}`);
  useProjectChanges(organisation, slug ?? "", again);
  const viewKey = `chatcms.editor.view.${who.person.id}.${slug ?? ""}`;
  const [layout, setLayoutNow] = useState<EditorLayout>(() => recall<EditorLayout>(viewKey, "split"));
  const setLayout = (next: EditorLayout) => {
    setLayoutNow(next);
    remember(viewKey, next);
  };
  const [running, setRunning] = useState(0);
  const [pills, setPills] = useState<RequestState[]>([]);
  const chat = useRef<ChatControl>(null);
  const [saves, setSaves] = useState(0);
  const [changes, setChanges] = useState(0);
  const [talking, setTalking] = useState(false);
  const [chatted, setChatted] = useState(false);
  const [showing, setShowing] = useState<"edit" | "preview">("edit");
  const [opened, setOpened] = useState<{ title: string; section: string; preview?: PagePreview } | null>(null);
  const [quote, setQuote] = useState<PreviewSelection | null>(null);
  const [highlight, setHighlight] = useState<{ before: string; after: string; id: number } | null>(null);
  const [request, setRequest] = useState<{ about: PreviewSelection; text: string; id: number } | null>(null);
  const [drafted, setDrafted] = useState<{ text: string; id: number } | null>(null);
  const [assistant] = useAsync(() => readAssistant(organisation, slug ?? ""), [organisation, slug]);
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

  const asking = (asked: PreviewRequest, how: "send" | "draft") => {
    const { prompt, ...about } = asked;
    setChatted(true);
    if (how === "send") {
      setRequest({ about, text: prompt, id: Date.now() });
      return;
    }
    setQuote(about);
    setDrafted({ text: prompt, id: Date.now() });
  };
  const pillAction = (action: "undo" | "stop" | "open-chat" | "seen", id: number) => {
    if (action === "undo") chat.current?.undo(id);
    if (action === "stop") chat.current?.stop(id);
    if (action === "seen") chat.current?.seen(id);
    if (action === "open-chat") {
      setLayout("chat");
      setTimeout(() => chat.current?.focus(), 0);
    }
  };
  const mayAsk = project.yourRole !== "viewer" && assistant.state === "ready" && assistant.value.available;
  if (!onAPhone) {
    return (
      <div>
        <EditorShell
          project={project}
          viewer={{ name: who.person.displayName || who.person.email }}
          title={opened ? opened.title : page}
          context={place.context ?? opened?.section}
          heading={<ContentBreadcrumb project={project.slug} position={place.position} />}
          backTo={place.backTo}
          backLabel={place.backLabel}
          stepper={<Stepper place={place} />}
          layout={layout}
          onLayoutChange={setLayout}
          chatCount={running}
          form={
            <PageEditor
              key={changes}
              organisation={organisation}
              headed={false}
              onOpened={setOpened}
              onSaved={() => setSaves((count) => count + 1)}
            />
          }
          preview={
            <Preview
              organisation={organisation}
              project={slug ?? ""}
              page={page}
              shows={opened?.preview}
              showAgain={saves + changes}
              onAsk={asking}
              chatOn={mayAsk && (layout === "preview" || layout === "chat")}
              highlight={highlight}
              pills={pills}
              onPill={pillAction}
            />
          }
          chat={
            <ChatPanel
              organisation={organisation}
              project={slug ?? ""}
              onChanged={changed}
              quote={quote}
              onQuoteDropped={() => setQuote(null)}
              onRunning={setRunning}
              onShowPlace={(about) => setHighlight({ before: about.before, after: about.after, id: Date.now() })}
              pageTitle={opened?.title ?? page}
              request={request}
              drafted={drafted}
              onRequests={setPills}
              control={chat}
            />
          }
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
              pageTitle={opened?.title ?? page}
              request={request}
              drafted={drafted}
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
              shows={opened?.preview}
              showAgain={saves + changes}
              onAsk={asking}
              chatOn={mayAsk}
              highlight={highlight}
            />
          </div>
        )}
      </div>
    </PhoneShell>
  );
}
