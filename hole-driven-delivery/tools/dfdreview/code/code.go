// Package code loads the module of a change with its types, and indexes its declarations,
// changed files and identifier links in both versions.
package code

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os/exec"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/design"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/dfdtext"
	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/repo"
)

// Index is the code of a change in both versions: the declarations and methods with the changed
// ones, the files the page shows with their identifier links, and the package errors.
type Index struct {
	Before, After               map[string]Place     // declarations by key, such as "analyze.sumTypes"
	BeforeMethods, AfterMethods map[string]Place     // methods by key, such as "render.renderer.text"
	ChangedDecls                map[string]bool      // the keys of the changed declarations and methods
	Reached                     map[string]bool      // the keys of those changed through their reach only
	Std                         map[string]bool      // the standard library's import paths
	Files                       map[string]repo.Pair // the files the page shows, by path
	Changed                     []repo.Change
	BeforeLinks, AfterLinks     map[string][]Link // the identifier links of each Go file, by path
	BeforeUses, AfterUses       map[string][]Use  // the uses of each declaration and method, by key
	Errors                      []string          // the package errors of both versions
	Uncovered                   []Uncovered       // the changed declarations and methods that no drawing covers, by key
}

// Use is an identifier that names a declaration or a method: its line, and the key of the
// function, type, variable, constant or method whose declaration holds it.
type Use struct {
	In   string // "" for an import
	File string
	Line int
}

// Uncovered is a changed declaration or method that no drawing covers: no drawing links it, and
// none links a function or method whose reach holds it.
type Uncovered struct {
	Key     string
	Mark    string // MarkChanged or MarkReached
	Place   Place
	Version string // "after", or "before" for one that the working tree removed
}

// Place is where a declaration or a method sits, its doc comment included, with its kind.
type Place struct {
	File       string
	Start, End int    // lines, from 1
	Kind       string // "func", "type", "var" or "const" for a declaration, "method" for a method
}

// Read loads the module with its types, then indexes the declarations and methods, the
// changed files, and the identifier links and uses of both versions. With a module, the page
// shows every Go file of both versions, test files included; without one, the files of the
// declarations that the boxes and the type comments name, and no uses. Every changed file
// shows too. Without a base, ChangedDecls is empty.
func Read(r *repo.Repo, d *design.Design) (*Index, error) {
	res, err := Load(r)
	if err != nil {
		return nil, err
	}
	before, after, err := r.GoFiles()
	if err != nil {
		return nil, err
	}
	idx := &Index{Files: map[string]repo.Pair{}}
	if idx.Before, err = Declarations(before); err != nil {
		return nil, fmt.Errorf("at the base: %w", err)
	}
	if idx.After, err = Declarations(after); err != nil {
		return nil, err
	}
	if idx.BeforeMethods, err = methodPlaces(before); err != nil {
		return nil, fmt.Errorf("at the base: %w", err)
	}
	if idx.AfterMethods, err = methodPlaces(after); err != nil {
		return nil, err
	}
	if r.Base() != "" {
		beforeAll, afterAll := union(idx.Before, idx.BeforeMethods), union(idx.After, idx.AfterMethods)
		idx.ChangedDecls = changedDecls(before, after, beforeAll, afterAll)
		var calls map[string][]string
		if res != nil && res.After != nil {
			calls = res.After.Calls
		}
		idx.Reached = map[string]bool{}
		for key := range reached(calls, idx.ChangedDecls) {
			if !idx.ChangedDecls[key] {
				idx.Reached[key] = true
			}
		}
		for key := range idx.Reached {
			idx.ChangedDecls[key] = true
		}
		idx.Uncovered = uncovered(idx, linked(d), calls, beforeAll, afterAll)
	}
	if idx.Std, err = stdPackages(); err != nil {
		return nil, err
	}
	if idx.Changed, err = r.Changed(); err != nil {
		return nil, err
	}
	show := func(file string) {
		p := idx.Files[file]
		if content, ok := before[file]; ok {
			p.Before = repo.Text{Content: content, Found: true}
		}
		if content, ok := after[file]; ok {
			p.After = repo.Text{Content: content, Found: true}
		}
		idx.Files[file] = p
	}
	for _, dg := range d.Diagrams {
		for _, src := range []repo.Text{dg.Source.Before, dg.Source.After} {
			keys := typeNames(dfdtext.Types(src.Content))
			for _, ref := range dfdtext.References(src.Content) {
				keys = append(keys, Key(ref))
			}
			for _, key := range keys {
				for _, decls := range []map[string]Place{idx.Before, idx.After} {
					if p, ok := decls[key]; ok {
						show(p.File)
					}
				}
			}
		}
	}
	for _, c := range idx.Changed {
		idx.Files[c.Path] = c.Content
	}
	if res != nil {
		// With a module, every Go file shows, in the texts that the load parsed.
		if res.Before != nil {
			for file, text := range res.Before.Files {
				p := idx.Files[file]
				p.Before = repo.Text{Content: text, Found: true}
				idx.Files[file] = p
			}
			idx.BeforeLinks = markLinks(res.Before.Links, idx)
			idx.BeforeUses = uses(idx.BeforeLinks)
		}
		if res.After != nil {
			for file, text := range res.After.Files {
				p := idx.Files[file]
				p.After = repo.Text{Content: text, Found: true}
				idx.Files[file] = p
			}
			idx.AfterLinks = markLinks(res.After.Links, idx)
			idx.AfterUses = uses(idx.AfterLinks)
		}
		idx.Errors = res.Errors
		// A version without a resolution keeps the texts from git.
		for file, p := range idx.Files {
			if content, ok := before[file]; ok && !p.Before.Found {
				p.Before = repo.Text{Content: content, Found: true}
			}
			if content, ok := after[file]; ok && !p.After.Found {
				p.After = repo.Text{Content: content, Found: true}
			}
			idx.Files[file] = p
		}
	}
	return idx, nil
}

// markLinks gives each link of files inside a changed declaration the mark of the changed
// functions and methods of the module that it names: MarkChanged when one of them changed
// in its own text, MarkReached when all of them changed through their reach only.
func markLinks(files map[string][]Link, idx *Index) map[string][]Link {
	for _, links := range files {
		for i := range links {
			l := &links[i]
			if !idx.ChangedDecls[l.In] {
				continue
			}
			for _, key := range l.Funcs {
				switch {
				case idx.ChangedDecls[key] && !idx.Reached[key]:
					l.Mark = MarkChanged
				case idx.Reached[key] && l.Mark == "":
					l.Mark = MarkReached
				}
			}
		}
	}
	return files
}

// uses returns the uses of each declaration and method in files, the identifier links of each
// Go file by path, in order of file and line: the links that name it by their key, or as one
// of their functions, which a call of an interface method adds for each implementation.
func uses(files map[string][]Link) map[string][]Use {
	out := map[string][]Use{}
	for file, links := range files {
		for _, l := range links {
			if l.End > 0 {
				continue // the name of a declaration, not a use
			}
			for _, key := range append([]string{l.Key}, l.Funcs...) {
				us := out[key]
				u := Use{In: l.In, File: file, Line: l.Line}
				// A key can repeat in a link, and a line can hold several links to it.
				if key != "" && (len(us) == 0 || us[len(us)-1] != u) {
					out[key] = append(us, u)
				}
			}
		}
	}
	for _, us := range out {
		sort.Slice(us, func(i, j int) bool {
			if us[i].File != us[j].File {
				return us[i].File < us[j].File
			}
			if us[i].Line != us[j].Line {
				return us[i].Line < us[j].Line
			}
			return us[i].In < us[j].In
		})
	}
	return out
}

// linked returns the keys of the declarations that the diagrams link in either version: their
// references, and the first named type of each type comment, which the page links.
func linked(d *design.Design) map[string]bool {
	keys := map[string]bool{}
	for _, dg := range d.Diagrams {
		for _, src := range []repo.Text{dg.Source.Before, dg.Source.After} {
			for _, ref := range dfdtext.References(src.Content) {
				keys[Key(ref)] = true
			}
			for _, t := range dfdtext.Types(src.Content) {
				if key := qualifiedType.FindString(t); key != "" {
					keys[key] = true
				}
			}
		}
	}
	return keys
}

// uncovered returns the changed declarations and methods of idx that no drawing covers: none of
// them is a key of linked, or lies in the reach over calls of one. before and after hold the
// places of each version's declarations and methods.
func uncovered(idx *Index, linked map[string]bool, calls map[string][]string, before, after map[string]Place) []Uncovered {
	covered := map[string]bool{}
	var queue []string
	for key := range linked {
		covered[key] = true
		queue = append(queue, key)
	}
	for len(queue) > 0 {
		key := queue[0]
		queue = queue[1:]
		for _, callee := range calls[key] {
			if !covered[callee] {
				covered[callee] = true
				queue = append(queue, callee)
			}
		}
	}
	var out []Uncovered
	for key := range idx.ChangedDecls {
		if covered[key] {
			continue
		}
		u := Uncovered{Key: key, Mark: MarkChanged}
		if idx.Reached[key] {
			u.Mark = MarkReached
		}
		if p, ok := after[key]; ok {
			u.Place, u.Version = p, "after"
		} else if p, ok := before[key]; ok {
			u.Place, u.Version = p, "before"
		}
		out = append(out, u)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

// reached returns the keys of the functions and methods whose reach over calls holds a key
// of changed.
func reached(calls map[string][]string, changed map[string]bool) map[string]bool {
	callers := map[string][]string{}
	for caller, callees := range calls {
		for _, callee := range callees {
			callers[callee] = append(callers[callee], caller)
		}
	}
	var queue []string
	for key := range changed {
		queue = append(queue, key)
	}
	out := map[string]bool{}
	for len(queue) > 0 {
		key := queue[0]
		queue = queue[1:]
		for _, caller := range callers[key] {
			if !out[caller] {
				out[caller] = true
				queue = append(queue, caller)
			}
		}
	}
	return out
}

// union returns the places of a and b in one map; b wins a shared key.
func union(a, b map[string]Place) map[string]Place {
	out := make(map[string]Place, len(a)+len(b))
	for k, p := range a {
		out[k] = p
	}
	for k, p := range b {
		out[k] = p
	}
	return out
}

// changedDecls returns the keys of the declarations whose lines, doc comment included,
// differ between the versions, and of those that exist in one version only.
func changedDecls(before, after map[string]string, beforeDecls, afterDecls map[string]Place) map[string]bool {
	keys := map[string]bool{}
	for key, b := range beforeDecls {
		if a, ok := afterDecls[key]; !ok || lines(before, b) != lines(after, a) {
			keys[key] = true
		}
	}
	for key := range afterDecls {
		if _, ok := beforeDecls[key]; !ok {
			keys[key] = true
		}
	}
	return keys
}

// lines returns the text of the lines at p in files, or "" when p lies outside its file.
func lines(files map[string]string, p Place) string {
	all := strings.Split(files[p.File], "\n")
	if p.Start < 1 || p.End > len(all) || p.Start > p.End {
		return ""
	}
	return strings.Join(all[p.Start-1:p.End], "\n")
}

// typeNames returns the declaration keys of the qualified names in types, such as lib.Item in []*lib.Item.
func typeNames(types map[string]string) []string {
	var keys []string
	for _, t := range types {
		keys = append(keys, qualifiedType.FindAllString(t, -1)...)
	}
	return keys
}

var qualifiedType = regexp.MustCompile(`\b[a-z][a-z0-9_]*\.[A-Za-z_][A-Za-z0-9_]*`)

// methodPlaces returns the places of the methods of files, doc comments included, keyed like
// render.renderer.text. The first file in path order wins a shared key.
//
// Example: it places render.renderer.text on lines 3 to 6, with the kind method.
//
//	package render
//
//	// text returns the source between two positions.
//	func (r *renderer) text(from, to token.Pos) string {
//		return r.src[from:to]
//	}
func methodPlaces(files map[string]string) (map[string]Place, error) {
	paths := make([]string, 0, len(files))
	for p := range files {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	fset := token.NewFileSet()
	out := map[string]Place{}
	for _, p := range paths {
		f, err := parser.ParseFile(fset, p, files[p], parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			return nil, err
		}
		for _, d := range f.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Recv == nil || len(fd.Recv.List) == 0 {
				continue
			}
			recv := receiverName(fd.Recv.List[0].Type)
			key := f.Name.Name + "." + recv + "." + fd.Name.Name
			if _, taken := out[key]; taken || recv == "" {
				continue
			}
			start, end := funcSpan(fd)
			out[key] = Place{File: p, Start: fset.PositionFor(start, false).Line, End: fset.PositionFor(end, false).Line, Kind: "method"}
		}
	}
	return out, nil
}

// receiverName returns the name of a receiver's type, through a pointer and type
// parameters, or "" for another expression.
func receiverName(t ast.Expr) string {
	for {
		switch e := t.(type) {
		case *ast.StarExpr:
			t = e.X
		case *ast.IndexExpr:
			t = e.X
		case *ast.IndexListExpr:
			t = e.X
		case *ast.ParenExpr:
			t = e.X
		case *ast.Ident:
			return e.Name
		default:
			return ""
		}
	}
}

// Key returns the declaration key of a reference: its package's name and its name.
func Key(ref dfdtext.Reference) string {
	return path.Base(ref.Qualifier) + "." + ref.Name
}

// Declarations returns the top-level declarations of files, keyed like "analyze.sumTypes".
// The key joins the package name and the identifier; the first file in path order wins.
// Methods are left out, since a box names a method by its type. The lines are the file's
// own, whatever its //line directives say.
//
// Example: it places alpha.F on lines 3 to 4 with the kind func, and alpha.T on lines 7 to
// 8 and alpha.U on line 9 with the kind type.
//
//	package alpha
//
//	// F does x.
//	func F() {}
//
//	type (
//		// T is t.
//		T int
//		U string
//	)
func Declarations(files map[string]string) (map[string]Place, error) {
	paths := make([]string, 0, len(files))
	for p := range files {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	fset := token.NewFileSet()
	decls := map[string]Place{}
	for _, p := range paths {
		f, err := parser.ParseFile(fset, p, files[p], parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			return nil, err
		}
		add := func(name, kind string, start, end token.Pos) {
			key := f.Name.Name + "." + name
			if _, taken := decls[key]; !taken {
				decls[key] = Place{File: p, Start: fset.PositionFor(start, false).Line, End: fset.PositionFor(end, false).Line, Kind: kind}
			}
		}
		for _, d := range f.Decls {
			switch d := d.(type) {
			case *ast.FuncDecl:
				if d.Recv == nil {
					start, end := funcSpan(d)
					add(d.Name.Name, "func", start, end)
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					start, end := specSpan(d, spec)
					switch s := spec.(type) {
					case *ast.TypeSpec:
						add(s.Name.Name, "type", start, end)
					case *ast.ValueSpec:
						for _, n := range s.Names {
							add(n.Name, d.Tok.String(), start, end)
						}
					}
				}
			}
		}
	}
	return decls, nil
}

// funcSpan returns the first and last positions of a function or a method, doc comment
// included.
//
// Example: it spans all four lines.
//
//	// Total sums the cost of the order's items.
//	func Total(o Order) int {
//		return sum(o.Items)
//	}
func funcSpan(d *ast.FuncDecl) (start, end token.Pos) {
	start = d.Pos()
	if d.Doc != nil {
		start = d.Doc.Pos()
	}
	return start, d.End()
}

// specSpan returns the first and last positions of a spec of a top-level declaration, doc
// comment included: the spec itself in a group, and the whole declaration without one.
//
// Example: it spans lines 1 to 2 for Order, and lines 5 to 6 for T.
//
//	// Order is the items a customer buys.
//	type Order struct{ Items []Item }
//
//	type (
//		// T is t.
//		T int
//	)
func specSpan(d *ast.GenDecl, spec ast.Spec) (start, end token.Pos) {
	var doc *ast.CommentGroup
	switch s := spec.(type) {
	case *ast.TypeSpec:
		doc = s.Doc
	case *ast.ValueSpec:
		doc = s.Doc
	}
	start, end = spec.Pos(), spec.End()
	if !d.Lparen.IsValid() {
		start, end, doc = d.Pos(), d.End(), d.Doc
	}
	if doc != nil {
		start = doc.Pos()
	}
	return start, end
}

// stdPackages returns the import paths of the standard library.
func stdPackages() (map[string]bool, error) {
	out, err := exec.Command("go", "list", "std").Output()
	if err != nil {
		return nil, fmt.Errorf("go list std: %w", err)
	}
	std := map[string]bool{}
	for _, p := range strings.Fields(string(out)) {
		std[p] = true
	}
	return std, nil
}
