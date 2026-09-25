// Package repo reads the files of a git repository at a base revision and in the working tree.
package repo

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

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
	Path    string // relative to the code directory
	Status  string // "added", "modified" or "deleted"
	Diff    string
	Content Pair // a binary file's versions hold no text
}

// Repo reads the files of a code directory in both versions.
type Repo struct {
	top  string // the repository's top level
	dir  string // the code directory, relative to top, with slashes
	base string // the base revision, "" for none
}

// Open opens the repository holding dir, to compare against base.
// An empty base means the page shows the working tree alone.
func Open(dir, base string) (*Repo, error) {
	abs, err := resolve(dir)
	if err != nil {
		return nil, err
	}
	out, err := git(abs, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, fmt.Errorf("%s: not in a git repository: %w", dir, err)
	}
	top := strings.TrimSpace(out)
	rel, err := filepath.Rel(top, abs)
	if err != nil {
		return nil, err
	}
	if base != "" {
		if _, err := git(top, "rev-parse", "--verify", "--quiet", base+"^{commit}"); err != nil {
			return nil, fmt.Errorf("base %q: not a commit in %s", base, top)
		}
	}
	return &Repo{top: top, dir: filepath.ToSlash(rel), base: base}, nil
}

// Base returns the base revision, "" when the page shows the working tree alone.
func (r *Repo) Base() string { return r.base }

// Read returns the file at path in both versions.
// Without a base, Before is not found.
func (r *Repo) Read(p string) (Pair, error) {
	var pair Pair
	after, err := os.ReadFile(p)
	switch {
	case err == nil:
		pair.After = Text{Content: string(after), Found: true}
	case !errors.Is(err, fs.ErrNotExist):
		return Pair{}, err
	}
	if r.base == "" {
		return pair, nil
	}
	rel, err := r.inRepo(p)
	if err != nil {
		return Pair{}, err
	}
	listed, err := r.list("--", rel)
	if err != nil || len(listed) == 0 {
		return pair, err
	}
	before, err := r.cat([]string{rel})
	if err != nil {
		return Pair{}, err
	}
	pair.Before = Text{Content: before[rel], Found: true}
	return pair, nil
}

// Names returns the names of the files in dir, in either version.
func (r *Repo) Names(dir string) ([]string, error) {
	seen := map[string]bool{}
	entries, err := os.ReadDir(dir)
	switch {
	case err == nil:
		for _, e := range entries {
			if !e.IsDir() {
				seen[e.Name()] = true
			}
		}
	case !errors.Is(err, fs.ErrNotExist):
		return nil, err
	}
	if r.base != "" {
		rel, err := r.inRepo(dir)
		if err != nil {
			return nil, err
		}
		listed, err := r.list("--", treeArg(rel))
		if err != nil {
			return nil, err
		}
		for _, p := range listed {
			seen[path.Base(p)] = true
		}
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

// GoFiles returns the Go sources of the code directory in both versions, without tests.
// The paths are relative to the code directory; testdata, vendor and hidden directories are left out.
func (r *Repo) GoFiles() (before, after map[string]string, err error) {
	root := filepath.Join(r.top, filepath.FromSlash(r.dir))
	after = map[string]string{}
	err = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if p != root && skipped(d.Name()) {
				return fs.SkipDir
			}
			return nil
		}
		if !source(rel) {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		after[rel] = string(b)
		return nil
	})
	if err != nil || r.base == "" {
		return nil, after, err
	}
	listed, err := r.list("-r", "--", treeArg(r.dir))
	if err != nil {
		return nil, nil, err
	}
	var paths []string
	for _, p := range listed {
		if source(r.inCode(p)) {
			paths = append(paths, p)
		}
	}
	contents, err := r.cat(paths)
	if err != nil {
		return nil, nil, err
	}
	before = map[string]string{}
	for p, content := range contents {
		before[r.inCode(p)] = content
	}
	return before, after, nil
}

// Changed returns the files changed under the code directory since the base, with their diffs.
// Untracked files count as added; without a base, nothing has changed.
func (r *Repo) Changed() ([]Change, error) {
	if r.base == "" {
		return nil, nil
	}
	out, err := git(r.top, "diff", "--name-status", "--no-renames", "-z", r.base, "--", treeArg(r.dir))
	if err != nil {
		return nil, err
	}
	var changes []Change
	fields := strings.Split(strings.TrimSuffix(out, "\x00"), "\x00")
	for i := 0; i+1 < len(fields); i += 2 {
		status := "modified"
		switch fields[i] {
		case "A":
			status = "added"
		case "D":
			status = "deleted"
		}
		p := fields[i+1]
		diff, err := git(r.top, "diff", "--no-color", "--no-ext-diff", "--src-prefix=a/", "--dst-prefix=b/", r.base, "--", p)
		if err != nil {
			return nil, err
		}
		var content Pair
		if status != "added" {
			before, err := r.cat([]string{p})
			if err != nil {
				return nil, err
			}
			content.Before = text(before[p])
		}
		if status != "deleted" {
			after, err := os.ReadFile(filepath.Join(r.top, filepath.FromSlash(p)))
			if err != nil {
				return nil, err
			}
			content.After = text(string(after))
		}
		changes = append(changes, Change{Path: r.inCode(p), Status: status, Diff: diff, Content: content})
	}
	out, err = git(r.top, "ls-files", "--others", "--exclude-standard", "-z", "--", treeArg(r.dir))
	if err != nil {
		return nil, err
	}
	for _, p := range strings.Split(strings.TrimSuffix(out, "\x00"), "\x00") {
		if p == "" {
			continue
		}
		content, err := os.ReadFile(filepath.Join(r.top, filepath.FromSlash(p)))
		if err != nil {
			return nil, err
		}
		changes = append(changes, Change{Path: r.inCode(p), Status: "added", Diff: added(p, string(content)), Content: Pair{After: text(string(content))}})
	}
	sort.Slice(changes, func(i, j int) bool { return changes[i].Path < changes[j].Path })
	return changes, nil
}

// text returns a found version of a file, without the content of a binary file.
func text(content string) Text {
	if strings.IndexByte(content, 0) >= 0 {
		return Text{Found: true}
	}
	return Text{Content: content, Found: true}
}

// added returns the diff that creates a file with content.
func added(p, content string) string {
	if strings.IndexByte(content, 0) >= 0 {
		return "Binary file b/" + p + " added\n"
	}
	lines := strings.SplitAfter(content, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	var b strings.Builder
	fmt.Fprintf(&b, "--- /dev/null\n+++ b/%s\n@@ -0,0 +1,%d @@\n", p, len(lines))
	for _, ln := range lines {
		b.WriteString("+" + ln)
		if !strings.HasSuffix(ln, "\n") {
			b.WriteString("\n\\ No newline at end of file\n")
		}
	}
	return b.String()
}

// list returns the paths of the files that git ls-tree lists at the base for args.
func (r *Repo) list(args ...string) ([]string, error) {
	out, err := git(r.top, append([]string{"ls-tree", "-z", r.base}, args...)...)
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, entry := range strings.Split(strings.TrimSuffix(out, "\x00"), "\x00") {
		// An entry reads "<mode> <type> <object>\t<path>".
		info, p, ok := strings.Cut(entry, "\t")
		if f := strings.Fields(info); ok && len(f) == 3 && f[1] == "blob" {
			paths = append(paths, p)
		}
	}
	return paths, nil
}

// cat returns the content of each path at the base, by path.
func (r *Repo) cat(paths []string) (map[string]string, error) {
	var in bytes.Buffer
	for _, p := range paths {
		fmt.Fprintf(&in, "%s:%s\n", r.base, p)
	}
	cmd := exec.Command("git", "cat-file", "--batch")
	cmd.Dir = r.top
	cmd.Stdin = &in
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git cat-file: %v: %s", err, strings.TrimSpace(stderr.String()))
	}
	contents := make(map[string]string, len(paths))
	rd := bufio.NewReader(bytes.NewReader(out))
	for _, p := range paths {
		header, err := rd.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("git cat-file %s: %w", p, err)
		}
		fields := strings.Fields(header)
		if len(fields) != 3 {
			return nil, fmt.Errorf("git cat-file %s: %s", p, strings.TrimSpace(header))
		}
		size, err := strconv.Atoi(fields[2])
		if err != nil {
			return nil, fmt.Errorf("git cat-file %s: %w", p, err)
		}
		content := make([]byte, size+1) // the object and its closing newline
		if _, err := io.ReadFull(rd, content); err != nil {
			return nil, fmt.Errorf("git cat-file %s: %w", p, err)
		}
		contents[p] = string(content[:size])
	}
	return contents, nil
}

// inRepo returns p relative to the repository's top level, with slashes.
func (r *Repo) inRepo(p string) (string, error) {
	abs, err := resolve(p)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(r.top, abs)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%s: outside the repository %s", p, r.top)
	}
	return filepath.ToSlash(rel), nil
}

// inCode returns a path relative to the top level as a path relative to the code directory.
func (r *Repo) inCode(p string) string {
	if r.dir == "." {
		return p
	}
	return strings.TrimPrefix(p, r.dir+"/")
}

// treeArg returns the pathspec of a directory relative to the top level.
func treeArg(dir string) string {
	if dir == "." {
		return "."
	}
	return dir + "/"
}

// skipped reports whether a directory holds no code of its own.
func skipped(name string) bool {
	return name == "testdata" || name == "vendor" || strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")
}

// source reports whether a path relative to the code directory is a Go source other than a test.
func source(rel string) bool {
	if !strings.HasSuffix(rel, ".go") || strings.HasSuffix(rel, "_test.go") {
		return false
	}
	for _, part := range strings.Split(path.Dir(rel), "/") {
		if part != "." && skipped(part) {
			return false
		}
	}
	return true
}

// resolve returns p as an absolute path with the symlinks of its existing part resolved.
func resolve(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	rest := ""
	for {
		real, err := filepath.EvalSymlinks(abs)
		if err == nil {
			return filepath.Join(real, rest), nil
		}
		parent := filepath.Dir(abs)
		if !errors.Is(err, fs.ErrNotExist) || parent == abs {
			return "", err
		}
		rest = filepath.Join(filepath.Base(abs), rest)
		abs = parent
	}
}

// git runs git in dir and returns its standard output.
func git(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return string(out), nil
}
