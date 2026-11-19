import { test, expect, onAPhone } from "./fixtures.mjs";

test.use({ previewApart: true });

const stream = "BT /F1 24 Tf 20 100 Td (Opening hours) Tj ET";
const pdf = `%PDF-1.4
1 0 obj << /Type /Catalog /Pages 2 0 R >> endobj
2 0 obj << /Type /Pages /Kids [3 0 R] /Count 1 >> endobj
3 0 obj << /Type /Page /Parent 2 0 R /MediaBox [0 0 300 200] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >> endobj
4 0 obj << /Type /Font /Subtype /Type1 /BaseFont /Helvetica >> endobj
5 0 obj << /Length ${stream.length} >> stream
${stream}
endstream endobj
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
  await expect(frame).toHaveAttribute("src", new RegExp(`^${chatcms.previewUrl}/v/[A-Za-z0-9_-]+/__chatcms/pdf/\\?file=report\\.pdf(&fit=width)?&lang=en$`), { timeout: 90_000 });

  await test.step("PDF.js shows its text, which can be selected", async () => {
    const viewer = page.frameLocator('iframe[title="Your site"]');
    await expect(viewer.locator(".textLayer")).toContainText("Opening hours", { timeout: 60_000 });
  });

  await test.step("the PDF itself comes from the preview", async () => {
    const src = await frame.getAttribute("src");
    const file = new URL("../../report.pdf", src).href;
    const answer = await page.request.get(file);
    expect(answer.status()).toBe(200);
    expect((await answer.body()).subarray(0, 5).toString()).toBe("%PDF-");
  });
});
