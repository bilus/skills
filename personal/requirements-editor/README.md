# requirements-editor

How an agent works with [requirements-editor](https://github.com/bilus/Scratchpad/tree/main/requirement-editor), one Go binary that serves a browser page for editing a repository's requirement YAML files and splices the changes back so a change to one requirement produces a diff touching only that requirement. The skill exists because an agent that does not know the tool edits the YAML by hand, and a requirements file with hundreds of `---` documents reflows into a diff nobody can review.

## What the agent does

- Opens the page for the user in the background, with `--mode review` after it has changed requirement files itself, and hands over the URL the server prints.
- Applies a `requirements-patch` document pasted from the page's *Copy patch* button: saved to a file, `--dry-run` first, then applied, then `git diff` and `REVIEW.md`.
- Writes a patch of its own instead of editing the YAML when asked to change a requirement, naming each one by `issue` or by `title`, never by line number.
- Reads `REVIEW.md`, where annotations typed into the page land, as the list of things a reviewer left for it.

## Why it is under personal/

It is tied to one binary installed through `devbox global` from a Nix flake in one source directory on one machine. The patch format and the workflow transfer; the paths do not.

## Installing

```bash
git clone https://github.com/bilus/skills.git ~/.claude/skills/bilus
```

Claude Code discovers `personal/requirements-editor/SKILL.md` inside the clone. The release workflow only packages top-level skills, so this one has no `.skill` asset.
