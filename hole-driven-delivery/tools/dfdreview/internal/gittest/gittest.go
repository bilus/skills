// Package gittest makes temporary git repositories for tests.
package gittest

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Repo is a git repository in a temporary directory.
type Repo struct {
	t   testing.TB
	Dir string
}

// New makes an empty repository with the user's git configuration left out.
func New(t testing.TB) *Repo {
	t.Helper()
	r := &Repo{t: t, Dir: t.TempDir()}
	r.git("init", "-q", "-b", "main")
	return r
}

// Write writes files, by path relative to the repository.
func (r *Repo) Write(files map[string]string) {
	r.t.Helper()
	for name, content := range files {
		path := filepath.Join(r.Dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			r.t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			r.t.Fatal(err)
		}
	}
}

// Remove removes files, by path relative to the repository.
func (r *Repo) Remove(names ...string) {
	r.t.Helper()
	for _, name := range names {
		if err := os.Remove(filepath.Join(r.Dir, name)); err != nil {
			r.t.Fatal(err)
		}
	}
}

// Commit commits every change and returns the commit's hash.
func (r *Repo) Commit(msg string) string {
	r.t.Helper()
	r.git("add", "-A")
	r.git("commit", "-q", "-m", msg)
	return strings.TrimSpace(r.git("rev-parse", "HEAD"))
}

// git runs git in the repository and returns its output.
func (r *Repo) git(args ...string) string {
	r.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = r.Dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null",
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com")
	out, err := cmd.CombinedOutput()
	if err != nil {
		r.t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}
