import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeAll, describe, expect, it, vi } from "vitest";

import type { ShortcodeDeclaration } from "../../lib/tenancy";
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

const figure: ShortcodeDeclaration = {
  key: "figure",
  title: "Picture",
  fields: [
    { path: "src", type: "string", title: "Address" },
    { path: "caption", type: "string", title: "Caption" },
  ],
};

describe("a declared shortcode", () => {
  beforeAll(async () => {
    await import("../markdown/MarkdownEditor");
  }, 30_000);

  const source = 'Before.\n\n{{< figure src="/a.jpg" class="wide" caption="Our loaf" >}}\n';

  it("opens its form, and Done writes back only what changed", async () => {
    const changed = vi.fn();
    const user = userEvent.setup();
    render(<BodyEditor value={source} onChange={changed} dialect="hugo" shortcodes={[figure]} />);
    await user.click(await screen.findByRole("button", { name: "Edit Picture" }));
    const caption = screen.getByRole("textbox", { name: "Caption" });
    expect(caption).toHaveValue("Our loaf");
    await user.clear(caption);
    await user.type(caption, "Fresh loaf");
    await user.click(screen.getByRole("button", { name: "Done" }));
    expect(changed).toHaveBeenLastCalledWith('Before.\n\n{{< figure src="/a.jpg" class="wide" caption="Fresh loaf" >}}\n');
    expect(screen.queryByRole("dialog")).toBeNull();
  });

  it("changes nothing on Cancel", async () => {
    const changed = vi.fn();
    const user = userEvent.setup();
    render(<BodyEditor value={source} onChange={changed} dialect="hugo" shortcodes={[figure]} />);
    await user.click(await screen.findByRole("button", { name: "Edit Picture" }));
    await user.type(screen.getByRole("textbox", { name: "Caption" }), "!");
    await user.click(screen.getByRole("button", { name: "Cancel" }));
    expect(changed).not.toHaveBeenCalled();
  });

  it("is inserted from the toolbar, and the menu is there only with declarations", async () => {
    const changed = vi.fn();
    const user = userEvent.setup();
    const { unmount } = render(<BodyEditor value={"First.\n"} onChange={changed} dialect="hugo" shortcodes={[figure]} />);
    await screen.findByRole("textbox", { name: "Page text" });
    await user.click(screen.getByRole("button", { name: "Insert" }));
    await user.click(screen.getByRole("menuitem", { name: "Picture" }));
    await user.type(screen.getByRole("textbox", { name: "Address" }), "/b.jpg");
    await user.click(screen.getByRole("button", { name: "Done" }));
    expect(changed).toHaveBeenLastCalledWith('First.\n\n{{< figure src="/b.jpg" >}}\n');
    unmount();

    render(<BodyEditor value={"First.\n"} onChange={() => {}} dialect="hugo" shortcodes={[]} />);
    await screen.findByRole("textbox", { name: "Page text" });
    expect(screen.queryByRole("button", { name: "Insert" })).toBeNull();
  });
});
