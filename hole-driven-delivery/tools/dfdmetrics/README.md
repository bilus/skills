# dfdmetrics

dfdmetrics measures a leveled data flow design written in the [dfd](https://github.com/bilus/dfd) format. It reports on the design's shared state and on the packages named in its boxes' references.

A state item accessed by two or more boxes gets a home: the deepest diagram that holds all of those boxes, directly or through child diagrams. The home's depth is the item's level, with the top diagram at depth 0. When the item has both a write and a read, the report also gives its live range, from the first write to the last read, and its read distance, the largest distance from a read back to the nearest write at or before it. Both are differences between box positions in the home. The tool also flags boxes and child diagrams whose references span two or more of the design's packages.

## Running it

Install it from this directory, or from anywhere with the module path:

```sh
go install ./cmd/dfdmetrics
go install github.com/bilus/skills/hole-driven-delivery/tools/dfdmetrics/cmd/dfdmetrics@latest
```

Both commands put the binary in `$(go env GOPATH)/bin`, which must be on the PATH for the bare name to work. Then run it on the top diagram of a design:

```sh
dfdmetrics docs/flow.dfd
```

With `-score`, the command prints only the score card: a tab-separated header and one row, where every value is better when lower. The columns are the sum of the live ranges, the largest read distance, the shared items at level 0, the shared items at any level, the largest pressure, the boxes and child diagrams spanning two or more design packages, and the number of findings.

The command exits with status 0 after a report, 2 after a usage error, and 1 after any other error, such as a parse error or a failed `go list std`. Under `go run`, the command exits with status 1 after any error. The tool runs `go list std` to recognize standard library packages, so it needs a Go toolchain on the PATH.

## Writing a design for dfdmetrics

The top diagram is a file such as flow.dfd. A file named flow.3.dfd beside it expands process 3, and flow.3.2.dfd expands process 3.2. A process's number is the explicit number at the start of its label, as in `[3.1. Parse the input]` in flow.3.dfd. Either every process of a diagram has an explicit number or none has one. A child diagram's numbers extend its parent's number by one part, and the top diagram's numbers have one part. Without explicit numbers, the tool follows dfd's own numbering: `dfd --number` for the top diagram, and `dfd --number --number-prefix 3.` for flow.3.dfd, where a repeated title keeps its first number.

A store arrow's label names state items, separated by commas. The store and the label together identify an item, so a write and a read of one item use the same store and the same label.

A store arrow on a box with a child diagram is a summary, and the tool does not count it as an access. The report flags a summary as unmatched unless a box without a child diagram, anywhere below the summary's box, makes the same access. It flags a missing summary when an ancestor box of an accessing box, in the item's home or below it, does not draw the same access. It flags a summary above the home when the summary sits on the box whose child diagram is the item's home, or on any ancestor of that box. An item accessed by one box does not have a home, and the report flags any summary on an ancestor of that box as a summary above home.

A box lists its references in its last top-level pair of parentheses, as in `(analyze.matchBindings, scope.Declares)`. Write the design's own packages by name and every other package by import path, as in `go/format.Source`. A single-element standard library package such as `fmt` does not need a path, because the tool recognizes it as external. A design package with the same name as one of these also counts as external, so it does not appear among the design packages or in the package findings.

## The report

The report lists the shared items grouped by home, then the private items, each accessed by one box only. After the items, it gives the pressure of each diagram with two or more boxes: among the items whose home is that diagram, the largest number live across one gap between consecutive boxes, with the number of gaps at that maximum and the first such gap. The findings, the design packages of each diagram, and totals for comparison between runs come last.

Besides the three summary findings and the two package findings, the report flags a store arrow whose label does not name an item, a dead item without a reader, an orphan item without a writer, a read positioned before the item's first write in the home, and an arrow candidate. An arrow candidate has a live range of 1. It does not have a read positioned before its first write, and its gap does not hold an entity, so the item could travel on the arrow between its two boxes.
