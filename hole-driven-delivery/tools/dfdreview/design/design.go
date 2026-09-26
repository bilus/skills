// Package design reads the diagrams and the vocabulary of a leveled dfd design in both versions.
package design

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/dfdtext"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/repo"
)

// Design is the diagrams and the vocabulary of a change, in both versions.
type Design struct {
	Diagrams   []Diagram // the top diagram first, then the child diagrams by number
	Vocabulary repo.Pair
	Base       string // the base revision, "" when there is nothing to compare against
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
	src, err := r.Read(top)
	if err != nil {
		return nil, err
	}
	if !src.Before.Found && !src.After.Found {
		return nil, fmt.Errorf("%s: no such diagram at the base or in the working tree", top)
	}
	d := &Design{Diagrams: []Diagram{{Path: top, Title: "Overview", Source: src}}, Base: r.Base()}
	children, err := childFiles(r, top)
	if err != nil {
		return nil, err
	}
	byNumber := map[string]repo.Pair{"": src}
	for _, number := range sortedNumbers(children) {
		src, err := r.Read(children[number])
		if err != nil {
			return nil, err
		}
		byNumber[number] = src
		title := number
		if action, ok := dfdtext.Titles(latest(byNumber[parent(number)]), prefix(parent(number)))[number]; ok {
			title += " " + action
		}
		d.Diagrams = append(d.Diagrams, Diagram{Number: number, Path: children[number], Title: title, Source: src})
	}
	if vocabulary != "" {
		if d.Vocabulary, err = r.Read(vocabulary); err != nil {
			return nil, err
		}
	}
	return d, nil
}

// Terms reads a vocabulary's "- term: definition" lines, by term.
func Terms(vocabulary string) map[string]string {
	terms := map[string]string{}
	for _, line := range strings.Split(vocabulary, "\n") {
		if m := termLine.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
			terms[m[1]] = m[2]
		}
	}
	return terms
}

var termLine = regexp.MustCompile(`^- ([^:]+?): (.+)$`)

// childFiles maps the number of each child diagram of top to its path, from either version.
func childFiles(r *repo.Repo, top string) (map[string]string, error) {
	dir := filepath.Dir(top)
	stem := strings.TrimSuffix(filepath.Base(top), ".dfd")
	child := regexp.MustCompile(`^` + regexp.QuoteMeta(stem) + `\.([0-9]+(?:\.[0-9]+)*)\.dfd$`)
	names, err := r.Names(dir)
	if err != nil {
		return nil, err
	}
	files := map[string]string{}
	for _, name := range names {
		if m := child.FindStringSubmatch(name); m != nil {
			files[m[1]] = filepath.Join(dir, name)
		}
	}
	return files, nil
}

// sortedNumbers returns the keys of files in process order, so each parent comes first.
func sortedNumbers(files map[string]string) []string {
	numbers := make([]string, 0, len(files))
	for n := range files {
		numbers = append(numbers, n)
	}
	sort.Slice(numbers, func(i, j int) bool { return lessNumber(numbers[i], numbers[j]) })
	return numbers
}

// lessNumber orders process numbers part by part, as numbers.
func lessNumber(a, b string) bool {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(as) && i < len(bs); i++ {
		x, errX := strconv.Atoi(as[i])
		y, errY := strconv.Atoi(bs[i])
		if errX != nil || errY != nil || x == y {
			continue
		}
		return x < y
	}
	return len(as) < len(bs)
}

// parent returns the number of the process whose child diagram holds number.
func parent(number string) string {
	if i := strings.LastIndex(number, "."); i >= 0 {
		return number[:i]
	}
	return ""
}

// prefix returns what dfd's own numbering puts before the numbers inside a diagram.
func prefix(number string) string {
	if number == "" {
		return ""
	}
	return number + "."
}

// latest returns the working tree's version of a file, or the base's for a deleted one.
func latest(p repo.Pair) string {
	if p.After.Found {
		return p.After.Content
	}
	return p.Before.Content
}
