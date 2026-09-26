# dfdreview

dfdreview builds the review page of a leveled data flow design written in the [dfd](https://github.com/bilus/dfd) format. The page shows each diagram and each changed file before a stage, as a diff and after it. Links lead from the diagrams into the code, and the page holds the score cards and the report of dfdmetrics and the plan's brief and metaphor.

## Running it

Install it from this directory:

```sh
go install ./cmd/dfdreview
```

The binary lands in `$(go env GOPATH)/bin`, which must be on the PATH for the bare name to work. At a review, run it on the top diagram of the design:

```sh
dfdreview -base <commit> -plan docs/plans/<plan>.md \
  -old-score docs/review/score.tsv -new-score docs/review/score.new.tsv \
  -report docs/review/report.txt docs/flow.dfd
```

| Flag | Gives |
|---|---|
| `-base REV` | The revision to compare against, such as the stage's start commit. Without it, the page shows the working tree alone. |
| `-plan FILE` | The plan. The page shows its first heading and its sections "The change in brief" and "Metaphor". |
| `-old-score FILE` | The cached score card of the last approved review. A file that does not exist yet counts as none. |
| `-new-score FILE` | This review's score card, from `dfdmetrics -score`. |
| `-report FILE` | This review's report, from `dfdmetrics`. |
| `-vocabulary FILE` | The vocabulary, by default `vocabulary.md` beside the design. |
| `-dfd PATH` | The dfd command, by default `dfd`. |
| `-per-row N` | The boxes per row of each drawing, 5 by default. |
| `-o FILE` | The page, by default `review/index.html` beside the design. |

The command exits with status 0 after writing the page, 2 after a usage error, and 1 after any other error. An error names its input: the file, the revision, or the diagram and view that dfd rejected, as in `docs/flow.3.dfd (diff):7: ...`.

dfdreview runs git, dfd and `go list std`. The dfd must have `--patch` and `--footnotes`: check that `dfd --help` lists both.

## What the page relies on

- The design has a top diagram such as `docs/flow.dfd`, with child diagrams named like `flow.3.dfd` beside it. The code directory is the directory above the design's directory, and dfdreview indexes its Go files.
- A process's number is the explicit number at the start of its label, as in `[3. Validate the sum types`. In a design without explicit numbers, it is dfd's own number, as dfdmetrics reads it. A child diagram's tab takes its title from its box, and a click on the box's number opens the tab.
- A box lists its references in its last top-level pair of parentheses, as in `(analyze.sumTypes, go/format.Source)`. A reference links to the Go declaration of that name, from either version. A reference to the standard library or to another module links to pkg.go.dev.
- A type comment, `# type: item = type`, links the item in the flow labels to its type, and the type's first qualified name opens its declaration.
- The vocabulary's lines, `- term: definition`, give each term in a box, an entity or a store name its definition on hover.
- A score card is two tab-separated lines, as `dfdmetrics -score` prints them: the column names and the values.

## How it works

dfdreview reads each diagram at the base and in the working tree and aligns the two versions line by line. It then wraps the references, the typed items and the terms of both versions in dfd footnote references, and writes each changed diagram's patch from its alignment with the linked lines. So a link that exists in one version alone does not show as a change. dfd draws every view with one shared file of footnote definitions. The page holds the drawings, the files, the declarations, the types, the terms and the notes as JSON, and its script adds the tabs and the code panel. Dragging the divider beside the code panel, or pressing the arrow keys on it, resizes the panel, and the browser keeps the width.

`docs/flow.dfd` and `docs/flow.4.dfd` hold dfdreview's own design, `docs/vocabulary.md` its terms, and `docs/plans` the plan and the ledger of its construction. `docs/review/index.html` is the review page of that design.
