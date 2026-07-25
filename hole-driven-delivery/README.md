# Hole-Driven Delivery — README

A staged-delivery workflow for AI coding agents, adapted from typed-hole programming (Haskell/Idris/Agda; `fatalError()` in Swift, `todo!()` in Rust).

Not to be confused with [jhhuh's hole-driven-delivery skill](https://github.com/jhhuh/hole-driven-delivery-skill), which is the complementary inner technique: an autonomous compiler loop (write holes, compile, read diagnostics, fill the most constrained hole, repeat) with no plan, git, tests, or review gates. This skill is the delivery process around that idea — plan approval, the skeleton as a reviewed commit, tests with every fill, and a halt at every stage boundary. The name says the difference: theirs is hole-driven development, this is hole-driven delivery. The agent plans first, then commits a complete *compiling skeleton* of the change — every type, signature, and module wired, every body a named `HOLE(id): contract` stub — and then fills holes one at a time, outside-in, keeping the build green after every single move. It halts for your review at every stage boundary.

## Why this approach

Review capacity is the bottleneck with agents, and the hardest thing to review late is architecture. Hole-driven moves the architecture review to the front and makes it cheap: the skeleton commit is a few hundred lines of pure declarations — names, types, seams, dependency direction — reviewed before any implementation exists to bias the discussion. After that, the compiler becomes the agent's oracle (each hole's type and call sites dictate what the body must be, so the agent reads the checker instead of inventing from memory), and `grep -rn "HOLE("` becomes a machine-checkable progress bar that survives any loss of agent context. Fills stay small because the only legal moves are "fill one hole" or "refine one hole into smaller named holes."

## What the agent will do

1. Write `docs/plans/<date>-<feature>.md`: numbered requirements, stages mapped to hole IDs, acceptance criteria, size estimates. **Stops for your approval.**
2. Branch, then commit the skeleton: full shape, zero logic, every body a tagged hole that typechecks and fails loudly if executed. Existing suite stays green. **Stops for review — this is the design review.** Approved signatures are frozen; changing one later is its own stage.
3. Each later stage fills one hole-cluster: one hole open at a time, build never red, each fill landing with the tests that pin its contract, one commit per hole. Newly discovered work becomes new named holes, never inline implementation. 100–300 LOC target / 400 hard cap per stage.
4. Each handoff quotes the fresh full-suite run and the hole census (filled / added / remaining). Done = zero grep matches + green suite. An append-only ledger next to the plan records every hole event and decision.
5. After your final sign-off, the agent squashes the branch to one commit per stage, each with an imperative subject and a one-or-two-sentence first- or third-person body describing the change itself — never the methodology (no "hole"/"fill"/"stage"/"ledger" vocabulary in messages). The history that lands reads like a human-maintained repo, without fabricating anything (real history collapsed, never invented).

## What you do

- Approve the plan, then review the skeleton hardest — it's the whole design in one small diff of declarations. Everything after is bodies for shapes you already approved.
- At each halt, answer the handoff's single "review focus" question and sanity-check the hole census delta.
- Approve any proposed signature change or existing-test change — the agent may not make these unilaterally.
- In dynamic languages, expect full type annotations plus a typechecker (mypy/pyright/tsc) in the gate; that's what makes the skeleton checkable.

## Usage

Save the skill (or drop the folder into your skills directory). Enable **only one** staged-delivery variant at a time — they have overlapping triggers and are meant to be A/B'd, not stacked. It triggers on any multi-edit coding task; you can also invoke it explicitly. Works best in typed languages; workable in dynamic ones with a typechecker in CI.

## Comparing against the other variants

Track per feature: number of stages, mean stage LOC, your review minutes per stage, defects caught vs. escaped, discipline breaks (red build, unnamed stubs, scope theft past a hole's contract), and — specific to this variant — how often the skeleton survived contact with implementation (signature churn after approval is the tell). Hole-driven should win when the design is the risk and types are strong; expect friction in untyped codebases or exploratory work where the shape genuinely can't be known upfront.
