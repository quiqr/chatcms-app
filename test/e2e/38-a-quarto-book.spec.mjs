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
});
