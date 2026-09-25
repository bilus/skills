// Package dfdtext reads and links the text of dfd sources, line by line.
package dfdtext

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
func Types(src string) map[string]string {
	panic("HOLE(2): one entry per comment; a later comment for an item replaces an earlier one")
}

// Titles returns the action line of each process with an explicit number in src, by number.
func Titles(src string) map[string]string {
	panic("HOLE(2): the first line of each numbered box, without its number and alias")
}

// Unified returns a patch from before to after that holds the whole file as context.
func Unified(path, before, after string) string {
	panic("HOLE(3): one hunk from a longest common subsequence of the lines; a missing side is an empty file")
}
