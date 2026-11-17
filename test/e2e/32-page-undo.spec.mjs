import { test, expect, onAPhone, saves, worksOnAPhone } from "./fixtures.mjs";

test("an edit can be undone, redone and gone back on from the undo history", async ({ signedIn: page, chatcms }, testInfo) => {
  const name = onAPhone(testInfo) ? "Undo on a phone" : "Undo on a desk";
  const headline = page.getByRole("textbox", { name: "Headline" });

  await test.step("a page to work on", async () => {
    await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/c/news`);
    await page.getByRole("button", { name: "New article" }).click();
    await page.getByLabel("Title").fill(name);
    await page.getByRole("button", { name: "Create and edit" }).click();
    await page.waitForURL(/content\/news\/undo-on-a-/);
    await expect(headline).toHaveValue(name);
    await expect(page.getByRole("button", { name: "Undo", exact: true })).toBeDisabled();
    await worksOnAPhone(page, testInfo);
  });

  await test.step("a cleared headline comes back with Undo, and stays after a reload", async () => {
    await saves(page, () => headline.fill(""));
    await saves(page, () => page.getByRole("button", { name: "Undo Headline" }).click());
    await expect(headline).toHaveValue(name);
    await page.reload();
    await expect(headline).toHaveValue(name);
    await expect(page.getByRole("button", { name: "Redo Headline" })).toBeEnabled();
  });

  await test.step("Redo clears it again", async () => {
    await saves(page, () => page.getByRole("button", { name: "Redo Headline" }).click());
    await expect(headline).toHaveValue("");
  });

  await test.step("the undo history goes back to the page as it was opened", async () => {
    await saves(page, () => headline.fill("Something else"));
    await page.getByRole("button", { name: "Undo history" }).click();
    await expect(page.getByRole("menuitem")).toHaveCount(3);
    await saves(page, () => page.getByRole("menuitem", { name: /Opened/ }).click());
    await expect(headline).toHaveValue(name);
    await worksOnAPhone(page, testInfo);
  });
});
