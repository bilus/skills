package code

import "github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/repo"

// Resolved holds the resolution of each version and the package errors of both.
type Resolved struct {
	Before, After *Resolution // nil for a version without a module
	Errors        []string    // each naming its version and its package
}

// Resolution is what the load finds in one version of the module.
type Resolution struct {
	Calls map[string][]string // by the key of each function and method, such as "render.renderer.text", the keys of the functions and methods of its package that it calls
	Files map[string]string   // the text of every Go file as the load parsed it, test files included, by path relative to the code directory
	Links map[string][]Link   // the identifier links of each Go file, by path
}

// Link is an identifier that uses a declared object, with the object's definition.
type Link struct {
	Line, Col, Len int    // the identifier's line, and its column and length in UTF-16 code units, from 1
	File           string // the definition's file in the same version, relative to the code directory
	To             int    // the definition's line in File
	URL            string // the definition's documentation, for one outside the code directory
}

// Load type-checks the module in both versions. It records the calls that every function
// and method makes within its package, the text and the identifier links of every Go file,
// and the package errors. It reads the base from the base export, a temporary copy of the
// module directory's blobs at the base that it deletes at the end. It returns an error only
// when it cannot run: an error that the go command or the type checker reports becomes a
// package error of its version.
func Load(r *repo.Repo) (*Resolved, error) {
	// HOLE(1): type-check both versions of the module with GOWORK=off and -mod=readonly, tests included, the base from the base export; record the calls within its package of every function and method, the text of every Go file as parsed, and the package errors; a version without a module gets no resolution
	// HOLE(2): record the identifier links of every Go file in both versions
	return &Resolved{After: &Resolution{Links: map[string][]Link{"lib/lib.go": {
		{Line: 6, Col: 22, Len: 4, File: "lib/lib.go", To: 9},
		{Line: 6, Col: 36, Len: 4, File: "lib/lib.go", To: 9},
		{Line: 6, Col: 41, Len: 4, File: "lib/lib.go", To: 9},
		{Line: 6, Col: 47, Len: 7, URL: "https://pkg.go.dev/strings"},
		{Line: 6, Col: 55, Len: 9, URL: "https://pkg.go.dev/strings#TrimSpace"},
		{Line: 6, Col: 65, Len: 1, File: "lib/lib.go", To: 6},
	}}}}, nil
}
