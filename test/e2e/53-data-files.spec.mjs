// Bean chatcms-hf29: a data file and the site's configuration are read and
// saved whole, in their own format, as the Quiqr desktop app writes them.
import { mkdtempSync, mkdirSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

import { test, expect, saves, worksOnAPhone } from "./fixtures.mjs";

test("a JSON data file and the TOML configuration are edited and published whole", async ({ signedIn: page, chatcms }, testInfo) => {
  const project = chatcms.projectId();
  const base = chatcms.onTheRemote(project, "cat-file", "blob", "refs/heads/main:quiqr/model/base.yaml");
  const extra = mkdtempSync(join(tmpdir(), "data-files-"));
  mkdirSync(join(extra, "quiqr", "model"), { recursive: true });
  mkdirSync(join(extra, "data"), { recursive: true });
  writeFileSync(
    join(extra, "quiqr", "model", "base.yaml"),
    String(base).replace(
      "singles:\n",
      "singles:\n  - key: hours\n    title: Opening hours\n    file: data/hours.json\n    fields:\n      - key: note\n        title: Note\n        type: string\n" +
        "  - key: siteConfig\n    title: Site settings\n    file: hugo.toml\n    fields:\n      - key: title\n        title: Site title\n        type: string\n",
    ),
  );
  writeFileSync(join(extra, "data", "hours.json"), '{\n  "note": "Closed on Sundays"\n}');
  chatcms.layOver(project, extra, "Opening hours and settings in the model");

  await page.goto(`${chatcms.url}/o/technative/p/korenaar/sync`);
  await page.getByRole("button", { name: "Check for updates" }).click();
  await expect(page.getByText(/Joris updated this site/)).toBeVisible({ timeout: 60_000 });

  await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/hours`);
  await expect(page.getByRole("textbox", { name: "Note" })).toHaveValue("Closed on Sundays");
  await saves(page, () => page.getByRole("textbox", { name: "Note" }).fill("Open on Sundays"));

  await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/siteConfig`);
  await expect(page.getByRole("textbox", { name: "Site title" })).toHaveValue("Korenaar Bakery");
  await saves(page, () => page.getByRole("textbox", { name: "Site title" }).fill("Korenaar Bakery & Co"));

  await page.goto(`${chatcms.url}/o/technative/p/korenaar/publish`);
  await page.getByRole("button", { name: /^Publish \d+ changes?$/ }).click();
  await expect(page.getByText("Your site is up to date")).toBeVisible({ timeout: 90_000 });

  expect(chatcms.onTheRemote(project, "cat-file", "blob", "refs/heads/main:data/hours.json")).toBe('{\n  "note": "Open on Sundays"\n}');
  expect(chatcms.onTheRemote(project, "cat-file", "blob", "refs/heads/main:hugo.toml")).toBe(
    'baseURL = "https://korenaar.example/"\nlanguageCode = "nl"\ntitle = "Korenaar Bakery & Co"',
  );

  await worksOnAPhone(page, testInfo);
});
