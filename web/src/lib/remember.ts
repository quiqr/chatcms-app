export function recall<T>(key: string, fallback: T): T {
  try {
    const held = window.localStorage.getItem(key);
    return held === null ? fallback : (JSON.parse(held) as T);
  } catch {
    return fallback;
  }
}

export function remember(key: string, value: unknown) {
  try {
    window.localStorage.setItem(key, JSON.stringify(value));
  } catch {
    return;
  }
}
