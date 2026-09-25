// Package notes reads the texts of a review from outside the design.
package notes

// Notes are the plan's brief and metaphor, the score cards and the report.
type Notes struct {
	Brief, Metaphor    string // sections of the plan
	OldScore, NewScore *Card  // nil when not given
	Report             string
}

// Card is a score card: the column names and their values.
type Card struct {
	Columns, Values []string
}

// Read reads the plan's brief and metaphor, the two score cards and the report.
// An empty path means none, and so does a cached score card that does not exist yet.
func Read(plan, oldScore, newScore, report string) (*Notes, error) {
	panic("HOLE(2): sections by heading, score cards as two tab-separated lines")
}
