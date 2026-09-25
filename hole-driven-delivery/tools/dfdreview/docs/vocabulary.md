# Vocabulary

Terms of dfdreview, one per line.

- review page: the HTML page a reviewer opens at the plan gate and at each stage boundary.
- base: the revision that the page compares against, given with -base; at a stage boundary, the commit at the start of the stage. Without a base, nothing has changed.
- version: one side of the comparison: the base, shown as Before, or the working tree, shown as After.
- top diagram: the diagram the command names, such as docs/flow.dfd; a child diagram expands one of its processes, as flow.3.dfd expands process 3.
- action line: the first line of a box, the step's action in the imperative.
- view: one drawing of a diagram, or one text of a source file: Before, Diff or After. An unchanged diagram or file has a single view.
- element: a process title, an entity name, a store name or an arrow label of a dfd source, with its text on each of its lines.
- reference: a qualified name in the last top-level parentheses of a box, such as analyze.sumTypes.
- item: one comma-separated part of a flow arrow's label, such as "sum types".
- type comment: a line "# type: item = type" of a dfd source; an item with one is a typed item.
- vocabulary: the file of terms, one "- term: definition" per line.
- term: a word or phrase that the vocabulary defines.
- target: where a link points: a declaration, a type, a term, or a page outside the repository.
- link: a span of a drawing that points at a target. dfdreview makes one by wrapping a reference, a typed item or a term in dfd's footnote syntax, {text:id}, with a footnote definition, {id} target.
- footnote definitions: the target of each footnote id that the links use, passed to dfd in one file.
- patch: the unified diff of one diagram, with the whole file as context, that dfd draws as the Diff view.
- design: the diagrams and the vocabulary, in both versions.
- code directory: the directory above the design's directory, whose Go files dfdreview indexes.
- declaration: a top-level Go function, type, variable or constant, with its file and lines, in one version.
- code index: the declarations of both versions, and the changed files with their diffs.
- changed file: a file of the code directory that differs between the versions, a new or a deleted file included.
- diff: the unified diff of a changed file, as git prints it.
- score card: the output of dfdmetrics -score. The page shows the old card, cached from the last approved review, beside the new card of this review.
- report: the output of dfdmetrics without -score.
- notes: the plan's sections "The change in brief" and "Metaphor", the score cards and the report.
- code panel: the part of the page that shows a source file in its views.
