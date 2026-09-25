// Package code indexes the Go declarations and the changed files of a change in both versions.
package code

import (
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/design"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/repo"
)

// Index is the Go declarations and the changed files of a change, in both versions.
type Index struct {
	Before, After map[string]Place     // declarations by key, such as "analyze.sumTypes"
	Std           map[string]bool      // the standard library's import paths
	Files         map[string]repo.Pair // the files the page shows, by path
	Changed       []repo.Change
}

// Place is where a declaration sits, its doc comment included.
type Place struct {
	File       string
	Start, End int // lines, from 1
}

// Read indexes the declarations of both versions and the files changed since the base.
func Read(r *repo.Repo, d *design.Design) (*Index, error) {
	panic("HOLE(4): the files of the referenced declarations and every changed file, with go list std")
}

// Declarations returns the top-level declarations of files, keyed like "analyze.sumTypes".
// The key joins the package name and the identifier; the first file in path order wins.
func Declarations(files map[string]string) (map[string]Place, error) {
	panic("HOLE(4): functions, types, variables and constants, by package clause; a parse error names its file")
}
