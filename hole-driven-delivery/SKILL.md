---
name: hole-driven-delivery
description: Use when implementing any feature, service, or multi-step coding task, before writing any code — including when the user asks for incremental delivery, small PRs, reviewable chunks, or "build X". Delivers work as a compiling skeleton full of named holes that are filled one at a time under a build-always-green invariant, halting for human review at every stage boundary. If there is even a 1% chance a coding task spans more than one edit, use this skill.
---

# Hole-Driven Delivery

You deliver code the way typed-hole programmers do: first commit the complete shape of the change — every type, every signature, every module wired together, every body a named hole — then fill holes one at a time, letting the compiler and the tests tell you what each hole must be. The skeleton commit IS the design review. The hole count IS the progress bar. Every stage is a one-sentence claim a human can verify in under an hour. You halt at every stage boundary and wait for review.

Announce at start: "Using hole-driven-delivery: plan → typechecked skeleton → fill holes outside-in, halting at every stage boundary." Create a todo per stage.

## The Iron Laws

```
1. NO CODE BEFORE AN APPROVED PLAN.
2. NO IMPLEMENTATION BEFORE A GREEN SKELETON. EVERY BODY STARTS AS A NAMED HOLE.
3. THE BUILD IS NEVER RED. ONE HOLE OPEN AT A TIME.
4. HALT AT EVERY STAGE BOUNDARY. REVIEW IS NOT OPTIONAL.
```

Violating the letter of these laws is violating their spirit. There are no exceptions without your human partner's explicit permission.

**REQUIRED BACKGROUND:** Apply the `avoiding-ai-tells` skill to everything a human will read — code, comments, tests, commit messages, handoffs, and PR descriptions. If it is not loaded, read `avoiding-ai-tells/SKILL.md` from this skill collection before writing any code. Staged delivery only works if reviewers judge each diff on its merits; machine-styled output turns them hostile before they read a line.

## The Outer Loop

### 1. Plan before code

Write `docs/plans/<YYYY-MM-DD>-<feature>.md` before touching any source file. If no spec exists, the plan's first section IS the spec: restate the request as numbered, testable requirements and get them confirmed.

Every stage in the plan states:

- **Goal** — one sentence, behavior-shaped: "the system now does X, observably." (Stage 1, the skeleton, is the single exception: its goal is the design itself.)
- **Spec reference** — which numbered requirement this stage advances.
- **Dependencies** — which stages must land first.
- **Holes** — the hole IDs (see below) this stage will fill.
- **Acceptance criterion** — an executable check (test name or command) that defines done.
- **Estimated size** — net hand-written LOC.

### 2. Plan approval gate

Present the plan and stop. Do not write code, scaffold projects, or "just set up the basics" while waiting. A human approves the plan; plan changes later are plan diffs, approved the same way. Never work on main/master: create a branch (or worktree) after approval, before stage 1.

### 3. Execute one stage at a time

Work the current stage to its acceptance criterion using the hole discipline below. Do not start stage N+1 on top of unreviewed stage N.

### 4. The ledger

Maintain `docs/plans/<YYYY-MM-DD>-<feature>.ledger.md`. First line: `# Ledger — plan: <plan path>`. Append one line per event: stage started, stage complete (with commit sha and remaining-hole count), hole added, hole filled, finding deferred, decision made, anything discarded. A silent discard is forbidden. After compaction or context loss, trust the ledger, `git log`, and the grep census (below) over your own recollection.

### 5. Stage boundary: verify, hand off, HALT

A stage is done only when every box checks:

- [ ] Build, lint, typecheck, and the FULL test suite pass — run fresh, in this message, output quoted. If you haven't run the command in this message, you cannot claim it passes.
- [ ] The stage's acceptance criterion demonstrably holds (test output or command output quoted).
- [ ] The hole census is quoted: `grep -rn "HOLE(" <src>` output, with the delta from last stage. Every hole this stage claimed is gone; every hole it added is in the plan.
- [ ] Single-revert safe: reverting this stage alone leaves the branch healthy.
- [ ] No unused code without a stated reason and a follow-up stage. (Holes named in the plan are not dead code; anything else unused is.)
- [ ] No existing test edited, weakened, skipped, or deleted. If a test seems wrong, propose the change and halt for sign-off.
- [ ] Diff within budget: target 100–300 net hand-written LOC, hard cap 400 (generated files and lockfiles excluded but labeled). Over the cap? Split the stage — do not ask forgiveness.
- [ ] Structure and behavior are separate: refactoring is its own stage with tests unchanged.

Then write the handoff and STOP. Open a PR if the repo uses PRs; otherwise present the handoff in conversation. Do not continue until review completes.

```
Stage: <n of N>   Type: <skeleton | fill | refactor | migration-phase>   Spec: <§x>
Claim: <one sentence: what this stage makes true>
Holes: <filled: ids / added: ids / remaining: count>
Risk addressed: <what could have gone wrong, and how this stage retires it>
Not done (deliberately): <deferred items + which stage picks them up>
Verify: <exact commands run, and their result>
Review focus: <the one question the reviewer should answer>
```

### 6. Drift and breakage rules

- If implementation reveals the plan or spec is wrong: STOP the stage, write a small plan diff, get it approved, then continue. Code never silently diverges from the plan.
- Breaking schema or published-interface changes use expand–contract: expand additively, migrate, contract. One phase per stage, every phase backward compatible and revertable.
- Circuit breakers: 3 consecutive failed attempts at the same hole → stop, write what you tried in the ledger, ask your human partner. 2 stages in a row that miss their acceptance criterion → the plan is wrong; replan.

### 7. Endgame: squash to stage commits

After the final stage is approved — and only then — reshape the branch so the history a human inherits is stage-sized:

- One commit per stage, in stage order. Fine-grained fill commits and ledger commits fold into their stage's commit (`git reset --soft` to the branch base and re-commit stage by stage, or an equivalent rebase). Each stage commit's tree must equal that stage's approved boundary tree, so every commit stays green and single-revert-safe.
- Each commit message: an imperative subject naming what changed, then a body of one or two sentences in the first person ("I added the sliding-window limiter and wired it into create.") or third person ("This commit fixes the visit-count race."). Brief. No bullet lists, no hunk enumeration. This body style is your human partner's explicit house style and overrides avoiding-ai-tells' "no 'This commit...' openers" rule.
- The final history must read like a human-maintained repository, bereft of methodology artifacts. Messages describe the functionality added, the change made, or the bug fixed — NEVER the process that produced it. Banned from final commit messages: "hole", "HOLE(", "fill", "census", "skeleton", "stage", "ledger", "plan approval", spec-section references. The scaffolding commit is described as what it is to a maintainer ("Set up the linkhub package: app factory, store API, route stubs") — not as a design-review artifact. A reader of `git log` alone should see a feature being built, not a methodology being followed.
- Squashing collapses real history; it never fabricates one. Do not backdate, reorder work across stages, or invent commits for work that happened differently.
- Re-run the full suite on the reshaped tip and quote the output — a green run proves only the tree it ran on.
- The plan and ledger files stay in the tree; their churn history need not survive, the files themselves are the record. Record the pre-squash tip sha in the ledger before reshaping.

## The Inner Strategy: Skeleton, Then Fill

### The hole convention

A hole is an unimplemented body that (a) satisfies the typechecker and (b) fails loudly if executed, carrying a greppable tag:

- Rust: `todo!("HOLE(3): parse the tenant header")`
- Python: `raise NotImplementedError("HOLE(3): parse the tenant header")`
- TypeScript/JS: `throw new Error("HOLE(3): parse the tenant header")`
- Go: `panic("HOLE(3): parse the tenant header")`
- Haskell: `error "HOLE(3): parse the tenant header"` (or a typed hole `_hole3` during local work)

The number is the plan-stage ID; the sentence is the hole's contract — what a correct body must do. In dynamically typed languages, every holed function gets full type annotations and the typechecker (mypy/pyright, tsc, etc.) joins the green gate; the annotations are what make the skeleton checkable at all.

### Stage 1: the skeleton

Before implementing anything, write the ENTIRE shape of the planned change: every new type, every function signature, every module and its wiring, every trait/interface — with every body a hole and zero logic. The skeleton must compile and typecheck, and the existing suite must stay green (holes are never executed by it).

Commit the skeleton alone. This commit is the cheapest design review your human partner will ever do: they review names, types, seams, and dependency direction over a few hundred declaration lines, before any implementation exists to bias the discussion. HALT here — skeleton approval is a mandatory review boundary. Signatures approved at this boundary are frozen; changing one later is itself a stage.

### The fill loop

Each subsequent stage fills the hole-cluster its plan entry names. Within a stage, fill one hole at a time:

1. **Pick the next hole**: outside-in — the unblocked hole nearest the system's entry point. Tie-break by most-constrained-first: the hole whose types, callers, and contract admit the fewest possible implementations. Let the code that consumes a hole exist before the code inside it.
2. **Interrogate the oracle before writing**: read the hole's signature, its contract sentence, its call sites, and what the typechecker says about the body you must produce. The compiler output is the spec of the hole; read it, do not invent from memory.
3. **Fill by refinement, not by leaps.** Legal moves: (a) direct fill, when the body is small and the types nearly dictate it; (b) refine — replace the hole with exactly one layer of structure (a branch, a match, a delegating call) whose gaps are NEW, smaller, named holes added to the plan and ledger. After every move the build and typecheck are green. Never have more than one hole torn open mid-edit.
4. **Every filled hole lands with its tests.** The tests that pin the hole's contract are part of the fill, in the same commit. A fill without tests is not filled; it is a hole wearing a body.
5. **Commit per hole** (or per tightly coupled hole pair) with the hole ID in the message: `fill HOLE(3): parse the tenant header`.

If mid-fill you discover a needed function that doesn't exist: do NOT implement it inline. Declare it as a new named hole, ledger it, keep filling the current hole against its signature. Implementation beyond the current hole's contract is scope theft from a later stage.

### The census

`grep -rn "HOLE(" <src>` is the authoritative to-do list and progress bar. It is what makes this workflow resumable: after any interruption, the census plus the ledger reconstructs exactly where the work stands. The feature is done when the census returns nothing, the suite is green, and every plan stage has a `complete` ledger line.

## Rationalizations

| Excuse | Reality |
|---|---|
| "I'll just implement this whole subtree while I'm here" | Everything past the current hole's contract belongs to a later stage. Declare holes and move on; scope theft now is an unreviewable diff later. |
| "The build can stay red while I edit these three files together" | A red build silences the one oracle telling you what each hole must be. One hole open at a time; green after every move. |
| "I know what type this needs, no need to run the checker" | Reading the checker is the method. Inventing from memory is how holes get filled with plausible wrong bodies. |
| "The skeleton is overhead, I can hold the design in my head" | The skeleton is the design review. Your head is not reviewable, and it does not survive compaction. |
| "I'll add tests once the holes are all filled" | An untested fill is indistinguishable from a wrong fill. Tests land in the same commit as the body they pin. |
| "This stage is too simple to need a review halt" | Simple stages are where unexamined assumptions hide. The halt costs minutes; a wrong signature costs stages. |
| "The cap is arbitrary; this 600-line diff is coherent" | Coherent is not reviewable. Defect detection collapses past ~400 lines. Split. |
| "I'll quietly adjust the plan; explaining takes longer" | A plan the code has silently left is a lie that costs every later stage. Amend first, then build. |
| "Tests passed earlier in the session" | A green run proves only the tree it ran on. Run the suite fresh at the boundary and quote it. |

## Red Flags — Stop and Reread the Laws

If you catch yourself thinking any of these, STOP: "I'll skip the skeleton for this one" · "let me rough out several bodies and fix types later" · "this stub doesn't need a name or a contract" · "grep says 12 holes but I'm sure it's fewer" · "I'll batch the whole fill into one commit at the end" · "the signature is wrong but changing it now is faster than a plan diff." All of these mean: return to the plan, restore the green invariant, halt at the boundary.

The through-line: the shape is committed and reviewed before any body exists; every body is demanded by a type, pinned by a test, and delivered inside a diff a human can hold in their head — and `grep` can always tell you exactly what remains.
