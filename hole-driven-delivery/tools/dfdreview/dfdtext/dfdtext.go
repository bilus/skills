// Package dfdtext reads and links the text of dfd sources, line by line.
package dfdtext

import (
	"regexp"
	"strings"
)

// Reference is a qualified name among a box's references, such as analyze.sumTypes.
type Reference struct {
	Qualifier string // a package name, or an import path
	Name      string
}

// CodeFn returns the target of a link to a reference.
type CodeFn func(Reference) string

// Linker gives the targets of the links that Link adds.
type Linker struct {
	Code  CodeFn            // the target of each reference
	Types map[string]string // the type of each flow label item
	Terms map[string]string // the definition of each term
}

// Link wraps each reference, typed item and term of src in a footnote reference.
// It returns the linked source and the footnotes it uses, by id.
func Link(src string, l Linker) (string, map[string]string) {
	panic("HOLE(3): references in a box's last parentheses, typed items of flow labels, terms in titles and store names")
}

// References returns the references of the boxes in src, in order.
func References(src string) []Reference {
	panic("HOLE(3): the qualified names in the last top-level parentheses of each box title")
}

// Types returns the type of each item, from the "# type: item = type" comments of src.
// A later comment for an item replaces an earlier one.
func Types(src string) map[string]string {
	types := map[string]string{}
	for _, line := range strings.Split(src, "\n") {
		if m := typeComment.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
			types[m[1]] = m[2]
		}
	}
	return types
}

var typeComment = regexp.MustCompile(`^#\s*type:\s*(.+?)\s*=\s*(.+?)$`)

// Titles returns the action line of each process with an explicit number in src, by number.
// The action line is the first line of the box, without its number.
func Titles(src string) map[string]string {
	lines := strings.Split(src, "\n")
	titles := map[string]string{}
	for _, e := range elements(lines) {
		if e.kind != process {
			continue
		}
		first := e.spans[0]
		m := numbered.FindStringSubmatch(lines[first.line][first.start:first.end])
		if m == nil {
			continue
		}
		if _, seen := titles[m[1]]; !seen {
			titles[m[1]] = strings.TrimSpace(m[2])
		}
	}
	return titles
}

// Unified returns a patch from before to after that holds the whole file as context.
func Unified(path, before, after string) string {
	panic("HOLE(3): one hunk from a longest common subsequence of the lines; a missing side is an empty file")
}
