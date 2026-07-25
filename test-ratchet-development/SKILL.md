---
name: test-ratchet-development
description: Use when implementing any feature, service, or multi-step coding task, before writing any code — including when the user asks for incremental delivery, small PRs, reviewable chunks, or "build X". Delivers work as a sequence of stages each opened by one failing acceptance test, with a test suite that only ever tightens, halting for human review at every stage boundary. If there is even a 1% chance a coding task spans more than one edit, use this skill.
---

# Test Ratchet Development

You deliver code the way a ratchet wrench turns: one click at a time, never backward. Each stage begins with ONE failing acceptance test that defines what the stage will make true; the stage ends when that test passes with the whole suite green. Existing tests only ever get stronger — never edited, weakened, skipped, or deleted to reach green. The suite is the record of every promise the code has made, and you are not allowed to break promises. Every stage is a one-sentence claim a human can verify in under an hour. You halt at every stage boundary and wait for review.

Announce at start: "Using test-ratchet-development: plan → one failing acceptance test per stage → red-green-refactor, halting at every stage boundary." Create a todo per stage.

## The Iron Laws

```
1. NO CODE BEFORE AN APPROVED PLAN.
2. NO PRODUCTION CODE WITHOUT A FAILING TEST DEMANDING IT.
3. THE RATCHET ONLY TIGHTENS: EXISTING TESTS ARE NEVER EDITED, WEAKENED,
   SKIPPED, OR DELETED TO GET TO GREEN.
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
- **Acceptance test** — the name and one-line intent of the single acceptance test that will open the stage and define done.
- **Estimated size** — net hand-written LOC.

### 2. Plan approval gate

Present the plan and stop. Do not write code, scaffold projects, or "just set up the basics" while waiting. A human approves the plan; plan changes later are plan diffs, approved the same way. Never work on main/master: create a branch (or worktree) after approval, before stage 1.

### 3. Execute one stage at a time

Work the current stage to green using the ratchet discipline below. Do not start stage N+1 on top of unreviewed stage N.

### 4. The ledger

Maintain `docs/plans/<YYYY-MM-DD>-<feature>.ledger.md`. First line: `# Ledger — plan: <plan path>`. Append one line per event: stage started (with the quoted failing output of its acceptance test), stage complete (with commit sha), finding deferred, decision made, test change proposed, anything discarded. A silent discard is forbidden. After compaction or context loss, trust the ledger and `git log` over your own recollection; resume at the first stage without a `complete` line.

### 5. Stage boundary: verify, hand off, HALT

A stage is done only when every box checks:

- [ ] Build, lint, typecheck, and the FULL test suite pass — run fresh, in this message, output quoted. If you haven't run the command in this message, you cannot claim it passes.
- [ ] The stage's acceptance test exists, was seen failing at stage open (quoted in the ledger), and now passes (quoted here).
- [ ] The test diff is append-only: `git diff <base> -- <test paths>` shows additions and new files only. Any modification to an existing test appears in the handoff with its approved sign-off, or the stage is not done.
- [ ] Single-revert safe: reverting this stage alone leaves the branch healthy.
- [ ] No unused code without a stated reason and a follow-up stage. Unused exports and speculative parameters are defects.
- [ ] Diff within budget: target 100–300 net hand-written LOC, hard cap 400 (generated files and lockfiles excluded but labeled). Over the cap? Split the stage — do not ask forgiveness.
- [ ] Structure and behavior are separate: refactoring is its own stage (or its own commits) with tests unchanged.

Then write the handoff and STOP. Open a PR if the repo uses PRs; otherwise present the handoff in conversation. Do not continue until review completes. Reviewers read the test diff first; write the handoff so they can.

```
Stage: <n of N>   Type: <slice | refactor | migration-phase | bugfix>   Spec: <§x>
Claim: <one sentence: what this stage makes true>
Acceptance test: <name> — failed at open (ledger), passes now (quoted below)
Test diff: <append-only | modification proposed & signed off: link>
Risk addressed: <what could have gone wrong, and how this stage retires it>
Not done (deliberately): <deferred items + which stage picks them up>
Verify: <exact commands run, and their result>
Review focus: <the one question the reviewer should answer>
```

### 6. Drift and breakage rules

- If implementation reveals the plan or spec is wrong: STOP the stage, write a small plan diff, get it approved, then continue. Code never silently diverges from the plan.
- Breaking schema or published-interface changes use expand–contract: expand additively, migrate, contract. One phase per stage, every phase backward compatible and revertable.
- Circuit breakers: 3 consecutive failed attempts to reach green on the same test → stop, write what you tried in the ledger, ask your human partner. 2 stages in a row that miss their acceptance test → the plan is wrong; replan.

## The Inner Strategy: The Ratchet

### Opening a stage: one failing acceptance test

Write exactly ONE acceptance test expressing the stage's goal from the outside — through the system's real interface (API call, CLI invocation, UI-level check), not against internals. Run it. Watch it fail, and verify it fails for the RIGHT reason: an assertion about missing behavior, not an import error, a typo, or a broken fixture. A test that errors instead of failing is not red; fix the test's plumbing first. A test that passes immediately is testing behavior that already exists — the stage is misdefined; go back to the plan.

Quote the failing output in the ledger. This quote is the stage's birth certificate: proof the test can catch the absence of the behavior it pins.

### Inside a stage: red–green–refactor

Drive the acceptance test to green through unit-level TDD cycles:

1. **Red** — write the smallest unit test that moves toward the acceptance test. Run it. See it fail for the right reason.
2. **Green** — write the MINIMAL production code that passes it. Test fails? Fix the code, not the test. Run the full suite, not just the new test.
3. **Refactor** — only on green. Improve structure; add no behavior; touch no test.
4. **Commit** — one commit per green cycle, test and code together. Never commit red.

Wrote production code with no failing test demanding it? Delete it. Not "keep it as reference," not "adapt it while writing the test" — delete it, write the test, watch it fail, write the code again. Code written before its test was never proven catchable by that test.

### The ratchet rules

- Existing tests are promises. You may ADD tests and you may STRENGTHEN assertions. You may not edit, loosen a tolerance of, delete, skip, `xfail`, or reorder-around an existing test to reach green — each of those is backsliding, and backsliding is the one thing this workflow exists to make impossible.
- Believe an existing test is genuinely wrong? STOP. Write the proposed test diff and your reasoning in the handoff, and halt for explicit human sign-off. Only a signed-off test change may land, and it lands as its own commit, labeled.
- A diff that modifies existing tests and adds production code in the same commit is a defect, no matter what it contains.
- Flaky test in your way? Flakiness is a finding for the ledger and the human — skipping it is deleting it with extra steps.

### Bug fixes

Every bug fix is a two-commit stage: commit 1 adds a failing test that reproduces the defect (quoted red output in the ledger); commit 2 is the minimal fix that turns it green. A reviewer can check out commit 1, watch the test fail, and know the fix is exactly the fix. No fix lands without its reproducing test — the ratchet must click over the bug so it can never return.

### Ordering

Order stages riskiest-behavior-first, then by user value. If a stage's behavior is genuinely uncertain (unfamiliar API, concurrency, performance), split off a spike stage first: its deliverable is a written conclusion in the ledger (what was tried, measured, decided), even if its code is discarded. Spike code never merges without coming back through the ratchet: test first, then implementation.

## Rationalizations

| Excuse | Reality |
|---|---|
| "The test is clearly wrong, I'll just fix it" | Maybe it is — and that decision belongs to your human partner. Propose the diff, halt, get sign-off. A workflow where you may rewrite the promises is not a ratchet. |
| "This assertion is too strict for my implementation" | The assertion was the spec before your implementation existed. Loosening it to fit the code inverts the entire direction of the ratchet. |
| "I'll write the tests after — it's faster" | A test written after passes immediately, which proves nothing: you never watched it catch the bug. The failing run is the only evidence the test tests anything. |
| "I'll skip this flaky test for now" | Skipping is deleting with extra steps. Flakiness goes in the ledger and to the human. |
| "The acceptance test can come once the pieces exist" | Then the pieces define the behavior and the test rubber-stamps it. The test comes first precisely so the code has something external to be wrong against. |
| "This stage is too simple to need a review halt" | Simple stages are where unexamined assumptions hide. The halt costs minutes; an unpinned behavior costs incidents. |
| "The cap is arbitrary; this 600-line diff is coherent" | Coherent is not reviewable. Defect detection collapses past ~400 lines. Split. |
| "I'll quietly adjust the plan; explaining takes longer" | A plan the code has silently left is a lie that costs every later stage. Amend first, then build. |
| "Tests passed earlier in the session" | A green run proves only the tree it ran on. Run the suite fresh at the boundary and quote it. |

## Red Flags — Stop and Reread the Laws

If you catch yourself thinking any of these, STOP: "I already manually tested it" · "the test passes immediately, close enough" · "I'll comment this assertion out temporarily" · "widening the tolerance isn't really weakening" · "it's about the spirit of TDD, not the ritual" · "this is different because…". All of these mean: delete the untested code, return to red, halt at the boundary.

The through-line: every behavior the system has is a test somebody watched fail; the suite only accumulates promises; and every stage delivers one new promise inside a diff a human can hold in their head.
