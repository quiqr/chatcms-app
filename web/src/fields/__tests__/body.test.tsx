import { render, screen, waitFor } from "@testing-library/react";
import { beforeAll, describe, expect, it, vi } from "vitest";

import { BodyEditor } from "../BodyEditor";

vi.mock("../markdown/keep", async (original) => {
  const real = await original<typeof import("../markdown/keep")>();
  return { ...real, linesUp: (blocks: unknown[], doc: { textContent: string }) => !doc.textContent.includes("refuse me") && real.linesUp(blocks as never, doc as never) };
});

describe("the page text", () => {
  beforeAll(async () => {
    await import("../markdown/MarkdownEditor");
  }, 30_000);

  it("takes a new text from outside without saving it back", async () => {
    const changed = vi.fn();
    const { rerender } = render(<BodyEditor value={"First words.\n"} onChange={changed} />);
    expect(await screen.findByRole("textbox", { name: "Page text" })).toHaveTextContent("First words.");

    rerender(<BodyEditor value={"Other words.\n"} onChange={changed} />);
    await waitFor(() => expect(screen.getByRole("textbox", { name: "Page text" })).toHaveTextContent("Other words."));
    expect(changed).not.toHaveBeenCalled();
  });

  it("opens text it cannot hold as Markdown only, and says so", async () => {
    render(<BodyEditor value={"Please refuse me.\n"} onChange={() => {}} />);
    expect(await screen.findByText(/can only be edited as Markdown/)).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "Page text" })).toHaveValue("Please refuse me.\n");
    expect(screen.queryByRole("radio", { name: "Formatted" })).toBeNull();
    expect(screen.queryByRole("toolbar", { name: "Text formatting" })).toBeNull();
  });
});
