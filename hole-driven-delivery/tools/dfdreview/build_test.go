package dfdreview_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/internal/gittest"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/page"
)

// echo is a stand-in for dfd that writes its input as an SVG comment.
const echo = "#!/bin/sh\nprintf '<svg><!--'\ncat\nprintf -- '--></svg>\\n'\n"

// readPage returns the data of the review page at path.
func readPage(t *testing.T, path string) page.Data {
	t.Helper()
	html, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	const open = `<script id="data" type="application/json">`
	_, rest, ok := strings.Cut(string(html), open)
	blob, _, closed := strings.Cut(rest, "</script>")
	if !ok || !closed {
		t.Fatalf("the page has no data script:\n%s", html)
	}
	var d page.Data
	if err := json.Unmarshal([]byte(blob), &d); err != nil {
		t.Fatalf("the page's data: %v", err)
	}
	return d
}

func TestBuild(t *testing.T) {
	g := gittest.New(t)
	g.Write(map[string]string{
		"docs/flow.dfd":         "[1. Read the items\n (lib.Read)]\n> items\n[2. Check]\n# type: items = []lib.Item\n",
		"docs/vocabulary.md":    "- item: one thing\n",
		"docs/plans/plan.md":    "# The plan\n\n## The change in brief\n\nIt reads.\n\n## Metaphor\n\nA packet.\n",
		"docs/review/score.tsv": "live_range\tfindings\n2\t1\n",
		"lib/lib.go":            "package lib\n\n// Read reads.\nfunc Read() {}\n\ntype Item int\n",
	})
	base := g.Commit("base")
	g.Write(map[string]string{
		"docs/flow.dfd":             "[1. Read the items\n (lib.Read)]\n> items\n[2. Check them]\n# type: items = []lib.Item\n",
		"docs/review/score.new.tsv": "live_range\tfindings\n1\t1\n",
		"docs/review/report.txt":    "findings\n  none\n",
		"lib/lib.go":                "package lib\n\n// Read reads the items.\nfunc Read() {}\n\ntype Item int\n",
	})
	dfd := filepath.Join(t.TempDir(), "dfd")
	if err := os.WriteFile(dfd, []byte(echo), 0o755); err != nil {
		t.Fatal(err)
	}
	docs := filepath.Join(g.Dir, "docs")
	opts := dfdreview.Options{
		Design:     filepath.Join(docs, "flow.dfd"),
		Base:       base,
		Vocabulary: filepath.Join(docs, "vocabulary.md"),
		Plan:       filepath.Join(docs, "plans", "plan.md"),
		OldScore:   filepath.Join(docs, "review", "score.tsv"),
		NewScore:   filepath.Join(docs, "review", "score.new.tsv"),
		Report:     filepath.Join(docs, "review", "report.txt"),
		DFD:        dfd,
		Out:        filepath.Join(docs, "review", "index.html"),
	}
	// The second build sees the first page as a new file, and leaves it out.
	for range 2 {
		if err := dfdreview.Build(opts); err != nil {
			t.Fatal(err)
		}
	}
	d := readPage(t, opts.Out)
	if d.Title != "The plan" || d.Brief != "It reads." || d.Metaphor != "A packet." || d.Report != "findings\n  none\n" {
		t.Errorf("notes: title %q, brief %q, metaphor %q, report %q", d.Title, d.Brief, d.Metaphor, d.Report)
	}
	if d.OldScore == nil || d.NewScore == nil || d.OldScore.Values[0] != "2" || d.NewScore.Values[0] != "1" {
		t.Errorf("score cards: old %+v, new %+v", d.OldScore, d.NewScore)
	}
	if len(d.Diagrams) != 1 {
		t.Fatalf("got %d diagrams, want 1", len(d.Diagrams))
	}
	top := d.Diagrams[0]
	if top.Title != "Overview" || top.Before == "" || top.Diff == "" || top.After == "" {
		t.Errorf("overview: %q, views drawn %v %v %v", top.Title, top.Before != "", top.Diff != "", top.After != "")
	}
	if top.Types.After["items"] != "[]lib.Item" {
		t.Errorf("types = %+v", top.Types)
	}
	if d.Terms.After["item"] != "one thing" {
		t.Errorf("terms = %+v", d.Terms)
	}
	read := d.Decls["lib.Read"]
	if read.Before == nil || read.After == nil || read.After.File != "lib/lib.go" || read.After.Start != 3 {
		t.Errorf("lib.Read declared at %+v", read)
	}
	if !read.Changed || d.Decls["lib.Item"].Changed {
		t.Errorf("changed: lib.Read %v, want true; lib.Item %v, want false", read.Changed, d.Decls["lib.Item"].Changed)
	}
	lib := d.Files["lib/lib.go"]
	if lib.Before == nil || lib.After == nil || !strings.Contains(lib.Diff, "+// Read reads the items.") {
		t.Errorf("lib/lib.go = %+v", lib)
	}
	var changed []string
	for _, c := range d.Changed {
		changed = append(changed, c.Path+" "+c.Status)
	}
	want := "docs/flow.dfd modified, docs/review/report.txt added, docs/review/score.new.tsv added, lib/lib.go modified"
	if got := strings.Join(changed, ", "); got != want {
		t.Errorf("changed = %s\nwant      %s", got, want)
	}
	if _, ok := d.Files["docs/review/index.html"]; ok {
		t.Errorf("the page holds itself")
	}
}
