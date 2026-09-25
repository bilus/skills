# Hole-driven delivery

A staged-delivery workflow for AI coding agents, adapted from typed-hole programming in Haskell, Idris and Agda (and its cousins, `fatalError()` in Swift and `todo!()` in Rust), with a data flow design that leads the code.

Not to be confused with [jhhuh's hole-driven-delivery skill](https://github.com/jhhuh/hole-driven-delivery-skill), which is the inner technique on its own: an autonomous compiler loop (write holes, compile, read diagnostics, fill the most constrained hole, repeat) without a plan, git, tests or review gates. This skill is the delivery process around that idea. The agent agrees the requirements, the vocabulary and a leveled data flow design with you, commits a compiling skeleton of the change with every body a named `HOLE(id): contract` stub, and then fills the holes one at a time, outside-in, with the build green after every move. It halts for your review at every stage boundary, and each review starts from a page that shows the design, links into the code, and a score.

## Why this approach

Review capacity is the bottleneck with agents, and the hardest thing to review late is architecture. This skill moves the architecture review to the front and makes it cheap in two steps. First comes the design: a set of dfd diagrams, one per level, in which every box names the functions that do the step and says why the later steps need it, every arrow names its type, and every change of state in place is a write to a store. Then comes the skeleton: a few hundred lines of declarations that implement exactly the functions the diagrams name. After that, the compiler becomes the agent's oracle, and `grep -rn "HOLE("` is a progress bar that survives any loss of agent context.

The design also gets measured. The bundled dfdmetrics tool reports how much state the design shares, at which level, and how far each read sits from its write, and it flags boxes that span two packages. Its score card goes into a history file at every review, so you can see whether a stage made the design better or worse.

## What the agent will do

1. Write `docs/plans/<date>-<feature>.md` with numbered requirements, then a vocabulary of the terms the change introduces, reviewed by a sub-agent for redundancy.
2. Draw the leveled design in `docs/flow.dfd` and its child diagrams, name a system metaphor and check the rules a reader would infer from it, and list every planned declaration.
3. Measure the design, record the score, build the review page, and stop for your approval of the plan.
4. Commit the skeleton: the planned declarations with every body a tagged hole that typechecks and fails loudly if executed. The existing suite stays green. It stops for review, and the approved signatures are frozen.
5. Fill one group of holes per stage: one hole open at a time, the build never red, each fill landing with the tests that pin its contract. The budget is 100 to 300 lines per stage, with a hard cap of 400. A change of shape starts in the diagram, as a plan diff.
6. At every boundary, quote a fresh full-suite run and the hole census, redraw the design to match the code, append a score row, rebuild the review page, write the handoff, and halt.
7. After your final sign-off, reshape the history to one commit per stage, with messages that describe the change and not the method.

## What you do

- Approve the requirements, the vocabulary, the design and the stages at the plan gate. The design and the skeleton are where your review time pays most.
- At each halt, open the review page, check the score's changes, and answer the handoff's one review question.
- Approve any proposed signature change, design change or change to an existing test. The agent may not make these alone.
- In dynamic languages, expect full type annotations and a typechecker in the gate.

## Tools

The skill needs Go, [dfd](https://github.com/bilus/dfd) for the diagrams (`go install github.com/bilus/dfd/cmd/dfd@latest`), and dfdmetrics for the score. dfdmetrics ships in `tools/dfdmetrics`: run `go install ./cmd/dfdmetrics` inside that directory. Its README describes the conventions a design must follow and the report it prints.

## Usage

Save the skill, or drop the folder into your skills directory. Enable only one staged-delivery variant at a time, because their triggers overlap. The skill triggers on any multi-edit coding task, and you can also invoke it by name. It works best in typed languages.

## Comparing against the other variants

Since version 0.2.0, this skill's outer loop adds the vocabulary, the design, the score and the review page, so it no longer shares the identical outer loop of the other two staged-delivery skills. Compare it against runs of the same version. Track per feature: the number of stages, the mean stage size, your review minutes per stage, defects caught and escaped, discipline breaks, how often the skeleton survived contact with the implementation, and the score history from `docs/review/scores.tsv`.
