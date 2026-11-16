# Releasing ChatCMS

A release is a version number, a dated changelog section, a tag and a GitHub
release. One command makes all four.

## The version

`VERSION` in the repository root is the one place it is written, as a bare
semver. `flake.nix` reads it for the package version and stamps it into the
binary, and a plain `go build` embeds the same file, so every build reports
the same thing. `chatcmsd --version` and `/api/health` both report it, so a
running machine can be asked what it is rather than inferred from a
deployment log.

## Cutting a release

Start from a working copy that is clean and on `origin/main`, which is where
`scripts/ship-change.sh` leaves it. Look first, then release:

```bash
scripts/release.sh minor --dry-run
scripts/release.sh minor
```

The argument is the part of the version that goes up:

| Argument | From `0.4.2` | Use it for                                     |
| -------- | ------------ | ---------------------------------------------- |
| `patch`  | `0.4.3`      | fixes only                                     |
| `minor`  | `0.5.0`      | new features; before 1.0, also breaking ones   |
| `major`  | `1.0.0`      | breaking changes, once there is a 1.0          |

The script refuses, before it changes anything, when the working copy has
changes, when it is not on `origin/main`, when `[Unreleased]` in
`CHANGELOG.md` is empty, when the tag already exists, or when `gh` is not
signed in. Then it:

1. writes the new version to `VERSION`;
2. renames `[Unreleased]` in `CHANGELOG.md` to the new version with today's
   date, and puts an empty `[Unreleased]` above it;
3. runs `nix flake check`, and puts both files back when it fails;
4. commits the two files as "Release x.y.z" and moves `main` to it;
5. tags the commit `vx.y.z` and pushes `main` and the tag;
6. creates the GitHub release, with the changelog section as its notes.

## When the GitHub release fails

The tag is already pushed by then, and the tag is what a machine pins, so the
release exists. The script keeps the notes in `release-notes-x.y.z.md` and
prints the `gh release create` command to run again. Delete the file after.

## Pinning a release

A machine that deploys ChatCMS takes the flake at a tag:

```nix
inputs.chatcms.url = "github:quiqr/chatcms-app/v0.1.0";
```

No binaries are attached to a release. The server only carries its frontend
when Nix builds it, and Nix is how ChatCMS is deployed.

## What goes in the changelog

What changed for somebody using it, not what changed in the code. An entry
nobody outside this repository could act on belongs in the OpenSpec archive,
which is where the reasoning already is.

Every archived change in `openspec/changes/archive/` is a release note waiting
to be written: its proposal says why, in the words the change was argued in.

## Versions before 1.0

The schema may change in either direction and no promise is made about
upgrading between pre-release versions other than by migration, which runs
automatically. Once a customer is on it, that stops being true, and this
section is rewritten rather than quietly ignored.
