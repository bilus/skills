package draw_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/code"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/design"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/draw"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/internal/gittest"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/repo"
)

// echo is a stand-in for dfd: it writes its arguments, its footnotes and its input as SVG comments.
const echo = `#!/bin/sh
defs=""
prev=""
for a in "$@"; do
  if [ "$prev" = "--footnotes" ]; then defs="$a"; fi
  prev="$a"
done
printf '<svg><!--args: %s-->\n<!--footnotes:\n' "$*"
cat "$defs"
printf -- '-->\n<!--input:\n'
cat
printf -- '-->\n</svg>\n'
`

// failing is a stand-in for dfd that rejects every patch.
const failing = `#!/bin/sh
case "$*" in *--patch*) echo "<stdin>:3: bad patch line" >&2; exit 1;; esac
echo '<svg/>'
`

func script(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "dfd")
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

// change makes a design with a changed top diagram, an unchanged, a new and a deleted child.
func change(t *testing.T, withBase bool) (*design.Design, *code.Index) {
	t.Helper()
	g := gittest.New(t)
	g.Write(map[string]string{
		"docs/flow.dfd":      "[1. Read the sum types\n (lib.Read)]\n> items\n[2. Check]\n> x\n[3. Write]\n# type: items = []lib.Item\n",
		"docs/flow.1.dfd":    "[1.1. Parse]\n",
		"docs/flow.3.dfd":    "[3.1. Old]\n",
		"docs/vocabulary.md": "- sum type: a closed set of variants\n",
		"lib/lib.go":         "package lib\n\nfunc Read() {}\n\ntype Item int\n",
	})
	base := g.Commit("base")
	g.Write(map[string]string{
		"docs/flow.dfd":   "[1. Read the sum types\n (lib.Read)]\n> items\n[2. Check again]\n> x\n[3. Write]\n# type: items = []lib.Item\n",
		"docs/flow.2.dfd": "[2.1. New]\n",
	})
	g.Remove("docs/flow.3.dfd")
	if !withBase {
		base = ""
	}
	r, err := repo.Open(g.Dir, base)
	if err != nil {
		t.Fatal(err)
	}
	docs := filepath.Join(g.Dir, "docs")
	d, err := design.Read(r, filepath.Join(docs, "flow.dfd"), filepath.Join(docs, "vocabulary.md"))
	if err != nil {
		t.Fatal(err)
	}
	c, err := code.Read(r, d)
	if err != nil {
		t.Fatal(err)
	}
	return d, c
}

func TestViews(t *testing.T) {
	d, c := change(t, true)
	views, err := draw.Views(d, c, draw.Command{Name: script(t, echo)})
	if err != nil {
		t.Fatal(err)
	}
	drawn := map[string][3]bool{}
	for _, v := range views {
		drawn[v.Number] = [3]bool{v.Before != "", v.Diff != "", v.After != ""}
	}
	want := map[string][3]bool{
		"":  {true, true, true},   // changed
		"1": {false, false, true}, // unchanged
		"2": {false, true, true},  // new
		"3": {true, true, false},  // deleted
	}
	for number, views := range want {
		if drawn[number] != views {
			t.Errorf("diagram %q: before, diff, after drawn = %v, want %v", number, drawn[number], views)
		}
	}
	top := views[0]
	for _, want := range []string{"{lib.Read:@c/lib.Read}", "{sum types:@v/sum type}", "{items:@t/items}"} {
		if !strings.Contains(top.After, want) {
			t.Errorf("the after view's input lacks %q:\n%s", want, top.After)
		}
	}
	for _, want := range []string{"--patch", "-[2. Check]", "+[2. Check again]", " [1. Read the {sum types:@v/sum type}"} {
		if !strings.Contains(top.Diff, want) {
			t.Errorf("the diff view's input lacks %q:\n%s", want, top.Diff)
		}
	}
	for _, want := range []string{"{@c/lib.Read} #code/lib.Read", "{@v/sum type} #term/sum%20type", "{@t/items} #type/items"} {
		if !strings.Contains(top.After, want) {
			t.Errorf("the footnotes lack %q:\n%s", want, top.After)
		}
	}
}

// dfd numbers an unnumbered child diagram only with the prefix of its process.
func TestViewsPrefixChildNumbers(t *testing.T) {
	d, c := change(t, true)
	views, err := draw.Views(d, c, draw.Command{Name: script(t, echo)})
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range views {
		prefixed := strings.Contains(v.After+v.Diff, "--number-prefix "+v.Number+".")
		if v.Number != "" && !prefixed {
			t.Errorf("diagram %q: dfd ran without --number-prefix %s.", v.Number, v.Number)
		}
		if v.Number == "" && strings.Contains(v.After, "--number-prefix") {
			t.Errorf("the top diagram ran with a number prefix")
		}
	}
}

func TestViewsBoxesPerRow(t *testing.T) {
	d, c := change(t, true)
	for perRow, want := range map[int]string{0: "--per-row 5", 3: "--per-row 3"} {
		views, err := draw.Views(d, c, draw.Command{Name: script(t, echo), PerRow: perRow})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(views[0].After, want) {
			t.Errorf("PerRow %d: dfd ran without %q:\n%s", perRow, want, views[0].After)
		}
	}
}

func TestViewsWithoutBase(t *testing.T) {
	d, c := change(t, false)
	views, err := draw.Views(d, c, draw.Command{Name: script(t, echo)})
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range views {
		if v.Before != "" || v.Diff != "" || v.After == "" {
			t.Errorf("diagram %q: without a base, want the after view alone", v.Number)
		}
	}
}

func TestViewsNamesTheFailingView(t *testing.T) {
	d, c := change(t, true)
	_, err := draw.Views(d, c, draw.Command{Name: script(t, failing)})
	if err == nil || !strings.Contains(err.Error(), "flow.dfd (diff):3: bad patch line") {
		t.Errorf("error = %v, want one naming the diagram and its view", err)
	}
}

// TestViewsWithDfd draws with the real dfd when one with --patch is on the PATH or in $DFD.
func TestViewsWithDfd(t *testing.T) {
	dfd := os.Getenv("DFD")
	if dfd == "" {
		dfd = "dfd"
	}
	help, _ := exec.Command(dfd, "--help").CombinedOutput() // dfd --help exits with status 0 or 2
	if !strings.Contains(string(help), "-patch") {
		t.Skip("no dfd with --patch; set DFD to one")
	}
	d, c := change(t, true)
	views, err := draw.Views(d, c, draw.Command{Name: dfd})
	if err != nil {
		t.Fatal(err)
	}
	top := views[0]
	for _, want := range []string{`<a href="#code/lib.Read">`, `text-decoration="line-through"`, `>Check again<`} {
		if !strings.Contains(top.Diff, want) {
			t.Errorf("the diff view lacks %q", want)
		}
	}
}
