import { test, expect, saves, worksOnAPhone } from "./fixtures.mjs";

test("a project starts from a template in the gallery", async ({ signedIn: page, chatcms }, testInfo) => {
  const remote = chatcms.emptyRemote("recipes");

  await test.step("the gallery lists the template and searches it", async () => {
    await page.goto(`${chatcms.url}/o/technative/projects`);
    await page.getByRole("link", { name: "Import a site or start from a template" }).click();

    await expect(page.getByRole("heading", { name: "Start from a template" })).toBeVisible();
    await expect(page.getByRole("article", { name: "Bakery" })).toBeVisible({ timeout: 60_000 });
    await expect(page.getByText("2 templates", { exact: true })).toBeVisible();

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

test("a template without a git address is kept in ChatCMS and connected later", async ({ signedIn: page, chatcms }, testInfo) => {
  await test.step("the git address may be left empty", async () => {
    await page.goto(`${chatcms.url}/o/technative/new/template/Bakery`);
    await expect(page.getByRole("heading", { name: "Start from Bakery" })).toBeVisible({ timeout: 60_000 });
    await page.getByLabel("Project name").fill("Kept");
    await expect(page.getByText(/Leave it empty and ChatCMS keeps the site/)).toBeVisible();

    await worksOnAPhone(page, testInfo);
    await page.getByRole("button", { name: "Continue", exact: true }).click();
    await page.waitForURL(/\/o\/technative\/p\/kept\/setup/, { timeout: 60_000 });
  });

  await test.step("the template is copied in at once, without a key", async () => {
    await expect(page.getByText("Ready to edit")).toBeVisible({ timeout: 60_000 });
    await expect(page.getByText(/No key is needed while ChatCMS keeps the site/)).toBeVisible();
    await expect(page.getByRole("button", { name: "Copy the template in" })).toHaveCount(0);
    await worksOnAPhone(page, testInfo);
  });

  await test.step("it can be edited, but not published yet", async () => {
    await page.goto(`${chatcms.url}/o/technative/p/kept/content/home`);
    await saves(page, () => page.getByRole("textbox", { name: "Page title" }).fill("Kept Bakery"));

    await page.goto(`${chatcms.url}/o/technative/p/kept/publish`);
    await page.getByRole("button", { name: /^Publish \d+ changes?$/ }).click();
    await expect(page.getByRole("alert")).toContainText("not connected yet");
    await worksOnAPhone(page, testInfo);
  });

  await test.step("connecting it in settings sends every version there", async () => {
    const remote = chatcms.emptyRemote("kept");
    await page.goto(`${chatcms.url}/o/technative/p/kept/settings`);
    await expect(page.getByText("Kept in ChatCMS, not connected yet")).toBeVisible();
    await worksOnAPhone(page, testInfo);

    await page.getByLabel("Empty git address to connect").fill(remote);
    await page.getByRole("button", { name: "Connect", exact: true }).click();
    await expect(page.getByText("Connected. Every version so far was sent there.")).toBeVisible({ timeout: 60_000 });
    await expect(page.getByText("Kept in ChatCMS, not connected yet")).toHaveCount(0);

    const subjects = chatcms.onTheRemote("kept", "log", "--format=%s", "main").trim().split("\n");
    expect(subjects.length).toBe(2);
    expect(subjects.at(-1)).toBe("Start from Bakery");
  });
});
