# The review page

The review page is the first thing your human partner opens at every review. dfdreview builds it from the design, the code, the vocabulary, the plan and the two score cards, and writes it to `docs/review/index.html`.

## What it shows

- One tab per diagram: "Overview" for the top diagram, then one tab per decomposed process, labeled with the process's number. The tab's tooltip and the heading above the drawing give the diagram's title: the text of the process's box, without its references. A red asterisk after a tab's label shows that its drawings link a declaration changed since the base. Clicking the number of a box with a child diagram opens that tab.
- For each diagram changed since the base commit, three views: Before, Diff and After. The Diff view is dfd's `--patch` drawing: added boxes, lines and arrows green, removed ones red and struck through, and boxes with added and removed lines amber.
- Function and type names in the boxes as links. Clicking one opens its source file, highlighted and scrolled to the declaration, with Before, Diff and After views when the file changed. In the After view, a strip beside the line numbers marks each added line green, and each line that replaces removed lines amber.
- A red asterisk after each function or type link whose declaration changed since the base, doc comment included, and after each flow item with a changed type. A function also gets one when a function or method that it calls in its own package changed, directly or through further calls in that package. A change to a type's methods leaves the type unmarked.
- In a Go file's Before and After views, each identifier that uses a declared object links to its definition: the same version of its file, at its line, for one in the code directory, and its documentation on pkg.go.dev for one outside it. With a Go module, the page holds every Go file of the code directory in both versions, test files included.
- The package errors of the type check, under "Package errors".
- Flow arrow labels as links. Clicking an item shows its type from its `# type:` comment and opens the declaration of the first named type.
- Vocabulary terms in boxes and store names underlined, with the definition on hover, and the vocabulary's changes since the base.
- The files changed since the base, each opening its diff.
- The new score card beside the cached one, with the change in each column and a mark on every rise.
- The dfdmetrics report, and "The change in brief" and "Metaphor" sections of the plan.

## Building it

```sh
dfdreview -base <commit> -plan <plan path> \
  -old-score docs/review/score.tsv -new-score docs/review/score.new.tsv \
  -report docs/review/report.txt docs/flow.dfd
```

The base is the commit at the start of the stage: the previous stage's boundary commit from the ledger, or at the plan gate, the commit at the start of the work. dfdreview reads the base versions with git, runs dfd for every drawing and type-checks the module with the go command, so all three must be on the PATH. A URL that ends in `#code/<key>`, such as `#code/analyze.sumTypes`, opens that declaration. `tools/dfdreview/README.md` lists the other flags and the conventions the page relies on.

The page loads highlight.js from cdnjs and holds everything else inline. Give your human partner the page's path in the handoff, and publish the page elsewhere, such as an artifact, only when your human partner asks for a link.
