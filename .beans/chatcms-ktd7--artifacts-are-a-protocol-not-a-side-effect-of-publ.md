---
# chatcms-ktd7
title: Artifacts are a protocol, not a side effect of publishing
status: todo
type: feature
priority: normal
created_at: 2026-11-20T15:03:56Z
updated_at: 2026-11-20T15:03:56Z
---

What an artifact is, where it lives and how long, said plainly enough that
other parts of ChatCMS can lean on it. Today keeping is only possible through
a publish, nothing is ever removed, and the files sit beside the database
outside any backup.

What already holds and must keep holding: a kept file is copied out of the
build workspace into `<state>/artifacts/<project>/<id>` with its name, size,
sha256 and time in Postgres, so it is downloadable after the workspace and the
VM are gone, and by anybody who may read the project.

Gaps to close:

1. **Keep on request.** Keeping is reachable from a publish only. A render on
   request (bean: a folder of documents rendered one at a time) should be able
   to keep its result without publishing the site, if that bean decides it
   should count as an artifact.
2. **Deduplicate by sha256.** Publishing twice keeps two copies of the same
   150 MB file. One file per content, referenced by more than one row.
3. **Remove.** A file outlives its row; deleting a project leaves its
   directory behind. Removal on deleting a project and a retention rule for
   files no row points at, with a dry run first.
4. **Say what it is.** A kind and a label, so "Offerte Van Dijk" is
   distinguishable from `book.pdf`, without inventing a second filing system.
5. **Say what a backup must cover.** The document that describes deployment
   names the artifact directory beside the database.

Left open on purpose: whether artifacts get their own page for deleting them,
and whether a publish may prune what it replaces.
