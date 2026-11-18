import { spokenNow } from "./say";

export function aNumber(value: number): string {
  return new Intl.NumberFormat(spokenNow()).format(value);
}
