// Package page writes the review page from its data.
package page

import (
	_ "embed"
	"encoding/json"
	"html"
	"io"
	"strings"
)

// Data is everything the review page shows.
type Data struct {
	Title    string          `json:"title"`
	Base     string          `json:"base"` // the revision the page compares against, "" for none
	Diagrams []Diagram       `json:"diagrams"`
	Files    map[string]File `json:"files"`
	Decls    map[string]Decl `json:"decls"`
	Terms    Versions        `json:"terms"`
	Changed  []Changed       `json:"changed"`
	Brief    string          `json:"brief"`
	Metaphor string          `json:"metaphor"`
	OldScore *Card           `json:"oldScore"`
	NewScore *Card           `json:"newScore"`
	Report   string          `json:"report"`
	Errors   []string        `json:"errors"` // the package errors of both versions
}

// Diagram is one tab: a diagram's views, each an SVG document or "", and its items' types.
type Diagram struct {
	Number string   `json:"number"`
	Title  string   `json:"title"`
	Before string   `json:"before"`
	Diff   string   `json:"diff"`
	After  string   `json:"after"`
	Types  Versions `json:"types"`
}

// File is a source file in both versions, nil where it does not exist, with its diff
// and the identifier links of each version.
type File struct {
	Before      *string `json:"before"`
	After       *string `json:"after"`
	Diff        string  `json:"diff"`
	BeforeLinks []Link  `json:"beforeLinks"`
	AfterLinks  []Link  `json:"afterLinks"`
}

// Link is an identifier and its definition: a line of a Go file of the code directory in
// the same version, or a URL. The short keys keep the page small.
type Link struct {
	Line int    `json:"l"`
	Col  int    `json:"c"` // in UTF-16 code units, from 1
	Len  int    `json:"n"` // in UTF-16 code units
	File string `json:"f,omitempty"`
	To   int    `json:"t,omitempty"`
	URL  string `json:"u,omitempty"`
	Mark string `json:"m,omitempty"` // "changed" or "reached" inside a changed declaration
}

// Decl is where a declaration sits in each version, nil for a version without it,
// and whether the declaration differs between the versions.
type Decl struct {
	Before  *Place `json:"before"`
	After   *Place `json:"after"`
	Changed bool   `json:"changed"`
	Reached bool   `json:"reached,omitempty"` // changed through its reach only
}

// Place is a declaration's file and lines.
type Place struct {
	File  string `json:"file"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

// Versions holds a map from each version, such as the vocabulary's definitions.
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

//go:embed page.html
var template string

// Write writes the review page for d to w.
// The data sits in a script element as JSON, which escapes every "<".
func Write(w io.Writer, d Data) error {
	blob, err := json.Marshal(d)
	if err != nil {
		return err
	}
	r := strings.NewReplacer("{{TITLE}}", html.EscapeString(d.Title), "{{DATA}}", string(blob))
	_, err = io.WriteString(w, r.Replace(template))
	return err
}
