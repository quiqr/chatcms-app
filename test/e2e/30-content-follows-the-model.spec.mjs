import { test, expect, worksOnAPhone } from "./fixtures.mjs";

test("a template made of page bundles is listed and changed the way Quiqr does it", async ({ signedIn: page, chatcms }, testInfo) => {
  await test.step("the template starts kept in ChatCMS", async () => {
    await page.goto(`${chatcms.url}/o/technative/new/template/Bundles`);
    await expect(page.getByRole("heading", { name: "Start from Bundles" })).toBeVisible({ timeout: 60_000 });
    await page.getByLabel("Project name").fill("Journal");
    await page.getByRole("button", { name: "Continue", exact: true }).click();
    await page.waitForURL(/\/o\/technative\/p\/journal\/setup/, { timeout: 60_000 });
    await expect(page.getByText("Ready to edit")).toBeVisible({ timeout: 60_000 });
  });

  await test.step("the content follows the model's menu", async () => {
    await page.goto(`${chatcms.url}/o/technative/p/journal/content`);
    await expect(page.getByRole("region", { name: "Journal" })).toBeVisible();
    await expect(page.getByRole("region", { name: "Developer" })).toBeVisible();
    await page.goto(`${chatcms.url}/o/technative/p/journal/content/c/posts`);
    await expect(page.getByRole("link", { name: "Spring sale", exact: true })).toBeVisible();
    await expect(page.getByRole("link", { name: "Autumn walk", exact: true })).toBeVisible();
    await expect(page.getByText("All posts")).toHaveCount(0);
    await worksOnAPhone(page, testInfo);
  });

  await test.step("a page bundle opens in the editor", async () => {
    await page.getByRole("link", { name: "Spring sale", exact: true }).click();
    await page.waitForURL(/\/content\/posts\/2024\/spring-sale$/);
    await expect(page.getByRole("textbox", { name: "Title" })).toHaveValue("Spring sale");
    await worksOnAPhone(page, testInfo);
  });

  await test.step("a new post is made as a page bundle and can be removed", async () => {
    await page.goto(`${chatcms.url}/o/technative/p/journal/content/c/posts`);
    await page.getByRole("button", { name: "New post" }).click();
    await page.getByLabel("Title").fill("Brand new");
    await expect(page.getByRole("button", { name: "Create and edit" })).toBeEnabled();
    await page.getByRole("button", { name: "Create and edit" }).click();
    await page.waitForURL(/\/content\/posts\/brand-new$/);

    await page.goto(`${chatcms.url}/o/technative/p/journal/content/c/posts`);
    await page.getByRole("button", { name: "What to do with Brand new" }).click();
    await page.getByRole("button", { name: "Remove…" }).click();
    await page.getByRole("button", { name: "Yes, remove it" }).click();
    await expect(page.getByRole("link", { name: "Brand new", exact: true })).toHaveCount(0);
    await worksOnAPhone(page, testInfo);
  });
});
