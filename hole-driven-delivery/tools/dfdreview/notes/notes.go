// Package notes reads the texts of a review from outside the design.
package notes

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

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
	n := &Notes{}
	if plan != "" {
		src, err := os.ReadFile(plan)
		if err != nil {
			return nil, err
		}
		n.Brief = section(string(src), "The change in brief")
		n.Metaphor = section(string(src), "Metaphor")
	}
	if oldScore != "" {
		c, err := readCard(oldScore)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		}
		n.OldScore = c
	}
	if newScore != "" {
		c, err := readCard(newScore)
		if err != nil {
			return nil, err
		}
		n.NewScore = c
	}
	if report != "" {
		src, err := os.ReadFile(report)
		if err != nil {
			return nil, err
		}
		n.Report = string(src)
	}
	return n, nil
}

// section returns the text under the Markdown heading named title,
// up to the next heading of its level or a higher one.
func section(src, title string) string {
	var body []string
	level, fenced := 0, false
	for _, line := range strings.Split(src, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			fenced = !fenced
		}
		if l, name, ok := heading(line); ok && !fenced {
			if level > 0 && l <= level {
				break
			}
			if level == 0 && strings.EqualFold(name, title) {
				level = l
				continue
			}
		}
		if level > 0 {
			body = append(body, line)
		}
	}
	return strings.TrimSpace(strings.Join(body, "\n"))
}

// heading reads a Markdown heading line such as "## Metaphor".
func heading(line string) (level int, name string, ok bool) {
	rest := strings.TrimLeft(line, "#")
	level = len(line) - len(rest)
	if level == 0 || level > 6 || !strings.HasPrefix(rest, " ") {
		return 0, "", false
	}
	return level, strings.TrimSpace(rest), true
}

// readCard reads a score card: a tab-separated header line and a line of values.
func readCard(path string) (*Card, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimRight(string(src), "\n"), "\n")
	if len(lines) != 2 {
		return nil, fmt.Errorf("%s: want a header line and a line of values", path)
	}
	c := &Card{Columns: strings.Split(lines[0], "\t"), Values: strings.Split(lines[1], "\t")}
	if len(c.Columns) != len(c.Values) {
		return nil, fmt.Errorf("%s: %d columns in the header, %d in the values", path, len(c.Columns), len(c.Values))
	}
	return c, nil
}
