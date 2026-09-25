package dfdmetrics_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/bilus/skills/hole-driven-delivery/tools/dfdmetrics"
)

// std is a fixed stand-in for the standard library set.
var std = map[string]bool{"fmt": true, "os": true, "strings": true}

func TestSpreadsCountOnlyDesignPackages(t *testing.T) {
	d := load(t, map[string]string{
		"flow.dfd": `
[Walk
 (analyze.Run, fmt.Errorf, go/format.Source,
  golang.org/x/tools/go/packages.Load, scope.Names)]
> a
[Render (render.site, strings.Builder)]
`,
		"flow.2.dfd": "[R1 (analyze.A)]\n> b\n[R2 (scope.B)]\n",
	})
	spreads, fs := dfdmetrics.Spreads(d, std)
	if got := strings.Join(spreads[0].Packages, " "); got != "analyze render scope" {
		t.Errorf("top packages = %q, want analyze render scope", got)
	}
	var mixedBoxes, mixedDiagrams []string
	for _, f := range fs {
		switch f.Kind {
		case dfdmetrics.MixedBox:
			mixedBoxes = append(mixedBoxes, f.Box.Number+" "+strings.Join(f.Packages, ","))
		case dfdmetrics.MixedDiagram:
			mixedDiagrams = append(mixedDiagrams, f.Diagram.Number+" "+strings.Join(f.Packages, ","))
		}
	}
	if got := strings.Join(mixedBoxes, "; "); got != "1 analyze,scope" {
		t.Errorf("mixed boxes = %q, want box 1 with analyze and scope only", got)
	}
	if got := strings.Join(mixedDiagrams, "; "); got != "2 analyze,scope" {
		t.Errorf("mixed diagrams = %q, want diagram 2 only, since the top diagram is exempt", got)
	}
}

func TestReportGroupsItemsByHomeAndRanksThem(t *testing.T) {
	d := load(t, map[string]string{
		"flow.dfd": `
[A]
    > x
    |S|
> a
[B]
> b
[C]
    < x
    |S|
    > p
    < p
    |S|
`,
		"flow.2.dfd": `
[B1]
    > y
    |S|
    > z
    |S|
> c
[B2]
    < z
    |S|
> d
[B3]
    < y
    |S|
`,
	})
	var out bytes.Buffer
	if err := dfdmetrics.Analyze(d, std).Write(&out); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	var at []int
	for _, line := range []string{"items at home flow.dfd, level 0", "S: x", "items at home flow.2.dfd, level 1", "S: y", "S: z", "private items", "S: p"} {
		i := strings.Index(text, line)
		if i < 0 {
			t.Fatalf("report lacks %q:\n%s", line, text)
		}
		at = append(at, i)
	}
	for i := 1; i < len(at); i++ {
		if at[i] < at[i-1] {
			t.Fatalf("report lines out of order:\n%s", text)
		}
	}
}

func TestReportOnTheFlatGoxFlow(t *testing.T) {
	d, err := dfdmetrics.Load("testdata/flow.dfd")
	if err != nil {
		t.Fatal(err)
	}
	r := dfdmetrics.Analyze(d, std)
	if !hasFinding(r.Findings, dfdmetrics.ArrowCandidate, "analyze.Package", "sum types", "5") {
		t.Error(`"sum types" on analyze.Package is not an arrow candidate after box 5`)
	}
	if !hasFinding(r.Findings, dfdmetrics.OrphanItem, "analyze.File", "matches", "") {
		t.Error(`"matches" on analyze.File is not an orphan item`)
	}
	for _, it := range r.Items {
		if it.Store == "analyze.Package" && it.Label == "sum types" && (it.LiveRange != 1 || it.ReadDistance != 1) {
			t.Errorf("sum types: live range %d, read distance %d; want 1 and 1", it.LiveRange, it.ReadDistance)
		}
	}
	var mixed []string
	for _, f := range r.Findings {
		if f.Kind == dfdmetrics.MixedBox {
			mixed = append(mixed, f.Box.Number)
		}
	}
	if got := strings.Join(mixed, " "); got != "8 9 12" {
		t.Errorf("mixed boxes = %s, want 8 9 12", got)
	}

	var out bytes.Buffer
	if err := r.Write(&out); err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("testdata/flow.report")
	if err != nil {
		t.Fatal(err)
	}
	if out.String() != string(want) {
		t.Errorf("report differs from testdata/flow.report:\n%s", out.String())
	}
}

func TestScoreCondensesTheReport(t *testing.T) {
	d, err := dfdmetrics.Load("testdata/flow.dfd")
	if err != nil {
		t.Fatal(err)
	}
	r := dfdmetrics.Analyze(d, std)
	want := dfdmetrics.Score{LiveRange: 7, ReadDistance: 6, TopShared: 3, Shared: 3, Pressure: 1, PackageSpread: 3, Findings: 15}
	if got := r.Score(); got != want {
		t.Errorf("score = %+v, want %+v", got, want)
	}
	var out bytes.Buffer
	if err := r.WriteScore(&out); err != nil {
		t.Fatal(err)
	}
	if want := "live_range\tread_distance\ttop_shared\tshared\tpressure\tpackage_spread\tfindings\n7\t6\t3\t3\t1\t3\t15\n"; out.String() != want {
		t.Errorf("score output = %q, want %q", out.String(), want)
	}
}
