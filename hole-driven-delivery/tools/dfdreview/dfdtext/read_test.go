package dfdtext_test

import (
	"reflect"
	"testing"

	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/dfdtext"
)

func TestTypes(t *testing.T) {
	src := "# type: sum types = []*analyze.Sum\n#type:diagnostics=[]analyze.Diagnostic\n[A]\n# not a type\n# type: sum types = map[string]int\n"
	want := map[string]string{"sum types": "map[string]int", "diagnostics": "[]analyze.Diagnostic"}
	if got := dfdtext.Types(src); !reflect.DeepEqual(got, want) {
		t.Errorf("types = %v, want %v", got, want)
	}
}

// A description is a box's text without its number and its references, on one line.
func TestDescriptions(t *testing.T) {
	src := "{3 Entity}\n> a\n[1. Read the input\n to parse it\n (p.Read,\n  p.Parse)]\n> b\n[v := 2. Check it (see the spec)]\n> c\n[v]\n> d\n[Unnumbered]\n> e\n[3.1. Nested (p.N)]\n"
	want := map[string]string{"1": "Read the input to parse it", "2": "Check it (see the spec)", "3.1": "Nested"}
	if got := dfdtext.Descriptions(src, ""); !reflect.DeepEqual(got, want) {
		t.Errorf("descriptions = %v, want %v", got, want)
	}
}

// Without explicit numbers, dfd numbers each distinct title in order, after the prefix,
// and an alias stands for the label it declares.
func TestDescriptionsFollowDfdsOwnNumbering(t *testing.T) {
	src := "{Client}\n> a\n[Read the input\n to parse it]\n> b\n[v := Check it]\n> c\n[Read the input\n to parse it]\n> d\n[v]\n> e\n[Write]\n"
	want := map[string]string{"2.1": "Read the input to parse it", "2.2": "Check it", "2.3": "Write"}
	if got := dfdtext.Descriptions(src, "2."); !reflect.DeepEqual(got, want) {
		t.Errorf("descriptions = %v, want %v", got, want)
	}
}
