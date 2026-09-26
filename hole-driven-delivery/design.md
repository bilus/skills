# Designing with leveled data flow diagrams

The design of a change is a set of dfd files: a top diagram, and a child diagram for each process whose function calls steps of its own. The top diagram comes before the code. A child diagram comes with the stage whose fill splits its process's function into steps, so every box stands for an existing function or for a hole declared in its stage. Keep every diagram true to the code at every stage boundary. dfdmetrics reads the same files, so the conventions below are also its input format, and `tools/dfdmetrics/README.md` defines its terms.

## The dfd format

```
{Entity}           an external source or sink
[Process]          a process box; document order is the flow; a box may span lines
|Store|            a datastore, attached to the process above it
> label            before a process: the flow arrow into it; before a store: a write
< label            before a store: a read
[A := Long name]   an alias; a bare [A] later means the same process
# comment          ignored by dfd; the design uses it for arrow types
```

`dfd --help` lists the flags. dfdreview renders every diagram for the review page. To render one by hand, run `dfd --box 300x150 --per-row 5 --number docs/flow.dfd -o flow.svg`.

Leave footnotes out of the design files: dfdreview adds the links to code, types and terms when it builds the page.

## Numbers

Write each process's number at the start of its label: `[3. Validate the sum types`, and in the child diagram of process 3, `[3.1. Collect the sum types`. dfd shows an explicit number in the box's number band. The number stays with its box when a later change inserts a box before it, so the child diagram's file name stays valid. Either every process of a diagram has a number or none has one, and the boxes of one process share its number and its label. dfdmetrics and dfdreview also follow dfd's own numbering, so a design drawn without explicit numbers keeps working. Add its numbers in the design step of its next plan, since the freeze of step 10 covers them.

## Process boxes

A box holds its number, an action line, a purpose line and the references:

1. The number and the action line, in the imperative: the step's action ("3. Validate the sum types").
2. The purpose line: why the later steps need this step ("to give patterns their variants"). Keep it short, and leave it out when the outgoing arrow's label already gives it.
3. The references, in the last top-level pair of parentheses: typically the process's own function, as in `(analyze.sumTypes)`. A box may name several functions of one package, as in `(analyze.sumTypes, analyze.checkCollisions)`. Write the design's own packages by name and other packages by import path (`go/format.Source`). A single-element standard library package such as `fmt` does not need a path. A design package named like one of them, such as `errors` or `log`, counts as external, so avoid such names.

The action line and the doc comment of the box's first reference say the same thing. "Analyze reads the source code and builds a list of declarations" becomes "Read the source code and build a list of declarations". Write one from the other, by hand or with a small model, and keep them in step.

## Arrows

- An arrow's label names the input data of the next process, as its function's parameters declare it: one item per parameter, with its declared type. A struct passed whole is one item, not a list of its fields.
- When a step changes something in place and returns nothing, an arrow does not carry the change: draw a store write instead.
- Every item of a flow label gets one `# type: item = type` comment somewhere in the file, so the label `sum types, diagnostics` gets `# type: sum types = []*analyze.Sum` and `# type: diagnostics = []analyze.Diagnostic`. A label with one item gets one comment.
- dfd draws a flow arrow between every two consecutive steps. When two steps communicate only through a store, label the arrow with the arguments of the second step.

## Stores and state

- A store stands for state that outlives one call: a struct filled in place, a map, a file, a directory.
- A store arrow's label lists state items, separated by commas. The store and the label together identify an item, so a write and the read of one item use the same store and the same label.
- Every write needs a read. A store read only from outside the flow, such as a build's input files, shows as a dead item. Record it in the ledger as expected, and keep the store in the diagram of its writer.
- Prefer arrows to stores. An item read only by the next step belongs on the arrow between the two steps.

## Levels and boundaries

- The top diagram is `docs/flow.dfd`. Process N with internal steps gets `docs/flow.N.dfd`, and process N.M gets `docs/flow.N.M.dfd`.
- A process with a child diagram appears once in its diagram.
- A child diagram may show the neighbours of its parent process as entities, so its inputs and outputs are visible.
- A few neighbouring steps may share state, as long as the last of them returns a finished value. Group them into one process. One level up, the group is a process with plain inputs and outputs, and its child diagram holds the store. Apply this rule at every level.
- A parent diagram draws a store arrow on a decomposed box only for a state item with accessing boxes both inside and outside that box. Such an arrow is a summary, and dfdmetrics checks each one against the leaf boxes anywhere below the decomposed box.
- Levels grow with the fills. The plan draws the top diagram, and the skeleton declares each top-level process's function with a hole for its body. When a fill replaces a hole with calls to new functions, the same stage draws that process's child diagram, with one box per new function.
- The code follows the levels. Each decomposed process is one function or method, and its body calls the functions of its child processes in order.

## Packages

- Each box references functions of one design package.
- When a decomposed process's function would call a function of another design package directly, give the process's package a wrapper function for the call. The box references the wrapper, and the wrapper's child diagram holds only the other package's steps. A call inside a leaf box's function may reach any package.
- The top diagram shows the architecture and may span several packages. Each child diagram stays within one package.

## Findings and their usual fixes

| dfdmetrics finding | Meaning | Usual fix |
|---|---|---|
| dead item | written, never read | Remove the write, or draw the reader. Record an external reader in the ledger. |
| orphan item | read, never written | Draw the writer, or give the write and the read the same label. |
| read before write | a read positioned before the first write in the item's home | Reorder the steps, or separate two items with one label. |
| arrow candidate | read only by the box after the writer in the item's home, with no entity between them | Return the item and put it on the arrow. |
| unmatched summary | a summary without a matching access below it | Draw the access in the child diagram, or drop the summary. |
| missing summary | state crossing a boundary without a summary on the decomposed box | Add the store arrow to that box. |
| summary above home | a summary on the box that expands the item's home or above it, or on an ancestor of an item's only accessing box | Drop the summary. |
| mixed box | a box spanning two or more design packages | Wrap the other package's call in a function of the box's package. |
| mixed diagram | a child diagram spanning two or more design packages | Push the other package one level down. |
| unlabeled arrow | a store arrow without an item | Name the item. |

## Common mistakes

- Drawing steps before their functions exist. A box without a declaration misleads the reviewer, and the review page reports it. Draw a finer step with the fill that creates its function.
- Listing a struct's fields on an arrow when the next function takes the struct itself.
- Drawing a change in place as an arrow out of the step, as if the step returned the changed value.
- Drawing every store of the design on the top diagram. Only state shared across top-level boxes belongs there.
- Writing an item under one label and reading it under another, such as "sites" and "matches".
- Ending a box with prose in parentheses, as in "(optional)". The references must be the last parentheses.
- Changing the code without redrawing the diagram, so the next review scores an outdated design.
