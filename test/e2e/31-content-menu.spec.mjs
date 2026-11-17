import { test, expect, onAPhone, worksOnAPhone } from "./fixtures.mjs";

test("the content menu, a collection and the editor work together", async ({ signedIn: page, chatcms }, testInfo) => {
  const phone = onAPhone(testInfo);
  const content = `${chatcms.url}/o/technative/p/diary/content`;

  await test.step("a kept site starts from the bundle template", async () => {
    await page.goto(`${chatcms.url}/o/technative/new/template/Bundles`);
    await expect(page.getByRole("heading", { name: "Start from Bundles" })).toBeVisible({ timeout: 60_000 });
    await page.getByLabel("Project name").fill("Diary");
    await page.getByRole("button", { name: "Continue", exact: true }).click();
    await page.waitForURL(/\/o\/technative\/p\/diary\/setup/, { timeout: 60_000 });
    await expect(page.getByText("Ready to edit")).toBeVisible({ timeout: 60_000 });
  });

  await test.step("1. the content menu lists the model's groups in order", async () => {
    await page.goto(content);
    const column = page.getByRole("navigation", { name: "Content" });
    await expect(column.getByRole("region")).toHaveCount(2);
    await expect(column.getByRole("region").first()).toHaveAttribute("aria-label", "Journal");
    await expect(column.getByRole("link", { name: /Posts/ })).toContainText("2");
    await worksOnAPhone(page, testInfo);
  });

  await test.step("2. a collection lists titles, counts, folders, search and sort", async () => {
    await page.goto(`${content}/c/posts`);
    await expect(page.getByRole("heading", { name: "Posts", level: 2 })).toBeVisible();
    await expect(page.getByText(/^2 posts · /)).toBeVisible();
    await expect(page.getByRole("radio", { name: "2024 · 1" })).toBeVisible();

    await page.getByLabel("Search in Posts").fill("autumn");
    await expect(page.getByRole("link", { name: "Spring sale", exact: true })).toHaveCount(0);
    await expect(page.getByRole("link", { name: "Autumn walk", exact: true })).toBeVisible();
    await page.getByLabel("Search in Posts").fill("");
    await expect(page.getByRole("link", { name: "Spring sale", exact: true })).toBeVisible();

    await page.getByRole("combobox", { name: "Sort" }).selectOption("title");
    await expect(page.getByRole("listitem").first()).toContainText("Autumn walk");
    await worksOnAPhone(page, testInfo);
  });

  await test.step("7. an implementer can see file names", async () => {
    await page.getByRole("switch", { name: "File names" }).check();
    await expect(page.getByText("content/post/autumn/index.md")).toBeVisible();
  });

  await test.step("3. an open item knows its place and steps through the list", async () => {
    await page.getByRole("link", { name: "Autumn walk", exact: true }).click();
    await page.waitForURL(/\/content\/posts\/autumn\?sort=title$/);
    await expect(page.getByText("Posts · 1 of 2").first()).toBeVisible();
    if (!phone) {
      await expect(page.getByRole("navigation", { name: "Content" }).getByRole("link", { name: /Posts/ })).toHaveAttribute(
        "aria-current",
        "page",
      );
    }
    await page.getByRole("link", { name: "Next" }).first().click();
    await page.waitForURL(/\/content\/posts\/2024\/spring-sale\?sort=title$/);
    await expect(page.getByText("Posts · 2 of 2").first()).toBeVisible();
    await expect(page.getByRole("button", { name: "Next" }).first()).toBeDisabled();
    await worksOnAPhone(page, testInfo);
  });

  if (!phone) {
    await test.step("4. the content column folds and stays folded after a reload", async () => {
      const panel = page.locator("[data-panel='content']");
      await page.getByRole("button", { name: "Hide the content menu" }).click();
      await expect(panel).toHaveAttribute("data-folded", "true");
      await page.reload();
      await expect(panel).toHaveAttribute("data-folded", "true");
      await expect(page.getByRole("navigation", { name: "Content" })).toHaveCount(0);
      await page.getByRole("button", { name: "Show the content menu" }).click();
      await expect(page.getByRole("navigation", { name: "Content" })).toBeVisible();
    });

    await test.step("4b. the navigation folds by hand and stays folded when an item opens", async () => {
      const site = page.locator("[data-panel='site']");
      await page.goto(`${content}/c/posts`);
      await expect(site).toHaveAttribute("data-folded", "false");
      await page.getByRole("button", { name: "Hide the navigation" }).click();
      await expect(site).toHaveAttribute("data-folded", "true");
      await page.getByRole("link", { name: "Autumn walk", exact: true }).click();
      await page.waitForURL(/\/content\/posts\/autumn$/);
      await expect(site).toHaveAttribute("data-folded", "true");
      await page.reload();
      await expect(site).toHaveAttribute("data-folded", "true");
      await page.getByRole("button", { name: "Show the navigation" }).click();
      await expect(site).toHaveAttribute("data-folded", "false");
      await page.goto(`${content}/c/posts`);
      await expect(site).toHaveAttribute("data-folded", "false");
    });
  }

  await test.step("5. quick find opens an item by its title", async () => {
    await page.goto(`${content}/c/posts`);
    await expect(page.getByRole("heading", { name: "Posts", level: 2 })).toBeVisible();
    await page.keyboard.press("Control+k");
    await page.getByRole("dialog", { name: "Find a page" }).getByLabel("Title").fill("spring");
    await expect(page.getByRole("dialog").getByRole("button", { name: "Spring sale" })).toBeVisible();
    await page.keyboard.press("Enter");
    await page.waitForURL(/\/content\/posts\/2024\/spring-sale$/);
  });

  await test.step("6. a taken address is refused, a free one opens the editor", async () => {
    await page.goto(`${content}/c/posts`);
    await page.getByRole("button", { name: "New post" }).click();
    const dialog = page.getByRole("dialog");
    await dialog.getByLabel("Title").fill("Autumn");
    await expect(dialog.getByText("There's already a post at this address. Change the title or the address.")).toBeVisible();
    await expect(dialog.getByRole("button", { name: "Create and edit" })).toBeDisabled();

    await dialog.getByLabel("Title").fill("Winter");
    await expect(dialog.getByRole("button", { name: "Create and edit" })).toBeEnabled();
    await dialog.getByRole("button", { name: "Create and edit" }).click();
    await page.waitForURL(/\/content\/posts\/winter$/);
    await worksOnAPhone(page, testInfo);
  });

  if (phone) {
    await test.step("8. on a phone: menu, then collection, then editor", async () => {
      await page.goto(content);
      await page.getByRole("link", { name: /Posts/ }).click();
      await page.waitForURL(/\/content\/c\/posts$/);
      await expect(page.getByRole("link", { name: "Back to the content menu" })).toBeVisible();
      await worksOnAPhone(page, testInfo);
      await page.getByRole("link", { name: "Autumn walk", exact: true }).click();
      await page.waitForURL(/\/content\/posts\/autumn$/);
      await worksOnAPhone(page, testInfo);
    });
  }
});
