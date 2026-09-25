package repo_test

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/internal/gittest"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/repo"
)

func open(t *testing.T, dir, base string) *repo.Repo {
	t.Helper()
	r, err := repo.Open(dir, base)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestReadInBothVersions(t *testing.T) {
	g := gittest.New(t)
	g.Write(map[string]string{"docs/a.dfd": "old a\n", "docs/b.dfd": "b\n"})
	base := g.Commit("base")
	g.Write(map[string]string{"docs/a.dfd": "new a\n", "docs/c.dfd": "c\n"})
	g.Remove("docs/b.dfd")
	r := open(t, g.Dir, base)
	cases := []struct {
		name          string
		before, after repo.Text
	}{
		{"a.dfd", repo.Text{Content: "old a\n", Found: true}, repo.Text{Content: "new a\n", Found: true}},
		{"b.dfd", repo.Text{Content: "b\n", Found: true}, repo.Text{}},
		{"c.dfd", repo.Text{}, repo.Text{Content: "c\n", Found: true}},
		{"missing.dfd", repo.Text{}, repo.Text{}},
	}
	for _, c := range cases {
		p, err := r.Read(filepath.Join(g.Dir, "docs", c.name))
		if err != nil {
			t.Fatalf("%s: %v", c.name, err)
		}
		if p.Before != c.before || p.After != c.after {
			t.Errorf("%s: read %+v, want before %+v and after %+v", c.name, p, c.before, c.after)
		}
	}
}

func TestReadWithoutBase(t *testing.T) {
	g := gittest.New(t)
	g.Write(map[string]string{"a.dfd": "a\n"})
	g.Commit("base")
	p, err := open(t, g.Dir, "").Read(filepath.Join(g.Dir, "a.dfd"))
	if err != nil {
		t.Fatal(err)
	}
	if p.Before.Found || p.After != (repo.Text{Content: "a\n", Found: true}) {
		t.Errorf("read %+v, want the working tree alone", p)
	}
}

func TestOpenErrors(t *testing.T) {
	if _, err := repo.Open(t.TempDir(), ""); err == nil || !strings.Contains(err.Error(), "not in a git repository") {
		t.Errorf("outside a repository: error %v", err)
	}
	g := gittest.New(t)
	g.Write(map[string]string{"a": "a\n"})
	g.Commit("base")
	if _, err := repo.Open(g.Dir, "no-such-rev"); err == nil || !strings.Contains(err.Error(), `"no-such-rev"`) {
		t.Errorf("unknown base: error %v", err)
	}
}

func TestNames(t *testing.T) {
	g := gittest.New(t)
	g.Write(map[string]string{"docs/flow.dfd": "", "docs/flow.2.dfd": "", "docs/sub/x.dfd": ""})
	base := g.Commit("base")
	g.Write(map[string]string{"docs/flow.3.dfd": ""})
	g.Remove("docs/flow.2.dfd")
	names, err := open(t, g.Dir, base).Names(filepath.Join(g.Dir, "docs"))
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"flow.2.dfd", "flow.3.dfd", "flow.dfd"}; !reflect.DeepEqual(names, want) {
		t.Errorf("names = %v, want %v", names, want)
	}
}

func TestGoFiles(t *testing.T) {
	g := gittest.New(t)
	g.Write(map[string]string{
		"tool/a.go": "package a // old\n", "tool/a_test.go": "package a\n",
		"tool/testdata/t.go": "package t\n", "tool/.hidden/h.go": "package h\n",
		"tool/vendor/v/v.go": "package v\n", "other/o.go": "package o\n",
	})
	base := g.Commit("base")
	g.Write(map[string]string{"tool/a.go": "package a // new\n", "tool/b/b.go": "package b\n"})
	before, after, err := open(t, filepath.Join(g.Dir, "tool"), base).GoFiles()
	if err != nil {
		t.Fatal(err)
	}
	if want := map[string]string{"a.go": "package a // old\n"}; !reflect.DeepEqual(before, want) {
		t.Errorf("before = %v, want %v", before, want)
	}
	if want := map[string]string{"a.go": "package a // new\n", "b/b.go": "package b\n"}; !reflect.DeepEqual(after, want) {
		t.Errorf("after = %v, want %v", after, want)
	}
}

func TestChanged(t *testing.T) {
	g := gittest.New(t)
	g.Write(map[string]string{"tool/kept.go": "same\n", "tool/edited.go": "old\n", "tool/gone.go": "gone\n", "other/x.go": "x\n"})
	base := g.Commit("base")
	g.Write(map[string]string{"tool/edited.go": "new\n", "tool/fresh.go": "fresh\n", "other/x.go": "y\n"})
	g.Remove("tool/gone.go")
	changed, err := open(t, filepath.Join(g.Dir, "tool"), base).Changed()
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, c := range changed {
		got = append(got, c.Path+" "+c.Status)
	}
	if want := []string{"edited.go modified", "fresh.go added", "gone.go deleted"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("changed = %v, want %v", got, want)
	}
	for i, want := range [][]string{{"-old", "+new"}, {"--- /dev/null", "+fresh"}, {"-gone", "+++ /dev/null"}} {
		for _, w := range want {
			if !strings.Contains(changed[i].Diff, w) {
				t.Errorf("%s: diff lacks %q:\n%s", changed[i].Path, w, changed[i].Diff)
			}
		}
	}
	none, err := open(t, filepath.Join(g.Dir, "tool"), "").Changed()
	if err != nil || len(none) != 0 {
		t.Errorf("without a base: %v, %v; want no changes", none, err)
	}
}
