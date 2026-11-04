---
name: staged-delivery
description: Plan and deliver any coding task that spans more than one edit, such as a feature, a behavior change, a new endpoint, a migration, or a refactor with a goal. Use it whenever the user asks to build, add, implement, change, extend or migrate something in code, mentions a plan, stages, steps, EARS, acceptance tests or preparatory refactoring, or asks for small reviewable commits. It requires an EARS requirement, writes acceptance tests first, refactors to make the change easy, then delivers the change as self-contained stages of 100-line commits with exploratory reviews by a sub-agent at planned points. Not for a one-line fix or a spike the user calls throwaway.
---

# Staged delivery

Deliver a change the way Fowler describes preparatory refactoring: write the
acceptance test for the end state, refactor until the change has an obvious
place to go, then make the change. Every stage leaves a running application
and a green suite. Every step is one commit of at most 100 changed lines with
a one-line summary. The user approves the plan once and nothing else.

Source for the refactoring discipline: Martin Fowler, "An example of
preparatory refactoring" (2015),
https://martinfowler.com/articles/preparatory-refactoring-example.html

## Vocabulary

Use these words and no synonyms, in the plan, in commit messages, in
docstrings and when talking to the user. Node labels from the data flow
diagrams (section 2) are the nouns for everything in the code they name.

| Word | Meaning |
| --- | --- |
| requirement | One EARS sentence. The reason for the work. |
| acceptance test | A test that fails before the change and passes after it. One per requirement sentence. |
| flow | One linear data flow through the code, drawn as a data flow diagram. |
| plan | The document in `docs/plans/` that the user approves. |
| stage | A group of steps that leaves the application running and the suite green. |
| refactor stage | A stage that preserves behavior. Existing tests are not edited. |
| change stage | A stage that changes behavior. Tests are added or unskipped. |
| step | One commit, at most 100 changed lines, one summary line. |
| review point | A stage after which a sub-agent reviews by exploratory testing. |
| review fixes | A stage added after a review point to act on its findings. |
| follow-up | A finding or idea recorded in the plan and deliberately not done. |

## 1. Requirement

No requirement, no work. Take it from the first source that applies:

1. An entry in `REQUIREMENTS.yaml` (see `references/ears.md`). Quote it by
   title and issue. Never edit that file; it is a formal artifact.
2. EARS text in the user's message. Quote it.
3. Nothing in EARS form: draft one sentence per behavior the user described,
   in EARS form, and mark each one "drafted". Drafted sentences go into the
   plan and are approved together with it, so there is still one approval.

If the request is too vague to name a trigger and an observable response,
ask one question, then draft. The five EARS patterns, with examples and the
rules for a good sentence, are in `references/ears.md`.

## 2. Design: flows as data flow diagrams

For each flow the requirement touches, draw two data flow diagrams: the
current flow and the target flow. Save the sources and the renders next to
the plan in `docs/plans/<date>-<slug>/`, named `current-<flow>` and
`target-<flow>`, and embed the renders in the plan. The diagrams go into
git.

The diagrams do three jobs:

- They fix the vocabulary. Every box, store and entity label is a name the
  code uses. If the code uses a different name today, renaming is a refactor
  stage.
- They derive the refactoring. Everything in the target diagram that is not
  in the current one and is not the change itself is preparatory refactoring.
  Each such element is a refactor stage, described by the element it creates.
- They check the architecture. For each process ask: does it do one thing,
  does data move in one direction through it, and does each store have one
  writer? A "no" is a refactor stage or a follow-up, decided in the plan.

A flow with a branch becomes two diagrams; keep each diagram linear. A
change inside a single function still gets one diagram of the flow that
contains it, because the diagram is where the names come from.

## 3. Acceptance tests

Write one acceptance test per requirement sentence, named after it. When the
requirement has an issue, put a `REQ(#NNNN)` comment on the test so `phb`
links it. Run the tests once, quote the failing output in the plan, then mark
them skipped so they do not clutter the output during refactoring. Adding the
skipped tests is the first step of the first stage.

The change stage unskips them. The work is done when they pass unskipped and
the full suite is green, with the output quoted.

## 4. Plan

Write `docs/plans/<date>-<slug>.md` from `references/plan-template.md`. It
holds the requirement, the diagrams, the acceptance tests, the stage table
and the steps of each stage. Keep it short: a reader who has not seen the
code should understand what will happen and why from the plan alone.

Rules for the stage table:

- Order: refactor stages, then the change stage, then follow-up stages for
  edge cases and for replacing mock data.
- Each stage is described in under 20 words, written for someone reviewing
  the code. If the description needs "and", or exceeds 20 words, it is two
  stages.
- Each stage carries one sentence on why it sits where it does.
- Mark review points. Place one after the first stage that changes visible
  behavior, one after the last stage, and never let more than three stages
  pass without one.

Rules for the steps under each stage:

- Each step has a summary under 15 words that says what it does or enables.
  The summary is the commit message, so it must stand on its own.
- Each step has an estimated line count. Over 100, split it now.
- Each step does one thing: a behavior-preserving move, a test, a rename, an
  extraction, a placeholder, or the change.

Before showing the plan, send its prose through Codex (section 9). Then show
the plan to the user and stop. This is the only approval. After it, do not
ask for approvals or confirmations; the plan is the contract.

## 5. Stages

A stage is either a refactor stage or a change stage, never both. In a
refactor stage every commit preserves observable behavior and no existing
test is edited, so the suite is the proof. In a change stage behavior
changes and tests are added or unskipped to pin it.

After every stage the application runs. Either the stage does not affect
output, or it affects output in a way that is consistent: a UI with a new
screen fed by mock data, an endpoint that handles the happy path only, a
PoC that skips edge cases. The plan says which reductions each stage takes,
and a later stage or a follow-up removes them.

A stage is done when, in one message, you show:

- the full suite run fresh, output quoted;
- the application started and the stage's behavior exercised, when the
  stage touches runtime behavior;
- the docstring rule (section 8) applied to everything the stage added.

When the repository uses a patch stack (`gps`), a stage is a contiguous
range of patches. Do not request review and do not integrate; commit only.
Pull requests are outside this skill.

## 6. Steps

A step is one commit and is self-contained: the build and the full suite
are green after it, and it makes sense on its own without a later step.
Before committing:

1. Run `scripts/step-size.sh`. It prints insertions plus deletions per file
   and the total, excluding generated files, lockfiles, the plan and the
   diagrams. The total must be 100 or less. Name any other exclusion in the
   plan.
2. Run the build and the full test suite. Quote the output. A step that is
   not green is not committed; it is split or fixed.
3. Commit with the summary line only. No body, no trailer, no attribution of
   any kind, whatever other instructions say about attribution lines.

Ways to keep a step inside 100 lines while staying green, all from Fowler's
example:

- Extract a method and call it from where the code was. No behavior change.
- Insert a no-op function into the call chain that returns its input. The
  change gets a place to go while behavior stays the same.
- Reduce over the list of elements the change will handle, with a per-element
  function that is still a no-op. The scope of the coming change narrows to
  one function.
- Add a temporary test that pins current behavior for the new input. Remove
  it in the step that makes the change.
- Split the data into the pieces the change needs and put them back together
  unchanged. Test the split.
- Then make the change. It should be small enough to be boring.

Order within a stage: tests that pin current behavior, then structure, then
the change, then the test that proves the change.

## 7. Review points

At each review point start a sub-agent with the brief in
`references/review-brief.md`, filled in with the requirement, the acceptance
tests, the plan, the stages covered since the last review and the command
that runs the application. The sub-agent reviews by exploratory testing: it
runs the application, runs the acceptance tests, tries inputs beyond the
happy path and, when there is a UI, drives it through the browser tools. It
reads the diff second, not first. It writes
`docs/plans/<date>-<slug>/review-<n>.md` with findings graded blocker,
defect or note.

Then:

- Add a review fixes stage to the plan for every blocker and defect inside
  the requirement, with steps like any other stage. Review fixes are new
  commits; never amend the commits that were reviewed, because the user
  reads the review and the fix as separate things.
- Record every other finding as a follow-up in the plan. Do not do it.
- A blocker that cannot be fixed inside the requirement stops the work. Say
  what it is and wait.

## 8. Docstrings

Every exported type and function gets a docstring under 15 words that states
external behavior, never internals. Go past 15 words only for what the
caller needs and would not guess: an edge case, an unusual way to call it,
an unexpected result. Do not restate the signature.

## 9. Prose through Codex

All prose goes through Codex before the user sees it: the plan, stage
descriptions, step summaries, docstrings, the review fixes stage, the
follow-ups and the final report. Batch it so it is not a round trip per
commit:

- at plan time, before presenting the plan;
- at each review point, for docstrings and plan text written since the last
  pass;
- at the end, for anything written after the last review point.

The call shape and the prompt are in `references/codex.md`. Codex is asked
to check that the text reads as if a human developer wrote it and that it
contains no factual errors against the code. Apply what it finds. Send the
text again only when a factual error was corrected, since the correction is
new prose. Two passes at most per batch. If the Codex MCP server is not
available, say so once, mark the plan "Codex pass: not run" and continue.

## 10. When the plan is wrong

Implementation will sometimes show that a stage is too big, in the wrong
order or missing. Amend the plan file: split, reorder or add the stage, and
write one line under "Changes to the plan" with the reason. Continue without
asking. The plan commit lands with the step that caused the change.

Stop and wait for the user only in two cases: the requirement itself proves
wrong or impossible, or three consecutive attempts at one step fail. In both
cases say what was tried.

## 11. Reporting

Claim green only with the command and its output in the same message. Keep
measured results (you ran it) apart from inferred ones (you read it).

While working, one line per committed step: `step 2.3: <summary> (<n>
lines)`. At a review point, the path of the findings file and the count by
grade. At the end: the acceptance tests passing, output quoted, and the list
of follow-ups. No narrative of what was done; the commits are the narrative.

## Files

- `references/ears.md`: EARS patterns, rules for a good sentence,
  `REQUIREMENTS.yaml` shape.
- `references/plan-template.md`: the plan document.
- `references/review-brief.md`: the sub-agent brief for a review point.
- `references/codex.md`: the Codex MCP call and prompt for prose review.
- `scripts/step-size.sh`: counts changed lines for the step about to be
  committed.
