// Package dfdreview builds the review page of a leveled dfd design.
package dfdreview

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/code"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/design"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/dfdtext"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/draw"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/notes"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/page"
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
	PerRow     int    // boxes per row in the drawings, 3 when 0
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
	views, err := draw.Views(d, c, draw.Command{Name: opts.DFD, PerRow: opts.PerRow})
	if err != nil {
		return err
	}
	return writePage(opts, views, d, c)
}

// writePage writes the review page to opts.Out, with the notes that opts names.
// The page leaves itself out of the files it shows.
func writePage(opts Options, views []draw.View, d *design.Design, c *code.Index) error {
	n, err := notes.Read(opts.Plan, opts.OldScore, opts.NewScore, opts.Report)
	if err != nil {
		return err
	}
	self, err := filepath.Rel(absolute(filepath.Dir(filepath.Dir(opts.Design))), absolute(opts.Out))
	if err != nil {
		return err
	}
	self = filepath.ToSlash(self)
	data := page.Data{
		Title: n.Title, Base: d.Base, Brief: n.Brief, Metaphor: n.Metaphor, Report: n.Report,
		OldScore: card(n.OldScore), NewScore: card(n.NewScore), Errors: c.Errors,
		Uncovered: uncovered(c.Uncovered),
		Files:     map[string]page.File{}, Decls: map[string]page.Decl{},
		Terms: page.Versions{Before: design.Terms(d.Vocabulary.Before.Content), After: design.Terms(d.Vocabulary.After.Content)},
	}
	if data.Title == "" {
		data.Title = filepath.Base(opts.Design)
	}
	sources := map[string]repo.Pair{}
	for _, dg := range d.Diagrams {
		sources[dg.Number] = dg.Source
	}
	for _, v := range views {
		src := sources[v.Number]
		data.Diagrams = append(data.Diagrams, page.Diagram{
			Number: v.Number, Title: v.Title, Before: v.Before, Diff: v.Diff, After: v.After,
			Types: page.Versions{Before: dfdtext.Types(src.Before.Content), After: dfdtext.Types(src.After.Content)},
		})
	}
	diffs := map[string]string{}
	for _, ch := range c.Changed {
		if ch.Path != self {
			data.Changed = append(data.Changed, page.Changed{Path: ch.Path, Status: ch.Status})
			diffs[ch.Path] = ch.Diff
		}
	}
	for path, p := range c.Files {
		if path != self {
			data.Files[path] = page.File{Before: content(p.Before), After: content(p.After), Diff: diffs[path],
				BeforeLinks: links(c.BeforeLinks[path]), AfterLinks: links(c.AfterLinks[path])}
		}
	}
	decls := func(places map[string]code.Place, after bool) {
		for key, pl := range places {
			decl := data.Decls[key]
			p := &page.Place{File: pl.File, Start: pl.Start, End: pl.End, Kind: pl.Kind}
			if after {
				decl.After = p
			} else {
				decl.Before = p
			}
			decl.Changed, decl.Reached = c.ChangedDecls[key], c.Reached[key]
			data.Decls[key] = decl
		}
	}
	decls(c.Before, false)
	decls(c.BeforeMethods, false)
	decls(c.After, true)
	decls(c.AfterMethods, true)
	data.Uses = page.Uses{Before: uses(c.BeforeUses), After: uses(c.AfterUses)}
	if err := os.MkdirAll(filepath.Dir(opts.Out), 0o755); err != nil {
		return err
	}
	f, err := os.Create(opts.Out)
	if err != nil {
		return err
	}
	if err := page.Write(f, data); err != nil {
		if cerr := f.Close(); cerr != nil {
			return fmt.Errorf("%w; close %s: %v", err, opts.Out, cerr)
		}
		return err
	}
	return f.Close()
}

// card turns a score card into the page's form.
func card(c *notes.Card) *page.Card {
	if c == nil {
		return nil
	}
	return &page.Card{Columns: c.Columns, Values: c.Values}
}

// uncovered turns the changes that no drawing covers into the page's form.
func uncovered(us []code.Uncovered) []page.Uncovered {
	out := make([]page.Uncovered, 0, len(us))
	for _, u := range us {
		out = append(out, page.Uncovered{Key: u.Key, Mark: u.Mark, File: u.Place.File, Line: u.Place.Start, Version: u.Version})
	}
	return out
}

// links turns a file's identifier links into the page's form.
func links(ls []code.Link) []page.Link {
	if len(ls) == 0 {
		return nil
	}
	out := make([]page.Link, 0, len(ls))
	for _, l := range ls {
		out = append(out, page.Link{Line: l.Line, Col: l.Col, Len: l.Len, File: l.File, To: l.To, URL: l.URL, Key: l.Key, Mark: l.Mark})
	}
	return out
}

// uses turns the uses of each declaration and method into the page's form.
func uses(byKey map[string][]code.Use) map[string][]page.Use {
	out := make(map[string][]page.Use, len(byKey))
	for key, us := range byKey {
		for _, u := range us {
			out[key] = append(out[key], page.Use{In: u.In, File: u.File, Line: u.Line})
		}
	}
	return out
}

// content returns a version's text, nil where the file does not exist.
func content(t repo.Text) *string {
	if !t.Found {
		return nil
	}
	s := t.Content
	return &s
}

// absolute returns p as an absolute path, or p itself when the working directory is gone.
func absolute(p string) string {
	if abs, err := filepath.Abs(p); err == nil {
		return abs
	}
	return p
}
