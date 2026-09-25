package dfdmetrics_test

import (
	"testing"

	"github.com/bilus/skills/hole-driven-delivery/tools/dfdmetrics"
)

// summariesOf loads files and returns the items by "store: label", with the summary findings.
func summariesOf(t *testing.T, files map[string]string) (map[string]*dfdmetrics.Item, []dfdmetrics.Finding) {
	t.Helper()
	d := load(t, files)
	items, _ := dfdmetrics.Items(d)
	byName := map[string]*dfdmetrics.Item{}
	for _, it := range items {
		byName[it.Store+": "+it.Label] = it
	}
	return byName, dfdmetrics.Summaries(d, items)
}

func TestSummariesFlagASummaryWithoutALeafAccessBelowIt(t *testing.T) {
	items, fs := summariesOf(t, map[string]string{
		"flow.dfd":   "[A]\n    > x\n    |S|\n> a\n[B]\n",
		"flow.1.dfd": "[A1]\n    > y\n    |S|\n> b\n[A2]\n    < y\n    |S|\n",
	})
	if !hasFinding(fs, dfdmetrics.UnmatchedSummary, "S", "x", "1") {
		t.Errorf("findings = %+v, want an unmatched summary of S: x on box 1", fs)
	}
	if items["S: x"] != nil {
		t.Errorf("an item mentioned only by summaries became an item: %+v", items["S: x"])
	}
}

func TestSummariesFlagAMissingSummaryOnTheBoundary(t *testing.T) {
	for _, tc := range []struct {
		name, top string
		want      bool
	}{
		{"missing", "[A]\n> a\n[B]\n    < x\n    |S|\n", true},
		{"drawn", "[A]\n    > x\n    |S|\n> a\n[B]\n    < x\n    |S|\n", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, fs := summariesOf(t, map[string]string{"flow.dfd": tc.top, "flow.1.dfd": "[A1]\n    > x\n    |S|\n"})
			if got := hasFinding(fs, dfdmetrics.MissingSummary, "S", "x", "1"); got != tc.want {
				t.Errorf("missing summary on box 1 = %v, want %v; findings %+v", got, tc.want, fs)
			}
			if hasFinding(fs, dfdmetrics.UnmatchedSummary, "S", "x", "") {
				t.Errorf("findings = %+v, want no unmatched summary", fs)
			}
		})
	}
}

func TestSummariesFlagStateDrawnAboveItsHome(t *testing.T) {
	for _, tc := range []struct {
		name  string
		child string
		label string
	}{
		{"shared inside the child", "[A1]\n    > y\n    |S|\n> b\n[A2]\n    < y\n    |S|\n", "y"},
		{"private to one box", "[A1]\n    > y\n    < y\n    |S|\n", "y"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, fs := summariesOf(t, map[string]string{"flow.dfd": "[A]\n    > y\n    |S|\n", "flow.1.dfd": tc.child})
			if !hasFinding(fs, dfdmetrics.SummaryAboveHome, "S", tc.label, "1") {
				t.Errorf("findings = %+v, want a summary above the home on box 1", fs)
			}
		})
	}
}

func TestPressuresCountTheItemsLiveAcrossEachGap(t *testing.T) {
	d := load(t, map[string]string{
		"flow.dfd": `
[A]
    > x
    |S|
> a
[B]
    > y
    |S|
> b
[C]
> c
[D]
    < x
    |S|
    < y
    |S|
`,
		"flow.3.dfd": "[C1]\n",
	})
	items, _ := dfdmetrics.Items(d)
	ps := dfdmetrics.Pressures(d, items)
	if len(ps) != 1 {
		t.Fatalf("pressures = %+v, want one for the top diagram only", ps)
	}
	if p := ps[0]; p.Diagram != d.Diagrams[0] || p.Max != 2 || p.Gaps != 2 || p.After.Number != "2" {
		t.Errorf("pressure = max %d in %d gaps after box %s, want max 2 in 2 gaps after box 2", p.Max, p.Gaps, p.After.Number)
	}
}
