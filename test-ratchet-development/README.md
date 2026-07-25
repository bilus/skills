# Test Ratchet Development — README

A staged-delivery workflow for AI coding agents. The agent plans first, then opens every stage by writing ONE acceptance test for the stage's behavior and watching it fail; it implements via red-green-refactor until that test passes with the whole suite green, and halts for your review at every stage boundary. The defining rule is the ratchet: existing tests may only be added to or strengthened — never edited, weakened, skipped, or deleted to reach green. Test changes require your explicit sign-off.

## Why this approach

Review capacity is the bottleneck with agents, and the most dangerous agent failure is the one CI can't see: quietly deleting an assertion, widening a tolerance, or rewriting an acceptance test so broken behavior passes. The ratchet makes that class of failure structurally impossible — the suite is an append-only record of every promise the code has made, and the agent physically cannot trade a promise away without your signature. Watching each test fail before implementation is the proof the test can catch anything at all; a test written after the code passes immediately and proves nothing. Each stage stays small (one behavior, one acceptance test, 400 LOC hard cap), so your review fits in one sitting and reverting any stage is safe.

## What the agent will do

1. Write `docs/plans/<date>-<feature>.md`: numbered requirements; stages each naming their single acceptance test, frozen interfaces, size estimates. **Stops for your approval.**
2. Branch. Per stage: write the acceptance test against the system's outside (API/CLI/UI, not internals), run it, verify it fails *for the right reason* (missing behavior, not a typo), and quote the red output in the ledger — the stage's birth certificate.
3. Drive to green in small red-green-refactor cycles, one commit per green cycle, test and code together. Production code written without a failing test demanding it gets deleted and redone properly. Refactoring happens only on green and touches no test.
4. At the boundary: fresh full-suite run quoted, test diff verified append-only, structured handoff (claim, acceptance-test status, risk retired, deliberately-not-done, review focus). **Stops for review.** Bug fixes are always two commits: failing repro test, then minimal fix.

## What you do

- Approve (or edit) the plan — especially each stage's acceptance test intent: these tests are the spec.
- At each halt, read the test diff first, then the implementation. Answer the handoff's single "review focus" question.
- When the agent proposes a change to an existing test, that's the highest-value decision you'll make in this workflow — it is asking permission to renegotiate a promise. Judge the reasoning, not just the diff.
- Flaky tests come to you as findings, not skips; decide their fate explicitly.

## Usage

Save the skill (or drop the folder into your skills directory). Enable **only one** staged-delivery variant at a time — they have overlapping triggers and are meant to be A/B'd, not stacked. It triggers on any multi-edit coding task; you can also invoke it explicitly. Works in any language; needs nothing beyond git and a test runner your CI also runs.

## Comparing against the other variants

Track per feature: number of stages, mean stage LOC, your review minutes per stage, defects caught vs. escaped, discipline breaks (code before test, test edits without sign-off, commits on red), and — specific to this variant — how often the agent proposed test changes and whether those proposals were legitimate. Test ratchet should win when behavioral correctness and regression-proofing are the risk; expect friction on exploratory work or UI-heavy code where acceptance tests are expensive to write first.
