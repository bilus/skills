package dfdtext

import "regexp"

// kind is what an element of a dfd source names.
type kind int

const (
	process    kind = iota // a process box's title
	entity                 // an external entity's name
	store                  // a datastore's name
	flowLabel              // the label of a flow arrow
	storeLabel             // the label of a store arrow
)

// span is the text of an element on one source line: lines[line][start:end].
type span struct {
	line, start, end int
}

// element is a title, a name or a label, over one span per source line.
type element struct {
	kind  kind
	spans []span
}

// elements reads the titles, names and labels of a dfd source, in order.
// A title's or a name's first span starts after its alias marker, if any.
func elements(lines []string) []element {
	var out []element
	var pending []element // arrows waiting for the line that says what they point at
	open := -1            // the index in out of a box still open, or -1
	var closer byte       // the closing delimiter of the open box
	cont := false         // the last line was an arrow or its continuation
	for i, raw := range lines {
		start, end := trim(raw)
		text := raw[start:end]
		if open >= 0 {
			sp := span{line: i, start: start, end: end}
			box := &out[open]
			if closes(text, closer) {
				sp.end--
				open = -1
			}
			box.spans = append(box.spans, sp)
			continue
		}
		switch {
		case text == "" || text[0] == '#':
			cont = false
		case definition.MatchString(text):
			cont = false
		case text[0] == '[' || text[0] == '{':
			cont = false
			for _, a := range pending {
				a.kind = flowLabel
				out = append(out, a)
			}
			pending = nil
			k, c := process, byte(']')
			if text[0] == '{' {
				k, c = entity, '}'
			}
			sp := afterAlias(raw, span{line: i, start: start + 1, end: end})
			if closes(text[1:], c) {
				sp.end--
			} else {
				open, closer = len(out), c
			}
			out = append(out, element{kind: k, spans: []span{sp}})
		case text[0] == '|':
			cont = false
			for _, a := range pending {
				a.kind = storeLabel
				out = append(out, a)
			}
			pending = nil
			sp := span{line: i, start: start + 1, end: end}
			if closes(text[1:], '|') {
				sp.end--
			}
			out = append(out, element{kind: store, spans: []span{afterAlias(raw, sp)}})
		case text[0] == '>' || text[0] == '<':
			cont = true
			s, _ := trim(raw[start+1 : end])
			pending = append(pending, element{spans: []span{{line: i, start: start + 1 + s, end: end}}})
		default:
			if cont && len(pending) > 0 {
				last := &pending[len(pending)-1]
				last.spans = append(last.spans, span{line: i, start: start, end: end})
			}
		}
	}
	return out
}

// definition matches a footnote definition line, "{id} target".
var definition = regexp.MustCompile(`^\{[^{}]*\}[ \t]+\S`)

// trim returns the bounds of s without its surrounding whitespace.
func trim(s string) (start, end int) {
	start, end = 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}
	return start, end
}

// closes reports whether text ends with an unescaped closing delimiter c.
func closes(text string, c byte) bool {
	n := len(text)
	return n > 0 && text[n-1] == c && (n < 2 || text[n-2] != '\\')
}

// afterAlias moves the start of sp past an alias marker, as in "R := Registration".
func afterAlias(line string, sp span) span {
	text := line[sp.start:sp.end]
	for i := 0; i+1 < len(text); i++ {
		switch {
		case text[i] == '\\':
			i++
		case text[i] == ':' && text[i+1] == '=':
			s, _ := trim(text[i+2:])
			sp.start += i + 2 + s
			return sp
		}
	}
	return sp
}

// numbered matches a label with an explicit process number, as in "3.2. Validate".
var numbered = regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)*)\. +(\S.*)$`)
