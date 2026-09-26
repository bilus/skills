// Package draw draws each diagram of a design before, as a diff and after, with dfd.
package draw

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

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
// Without a base, nothing has changed.
func align(d *design.Design) map[string][]dfdtext.Op {
	alignments := map[string][]dfdtext.Op{}
	if d.Base == "" {
		return alignments
	}
	for _, dg := range d.Diagrams {
		b, a := dg.Source.Before, dg.Source.After
		if b.Found && a.Found && b.Content == a.Content {
			continue
		}
		alignments[dg.Path] = dfdtext.Align(b.Content, a.Content)
	}
	return alignments
}

// link links both versions of each diagram and writes each patch with the linked lines.
// It returns the sheets and the footnote definitions their links use.
func link(d *design.Design, c *code.Index, alignments map[string][]dfdtext.Op) ([]sheet, map[string]string) {
	footnotes := map[string]string{}
	target := func(ref dfdtext.Reference) string {
		key := code.Key(ref)
		_, before := c.Before[key]
		_, after := c.After[key]
		if !before && !after && (c.Std[ref.Qualifier] || strings.Contains(ref.Qualifier, ".")) {
			return "https://pkg.go.dev/" + ref.Qualifier + "#" + ref.Name
		}
		// A name without a declaration still links, so the page can report it.
		return "#code/" + key
	}
	linked := func(t repo.Text, vocabulary repo.Text) repo.Text {
		if !t.Found {
			return t
		}
		src, used := dfdtext.Link(t.Content, dfdtext.Linker{
			Code:  target,
			Types: dfdtext.Types(t.Content),
			Terms: design.Terms(vocabulary.Content),
		})
		for id, to := range used {
			footnotes[id] = to
		}
		return repo.Text{Content: src, Found: true}
	}
	var sheets []sheet
	for _, dg := range d.Diagrams {
		s := sheet{number: dg.Number, title: dg.Title, path: dg.Path}
		s.source.Before = linked(dg.Source.Before, d.Vocabulary.Before)
		s.source.After = linked(dg.Source.After, d.Vocabulary.After)
		if ops, changed := alignments[dg.Path]; changed {
			s.patch = dfdtext.Patch(filepath.Base(dg.Path), ops, s.source.Before.Content, s.source.After.Content)
		}
		sheets = append(sheets, s)
	}
	return sheets, footnotes
}

// render runs dfd for each view of each sheet, with the footnote definitions in one file.
func render(dfd string, sheets []sheet, footnotes map[string]string) ([]View, error) {
	dir, err := os.MkdirTemp("", "dfdreview")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir) // a leftover temporary directory harms nothing
	defs := filepath.Join(dir, "footnotes.dfd")
	ids := make([]string, 0, len(footnotes))
	for id := range footnotes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	var b strings.Builder
	for _, id := range ids {
		fmt.Fprintf(&b, "{%s} %s\n", id, footnotes[id])
	}
	if err := os.WriteFile(defs, []byte(b.String()), 0o644); err != nil {
		return nil, err
	}
	run := func(s sheet, view, input string, patch bool) (string, error) {
		args := []string{"--box", "300x150", "--per-row", "5", "--number", "--footnotes", defs, "-o", "-"}
		if s.number != "" {
			// dfd's own numbering needs the prefix; explicit numbers ignore it.
			args = append(args, "--number-prefix", s.number+".")
		}
		if patch {
			args = append(args, "--patch")
		}
		cmd := exec.Command(dfd, args...)
		cmd.Stdin = strings.NewReader(input)
		var out, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &out, &stderr
		name := s.path + " (" + view + ")"
		if err := cmd.Run(); err != nil {
			if msg := strings.TrimSpace(stderr.String()); msg != "" {
				return "", errors.New(strings.ReplaceAll(msg, "<stdin>", name))
			}
			return "", fmt.Errorf("%s: %s: %w", name, dfd, err)
		}
		return out.String(), nil
	}
	var views []View
	for _, s := range sheets {
		v := View{Number: s.number, Title: s.title}
		switch {
		case s.patch == "" && s.source.After.Found:
			v.After, err = run(s, "after", s.source.After.Content, false)
		case s.patch == "":
			v.Before, err = run(s, "before", s.source.Before.Content, false)
		default:
			if s.source.Before.Found {
				if v.Before, err = run(s, "before", s.source.Before.Content, false); err != nil {
					return nil, err
				}
			}
			if v.Diff, err = run(s, "diff", s.patch, true); err != nil {
				return nil, err
			}
			if s.source.After.Found {
				v.After, err = run(s, "after", s.source.After.Content, false)
			}
		}
		if err != nil {
			return nil, err
		}
		views = append(views, v)
	}
	return views, nil
}
