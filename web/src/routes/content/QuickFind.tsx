import { FileText, Layers } from "lucide-react";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router";

import { Dialog, DialogContent } from "../../components/ui/dialog";
import { Field } from "../../components/ui/field";
import { say } from "../../lib/say";
import { searchContent, type SearchResults } from "../../lib/tenancy";
import { useHere } from "../../lib/whoami";

const opening = "chatcms:quick-find";
const pause = 200;

export function openQuickFind() {
  window.dispatchEvent(new Event(opening));
}

export function QuickFind({ project }: { project: string }) {
  const { organisation } = useHere();
  const goTo = useNavigate();
  const [open, setOpen] = useState(false);
  const [q, setQ] = useState("");
  const [found, setFound] = useState<SearchResults | null>(null);
  const [looking, setLooking] = useState(false);

  useEffect(() => {
    function keys(event: KeyboardEvent) {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
        event.preventDefault();
        setOpen(true);
      }
    }
    function asked() {
      setOpen(true);
    }
    window.addEventListener("keydown", keys);
    window.addEventListener(opening, asked);
    return () => {
      window.removeEventListener("keydown", keys);
      window.removeEventListener(opening, asked);
    };
  }, []);

  useEffect(() => {
    if (!open || q.trim() === "") {
      setFound(null);
      return;
    }
    let cancelled = false;
    setLooking(true);
    const timer = setTimeout(() => {
      searchContent(organisation, project, q)
        .then((answer) => {
          if (!cancelled) setFound(answer);
        })
        .catch(() => {
          if (!cancelled) setFound({ groups: [] });
        })
        .finally(() => {
          if (!cancelled) setLooking(false);
        });
    }, pause);
    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  }, [open, q, organisation, project]);

  function close(next: boolean) {
    setOpen(next);
    if (!next) {
      setQ("");
      setFound(null);
    }
  }

  function go(key: string) {
    close(false);
    goTo(`/o/${organisation}/p/${project}/content/${key}`);
  }

  const first = found?.groups[0]?.items[0];

  return (
    <Dialog open={open} onOpenChange={close}>
      <DialogContent title={say("quick-find.title")}>
        <form
          onSubmit={(event) => {
            event.preventDefault();
            if (first) go(first.key);
          }}
          className="flex flex-col gap-4"
        >
          <Field
            label={say("quick-find.field")}
            type="search"
            autoFocus
            value={q}
            placeholder={say("quick-find.placeholder")}
            onChange={(event) => setQ(event.target.value)}
          />
          {looking && !found && <p className="text-[15px] text-subtle-foreground">{say("quick-find.looking")}</p>}
          {found && found.groups.length === 0 && (
            <p className="text-[15px] text-subtle-foreground">{say("quick-find.nothing")}</p>
          )}
          {found?.groups.map((group) => (
            <section key={group.title} aria-label={group.title} className="flex flex-col gap-0.5">
              <h3 className="text-xs font-semibold uppercase tracking-wide text-subtle-foreground">{group.title}</h3>
              {group.items.map((item) => {
                const Icon = item.kind === "single" ? FileText : Layers;
                return (
                  <button
                    key={item.key}
                    type="button"
                    onClick={() => go(item.key)}
                    className="flex min-h-11 items-center gap-2 rounded-lg px-2 text-left text-[15px] hover:bg-muted"
                  >
                    <Icon aria-hidden="true" className="size-4 shrink-0 text-subtle-foreground" />
                    <span className="min-w-0 flex-1 truncate">{item.title}</span>
                  </button>
                );
              })}
            </section>
          ))}
        </form>
      </DialogContent>
    </Dialog>
  );
}
