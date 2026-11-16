import { ChevronLeft, Copy, ExternalLink, MoreHorizontal, Pencil, Plus, Trash2 } from "lucide-react";
import { useEffect, useState } from "react";
import { NavLink, useNavigate, useSearchParams } from "react-router";

import { Badge } from "../../components/ui/badge";
import { Button } from "../../components/ui/button";
import { Field } from "../../components/ui/field";
import { Segmented } from "../../components/ui/segmented";
import { whatWentWrong } from "../../lib/api";
import { recall, remember } from "../../lib/remember";
import { plural, say, type Key } from "../../lib/say";
import {
  createItem,
  deletePage,
  imageUrl,
  readCollectionItems,
  renamePage,
  type CollectionItem,
  type CollectionItems,
} from "../../lib/tenancy";
import { useAsync } from "../../lib/useAsync";
import { useDebounced } from "../../lib/useDebounced";
import { onADay } from "../../lib/when";
import { useHere } from "../../lib/whoami";
import { AskAboutItem, NewItemDialog, type Asking } from "./ItemDialogs";

export const pageSize = 20;

const sorts: { value: string; label: Key }[] = [
  { value: "newest", label: "collection.sort-newest" },
  { value: "oldest", label: "collection.sort-oldest" },
  { value: "title", label: "collection.sort-title" },
  { value: "edited", label: "collection.sort-edited" },
];

const badges: Record<CollectionItem["status"], { label: Key; tone: "changed" | "new" | "neutral" | "live" }> = {
  changed: { label: "collection.status-changed", tone: "changed" },
  new: { label: "collection.status-new", tone: "new" },
  draft: { label: "collection.status-draft", tone: "neutral" },
  published: { label: "collection.status-published", tone: "live" },
};

export function listQuery(params: URLSearchParams) {
  return { q: params.get("q") ?? "", folder: params.get("folder") ?? "", sort: params.get("sort") ?? "" };
}

export function carried(query: { q: string; folder: string; sort: string }) {
  const params = new URLSearchParams();
  if (query.q) params.set("q", query.q);
  if (query.folder) params.set("folder", query.folder);
  if (query.sort) params.set("sort", query.sort);
  const said = params.toString();
  return said ? `?${said}` : "";
}

function sortLabel(sort: string) {
  if (sort === "model") return say("collection.sort-model");
  const found = sorts.find((one) => one.value === sort);
  return say(found ? found.label : "collection.sort-newest");
}

function Row({
  project,
  onAPhone,
  item,
  listing,
  to,
  mayEdit,
  showPaths,
  onAsk,
  onOpen,
}: {
  project: string;
  onAPhone: boolean;
  item: CollectionItem;
  listing: CollectionItems;
  to: string;
  mayEdit: boolean;
  showPaths: boolean;
  onAsk: (asking: Asking) => void;
  onOpen: () => void;
}) {
  const { organisation } = useHere();
  const [menu, setMenu] = useState(false);
  const [broken, setBroken] = useState(false);
  const meta = [onADay(item.date), item.author, item.folder].filter(Boolean).join(" · ");
  const badge = badges[item.status];

  return (
    <li className="flex items-center gap-3 border-b border-border px-3 py-2 last:border-b-0">
      {listing.hasImages &&
        (item.thumbnail && !broken ? (
          <img
            src={imageUrl(organisation, project, item.thumbnail, true)}
            alt=""
            onError={() => setBroken(true)}
            className="h-[42px] w-14 shrink-0 rounded object-cover"
          />
        ) : (
          <span className="h-[42px] w-14 shrink-0 rounded bg-muted" aria-hidden="true" />
        ))}
      <div className="flex min-w-0 flex-1 flex-col">
        <NavLink to={to} className="flex min-h-11 items-center truncate text-[15px] font-medium hover:underline">
          {item.title}
        </NavLink>
        {meta && <span className="truncate text-[13px] text-subtle-foreground">{meta}</span>}
        {showPaths && item.path && <span className="truncate font-mono text-xs text-subtle-foreground">{item.path}</span>}
      </div>
      {!(onAPhone && item.status === "published") && <Badge tone={badge.tone}>{say(badge.label)}</Badge>}
      <div className="relative">
        <button
          type="button"
          aria-label={say("collection.actions-for", { title: item.title })}
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
                onOpen();
              }}
              className="flex min-h-11 items-center gap-2 rounded px-2 text-left text-[15px] hover:bg-muted"
            >
              <ExternalLink aria-hidden="true" className="size-4" />
              {say("collection.open")}
            </button>
            {mayEdit && (
              <>
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
              </>
            )}
          </div>
        )}
      </div>
    </li>
  );
}

export function CollectionView({
  project,
  collection,
  mayEdit,
  onAPhone,
}: {
  project: string;
  collection: string;
  mayEdit: boolean;
  onAPhone?: boolean;
}) {
  const { organisation, who } = useHere();
  const goTo = useNavigate();
  const [params, setParams] = useSearchParams();
  const query = listQuery(params);
  const shown = Math.max(1, Number(params.get("shown") ?? "1") || 1);
  const [typed, setTyped] = useState(query.q);
  const settled = useDebounced(typed, 250);
  const [refresh, setRefresh] = useState(0);
  const [asking, setAsking] = useState<Asking | null>(null);
  const [making, setMaking] = useState(false);
  const [busy, setBusy] = useState(false);
  const [trouble, setTrouble] = useState("");
  const pathsKey = `chatcms.content.paths.${who.person.id}`;
  const [showPaths, setShowPaths] = useState(() => recall(pathsKey, false));

  function set(changes: Record<string, string>) {
    const next = new URLSearchParams(params);
    for (const [name, value] of Object.entries(changes)) {
      if (value) next.set(name, value);
      else next.delete(name);
    }
    setParams(next, { replace: true });
  }

  useEffect(() => {
    if (settled !== query.q) set({ q: settled, shown: "" });
  }, [settled]);

  useEffect(() => {
    setTyped(query.q);
  }, [collection]);

  const [listing] = useAsync<CollectionItems>(
    () =>
      readCollectionItems(organisation, project, collection, {
        q: query.q,
        folder: query.folder,
        sort: query.sort,
        page: 1,
        pageSize: pageSize * shown,
      }),
    [organisation, project, collection, query.q, query.folder, query.sort, shown, refresh],
  );

  const base = `/o/${organisation}/p/${project}/content`;
  const itemAddress = (key: string) => `${base}/${key}${carried(query)}`;

  async function act(title: string) {
    if (!asking) return;
    setBusy(true);
    setTrouble("");
    try {
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

  if (listing.state === "loading") {
    return <p className="text-[15px] text-subtle-foreground">{say("common.looking")}</p>;
  }
  if (listing.state === "failed") {
    return <p className="text-[15px]">{listing.message}</p>;
  }

  const view = listing.value;
  const noun = view.itemName;
  const folders = view.folders;

  return (
    <section aria-label={view.title} className="flex min-w-0 flex-col gap-4">
      {onAPhone && (
        <NavLink to={base} className="flex min-h-11 w-fit items-center gap-1 text-[15px] text-primary">
          <ChevronLeft aria-hidden="true" className="size-4" />
          {say("collection.back-to-menu")}
        </NavLink>
      )}

      {trouble && (
        <p role="alert" className="rounded-md bg-destructive-soft px-3 py-2 text-sm text-destructive-foreground">
          {trouble}
        </p>
      )}

      <header className="flex flex-wrap items-end justify-between gap-3">
        <div className="min-w-0">
          <h2 className="text-[26px] font-semibold tracking-tight">{view.title}</h2>
          <p className="text-[15px] text-muted-foreground">
            {say("collection.line", {
              count: plural(view.all, "collection.count-one", "collection.count-many", { noun }),
              sort: sortLabel(view.sort),
            })}
          </p>
        </div>
        {mayEdit && (
          <Button variant="ink" onClick={() => setMaking(true)}>
            <Plus aria-hidden="true" className="size-4" />
            {say("collection.new", { noun })}
          </Button>
        )}
      </header>

      <div className="flex flex-wrap items-end gap-3">
        <div className="min-w-[200px] flex-1">
          <Field
            label={say("collection.search", { collection: view.title })}
            type="search"
            value={typed}
            onChange={(event) => setTyped(event.target.value)}
          />
        </div>

        {folders.length > 0 && folders.length <= 5 && (
          <Segmented
            label={say("collection.folder")}
            value={query.folder}
            onValueChange={(value) => set({ folder: value, shown: "" })}
            options={[
              { value: "", label: say("collection.all", { count: view.all }) },
              ...folders.map((folder) => ({
                value: folder.name,
                label: say("collection.one-folder", { folder: folder.label, count: folder.count }),
              })),
            ]}
          />
        )}
        {folders.length > 5 && (
          <div className="flex flex-col gap-1.5">
            <label htmlFor="collection-folder" className="text-sm font-medium">
              {say("collection.folder")}
            </label>
            <select
              id="collection-folder"
              value={query.folder}
              onChange={(event) => set({ folder: event.target.value, shown: "" })}
              className="h-11 rounded-md border border-input bg-card px-3 text-base sm:text-[15px]"
            >
              <option value="">{say("collection.all", { count: view.all })}</option>
              {folders.map((folder) => (
                <option key={folder.name} value={folder.name}>
                  {say("collection.one-folder", { folder: folder.label, count: folder.count })}
                </option>
              ))}
            </select>
          </div>
        )}

        <div className="flex flex-col gap-1.5">
          <label htmlFor="collection-sort" className="text-sm font-medium">
            {say("collection.sort")}
          </label>
          <select
            id="collection-sort"
            value={view.sort}
            onChange={(event) => set({ sort: event.target.value, shown: "" })}
            className="h-11 rounded-md border border-input bg-card px-3 text-base sm:text-[15px]"
          >
            {view.sort === "model" && <option value="model">{say("collection.sort-model")}</option>}
            {sorts.map((sort) => (
              <option key={sort.value} value={sort.value}>
                {say(sort.label)}
              </option>
            ))}
          </select>
        </div>

        {view.showsPaths && (
          <label className="flex min-h-11 items-center gap-2 text-[15px]">
            <input
              type="checkbox"
              role="switch"
              checked={showPaths}
              onChange={(event) => {
                setShowPaths(event.target.checked);
                remember(pathsKey, event.target.checked);
              }}
              className="size-5"
            />
            {say("collection.file-names")}
          </label>
        )}
      </div>

      {view.total === 0 && query.q !== "" && (
        <div className="flex flex-col items-start gap-2 rounded-xl border border-border bg-card p-5">
          <h3 className="text-base font-semibold">{say("collection.none-match", { noun, query: query.q })}</h3>
          <p className="text-[15px] text-muted-foreground">{say("collection.none-match-help")}</p>
          <Button
            variant="outline"
            onClick={() => {
              setTyped("");
              set({ q: "", shown: "" });
            }}
          >
            {say("collection.clear-search")}
          </Button>
        </div>
      )}

      {view.total === 0 && query.q === "" && (
        <div className="flex flex-col items-start gap-2 rounded-xl border border-border bg-card p-5">
          <h3 className="text-base font-semibold">{say("collection.empty", { noun })}</h3>
          {mayEdit && (
            <Button variant="ink" onClick={() => setMaking(true)}>
              <Plus aria-hidden="true" className="size-4" />
              {say("collection.new", { noun })}
            </Button>
          )}
        </div>
      )}

      {view.items.length > 0 && (
        <ul className="flex flex-col rounded-xl border border-border bg-card">
          {view.items.map((item) => (
            <Row
              key={item.key}
              project={project}
              onAPhone={onAPhone ?? false}
              item={item}
              listing={view}
              to={itemAddress(item.key)}
              mayEdit={mayEdit}
              showPaths={showPaths}
              onAsk={setAsking}
              onOpen={() => goTo(itemAddress(item.key))}
            />
          ))}
        </ul>
      )}

      {view.total > 0 && (
        <div className="flex flex-wrap items-center justify-between gap-2">
          <p className="text-sm text-subtle-foreground">
            {say("collection.showing", { shown: view.items.length, total: view.total })}
          </p>
          {view.items.length < view.total && (
            <Button variant="outline" onClick={() => set({ shown: String(shown + 1) })}>
              {say("content.show-more", { count: Math.min(pageSize, view.total - view.items.length) })}
            </Button>
          )}
        </div>
      )}

      {asking && <AskAboutItem asking={asking} busy={busy} onDo={act} onCancel={() => setAsking(null)} />}
      {making && (
        <NewItemDialog
          organisation={organisation}
          project={project}
          collection={collection}
          itemName={noun}
          folders={folders.map((folder) => folder.name)}
          folder={query.folder}
          onCancel={() => setMaking(false)}
          onMake={async (asked) => {
            const made = await createItem(organisation, project, collection, asked);
            goTo(itemAddress(made.key));
          }}
        />
      )}
    </section>
  );
}
