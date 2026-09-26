package design_test

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/design"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/internal/gittest"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/repo"
)

func TestReadFindsChildDiagramsInBothVersions(t *testing.T) {
	g := gittest.New(t)
	g.Write(map[string]string{
		"docs/flow.dfd":      "[1. Read\n (p.Read)]\n> x\n[2. Check]\n> y\n[3. Write]\n",
		"docs/flow.2.dfd":    "[2.1. Parse]\n",
		"docs/flow.2.1.dfd":  "[2.1.1. Split]\n",
		"docs/flow.3.dfd":    "[3.1. Old]\n",
		"docs/other.2.dfd":   "[not a child]\n",
		"docs/vocabulary.md": "- term: old\n",
	})
	base := g.Commit("base")
	g.Write(map[string]string{"docs/flow.1.dfd": "[1.1. New]\n", "docs/vocabulary.md": "- term: new\n"})
	g.Remove("docs/flow.3.dfd")
	r, err := repo.Open(g.Dir, base)
	if err != nil {
		t.Fatal(err)
	}
	docs := filepath.Join(g.Dir, "docs")
	d, err := design.Read(r, filepath.Join(docs, "flow.dfd"), filepath.Join(docs, "vocabulary.md"))
	if err != nil {
		t.Fatal(err)
	}
	var numbers, titles, paths []string
	for _, dg := range d.Diagrams {
		numbers = append(numbers, dg.Number)
		titles = append(titles, dg.Title)
		paths = append(paths, filepath.Base(dg.Path))
	}
	if want := []string{"", "1", "2", "2.1", "3"}; !reflect.DeepEqual(numbers, want) {
		t.Errorf("numbers = %q, want %q", numbers, want)
	}
	if want := []string{"Overview", "Read", "Check", "Parse", "Write"}; !reflect.DeepEqual(titles, want) {
		t.Errorf("titles = %q, want %q", titles, want)
	}
	if want := []string{"flow.dfd", "flow.1.dfd", "flow.2.dfd", "flow.2.1.dfd", "flow.3.dfd"}; !reflect.DeepEqual(paths, want) {
		t.Errorf("paths = %q, want %q", paths, want)
	}
	if one := d.Diagrams[1].Source; one.Before.Found || one.After.Content != "[1.1. New]\n" {
		t.Errorf("flow.1.dfd, new since the base: %+v", one)
	}
	if three := d.Diagrams[4].Source; !three.Before.Found || three.After.Found {
		t.Errorf("flow.3.dfd, deleted since the base: %+v", three)
	}
	if v := d.Vocabulary; v.Before.Content != "- term: old\n" || v.After.Content != "- term: new\n" {
		t.Errorf("vocabulary = %+v", v)
	}
}

func TestReadTitlesAnUnnumberedDesign(t *testing.T) {
	g := gittest.New(t)
	g.Write(map[string]string{
		"docs/flow.dfd":     "[Read]\n> x\n[Check the items\n (p.Check)]\n",
		"docs/flow.2.dfd":   "[Parse]\n> y\n[Validate]\n",
		"docs/flow.2.2.dfd": "[Split]\n",
	})
	r, err := repo.Open(g.Dir, "")
	if err != nil {
		t.Fatal(err)
	}
	d, err := design.Read(r, filepath.Join(g.Dir, "docs", "flow.dfd"), "")
	if err != nil {
		t.Fatal(err)
	}
	var titles []string
	for _, dg := range d.Diagrams {
		titles = append(titles, dg.Title)
	}
	if want := []string{"Overview", "Check the items", "Validate"}; !reflect.DeepEqual(titles, want) {
		t.Errorf("titles = %q, want %q", titles, want)
	}
}

func TestReadWithoutVocabulary(t *testing.T) {
	g := gittest.New(t)
	g.Write(map[string]string{"docs/flow.dfd": "[A]\n"})
	r, err := repo.Open(g.Dir, "")
	if err != nil {
		t.Fatal(err)
	}
	d, err := design.Read(r, filepath.Join(g.Dir, "docs", "flow.dfd"), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Diagrams) != 1 || d.Vocabulary.After.Found {
		t.Errorf("design = %+v", d)
	}
}

func TestReadRejectsAMissingTopDiagram(t *testing.T) {
	g := gittest.New(t)
	g.Write(map[string]string{"docs/other.dfd": "[A]\n"})
	r, err := repo.Open(g.Dir, "")
	if err != nil {
		t.Fatal(err)
	}
	_, err = design.Read(r, filepath.Join(g.Dir, "docs", "flow.dfd"), "")
	if err == nil || !strings.Contains(err.Error(), "flow.dfd") {
		t.Errorf("error = %v, want one naming flow.dfd", err)
	}
}
