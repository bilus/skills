# Hole-driven delivery

A staged-delivery workflow for AI coding agents, adapted from typed-hole programming in Haskell, Idris and Agda (and its cousins, `fatalError()` in Swift and `todo!()` in Rust) and from tracer-bullet development, with a data flow design that leads the code.

Not to be confused with [jhhuh's hole-driven-delivery skill](https://github.com/jhhuh/hole-driven-delivery-skill), which is the inner technique on its own: an autonomous compiler loop (write holes, compile, read diagnostics, fill the most constrained hole, repeat) without a plan, git, tests or review gates. This skill is the delivery process around that idea. The agent agrees the requirements, the vocabulary and a leveled data flow design with you, and commits a compiling skeleton of the change: every body a stub that returns mock data under a `HOLE(id): contract` comment, with a smoke test that runs the whole flow on that data. Then it fills the holes one at a time, outside-in, with the build green after every move and the suite green at every commit. It halts for your review at every stage boundary, and each review starts from a page that shows the design and the code before the stage, as a diff and after it, with the design's score.

## Why this approach

Review capacity is the bottleneck with agents, and the hardest thing to review late is architecture. This skill moves the architecture review to the front and makes it cheap in two steps. The first review shows the top level of the design together with its skeleton. The design is a dfd diagram in which every box names the function that does the step and says why the later steps need it, every arrow names its type, and every change of state in place is a write to a store. The skeleton declares exactly the functions that the diagram names, with every body a hole, and the holes are what you review: names, types, seams and each body's contract, beside the diagram. Each hole returns mock data, so a smoke test runs the whole flow from the first review on, as in tracer-bullet development, and the suite stays green while the agent fills the holes. Each lower level of the design comes with the stage whose fill creates its functions, again together with its code. After that, the compiler becomes the agent's oracle, and `grep -rn "HOLE("` is a progress bar that survives any loss of agent context.

The design also gets measured. The bundled dfdmetrics tool reports how much state the design shares, at which level, and how far each read sits from its write, and it flags boxes that span two packages. At every review, the page shows the new score card beside the last approved one, so you can see whether a stage made the design better or worse.

## What the agent will do

1. Write `docs/plans/<date>-<feature>.md` with numbered requirements, then a vocabulary of the terms the change introduces, reviewed by a sub-agent for redundancy.
2. Draw the top diagram of the design in `docs/flow.dfd`, write a system metaphor of at most 60 words, and declare the diagram's functions and types as a compiling skeleton, with every body a tagged hole that returns mock data. A smoke test runs the top function on that data and passes, and each acceptance test that needs a real body is written and skipped under the tag of the stage that fills its holes. The suite stays green. A sub-agent reviews the design and checks it against the rules implied by the metaphor.
3. Measure the design, save the score card, build the review page with the bundled dfdreview tool, which links every box to its declaration, and stop for your approval of the plan, the design and the skeleton together.
4. After your approval, commit the plan and the skeleton. The approved signatures, the top diagram and the vocabulary are frozen.
5. Fill one group of holes per stage: one hole open at a time, the build never red and the suite green at every commit, each fill landing with the tests that pin its contract. A skipped acceptance test starts to run in the commit that fills the last hole on its path. A fill that splits a function into steps draws that process's child diagram, declares the new functions with holes, writes a smoke test of the process, and halts there, so you review the new level while its bodies are still holes. The budget is 100 to 300 lines per stage, with a hard cap of 400. After the plan gate, a change to the design or the vocabulary waits for your approval, shown with its code, and the agent prefers a change inside a function to a change of an interface between processes.
6. At every boundary, quote a fresh full-suite run and the hole census, check the design against the code, save the score card, rebuild the review page against the stage's start, write the handoff, and halt.
7. After your final sign-off, reshape the history to one commit per stage, with messages that describe the change and not the method.
8. For the next feature, run the same loop on the existing design: propose its changes to the diagrams and the vocabulary, then add the first layer of holes, some of them in existing functions.

## What you do

- Approve the requirements, the vocabulary, the design with its skeleton, and the stages at the plan gate. That review is where your time pays most.
- At each halt, open the review page, check the score's changes, and answer the handoff's one review question.
- Approve any proposed signature change, design change or change to an existing test. The agent may not make these alone.
- In dynamic languages, expect full type annotations and a typechecker in the gate.

## Tools

The skill needs Go, git, [dfd](https://github.com/bilus/dfd) for the diagrams (`go install github.com/bilus/dfd/cmd/dfd@latest`), dfdmetrics for the score, and dfdreview for the review page. The page needs a dfd with `--patch`. dfdmetrics ships in `tools/dfdmetrics` and dfdreview in `tools/dfdreview`: run `go install ./cmd/dfdmetrics` or `go install ./cmd/dfdreview` inside each directory. Their READMEs describe the conventions a design must follow, the report and the page.

## Usage

Save the skill, or drop the folder into your skills directory. Enable only one staged-delivery variant at a time, because their triggers overlap. The skill triggers on any multi-edit coding task, and you can also invoke it by name. It works best in typed languages.

## Comparing against the other variants

Since version 0.2.0, this skill's outer loop adds the vocabulary, the design, the score and the review page, so it no longer shares the identical outer loop of the other two staged-delivery skills. Compare it against runs of the same version. Since version 0.4.0, its holes return mock data under a smoke test, like the hardcoded responses of tracer-bullet-delivery's walking skeleton. Track per feature: the number of stages, the mean stage size, your review minutes per stage, defects caught and escaped, discipline breaks (with the variant's own: a hole that panics, an untagged skip, a skip kept after the fill of its last hole, a smoke test changed to fit a fill), how often the skeleton survived contact with the implementation, and the score cards of the reviews.
