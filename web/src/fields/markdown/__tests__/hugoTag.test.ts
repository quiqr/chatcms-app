import { describe, expect, it } from "vitest";

import { build, readTag, rewrite, valuesOf, type Declared } from "../hugoTag";

const figure: Declared = { key: "figure", order: ["src", "caption", "alt"], positions: {} };
const notice: Declared = { key: "notice", inner: "markdown", order: ["type"], positions: { type: 0 } };
const icon: Declared = { key: "icon", order: ["name"], positions: { name: 0 } };

describe("reading a Hugo tag", () => {
  it("reads named, positional, quoted, backticked and bare parameters", () => {
    const tag = readTag('{{< figure src="/a.jpg" class=wide caption=`Our "loaf"` >}}')!;
    expect(tag.name).toBe("figure");
    expect(tag.params.map((p) => [p.name, p.value, p.quote])).toEqual([
      ["src", "/a.jpg", '"'],
      ["class", "wide", ""],
      ["caption", 'Our "loaf"', "`"],
    ]);
    expect(readTag('{{< icon "bread" >}}')!.params[0]).toMatchObject({ name: undefined, value: "bread" });
  });

  it("reads escaped quotes and the text between paired tags", () => {
    expect(readTag('{{< x a="say \\"hi\\"" >}}')!.params[0]!.value).toBe('say "hi"');
    const raw = "{{% notice info %}}\nInside **this**.\n{{% /notice %}}";
    expect(valuesOf(raw, notice)).toEqual({ values: { type: "info" }, inner: "Inside **this**." });
  });

  it("refuses what it cannot read rather than guessing", () => {
    expect(readTag('{{< figure src="/a.jpg >}}')).toBeNull();
    expect(readTag("{{< figure >}} trailing words")).toBeNull();
    expect(readTag("{{< /figure >}}")).toBeNull();
  });
});

describe("writing a Hugo tag back", () => {
  const raw = '{{< figure src="/images/loaf.jpg" class="wide"  caption="Our loaf" >}}';

  it("changes only the value that changed, keeping everything else", () => {
    expect(rewrite(raw, figure, { src: "/images/loaf.jpg", caption: "Fresh loaf" }, "")).toBe(
      '{{< figure src="/images/loaf.jpg" class="wide"  caption="Fresh loaf" >}}',
    );
  });

  it("adds a new value at the end and removes a cleared one", () => {
    expect(rewrite(raw, figure, { src: "/images/loaf.jpg", caption: "Our loaf", alt: "A loaf" }, "")).toBe(
      '{{< figure src="/images/loaf.jpg" class="wide"  caption="Our loaf" alt="A loaf" >}}',
    );
    expect(rewrite(raw, figure, { src: "/images/loaf.jpg" }, "")).toBe('{{< figure src="/images/loaf.jpg" class="wide" >}}');
  });

  it("keeps a bare value bare, and quotes it when it needs quotes", () => {
    expect(rewrite("{{< icon bread >}}", icon, { name: "rye" }, "")).toBe("{{< icon rye >}}");
    expect(rewrite("{{< icon bread >}}", icon, { name: "dark rye" }, "")).toBe('{{< icon "dark rye" >}}');
  });

  it("changes a positional value and the text between", () => {
    const paired = "{{% notice info %}}\nOld text.\n{{% /notice %}}";
    expect(rewrite(paired, notice, { type: "warning" }, "New text.")).toBe("{{% notice warning %}}\nNew text.\n{{% /notice %}}");
  });

  it("leaves an untouched tag exactly as it was", () => {
    expect(rewrite(raw, figure, { src: "/images/loaf.jpg", caption: "Our loaf" }, "")).toBe(raw);
  });
});

describe("building a new Hugo tag", () => {
  it("writes named values in the model's order, leaving out empty ones", () => {
    expect(build(figure, { caption: "A loaf", src: "/a.jpg" }, "")).toBe('{{< figure src="/a.jpg" caption="A loaf" >}}');
  });

  it("uses % and a closing tag for Markdown between the tags", () => {
    expect(build(notice, { type: "tip" }, "Bake **early**.")).toBe("{{% notice tip %}}\nBake **early**.\n{{% /notice %}}");
  });
});
