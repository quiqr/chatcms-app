import { Monitor, Smartphone } from "lucide-react";
import { useEffect, useRef, useState, type RefObject } from "react";
import { Link } from "react-router";

import { Button } from "../components/ui/button";
import { Segmented } from "../components/ui/segmented";
import { useIsPhone } from "../shells/useIsPhone";
import { whatWentWrong } from "../lib/api";
import { maybe, say, spokenNow } from "../lib/say";
import { logAddress, openWorkspace, readWorkspace, type PagePreview, type PreviewRequest, type Workspace } from "../lib/tenancy";
import { useRestarted } from "../lib/workspace";
import type { Pill, Words } from "../preview/select";

/**
 * Preview shows the site as a reader will see it. It is served through the
 * app's own address, or through an origin of its own when the installation
 * has one, and either way only to somebody who may see the project. A site
 * whose runtime makes a PDF shows that; one that makes something else is
 * offered to open.
 */
export function Preview({
  organisation,
  project,
  page,
  showAgain = 0,
  settle = 0,
  onAsk,
  chatOn = false,
  highlight,
  pills = [],
  onPill,
  shows,
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
  settle?: number;
  onAsk?: (request: PreviewRequest, how: "send" | "draft") => void;
  chatOn?: boolean;
  pills?: Pill[];
  onPill?: (action: "undo" | "stop" | "open-chat" | "seen", id: number) => void;
  highlight?: { before: string; after: string; id: number } | null;
  shows?: PagePreview;
}) {
  const [device, setDevice] = useState<"desktop" | "phone">("desktop");
  const [homeAnyway, setHomeAnyway] = useState(false);
  useEffect(() => setHomeAnyway(false), [page]);
  const [reloads, setReloads] = useState(0);
  useEffect(() => {
    if (settle === 0) return;
    const timers = settleAfter.map((after) => setTimeout(() => setReloads((count) => count + 1), after));
    return () => timers.forEach(clearTimeout);
  }, [settle]);
  const onAPhone = useIsPhone();

  const { workspace, refused } = useWorkspace(organisation, project);
  const ready = workspace?.state === "warm" || (workspace?.state === "sleeping" && !workspace.reason);
  const kind = workspace?.type ?? "web";
  const base = baseOf(workspace?.address, `/preview/${organisation}/${project}/`);
  const opened = (workspace?.open ?? "/").replace(/^\//, "");
  const pagePath = shows?.kind === "declared" ? pathOf((shows.address ?? "").replace(/^\/+/, "")) : homeAnyway ? "" : pathOf(page);
  const without = kind === "web" && !workspace?.ownPages && (shows?.kind === "hidden" || (shows?.kind === "none" && !homeAnyway));
  const at =
    kind === "web" ? base + (workspace?.ownPages ? opened : pagePath) : kind === "pdf" ? `${base}__chatcms/pdf/?file=${encodeURIComponent(opened)}${onAPhone ? "&fit=width" : ""}&lang=${spokenNow()}` : base + opened;
  const frame = useRef<HTMLIFrameElement>(null);
  useFrameMessages(frame, base, chatOn && Boolean(onAsk), onAsk, highlight, pills, onPill);

  return (
    <div className="flex h-full flex-col gap-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span className="truncate rounded-full bg-muted px-3 py-1 text-[13px] text-subtle-foreground">
          {say("preview.preview")} {at}
        </span>
        <div className="flex items-center gap-2">
          {kind === "web" && (
            <Segmented
              label={say("preview.how-wide-to-show-the-site")}
              value={device}
              options={[
                { value: "desktop", label: "Desktop" },
                { value: "phone", label: "Phone" },
              ]}
              onValueChange={(next) => setDevice(next as "desktop" | "phone")}
            />
          )}
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
        {without ? (
          <NoPage hidden={shows?.kind === "hidden"} onHome={() => setHomeAnyway(true)} />
        ) : ready && kind === "other" ? (
          <div className="flex w-full max-w-[520px] flex-col items-center justify-center gap-4 p-6 text-center">
            <p className="text-[15px] text-muted-foreground">{say("preview.cannot-show-here")}</p>
            <a href={at} target="_blank" rel="noreferrer" className="flex min-h-11 items-center font-medium text-primary">
              {say("preview.open-the-result")}
            </a>
          </div>
        ) : ready ? (
          <iframe
            key={`${reloads}-${showAgain}`}
            ref={frame}
            title={say("preview.your-site")}
            src={at}
            className={
              "h-full border-0 bg-card " + (device === "phone" && kind === "web" ? "w-[390px]" : "w-full")
            }
          />
        ) : (
          <Preparing workspace={workspace} refused={refused} log={workspace?.entry ? logAddress(organisation, project, workspace.entry) : ""} />
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

function useFrameMessages(
  frame: RefObject<HTMLIFrameElement | null>,
  base: string,
  on: boolean,
  onAsk?: (request: PreviewRequest, how: "send" | "draft") => void,
  highlight?: { before: string; after: string; id: number } | null,
  pills: Pill[] = [],
  onPill?: (action: "undo" | "stop" | "open-chat" | "seen", id: number) => void,
) {
  const latest = useRef({ onAsk, highlight, base, on, pills, onPill });
  latest.current = { onAsk, highlight, base, on, pills, onPill };

  useEffect(() => {
    const heard = (event: MessageEvent) => {
      const here = frame.current;
      if (!here || event.source !== here.contentWindow) return;
      const origin = new URL(here.src, window.location.href).origin;
      if (event.origin !== origin) return;
      const said = event.data as Record<string, unknown> | null;
      const { onAsk, highlight, base, on, pills, onPill } = latest.current;
      if (!said) return;
      if (said.chatcms === "ready") {
        here.contentWindow?.postMessage({ chatcms: "setup", words: overlayWords(), on }, origin);
        here.contentWindow?.postMessage({ chatcms: "pills", pills }, origin);
        if (highlight && Date.now() - highlight.id < highlightWhile) {
          here.contentWindow?.postMessage({ chatcms: "highlight", before: highlight.before, after: highlight.after }, origin);
        }
        return;
      }
      if (said.chatcms === "undo" || said.chatcms === "stop" || said.chatcms === "open-chat" || said.chatcms === "seen") {
        if (typeof said.id === "number") onPill?.(said.chatcms, said.id);
        return;
      }
      if ((said.chatcms !== "send" && said.chatcms !== "draft") || !onAsk || !on) return;
      const { text, before, after, path, prompt } = said;
      if (typeof text !== "string" || typeof before !== "string" || typeof after !== "string" || typeof path !== "string" || typeof prompt !== "string") return;
      onAsk(
        { text: text.slice(0, 2000), before: before.slice(-200), after: after.slice(0, 200), path: within(base, path), prompt: prompt.slice(0, 4000) },
        said.chatcms,
      );
    };
    window.addEventListener("message", heard);
    return () => window.removeEventListener("message", heard);
  }, [frame]);

  useEffect(() => {
    const here = frame.current;
    if (!here?.contentWindow || !here.src) return;
    const origin = new URL(here.src, window.location.href).origin;
    here.contentWindow.postMessage({ chatcms: "setup", words: overlayWords(), on }, origin);
  }, [frame, on]);

  const pillsKey = JSON.stringify(pills);
  useEffect(() => {
    const here = frame.current;
    if (!here?.contentWindow || !here.src) return;
    const origin = new URL(here.src, window.location.href).origin;
    here.contentWindow.postMessage({ chatcms: "pills", pills: JSON.parse(pillsKey) as Pill[] }, origin);
  }, [frame, pillsKey]);

  useEffect(() => {
    const here = frame.current;
    if (!highlight || !here?.contentWindow || !here.src) return;
    const origin = new URL(here.src, window.location.href).origin;
    here.contentWindow.postMessage({ chatcms: "highlight", before: highlight.before, after: highlight.after }, origin);
  }, [frame, highlight]);
}

const highlightWhile = 15_000;

function overlayWords(): Words {
  return {
    chip: say("preview.chat"),
    chipLabel: say("preview.chat-about-the-selection"),
    card: say("preview.chat-about-the-selection"),
    placeholder: say("preview.what-should-change"),
    send: say("preview.send-to-the-assistant"),
    quick: [say("preview.make-it-shorter"), say("preview.simpler-words"), say("preview.fix-spelling")],
    goesTo: say("preview.goes-to-your-chat"),
    keys: say("preview.enter-to-send"),
    working: say("preview.working"),
    stop: say("preview.stop"),
    changed: say("preview.changed"),
    undo: say("preview.undo"),
    showInChat: say("preview.show-in-chat"),
    question: say("preview.question"),
    answerInChat: say("preview.answer-in-chat"),
  };
}

function within(base: string, path: string) {
  const prefix = new URL(base, window.location.href).pathname;
  return path.startsWith(prefix) ? "/" + path.slice(prefix.length) : path;
}

function baseOf(address: string | undefined, own: string) {
  if (!address) return own;
  try {
    const parsed = new URL(address, window.location.href);
    if (parsed.origin === window.location.origin) return parsed.pathname.endsWith("/") ? parsed.pathname : parsed.pathname + "/";
    return parsed.href.endsWith("/") ? parsed.href : parsed.href + "/";
  } catch {
    return own;
  }
}

/**
 * pathOf turns a page's key into the address Hugo serves it at: a single is
 * the site's front page unless it says otherwise, and an item of a collection
 * sits under its collection.
 */
const settleAfter = [2_500, 8_000];

function pathOf(page?: string) {
  if (!page || page === "home") return "";
  return page.endsWith("/") ? page : page + "/";
}


function useWorkspace(organisation: string, project: string) {
  const [workspace, setWorkspace] = useState<Workspace | null>(null);
  const [refused, setRefused] = useState("");
  const restarted = useRestarted();

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
  }, [organisation, project, restarted]);

  return { workspace, refused };
}

function NoPage({ hidden, onHome }: { hidden: boolean; onHome: () => void }) {
  return (
    <div data-testid="preview-no-page" className="flex w-full max-w-[520px] flex-col items-center justify-center gap-3 p-6 text-center">
      {hidden ? (
        <p className="text-[15px] text-muted-foreground">{say("preview.switched-off")}</p>
      ) : (
        <>
          <p className="text-[15px] font-semibold">{say("preview.no-page-title")}</p>
          <p className="text-[15px] text-muted-foreground">{say("preview.no-page-explained")}</p>
          <Button type="button" variant="outline" size="secondary" onClick={onHome}>
            {say("preview.show-home")}
          </Button>
        </>
      )}
    </div>
  );
}

function Preparing({ workspace, refused, log }: { workspace: Workspace | null; refused: string; log: string }) {
  const failed = refused || (workspace?.state === "sleeping" && workspace.reason
    ? maybe(workspace.reason) ?? say("preview.did-not-start")
    : "");

  return (
    <div className="flex w-full max-w-[520px] flex-col items-center justify-center gap-4 p-6 text-center">
      <p role="status" className="text-[15px] text-muted-foreground">
        {failed || say("preview.preparing")}
      </p>
      {log && (
        <Link to={log} className="text-[13px] text-primary">
          {say(failed ? "log.see-what-went-wrong" : "log.show-the-log")}
        </Link>
      )}
    </div>
  );
}
