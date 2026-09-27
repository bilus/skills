package code

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"golang.org/x/tools/go/packages"

	"github.com/bilus/skills/hole-driven-delivery/tools/dfdreview/repo"
)

// Resolved holds the resolution of each version and the package errors of both.
type Resolved struct {
	Before, After *Resolution // nil for a version without a module
	Errors        []string    // each naming its version and its package
}

// Resolution is what the load finds in one version of the module.
type Resolution struct {
	Calls map[string][]string // by the key of each function and method, such as "render.renderer.text", the keys of the functions and methods of the module that it calls
	Files map[string]string   // the text of every Go file as the load parsed it, test files included, by path relative to the code directory
	Links map[string][]Link   // the identifier links of each Go file, by path
}

// Link is an identifier that uses a declared object, with the object's definition.
type Link struct {
	Line, Col, Len int      // the identifier's line, and its column and length in UTF-16 code units, from 1
	File           string   // the definition's file in the same version, relative to the code directory
	To             int      // the definition's line in File
	URL            string   // the definition's documentation, for one outside the code directory
	Funcs          []string // the keys of the functions and methods of the module that it names; for an interface method, those of the implementations
	In             string   // the key of the function or method whose declaration holds the identifier, "" outside one
	Mark           string   // MarkChanged or MarkReached for a link inside a changed declaration to a changed function or method, "" otherwise
}

// The marks of a changed declaration: a change in its own text, or a change in its reach only.
const (
	MarkChanged = "changed"
	MarkReached = "reached"
)

// Load type-checks the module in both versions. It records the calls that every function
// and method makes within the module, the text and the identifier links of every Go file,
// and the package errors. It reads the base from the base export, a temporary copy of the
// module directory's blobs at the base that it deletes at the end. It returns an error only
// when it cannot run: an error that the go command or the type checker reports becomes a
// package error of its version.
func Load(r *repo.Repo) (*Resolved, error) {
	beforeMod, afterMod, err := r.ModuleDirs()
	if err != nil {
		return nil, err
	}
	res := &Resolved{}
	if beforeMod != "" {
		var errs []string
		if res.Before, errs, err = loadBase(r, beforeMod); err != nil {
			return nil, err
		}
		res.Errors = append(res.Errors, errs...)
	}
	if afterMod != "" {
		var errs []string
		res.After, errs = load(r.Dir(), "working tree")
		res.Errors = append(res.Errors, errs...)
	}
	return res, nil
}

// loadBase type-checks the base from the base export of the module directory mod, relative
// to the code directory, and deletes the export.
func loadBase(r *repo.Repo, mod string) (res *Resolution, errs []string, err error) {
	tmp, err := os.MkdirTemp("", "dfdreview-base-*")
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if rerr := os.RemoveAll(tmp); rerr != nil && err == nil {
			err = rerr
		}
	}()
	if err := r.Export(mod, tmp); err != nil {
		return nil, nil, err
	}
	// The code directory sits at the same place below the module directory in both versions.
	below, err := filepath.Rel(filepath.Join(r.Dir(), mod), r.Dir())
	if err != nil {
		return nil, nil, err
	}
	res, errs = load(filepath.Join(tmp, below), "base")
	return res, errs, nil
}

// load type-checks the packages of dir and below, test files included, and records their
// calls, and the text and the identifier links of each of their Go files. Each error it
// returns names the version.
// A load that fails as a whole gives no resolution.
func load(dir, version string) (*Resolution, []string) {
	root, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return nil, []string{version + ": " + err.Error()}
	}
	res := &Resolution{Calls: map[string][]string{}, Files: map[string]string{}}
	var mu sync.Mutex
	record := func(file string, text []byte) {
		if rel, ok := within(root, file); ok {
			mu.Lock()
			res.Files[rel] = string(text)
			mu.Unlock()
		}
	}
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes |
			packages.NeedTypesInfo | packages.NeedImports,
		Dir:        root,
		Env:        append(os.Environ(), "GOWORK=off"),
		BuildFlags: []string{"-mod=readonly"},
		Tests:      true,
		ParseFile: func(fset *token.FileSet, file string, src []byte) (*ast.File, error) {
			record(file, src)
			return parser.ParseFile(fset, file, src, parser.AllErrors|parser.ParseComments)
		},
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, []string{version + ": " + err.Error()}
	}
	var errs []string
	seen := map[string]bool{}
	packages.Visit(pkgs, nil, func(p *packages.Package) {
		checked := false
		for _, e := range p.Errors {
			checked = checked || e.Kind == packages.TypeError || e.Kind == packages.ParseError
		}
		for _, e := range p.Errors {
			// The go command's compiler output repeats the type checker's errors.
			if checked && e.Kind == packages.ListError && strings.HasPrefix(e.Msg, "# ") {
				continue
			}
			if msg := version + ": " + p.PkgPath + ": " + relative(root, e); !seen[msg] {
				seen[msg] = true
				errs = append(errs, msg)
			}
		}
	})
	// The packages of the module, without their tests, and the methods of their types.
	module := map[string]bool{}
	var methods []*types.Func
	for _, p := range pkgs {
		if p.ID == p.PkgPath && !strings.HasSuffix(p.PkgPath, ".test") && p.Types != nil {
			module[p.PkgPath] = true
			methods = append(methods, concreteMethods(p)...)
		}
	}
	for _, p := range pkgs {
		for _, file := range p.IgnoredFiles {
			text, err := os.ReadFile(file)
			if err != nil {
				errs = append(errs, version+": "+p.PkgPath+": "+err.Error())
				continue
			}
			record(file, text)
		}
		// The reach follows the module without its tests.
		if module[p.PkgPath] && p.ID == p.PkgPath && p.TypesInfo != nil {
			calls(p, module, methods, res.Calls)
		}
	}
	res.Links = links(root, pkgs, res.Files, module, methods)
	return res, errs
}

// calls records, by the key of each function and method of p, the functions and methods of
// the packages in module that it calls or refers to. A call of an interface method stands for
// every one of methods that implements the interface.
func calls(p *packages.Package, module map[string]bool, methods []*types.Func, out map[string][]string) {
	for _, f := range p.Syntax {
		for _, d := range f.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok {
				continue
			}
			fn, ok := p.TypesInfo.Defs[fd.Name].(*types.Func)
			if !ok {
				continue
			}
			callees := map[string]bool{}
			if fd.Body != nil {
				ast.Inspect(fd.Body, func(n ast.Node) bool {
					if id, ok := n.(*ast.Ident); ok {
						if callee, ok := p.TypesInfo.Uses[id].(*types.Func); ok && inModule(callee, module) {
							for _, m := range implementations(callee.Origin(), methods) {
								callees[funcKey(m)] = true
							}
						}
					}
					return true
				})
			}
			out[funcKey(fn)] = sorted(callees)
		}
	}
}

// inModule reports whether fn belongs to one of the packages in module.
func inModule(fn *types.Func, module map[string]bool) bool {
	return fn.Pkg() != nil && module[fn.Pkg().Path()]
}

// concreteMethods returns the methods of the types of p that are not interfaces.
func concreteMethods(p *packages.Package) []*types.Func {
	var methods []*types.Func
	scope := p.Types.Scope()
	for _, name := range scope.Names() {
		tn, ok := scope.Lookup(name).(*types.TypeName)
		if !ok {
			continue
		}
		if named, ok := types.Unalias(tn.Type()).(*types.Named); ok && !types.IsInterface(named) {
			for i := range named.NumMethods() {
				methods = append(methods, named.Method(i))
			}
		}
	}
	return methods
}

// implementations returns fn itself, or for a method of an interface, the method of every
// type in methods that implements the interface.
func implementations(fn *types.Func, methods []*types.Func) []*types.Func {
	recv := fn.Signature().Recv()
	if recv == nil || !types.IsInterface(recv.Type()) {
		return []*types.Func{fn}
	}
	iface, ok := recv.Type().Underlying().(*types.Interface)
	if !ok {
		return nil
	}
	var out []*types.Func
	for _, m := range methods {
		t := m.Signature().Recv().Type()
		if m.Name() == fn.Name() && (types.Implements(t, iface) || types.Implements(types.NewPointer(t), iface)) {
			out = append(out, m)
		}
	}
	return out
}

// funcKey returns the key of a function or a method, such as analyze.sumTypes or
// render.renderer.text.
func funcKey(fn *types.Func) string {
	if recv := fn.Signature().Recv(); recv != nil {
		if name := typeName(recv.Type()); name != "" {
			return fn.Pkg().Name() + "." + name + "." + fn.Name()
		}
	}
	return fn.Pkg().Name() + "." + fn.Name()
}

// sorted returns the keys of set in order, nil for an empty set.
func sorted(set map[string]bool) []string {
	var keys []string
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// relative returns the text of e with its file relative to root.
func relative(root string, e packages.Error) string {
	if e.Pos == "" || e.Pos == "-" {
		return e.Msg
	}
	pos := e.Pos
	if rel, ok := within(root, pos); ok {
		pos = rel
	}
	return pos + ": " + e.Msg
}

// within returns file relative to root, with slashes, when root holds it.
func within(root, file string) (string, bool) {
	rel, err := filepath.Rel(root, file)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return filepath.ToSlash(rel), true
}
