// Package repo reads the files of a git repository at a base revision and in the working tree.
package repo

// Text is a file's content in one version.
type Text struct {
	Content string
	Found   bool // the file exists in this version
}

// Pair is a file in both versions: at the base and in the working tree.
type Pair struct {
	Before, After Text
}

// Change is a file changed since the base, with its unified diff.
type Change struct {
	Path   string // relative to the code directory
	Status string // "added", "modified" or "deleted"
	Diff   string
}

// Repo reads the files of a code directory in both versions.
type Repo struct {
	top  string // the repository's top level
	dir  string // the code directory, relative to top
	base string // the base revision, "" for none
}

// Open opens the repository holding dir, to compare against base.
// An empty base means the page shows the working tree alone.
func Open(dir, base string) (*Repo, error) {
	panic("HOLE(2): the top level from git rev-parse; a base that git cannot resolve is an error")
}

// Read returns the file at path in both versions.
// Without a base, Before is not found.
func (r *Repo) Read(path string) (Pair, error) {
	panic("HOLE(2): git show for the base, the file system for the working tree; a missing file is not an error")
}

// Names returns the names of the files in dir, in either version.
func (r *Repo) Names(dir string) ([]string, error) {
	panic("HOLE(2): git ls-tree for the base, the directory for the working tree, sorted and without repeats")
}

// GoFiles returns the Go sources of the code directory in both versions, without tests.
func (r *Repo) GoFiles() (before, after map[string]string, err error) {
	panic("HOLE(2): paths relative to the code directory; testdata, vendor and hidden directories left out")
}

// Changed returns the files changed under the code directory since the base.
func (r *Repo) Changed() ([]Change, error) {
	panic("HOLE(2): git diff against the base plus untracked files, each with its diff; none without a base")
}
