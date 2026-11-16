import { test, expect, worksOnAPhone } from "./fixtures.mjs";

test("a project starts from a template in the gallery", async ({ signedIn: page, chatcms }, testInfo) => {
  const remote = chatcms.emptyRemote("recipes");

  await test.step("the gallery lists the template and searches it", async () => {
    await page.goto(`${chatcms.url}/o/technative/projects`);
    await page.getByRole("link", { name: "Import a site or start from a template" }).click();

    await expect(page.getByRole("heading", { name: "Start from a template" })).toBeVisible();
    await expect(page.getByRole("article", { name: "Bakery" })).toBeVisible({ timeout: 60_000 });
    await expect(page.getByText("1 template", { exact: true })).toBeVisible();

    await page.getByPlaceholder("Search name, description or author").fill("spaceship");
    await expect(page.getByText('No templates match "spaceship".')).toBeVisible();
    await page.getByPlaceholder("Search name, description or author").fill("bakery");
    await expect(page.getByRole("article", { name: "Bakery" })).toBeVisible();

    await worksOnAPhone(page, testInfo);
  });

  await test.step("choosing it asks for a name and an empty place in git", async () => {
    await page.getByRole("link", { name: "Use template: Bakery" }).click();
    await expect(page.getByRole("heading", { name: "Start from Bakery" })).toBeVisible();
    await expect(page.getByRole("note")).toContainText("0.101.0 extended");

    await page.getByLabel("Project name").fill("Recipes");
    await page.getByLabel("Empty git address for your copy").fill(remote);

    await worksOnAPhone(page, testInfo);
    await page.getByRole("button", { name: "Continue to the key" }).click();
    await page.waitForURL(/\/o\/technative\/p\/recipes\/setup/);
  });

  await test.step("after the key, the template is copied in as one version", async () => {
    await expect(page.getByText(/Starting from Bakery/)).toBeVisible();
    await page.getByRole("button", { name: "Make a key" }).click();
    await expect(page.getByText(/ssh-ed25519/)).toBeVisible();

    await page.getByRole("button", { name: "Copy the template in" }).click();
    await expect(page.getByText("Ready to edit")).toBeVisible({ timeout: 60_000 });

    expect(chatcms.onTheRemote("recipes", "log", "--format=%s", "main").trim()).toBe("Start from Bakery");
    await worksOnAPhone(page, testInfo);
  });
});
