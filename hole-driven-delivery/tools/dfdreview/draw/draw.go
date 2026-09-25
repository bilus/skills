// Package draw draws each diagram of a design before, as a diff and after, with dfd.
package draw

import (
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/code"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/design"
)

// View is one diagram drawn in its views, each an SVG document or "".
type View struct {
	Number, Title       string
	Before, Diff, After string
}

// Views draws each diagram of d in its views, with links to code, types and terms.
// dfd is the command that draws.
func Views(d *design.Design, c *code.Index, dfd string) ([]View, error) {
	l, footnotes := link(d, c)
	patches := diff(l)
	return render(dfd, l, patches, footnotes)
}

// linked is the design's diagrams with their sources linked in both versions.
type linked []design.Diagram

// link links each version of each diagram, and returns the footnotes the links use.
func link(d *design.Design, c *code.Index) (linked, map[string]string) {
	panic("HOLE(4): dfdtext.Link with each version's types and terms; references resolve to code or pkg.go.dev")
}

// diff returns the patch of each diagram whose versions differ, by path.
func diff(l linked) map[string]string {
	panic("HOLE(4): dfdtext.Unified for a changed, a new or a deleted diagram, none for an unchanged one")
}

// render runs dfd for each view of each diagram.
func render(dfd string, l linked, patches, footnotes map[string]string) ([]View, error) {
	panic("HOLE(4): one dfd run per view with a shared footnotes file; an error names the diagram and the version")
}
