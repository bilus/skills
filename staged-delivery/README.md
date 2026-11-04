# Staged delivery

A delivery workflow for AI coding agents built on Martin Fowler's
[example of preparatory refactoring](https://martinfowler.com/articles/preparatory-refactoring-example.html):
write the acceptance test for the end state, refactor until the change has
an obvious place to go, then make the change. The agent plans once, the
human approves once, and the rest runs without approvals: stages of
100-line commits, each leaving a running application and a green suite,
with a sub-agent doing exploratory testing at points the plan marks.

## Why this approach

Review capacity is the bottleneck with agents. This skill spends it in two
places and nowhere else. The plan is the first: a requirement in EARS form,
a current and a target data flow diagram, the acceptance tests, and a table
of stages each described in under 20 words with one sentence on why it sits
where it does. The commit log is the second: every commit is one step, at
most 100 changed lines, with a summary under 15 words as its whole message,
so a reviewer can read the history the way they would read a colleague's.

Preparatory refactoring is what keeps the commits small. The difference
between the current and the target diagram, minus the change itself, is the
list of refactor stages; by the time the change stage comes, the change is
one small function. Fowler's moves (extract a method, insert a no-op into
the call chain, pin current behavior with a temporary test, split and
rejoin the data) are the ways a step stays inside its budget while the
suite stays green.

## What the agent will do

1. Take the requirement from `REQUIREMENTS.yaml`, from the message, or
   draft it in EARS form and mark it drafted.
2. Draw the current and the target data flow of each flow the requirement
   touches, and derive the refactor stages from their difference.
3. Write one acceptance test per requirement sentence, watch it fail, skip
   it.
4. Write `docs/plans/<date>-<slug>.md`: requirement, diagrams, tests, the
   stage table and the steps under each stage, with their estimated line
   counts. Send the prose through Codex. Show the plan and stop. This is the
   only approval.
5. Deliver stage by stage, step by step. Each step: `scripts/step-size.sh`
   under 100, build and suite green with the output quoted, one commit with
   the summary line only and no attribution.
6. At each review point, start a sub-agent with
   `references/review-brief.md`. It runs the application, runs the
   acceptance tests, tries inputs beyond the happy path and drives any UI
   through the browser. Blockers and defects inside the requirement become
   a review fixes stage of new commits; everything else becomes a
   follow-up in the plan.
7. Send docstrings and plan text through Codex at each review point and at
   the end. Finish with the acceptance tests passing unskipped and the list
   of follow-ups.

## What you do

- Approve the plan. Spend your time on the stage table: the order and the
  "why here" column are where a wrong plan shows.
- Read the commits as they land, by their summaries.
- Answer when the agent stops, which happens only when the requirement
  turns out to be wrong or a step fails three times.

## Files

- `SKILL.md`: the rules.
- `references/ears.md`: EARS patterns and the `REQUIREMENTS.yaml` shape.
- `references/plan-template.md`: the plan document.
- `references/review-brief.md`: the sub-agent brief for a review point.
- `references/codex.md`: the Codex MCP call and prompt for prose review.
- `scripts/step-size.sh`: counts changed lines for the step about to be
  committed, excluding lockfiles, generated code, the plan folder and
  diagrams.

## Requirements

Codex available as an MCP server, for the prose review. Without it the
skill says so once and continues. A `REQUIREMENTS.yaml` managed by `phb` is
used when present and never required.
