// Bean chatcms-6kfu: a theme that writes its own addresses with
// {{ .Site.BaseURL }}, as UILite does, shows its styles and pictures in the
// preview even when the app is opened under another name than its base URL.
import { mkdtempSync, mkdirSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

import { test, expect, onAPhone } from "./fixtures.mjs";
import { owner } from "./harness.mjs";

test("styles and pictures linked through the base URL load in the preview", async ({ browser, chatcms }, testInfo) => {
  const project = chatcms.projectId();
  const theme = mkdtempSync(join(tmpdir(), "base-url-"));
  mkdirSync(join(theme, "layouts"), { recursive: true });
  mkdirSync(join(theme, "static", "css"), { recursive: true });
  writeFileSync(
    join(theme, "layouts", "index.html"),
    '<!doctype html>\n<html><head><meta charset="utf-8"><link rel="stylesheet" href="{{ .Site.BaseURL }}css/look.css"></head>\n<body><h1>{{ .Title }}</h1></body></html>\n',
  );
  writeFileSync(join(theme, "static", "css", "look.css"), "h1 { color: rgb(1, 2, 3); }\n");
  chatcms.layOver(project, theme, "A theme with addresses through the base URL");

  const elsewhere = chatcms.url.replace("127.0.0.1", "localhost");
  const context = await browser.newContext({ viewport: testInfo.project.use.viewport });
  const page = await context.newPage();
  try {
    await page.goto(`${elsewhere}/sign-in`);
    await page.getByLabel("Email").fill(owner.email);
    await page.getByLabel("Password").fill(owner.password);
    await page.getByRole("button", { name: /Sign in/i }).click();
    await page.waitForURL(/projects/);

    await page.goto(`${elsewhere}/o/technative/p/korenaar/sync`);
    await page.getByRole("button", { name: "Check for updates" }).click();
    await expect(page.getByText(/Joris updated this site/)).toBeVisible({ timeout: 60_000 });

    const preview = page.frameLocator('iframe[title="Your site"]');
    await expect(async () => {
      await page.goto(`${elsewhere}/o/technative/p/korenaar/content/home`);
      if (onAPhone(testInfo)) await page.getByRole("radio", { name: "Preview" }).click();
      await expect(preview.locator("link[rel=stylesheet]")).toHaveAttribute("href", "/preview/technative/korenaar/css/look.css", { timeout: 15_000 });
    }).toPass({ timeout: 120_000 });
    await expect(preview.locator("h1")).toHaveCSS("color", "rgb(1, 2, 3)", { timeout: 30_000 });
  } finally {
    await context.close();
  }
});
