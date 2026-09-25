// Package dfdmetrics measures shared state and package spread in a leveled dfd design.
package dfdmetrics

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/bilus/dfd/ast"
	"github.com/bilus/dfd/parse"
)

// Design is a top diagram with the child diagrams of its processes.
type Design struct {
	Diagrams []*Diagram // the top diagram first, then child diagrams by number
}

// Diagram is one dfd file of a design.
type Diagram struct {
	Path   string
	Number string // the expanded process's number, "" for the top diagram
	Depth  int
	Parent *Box   // the expanded box, nil for the top diagram
	Boxes  []*Box // process boxes in flow order
}

// Box is one drawing of a process.
type Box struct {
	Diagram     *Diagram
	Number      string // dfd's number, shared by boxes with one title
	Position    int    // index among the diagram's process boxes, from 1
	Title       string
	Refs        []Ref
	Accesses    []Access
	Child       *Diagram // nil for a box of a leaf process
	EntityAfter bool     // an entity sits between this box and the next process box
}

// Access is one read or write of a state item by a box.
type Access struct {
	Store string
	Item  string // "" for an arrow whose label gives no item
	Write bool
}

// Ref is a qualified name among a box's references.
type Ref struct {
	Qualifier string // a package name, or an import path
	Name      string
}

// Load reads the design rooted at the dfd file path.
func Load(path string) (*Design, error) {
	top := &Diagram{Path: path}
	if err := parseInto(top); err != nil {
		return nil, err
	}
	children, err := childFiles(path)
	if err != nil {
		return nil, err
	}
	numbers := make([]string, 0, len(children))
	for n := range children {
		numbers = append(numbers, n)
	}
	sort.Slice(numbers, func(i, j int) bool { return lessNumber(numbers[i], numbers[j]) })

	d := &Design{Diagrams: []*Diagram{top}}
	byNumber := map[string]*Diagram{"": top}
	for _, n := range numbers {
		parentNumber := ""
		if i := strings.LastIndex(n, "."); i >= 0 {
			parentNumber = n[:i]
		}
		parent, ok := byNumber[parentNumber]
		if !ok {
			return nil, fmt.Errorf("%s: missing parent diagram %s", children[n], siblingFile(path, parentNumber))
		}
		var drawn []*Box
		for _, b := range parent.Boxes {
			if b.Number == n {
				drawn = append(drawn, b)
			}
		}
		if len(drawn) == 0 {
			return nil, fmt.Errorf("%s: %s has no process %s", children[n], parent.Path, n)
		}
		if len(drawn) > 1 {
			return nil, fmt.Errorf("%s: process %s is drawn %d times in %s", children[n], n, len(drawn), parent.Path)
		}
		dg := &Diagram{Path: children[n], Number: n, Depth: parent.Depth + 1, Parent: drawn[0]}
		if err := parseInto(dg); err != nil {
			return nil, err
		}
		drawn[0].Child = dg
		byNumber[n] = dg
		d.Diagrams = append(d.Diagrams, dg)
	}
	return d, nil
}

// parseInto reads dg's file and sets its boxes.
func parseInto(dg *Diagram) error {
	src, err := os.ReadFile(dg.Path)
	if err != nil {
		return err
	}
	parsed, err := parse.Parse(bytes.NewReader(src), dg.Path)
	if err != nil {
		return err
	}
	dg.Boxes, err = boxes(dg, parsed.Steps)
	return err
}

// siblingFile returns the path of the diagram that expands process number.
func siblingFile(path, number string) string {
	dir, base := filepath.Split(path)
	return filepath.Join(dir, strings.TrimSuffix(base, ".dfd")+"."+number+".dfd")
}

// childFiles maps each child number to its file, for the design rooted at path.
func childFiles(path string) (map[string]string, error) {
	dir, base := filepath.Split(path)
	stem := strings.TrimSuffix(base, ".dfd")
	child := regexp.MustCompile(`^` + regexp.QuoteMeta(stem) + `\.([0-9]+(?:\.[0-9]+)*)\.dfd$`)
	listed := dir
	if listed == "" {
		listed = "."
	}
	entries, err := os.ReadDir(listed)
	if err != nil {
		return nil, err
	}
	files := map[string]string{}
	for _, e := range entries {
		if m := child.FindStringSubmatch(e.Name()); m != nil && !e.IsDir() {
			files[m[1]] = filepath.Join(dir, e.Name())
		}
	}
	return files, nil
}

// boxes turns parsed steps into the numbered boxes of dg.
// An explicit number, as in "3.1. Parse", replaces dfd's own numbering.
func boxes(dg *Diagram, steps []ast.Step) ([]*Box, error) {
	prefix := ""
	if dg.Number != "" {
		prefix = dg.Number + "."
	}
	// dfd identifies a process by its title, so a repeated title keeps its number.
	numbers := map[string]string{}
	labels := map[string]string{} // the label of each explicit number
	var unnumbered *Box           // the first box without an explicit number
	var out []*Box
	entity := false
	for _, st := range steps {
		if st.Kind == ast.Entity {
			entity = true
			continue
		}
		title := st.Title
		num, rest, explicit := explicitNumber(title)
		switch {
		case !explicit:
			var ok bool
			if num, ok = numbers[title]; !ok {
				num = fmt.Sprintf("%s%d", prefix, len(numbers)+1)
				numbers[title] = num
			}
		case prefix == "" && strings.Contains(num, "."):
			return nil, fmt.Errorf("%s: process %s does not belong in the top diagram", dg.Path, num)
		case !strings.HasPrefix(num, prefix) || strings.Contains(num[len(prefix):], "."):
			return nil, fmt.Errorf("%s: process %s does not belong to process %s", dg.Path, num, dg.Number)
		case labels[num] != "" && labels[num] != rest:
			return nil, fmt.Errorf("%s: number %s already belongs to %q", dg.Path, num, labels[num])
		default:
			labels[num], title = rest, rest
		}
		if n := len(out); n > 0 && entity {
			out[n-1].EntityAfter = true
		}
		entity = false
		b := &Box{
			Diagram:  dg,
			Number:   num,
			Position: len(out) + 1,
			Title:    title,
			Refs:     refs(title),
			Accesses: accesses(st.Stores),
		}
		if !explicit && unnumbered == nil {
			unnumbered = b
		}
		out = append(out, b)
	}
	if len(labels) > 0 && unnumbered != nil {
		return nil, fmt.Errorf("%s: process %q has no number; number every process or none", dg.Path, unnumbered.Title)
	}
	return out, nil
}

// explicitNumber splits a leading number such as "3.2. " off a process label.
// The number is groups of digits joined by dots, ended by a period and a space.
func explicitNumber(label string) (number, rest string, ok bool) {
	m := explicitPrefix.FindStringSubmatch(label)
	if m == nil {
		return "", "", false
	}
	return m[1], m[2], true
}

var explicitPrefix = regexp.MustCompile(`(?s)^([0-9]+(?:\.[0-9]+)*)\. +(\S.*)$`)

// accesses returns one access per state item of each store arrow, without repeats.
func accesses(links []ast.StoreLink) []Access {
	var out []Access
	seen := map[Access]bool{}
	add := func(a Access) {
		if !seen[a] {
			seen[a] = true
			out = append(out, a)
		}
	}
	for _, l := range links {
		for _, arrow := range []struct {
			a     *ast.Arrow
			write bool
		}{{l.Put, true}, {l.Get, false}} {
			if arrow.a == nil {
				continue
			}
			items := splitItems(arrow.a.Label)
			if len(items) == 0 {
				add(Access{Store: l.Name, Write: arrow.write})
			}
			for _, it := range items {
				add(Access{Store: l.Name, Item: it, Write: arrow.write})
			}
		}
	}
	return out
}

// splitItems returns the normalized state items of a store arrow's label.
func splitItems(label string) []string {
	var items []string
	for _, part := range strings.Split(label, ",") {
		if it := strings.Join(strings.Fields(part), " "); it != "" {
			items = append(items, it)
		}
	}
	return items
}

// qualified matches an optional import path, a package name, a dot and an identifier.
var qualified = regexp.MustCompile(`((?:[A-Za-z0-9_.~-]+/)*)([a-z][a-z0-9_]*)\.([A-Za-z_][A-Za-z0-9_]*)`)

// refs returns the qualified names in the last top-level parentheses of title.
func refs(title string) []Ref {
	group, ok := lastGroup(title)
	if !ok {
		return nil
	}
	var out []Ref
	for _, m := range qualified.FindAllStringSubmatchIndex(group, -1) {
		// Skip a match inside a longer name, such as b.C in a.b.C.
		if m[0] > 0 && strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_.", rune(group[m[0]-1])) {
			continue
		}
		out = append(out, Ref{Qualifier: group[m[2]:m[5]], Name: group[m[6]:m[7]]})
	}
	return out
}

// lastGroup returns the text inside the last top-level pair of parentheses in s.
func lastGroup(s string) (string, bool) {
	depth, start := 0, 0
	group, found := "", false
	for i, r := range s {
		switch r {
		case '(':
			if depth == 0 {
				start = i + 1
			}
			depth++
		case ')':
			if depth == 0 {
				continue
			}
			depth--
			if depth == 0 {
				group, found = s[start:i], true
			}
		}
	}
	return group, found
}

// lessNumber orders process numbers part by part, as numbers.
func lessNumber(a, b string) bool {
	if a == "" || b == "" {
		return a == "" && b != ""
	}
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(as) && i < len(bs); i++ {
		if as[i] == bs[i] {
			continue
		}
		x, errX := strconv.Atoi(as[i])
		y, errY := strconv.Atoi(bs[i])
		if errX != nil || errY != nil {
			return as[i] < bs[i]
		}
		return x < y
	}
	return len(as) < len(bs)
}
