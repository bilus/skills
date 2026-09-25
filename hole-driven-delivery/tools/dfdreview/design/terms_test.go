package design_test

import (
	"reflect"
	"testing"

	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/design"
)

func TestTerms(t *testing.T) {
	src := "# Vocabulary\n\nTerms, one per line.\n\n- base: the revision to compare against.\n- score card: dfdmetrics -score output: two lines.\n-not a term\n"
	want := map[string]string{"base": "the revision to compare against.", "score card": "dfdmetrics -score output: two lines."}
	if got := design.Terms(src); !reflect.DeepEqual(got, want) {
		t.Errorf("terms = %v, want %v", got, want)
	}
}
