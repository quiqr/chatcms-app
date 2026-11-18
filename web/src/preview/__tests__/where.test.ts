import { describe, expect, it } from "vitest";

import { fileOf, remember, remembered } from "../pdf/where";

function at(href: string) {
  return new URL(href) as unknown as Location;
}

class Memory {
  kept = new Map<string, string>();
  getItem(key: string) {
    return this.kept.get(key) ?? null;
  }
  setItem(key: string, value: string) {
    this.kept.set(key, value);
  }
}

describe("the PDF viewer's file", () => {
  it("is found under the preview's own address", () => {
    expect(fileOf(at("https://preview.example/v/k1/__chatcms/pdf/?file=report.pdf"))?.href).toBe("https://preview.example/v/k1/report.pdf");
    expect(fileOf(at("https://preview.example/v/k1/__chatcms/pdf/?file=%2Fbook%2Fall.pdf"))?.href).toBe("https://preview.example/v/k1/book/all.pdf");
  });

  it("is never taken from elsewhere", () => {
    expect(fileOf(at("https://preview.example/v/k1/__chatcms/pdf/?file=https%3A%2F%2Fevil.example%2Fx.pdf"))).toBeNull();
    expect(fileOf(at("https://preview.example/v/k1/__chatcms/pdf/?file=..%2F..%2Fv%2Fk2%2Fx.pdf"))).toBeNull();
    expect(fileOf(at("https://preview.example/somewhere/?file=x.pdf"))).toBeNull();
  });
});

describe("the page the viewer opens at", () => {
  const file = new URL("https://preview.example/v/k1/report.pdf");

  it("is the one asked for, then the one last seen, then the first", () => {
    const storage = new Memory() as unknown as Storage;
    expect(remembered(storage, file, "")).toBe(1);
    remember(storage, file, 3);
    expect(remembered(storage, file, "")).toBe(3);
    expect(remembered(storage, file, "#page=5")).toBe(5);
    expect(remembered(null, file, "")).toBe(1);
  });
});
