import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { FileTree } from "../files/FileTree";

const entries = [
  { path: "hugo.toml" },
  { path: "layouts", dir: true },
  { path: "layouts/index.html" },
  { path: "layouts/partials", dir: true },
  { path: "layouts/partials/head.html" },
  { path: "content", dir: true },
  { path: "content/_index.md" },
];

function tree(onOpen = vi.fn(), chosen = "") {
  render(<FileTree entries={entries} label="Files of the site" chosen={chosen} onOpen={onOpen} />);
  return onOpen;
}

describe("the file tree", () => {
  it("lists folders first and has one tab stop", () => {
    tree();
    const items = screen.getAllByRole("treeitem");
    expect(items.map((item) => item.textContent)).toEqual(["content", "layouts", "hugo.toml"]);
    expect(items.filter((item) => item.tabIndex === 0)).toHaveLength(1);
    expect(items[0]).toHaveAttribute("aria-expanded", "false");
    expect(items[2]).not.toHaveAttribute("aria-expanded");
  });

  it("is walked with the arrow keys and opens a file with Enter", async () => {
    const onOpen = tree();
    await userEvent.tab();
    expect(screen.getByRole("treeitem", { name: "content" })).toHaveFocus();
    await userEvent.keyboard("{ArrowDown}");
    expect(screen.getByRole("treeitem", { name: "layouts" })).toHaveFocus();
    await userEvent.keyboard("{ArrowRight}");
    expect(screen.getByRole("treeitem", { name: "layouts" })).toHaveAttribute("aria-expanded", "true");
    await userEvent.keyboard("{ArrowRight}");
    expect(screen.getByRole("treeitem", { name: "partials" })).toHaveFocus();
    await userEvent.keyboard("{ArrowDown}");
    expect(screen.getByRole("treeitem", { name: "index.html" })).toHaveFocus();
    await userEvent.keyboard("{Enter}");
    expect(onOpen).toHaveBeenCalledWith("layouts/index.html");
    await userEvent.keyboard("{ArrowLeft}");
    expect(screen.getByRole("treeitem", { name: "layouts" })).toHaveFocus();
    await userEvent.keyboard("{ArrowLeft}");
    expect(screen.getByRole("treeitem", { name: "layouts" })).toHaveAttribute("aria-expanded", "false");
    await userEvent.keyboard("{End}");
    expect(screen.getByRole("treeitem", { name: "hugo.toml" })).toHaveFocus();
  });

  it("opens the folders of the chosen file and marks it", () => {
    tree(vi.fn(), "layouts/partials/head.html");
    expect(screen.getByRole("treeitem", { name: "head.html" })).toHaveAttribute("aria-selected", "true");
    expect(screen.getByRole("treeitem", { name: "head.html" })).toHaveAttribute("aria-level", "3");
  });
});
