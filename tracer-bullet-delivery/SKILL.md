---
name: tracer-bullet-delivery
description: Use when implementing any feature, service, or multi-step coding task, before writing any code — including when the user asks for incremental delivery, small PRs, reviewable chunks, or "build X". Delivers work as a walking skeleton followed by vertical slices, halting for human review at every stage boundary. If there is even a 1% chance a coding task spans more than one edit, use this skill.
---

# Tracer Bullet Delivery

You deliver code the way a tracer round finds a target: fire the thinnest possible shot through the entire system first, observe where it lands, then follow with slices that each carry one real behavior end to end. Every stage is a one-sentence claim a human can verify in under an hour. You halt at every stage boundary and wait for review.

Announce at start: "Using tracer-bullet-delivery: plan → skeleton → vertical slices, halting at every stage boundary." Create a todo per stage.

## The Iron Laws

```
1. NO CODE BEFORE AN APPROVED PLAN.
2. STAGE 1 IS A WALKING SKELETON. EVERY LATER STAGE IS A VERTICAL SLICE.
3. NO LAYER-SHAPED WORK. EVER.
4. HALT AT EVERY STAGE BOUNDARY. REVIEW IS NOT OPTIONAL.
```

Violating the letter of these laws is violating their spirit. There are no exceptions without your human partner's explicit permission.

**REQUIRED BACKGROUND:** Apply the `avoiding-ai-tells` skill to everything a human will read — code, comments, tests, commit messages, handoffs, and PR descriptions. If it is not loaded, read `avoiding-ai-tells/SKILL.md` from this skill collection before writing any code. Staged delivery only works if reviewers judge each diff on its merits; machine-styled output turns them hostile before they read a line.

## The Outer Loop

### 1. Plan before code

Write `docs/plans/<YYYY-MM-DD>-<feature>.md` before touching any source file. If no spec exists, the plan's first section IS the spec: restate the request as numbered, testable requirements and get them confirmed.

Every stage in the plan states:

- **Goal** — one sentence, behavior-shaped: "the system now does X, observably." If you cannot phrase the goal this way, the stage is wrong-shaped; reshape it.
- **Spec reference** — which numbered requirement this stage advances.
- **Dependencies** — which stages must land first.
- **Interfaces** — the exact signatures this stage exposes to later stages. These are frozen at the stage boundary; changing a frozen seam later is itself a stage.
- **Acceptance criterion** — an executable check (test name or command) that defines done.
- **Estimated size** — net hand-written LOC.

### 2. Plan approval gate

Present the plan and stop. Do not write code, scaffold projects, or "just set up the basics" while waiting. A human approves the plan; plan changes later are plan diffs, approved the same way. Never work on main/master: create a branch (or worktree) after approval, before stage 1.

### 3. Execute one stage at a time

Work the current stage to its acceptance criterion using the slice discipline below. Do not start stage N+1 on top of unreviewed stage N.

### 4. The ledger

Maintain `docs/plans/<YYYY-MM-DD>-<feature>.ledger.md`. First line: `# Ledger — plan: <plan path>`. Append one line per event: stage started, stage complete (with commit sha), finding deferred, decision made, anything discarded. A silent discard is forbidden — every dropped idea or deferred fix is a ledger line. After compaction or context loss, trust the ledger and `git log` over your own recollection; resume at the first stage without a `complete` line.

### 5. Stage boundary: verify, hand off, HALT

A stage is done only when every box checks:

- [ ] Build, lint, typecheck, and the FULL test suite pass — run fresh, in this message, output quoted. If you haven't run the command in this message, you cannot claim it passes.
- [ ] The stage's acceptance criterion demonstrably holds (test output or command output quoted).
- [ ] Single-revert safe: reverting this stage alone leaves the branch healthy.
- [ ] No unused code without a stated reason and a follow-up stage. Unused exports and speculative parameters are defects.
- [ ] No existing test edited, weakened, skipped, or deleted. If a test seems wrong, propose the change and halt for sign-off.
- [ ] Diff within budget: target 100–300 net hand-written LOC, hard cap 400 (generated files and lockfiles excluded but labeled). Over the cap? Split the stage — do not ask forgiveness.
- [ ] Structure and behavior are separate: refactoring is its own stage with tests unchanged.

Then write the handoff and STOP. Open a PR if the repo uses PRs; otherwise present the handoff in conversation. Do not continue until review completes.

```
Stage: <n of N>   Type: <skeleton | slice | refactor | migration-phase>   Spec: <§x>
Claim: <one sentence: what this stage makes true>
Risk addressed: <what could have gone wrong, and how this stage retires it>
Not done (deliberately): <deferred items + which stage picks them up>
Verify: <exact commands run, and their result>
Review focus: <the one question the reviewer should answer>
```

### 6. Drift and breakage rules

- If implementation reveals the plan or spec is wrong: STOP the stage, write a small plan diff, get it approved, then continue. Code never silently diverges from the plan.
- Breaking schema or published-interface changes use expand–contract: expand additively, migrate, contract. One phase per stage, every phase backward compatible and revertable.
- Circuit breakers: 3 consecutive failed attempts at the same problem → stop, write what you tried in the ledger, ask your human partner. 2 stages in a row that miss their acceptance criterion → the plan is wrong; replan.

## The Inner Strategy: Skeleton, Then Slices

### Stage 1: the walking skeleton

For any new project, service, or feature that crosses a new architectural boundary, stage 1 is always the thinnest possible end-to-end path through the REAL architecture: one input in, one output out, using the real entry point, real build, real test harness, real deploy configuration if deploy exists, and the real datastore (or a stub standing at the real seam). The response can be hardcoded. Everything off the path is stubbed.

The skeleton's claim is architectural: "these are the system's boundaries, this is how it runs, this is the seam everything else hangs on." A reviewer settles boundary arguments over 150 lines here that would otherwise surface inside a 3,000-line stage 5. For work inside an existing system where the end-to-end path already runs, stage 1 is a thin tracer through each NEW seam the feature introduces — a stub wired all the way through, invocable, tested.

The skeleton must be demoable: state the exact command that exercises the path end to end, and quote its output in the handoff.

### Every later stage: one vertical slice

A slice implements exactly one observable behavior through every layer it needs — and only the parts of each layer it needs. One handler, one service function, one query, one acceptance test. Not the full repository interface: the one method this behavior calls. The slice is how interfaces earn their existence: an interface member and its first caller land in the same diff, so the reviewer can judge whether the shape is right. An interface member with no caller in the stage is a defect.

Slice rules:

- Order slices riskiest-seam-first, then by user value. Retire uncertainty early; do not save the hard slice for last.
- Extend a shared interface only at the moment a slice's caller needs the extension. "Later slices will need it" is not a caller.
- If one part of a slice is genuinely uncertain (unfamiliar API, concurrency question, performance question), split off a spike stage first: its deliverable is a written conclusion in the ledger (what was tried, what was measured, what was decided), even if its code is discarded.
- A bug found during a slice gets its own two-commit treatment inside a stage: first a failing test reproducing it, then the minimal fix.

### The shape test

Before starting any stage, say its goal out loud. "Implement the storage layer," "add the models," "set up the service classes," "build out the API" — all layer-shaped, all forbidden. "A user can fetch their profile," "the pipeline rejects malformed input with a 400" — behavior-shaped, allowed. If the goal names a layer instead of a behavior, restructure before writing a line.

## Rationalizations

| Excuse | Reality |
|---|---|
| "Building the full storage layer now saves rework later" | A reviewer shown 15 interface methods and zero callers cannot evaluate any of them. The slice's caller is the evidence. Interfaces earn members one caller at a time. |
| "The skeleton is trivial, skip to the real features" | The skeleton settles deploy, wiring, and boundary arguments at 150 lines. Skipping it defers those arguments into your largest, least reviewable stage. |
| "This helper will be needed by three later slices" | Land it with the first slice that calls it. Until then it is speculation the reviewer cannot check. |
| "This stage is too simple to need a review halt" | Simple stages are where unexamined assumptions hide. The halt costs minutes; a wrong seam costs stages. |
| "I'll batch these three behaviors into one stage, review once" | Three claims cannot be verified as one. The reviewer's question must fit in one sentence. |
| "The cap is arbitrary; this 600-line diff is coherent" | Coherent is not reviewable. Defect detection collapses past ~400 lines. Split. |
| "I'll quietly adjust the plan; explaining takes longer" | A plan the code has silently left is a lie that costs every later stage. Amend first, then build. |
| "Tests passed earlier in the session" | A green run proves only the tree it ran on. Run the suite fresh at the boundary and quote it. |

## Red Flags — Stop and Reread the Laws

If you catch yourself thinking any of these, STOP: "let me just scaffold all the entities first" · "I'll wire up review after a few more stages" · "the interface is obvious, callers can come later" · "this doesn't need to actually run yet" · "I'll demo it at the end" · "one big PR is easier for everyone." All of these mean: return to the plan, reshape the stage, halt at the boundary.

The through-line: every stage is a one-sentence claim about a running system, proven by a command you ran in this message, delivered inside a diff a human can hold in their head.
