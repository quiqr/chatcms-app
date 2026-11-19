import { describe, expect, it } from "vitest";

import { arrange } from "../useEditorRoom";

describe("what the editor shows in the room it has", () => {
  it("shows form and preview side by side only where both fit", () => {
    expect(arrange(1200, "split", false)).toEqual({ shown: "split", assistant: "closed", splitFits: true });
    expect(arrange(700, "split", false)).toEqual({ shown: "form", assistant: "closed", splitFits: false });
    expect(arrange(700, "preview", false).shown).toBe("preview");
  });

  it("gives the assistant a column where it fits beside what is shown", () => {
    expect(arrange(1200, "split", true).assistant).toBe("column");
    expect(arrange(800, "form", true).assistant).toBe("column");
  });

  it("puts the assistant in the form's place where it does not fit, so the preview stays beside it", () => {
    expect(arrange(1000, "split", true)).toEqual({ shown: "preview", assistant: "instead-of-form", splitFits: true });
    expect(arrange(800, "preview", true).assistant).toBe("column");
    expect(arrange(700, "split", true).assistant).toBe("alone");
  });
});
