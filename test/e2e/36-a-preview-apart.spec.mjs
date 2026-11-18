import { test, expect, onAPhone } from "./fixtures.mjs";

test.use({ previewApart: true });

const pdf = `%PDF-1.4
1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj
2 0 obj << /Type /Pages /Kids [3 0 R] /Count 1 >> endobj
3 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] >> endobj
trailer << /Root 1 0 R >>
%%EOF
`;

const runtime = `packages:
  hugo: "0.139.0"
preview:
  type: pdf
  open: /report.pdf
  command: [hugo, server, --source, "{source}", --bind, "{host}", --port, "{port}", --baseURL, "{url}", --appendPort=false, --renderToMemory, --noBuildLock, --cacheDir, "{cache}", --liveReloadPort, "{reload}", --poll, 500ms]
`;

async function showThePreview(page, testInfo) {
  if (onAPhone(testInfo)) await page.getByRole("radio", { name: "Preview" }).click();
}

test("the preview comes from an origin of its own", async ({ signedIn: page, chatcms }, testInfo) => {
  await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/home`);
  await showThePreview(page, testInfo);

  const frame = page.locator('iframe[title="Your site"]');
  await expect(frame).toHaveAttribute("src", new RegExp(`^${chatcms.previewUrl}/v/[A-Za-z0-9_-]+/$`), { timeout: 90_000 });
  await expect(page.frameLocator('iframe[title="Your site"]').locator("body")).toContainText("Korenaar Bakery", { timeout: 90_000 });

  await test.step("the app's own address no longer serves previews", async () => {
    const answer = await page.request.get(`${chatcms.url}/preview/technative/korenaar/`);
    expect(await answer.text()).not.toContain("Korenaar Bakery");
  });

  await test.step("the preview origin serves nothing of the app", async () => {
    for (const path of ["/api/me", "/sign-in", "/v/not-a-key/"]) {
      const answer = await page.request.get(`${chatcms.previewUrl}${path}`);
      expect(answer.status(), path).toBe(404);
    }
  });
});

test("a site whose runtime makes a PDF shows the PDF", async ({ signedIn: page, chatcms }, testInfo) => {
  chatcms.pushFromSomewhereElse(chatcms.projectId(), "quiqr/runtime.yaml", runtime, "Preview the report");
  chatcms.pushFromSomewhereElse(chatcms.projectId(), "static/report.pdf", pdf, "The report");

  await page.goto(`${chatcms.url}/o/technative/p/korenaar/sync`);
  await page.getByRole("button", { name: "Check for updates" }).click();
  await expect(page.getByText(/Joris updated this site/)).toBeVisible({ timeout: 60_000 });

  await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/home`);
  await showThePreview(page, testInfo);

  const frame = page.locator('iframe[title="Your site"]');
  await expect(frame).toHaveAttribute("src", new RegExp(`^${chatcms.previewUrl}/v/[A-Za-z0-9_-]+/report\\.pdf$`), { timeout: 90_000 });

  const src = await frame.getAttribute("src");
  await expect(async () => {
    const answer = await page.request.get(src);
    expect(answer.status()).toBe(200);
    expect((await answer.body()).subarray(0, 5).toString()).toBe("%PDF-");
  }).toPass({ timeout: 60_000 });
});
