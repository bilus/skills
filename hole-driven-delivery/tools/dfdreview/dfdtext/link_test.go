package dfdtext_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/dfdtext"
)

// code links the references of package p and of import paths, and no others.
func code(r dfdtext.Reference) string {
	switch {
	case r.Qualifier == "p" || r.Qualifier == "analyze":
		return "#code/" + r.Qualifier + "." + r.Name
	case strings.Contains(r.Qualifier, "/"):
		return "https://pkg.go.dev/" + r.Qualifier + "#" + r.Name
	}
	return ""
}

func TestLinkReferences(t *testing.T) {
	src := "[3. Validate the sum types\n to give patterns their variants\n (analyze.sumTypes,\n  analyze.checkCollisions, go/format.Source, q.Unlinked)]\n"
	got, footnotes := dfdtext.Link(src, dfdtext.Linker{Code: code})
	want := "[3. Validate the sum types\n to give patterns their variants\n" +
		" ({analyze.sumTypes:@c/analyze.sumTypes},\n" +
		"  {analyze.checkCollisions:@c/analyze.checkCollisions}, {go/format.Source:@c/go/format.Source}, q.Unlinked)]\n"
	if got != want {
		t.Errorf("linked:\n%s\nwant:\n%s", got, want)
	}
	wantNotes := map[string]string{
		"@c/analyze.sumTypes":        "#code/analyze.sumTypes",
		"@c/analyze.checkCollisions": "#code/analyze.checkCollisions",
		"@c/go/format.Source":        "https://pkg.go.dev/go/format#Source",
	}
	if !reflect.DeepEqual(footnotes, wantNotes) {
		t.Errorf("footnotes = %v, want %v", footnotes, wantNotes)
	}
}

func TestLinkTerms(t *testing.T) {
	terms := map[string]string{"sum type": "a closed set of variants", "client": "the caller", "registration": "a row", "row": "a record"}
	src := "{Client app}\n> sum types\n[1. Check the sum types (p.Check)]\n    > row\n    |R := Registration|\n> x\n[R]\n# type: sum types = []X\n"
	got, footnotes := dfdtext.Link(src, dfdtext.Linker{Code: code, Terms: terms})
	want := "{{Client:@v/client} app}\n> sum types\n[1. Check the {sum types:@v/sum type} ({p.Check:@c/p.Check})]\n" +
		"    > row\n    |R := {Registration:@v/registration}|\n> x\n[R]\n# type: sum types = []X\n"
	if got != want {
		t.Errorf("linked:\n%s\nwant:\n%s", got, want)
	}
	for id, target := range map[string]string{"@v/client": "#term/client", "@v/sum type": "#term/sum%20type", "@v/registration": "#term/registration"} {
		if footnotes[id] != target {
			t.Errorf("footnote %q = %q, want %q", id, footnotes[id], target)
		}
	}
	if _, ok := footnotes["@v/row"]; ok {
		t.Errorf("a store arrow's label got a term link")
	}
}

// dfd reads a line that opens with a footnote reference as an entity or a footnote
// definition, so an item that opens a continuation line stays unlinked, all its lines
// included. The other items keep their links.
func TestLinkTypedItems(t *testing.T) {
	types := map[string]string{"sum types": "[]X", "test files": "[]string", "sizes": "[]int"}
	src := "[A]\n> sum types, diagnostics,\n  test\n  files, sizes\n[B]\n"
	got, footnotes := dfdtext.Link(src, dfdtext.Linker{Code: code, Types: types})
	want := "[A]\n> {sum types:@t/sum types}, diagnostics,\n  test\n  files, {sizes:@t/sizes}\n[B]\n"
	if got != want {
		t.Errorf("linked:\n%s\nwant:\n%s", got, want)
	}
	if footnotes["@t/sum types"] != "#type/sum%20types" {
		t.Errorf("footnotes = %v", footnotes)
	}
}

// As in a flow label, an item that opens a continuation line stays unlinked.
func TestLinkTypedStoreItems(t *testing.T) {
	types := map[string]string{"walked template": "*T", "package": "P", "form structs": "S", "recv": "*R"}
	src := "[A]\n    > walked template,\n      package,\n      form structs, recv\n    < walked template\n    |Template walk|\n> x\n[B]\n"
	got, footnotes := dfdtext.Link(src, dfdtext.Linker{Code: code, Types: types})
	want := "[A]\n    > {walked template:@t/walked template},\n      package,\n      form structs, {recv:@t/recv}\n" +
		"    < {walked template:@t/walked template}\n    |Template walk|\n> x\n[B]\n"
	if got != want {
		t.Errorf("linked:\n%s\nwant:\n%s", got, want)
	}
	if footnotes["@t/walked template"] != "#type/walked%20template" {
		t.Errorf("footnotes = %v", footnotes)
	}
}

// A store comment links the store's name. dfd draws a bare alias with its label and the
// label's links, and reads a reference around the alias as another store, so the alias keeps
// its plain text, even under a comment of its own.
func TestLinkStores(t *testing.T) {
	stores := map[string]dfdtext.Reference{
		"Live template": {"p", "liveTemplate"},
		"Registration":  {"p", "registrations"},
		"R":             {"p", "R"},
		"Unlinked":      {"q", "Unlinked"},
	}
	src := "[A]\n    > tree\n    | Live template |\n> x\n[B]\n    > row\n    |R := Registration|\n> y\n" +
		"[C]\n    < row\n    |R|\n    < z\n    |Unlinked|\n"
	got, footnotes := dfdtext.Link(src, dfdtext.Linker{Code: code, Stores: stores})
	want := "[A]\n    > tree\n    | {Live template:@c/p.liveTemplate} |\n> x\n[B]\n    > row\n    |R := {Registration:@c/p.registrations}|\n> y\n" +
		"[C]\n    < row\n    |R|\n    < z\n    |Unlinked|\n"
	if got != want {
		t.Errorf("linked:\n%s\nwant:\n%s", got, want)
	}
	wantNotes := map[string]string{"@c/p.liveTemplate": "#code/p.liveTemplate", "@c/p.registrations": "#code/p.registrations"}
	if !reflect.DeepEqual(footnotes, wantNotes) {
		t.Errorf("footnotes = %v, want %v", footnotes, wantNotes)
	}
}

// The code link takes a store's whole name, including any term in the name. A store without a
// comment keeps its terms, and a bare alias, drawn by dfd with its label, takes none.
func TestLinkStoreNamesBeforeTerms(t *testing.T) {
	terms := map[string]string{"registration": "a row", "template": "a page with holes", "cache": "a copy"}
	stores := map[string]dfdtext.Reference{"Registration": {"p", "registrations"}, "Live template": {"p", "liveTemplate"}}
	src := "[A]\n    > row\n    |Registration|\n    > tree\n    |Live template|\n    > page\n    |Cache := Page cache|\n> x\n" +
		"[B]\n    < page\n    |Cache|\n"
	got, footnotes := dfdtext.Link(src, dfdtext.Linker{Code: code, Terms: terms, Stores: stores})
	want := "[A]\n    > row\n    |{Registration:@c/p.registrations}|\n    > tree\n    |{Live template:@c/p.liveTemplate}|\n" +
		"    > page\n    |Cache := Page {cache:@v/cache}|\n> x\n[B]\n    < page\n    |Cache|\n"
	if got != want {
		t.Errorf("linked:\n%s\nwant:\n%s", got, want)
	}
	for _, id := range []string{"@v/registration", "@v/template"} {
		if _, ok := footnotes[id]; ok {
			t.Errorf("a store's name with a code link kept the term link %s", id)
		}
	}
}

// Without a Code function, nothing links to code.
func TestLinkWithoutCode(t *testing.T) {
	src := "[1. A (p.A)]\n    > row\n    |Rows|\n> x\n[B]\n"
	got, footnotes := dfdtext.Link(src, dfdtext.Linker{Stores: map[string]dfdtext.Reference{"Rows": {"p", "rows"}}})
	if got != src || len(footnotes) != 0 {
		t.Errorf("linked:\n%s\nfootnotes %v, want the source unchanged and none", got, footnotes)
	}
}

func TestLinkLeavesOtherLinesAlone(t *testing.T) {
	terms := map[string]string{"map": "a table"}
	src := "# Map comment\n[Map\\{k\\} (p.F)]\n{Map} http://example.com\n> x\n[v := 2. Map it (p.G)]\n> y\n[v]\n"
	got, _ := dfdtext.Link(src, dfdtext.Linker{Code: code, Terms: terms})
	want := "# Map comment\n[{Map:@v/map}\\{k\\} ({p.F:@c/p.F})]\n{Map} http://example.com\n> x\n[v := 2. {Map:@v/map} it ({p.G:@c/p.G})]\n> y\n[v]\n"
	if got != want {
		t.Errorf("linked:\n%s\nwant:\n%s", got, want)
	}
}

// An external system beside a process is a name for terms, and its arrows are not flow labels.
func TestLinkExternalSystems(t *testing.T) {
	src := "[A]\n    > event\n    <Event bus>\n> x\n[B]\n"
	got, _ := dfdtext.Link(src, dfdtext.Linker{Code: code, Types: map[string]string{"event": "E"}, Terms: map[string]string{"bus": "a queue"}})
	want := "[A]\n    > event\n    <Event {bus:@v/bus}>\n> x\n[B]\n"
	if got != want {
		t.Errorf("linked:\n%s\nwant:\n%s", got, want)
	}
}

func TestReferences(t *testing.T) {
	src := "{E (e.NotAReference)}\n> x\n[1. A (p.A, q.B)]\n> y\n[2. B\n (see (p.Inner))\n (r.C,\n  a.b.C)]\n"
	want := []dfdtext.Reference{{"p", "A"}, {"q", "B"}, {"r", "C"}, {"a", "b"}}
	if got := dfdtext.References(src); !reflect.DeepEqual(got, want) {
		t.Errorf("references = %v, want %v", got, want)
	}
}

func TestAlignAndPatch(t *testing.T) {
	cases := []struct {
		name, before, after, want string
	}{
		{"changed line", "a\nb\nc\n", "a\nx\nc\n", "@@ -1,3 +1,3 @@\n a\n-b\n+x\n c\n"},
		{"removed before added", "a\nb\nc\n", "x\ny\n", "@@ -1,3 +1,2 @@\n-a\n-b\n-c\n+x\n+y\n"},
		{"new file", "", "a\nb\n", "@@ -0,0 +1,2 @@\n+a\n+b\n"},
		{"deleted file", "a\n", "", "@@ -1,1 +0,0 @@\n-a\n"},
		{"final newline ignored", "a\nb", "a\nb\n", "@@ -1,2 +1,2 @@\n a\n b\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			want := "--- a/docs/flow.dfd\n+++ b/docs/flow.dfd\n" + c.want
			ops := dfdtext.Align(c.before, c.after)
			if got := dfdtext.Patch("docs/flow.dfd", ops, c.before, c.after); got != want {
				t.Errorf("patch:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}

// The alignment of the raw texts decides what changed, so a link added to one
// version alone, as for a term new in its vocabulary, does not show as a change.
func TestPatchWritesLinkedLines(t *testing.T) {
	before, after := "[A]\n> x\n[B]\n", "[A]\n> x\n[C]\n"
	linkedBefore, linkedAfter := "[A]\n> x\n[B]\n", "[{A:@v/a}]\n> x\n[C]\n"
	got := dfdtext.Patch("f.dfd", dfdtext.Align(before, after), linkedBefore, linkedAfter)
	want := "--- a/f.dfd\n+++ b/f.dfd\n@@ -1,3 +1,3 @@\n [{A:@v/a}]\n > x\n-[B]\n+[C]\n"
	if got != want {
		t.Errorf("patch:\n%s\nwant:\n%s", got, want)
	}
}
