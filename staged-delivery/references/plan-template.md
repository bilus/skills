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

## Risks

| Risk | Why it is risky | Milestone |
| --- | --- | --- |
| An invoice number lookup per row may slow exports of 100,000 orders | The numbers live in another table | 1 |

## Why this order

<One short paragraph. Why the first milestone comes first, which
refactoring makes the change easy, and why the feature stages wait for it.>

## Milestones

| # | Win (what the user can do after it) | Requirement | Why here (one sentence) | Stages |
| --- | --- | --- | --- | --- |
| 1 | Invoicing tenants see the invoice number column in their export | 1 | Core: the column is the request, and its per-row lookup is the biggest risk. | 1-5 |
| 2 | Invoicing tenants choose the position of the invoice number column | 2 | Value: the next request from invoicing tenants. | 6-7 |

Feature flag: `invoice_column`, off by default. Stage 3 adds it with the
project's flag mechanism; stage 5 removes it.

## Stages

| # | Kind | Description (under 20 words) | Why here (one sentence) | Review |
| --- | --- | --- | --- | --- |
| 1 | refactor | Extract row formatting into a function called from the exporter | The change needs one place to add a column. | |
| 2 | refactor | Add a no-op column hook between formatting and writing | Gives the change a slot while output stays the same. | |
| 3 | feature | Add the invoice number column behind the flag, benchmarked on 100,000 orders | The lookup is the risk, so it comes before the real settings. | yes |
| 4 | feature | Read invoicing settings from the settings service instead of a mock | Stage 3 used mock settings to reach the lookup first. | |
| 5 | feature | Remove the invoice_column flag so invoicing tenants get the column | The column is complete and fast enough. | yes |
| 6 | feature | Add the column position to the tenant settings model | Stacked: the data model comes before the logic that reads it. | |
| 7 | feature | Place the invoice number column at the position from the tenant settings | Stacked on the model of stage 6. | yes |

Reductions taken: stage 3 reads invoicing settings from a mock behind the
flag; stage 4 removes it. Edge case X is a follow-up.

## Steps

### Stage 1

| # | Summary (at most 15 words) | Kind | Est. lines |
| --- | --- | --- | --- |
| 1.1 | Add skipped acceptance tests for the invoice number column | test | 40 |
| 1.2 | Extract formatRow from Exporter.Write to add columns in one place | refactor | 60 |

### Stage 2

| # | Summary | Kind | Est. lines |
| --- | --- | --- | --- |
| 2.1 | Add addColumns hook returning its input, called from Write | refactor | 30 |
| 2.2 | Pin current output for invoicing tenants with a temporary test | test | 35 |

### Stage 3

| # | Summary | Kind | Est. lines |
| --- | --- | --- | --- |
| 3.1 | Add invoice number in addColumns behind the invoice_column flag, replacing the pin test | feature | 60 |
| 3.2 | Benchmark an export of 100,000 orders with the invoice column on | test | 40 |
| 3.3 | Unskip invoice column acceptance tests with the flag on | test | 10 |

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
- Keep the step number and the summary in the first two columns of a step
  table; tools read them by position.
- The estimated lines are insertions plus deletions as
  `scripts/step-size.sh` counts them. An estimate over 100 means the step is
  split before the plan is shown. A stage whose estimates add up to 400 or
  more is split the same way.
- The "Why here" columns are for the user, who has not read the code. They
  are the part of the plan to spend effort on.
- The stage table is also the task list: tick off stages as they land by
  appending the commit range to the row, for example `3: a1b2c3..d4e5f6`.
