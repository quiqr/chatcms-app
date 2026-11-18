---
# chatcms-pu3a
title: OpenSpec guardrails per project
status: draft
type: epic
priority: deferred
created_at: 2026-11-18T12:51:55Z
updated_at: 2026-11-18T12:51:55Z
parent: chatcms-2udf
---

## Intent

Optional per project, off by default. An `openspec/` directory in a site repo
already works with OpenCode; this integrates it deeper:

1. A guardrails screen: specs as readable rules, editable by implementers,
   read-only for editors.
2. Pre-publish checks: a workflow step validates changes against the specs
   (deterministic checks plus an AI reviewer), shown as friendly warnings.
3. Larger changes become an OpenSpec change proposal an organisation admin
   approves; small edits apply directly.
4. Templates can ship default specs.

The agent always loads the project's specs before editing.

## Room left for it

Needs the chat epic of milestone 07 (chatcms-6jxg). See
`docs/BRIEFING-M07.md`, "Future module: OpenSpec guardrails".
