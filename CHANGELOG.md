# Changelog

All notable changes to Quiqr ChatCMS are documented here.
The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Everyone edits their own copy of a site, with a preview of their own; two
  people can preview different unfinished changes at once.
- Saves reach the shared version by themselves; the same field changed by two
  people is asked about, and each keeps seeing their own until answered.
- Opening a project starts its preview; the preview says it is being prepared,
  and implementers can unfold the steps and the log.
- `services.chatcms.workspaceIdle` and `workspacesAtOnce` set how long an
  unused preview stays and how many run.
- Previews can run in a micro VM each (`workspaceBackend = "microvm"`), seeing
  only the package store and the person's copy, reaching nothing but the host.
- A project's own Hugo version is used for its preview and its builds, found
  once through Nixhub and built with Nix (`pinHugo`).
- An AI gateway on the host: workspaces reach the model with a token of their
  own, never the provider's key (`services.chatcms.ai`).
- Per organisation, the platform admin can cap AI requests a minute and tokens
  a month, and sees this month's use.
- A chat assistant in the editor: ask for a change in words, see the reply as
  it is written, and Show or Undo each page it changed.
- A site can declare its packages, preview and build in `quiqr/runtime.yaml`
  (`docs/RUNTIME.md`); a preview can be a web page, a PDF or a file.
- Previews can come from an origin of their own (`services.chatcms.previewUrl`,
  `CHATCMS_PREVIEW_URL`), away from the app's cookies.
- The preview pane shows a PDF preview, and offers to open any other result.
- Select text in the preview to ask the assistant about it; it changes the
  right page, and the new text lights up when the preview reloads.
- PDF previews show through PDF.js, selectable and at the same page after a
  refresh.
- Quarto sites: preview, build in the sandbox with Python and R cells, and a
  book's website and PDF from one publish.
- A workflow's `outputs` send each part of a build to publish, to Artifacts or
  to both; Artifacts lists and downloads what each publish kept.
- A runtime package may name its nixpkgs revision, and an installation may
  provide packages itself (`services.chatcms.packages`).
- OpenCode Go works as the AI provider (`ai.sessionHeader`); developers can
  run the assistant and its tests against a real model (`docs/TESTING.md`).
- When two people changed the same field, the AI can combine the two; the
  person edits and uses the result, or keeps either version.
- A PDF preview has a toolbar: the page in view and going to any page, zoom
  with a way back to the actual size, the PDF's contents, and download.
- The page text is edited formatted; an edit rewrites only the blocks it
  touched, and Hugo shortcodes show as locked blocks.

### Changed

- The page no longer scrolls: the header stays and each panel scrolls on its
  own, so the preview's end and the assistant's message box stay in view.
- A narrow editor shows form or preview instead of stacking them; the open
  assistant takes the form's place beside the preview.
- The folded content menu names the open page or collection, with its count
  and a dot for what waits to be published.
- A breadcrumb above the collection view and the editor opens a picker with
  quick find and the whole content menu, also while the menu is folded.
- `CHATCMS_WORKSPACE_IDLE` and `CHATCMS_WORKSPACES_AT_ONCE` replace the preview
  names, which are still read for one release.
- Publishing first brings the publisher's own saves in, and waits while a
  question about them is open.

### Fixed

- The preview's live reload passes through the app again, and event streams
  are flushed as they are written.
- Publishing the built site no longer fails when its settings name no author.
- The preview shows when the app is reached by another name than its base URL.
- A project's card says what it runs on, Quarto too; a PDF preview keeps its
  size; a Quarto preview has time to render before it is given up on.
- A Quarto preview reloads by itself when its sources change.
- After an answer replays waiting work, every person's copy shows the answered
  version, instead of keeping the old one for good.
- Output a preview writes into a person's copy no longer stops other people's
  changes from arriving, or shows as saving.
- A person without a display name can save; their address is used as the
  author.
- A question about what two people changed can be answered later, and
  checking for updates while asked shows it again.

## [0.2.0] - 2026-11-17

### Added

- A new project can start from a community template: a gallery with search,
  then one fresh version copied into an empty place in git.
- A project records the Hugo version it asks for, shown and changeable in its
  settings.
- `services.chatcms.templateFeed` sets the template feed, or switches the
  gallery off.
- A template or Quiqr Desktop export can start without a git address; ChatCMS
  keeps it until it is connected.
- Project settings connect a kept site to an empty repository, sending every
  version there; publishing waits until then.
- The content screen follows the model's menu; a developer's section is shown
  only to implementers.
- A content column beside every content screen, folding away in the editor;
  a drawer on tablets and its own screens on a phone.
- A collection view with search, folder tabs, sort, status badges and paging;
  file names for implementers.
- A new-item dialog that checks the web address before making anything.
- The editor says "12 of 48" and steps to the previous and next item; a
  "Show on website" switch for a draft field.
- Quick find: Ctrl+K or Cmd+K opens any page by its title.
- Undo and redo in the editor, with a list of the steps since the page was
  opened; an undo is saved like any edit.
- An About screen in the platform admin area: version, what it runs on, the
  license, the source code and the credits.

### Changed

- Collections list items as Quiqr Desktop does: subfolders, page bundles,
  `hideIndex`, `sortkey`, and titles from front matter.
- The navigation and the content menu fold by hand on every screen, to a rail
  and a strip, and stay as left.
- Shipping a change runs the ship tier, with the acceptance tests side by
  side; a release still runs every check.
- A new item under `content/` is a page bundle, `<slug>/index.md`; renaming,
  copying and removing a bundle move its whole folder.

### Removed

- "Sign out everywhere" is gone from the person menu; "Sign out" is the one
  way out.

### Fixed

- A project with nothing brought in yet says so on every screen, instead of
  "Something went wrong".
- A model with a collection called `pages` no longer clashes with the Pages
  group.
- Opening a page no longer folds the navigation by itself.

## [0.1.0] - 2026-11-16

### Added

- Releases: `scripts/release.sh patch|minor|major` bumps, gates, tags and
  publishes; see `docs/RELEASING.md`.
- In development, the sign-in and forgot-password screens link to the
  instance's Mailpit inbox.
- `scripts/dev-fe.sh` starts the frontend for the backend's instance and host,
  and warns when no backend answers.
- In development, the sign-in screen signs in as each seeded role in one press;
  a production build carries none of it.
- `scripts/reference.sh` runs the old Quiqr with a site, and a gated test
  replays its recorded edits; only strings are recorded so far.
- A site can be duplicated, moved to another organisation, downloaded as a zip
  or archived.
- The platform admin manages every account: fix a name or address, reset,
  block, delete, and grant platform admin.
- An organisation has its own screens: its sites, its people and its usage
  against its limits.
- An organisation can be limited in members and sites; a refusal names the
  limit and how to get under it.
- Dutch, with every string in one catalogue in `web/src/strings/`; a missing
  key fails the build.
- The platform admin can act as a customer, with a band on every screen and
  history naming both people.
- The platform admin sees every person with their organisations and sites, and
  can invite one or turn one off.
- The platform admin can rename an organisation, change its people, and remove
  it once it holds no sites.
- Editing: a page opens as a form built from the site's Quiqr model and saves
  by itself, byte for byte.
- A content browser: pages and collections, search, unpublished marks, and
  making, renaming, copying and removing pages.
- A picture library: add, search, describe, see where a picture is used, and
  pick one from a field.
- Publishing: a review of what changed per page and field, an optional note,
  and the destination and form set per project.
- History in plain language, and going back to an earlier version, which can
  itself be undone.
- Live preview while editing, served through the app so unpublished work stays
  private.
- Declarative workflows in `quiqr/workflows/` for what happens on save, preview
  and publish.
- Site builds run isolated, with time and memory limits and no network.
- Sync status in five states, and a question per field when two people changed
  the same thing.
- A NixOS module, with a VM test that boots it from an empty database; see
  `docs/DEPLOYING.md`.
- A dark appearance, or following the device, kept with the account.
- Every date and time is written one way, in one place.
- An account has a picture, kept outside every site so it is never published.
- An email address can be changed once the link sent to the new address is
  followed.
- You can see where you are signed in and end any of those sessions.
- An organisation admin invites, changes and removes its people, and can resend
  or withdraw an invitation.
- Every screen has a person menu with your account, signing out and signing
  out everywhere.
- An account can be turned off: its sessions stop at once and its history
  keeps its name.
- You can change your own name and password; the current password is asked
  for first.
- Project settings: rename, move the site, change publishing, or remove the
  project without touching the site.
- Organisations are real: the address names one, the switcher moves between
  them, and renaming keeps the address.
- Invitations and password resets are sent by mail through the job queue;
  without SMTP the link goes to the log.
- Screens to choose a password from an invitation or reset and to ask for a
  reset; setting one signs you in.
- An application administrator runs the installation without any power over
  customer content.
- The nine acceptance tests run in a real browser at desk and phone width in
  `nix flake check`.
- Project scaffold: OpenSpec, beans, Nix flake, ship gate and the milestone 01
  backlog.
- `chatcmsd` serves a JSON API, reports its health at `/api/health` and stops
  without dropping a request.
- The server creates and upgrades its database schema on start-up, and refuses
  a schema newer than itself.
- A deployment is one binary, with the browser application inside it.
- Configuration values can come from files, which keeps secrets out of the Nix
  store.
- Quiqr models are read as the desktop app reads them: YAML, TOML or JSON, with
  includes and partials.
- All 27 field types of the desktop app are understood; an unknown type is kept
  as a placeholder.
- An unusable model says which file and which part of it is wrong.
- The pages and collections to browse, with counts and labels, come from the
  model.
- Every edit makes a version attributed to whoever made it.
- Changes made outside ChatCMS arrive by themselves, beneath the work in
  progress.
- When two people change one page, both versions are offered; neither is lost
  and no markers appear.
- Publishing and syncing run in the background, are retried on failure and
  resume after a restart.
- No failure record carries a password or a key.
- A saved page is spelled exactly as the Quiqr desktop app spells it, so
  adopting ChatCMS rewrites nothing.
- Editing one field changes one line, and every other value stays where it
  was.
- TOML and JSON front matter can be read but not saved, and the page says so.
- The interface has the Quiqr look, with its own phone frame and bottom tab
  bar.
- Typefaces are served by ChatCMS itself, so no browser contacts a third party.
- The build fails on a version control word in editor screens, a colour off
  the palette or a control too small.
- Signing in, staying signed in on a device, and signing out, which ends the
  session on the server.
- Invitation and reset links work once, and a reset ends every session of that
  person.
- Nothing ChatCMS says reveals whether an address has an account.
- Repeated guessing is slowed down, per address and per client.
- Organisations hold projects, with roles that each have exactly the powers
  the design gives them.
- Inviting by address and role; someone new gets an account and a link to set
  a password.
- A viewer can read and change nothing, whatever the screen shows.
- What someone may not do is left out of the interface, not greyed out.
- Another customer's project looks exactly like one that does not exist.
- The project list and the people and roles screen show real data.
- An existing site can be brought in from git, from a Quiqr desktop export or
  from a template.
- Each project gets its own deploy key; the private half is encrypted and never
  shown or logged.
- After an import, ChatCMS says whether it found a Hugo site and a Quiqr model,
  and what to do if not.
- The connection can be tested at any time, with an answer that says what to
  do next.
- Where and in what form a site is published is set per project.

### Changed

- Changelog entries are at most two lines of 80 characters; the gate checks it.
- The version lives only in `VERSION`, and a plain `go build` reports it.
- **Breaking for developers:** `scripts/dev.sh` is now `scripts/dev-be.sh`.
- The browser tab shows the Quiqr icon, and the frontend check fails without
  it.
- The project switcher opens the organisation's projects, on desktop and on
  the phone.
- Everything clickable shows the pointing hand, and the interface check guards
  it.
- Every screen has a top bar with the person menu at its right, naming the role
  held there.
- The frontend check gives each test 30 seconds instead of 5, for the load of
  the whole gate.
- The frontend tests run with at most four workers and paste text instead of
  typing it.
- Building the server no longer runs the frontend tests; `nix flake check`
  still does.
- The platform admin has an area: every organisation with its admins, usage
  and limits, and a full removal.
- Roles have three levels, platform, organisation and project; existing owners
  become admins or implementers.

### Fixed

- Choosing an organisation in the switcher opens it, where it did nothing
  before.
- The count on the Publish button is the real one, the same as on the publish
  screen.
