// What can be clicked shows the hand, whether it is built as a link or as a
// button.
import { test, expect } from "./fixtures.mjs";

const cursor = (locator) => locator.evaluate((element) => getComputedStyle(element).cursor);

test("a button, a link and a menu item all show the hand", async ({ administering: page, chatcms }) => {
  await page.goto(`${chatcms.url}/projects`);

  const newProject = page.getByRole("button", { name: /New project/ });
  await expect(newProject).toBeVisible();
  expect(await cursor(newProject)).toBe("pointer");

  expect(await cursor(page.getByRole("link", { name: "Help" }))).toBe("pointer");

  const person = page.getByTestId("person-menu");
  expect(await cursor(person)).toBe("pointer");

  await person.click();
  const account = page.getByRole("menuitem", { name: "Your account" });
  await expect(account).toBeVisible();
  expect(await cursor(account)).toBe("pointer");
});

test("a disabled button does not", async ({ administering: page, chatcms }) => {
  await page.goto(`${chatcms.url}/projects`);
  await page.getByRole("button", { name: /New project/ }).click();

  const dialog = page.getByRole("dialog", { name: "New project" });
  await expect(dialog).toBeVisible();

  const disabled = dialog.locator("button:disabled").first();
  await expect(disabled).toBeVisible();
  expect(await cursor(disabled)).not.toBe("pointer");
});
