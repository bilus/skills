package notes_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/notes"
)

func write(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRead(t *testing.T) {
	dir := t.TempDir()
	plan := write(t, dir, "plan.md", "# Plan\n\n## The change in brief\n\nA brief\nover two lines.\n\n## Metaphor\n\nA light table.\n\n### Rules\n\nBoth sheets match.\n\n## Stages\n\nNot this.\n")
	old := write(t, dir, "old.tsv", "live_range\tread_distance\n3\t1\n")
	fresh := write(t, dir, "new.tsv", "live_range\tread_distance\n2\t1\n")
	report := write(t, dir, "report.txt", "findings\n  none\n")
	n, err := notes.Read(plan, old, fresh, report)
	if err != nil {
		t.Fatal(err)
	}
	want := &notes.Notes{
		Title:    "Plan",
		Brief:    "A brief\nover two lines.",
		Metaphor: "A light table.\n\n### Rules\n\nBoth sheets match.",
		OldScore: &notes.Card{Columns: []string{"live_range", "read_distance"}, Values: []string{"3", "1"}},
		NewScore: &notes.Card{Columns: []string{"live_range", "read_distance"}, Values: []string{"2", "1"}},
		Report:   "findings\n  none\n",
	}
	if !reflect.DeepEqual(n, want) {
		t.Errorf("notes = %+v\nwant    %+v", n, want)
	}
}

func TestReadWithoutInputs(t *testing.T) {
	n, err := notes.Read("", filepath.Join(t.TempDir(), "score.tsv"), "", "")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(n, &notes.Notes{}) {
		t.Errorf("notes = %+v, want none, since the cached card does not exist yet", n)
	}
}

func TestReadErrors(t *testing.T) {
	dir := t.TempDir()
	short := write(t, dir, "short.tsv", "live_range\n")
	uneven := write(t, dir, "uneven.tsv", "a\tb\n1\n")
	missing := filepath.Join(dir, "missing")
	cases := []struct {
		name                  string
		plan, old, fresh, rep string
		want                  string
	}{
		{"missing plan", missing, "", "", "", "missing"},
		{"missing new card", "", "", missing, "", "missing"},
		{"missing report", "", "", "", missing, "missing"},
		{"one line", "", "", short, "", "short.tsv: want a header line and a line of values"},
		{"uneven", "", "", uneven, "", "uneven.tsv: 2 columns in the header, 1 in the values"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := notes.Read(c.plan, c.old, c.fresh, c.rep)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Errorf("error = %v, want one containing %q", err, c.want)
			}
		})
	}
}
