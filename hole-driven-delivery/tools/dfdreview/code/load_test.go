package code_test

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/code"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/design"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/internal/gittest"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/repo"
)

const goMod = "module example.com/m\n\ngo 1.22\n"

// pGo is package p at step n: T.M returns n, and U.M, of the same name, stays the same.
func pGo(n int) string {
	return fmt.Sprintf(`package p

// I has M.
type I interface{ M() int }

// T is one implementation of I.
type T struct{}

// M returns the step.
func (T) M() int { return %d }

// U is another implementation of I.
type U struct{}

// M returns zero.
func (U) M() int { return 0 }

// F calls T.M.
func F() int { return T{}.M() }

// G calls U.M.
func G() int { return U{}.M() }

// H calls M through I.
func H(i I) int { return i.M() }
`, n)
}

// qGo calls p.F from another package.
const qGo = `package q

import "example.com/m/p"

// K calls p.F.
func K() int { return p.F() }
`

// module makes a module whose p.T.M changes after the base, and opens it at the base.
func module(t *testing.T) (*repo.Repo, string) {
	t.Helper()
	g := gittest.New(t)
	g.Write(map[string]string{"go.mod": goMod, "docs/flow.dfd": "[1. Run\n (p.F)]\n", "p/p.go": pGo(1), "q/q.go": qGo})
	base := g.Commit("base")
	g.Write(map[string]string{"p/p.go": pGo(2)})
	r, err := repo.Open(g.Dir, base)
	if err != nil {
		t.Fatal(err)
	}
	return r, g.Dir
}

// index indexes the code of the design at dir/docs/flow.dfd.
func index(t *testing.T, r *repo.Repo, dir string) *code.Index {
	t.Helper()
	d, err := design.Read(r, filepath.Join(dir, "docs", "flow.dfd"), "")
	if err != nil {
		t.Fatal(err)
	}
	c, err := code.Read(r, d)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestLoadResolvesCalls(t *testing.T) {
	tmp := t.TempDir()
	r, dir := module(t)
	t.Setenv("TMPDIR", tmp)
	t.Setenv("GOFLAGS", "-mod=mod") // the load must not update go.mod even so
	res, err := code.Load(r)
	if err != nil {
		t.Fatal(err)
	}
	if res.Before == nil || res.After == nil {
		t.Fatalf("resolutions: before %v, after %v", res.Before != nil, res.After != nil)
	}
	for key, want := range map[string][]string{
		"p.F": {"p.T.M"},
		"p.G": {"p.U.M"},
		"p.H": {"p.T.M", "p.U.M"}, // both types implement I
		"q.K": nil,                // p.F lies in another package
	} {
		got := append([]string(nil), res.After.Calls[key]...)
		sort.Strings(got)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s calls %v, want %v", key, got, want)
		}
	}
	if res.Before.Files["p/p.go"] != pGo(1) || res.After.Files["p/p.go"] != pGo(2) {
		t.Errorf("p/p.go as parsed: before %q, after %q", res.Before.Files["p/p.go"], res.After.Files["p/p.go"])
	}
	if mod, err := os.ReadFile(filepath.Join(dir, "go.mod")); err != nil || string(mod) != goMod {
		t.Errorf("the load changed go.mod to %q (%v)", mod, err)
	}
	if left, err := os.ReadDir(tmp); err != nil || len(left) > 0 {
		t.Errorf("the temporary directory holds %d entries after the load (%v)", len(left), err)
	}
}

func TestLoadWithoutModule(t *testing.T) {
	g := gittest.New(t)
	g.Write(map[string]string{"docs/flow.dfd": "[1. Run]\n", "lib/lib.go": "package lib\n"})
	base := g.Commit("base")
	r, err := repo.Open(g.Dir, base)
	if err != nil {
		t.Fatal(err)
	}
	res, err := code.Load(r)
	if err != nil {
		t.Fatal(err)
	}
	if res.Before != nil || res.After != nil || len(res.Errors) > 0 {
		t.Errorf("without a module: %+v", res)
	}
}

func TestLoadReportsPackageErrors(t *testing.T) {
	g := gittest.New(t)
	g.Write(map[string]string{
		"go.mod":        goMod,
		"docs/flow.dfd": "[1. Run]\n",
		"p/p.go":        "package p\n\n// F returns a string as an int.\nfunc F() int { return \"s\" }\n\n// G returns one.\nfunc G() int { return 1 }\n",
	})
	base := g.Commit("base")
	r, err := repo.Open(g.Dir, base)
	if err != nil {
		t.Fatal(err)
	}
	res, err := code.Load(r)
	if err != nil {
		t.Fatal(err)
	}
	all := strings.Join(res.Errors, "\n")
	if len(res.Errors) != 2 || !strings.Contains(all, "base") || !strings.Contains(all, "working tree") || strings.Count(all, "p/p.go") != 2 {
		t.Errorf("package errors:\n%s", all)
	}
	if _, ok := res.After.Calls["p.G"]; !ok {
		t.Errorf("the package with an error lost its functions: %v", res.After.Calls)
	}
}

func TestLoadLinksIdentifiers(t *testing.T) {
	g := gittest.New(t)
	g.Write(map[string]string{
		"go.mod":        goMod,
		"docs/flow.dfd": "[1. Run]\n",
		"p/a.go": "package p\n\nimport \"strings\"\n\n// A trims s.\n" +
			"func A(s string) string { return strings.TrimSpace(s) + b }\n\n" +
			"var c = \"\u00e9\" + b\n\n" +
			"func w(sb *strings.Builder) { sb.WriteString(\"x\") }\n",
		"p/b.go": "package p\n\nvar b = \"x\"\n",
		"q/q.go": "package q\n\nimport \"example.com/m/p\"\n\nvar d = p.A\n",
	})
	base := g.Commit("base")
	// The working tree moves b down two lines.
	g.Write(map[string]string{"p/b.go": "package p\n\n// b ends each result.\n\nvar b = \"x\"\n"})
	r, err := repo.Open(g.Dir, base)
	if err != nil {
		t.Fatal(err)
	}
	res, err := code.Load(r)
	if err != nil {
		t.Fatal(err)
	}
	links := func(b int) []code.Link {
		return []code.Link{
			{Line: 6, Col: 34, Len: 7, URL: "https://pkg.go.dev/strings"},
			{Line: 6, Col: 42, Len: 9, URL: "https://pkg.go.dev/strings#TrimSpace"},
			{Line: 6, Col: 52, Len: 1, File: "p/a.go", To: 6},
			{Line: 6, Col: 57, Len: 1, File: "p/b.go", To: b},
			{Line: 8, Col: 15, Len: 1, File: "p/b.go", To: b}, // after a two-byte letter, one UTF-16 unit
			{Line: 10, Col: 12, Len: 7, URL: "https://pkg.go.dev/strings"},
			{Line: 10, Col: 20, Len: 7, URL: "https://pkg.go.dev/strings#Builder"},
			{Line: 10, Col: 31, Len: 2, File: "p/a.go", To: 10},
			{Line: 10, Col: 34, Len: 11, URL: "https://pkg.go.dev/strings#Builder.WriteString"},
		}
	}
	if got, want := res.Before.Links["p/a.go"], links(3); !reflect.DeepEqual(got, want) {
		t.Errorf("before:\n got %+v\nwant %+v", got, want)
	}
	if got, want := res.After.Links["p/a.go"], links(5); !reflect.DeepEqual(got, want) {
		t.Errorf("after:\n got %+v\nwant %+v", got, want)
	}
	// A package of the code directory opens at the package clause of its first file.
	q := []code.Link{{Line: 5, Col: 9, Len: 1, File: "p/a.go", To: 1}, {Line: 5, Col: 11, Len: 1, File: "p/a.go", To: 6}}
	for _, v := range []*code.Resolution{res.Before, res.After} {
		if got := v.Links["q/q.go"]; !reflect.DeepEqual(got, q) {
			t.Errorf("q/q.go:\n got %+v\nwant %+v", got, q)
		}
	}
}

func TestReadMarksByReach(t *testing.T) {
	t.Skip("HOLE(3): mark each function whose reach holds a changed function or method")
	r, dir := module(t)
	c := index(t, r, dir)
	for key, want := range map[string]bool{
		"p.F": true,  // reaches T.M, which changed
		"p.G": false, // reaches U.M only, though T.M shares its name
		"p.H": true,  // reaches T.M through I
		"q.K": false, // p.F lies in another package
		"p.T": false, // a type keeps its own rule
	} {
		if got := c.ChangedDecls[key]; got != want {
			t.Errorf("%s changed: %v, want %v", key, got, want)
		}
	}
}

func TestReadMarksByReachWithoutBaseModule(t *testing.T) {
	t.Skip("HOLE(3): the reach marks functions when only the working tree has a module")
	g := gittest.New(t)
	g.Write(map[string]string{"docs/flow.dfd": "[1. Run\n (p.F)]\n", "p/p.go": pGo(1)})
	base := g.Commit("base")
	g.Write(map[string]string{"go.mod": goMod, "p/p.go": pGo(2)})
	r, err := repo.Open(g.Dir, base)
	if err != nil {
		t.Fatal(err)
	}
	c := index(t, r, g.Dir)
	for key, want := range map[string]bool{"p.F": true, "p.G": false, "p.H": true} {
		if got := c.ChangedDecls[key]; got != want {
			t.Errorf("%s changed: %v, want %v", key, got, want)
		}
	}
}

func TestReadHoldsEveryGoFile(t *testing.T) {
	t.Skip("HOLE(3): hold every Go file of both versions, test files included")
	g := gittest.New(t)
	g.Write(map[string]string{
		"go.mod":        goMod,
		"docs/flow.dfd": "[1. Run\n (p.F)]\n",
		"p/p.go":        "package p\n\n// F returns one.\nfunc F() int { return 1 }\n",
		"p/p_test.go":   "package p\n\nimport \"testing\"\n\nfunc TestF(t *testing.T) {}\n",
		"q/q.go":        "package q\n",
		"q/r.go":        "package q\n\n// R is not named by any box.\nfunc R() {}\n",
	})
	base := g.Commit("base")
	g.Write(map[string]string{"q/q.go": "package q\n\n// K is new.\nfunc K() {}\n"})
	r, err := repo.Open(g.Dir, base)
	if err != nil {
		t.Fatal(err)
	}
	c := index(t, r, g.Dir)
	for _, path := range []string{"p/p.go", "p/p_test.go", "q/q.go", "q/r.go"} {
		if f, ok := c.Files[path]; !ok || !f.Before.Found || !f.After.Found {
			t.Errorf("%s: %+v", path, f)
		}
	}
}

func TestLoadLeavesGoModAlone(t *testing.T) {
	// Under -mod=mod, the go command adds a go directive to a go.mod without one.
	const mod = "module example.com/m\n"
	g := gittest.New(t)
	g.Write(map[string]string{"go.mod": mod, "docs/flow.dfd": "[1. Run]\n", "p/p.go": "package p\n\n// F does nothing.\nfunc F() {}\n"})
	base := g.Commit("base")
	r, err := repo.Open(g.Dir, base)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOFLAGS", "-mod=mod")
	res, err := code.Load(r)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := os.ReadFile(filepath.Join(g.Dir, "go.mod")); err != nil || string(got) != mod {
		t.Errorf("the load changed go.mod to %q (%v)", got, err)
	}
	if res.Before == nil || res.After == nil || len(res.Errors) > 0 {
		t.Errorf("resolutions: before %v, after %v, errors %v", res.Before != nil, res.After != nil, res.Errors)
	}
}
