# dfdreview: calls and identifiers resolved by type

dfdreview loads the module of the code directory with its types, in both versions, through go/packages and go/types. The changed mark on a function's link then follows the function's calls into the functions and methods of its own package, and every identifier in the code panel opens its definition.

## Requirements

1. A function's link in a box carries the changed mark when the function is a changed declaration: when it differs between the versions, or when its reach in the working tree holds a function or method that does. The reach stays within the function's own package. A method call resolves by the receiver's type, so a change to another type's method of the same name leaves the caller unmarked. A call of an interface method reaches the method of every type of the package that implements the interface. A type's link keeps its rule: the mark follows the type's own declaration.
2. In the code panel's Before and After views of a Go file, each identifier that uses an object declared in Go code is an identifier link. A predeclared object, such as `int` or `len`, gives no link, and neither does an identifier that declares its object. An object of the code directory opens the same version of its file in the code panel, at its line, and a package of the code directory opens the package clause of its first Go file in path order. An object outside the code directory opens its documentation on pkg.go.dev: `https://pkg.go.dev/<import path>#<name>`, with `<type>.<name>` for a method or a field, and with no anchor for a package name.
3. The page holds every Go file of the code directory in both versions, test files included, so every identifier link into the code directory opens in the code panel.
4. The views with identifier links keep the colors of their Go syntax.
5. The load reads the working tree in place and the base from the base export, which it deletes at the end. It runs the go command with -mod=readonly, so a build changes neither the repository, its go.mod and go.sum included, nor the temporary directory.
6. Without a module, the page builds as before: no identifier links, and the changed mark only on a declaration that differs itself. A package error leaves the build to finish: the package's files still show in the code panel, their identifiers without type information are not links, and the page lists each package error under "Package errors". An error that fails a whole version's load, such as a go.mod that the go command rejects, is a package error of that version, which then has no resolution. The build stops on a load error only when the load cannot run at all, as when git cannot read the base. A Go syntax error in a file of the declaration index stops the build, as it does today. A file outside the build configuration shows without identifier links.
7. The README and review-page.md describe the reach, the identifier links and the package errors.

## Questions and assumptions

- Build configuration: the load runs the go command with its defaults, so GOOS, GOARCH and GOFLAGS, such as `-tags`, apply. A file for another build configuration shows without identifier links, and a `-tags` flag can come later.
- go.work and go.mod: the load sets GOWORK=off, so each version loads its own module alone, since the base export holds no go.work. It also passes -mod=readonly, which overrides a -mod=mod in GOFLAGS, so the go command never writes go.mod or go.sum. A module that needs a go.mod update gets package errors instead.
- Test files: the load includes them, so identifiers in tests are links too. The reach follows the package without its tests.
- Base export: the load writes the module directory's blobs at the base, as `git cat-file` returns them, into a temporary directory. `git archive` would apply `.gitattributes`, such as export-subst and export-ignore, and then the Before texts and their identifier links could disagree. A `replace` directive that points outside the module directory gives the base's packages package errors, and their Before views then show without identifier links, as requirement 6 says.
- Dependencies: go/packages runs the go command, which may download missing modules into the module cache.
- The pkg.go.dev links carry no version, like the existing reference links.
- Design: code.Read loads the module in its own body, and box 3 names code.Load beside code.Read, so the arrows and code.Read's signature stay as they are. Process 3 stays a leaf box: code.Load and its helpers belong to it.
- I/O: the existing design predates the rule that every piece of I/O is drawn. Processes 1, 2 and 5 read git and files, process 4 runs the dfd command and writes a file of footnote definitions, and process 5 writes the page, all without an external system. This plan draws process 3's I/O and leaves the rest to a separate change. dfd's layout overlaps the labels of close arrows, so each external system shows one label: the reply of the module directory, the reply of the git command, and the base export written for the go command.
- Size: the page grows by every Go file of both versions and their identifier links. dfdreview's own page, about 5,000 lines of Go, should stay near 1 MB.
- The work stays on the skills repository's branch hdd-review-page, which holds all of dfdreview.
- The installed dfdreview, at 7e87621, builds this plan's review pages until the last stage, so other sessions keep a working tool.

## The change in brief

code.Read now loads the module with its types in both versions. Each version's resolution holds the calls that every function and method makes in its own package, and the text and the identifier links of every Go file, and the load also reports the package errors. The code index compares the declarations and the methods of both versions, parsed without types, and marks each changed declaration, now with every function whose reach in the working tree holds a changed function or method. The page holds every Go file of both versions with its identifier links, so each identifier in the code panel opens its definition, and the page lists the package errors.

## Metaphor

The review page is an editor's proof packet. Each changed drawing and file comes as old, marked-up and new copies, and every citation in them opens its source.

## Planned declarations

Package `code`:

- `type Resolved struct`: `Before, After *Resolution`, nil for a version without a module, and `Errors []string`, the package errors of both versions.
- `type Resolution struct`, with three fields. `Calls map[string][]string` gives, by the key of each function and method, such as `render.renderer.text`, the keys of the functions and methods of its package that it calls. `Files map[string]string` holds the text of every Go file as the load parsed it, test files included. `Links map[string][]Link` holds the identifier links of each Go file.
- `type Link struct`: the identifier's `Line`, `Col` and `Len`, in UTF-16 code units, and its definition: `File` and `To`, a line of a Go file of the code directory in the same version, or `URL`.
- `func Load(r *repo.Repo) (*Resolved, error)`: Load type-checks the module in both versions. code.Read calls it first. It returns an error only when it cannot run.
- `Index` gains `BeforeLinks, AfterLinks map[string][]Link` and `Errors []string`. `code.Read` keeps its signature.

Package `page`: `type Link struct`, the page's form of a link, and the fields `File.BeforeLinks`, `File.AfterLinks` and `Data.Errors`.

Package `dfdreview`: `writePage` copies the links and the package errors into the page.

The page's script: `linkLine(html, n, version)` wraps the identifiers of line n of the open file in identifier links, `openDefinition(link, version)` opens a link's definition, and `packageErrors()` lists the package errors.

The helpers of code.Load, such as the base export, come with their fills as unexported functions of package code and methods of repo.Repo, and they stay inside box 3.

## Stages

### Stage 1: the load

Goal: code.Load type-checks both versions of the module and records the calls of every function and method within its package, the text of every Go file, and the package errors.
Requirement: 1, 5, 6.
Dependencies: plan approval.
Holes: 1 code.Load.
Acceptance: TestLoadResolvesCalls, TestLoadWithoutModule and TestLoadReportsPackageErrors in package code.
Size: 250 lines.

### Stage 2: identifier links

Goal: code.Load records the identifier links of every Go file in both versions.
Requirement: 2.
Dependencies: stage 1.
Holes: 2 code.Load#2.
Acceptance: TestLoadLinksIdentifiers in package code.
Size: 150 lines.

### Stage 3: the reach and every Go file

Goal: the code index compares the methods of both versions too, marks each function whose reach holds a changed function or method, and holds every Go file of both versions.
Requirement: 1, 3.
Dependencies: stage 1.
Holes: 3 code.Read, 3 code.Read#2.
Acceptance: TestReadMarksByReach, TestReadMarksByReachWithoutBaseModule and TestReadHoldsEveryGoFile in package code.
Size: 120 lines.

### Stage 4: the code panel

Goal: each identifier link in the code panel's Before and After views opens its definition, the syntax colors stay, the changed mark's tooltip covers a change that comes through a call, and the page lists the package errors.
Requirement: 2, 4, 6, 7.
Dependencies: stages 2 and 3.
Holes: 4 linkLine, 4 openDefinition, 4 packageErrors.
Acceptance: the smoke test TestBuildLinksIdentifiers, and a check in a browser on dfdreview's own page against the base of this plan: identifiers are links, each opens its definition, the syntax colors stay, and the console shows no error.
Size: 150 lines.
