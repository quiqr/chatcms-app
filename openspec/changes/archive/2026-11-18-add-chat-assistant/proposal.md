# Proposal

Bean: [chatcms-6jxg](../../../.beans/chatcms-6jxg--chat-with-opencode-in-the-workspace.md)
(chat with OpenCode in the workspace), milestone 07
([chatcms-z0p7](../../../.beans/chatcms-z0p7--07-workspaces-micro-vms-and-chat.md)).

## Why

The editor has kept a slot for an assistant since milestone 01. With a
workspace per person, a micro VM to run it in and a gateway that holds the
model's key, the assistant can now run: somebody asks for a change in words,
it appears in their preview, and it is saved like any other edit.

## What Changes

- **OpenCode runs in each workspace** as a headless server beside the
  preview: in the micro VM (from the guest's own closure) or as a process of
  the local backend. Its whole configuration is passed in; a site's own
  `opencode.json` is ignored. It talks to the model only through the AI
  gateway, with the workspace's token, and its API is protected by a password
  only the server knows.
- **The agent writes through tools on the host, not into files.** A Quiqr MCP
  server on the gateway's listener offers `list_pages`, `read_page` and
  `update_page`, which go through the same model, validation and saving as
  the forms do. OpenCode's own file editing, shell and web tools are denied;
  it may read and search the site. The VM's share of the site stays
  read-only.
- **A chat panel in the editor's assistant slot.** A message goes to the
  person's OpenCode session; the reply streams back as it is written; each
  page the assistant changed is listed with Show and Undo. Undo goes back to
  the version before the change without publishing.
- `CHATCMS_OPENCODE_BIN` gives the local backend an OpenCode; without it, or
  without an AI gateway, the panel says the assistant is not available.

## Capabilities

### New Capabilities

- `assistant/agent`: the agent in a workspace, what it may reach and what it
  is told.
- `assistant/tools`: what the agent may do to a site, and how.
- `assistant/chat`: talking to the agent from the editor, and what is shown.

### Modified Capabilities

- `workspace/microvm`: a VM also runs the agent, reachable by the host only.

## Impact

- New `internal/assistant` (MCP tools, chat sessions, event relay, web).
- `internal/workspace`: `Endpoint.AgentPort`, `Workspaces.Agent(key)`; both
  backends start OpenCode when they have one.
- `nix/workspace-guest.nix`: an agent unit; OpenCode in the guest's closure.
- `internal/httpapi`: `POST /api/organisations/{org}/projects/{project}/chat`
  and `GET …/chat/events` (server-sent events).
- Frontend: the assistant slot opens a chat panel; strings in both languages.
- e2e: a scripted model behind the gateway; `test/e2e/35-asking-the-assistant`.
- Acceptance (milestone 07, epic 6): "change X on page Y" through chat changes
  the file, shows in the preview and is saved on the person's copy.

## Assumptions

- **Edits only through the tools.** The briefing mentions direct Markdown
  edits; a writable share would let the VM write into a git directory the
  host then runs git in. `update_page` takes the page's text as well as its
  fields, which covers what a direct edit would.
- **One session per person per project**, held in memory; a restart starts a
  new conversation. History of conversations is not kept.
- **"Let the AI combine conflicting changes" (chatcms-izhr) is left to the
  merging epic**, which owns the question it would answer.
- OpenCode's model list is not fetched and its updates are off; it runs fully
  without internet, which the VM requires.
