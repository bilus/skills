package dfdmetrics

// Summaries checks the summaries of d against items.
func Summaries(d *Design, items []*Item) []Finding {
	type key struct{ store, label string }
	byKey := map[key]*Item{}
	for _, it := range items {
		byKey[key{it.Store, it.Label}] = it
	}
	type summary struct {
		box          *Box
		store, label string
		write        bool
	}
	drawn := map[summary]bool{}
	var fs []Finding
	for _, dg := range d.Diagrams {
		for _, b := range dg.Boxes {
			if b.Child == nil {
				continue
			}
			for _, a := range b.Accesses {
				if a.Item == "" {
					continue
				}
				drawn[summary{b, a.Store, a.Item, a.Write}] = true
				it := byKey[key{a.Store, a.Item}]
				f := Finding{Store: a.Store, Label: a.Item, Box: b, Write: a.Write}
				if !matched(it, b, a.Write) {
					f.Kind = UnmatchedSummary
					fs = append(fs, f)
				}
				if it != nil && aboveHome(it, b) {
					f.Kind = SummaryAboveHome
					fs = append(fs, f)
				}
			}
		}
	}
	reported := map[summary]bool{}
	for _, it := range items {
		if it.Home == nil {
			continue
		}
		for _, side := range []struct {
			boxes []*Box
			write bool
		}{{it.Writers, true}, {it.Readers, false}} {
			for _, b := range side.boxes {
				for dg := b.Diagram; dg != it.Home; dg = dg.Parent.Diagram {
					s := summary{dg.Parent, it.Store, it.Label, side.write}
					if !drawn[s] && !reported[s] {
						reported[s] = true
						fs = append(fs, Finding{Kind: MissingSummary, Store: it.Store, Label: it.Label, Box: dg.Parent, Write: side.write})
					}
				}
			}
		}
	}
	return fs
}

// matched reports whether a leaf box below s makes the access that s summarizes.
func matched(it *Item, s *Box, write bool) bool {
	if it == nil {
		return false
	}
	boxes := it.Readers
	if write {
		boxes = it.Writers
	}
	for _, b := range boxes {
		if below(b, s) {
			return true
		}
	}
	return false
}

// aboveHome reports whether s sits on or above the box expanding the item's home.
func aboveHome(it *Item, s *Box) bool {
	if it.Home == nil {
		for _, boxes := range [][]*Box{it.Writers, it.Readers} {
			for _, b := range boxes {
				if below(b, s) {
					return true
				}
			}
		}
		return false
	}
	for dg := it.Home; dg.Parent != nil; dg = dg.Parent.Diagram {
		if dg.Parent == s {
			return true
		}
	}
	return false
}

// below reports whether b lies inside the child diagrams of s.
func below(b, s *Box) bool {
	for x := b.Diagram.Parent; x != nil; x = x.Diagram.Parent {
		if x == s {
			return true
		}
	}
	return false
}

// Pressure is the busiest gap of one diagram.
type Pressure struct {
	Diagram *Diagram
	Max     int
	Gaps    int  // the number of gaps with the maximum
	After   *Box // the box before the first such gap
}

// Pressures returns the pressure of each diagram with two or more boxes.
func Pressures(d *Design, items []*Item) []Pressure {
	var out []Pressure
	for _, dg := range d.Diagrams {
		n := len(dg.Boxes)
		if n < 2 {
			continue
		}
		live := make([]int, n) // live[k] counts the items across the gap after position k
		for _, it := range items {
			if it.Home != dg || !it.HasLiveRange {
				continue
			}
			first, last := n, 0
			for _, w := range it.Writers {
				first = min(first, position(w, dg))
			}
			for _, r := range it.Readers {
				last = max(last, position(r, dg))
			}
			for k := first; k < last; k++ {
				live[k]++
			}
		}
		p := Pressure{Diagram: dg}
		for k := 1; k < n; k++ {
			switch {
			case p.After == nil || live[k] > p.Max:
				p.Max, p.Gaps, p.After = live[k], 1, dg.Boxes[k-1]
			case live[k] == p.Max:
				p.Gaps++
			}
		}
		out = append(out, p)
	}
	return out
}
