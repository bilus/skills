package dfdmetrics_test

import (
	"path/filepath"
	"testing"

	"github.com/bilus/skills/hole-driven-delivery/tools/dfdmetrics"
)

// itemsOf loads files and returns the items by "store: label", with the findings.
func itemsOf(t *testing.T, files map[string]string) (map[string]*dfdmetrics.Item, []dfdmetrics.Finding) {
	t.Helper()
	items, findings := dfdmetrics.Items(load(t, files))
	byName := map[string]*dfdmetrics.Item{}
	for _, it := range items {
		byName[it.Store+": "+it.Label] = it
	}
	return byName, findings
}

func numbers(boxes []*dfdmetrics.Box) []string {
	var out []string
	for _, b := range boxes {
		out = append(out, b.Number)
	}
	return out
}

func hasFinding(fs []dfdmetrics.Finding, kind dfdmetrics.FindingKind, store, label, box string) bool {
	for _, f := range fs {
		if f.Kind == kind && f.Store == store && f.Label == label && (box == "" || f.Box != nil && f.Box.Number == box) {
			return true
		}
	}
	return false
}

func TestItemsShareStateInsideAChildDiagram(t *testing.T) {
	items, _ := itemsOf(t, map[string]string{
		"flow.dfd":   "[A]\n> x\n[B]\n",
		"flow.2.dfd": "[B1]\n    > y\n    |S|\n> z\n[B2]\n    < y\n    |S|\n",
	})
	it := items["S: y"]
	if it == nil || it.Home == nil {
		t.Fatalf("item S: y = %+v, want a shared item", it)
	}
	if filepath.Base(it.Home.Path) != "flow.2.dfd" || it.Level != 1 || it.LiveRange != 1 || it.ReadDistance != 1 {
		t.Errorf("home %s, level %d, live range %d, read distance %d; want flow.2.dfd, 1, 1, 1", filepath.Base(it.Home.Path), it.Level, it.LiveRange, it.ReadDistance)
	}
}

func TestItemsMeasureALiveRangeThroughAChildDiagram(t *testing.T) {
	items, _ := itemsOf(t, map[string]string{
		"flow.dfd":   "[A]\n    > x\n    |S|\n> a\n[M]\n> b\n[B]\n",
		"flow.3.dfd": "[B1]\n> c\n[B2]\n    < x\n    |S|\n",
	})
	it := items["S: x"]
	if it == nil || it.Home == nil || it.Home.Depth != 0 {
		t.Fatalf("item S: x = %+v, want a shared item homed in the top diagram", it)
	}
	if got := numbers(it.Readers); len(got) != 1 || got[0] != "3.2" {
		t.Errorf("readers = %v, want [3.2]", got)
	}
	if it.LiveRange != 2 || it.ReadDistance != 2 {
		t.Errorf("live range %d, read distance %d; want 2, 2", it.LiveRange, it.ReadDistance)
	}
}

func TestItemsMeasureReadDistanceAgainstTheNearestWrite(t *testing.T) {
	items, _ := itemsOf(t, map[string]string{"flow.dfd": `
[A]
    > x
    |S|
> a
[B]
    < x
    |S|
> b
[C]
> c
[D]
    > x
    |S|
> d
[E]
    < x
    |S|
`})
	it := items["S: x"]
	if !it.HasLiveRange || it.LiveRange != 4 || !it.HasReadDistance || it.ReadDistance != 1 {
		t.Errorf("live range %d (%v), read distance %d (%v); want 4 and 1", it.LiveRange, it.HasLiveRange, it.ReadDistance, it.HasReadDistance)
	}
}

func TestItemsTreatAProcessDrawnTwiceAsTwoBoxes(t *testing.T) {
	items, _ := itemsOf(t, map[string]string{"flow.dfd": "[A]\n    > x\n    |S|\n> a\n[B]\n> b\n[A]\n    < x\n    |S|\n"})
	it := items["S: x"]
	if it.Home == nil || it.LiveRange != 2 {
		t.Errorf("home %v, live range %d; want a shared item with live range 2", it.Home, it.LiveRange)
	}
}

func TestItemsMarkStateInOneBoxPrivate(t *testing.T) {
	items, findings := itemsOf(t, map[string]string{"flow.dfd": "[A]\n    > x\n    < x\n    |S|\n> a\n[B]\n"})
	it := items["S: x"]
	if it.Home != nil || it.HasLiveRange || it.HasReadDistance {
		t.Errorf("item = %+v, want a private item without distances", it)
	}
	if len(findings) != 0 {
		t.Errorf("findings = %+v, want none", findings)
	}
}

func TestItemsReportTheItemFindings(t *testing.T) {
	for _, tc := range []struct {
		name, src string
		kind      dfdmetrics.FindingKind
		box       string
		want      bool
	}{
		{"dead item", "[A]\n    > x\n    |S|\n", dfdmetrics.DeadItem, "", true},
		{"orphan item", "[A]\n    < x\n    |S|\n", dfdmetrics.OrphanItem, "", true},
		{"read before write", "[A]\n    < x\n    |S|\n> a\n[B]\n    > x\n    |S|\n> b\n[C]\n    < x\n    |S|\n", dfdmetrics.ReadBeforeWrite, "1", true},
		{"no candidate after an early read", "[A]\n    < x\n    |S|\n> a\n[B]\n    > x\n    |S|\n> b\n[C]\n    < x\n    |S|\n", dfdmetrics.ArrowCandidate, "", false},
		{"arrow candidate", "[A]\n    > x\n    |S|\n> a\n[B]\n    < x\n    |S|\n", dfdmetrics.ArrowCandidate, "1", true},
		{"entity in the gap", "[A]\n    > x\n    |S|\n> a\n{E}\n> b\n[B]\n    < x\n    |S|\n", dfdmetrics.ArrowCandidate, "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, findings := itemsOf(t, map[string]string{"flow.dfd": tc.src})
			if got := hasFinding(findings, tc.kind, "S", "x", tc.box); got != tc.want {
				t.Errorf("finding %d on S: x present = %v, want %v; findings %+v", tc.kind, got, tc.want, findings)
			}
		})
	}
}

// An external system beside a process holds none of the design's state.
func TestItemsIgnoreExternalSystems(t *testing.T) {
	items, findings := itemsOf(t, map[string]string{"flow.dfd": "[A]\n    > event\n    < ack\n    <Bus>\n> a\n[B]\n"})
	if len(items) != 0 || len(findings) != 0 {
		t.Errorf("items = %+v, findings = %+v, want none", items, findings)
	}
}

func TestItemsFlagAStoreArrowWithoutItems(t *testing.T) {
	_, findings := itemsOf(t, map[string]string{"flow.dfd": "[A]\n    >\n    |S|\n"})
	if !hasFinding(findings, dfdmetrics.UnlabeledArrow, "S", "", "1") {
		t.Errorf("findings = %+v, want an unlabeled arrow on box 1", findings)
	}
}
