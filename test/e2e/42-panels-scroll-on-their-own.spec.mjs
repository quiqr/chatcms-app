import { test, expect, onAPhone } from "./fixtures.mjs";

test.use({ assistant: true });

async function inView(page, locator) {
  const box = await locator.boundingBox();
  const { height } = page.viewportSize();
  expect(box, "the element is not shown at all").not.toBeNull();
  expect(box.y).toBeGreaterThanOrEqual(-1);
  expect(box.y + box.height).toBeLessThanOrEqual(height + 1);
}

test("only the form scrolls; the header, the preview's end and the message box stay in view", async ({ signedIn: page, chatcms }, testInfo) => {
  test.skip(onAPhone(testInfo), "the phone frame keeps its own layout");
  await page.setViewportSize({ width: 1440, height: 600 });
  await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/home`);
  await expect(page.getByRole("textbox", { name: "Page title" })).toBeVisible();

  const form = page.getByTestId("editor-slot-form");
  const tall = await form.evaluate((element) => element.scrollHeight > element.clientHeight);
  expect(tall, "the form fits the window, so this proves nothing").toBe(true);

  await form.evaluate((element) => element.scrollTo(0, element.scrollHeight));
  expect(await page.evaluate(() => document.scrollingElement.scrollTop)).toBe(0);
  await inView(page, page.getByRole("heading", { level: 1 }).first());
  await inView(page, page.locator('iframe[title="Your site"]').or(page.getByRole("status").filter({ hasText: /preview/i })).first());

  await page.getByRole("button", { name: "Open the assistant" }).click();
  await inView(page, page.getByRole("textbox", { name: "Your message" }));
});

test("a window too narrow for form and preview shows the form, with the preview one choice away", async ({ signedIn: page, chatcms }, testInfo) => {
  test.skip(onAPhone(testInfo), "the phone frame keeps its own layout");
  await page.setViewportSize({ width: 900, height: 800 });
  await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/home`);
  await expect(page.getByRole("textbox", { name: "Page title" })).toBeVisible();
  await expect(page.getByTestId("editor-slot-preview")).toHaveCount(0);

  const views = page.getByRole("radiogroup", { name: "How to show this page" });
  await expect(views.getByRole("radio", { name: "Side by side" })).toHaveCount(0);
  await views.getByRole("radio", { name: "Preview" }).click();
  await expect(page.getByTestId("editor-slot-preview")).toBeVisible();
  await expect(page.getByTestId("editor-slot-form")).toHaveCount(0);

  await expect(page.locator("[data-panel='site']")).toHaveAttribute("data-folded", "true");
});
