# Plan template

Copy this into `docs/plans/<date>-<slug>.md`. Keep every section; write
"none" where a section has nothing. Diagrams and review findings live in
`docs/plans/<date>-<slug>/`.

```markdown
# <Title: the change in one line>

Requirement source: REQUIREMENTS.yaml "<title>" (#NNNN) | user message | drafted
Codex pass: done <date> | not run

## Requirement

1. When <trigger>, the <system> shall <response>.   (drafted)
2. If <condition>, then the <system> shall <response>.   (drafted)

## Flows

### <flow name>

Current:

![current](<date>-<slug>/current-<flow>.svg)

Target:

![target](<date>-<slug>/target-<flow>.svg)

What the target adds that is not the change itself: <one line per element>.

Architecture check: <for each process that fails the one-thing, one-direction
or one-writer question, which stage or follow-up handles it, or "none">.

## Acceptance tests

| Test | Requirement | Status |
| --- | --- | --- |
| Test<Name> | 1 | skipped, fails as expected (output below) |

```
<quoted output of the first run, failing>
```

## Why this order

<One short paragraph. Which refactoring makes the change easy and why the
change stage waits for it.>

## Stages

| # | Kind | Description (under 20 words) | Why here (one sentence) | Review |
| --- | --- | --- | --- | --- |
| 1 | refactor | Extract row formatting into a function called from the exporter | The change needs one place to add a column. | |
| 2 | refactor | Add a no-op column hook between formatting and writing | Gives the change a slot while output stays the same. | |
| 3 | change | Add the invoice number column for invoicing tenants | The slot exists; the change is one function. | yes |
| 4 | change | Replace mock tenant settings with the settings service | Stage 3 used mock data to stay within the happy path. | yes |

Reductions taken: stage 3 reads tenant settings from a mock; stage 4
removes it. Edge case X is a follow-up.

## Steps

### Stage 1

| # | Summary (under 15 words) | Est. lines |
| --- | --- | --- |
| 1.1 | Add skipped acceptance tests for invoice number column | 40 |
| 1.2 | Extract formatRow from Exporter.Write | 60 |

### Stage 2

| # | Summary | Est. lines |
| --- | --- | --- |
| 2.1 | Add addColumns hook returning its input, called from Write | 30 |
| 2.2 | Pin current output for invoicing tenants with a temporary test | 35 |

### Stage 3

| # | Summary | Est. lines |
| --- | --- | --- |
| 3.1 | Add invoice number in addColumns, replacing the temporary pin test | 45 |
| 3.2 | Unskip invoice column acceptance tests | 10 |

## Changes to the plan

none

## Follow-ups

none

## Reviews

none yet
```

Notes on filling it in:

- Step summaries are commit messages. Write them in the imperative, without
  a trailing period, and make them specific: "Extract formatRow from
  Exporter.Write", not "Refactor exporter".
- The estimated lines are insertions plus deletions as
  `scripts/step-size.sh` counts them. An estimate over 100 means the step is
  split before the plan is shown.
- The "Why here" column is for the user, who has not read the code. It is
  the part of the plan to spend effort on.
- The stage table is also the task list: tick off stages as they land by
  appending the commit range to the row, for example `3: a1b2c3..d4e5f6`.
