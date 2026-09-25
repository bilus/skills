# Vocabulary

Terms of dfdreview, one per line.

- review page: the HTML page a reviewer opens at a stage boundary.
- base: the revision the page compares against, the commit at the start of the stage.
- version: a file as it is at the base (before) or in the working tree (after).
- view: one drawing of a diagram: Before, Diff or After.
- element: a box title, a store name or an arrow label of a dfd source, over one or more consecutive lines.
- reference: a qualified name in the last top-level parentheses of a box, such as analyze.sumTypes.
- item: one comma-separated part of a flow arrow's label, such as "sum types".
- term: a word or phrase that the vocabulary defines.
- link: a footnote reference that dfdreview wraps around a reference, an item or a term, so dfd draws the span as a link.
- target: the address of a link: a declaration, a type, a term, or a page outside the repository.
- declaration: a top-level Go function or type, with its file and lines, in one version.
- score card: the output of dfdmetrics -score, cached for the last approved review and new for this one.
- changed file: a file whose content differs between the base and the working tree.
