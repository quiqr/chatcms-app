import { ArrowLeft } from "lucide-react";
import { lazy, Suspense, useEffect, useState } from "react";
import { useParams, useSearchParams } from "react-router";

import { Button } from "../components/ui/button";
import { whatWentWrong, RequestFailed } from "../lib/api";
import { say } from "../lib/say";
import { readFiles, readProjectFile, saveProjectFile, type FileEntry, type ProjectFile } from "../lib/tenancy";
import { useHere } from "../lib/whoami";
import { useIsPhone } from "../shells/useIsPhone";
import { FileTree } from "./files/FileTree";

const CodeEditor = lazy(() => import("./files/CodeEditor"));

type Saving = "idle" | "saving" | "saved";

function OpenFile({ organisation, project, path, onDirty }: { organisation: string; project: string; path: string; onDirty: (dirty: boolean) => void }) {
  const [file, setFile] = useState<ProjectFile | null>(null);
  const [draft, setDraft] = useState("");
  const [trouble, setTrouble] = useState("");
  const [stale, setStale] = useState(false);
  const [saving, setSaving] = useState<Saving>("idle");
  const [reads, setReads] = useState(0);

  useEffect(() => {
    let live = true;
    setFile(null);
    setTrouble("");
    setStale(false);
    setSaving("idle");
    readProjectFile(organisation, project, path).then(
      (answer) => {
        if (!live) return;
        setFile(answer);
        setDraft(answer.text);
      },
      (error: unknown) => live && setTrouble(whatWentWrong(error)),
    );
    return () => {
      live = false;
    };
  }, [organisation, project, path, reads]);

  const dirty = file !== null && draft !== file.text;
  useEffect(() => onDirty(dirty), [dirty, onDirty]);

  const save = () => {
    if (!file || !dirty || saving === "saving") return;
    setSaving("saving");
    setTrouble("");
    saveProjectFile(organisation, project, path, draft, file.sum ?? "").then(
      (answer) => {
        setFile({ ...answer, text: draft });
        setSaving("saved");
      },
      (error: unknown) => {
        setSaving("idle");
        setStale(error instanceof RequestFailed && error.key === "refused.file-changed");
        setTrouble(whatWentWrong(error));
      },
    );
  };

  if (trouble && !file) {
    return (
      <p role="alert" className="m-4 rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
        {trouble}
      </p>
    );
  }
  if (!file) return <p className="m-4 text-[13px] text-muted-foreground">{say("common.looking")}</p>;
  if (file.unusable) {
    return <p className="m-4 text-[15px] text-muted-foreground">{say(file.unusable === "too-large" ? "files.too-large" : "files.binary")}</p>;
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <div className="flex flex-wrap items-center gap-3 border-b border-border px-4 py-2">
        <p className="min-w-0 flex-1 truncate font-mono text-[13px]">{path}</p>
        <span role="status" className="text-[13px] text-subtle-foreground">
          {saving === "saving" ? say("files.saving") : dirty ? say("files.unsaved") : saving === "saved" ? say("files.saved") : ""}
        </span>
        <Button type="button" size="secondary" onClick={save} disabled={!dirty || saving === "saving"}>
          {say("files.save")}
        </Button>
      </div>
      {trouble && (
        <div role="alert" className="flex flex-wrap items-center gap-3 bg-destructive-soft px-4 py-2 text-sm text-destructive-foreground">
          <span className="flex-1">{trouble}</span>
          {stale && (
            <Button type="button" variant="outline" size="secondary" onClick={() => setReads((count) => count + 1)}>
              {say("files.open-again")}
            </Button>
          )}
        </div>
      )}
      <div className="min-h-0 flex-1">
        <Suspense fallback={<p className="m-4 text-[13px] text-muted-foreground">{say("common.looking")}</p>}>
          <CodeEditor key={`${path}-${reads}`} path={path} text={file.text} label={say("files.editor", { path })} onChange={setDraft} onSave={save} />
        </Suspense>
      </div>
    </div>
  );
}

export function FilesPage() {
  const { organisation } = useHere();
  const { project = "" } = useParams();
  const onAPhone = useIsPhone();
  const [asked, setAsked] = useSearchParams();
  const opened = asked.get("file") ?? "";
  const [entries, setEntries] = useState<FileEntry[] | null>(null);
  const [trouble, setTrouble] = useState("");
  const [dirty, setDirty] = useState(false);

  useEffect(() => {
    readFiles(organisation, project).then(
      (answer) => setEntries(answer.entries),
      (error: unknown) => setTrouble(whatWentWrong(error)),
    );
  }, [organisation, project]);

  const open = (path: string) => {
    if (path === opened) return;
    if (dirty && !window.confirm(say("files.leave-unsaved"))) return;
    setDirty(false);
    const query = new URLSearchParams(asked);
    if (path) query.set("file", path);
    else query.delete("file");
    setAsked(query);
  };

  if (trouble) {
    return (
      <p role="alert" className="m-6 rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
        {trouble}
      </p>
    );
  }
  if (!entries) return <p className="p-6 text-[15px] text-muted-foreground">{say("common.looking")}</p>;

  const tree =
    entries.length === 0 ? (
      <p className="p-4 text-[15px] text-muted-foreground">{say("files.empty")}</p>
    ) : (
      <FileTree entries={entries} label={say("files.tree")} chosen={opened} onOpen={open} />
    );

  if (onAPhone) {
    if (opened) {
      return (
        <div className="flex h-full min-h-[70vh] flex-col">
          <Button type="button" variant="ghost" size="secondary" className="m-2 self-start" onClick={() => open("")}>
            <ArrowLeft aria-hidden="true" className="size-4" />
            {say("files.back")}
          </Button>
          <OpenFile organisation={organisation} project={project} path={opened} onDirty={setDirty} />
        </div>
      );
    }
    return (
      <div className="flex flex-col gap-2 p-4">
        <h1 className="text-[28px] font-semibold tracking-tight">{say("files.title")}</h1>
        <p className="text-[15px] text-muted-foreground">{say("files.intro")}</p>
        <div className="rounded-xl border border-border bg-card">{tree}</div>
      </div>
    );
  }

  return (
    <div className="flex h-full min-h-0 flex-col gap-4 p-6">
      <div className="flex flex-col gap-1">
        <h1 className="text-[28px] font-semibold tracking-tight">{say("files.title")}</h1>
        <p className="text-[15px] text-muted-foreground">{say("files.intro")}</p>
      </div>
      <div className="flex min-h-[60vh] flex-1 overflow-hidden rounded-xl border border-border bg-card">
        <div className="w-72 shrink-0 overflow-auto border-r border-border">{tree}</div>
        <div className="flex min-w-0 flex-1 flex-col">
          {opened ? (
            <OpenFile organisation={organisation} project={project} path={opened} onDirty={setDirty} />
          ) : (
            <p className="m-auto text-[15px] text-muted-foreground">{say("files.choose")}</p>
          )}
        </div>
      </div>
    </div>
  );
}
