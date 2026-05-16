# Changelog

Notable changes to skills in this repo. Each skill is versioned independently via git tags of the form `<skill-name>/vX.Y.Z`.

The format is loosely based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

---

## three-layer-cake-haskell

### [0.1.0] - 2026-05-16

Initial release.

- `SKILL.md` — default track: pure typeclasses, `MonadError`, `tryIO`, `withDb`.
- `testing.md` — testing business logic without `IO` via `TestM`.
- `alternative-exceptions.md` — exception-based variant for concurrent code.
- `handles-upgrade.md` — adding handles for decoration and runtime swapping.
- `compensations.md` — DB transactions, `bracket`, sagas, and the outbox pattern.
- `starter-template.hs` — copy-pasteable skeleton.
