# Changelog

Notable changes to skills in this repo. Each skill is versioned independently via git tags of the form `<skill-name>/vX.Y.Z`.

The format is loosely based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

---

## otterwiki-search

### [0.1.0] - 2026-09-06

Initial release. Query procedure for the OtterWiki MCP tools under fusion-only retrieval with no reranker: two to four terms in the answer page's own vocabulary, stop at a rank-1 answer, one targeted follow-up (a snippet link, `wiki_list` by prefix or tag, or one rephrasing), never a vote-count merge. Measured at 18 of 18 on `evals/otterwiki-search/`. Lives under `personal/` because it is tied to one wiki and one server. Ships with a human-facing `README.md`.

## prose-contract

### [0.1.0] - 2026-08-31

Initial release. A style contract for agent-written prose: actor-and-action discipline, an outline-review-expand-review-draft-verify process with reviewer sub-agents at each gate, section A machine-checkable rules (banlist, at-most-once list, transition caps, typography), section B judgment rules with bad/good pairs, and a section C leak check against the skill's own examples. Ships with a human-facing `README.md`.

## tracer-bullet-delivery

### [0.1.0] - 2026-07-25

Initial release. Staged delivery: plan approval gate, walking skeleton as stage 1, one vertical slice per stage, 400 LOC hard cap, append-only ledger, halt for review at every boundary. Ships with a human-facing `README.md`.

## hole-driven-delivery

### [0.2.0] - 2026-09-25

The design now leads the code. Before the skeleton, the agent writes a vocabulary of the change's terms, reviewed by a sub-agent for redundancy, and a leveled data flow design in the dfd format: each box names its functions and says why the later steps need it, each arrow has a type, each change of state in place is a store write, and state stays inside the smallest process that uses it. The plan names a system metaphor and checks the rules it implies. Every review now records a score row from the bundled `tools/dfdmetrics` (shared state by level, live range, read distance, pressure, package spread, findings) and rebuilds a review page with the design, links into the code, term definitions and the score history. New reference files: `design.md` (the design rules), `review-page.md` (the page's content) and `prompts.md` (sub-agent prompts for the vocabulary review and the design review). A hole's ID is now its stage number with the function that holds it, and a fill that needs a new function adds it to the diagram first. The endgame keeps pushed history intact on repositories that keep all work on one shared branch, and reshapes other branches with `git read-tree`.

### [0.1.0] - 2026-07-25

Initial release. Staged delivery: plan approval gate, typechecked skeleton of named `HOLE(id)` stubs as the reviewable design, outside-in fill under a build-never-red invariant, grep census as progress bar, halt for review at every boundary. Ships with a human-facing `README.md`.

## test-ratchet-development

### [0.1.0] - 2026-07-25

Initial release. Staged delivery: plan approval gate, one failing acceptance test opens each stage, red-green-refactor with commit-per-green, append-only test diffs (test changes need human sign-off), halt for review at every boundary. Ships with a human-facing `README.md`.

All three share an identical outer loop and differ only in inner strategy, for A/B comparison — see `evals/staged-delivery/` for the standard task and rerun instructions. All three require `avoiding-ai-tells` as background.

## avoiding-ai-tells

### [0.1.0] - 2026-07-25

Initial release. Style rules for AI-generated code under human review: comment gradient instead of uniform narration, ASCII typography, no gold-plating or premature abstraction, behavior-named tests, human-shaped READMEs and commit history, no faked struggle, one-sentence disclosure of substantial AI assistance. Ships with a human-facing `README.md`.

## three-layer-cake-haskell

### [0.1.0] - 2026-05-16

Initial release.

- `SKILL.md` — default track: pure typeclasses, `MonadError`, `tryIO`, `withDb`.
- `testing.md` — testing business logic without `IO` via `TestM`.
- `alternative-exceptions.md` — exception-based variant for concurrent code.
- `handles-upgrade.md` — adding handles for decoration and runtime swapping.
- `compensations.md` — DB transactions, `bracket`, sagas, and the outbox pattern.
- `starter-template.hs` — copy-pasteable skeleton.
