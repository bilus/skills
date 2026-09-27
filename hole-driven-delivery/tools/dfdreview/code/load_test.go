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
		"q.K": {"p.F"},            // the reach crosses packages
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
			{Line: 6, Col: 6, Len: 1, File: "p/a.go", To: 5, End: 6, Key: "p.A", In: "p.A"},
			{Line: 6, Col: 34, Len: 7, URL: "https://pkg.go.dev/strings", In: "p.A"},
			{Line: 6, Col: 42, Len: 9, URL: "https://pkg.go.dev/strings#TrimSpace", In: "p.A"},
			{Line: 6, Col: 52, Len: 1, File: "p/a.go", To: 6, In: "p.A"},
			{Line: 6, Col: 57, Len: 1, File: "p/b.go", To: b, Key: "p.b", In: "p.A"},
			{Line: 8, Col: 5, Len: 1, File: "p/a.go", To: 8, End: 8, Key: "p.c", In: "p.c"},
			{Line: 8, Col: 15, Len: 1, File: "p/b.go", To: b, Key: "p.b", In: "p.c"}, // after a two-byte letter, one UTF-16 unit
			{Line: 10, Col: 6, Len: 1, File: "p/a.go", To: 10, End: 10, Key: "p.w", In: "p.w"},
			{Line: 10, Col: 12, Len: 7, URL: "https://pkg.go.dev/strings", In: "p.w"},
			{Line: 10, Col: 20, Len: 7, URL: "https://pkg.go.dev/strings#Builder", In: "p.w"},
			{Line: 10, Col: 31, Len: 2, File: "p/a.go", To: 10, In: "p.w"},
			{Line: 10, Col: 34, Len: 11, URL: "https://pkg.go.dev/strings#Builder.WriteString", In: "p.w"},
		}
	}
	if got, want := res.Before.Links["p/a.go"], links(3); !reflect.DeepEqual(got, want) {
		t.Errorf("before:\n got %+v\nwant %+v", got, want)
	}
	if got, want := res.After.Links["p/a.go"], links(5); !reflect.DeepEqual(got, want) {
		t.Errorf("after:\n got %+v\nwant %+v", got, want)
	}
	// A package of the code directory opens at the package clause of its first file.
	q := []code.Link{
		{Line: 5, Col: 5, Len: 1, File: "q/q.go", To: 5, End: 5, Key: "q.d", In: "q.d"},
		{Line: 5, Col: 9, Len: 1, File: "p/a.go", To: 1, In: "q.d"},
		{Line: 5, Col: 11, Len: 1, File: "p/a.go", To: 6, Key: "p.A", Funcs: []string{"p.A"}, In: "q.d"},
	}
	for _, v := range []*code.Resolution{res.Before, res.After} {
		if got := v.Links["q/q.go"]; !reflect.DeepEqual(got, q) {
			t.Errorf("q/q.go:\n got %+v\nwant %+v", got, q)
		}
	}
}

func TestReadMarksByReach(t *testing.T) {
	r, dir := module(t)
	c := index(t, r, dir)
	for key, want := range map[string]bool{
		"p.F": true,  // reaches T.M, which changed
		"p.G": false, // reaches U.M only, though T.M shares its name
		"p.H": true,  // reaches T.M through I
		"q.K": true,  // it reaches T.M through p.F, in another package
		"p.T": false, // a type keeps its own rule
	} {
		if got := c.ChangedDecls[key]; got != want {
			t.Errorf("%s changed: %v, want %v", key, got, want)
		}
	}
}

func TestReadMarksByReachWithoutBaseModule(t *testing.T) {
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

// chainGo is package p at step n, with the call chain A, B, C: only C changes between steps.
func chainGo(n int) string {
	return fmt.Sprintf(`package p

// A calls B.
func A() int { return B() }

// B calls C and D.
func B() int { return C() + D() }

// C returns the step.
func C() int { return %d }

// D stays the same.
func D() int { return 0 }
`, n)
}

// TestLoadLinksDeclarations checks the links of the names of declarations and methods: each
// links to the lines of its own declaration, doc comment included.
func TestLoadLinksDeclarations(t *testing.T) {
	g := gittest.New(t)
	g.Write(map[string]string{
		"go.mod":        goMod,
		"docs/flow.dfd": "[1. Run]\n",
		"p/p.go": "package p\n\n// Shape has an area.\ntype Shape interface {\n\t// Area returns the area.\n\tArea() int\n}\n\n" +
			"type (\n\t// Sq is a square.\n\tSq struct{ S int }\n\tn int\n)\n\n" +
			"// Area returns the area of a square.\nfunc (s Sq) Area() int { return s.S * s.S }\n\nconst k, _ = 1, 2\n",
	})
	r, err := repo.Open(g.Dir, "")
	if err != nil {
		t.Fatal(err)
	}
	res, err := code.Load(r)
	if err != nil {
		t.Fatal(err)
	}
	var got []code.Link
	for _, l := range res.After.Links["p/p.go"] {
		if l.End > 0 {
			got = append(got, l)
		}
	}
	// A blank name, a field and a parameter have no key, and so no link of their own.
	want := []code.Link{
		{Line: 4, Col: 6, Len: 5, File: "p/p.go", To: 3, End: 7, Key: "p.Shape", In: "p.Shape"},
		{Line: 6, Col: 2, Len: 4, File: "p/p.go", To: 5, End: 6, Key: "p.Shape.Area", In: "p.Shape"}, // an interface's method
		{Line: 11, Col: 2, Len: 2, File: "p/p.go", To: 10, End: 11, Key: "p.Sq", In: "p.Sq"},         // a spec of a group
		{Line: 12, Col: 2, Len: 1, File: "p/p.go", To: 12, End: 12, Key: "p.n", In: "p.n"},
		{Line: 16, Col: 13, Len: 4, File: "p/p.go", To: 15, End: 16, Key: "p.Sq.Area", In: "p.Sq.Area"},
		{Line: 18, Col: 7, Len: 1, File: "p/p.go", To: 18, End: 18, Key: "p.k", In: "p.k"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("the links of the names:\n got %+v\nwant %+v", got, want)
	}
}

func TestReadMarksKindsAndLinks(t *testing.T) {
	g := gittest.New(t)
	g.Write(map[string]string{
		"go.mod":        goMod,
		"docs/flow.dfd": "[1. Run\n (p.A)]\n",
		"p/p.go":        chainGo(1),
		"q/q.go":        "package q\n\nimport \"example.com/m/p\"\n\n// E calls p.A.\nfunc E() int { return p.A() }\n",
	})
	base := g.Commit("base")
	g.Write(map[string]string{"p/p.go": chainGo(2)})
	r, err := repo.Open(g.Dir, base)
	if err != nil {
		t.Fatal(err)
	}
	c := index(t, r, g.Dir)
	for key, want := range map[string]string{"p.A": code.MarkReached, "p.B": code.MarkReached, "p.C": code.MarkChanged, "p.D": "", "q.E": code.MarkReached} {
		got := ""
		switch {
		case c.Reached[key]:
			got = code.MarkReached
		case c.ChangedDecls[key]:
			got = code.MarkChanged
		}
		if got != want {
			t.Errorf("%s: mark %q, want %q", key, got, want)
		}
	}
	marks := map[string]string{}
	for path, links := range c.AfterLinks {
		for _, l := range links {
			if l.Mark != "" {
				marks[fmt.Sprintf("%s:%d:%d", path, l.Line, l.Col)] = l.Mark
			}
		}
	}
	want := map[string]string{
		"p/p.go:4:23": code.MarkReached, // B in A
		"p/p.go:7:23": code.MarkChanged, // C in B; D, at 7:29, stays unmarked
		"q/q.go:6:25": code.MarkReached, // p.A in E, across packages
	}
	if !reflect.DeepEqual(marks, want) {
		t.Errorf("marked links %v, want %v", marks, want)
	}
}

func TestReadReachesImplementationsInOtherPackages(t *testing.T) {
	shapes := func(area string) map[string]string {
		return map[string]string{
			"go.mod":         goMod,
			"docs/flow.dfd":  "[1. Sum the areas\n (use.Total)]\n",
			"iface/iface.go": "package iface\n\n// Shape has an area.\ntype Shape interface{ Area() int }\n",
			"sq/sq.go":       "package sq\n\n// Square is a shape.\ntype Square struct{ S int }\n\n// Area returns the area.\nfunc (s Square) Area() int { return " + area + " }\n",
			"emb/emb.go":     "package emb\n\nimport \"example.com/m/sq\"\n\n// Big gets Area from sq.Square.\ntype Big struct{ sq.Square }\n",
			"ptr/ptr.go":     "package ptr\n\n// Circle is a shape through its pointer.\ntype Circle struct{ R int }\n\n// Area returns the area.\nfunc (c *Circle) Area() int { return 3 * c.R * c.R }\n",
			"use/use.go":     "package use\n\nimport \"example.com/m/iface\"\n\n// Total calls Area through the interface.\nfunc Total(s iface.Shape) int { return s.Area() }\n",
		}
	}
	g := gittest.New(t)
	g.Write(shapes("s.S * s.S"))
	base := g.Commit("base")
	g.Write(shapes("s.S * s.S * 1"))
	r, err := repo.Open(g.Dir, base)
	if err != nil {
		t.Fatal(err)
	}
	res, err := code.Load(r)
	if err != nil {
		t.Fatal(err)
	}
	got := append([]string(nil), res.After.Calls["use.Total"]...)
	sort.Strings(got)
	if want := []string{"ptr.Circle.Area", "sq.Square.Area"}; !reflect.DeepEqual(got, want) {
		t.Errorf("use.Total calls %v, want %v", got, want)
	}
	c := index(t, r, g.Dir)
	if !c.Reached["use.Total"] || c.ChangedDecls["ptr.Circle.Area"] {
		t.Errorf("use.Total reached: %v, ptr.Circle.Area changed: %v", c.Reached["use.Total"], c.ChangedDecls["ptr.Circle.Area"])
	}
}

func TestReadListsUses(t *testing.T) {
	g := gittest.New(t)
	g.Write(map[string]string{
		"go.mod":        goMod,
		"docs/flow.dfd": "[1. Run\n (p.F)]\n",
		"p/p.go":        pGo(1),
		"p/box.go":      "package p\n\n// Box holds a T.\ntype Box struct{ T T }\n",
		"p/p_test.go":   "package p\n\nimport \"testing\"\n\nfunc TestF(t *testing.T) {\n\tif F() != F() {\n\t\tt.Error(\"differs\")\n\t}\n}\n",
		"q/q.go":        qGo,
	})
	r, err := repo.Open(g.Dir, "")
	if err != nil {
		t.Fatal(err)
	}
	c := index(t, r, g.Dir)
	for key, want := range map[string][]code.Use{
		// A method's receiver and a field's type use a type too.
		"p.T": {{In: "p.Box", File: "p/box.go", Line: 4}, {In: "p.T.M", File: "p/p.go", Line: 10}, {In: "p.F", File: "p/p.go", Line: 19}},
		// H calls T.M through I, and so uses I.M, T.M and U.M.
		"p.T.M": {{In: "p.F", File: "p/p.go", Line: 19}, {In: "p.H", File: "p/p.go", Line: 25}},
		"p.I.M": {{In: "p.H", File: "p/p.go", Line: 25}},
		// Two calls on one line of a test make one use.
		"p.F": {{In: "p.TestF", File: "p/p_test.go", Line: 6}, {In: "q.K", File: "q/q.go", Line: 6}},
	} {
		if got := c.AfterUses[key]; !reflect.DeepEqual(got, want) {
			t.Errorf("uses of %s:\n got %+v\nwant %+v", key, got, want)
		}
	}
	if got, want := c.AfterMethods["p.T.M"], (code.Place{File: "p/p.go", Start: 9, End: 10, Kind: "method"}); got != want {
		t.Errorf("the place of p.T.M: %+v, want %+v", got, want)
	}
	if c.BeforeUses != nil {
		t.Errorf("uses without a base: %v", c.BeforeUses)
	}
}

func TestReadListsUncoveredChanges(t *testing.T) {
	step := func(n int) map[string]string {
		return map[string]string{
			"go.mod":        goMod,
			"docs/flow.dfd": "[1. Run\n (p.F)]\n> t\n[2. Next]\n# type: t = p.T\n",
			"p/p.go": fmt.Sprintf("package p\n\nimport \"example.com/m/s\"\n\n// T is a thing.\ntype T struct{}\n\n"+
				"// M returns the step.\nfunc (T) M() int { return %d }\n\n// F calls g.\nfunc F() int { return g() }\n\n"+
				"func g() int { return %d }\n\n// H calls s.S.\nfunc H() int { return s.S() }\n", n, n),
			"s/s.go": fmt.Sprintf("package s\n\n// S returns the step.\nfunc S() int { return %d }\n", n),
		}
	}
	g := gittest.New(t)
	g.Write(step(1))
	base := g.Commit("base")
	g.Write(step(2))
	r, err := repo.Open(g.Dir, base)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, u := range index(t, r, g.Dir).Uncovered {
		got = append(got, fmt.Sprintf("%s %s %s:%d %s", u.Key, u.Mark, u.Place.File, u.Place.Start, u.Version))
	}
	// p.F is linked, and p.g lies in its reach; T.M belongs to a linked type, which has no calls.
	want := []string{
		"p.H reached p/p.go:16 after",
		"p.T.M changed p/p.go:8 after",
		"s.S changed s/s.go:3 after",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("uncovered:\n got %q\nwant %q", got, want)
	}
}
