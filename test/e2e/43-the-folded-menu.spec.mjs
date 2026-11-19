import { mkdirSync } from "node:fs";
import { join } from "node:path";

import { test, expect, onAPhone } from "./fixtures.mjs";

const into = process.env.CHATCMS_SHOTS;

async function shoot(page, name) {
  if (!into) return;
  mkdirSync(join(into, "folded-menu"), { recursive: true });
  await page.screenshot({ path: join(into, "folded-menu", `${name}.png`) });
}

test("the folded column names where you are, and the picker goes anywhere", async ({ signedIn: page, chatcms }, testInfo) => {
  test.skip(onAPhone(testInfo), "a phone has no column to fold and no breadcrumb");
  const content = `${chatcms.url}/o/technative/p/korenaar/content`;
  const panel = page.locator("[data-panel='content']");
  const where = page.getByRole("navigation", { name: "Where you are" });

  await test.step("folded, the strip says what is open", async () => {
    await page.goto(`${content}/c/news`);
    await page.getByRole("button", { name: "Hide the content menu" }).click();
    await expect(panel).toHaveAttribute("data-folded", "true");
    const strip = page.getByRole("button", { name: /^Show the content menu \(News/ });
    await expect(strip).toBeVisible();
    await expect.poll(async () => Math.round((await panel.boundingBox()).width)).toBe(48);
    await expect(where).toContainText("News");
    await shoot(page, "folded");
  });

  await test.step("the picker switches while the column stays folded", async () => {
    await where.getByRole("button", { name: /^Choose another place/ }).click();
    const picker = page.getByRole("dialog", { name: "Go to" });
    await expect(picker.locator("[aria-current='page']")).toContainText("News");
    await shoot(page, "picker");
    await picker.getByRole("button", { name: /Home/ }).first().click();
    await page.waitForURL(/\/content\/home$/);
    await expect(picker).toHaveCount(0);
    await expect(panel).toHaveAttribute("data-folded", "true");
  });

  await test.step("by keyboard: open, move, close, and focus comes back", async () => {
    const opener = where.getByRole("button", { name: /^Choose another place/ });
    await opener.focus();
    await page.keyboard.press("Enter");
    const picker = page.getByRole("dialog", { name: "Go to" });
    await expect(picker.getByRole("searchbox", { name: "Title" })).toBeFocused();
    await page.keyboard.press("ArrowDown");
    await expect(picker.locator("[data-picker-entry]").first()).toBeFocused();
    await page.keyboard.press("Escape");
    await expect(picker).toHaveCount(0);
    await expect(opener).toBeFocused();
  });

  await test.step("switching between a collection and a page does not blank the screen", async () => {
    await page.goto(`${content}/c/news`);
    await expect(where).toContainText("News");
    await page.evaluate(() => {
      window.__blanked = false;
      new MutationObserver(() => {
        if (!document.querySelector("[data-testid='project-shell'], [data-testid='editor-shell']")) window.__blanked = true;
      }).observe(document.body, { childList: true, subtree: true });
    });
    await where.getByRole("button", { name: /^Choose another place/ }).click();
    await page.getByRole("dialog", { name: "Go to" }).getByRole("button", { name: /Home/ }).first().click();
    await page.waitForURL(/\/content\/home$/);
    await expect(page.getByTestId("editor-shell")).toBeVisible();
    await where.getByRole("button", { name: /^Choose another place/ }).click();
    await page.getByRole("dialog", { name: "Go to" }).getByRole("button", { name: /News/ }).first().click();
    await page.waitForURL(/\/content\/c\/news$/);
    await expect(page.getByTestId("project-shell")).toBeVisible();
    expect(await page.evaluate(() => window.__blanked)).toBe(false);
  });

  await test.step("an item keeps its position beside the breadcrumb", async () => {
    await page.goto(`${content}/c/news`);
    await page.getByRole("listitem").getByRole("link").first().click();
    await page.waitForURL(/\/content\/news\//);
    await expect(where).toContainText(/News· 1 of \d+/);
  });

  await test.step("the footer unfolds the column", async () => {
    await where.getByRole("button", { name: /^Choose another place/ }).click();
    await page.getByRole("button", { name: "Show the full content menu" }).click();
    await expect(panel).toHaveAttribute("data-folded", "false");
  });
});
