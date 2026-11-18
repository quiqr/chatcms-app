# Design

## Context

See proposal.md. Workspaces have a manager (`internal/workspace`) that starts
and stops them through a `Backend`; micro VMs reach only replies from the host
(nftables table `inet chatcms_workspaces`, NixOS firewall). Organisation
limits live on `organisations` and are set by the platform admin through
`PUT /api/administration/organisations/{org}/limits`.

## Decisions

### 1. `internal/aigateway`

`Gateway` is an `http.Handler`:

- `GET /v1/models` lists the configured model.
- `POST /v1/chat/completions`: check the caller's address (loopback or
  `10.231.0.0/16`), read `Authorization: Bearer <token>`, look the token up,
  check the organisation's limits, decode the body as a JSON object, set
  `model` to the configured one and, when `stream` is true, set
  `stream_options.include_usage`, then forward to `<provider>/chat/completions`
  with `Authorization: Bearer <key>`. A non-streamed answer is read and its
  `usage` recorded. A streamed answer is copied line by line, flushed after
  each, and the last `data:` chunk carrying `usage` is recorded.
- Refusals are OpenAI-shaped (`{"error":{"message","type","code"}}`), 401 for
  tokens, 429 for limits, 403 for an address, so an OpenAI client shows the
  message as is.
- Bodies above 1 MiB are refused; the provider call has a 5-minute timeout.

### 2. Tokens

`ai_tokens (token_hash BYTEA PRIMARY KEY, person_id, project_id,
organisation_id, issued_at, revoked_at)`. A token is 32 random bytes,
base64url, stored as its SHA-256. `Issue(key)` revokes any live token of that
(person, project) and stores a new one; `Revoke(key)` revokes it. The
manager's `Settings.Tokens` issues before the backend starts and revokes in
`stopOne`, after a failed start, and for every row reconcile puts to sleep.
`Spec.Token` carries it to the backend.

- local: nothing reads it yet; the chat epic passes it to OpenCode.
- microvm: `chatcms.token` and `chatcms.gateway=http://10.231.<i>.1:8091` on
  the kernel command line; the guest writes `/run/chatcms/ai.env` (mode 0600)
  and leaves the rest of the command line alone.

### 3. Limits and usage

`organisations` gains `ai_requests_per_minute INTEGER` and
`ai_tokens_per_month BIGINT`, null for none. The rate is counted in memory per
organisation over a sliding minute. The budget is
`SUM(tokens_in + tokens_out)` from `ai_usage` since the first of the month,
UTC, read before each request. `ai_usage (id, at, organisation_id, project_id,
person_id, model, tokens_in, tokens_out, status)`, indexed on
`(organisation_id, at)`.

Admin: `GET`/`PUT /api/administration/organisations/{org}/ai`, platform admin
only (`ActionAdministerInstallation`, as the existing limits), answering the
limits and this month's requests and tokens. The organisation page gets an
"AI use" section with two steppers like the existing limits.

### 4. Configuration and deployment

`CHATCMS_AI_LISTEN` (default `127.0.0.1:8091`), `CHATCMS_AI_PROVIDER_URL`,
`CHATCMS_AI_MODEL`, `CHATCMS_AI_KEY_FILE`. Without a provider URL and key the
gateway does not start and workspaces get no token. The NixOS module's `ai`
options pass the key as a systemd credential; with the microvm backend the
listener is `0.0.0.0:8091`, nftables accepts `tcp dport 8091` on `chv*`, and
`networking.firewall.interfaces."chv+".allowedTCPPorts` opens it in the NixOS
firewall too, because an accept in one nftables table does not undo a drop in
another.

## Risks / Trade-offs

- [A token on a kernel command line is readable inside the VM] → the VM is
  the token's holder; it dies with the workspace.
- [The budget check and the request are not atomic] → a few concurrent
  requests can overshoot by their own size; acceptable for a monthly budget.
- [Usage missing when a provider sends none] → recorded with zero tokens and
  the status, so the request count still shows.
