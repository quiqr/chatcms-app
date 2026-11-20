// Bean chatcms-nwic: an implementer looks through a project's files, edits a
// template in a code editor, and the change is published like any other.
import { test, expect, worksOnAPhone } from "./fixtures.mjs";
import { owner } from "./harness.mjs";

const marieke = { email: "marieke@korenaar.example", password: owner.password };

test("an implementer opens a template with the keyboard, edits it and publishes it", async ({ signedIn: page, chatcms }, testInfo) => {
  const project = chatcms.projectId();
  await page.goto(`${chatcms.url}/o/technative/p/korenaar/publish`);
  await page.getByRole("button", { name: /^Publish \d+ changes?$/ }).click();
  await expect(page.getByText("Your site is up to date")).toBeVisible({ timeout: 60_000 });

  await page.goto(`${chatcms.url}/o/technative/p/korenaar/files`);
  const tree = page.getByRole("tree", { name: "Files of the site" });
  await expect(tree).toBeVisible();
  await expect(tree.getByRole("treeitem", { name: ".git" })).toHaveCount(0);

  await tree.getByRole("treeitem", { name: "layouts" }).focus();
  await page.keyboard.press("ArrowRight");
  await page.keyboard.press("ArrowDown");
  await page.keyboard.press("ArrowDown");
  await expect(tree.getByRole("treeitem", { name: "index.html" })).toBeFocused();
  await page.keyboard.press("Enter");
  await page.waitForURL(/file=layouts%2Findex.html/);

  const editor = page.getByRole("textbox", { name: "Text of layouts/index.html" });
  await expect(editor).toContainText("{{ .Content }}");
  await expect(page.locator(".cm-go-template").first()).toBeVisible();
  await editor.click();
  await page.keyboard.press("ControlOrMeta+End");
  await page.keyboard.type("<!-- baked by night -->");
  await expect(page.getByRole("status").filter({ hasText: "Not saved yet" })).toBeVisible();
  await page.keyboard.press("ControlOrMeta+s");
  await expect(page.getByRole("status").filter({ hasText: /^Saved$/ })).toBeVisible({ timeout: 30_000 });

  const named = { name: /^(Publish\s*1|1\s*Publish)(\D|$)/ };
  await expect(page.getByRole("link", named).or(page.getByRole("button", named)).first()).toBeVisible({ timeout: 15_000 });

  await page.goto(`${chatcms.url}/o/technative/p/korenaar/publish`);
  await page.getByRole("button", { name: /^Publish \d+ changes?$/ }).click();
  await expect(page.getByText("Your site is up to date")).toBeVisible({ timeout: 90_000 });
  expect(chatcms.onTheRemote(project, "cat-file", "blob", "refs/heads/main:layouts/index.html")).toContain("<!-- baked by night -->");

  await page.goto(`${chatcms.url}/o/technative/p/korenaar/files`);
  await worksOnAPhone(page, testInfo);
});

test("an editor has no files and is refused them", async ({ browser, chatcms }, testInfo) => {
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
    await expect(page.getByRole("link", { name: "Files", exact: true })).toHaveCount(0);
    for (const path of ["files", "files/hugo.toml", "files/../../etc/passwd"]) {
      const refused = await page.request.get(`${chatcms.url}/api/organisations/technative/projects/korenaar/${path}`, { failOnStatusCode: false });
      expect(refused.status(), `an editor read ${path}`).not.toBe(200);
    }
  } finally {
    await context.close();
  }
});
