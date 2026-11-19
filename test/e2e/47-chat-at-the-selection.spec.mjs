import { mkdirSync } from "node:fs";
import { join } from "node:path";

import { test, expect, onAPhone } from "./fixtures.mjs";
import { owner } from "./harness.mjs";

test.use({ assistant: true });

const shots = process.env.CHATCMS_SHOTS;

async function shoot(page, name) {
  if (!shots) return;
  mkdirSync(join(shots, "chat-at-the-selection"), { recursive: true });
  await page.screenshot({ path: join(shots, "chat-at-the-selection", `${name}.png`) });
}

test("a chip waits for the selection, opens an input there, and Esc keeps the draft", async ({ signedIn: page, chatcms }, testInfo) => {
  test.skip(onAPhone(testInfo), "the phone's selection comes with its own design");
  test.skip(!process.env.CHATCMS_OPENCODE_BIN, "no OpenCode to run the assistant with");
  test.setTimeout(180_000);
  await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/home`);
  const views = page.getByRole("radiogroup", { name: "How to show this page" });
  const preview = page.frameLocator('iframe[title="Your site"]');
  const words = preview.getByText("Come in and smell it.");
  const chip = preview.getByRole("button", { name: "Chat about the selected text" });

  await test.step("no chip in Form + preview", async () => {
    await views.getByRole("radio", { name: "Form + preview" }).click();
    await expect(words).toBeVisible({ timeout: 90_000 });
    await words.selectText();
    await page.waitForTimeout(600);
    await expect(chip).toBeHidden();
  });

  await test.step("in Preview the chip shows once the mouse is let go", async () => {
    await views.getByRole("radio", { name: "Preview", exact: true }).click();
    await expect(words).toBeVisible({ timeout: 90_000 });
    await words.dispatchEvent("pointerdown", { pointerType: "mouse", bubbles: true });
    await words.evaluate((element) => {
      const range = element.ownerDocument.createRange();
      range.selectNodeContents(element);
      element.ownerDocument.getSelection().removeAllRanges();
      element.ownerDocument.getSelection().addRange(range);
    });
    await page.waitForTimeout(600);
    await expect(chip).toBeHidden();
    await words.dispatchEvent("pointerup", { pointerType: "mouse", bubbles: true });
    await expect(chip).toBeVisible();
    await shoot(page, "01-chip");
  });

  await test.step("the chip opens the input below the selection", async () => {
    await chip.click();
    const card = preview.getByRole("dialog", { name: "Chat about the selected text" });
    await expect(card).toBeVisible();
    await expect(card).toHaveAttribute("data-above", "false");
    const words_ = await words.boundingBox();
    const cardBox = await card.boundingBox();
    expect(cardBox.y).toBeGreaterThan(words_.y + words_.height);
    await expect(card.getByRole("button", { name: "Fix spelling" })).toBeVisible();
    await card.getByRole("textbox", { name: "What should change?" }).fill("Half a thought");
    await shoot(page, "02-input");
    await page.keyboard.press("Escape");
    await expect(card).toBeHidden();
  });

  await test.step("the draft waits in the chat with the selection quoted", async () => {
    await views.getByRole("radio", { name: /^Chat \+ preview/ }).click();
    const panel = page.getByRole("region", { name: "Assistant" });
    await expect(panel.getByRole("textbox", { name: "Your message" })).toHaveValue("Half a thought");
    await expect(panel.getByLabel("About this part of the page")).toContainText("Come in and smell it");
  });
});

test("a viewer gets no chip", async ({ browser, chatcms }, testInfo) => {
  test.skip(onAPhone(testInfo), "the phone's selection comes with its own design");
  test.setTimeout(180_000);
  const context = await browser.newContext({ viewport: { width: 1440, height: 900 } });
  const page = await context.newPage();
  try {
    await page.goto(`${chatcms.url}/sign-in`);
    await page.getByLabel("Email").fill("sam@korenaar.example");
    await page.getByLabel("Password").fill(owner.password);
    await page.getByRole("button", { name: /Sign in/i }).click();
    await page.waitForURL(/projects/);
    await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/home`);
    await page.getByRole("radiogroup", { name: "How to show this page" }).getByRole("radio", { name: "Preview", exact: true }).click();
    const preview = page.frameLocator('iframe[title="Your site"]');
    const words = preview.getByText("Come in and smell it.");
    await expect(words).toBeVisible({ timeout: 90_000 });
    await words.selectText();
    await page.waitForTimeout(600);
    await expect(preview.getByRole("button", { name: "Chat about the selected text" })).toBeHidden();
  } finally {
    await context.close();
  }
});
