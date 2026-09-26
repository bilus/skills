// Package code indexes the Go declarations and the changed files of a change in both versions.
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

// Index is the Go declarations and the changed files of a change, in both versions.
type Index struct {
	Before, After map[string]Place     // declarations by key, such as "analyze.sumTypes"
	ChangedDecls  map[string]bool      // the keys of the declarations that differ between the versions
	Std           map[string]bool      // the standard library's import paths
	Files         map[string]repo.Pair // the files the page shows, by path
	Changed       []repo.Change
}

// Place is where a declaration sits, its doc comment included.
type Place struct {
	File       string
	Start, End int // lines, from 1
}

// Read indexes the declarations of both versions and the files changed since the base.
// The page shows the files of the declarations that the boxes and the type comments
// name, and every changed file. Without a base, ChangedDecls is empty.
func Read(r *repo.Repo, d *design.Design) (*Index, error) {
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
		idx.ChangedDecls = changedDecls(before, after, idx.Before, idx.After)
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
	return idx, nil
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
