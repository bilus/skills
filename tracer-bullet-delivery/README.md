# Tracer Bullet Delivery — README

A staged-delivery workflow for AI coding agents. The agent plans first, ships a walking skeleton (the thinnest end-to-end path through the real architecture), then delivers one vertical slice — one observable behavior through every layer it needs — per stage, halting for your review at every stage boundary.

## Why this approach

Review capacity, not code production, is the bottleneck with agents. Defect detection in code review collapses past roughly 400 changed lines, so the only lever that scales is forcing the agent to deliver in small, self-contained, revertable stages. Tracer bullet answers the "what shape should a stage be?" question with: *always a running behavior, never a layer*. Layer-shaped PRs (a storage interface with 15 methods and zero callers) are structurally unreviewable — you can't judge an interface without seeing its consumer. Slices put interface and caller in the same diff. The skeleton exists to settle architecture arguments over ~150 lines instead of inside a 3,000-line stage 5.

## What the agent will do

1. Write `docs/plans/<date>-<feature>.md`: numbered requirements, stages with behavior-shaped goals, frozen interfaces, acceptance criteria, size estimates. **Stops for your approval.**
2. Branch, then build stage 1: a walking skeleton — real entry point, real build, real tests, hardcoded response, everything off the path stubbed. Demoable with one command. **Stops for review.**
3. Each later stage: one vertical slice with its acceptance test, 100–300 LOC target / 400 hard cap, full suite run fresh and quoted, structured handoff (claim, risk retired, deliberately-not-done list, verify commands, review focus). **Stops for review.**
4. Keeps an append-only ledger file next to the plan so progress survives context loss; discovered spec problems become plan diffs you approve before code continues.

## What you do

- Approve (or edit) the plan — this is the cheapest review you'll ever do; reject layer-shaped stages here.
- Review the skeleton hardest: it fixes the boundaries everything else hangs on.
- At each halt, answer the handoff's single "review focus" question. "Too big to review" is valid feedback; the required fix is "split it".
- Approve any proposed change to an existing test or frozen interface — the agent may not make these unilaterally.

## Usage

Save the skill (or drop the folder into your skills directory). Enable **only one** staged-delivery variant at a time — they have overlapping triggers and are meant to be A/B'd, not stacked. It triggers on any multi-edit coding task; you can also invoke it explicitly. Works in any language/repo; no extra tooling required beyond git and your test runner.

## Comparing against the other variants

Track per feature: number of stages, mean stage LOC, your review minutes per stage, defects you caught at review vs. escaped, how often the agent broke discipline (skipped halt, layer-shaped stage, dead code), and rework caused by a wrong early interface. Tracer bullet should win when architecture/seams are the risk; expect it to feel slower on small, well-understood changes.
