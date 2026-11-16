import { useEffect, useState } from "react";

import { Button } from "../../components/ui/button";
import { Dialog, DialogContent } from "../../components/ui/dialog";
import { Field } from "../../components/ui/field";
import { useDebounced } from "../../lib/useDebounced";
import { say } from "../../lib/say";
import { readItemAddress, type ItemAddress } from "../../lib/tenancy";
import { whatWentWrong } from "../../lib/api";

export type Asking =
  | { what: "rename" | "copy"; page: string; title: string }
  | { what: "delete"; page: string; title: string };

export function AskAboutItem({
  asking,
  busy,
  onDo,
  onCancel,
}: {
  asking: Asking;
  busy: boolean;
  onDo: (title: string) => void;
  onCancel: () => void;
}) {
  const [title, setTitle] = useState(asking.title);

  if (asking.what === "delete") {
    return (
      <Dialog open onOpenChange={(open) => !open && onCancel()}>
        <DialogContent title={say("content.remove-page", { title: asking.title })}>
          <p className="text-[15px]">{say("content.it-will-go-from-your-site")}</p>
          <div className="mt-5 flex flex-wrap gap-2">
            <Button onClick={() => onDo(asking.title)} disabled={busy}>
              {busy ? say("content.removing") : say("common.yes-remove-it")}
            </Button>
            <Button variant="ghost" onClick={onCancel} disabled={busy}>
              {say("common.keep-it")}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    );
  }

  return (
    <Dialog open onOpenChange={(open) => !open && onCancel()}>
      <DialogContent title={asking.what === "copy" ? say("content.make-a-copy") : say("content.rename")}>
        <form
          className="flex flex-col gap-4"
          onSubmit={(event) => {
            event.preventDefault();
            onDo(title);
          }}
        >
          <Field
            label={say("common.name")}
            hint={say("content.this-is-what-it-is-called")}
            type="text"
            value={title}
            autoFocus
            onChange={(event) => setTitle(event.target.value)}
          />
          <div className="flex flex-wrap gap-2">
            <Button type="submit" disabled={busy || title.trim() === ""}>
              {busy ? say("content.working") : say("common.save-the-name")}
            </Button>
            <Button type="button" variant="ghost" onClick={onCancel} disabled={busy}>
              {say("content.cancel")}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}

export function NewItemDialog({
  organisation,
  project,
  collection,
  itemName,
  folders,
  folder: startIn,
  onMake,
  onCancel,
}: {
  organisation: string;
  project: string;
  collection: string;
  itemName: string;
  folders: string[];
  folder: string;
  onMake: (asked: { title: string; folder: string; slug: string }) => Promise<void>;
  onCancel: () => void;
}) {
  const [title, setTitle] = useState("");
  const [folder, setFolder] = useState(startIn);
  const [slug, setSlug] = useState("");
  const [editing, setEditing] = useState(false);
  const [address, setAddress] = useState<ItemAddress | null>(null);
  const [checking, setChecking] = useState(false);
  const [busy, setBusy] = useState(false);
  const [trouble, setTrouble] = useState("");
  const asked = useDebounced({ title, folder, slug: editing ? slug : "" }, 250);

  useEffect(() => {
    if (asked.title.trim() === "" && asked.slug.trim() === "") {
      setAddress(null);
      return;
    }
    let cancelled = false;
    setChecking(true);
    readItemAddress(organisation, project, collection, asked)
      .then((answer) => {
        if (cancelled) return;
        setAddress(answer);
        setTrouble("");
        if (!editing) setSlug(answer.slug);
      })
      .catch((error: unknown) => {
        if (!cancelled) {
          setAddress(null);
          setTrouble(whatWentWrong(error));
        }
      })
      .finally(() => {
        if (!cancelled) setChecking(false);
      });
    return () => {
      cancelled = true;
    };
  }, [organisation, project, collection, asked.title, asked.folder, asked.slug]);

  const current =
    address !== null && asked.title === title && asked.folder === folder && (!editing || asked.slug === slug);
  const ready = current && address.free && !checking && title.trim() !== "";

  async function make(event: React.FormEvent) {
    event.preventDefault();
    if (!ready) return;
    setBusy(true);
    try {
      await onMake({ title: title.trim(), folder, slug: editing ? slug : "" });
    } catch (error: unknown) {
      setTrouble(whatWentWrong(error));
      setBusy(false);
    }
  }

  return (
    <Dialog open onOpenChange={(open) => !open && onCancel()}>
      <DialogContent title={say("collection.new", { noun: itemName })}>
        <form className="flex flex-col gap-4" onSubmit={make}>
          {trouble && (
            <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
              {trouble}
            </p>
          )}
          <Field
            label={say("new-item.title")}
            required
            autoFocus
            value={title}
            onChange={(event) => setTitle(event.target.value)}
          />

          {folders.length > 0 && (
            <div className="flex flex-col gap-1.5">
              <label htmlFor="new-item-folder" className="text-sm font-medium">
                {say("new-item.folder")}
              </label>
              <select
                id="new-item-folder"
                value={folder}
                onChange={(event) => setFolder(event.target.value)}
                className="h-11 rounded-md border border-input bg-card px-3 text-base sm:text-[15px]"
              >
                <option value="">{say("new-item.top-folder")}</option>
                {folders.map((name) => (
                  <option key={name} value={name}>
                    {name}
                  </option>
                ))}
              </select>
            </div>
          )}

          {editing ? (
            <Field
              label={say("new-item.address-field")}
              className="font-mono text-sm"
              value={slug}
              onChange={(event) => setSlug(event.target.value)}
            />
          ) : (
            <div className="flex flex-wrap items-center gap-2">
              <span className="text-sm font-medium">{say("new-item.address")}</span>
              <span className="min-w-0 flex-1 truncate font-mono text-sm text-muted-foreground">
                {address?.address ?? ""}
              </span>
              <Button type="button" variant="ghost" size="compact" onClick={() => setEditing(true)}>
                {say("new-item.edit-address")}
              </Button>
            </div>
          )}

          {address?.file && (
            <p className="text-[13px] text-subtle-foreground">
              {say("new-item.file")}: <span className="font-mono">{address.file}</span>
            </p>
          )}
          {checking && <p className="text-[13px] text-subtle-foreground">{say("new-item.checking")}</p>}
          {current && !address.free && (
            <p role="alert" className="text-[13px] text-destructive-foreground">
              {say("new-item.taken", { noun: itemName })}
            </p>
          )}

          <div className="flex flex-wrap justify-end gap-2">
            <Button type="button" variant="ghost" onClick={onCancel} disabled={busy}>
              {say("content.cancel")}
            </Button>
            <Button type="submit" variant="ink" disabled={!ready || busy}>
              {busy ? say("content.working") : say("new-item.create")}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
