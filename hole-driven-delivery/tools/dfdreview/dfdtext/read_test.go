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

func TestTitles(t *testing.T) {
	src := "{3 Entity}\n> a\n[1. Read the input\n to parse it\n (p.Read)]\n> b\n[v := 2. Check it]\n> c\n[v]\n> d\n[Unnumbered]\n> e\n[3.1. Nested (p.N)]\n"
	want := map[string]string{"1": "Read the input", "2": "Check it", "3.1": "Nested (p.N)"}
	if got := dfdtext.Titles(src); !reflect.DeepEqual(got, want) {
		t.Errorf("titles = %v, want %v", got, want)
	}
}
