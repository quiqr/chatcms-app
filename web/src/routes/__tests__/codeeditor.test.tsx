import { render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import CodeEditor from "../files/CodeEditor";
import { extensionOf, languageFor } from "../files/languages";

describe("the code editor", () => {
  it("shows the text with line numbers and marks Go template actions", async () => {
    render(<CodeEditor path="layouts/index.html" text={"<h1>{{ .Title }}</h1>\n<p>{{- .Summary -}}</p>"} label="Text of layouts/index.html" onChange={vi.fn()} onSave={vi.fn()} />);
    const editor = screen.getByTestId("code-editor");
    await waitFor(() => expect(editor.querySelectorAll(".cm-go-template")).toHaveLength(2));
    expect(editor.querySelector(".cm-lineNumbers")).not.toBeNull();
    expect(screen.getByRole("textbox", { name: "Text of layouts/index.html" })).toBeInTheDocument();
  });

  it("chooses a language by extension", () => {
    for (const path of ["a.html", "b.css", "c.js", "d.ts", "e.json", "f.yaml", "g.yml", "h.toml", "i.md", "j.qmd"]) {
      expect(languageFor(path).length, path).toBeGreaterThan(0);
    }
    expect(languageFor("notes.txt")).toEqual([]);
    expect(extensionOf("layouts/_default/baseof.HTML")).toBe("html");
  });
});
