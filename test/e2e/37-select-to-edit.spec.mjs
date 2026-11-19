import { test, expect, onAPhone } from "./fixtures.mjs";
import { realAI } from "./harness.mjs";

test.use({ assistant: true, previewApart: true });

test("a sentence selected in the preview is changed on its own page, and the change is highlighted", async ({ signedIn: page, chatcms }, testInfo) => {
  test.skip(!process.env.CHATCMS_OPENCODE_BIN, "no OpenCode to run the assistant with");
  test.setTimeout(300_000);
  const phone = onAPhone(testInfo);
  const showThePreview = async () => {
    if (phone) await page.getByRole("radio", { name: "Preview" }).click();
    else await page.getByRole("radio", { name: /^Chat \+ preview/ }).click();
  };

  await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/home`);
  await showThePreview();
  const preview = page.frameLocator('iframe[title="Your site"]');
  const heading = preview.getByText("Korenaar Bakery").first();
  await expect(heading).toBeVisible({ timeout: 90_000 });

  await test.step("selecting the text offers a chip, and the chip an input at the selection", async () => {
    await heading.evaluate((element) => {
      const range = element.ownerDocument.createRange();
      range.selectNodeContents(element);
      const selection = element.ownerDocument.getSelection();
      selection.removeAllRanges();
      selection.addRange(range);
    });
    await preview.getByRole("button", { name: "Chat about the selected text" }).click();
    const card = preview.getByRole("dialog", { name: "Chat about the selected text" });
    await card.getByRole("textbox", { name: "What should change?" }).fill("Change it to Korenaar by night");
    await card.getByRole("textbox", { name: "What should change?" }).press("Enter");
    await expect(card).toBeHidden();
  });

  const panel = page.getByRole("region", { name: "Assistant" });
  await test.step("the request is in the chat with its selection, and changes that page", async () => {
    if (phone) await page.getByRole("button", { name: "Open the assistant" }).click();
    await expect(panel.getByRole("button", { name: /^Show in the preview: Korenaar Bakery/ })).toBeVisible();
    if (!realAI) await expect(panel.getByText("Done. The home page has its new title.")).toBeVisible({ timeout: 150_000 });
    await expect(panel.getByText(/^Changed /)).toBeVisible({ timeout: 150_000 });
    if (realAI) return;
    const asked = chatcms.asked.flatMap((one) => one.messages ?? []).map((one) => (typeof one.content === "string" ? one.content : JSON.stringify(one.content)));
    expect(asked.some((text) => text.includes('with key "home"') && text.includes("Korenaar Bakery"))).toBe(true);
  });

  await test.step("the preview shows the new text, highlighted", async () => {
    if (phone) await page.getByRole("button", { name: "Close the assistant" }).click();
    await showThePreview();
    await expect(async () => {
      await expect(preview.locator("body")).toContainText("Korenaar by night", { timeout: 2_000 });
      const lit = await preview.locator("body").evaluate(() => CSS.highlights?.has("chatcms-changed") ?? false);
      expect(lit).toBe(true);
    }).toPass({ timeout: 60_000, intervals: [250] });
  });
});
