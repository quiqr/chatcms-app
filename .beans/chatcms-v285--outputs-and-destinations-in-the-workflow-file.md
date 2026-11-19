---
# chatcms-v285
title: Outputs and destinations in the workflow file
status: completed
type: feature
priority: normal
created_at: 2026-11-18T12:51:49Z
updated_at: 2026-11-19T07:19:37Z
parent: chatcms-q9qu
blocked_by:
    - chatcms-7wcg
---

The workflow file declares a build's outputs and, for each, a destination:
artifact, publish or both. The publish flow runs the build through the Runner
and hands each output to its destination.
