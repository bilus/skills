package dfdmetrics

import "sort"

// Item is one state item with the boxes that access it.
type Item struct {
	Store           string
	Label           string
	Writers         []*Box
	Readers         []*Box
	Home            *Diagram // nil for a private item
	Level           int      // the depth of Home
	LiveRange       int
	HasLiveRange    bool
	ReadDistance    int
	HasReadDistance bool
}

// Finding is one problem the analysis reports.
type Finding struct {
	Kind     FindingKind
	Store    string   // the item's store, "" for a package finding
	Label    string   // the item's label, "" for a package finding
	Box      *Box     // nil when the finding concerns a whole item
	Write    bool     // the direction of a summary finding
	Diagram  *Diagram // the diagram of a package finding
	Packages []string // the design packages of a package finding
}

// FindingKind is the rule behind a finding.
type FindingKind int

// The findings, one per rule of the plan.
const (
	UnlabeledArrow   FindingKind = iota // a store arrow without an item
	DeadItem                            // an item without a reader
	OrphanItem                          // an item without a writer
	ReadBeforeWrite                     // a read before the item's first write
	ArrowCandidate                      // an item that could travel on an arrow
	UnmatchedSummary                    // a summary without a leaf access below it
	MissingSummary                      // an ancestor box without the item's summary
	SummaryAboveHome                    // a summary above the item's home
	MixedBox                            // a box whose references name several design packages
	MixedDiagram                        // a child diagram whose boxes name several design packages
)

// Items collects the state items of d with their findings.
func Items(d *Design) ([]*Item, []Finding) {
	type key struct{ store, label string }
	byKey := map[key]*Item{}
	var findings []Finding
	for _, dg := range d.Diagrams {
		for _, b := range dg.Boxes {
			for _, a := range b.Accesses {
				if a.Item == "" {
					findings = append(findings, Finding{Kind: UnlabeledArrow, Store: a.Store, Box: b})
					continue
				}
				if b.Child != nil {
					continue // a summary, checked by Summaries
				}
				k := key{a.Store, a.Item}
				it := byKey[k]
				if it == nil {
					it = &Item{Store: a.Store, Label: a.Item}
					byKey[k] = it
				}
				if a.Write {
					it.Writers = addBox(it.Writers, b)
				} else {
					it.Readers = addBox(it.Readers, b)
				}
			}
		}
	}
	items := make([]*Item, 0, len(byKey))
	for _, it := range byKey {
		items = append(items, it)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Store != items[j].Store {
			return items[i].Store < items[j].Store
		}
		return items[i].Label < items[j].Label
	})
	for _, it := range items {
		findings = append(findings, measure(it)...)
	}
	return items, findings
}

// measure sets the home, level and distances of it, and returns its findings.
func measure(it *Item) []Finding {
	var fs []Finding
	finding := func(k FindingKind, b *Box) {
		fs = append(fs, Finding{Kind: k, Store: it.Store, Label: it.Label, Box: b})
	}
	if len(it.Readers) == 0 {
		finding(DeadItem, nil)
	}
	if len(it.Writers) == 0 {
		finding(OrphanItem, nil)
	}
	accessors := append([]*Box(nil), it.Writers...)
	for _, r := range it.Readers {
		accessors = addBox(accessors, r)
	}
	if len(accessors) < 2 {
		return fs
	}
	it.Home = home(accessors)
	it.Level = it.Home.Depth
	if len(it.Writers) == 0 || len(it.Readers) == 0 {
		return fs
	}
	writes := make([]int, 0, len(it.Writers))
	for _, w := range it.Writers {
		writes = append(writes, position(w, it.Home))
	}
	sort.Ints(writes)
	lastRead, early := 0, false
	for _, r := range it.Readers {
		p := position(r, it.Home)
		lastRead = max(lastRead, p)
		// SearchInts finds the first write after p, so the one before is nearest.
		i := sort.SearchInts(writes, p+1) - 1
		if i < 0 {
			finding(ReadBeforeWrite, r)
			early = true
			continue
		}
		if d := p - writes[i]; !it.HasReadDistance || d > it.ReadDistance {
			it.ReadDistance, it.HasReadDistance = d, true
		}
	}
	it.LiveRange, it.HasLiveRange = lastRead-writes[0], true
	if before := it.Home.Boxes[writes[0]-1]; it.LiveRange == 1 && !early && !before.EntityAfter {
		finding(ArrowCandidate, before)
	}
	return fs
}

// addBox appends b to boxes unless boxes already holds it.
func addBox(boxes []*Box, b *Box) []*Box {
	for _, x := range boxes {
		if x == b {
			return boxes
		}
	}
	return append(boxes, b)
}

// home returns the deepest diagram that holds every box in boxes.
func home(boxes []*Box) *Diagram {
	if len(boxes) == 0 {
		return nil
	}
	for dg := boxes[0].Diagram; ; dg = dg.Parent.Diagram {
		holdsAll := true
		for _, b := range boxes {
			if position(b, dg) == 0 {
				holdsAll = false
				break
			}
		}
		if holdsAll || dg.Parent == nil {
			return dg
		}
	}
}

// position returns the position in dg of b or its ancestor box, else 0.
func position(b *Box, dg *Diagram) int {
	for x := b; x != nil; x = x.Diagram.Parent {
		if x.Diagram == dg {
			return x.Position
		}
	}
	return 0
}
