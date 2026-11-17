import { useState } from "react";

import { recall, remember } from "./remember";

export function foldKey(panel: string, person: string) {
  return `chatcms.panels.${panel}.${person}`;
}

function start(key: string, fallbackKey?: string) {
  const own = recall<boolean | null>(key, null);
  if (own !== null) return own;
  return fallbackKey ? recall(fallbackKey, false) : false;
}

export function useFold(panel: string, person: string, fallbackKey?: string): [boolean, (next: boolean) => void] {
  const key = foldKey(panel, person);
  const [folded, setFolded] = useState(() => start(key, fallbackKey));

  function set(next: boolean) {
    setFolded(next);
    remember(key, next);
  }

  return [folded, set];
}
