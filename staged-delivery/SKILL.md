---
name: staged-delivery
description: Plan and deliver any coding task that spans more than one edit, such as a feature, a behavior change, a new endpoint, a migration, or a refactor with a goal. Use it whenever the user asks to build, add, implement, change, extend or migrate something in code, mentions a plan, milestones, stages, steps, stacked PRs, feature flags, EARS, acceptance tests or preparatory refactoring, or asks for small reviewable commits. It requires an EARS requirement, writes acceptance tests first, refactors to make the change easy, then delivers the change as milestones of stacked PR-sized stages under 400 lines, each a series of single-purpose commits, with incomplete features behind a feature flag and exploratory reviews by a sub-agent at planned points. Not for a one-line fix or a spike the user calls throwaway.
---

# Staged delivery

Deliver a change the way Fowler describes preparatory refactoring: write the
acceptance test for the end state, refactor until the change has an obvious
place to go, then make the change. Order the work as milestones, each a win
visible to the user: the simplest working version first, including its
risky parts, then one increment at a time. Every stage is one pull request
under 400 changed lines that leaves the application running and the suite
green, with incomplete features behind a feature flag. Every step is one
commit that does one thing, with a one-line summary. The user approves the
plan once and nothing else.

## Vocabulary

Use these words and no synonyms, in the plan, in commit messages, in
docstrings and when talking to the user. Node labels from the data flow
diagrams (section 2) are the nouns for everything in the code they name.

| Word | Meaning |
| --- | --- |
| requirement | One EARS sentence. The reason for the work. |
| acceptance test | A test that fails before the change and passes after it. One per requirement sentence. |
| flow | One linear data flow through the code, drawn as a data flow diagram. |
| risk | A part that is hard to build or could change the design most. |
| plan | The document in `docs/plans/` that the user approves. |
| milestone | A win: a new capability for the user, or a settled risk. A group of stages. |
| stage | One pull request: a contiguous range of steps, under 400 changed lines, that leaves the application running and the suite green. |
| refactor stage | A stage that preserves behavior. Existing tests are not edited. |
| infrastructure stage | A stage of database migrations, configuration changes or dependency updates, and nothing else. |
| feature stage | A stage that adds behavior. |
| fix stage | A stage that fixes a bug. |
| feature flag | A switch, off by default, that keeps an incomplete feature inactive in production. |
| step | One commit of one kind, at most 100 changed lines, one summary line. |
| review point | A stage after which a sub-agent reviews by exploratory testing. |
| review fixes | A fix stage added after a review point to act on its findings. |
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
  So is a reasonable generalization or simplification of existing code that
  makes the change simpler.
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

Each milestone unskips the tests of its requirement sentences, at the latest
in its last stage. A test of a feature behind a feature flag turns the flag
on. The work is done when every acceptance test passes unskipped and the
full suite is green, with the output quoted.

## 4. Plan

Write `docs/plans/<date>-<slug>.md` from `references/plan-template.md`. It
holds the requirement, the diagrams, the acceptance tests, the risks, the
milestone table, the stage table and the steps of each stage. Keep it
short: a reader who has not seen the code should understand what will
happen and why from the plan alone.

Rules for the risks:

- List the parts that are hard to build or could change the design most,
  each with one sentence on why.
- Each risk names the milestone that settles it. A performance risk is
  settled by a benchmark.

Rules for the milestone table:

- Each milestone is a win, described as what the user can do after it:
  "User can select a line and move it". A settled risk is a win too:
  "Selecting one shape among 5,000 takes under 16 ms".
- Each milestone is the simplest version of its win, such as two buttons
  that add shapes and a drag to move them, instead of the full UI.
  Complexity comes in later milestones.
- Order the milestones in this priority:
  1. The simplest version of the core that works end to end, including
     every risky part needed for a minimal working product. A risky part
     may start as a rough proof of concept with mocked parts behind the
     feature flag.
  2. Then one milestone at a time, each chosen by business value, by
     architectural drivers or by risk.
- Each milestone carries one sentence on why it sits at its place in the
  order, naming the reason: the core, value, architecture or risk.
- Each milestone names the requirement sentences whose acceptance tests pass
  at its end, and lists its stages. A milestone holds any number of stages.

Rules for the stage table:

- A stage is one pull request, under 400 changed lines as
  `scripts/step-size.sh` counts them. This is a hard limit: when the
  estimates of a stage's steps add up to 400 or more, split the stage now.
- Stages stack: each one builds on the stage before it, adds one increment,
  and is reviewable on its own.
- A stage has one kind: refactor, infrastructure, feature or fix. A
  feature's preparatory refactoring goes in a refactor stage before it.
  Migrations, configuration changes and dependency updates go in an
  infrastructure stage, never in a feature stage.
- Within a milestone, draw the stages from this list, in an order suited to
  the change. Not every milestone needs every item.
  - preparatory refactoring that makes the change simple;
  - infrastructure: a migration, a configuration value, a dependency;
  - top down: interfaces changed for the feature, returning mock data
    behind the feature flag;
  - bottom up: the supporting functions, types and packages;
  - the feature in layers: the data model, then the business logic, then
    the API or the UI;
  - glue: wire the pieces together and unskip the milestone's acceptance
    tests;
  - a refactor for cohesion and readability once the milestone runs end to
    end.
- Each stage is described in under 20 words, written for someone reviewing
  the code. If the description needs "and", or exceeds 20 words, it is two
  stages.
- Each stage carries one sentence on why it sits where it does.
- Mark review points. Place one after the first stage that changes visible
  behavior, one after the last stage, and never let more than three stages
  pass without one.

Rules for the steps under each stage:

- Each step has one kind: refactor, cleanup (comments, documentation,
  renames), test, feature, fix, or infrastructure (a migration, a
  configuration change, a dependency update). A step of two kinds is two
  steps.
- Each step has a summary of at most 15 words that names its change and,
  when it prepares a later step, the purpose: "Extract formatRow from
  Exporter.Write to reuse it for the invoice column". The summary is the
  commit message, so it must stand on its own.
- Each step has an estimated line count. Over 100, split it now.

Before showing the plan, send its prose through Codex (section 9). Then show
the plan to the user and stop. This is the only approval. After it, do not
ask for approvals or confirmations; the plan is the contract.

## 5. Stages

A stage has one kind and never mixes them. In a refactor stage every commit
preserves observable behavior and no existing test is edited, so the suite
is the proof. In a feature or fix stage behavior changes and tests are added
or unskipped to pin it. An infrastructure stage holds migrations,
configuration changes or dependency updates and nothing else.

After every stage the application builds, runs, and has a working UI. Until
its last stage, a feature's new behavior stays behind a feature flag, off by
default: the code is in production but not active. Use the project's flag
mechanism; without one, add a configuration value in an infrastructure
stage. Behind the flag a stage may take reductions: a screen fed by mock
data, an endpoint that handles the happy path only, a proof of concept that
skips edge cases. The plan names the flag, lists the reductions of each
stage, and names the later stage or follow-up that removes each reduction
and the flag.

A stage is done when, in one message, you show:

- the full suite run fresh, output quoted;
- `sh scripts/step-size.sh <base>..HEAD` under 400, output quoted, where
  `<base>` is the commit before the stage's first step;
- the application started and the stage's behavior exercised with the flag
  on, when the stage touches runtime behavior;
- the docstring rule (section 8) applied to everything the stage added.

Each stage is one pull request, stacked on the stage before it. Open, push
and merge pull requests only on the user's or the repository's instruction;
otherwise commit only, and the stage's commit range in the plan stands for
its pull request. When the repository uses a patch stack (`gps`), a stage
is a contiguous range of patches.

## 6. Steps

A step is one commit and is self-contained: the build and the full suite
are green after it, and it makes sense on its own without a later step.
Before committing:

1. Run `sh scripts/step-size.sh`. It prints insertions plus deletions per
   file and the total, excluding generated files, lockfiles, the plan and
   the diagrams. The total must be 100 or less. Name any other exclusion in
   the plan.
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

Order: in a refactor stage, the tests that pin current behavior come first,
then the structural steps. In a feature or fix stage, the change comes
first, then the test that proves it.

## 7. Review points

At each review point start a sub-agent with the brief in
`references/review-brief.md`, filled in with the requirement, the acceptance
tests, the plan, the stages covered since the last review, the feature flag
and how to turn it on, and the command that runs the application. The
sub-agent reviews by exploratory testing: it runs the application, runs the
acceptance tests, tries inputs beyond the happy path and, when there is a
UI, drives it through the browser tools. It reads the diff second, not
first. It writes `docs/plans/<date>-<slug>/review-<n>.md` with findings
graded blocker, defect or note.

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

When a stage reaches 400 lines, end it at its last step under the limit and
start the next stage with the step that crossed it; record the split the
same way. At the end of each milestone, check the order of the remaining
milestones against business value, architectural drivers and risk, given
the findings of the finished milestone, and record any change of order the
same way.

Stop and wait for the user only in two cases: the requirement itself proves
wrong or impossible, or three consecutive attempts at one step fail. In both
cases say what was tried.

## 11. Reporting

Claim green only with the command and its output in the same message. Keep
measured results (you ran it) apart from inferred ones (you read it).

While working, one line per committed step: `step 2.3: <summary> (<n>
lines)`. At the end of a stage, its total from `scripts/step-size.sh`. At
the end of a milestone, its win and its unskipped acceptance tests, output
quoted. At a review point, the path of the findings file and the count by
grade. At the end: the acceptance tests passing, output quoted, and the list
of follow-ups. No narrative of what was done; the commits are the narrative.

## Files

- `references/ears.md`: EARS patterns, rules for a good sentence,
  `REQUIREMENTS.yaml` shape.
- `references/plan-template.md`: the plan document.
- `references/review-brief.md`: the sub-agent brief for a review point.
- `references/codex.md`: the Codex MCP call and prompt for prose review.
- `scripts/step-size.sh`: counts changed lines for the step about to be
  committed, or for a stage's range of commits.
