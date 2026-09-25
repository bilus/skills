package dfdtext

import (
	"regexp"
	"sort"
	"strings"
)

// wrapper collects the spans to wrap in footnote references, and their footnotes.
type wrapper struct {
	spans     []wrap
	footnotes map[string]string
}

// wrap is one span to wrap, with the id of its footnote.
type wrap struct {
	at span
	id string
}

// add wraps the span at in a reference to id, unless its text would break the syntax.
func (w *wrapper) add(at span, id, target string) {
	w.spans = append(w.spans, wrap{at: at, id: id})
	w.footnotes[id] = target
}

// apply returns lines with every span wrapped, joined into a source.
func (w *wrapper) apply(lines []string) string {
	// Wrapping from the end of each line keeps the earlier offsets valid.
	sort.Slice(w.spans, func(i, j int) bool {
		a, b := w.spans[i].at, w.spans[j].at
		return a.line < b.line || (a.line == b.line && a.start > b.start)
	})
	out := append([]string(nil), lines...)
	last := span{line: -1}
	for _, s := range w.spans {
		ln := out[s.at.line]
		text := ln[s.at.start:s.at.end]
		overlaps := s.at.line == last.line && s.at.end > last.start
		if overlaps || strings.ContainsAny(text, `{}\`) || strings.Contains(s.id, ":") {
			continue
		}
		out[s.at.line] = ln[:s.at.start] + "{" + text + ":" + s.id + "}" + ln[s.at.end:]
		last = s.at
	}
	return strings.Join(out, "\n")
}

// joined is an element's text with its lines joined by newlines, and the offset of each span in it.
func joined(lines []string, e element) (string, []int) {
	var b strings.Builder
	offsets := make([]int, len(e.spans))
	for k, sp := range e.spans {
		if k > 0 {
			b.WriteByte('\n')
		}
		offsets[k] = b.Len()
		b.WriteString(lines[sp.line][sp.start:sp.end])
	}
	return b.String(), offsets
}

// at returns the span of the joined text's range [start, end), which lies on one line.
func at(e element, offsets []int, start, end int) span {
	k := sort.Search(len(offsets), func(k int) bool { return offsets[k] > start }) - 1
	sp := e.spans[k]
	return span{line: sp.line, start: sp.start + start - offsets[k], end: sp.start + end - offsets[k]}
}

// lastGroup returns the spans inside the last top-level pair of parentheses of a box,
// and a test of whether a span lies inside them.
func lastGroup(lines []string, e element) ([]span, func(span) bool) {
	text, offsets := joined(lines, e)
	depth, start, from, to := 0, 0, -1, -1
	for i, c := range text {
		switch c {
		case '(':
			if depth == 0 {
				start = i + 1
			}
			depth++
		case ')':
			if depth == 0 {
				continue
			}
			depth--
			if depth == 0 {
				from, to = start, i
			}
		}
	}
	if from < 0 {
		return nil, func(span) bool { return false }
	}
	var group []span
	for k, sp := range e.spans {
		lo, hi := max(from, offsets[k]), min(to, offsets[k]+sp.end-sp.start)
		if lo < hi {
			group = append(group, span{line: sp.line, start: sp.start + lo - offsets[k], end: sp.start + hi - offsets[k]})
		}
	}
	inside := func(s span) bool {
		for _, g := range group {
			if s.line == g.line && s.start < g.end && g.start < s.end {
				return true
			}
		}
		return false
	}
	return group, inside
}

// found is a reference and where it sits.
type found struct {
	ref Reference
	at  span
}

// qualified matches a qualified name, as dfdmetrics reads it.
var qualified = regexp.MustCompile(`((?:[A-Za-z0-9_.~-]+/)*)([a-z][a-z0-9_]*)\.([A-Za-z_][A-Za-z0-9_]*)`)

// references returns the qualified names in the spans of a box's last parentheses.
func references(lines []string, group []span) []found {
	var out []found
	for _, sp := range group {
		text := lines[sp.line][sp.start:sp.end]
		for _, m := range qualified.FindAllStringSubmatchIndex(text, -1) {
			// A match inside a longer name, such as b.C in a.b.C, is not a reference.
			if m[0] > 0 && strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_.", rune(text[m[0]-1])) {
				continue
			}
			out = append(out, found{
				ref: Reference{Qualifier: text[m[2]:m[5]], Name: text[m[6]:m[7]]},
				at:  span{line: sp.line, start: sp.start + m[0], end: sp.start + m[1]},
			})
		}
	}
	return out
}

// item is a flow label item, with its parts on each line.
type item struct {
	name  string // the item's words, joined by single spaces
	parts []span
}

// items splits a flow label at its commas.
func items(lines []string, e element) []item {
	text, offsets := joined(lines, e)
	var out []item
	start := 0
	for i := 0; i <= len(text); i++ {
		if i < len(text) && text[i] != ',' {
			continue
		}
		it := item{name: strings.Join(strings.Fields(text[start:i]), " ")}
		// Each line of the item is one part, without its surrounding whitespace.
		for lo := start; lo < i; {
			hi := strings.IndexByte(text[lo:i], '\n')
			if hi < 0 {
				hi = i
			} else {
				hi += lo
			}
			s, t := trim(text[lo:hi])
			if s < t {
				it.parts = append(it.parts, at(e, offsets, lo+s, lo+t))
			}
			lo = hi + 1
		}
		if it.name != "" {
			out = append(out, it)
		}
		start = i + 1
	}
	return out
}

// termMatcher finds the terms of a vocabulary, and their plurals, in text.
type termMatcher struct {
	re     *regexp.Regexp // nil for an empty vocabulary
	byForm map[string]string
}

// newTermMatcher builds a matcher for the terms, the longest first.
func newTermMatcher(terms map[string]string) termMatcher {
	m := termMatcher{byForm: map[string]string{}}
	var forms []string
	for term := range terms {
		for _, form := range []string{term, plural(term)} {
			m.byForm[strings.ToLower(form)] = term
			forms = append(forms, strings.ReplaceAll(regexp.QuoteMeta(form), " ", `[ \t]+`))
		}
	}
	if len(forms) == 0 {
		return m
	}
	sort.Slice(forms, func(i, j int) bool {
		return len(forms[i]) > len(forms[j]) || (len(forms[i]) == len(forms[j]) && forms[i] < forms[j])
	})
	m.re = regexp.MustCompile(`(?i)\b(?:` + strings.Join(forms, "|") + `)\b`)
	return m
}

// termAt is a term and where it sits.
type termAt struct {
	term string
	at   span
}

// find returns the terms on each line of an element.
func (m termMatcher) find(lines []string, e element) []termAt {
	if m.re == nil {
		return nil
	}
	var out []termAt
	for _, sp := range e.spans {
		text := lines[sp.line][sp.start:sp.end]
		for _, loc := range m.re.FindAllStringIndex(text, -1) {
			form := strings.ToLower(strings.Join(strings.Fields(text[loc[0]:loc[1]]), " "))
			out = append(out, termAt{term: m.byForm[form], at: span{line: sp.line, start: sp.start + loc[0], end: sp.start + loc[1]}})
		}
	}
	return out
}

// plural returns the regular plural of a term's last word.
func plural(term string) string {
	for _, end := range []string{"s", "x", "ch", "sh"} {
		if strings.HasSuffix(term, end) {
			return term + "es"
		}
	}
	return term + "s"
}
