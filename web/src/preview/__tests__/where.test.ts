import { describe, expect, it } from "vitest";

import { fileOf, fitsTheWidth, languageOf, nextZoom, ownSize, pageWithin, percentOf, remember, remembered, scaleFor } from "../pdf/where";

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

describe("the size of a page", () => {
  it("is its own size unless asked to fit", () => {
    expect(scaleFor(420, 300, false)).toBe(ownSize);
    expect(fitsTheWidth(at("https://preview.example/v/k1/__chatcms/pdf/?file=a.pdf"))).toBe(false);
    expect(fitsTheWidth(at("https://preview.example/v/k1/__chatcms/pdf/?file=a.pdf&fit=width"))).toBe(true);
  });

  it("shrinks to the room, keeping a margin, and never grows past its own size", () => {
    expect(scaleFor(420, 390, true)).toBeCloseTo(366 / 420);
    expect(scaleFor(420, 2000, true)).toBe(ownSize);
    expect(scaleFor(0, 390, true)).toBe(ownSize);
  });
});

describe("the tools", () => {
  it("zoom by steps from any start, and stop at the ends", () => {
    expect(percentOf(nextZoom(ownSize, 1))).toBe(125);
    expect(percentOf(nextZoom(ownSize, -1))).toBe(75);
    expect(percentOf(nextZoom(ownSize * 0.87, 1))).toBe(100);
    expect(percentOf(nextZoom(ownSize * 0.87, -1))).toBe(75);
    expect(percentOf(nextZoom(ownSize * 3, 1))).toBe(300);
    expect(percentOf(nextZoom(ownSize * 0.5, -1))).toBe(50);
  });

  it("go to the nearest page there is", () => {
    expect(pageWithin(7, 12)).toBe(7);
    expect(pageWithin(40, 12)).toBe(12);
    expect(pageWithin(0, 12)).toBe(1);
    expect(pageWithin(Number("seven"), 12)).toBe(1);
  });

  it("speak the language the pane asks for", () => {
    expect(languageOf(at("https://preview.example/v/k1/__chatcms/pdf/?file=a.pdf&lang=nl"))).toBe("nl");
    expect(languageOf(at("https://preview.example/v/k1/__chatcms/pdf/?file=a.pdf"))).toBe("en");
  });
});
