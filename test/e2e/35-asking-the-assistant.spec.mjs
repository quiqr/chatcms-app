// Milestone 07, epic 6: "change X on page Y" through chat changes the page,
// shows in the preview and is saved; Undo puts it back.
import { test, expect, onAPhone, worksOnAPhone } from "./fixtures.mjs";

test.use({ assistant: true });

test("the assistant changes a page when asked, and Undo puts it back", async ({ signedIn: page, chatcms }, testInfo) => {
  test.skip(!process.env.CHATCMS_OPENCODE_BIN, "no OpenCode to run the assistant with");
  test.setTimeout(300_000);
  const phone = onAPhone(testInfo);

  await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/home`);
  await expect(page.getByRole("textbox", { name: "Page title" })).toHaveValue("Korenaar Bakery");

  await page.getByRole("button", { name: "Open the assistant" }).click();
  const panel = page.getByRole("region", { name: "Assistant" });
  const message = panel.getByRole("textbox", { name: "Your message" });

  await test.step("asking for a change", async () => {
    await message.fill("Change the title of the home page to Korenaar by night");
    await message.press("Enter");
    await expect(panel.getByText("Done. The home page has its new title.")).toBeVisible({ timeout: 150_000 });
    await expect(panel.getByText(/^Changed /)).toBeVisible();
  });

  const reopen = async () => {
    if (phone) await page.getByRole("button", { name: "Close the assistant" }).click();
  };

  await test.step("the page holds the new title, and the preview shows it", async () => {
    await reopen();
    await expect(page.getByRole("textbox", { name: "Page title" })).toHaveValue("Korenaar by night", { timeout: 20_000 });
    if (phone) await page.getByRole("radio", { name: "Preview" }).click();
    const preview = page.frameLocator('iframe[title="Your site"]');
    await expect(async () => {
      await page.getByRole("button", { name: "Refresh" }).click();
      await expect(preview.locator("body")).toContainText("Korenaar by night", { timeout: 2_000 });
    }).toPass({ timeout: 60_000, intervals: [500, 1_000, 2_000] });
    if (phone) await page.getByRole("radio", { name: "Edit" }).click();
  });

  await test.step("Undo puts it back without publishing", async () => {
    if (phone) await page.getByRole("button", { name: "Open the assistant" }).click();
    await panel.getByRole("button", { name: "Undo" }).click();
    await expect(panel.getByText(/is back as it was/)).toBeVisible({ timeout: 20_000 });
    await reopen();
    await expect(page.getByRole("textbox", { name: "Page title" })).toHaveValue("Korenaar Bakery", { timeout: 20_000 });
  });

  expect(chatcms.asked.some((asked) => (asked.tools ?? []).some((tool) => /update_page$/.test(tool.function?.name ?? "")))).toBe(true);
  await worksOnAPhone(page, testInfo);
});

test.describe("without a gateway", () => {
  test.use({ assistant: false });

  test("the panel says the assistant is not available", async ({ signedIn: page, chatcms }) => {
    await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/home`);
    await page.getByRole("button", { name: "Open the assistant" }).click();
    await expect(page.getByText("The assistant is not available on this installation.")).toBeVisible();
  });
});
