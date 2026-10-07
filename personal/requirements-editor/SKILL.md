---
name: requirements-editor
description: Use when editing, reviewing or annotating requirement YAML files (REQUIREMENTS.yaml, EARS requirements) in a repository, when asked to open the requirements editor or review requirements in a browser, or when handed a "requirements-patch" document to apply.
---

# requirements-editor

One Go binary, on PATH through devbox global. It serves a browser page for
editing a repository's requirement YAML files and applies the changes back by
splicing byte ranges. **A change to one requirement produces a diff touching
only that requirement.** Hand-editing these files, or round-tripping them
through a YAML library, reflows every document and produces a diff nobody can
review. Use the tool.

Source and design notes: `~/dev/bilus/Scratchpad/requirement-editor/README.md`.
Flags: `requirements-editor --help`.

## Quick reference

| Task | Command (from the repository root) |
| :--- | :--- |
| Open the editor for the user | `requirements-editor` |
| Review only what the working tree changed, diff view on | `requirements-editor --mode review` |
| One file instead of searching the tree | `requirements-editor --file delivery/ads/REQUIREMENTS.yaml` |
| Validate a patch, write nothing | `requirements-editor --apply-clipboard --dry-run < patch.yaml` |
| Apply a patch | `requirements-editor --apply-clipboard < patch.yaml` |
| Apply a patch that deletes a requirement | add `--allow-remove` |
| Throwaway repo to try it on | `bash ~/dev/bilus/Scratchpad/requirement-editor/testdata/seed-demo.sh /tmp/demo` |

Flags only, no positional arguments. `--root` defaults to `.`.

## Opening it for the user

Run it in the background: it serves until the page stops sending heartbeats
(30 minutes idle by default), so a foreground call blocks the session. It
picks a free port, mints a token for the run, and prints the URL carrying
that token; give the user that line. It opens a browser unless `--open=false`;
pass that when there is none (a subagent, a remote shell). `--addr` accepts
loopback only.

After you have edited requirement files yourself, `--mode review` is the way
to hand the result to the user: it loads just the changed requirements with
the diff showing.

## Applying a patch

The page's *Copy patch* button puts a `requirements-patch` document on the
clipboard. When one is pasted to you:

1. Save it verbatim to a file. The backtick fence is fine, the parser strips it.
2. `--apply-clipboard --dry-run < patch.yaml` and read the report. Despite the
   name, the flag reads stdin, not the clipboard.
3. Apply. Then read `git diff` and `REVIEW.md`.

A `set` whose value already matches is reported as skipped, and so is
`field: null` on a field the document never had.

The whole patch is validated before any byte is written; a refusal writes
nothing. Applying the same patch twice changes nothing the second time.

## Writing a patch yourself

When you want to change a requirement, write a patch and apply it rather than
editing the YAML. The format:

```yaml
requirements-patch: 1
changes:
  - file: delivery/ads/REQUIREMENTS.yaml
    issue: https://github.com/iBiquity/conrad/issues/21065   # identity: issue when it has one
    set:
      requirement: |
        When a receiver provides a UID2 token in an ad request, the AutoStage
        API shall pass this value to the ad provider.
      labels: [question, v2]
      notes: null                                            # null removes the field
  - file: delivery/config/REQUIREMENTS.yaml
    add:                                                     # a new draft; no issue yet
      title: "Config API: Require an AutoStage token to write consent"
      requirement: |
        When a consent write is received, the Config API shall require an
        AutoStage token naming the consent API as its audience.
      scope: iBiquity/conrad
      status: draft
  - file: delivery/ads/REQUIREMENTS.yaml
    title: "Preload Ads API: Daypart enforcement"            # identity: title when there is no issue
    remove: true                                             # needs --allow-remove
review:
  - path: delivery/ads/REQUIREMENTS.yaml
    line: 3820
    annotation: Split this: the Maryland rule and the consent rule are different things.
```

A requirement is named by `issue`, or by `title` when it has none. Never by
line number. A title matching two documents is refused, not guessed.

Fields the tool knows: `issue`, `parent`, `old_req_id`, `title`, `requirement`,
`scope`, `labels`, `test`, `value`, `notes`, `status`. Anything else in a
document is carried through untouched. A new document gets its fields in that
order.

## REVIEW.md

Annotations typed into the page never enter the YAML. They travel in the
patch's `review` list and land in `REVIEW.md` at the repository root, one line
per note, under a heading for the file's git state (committed, staged,
unstaged):

```
delivery/ads/REQUIREMENTS.yaml:3820: [31af574a] Split this: the Maryland rule and the consent rule are different things.
```

A patch with only a `review` list and no `changes` is the "leave this for an
agent" case. When asked to work through a review, read `REVIEW.md`, act on each
line, and delete the lines you have handled.

## Common mistakes

- **Editing the YAML by hand** to make a small change. The file reflows; the
  diff is unreviewable. Write a patch.
- **Running it in the foreground** for the user. It does not exit until idle.
- **Running from the wrong directory.** `--root` is where it searches and where
  `REVIEW.md` goes. Use the repository root.
- **Expecting `remove` to work** without `--allow-remove`.
- **Looking for Makefile targets, issue creation or docx generation.** Out of
  scope; the repository's existing requirements tooling does those.

## Rebuilding after changing the source

The binary is a Nix flake in the source directory, installed with
`devbox global add path:$HOME/dev/bilus/Scratchpad/requirement-editor#default`.
A `path:` flake is not pinned in `devbox.lock`, so after changing the source
`devbox global install` rebuilds it. If `go.mod` changed, the `vendorHash` in
`flake.nix` must change too: set it to `pkgs.lib.fakeHash`, build, and copy
the hash from the mismatch error.
