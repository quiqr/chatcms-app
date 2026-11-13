// The project at the top of a project's frame switches to another project,
// on a desktop and on a phone, and its last item leads to all of them.
import { test, expect, worksOnAPhone } from "./fixtures.mjs";

test("from one site to another, and back to all of them", async ({ signedIn: page, chatcms }, testInfo) => {
  await page.goto(`${chatcms.url}/o/technative/p/korenaar/content`);

  await test.step("another project is one choice away", async () => {
    await page.getByRole("button", { name: "Korenaar Bakery, switch project" }).click();

    const menu = page.getByRole("menu");
    await expect(menu.getByRole("menuitem", { name: /Korenaar Bakery/ })).toHaveAttribute("aria-current", "true");
    await menu.getByRole("menuitem", { name: /Studio Noord/ }).click();

    await expect(page).toHaveURL(/\/o\/technative\/p\/studio-noord\/content$/);
  });

  await test.step("and every project is at the end of the same menu", async () => {
    await page.getByRole("button", { name: "Studio Noord, switch project" }).click();
    await page.getByRole("menuitem", { name: "All projects" }).click();

    await expect(page).toHaveURL(/\/o\/technative\/projects$/);
  });

  await worksOnAPhone(page, testInfo);
});
