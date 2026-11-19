import { describe, expect, it } from "vitest";

import { arrange } from "../useEditorRoom";

describe("the editor's room", () => {
  it("shows form and preview side by side only where both fit", () => {
    expect(arrange(1200, "split")).toEqual({ shown: "split", chat: "closed", splitFits: true });
    expect(arrange(700, "split")).toEqual({ shown: "form", chat: "closed", splitFits: false });
    expect(arrange(700, "preview").shown).toBe("preview");
  });

  it("puts the chat beside the preview where both fit, and alone where they do not", () => {
    expect(arrange(1200, "chat")).toEqual({ shown: "chat", chat: "column", splitFits: true });
    expect(arrange(820, "chat").chat).toBe("column");
    expect(arrange(819, "chat").chat).toBe("alone");
  });
});
