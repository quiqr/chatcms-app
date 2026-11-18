# Design

## Context

See proposal.md. Both workspace backends start Hugo with flags of their own;
the micro VM gets `chatcms.hugo` on its kernel command line; `sitebuild`
runs Hugo with fixed flags; `app.hugos` gives a project's Hugo (its recorded
version through `nixpin`, else the installation's). Previews are served on
the app's origin under `/preview/{org}/{project}/` behind the session.

## Decisions

### 1. The format

```yaml
# quiqr/runtime.yaml
packages:
  hugo: "0.145.0"
  dart-sass: "1.86.0"
preview:
  type: web                     # web | pdf | other
  open: /                       # what the pane shows, under the preview
  command: [hugo, server, --bind, "{host}", --port, "{port}",
            --baseURL, "{url}", --appendPort=false, --renderToMemory,
            --buildDrafts, --noBuildLock, --cacheDir, "{cache}",
            --liveReloadPort, "{reload}", --poll, 500ms]
build:
  output: public
  command: [hugo, --destination, "{output}", --baseURL, "{url}",
            --cacheDir, "{cache}", --noBuildLock]
```

`internal/runtime` parses it (strict: unknown keys are refused), checks
types, placeholders and that each command's program is one the packages
provide (`hugo` by `hugo`, otherwise the package's name or its `meta.mainProgram`
is not knowable without evaluation, so a program is accepted when a file of
that name is in one of the packages' `bin` directories after realising).
`Default(hugoVersion)` is today's Hugo setup written in this format, so there
is one code path.

`runtime.Resolved` holds the program's absolute path, the arguments with
placeholders still in them, the packages' `bin` directories, the type and the
page to open. `Fill(words, values)` replaces whole-word placeholders only.

### 2. Who resolves it

`app.runtimes.For(ctx, project, dir)` reads `quiqr/runtime.yaml` from the
given copy (a person's for previews, the shared one for builds), falls back
to `Default` with the project's recorded Hugo, realises each package through
`nixpin` (or, without Nix, accepts only `hugo` and gives the installation's),
and finds the programs. The manager's `Settings.HugoFor` becomes
`Settings.RuntimeFor(ctx, key, dir)`; `Spec.Hugo` becomes `Spec.Run`.
`sitebuild.Builder.UseHugoOf` becomes `UseRuntimeOf`.

### 3. Backends

- local: `{host}` 127.0.0.1, `{port}` a free port, `{url}` the preview
  address, `{reload}` the browser's port, `{source}` the person's copy,
  `{cache}` beside it; `PATH` is the packages' `bin` directories.
- microvm: filled on the host with 0.0.0.0, 1313, `/site` and
  `/run/preview/cache`, then passed as `chatcms.run=<base64url JSON
  {"cmd": [...], "path": [...]}>`; the guest decodes it with `jq` and runs it
  as uid 65534. A command line past 1900 bytes is refused with a reason.
  `chatcms.hugo` goes away.

### 4. Builds

`sitebuild.Build` runs the resolved build command through the runner with
`{output}` as the declared output directory, collects that directory, and
`Pages` reads the declared output: `runner.Result` gains `Root`, the
directory inside `Outputs` that holds the pages.

### 5. The preview origin

`CHATCMS_PREVIEW_URL` (e.g. `https://preview.chatcms.eu`). When set:

- the manager makes a random view key (32 bytes, base64url) at each start;
  the preview address is `<preview>/v/<key>/` and Hugo's base URL is that;
- requests whose `Host` is the preview host are served by the preview origin
  only: `/v/<key>/…` finds the warm workspace by its key, checks its person
  may still read the project, and proxies the rest of the path; anything else
  is not found;
- the app origin's `/preview/` route answers not found;
- `GET …/workspace` gives the pane `address` and `type` and `open`.

Without it, `address` is `/preview/{org}/{project}/` as before.

### 6. The pane

`Preview.tsx` takes `address`, `type` and `open` from the workspace:
`web` frames `address + page`, `pdf` frames `address + open` (the browser's
own viewer; refreshed on every save), `other` offers `address + open` as a
download.

## Risks / Trade-offs

- [A view key in an address can be passed on] → it lasts as long as the
  workspace, every request checks the person's right, and it is never
  logged.
- [The preview origin on another site] → nothing depends on cookies there.
- [Kernel command line size] → measured and refused above 1900 bytes; a few
  packages fit easily.
