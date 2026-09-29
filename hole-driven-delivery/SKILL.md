---
name: hole-driven-delivery
description: Use when implementing any feature, service, or multi-step coding task, before writing any code, including when the user asks for incremental delivery, small PRs, reviewable chunks, a design review, or "build X". Use it for any coding task that may span more than one edit.
---

# Hole-driven delivery

You deliver a change as a sequence of reviewed stages, and every review shows a diagram together with its code. The first review, the plan gate, shows the requirements, the vocabulary, the top level of a data flow design and its compiling skeleton: each box's function declared, with a named hole for its body. A hole returns mock data, so a smoke test runs the whole flow on the skeleton and passes, as in tracer-bullet development, and each acceptance test that needs a real body is skipped with the tag of the stage that fills its holes. Your human partner reviews the holes and the tests. Then you fill the holes one at a time, and the suite stays green at every commit. The compiler's errors and the tests constrain each hole's body. The design comes before the code: you draw a diagram first and declare its functions second, and your human partner reviews the two at once. A fill that splits a function into steps draws that process's child diagram and declares the new functions with holes, in the same stage, so the design grows one level at a time. An approved diagram changes only with your human partner's approval, and so does the vocabulary. At every stage boundary you rebuild a review page that shows the design and the code before the stage, as a diff and after it, with the new score beside the last approved one, and you halt until your human partner approves the stage.

Announce at start: "Using hole-driven-delivery: requirements, vocabulary, and a top diagram reviewed with its typechecked skeleton and a passing smoke test, then holes filled outside-in with the suite green, each lower diagram drawn beside its code, halting at every stage boundary with a review page and a score." Create one todo per outer-loop step now, and one per plan stage after approval.

## The iron laws

```
1. NO IMPLEMENTATION BEFORE AN APPROVED PLAN. THE PLAN GATE REVIEWS THE REQUIREMENTS, THE VOCABULARY, THE TOP DIAGRAM AND ITS SKELETON OF HOLES.
2. A DIAGRAM AND ITS CODE CHANGE TOGETHER. A NEW BOX GOES TO REVIEW AS ITS FUNCTION'S SIGNATURE WITH A HOLE, AND A LATER STAGE FILLS IT. AN APPROVED DIAGRAM OR TERM CHANGES ONLY WITH APPROVAL.
3. NO IMPLEMENTATION BEFORE A GREEN SKELETON. EVERY NEW BODY STARTS AS A NAMED HOLE THAT RETURNS MOCK DATA, AND A SMOKE TEST RUNS EACH NEW DIAGRAM'S PROCESS ON THAT DATA.
4. THE BUILD IS NEVER RED, AND EVERY COMMIT PASSES THE SUITE. ONE HOLE OPEN AT A TIME.
5. HALT AT EVERY STAGE BOUNDARY, WITH THE REVIEW PAGE AND THE SCORE UP TO DATE.
```

A workaround that keeps a law's wording but defeats its purpose is a violation. Only your human partner can grant an exception to a law. When your human partner lifts law 5 for a run of stages ("continue without asking"), you still do the boundary work of step 9 at every boundary.

Required background: apply the `avoiding-ai-tells` skill to all human-facing text: code, comments, commit messages, handoffs and the review page. If that skill is not in your context, read `../avoiding-ai-tells/SKILL.md`, relative to this skill's directory, before writing anything.

Required reference: `design.md` in this skill's directory holds the design rules, `review-page.md` the page's content, and `prompts.md` the prompts for sub-agents. `tools/dfdmetrics/README.md` defines the metric terms: home, level, live range, read distance, pressure and the findings. `tools/dfdreview/README.md` describes the page builder. Read each file before the step that uses it.

## Tools

- dfd renders the diagrams. Its source is at https://github.com/bilus/dfd; install it with `go install github.com/bilus/dfd/cmd/dfd@latest`. The review page needs a dfd with `--patch`, footnotes and explicit numbers.
- dfdmetrics measures the design and prints the score. It is a Go module of its own in this skill's `tools/dfdmetrics` directory: run `go install ./cmd/dfdmetrics` inside that directory.
- dfdreview builds the review page. It is a Go module in `tools/dfdreview`: run `go install ./cmd/dfdreview` inside that directory, with Go 1.27 or later. It runs `git`, `dfd` and the `go` command.

`go install` puts the binaries in `$(go env GOBIN)`, or in `$(go env GOPATH)/bin` when GOBIN is empty. Put that directory on the PATH, or call the binaries by their full path. Before step 1, check all three: `dfd --help` lists `-patch`, and `dfdmetrics` and `dfdreview` without arguments print their usage.

The design rules use Go's terms: types, import paths, the standard library. In another language, write the `# type:` comments in that language and the references as `module.Name`. dfdmetrics then still measures the state, but it classifies packages by Go's standard library, so check its package findings by hand.

## Files

| File | Holds |
|---|---|
| `docs/plans/<date>-<feature>.md` | the plan: requirements, questions and assumptions, the change in brief, the metaphor, the stages |
| `docs/plans/<date>-<feature>.ledger.md` | the ledger, one line per event |
| `docs/vocabulary.md` | the terms, one per line: `- term: definition` |
| `docs/flow.dfd`, `docs/flow.N.dfd`, `docs/flow.N.M.dfd` and so on | the leveled design: the top diagram and one child diagram per decomposed process |
| `docs/review/score.tsv` | the score card of the last approved review |
| `docs/review/score.new.tsv`, `docs/review/report.txt` | the score card and the dfdmetrics report of the current review |
| `docs/review/index.html` | the review page, rebuilt at every review |

Another agent must be able to resume the work from these files, `git log` and the hole census alone.

## The outer loop

### 1. Requirements

Write the plan before touching any source file, and create the ledger beside it (step 8). If the repository's instructions name a branch for all work, use it. Otherwise create a branch or worktree now, never main or master, since the skeleton comes before the gate. If the task has no spec, the plan's first section is the spec: numbered, testable requirements. For each question you would ask your human partner, write the question and your assumption in the plan, and continue to the gate.

### 2. Vocabulary

Start `docs/vocabulary.md` from the requirements, with one line per term and a definition checked against the requirements. Add the names of processes, data and stores during step 3, and the words of the metaphor during step 4. Step 5 reviews the whole vocabulary before the gate.

### 3. Design and skeleton

Draw the top diagram in the dfd format, following `design.md`: one box per top-level process, each typically with a function of its own. Child diagrams come later, with the fills that create their functions (see the fill loop). In short:

- A process box starts with its number, as in `[3. Validate the sum types`, so the number stays put when a box is inserted before it.
- A process box states the step's action, its purpose for the later steps (unless the outgoing arrow's label already gives it), and the functions behind it.
- An arrow label names the input data of the next step, and a `# type:` comment gives its type.
- Every piece of I/O is drawn. The program's own state is a store. Anything outside the program is an external system beside the process, as in `<Event bus>`, and that includes the user's own files and directories. dfdmetrics counts store items only, and every store write has a matching read.
- A process gets a child diagram in the stage whose fill splits its function into steps. State stays inside the smallest process that uses it, and the top diagram shows only the data and the state shared across its boxes.
- Each box references one package of the design.

The code follows the diagram. Each decomposed process is one named function or method, and its body calls the functions of its child processes in the order of its diagram, so the top function reads like the top diagram. A leaf box typically references its process's own function, and it may reference several functions of one package. Its action line matches the doc comment of its first reference.

Then write the skeleton (see the inner strategy): declare every function and type named by the boxes' references and the `# type:` comments, with each body a named hole whose sentence is its contract, and write the smoke test and the acceptance tests. Draw first and declare second, so each signature takes its shape from the diagram. The holes are the input of the review: a box without its declared function, or a function without its box, gives your human partner nothing to check.

### 4. Metaphor

Write one system metaphor in the plan: the top diagram as one familiar thing, such as a tally sheet, a compiler's passes, or a pipeline with a barrier. In at most 60 words, name the thing and match three or four of its parts to parts of the design, as in "Components are bricks, and the code between them is mortar." Choose a thing whose most obvious behavior matches the design's main property, and let the reader imagine the rest.

A reader infers rules from a metaphor, as "stack" implies push and pop at one end. The design review (step 5), not the metaphor, lists those rules and checks each against the design. For each broken rule, fix the design or change the metaphor, and record the decision and the level of the first break in the ledger.

### 5. Reviews and score

Measure the design with dfdmetrics:

```sh
dfdmetrics docs/flow.dfd
dfdmetrics -score docs/flow.dfd
```

The first command prints the report. The second prints the score card as two tab-separated lines: a header (`live_range read_distance top_shared shared pressure package_spread findings`) and one row of values. Every value is better when lower.

Then run the vocabulary review and the design review from `prompts.md` in parallel, each with a sub-agent, and give the design review the report. Set each sub-agent's reasoning effort where the harness allows it, because a sub-agent otherwise inherits the session's effort. In Claude Code, set it to high for the design review and to medium for the vocabulary review. During a review, change none of its input files, for example by editing or regenerating code. If you must change them, give the review a snapshot: a commit or a worktree. Decide on each finding, and record every applied, deferred or discarded finding in the ledger. If you cannot run a sub-agent, run each review yourself with the same prompt, and record that in the ledger. Then write "The change in brief" in the plan: one paragraph that describes the change in the vocabulary's terms.

At the plan gate and at each stage boundary, never in between, save this review's report and score card:

```sh
mkdir -p docs/review
dfdmetrics docs/flow.dfd > docs/review/report.txt
dfdmetrics -score docs/flow.dfd > docs/review/score.new.tsv
```

`docs/review/score.tsv` caches the score card of the last approved review, and the review page shows it beside the new card. When your human partner approves a review, move the new card over the cached one (`git mv -f docs/review/score.new.tsv docs/review/score.tsv`) and commit the move with the ledger line of the approval. At the plan gate, the cache does not exist yet, and the page shows the new card alone.

Keep the columns separate: a weighted total hides the trade-offs between columns from your human partner. Every rise in a column needs a ledger line with the reason, or a design fix.

The score cannot see data carried through a step that does not use it: an item that a step receives and passes on unchanged. Carrying data through a step is the alternative to a store, so list each carried item in the handoff for your human partner's judgment.

### 6. Stages

Every stage in the plan has these fields:

- Goal: one sentence about behavior, "the system now does X, observably". A stage that splits a process into steps is the exception: its goal is the new level, and its acceptance test is the split process's smoke test, which passes on the new holes' mock data.
- Requirement: the number of the stage's requirement.
- Dependencies: the stages that must land first.
- Holes: the IDs of the holes filled in this stage (see the hole convention).
- Acceptance: the name of the test that defines done. For a split stage, it is the split process's smoke test, written in the stage. For any other stage, the skeleton or an earlier stage writes it, skipped with this stage's tag, and this stage deletes the skip (see the hole convention).
- Size: the estimated net hand-written lines.

When a hole's fill will split its process into steps, plan two stages. The first draws the process's child diagram, declares one function per box with a hole, writes the process's smoke test, and ends. The second fills those holes and deletes the skip of an acceptance test that depends on them.

The budget is 100 to 300 net hand-written lines per stage, with a hard cap of 400. Generated files, lockfiles and fixtures copied from a spec fall outside the budget, and the handoff labels them. Split any stage whose estimate exceeds the cap. If a stage's diff crosses the cap during the fill, stop at the last green commit, and halt the stage with a proposal to move the remaining holes into a new stage.

### 7. Plan approval gate

Save the first report and score card (step 5), run the suite, build the review page with dfdreview (see `review-page.md`), then present the plan, the suite's output and the page, and wait for your human partner's approval. Every function or type link on the page opens a declaration of the skeleton. Fix any missing declaration that the page reports before you present the plan. Do not fill any hole while you wait for approval.

Your human partner approves the requirements, the vocabulary, the top diagram with its skeleton and tests, and the stages. Approval freezes the signatures, the top diagram and the vocabulary. A later change to any of them is a plan diff, approved the same way. Then commit the plan, the ledger, the vocabulary, the design, the skeleton, the report, the score card and the review page, and start stage 1.

### 8. The ledger

Keep `docs/plans/<date>-<feature>.ledger.md`, with the first line `# Ledger, plan: <plan path>`. Append one line per event: the start of a stage (with the commit it starts from, which is the previous stage's boundary commit), the completion of a stage (with the remaining holes), a new or filled hole, a change to the design or the vocabulary, a rise in a score column (with the reason), a decision (including one on a review finding), and a discarded finding or abandoned attempt (with the reason). Every discard needs a ledger line. After compaction or context loss, trust the ledger, `git log` and the census over your recollection.

### 9. Stage boundary

A stage is done only when it meets every item of this checklist:

- [ ] The build, lint, typecheck and full test suite all pass in a fresh run in this message, and the handoff quotes the output.
- [ ] The stage's acceptance test runs without a skip and passes in a fresh run, and the handoff quotes its output.
- [ ] The handoff quotes the hole census with its change since the last stage. The census shows no tag with the stage's number, and the plan lists each hole added in the stage.
- [ ] Every box added in this stage comes with its function declared with a hole in this stage's diff, and this stage filled none of those holes. Every function declared in this stage has its box. A smoke test runs the process that holds the new boxes, and it passes on their mock data.
- [ ] The design matches the code. Every reference in a box resolves to a declaration, every `# type:` names a declared type, every function called directly by the top function or by a decomposed process's function has a box in the matching diagram, and the diagrams show the current form of every flow changed in this stage. A call inside a leaf box's function needs no box.
- [ ] After a change to the diagrams, the design review from `prompts.md` ran again, and the ledger records a decision on each finding.
- [ ] The stage's report and score card are in `docs/review/`, and every rise over the cached card has a ledger line.
- [ ] The review page is rebuilt with dfdreview against the stage's start commit, so it shows the stage's changes to the design and the code.
- [ ] Reverting this stage alone leaves the build and the full suite green.
- [ ] Every piece of unused code has a stated reason and a follow-up stage. Planned holes are not dead code.
- [ ] The stage did not edit, weaken, skip or delete any existing test, except to delete the skip of an acceptance test after filling its holes. A new test is skipped only with a HOLE tag, while a hole on its path is open. If a test seems wrong, propose the change and halt for sign-off.
- [ ] The diff is within budget.
- [ ] Structure and behavior are separate: a refactoring is its own stage, with the tests unchanged.

Then write the handoff and wait for your human partner's review. If the repository uses pull requests, open one pull request for the change branch at the first boundary, and post each later handoff on it as a comment. Otherwise present the handoff in the conversation.

```
Stage: <n of N>   Type: <fill | refactor | fixtures | migration-phase>   Requirement: <number>
Claim: <one sentence: what this stage makes true>
Holes: <filled: ids / added: ids / unskipped tests: names / remaining: count>
Size: <net hand-written lines>, generated: <files, or none>
Score: <each column with its change, for example live_range 7 (-2)>
Carried: <items carried through steps that do not use them, or none>
Risk addressed: <what could have gone wrong, and how this stage retires it>
Not done (deliberately): <deferred items and the stage that picks each up>
Verify: <exact commands run, and their result>
Review page: <path or link>
Review focus: <the one question the reviewer should answer>
```

### 10. Changes after the plan gate

The plan's approval freezes the top diagram and the vocabulary, and a stage's approval freezes the child diagrams it drew. A child diagram that a fill draws for its own process belongs to that fill, and it ends the stage: your human partner reviews the new level, with its holes, at that boundary. When the work needs a change to a frozen diagram or to the vocabulary, choose the first of these that does the job:

1. A change inside a function's body. The body may call new unexported helpers of its own package: they belong to the box of the function, and the diagrams do not change.
2. A change to an interface between processes in the deepest diagram that holds it: an arrow's label or type, a store item, a new child process.
3. A change to an interface in a higher diagram.

A change of kind 2 or 3 changes the design, and so does a new or changed term in the vocabulary. Stop the stage, and ask your human partner for approval with the reason, as a plan diff that shows the changed diagram together with its code: the changed signatures, and a hole for each new or changed body. The code never silently leaves the diagram.

- If the code shows that the design or the requirements are wrong, stop the stage in the same way.
- A new hole within the stage's scope needs no approval when the frozen diagrams stay as they are: a hole in a function body, a new unexported helper, or a function in the child diagram that the fill draws for its own process. Declare it (see the fill loop), and your human partner reviews it at the boundary. A function of a new child diagram waits for that review before a later stage fills it. A change to an approved signature, requirement or stage is a plan diff, and it halts the stage.
- Breaking changes to a schema or a published interface go in phases: add the new form beside the old one, migrate every caller, then remove the old form. One phase per stage, each phase backward compatible and revertable.
- After three failed attempts at one hole, stop work on the hole, record each attempt and its failure in the ledger, and ask your human partner. When two stages in a row halt without meeting their acceptance check, the plan is wrong: revise it as a plan diff for approval.

### 11. Endgame

Commit messages follow these rules on every path:

- An imperative subject names the change, and a body of one or two sentences follows, in the first person ("I added the sliding-window limiter and wired it into create.") or the third person ("This commit fixes the visit-count race."). The body does not use bullet lists or list the hunks. This body style is your human partner's house style, and it overrides two rules of avoiding-ai-tells: the rule against "This commit..." openers, and the rule that a body appears only for a diagnosis.
- The final history reads like a human-maintained repository. Messages describe the change to the software, never the process. Final commit messages do not use "hole", "HOLE(", "fill", "census", "skeleton", "stage", "ledger", "plan approval", "score", "review page" or requirement numbers. The scaffolding commit describes its content to a maintainer ("Set up the linkhub package: app factory, store API, route stubs").

If the repository's rules keep all work on one shared, already-pushed branch, do not rewrite its history. Push each stage after its approval as one squashed commit.

Otherwise, only after your human partner approves the final stage, reshape the branch to one commit per stage:

- Record the pre-squash tip in the ledger, and commit that line.
- Start a new branch at the base. For each stage in order, run `git read-tree -u --reset <boundary commit>` and commit the result with a message under the rules above. A stage's boundary commit is the commit on the next stage's start line in the ledger. For the last stage, use the commit with the ledger line on the pre-squash tip, the one allowed difference from an approved tree.
- Point the change branch at the new tip, and force-push it if a pull request is open.
- Squashing collapses real history and never fabricates one: do not backdate commits, move work between stages or invent commits.
- Re-run the full suite on the reshaped tip and quote the output.
- The plan, ledger, vocabulary, design, score card and review page stay in the tree.

### 12. Later features

When the last stage is done, the next feature runs the outer loop again, on the existing design and code:

- Its plan proposes changes to the existing diagrams and vocabulary, following the order of step 10: the smallest change that does the job.
- Its plan gate shows those changes together with their code: the new declarations and the first layer of holes, including holes in existing functions, with a smoke test through the new boxes and the acceptance tests of the new requirements. A new branch or a new call in an existing body starts as a hole, whose mock data keep the existing tests green. Where an inline change reads better than a hole, as in a renamed call or one new argument, change the code inline. Build the review page with the commit before the feature as the base, so the Diff views show the proposal in the diagrams and in the code.
- After its plan gate, the rules of step 10 apply again.

## The inner strategy: skeleton, then fill

### The hole convention

A hole is an unimplemented body that satisfies the typechecker, returns mock data, and carries a comment tagged for grep:

```go
func Parse(r *http.Request) (Tenant, error) {
	// HOLE(3): parse the tenant header
	return Tenant{ID: "acme"}, nil
}
```

Other languages write the same comment in their own syntax: `#` in Python, `--` in Haskell, and `//` in Rust, TypeScript and JavaScript. A hole never panics or throws, since a failing hole turns every test through it red. In Haskell, a typed hole such as `_hole3` fails the build, so use one only to read GHC's report, and replace it before the next build.

On the smoke test's path, a hole's mock data satisfy the smoke test's assertions (see the smoke test). Use the values that a correct body returns for the smoke test's input, so the test keeps passing as the fills replace the mock data. On that path, a hole that writes to a file or a store makes the writes checked by the smoke test. Off the path, a hole returns the simplest value of its type, such as a zero value, and a hole without a result keeps only the comment.

The number is the stage that fills the hole, and the sentence is its contract: the required behavior of a correct body. A hole declared by a split, in a new child diagram, takes the number of a later stage, never the current one. A hole's ID is its stage number and the name of the function that holds it, as in `3 header.Parse`. When one function holds several holes, number them in the order of the body: `3 header.Parse#2`. In a dynamically typed language, every function with a hole gets full type annotations, and the plan names the typechecker, such as mypy or tsc, for the stage-boundary checks.

An acceptance test that needs a real body is written in full and skipped, with the tag in its skip message:

- Go: `t.Skip("HOLE(4): reject a request without a tenant header")`
- Python: `pytest.skip("HOLE(4): ...")`
- Rust: `#[ignore = "HOLE(4): ..."]`
- TypeScript and JavaScript: `it.skip`, with the tag in a comment above it
- Haskell, with hspec: `pendingWith "HOLE(4): ..."`

The number names the stage that fills the last hole on the test's path, and the test's ID is that number with the test's name, as in `4 TestRejectMissingTenant`. That stage deletes the skip and its tag, and leaves the rest of the test unchanged. A tagged skip is only for a test with an open hole on its path: a test that fails with its holes filled shows a bug, and a skip would hide it.

### The smoke test

A smoke test is the acceptance test of a process with a diagram. It calls the process's function on one fixed input, which passes through every box of the diagram, and checks the output. The skeleton writes the smoke test of the top function, and a stage that splits a process writes the smoke test of that process. The test passes on the mock data from its first run, and it keeps passing as the fills replace the mock data with real bodies.

The smoke test asserts only what a correct program produces for its input. When a fill turns the smoke test red, either the fill is wrong or the test asserts a value of the mock data, not of the program. Fix the fill, or propose the change to the test and halt for sign-off.

### The skeleton

Write the declarations named by the top diagram and nothing else: each type, function signature, module and piece of wiring, with every new body a hole and zero logic. A body that only calls the declared functions in the order of its diagram, such as the top function, may be real code in the skeleton, because it is the diagram written in code. Then write the top function's smoke test and the acceptance tests of the requirements, and skip each acceptance test that needs a real body, with the tag of the stage that fills the last hole on its path. The build, the typecheck and the suite all stay green, with the smoke test passing. Before the plan gate, check the design against the skeleton with the design items of the stage-boundary checklist.

The plan gate presents the skeleton beside its diagram. Your human partner reviews the names, types, seams and dependency direction over a few hundred lines of declarations, and each hole's sentence says what its body will do. Approval freezes the signatures, the top diagram and the vocabulary. A later change to any of them is a plan diff that shows the diagram and the code together (step 10).

### The fill loop

Each stage fills the holes listed in its plan entry, one hole at a time:

1. Pick the next hole outside-in: the open hole nearest the entry point whose tests do not execute another open hole, since a test through another open hole depends on that hole's mock data. Break ties by taking the most constrained hole, the one whose types, callers and contract admit the fewest implementations.
2. Before you write the body, read the hole's signature, its contract sentence, its call sites, and the typechecker's report on the body. The compiler's output is the spec of the hole.
3. Fill by refinement. Either fill the hole directly, when the types nearly dictate the body, or replace it with one layer of structure (a branch, a match, a delegating call) whose gaps are new, smaller, named holes. When the layer is a sequence of calls to new functions, the process now has steps of its own: draw its child diagram first, one box per new function with its arrows and their types, then declare each function as a named hole, write the process's smoke test, and end the stage there. Those holes take the next stage's number, and the plan adds them to that stage's Holes field. A helper called among a body's own statements stays inside its box, and its hole joins the current stage. Record each new hole in the ledger. Every move leaves the build and the typecheck green.
4. Every filled hole lands with the tests that pin its contract, in the same commit. When a fill closes the last open hole on a skipped acceptance test's path, the same commit deletes the test's skip, and the test passes unchanged.
5. Commit per hole, or per tightly coupled pair, with the hole's ID in the message.

If a fill needs a missing function, declare it as a named hole and record it in the ledger. A helper stays inside its box: add its ID to the current stage's Holes field, and return to the current hole. A new step of the process being filled gets a box in that process's child diagram, which ends the stage, and its hole goes to the next stage. A missing function that needs a new box, arrow or store in a frozen diagram changes the design, and it waits for approval (step 10). Implementation beyond the current hole's contract takes scope from a later stage.

### The census

`grep -rn "HOLE(" <src>`, over the source and test directories, lists the open holes and the skipped acceptance tests. It is the to-do list and the progress bar. After any interruption, the census and the ledger together show the state of the work. A hole returns mock data instead of failing, so the census is the only record of an unfinished body. The change is done when the census returns nothing, the full suite is green, and every stage has a completed line in the ledger.

## Code comments

A code comment describes the thing or its behavior in fewer than fifteen words. The reason for a step lives in the design, on its box's purpose line. A function's doc comment starts with the function's name and a verb, and the box's action line says the same in the imperative ("Analyze reads the source" beside "Read the source"), so either one can be written from the other.

## Rationalizations

| Excuse | Reality |
|---|---|
| "I'll update the diagram once the code settles" | The diagram comes before the code. With a stale diagram, the reviewer judges, and the score measures, a design missing from the code. |
| "The reviewer can judge the design from the diagram alone, and the skeleton can wait for approval" | A diagram without its code hides the signatures, and no box links to anything. The holes are what the reviewer checks, so the diagram and its skeleton go to the plan gate together. |
| "The new level's functions are small, so I'll fill them before the boundary" | Then your human partner never sees the new level's signatures and contracts before its bodies exist. End the stage with the holes, and fill them in the next one. |
| "The stage's acceptance check needs the new functions filled" | The plan put a split and its fill in one stage. Split the stage: the first ends with the new level's holes and its smoke test, and the second fills the holes and runs that check. |
| "I'll draw the lower levels now, so the reviewer sees the whole design" | A box without a function misleads the reviewer, and the review page reports its missing declaration. A child diagram comes with the fill that creates its functions. |
| "This new arrow is small, so I'll mention it at the boundary" | After the plan gate, a change to a frozen diagram or to the vocabulary needs approval, shown with its code. Ask first. |
| "Changing the top diagram is cleaner than working inside this box" | A change inside a body beats a change to an interface, and a change deep in the design beats one near the top. Take the first that does the job. |
| "The arrow can show the updated value" | An arrow claims that the step returns the value. A change in place is a store write. |
| "One weighted score is easier to track" | Invented weights hide the trade-offs between columns. Track each column, and explain every rise. |
| "The review page is overhead for this stage" | The reviewer reads the page first, so a boundary without the page is incomplete. |
| "Everyone knows what these terms mean" | Without definitions, the reviewer and the agent read one term in two different ways. |
| "I'll just implement this whole subtree while I'm here" | Work past the current hole's contract belongs to a later stage. Declare new holes for that work and return to the current hole. |
| "The build can stay red while I edit these three files together" | While the build is red, the typechecker cannot report the required type of each hole. |
| "I know what type this needs, no need to run the checker" | Read the typechecker's output for each hole. A body written from memory can typecheck and still do the wrong thing. |
| "I'll add tests once the holes are all filled" | Without its tests, a wrong fill looks the same as a right one. |
| "A panic marks an unfinished body more clearly than mock data" | A panic fails every test that runs through it, so the suite stays red until the last fill. The census finds every hole by its tag. |
| "This acceptance test fails, so I'll skip it with a HOLE tag" | A tagged skip is only for a test with an open hole on its path. A test that fails with its holes filled shows a bug, and a skip hides it. |
| "The fill changed the smoke test's output, so I'll update the expected value" | The smoke test asserts what a correct program produces. Either the fill is wrong or the test asserted mock data: find the wrong one, and propose any change to the test for sign-off. |
| "This stage is too simple to need a review halt" | The halt costs little on a simple stage, and it still catches a wrong assumption. |
| "The cap is arbitrary, and this 600-line diff is coherent" | A 600-line diff exceeds the 400-line cap. Split the stage. |
| "I'll quietly adjust the plan, since explaining takes longer" | A plan out of step with the code misleads every later stage. Amend the plan first. |
| "Tests passed earlier in the session" | A green run proves only the state of the tree at that run. Run the suite fresh at the boundary. |

## Red flags

Stop and reread the laws if you catch yourself thinking any of these: "I'll draw the diagram after the code". "This mutation can go on the arrow". "The top diagram can show every store". "This box can call two packages". "I'll skip the score this stage". "I'll skip the skeleton for this one". "Let me rough out several bodies and fix the types later". "This stub does not need a name or a contract". "I'll batch the whole fill into one commit at the end". "The signature is wrong, but changing it now is faster than a plan diff". "I'll add a box now and explain it at the boundary". "I'll sketch the lower levels now and write their functions later". "I'll get the design approved first and write the skeleton after". "I'll draw the child diagram and fill its boxes in the same stage". "This hole can panic until I fill it". "I'll skip this failing test for now". "I'll change the smoke test to match the new output". Each of these means: return to the plan and the diagram, restore the green build, and halt now, with a ledger line on the cause.
