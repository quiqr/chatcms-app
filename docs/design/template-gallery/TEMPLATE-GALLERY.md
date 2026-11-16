# Quiqr ChatCMS — Template gallery ("Start from a template")

Spec for the third way to create a project, next to *Import Git repository*
and *From Quiqr Desktop*: pick a Quiqr community template, name the project,
point to an empty repository, then continue with the normal onboarding steps
(deploy key → publishing → invite people).

Reference screens: `reference/html/*.html` and
`reference/screenshots/{desktop,phone}`. Visual system: the milestone-1 tokens
and components (`tokens/tokens.css`, main `DESIGN.md`).

Template names in the references are the real catalogue entries; screenshots
are coloured placeholders, and only *Summer* has real description/author
data. Render everything from the feed.

---

## 1. Data source

Public feed (same one Quiqr Desktop uses):

```
GET https://quiqr.github.io/quiqr-community-templates/templates.json
```

A JSON array. Fields per entry (from the catalogue's build script):

| Field | Example | Used for |
| --- | --- | --- |
| `QuiqrEtalageName` | `Summer` | Card title |
| `NormalizedName` | `Summer` | Id, screenshot path |
| `QuiqrEtalageDescription` | `Simple fresh looking recipe site, …` | Card text, search |
| `QuiqrEtalageAuthor` | `Connor Bär (Ported by Pim Snel)` | Card meta, search |
| `QuiqrEtalageAuthorHomepage` | URL | Author link (details) |
| `QuiqrEtalageLicense` / `QuiqrEtalageLicenseURL` | `MIT` / URL | Card meta, details |
| `QuiqrEtalageHomepage` | URL | Details |
| `QuiqrEtalageDemoUrl` | URL | **Live demo** button (hide when empty) |
| `HugoTheme` | `summer-qremix` | Details |
| `HugoVersion` | `extended_0.101.0` | Details + project build setting ("0.101.0 extended") |
| `QuiqrModel` | `base collections menu singles` | Details, shown as chips |
| `QuiqrFormsEndPoints` | `0` | Details ("Forms: None" / "3 forms") |
| `ScreenshotImageType` | `jpg` | Screenshot URL |
| `SourceLink` | `https://github.com/mipmip/summer-qremix` | What gets cloned |

Screenshot URL:
`https://quiqr.github.io/quiqr-community-templates/templates/<NormalizedName>/screenshot.<ScreenshotImageType>`

**Backend, not browser:** fetch the feed server-side, cache it (e.g. 1 hour,
serve stale on failure), validate each entry, drop entries without
`SourceLink` or name, and proxy/cache the screenshots so the app doesn't
depend on github.io at view time. Expose:

```
GET /api/templates            → [{ id, name, description, author, authorUrl,
                                   license, licenseUrl, homepage, demoUrl,
                                   hugoTheme, hugoVersion, extended, model: [..],
                                   forms, sourceUrl, screenshotUrl }]
POST /api/orgs/{org}/projects/from-template
     { templateId, name, repoUrl }  → 201 { projectId, deployKey }
                                    | 409 { error: "project_limit" | "name_taken" }
                                    | 422 { error: "repo_not_empty" | "repo_unreachable" }
```

## 2. Gallery (`gallery`, `gallery-loading`, `gallery-error`)

- Header: logo, breadcrumb `<Organisation> / New project / Choose a template`.
- Segmented control at the top: Import Git repository · From Quiqr Desktop ·
  **Start from a template** (selected). Same control as onboarding.
- Title "Start from a template" + one-line intro; search field (name,
  description, author; client-side filter, case-insensitive).
- Count line: "9 templates" / "3 of 9 templates" while searching.
- Card grid `repeat(auto-fill, minmax(270px, 1fr))`, gap 18:
  - Screenshot 16:10, `object-fit: cover`, top aligned; whole image links
    to step 2.
  - Name (16/600), description (14, max 3 lines), "author · licence" (13 muted).
  - Buttons: **Live demo ↗** (outline, opens new tab) and **Use template**
    (primary → step 2).
- Sort: alphabetical by name (feed order is not guaranteed).
- **Loading**: 6 skeleton cards. **Error**: alert card "The template list
  couldn't be loaded" + Try again + "Import a Git repository instead".
  **No match**: "No templates match “…”".
- Footer: link to the community repository ("Want yours listed?").

## 3. Step 2 · Name & repository (`use-template`)

Uses the onboarding step list: **Template ✓ → Name & repository → Deploy key
→ Publishing → Invite people**.

- Left card: screenshot, name, description, details list (Author + link,
  Licence + link, Content model chips, Forms, Hugo version + theme, Source
  link), Live demo, "Choose another template".
- Right form (active card style):
  - **Project name** — required, prefilled empty or with the template name;
    unique within the organisation ("A project with this name already exists
    in TechNative").
  - **Organisation** — select of organisations where the user may create
    projects, with limit info ("4 of 5 projects"). Warn when it uses the last
    free project; block with the hard-limit error from DESIGN.md when full.
  - **Empty repository for your copy** — git URL (mono). Help: create it
    first without a README; next step adds the deploy key.
  - Info: builds with the template's Hugo version; changeable later.
  - Buttons: Back / **Continue to deploy key**.
- On continue: create the project record + deploy key, then show the existing
  Deploy key step. After the key is verified: clone `SourceLink`, copy the
  files into the empty repository with a **fresh history** (one initial
  commit "Start from <template>"), push, then continue to Publishing.
  Errors in plain language: "That repository isn't empty", "ChatCMS can't
  reach that repository yet — check the deploy key".

## 4. Phone (`phone-gallery`)

Single column of cards (screenshot 150 px high, name, one-line description,
**Use** button). Search at the top. Step 2 stacks: template summary, then form.

## 5. Roles

Only users who may create projects in an organisation (organisation admins,
platform admins) see *Start from a template*; others don't get the New
project entry at all.

## 6. Open decisions (assumed for now)

1. The user supplies an empty repository (no ChatCMS-hosted repositories in
   milestone 1). Changed by `hosted-until-connected`: the address may be left
   empty, and ChatCMS then keeps the site in `<state>/hosted/` until it is
   connected to an empty repository in the project settings.
2. The copy starts with a fresh git history and is not linked to the template
   for later updates.

## 7. Acceptance checks

1. The gallery lists every valid entry from the feed, alphabetically, with
   screenshot, description, author, licence and working demo links.
2. Search filters by name, description and author; empty state shows.
3. Feed unavailable → error card with Try again; cached data is used when
   available.
4. Use template → step 2 shows the template's details from the feed.
5. Duplicate project name, full project limit, non-empty or unreachable
   repository each give their message and block continuing.
6. After the deploy key step the repository contains the template files in
   one initial commit, and the project builds with the template's Hugo version.
7. Works at 390 px without horizontal scrolling.
