import { test, expect, onAPhone } from "./fixtures.mjs";

const marieke = { email: "marieke@korenaar.example", password: "a long enough passphrase" };

test("the preview works when the app is reached by another name than its own", async ({ page, chatcms }, testInfo) => {
  test.setTimeout(150_000);
  const other = chatcms.url.replace("127.0.0.1", "localhost");
  await page.goto(`${other}/sign-in`);
  await page.getByLabel("Email").fill(marieke.email);
  await page.getByLabel("Password").fill(marieke.password);
  await page.getByRole("button", { name: /Sign in/i }).click();
  await page.waitForURL(/projects/);

  await page.goto(`${other}/o/technative/p/korenaar/content/news/spring-prices`);
  if (onAPhone(testInfo)) await page.getByRole("radio", { name: "Preview" }).click();
  const frame = page.locator('iframe[title="Your site"]');
  await expect(frame).toHaveAttribute("src", /^\/preview\/technative\/korenaar\//, { timeout: 90_000 });
  await expect(page.frameLocator('iframe[title="Your site"]').locator("body")).toContainText("Spring prices", { timeout: 90_000 });
});
