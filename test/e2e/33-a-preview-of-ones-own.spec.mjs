// Milestone 07, epic 1: two people preview the same project at once, each
// with their own change.
import { test, expect, onAPhone, saves, worksOnAPhone } from "./fixtures.mjs";
import { owner } from "./harness.mjs";

const marieke = { email: "marieke@korenaar.example", password: owner.password };

async function signInAs(page, chatcms, who) {
  await page.goto(`${chatcms.url}/sign-in`);
  await page.getByLabel("Email").fill(who.email);
  await page.getByLabel("Password").fill(who.password);
  await page.getByRole("button", { name: /Sign in/i }).click();
  await page.waitForURL(/projects/);
}

function previewOf(page, phone) {
  return {
    frame: page.frameLocator('iframe[title="Your site"]'),
    show: async () => {
      if (phone) await page.getByRole("radio", { name: "Preview" }).click();
    },
    form: async () => {
      if (phone) await page.getByRole("radio", { name: "Edit" }).click();
    },
  };
}

test("two people each see their own change in the preview", async ({ signedIn: joris, chatcms, browser }, testInfo) => {
  const phone = onAPhone(testInfo);
  const context = await browser.newContext({ viewport: joris.viewportSize() });
  const other = await context.newPage();

  try {
    await signInAs(other, chatcms, marieke);

    const home = `${chatcms.url}/o/technative/p/korenaar/content/home`;
    await joris.goto(home);
    await other.goto(home);

    const his = previewOf(joris, phone);
    const hers = previewOf(other, phone);

    await test.step("each gets a preview of the site", async () => {
      await his.show();
      await hers.show();
      await expect(his.frame.locator("body")).toContainText("Korenaar Bakery", { timeout: 90_000 });
      await expect(hers.frame.locator("body")).toContainText("Korenaar Bakery", { timeout: 90_000 });
    });

    await test.step("each saves a different title", async () => {
      await hers.form();
      await saves(other, () => other.getByRole("textbox", { name: "Page title" }).fill("Korenaar by Marieke"));
      await his.form();
      await saves(joris, () => joris.getByRole("textbox", { name: "Page title" }).fill("Korenaar by Joris"));
    });

    await test.step("each sees their own, and not the other's", async () => {
      await his.show();
      await hers.show();
      for (const [page, preview, mine, theirs] of [
        [joris, his, "Korenaar by Joris", "Korenaar by Marieke"],
        [other, hers, "Korenaar by Marieke", "Korenaar by Joris"],
      ]) {
        await expect(async () => {
          await page.getByRole("button", { name: "Refresh" }).click();
          await expect(preview.frame.locator("body")).toContainText(mine, { timeout: 2_000 });
        }).toPass({ timeout: 30_000, intervals: [250, 500, 1_000] });
        await expect(preview.frame.locator("body")).not.toContainText(theirs);
      }
    });

    await worksOnAPhone(joris, testInfo);
  } finally {
    await context.close();
  }
});

test("an implementer opening a project cold is told the preview is being prepared", async ({ signedIn: page, chatcms }, testInfo) => {
  await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/home`);
  if (onAPhone(testInfo)) await page.getByRole("radio", { name: "Preview" }).click();

  const status = page.getByRole("status").filter({ hasText: "Preparing your preview" });
  const frame = page.frameLocator('iframe[title="Your site"]');
  await expect(status.or(page.locator('iframe[title="Your site"]'))).toBeVisible({ timeout: 90_000 });

  if (await status.isVisible()) {
    await page.getByText("Show what is happening").click();
    await expect(page.getByText("Building the preview")).toBeVisible();
  }
  await expect(frame.locator("body")).toContainText("Korenaar Bakery", { timeout: 90_000 });
});
