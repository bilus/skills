package code_test

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/code"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/design"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/internal/gittest"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/repo"
)

func TestDeclarations(t *testing.T) {
	files := map[string]string{
		"a/a.go": "package alpha\n\n// F does x.\nfunc F() {}\n\ntype (\n\t// T is t.\n\tT int\n\tU string\n)\n\nvar V = 1\n\nconst C = 2\n\nfunc (T) M() {}\n",
		"b/b.go": "package beta\n\nfunc F() {}\n",
		"c/c.go": "package alpha\n\nfunc F() {}\n",
	}
	got, err := code.Declarations(files)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]code.Place{
		"alpha.F": {File: "a/a.go", Start: 3, End: 4, Kind: "func"},
		"alpha.T": {File: "a/a.go", Start: 7, End: 8, Kind: "type"},
		"alpha.U": {File: "a/a.go", Start: 9, End: 9, Kind: "type"},
		"alpha.V": {File: "a/a.go", Start: 12, End: 12, Kind: "var"},
		"alpha.C": {File: "a/a.go", Start: 14, End: 14, Kind: "const"},
		"beta.F":  {File: "b/b.go", Start: 3, End: 3, Kind: "func"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("declarations = %v\nwant           %v", got, want)
	}
	// A line directive, which would put F at other.go:41, leaves the lines of the file's own text.
	directed, err := code.Declarations(map[string]string{"d.go": "package delta\n\n//line other.go:40\n\nfunc F() {}\n"})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := directed["delta.F"], (code.Place{File: "d.go", Start: 5, End: 5, Kind: "func"}); got != want {
		t.Errorf("under a line directive: %+v, want %+v", got, want)
	}
	if _, err := code.Declarations(map[string]string{"bad.go": "package"}); err == nil || !strings.Contains(err.Error(), "bad.go") {
		t.Errorf("parse error = %v, want one naming bad.go", err)
	}
}

func TestRead(t *testing.T) {
	g := gittest.New(t)
	g.Write(map[string]string{
		"docs/flow.dfd":  "[1. Read (lib.Read, fmt.Println)]\n> x\n[2. Gone (lib.Gone)]\n# type: x = []*types.Item\n",
		"types/types.go": "package types\n\ntype Item int\n",
		"lib/lib.go":     "package lib\n\nfunc Read() {}\n\nfunc Gone() {}\n",
		"lib/other.go":   "package lib\n\nfunc Other() {}\n",
		"README.md":      "old\n",
	})
	base := g.Commit("base")
	g.Write(map[string]string{
		"docs/flow.dfd": "[1. Read (lib.Read, fmt.Println)]\n> x\n[2. New (lib.New)]\n",
		"lib/lib.go":    "package lib\n\nfunc Read() {}\n\nfunc New() {}\n",
		"README.md":     "new\n",
	})
	r, err := repo.Open(g.Dir, base)
	if err != nil {
		t.Fatal(err)
	}
	d, err := design.Read(r, filepath.Join(g.Dir, "docs", "flow.dfd"), "")
	if err != nil {
		t.Fatal(err)
	}
	c, err := code.Read(r, d)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := c.Before["lib.Gone"]; !ok {
		t.Errorf("before: lib.Gone missing from %v", c.Before)
	}
	if _, ok := c.After["lib.New"]; !ok {
		t.Errorf("after: lib.New missing from %v", c.After)
	}
	if !c.Std["fmt"] || c.Std["lib"] {
		t.Errorf("std knows fmt: %v, lib: %v", c.Std["fmt"], c.Std["lib"])
	}
	var files []string
	for f := range c.Files {
		files = append(files, f)
	}
	for _, want := range []string{"lib/lib.go", "README.md", "docs/flow.dfd", "types/types.go"} {
		if _, ok := c.Files[want]; !ok {
			t.Errorf("files %v lack %s", files, want)
		}
	}
	if _, ok := c.Files["lib/other.go"]; ok {
		t.Errorf("files %v hold lib/other.go, which no box names and no change touched", files)
	}
	if lib := c.Files["lib/lib.go"]; !strings.Contains(lib.Before.Content, "Gone") || !strings.Contains(lib.After.Content, "New") {
		t.Errorf("lib/lib.go = %+v", lib)
	}
	if readme := c.Files["README.md"]; readme.Before.Content != "old\n" || readme.After.Content != "new\n" {
		t.Errorf("README.md = %+v", readme)
	}
}

func TestReadChangedDecls(t *testing.T) {
	g := gittest.New(t)
	g.Write(map[string]string{
		"docs/flow.dfd": "[1. Read (lib.Read)]\n",
		"lib/lib.go":    "package lib\n\n// Read reads.\nfunc Read() {}\n\nfunc Same() {}\n\nfunc Gone() {}\n\ntype Item int\n",
	})
	base := g.Commit("base")
	g.Write(map[string]string{
		// Same moves down a few lines and stays the same.
		"lib/lib.go": "package lib\n\n// Read reads the input.\nfunc Read() {}\n\nfunc New() {}\n\nfunc Same() {}\n\ntype Item string\n",
	})
	for _, tc := range []struct {
		base string
		want map[string]bool
	}{
		{base, map[string]bool{"lib.Read": true, "lib.Gone": true, "lib.New": true, "lib.Item": true}},
		{"", nil},
	} {
		r, err := repo.Open(g.Dir, tc.base)
		if err != nil {
			t.Fatal(err)
		}
		d, err := design.Read(r, filepath.Join(g.Dir, "docs", "flow.dfd"), "")
		if err != nil {
			t.Fatal(err)
		}
		c, err := code.Read(r, d)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(c.ChangedDecls, tc.want) {
			t.Errorf("base %q: changed declarations %v, want %v", tc.base, c.ChangedDecls, tc.want)
		}
	}
}
