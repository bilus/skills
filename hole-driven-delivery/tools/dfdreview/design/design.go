// Package design reads the diagrams and the vocabulary of a leveled dfd design in both versions.
package design

import "github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/repo"

// Design is the diagrams and the vocabulary of a change, in both versions.
type Design struct {
	Diagrams   []Diagram // the top diagram first, then the child diagrams by number
	Vocabulary repo.Pair
}

// Diagram is one dfd file of the design, in both versions.
type Diagram struct {
	Number string // "" for the top diagram
	Path   string // as the design names it
	Title  string // the tab's name: "Overview", or the number and the action line
	Source repo.Pair
}

// Read reads each diagram of the design rooted at top, and the vocabulary, in both versions.
// An empty vocabulary path means none.
func Read(r *repo.Repo, top, vocabulary string) (*Design, error) {
	panic("HOLE(2): child diagrams by file name from either version, titled from their parents' boxes")
}
