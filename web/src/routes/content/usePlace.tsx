import { ChevronDown, ChevronUp } from "lucide-react";
import { useEffect } from "react";
import { NavLink, useNavigate, useSearchParams } from "react-router";

import { Button } from "../../components/ui/button";
import { say } from "../../lib/say";
import { readCollectionItems, readContentMenu } from "../../lib/tenancy";
import { useAsync } from "../../lib/useAsync";
import { carried, listQuery } from "./CollectionView";
import { groupTitle } from "./ContentColumn";

export type Place = {
  context?: string;
  contextTo?: string;
  backTo?: string;
  backLabel?: string;
  previous?: string;
  next?: string;
  isItem: boolean;
};

export function usePlace(organisation: string, project: string, page: string): Place {
  const [params] = useSearchParams();
  const query = listQuery(params);
  const collection = page.includes("/") ? page.split("/")[0] : "";
  const base = `/o/${organisation}/p/${project}/content`;

  const [around] = useAsync(
    () =>
      collection
        ? readCollectionItems(organisation, project, collection, { ...query, around: page })
        : Promise.resolve(null),
    [organisation, project, collection, page, query.q, query.folder, query.sort],
  );
  const [menu] = useAsync(
    () => (collection ? Promise.resolve(null) : readContentMenu(organisation, project)),
    [organisation, project, collection],
  );

  if (collection) {
    const listing = around.state === "ready" ? around.value : null;
    const list = `${base}/c/${collection}${carried(query)}`;
    const place: Place = { isItem: true, backTo: list, contextTo: list };
    if (listing) {
      place.backLabel = say("editor.back-to", { collection: listing.title });
      place.context = listing.around
        ? say("editor.place", { collection: listing.title, index: listing.around.index, total: listing.total })
        : listing.title;
      if (listing.around?.previous) place.previous = `${base}/${listing.around.previous}${carried(query)}`;
      if (listing.around?.next) place.next = `${base}/${listing.around.next}${carried(query)}`;
    }
    return place;
  }

  if (menu.state === "ready" && menu.value) {
    for (const group of menu.value.groups) {
      if (group.items.some((entry) => entry.kind === "single" && entry.key === page)) {
        return { isItem: false, context: groupTitle(menu.value, group), backTo: base };
      }
    }
  }
  return { isItem: false, backTo: base };
}

function typing(target: EventTarget | null) {
  if (!(target instanceof HTMLElement)) return false;
  return target.isContentEditable || ["INPUT", "TEXTAREA", "SELECT"].includes(target.tagName);
}

export function Stepper({ place }: { place: Place }) {
  const goTo = useNavigate();

  useEffect(() => {
    function keys(event: KeyboardEvent) {
      if (!event.altKey || typing(event.target)) return;
      if (event.key === "ArrowUp" && place.previous) {
        event.preventDefault();
        goTo(place.previous);
      }
      if (event.key === "ArrowDown" && place.next) {
        event.preventDefault();
        goTo(place.next);
      }
    }
    window.addEventListener("keydown", keys);
    return () => window.removeEventListener("keydown", keys);
  }, [place.previous, place.next, goTo]);

  if (!place.isItem) return null;

  return (
    <div className="flex items-center gap-1">
      {place.previous ? (
        <Button variant="ghost" size="icon" asChild aria-label={say("editor.previous")}>
          <NavLink to={place.previous}>
            <ChevronUp aria-hidden="true" className="size-4" />
          </NavLink>
        </Button>
      ) : (
        <Button variant="ghost" size="icon" disabled aria-label={say("editor.previous")}>
          <ChevronUp aria-hidden="true" className="size-4" />
        </Button>
      )}
      {place.next ? (
        <Button variant="ghost" size="icon" asChild aria-label={say("editor.next")}>
          <NavLink to={place.next}>
            <ChevronDown aria-hidden="true" className="size-4" />
          </NavLink>
        </Button>
      ) : (
        <Button variant="ghost" size="icon" disabled aria-label={say("editor.next")}>
          <ChevronDown aria-hidden="true" className="size-4" />
        </Button>
      )}
    </div>
  );
}
