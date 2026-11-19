import { readdirSync, readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { join } from "node:path";

import { test, expect, onAPhone, saves, worksOnAPhone } from "./fixtures.mjs";

const tricky = readFileSync(process.env.CHATCMS_MARKDOWN_TRICKY ?? fileURLToPath(new URL("../fixtures/markdown/tricky.md", import.meta.url)), "utf8");
const page = `---\ntitle: Tricky text\ndate: 2026-11-19\n---\n\n${tricky}`;

function savedCopies(chatcms, project, path) {
  const people = join(chatcms.state, "sites", project, "people");
  return readdirSync(people).map((person) => readFileSync(join(people, person, "working-copy", path), "utf8"));
}

test("the page text is edited formatted and keeps its own writing", async ({ signedIn: browser, chatcms }, testInfo) => {
  test.setTimeout(180_000);
  const project = chatcms.projectId();
  chatcms.pushFromSomewhereElse(project, "content/news/tricky.md", page, "A page with tricky text");
  await browser.goto(`${chatcms.url}/o/technative/p/korenaar/sync`);
  await browser.getByRole("button", { name: "Check for updates" }).click();
  await expect(browser.getByText(/Joris updated this site/)).toBeVisible({ timeout: 60_000 });

  await browser.goto(`${chatcms.url}/o/technative/p/korenaar/content/news/tricky`);
  const text = browser.getByRole("textbox", { name: "Page text" });

  await test.step("the breadcrumb keeps its place's name in a full header", async () => {
    const choose = browser.getByRole("navigation", { name: "Where you are" }).getByRole("button", { name: /^Choose another place, now News/ });
    if (!onAPhone(testInfo)) await expect(choose).toContainText("News");
    if (!onAPhone(testInfo)) expect((await choose.boundingBox()).width).toBeGreaterThanOrEqual(60);
  });

  await test.step("the text shows formatted, shortcodes as locked blocks", async () => {
    await expect(text.getByRole("heading", { name: "A heading" })).toBeVisible({ timeout: 30_000 });
    await expect(text.locator("strong").first()).toHaveText("strong");
    await expect(text.locator("[data-shortcode='figure']")).toHaveAttribute("contenteditable", "false");
    await expect(text.locator("[data-shortcode='notice']")).toBeVisible();
    if (process.env.CHATCMS_SHOTS) {
      await text.locator("[data-shortcode='figure']").scrollIntoViewIfNeeded();
      await browser.screenshot({ path: `${process.env.CHATCMS_SHOTS}/markdown-editor/${testInfo.project.name}.png` });
    }
  });

  await test.step("bold from the toolbar changes only that paragraph", async () => {
    const paragraph = text.locator("p", { hasText: "Escapes:" });
    await paragraph.click();
    await paragraph.evaluate((element) => {
      const word = element.firstChild;
      const range = document.createRange();
      range.setStart(word, 0);
      range.setEnd(word, "Escapes".length);
      const selection = window.getSelection();
      selection.removeAllRanges();
      selection.addRange(range);
    });
    await browser.waitForTimeout(100);
    await saves(browser, () => browser.getByRole("button", { name: "Bold" }).click());
    await expect(text.locator("strong", { hasText: "Escapes" })).toBeVisible();

    const expected = `\n${tricky}`.replace("Escapes: \\* not a list, 3 \\< 4", "**Escapes**: \\* not a list, 3 < 4");
    const bodies = () => savedCopies(chatcms, project, "content/news/tricky.md").map((file) => file.replace(/^---\n[\s\S]*?\n---\n/, ""));
    await expect.poll(bodies).toContain(expected);
  });

  await test.step("the Markdown view shows exactly what was saved", async () => {
    await browser.getByRole("radio", { name: "Markdown" }).click();
    await expect(browser.getByRole("textbox", { name: "Page text" })).toHaveValue(/\*\*Escapes\*\*: /);
    await expect(browser.getByRole("textbox", { name: "Page text" })).toHaveValue(/\{\{< figure src="\/images\/loaf.jpg"/);
  });

  if (onAPhone(testInfo)) await worksOnAPhone(browser, testInfo);
});
