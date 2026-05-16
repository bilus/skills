# three-layer-cake-haskell

A [Claude skill](https://www.anthropic.com/news/skills) for the three-layer cake architecture pattern in Haskell: capability typeclasses, `MonadError` for domain errors, IO exceptions for infrastructure errors caught at the boundary.

Part of the [bilus/skills](https://github.com/bilus/skills) collection.

## Install

### Claude Code

```bash
mkdir -p ~/.claude/skills
curl -L https://github.com/bilus/skills/archive/refs/heads/main.tar.gz | \
  tar xz --strip-components=1 -C ~/.claude/skills \
  skills-main/three-layer-cake-haskell
```

Or clone the whole collection and let Claude Code discover everything:

```bash
git clone https://github.com/bilus/skills.git ~/.claude/skills/bilus
```

### Claude.ai

Download `three-layer-cake-haskell.skill` from a release and upload via **Settings → Capabilities → Skills**. The latest release with a `three-layer-cake-haskell/v*` tag will have it as an asset.

## What's here

| File | Purpose |
|------|---------|
| `SKILL.md` | Main file — default track: pure typeclasses, `MonadError`, `tryIO`, `withDb` |
| `testing.md` | Testing business logic without `IO` |
| `alternative-exceptions.md` | Exception-based variant for code with heavy `async` composition |
| `handles-upgrade.md` | Adding handles for decoration and runtime swapping |
| `compensations.md` | Undoing side effects: transactions, `bracket`, sagas, outbox |
| `starter-template.hs` | Copy-pasteable Haskell skeleton |

`SKILL.md` is self-contained; the other files are referenced from it with "read this when…" pointers, so Claude only loads them when relevant.

## When the skill triggers

Claude consults this skill when:

- You **explicitly ask** for the three-layer cake, the ReaderT pattern, or capability typeclasses
- You discuss how to **structure, organize, or architect** a non-trivial Haskell application
- You ask about **testing business logic**, **rolling back side effects**, or **sagas**

Claude will **not** push the pattern on:

- One-off scripts, CLI utilities, prototypes
- Single-file demos or learning exercises

If you ask for it by name in a context where it's arguably overkill, Claude applies it without protest.

## Acknowledgments

The pattern is informed by Matt Parsons' [The Three Layer Haskell Cake](https://www.parsonsmatt.org/2018/03/22/three_layer_haskell_cake.html) and [The ReaderT Design Pattern](https://www.fpcomplete.com/blog/2017/06/readert-design-pattern/) by Michael Snoyman. This skill is a synthesis of those ideas adapted for AI-assisted development.
