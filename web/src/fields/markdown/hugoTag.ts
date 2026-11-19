export type Quote = '"' | "`" | "";

export type Param = {
  name?: string;
  value: string;
  quote: Quote;
  from: number;
  to: number;
  valueFrom: number;
  valueTo: number;
};

export type Tag = {
  delimiter: "<" | "%";
  name: string;
  params: Param[];
  closerAt: number;
  openEnd: number;
  inner?: { from: number; to: number };
};

export type Declared = {
  key: string;
  inner?: string;
  order: string[];
  positions: Record<string, number>;
};

const closerOf = { "<": ">}}", "%": "%}}" } as const;

export function readTag(raw: string): Tag | null {
  const head = raw.match(/^\{\{([<%])\s*([\w./-]+)/);
  if (!head) return null;
  const delimiter = head[1] as "<" | "%";
  const name = head[2]!;
  if (name.startsWith("/")) return null;
  const closer = closerOf[delimiter];
  const params: Param[] = [];
  let at = head[0].length;

  for (;;) {
    while (at < raw.length && /\s/.test(raw[at]!)) at++;
    if (raw.startsWith(closer, at)) break;
    if (raw.startsWith("/" + closer, at)) {
      at += 1;
      break;
    }
    if (at >= raw.length) return null;
    const from = at;
    let paramName: string | undefined;
    const named = raw.slice(at).match(/^([\w-]+)=/);
    if (named) {
      paramName = named[1];
      at += named[0].length;
    }
    const value = readValue(raw, at, closer);
    if (!value) return null;
    params.push({ name: paramName, value: value.value, quote: value.quote, from, to: value.end, valueFrom: at, valueTo: value.end });
    at = value.end;
  }

  const closerAt = at;
  const openEnd = at + closer.length;
  const tag: Tag = { delimiter, name, params, closerAt, openEnd };
  const rest = raw.slice(openEnd);
  const closing = rest.match(new RegExp(`\\{\\{${delimiter === "<" ? "<" : "%"}\\s*/\\s*${escape(name)}\\s*${delimiter === "<" ? ">" : "%"}\\}\\}\\s*$`));
  if (closing) {
    tag.inner = { from: openEnd, to: openEnd + (closing.index ?? 0) };
  } else if (rest.trim() !== "") {
    return null;
  }
  return tag;
}

function escape(text: string) {
  return text.replace(/[.*+?^${}()|[\]\\/]/g, "\\$&");
}

function readValue(raw: string, at: number, closer: string): { value: string; quote: Quote; end: number } | null {
  const first = raw[at];
  if (first === '"') {
    let value = "";
    let i = at + 1;
    while (i < raw.length) {
      const c = raw[i]!;
      if (c === "\\" && i + 1 < raw.length) {
        value += raw[i + 1];
        i += 2;
        continue;
      }
      if (c === '"') return { value, quote: '"', end: i + 1 };
      value += c;
      i++;
    }
    return null;
  }
  if (first === "`") {
    const end = raw.indexOf("`", at + 1);
    if (end === -1) return null;
    return { value: raw.slice(at + 1, end), quote: "`", end: end + 1 };
  }
  let i = at;
  while (i < raw.length && !/\s/.test(raw[i]!) && !raw.startsWith(closer, i) && !raw.startsWith("/" + closer, i)) {
    if (raw[i] === '"' || raw[i] === "`") return null;
    i++;
  }
  if (i === at) return null;
  return { value: raw.slice(at, i), quote: "", end: i };
}

export function valuesOf(raw: string, declared: Declared): { values: Record<string, string>; inner: string } | null {
  const tag = readTag(raw);
  if (!tag || tag.name !== declared.key) return null;
  const values: Record<string, string> = {};
  const positional = tag.params.filter((param) => param.name === undefined);
  for (const key of declared.order) {
    const position = declared.positions[key];
    const found = position !== undefined ? positional[position] : tag.params.find((param) => param.name === key);
    if (found) values[key] = found.value;
  }
  return { values, inner: tag.inner ? raw.slice(tag.inner.from, tag.inner.to).replace(/^\n/, "").replace(/\n$/, "") : "" };
}

function written(value: string, quote: Quote): string {
  if (quote === "`" && !value.includes("`")) return "`" + value + "`";
  if (quote === "" && value !== "" && /^[^\s"`]+$/.test(value) && !value.includes("}}")) return value;
  return '"' + value.replace(/\\/g, "\\\\").replace(/"/g, '\\"') + '"';
}

export function rewrite(raw: string, declared: Declared, values: Record<string, string>, inner: string): string | null {
  const tag = readTag(raw);
  if (!tag || tag.name !== declared.key) return null;
  const before = valuesOf(raw, declared)!;
  const positional = tag.params.filter((param) => param.name === undefined);
  type Edit = { from: number; to: number; text: string };
  const edits: Edit[] = [];
  const added: string[] = [];

  for (const key of declared.order) {
    const next = values[key] ?? "";
    if ((before.values[key] ?? "") === next) continue;
    const position = declared.positions[key];
    const found = position !== undefined ? positional[position] : tag.params.find((param) => param.name === key);
    if (found && next === "") {
      let from = found.from;
      while (from > 0 && /\s/.test(raw[from - 1]!)) from--;
      edits.push({ from, to: found.to, text: "" });
    } else if (found) {
      edits.push({ from: found.valueFrom, to: found.valueTo, text: written(next, found.quote) });
    } else if (next !== "") {
      added.push(position !== undefined ? written(next, "") : `${key}=${written(next, '"')}`);
    }
  }

  if (added.length > 0) {
    let at = tag.closerAt;
    while (at > 0 && /\s/.test(raw[at - 1]!)) at--;
    edits.push({ from: at, to: at, text: added.map((one) => " " + one).join("") });
  }

  if (tag.inner && before.inner !== inner) {
    const now = raw.slice(tag.inner.from, tag.inner.to);
    const lead = now.startsWith("\n") ? "\n" : "";
    const trail = now.endsWith("\n") ? "\n" : "";
    edits.push({ from: tag.inner.from, to: tag.inner.to, text: lead + inner + trail });
  }

  let out = raw;
  for (const edit of edits.sort((a, b) => b.from - a.from)) {
    out = out.slice(0, edit.from) + edit.text + out.slice(edit.to);
  }
  return out;
}

export function build(declared: Declared, values: Record<string, string>, inner: string): string {
  const delimiter = declared.inner === "markdown" ? "%" : "<";
  const closer = delimiter === "<" ? ">" : "%";
  const positional = declared.order
    .filter((key) => declared.positions[key] !== undefined)
    .sort((a, b) => declared.positions[a]! - declared.positions[b]!)
    .map((key) => values[key] ?? "");
  while (positional.length > 0 && positional[positional.length - 1] === "") positional.pop();
  const named = declared.order
    .filter((key) => declared.positions[key] === undefined && (values[key] ?? "") !== "")
    .map((key) => `${key}=${written(values[key]!, '"')}`);
  const params = [...positional.map((value) => written(value, "")), ...named];
  const open = `{{${delimiter} ${[declared.key, ...params].join(" ")} ${closer}}}`;
  if (!declared.inner) return open;
  return `${open}\n${inner}\n{{${delimiter} /${declared.key} ${closer}}}`;
}
