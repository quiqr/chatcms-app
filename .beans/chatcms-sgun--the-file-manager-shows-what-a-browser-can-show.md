---
# chatcms-sgun
title: The file manager shows what a browser can show
status: completed
type: feature
created_at: 2026-11-27T16:56:42Z
updated_at: 2026-11-27T18:30:00Z
openspec-link: openspec/changes/archive/2026-11-27-file-previews
---

In Setup, Files a picture, a film or a PDF now says it cannot be opened
here, because it is not text. Whatever a browser can show by itself should
be shown instead: images, video, audio and PDF.

- The server serves the file's bytes for a known kind only, chosen by its
  extension, with its own content type and `nosniff`, so a file cannot pass
  for something else. Video can be scrubbed through (range requests).
- The file screen shows an image as a picture, video and audio with the
  browser's own controls, and a PDF in a frame, with a link to open the file
  in a new tab.
- An SVG stays text and opens in the editor.
- Anything else still says it cannot be opened here.

## Summary of Changes

- Setup, Files shows images, video and audio with the browser's own
  controls, and a PDF in a frame, each with a link to a new tab.
- A new route serves a file's bytes for a known kind only, chosen by
  extension, with its own content type, `nosniff`, ranges, and a sandbox
  policy for everything but PDF. Other files are refused.
- SVG stays text and opens in the editor.
