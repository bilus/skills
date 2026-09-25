// Package page writes the review page from its data.
package page

import "io"

// Data is everything the review page shows.
type Data struct {
	Title    string          `json:"title"`
	Diagrams []Diagram       `json:"diagrams"`
	Files    map[string]File `json:"files"`
	Decls    map[string]Decl `json:"decls"`
	Types    Versions        `json:"types"`
	Terms    Versions        `json:"terms"`
	Changed  []Changed       `json:"changed"`
	Brief    string          `json:"brief"`
	Metaphor string          `json:"metaphor"`
	OldScore *Card           `json:"oldScore"`
	NewScore *Card           `json:"newScore"`
	Report   string          `json:"report"`
}

// Diagram is one tab: a diagram's views, each an SVG document or "".
type Diagram struct {
	Number string `json:"number"`
	Title  string `json:"title"`
	Before string `json:"before"`
	Diff   string `json:"diff"`
	After  string `json:"after"`
}

// File is a source file in both versions, with its diff; nil where it does not exist.
type File struct {
	Before *string `json:"before"`
	After  *string `json:"after"`
	Diff   string  `json:"diff"`
}

// Decl is where a declaration sits in each version; nil where it does not exist.
type Decl struct {
	Before *Place `json:"before"`
	After  *Place `json:"after"`
}

// Place is a declaration's file and lines.
type Place struct {
	File  string `json:"file"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

// Versions holds a map from each version, such as the types of the flow label items.
type Versions struct {
	Before map[string]string `json:"before"`
	After  map[string]string `json:"after"`
}

// Changed is a file changed since the base.
type Changed struct {
	Path   string `json:"path"`
	Status string `json:"status"`
}

// Card is a score card: the column names and their values.
type Card struct {
	Columns []string `json:"columns"`
	Values  []string `json:"values"`
}

// Write writes the review page for d to w.
func Write(w io.Writer, d Data) error {
	panic("HOLE(5): the embedded page with d as JSON, safe inside a script element")
}
