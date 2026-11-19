import { mkdirSync } from "node:fs";
import { join } from "node:path";

import { test, expect, onAPhone } from "./fixtures.mjs";
import { realAI } from "./harness.mjs";

test.use({ assistant: true, previewApart: true });

test("Chat + preview shows the chat beside the preview, with a request card that finds its place", async ({ signedIn: page, chatcms }, testInfo) => {
  test.skip(onAPhone(testInfo), "the phone keeps its own chat until its design is built");
  test.skip(!process.env.CHATCMS_OPENCODE_BIN, "no OpenCode to run the assistant with");
  test.setTimeout(300_000);

  await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/home`);
  const views = page.getByRole("radiogroup", { name: "How to show this page" });
  await expect(views.getByRole("radio")).toHaveText(["Form", "Form + preview", "Chat + preview", "Preview"]);
  await expect(page.getByRole("button", { name: "Open the assistant" })).toHaveCount(0);

  await test.step("the chat stands beside the preview, without the form, and is remembered", async () => {
    await views.getByRole("radio", { name: /^Chat \+ preview/ }).click();
    const chat = page.getByTestId("editor-slot-chat");
    await expect(chat).toBeVisible();
    expect(Math.round((await chat.boundingBox()).width)).toBe(400);
    await expect(page.getByTestId("editor-slot-form")).toHaveCount(0);
    await expect(page.getByTestId("editor-slot-preview")).toBeVisible();
    await page.reload();
    await expect(page.getByTestId("editor-slot-chat")).toBeVisible();
  });

  const preview = page.frameLocator('iframe[title="Your site"]');
  const heading = preview.getByText("Korenaar Bakery").first();
  await expect(heading).toBeVisible({ timeout: 90_000 });

  await test.step("a selection becomes a request card", async () => {
    await heading.evaluate((element) => {
      const range = element.ownerDocument.createRange();
      range.selectNodeContents(element);
      const selection = element.ownerDocument.getSelection();
      selection.removeAllRanges();
      selection.addRange(range);
    });
    await preview.getByRole("button", { name: "Ask the assistant about this" }).click();
    const panel = page.getByRole("region", { name: "Assistant" });
    await panel.getByRole("textbox", { name: "Your message" }).fill("Change it to Korenaar by night");
    await panel.getByRole("textbox", { name: "Your message" }).press("Enter");
    const card = panel.getByRole("button", { name: /^Show in the preview: Korenaar Bakery/ });
    await expect(card).toBeVisible();
    await expect(card.locator("mark")).toHaveText("Korenaar Bakery");
    await expect(panel.getByText(/^From the preview · /)).toBeVisible();
    if (!realAI) await expect(panel.getByText("Done. The home page has its new title.")).toBeVisible({ timeout: 150_000 });
    await expect(panel.getByRole("button", { name: "Show changes" })).toBeVisible({ timeout: 150_000 });
    if (process.env.CHATCMS_SHOTS) {
      mkdirSync(join(process.env.CHATCMS_SHOTS, "chat-and-preview"), { recursive: true });
      await page.screenshot({ path: join(process.env.CHATCMS_SHOTS, "chat-and-preview", "desk.png") });
    }
  });

  await test.step("choosing the card marks its place in the preview", async () => {
    await expect(preview.locator("body")).toContainText("Korenaar by night", { timeout: 60_000 });
    await expect
      .poll(() => preview.locator("body").evaluate(() => CSS.highlights?.has("chatcms-changed") ?? false), { timeout: 30_000 })
      .toBe(false);
    await page.getByRole("region", { name: "Assistant" }).getByRole("button", { name: /^Show in the preview/ }).click();
    await expect.poll(() => preview.locator("body").evaluate(() => CSS.highlights?.has("chatcms-changed") ?? false)).toBe(true);
  });
});
