# Proposal

Bean: [chatcms-sgun](../../../.beans/chatcms-sgun--the-file-manager-shows-what-a-browser-can-show.md)
(the file manager shows what a browser can show).

## Why

In Setup, Files a picture, a film or a PDF says it cannot be opened here.
An implementer looking for the right logo or checking a document has to
leave ChatCMS for it, while the browser can show these files by itself.

## What Changes

- **The server** serves the bytes of one file of a known kind, chosen by its
  extension: images, video, audio and PDF. It sets the kind's own content
  type, `nosniff`, and serves ranges so video can be scrubbed. Other files
  are refused.
- **Reading a file** says which kind of preview it has, for a file that is
  not text or too large to open as text.
- **The file screen** shows an image as a picture, video and audio with the
  browser's controls, and a PDF in a frame, each with a link that opens the
  file in a new tab.

Assumptions: an SVG is text and keeps opening in the editor. A PDF is served
without a sandbox policy, because browsers refuse to show a PDF under one;
`nosniff` and the extension list keep any other kind from being served.

## Capabilities

### New Capabilities

### Modified Capabilities

- `content/project-files`: a file the browser can show is shown.

## Impact

- `internal/content`: the preview kind and opening a file's bytes.
- `internal/httpapi`: one route, `GET .../file-content/{path...}`.
- `web/src/routes/FilesPage.tsx`: the preview.
