# Design

## Context

See proposal.md. A workspace has a manager, two backends and a token issued
per start; the AI gateway listens apart on port 8091 and is the one host port
a VM may reach. OpenCode 1.15 (nixpkgs) serves an HTTP API: `POST /session`,
`POST /session/{id}/prompt_async`, `GET /event` (server-sent events:
`message.part.delta`, `session.idle`, `session.error`), protected by
`OPENCODE_SERVER_PASSWORD` (basic auth). Tested: with
`OPENCODE_DISABLE_MODELS_FETCH`, autoupdate off and its bundled
OpenAI-compatible provider, it runs with no network at all.

## Decisions

### 1. Configuration passed whole, the site's ignored

Both backends start `opencode serve` with
`OPENCODE_CONFIG_CONTENT` (provider `chatcms` pointing at the gateway with
the token as key, model `chatcms/chatcms`, MCP server `quiqr` at the
gateway's `/mcp` with the token as bearer, permissions `edit`, `bash`,
`webfetch` denied, `autoupdate` false, `share` disabled),
`OPENCODE_DISABLE_PROJECT_CONFIG=1`, `OPENCODE_DISABLE_CLAUDE_CODE=1`,
`OPENCODE_DISABLE_EXTERNAL_SKILLS=1`, `OPENCODE_DISABLE_MODELS_FETCH=1`,
`OPENCODE_DISABLE_AUTOUPDATE=1`, `OPENCODE_SERVER_PASSWORD=<token>`, and a
`HOME` and XDG directories of its own. `assistant.OpenCodeConfig(gateway,
token)` builds the JSON; the guest builds the same from its command line.

- local: a second process beside Hugo, on 127.0.0.1 and a free port, in the
  person's copy; its handle gains the pid and port.
- microvm: the guest's `chatcms-agent` unit runs OpenCode from the guest's
  own closure as uid 65534 on port 4096 with `/site` (read-only) as its
  directory, when `/run/chatcms/ai.env` exists.

`workspace.Endpoint` gains `AgentPort`; `Workspaces.Agent(key)` answers the
agent's address and secret (the token) of a warm workspace.

### 2. The Quiqr MCP server on the gateway's listener

`/mcp` on the gateway's listener speaks MCP over streamable HTTP with JSON
answers: `initialize`, `notifications/initialized`, `ping`, `tools/list`,
`tools/call`. The bearer token is looked up like the gateway's; the person's
`policy.Actor` comes from tenancy. Tools:

- `list_pages`: the content sections with each page's key and title.
- `read_page {key}`: the page's fields (path, label, value) and text.
- `update_page {key, fields?, text?}`: `content.Service.Save` with those
  values. Before and after, the person's copy's latest version is read; a
  change is reported to the chat service as (page, title, version before,
  version after).

A refusal (unknown field, not allowed) is a tool result with `isError`, in
plain words, so the model can correct itself.

### 3. Chat sessions and the event relay

`assistant.Chat` keeps one OpenCode session per (person, project) in memory.
`Send` makes sure the workspace is warm, creates the session on first use and
posts `prompt_async`. `Events` opens OpenCode's `/event`, keeps what belongs to
the session and writes server-sent events to the browser: `text` (a delta of
the reply), `change` (from the tools), `done` (`session.idle`) and `error`.
Changes are fanned out from the tools to every open stream of that person and
project.

### 4. The panel

The editor's assistant slot and the Assistant button open `ChatPanel`: the
conversation, a text box, the reply growing as `text` events arrive, and a
chip per changed page with Show (opens the page in the editor) and Undo
(`POST versions/{before}/go-back` with `publishRightAway: false`). Without an
assistant (`GET …/chat` says `available: false`) it says so.

### 5. Testing

The e2e harness starts a scripted OpenAI-compatible model: asked with tools
for the first time it answers with an `update_page` call; asked again (with
the tool's result) it answers in words. ChatCMS runs the local backend with
`CHATCMS_OPENCODE_BIN`, so the real OpenCode, the real MCP server and the real
gateway are all in the loop.

## Risks / Trade-offs

- [OpenCode's API changes between versions] → pinned by nixpkgs; the relay
  uses five calls and three event types, each covered by the e2e test.
- [A session per person in memory] → lost on restart; acceptable for the PoC.
- [The local backend runs OpenCode on the host] → its tools are denied
  writing and commands, and the site's configuration is ignored; the local
  backend is for development.
