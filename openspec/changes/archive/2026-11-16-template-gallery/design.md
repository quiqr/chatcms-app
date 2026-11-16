# Design

## Context

`internal/onboarding/import.go` has a `SourceTemplate` kind that looks a name
up in an empty `Templates` map and clones the result as the site. The export
kind already creates a site from files with `sites.Create` and sends it to the
project's remote. `internal/jobs` has a queue with `EnqueueAt`, retries, and a
worker that runs registered handlers. The visual source of truth is
`docs/design/template-gallery/`.

## Goals / Non-Goals

**Goals:**

- The gallery works from the last good feed even when the feed host is down.
- One import path: a template is a source like the other two.

**Non-Goals:**

- Building with the template's Hugo version (microVMs, later).
- Linking a project to its template for later updates.
- Hosting repositories in ChatCMS.

## Decisions

### The feed is refreshed by a self-scheduling job and kept in the database

A `templates-refresh` job fetches the feed, validates it, stores the result,
and enqueues itself again with `EnqueueAt`: an hour later, or five minutes
later when a fetch failed and no copy exists yet. The server enqueues one at
start-up when the feed is configured. The last good copy lives in a
`template_feed` row (feed address, JSON, fetched at), so a restart does not
empty the gallery and two servers share one copy. A failed fetch leaves the
row alone. The handler never returns an error: a retried job would be set
back to queued and collide with the next run it already queued.

Alternative: an in-memory cache with a ticker. Simpler, but a restart while
the feed host is down would show the error state for no reason.

### Screenshots are fetched on first request and kept on disk

`GET /api/organisations/{org}/templates/{id}/screenshot` serves from
`<state>/template-screenshots/<id>@<feed fetched at>`, fetching it the first
time and again after each new copy of the feed. The fetch time is in the name
because comparing file times with database times is not reliable to the
millisecond. The response carries a long cache
lifetime. The type is checked by reading the bytes, as the media library
does, never trusted from the feed.

Alternative: fetch all screenshots in the refresh job. That fetches images
nobody looks at, and slows the job that matters.

### Starting from a template creates the site like an export does

The template import clones `SourceLink` into a temporary directory (a full
clone of one line: git's simple HTTP protocol cannot do shallow ones), drops
its `.git`, and calls `sites.Create` with those files and the commit message
"Start from <name>", which sends it to the project's remote. Before that it
asks the remote for its refs; any ref at all means the repository is not
empty, and the import is refused before anything is sent.

The `sync/repository` seam stays as it is: this is the export path with
another source of files.

### The template's Hugo version is stored on the project, not used yet

A migration adds `hugo_version` (text, empty when unknown) and
`hugo_extended` (bool) to projects. The template import sets them. The build
runner does not read them; it keeps the pinned Hugo. Settings and the step
compare them to the installation's Hugo version, which the server reads once
at start-up from `hugo version`.

### The two existing calls are reused

The project's Hugo version has its own `GET/PUT …/projects/{project}/hugo`,
served beside the template list, so the project settings call keeps its
shape.

The step creates the project with `POST /api/organisations/{org}/projects`,
then the onboarding continues as for a repository: deploy key, then
`POST …/import` with kind `template`, the template id and the remote. The
hand-off's single `from-template` call is not added. Name, limit and
emptiness refusals come from these calls with their own keys and English
reasons, so the screens say them in the reader's language.

### Feed fields are mapped once, in Go

The feed's `QuiqrEtalage*` names stay at the boundary. One function maps an
entry to the template the API returns (`name`, `description`, `author`,
`hugoVersion`, `extended`, `model` as a list, …) and drops invalid entries.
It is tested against a recorded copy of the live feed in
`test/fixtures/templates/templates.json`.

## Risks / Trade-offs

- [A template's Hugo version is newer or older than the pinned one, and its
  build fails] → The step and the settings say so before anything is built.
  The build log already names the error.
- [The feed host serves something hostile] → Entries are validated, URLs are
  only used server-side or as links, screenshots are type-checked from their
  bytes, and a template's site builds in the existing sandbox.
- [The remote is empty but the deploy key cannot write] → The deploy key step
  already tests read and write before the import runs.
