import { useCallback, useEffect, useRef, useState } from "react";
import { whatWentWrong } from "./api";

export type Loading<T> =
  | { state: "loading" }
  | { state: "ready"; value: T }
  | { state: "failed"; message: string };

/**
 * useAsync loads something once and gives back a way to load it again. It
 * exists so that every screen reports loading and failure the same way.
 *
 * Asking again keeps what it already has on the screen. Going back to
 * "loading" would blank the screen below, which unmounts whatever somebody
 * was in the middle of typing: a half-filled form disappears because
 * something else was saved.
 */
const kept = new Map<string, unknown>();

export function forgetKept() {
  kept.clear();
}

export function useAsync<T>(load: () => Promise<T>, keys: unknown[], keep?: string): [Loading<T>, () => void] {
  const [result, setResult] = useState<Loading<T>>(() =>
    keep !== undefined && kept.has(keep) ? { state: "ready", value: kept.get(keep) as T } : { state: "loading" },
  );
  const [attempt, setAttempt] = useState(0);
  const asking = useRef(false);

  const again = useCallback(() => {
    asking.current = true;
    setAttempt((count) => count + 1);
  }, []);

  useEffect(() => {
    let cancelled = false;
    if (!asking.current) {
      setResult(
        keep !== undefined && kept.has(keep) ? { state: "ready", value: kept.get(keep) as T } : { state: "loading" },
      );
    }
    asking.current = false;

    load()
      .then((value) => {
        if (keep !== undefined) kept.set(keep, value);
        if (!cancelled) setResult({ state: "ready", value });
      })
      .catch((error: unknown) => {
        if (cancelled) return;
        setResult({
          state: "failed",
          message: whatWentWrong(error, "common.cannot-reach-the-server"),
        });
      });

    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [...keys, attempt]);

  return [result, again];
}
