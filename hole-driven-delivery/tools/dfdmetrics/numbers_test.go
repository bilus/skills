package dfdmetrics_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/bilus/skills/hole-driven-delivery/tools/dfdmetrics"
)

func TestLoadReadsExplicitNumbers(t *testing.T) {
	d := load(t, map[string]string{
		"flow.dfd":   "[1. Read (p.Read)]\n> x\n[3. Check]\n> y\n[1. Read (p.Read)]\n",
		"flow.3.dfd": "[3.1. Parse]\n> z\n[v := 3.2. Validate]\n> w\n[v]\n",
	})
	top, three := d.Diagrams[0], d.Diagrams[1]
	var numbers, titles []string
	for _, b := range append(top.Boxes, three.Boxes...) {
		numbers = append(numbers, b.Number)
		titles = append(titles, b.Title)
	}
	if want := []string{"1", "3", "1", "3.1", "3.2", "3.2"}; !reflect.DeepEqual(numbers, want) {
		t.Errorf("numbers = %v, want %v", numbers, want)
	}
	if want := []string{"Read (p.Read)", "Check", "Read (p.Read)", "Parse", "Validate", "Validate"}; !reflect.DeepEqual(titles, want) {
		t.Errorf("titles = %q, want %q", titles, want)
	}
	if three.Parent != top.Boxes[1] {
		t.Errorf("flow.3.dfd expands box %q, want the box of process 3", three.Parent.Title)
	}
	if refs := top.Boxes[0].Refs; len(refs) != 1 || refs[0].Name != "Read" {
		t.Errorf("refs = %v", refs)
	}
}

func TestLoadRejectsBadExplicitNumbers(t *testing.T) {
	cases := []struct {
		name  string
		files map[string]string
		want  string
	}{
		{"mixed", map[string]string{"flow.dfd": "[1. A]\n> x\n[B]\n"},
			`flow.dfd:3: process "B" has no number; number every process or none`},
		{"one number, two labels", map[string]string{"flow.dfd": "[1. A]\n> x\n[1. B]\n"},
			`flow.dfd:3: number 1 already belongs to "A"`},
		{"dotted in the top diagram", map[string]string{"flow.dfd": "[1.1. A]\n"},
			`flow.dfd: process 1.1 does not belong in the top diagram`},
		{"outside its parent", map[string]string{"flow.dfd": "[1. A]\n> x\n[2. B]\n", "flow.2.dfd": "[3.1. C]\n"},
			`flow.2.dfd: process 3.1 does not belong to process 2`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := dfdmetrics.Load(writeDesign(t, c.files))
			if err == nil || !strings.HasSuffix(err.Error(), c.want) {
				t.Fatalf("error = %v, want one ending in %q", err, c.want)
			}
		})
	}
}
