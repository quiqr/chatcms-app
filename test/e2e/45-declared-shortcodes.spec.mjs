import { readdirSync, readFileSync } from "node:fs";
import { join } from "node:path";

import { test, expect, onAPhone, saves, worksOnAPhone } from "./fixtures.mjs";

const declarations = `
shortcodes:
  - key: figure
    title: Picture
    fields:
      - key: src
        type: image-select
        title: Image
        path: static/images
      - key: caption
        type: string
        title: Caption
  - key: notice
    title: Notice
    inner: markdown
    fields:
      - key: type
        type: select
        title: Kind
        position: 0
        options:
          - value: info
            text: Information
          - value: warning
            text: Warning
`;

const body = `
Intro paragraph.

{{< figure src="/images/old.png" class="wide" caption="Our loaf" >}}

{{% notice info %}}
Fresh every day.
{{% /notice %}}

Last line.
`;

function bodies(chatcms, project, path) {
  const people = join(chatcms.state, "sites", project, "people");
  return readdirSync(people).map((person) =>
    readFileSync(join(people, person, "working-copy", path), "utf8").replace(/^---\n[\s\S]*?\n---\n/, ""),
  );
}

test("a declared shortcode is edited through its form, and inserted", async ({ signedIn: page, chatcms }, testInfo) => {
  test.setTimeout(240_000);
  const project = chatcms.projectId();
  const model = chatcms.onTheRemote(project, "show", "main:quiqr/model/base.yaml");
  chatcms.pushFromSomewhereElse(project, "quiqr/model/base.yaml", model + declarations, "Declare two shortcodes");
  chatcms.pushFromSomewhereElse(project, "content/news/bakery.md", `---\ntitle: The bakery\n---\n${body}`, "A page with shortcodes");
  await page.goto(`${chatcms.url}/o/technative/p/korenaar/sync`);
  await page.getByRole("button", { name: "Check for updates" }).click();
  await expect(page.getByText(/Joris updated this site/)).toBeVisible({ timeout: 60_000 });

  await page.goto(`${chatcms.url}/o/technative/p/korenaar/content/news/bakery`);
  const text = page.getByRole("textbox", { name: "Page text" });
  const saved = () => bodies(chatcms, project, "content/news/bakery.md");

  await test.step("the figure's caption and picture are changed in its form", async () => {
    await text.getByRole("button", { name: "Edit Picture" }).click();
    const form = page.getByRole("dialog", { name: "Picture" });
    await form.getByRole("textbox", { name: "Caption" }).fill("Fresh loaf");
    await form.getByRole("button", { name: "Replace" }).click();
    const library = page.getByRole("dialog", { name: "Choose a picture" });
    await library.getByText("loaf.png").first().click();
    await library.getByRole("button", { name: "Use this image" }).click();
    await expect(page.getByRole("dialog", { name: "Picture" }).getByRole("textbox", { name: "Image" })).toHaveValue("/images/loaf.png");
    await saves(page, () => page.getByRole("dialog", { name: "Picture" }).getByRole("button", { name: "Done" }).click());
    await expect.poll(saved).toContain(body.replace('src="/images/old.png" class="wide" caption="Our loaf"', 'src="/images/loaf.png" class="wide" caption="Fresh loaf"'));
  });

  await test.step("the notice's kind and text are changed", async () => {
    await text.getByRole("button", { name: "Edit Notice" }).click();
    const form = page.getByRole("dialog", { name: "Notice" });
    await form.getByLabel("Kind").selectOption("warning");
    await form.getByRole("textbox", { name: "Text inside" }).fill("Fresh **every** day.");
    await saves(page, () => form.getByRole("button", { name: "Done" }).click());
    await expect.poll(saved).toContain(
      body
        .replace('src="/images/old.png" class="wide" caption="Our loaf"', 'src="/images/loaf.png" class="wide" caption="Fresh loaf"')
        .replace("{{% notice info %}}\nFresh every day.", "{{% notice warning %}}\nFresh **every** day."),
    );
  });

  if (process.env.CHATCMS_SHOTS) {
    await text.getByRole("button", { name: "Edit Picture" }).scrollIntoViewIfNeeded();
    await page.screenshot({ path: `${process.env.CHATCMS_SHOTS}/declared-shortcodes/${testInfo.project.name}.png` });
  }

  await test.step("a figure is inserted from the toolbar", async () => {
    await text.getByText("Intro paragraph.").click();
    await page.getByRole("button", { name: "Insert" }).click();
    await page.getByRole("menuitem", { name: "Picture" }).click();
    const form = page.getByRole("dialog", { name: "Picture" });
    await form.getByRole("textbox", { name: "Caption" }).fill("A new one");
    await saves(page, () => form.getByRole("button", { name: "Done" }).click());
    await expect.poll(() => saved().some((one) => one.includes('Intro paragraph.\n\n{{< figure caption="A new one" >}}\n\n{{< figure src='))).toBe(true);
  });

  if (onAPhone(testInfo)) await worksOnAPhone(page, testInfo);
});
