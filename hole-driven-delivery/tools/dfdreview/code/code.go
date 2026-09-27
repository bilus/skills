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

// Index is the code of a change in both versions: the declarations with the changed ones,
// the files the page shows with their identifier links, and the package errors.
type Index struct {
	Before, After           map[string]Place     // declarations by key, such as "analyze.sumTypes"
	ChangedDecls            map[string]bool      // the keys of the changed declarations
	Std                     map[string]bool      // the standard library's import paths
	Files                   map[string]repo.Pair // the files the page shows, by path
	Changed                 []repo.Change
	BeforeLinks, AfterLinks map[string][]Link // the identifier links of each Go file, by path
	Errors                  []string          // the package errors of both versions
}

// Place is where a declaration sits, its doc comment included.
type Place struct {
	File       string
	Start, End int // lines, from 1
}

// Read loads the module with its types, then indexes the declarations, the changed files
// and the identifier links of both versions. With a module, the page shows every Go file
// of both versions, test files included; without one, the files of the declarations that
// the boxes and the type comments name. Every changed file shows too. Without a base,
// ChangedDecls is empty.
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
	if r.Base() != "" {
		if idx.ChangedDecls, err = changedDeclsAndMethods(before, after, idx); err != nil {
			return nil, err
		}
		if res != nil && res.After != nil {
			for key := range reached(res.After.Calls, idx.ChangedDecls) {
				idx.ChangedDecls[key] = true
			}
		}
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
			idx.BeforeLinks = res.Before.Links
		}
		if res.After != nil {
			for file, text := range res.After.Files {
				p := idx.Files[file]
				p.After = repo.Text{Content: text, Found: true}
				idx.Files[file] = p
			}
			idx.AfterLinks = res.After.Links
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

// changedDeclsAndMethods returns the keys of the changed declarations and methods of idx,
// methods keyed like render.renderer.text.
func changedDeclsAndMethods(before, after map[string]string, idx *Index) (map[string]bool, error) {
	beforeMethods, err := methodPlaces(before)
	if err != nil {
		return nil, fmt.Errorf("at the base: %w", err)
	}
	afterMethods, err := methodPlaces(after)
	if err != nil {
		return nil, err
	}
	return changedDecls(before, after, union(idx.Before, beforeMethods), union(idx.After, afterMethods)), nil
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
			start := fd.Pos()
			if fd.Doc != nil {
				start = fd.Doc.Pos()
			}
			out[key] = Place{File: p, Start: fset.PositionFor(start, false).Line, End: fset.PositionFor(fd.End(), false).Line}
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
// Methods are left out, since a box names a method by its type.
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
		add := func(name string, start, end token.Pos, doc *ast.CommentGroup) {
			if doc != nil {
				start = doc.Pos()
			}
			key := f.Name.Name + "." + name
			if _, taken := decls[key]; !taken {
				decls[key] = Place{File: p, Start: fset.Position(start).Line, End: fset.Position(end).Line}
			}
		}
		for _, d := range f.Decls {
			switch d := d.(type) {
			case *ast.FuncDecl:
				if d.Recv == nil {
					add(d.Name.Name, d.Pos(), d.End(), d.Doc)
				}
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					var names []*ast.Ident
					var doc *ast.CommentGroup
					switch s := spec.(type) {
					case *ast.TypeSpec:
						names, doc = []*ast.Ident{s.Name}, s.Doc
					case *ast.ValueSpec:
						names, doc = s.Names, s.Doc
					default:
						continue
					}
					// In a group, a declaration is its own spec; alone, it is the whole declaration.
					start, end := spec.Pos(), spec.End()
					if !d.Lparen.IsValid() {
						start, end, doc = d.Pos(), d.End(), d.Doc
					}
					for _, n := range names {
						add(n.Name, start, end, doc)
					}
				}
			}
		}
	}
	return decls, nil
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
