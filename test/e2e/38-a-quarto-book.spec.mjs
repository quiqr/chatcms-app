import { fileURLToPath } from "node:url";

import { test, expect, onAPhone, saves } from "./fixtures.mjs";

const book = process.env.CHATCMS_QUARTO_BOOK ?? fileURLToPath(new URL("../fixtures/quarto-book", import.meta.url));

test("a Quarto book is brought in as a site with its model", async ({ signedIn: page, chatcms }) => {
  test.skip(!/quarto=/.test(process.env.CHATCMS_PACKAGES ?? ""), "no Quarto provided to this installation");
  const remote = chatcms.remoteFrom(book, "a-book");
  await page.goto(`${chatcms.url}/o/technative/p/studio-noord/setup`);
  await page.getByRole("button", { name: "Make a key" }).click();
  await page.getByLabel("Address of the site in git").fill(remote);
  await page.getByRole("button", { name: "Bring it in" }).click();
  await expect(page.getByText("Quarto site, Quiqr model found: 1 pages and 1 collections.")).toBeVisible({ timeout: 60_000 });
});

test("a Quarto book previews, publishes its website and keeps its PDF", async ({ signedIn: page, chatcms }, testInfo) => {
  test.skip(!/quarto=/.test(process.env.CHATCMS_PACKAGES ?? ""), "no Quarto provided to this installation");
  test.setTimeout(300_000);
  const phone = onAPhone(testInfo);
  const project = chatcms.projectId();

  await test.step("the site becomes a Quarto book", async () => {
    chatcms.layOver(project, book, "Turn the site into a book");
    await page.goto(`${chatcms.url}/o/technative/p/korenaar/sync`);
    await page.getByRole("button", { name: "Check for updates" }).click();
    await expect(page.getByText(/Joris updated this site/)).toBeVisible({ timeout: 60_000 });
  });

  await test.step("it publishes the built site to a line of its own", async () => {
    await page.goto(`${chatcms.url}/o/technative/p/korenaar/settings`);
    await page.getByRole("radio", { name: /A line of its own/ }).click();
    await expect(async () => {
      await page.reload();
      await expect(page.getByRole("radio", { name: /A line of its own/ })).toBeChecked({ timeout: 2_000 });
    }).toPass({ timeout: 20_000 });
    await page.getByRole("radio", { name: /The built site/ }).click();
    await page.reload();
    await expect(page.getByRole("radio", { name: /A line of its own/ })).toBeChecked();
    await expect(page.getByRole("radio", { name: /The built site/ })).toBeChecked();
  });

  await test.step("a chapter shows in the preview", async () => {
    await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/preface`);
    await expect(page.getByRole("textbox", { name: "Chapter title" })).toHaveValue("Preface");
    if (phone) await page.getByRole("radio", { name: "Preview" }).click();
    const preview = page.frameLocator('iframe[title="Your site"]');
    await expect(preview.locator("body")).toContainText("Welcome to the bakery", { timeout: 150_000 });
    if (phone) await page.getByRole("radio", { name: "Edit" }).click();
  });

  await test.step("a change is saved", async () => {
    await saves(page, () => page.getByRole("textbox", { name: "Chapter title" }).fill("A word first"));
  });

  let pdf;
  await test.step("publishing puts the website live and keeps the PDF", async () => {
    await page.goto(`${chatcms.url}/o/technative/p/korenaar/publish`);
    await page.getByRole("button", { name: /^Publish \d+ changes?$/ }).click();
    await expect(page.getByText("Your site is up to date")).toBeVisible({ timeout: 180_000 });
    const made = page.getByRole("region", { name: "Files made by this publish" });
    pdf = made.getByRole("link", { name: "Download Korenaar-Recipes.pdf" });
    await expect(pdf).toBeVisible();

    const bread = chatcms.onTheRemote(project, "cat-file", "blob", "refs/heads/quiqr/publish:chapters/bread.html");
    expect(bread).toContain("Loaves this week: 39");
    const index = chatcms.onTheRemote(project, "cat-file", "blob", "refs/heads/quiqr/publish:index.html");
    expect(index).toContain("A word first");
    expect(chatcms.onTheRemote(project, "cat-file", "-t", "refs/heads/quiqr/publish:Korenaar-Recipes.pdf").trim()).toBe("blob");
  });

  await test.step("the PDF downloads", async () => {
    const answer = await page.request.get(new URL(await pdf.getAttribute("href"), chatcms.url).href);
    expect(answer.status()).toBe(200);
    expect(answer.headers()["content-disposition"]).toContain("Korenaar-Recipes.pdf");
    expect((await answer.body()).subarray(0, 5).toString()).toBe("%PDF-");
  });

  await test.step("the PDF is kept under Artifacts", async () => {
    await page.goto(`${chatcms.url}/o/technative/p/korenaar/artifacts`);
    await expect(page.getByRole("link", { name: "Download Korenaar-Recipes.pdf" })).toBeVisible();
    if (!phone) {
      await expect(page.getByRole("group", { name: "Site" }).getByRole("link", { name: "Artifacts" })).toBeVisible();
    }
  });
});

test("a book previewed as a PDF shows its text through PDF.js", async ({ signedIn: page, chatcms }, testInfo) => {
  test.skip(!/quarto=/.test(process.env.CHATCMS_PACKAGES ?? ""), "no Quarto provided to this installation");
  test.setTimeout(300_000);
  const project = chatcms.projectId();
  chatcms.layOver(project, book, "Turn the site into a book");
  chatcms.pushFromSomewhereElse(
    project,
    "quiqr/runtime.yaml",
    [
      "packages:",
      '  quarto: { version: "1.10.18", nixpkgs: 02c087f3fe673c9e317f206a627cb6b910336dbe }',
      "preview:",
      "  type: pdf",
      "  open: /Korenaar-Recipes.pdf",
      '  command: [quarto, preview, "{source}", --host, "{host}", --port, "{port}", --no-browser, --to, typst]',
      "",
    ].join("\n"),
    "Preview the book as a PDF",
  );
  await page.goto(`${chatcms.url}/o/technative/p/korenaar/sync`);
  await page.getByRole("button", { name: "Check for updates" }).click();
  await expect(page.getByText(/Joris updated this site/)).toBeVisible({ timeout: 60_000 });

  await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/preface`);
  if (onAPhone(testInfo)) await page.getByRole("radio", { name: "Preview" }).click();
  const viewer = page.frameLocator('iframe[title="Your site"]');
  await expect(viewer.locator(".textLayer").first()).toContainText("Korenaar", { timeout: 150_000 });
  const page1 = viewer.locator(".page").first();
  const room = await viewer.locator("html").evaluate((html) => html.clientWidth);
  const wide = await page1.evaluate((box) => box.getBoundingClientRect().width);
  if (onAPhone(testInfo)) {
    expect(wide).toBeLessThanOrEqual(room - 24 + 1);
  } else {
    expect(Number(await page1.evaluate((box) => box.style.getPropertyValue("--scale-factor")))).toBeCloseTo(96 / 72);
  }

  const tools = viewer.getByRole("toolbar", { name: "PDF tools" });
  const field = tools.getByRole("textbox", { name: "Page" });

  await test.step("the toolbar says the page and goes to another", async () => {
    await expect(field).toHaveValue("1");
    await expect(tools.getByRole("button", { name: "Previous page" })).toBeDisabled();
    const total = Number((await tools.getByText(/^\/ \d+$/).textContent()).replace(/\D/g, ""));
    expect(total).toBeGreaterThan(1);
    await field.fill(String(total));
    await field.press("Enter");
    await expect(field).toHaveValue(String(total));
    await expect(tools.getByRole("button", { name: "Next page" })).toBeDisabled();
    await tools.getByRole("button", { name: "Previous page" }).click();
    await expect(field).toHaveValue(String(total - 1));
  });

  await test.step("zoom changes the size, keeps the proportions, and 100% goes back", async () => {
    const before = await page1.evaluate((box) => box.getBoundingClientRect());
    const start = await tools.getByRole("button", { name: "Back to the actual size" }).textContent();
    await tools.getByRole("button", { name: "Zoom in" }).click();
    await expect(tools.getByRole("button", { name: "Back to the actual size" })).not.toHaveText(start);
    await expect(async () => {
      const after = await page1.evaluate((box) => box.getBoundingClientRect());
      expect(after.width).toBeGreaterThan(before.width);
      expect(after.height / after.width).toBeCloseTo(before.height / before.width, 2);
    }).toPass({ timeout: 10_000 });
    await expect(viewer.locator(".textLayer").first()).toContainText("Korenaar", { timeout: 30_000 });
    await tools.getByRole("button", { name: "Back to the actual size" }).click();
    await expect(tools.getByRole("button", { name: "Back to the actual size" })).toHaveText("100%");
    await expect(async () => {
      expect(Number(await page1.evaluate((box) => box.style.getPropertyValue("--scale-factor")))).toBeCloseTo(96 / 72);
    }).toPass({ timeout: 30_000 });
  });

  await test.step("the contents go to a chapter", async () => {
    await tools.getByRole("button", { name: "Contents" }).click();
    const contents = viewer.getByRole("navigation", { name: "Contents" });
    if (process.env.CHATCMS_SHOTS) {
      await page.screenshot({ path: `${process.env.CHATCMS_SHOTS}/pdf-viewer/${testInfo.project.name}.png` });
    }
    await contents.getByRole("button", { name: /Rye$/ }).click();
    const rye = viewer.locator(".page").filter({ has: viewer.locator(".textLayer", { hasText: "A dark rye" }) });
    await expect(rye).toBeInViewport({ timeout: 30_000 });
    if (onAPhone(testInfo)) await expect(contents).toBeHidden();
  });

  await test.step("the PDF can be downloaded", async () => {
    await expect(tools.getByRole("link", { name: "Download the PDF" })).toHaveAttribute("download", "Korenaar-Recipes.pdf");
  });
});

test.describe("on a preview origin of its own", () => {
  test.use({ previewApart: true });

  test("a Quarto preview reloads by itself when its sources change", async ({ signedIn: page, chatcms }, testInfo) => {
    test.skip(!/quarto=/.test(process.env.CHATCMS_PACKAGES ?? ""), "no Quarto provided to this installation");
    test.setTimeout(300_000);
    const project = chatcms.projectId();
    chatcms.layOver(project, book, "Turn the site into a book");
    await page.goto(`${chatcms.url}/o/technative/p/korenaar/sync`);
    await page.getByRole("button", { name: "Check for updates" }).click();
    await expect(page.getByText(/Joris updated this site/)).toBeVisible({ timeout: 60_000 });

    await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/preface`);
    if (onAPhone(testInfo)) await page.getByRole("radio", { name: "Preview" }).click();
    const preview = page.frameLocator('iframe[title="Your site"]');
    await expect(preview.locator("body")).toContainText("Welcome to the bakery", { timeout: 150_000 });

    chatcms.writeInEveryCopy(project, "index.qmd", "---\ntitle: Preface\n---\n\nThe oven is warm before dawn.\n");
    await expect(preview.locator("body")).toContainText("The oven is warm before dawn", { timeout: 60_000 });
  });
});
