import { test, expect, saves, worksOnAPhone } from "./fixtures.mjs";
import { owner } from "./harness.mjs";

const marieke = { email: "marieke@korenaar.example", password: owner.password };

test("two people edit different pages, and one publish carries both", async ({ signedIn: joris, chatcms, browser }, testInfo) => {
  test.setTimeout(180_000);
  const project = chatcms.projectId();
  const context = await browser.newContext({ viewport: joris.viewportSize() });
  const other = await context.newPage();

  try {
    await test.step("Marieke changes a news article", async () => {
      await other.goto(`${chatcms.url}/sign-in`);
      await other.getByLabel("Email").fill(marieke.email);
      await other.getByLabel("Password").fill(marieke.password);
      await other.getByRole("button", { name: /Sign in/i }).click();
      await other.waitForURL(/projects/);
      await other.goto(`${chatcms.url}/o/technative/p/korenaar/content/news/spring-prices`);
      await saves(other, () => other.getByRole("textbox", { name: "Headline" }).fill("Spring prices, by Marieke"));
    });

    await test.step("Joris changes the home page", async () => {
      await joris.goto(`${chatcms.url}/o/technative/p/korenaar/content/home`);
      await saves(joris, () => joris.getByRole("textbox", { name: "Page title" }).fill("Korenaar, by Joris"));
    });

    await test.step("Joris's publish shows both changes", async () => {
      await expect(async () => {
        await joris.goto(`${chatcms.url}/o/technative/p/korenaar/publish`);
        await expect(joris.getByRole("listitem").filter({ hasText: "Spring prices" }).filter({ hasText: marieke.email })).toBeVisible({ timeout: 2_000 });
      }).toPass({ timeout: 60_000, intervals: [1_000, 2_000] });
      await expect(joris.getByRole("listitem").filter({ hasText: "Home page" })).toBeVisible();
    });

    await test.step("one publish carries both", async () => {
      await joris.getByRole("button", { name: /^Publish \d+ changes?$/ }).click();
      await expect(joris.getByText("Your site is up to date")).toBeVisible({ timeout: 60_000 });
      expect(chatcms.onTheRemote(project, "cat-file", "blob", "refs/heads/main:content/_index.md")).toContain("Korenaar, by Joris");
      expect(chatcms.onTheRemote(project, "cat-file", "blob", "refs/heads/main:content/news/spring-prices.md")).toContain("Spring prices, by Marieke");
    });
  } finally {
    await context.close();
  }

  await worksOnAPhone(joris, testInfo);
});
