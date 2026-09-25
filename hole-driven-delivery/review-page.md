# The review page

The review page is the first thing your human partner opens at every review. It shows the leveled design with links into the code, the definitions of the vocabulary, and the score with its history. Rebuild it at the plan approval gate and at every stage boundary, and write it to `docs/review/index.html`.

## What it shows

- One tab per diagram: "Overview" for the top diagram, then one tab per decomposed process, named by its number and action line ("3 Validate the sum types"). Clicking a numbered box with a child diagram opens that tab.
- Each diagram rendered by dfd with process numbers and five boxes per row: `dfd --box 300x150 --per-row 5 --number`, plus `--number-prefix N.` for the child diagram of process N.
- Function and type names in the boxes as links. Clicking one shows the whole source file, syntax highlighted, scrolls to the declaration, and marks the declaration's lines. Until the skeleton commit, the page embeds the plan's planned declarations, and the link shows the name's entry there.
- Flow arrow labels as links. Clicking one shows the type of each of the label's items from its `# type:` comment, and opens the declaration of the first named type.
- Every vocabulary term in a box or a store name underlined, with its definition shown on hover.
- The score card: the latest row of `docs/review/scores.tsv`, with each column's change since the previous row and a mark on every rise. Below it, the whole history as a table.
- The findings from dfdmetrics.
- "The change in brief" and the metaphor from the plan.

## Building it

Build the page by hand or with a small script. The method fixes the page's content, not its builder, but a script is quicker to rerun at every boundary. A script can follow these steps:

1. Render each diagram to SVG with dfd. Pipe the output with `-o -`, or write the files under `docs/review/`.
2. Walk the SVG's `<text>` elements in order. A text after a `<rect>` belongs to a box. A text after a `<line>` with an arrowhead belongs to an arrow label: a store arrow when an end of the line lies within a few pixels of a store's lines, a flow arrow otherwise. A text after a plain `<line>` belongs to a store name. With `--number`, store names start with a prefix such as `D1 `: strip it before matching terms. Join the wrapped lines of a flow label, and split it at commas into items, before matching it against the `# type:` comments.
3. Wrap each reference and each flow label in a clickable element for the page's script, and each vocabulary term in a `<tspan>` with a `<title>` child for the hover definition.
4. Find each declaration's line range with the language's own parser, and embed the source files in the page as JSON.
5. Load highlight.js from cdnjs to color the source by syntax.

Keep the page self-contained apart from that script and any web fonts, and readable in both light and dark themes. Keep the builder out of the build: in Go, start it with `//go:build ignore`. Give your human partner the page's path in the handoff, and publish the page elsewhere, such as an artifact, only when your human partner asks for a link.
