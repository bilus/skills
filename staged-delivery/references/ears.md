# EARS requirements

EARS (Easy Approach to Requirements Syntax) gives a requirement a fixed
shape so that the trigger, the system and the response are each in one
place. One sentence is one requirement and maps to one acceptance test.

## The five patterns

| Pattern | Shape | Example |
| --- | --- | --- |
| Ubiquitous | The <system> shall <response>. | The exporter shall write one CSV row per order. |
| Event-driven | When <trigger>, the <system> shall <response>. | When an order is paid, the exporter shall append its row within one minute. |
| State-driven | While <state>, the <system> shall <response>. | While the export file is locked, the exporter shall queue new rows. |
| Unwanted behavior | If <condition>, then the <system> shall <response>. | If the CSV directory is missing, then the exporter shall create it and log a warning. |
| Optional feature | Where <feature>, the <system> shall <response>. | Where the tenant has invoicing enabled, the exporter shall add the invoice number column. |

Patterns combine when needed: "While the file is locked, when an order is
paid, the exporter shall queue the row."

## Rules for a sentence

- Name the system by the word the data flow diagram uses for it.
- The trigger and the response are observable from outside the system. "The
  parser shall use a state machine" is an implementation note, not a
  requirement.
- Use "shall" and nothing else; no "should", "must" or "will".
- One response per sentence. A sentence with "and" between two responses is
  two sentences.
- The response is testable: a reader can write the acceptance test from the
  sentence alone.
- Quantities are concrete: "within one minute", not "quickly".

## Drafting

When the user gave no requirement, write one sentence per behavior they
described, using the patterns above, and mark each one "drafted" in the
plan. Do not add behaviors the user did not mention; an edge case the user
did not ask for is a follow-up, not a requirement. The user approves the
drafted sentences together with the plan.

## REQUIREMENTS.yaml

Projects that track requirements keep them in `REQUIREMENTS.yaml` at the
repository root (or the project root in a monorepo), managed by the `phb`
CLI. Shape of an entry:

```yaml
reqs:
  - issue: https://github.com/org/repo/issues/12345
    parent: PRD-42
    title: Short descriptive name
    requirement: |
      When the system receives X, the system shall do Y so that Z.
    notes: |
      Implementation guidance or rationale.
```

When the requirement comes from this file, quote it in the plan by title and
issue. Never edit the file. Tag each acceptance test with a comment
`REQ(#NNNN)` where NNNN is the issue number, on the line before the test
function, so `phb reqs sync` links the test to the issue:

```go
// REQ(#12345): paid order appends a CSV row within one minute
func TestPaidOrderAppendsRow(t *testing.T) { ... }
```
