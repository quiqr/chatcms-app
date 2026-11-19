# Proposal

Bean: [chatcms-izhr](../../../.beans/chatcms-izhr--let-the-ai-combine-conflicting-changes.md)
(let the AI combine conflicting changes), in epic
[chatcms-9iuz](../../../.beans/chatcms-9iuz--changes-from-everyone-come-together.md)
(changes from everyone come together), milestone 07
([chatcms-z0p7](../../../.beans/chatcms-z0p7--07-workspaces-micro-vms-and-chat.md)).

## Why

When two people change the same field, the choice is one version or the
other. Often both changes are wanted: one fixed a typo, the other added a
sentence. The assistant can propose a text that keeps both, and the person
decides.

## What Changes

- **Let the AI combine them**: a third button on each field question. The
  combined text appears in a box the person can change, with "Use the
  combination"; keep mine and use theirs stay. It is offered only where the
  assistant is available.
- **A combined answer**: an answer may give a text for a field instead of
  choosing a version, and that text is written.
- **The gateway answers the host once**, for a request that is not a
  workspace's: the same model, key, headers, limits and usage records.
- **Fixed on the way: a copy followed a rewritten shared version.** When an
  answer replays waiting work on top of what arrived, the shared version's
  history changes; a person's copy that held the earlier history looked as
  if it had work of its own and never moved forward again. A copy now
  remembers the version it last took, and when it holds nothing since, it
  takes the shared version as it is.

## Capabilities

### New Capabilities

### Modified Capabilities

- `sync/settling`: combining a field with the AI.
- `sync/personal-copies`: a copy with nothing of its own follows the shared
  version.
- `ai/gateway`: a single answer for the host.

## Impact

- `internal/aigateway`: `Ask`.
- `internal/sitesync`: `Answer.Combined`, `Combine`, a `Combiner`.
- `internal/app`: the combiner's prompt; `internal/httpapi`:
  `POST …/standing/combine`.
- `internal/gitsync`: `refs/chatcms/taken` in a person's copy.
- Frontend: the sync screen's third choice; strings.
- e2e `39-letting-the-ai-combine`, scripted and with `E2E_REAL_AI`.

## Assumptions

- **Combining is for one field at a time**, never a whole page: a person
  reads what they accept.
- **The combination is not saved until the answers are**, so asking costs
  nothing but the call; asking again gives a new proposal.
- **The prompt asks for the text only**, prefers the second version's facts
  where they contradict, and keeps the form of the field. A fenced answer is
  unwrapped.
- **yr88** (rebasing each person's work in the background) was delivered by
  the per-person workspace change: saves are shared in the background, by
  merging. **Soft page locking**, optional in the epic, is left out and
  filed on its own.
