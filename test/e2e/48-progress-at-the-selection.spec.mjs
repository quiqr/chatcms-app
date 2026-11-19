import { mkdirSync } from "node:fs";
import { join } from "node:path";

import { test, expect, onAPhone } from "./fixtures.mjs";
import { realAI } from "./harness.mjs";

test.use({ assistant: true, previewApart: true });

const shots = process.env.CHATCMS_SHOTS;
async function shoot(page, name) {
  if (!shots) return;
  mkdirSync(join(shots, "progress-at-the-selection"), { recursive: true });
  await page.screenshot({ path: join(shots, "progress-at-the-selection", `${name}.png`) });
}

async function ask(preview, words, prompt) {
  await words.evaluate((element) => {
    const range = element.ownerDocument.createRange();
    range.selectNodeContents(element);
    element.ownerDocument.getSelection().removeAllRanges();
    element.ownerDocument.getSelection().addRange(range);
  });
  await preview.getByRole("button", { name: "Chat about the selected text" }).click();
  const card = preview.getByRole("dialog", { name: "Chat about the selected text" });
  await card.getByRole("textbox", { name: "What should change?" }).fill(prompt);
  await card.getByRole("textbox", { name: "What should change?" }).press("Enter");
  await expect(card).toBeHidden();
}

test("a request shows its progress at the selection, and Undo there puts it back", async ({ signedIn: page, chatcms }, testInfo) => {
  test.skip(onAPhone(testInfo), "the phone's selection comes with its own design");
  test.skip(!process.env.CHATCMS_OPENCODE_BIN, "no OpenCode to run the assistant with");
  test.setTimeout(300_000);
  await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/home`);
  const views = page.getByRole("radiogroup", { name: "How to show this page" });
  await views.getByRole("radio", { name: "Preview", exact: true }).click();
  const preview = page.frameLocator('iframe[title="Your site"]');
  const heading = preview.getByText("Korenaar Bakery").first();
  await expect(heading).toBeVisible({ timeout: 90_000 });

  await test.step("two requests give two pills and a count on Chat + preview", async () => {
    await ask(preview, heading, "Change it to Korenaar by night");
    await ask(preview, preview.getByText("Come in and smell it."), "Make it warmer");
    await expect(preview.getByRole("status").filter({ hasText: "Working on it…" })).toHaveCount(2);
    await expect(views.getByRole("radio", { name: /^Chat \+ preview\s?[12]$/ })).toBeVisible();
    await shoot(page, "07-several");
  });

  await test.step("the first turns into Changed with Undo", async () => {
    const changed = preview.getByRole("status").filter({ hasText: "Changed" }).first();
    await expect(changed).toBeVisible({ timeout: 150_000 });
    await expect(preview.locator("body")).toContainText("Korenaar by night", { timeout: 60_000 });
    await shoot(page, "05-done");
  });

  await test.step("Stop takes the waiting one away", async () => {
    const waiting = preview.getByRole("status").filter({ hasText: "Working on it…" });
    if ((await waiting.count()) > 0) {
      await waiting.first().getByRole("button", { name: "Stop" }).click();
      await expect(waiting).toHaveCount(0, { timeout: 30_000 });
    }
  });

  await test.step("Undo at the selection puts the title back", async () => {
    await preview.getByRole("status").filter({ hasText: "Changed" }).first().getByRole("button", { name: "Undo" }).click();
    await expect(async () => {
      await page.getByRole("button", { name: "Refresh" }).click();
      await expect(preview.locator("body")).toContainText("Korenaar Bakery", { timeout: 2_000 });
    }).toPass({ timeout: 60_000, intervals: [500, 1_000, 2_000] });
    await views.getByRole("radio", { name: /^Chat \+ preview/ }).click();
    await expect(page.getByRole("region", { name: "Assistant" }).getByText(/is back as it was/).first()).toBeVisible();
  });

  if (realAI) return;
});
