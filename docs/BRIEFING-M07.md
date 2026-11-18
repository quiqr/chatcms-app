---
status:
    - agent-draft
title: 'Claude Code briefing Quiqr ChatCMS: workspaces, micro VMs and chat'
onderwerp: quiqr
project: chatcms
---
# Claude Code briefing — Quiqr ChatCMS: workspaces, micro VMs & chat

Repo: github.com/quiqr/chatcms-app. Follow the usual 1shotpoc workflow and gates; spec every step with OpenSpec. Read docs/ARCHITECTURE.md and docs/BRIEFING-M01.md first.

## Goal
A working proof of concept: a user opens a project, gets their own workspace (micro VM) that runs a Hugo **or** Quarto preview plus an OpenCode chat agent. Edits made through chat show up live in the preview.

## Current state (from code analysis)
- Builds/checks go through `internal/runner.Runner` (`Run(ctx, Task)`, fs.FS in/out), implemented by `runner.Sandbox` (bubblewrap + prlimit). This is the intended swap point for micro VMs.
- Preview (`internal/preview`) is hard-coded: `exec hugo server` per **project**, reverse proxy at /preview/{org}/{project}/, idle sweep (15 min), cap `PreviewsAtOnce` (8), TCP dial readiness, live reload proxied.
- Workflows (`internal/workflows`): only OnPublish is wired up; Publish runs unsandboxed (deploy key + network).
- No chat/agent runtime exists yet.

## Core concept: Workspace
One workspace per **user per project**. It holds:
1. the user's working copy on their own branch,
2. the preview server (Hugo/Quarto),
3. the AI agent (OpenCode).

The backend asks for "workspace(user, project)" and gets back an address plus a lifecycle (start → warm → sleep after idle → cleanup). Where it runs is hidden behind the interface. Builds stay separate, as one-shot jobs through Runner.

## Steps (each independently testable)

### Step 1 — Workspace interface, local backend
- Introduce a Workspace abstraction; refactor `internal/preview` to sit behind it.
- Key on (user, project) instead of project; each user gets their own working copy/branch.
- Keep existing behaviour: proxy, live reload, idle sweep, cap (now workspace limits).
- Acceptance: two users preview the same project at the same time with different changes.

### Step 2 — Micro VM backend
- microvm.nix on **Cloud Hypervisor**; host /nix/store shared read-only via virtio-fs.
- Networking via TAP with a private IP per VM (no vsock for now). The VM can reach only the host (git remote, AI gateway).
- Working copy: the host keeps a bare repo per project; the VM clones it, works on the user's branch and pushes back to the host. Deploy keys never enter the VM; the host does external publishing.
- The proxy maps workspace → VM IP:port; readiness via TCP dial; live-reload websocket proxied.
- Hugo version per project: pinned in the project's quiqr config; the host realises that version (via Pim's "Very Nix" tool) before VM start; the VM runs it from the shared store.
- Acceptance: preview runs inside a VM with the project's pinned Hugo version.

### Step 3 — Chat
- OpenCode (from nixpkgs) runs in headless server mode inside the workspace VM, working in the working copy.
- AI gateway on the host: OpenAI-compatible endpoint holding the real API key (start with the DeepSeek V4.1 Flash API); VMs get only a per-workspace token. Per-organisation rate/budget limits and usage logging. Default: a small Go proxy in the backend (LiteLLM is an alternative).
- OpenCode is configured with the gateway as a custom provider and gets Quiqr MCP tools so model/forms stay in sync (combination of MCP and direct Markdown edits).
- Backend ↔ OpenCode via its HTTP API; stream replies to a basic chat panel in the React frontend.
- Acceptance: "change X on page Y" through chat → file changed → live in preview → committed on the user's branch.

### Step 4 — Quarto runtime (PDF)
- Simple declarative runtime definition in the site repo (in the quiqr directory): packages + versions + preview/build commands. Simple mode only (nixpkgs packages); full Nix flakes later.
- Quarto + Typst as the PDF engine (TeX Live optional).
- Preview: `quarto preview` in the workspace, PDF shown via the proxy. Final PDF: one-shot build through Runner → artifacts system.
- Acceptance: a Quarto project renders a PDF preview and produces a PDF artifact.

## Collaboration / merging (defaults for the PoC)
- Users must never see merge conflicts; publishing must be idiot proof.
- Small auto-commits per user branch, background rebase onto main.
- Optional page-level soft locking (project setting, on/off).
- On a real conflict: plain choice (keep mine / theirs / let AI combine), never git terminology.

## Out of scope (future modules)
- vsock networking
- Remote micro VM hosts (host agent over Nebula, shared binary cache)
- Customer self-hosted runners (outbound-connecting, untrusted, org-scoped tokens)
- Advanced runtime definitions (custom Nix flakes, evaluated in an isolated builder VM)
- Firecracker Runner for builds (the Runner interface is ready for it)

## Open decisions (pick defaults, flag in the spec)
- Exact format of the runtime definition file
- Own Go gateway vs LiteLLM
- Final merge strategy
- Check the OpenCode Go terms before using it for customers; the PoC uses the pay-as-you-go API



## Addendum — backends, lifecycle, startup UX, hosting

### Workspace backends
The Workspace interface must support several backends:
- **local** (processes / bubblewrap): dev and fallback
- **gVisor containers** (runsc, Apache 2.0, in nixpkgs): candidate first production backend. Runs on any VPS without KVM, is cheap and easy to scale. Nix store as a read-only bind mount, normal container networking. Caveats: slower file I/O on big Hugo builds; weaker isolation than a VM.
- **micro VM** (microvm.nix + Cloud Hypervisor): strongest isolation, needs KVM (bare metal or nested virtualization).
Design so that switching backends needs no changes outside the workspace package.

### Lifecycle (owned by ChatCMS)
- The backend owns all lifecycle decisions (start, warm, sleep after idle, cleanup); state is tracked in Postgres.
- On server startup: reconcile, i.e. reattach to running workspaces or clean them up (same idea as builds' InterruptLeftovers).
- Start the workspace eagerly when the user opens a project, so the preview is usually ready before it's requested.
- Workspace limits replace PreviewsAtOnce; idle sweep carries over.

### Startup UX
- Normal users: a simple "preparing preview" progress indicator only.
- Advanced users / implementers: an expandable view with the startup steps (starting workspace, fetching site, building preview) and logs.
- Target: about 1–3 s cold start for a typical site; instant when warm.

### Resource sizing (estimate)
- Per active workspace: about 1 GB RAM and 1 vCPU, mostly idle (VM overhead ~50–100 MB, Hugo 200–500 MB, OpenCode 300–500 MB, Quarto 1–2 GB while rendering).
- 10 companies / 20 users: one host with 16–32 GB RAM and 8 cores is enough.

### Future: autoscaling
An always-on base host plus hourly-billed burst hosts registered as workspace hosts (builds on the remote-hosts module). Out of scope for now.



## Addendum — Quarto support

- Support all Quarto project types (website, book, manuscript, presentation).
- **Outputs and destinations:** a build produces outputs; each output has a destination, declared in the workflow file in the quiqr directory:
  - artifact: stored in ChatCMS, versioned and downloadable (e.g. PDF, DOCX)
  - publish: pushed to a hosting endpoint (e.g. website to Cloudflare)
  - both (e.g. a Quarto book: website published + PDF artifact)
  Some projects have no artifacts; show the artifacts panel only when a project defines them.
- **Code cells (Python/R) are allowed** during rendering. They run only inside the workspace/build sandbox, with no network and no secrets.
- Python/R dependencies are pinned in the runtime definition through Nix, for reproducible renders.
- **Responsibility split:** preview and build scripts/commands are the site developer's responsibility (in their runtime definition), not ChatCMS's. ChatCMS provides the runtime, sandbox, lifecycle, proxy and artifact/publish handling, and runs whatever commands the developer declares. Recommend Quarto's freeze in docs/templates, but don't enforce it.



## Addendum — Generic preview pane & select-to-edit

### Generic preview pane
- One preview pane for all runtimes. The preview is always a URL behind the ChatCMS proxy, shown in an iframe.
- The runtime definition declares the preview type (web, pdf, other); the pane picks the matching viewer:
  - web: plain iframe with live reload
  - pdf: the browser's built-in viewer first; PDF.js later (keeps the page position on refresh, allows text selection)
  - other (e.g. DOCX): download button, or conversion to PDF for the preview
- Serve previews from a separate origin than the app: site JavaScript must never get access to the ChatCMS app.

### Select-to-edit
1. The proxy injects a small script into HTML previews. On text selection it shows a floating icon and sends the selected text plus surrounding context to the parent via postMessage (origin-checked).
2. Clicking the icon opens an inline prompt, or quotes the selection into the chat panel; the user types the request.
3. Map preview URL → content file using Hugo's `hugo list all` (path ↔ permalink). Fallback: the agent searches the repo (text from templates/data files).
4. OpenCode gets the file, the selection, the context and the request; it edits the source; the preview reloads; the changed text is briefly highlighted.
5. A short-lived "undo / view change" option. Every edit is a commit on the user's branch.
- PDF (Quarto): selection via PDF.js; the agent searches the chapter sources (fuzzier matching).



## Future module — OpenSpec guardrails (optional)

Optional per project (off by default; implementers enable it per project or through a template). An `openspec/` directory in a site or document repo already works with OpenCode as-is; this module integrates it deeper:
1. **Guardrails screen:** specs shown as readable rules; implementers edit them, editors get read-only access.
2. **Pre-publish checks:** a workflow check step validates changes against the specs (deterministic checks plus an AI reviewer); problems are shown as friendly warnings (e.g. "this page no longer has a call to action").
3. **Proposals for big changes:** small edits apply directly; larger ones (e.g. restructuring the menu) become an OpenSpec change proposal that an organisation admin approves.
4. **Template defaults:** templates can ship default specs.
The agent always loads the project's specs before editing.

