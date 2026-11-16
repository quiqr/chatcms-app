import { Copy, FileText, MoreHorizontal, Pencil, Plus, Trash2 } from "lucide-react";
import { useState } from "react";
import { useNavigate, useParams } from "react-router";

import { Badge } from "../components/ui/badge";
import { Button } from "../components/ui/button";
import { Field } from "../components/ui/field";
import { useAsync } from "../lib/useAsync";
import { onADay } from "../lib/when";
import { useHere } from "../lib/whoami";
import { say } from "../lib/say";
import {
  createPage,
  deletePage,
  readCollectionItems,
  readContentMenu,
  renamePage,
  type CollectionItems,
  type ContentItem,
  type ContentMenu,
  type MenuEntry,
  type MenuGroup,
} from "../lib/tenancy";
import { whatWentWrong } from "../lib/api";

const atATime = 20;

type Asking =
  | { what: "new"; section: string; itemName: string }
  | { what: "rename" | "copy"; page: string; title: string }
  | { what: "delete"; page: string; title: string };

function Ask({
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
  const [title, setTitle] = useState(asking.what === "new" ? "" : asking.title);

  if (asking.what === "delete") {
    return (
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="ask-title"
        className="flex flex-col gap-4 rounded-lg border border-border bg-card p-4"
      >
        <h3 id="ask-title" className="text-base font-semibold">
          {say("content.remove-page", { title: asking.title })}
        </h3>
        <p className="text-[15px]">
          {say("content.it-will-go-from-your-site")}
        </p>
        <div className="flex flex-wrap gap-2">
          <Button onClick={() => onDo(asking.title)} disabled={busy}>
            {busy ? say("content.removing") : say("common.yes-remove-it")}
          </Button>
          <Button variant="ghost" onClick={onCancel} disabled={busy}>
            {say("common.keep-it")}
          </Button>
        </div>
      </div>
    );
  }

  const heading =
    asking.what === "new"
      ? `New ${asking.itemName}`
      : asking.what === "copy"
        ? `Make a copy of “${asking.title}”`
        : `Rename “${asking.title}”`;

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-labelledby="ask-title"
      className="flex flex-col gap-4 rounded-lg border border-border bg-card p-4"
    >
      <h3 id="ask-title" className="text-base font-semibold">
        {heading}
      </h3>
      <Field
        label={say("common.name")}
        hint={say("content.this-is-what-it-is-called")}
        type="text"
        value={title}
        autoFocus
        onChange={(event) => setTitle(event.target.value)}
      />
      <div className="flex flex-wrap gap-2">
        <Button onClick={() => onDo(title)} disabled={busy || title.trim() === ""}>
          {busy ? say("content.working") : asking.what === "new" ? say("content.make-it") : say("common.save-the-name")}
        </Button>
        <Button variant="ghost" onClick={onCancel} disabled={busy}>
          {say("content.cancel")}
        </Button>
      </div>
    </div>
  );
}

function Row({
  item,
  onOpen,
  onAsk,
  canChange,
}: {
  item: ContentItem;
  onOpen: () => void;
  onAsk: (asking: Asking) => void;
  canChange: boolean;
}) {
  const [menu, setMenu] = useState(false);

  return (
    <li className="flex flex-wrap items-center gap-2 border-b border-border py-2 last:border-b-0">
      <FileText aria-hidden="true" className="size-4 shrink-0 text-subtle-foreground" />
      <button
        type="button"
        onClick={onOpen}
        className="flex min-h-11 flex-1 items-center text-left text-[15px] font-medium"
      >
        {item.title}
      </button>
      {item.waiting && <Badge tone="changed">{say("content.not-published")}</Badge>}
      {item.by && (
        <span className="text-sm text-subtle-foreground">
          {item.by}
          {onADay(item.at) ? ` · ${onADay(item.at)}` : ""}
        </span>
      )}
      {canChange && (
        <div className="relative">
          <button
            type="button"
            aria-label={`What to do with ${item.title}`}
            aria-expanded={menu}
            onClick={() => setMenu(!menu)}
            className="flex size-11 items-center justify-center rounded-md text-subtle-foreground hover:bg-muted"
          >
            <MoreHorizontal aria-hidden="true" className="size-4" />
          </button>
          {menu && (
            <div className="absolute right-0 z-10 mt-1 flex w-48 flex-col rounded-md border border-border bg-card p-1 shadow-md">
              <button
                type="button"
                onClick={() => {
                  setMenu(false);
                  onAsk({ what: "rename", page: item.key, title: item.title });
                }}
                className="flex min-h-11 items-center gap-2 rounded px-2 text-left text-[15px] hover:bg-muted"
              >
                <Pencil aria-hidden="true" className="size-4" />
                {say("content.rename")}
              </button>
              <button
                type="button"
                onClick={() => {
                  setMenu(false);
                  onAsk({ what: "copy", page: item.key, title: item.title });
                }}
                className="flex min-h-11 items-center gap-2 rounded px-2 text-left text-[15px] hover:bg-muted"
              >
                <Copy aria-hidden="true" className="size-4" />
                {say("content.make-a-copy")}
              </button>
              <button
                type="button"
                onClick={() => {
                  setMenu(false);
                  onAsk({ what: "delete", page: item.key, title: item.title });
                }}
                className="flex min-h-11 items-center gap-2 rounded px-2 text-left text-[15px] text-destructive-foreground hover:bg-muted"
              >
                <Trash2 aria-hidden="true" className="size-4" />
                {say("content.remove-2")}
              </button>
            </div>
          )}
        </div>
      )}
    </li>
  );
}

function CollectionBlock({
  entry,
  looking,
  refresh,
  onOpen,
  onAsk,
}: {
  entry: MenuEntry;
  looking: string;
  refresh: number;
  onOpen: (key: string) => void;
  onAsk: (asking: Asking) => void;
}) {
  const { organisation } = useHere();
  const { project = "" } = useParams();
  const [limit, setLimit] = useState(atATime);
  const [listing] = useAsync<CollectionItems>(
    () => readCollectionItems(organisation, project, entry.key),
    [organisation, project, entry.key, refresh],
  );

  const items: ContentItem[] =
    listing.state === "ready"
      ? listing.value.items.map((item) => ({
          key: item.key,
          title: item.title,
          waiting: item.status !== "published",
          by: item.by,
        }))
      : [];
  const matching = items.filter((item) => (looking === "" ? true : item.title.toLowerCase().includes(looking)));
  const shown = matching.slice(0, limit);

  return (
    <div className="flex flex-col gap-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h4 className="text-[15px] font-semibold">
          {entry.title}
          <span className="ml-2 text-sm font-normal text-subtle-foreground">{entry.count ?? 0}</span>
        </h4>
        <Button
          variant="ghost"
          onClick={() => onAsk({ what: "new", section: entry.key, itemName: entry.itemName ?? "" })}
        >
          <Plus aria-hidden="true" className="size-4" />
          {say("content.new")} {entry.itemName}
        </Button>
      </div>

      {listing.state === "failed" && <p className="text-[15px]">{listing.message}</p>}
      {listing.state === "ready" &&
        (matching.length === 0 ? (
          <p className="text-[15px] text-subtle-foreground">
            {looking === "" ? say("content.nothing-here-yet") : say("content.nothing-here-matches-that")}
          </p>
        ) : (
          <ul className="flex flex-col">
            {shown.map((item) => (
              <Row key={item.key} item={item} canChange onOpen={() => onOpen(item.key)} onAsk={onAsk} />
            ))}
          </ul>
        ))}

      {matching.length > shown.length && (
        <button
          type="button"
          onClick={() => setLimit(limit + atATime)}
          className="flex min-h-11 w-fit items-center text-sm font-medium text-primary"
        >
          {say("content.show-more", { count: Math.min(atATime, matching.length - shown.length) })}
        </button>
      )}
    </div>
  );
}

function groupTitle(menu: ContentMenu, group: MenuGroup) {
  if (!menu.fallback) return group.title;
  return group.key === "collections" ? say("content.collections") : say("content.pages");
}

export function ContentPage() {
  const { organisation } = useHere();
  const { project = "" } = useParams();
  const goTo = useNavigate();
  const [refresh, setRefresh] = useState(0);
  const [search, setSearch] = useState("");
  const [asking, setAsking] = useState<Asking | null>(null);
  const [busy, setBusy] = useState(false);
  const [trouble, setTrouble] = useState("");

  const [menu] = useAsync<ContentMenu>(
    () => readContentMenu(organisation, project),
    [organisation, project, refresh],
  );

  function open(key: string) {
    goTo(`/o/${organisation}/p/${project}/content/${key}`);
  }

  async function act(title: string) {
    if (!asking) return;
    setBusy(true);
    setTrouble("");
    try {
      if (asking.what === "new") {
        const made = await createPage(organisation, project, asking.section, title);
        open(made.key);
        return;
      }
      if (asking.what === "delete") {
        await deletePage(organisation, project, asking.page);
      } else {
        await renamePage(organisation, project, asking.page, title, asking.what === "copy");
      }
      setAsking(null);
      setRefresh(refresh + 1);
    } catch (error: unknown) {
      setTrouble(whatWentWrong(error));
    } finally {
      setBusy(false);
    }
  }

  if (menu.state === "loading") {
    return <p className="text-[15px] text-subtle-foreground">{say("content.looking-at-what-is-here")}</p>;
  }
  if (menu.state === "failed") {
    return <p className="text-[15px]">{menu.message}</p>;
  }

  const looking = search.trim().toLowerCase();

  return (
    <section className="flex flex-col gap-5">
      {trouble && (
        <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          {trouble}
        </p>
      )}

      {asking && <Ask asking={asking} busy={busy} onDo={act} onCancel={() => setAsking(null)} />}

      <Field
        label={say("content.search")}
        type="search"
        value={search}
        placeholder={say("content.find-a-page")}
        onChange={(event) => setSearch(event.target.value)}
      />

      {menu.value.groups.map((group) => {
        const singles = group.items.filter(
          (entry) => entry.kind === "single" && (looking === "" || entry.title.toLowerCase().includes(looking)),
        );
        const collections = group.items.filter((entry) => entry.kind === "collection");

        return (
          <section key={group.key} aria-label={groupTitle(menu.value, group)} className="flex flex-col gap-3">
            <h3 className="text-base font-semibold">{groupTitle(menu.value, group)}</h3>
            {singles.length > 0 && (
              <ul className="flex flex-col">
                {singles.map((entry) => (
                  <Row
                    key={entry.key}
                    item={{ key: entry.key, title: entry.title, waiting: entry.hasChanges }}
                    canChange={false}
                    onOpen={() => open(entry.key)}
                    onAsk={setAsking}
                  />
                ))}
              </ul>
            )}
            {collections.map((entry) => (
              <CollectionBlock
                key={entry.key}
                entry={entry}
                looking={looking}
                refresh={refresh}
                onOpen={open}
                onAsk={setAsking}
              />
            ))}
          </section>
        );
      })}
    </section>
  );
}
