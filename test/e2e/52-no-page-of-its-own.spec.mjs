// Bean chatcms-a6on: an item with no page of its own says so in the preview,
// and can show the home page instead.
import { mkdtempSync, mkdirSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

import { test, expect, onAPhone, worksOnAPhone } from "./fixtures.mjs";

test("data without a page of its own is explained beside the form", async ({ signedIn: page, chatcms }, testInfo) => {
  const project = chatcms.projectId();
  const base = chatcms.onTheRemote(project, "cat-file", "blob", "refs/heads/main:quiqr/model/base.yaml");
  const extra = mkdtempSync(join(tmpdir(), "no-page-"));
  mkdirSync(join(extra, "quiqr", "model"), { recursive: true });
  mkdirSync(join(extra, "data"), { recursive: true });
  writeFileSync(
    join(extra, "quiqr", "model", "base.yaml"),
    String(base).replace("singles:\n", "singles:\n  - key: hours\n    title: Opening hours\n    file: data/hours.json\n    fields:\n      - key: note\n        title: Note\n        type: string\n"),
  );
  writeFileSync(join(extra, "data", "hours.json"), '{\n  "note": "Closed on Sundays"\n}\n');
  chatcms.layOver(project, extra, "Opening hours as data");

  await page.goto(`${chatcms.url}/o/technative/p/korenaar/sync`);
  await page.getByRole("button", { name: "Check for updates" }).click();
  await expect(page.getByText(/Joris updated this site/)).toBeVisible({ timeout: 60_000 });

  await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/hours`);
  await expect(page.getByRole("textbox", { name: "Note" })).toHaveValue("Closed on Sundays");
  if (onAPhone(testInfo)) await page.getByRole("radio", { name: "Preview" }).click();
  else await page.getByRole("radio", { name: "Side by side" }).click().catch(() => undefined);

  await expect(page.getByText("No page of its own")).toBeVisible({ timeout: 60_000 });
  await page.getByRole("button", { name: "Show the home page" }).click();
  await expect(page.frameLocator('iframe[title="Your site"]').locator("body")).toContainText("Korenaar", { timeout: 90_000 });

  await worksOnAPhone(page, testInfo);
});
