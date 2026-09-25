package dfdmetrics

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
)

// Spread is the set of design packages among a diagram's own boxes.
type Spread struct {
	Diagram  *Diagram
	Packages []string // sorted
}

// Spreads returns each diagram's design packages, with the package findings.
func Spreads(d *Design, std map[string]bool) ([]Spread, []Finding) {
	var spreads []Spread
	var fs []Finding
	for _, dg := range d.Diagrams {
		all := map[string]bool{}
		flagged := map[string]bool{}
		for _, b := range dg.Boxes {
			pkgs := designPackages(b.Refs, std)
			for _, p := range pkgs {
				all[p] = true
			}
			if len(pkgs) >= 2 && !flagged[b.Number] {
				flagged[b.Number] = true
				fs = append(fs, Finding{Kind: MixedBox, Box: b, Diagram: dg, Packages: pkgs})
			}
		}
		pkgs := sortedKeys(all)
		spreads = append(spreads, Spread{Diagram: dg, Packages: pkgs})
		if dg.Depth > 0 && len(pkgs) >= 2 {
			fs = append(fs, Finding{Kind: MixedDiagram, Diagram: dg, Packages: pkgs})
		}
	}
	return spreads, fs
}

// designPackages returns the sorted qualifiers of refs that name design packages.
func designPackages(refs []Ref, std map[string]bool) []string {
	set := map[string]bool{}
	for _, r := range refs {
		if !strings.Contains(r.Qualifier, "/") && !std[r.Qualifier] {
			set[r.Qualifier] = true
		}
	}
	return sortedKeys(set)
}

func sortedKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// StdPackages returns the standard library packages without a slash in their path.
func StdPackages() (map[string]bool, error) {
	out, err := exec.Command("go", "list", "std").Output()
	if err != nil {
		return nil, fmt.Errorf("go list std: %w", err)
	}
	std := map[string]bool{}
	for _, path := range strings.Fields(string(out)) {
		if !strings.Contains(path, "/") {
			std[path] = true
		}
	}
	return std, nil
}

// Report is the complete analysis of one design.
type Report struct {
	Design    *Design
	Items     []*Item
	Pressures []Pressure
	Spreads   []Spread
	Findings  []Finding
}

// Analyze runs every measurement on d.
func Analyze(d *Design, std map[string]bool) *Report {
	items, fs := Items(d)
	fs = append(fs, Summaries(d, items)...)
	spreads, pf := Spreads(d, std)
	fs = append(fs, pf...)
	sort.SliceStable(fs, func(i, j int) bool { return lessFinding(fs[i], fs[j]) })
	return &Report{Design: d, Items: items, Pressures: Pressures(d, items), Spreads: spreads, Findings: fs}
}

// lessFinding orders findings by kind, then by item, then by box and diagram number.
func lessFinding(a, b Finding) bool {
	if a.Kind != b.Kind {
		return a.Kind < b.Kind
	}
	if a.Store != b.Store {
		return a.Store < b.Store
	}
	if a.Label != b.Label {
		return a.Label < b.Label
	}
	if an, bn := boxNumber(a.Box), boxNumber(b.Box); an != bn {
		return lessNumber(an, bn)
	}
	if an, bn := diagramNumber(a.Diagram), diagramNumber(b.Diagram); an != bn {
		return lessNumber(an, bn)
	}
	return a.Write && !b.Write
}

func boxNumber(b *Box) string {
	if b == nil {
		return ""
	}
	return b.Number
}

func diagramNumber(dg *Diagram) string {
	if dg == nil {
		return ""
	}
	return dg.Number
}

// Score is the score card of a report. Every value is better when lower.
type Score struct {
	LiveRange     int // the sum of positive live ranges
	ReadDistance  int // the largest read distance
	TopShared     int // shared items at level 0
	Shared        int // shared items at any level
	Pressure      int // the largest pressure of any diagram
	PackageSpread int // mixed boxes and mixed diagrams
	Findings      int // every finding
}

// Score returns the score card of r.
func (r *Report) Score() Score {
	var s Score
	for _, it := range r.Items {
		if it.HasLiveRange && it.LiveRange > 0 {
			s.LiveRange += it.LiveRange
		}
		if it.HasReadDistance {
			s.ReadDistance = max(s.ReadDistance, it.ReadDistance)
		}
		if it.Home != nil {
			s.Shared++
			if it.Level == 0 {
				s.TopShared++
			}
		}
	}
	for _, p := range r.Pressures {
		s.Pressure = max(s.Pressure, p.Max)
	}
	for _, f := range r.Findings {
		if f.Kind == MixedBox || f.Kind == MixedDiagram {
			s.PackageSpread++
		}
	}
	s.Findings = len(r.Findings)
	return s
}

// WriteScore prints the score card as a tab-separated header and one row.
func (r *Report) WriteScore(w io.Writer) error {
	s := r.Score()
	p := &printer{w: w}
	p.printf("live_range\tread_distance\ttop_shared\tshared\tpressure\tpackage_spread\tfindings\n")
	p.printf("%d\t%d\t%d\t%d\t%d\t%d\t%d\n", s.LiveRange, s.ReadDistance, s.TopShared, s.Shared, s.Pressure, s.PackageSpread, s.Findings)
	return p.err
}

// String returns the finding kind as the report names it.
func (k FindingKind) String() string {
	names := []string{"unlabeled arrow", "dead item", "orphan item", "read before write", "arrow candidate",
		"unmatched summary", "missing summary", "summary above home", "mixed box", "mixed diagram"}
	if int(k) < 0 || int(k) >= len(names) {
		return fmt.Sprintf("finding %d", int(k))
	}
	return names[k]
}

// printer writes formatted text and keeps the first write error.
type printer struct {
	w   io.Writer
	err error
}

func (p *printer) printf(format string, args ...any) {
	if p.err == nil {
		_, p.err = fmt.Fprintf(p.w, format, args...)
	}
}

// Write prints the report as plain text to w.
func (r *Report) Write(w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	p := &printer{w: tw}
	boxCount := 0
	for _, dg := range r.Design.Diagrams {
		boxCount += len(dg.Boxes)
	}
	p.printf("design %s: %s, %s\n", name(r.Design.Diagrams[0]), count(len(r.Design.Diagrams), "diagram"), count(boxCount, "box"))
	r.writeItems(p)
	p.printf("\npressure\n")
	if len(r.Pressures) == 0 {
		p.printf("  none\n")
	}
	for _, pr := range r.Pressures {
		next := pr.Diagram.Boxes[pr.After.Position]
		p.printf("  %s\tmax %d in %d gaps, first between %s and %s\n", name(pr.Diagram), pr.Max, pr.Gaps, pr.After.Number, next.Number)
	}
	p.printf("\nfindings\n")
	if len(r.Findings) == 0 {
		p.printf("  none\n")
	}
	for _, f := range r.Findings {
		p.printf("  %s\t%s\n", f.Kind, subject(f))
	}
	p.printf("\ndesign packages\n")
	for _, s := range r.Spreads {
		p.printf("  %s, depth %d\t%s\n", name(s.Diagram), s.Diagram.Depth, list(s.Packages))
	}
	r.writeTotals(p)
	if p.err != nil {
		return p.err
	}
	return tw.Flush()
}

// writeItems prints the shared items grouped by home, then the private items.
func (r *Report) writeItems(p *printer) {
	var homes []*Diagram
	byHome := map[*Diagram][]*Item{}
	var private []*Item
	for _, it := range r.Items {
		if it.Home == nil {
			private = append(private, it)
			continue
		}
		if byHome[it.Home] == nil {
			homes = append(homes, it.Home)
		}
		byHome[it.Home] = append(byHome[it.Home], it)
	}
	sort.Slice(homes, func(i, j int) bool {
		if homes[i].Depth != homes[j].Depth {
			return homes[i].Depth < homes[j].Depth
		}
		return lessNumber(homes[i].Number, homes[j].Number)
	})
	for _, h := range homes {
		items := byHome[h]
		sort.SliceStable(items, func(i, j int) bool { return rankedBefore(items[i], items[j]) })
		p.printf("\nitems at home %s, level %d\n", name(h), h.Depth)
		for _, it := range items {
			p.printf("  %s: %s\tlive range %s\tread distance %s\twriters %s\treaders %s\n", it.Store, it.Label,
				optional(it.LiveRange, it.HasLiveRange), optional(it.ReadDistance, it.HasReadDistance), boxList(it.Writers), boxList(it.Readers))
		}
	}
	p.printf("\nprivate items\n")
	if len(private) == 0 {
		p.printf("  none\n")
	}
	for _, it := range private {
		p.printf("  %s: %s\twriters %s\treaders %s\n", it.Store, it.Label, boxList(it.Writers), boxList(it.Readers))
	}
}

// rankedBefore orders items by live range, then read distance, longest first.
func rankedBefore(a, b *Item) bool {
	if a.HasLiveRange != b.HasLiveRange {
		return a.HasLiveRange
	}
	if a.LiveRange != b.LiveRange {
		return a.LiveRange > b.LiveRange
	}
	if a.HasReadDistance != b.HasReadDistance {
		return a.HasReadDistance
	}
	if a.ReadDistance != b.ReadDistance {
		return a.ReadDistance > b.ReadDistance
	}
	if a.Store != b.Store {
		return a.Store < b.Store
	}
	return a.Label < b.Label
}

// writeTotals prints the counts that runs compare against each other.
func (r *Report) writeTotals(p *printer) {
	p.printf("\ntotals\n")
	shared, private := map[int]int{}, 0
	for _, it := range r.Items {
		if it.Home == nil {
			private++
		} else {
			shared[it.Level]++
		}
	}
	p.printf("  shared items by level\t%s\n", byDepth(shared))
	p.printf("  private items\t%d\n", private)
	counts := map[FindingKind]int{}
	for _, f := range r.Findings {
		counts[f.Kind]++
	}
	for k := UnlabeledArrow; k <= MixedDiagram; k++ {
		p.printf("  %s\t%d\n", k, counts[k])
	}
	mixed := map[int]int{}
	for _, s := range r.Spreads {
		if len(s.Packages) >= 2 {
			mixed[s.Diagram.Depth]++
		}
	}
	p.printf("  diagrams with two or more design packages by depth\t%s\n", byDepth(mixed))
}

// subject names what a finding is about.
func subject(f Finding) string {
	direction := "read"
	if f.Write {
		direction = "write"
	}
	switch f.Kind {
	case UnlabeledArrow:
		return fmt.Sprintf("store %s on box %s", f.Store, f.Box.Number)
	case DeadItem, OrphanItem:
		return fmt.Sprintf("%s: %s", f.Store, f.Label)
	case ReadBeforeWrite:
		return fmt.Sprintf("%s: %s, read on box %s", f.Store, f.Label, f.Box.Number)
	case ArrowCandidate:
		next := f.Box.Diagram.Boxes[f.Box.Position]
		return fmt.Sprintf("%s: %s, between %s and %s", f.Store, f.Label, f.Box.Number, next.Number)
	case UnmatchedSummary, MissingSummary, SummaryAboveHome:
		return fmt.Sprintf("%s: %s, %s on box %s", f.Store, f.Label, direction, f.Box.Number)
	case MixedBox:
		return fmt.Sprintf("box %s in %s: %s", f.Box.Number, name(f.Diagram), list(f.Packages))
	case MixedDiagram:
		return fmt.Sprintf("%s: %s", name(f.Diagram), list(f.Packages))
	}
	return ""
}

func name(dg *Diagram) string { return filepath.Base(dg.Path) }

func list(xs []string) string {
	if len(xs) == 0 {
		return "none"
	}
	return strings.Join(xs, ", ")
}

// boxList returns the distinct numbers of boxes in number order, or "-".
func boxList(boxes []*Box) string {
	seen := map[string]bool{}
	var nums []string
	for _, b := range boxes {
		if !seen[b.Number] {
			seen[b.Number] = true
			nums = append(nums, b.Number)
		}
	}
	if len(nums) == 0 {
		return "-"
	}
	sort.Slice(nums, func(i, j int) bool { return lessNumber(nums[i], nums[j]) })
	return strings.Join(nums, ", ")
}

// count returns n with noun, adding the plural ending when n is not 1.
func count(n int, noun string) string {
	switch {
	case n == 1:
		return "1 " + noun
	case strings.HasSuffix(noun, "x"):
		return fmt.Sprintf("%d %ses", n, noun)
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

func optional(n int, ok bool) string {
	if !ok {
		return "-"
	}
	return fmt.Sprint(n)
}

// byDepth formats counts by depth as "0: 2, 1: 3", or "none".
func byDepth(counts map[int]int) string {
	var depths []int
	for d, n := range counts {
		if n > 0 {
			depths = append(depths, d)
		}
	}
	if len(depths) == 0 {
		return "none"
	}
	sort.Ints(depths)
	parts := make([]string, 0, len(depths))
	for _, d := range depths {
		parts = append(parts, fmt.Sprintf("%d: %d", d, counts[d]))
	}
	return strings.Join(parts, ", ")
}
