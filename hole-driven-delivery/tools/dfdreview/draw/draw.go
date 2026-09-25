// Package draw draws each diagram of a design before, as a diff and after, with dfd.
package draw

import (
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/code"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/design"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/dfdtext"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/repo"
)

// View is one diagram drawn in its views, each an SVG document or "".
type View struct {
	Number, Title       string
	Before, Diff, After string
}

// Views draws each diagram of d in its views, with links to code, types and terms.
// dfd is the command that draws.
func Views(d *design.Design, c *code.Index, dfd string) ([]View, error) {
	alignments := align(d)
	sheets, footnotes := link(d, c, alignments)
	return render(dfd, sheets, footnotes)
}

// sheet is a diagram ready to draw: both versions and the patch, linked.
type sheet struct {
	number, title, path string
	source              repo.Pair // the linked versions
	patch               string    // "" for a diagram without changes
}

// align matches the lines of the two versions of each changed diagram, by path.
func align(d *design.Design) map[string][]dfdtext.Op {
	panic("HOLE(4): dfdtext.Align on the raw sources of a changed, a new or a deleted diagram, none for an unchanged one")
}

// link links both versions of each diagram and writes each patch with the linked lines.
// It returns the sheets and the footnote definitions their links use.
func link(d *design.Design, c *code.Index, alignments map[string][]dfdtext.Op) ([]sheet, map[string]string) {
	panic("HOLE(4): dfdtext.Link with each version's types and terms; references resolve to code or pkg.go.dev")
}

// render runs dfd for each view of each sheet.
func render(dfd string, sheets []sheet, footnotes map[string]string) ([]View, error) {
	panic("HOLE(4): one dfd run per view with a shared footnotes file; an error names the diagram and the version")
}
