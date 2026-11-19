import { test, expect, worksOnAPhone } from "./fixtures.mjs";
import { realAI } from "./harness.mjs";

test.use({ assistant: true });


test("two titles written at once are combined by the AI, and the combination is kept", async ({ signedIn: page, chatcms }, testInfo) => {
  test.setTimeout(300_000);

  await test.step("the title waiting to be published is changed somewhere else too", async () => {
    await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/home`);
    await expect(page.getByRole("textbox", { name: "Page title" })).toHaveValue("Korenaar Bakery");
    const project = chatcms.projectId();
    const live = chatcms.onTheRemote(project, "cat-file", "blob", "refs/heads/main:content/_index.md");
    const renamed = live.replace(/^title: .*$/m, "title: Korenaar on the Oude Gracht");
    expect(renamed).not.toBe(live);
    chatcms.pushFromSomewhereElse(project, "content/_index.md", renamed, "Rename the bakery");
  });

  await test.step("checking for updates asks which title to keep", async () => {
    await page.goto(`${chatcms.url}/o/technative/p/korenaar/sync`);
    await expect(async () => {
      await page.getByRole("button", { name: "Check for updates" }).click();
      await expect(page.getByRole("button", { name: "Let the AI combine them" }).first()).toBeVisible({ timeout: 3_000 });
    }).toPass({ timeout: 60_000, intervals: [500, 1_000, 2_000] });
  });

  let combined = "";
  await test.step("the AI offers a combination", async () => {
    const question = page.getByRole("listitem").filter({ hasText: /page title/i });
    await question.getByRole("button", { name: "Let the AI combine them" }).click();
    const box = question.getByRole("textbox", { name: "Combined Page title" });
    await expect(box).not.toHaveValue("", { timeout: 120_000 });
    combined = await box.inputValue();
    if (realAI) console.log(`real AI combined: ${JSON.stringify(combined)}`);
    if (!realAI) expect(combined).toBe("Korenaar Bakery & Korenaar on the Oude Gracht");
    expect(combined.length).toBeLessThan(200);
  });

  await test.step("saving the answers keeps the combination", async () => {
    await page.getByRole("button", { name: "Save these answers" }).click();
    await expect(page.getByRole("button", { name: "Let the AI combine them" })).toHaveCount(0, { timeout: 20_000 });
    expect(chatcms.inTheSharedVersion(chatcms.projectId(), "content/_index.md")).toContain(`title: ${combined}`);
    await expect(async () => {
      await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/home`);
      await expect(page.getByRole("textbox", { name: "Page title" })).toHaveValue(combined, { timeout: 2_000 });
    }).toPass({ timeout: 30_000, intervals: [1_000, 2_000] });
  });

  await worksOnAPhone(page, testInfo);
});
