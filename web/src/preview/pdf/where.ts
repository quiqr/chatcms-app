export function fileOf(location: Location): URL | null {
  const at = location.pathname.indexOf("__chatcms/");
  if (at < 0) return null;
  const base = location.origin + location.pathname.slice(0, at);
  const asked = new URLSearchParams(location.search).get("file") ?? "";
  const file = new URL(asked.replace(/^\/+/, ""), base);
  if (file.origin !== location.origin || !file.pathname.startsWith(location.pathname.slice(0, at))) return null;
  return file;
}

export function remembered(storage: Storage | null, file: URL, hash: string): number {
  const asked = Number(new URLSearchParams(hash.replace(/^#/, "")).get("page"));
  if (asked > 0) return asked;
  try {
    return Number(storage?.getItem("chatcms-pdf:" + file.pathname)) || 1;
  } catch {
    return 1;
  }
}

export function remember(storage: Storage | null, file: URL, page: number) {
  try {
    storage?.setItem("chatcms-pdf:" + file.pathname, String(page));
  } catch {
    return;
  }
}
