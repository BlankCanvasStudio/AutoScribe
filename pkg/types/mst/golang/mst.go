package golang

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"go/ast"
	"go/build"
	gtypes "go/types"

	"golang.org/x/tools/go/packages"

	log "github.com/sirupsen/logrus"

	"github.com/BlankCanvasStudio/AutoScribe/pkg/types/mst"
)

// Compile-time check
// var _ mst.MST = (*MST)(nil)
type MST struct {
	PackageNodes []mst.PackageNode
	FunctionMap  map[string]mst.FunctionInfo
}

type PackageNode struct {
	*packages.Package
	MST           mst.MST
	Path          string
	FunctionDecls []mst.FunctionDecl
	TypeDefs      []mst.TypeDefinition
	Imports       map[string]string
	CurrentFile   string
}

type FunctionInfo struct {
	Package mst.PackageNode

	Name        string
	ResolvedPkg string
	Object      string
	File        string

	Documentation string

	HasDocumentation     bool // Did we find documentation for it
	DocumentedInThisPass bool // Did we write documentation for it
	IsAiAware            bool

	Declaration mst.FunctionDecl // Where the function is declared & all that jazz
}

type FunctionDecl struct {
	Info  mst.FunctionInfo
	Calls []mst.FunctionCall

	DocInsertLocation uint

	Package mst.PackageNode

	Node *ast.FuncDecl
}

type FunctionCall struct {
	Info mst.FunctionInfo
	Kind mst.FunctionCallKind

	Package mst.PackageNode

	Node *ast.CallExpr
}

type TypeDef struct {
	Name string
}

/*
GetPackages returns the package nodes stored in m as a []mst.PackageNode.
This method returns a direct view into the internal m.PackageNodes slice; it does not
allocate a new slice. Modifications to the elements of the returned slice will affect
the MST, and changes to the slice length/updating the slice header may not be reflected
in the internal state without using dedicated mutators.
Signature: func (m *MST) GetPackages() []mst.PackageNode
Parameters: none
Returns: []mst.PackageNode containing the package nodes; may be nil if uninitialized.
Errors/Exceptions: none
Side Effects: exposes internal state via the returned slice (no defensive copy).
Edge Cases & Assumptions: if m.PackageNodes is nil, GetPackages returns nil; otherwise the
slice views the same underlying array as m.PackageNodes.

*/
func (m *MST) GetPackages() []mst.PackageNode {
	return m.PackageNodes
}

/*
Summary:
Replaces the MST's PackageNodes with the provided pkgs slice. Use this to overwrite the current set of package nodes maintained by m.

Signature:
func (m *MST) SetPackages(pkgs []mst.PackageNode) error

Parameters:
- m: *MST — the receiver; must be non-nil.
- pkgs: []mst.PackageNode — the new package node list to assign to m.PackageNodes.

Returns:
- error: always nil.

Errors/Exceptions:
- Panics if called with a nil receiver (m == nil) since the method dereferences m to assign to m.PackageNodes.

Side Effects:
- Mutates m.PackageNodes by assigning it the value of pkgs.

Edge Cases & Assumptions:
- If pkgs is nil, m.PackageNodes becomes nil.
- No validation of elements in pkgs is performed.
- No concurrency controls are applied.

*/
func (m *MST) SetPackages(pkgs []mst.PackageNode) error {
	m.PackageNodes = pkgs
	return nil
}

/*
Summary: AddPackage appends the given mst.PackageNode to the MST's PackageNodes collection, initializing the slice if needed.
Use when you want to associate a new package node with the MST.
Signature: func (m *MST) AddPackage(n mst.PackageNode) error

Parameters:
- m: receiver of type *MST; mutates the MST instance.
- n: mst.PackageNode; the package node to add.
Returns:
- error: always nil in the current implementation; there are no failure conditions.
Side Effects:
- Mutates m.PackageNodes by appending n.
- Initializes m.PackageNodes if it is nil.
Edge Cases & Assumptions:
- No duplicate detection; duplicates may be added if invoked repeatedly.
- No validation of n beyond being a valid mst.PackageNode value.

*/
func (m *MST) AddPackage(n mst.PackageNode) error {
	if m.PackageNodes == nil {
		m.PackageNodes = make([]mst.PackageNode, 0)
	}

	m.PackageNodes = append(m.PackageNodes, n)

	return nil
}

/*
Summary: Adds or updates an entry in MST.FunctionMap, initializing the map if needed.
         Associates the provided name with the given FunctionInfo value.
Signature: func (m *MST) AddToFunctionMap(name string, info mst.FunctionInfo) error
Parameters:
  - name: string — the key under which info is stored; may overwrite an existing entry.
  - info: mst.FunctionInfo — the value to associate with the given name.
Returns:
  - error: always nil (present for API compatibility; no error conditions are currently defined).
Errors/Exceptions:
  - None produced by this function.
Side Effects:
  - Mutates m.FunctionMap; initializes it if nil, then sets m.FunctionMap[name] = info.
Edge Cases & Assumptions:
  - If m.FunctionMap is nil, it will be initialized to an empty map.
  - If name already exists, its value is overwritten with info.

*/
func (m *MST) AddToFunctionMap(name string, info mst.FunctionInfo) error {
	if m.FunctionMap == nil {
		m.FunctionMap = make(map[string]mst.FunctionInfo)
	}

	m.FunctionMap[name] = info

	return nil
}

/*
Summary:
GetFromFunctionMap ensures m.FunctionMap is initialized and returns the FunctionInfo for the given name from the map.

Signature:
func (m *MST) GetFromFunctionMap(name string) (mst.FunctionInfo, bool, error)

Parameters:
name string
  The function name to look up in m.FunctionMap.

Returns:
val mst.FunctionInfo
  The value associated with the provided name in m.FunctionMap (zero value if not present).
ok bool
  True if the value was found in m.FunctionMap; false otherwise.
err error
  Always nil in this implementation.

Errors/Exceptions:
  nil always.

Side Effects:
  May initialize m.FunctionMap via make(map[string]mst.FunctionInfo) if it is nil.

Edge Cases & Assumptions:
  If FunctionMap already exists, it is used as is. If name is not present, ok is false and val is the zero value of mst.FunctionInfo.

*/
func (m *MST) GetFromFunctionMap(name string) (mst.FunctionInfo, bool, error) {
	if m.FunctionMap == nil {
		m.FunctionMap = make(map[string]mst.FunctionInfo)
	}

	val, ok := m.FunctionMap[name]

	return val, ok, nil
}

/*
Summary
PrettyPrint is a method on *MST that, when implemented, would output a human-readable, pretty-printed representation of the MST. The current implementation is a no-op.

Signature
func (m *MST) PrettyPrint(string)

Parameters
- string: unnamed string parameter; intended as a header/label for the printed output (its use is implementation-defined and may be ignored).

Returns
- none

Errors/Exceptions
- none

Side Effects
- none in the current implementation (no output or state changes).

Edge Cases & Assumptions
- Assumes MST is valid; actual printing behavior is implementation-defined when/if this method is completed.

*/
func (m *MST) PrettyPrint(string) {

}

/*
Summary: Build and populate m.PackageNodes by loading Go packages from the provided folders, creating a PackageNode for each discovered package, and invoking PopulatePackageInformation to derive imports and function declarations.

Signature: func (m *MST) Populate(folders []string) error

Parameters:
- folders: []string; filesystem paths used as package roots to load and process.

Returns:
- error: non-nil if any package information population fails for a processed package.

Errors/Exceptions:
- If pkgNode.PopulatePackageInformation() returns an error, this method returns fmt.Errorf("failed to populate package information: %v", err).
- Errors from packages.Load are not explicitly surfaced by this function (err from Load is assigned but not checked here).

Side Effects:
- Mutates m.PackageNodes to contain all created PackageNode instances.
- Each PackageNode is mutated by PopulatePackageInformation(), updating its Imports and FunctionDecls (and any other populated metadata).
- Logs progress via log.Debugf with the number of functions declared for each package.

Edge Cases & Assumptions:
- Assumes folders contains valid package roots and that packages.Load returns packages for each folder.
- If no packages are found, the function returns nil with m.PackageNodes left empty or partially populated.
- Assumes PackageNode.PopulatePackageInformation handles its own internal processing and error reporting.

*/
func (m *MST) Populate(folders []string) error {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedSyntax |
			packages.NeedCompiledGoFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedImports |
			packages.NeedDeps,
	}

	pkgNodes := []mst.PackageNode{}

	for _, f := range folders {
		pkgs, err := packages.Load(cfg, f)

		for _, pkg := range pkgs {
			pkgNode := &PackageNode{
				MST:           m,
				Package:       pkg,
				FunctionDecls: []mst.FunctionDecl{},
				// TypeDefinitions:      []*ast.TypeSpec{},
				Imports: make(map[string]string),
			}

			err = pkgNode.PopulatePackageInformation()
			if err != nil {
				return fmt.Errorf("failed to populate package information: %v", err)
			}

			log.Debugf("%v functions declared", len(pkgNode.FunctionDecls))

			pkgNodes = append(pkgNodes, pkgNode)
		}
	}

	m.PackageNodes = pkgNodes

	return nil
}

/*
Summary: Clips cyclic dependencies across all packages by clipping function cycles in every function declaration.
This method iterates over the package nodes returned by GetPackages and calls ClipFunctionCycles for each function declaration.
On any error, it returns an error wrapping the underlying error.
"GetPackages returns the package nodes stored in m as a []mst.PackageNode. This method returns a direct view into the internal m.PackageNodes slice; it does not allocate a new slice. Modifications to the elements of the returned slice will affect the MST, and changes to the slice length/updating the slice header may not be reflected in the internal state without using dedicated mutators."
Signature: func (m *MST) HandleCyclicDependencies() error
Parameters: none
Returns: error; nil on success; non-nil on failure.
Errors/Exceptions: returns fmt.Errorf("failed to clip function cycles: %v", err) if ClipFunctionCycles fails.
Side Effects: May mutate internal function call graphs via ClipFunctionCycles; relies on GetPackages returning a direct view into the internal slice.
Edge Cases & Assumptions: If GetPackages returns nil, the outer loop is a no-op. Each package may have zero or more function declarations; in that case, no clipping occurs for that package.

*/
func (m *MST) HandleCyclicDependencies() error {
	for _, p := range m.GetPackages() {
		for _, decl := range p.GetFunctionDecls() {
			callStack := []string{}
			err := p.ClipFunctionCycles(decl.GetInfo(), callStack)
			if err != nil {
				return fmt.Errorf("failed to clip function cycles: %v", err)
			}
		}
	}

	return nil
}

/*
FindLineNo returns the absolute byte offset of the start of the line that contains the given AST node.
It uses p.Fset to map n.Pos() to the corresponding file and locate the line's start.
If n is nil, it returns -1.

Signature: func (p *PackageNode) FindLineNo(n ast.Node) int

Parameters:
  n: ast.Node - the AST node whose starting line offset is sought. If n is nil, the result is -1.
  Constraints: n.Pos() must be a position known to p.Fset (i.e., within the same FileSet).

Returns:
  int - the absolute offset (in bytes) of the start of the line containing n.Pos(); -1 if n is nil.

Errors/Exceptions:
  No explicit errors are returned. A nil n is handled. If n.Pos() is not registered in p.Fset,
  or p.Fset.File(n.Pos()) returns nil, this function may panic at runtime.

Side Effects:
  None.

Edge Cases & Assumptions:
  Assumes n is a non-nil ast.Node with a valid Pos() registered in p.Fset. LineStart uses startPos.Line.

*/
func (p *PackageNode) FindLineNo(n ast.Node) int {
	if n == nil {
		return -1
	}
	startPos := p.Fset.Position(n.Pos())
	// get the File for this position
	file := p.Fset.File(n.Pos())

	// find offset of the start of the line
	lineStart := file.LineStart(startPos.Line)

	// convert to absolute offset
	lineOffset := p.Fset.Position(lineStart).Offset

	return lineOffset
}

/*
Summary: Prunes cycles in a function's call graph by removing calls that would re-enter a function already on the current call stack. It traverses non-cyclic calls recursively and removes cycle-inducing entries from the function's declaration.

Signature: func (p *PackageNode) ClipFunctionCycles(f mst.FunctionInfo, callStack []string) error

Parameters:
  f mst.FunctionInfo - the function to process; its decl is examined for calls.
  callStack []string - the sequence of fully-qualified function names on the current call path; used to detect cycles.

Returns:
  error - always nil (no error conditions produced by this function).

Errors/Exceptions:
  None. If decl == nil, the function is a no-op and returns nil.

Side Effects:
  Mutates f.GetDecl().SetCalls(...) to remove cycle-inducing calls.
  Performs recursive descent on non-cycle calls via p.ClipFunctionCycles(call.GetInfo(), append(callStack, f.GetFullName())).

Edge Cases & Assumptions:
  If decl == nil, the function returns nil without changes.
  Assumes f.GetDecl().GetCalls() returns a slice of calls; each call provides GetFullName() and GetInfo().
  Removals from the calls slice are performed in reverse index order to preserve correct indices.

*/
func (p *PackageNode) ClipFunctionCycles(f mst.FunctionInfo, callStack []string) error {
	to_remove := []int{}
	decl := f.GetDecl()
	if decl == nil {
		return nil
	}

	for i, call := range decl.GetCalls() {
		// Remove the repeated node from the calls array, but don't descend it.
		// If you descend it, all the nodes above it will be removed as well (since they've
		//   already been included in the list)
		if slices.Contains(callStack, call.GetFullName()) {
			to_remove = append(to_remove, i)
		} else {
			p.ClipFunctionCycles(call.GetInfo(), append(callStack, f.GetFullName()))
		}
	}

	for i := len(to_remove) - 1; i >= 0; i-- {
		decl := f.GetDecl()

		calls := decl.GetCalls()

		decl.SetCalls(append(calls[:to_remove[i]], calls[to_remove[i]+1:]...))

	}

	return nil
}

/*

*/
func (p *PackageNode) GetMST() mst.MST {
	return p.MST
}

/*

*/
func (p *FunctionDecl) SetDocumentation(docs string) error {
	p.Info.(*FunctionInfo).Documentation = docs
	return nil
}

/*
Summary: Sets the Documentation field of the underlying FunctionInfo for this FunctionCall by assigning the provided docs string.

Signature: func (p *FunctionCall) SetDocumentation(docs string) error

Parameters:
- docs: string — the documentation text to set on p.Info.(*FunctionInfo).Documentation.

Returns:
- error — always nil.

Errors/Exceptions:
- Panics if p.Info is nil or not a *FunctionInfo (due to the type assertion p.Info.(*FunctionInfo)).

Side Effects:
- Mutates p.Info.(*FunctionInfo).Documentation.

Edge Cases & Assumptions:
- p.Info must be non-nil and of type *FunctionInfo before calling SetDocumentation.
- docs may be empty; the field will be updated accordingly.

*/
func (p *FunctionCall) SetDocumentation(docs string) error {
	p.Info.(*FunctionInfo).Documentation = docs
	return nil
}

/*
Summary: GetDocumentation returns the documentation string for this FunctionDecl by delegating to p.Info.GetDocumentation(). Use this when you need the underlying Info documentation for the FunctionDecl.

Signature: func (p *FunctionDecl) GetDocumentation() (string, error)

Parameters: none

Returns:
  - string: the documentation text.
  - error: any error returned by p.Info.GetDocumentation().

Errors/Exceptions:
  - Propagates error from p.Info.GetDocumentation().
  - May panic if p == nil or p.Info == nil (nil checks are not performed here).

Side Effects: none

Edge Cases & Assumptions:
  - Assumes p != nil and p.Info != nil.
  - This method is a thin wrapper that delegates to p.Info.GetDocumentation() for its result.

*/
func (p *FunctionDecl) GetDocumentation() (string, error) {
	return p.Info.GetDocumentation()
}

/*
Summary
GetDocumentation returns the documentation for this FunctionCall by delegating to p.Info.GetDocumentation() and returning its results.

Signature
func (p *FunctionCall) GetDocumentation() (string, error)

Parameters
- none: method has a receiver *FunctionCall; no input parameters.

Returns
- (string, error): the documentation text and any error from p.Info.GetDocumentation().

Errors/Exceptions
- Propagates any error returned by p.Info.GetDocumentation().

Side Effects
- None beyond calling p.Info.GetDocumentation().

Edge Cases & Assumptions
- Assumes p.Info is non-nil and implements GetDocumentation() (string, error). If p.Info is nil, this will panic at runtime when invoking p.Info.GetDocumentation().

*/
func (p *FunctionCall) GetDocumentation() (string, error) {
	return p.Info.GetDocumentation()
}

/*
Summary: GetFunctionDecls returns the FunctionDecls stored on the PackageNode.

When to use: Retrieve the list of function declarations associated with a PackageNode.

Signature:
func (p *PackageNode) GetFunctionDecls() []mst.FunctionDecl

Parameters:
p - *PackageNode: the receiver containing FunctionDecls.

Returns:
[]mst.FunctionDecl: the slice of FunctionDecls contained in p. May be nil or empty if none.

Errors/Exceptions:
If p is nil, calling this method will panic when accessing p.FunctionDecls.

Side Effects:
None.

Edge Cases & Assumptions:
Assumes non-nil receiver when called; returns a direct reference to the underlying slice (not a copy).

*/
func (p *PackageNode) GetFunctionDecls() []mst.FunctionDecl {
	return p.FunctionDecls
}

/*
SetFunctionDecls replaces the PackageNode.FunctionDecls with the provided decls.

Use this to update the list of function declarations associated with a PackageNode.

Signature: func (p *PackageNode) SetFunctionDecls(decls []mst.FunctionDecl) error

Parameters:
- decls: []mst.FunctionDecl — the new function declarations to assign to p.FunctionDecls.

Returns:
- error: always nil in the current implementation.

Side Effects:
- Mutates the receiver p by assigning FunctionDecls = decls.

Edge Cases & Assumptions:
- decls may be nil or empty; the field is set to the provided value as-is.
- No validation is performed on the contents of decls.

*/
func (p *PackageNode) SetFunctionDecls(decls []mst.FunctionDecl) error {
	p.FunctionDecls = decls
	return nil
}

/*
Summary: GetImports returns the Imports map stored on the PackageNode.
Use when you need access to the mapping of import aliases to import paths for this node.
Signature: func (p *PackageNode) GetImports() map[string]string
Parameters: none
Returns: map[string]string - the Imports map from the PackageNode. This is the internal map; callers may mutate it and such mutations affect the node.
Errors/Exceptions: none
Side Effects: none
Edge Cases & Assumptions:
- The function returns the actual internal p.Imports map (not a copy).
- p is a valid, initialized *PackageNode when called; no nil handling is performed.

*/
func (p *PackageNode) GetImports() map[string]string {
	return p.Imports
}

/*
Summary: Adds or updates an import alias mapping on PackageNode by associating the short
identifier with its fully-qualified name (fqn). Initializes the Imports map if needed.
Use to accumulate import mappings for a package during generation or transformation.
Signature: func (p *PackageNode) AddToImports(short string, fqn string) error
Parameters:
- short: string, the local/import alias used in code.
- fqn: string, the fully-qualified import path corresponding to the alias.
Returns:
- error: always nil in current implementation.
Errors/Exceptions:
- None (function always returns nil).
Side Effects:
- If p.Imports is nil, initializes it as make(map[string]string).
- Sets p.Imports[short] = fqn, mutating the PackageNode.
- Overwrites any existing value for the same short key.
Edge Cases & Assumptions:
- If called concurrently, there is no synchronization; concurrent writes may race.
- If the same short is added multiple times, the last fqn wins.

*/
func (p *PackageNode) AddToImports(short string, fqn string) error {
	if p.Imports == nil {
		p.Imports = make(map[string]string)
	}

	p.Imports[short] = fqn

	return nil
}

/*
Summary: GetTypeDefs returns the TypeDefs associated with this PackageNode.
Use GetTypeDefs to access the underlying []mst.TypeDefinition stored in p.TypeDefs.
Signature: func (p *PackageNode) GetTypeDefs() []mst.TypeDefinition
Returns: []mst.TypeDefinition containing the TypeDefs for this package node.
Side Effects: None; this is a read-only accessor that does not modify the receiver.
Edge Cases & Assumptions: p must be non-nil; calling on a nil *PackageNode will panic when accessing p.TypeDefs.

*/
func (p *PackageNode) GetTypeDefs() []mst.TypeDefinition {
	return p.TypeDefs
}

/*
Summary: Sets p.TypeDefs to the provided in slice, replacing any existing value.
Use this to update the type definitions associated with a PackageNode.

Signature: func (p *PackageNode) SetTypeDefs(in []mst.TypeDefinition) error

Parameters:
  in: []mst.TypeDefinition - new type definitions to assign to p.TypeDefs.
Returns:
  error - always nil. The function does not perform validation and always succeeds.
Errors/Exceptions:
  None
Side Effects:
  Mutates p.TypeDefs by replacing it with the provided slice.
Edge Cases & Assumptions:
  If in is nil, p.TypeDefs is set to nil; no validation is performed.

*/
func (p *PackageNode) SetTypeDefs(in []mst.TypeDefinition) error {
	p.TypeDefs = in
	return nil
}

/*
GetResolvedPackageName returns the resolved package name for this PackageNode.
It does so by returning the p.ID field of the receiver.

Signature: func (p *PackageNode) GetResolvedPackageName() string
Parameters: none
Returns: string - the resolved package name (p.ID)
Errors/Exceptions: none
Side Effects: none
Edge Cases & Assumptions: calling on a nil *PackageNode will panic; ensure the receiver is non-nil when invoking.

*/
func (p *PackageNode) GetResolvedPackageName() string {
	return p.ID
}

/*
Summary: Placeholder setter for the resolved package name on PackageNode. This implementation performs no operation and ignores the name argument.
When to use: Use when an API requires a SetResolvedPackageName method, but the value is not needed yet or is managed elsewhere.
Signature: func (p *PackageNode) SetResolvedPackageName(name string)
Parameters:
  name string: the proposed resolved package name; currently unused.
Side Effects: none
Edge Cases & Assumptions:
  - name is ignored; no validation performed.
  - This may be implemented in the future to mutate p or update internal state.

*/
func (p *PackageNode) SetResolvedPackageName(name string) {
	_ = name
	return
	// return p.PkgPath
}

/*
Summary: GetPath returns p.PkgPath, the package path associated with this PackageNode.
Signature: func (p *PackageNode) GetPath() string
Returns: string - the value of p.PkgPath.
Edge Cases & Assumptions: If PkgPath is empty, this returns "".

*/
func (p *PackageNode) GetPath() string {
	return p.PkgPath
}

/*
Summary: Disallows setting the path on PackageNode; this method always returns an error indicating that path assignment is unsupported for golang packages.
Use when you need to guard against attempting to set a path on golang packages.
Signature: func (p *PackageNode) SetPath(s string) error
Parameters:
  p: *PackageNode; role: receiver; constraints: none
  s: string; role: proposed path; constraints: any string (not used in operation)
Returns: error; non-nil always; the error message is generated as "do not set path %v on golang packages" with s substituted
Errors/Exceptions: Always returns an error; never succeeds
Side Effects: none; no mutation occurs
Edge Cases & Assumptions: This is a hard fail for all inputs; caller must handle the error accordingly

*/
func (p *PackageNode) SetPath(s string) error {
	return fmt.Errorf("do not set path %v on golang packages", s)
}

/*
Summary
SetCurrentFile prevents setting the current file path on PackageNode instances. It always returns an error to enforce that Go package nodes do not have their current file path set.

Signature
func (p *PackageNode) SetCurrentFile(file string) error

Parameters
- file string
  Path to the file to set (ignored).

Returns
- error
  A non-nil error with the exact message: "don't set path on go packages; I got it"

Errors/Exceptions
Always returns an error; no mutation occurs.

Side Effects
No state changes; only returns an error.

Edge Cases & Assumptions
The receiver is a *PackageNode. The method does not inspect or use p or file.
This function exists to guard against invalid attempts to assign a current file to Go packages.

*/
func (p *PackageNode) SetCurrentFile(file string) error {
	return fmt.Errorf("don't set path on go packages; I got it")
}

/*
Summary: PrettyPrint is a placeholder method on *PackageNode that accepts a string parameter and currently has no implementation.
Use this as a hook for future pretty-printing logic of PackageNode when implemented.

Signature: func (p *PackageNode) PrettyPrint(string)

Parameters:
- (unnamed): string — input that may influence formatting/output when the function is implemented.

Returns:
- None (no return values).

Errors/Exceptions:
- None.

Side Effects:
- None in current implementation (no operation).

Edge Cases & Assumptions:
- This is a no-op until an implementation is added. The exact purpose of the string parameter and the output format are unspecified in the current code.

*/
func (p *PackageNode) PrettyPrint(string) {

}

/*
Summary:
PopulatePackageInformation processes each AST file in p.Syntax to derive and store
package-level metadata by delegating to AddToImportMap, AddToTypeDefinitions, and
AddToFunctionDeclarations. It updates internal structures (imports, type defs,
and function declarations) to enable further analysis.

Signature:
func (p *PackageNode) PopulatePackageInformation() error

Parameters:
- none (receiver is *PackageNode; all data is read from and written to by the method)

Returns:
- error: non-nil if any step fails while populating imports, type definitions, or function declarations.

Errors/Exceptions:
- If AddToImportMap(syn_ast) fails, returns fmt.Errorf("failed to add to import map: %v", err)
- If AddToTypeDefinitions(syn_ast) fails, returns fmt.Errorf("failed to expand type definitions: %v", err)
- If AddToFunctionDeclarations(syn_ast) fails, returns fmt.Errorf("failed to expand function definitions: %v", err)

Side Effects:
- Mutates p.CurrentFile for each processed file.
- Mutates p.Imports, p.TypeDefs, and p.FunctionDecls based on the parsed ASTs.
- Logs progress via log.Debugf.

Edge Cases & Assumptions:
- Assumes p.Syntax and p.CompiledGoFiles are aligned by index; for each i in range of p.Syntax,
  p.CurrentFile is set to p.CompiledGoFiles[i].
- If no syntax files are present, the function returns nil after performing no iterations.
- If any of the AddTo* helper methods return an error, processing stops and the error is propagated.

*/
func (p *PackageNode) PopulatePackageInformation() error {
	for i, syn_ast := range p.Syntax {
		p.CurrentFile = p.CompiledGoFiles[i]
		log.Debugf("Stripping ASTs from %v: ", p.CurrentFile)

		err := p.AddToImportMap(syn_ast)
		if err != nil {
			return fmt.Errorf("failed to add to import map: %v", err)
		}

		err = p.AddToTypeDefinitions(syn_ast)
		if err != nil {
			return fmt.Errorf("failed to expand type definitions: %v", err)
		}

		err = p.AddToFunctionDeclarations(syn_ast)
		if err != nil {
			return fmt.Errorf("failed to expand function definitions: %v", err)
		}

	}

	for _, decl := range p.FunctionDecls {
		log.Debugf("defined function: %v", decl.GetName())
	}

	return nil
}

/*
Summary:
Populate p.Imports from the imports declared in f_ast. If p.Imports is nil, initialize it. For aliased imports, store the alias name as the key and the import path as the value. For non-aliased imports, determine the default package name via the build system and map that name to the import path.

Signature:
func (p *PackageNode) AddToImportMap(f_ast *ast.File) error

Parameters:
- p: *PackageNode
  Role: receiver; mutated to store mapping of import names to paths.
  Constraints: must be a valid pointer to PackageNode.
- f_ast: *ast.File
  Role: input AST containing import declarations to process.
  Constraints: non-nil; expected to contain f_ast.Imports.

Returns:
- error
  Description: non-nil if a non-aliased import name cannot be resolved via the build system or other failures occur during import processing.

Errors/Exceptions:
- If an import is not aliased and build.Import(path, "", build.ImportComment) fails, returns an error wrapping the underlying cause:
  "failed to build imports: %v"

- On success, returns nil.

Side Effects:
- Mutates p.Imports by inserting or updating entries.

Edge Cases & Assumptions:
- If p.Imports == nil, it is initialized to an empty map.
- Aliased imports: uses imp.Name.Name as the key and the trimmed path as the value.
  Comment in code: "Set name if they alias the package" and "If they alias the package"
- Non-aliased imports: derives the default package name from the build system and uses it as the key, with the trimmed path as the value.
  Comment in code: "Get default import name from build system if they don't alias"
- Paths are trimmed of surrounding quotes; for aliased imports the path is trimmed with "\"", and for non-aliased imports similarly.
- If build.Import reports an error for a non-aliased import, processing stops and the error is returned.

*/
func (p *PackageNode) AddToImportMap(f_ast *ast.File) error {
	if p.Imports == nil {
		p.Imports = map[string]string{}
	}

	for _, imp := range f_ast.Imports {
		// Set name if they alias the package
		if imp.Name != nil { // If they alias the package
			p.Imports[imp.Name.Name] = strings.Trim(imp.Path.Value, "\"")
			continue
		}

		// Get default import name from build system if they don't alias
		path := strings.Trim(imp.Path.Value, `"`)

		pkg, err := build.Import(path, "", build.ImportComment)
		if err != nil {
			return fmt.Errorf("failed to build imports: %v", err)
		}

		p.Imports[pkg.Name] = path
	}

	return nil
}

/*
Summary: Collects all named type declarations (TypeSpec) from the given AST node and stores them as TypeDef entries in p.TypeDefs. It traverses the subtree rooted at f and appends a TypeDef for each TypeSpec found.

Signature: func (p *PackageNode) AddToTypeDefinitions(f ast.Node) error

Parameters:
- f: ast.Node to inspect. The subtree rooted at f is traversed to locate *ast.TypeSpec nodes.

Returns:
- error: always nil in the current implementation.

Errors/Exceptions: none (function always returns nil).

Side Effects:
- Mutates p.TypeDefs by appending a new TypeDef for each *ast.TypeSpec found: TypeDef{Name: fd.Name.Name}.

Edge Cases & Assumptions:
- Assumes fd.Name.Name is non-empty for each found TypeSpec.
- All TypeSpec nodes found under f are added, including nested ones.

*/
func (p *PackageNode) AddToTypeDefinitions(f ast.Node) error {
	ast.Inspect(f, func(n ast.Node) bool {
		fd, ok := n.(*ast.TypeSpec)
		if ok {
			td := &TypeDef{Name: fd.Name.Name}
			p.TypeDefs = append(p.TypeDefs, td)
		}

		return true
	})

	return nil
}

/*
Summary: Collects all function declarations from the provided AST file, resolves the function invocations within each function, converts each declaration into an mst.FunctionDecl with its associated Calls, and appends them to p.FunctionDecls for later analysis.

Signature: func (p *PackageNode) AddToFunctionDeclarations(f *ast.File) error

Parameters:
- f: *ast.File - the AST of the Go source file to scan for function declarations. The function inspects f to extract *ast.FuncDecl entries. If f contains no functions, this yields no work.

Returns:
- error: non-nil if obtaining function invocations for any declaration fails (wrapped with context); nil on success.

Errors/Exceptions:
- Returns an error if p.GetFunctionInvocations(decl) fails for any function declaration, with context: "failed to get function invocations: %v".

Side Effects:
- Mutates p.FunctionDecls by appending new mst.FunctionDecl entries created from the AST declarations.
- May interact with the MST through p.CreateFunctionDecl and p.CreateFunctionCall via underlying state (FunctionDecl creation and caching of FunctionInfo).

Edge Cases & Assumptions:
- If f contains no function declarations, the function performs no changes and returns nil.
- For each decl, function invocations are retrieved by p.GetFunctionInvocations; if an error occurs, processing stops and is propagated.
- For each invocation, a corresponding mst.FunctionCall is created with p.CreateFunctionCall; if it returns nil (e.g., non-call expressions such as type casts), that invocation is skipped.
- The function initializes p.FunctionDecls if nil and relies on p.Syntax to size intermediate slices.

*/
func (p *PackageNode) AddToFunctionDeclarations(f *ast.File) error {
	if p.FunctionDecls == nil {
		p.FunctionDecls = make([]mst.FunctionDecl, 0, len(p.Syntax))
	}

	// Get all the function Declarations from the AST
	decl_funcs := make([]*ast.FuncDecl, 0, len(p.Syntax))

	ast.Inspect(f, func(n ast.Node) bool {
		fd, ok := n.(*ast.FuncDecl)
		if ok {
			decl_funcs = append(decl_funcs, fd)
		}

		return true
	})

	for _, decl := range decl_funcs {
		// Get all the function invocations called in this function
		invocations, err := p.GetFunctionInvocations(decl)
		if err != nil {
			return fmt.Errorf("failed to get function invocations: %v", err)
		}

		Calls := make([]mst.FunctionCall, 0, len(invocations)) // Not necessarily full length. Typecasts

		// Iterate over all the function infos
		for _, invoc := range invocations {
			// Create a new node from the call (proper finfo set up in the call)
			newNode := p.CreateFunctionCall(invoc)
			if newNode == nil { // Technically a typecast
				continue
			}

			Calls = append(Calls, newNode)
		}

		// Create a new function node for the newly declared function
		newFuncNode := p.CreateFunctionDecl(decl)

		newFuncNode.SetCalls(Calls)

		// Save our newly declared function to the package object
		p.FunctionDecls = append(p.FunctionDecls, newFuncNode)
	}

	return nil
}

/*
Summary: Returns all function invocation expressions (ast.CallExpr) found within the provided AST node f.

Signature: func (p *PackageNode) GetFunctionInvocations(f ast.Node) ([]*ast.CallExpr, error)

Parameters:
- f: ast.Node. Root of the subtree to search for function invocations.

Returns:
- ([]*ast.CallExpr, error): slice of all found *ast.CallExpr nodes. Error is always nil with current implementation.

Errors/Exceptions:
- None produced by this function.

Side Effects:
- None. Reads the AST and returns results without mutation.

Edge Cases & Assumptions:
- If no function invocations exist in f, returns an empty slice.
- If f is nil, the search yields an empty result.

*/
func (p *PackageNode) GetFunctionInvocations(f ast.Node) ([]*ast.CallExpr, error) {
	funcs := make([]*ast.CallExpr, 0, 10)

	ast.Inspect(f, func(n ast.Node) bool {
		fd, ok := n.(*ast.CallExpr)
		if ok {
			funcs = append(funcs, fd)
		}

		return true
	})

	return funcs, nil
}

/*
Summary: Builds an mst.FunctionCall from an ast.CallExpr within a PackageNode, populating FunctionInfo and caching it for reuse. Use to convert a Go AST call into the internal function-call representation and to resolve the function’s name, object/package context, and associated package.
Signature: func (p *PackageNode) CreateFunctionCall(fun *ast.CallExpr) mst.FunctionCall
Parameters:
  p *PackageNode: the package node context (receiver); provides type information, imports, and the function-name cache.
  fun *ast.CallExpr: the AST node representing the function call to translate.
Returns:
  mst.FunctionCall: the constructed function-call representation, or nil for certain non-call expressions (e.g., type casts, paren expressions, or function literals).
Errors/Exceptions:
  May log a fatal error if fun.Fun has an unrecognized type (default case in the type switch).
Side Effects:
  Reads and updates p.MST via GetFromFunctionMap / AddToFunctionMap; reads p.TypesInfo, p.Imports, and p.ID; does not mutate the input AST.
Edge Cases & Assumptions:
  - For *ast.Ident, creates an InternalCall with the function name and the current package as the resolved package.
  - For *ast.SelectorExpr, distinguishes PackageCall vs ObjectCall by inspecting p.TypesInfo.Uses; may resolve to a package path via Imports or to an associated object, including object name and package path if available.
  - For *ast.ArrayType, *ast.ParenExpr, and *ast.FuncLit, returns nil (not a resolvable FunctionCall).
  - If an existing FunctionInfo with the same full name exists in MST, reuses it; otherwise adds the new FunctionInfo to MST before returning.

*/
func (p *PackageNode) CreateFunctionCall(fun *ast.CallExpr) mst.FunctionCall {
	var fInfo mst.FunctionInfo

	// Make FCall and FInfo
	// So this is wrong, but it doesn't matter cause we update it later
	fInfo = &FunctionInfo{Package: p}
	fCall := &FunctionCall{Node: fun, Package: p}
	// fCall := &FunctionCall{Node: fun}

	// Populate FInfo and FCall so we can look things up
	switch fd := fun.Fun.(type) {
	case *ast.Ident:
		fCall.Kind = mst.InternalCall

		fInfo.SetName(fd.Name)
		fInfo.SetResolvedPkg(p.ID)

		// fInfo.Package = p

	case *ast.SelectorExpr:
		sel, _ := fd.X.(*ast.Ident)
		obj := p.TypesInfo.Uses[sel]

		if _, isPkg := obj.(*gtypes.PkgName); isPkg { // Then its a package call
			fCall.Kind = mst.PackageCall

			pkg_name := gtypes.ExprString(fd.X)

			fInfo.SetName(fd.Sel.Name)
			fInfo.SetResolvedPkg(p.Imports[pkg_name])

		} else { // Its an object call

			fCall.Kind = mst.ObjectCall

			selInfo := p.TypesInfo.Selections[fd]
			obj := selInfo.Obj()

			fInfo.SetName(obj.Name())

			recv := selInfo.Recv()
			if p, ok := recv.(*gtypes.Pointer); ok {
				recv = p.Elem()
			}

			if n, ok := recv.(*gtypes.Named); ok {
				fInfo.SetObjectName(n.Obj().Name())
				if n.Obj().Pkg() != nil {
					fInfo.SetResolvedPkg(n.Obj().Pkg().Path())
				}
			} else {
				fInfo.SetObjectName(gtypes.TypeString(recv, func(*gtypes.Package) string { return "" }))
			}

			if fInfo.GetResolvedPkg() == "" && obj.Pkg() != nil {
				fInfo.SetResolvedPkg(obj.Pkg().Path())
			}

		}

	case *ast.ArrayType: // Type cast
		return nil

	case *ast.ParenExpr: // I guess this is fine. Should just be math; we'd catch the rest later in the tree
		return nil

        case *ast.FuncLit:
            // Not sure this is the correct answer
            return nil

	default:
		log.Fatalf("Failed to switch types on function %T: %v", fun.Fun, fun.Fun)
	}

	// Check if FInfo already exists. If it does, swap them out and use existing pointer
	// Would do this first, but getting the object name takes so long, that I'd rather not do
	//   it twice
	possibleInfo, exists, _ := p.MST.GetFromFunctionMap(fInfo.GetFullName())
	if exists {
		fInfo = possibleInfo
	} else {
		p.MST.AddToFunctionMap(fInfo.GetFullName(), fInfo)
	}

	fCall.Info = fInfo

	return fCall
}

/*
Summary: Creates a new mst.FunctionDecl for the given *ast.FuncDecl. It determines the receiver's named type when f is a member function (via IsMemberFunction), initializes a FunctionInfo (reusing an existing one from the MST if present), and returns a FunctionDecl linked to that info and the AST node. It also registers the FunctionInfo in the MST under its full name and sets the FunctionInfo.File to the current file.

Signature: func (p *PackageNode) CreateFunctionDecl(f *ast.FuncDecl) mst.FunctionDecl

Parameters:
  f *ast.FuncDecl - the function declaration to convert; must be non-nil. If f is nil, the call will panic.

Returns:
  mst.FunctionDecl - the created function declaration object.

Errors/Exceptions:
  None returned. May panic on nil f or on other assumed non-nil fields of f (e.g., f.Doc).

Side Effects:
  - May modify the MST function map via GetFromFunctionMap and AddToFunctionMap.
  - May overwrite fInfo with a previously stored FunctionInfo from the MST.
  - Establishes the association between fInfo and the FunctionDecl (fDecl).
  - Sets fInfo.File to p.CurrentFile via fInfo.SetFile.

Edge Cases & Assumptions:
  - If f is a member function, the receiver’s named type is used to populate the Object field for the function’s full name.
  - f.Doc.Text() is used for Documentation; assumes f.Doc is non-nil.
  - Requires p.GetMST() and p.CurrentFile to be valid; relies on p.IsMemberFunction(f) for receiver-type resolution.

*/
func (p *PackageNode) CreateFunctionDecl(f *ast.FuncDecl) mst.FunctionDecl {
	obj := ""

	typeName, found := p.IsMemberFunction(f)
	if found {
		obj = typeName.Obj().Id()
	}

	// Create a POSSIBLE object
	fInfo := &FunctionInfo{
		Package: p,

		Name:        f.Name.String(),
		ResolvedPkg: p.ID,
		Object:      obj,
		File:        p.CurrentFile,

		Documentation: f.Doc.Text(),

		HasDocumentation:     f.Doc.Text() != "",
		DocumentedInThisPass: false,
		IsAiAware:            false,
	}

	// Check if we have a function definition set up in the map.
	//   This would happen if call was experienced before we saw definition
	if possibleInfo, exists, _ := p.GetMST().GetFromFunctionMap(fInfo.GetFullName()); exists {
		info := (possibleInfo).(*FunctionInfo)
		docs := f.Doc.Text()
		fInfo = info
		fInfo.Documentation = docs
		fInfo.File = p.CurrentFile
		fInfo.HasDocumentation = docs != ""
		fInfo.Package = p
	}

	fDecl := &FunctionDecl{
		Info:    fInfo,
		Node:    f,
		Package: p,

		Calls: []mst.FunctionCall{},
	}

	fInfo.Declaration = fDecl

	p.GetMST().AddToFunctionMap(fInfo.GetFullName(), fInfo)

	fInfo.SetFile(p.CurrentFile)

	return fDecl
}

/*
Summary: Determines if fd is a member function declaration (has a receiver) and,
if so, returns the receiver's named type as *gtypes.Named along with true.
It first uses go/types via p.TypesInfo.Defs[fd.Name], and if that yields a
receiver type, returns it when it is a *gtypes.Named. If not, it falls back to
syntactic analysis of the receiver expression to resolve a named receiver type.
Intended for use when you need the receiver's named type for a method on a type.

Signature: func (p *PackageNode) IsMemberFunction(fd *ast.FuncDecl) (*gtypes.Named, bool)

Parameters:
  fd *ast.FuncDecl - the function declaration to analyze. Must have a receiver to be considered a member function.
    If fd is nil or has no receiver, the function returns (nil, false).

Returns:
  n *gtypes.Named - the receiver's named type when fd is a member function; nil otherwise.
  ok bool - true if the function is a member function and the receiver resolves to a named type; otherwise false.

Errors/Exceptions: none returned; function returns (nil, false) on failure to resolve a named receiver.

Side Effects: none.

Edge Cases & Assumptions:
  - Pointer receivers are unwrapped to their element type before checking for a named type.
  - The function is considered a member function only if the receiver type resolves to a *gtypes.Named.
  - If fd is nil or has no receiver, returns (nil, false).
  - Relies on p.TypesInfo being populated; if unavailable, the fallback resolution may fail.
  - If a direct TypeName/Named resolution via the semantic (Defs) path fails, a fallback via
    p.TypesInfo.Uses is attempted by peeling the receiver expression (StarExpr, ParenExpr,
    IndexExpr, IndexListExpr) to obtain a base type. If that resolves to a *gtypes.Named, it is returned.
Note: This follows the guidance in the code: "Preferred: use go/types from the function object."

*/
func (p *PackageNode) IsMemberFunction(fd *ast.FuncDecl) (*gtypes.Named, bool) {
	if fd == nil || fd.Recv == nil || len(fd.Recv.List) == 0 {
		return nil, false
	}

	// Preferred: use go/types from the function object.
	if obj, ok := p.TypesInfo.Defs[fd.Name].(*gtypes.Func); ok && obj != nil {
		if sig, ok := obj.Type().(*gtypes.Signature); ok && sig.Recv() != nil {
			t := sig.Recv().Type()
			if p, ok := t.(*gtypes.Pointer); ok {
				t = p.Elem()
			}
			if n, ok := t.(*gtypes.Named); ok {
				return n, true
			}
		}
	}

	// Fallback: peel syntax and resolve via info.Uses.
	baseRecvExpr := func(e ast.Expr) ast.Expr {
		for {
			switch x := e.(type) {
			case *ast.StarExpr:
				e = x.X
			case *ast.ParenExpr:
				e = x.X
			case *ast.IndexExpr:
				e = x.X
			case *ast.IndexListExpr:
				e = x.X
			default:
				return e
			}
		}
	}

	switch e := baseRecvExpr(fd.Recv.List[0].Type).(type) {
	case *ast.Ident:
		if tn, ok := p.TypesInfo.Uses[e].(*gtypes.TypeName); ok {
			if n, ok := tn.Type().(*gtypes.Named); ok {
				return n, true
			}
		}
	case *ast.SelectorExpr:
		if tn, ok := p.TypesInfo.Uses[e.Sel].(*gtypes.TypeName); ok {
			if n, ok := tn.Type().(*gtypes.Named); ok {
				return n, true
			}
		}
	}

	return nil, false
}

/*
Summary:
UpdateDocsInFile iterates over FunctionDecls associated with the PackageNode and inserts generated documentation
for those that are marked as documented in this pass and currently lack inline documentation in their source file.
The updated documentation is written directly into the corresponding files at the appropriate byte offsets.

Signature:
func (p *PackageNode) UpdateDocsInFile() error

Parameters:
- p *PackageNode: the receiver containing FunctionDecls; no additional parameters.

Returns:
- error: non-nil if the insertion into a file fails; nil on success.

Errors/Exceptions:
- May panic if p is nil (no nil-check before accessing p.FunctionDecls).
- May return an error from insertIntoFile if reading, validating the offset, or writing the file fails.
- If FindStartEnd cannot resolve a start offset (e.g., fd is nil or p.Fset.File(fd.Pos()) is nil), the resulting start may be -1, which can cause insertion to fail.

Side Effects:
- Modifies target source files on disk by inserting documentation text at computed byte offsets.

Edge Cases & Assumptions:
- Assumes a non-nil receiver p and non-nil GetFunctionDecls content.
- Processes only FunctionDecls for which GetInfo().GetDocumentedInThisPass() is true.
- Skips FunctionDecls whose Node.Doc already contains non-empty text.
- Retrieves documentation via f.GetDocumentation() but ignores its error (uses the value returned).
- Uses f.GetFile() to determine the target file and FindStartEnd(fd) to compute the insertion offset; insertion occurs at that offset.

*/
func (p *PackageNode) UpdateDocsInFile() error {
	decls := p.GetFunctionDecls()
	for i := len(decls) - 1; i >= 0; i-- {
		if !decls[i].GetInfo().GetDocumentedInThisPass() {
			continue
		}

		f := decls[i].(*FunctionDecl)

		fd := f.Node

		// We read in pre-existing docs
		if fd.Doc != nil && strings.TrimSpace(fd.Doc.Text()) != "" {
			continue
		}

		start, _ := p.FindStartEnd(fd)

		docs, _ := f.GetDocumentation()

		log.Debugf("Documentation string for %v: %v", f.GetName(), docs)

		toAdd := fmt.Sprintf("%v\n", docs)

		err := insertIntoFile(f.GetFile(), start, toAdd)
		if err != nil {
			return fmt.Errorf("failed to update docs in file: %v", err)
		}
	}

	return nil
}

/*
Summary
Inserts the string insertion into the file at path at the byte offset offset.
The file is read, the insertion is placed at the specified position, and the result is written back to disk.

Signature
  func insertIntoFile(path string, offset int, insertion string) error

Parameters
  path: string — filesystem path of the target file.
  offset: int — byte offset at which to insert; must satisfy 0 <= offset <= len(data).
  insertion: string — text to insert at the offset.

Returns
  error — non-nil on read failure, offset out of range, or write failure;
           nil on success.

Errors/Exceptions
  - non-nil error from os.ReadFile on read failure
  - non-nil error when offset < 0 or offset > len(data)
  - non-nil error from os.WriteFile on write failure

Side Effects
  Reads and writes the target file on disk; the file content is modified in place.
  Uses file mode 0644 for the written file.

Edge Cases & Assumptions
  - offset is validated to be within [0, len(data)]
  - path must exist and be readable/writable by the running process
  - assumes insertion does not require creating directories or altering permissions beyond 0644

*/
func insertIntoFile(path string, offset int, insertion string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if offset < 0 || offset > len(data) {
		return fmt.Errorf("offset out of range")
	}
	out := append(append([]byte{}, data[:offset]...), append([]byte(insertion), data[offset:]...)...)
	return os.WriteFile(path, out, 0644)
}

/*
Summary: GetInfo returns the current FunctionInfo as mst.FunctionInfo.
Use this to obtain the information held by this FunctionInfo instance as the mst.FunctionInfo interface.
Signature: func (f *FunctionInfo) GetInfo() mst.FunctionInfo
Parameters:
  f *FunctionInfo: receiver containing the function information.
Returns:
  mst.FunctionInfo: this FunctionInfo promoted to the mst.FunctionInfo interface.
Errors/Exceptions: none.
Side Effects: none.
Edge Cases & Assumptions:
  - The result is the receiver captured as an mst.FunctionInfo interface value.
  - If called on a nil receiver, the returned interface has dynamic type *FunctionInfo and a nil value.
  - mst.FunctionInfo is implemented by *FunctionInfo.

*/
func (f *FunctionInfo) GetInfo() mst.FunctionInfo {
	return f
}

/*
Summary: GetPackage returns the mst.PackageNode associated with this FunctionInfo.
Use this method to access the package context for the function being described.

Signature: func (f *FunctionInfo) GetPackage() mst.PackageNode

Parameters: none.

Returns: mst.PackageNode
  - The PackageNode stored in f.Package for this FunctionInfo.
  - If f is nil, the call would cause a nil pointer dereference (runtime panic).

Errors/Exceptions: None returned directly. Potential panic if f is nil.

Side Effects: None.

Edge Cases & Assumptions:
  - Assumes the receiver f is non-nil when invoked.

*/
func (f *FunctionInfo) GetPackage() mst.PackageNode {
	return f.Package
}

/*
Summary: GetName returns the Name field value from the FunctionInfo receiver.

Use GetName to access the underlying Name stored in a FunctionInfo instance.

Signature: func (f *FunctionInfo) GetName() string

Parameters:
  - f: *FunctionInfo - the receiver containing the Name field; must be non-nil when invoked.

Returns:
  - string: the value of f.Name.

Errors/Exceptions:
  - Calling this method on a nil receiver will panic due to dereferencing a nil pointer.

Side Effects:
  - None.

Edge Cases & Assumptions:
  - Assumes the FunctionInfo instance is properly initialized and its Name field holds a valid string.

*/
func (f *FunctionInfo) GetName() string {
	return f.Name
}

/*
SetName sets the FunctionInfo.Name to the provided name. It always returns nil.

Signature: func (f *FunctionInfo) SetName(name string) error

Parameters:
- name: string. The new name to assign to f.Name.

Returns:
- error: always nil.

Errors/Exceptions:
- None. This method does not perform validation and cannot fail.

Side Effects:
- Mutates f.Name on the receiver.

Edge Cases & Assumptions:
- No validation is performed; empty strings are allowed.
- The method assumes f is non-nil; calling on a nil receiver will cause a runtime panic.

*/
func (f *FunctionInfo) SetName(name string) error {
	f.Name = name
	return nil
}

/*
Summary: Returns the fully-qualified name for this FunctionInfo. Uses f.GetResolvedPkg() as the package path; if f.Object is non-empty, the name includes the Object and the function Name, otherwise it includes only the package path and Name. This is useful for code generation or analysis needing a unique function identifier.

Signature: func (f *FunctionInfo) GetFullName() string

Parameters:
- f *FunctionInfo: receiver; the FunctionInfo instance to inspect.

Returns:
- string: the fully-qualified name. If f.Object != "" the form is <resolvedPkg>.<Object>.<Name>; otherwise <resolvedPkg>.<Name>.

Errors/Exceptions: none.

Side Effects: none.

Edge Cases & Assumptions:
- Assumes f.GetResolvedPkg() provides a non-empty package path (via f.ResolvedPkg or f.Package.GetPath()).
- Assumes f.Object and f.Name reflect the function's identifiers.

*/
func (f *FunctionInfo) GetFullName() string {
	if f.Object != "" {
		return fmt.Sprintf("%s.%s.%s", f.GetResolvedPkg(), f.Object, f.Name)
	}

	return fmt.Sprintf("%s.%s", f.GetResolvedPkg(), f.Name)
}

/*
Summary: Return the package path resolved for this FunctionInfo.
If f.ResolvedPkg is non-empty, return that value; otherwise return f.Package.GetPath().
Use this to obtain the effective package path for code generation or analysis.
Signature: func (f *FunctionInfo) GetResolvedPkg() string
Parameters:
- f *FunctionInfo: receiver; the FunctionInfo instance to inspect.
Returns:
- string: the resolved package path. If f.ResolvedPkg != "" returns f.ResolvedPkg; otherwise returns f.Package.GetPath().
Errors/Exceptions: none.
Side Effects: none.
Edge Cases & Assumptions:
- Assumes f.Package is non-nil and provides GetPath().

*/
func (f *FunctionInfo) GetResolvedPkg() string {
	if f.ResolvedPkg == "" {
		return f.Package.GetPath()
	}
	return f.ResolvedPkg
}

/*
Summary: Sets f.ResolvedPkg to the provided in string, persisting the resolved package name for this FunctionInfo.
Use when you have determined the resolved package name and need to store it on the FunctionInfo instance.

Signature: func (f *FunctionInfo) SetResolvedPkg(in string) error

Parameters:
- in: string; the resolved package name to assign to f.ResolvedPkg.

Returns:
- error: nil. This function always returns nil.

Errors/Exceptions:
- None produced by the function; it always returns nil.
- Precondition: f != nil.

Side Effects:
- Mutates the receiver: f.ResolvedPkg is set to in.

Edge Cases & Assumptions:
- If f is nil, the call will panic.
- Empty string is allowed; f.ResolvedPkg will be set to "".

*/
func (f *FunctionInfo) SetResolvedPkg(in string) error {
	f.ResolvedPkg = in
	return nil
}

/*

*/
func (f *FunctionInfo) GetObjectName() (string, error) {
	return f.Object, nil
}

/*
Summary: Sets f.Object to the provided name.
Signature: func (f *FunctionInfo) SetObjectName(name string) error
Parameters:
  name: string - the new object name
Returns:
  error - always nil
Side Effects:
  Mutates f.Object
Edge Cases & Assumptions:
  None

*/
func (f *FunctionInfo) SetObjectName(name string) error {
	f.Object = name
	return nil
}

/*
Summary: GetFile returns the value of the FunctionInfo.File field. Use this accessor when you need the associated file name or path without directly accessing the field.
Signature: func (f *FunctionInfo) GetFile() string
Parameters: none
Returns: string — the value of f.File; may be "" if not set.
Errors/Exceptions: none
Side Effects: none
Edge Cases & Assumptions:
- Precondition: f != nil. Accessing f.File requires a non-nil receiver; calling on a nil *FunctionInfo will trigger a nil dereference.
- If File is not initialized, GetFile returns "".

*/
func (f *FunctionInfo) GetFile() string {
	return f.File
}

/*
Summary
    Sets the FunctionInfo.File field to the provided file path. No validation is performed; this method always returns nil.

Signature
    func (f *FunctionInfo) SetFile(file string) error

Parameters
    file string - path to the file to associate with this FunctionInfo.

Returns
    error - Always nil.

Errors/Exceptions
    None produced by this method. If the receiver is nil, the method will panic due to dereferencing a nil pointer.

Side Effects
    Mutates the receiver by assigning f.File = file.

Edge Cases & Assumptions
    f must be non-nil when calling this method.
    Empty string is accepted and sets f.File to "".

*/
func (f *FunctionInfo) SetFile(file string) error {
	f.File = file
	return nil
}

/*
Summary
SetDocumentation assigns the provided docs string to f.Documentation.

When to use
Use this method to attach or update descriptive documentation text on a FunctionInfo instance.

Signature
func (f *FunctionInfo) SetDocumentation(docs string) error

Parameters
- f: *FunctionInfo — the receiver.
- docs: string — the documentation text to store in f.Documentation.

Returns
- error: always nil.

Errors/Exceptions
- None. This method never returns a non-nil error.

Side Effects
- Mutates f.Documentation.

Edge Cases & Assumptions
- If docs is "", f.Documentation becomes "".
- No validation is performed on docs.
- Not safe for concurrent use without external synchronization.

*/
func (f *FunctionInfo) SetDocumentation(docs string) error {
	f.Documentation = docs
	return nil
}

/*

*/
func (f *FunctionInfo) GetDocumentation() (string, error) {
	return f.Documentation, nil
}

/*
Summary: GetHasDocumentation reports whether the FunctionInfo has documentation by returning f.HasDocumentation.
Use when you need to check if a FunctionInfo instance has documentation.
Signature: func (f *FunctionInfo) GetHasDocumentation() bool
Parameters: none
Returns: bool - the value of f.HasDocumentation; true if documentation is present, otherwise false.
Side Effects: none
Edge Cases & Assumptions: Assumes FunctionInfo.HasDocumentation is a boolean field; no mutation or I/O occurs.

*/
func (f *FunctionInfo) GetHasDocumentation() bool {
	return f.HasDocumentation
}

/*
Summary:
SetHasDocumentation updates the FunctionInfo.HasDocumentation flag to the provided value.

Signature:
func (f *FunctionInfo) SetHasDocumentation(t bool) error

Parameters:
t bool - value to assign to FunctionInfo.HasDocumentation.

Returns:
error - always nil.

Errors/Exceptions:
- None produced by this function; however, if f is nil, the method will panic due to dereferencing a nil pointer.

Side Effects:
- Mutates f.HasDocumentation.

Edge Cases & Assumptions:
- f must be a non-nil *FunctionInfo before invocation.

*/
func (f *FunctionInfo) SetHasDocumentation(t bool) error {
	f.HasDocumentation = t
	return nil
}

/*

*/
func (f *FunctionInfo) GetDocumentedInThisPass() bool {
	return f.DocumentedInThisPass
}

/*
Summary: Sets f.DocumentedInThisPass to the provided value, marking whether this FunctionInfo has been documented in the current pass.
Use this to track documentation generation state for a FunctionInfo during a pass.
Signature: func (f *FunctionInfo) SetDocumentedInThisPass(t bool)
Parameters:
  t: bool - true to indicate the function has been documented in this pass; false to indicate it has not.
Returns: none.
Side Effects: mutates f.DocumentedInThisPass.
Edge Cases & Assumptions: f must be non-nil when calling this method.

*/
func (f *FunctionInfo) SetDocumentedInThisPass(t bool) {
	f.DocumentedInThisPass = t
}

/*
Summary
GetIsAiAware reports whether this FunctionInfo instance is AI-aware by returning the IsAiAware field.

Signature
func (f *FunctionInfo) GetIsAiAware() bool

Returns
bool — true if f.IsAiAware is set to true (AI-aware), otherwise false.

Errors/Exceptions
This function does not return errors.

Side Effects
No side effects.

Edge Cases & Assumptions
Relies on the f.IsAiAware field value accurately representing AI-awareness at the time of the call.

*/
func (f *FunctionInfo) GetIsAiAware() bool {
	return f.IsAiAware
}

/*
Summary:
Sets the FunctionInfo.IsAiAware flag to the provided boolean value.

Signature:
func (f *FunctionInfo) SetIsAiAware(t bool) error

Parameters:
- t: bool — value to assign to f.IsAiAware.

Returns:
- error — always nil.

Errors/Exceptions:
- none.

Side Effects:
- Mutates f.IsAiAware.

Edge Cases & Assumptions:
- None; this is a straightforward setter.

*/
func (f *FunctionInfo) SetIsAiAware(t bool) error {
	f.IsAiAware = t
	return nil
}

/*
Summary: GetDecl returns the FunctionDecl stored in FunctionInfo.Declaration.

Signature: func (f *FunctionInfo) GetDecl() mst.FunctionDecl

Parameters:
- none

Returns:
mst.FunctionDecl — the value held by f.Declaration.

Errors/Exceptions:
None. A nil receiver will panic at runtime.

Side Effects:
None.

Edge Cases & Assumptions:
Assumes the receiver f is non-nil when called; returns the zero value of mst.FunctionDecl if f.Declaration is zero-valued.

*/
func (f *FunctionInfo) GetDecl() mst.FunctionDecl {
	return f.Declaration
}

/*
Summary: Sets the FunctionDecl for the FunctionInfo by assigning to f.Declaration and returns nil.

Signature: func (f *FunctionInfo) SetDecl(d mst.FunctionDecl) error

Parameters:
- f: *FunctionInfo, receiver of the method.
- d: mst.FunctionDecl, the declaration to assign to f.Declaration.

Returns:
- error: always nil

Errors/Exceptions:
- None (always returns nil)

Side Effects:
- Mutates f.Declaration by assignment.

Edge Cases & Assumptions:
- f must be non-nil; calling with a nil receiver will panic when accessing f.Declaration.
- d is assigned directly without validation.

*/
func (f *FunctionInfo) SetDecl(d mst.FunctionDecl) error {
	f.Declaration = d
	return nil
}

/*
Summary:
GetCalls returns the list of mst.FunctionCall produced by this FunctionInfo by delegating to its Declaration. If f.Declaration is nil, nil is returned.

Signature:
func (f *FunctionInfo) GetCalls() []mst.FunctionCall

Parameters:
- none

Returns:
[]mst.FunctionCall: the result of f.Declaration.GetCalls() when f.Declaration != nil; otherwise nil.

Errors/Exceptions:
none

Side Effects:
none

Edge Cases & Assumptions:
- If f.Declaration == nil, GetCalls returns nil.
- If non-nil, GetCalls delegates to f.Declaration.GetCalls().

*/
func (f *FunctionInfo) GetCalls() []mst.FunctionCall {
	if f.Declaration == nil {
		return nil
	}

	return f.Declaration.GetCalls()
}

/*
Summary: PrettyPrint formats the FunctionInfo receiver as a human-readable representation using the provided string parameter.

Signature: func (f *FunctionInfo) PrettyPrint(string)

Parameters:
- string: format or mode guiding the pretty-printed output.

Returns: none

Errors/Exceptions: none observed.

Side Effects: none guaranteed by the current empty implementation; may perform I/O in future.

Edge Cases & Assumptions:
- Assumes f != nil when called; behavior for a nil receiver is undefined.

*/
func (f *FunctionInfo) PrettyPrint(string) {

}

/*
Summary: GetDocInsertLocation returns the doc insertion location for the FunctionInfo's FunctionDecl.
Signature: func (f *FunctionInfo) GetDocInsertLocation() uint
Parameters: none
Returns: uint — the document insertion location as provided by the underlying FunctionDecl.
Errors/Exceptions: None. A nil receiver will panic at runtime.
Side Effects: None.
Edge Cases & Assumptions: Assumes the receiver f is non-nil when called; returns the zero value of mst.FunctionDecl if f.Declaration is zero-valued.

*/
func (f *FunctionInfo) GetDocInsertLocation() uint {
	return f.GetDecl().GetDocInsertLocation()
}

/*
Summary:
CreateComment returns a string that wraps the input docs in a C-style block comment and returns that string.

Signature:
func (f *FunctionInfo) CreateComment(docs string) string

Parameters:
- docs: string — text to embed inside the block comment.

Returns:
- string — a block-comment string containing the docs.

Notes:
- No error handling; content is not escaped.
- If docs itself contains a comment end sequence, the result may not be a valid block comment when embedded.

Edge cases:
- Empty docs yields an empty block comment body inside the delimiters.

Side Effects:
- None.

*/
func (f *FunctionInfo) CreateComment(docs string) string {
	return "/*" + docs + "*/"
}

/*
Summary:
Returns the comment string for this FunctionDecl by delegating to the associated
FunctionInfo via f.GetInfo().CreateComment(docs).
When to use:
Retrieve the generated comment for a FunctionDecl without modifying the receiver.
Signature:
func (f *FunctionDecl) CreateComment(docs string) string
Parameters:
- f: *FunctionDecl - receiver; no additional parameters.
- docs: string - additional documentation text to incorporate when creating the comment.
Returns:
string - the value produced by FunctionInfo.CreateComment(docs).
Errors/Exceptions:
None. Note: calling on a nil *FunctionDecl will panic when accessing f.Info.
Side Effects:
None.
Edge Cases & Assumptions:
- f must be non-nil; calling on a nil *FunctionDecl will panic when accessing f.Info.
- Returns the same value as f.GetInfo().CreateComment(docs).

*/
func (f *FunctionDecl) CreateComment(docs string) string {
	return f.GetInfo().CreateComment(docs)
}

/*
Summary: CreateComment returns a generated comment for the underlying FunctionCall by delegating to the stored FunctionInfo.
Use when you need to produce a comment that describes the underlying function call with the given docs.
Signature: func (f *FunctionCall) CreateComment(docs string) string
Parameters:
  docs string: additional documentation to embed in the generated comment.
Returns:
  string: the comment text produced by FunctionInfo.CreateComment(docs).
Errors/Exceptions: none
Side Effects: none
Edge Cases & Assumptions:
  f.Info is assumed to be a valid mst.FunctionInfo value. This method does not mutate state.

*/
func (f *FunctionCall) CreateComment(docs string) string {
	return f.GetInfo().CreateComment(docs)
}

/*
Summary: Sets the FunctionDecl's Info field to the provided FunctionInfo and returns nil.
Use SetInfo when you want to associate or update the FunctionInfo for a FunctionDecl without validation.

Signature: func (f *FunctionDecl) SetInfo(i mst.FunctionInfo) error

Parameters:
- f: *FunctionDecl, the receiver on which SetInfo is invoked.
- i: mst.FunctionInfo, the value to assign to f.Info.

Returns:
- error: always nil.

Side Effects:
- Mutates f.Info to i.

Edge Cases & Assumptions:
- No validation is performed on i; this method always succeeds.

*/
func (f *FunctionDecl) SetInfo(i mst.FunctionInfo) error {
	f.Info = i
	return nil
}

/*
Summary:
Returns the FunctionInfo stored in f.Info for this FunctionDecl.

When to use:
Access the metadata of the FunctionDecl without modifying it.

Signature:
func (f *FunctionDecl) GetInfo() mst.FunctionInfo

Parameters:
- f: *FunctionDecl - receiver; no additional parameters.

Returns:
mst.FunctionInfo - the value stored in f.Info.

Errors/Exceptions:
None.

Side Effects:
None.

Edge Cases & Assumptions:
- f must be non-nil; calling on a nil *FunctionDecl will panic when accessing f.Info.
- Returns by value; if mst.FunctionInfo is a struct, this is a copy of the field.

*/
func (f *FunctionDecl) GetInfo() mst.FunctionInfo {
	return f.Info
}

/*
Summary: GetName returns the name of this FunctionDecl by delegating to f.Info.GetName().
Signature: func (f *FunctionDecl) GetName() string
Parameters:
  name: f
  type: *FunctionDecl
  role: receiver
  constraints: none
Returns:
  string: the function name.
Side Effects: None.

*/
func (f *FunctionDecl) GetName() string {
	return f.Info.GetName()
}

/*

*/
func (f *FunctionDecl) GetFullName() string {
	return f.Info.GetFullName()
}

/*
Summary: Returns the DocInsertLocation value stored in f.
Signature: func (f *FunctionDecl) GetDocInsertLocation() uint
Returns: uint - the value of f.DocInsertLocation.
Side Effects: none
Edge Cases & Assumptions:
- This method requires a non-nil receiver; calling it with a nil *FunctionDecl will panic due to nil dereference.

*/
func (f *FunctionDecl) GetDocInsertLocation() uint {
	return f.DocInsertLocation
}

/*
func (f *FunctionDecl) GetNode() *ast.FuncDecl
func (f *FunctionDecl) SetNode() *ast.FuncDecl
*/

/*
Summary: GetCalls returns the Calls field of the FunctionDecl. It is nil-safe and returns nil when the receiver is nil.

Signature: func (f *FunctionDecl) GetCalls() []mst.FunctionCall

Parameters:
  - f: *FunctionDecl — the receiver; may be nil. No additional input parameters.

Returns:
  - []mst.FunctionCall: the Calls slice from f; nil if f == nil.

Errors/Exceptions:
  - None. This method does not return an error.

Side Effects:
  - None. This is a read-only accessor.

Edge Cases & Assumptions:
  - If f == nil, returns nil and does not panic.
  - The returned value is the underlying f.Calls slice; callers should treat it as a read-only view of that data unless they intend to mutate the elements.

*/
func (f *FunctionDecl) GetCalls() []mst.FunctionCall {
	if f == nil {
		return nil
	}

	return f.Calls
}

/*
Summary: Sets the f.Calls field of FunctionDecl to the provided []mst.FunctionCall.
Use when you want to replace the list of function calls associated with this function declaration.

Signature: func (f *FunctionDecl) SetCalls(c []mst.FunctionCall) error

Parameters:
- f: *FunctionDecl, receiver; role: target instance to mutate; constraints: non-nil when called.
- c: []mst.FunctionCall; role: new calls to assign; constraints: may be nil.

Returns:
- error: always nil in this implementation.

Side Effects:
- Mutates the receiver by assigning f.Calls = c.
- Assigns the slice header; underlying array may be shared with caller.
- No I/O or concurrency side effects.

Edge Cases & Assumptions:
- c == nil results in f.Calls == nil.
- No validation of c; caller is responsible for ensuring the contents.

*/
func (f *FunctionDecl) SetCalls(c []mst.FunctionCall) error {
	f.Calls = c
	return nil
}

/*
Summary: Appends the provided FunctionCall to f.Calls, initializing the slice if nil.
Signature: func (f *FunctionDecl) AddCall(c mst.FunctionCall) error
Parameters:
  f: *FunctionDecl - the receiver being mutated; must be non-nil when called.
  c: mst.FunctionCall - the call to add to f.Calls.
Returns: error - always nil.
Errors/Exceptions: None.
Side Effects: Mutates f.Calls by appending c; may allocate a new slice when f.Calls is nil.
Edge Cases & Assumptions: The receiver f must be non-nil when invoked; no validation or deduplication is performed.

*/
func (f *FunctionDecl) AddCall(c mst.FunctionCall) error {
	if f.Calls == nil {
		f.Calls = make([]mst.FunctionCall, 0)
	}

	f.Calls = append(f.Calls, c)
	return nil
}

/*
Summary:
FindStartEnd returns the byte offsets [start, end) of the given ast.Node within the file associated with p.Fset. Use these offsets to slice the source corresponding to the node. If the node is nil or the file cannot be determined, the function returns -1, -1.

Signature:
func (p *PackageNode) FindStartEnd(n ast.Node) (int, int)

Parameters:
- n: ast.Node — the node whose span to compute.

Returns:
- start int: byte offset of the node's start (file.Offset(n.Pos())).
- end int: byte offset just after the node (file.Offset(n.End())).

Errors/Exceptions:
- Returns (-1, -1) when n == nil or p.Fset.File(n.Pos()) == nil.

Side Effects:
- None.

Edge Cases & Assumptions:
- End() is the position after the node; thus [start:end] can safely slice the source for the node.
- If the node's position cannot be resolved to a file, both values are -1.

*/
func (p *PackageNode) FindStartEnd(n ast.Node) (int, int) {
	if n == nil {
		return -1, -1
	}
	file := p.Fset.File(n.Pos())
	if file == nil {
		return -1, -1
	}
	// End() is the position *after* the node; safe for slicing [start:end].
	start := file.Offset(n.Pos())
	end := file.Offset(n.End())
	return start, end
}

/*
Summary: ToStringForAi returns the AI string representation of the FunctionInfo.Declaration by delegating to f.GetDecl().ToStringForAi().

Signature: func (f *FunctionInfo) ToStringForAi() (string, error)

Parameters:
- none

Returns:
- string: the result of FunctionDecl.ToStringForAi()
- error: propagated from FunctionDecl.ToStringForAi()

Errors/Exceptions:
- A nil receiver will panic at runtime.

Side Effects:
- None.

Edge Cases & Assumptions:
- Assumes the receiver f is non-nil when called.
- If f.Declaration is zero-valued, GetDecl() yields the zero-value mst.FunctionDecl and its ToStringForAi() is invoked accordingly.

*/
func (f *FunctionInfo) ToStringForAi() (string, error) {
	return f.GetDecl().ToStringForAi()
}

/*
Summary: ToStringForAi returns the source text of the FunctionDecl from its associated file, with per-call documentation inlined at the corresponding line positions. It builds the result by slicing the function's source from f.Info.GetFile() and inserting documentation retrieved from each call's GetInfo().GetDocumentation().
Signature: func (f *FunctionDecl) ToStringForAi() (string, error)
Returns:
- string: the function's source text with inlined documentation for each Call.
- error: non-nil if the underlying file cannot be read (read file: ...).
Errors/Exceptions: The function returns an error only when reading the file fails. Other steps ignore per-call documentation retrieval errors (they skip if docs are empty).
Side Effects: Reads the file from disk; constructs and returns a modified string without writing to disk.
Edge Cases & Assumptions: Assumes f.FindStartEnd() yields valid [fd_start, fd_end] offsets into the file; assumes fc_start/fc_end and fc_line_no become valid indices within fd_text after subtracting fd_start; processes f.Calls in reverse order to avoid index shifting when inserting docs. Each per-call documentation is retrieved via f.Calls[i].GetInfo().GetDocumentation() and then passed through UnescapeCommonChars before insertion.

*/
func (f *FunctionDecl) ToStringForAi() (string, error) {

	// // This should only be one layer deep. We are using comments to avoid the recursion

	raw, err := os.ReadFile(f.Info.GetFile())
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}

	fd_start, fd_end := f.FindStartEnd()
	fd_text := string(raw)[fd_start:fd_end]

	for i := len(f.Calls) - 1; i >= 0; i-- {
		fc_line_no := f.Calls[i].FindLineNo()
		fc_start, fc_end := f.Calls[i].FindStartEnd()

		fc_start -= fd_start
		fc_end -= fd_start
		fc_line_no -= fd_start

		docs, _ := f.Calls[i].GetInfo().GetDocumentation()
		if strings.TrimSpace(docs) == "" {
			continue
		}
		docs = UnescapeCommonChars(docs)
		fd_text = fd_text[:fc_line_no] + docs + "\n" + fd_text[fc_line_no:]
	}

	return fd_text, nil
}

/*
Summary: Unescapes common backslash-escaped sequences in s by replacing them with their literal characters (e.g., \n -> newline, \\ -> \).
Use when you need to convert a string containing escaped sequences into its actual character representation.

Signature: func UnescapeCommonChars(s string) string

Parameters:
- s: string to process. May contain the escape sequences: \\, \", \', \n, \t, \r.
  Unrecognized escapes are left unchanged.

Returns:
- string: the input with the listed escape sequences replaced by their corresponding characters.

Errors/Exceptions: none. This function does not return an error.

Side Effects: none beyond allocation for the replacer and the resulting string.

Edge Cases & Assumptions:
- Only the specified escapes are handled; other backslash sequences remain as-is.
- If s already contains real newline/tab/etc., they are preserved unless escaped as \n, \t, etc.

*/
func UnescapeCommonChars(s string) string {
	replacer := strings.NewReplacer(
		`\\`, `\`, // backslash
		`\"`, `"`, // double quote
		`\'`, `'`, // single quote
		`\n`, "\n", // newline
		`\t`, "\t", // tab
		`\r`, "\r", // carriage return
	)
	return replacer.Replace(s)
}

/*

*/
func (f *FunctionDecl) PrettyPrint(string) {

}

/*
Summary: Sets the FunctionCall's Info to the provided mst.FunctionInfo and returns nil.
Use SetInfo to attach or update the metadata represented by mst.FunctionInfo on a FunctionCall.

Signature: func (f *FunctionCall) SetInfo(i mst.FunctionInfo) error

Parameters:
  i: mst.FunctionInfo - the new information to assign to f.Info.

Returns:
  error - always nil; this method does not error in the current implementation.

Errors/Exceptions:
  None. This method always succeeds.

Side Effects:
  Mutates the FunctionCall receiver by assigning f.Info = i.

Edge Cases & Assumptions:
  The method requires a non-nil receiver. Calling on a nil *FunctionCall will panic.
  If i has zero-values, f.Info is updated to those zero-values.

*/
func (f *FunctionCall) SetInfo(i mst.FunctionInfo) error {
	f.Info = i
	return nil
}

/*
GetInfo returns the FunctionInfo stored on this FunctionCall.
It exposes the metadata for the underlying function call.

Signature: func (f *FunctionCall) GetInfo() mst.FunctionInfo

Parameters: none

Returns: mst.FunctionInfo - the FunctionInfo value held by f.Info

Errors/Exceptions: none

Side Effects: none

Edge Cases & Assumptions:
- f.Info is assumed to be a valid mst.FunctionInfo value.
- This method does not mutate state.

*/
func (f *FunctionCall) GetInfo() mst.FunctionInfo {
	return f.Info
}

/*
Summary: GetName returns the name of this FunctionCall by delegating to f.Info.GetName().
Signature: func (f *FunctionCall) GetName() string
Returns: string - the name provided by f.Info.GetName().
Errors/Exceptions: none returned; may panic if f is nil or f.Info is nil.
Side Effects: none.
Edge Cases & Assumptions: assumes f != nil and f.Info != nil; a nil receiver or nil f.Info may cause a panic.

*/
func (f *FunctionCall) GetName() string {
	return f.Info.GetName()
}

/*
Summary: Returns the full name of the FunctionCall by delegating to f.Info.GetFullName().
Use when you need the canonical full name as determined by the underlying Info object.

Signature: func (f *FunctionCall) GetFullName() string

Parameters: none

Returns: string - the full name as provided by f.Info.GetFullName()

Errors/Exceptions: May panic if f is nil or if f.Info is nil due to nil pointer dereference.

Side Effects: none

Edge Cases & Assumptions:
- Assumes f.Info implements GetFullName() and is non-nil when called.
- If f is nil, or f.Info is nil, a runtime panic may occur due to dereferencing a nil pointer.
- The returned value reflects the underlying Info.GetFullName() implementation.

*/
func (f *FunctionCall) GetFullName() string {
	return f.Info.GetFullName()
}

/*
func (f *FunctionCall) GetNode() *ast.FuncDecl
func (f *FunctionCall) SetNode() *ast.FuncDecl
*/

/*
Summary
GetKind returns the mst.FunctionCallKind stored in f.Kind for the FunctionCall instance f. Use this accessor to obtain the kind without modifying the object.

Signature
func (f *FunctionCall) GetKind() mst.FunctionCallKind

Returns
mst.FunctionCallKind representing the kind of this FunctionCall (the value of f.Kind).

Side Effects
None. This is a read-only accessor.

Edge Cases & Assumptions
- f must be non-nil; calling GetKind on a nil *FunctionCall will panic due to dereferencing f.Kind.
- This method does not mutate state or perform any transformation; it simply returns the stored kind.

*/
func (f *FunctionCall) GetKind() mst.FunctionCallKind {
	return f.Kind
}

/*
Summary:
Sets the FunctionCall's Kind to the provided mst.FunctionCallKind value and returns nil. Use to assign or update the kind of a FunctionCall instance.

Signature:
func (f *FunctionCall) SetKind(k mst.FunctionCallKind) error

Parameters:
- f: *FunctionCall, receiver (implicit)
- k: mst.FunctionCallKind, the new kind to assign

Returns:
- error: always nil

Errors/Exceptions:
- None (this method never returns an error)

Side Effects:
- Mutates the receiver by setting f.Kind

Edge Cases & Assumptions:
- Calling on a nil *FunctionCall will panic.
- No validation on k; any mst.FunctionCallKind value is accepted.

*/
func (f *FunctionCall) SetKind(k mst.FunctionCallKind) error {
	f.Kind = k
	return nil
}

/*

*/
func (f *FunctionCall) PrettyPrint(string) {

}

/*
Summary: Returns the document insertion location for this FunctionCall by delegating to the underlying FunctionDecl.
Use when you need the position to insert documentation related to this call's declaration.
Signature: func (f *FunctionCall) GetDocInsertLocation() uint
Parameters: none
Returns: uint - the doc insertion location within the underlying FunctionDecl.
Errors/Exceptions: none
Side Effects: none
Edge Cases & Assumptions: Precondition: f.Info != nil; a nil pointer dereference will occur if f.Info is nil. This method relies on GetDecl(), which requires a valid f.Info.

*/
func (f *FunctionCall) GetDocInsertLocation() uint {
	return f.GetDecl().GetDocInsertLocation()
}

/*
Summary:
FindLineNo returns the absolute byte offset of the start of the line that contains f.Node by delegating to f.Package.(*PackageNode).FindLineNo(f.Node). Use this to locate the line start for a FunctionDecl's node.

Signature:
func (f *FunctionDecl) FindLineNo() int

Parameters:
  - none

Returns:
  int - the absolute offset (in bytes) of the start of the line containing f.Node; -1 if f.Node is nil.

Errors/Exceptions:
  May panic if f.Package is not a *PackageNode, or if f.Node.Pos() is not registered in the PackageNode's FileSet; underlying FindLineNo may also panic when Pos() is invalid.

Side Effects:
  None.

Edge Cases & Assumptions:
  Assumes f.Node is a non-nil ast.Node with a valid Pos() registered in f.Package's FileSet; nil f.Node yields -1 via the underlying behavior. If the PackageNode type assertion fails, a panic may occur.

*/
func (f *FunctionDecl) FindLineNo() int {
	return f.Package.(*PackageNode).FindLineNo(f.Node)
}

/*
FindLineNo returns the absolute byte offset of the start of the line that contains
the AST node associated with this FunctionCall.
It delegates to f.Package.(*PackageNode).FindLineNo(f.Node) to compute the offset
from the FileSet position information.
Signature: func (f *FunctionCall) FindLineNo() int

Returns:
  int - the absolute offset (in bytes) of the start of the line containing f.Node;
        -1 if the node is nil or the underlying calculation cannot resolve the line.

Errors/Exceptions:
  No explicit errors are returned. If the underlying PackageNode.FindLineNo panics due to
  invalid inputs or FileSet state, the panic will propagate.

Side Effects:
  None.

Edge Cases & Assumptions:
  Assumes f.Node is a valid ast.Node registered in the FileSet associated with f.Package.
  If f.Node is nil, the behavior follows that of PackageNode.FindLineNo for a nil node.

*/
func (f *FunctionCall) FindLineNo() int {
	return f.Package.(*PackageNode).FindLineNo(f.Node)
}

/*
Summary:
Return the byte offsets [start, end) of f.Node within the file associated with f.Info.(*FunctionInfo).Package.(*PackageNode). This enables slicing the source corresponding to the node. The result is obtained by delegating to the underlying PackageNode.FindStartEnd.

Signature:
func (f *FunctionDecl) FindStartEnd() (int, int)

Returns:
- start int: byte offset of the node's start (file.Offset(n.Pos())).
- end int: byte offset just after the node (file.Offset(n.End())).

Errors/Exceptions:
- Returns (-1, -1) when the node cannot be resolved to a file (e.g., nil node or unresolved file).

Side Effects:
- None.

Edge Cases & Assumptions:
- End() is the position after the node; thus [start:end] can safely slice the source for the node.
- If the node's position cannot be resolved to a file, both values are -1.

*/
func (f *FunctionDecl) FindStartEnd() (int, int) {
	return f.Info.(*FunctionInfo).Package.(*PackageNode).FindStartEnd(f.Node)
}

/*
FindStartEnd returns the byte-offset span [start, end) for the ast.Node
referenced by f.Node, as resolved through the enclosing FunctionCall's context.
It delegates to f.Info.(*FunctionInfo).Package.(*PackageNode).FindStartEnd(f.Node)
to compute the exact offsets within the source file.
Use these offsets to slice the source corresponding to the node.
If the node or the underlying file cannot be determined, the result is (-1, -1).

Signature:
func (f *FunctionCall) FindStartEnd() (int, int)

Returns:
- start int: byte offset of the node's start (as provided by file.Offset(n.Pos()))
- end   int: byte offset just after the node (as provided by file.Offset(n.End()))

Errors/Exceptions:
- Returns (-1, -1) if f.Node is nil or the chain to the in-file node cannot be resolved.
- May panic if the concrete types of f.Info or its PackageNode are not as expected.

Side Effects:
- None.

Edge Cases & Assumptions:
- End() yields the position immediately after the node; [start:end] selects the node's source.
- Relies on the integrity of f.Info, f.Node, and the nested PackageNode to be non-nil and of expected types.

*/
func (f *FunctionCall) FindStartEnd() (int, int) {
	return f.Info.(*FunctionInfo).Package.(*PackageNode).FindStartEnd(f.Node)
}

/*
Summary: Returns the slice of mst.FunctionCall associated with this FunctionCall by delegating to f.Info.GetCalls() when Info is present; otherwise returns nil.
Signature: func (f *FunctionCall) GetCalls() []mst.FunctionCall
Parameters: none
Returns: []mst.FunctionCall; nil when f.Info is nil; otherwise the result of f.Info.GetCalls()
Errors/Exceptions: none
Side Effects: none
Edge Cases & Assumptions: if f.Info is nil, this function returns nil; otherwise it relies on f.Info.GetCalls() for the result

*/
func (f *FunctionCall) GetCalls() []mst.FunctionCall {
	if f.Info == nil {
		return nil
	}

	return f.Info.GetCalls()
}

/*
GetDecl returns the FunctionDecl describing this FunctionCall.
It delegates to f.Info.GetDecl().

Signature: func (f *FunctionCall) GetDecl() mst.FunctionDecl
Parameters: none
Returns: mst.FunctionDecl - the underlying function declaration.
Errors/Exceptions: none
Side Effects: none
Edge Cases & Assumptions: Precondition: f.Info != nil; a nil pointer dereference will occur if f.Info is nil.

*/
func (f *FunctionCall) GetDecl() mst.FunctionDecl {
	return f.Info.GetDecl()
}

/*
Summary: Returns the file path associated with this FunctionCall by delegating to f.Info.GetFile(). Use when you need the source file for this function call.
Signature: func (f *FunctionCall) GetFile() string
Parameters: none
Returns: string — the file path as provided by f.Info.GetFile().
Errors/Exceptions: none (Note: a nil f.Info would cause a panic when calling f.Info.GetFile()).
Side Effects: none
Edge Cases & Assumptions:
- Precondition: f.Info != nil.
- If f.Info.GetFile() returns an empty string, this method returns that value.

*/
func (f *FunctionCall) GetFile() string {
	return f.Info.GetFile()
}

/*
Summary:
SetDocumentedInThisPass records whether this FunctionCall has been documented in the current pass. It delegates the state change to the embedded Info via f.Info.SetDocumentedInThisPass(in).

Signature:
func (f *FunctionCall) SetDocumentedInThisPass(in bool)

Parameters:
- in: bool. True to mark as documented in this pass; false to clear.

Returns:
- none.

Errors/Exceptions:
- None documented.

Side Effects:
- Mutates f.Info by calling f.Info.SetDocumentedInThisPass(in).

Edge Cases & Assumptions:
- Assumes f and f.Info are non-nil when called.
- No additional validation beyond delegation.

*/
func (f *FunctionCall) SetDocumentedInThisPass(in bool) {
	f.Info.SetDocumentedInThisPass(in)
}

/*
Summary:
Marks this FunctionCall as documented in the current pass by setting the internal documented flag to the value of in.

Signature:
func (f *FunctionCall) AddCall(in bool)

Parameters:
in bool - Indicates whether this FunctionCall is documented in this pass.

Returns:
None.

Errors/Exceptions:
Potential panic if f or f.Info is nil.

Side Effects:
Updates f.Info via SetDocumentedInThisPass(in).

Edge Cases & Assumptions:
Assumes f and f.Info are non-nil. If in is true, the FunctionCall is flagged as documented in this pass; if false, it is flagged as not documented.

*/
func (f *FunctionCall) AddCall(in bool) {
	f.Info.SetDocumentedInThisPass(in)
}

/*
Summary: ToStringForAi returns the AI-ready string representation of this FunctionCall by delegating to f.Info.GetDecl().ToStringForAi().
Signature: func (f *FunctionCall) ToStringForAi() (string, error)
Parameters:
  - f: *FunctionCall — receiver; must be non-nil.
Returns:
  - string: AI-friendly string representation produced by the underlying declaration.
  - error: propagated from f.Info.GetDecl().ToStringForAi().
Errors/Exceptions:
  - Propagates any error returned by f.Info.GetDecl().ToStringForAi().
Side Effects:
  - None.
Edge Cases & Assumptions:
  - f != nil; otherwise a nil-pointer dereference occurs.
  - f.Info != nil and f.Info.GetDecl() != nil; otherwise panic at runtime.

*/
func (f *FunctionCall) ToStringForAi() (string, error) {
	return f.Info.GetDecl().ToStringForAi()
}

/*
Summary: Returns the underlying mst.FunctionDecl value for this FunctionDecl.

Signature: func (f *FunctionDecl) GetDecl() mst.FunctionDecl

Returns: mst.FunctionDecl — the function declaration value represented by this receiver.

Side Effects: none.

Edge Cases & Assumptions: Safe to call on a nil receiver; the exact returned value when f is nil depends on mst.FunctionDecl's underlying type (e.g., interface vs. concrete type).

*/
func (f *FunctionDecl) GetDecl() mst.FunctionDecl {
	return f
}

/*
Summary: GetFile returns the file path associated with the FunctionDecl by delegating to f.Info.GetFile().
Use this to determine the source file where the function declaration is defined.
Signature: func (f *FunctionDecl) GetFile() string
Parameters:
- f: *FunctionDecl — receiver; the function declaration instance. Assumes f != nil.
Returns:
- string: the file path as reported by f.Info.GetFile().
Errors/Exceptions: None.
Side Effects: None.
Edge Cases & Assumptions: Assumes f.Info is non-nil; f.Info.GetFile() returns a string representing the file path.

*/
func (f *FunctionDecl) GetFile() string {
	return f.Info.GetFile()
}

/*
Summary: Marks whether this FunctionDecl has been documented in the current pass by updating its internal state.
This method delegates to f.Info.SetDocumentedInThisPass(in) for the actual state mutation.
Use this to track documentation progress during incremental passes.

Signature: func (f *FunctionDecl) SetDocumentedInThisPass(in bool)

Parameters:
  f: *FunctionDecl, the receiver on which the method is invoked.
  in: bool, when true indicates the function was documented in this pass; false otherwise.

Returns: none

Errors/Exceptions: none

Side Effects: Mutates f.Info via SetDocumentedInThisPass(in).

Edge Cases & Assumptions: Assumes f and f.Info are non-nil and that f.Info implements SetDocumentedInThisPass(bool).

*/
func (f *FunctionDecl) SetDocumentedInThisPass(in bool) {
	f.Info.SetDocumentedInThisPass(in)
}

/*
Summary: GetName returns the TypeDef.Name field value. Use as a simple accessor to retrieve the name without direct field access.

Signature: func (t *TypeDef) GetName() string

Parameters:
  t *TypeDef: receiver; no additional parameters.

Returns:
  string: the TypeDef.Name value.

Errors/Exceptions:
  None.

Side Effects:
  None.

Edge Cases & Assumptions:
  The receiver t must be non-nil; calling on a nil *TypeDef will panic due to dereferencing t.Name.

*/
func (t *TypeDef) GetName() string {
	return t.Name
}
