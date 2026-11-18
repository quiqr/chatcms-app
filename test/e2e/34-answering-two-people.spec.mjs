// Milestone 07: when two people changed the same field, the question can be
// answered at any later moment, not only in the request that asked it.
import { test, expect, saves, worksOnAPhone } from "./fixtures.mjs";

const fromElsewhere = `---
title: Korenaar from elsewhere
intro: We bake with apples and pears every morning.
openToday: true
style: traditional
accent: '#8b5a2b'
specialities:
  - Sourdough
  - Apple tart
opening:
  - day: Monday
    hours: 7 to 18
  - day: Saturday
    hours: 7 to 16
seo:
  description: A bakery on the Oude Gracht.
---

Come in and smell it.
`;

test("a question about the same field is answered later, and the answer is kept", async ({ signedIn: page, chatcms }, testInfo) => {
  const sync = `${chatcms.url}/o/technative/p/korenaar/sync`;

  await test.step("the title is changed here and somewhere else", async () => {
    await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/home`);
    await saves(page, () => page.getByRole("textbox", { name: "Page title" }).fill("Korenaar by Joris"));
    chatcms.pushFromSomewhereElse(chatcms.projectId(), "content/_index.md", fromElsewhere, "Rename the bakery");
  });

  await test.step("checking for updates asks which title to keep", async () => {
    await page.goto(sync);
    await expect(async () => {
      await page.getByRole("button", { name: "Check for updates" }).click();
      await expect(page.getByRole("button", { name: "Keep mine" }).first()).toBeVisible({ timeout: 3_000 });
    }).toPass({ timeout: 60_000, intervals: [500, 1_000, 2_000] });
  });

  await test.step("coming back later, the same question is still asked", async () => {
    await page.goto(`${chatcms.url}/o/technative/p/korenaar/content`);
    await page.goto(sync);
    await expect(page.getByRole("button", { name: "Keep mine" }).first()).toBeVisible({ timeout: 20_000 });
    await page.getByRole("button", { name: "Check for updates" }).click();
    await expect(page.getByRole("button", { name: "Keep mine" }).first()).toBeVisible();
    await expect(page.getByRole("alert")).toHaveCount(0);
  });

  await test.step("keeping mine finishes it", async () => {
    await page.getByRole("button", { name: "Keep mine" }).first().click();
    await page.getByRole("button", { name: "Save these answers" }).click();
    await expect(page.getByRole("button", { name: "Keep mine" })).toHaveCount(0, { timeout: 20_000 });

    await page.goto(sync);
    await expect(page.getByRole("button", { name: "Save these answers" })).toHaveCount(0, { timeout: 20_000 });
  });

  await test.step("the page holds the kept title", async () => {
    await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/home`);
    await expect(page.getByRole("textbox", { name: "Page title" })).toHaveValue("Korenaar by Joris", { timeout: 20_000 });
  });

  await worksOnAPhone(page, testInfo);
});
