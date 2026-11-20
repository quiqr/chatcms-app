// Bean chatcms-qru3: a project's log keeps every preview start with what Hugo
// it used, beside builds and publications, for implementers only.
import { test, expect, worksOnAPhone } from "./fixtures.mjs";
import { owner } from "./harness.mjs";

const marieke = { email: "marieke@korenaar.example", password: owner.password };

test("an implementer opens this start in the log and reads what Hugo it used", async ({ signedIn: page, chatcms }, testInfo) => {
  await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/home`);
  const symbol = page.getByRole("button", { name: "Your preview: Preview ready" });
  await expect(symbol).toBeVisible({ timeout: 60_000 });
  await symbol.click();
  await page.getByRole("dialog").getByRole("link", { name: "Open the log" }).click();
  await page.waitForURL(/\/log\?entry=start/);

  const detail = page.getByTestId("log-detail");
  const hugo = detail.getByRole("region", { name: "Hugo" });
  await expect(hugo).toContainText("Program");
  await expect(hugo).toContainText(/hugo/);
  await expect(hugo).toContainText("Why");
  await expect(detail.getByRole("region", { name: "Log" })).toContainText("backend");

  if (testInfo.project.name === "phone") {
    await page.getByRole("button", { name: "Back to the log" }).click();
  }
  await expect(page.getByRole("list", { name: "Log" })).toContainText("Preview start");
  await page.getByRole("radio", { name: "Publications" }).click();
  await expect(page.getByRole("list", { name: "Log" }).or(page.getByText("Nothing has been logged yet.")).first()).toBeVisible();

  await worksOnAPhone(page, testInfo);
});

test("an editor has no log and is refused it", async ({ browser, chatcms }, testInfo) => {
  const context = await browser.newContext({ viewport: testInfo.project.use.viewport });
  const page = await context.newPage();
  try {
    await page.goto(`${chatcms.url}/sign-in`);
    await page.getByLabel("Email").fill(marieke.email);
    await page.getByLabel("Password").fill(marieke.password);
    await page.getByRole("button", { name: /Sign in/i }).click();
    await page.waitForURL(/projects/);

    await page.goto(`${chatcms.url}/o/technative/p/korenaar/content`);
    await expect(page.getByRole("link", { name: "Content" }).first()).toBeVisible();
    await expect(page.getByRole("link", { name: "Log", exact: true })).toHaveCount(0);

    const refused = await page.request.get(`${chatcms.url}/api/organisations/technative/projects/korenaar/log`, { failOnStatusCode: false });
    expect(refused.status(), "an editor read the project log").toBe(403);
  } finally {
    await context.close();
  }
});
