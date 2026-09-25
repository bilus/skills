package dfdmetrics_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/bilus/skills/hole-driven-delivery/tools/dfdmetrics"
)

// writeDesign writes files into a temporary directory and returns the path of flow.dfd there.
func writeDesign(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, src := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(dir, "flow.dfd")
}

func load(t *testing.T, files map[string]string) *dfdmetrics.Design {
	t.Helper()
	d, err := dfdmetrics.Load(writeDesign(t, files))
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestLoadNumbersBoxesLikeDfd(t *testing.T) {
	d := load(t, map[string]string{"flow.dfd": `
{In}
> a
[A]
> b
{E}
> c
[B]
> d
[A]
> e
[C]
`})
	top := d.Diagrams[0]
	var numbers []string
	var positions []int
	var entities []bool
	for _, b := range top.Boxes {
		numbers = append(numbers, b.Number)
		positions = append(positions, b.Position)
		entities = append(entities, b.EntityAfter)
	}
	if want := []string{"1", "2", "1", "3"}; !reflect.DeepEqual(numbers, want) {
		t.Errorf("numbers = %v, want %v", numbers, want)
	}
	if want := []int{1, 2, 3, 4}; !reflect.DeepEqual(positions, want) {
		t.Errorf("positions = %v, want %v", positions, want)
	}
	if want := []bool{true, false, false, false}; !reflect.DeepEqual(entities, want) {
		t.Errorf("entity after = %v, want %v", entities, want)
	}
}

func TestLoadFindsChildDiagramsByNumber(t *testing.T) {
	d := load(t, map[string]string{
		"flow.dfd":     "[A]\n> x\n[B]\n",
		"flow.2.dfd":   "[B1]\n> y\n[B2]\n",
		"flow.2.1.dfd": "[C]\n",
		"flow.v2.dfd":  "not a child diagram",
		"other.2.dfd":  "not a child diagram either",
	})
	if len(d.Diagrams) != 3 {
		t.Fatalf("loaded %d diagrams, want 3", len(d.Diagrams))
	}
	top, two, twoOne := d.Diagrams[0], d.Diagrams[1], d.Diagrams[2]
	for i, want := range []struct {
		base, number string
		depth        int
	}{{"flow.dfd", "", 0}, {"flow.2.dfd", "2", 1}, {"flow.2.1.dfd", "2.1", 2}} {
		dg := d.Diagrams[i]
		if filepath.Base(dg.Path) != want.base || dg.Number != want.number || dg.Depth != want.depth {
			t.Errorf("diagram %d = %s %q depth %d, want %s %q depth %d", i, filepath.Base(dg.Path), dg.Number, dg.Depth, want.base, want.number, want.depth)
		}
	}
	if top.Parent != nil || two.Parent != top.Boxes[1] || twoOne.Parent != two.Boxes[0] {
		t.Error("child diagrams do not point at the boxes they expand")
	}
	if top.Boxes[1].Child != two || two.Boxes[0].Child != twoOne || top.Boxes[0].Child != nil {
		t.Error("boxes do not point at their child diagrams")
	}
	if two.Boxes[0].Number != "2.1" || two.Boxes[1].Number != "2.2" || twoOne.Boxes[0].Number != "2.1.1" {
		t.Errorf("child numbers = %s %s %s", two.Boxes[0].Number, two.Boxes[1].Number, twoOne.Boxes[0].Number)
	}
}

func TestLoadRejectsInconsistentChildDiagrams(t *testing.T) {
	for _, tc := range []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{"no such process", map[string]string{"flow.dfd": "[A]\n> x\n[B]\n", "flow.7.dfd": "[Z]\n"}, []string{"has no process 7"}},
		{"missing parent", map[string]string{"flow.dfd": "[A]\n", "flow.1.2.dfd": "[Z]\n"}, []string{"missing parent diagram", "flow.1.dfd"}},
		{"drawn twice", map[string]string{"flow.dfd": "[A]\n> x\n[B]\n> y\n[A]\n", "flow.1.dfd": "[Z]\n"}, []string{"process 1 is drawn 2 times"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := dfdmetrics.Load(writeDesign(t, tc.files))
			if err == nil {
				t.Fatalf("Load succeeded, want an error containing %q", tc.want)
			}
			for _, w := range tc.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("error = %v, want it to contain %q", err, w)
				}
			}
		})
	}
}

func TestLoadSplitsStoreArrowsIntoItems(t *testing.T) {
	d := load(t, map[string]string{"flow.dfd": `
[A]
    > sum types,
      diagnostics
    |P|
    > a,  , b
    |Q|
    >
    |R|
    > sum types
    |P|
    < x
    |Q|
> next
[B]
`})
	got := d.Diagrams[0].Boxes[0].Accesses
	want := []dfdmetrics.Access{
		{Store: "P", Item: "sum types", Write: true},
		{Store: "P", Item: "diagnostics", Write: true},
		{Store: "Q", Item: "a", Write: true},
		{Store: "Q", Item: "b", Write: true},
		{Store: "R", Item: "", Write: true},
		{Store: "Q", Item: "x", Write: false},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("accesses = %+v\nwant %+v", got, want)
	}
}

func TestLoadExtractsReferencesFromTheLastParentheses(t *testing.T) {
	d := load(t, map[string]string{"flow.dfd": `
[Walk
 (analyze.Analyze,
  go/format.Source,
  golang.org/x/tools/go/packages.Load)]
> a
[Close (analyze.Run, (*analyze.File).Close)]
> b
[Check (optional) flags
 (render.site)]
> c
[Plain step]
`})
	boxes := d.Diagrams[0].Boxes
	for i, want := range [][]dfdmetrics.Ref{
		{{Qualifier: "analyze", Name: "Analyze"}, {Qualifier: "go/format", Name: "Source"}, {Qualifier: "golang.org/x/tools/go/packages", Name: "Load"}},
		{{Qualifier: "analyze", Name: "Run"}, {Qualifier: "analyze", Name: "File"}},
		{{Qualifier: "render", Name: "site"}},
		nil,
	} {
		if !reflect.DeepEqual(boxes[i].Refs, want) {
			t.Errorf("box %d refs = %+v, want %+v", i+1, boxes[i].Refs, want)
		}
	}
	if !strings.Contains(boxes[0].Title, "\n") {
		t.Errorf("title %q lost its line breaks", boxes[0].Title)
	}
}
