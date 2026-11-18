# Proposal

Bean: [chatcms-mizc](../../../.beans/chatcms-mizc--ai-gateway-on-the-host.md)
(AI gateway on the host), milestone 07
([chatcms-z0p7](../../../.beans/chatcms-z0p7--07-workspaces-micro-vms-and-chat.md)).

## Why

The chat agent runs inside a workspace, which runs code nobody reviewed. A
model provider's API key in there would be the installation's key in a
customer's hands. The host keeps the key and answers for it; a workspace gets
a token that only works through the host, only while the workspace runs, and
only within what its organisation may spend.

## What Changes

- **An OpenAI-compatible gateway in the server**, on a listener of its own
  (`CHATCMS_AI_LISTEN`, port 8091), answering `POST /v1/chat/completions` and
  `GET /v1/models`. It forwards to one provider (`CHATCMS_AI_PROVIDER_URL`,
  key from `CHATCMS_AI_KEY_FILE`, model `CHATCMS_AI_MODEL`), streaming
  included, with the provider's key in place of the workspace's token.
- **A token per workspace**, issued when it starts and revoked when it sleeps,
  fails or is replaced; stored hashed. The workspace gets the token and the
  gateway's address; a micro VM gets them on its kernel command line and
  reaches the gateway at its host address, which the firewall opens for that
  port only.
- **Limits per organisation**: requests a minute and tokens a month, set by
  the platform admin beside the existing limits, none by default. Over a
  limit, a request is refused with a plain reason.
- **Usage logged** per request: organisation, project, person, model, tokens
  in and out, status. The platform admin sees this month's totals per
  organisation.

## Capabilities

### New Capabilities

- `ai/gateway`: who may call the model through the host, with what, and what
  is passed on.
- `ai/limits`: how much an organisation may use, and what happens past it.
- `ai/usage`: what is recorded about each call and who sees it.

### Modified Capabilities

- `workspace/microvm`: a VM may reach the host's gateway port, and nothing
  else of the host.

## Impact

- New `internal/aigateway`; migration `00020_ai.sql`; queries.
- `internal/workspace`: `Spec.Token`, a `Tokens` setting the manager issues and
  revokes through; both backends pass the token and gateway address on.
- The guest writes the token and address to `/run/chatcms/ai.env` (root
  only). Nothing in this change reads it yet; the chat epic's OpenCode will.
- Admin API `GET`/`PUT /api/administration/organisations/{org}/ai` and an
  "AI use" section on the platform admin's organisation page.
- NixOS module: the gateway listener, the key file as a credential, port 8091
  open on `chv*` in both firewalls.
- Seams: `workspace.Backend` gains nothing; `policy` decides who sees usage.

## Assumptions

- **Own Go gateway, not LiteLLM**: one provider and one route is a few hundred
  lines with no new service to deploy; the briefing names this the default.
- **Provider and model are configuration.** The briefing's DeepSeek V4.1 Flash
  is the intended first one; any OpenAI-compatible endpoint works. No
  provider configured means no gateway, and the server says so once.
- **The model a workspace asks for is replaced by the configured one**, so a
  token cannot pick a more expensive model.
- **A budget is counted in tokens, not money**, per calendar month in UTC;
  prices differ per provider and change.
- **Rate limiting is per server process**, which is what one host runs.
- The OpenCode Go terms question stays open; nothing here depends on it.
