// Bean chatcms-9m9u: the number of changes waiting to be published follows a
// save without the page being opened again.
import { test, expect, saves } from "./fixtures.mjs";

test("a save shows in the publish count without reloading", async ({ signedIn: page, chatcms }) => {
  await page.goto(`${chatcms.url}/o/technative/p/korenaar/publish`);
  await page.getByRole("button", { name: /^Publish \d+ changes?$/ }).click();
  await expect(page.getByText("Your site is up to date")).toBeVisible({ timeout: 60_000 });

  await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/home`);
  const named = { name: /^Publish\s*1(\D|$)/ };
  const publish = page.getByRole("link", named).or(page.getByRole("button", named));
  await expect(publish).toHaveCount(0);

  await saves(page, () => page.getByRole("textbox", { name: "Page title" }).fill("Korenaar Bakery, fresh"));
  await expect(publish.first()).toBeVisible({ timeout: 10_000 });
});
