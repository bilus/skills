// Package dfdtext reads and links the text of dfd sources, line by line.
package dfdtext

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// Reference is a qualified name among a box's references, such as analyze.sumTypes.
type Reference struct {
	Qualifier string // a package name, or an import path
	Name      string
}

// CodeFn returns the target of a link to a reference.
type CodeFn func(Reference) string

// Linker gives the targets of the links that Link adds.
type Linker struct {
	Code  CodeFn            // the target of each reference
	Types map[string]string // the type of each flow label item
	Terms map[string]string // the definition of each term
}

// Link wraps each reference, typed item and term of src in a footnote reference.
// It returns the linked source and the footnotes it uses, by id. A typed item
// links to "#type/<item>" and a term to "#term/<term>", with the name path-escaped.
func Link(src string, l Linker) (string, map[string]string) {
	lines := strings.Split(src, "\n")
	w := wrapper{footnotes: map[string]string{}}
	terms := newTermMatcher(l.Terms)
	for _, e := range elements(lines) {
		switch e.kind {
		case process:
			group, inGroup := lastGroup(lines, e)
			for _, r := range references(lines, group) {
				if target := l.Code(r.ref); target != "" {
					w.add(r.at, "@c/"+r.ref.Qualifier+"."+r.ref.Name, target)
				}
			}
			for _, t := range terms.find(lines, e) {
				if !inGroup(t.at) {
					w.add(t.at, "@v/"+t.term, "#term/"+url.PathEscape(t.term))
				}
			}
		case entity, store:
			for _, t := range terms.find(lines, e) {
				w.add(t.at, "@v/"+t.term, "#term/"+url.PathEscape(t.term))
			}
		case flowLabel:
			for _, it := range items(lines, e) {
				if _, typed := l.Types[it.name]; typed {
					for _, at := range it.parts {
						w.add(at, "@t/"+it.name, "#type/"+url.PathEscape(it.name))
					}
				}
			}
		}
	}
	return w.apply(lines), w.footnotes
}

// References returns the references of the boxes in src, in order.
func References(src string) []Reference {
	lines := strings.Split(src, "\n")
	var out []Reference
	for _, e := range elements(lines) {
		if e.kind == process {
			group, _ := lastGroup(lines, e)
			for _, r := range references(lines, group) {
				out = append(out, r.ref)
			}
		}
	}
	return out
}

// Types returns the type of each item, from the "# type: item = type" comments of src.
// A later comment for an item replaces an earlier one.
func Types(src string) map[string]string {
	types := map[string]string{}
	for _, line := range strings.Split(src, "\n") {
		if m := typeComment.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
			types[m[1]] = m[2]
		}
	}
	return types
}

var typeComment = regexp.MustCompile(`^#\s*type:\s*(.+?)\s*=\s*(.+?)$`)

// Titles returns the action line of each process in src, by number. The action
// line is the first line of the box, without its number. Explicit numbers come
// from the labels; without them, the numbers follow dfd's own numbering, which
// numbers each distinct title in order after prefix, such as "2." in flow.2.dfd.
func Titles(src, prefix string) map[string]string {
	lines := strings.Split(src, "\n")
	explicit, automatic := map[string]string{}, map[string]string{}
	numbers := map[string]bool{}   // the titles dfd has numbered
	aliases := map[string]string{} // the label each alias declares
	for _, e := range elements(lines) {
		if e.kind != process {
			continue
		}
		title, _ := joined(lines, e)
		if m := numbered.FindStringSubmatch(firstLine(title)); m != nil {
			if _, seen := explicit[m[1]]; !seen {
				explicit[m[1]] = strings.TrimSpace(m[2])
			}
			continue
		}
		if alias := aliasOf(lines, e); alias != "" {
			aliases[alias] = title
		} else if label, ok := aliases[title]; ok {
			title = label
		}
		if !numbers[title] {
			numbers[title] = true
			automatic[prefix+strconv.Itoa(len(numbers))] = strings.TrimSpace(firstLine(title))
		}
	}
	if len(explicit) > 0 {
		return explicit
	}
	return automatic
}

// firstLine returns text up to its first line break.
func firstLine(text string) string {
	line, _, _ := strings.Cut(text, "\n")
	return line
}

// aliasOf returns the alias that a box's opening line declares, or "".
func aliasOf(lines []string, e element) string {
	first := e.spans[0]
	head := lines[first.line][:first.start]
	open := strings.IndexAny(head, "[{")
	marker := strings.LastIndex(head, ":=")
	if open < 0 || marker < open {
		return ""
	}
	return strings.TrimSpace(head[open+1 : marker])
}

// Op is one line of an alignment: ' ' for a line of both versions, '-' for a
// line of before alone, and '+' for a line of after alone.
type Op byte

// Align matches the lines of before and after along a longest common subsequence.
// Within a change, removed lines come before added ones, as in git's diffs.
func Align(before, after string) []Op {
	a, b := fileLines(before), fileLines(after)
	// common[i][j] is the length of the longest common subsequence of a[i:] and b[j:].
	common := make([][]int, len(a)+1)
	for i := range common {
		common[i] = make([]int, len(b)+1)
	}
	for i := len(a) - 1; i >= 0; i-- {
		for j := len(b) - 1; j >= 0; j-- {
			if a[i] == b[j] {
				common[i][j] = common[i+1][j+1] + 1
			} else {
				common[i][j] = max(common[i+1][j], common[i][j+1])
			}
		}
	}
	var ops []Op
	i, j := 0, 0
	for i < len(a) || j < len(b) {
		switch {
		case i < len(a) && j < len(b) && a[i] == b[j]:
			ops = append(ops, ' ')
			i++
			j++
		case j == len(b) || (i < len(a) && common[i+1][j] >= common[i][j+1]):
			ops = append(ops, '-')
			i++
		default:
			ops = append(ops, '+')
			j++
		}
	}
	return ops
}

// Patch writes the patch of an alignment, with the whole file as context. It takes
// removed lines from before and the other lines from after, so both texts may be
// linked versions of the aligned ones, with the same lines.
func Patch(path string, ops []Op, before, after string) string {
	a, b := fileLines(before), fileLines(after)
	var out strings.Builder
	fmt.Fprintf(&out, "--- a/%s\n+++ b/%s\n@@ -%s +%s @@\n", path, path, hunkRange(len(a)), hunkRange(len(b)))
	i, j := 0, 0
	for _, op := range ops {
		switch {
		case op == '-' && i < len(a):
			out.WriteString("-" + a[i] + "\n")
			i++
		case op == '+' && j < len(b):
			out.WriteString("+" + b[j] + "\n")
			j++
		case op == ' ' && i < len(a) && j < len(b):
			out.WriteString(" " + b[j] + "\n")
			i++
			j++
		}
	}
	return out.String()
}

// fileLines splits a file's text into lines, without the empty text after a final newline.
// A missing final newline makes no difference to an alignment.
func fileLines(text string) []string {
	if text == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(text, "\n"), "\n")
}

// hunkRange writes the range of a hunk header for a file of n lines.
func hunkRange(n int) string {
	if n == 0 {
		return "0,0"
	}
	return fmt.Sprintf("1,%d", n)
}
