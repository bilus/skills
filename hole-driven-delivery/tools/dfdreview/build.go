// Package dfdreview builds the review page of a leveled dfd design.
package dfdreview

import (
	"path/filepath"

	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/code"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/design"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/draw"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/repo"
)

// Options are the inputs of one build of the review page.
type Options struct {
	Design     string // the top diagram, such as docs/flow.dfd
	Base       string // the revision to compare against, "" for none
	Vocabulary string // the vocabulary, "" for none
	Plan       string // the plan, "" for none
	OldScore   string // the cached score card, "" for none
	NewScore   string // this review's score card, "" for none
	Report     string // this review's dfdmetrics report, "" for none
	DFD        string // the dfd command
	Out        string // the page to write
}

// Build writes the review page that opts describes.
// The code is the directory above the design's directory.
func Build(opts Options) error {
	r, err := repo.Open(filepath.Dir(filepath.Dir(opts.Design)), opts.Base)
	if err != nil {
		return err
	}
	d, err := design.Read(r, opts.Design, opts.Vocabulary)
	if err != nil {
		return err
	}
	c, err := code.Read(r, d)
	if err != nil {
		return err
	}
	views, err := draw.Views(d, c, opts.DFD)
	if err != nil {
		return err
	}
	return writePage(opts, views, d, c)
}

// writePage writes the review page to opts.Out, with the notes that opts names.
func writePage(opts Options, views []draw.View, d *design.Design, c *code.Index) error {
	panic("HOLE(5): notes.Read, then page.Data from the views, the design, the code and the notes, without the page itself, then page.Write")
}
