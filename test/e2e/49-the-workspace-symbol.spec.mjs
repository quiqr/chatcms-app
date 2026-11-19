// Bean chatcms-e6lo: a symbol beside the save status says how the person's
// own workspace stands; an implementer sees more and can start it again.
import { test, expect, worksOnAPhone } from "./fixtures.mjs";
import { owner } from "./harness.mjs";

const marieke = { email: "marieke@korenaar.example", password: owner.password };

test("an implementer sees what the workspace runs on and starts it again", async ({ signedIn: page, chatcms }, testInfo) => {
  await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/home`);

  const symbol = page.getByRole("button", { name: "Your preview: Preview ready" });
  await expect(symbol).toBeVisible({ timeout: 60_000 });
  await symbol.click();

  const popover = page.getByRole("dialog");
  await expect(popover).toContainText(/Started/);
  await expect(popover).toContainText("Runs on this server");
  await expect(popover).toContainText(/The assistant (can be asked|is not available)\./);

  await popover.getByRole("button", { name: "Start again" }).click();
  await expect(page.getByRole("button", { name: /Your preview: Preview (being prepared|ready)/ })).toBeVisible();
  await expect(page.getByRole("button", { name: "Your preview: Preview ready" })).toBeVisible({ timeout: 60_000 });

  await page.keyboard.press("Escape");
  await worksOnAPhone(page, testInfo);
});

test("an editor is told in plain words, and cannot start it again", async ({ browser, chatcms }, testInfo) => {
  const context = await browser.newContext({ viewport: testInfo.project.use.viewport });
  const page = await context.newPage();
  try {
    await page.goto(`${chatcms.url}/sign-in`);
    await page.getByLabel("Email").fill(marieke.email);
    await page.getByLabel("Password").fill(marieke.password);
    await page.getByRole("button", { name: /Sign in/i }).click();
    await page.waitForURL(/projects/);

    await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/home`);
    const symbol = page.getByRole("button", { name: "Your preview: Preview ready" });
    await expect(symbol).toBeVisible({ timeout: 60_000 });
    await symbol.click();

    const popover = page.getByRole("dialog");
    await expect(popover).toContainText(/Started/);
    await expect(popover).not.toContainText("Runs on");
    await expect(popover.getByRole("button", { name: "Start again" })).toHaveCount(0);

    const refused = await page.request.post(`${chatcms.url}/api/organisations/technative/projects/korenaar/workspace/restart`, {
      failOnStatusCode: false,
    });
    expect(refused.status(), "an editor started a workspace again").toBe(403);
  } finally {
    await context.close();
  }
});
