import type { Node } from "@milkdown/kit/prose/model";

export type Range = { from: number; to: number };

type Positioned = { type: string; position?: { start: { offset?: number }; end: { offset?: number } } };

export function blocksOf(tree: { children: Positioned[] }): Range[] {
  return tree.children
    .filter((node) => node.type !== "definition")
    .map((node) => ({ from: node.position?.start.offset ?? 0, to: node.position?.end.offset ?? 0 }));
}

function topNodes(doc: Node): Node[] {
  const nodes: Node[] = [];
  doc.forEach((node) => nodes.push(node));
  if (nodes.length === 1 && nodes[0]?.type.name === "paragraph" && nodes[0].content.size === 0) return [];
  return nodes;
}

export function linesUp(blocks: Range[], doc: Node): boolean {
  return topNodes(doc).length === blocks.length;
}

function matched(old: Node[], now: Node[]): Array<[number, number]> {
  const n = old.length;
  const m = now.length;
  const table: number[][] = Array.from({ length: n + 1 }, () => new Array<number>(m + 1).fill(0));
  for (let i = n - 1; i >= 0; i--) {
    for (let j = m - 1; j >= 0; j--) {
      table[i]![j] = old[i]!.eq(now[j]!) ? table[i + 1]![j + 1]! + 1 : Math.max(table[i + 1]![j]!, table[i]![j + 1]!);
    }
  }
  const pairs: Array<[number, number]> = [];
  let i = 0;
  let j = 0;
  while (i < n && j < m) {
    if (old[i]!.eq(now[j]!)) {
      pairs.push([i, j]);
      i++;
      j++;
    } else if (table[i + 1]![j]! >= table[i]![j + 1]!) {
      i++;
    } else {
      j++;
    }
  }
  return pairs;
}

export function keepSource(
  source: string,
  blocks: Range[],
  before: Node,
  after: Node,
  write: (node: Node) => string,
): string | null {
  const old = topNodes(before);
  const now = topNodes(after);
  if (old.length !== blocks.length) return null;

  const head = blocks.length > 0 ? source.slice(0, blocks[0]!.from) : source.replace(/\s+$/, "");
  const tail = blocks.length > 0 ? source.slice(blocks[blocks.length - 1]!.to) : source.slice(head.length);
  let out = head;
  let wrote = false;
  let lastOld = -1;
  let lastNew = -1;

  for (const [i, j] of [...matched(old, now), [old.length, now.length] as [number, number]]) {
    for (let k = lastNew + 1; k < j; k++) {
      out += (wrote ? "\n\n" : "") + write(now[k]!).replace(/\n+$/, "");
      wrote = true;
    }
    if (i < old.length) {
      if (wrote) {
        const adjacent = i === lastOld + 1 && j === lastNew + 1 && i > 0;
        out += adjacent ? source.slice(blocks[i - 1]!.to, blocks[i]!.from) : "\n\n";
      }
      out += source.slice(blocks[i]!.from, blocks[i]!.to);
      wrote = true;
    }
    lastOld = i;
    lastNew = j;
  }

  if (blocks.length === 0 && wrote && !/\n$/.test(out)) return out + (tail || "\n");
  return out + tail;
}
