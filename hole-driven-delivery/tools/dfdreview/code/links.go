package code

import (
	"go/ast"
	"go/token"
	"go/types"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

// clause is the package clause of a package's first file in path order.
type clause struct {
	file string
	line int
}

// links returns the identifier links of each Go file under root in pkgs, by path. texts holds
// the files' texts, which give the columns in UTF-16 code units. module holds the paths of the
// module's packages, and methods the methods of their types.
func links(root string, pkgs []*packages.Package, texts map[string]string, module map[string]bool, methods []*types.Func) map[string][]Link {
	clauses := firstClauses(root, pkgs)
	out := map[string][]Link{}
	for _, p := range pkgs {
		if p.TypesInfo == nil || p.Types == nil {
			continue
		}
		for _, f := range p.Syntax {
			rel, ok := within(root, p.Fset.PositionFor(f.Pos(), false).Filename)
			if _, done := out[rel]; !ok || done {
				continue
			}
			out[rel] = fileLinks(root, p, f, clauses, module, methods, strings.Split(texts[rel], "\n"))
		}
	}
	return out
}

// firstClauses returns, by import path, the package clause of the first file in path order
// of each package under root, without its tests.
func firstClauses(root string, pkgs []*packages.Package) map[string]clause {
	out := map[string]clause{}
	for _, p := range pkgs {
		if p.ID != p.PkgPath {
			continue
		}
		for _, f := range p.Syntax {
			pos := p.Fset.PositionFor(f.Package, false)
			rel, ok := within(root, pos.Filename)
			if c, seen := out[p.PkgPath]; ok && (!seen || rel < c.file) {
				out[p.PkgPath] = clause{file: rel, line: pos.Line}
			}
		}
	}
	return out
}

// fileLinks returns the identifier links of file f of p, in order. module holds the paths of
// the module's packages, methods the methods of their types, and lines the file's text.
//
// Example: it links strings.TrimSpace and b in p.A, and b in p.c, the key of each link's
// declaration or method.
//
//	package p
//
//	func A(s string) string { return strings.TrimSpace(s) + b }
//
//	var c = "x" + b
func fileLinks(root string, p *packages.Package, f *ast.File, clauses map[string]clause, module map[string]bool, methods []*types.Func, lines []string) []Link {
	owners := fieldOwners(p.TypesInfo, f)
	var out []Link
	visit := func(n ast.Node, in string) {
		ast.Inspect(n, func(n ast.Node) bool {
			id, ok := n.(*ast.Ident)
			if !ok {
				return true
			}
			obj := p.TypesInfo.Uses[id]
			if obj == nil || obj.Pkg() == nil {
				return true
			}
			pos := p.Fset.PositionFor(id.Pos(), false)
			if pos.Line < 1 || pos.Line > len(lines) || pos.Column-1 > len(lines[pos.Line-1]) {
				return true
			}
			link := definition(root, p.Fset, obj, owners[id], clauses)
			link.Line = pos.Line
			link.Col = utf16Len(lines[pos.Line-1][:pos.Column-1]) + 1
			link.Len = utf16Len(id.Name)
			link.In = in
			if link.File != "" {
				link.Key = objectKey(obj)
			}
			if fn, ok := obj.(*types.Func); ok && inModule(fn, module) {
				for _, m := range implementations(fn.Origin(), methods) {
					link.Funcs = append(link.Funcs, funcKey(m))
				}
			}
			out = append(out, link)
			return true
		})
	}
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.FuncDecl:
			in := ""
			if fn, ok := p.TypesInfo.Defs[d.Name].(*types.Func); ok {
				in = funcKey(fn)
			}
			visit(d, in)
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				visit(spec, specKey(f.Name.Name, spec))
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Line != out[j].Line {
			return out[i].Line < out[j].Line
		}
		return out[i].Col < out[j].Col
	})
	return out
}

// specKey returns the key of a top-level spec of package pkg.
// A value spec with several names takes the first; an import has none.
//
// Example: it returns cart.Order for the first spec and cart.a for the second.
//
//	package cart
//
//	type Order struct{ Items []Item }
//
//	var a, b = f(), g()
func specKey(pkg string, spec ast.Spec) string {
	switch s := spec.(type) {
	case *ast.TypeSpec:
		return pkg + "." + s.Name.Name
	case *ast.ValueSpec:
		return pkg + "." + s.Names[0].Name
	}
	return ""
}

// objectKey returns the key of a top-level function, type, variable or constant, or of a
// method of a named type. Any other object, such as a field or a local variable, has none.
func objectKey(obj types.Object) string {
	switch o := obj.(type) {
	case *types.Func:
		if recv := o.Signature().Recv(); recv != nil && typeName(recv.Type()) == "" {
			return "" // a method of an interface without a name
		}
		return funcKey(o.Origin())
	case *types.TypeName, *types.Var, *types.Const:
		if obj.Parent() == obj.Pkg().Scope() {
			return obj.Pkg().Name() + "." + obj.Name()
		}
	}
	return ""
}

// definition returns where obj opens: a line of a file under root, the package clause of its
// package's first file for a package name, or its documentation on pkg.go.dev. owner names
// the struct type of a field.
func definition(root string, fset *token.FileSet, obj types.Object, owner string, clauses map[string]clause) Link {
	if pkg, ok := obj.(*types.PkgName); ok {
		path := pkg.Imported().Path()
		if c, ok := clauses[path]; ok {
			return Link{File: c.file, To: c.line}
		}
		return Link{URL: "https://pkg.go.dev/" + path}
	}
	pos := fset.PositionFor(obj.Pos(), false)
	if rel, ok := within(root, pos.Filename); ok && pos.IsValid() {
		return Link{File: rel, To: pos.Line}
	}
	return Link{URL: "https://pkg.go.dev/" + obj.Pkg().Path() + "#" + anchor(obj, owner)}
}

// anchor returns obj's anchor on its package's page on pkg.go.dev: its name, or for a method
// or a field, the name of its type and its own name.
func anchor(obj types.Object, owner string) string {
	if fn, ok := obj.(*types.Func); ok {
		if recv := fn.Signature().Recv(); recv != nil {
			if name := typeName(recv.Type()); name != "" {
				return name + "." + fn.Name()
			}
		}
	}
	if v, ok := obj.(*types.Var); ok && v.IsField() && owner != "" {
		return owner + "." + v.Name()
	}
	return obj.Name()
}

// fieldOwners returns, for each identifier of f that names a field in a selector or in a
// composite literal's key, the name of the struct type that the field belongs to.
func fieldOwners(info *types.Info, f *ast.File) map[*ast.Ident]string {
	out := map[*ast.Ident]string{}
	ast.Inspect(f, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.SelectorExpr:
			if sel := info.Selections[n]; sel != nil && sel.Kind() == types.FieldVal {
				out[n.Sel] = typeName(sel.Recv())
			}
		case *ast.CompositeLit:
			name := typeName(info.TypeOf(n))
			for _, e := range n.Elts {
				if kv, ok := e.(*ast.KeyValueExpr); ok {
					if key, ok := kv.Key.(*ast.Ident); ok {
						out[key] = name
					}
				}
			}
		}
		return true
	})
	return out
}

// typeName returns the name of t's named type, through a pointer, or "" for another type.
func typeName(t types.Type) string {
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	if named, ok := types.Unalias(t).(*types.Named); ok {
		return named.Obj().Name()
	}
	return ""
}

// utf16Len returns the length of s in UTF-16 code units.
func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		n++
		if r >= 0x10000 {
			n++
		}
	}
	return n
}
