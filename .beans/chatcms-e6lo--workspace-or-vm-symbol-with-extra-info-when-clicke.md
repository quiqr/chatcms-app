---
# chatcms-e6lo
title: Workspace or VM symbol, with extra info when clicked.
status: completed
type: feature
priority: normal
created_at: 2026-11-19T11:39:24Z
updated_at: 2026-11-19T23:34:57Z
openspec-link: openspec/changes/archive/2026-11-20-workspace-symbol
parent: chatcms-iqog
---

## Summary of Changes

- A workspace symbol beside the save status in the editor, project and phone
  shells, coloured by state: ready, being prepared, asleep, did not start.
- Its popover says since when and whether the assistant can be asked, in
  plain words.
- Implementers also see what it runs on, the start steps and the log, and
  can start it again (`POST …/workspace/restart`, others get 403); the
  preview column follows a restart.
