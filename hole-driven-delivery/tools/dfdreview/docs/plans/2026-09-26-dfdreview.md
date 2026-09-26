# dfdreview, the review page builder

dfdreview builds the review page of hole-driven delivery: the leveled design and the code of a change, each before the stage, as a diff and after it, with the score cards, the dfdmetrics report and the plan's brief and metaphor.

## Requirements

1. `dfdreview [flags] docs/flow.dfd` writes one HTML page, by default `docs/review/index.html`, for the design rooted at the given top diagram. The child diagrams are the files named like `flow.3.dfd` beside it.
2. With `-base REV`, a diagram changed since REV gets three views: Before, Diff and After. The Diff view is dfd's `--patch` drawing. An unchanged diagram gets one view, a diagram new since REV gets Diff and After, and a deleted diagram gets Before and Diff. Without `-base`, every diagram gets one view.
3. A reference in the last top-level parentheses of a box links to its Go declaration. Clicking it shows the source file, highlighted and scrolled to the declaration, with the declaration's lines marked, and with Before, Diff and After views when the file changed. A reference to a package outside the repository links to pkg.go.dev, and the page reports a reference without a declaration in either version.
4. An item of a flow arrow's label with a `# type:` comment links to its type. Clicking it shows the type and opens the declaration of the first named type.
5. A vocabulary term in a box, an entity or a store name shows its definition on hover, and the page lists the vocabulary's changes since the base.
6. Each child diagram has a tab named by its number and its action line, and clicking the number of its box opens the tab.
7. The page lists the files changed since the base, untracked files included, and each file opens its diff.
8. The page shows the new score card beside the old one, with the change in each column and a mark on every rise, the dfdmetrics report, and the plan's sections "The change in brief" and "Metaphor".
9. Every error names its input: the file, the revision, or the diagram and version that dfd rejected.
10. `README.md` in the tool's directory describes the flags and the conventions the page relies on.

## Questions and assumptions

- dfdreview runs the dfd command instead of importing dfd, so it works with any dfd that has `--patch` and footnotes, and pins no dfd version.
- dfdreview adds its links as footnotes on source lines, with a line-level reading of the dfd format, so it needs no dfd parser. A context line that the two versions read differently shows as changed.
- The code directory is the directory above the design's directory, and dfdreview indexes only its Go files.
- The page loads highlight.js from cdnjs, like the page builder it replaces.
- The review page for this plan waits for the builder itself: the first page comes from stage 5.

## The change in brief

A stage boundary reads each diagram of the design at the base and in the working tree, links its references, flow label items and terms as footnotes, and draws it with dfd before, as a diff and after. The Go declarations of both versions give the references their code, and the changed files give the code its diffs. The page puts the drawings, the code, the score cards, the report and the plan's brief in one file.

## Metaphor

The review page is an editor's proof packet. Each changed drawing and file comes as the old copy, a marked-up copy and the new copy. A reference in a box is a citation that opens its source. The packet ends with the editor's notes: the plan's brief, the score cards and the report.

## Planned declarations

Package `dfdreview`, the command's library:

- `type Options struct`: the inputs of one build: Design, Base, Vocabulary, Plan, OldScore, NewScore, Report, DFD, Out.
- `func Build(opts Options) error`: Build writes the review page that opts describes.
- `func writePage(opts Options, views []draw.View, d *design.Design, c *code.Index) error`: reads the notes, assembles the page data and writes the page.

Package `repo`: `type Repo`, `func Open(dir, base string) (*Repo, error)`, and the methods `Base() string`, `Read(path) (Pair, error)`, `Names(dir) ([]string, error)`, `GoFiles() (before, after map[string]string, err error)` and `Changed() ([]Change, error)`, with the types `Text`, `Pair` and `Change`; a change holds its diff and both versions of the file.

Package `dfdtext`, dfd source text line by line: `Link(src, Linker) (string, map[string]string)`, `References(src) []Reference`, `Types(src) map[string]string`, `Titles(src) map[string]string`, `Align(before, after) []Op` and `Patch(path, ops, before, after) string`, with the types `Linker`, `CodeFn`, `Reference` and `Op`.

Package `design`: `type Design` (with the base revision), `type Diagram`, `func Read(r *repo.Repo, top, vocabulary string) (*Design, error)`, and `func Terms(vocabulary string) map[string]string`.

Package `code`: `type Index`, `type Place`, `func Read(r *repo.Repo, d *design.Design) (*Index, error)`, `func Declarations(files map[string]string) (map[string]Place, error)`, and `func Key(dfdtext.Reference) string`.

Package `draw`: `type View`, `func Views(d *design.Design, c *code.Index, dfd string) ([]View, error)`, and the steps of process 4: `align`, `link` and `render`, with the type `sheet`.

Package `notes`: `type Notes`, `type Card`, `func Read(plan, oldScore, newScore, report string) (*Notes, error)`, called by `writePage`.

Package `page`: `func Write(w io.Writer, d Data) error` and the page's data types, with the page embedded.

Every step of the top diagram lives in its own package, so the tests of each stage reach its code through the package's exported API. The boxes' functions call repo, dfdtext, notes and page inside their own bodies, so no box names those packages.

## Stages

### Stage 1: skeleton

Goal: the vocabulary, the design and every planned declaration, with every body a hole.
Requirement: all.
Dependencies: none.
Holes: adds the holes of stages 2 to 5.
Acceptance: `go vet ./... && go test ./...` passes and the census lists every hole.
Size: 250 lines.

### Stage 2: reading

Goal: dfdreview reads the design, the vocabulary, the notes and every file of both versions through git.
Requirement: 1, 8, 9.
Dependencies: stage 1.
Holes: 2 repo.Open, 2 repo.Repo.Read, 2 repo.Repo.Names, 2 repo.Repo.GoFiles, 2 repo.Repo.Changed, 2 dfdtext.Types, 2 dfdtext.Titles, 2 design.Read, 2 notes.Read.
Acceptance: tests on a temporary git repository for a modified, a new, a deleted and an untracked file, and for the plan's sections and the score cards.
Size: 250 lines.

### Stage 3: linking and diffs

Goal: every reference, typed item and term of a dfd source becomes a footnote reference, and two versions of a source become a full-context patch.
Requirement: 3, 4, 5.
Dependencies: stage 1.
Holes: 3 dfdtext.Link, 3 dfdtext.References, 3 dfdtext.Unified (filled as Align and Patch).
Acceptance: tests of Link on titles over several lines, aliases, entities, store names and flow labels over several lines, and of Align and Patch.
Size: 300 lines.

### Stage 4: code and drawings

Goal: the Go declarations of both versions are indexed, and every diagram is drawn in its views by dfd.
Requirement: 2, 3, 9.
Dependencies: stages 2 and 3.
Holes: 4 code.Read, 4 code.Declarations, 4 draw.align, 4 draw.link, 4 draw.render.
Acceptance: tests of code.Read and code.Declarations on a temporary repository, and of draw.Views with a stand-in dfd and, when it supports `--patch`, the real one.
Size: 250 lines.

### Stage 5: the page data

Goal: dfdreview writes a page that holds every view, source file, declaration, type, term, change and note, and shows the drawings in tabs.
Requirement: 1, 2, 7, 8.
Dependencies: stage 4.
Holes: 5 dfdreview.writePage, 5 page.Write.
Acceptance: a test of Build on a temporary repository with a stand-in dfd, which reads the page's data back.
Size: 250 lines.

### Stage 6: the page's script

Goal: the page's links open code, types, terms and tabs, and its code panel shows each file before, as a diff and after.
Requirement: 3 to 6.
Dependencies: stage 5.
Holes: none; the script is one file, written in one piece.
Acceptance: the page of this tool's own design, checked in a browser.
Size: 350 lines.

### Stage 7: documentation

Goal: the README describes the flags and the conventions.
Requirement: 10.
Dependencies: stage 6.
Holes: none.
Acceptance: the skill's and the tool's text match the flags that `dfdreview -h` prints.
Size: 80 lines.
