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
Summary: Returns the internal PackageNodes slice stored in the MST instance.
Signature: func (m *MST) GetPackages() []mst.PackageNode
Returns: []mst.PackageNode - the PackageNodes stored in the MST.
Side Effects: None.
Edge Cases & Assumptions: If the receiver m is nil, this will panic when accessing m.PackageNodes.

*/
func (m *MST) GetPackages() []mst.PackageNode {
	return m.PackageNodes
}

/*
Summary: Sets the MST's PackageNodes field to the provided pkgs.
Use to update the list of package nodes tracked by the MST.

Signature: func (m *MST) SetPackages(pkgs []mst.PackageNode) error

Parameters:
- m: *MST, the receiver instance to modify.
- pkgs: []mst.PackageNode, the new package nodes to assign; may be nil.

Returns:
- error: Always nil.

Side Effects:
- Mutates m.PackageNodes to equal pkgs.

Edge Cases & Assumptions:
- pkgs == nil results in m.PackageNodes being nil.
- No validation of pkgs contents is performed.
- No concurrency control is indicated; caller is responsible for synchronization.

*/
func (m *MST) SetPackages(pkgs []mst.PackageNode) error {
	m.PackageNodes = pkgs
	return nil
}

/*
Summary: Appends a mst.PackageNode to the MST's PackageNodes slice, initializing the slice if nil.
Signature: func (m *MST) AddPackage(n mst.PackageNode) error
Parameters:
- m: *MST, the receiver.
- n: mst.PackageNode, the node to add.
Returns:
- error: always nil in this implementation.
Errors/Exceptions:
- None (no error path; function always returns nil).
Side Effects:
- Mutates m.PackageNodes by appending n; initializes m.PackageNodes if it is nil.
- May trigger allocation when the underlying slice grows.
Edge Cases & Assumptions:
- If m.PackageNodes is nil, it is initialized as make([]mst.PackageNode, 0) before appending.
- No validation on n; the provided node is added as-is.

*/
func (m *MST) AddPackage(n mst.PackageNode) error {
	if m.PackageNodes == nil {
		m.PackageNodes = make([]mst.PackageNode, 0)
	}

	m.PackageNodes = append(m.PackageNodes, n)

	return nil
}

/*
Summary: AddToFunctionMap adds an entry to the MST.FunctionMap, initializing the map on first use if needed, and stores the provided info under the given name.

Signature: func (m *MST) AddToFunctionMap(name string, info mst.FunctionInfo) error

Parameters:
- name: string
  role: key under which info is stored in m.FunctionMap
  constraints: no validation; can be empty

- info: mst.FunctionInfo
  role: value to store at m.FunctionMap[name]

Returns:
- error: nil

Errors/Exceptions:
- None (the function always returns nil). Note: if m == nil, the call will panic when accessing m.FunctionMap.

Side Effects:
- Mutates m.FunctionMap: initializes it when nil, then assigns m.FunctionMap[name] = info.

Edge Cases & Assumptions:
- Requires m != nil
- If name already exists, its value is overwritten
- No validation on name or info

*/
func (m *MST) AddToFunctionMap(name string, info mst.FunctionInfo) error {
	if m.FunctionMap == nil {
		m.FunctionMap = make(map[string]mst.FunctionInfo)
	}

	m.FunctionMap[name] = info

	return nil
}

/*
Summary: GetFromFunctionMap returns the FunctionInfo for the given name from m.FunctionMap, initializing the map if nil.

Signature: func (m *MST) GetFromFunctionMap(name string) (mst.FunctionInfo, bool, error)

Parameters:
- name: string — the key to look up in m.FunctionMap.

Returns:
- val: mst.FunctionInfo — the value associated with name if present; zero-value if not.
- ok: bool — true if name exists in m.FunctionMap; false otherwise.
- err: error — always nil in this implementation.

Errors/Exceptions: nil

Side Effects:
- If m.FunctionMap is nil, it is initialized to an empty map.

Edge Cases & Assumptions:
- If name is not present, ok is false and val is the zero-value of mst.FunctionInfo.

*/
func (m *MST) GetFromFunctionMap(name string) (mst.FunctionInfo, bool, error) {
	if m.FunctionMap == nil {
		m.FunctionMap = make(map[string]mst.FunctionInfo)
	}

	val, ok := m.FunctionMap[name]

	return val, ok, nil
}

/*

*/
/*

*/
func (m *MST) PrettyPrint(string) {

}

/*
Summary:
Populate loads Go packages from the provided folders, constructs PackageNode entries for each loaded package, populates their detailed information via PopulatePackageInformation, and stores the resulting nodes in MST.PackageNodes. Use this to initialize an MST with metadata from a set of folders.

Signature:
func (m *MST) Populate(folders []string) error

Parameters:
- m: *MST — receiver; MST instance to populate; mutated to hold PackageNodes.
- folders: []string — input folder paths to load packages from.

Returns:
- error — non-nil on failure; nil on success.

Errors/Exceptions:
- Propagates errors from pkgNode.PopulatePackageInformation() with a wrapper: "failed to populate package information: %v".
- Note: errors from packages.Load are not explicitly handled in this function as written.

Side Effects:
- Mutates m.PackageNodes by appending a PackageNode per loaded package.
- Creates new PackageNode instances with fields MST, Package, FunctionDecls, and Imports initialized.
- Logs the number of functions declared per package via log.Debugf.

Edge Cases & Assumptions:
- If no folders or no packages are found, m.PackageNodes remains unchanged.
- Assumes the per-package PopulatePackageInformation handles its own internal errors; this function stops and returns on the first error encountered.
- The function processes each folder and each package yielded by packages.Load; errors from Load itself are not surfaced here.

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
Summary
Clips cyclic dependencies in all packages tracked by the MST by invoking ClipFunctionCycles on every function declaration.

Signature
func (m *MST) HandleCyclicDependencies() error

Parameters
- m: *MST, receiver. The MST instance on which this method operates. Non-nil precondition.

Returns
- error: nil on success; non-nil if any ClipFunctionCycles invocation fails, with an explanatory message.

Errors/Exceptions
- Returns an error wrapping the underlying error from ClipFunctionCycles when a cycle clipping operation fails.
- May panic if called on a nil *MST due to dereferencing in GetPackages() (precondition: m != nil).

Side Effects
- May mutate internal state of PackageNode(s) via ClipFunctionCycles as part of clipping cycles.

Edge Cases & Assumptions
- Assumes the receiver m is non-nil.
- If a package contains no function declarations, it is skipped.
- GetPackages() is assumed to return the current PackageNodes; order is not guaranteed.

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
Summary: Returns the absolute offset (in bytes) of the start of the line that contains the given AST node n within the PackageNode's file set. Returns -1 if n is nil.
Signature: func (p *PackageNode) FindLineNo(n ast.Node) int
Parameters:
  - n: ast.Node; the AST node whose line start offset is requested. If nil, the function returns -1.
Returns:
  - int: the absolute offset of the start of the line containing n.Pos(), or -1 when n is nil.
Errors/Exceptions: none (no error return value).
Side Effects: reads p.Fset, its File, and position information to compute the offset.
Edge Cases & Assumptions: none beyond handling a nil n; assumes n.Pos() is within p.Fset.

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
ClipFunctionCycles removes cycles in function declarations by pruning calls that recreate recursion along the provided callStack.
Use during call graph processing to break cycles without descending into cycle-causing nodes.
Signature: func (p *PackageNode) ClipFunctionCycles(f mst.FunctionInfo, callStack []string) error
Parameters:
- f: mst.FunctionInfo; the function to process.
- callStack: []string; the current recursion path as a slice of function full names.
Returns:
- error: nil (function always returns nil; no error is produced).
Side Effects:
- Mutates f.GetDecl().GetCalls() by removing entries at positions identified by to_remove.
- May recursively call ClipFunctionCycles on non-cyclic calls.
- May mutate declarations in nested function infos as cycles are pruned.
Implementation notes:
- If f.GetDecl() == nil, returns nil immediately.
- For each call in decl.GetCalls(), if call.GetFullName() is in callStack, mark its index for removal;
  otherwise, recursively descend with p.ClipFunctionCycles(call.GetInfo(), append(callStack, f.GetFullName())).
- After traversal, remove marked entries from decl.GetCalls() in reverse order to preserve remaining indices.
- Uses slices.Contains to detect whether a call is already on the callStack.
Edge Cases & Assumptions:
- If there are no calls or to_remove is empty, nothing is changed.
- Assumes f.GetDecl(), decl.GetCalls(), call.GetFullName(), call.GetInfo(), and f.GetFullName() behave as in the code.

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
Summary: GetMST returns the mst.MST stored in the PackageNode.
Use this getter to access the MST value without exposing internal fields.

Signature: func (p *PackageNode) GetMST() mst.MST

Returns: mst.MST - the MST value held by p.

Edge Cases & Assumptions: Assumes the receiver p is a non-nil *PackageNode; calling on a nil receiver will panic.

*/
func (p *PackageNode) GetMST() mst.MST {
	return p.MST
}

/*
Summary: Sets the Documentation field of the underlying FunctionInfo for this FunctionDecl to the provided docs string.
Signature: func (p *FunctionDecl) SetDocumentation(docs string) error
Parameters:
- p: *FunctionDecl, receiver; the target object to modify
- docs: string; the documentation text to assign to p.Info.(*FunctionInfo).Documentation
Returns:
- error: nil (always)
Errors/Exceptions:
- None returned; runtime panics may occur if p.Info is nil or not of type *FunctionInfo
Side Effects:
- Mutates p.Info.(*FunctionInfo).Documentation
Edge Cases & Assumptions:
- Precondition: p.Info is non-nil and of type *FunctionInfo
- If the type assertion fails at runtime, a panic will occur

*/
func (p *FunctionDecl) SetDocumentation(docs string) error {
	p.Info.(*FunctionInfo).Documentation = docs
	return nil
}

/*
Summary: Sets the Documentation field of the embedded FunctionInfo for this FunctionCall to the provided docs string.
Use when you want to attach or update the function's documentation text on a FunctionCall.
Signature: func (p *FunctionCall) SetDocumentation(docs string) error
Parameters:
  docs: string; the documentation text to assign to p.Info.(*FunctionInfo).Documentation.
Returns:
  error: nil (this function always returns nil).
Side Effects:
  Mutates p.Info.(*FunctionInfo).Documentation.
Edge Cases & Assumptions:
  Assumes p.Info is a *FunctionInfo; a type mismatch would trigger a runtime panic.

*/
func (p *FunctionCall) SetDocumentation(docs string) error {
	p.Info.(*FunctionInfo).Documentation = docs
	return nil
}

/*
Summary: GetDocumentation returns the documentation string for this FunctionDecl by delegating to p.Info.GetDocumentation().
Use when you need the textual documentation associated with a FunctionDecl object.

Signature: func (p *FunctionDecl) GetDocumentation() (string, error)

Parameters: none (method receiver only; p is the receiver).

Returns: (string, error) where the string is the documentation and error is non-nil on failure.
The return values mirror the results of p.Info.GetDocumentation().

Errors/Exceptions: Propagates any error returned by p.Info.GetDocumentation(). May panic if p.Info is nil.

Side Effects: None; this method does not modify state.

Edge Cases & Assumptions: Assumes p.Info is non-nil and implements GetDocumentation() (string, error).

*/
func (p *FunctionDecl) GetDocumentation() (string, error) {
	return p.Info.GetDocumentation()
}

/*

*/
/*
Summary: GetDocumentation returns the documentation for this FunctionCall by delegating to p.Info.GetDocumentation().
Signature: func (p *FunctionCall) GetDocumentation() (string, error)
Parameters: none
Returns: (string, error) where the string is the documentation text and error indicates failure.
Errors/Exceptions: Any error returned by p.Info.GetDocumentation() will be returned.
Side Effects: none
Edge Cases & Assumptions: Requires p.Info != nil; behavior depends on p.Info.GetDocumentation().

*/
func (p *FunctionCall) GetDocumentation() (string, error) {
	return p.Info.GetDocumentation()
}

/*

*/
/*

*/
func (p *PackageNode) GetFunctionDecls() []mst.FunctionDecl {
	return p.FunctionDecls
}

/*

*/
/*
Summary
SetFunctionDecls assigns the provided decls to p.FunctionDecls and returns nil.

Signature
func (p *PackageNode) SetFunctionDecls(decls []mst.FunctionDecl) error

Parameters
- decls: []mst.FunctionDecl; function declarations to store in p.FunctionDecls.

Returns
- error: always nil.

Errors/Exceptions
- none

Side Effects
- Mutates p.FunctionDecls by replacing it with decls.

Edge Cases & Assumptions
- If decls is nil, p.FunctionDecls is set to nil.
- No concurrency guarantees are implied; callers may need to synchronize access if used concurrently.

*/
func (p *PackageNode) SetFunctionDecls(decls []mst.FunctionDecl) error {
	p.FunctionDecls = decls
	return nil
}

/*
Summary: Returns the imports for the PackageNode.
Use this to access the map of import aliases to paths stored in the PackageNode.
Signature: func (p *PackageNode) GetImports() map[string]string
Parameters: none.
Returns: map[string]string — the internal p.Imports map (not a copy). May be nil.
Errors/Exceptions: none.
Side Effects: none, though the returned map is the internal state; mutating it will mutate the PackageNode's Imports.
Edge Cases & Assumptions: If no imports were set, the returned value may be nil; the function does not allocate a new map.

*/
func (p *PackageNode) GetImports() map[string]string {
	return p.Imports
}

/*
Summary: Adds an entry to p.Imports, initializing the map when needed.
Signature: func (p *PackageNode) AddToImports(short string, fqn string) error
Parameters:
  short string: short key for the import alias.
  fqn string: fully-qualified name to associate with short.
Returns:
  error: always nil in current implementation.
Side Effects:
  Mutates p.Imports; if nil, initializes with make(map[string]string) and then sets p.Imports[short] = fqn.
Edge Cases & Assumptions:
  If short already exists, its value will be overwritten with fqn.
  p must be non-nil; calling on a nil receiver will panic.

*/
func (p *PackageNode) AddToImports(short string, fqn string) error {
	if p.Imports == nil {
		p.Imports = make(map[string]string)
	}

	p.Imports[short] = fqn

	return nil
}

/*

*/
/*

*/
func (p *PackageNode) GetTypeDefs() []mst.TypeDefinition {
	return p.TypeDefs
}

/*

*/
/*

*/
func (p *PackageNode) SetTypeDefs(in []mst.TypeDefinition) error {
	p.TypeDefs = in
	return nil
}

/*

*/
/*
Summary:
Return the resolved package name for this PackageNode by returning the ID field. Use this when you need the canonical package name associated with the node.

Signature:
func (p *PackageNode) GetResolvedPackageName() string

Parameters:
- p: *PackageNode — the receiver instance from which to obtain the ID. No additional parameters.

Returns:
- string: the value of p.ID, representing the resolved package name.

Errors/Exceptions:
- None. This function does not return an error. Note: if p is nil, calling this method will panic due to dereferencing a nil pointer.

Side Effects:
- None.

Edge Cases & Assumptions:
- Assumes a non-nil receiver; nil receiver will cause a panic.
- If p.ID is empty, the function returns "".

*/
func (p *PackageNode) GetResolvedPackageName() string {
	return p.ID
}

/*
Summary: SetResolvedPackageName is a no-op setter; it accepts name but does not store or apply it to PackageNode.
Signature: func (p *PackageNode) SetResolvedPackageName(name string)
Parameters:
- name: string, input to set (ignored).
Edge Cases & Assumptions:
- The input is ignored; the function does not mutate p.
- The local statement _ = name exists to silence the unused parameter.
- The commented line // return p.PkgPath hints at a possible intended behavior, but it is not active.

*/
func (p *PackageNode) SetResolvedPackageName(name string) {
	_ = name
	return
	// return p.PkgPath
}

/*

*/
/*
Summary:
Returns the package path stored in the PackageNode (p.PkgPath).

Signature:
func (p *PackageNode) GetPath() string

Returns:
string - the package path value from p.PkgPath

Side Effects:
none

Edge Cases & Assumptions:
- If PkgPath is empty, returns "".
- Called on a non-nil *PackageNode.

*/
func (p *PackageNode) GetPath() string {
	return p.PkgPath
}

/*
Summary: SetPath is not supported for PackageNode; it always returns an error indicating that paths cannot be set on golang packages.
Use when attempting to set a path on a golang package to obtain a clear failure.

Signature: func (p *PackageNode) SetPath(s string) error

Parameters:
  s: string - the path to set; its value is included in the error message.

Returns:
  error - always non-nil. The error value is fmt.Errorf("do not set path %v on golang packages", s).

Errors/Exceptions:
  Always returns the above error; no other error conditions.

Side Effects:
  None. No mutation of p or other state, and no I/O occurs.

Edge Cases & Assumptions:
  s may be any string; the method ignores it beyond formatting in the error message.

*/
func (p *PackageNode) SetPath(s string) error {
	return fmt.Errorf("do not set path %v on golang packages", s)
}

/*
Summary: Prevents setting a current file path on Go package nodes; the operation is disallowed.
         This method always returns an error to signal misuse for Go packages.
Signature: func (p *PackageNode) SetCurrentFile(file string) error
Parameters:
- file: string — the path to set (ignored); path is not accepted for Go packages.
Returns:
- error: non-nil; always fmt.Errorf("don't set path on go packages; I got it")
Errors/Exceptions:
- always returns the above error to indicate that setting a path on Go packages is not allowed
Side Effects:
- no state mutation occurs; the function immediately returns an error
Edge Cases & Assumptions:
- The error message is constant and does not depend on input; this method never changes internal state.

*/
func (p *PackageNode) SetCurrentFile(file string) error {
	return fmt.Errorf("don't set path on go packages; I got it")
}

/*
Summary: PrettyPrint is a method on *PackageNode intended to pretty-print the node. The current implementation is a no-op placeholder.

Signature: func (p *PackageNode) PrettyPrint(string)

Parameters:
- string: unnamed string parameter; currently unused by the implementation.

Returns:
- none

Errors/Exceptions:
- none

Side Effects:
- none

Edge Cases & Assumptions:
- No input validation is performed.
- The method body is empty; future implementations may perform formatting or output when provided a string context.

*/
func (p *PackageNode) PrettyPrint(string) {

}

/*
Summary: PopulatePackageInformation gathers and stores package metadata by walking each AST in p.Syntax. For each AST, it sets the current file, logs progress, and delegates to AddToImportMap, AddToTypeDefinitions, and AddToFunctionDeclarations to collect imports, type definitions, and function declarations needed by MST. It finishes by listing defined functions.

Signature: func (p *PackageNode) PopulatePackageInformation() error

Parameters:
- p: *PackageNode - receiver; uses p.Syntax and p.CompiledGoFiles to drive processing.

Returns:
- error: non-nil if any of the component steps fail; otherwise nil.

Errors/Exceptions:
- Propagates errors from AddToImportMap as "failed to add to import map: %v".
- Propagates errors from AddToTypeDefinitions as "failed to expand type definitions: %v".
- Propagates errors from AddToFunctionDeclarations as "failed to expand function definitions: %v".

Side Effects:
- Mutates p.CurrentFile, p.Imports, p.TypeDefs, and p.FunctionDecls.
- Logs progress with log.Debugf.

Edge Cases & Assumptions:
- Assumes p.Syntax and p.CompiledGoFiles are aligned in length; if p.Syntax is empty, the function completes with nil.
- p.Imports is initialized by AddToImportMap if it is nil.
- The function relies on AddToImportMap to handle aliased vs non-aliased imports and to resolve default names for non-aliased imports.
- Type definitions may contain duplicates due to non-deduplicating behavior described in AddToTypeDefinitions.

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
Summary: Populates p.Imports by extracting import declarations from f_ast. For aliased imports, uses the alias as the key; for non-aliased imports, resolves the default package name via build.Import and uses that as the key, mapping to the import path.
Signature: func (p *PackageNode) AddToImportMap(f_ast *ast.File) error
Parameters:
- p: *PackageNode - the receiver whose Import map will be populated.
- f_ast: *ast.File - the AST of the file whose imports are to be processed.
Returns:
- error: non-nil if, for any non-aliased import, the default package name cannot be resolved via build.Import; otherwise nil.
Errors/Exceptions:
- Returns fmt.Errorf("failed to build imports: %v", err) when build.Import fails for a non-aliased import.
Side Effects:
- Modifies p.Imports (initializes it if nil, then adds entries for each import).
- May perform I/O via build.Import to determine default import names.
Edge Cases & Assumptions:
- If f_ast.Imports contains imports with imp.Name != nil, the alias is used as the key and the Path is trimmed of surrounding quotes.
- If an import has no alias, the function relies on build.Import(path, "", build.ImportComment) to determine the default import name; returns an error on failure.
- If p.Imports is nil, it will be initialized to an empty map before processing.
- The function assumes f_ast is non-nil and that f_ast.Imports may be empty.

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
Summary: Traverse the AST rooted at f and append a TypeDef for each *ast.TypeSpec found to p.TypeDefs.
It creates TypeDef{Name: fd.Name.Name} for every TypeSpec encountered during the inspection.
Signature: func (p *PackageNode) AddToTypeDefinitions(f ast.Node) error
Parameters:
  p: *PackageNode - receiver; the target object to update with discovered type definitions.
  f: ast.Node - root of the AST subtree to inspect for TypeSpec declarations.
Returns:
  error: always nil (no error conditions are produced by this function).
Errors/Exceptions:
  None; the function always returns nil.
Side Effects:
  Mutates p.TypeDefs by appending new *TypeDef entries corresponding to TypeSpec nodes.
Edge Cases & Assumptions:
  Assumes f is a non-nil ast.Node representing the root of an AST to inspect.
  Does not deduplicate; identical TypeSpec names may yield duplicate TypeDef entries.
  Uses fd.Name.Name to populate TypeDef.Name; assumes TypeSpec and its Name are well-formed identifiers.

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
Summary: Collects all function declarations from the provided *ast.File, resolves their
         function invocations, constructs corresponding mst.FunctionDecls, associates the
         collected Calls with each function, and appends them to p.FunctionDecls for MST usage.
Signature: func (p *PackageNode) AddToFunctionDeclarations(f *ast.File) error
Parameters:
  - p: *PackageNode — receiver; the target node to augment with function declarations.
  - f: *ast.File — AST root to scan for function declarations; assumed non-nil.
Returns:
  - error: always nil in current implementation (no error path is produced).
Errors/Exceptions: None.
Side Effects: Mutates p.FunctionDecls by appending new FunctionDecls; may mutate MST state via
              CreateFunctionDecl and related calls; reads f and internal AST/type information.
Edge Cases & Assumptions: Assumes f != nil; nil input would lead to runtime panic in ast.Inspect.
                        For each *ast.FuncDecl, function invocations are collected via p.GetFunctionInvocations
                        and transformed into FunctionCall via p.CreateFunctionCall (skipping nil results).
                        If an existing FunctionInfo with the same full name exists, it may be reused by
                        CreateFunctionDecl.

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
Summary: Collects all function call expressions (ast.CallExpr) within the subtree rooted at the provided ast.Node f and returns them as a slice. Use this to analyze invocation sites within a given AST fragment.
Signature: func (p *PackageNode) GetFunctionInvocations(f ast.Node) ([]*ast.CallExpr, error)
Parameters:
  - f: ast.Node - the root AST node to search under (the subtree to inspect).
Returns:
  - []*ast.CallExpr: the found call expressions in the subtree.
  - error: always nil in current implementation (no error path).
Errors/Exceptions: None (the function does not produce errors).
Side Effects: None.
Edge Cases & Assumptions: If f is nil, the result is an empty slice. The receiver p is unused in this implementation. The function collects all nested *ast.CallExpr nodes, including method and function calls.

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
Summary:
CreateFunctionCall builds a FunctionCall from an ast.CallExpr for a PackageNode.
It populates FunctionCall and its FunctionInfo based on the kind of function being called
(internal, package, or object call) and wires in existing FunctionInfo when available.
Use it when you need to translate a Go AST call expression into the runtime FunctionCall/FunctionInfo
representation used by the MST system.
Signature:
func (p *PackageNode) CreateFunctionCall(fun *ast.CallExpr) mst.FunctionCall
Parameters:
- p: *PackageNode
    The receiver containing MST, TypesInfo, and Imports used to resolve function information.
- fun: *ast.CallExpr
    The AST node representing the function call to analyze and wrap as a FunctionCall.
Returns:
- mst.FunctionCall
    A FunctionCall with its Info populated to reflect the call. Returns nil for certain
    non-call expressions (see edge cases below).
Errors/Exceptions:
- log.Fatalf is invoked for unexpected fun.Fun types (panic-like fatal error) if the switch covers a
  type not handled by the implementation.
Side Effects:
- May mutate p.MST by adding or reusing a FunctionInfo entry.
- Sets fCall.Info to the computed fInfo and returns the constructed fCall.
- Reads p.TypesInfo, p.Imports, and other AST/type information during resolution.
Edge Cases & Assumptions:
- Handles these cases:
    * *ast.Ident: internal function call (mst.InternalCall) with package-less resolution.
    * *ast.SelectorExpr: could be a package call (pkg.*) or an object method call.
      - If X names a package (uses map yields a *gtypes.PkgName): mst.PackageCall
        and resolves the package via p.Imports and the package name.
      - Otherwise: mst.ObjectCall
        resolves the object name and, if possible, its package path; otherwise stores
        object type name as the object name.
- Type casts via *ast.ArrayType and *ast.ParenExpr immediately return nil (no FunctionCall created).
- If an existing FunctionInfo with the same full name exists in MST, it will be reused to avoid duplication.
- When resolving selectors on object calls, if the receiver is a pointer, it is dereferenced for naming.
- If resolved package information is not found in cased paths, GetResolvedPkg may remain empty.

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
Summary: Creates a mst.FunctionDecl from the given *ast.FuncDecl within the PackageNode.
It initializes a FunctionInfo, optionally reuses an existing FunctionInfo from the MST,
associates the declaration, registers it in the MST, and returns the created FunctionDecl.
Use this when converting an AST function declaration into the internal representation and wiring it
into the package's MST bookkeeping.

Signature: func (p *PackageNode) CreateFunctionDecl(f *ast.FuncDecl) mst.FunctionDecl

Parameters:
- f: *ast.FuncDecl — the function declaration to convert; must be non-nil.

Returns:
- mst.FunctionDecl — the constructed function declaration, with Info and Declaration linked.

Errors/Exceptions:
- Does not return an error value; may panic if f is nil or if internal structures are not initialized.

Side Effects:
- Mutates the MST function map via p.GetMST().AddToFunctionMap(...).
- Mutates fInfo (Documentation, File, HasDocumentation, Package) and fDecl.
- Sets fInfo.Declaration to the created FunctionDecl.
- Calls fInfo.SetFile(p.CurrentFile) to record the current file context.

Edge Cases & Assumptions:
- Assumes f != nil; passing nil will panic.
- If the function has a receiver that resolves to a named type, IsMemberFunction is used to populate fInfo.Object with the receiver name.
- If an existing FunctionInfo with the same full name already exists in the MST, it is reused
  (fInfo is replaced with the existing entry, with updated Documentation and File) before finalizing.
- The function relies on the receiver resolution order described in the code: first via function object,
  then via syntax-based resolution through p.TypesInfo.

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
Summary
IsMemberFunction reports whether the given *ast.FuncDecl fd is a method with a receiver that is a named type, returning the receiver as *gtypes.Named and a boolean indicating success. Use it to determine the concrete receiver type of a member function.

Signature
func (p *PackageNode) IsMemberFunction(fd *ast.FuncDecl) (*gtypes.Named, bool)

Parameters
- fd: *ast.FuncDecl — the function declaration to analyze; must have a receiver to be considered a member function.

Returns
- *gtypes.Named: the receiver's named type when fd is a member method whose receiver resolves to a named type; nil otherwise.
- bool: true if a named receiver type was found, false otherwise.

Errors/Exceptions
- Returns (nil, false) if fd is nil, has no receiver, or the receiver cannot be resolved to a named type.

Side Effects
- No external I/O or state changes beyond read-only use of p.TypesInfo.

Edge Cases & Assumptions
- If the receiver type is a pointer to a named type, the pointer is unwrapped to obtain the underlying named type.
- Resolution occurs via two strategies:
  - Preferred: use go/types from the function object. The code path checks p.TypesInfo.Defs[fd.Name] for a *gtypes.Func and inspects its Signature for a receiver.
  - Fallback: peel syntax wrappers (via baseRecvExpr) and resolve through p.TypesInfo.Uses for *gtypes.TypeName, then extract a *gtypes.Named if available.
- The function returns (nil, false) when the receiver is not a named type or resolution fails.

Notes
- Explicitly follows the internal resolution order documented in the code:
  "Preferred: use go/types from the function object." and
  "Fallback: peel syntax and resolve via info.Uses."

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
Summary: Updates in-file documentation for FunctionDecls in the PackageNode by inserting the function's documentation text into the source file at the start position of the function's AST node. Only FunctionDecls that are marked as documented in this pass and that currently lack a doc comment in the AST are processed.

Signature: func (p *PackageNode) UpdateDocsInFile() error

Parameters:
- none

Returns:
- error: non-nil if inserting documentation into any file fails; nil on success

Errors/Exceptions:
- Returns an error if insertIntoFile fails.

Side Effects:
- Writes changes to the underlying source files by inserting documentation text.

Edge Cases & Assumptions:
- Iterates FunctionDecls in reverse order; only those with GetInfo().GetDocumentedInThisPass() true are updated.
- If fd.Doc exists and is non-empty, the function skips that declaration.
- Documentation text is obtained via f.GetDocumentation(), with the error ignored.
- The insertion point is the start offset of the function's AST node as resolved by p.FindStartEnd(fd). The actual writing uses insertIntoFile on the file path f.GetFile().
- Insertion is performed as plain bytes, not atomically; concurrent edits may race.

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
Summary: Inserts the string insertion into the file at the given byte offset, rewriting the file in place. Use to inject data at a precise position.

Signature: func insertIntoFile(path string, offset int, insertion string) error

Parameters:
- path string: path to the target file
- offset int: byte offset at which to insert; must satisfy 0 <= offset <= len(data)
- insertion string: data to insert at offset

Returns:
- error: non-nil if an error occurs

Errors/Exceptions:
- offset out of range: when offset < 0 or offset > len(data)
- read error: if os.ReadFile fails
- write error: if os.WriteFile fails

Side Effects:
- Reads the file contents, inserts insertion at offset, and writes the result back to the same path
- Writes the entire file with permissions 0644

Edge Cases & Assumptions:
- If offset == 0, insertion is prepended; if offset == len(data), insertion is appended
- Operates on raw bytes; insertion is inserted as-is
- No atomicity; concurrent access may race
- The path must refer to an existing file read successfully; failure to read will prevent insertion

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
Summary: GetInfo returns the mst.FunctionInfo representation of the receiver.

When to use: when you need the receiver exposed as an mst.FunctionInfo interface value.

Signature: func (f *FunctionInfo) GetInfo() mst.FunctionInfo

Parameters: none.

Returns: mst.FunctionInfo - the receiver presented as mst.FunctionInfo. If f is nil, the return is a non-nil interface value with dynamic type *FunctionInfo and a nil underlying value.

Errors/Exceptions: none.

Side Effects: none.

Edge Cases & Assumptions: Assumes FunctionInfo implements mst.FunctionInfo. A nil receiver yields a non-nil interface value due to the dynamic type information.

*/
func (f *FunctionInfo) GetInfo() mst.FunctionInfo {
	return f
}

/*
Summary: Returns the package node associated with this FunctionInfo.

Use GetPackage when you need the mst.PackageNode representing the function's package.

Signature: func (f *FunctionInfo) GetPackage() mst.PackageNode

Returns: mst.PackageNode - the package node stored in f.Package.

Side Effects: none.

Edge Cases & Assumptions: Assumes f != nil; invoking this method on a nil receiver will panic due to dereferencing the receiver.

*/
func (f *FunctionInfo) GetPackage() mst.PackageNode {
	return f.Package
}

/*
Summary: GetName returns the Name field of the FunctionInfo as a string.
Signature: func (f *FunctionInfo) GetName() string
Parameters: none
Returns: string - the Name value of f
Errors/Exceptions: nil, except that calling on a nil receiver will cause a nil pointer dereference.
Side Effects: none
Edge Cases & Assumptions:
- Assumes f != nil when called.
- Name is the FunctionInfo.Name field.

*/
func (f *FunctionInfo) GetName() string {
	return f.Name
}

/*
Summary: Sets the FunctionInfo.Name field to the provided name.
Signature: func (f *FunctionInfo) SetName(name string) error
Parameters:
- name: string — the new value to assign to f.Name.
Returns:
- error — always nil.
Errors/Exceptions:
- None (the function returns nil unconditionally).
Side Effects:
- Mutates the receiver: f.Name is updated to name.
Edge Cases & Assumptions:
- name may be any string; no validation is performed.
- f must be a non-nil pointer; calling on a nil receiver will panic.

*/
func (f *FunctionInfo) SetName(name string) error {
	f.Name = name
	return nil
}

/*
Summary: Returns the fully-qualified name of a FunctionInfo. If f.Object is non-empty, the name includes the package path, the Object (receiver), and the function name; otherwise it includes only the package path and the function name.

Signature: func (f *FunctionInfo) GetFullName() string

Parameters:
- f: *FunctionInfo
  Role: receiver; the FunctionInfo instance for which GetFullName is invoked.
  Constraints: non-nil to avoid panic.

Returns:
- string: The full function name in one of these forms:
  - "<ResolvedPkg>.<Object>.<Name>" when f.Object != ""
  - "<ResolvedPkg>.<Name>" when f.Object == ""

Errors/Exceptions:
- May panic if the receiver is nil or if required fields (e.g., f.Package) are nil; the method does not guard against these cases.

Side Effects:
- None.

Edge Cases & Assumptions:
- Relies on GetResolvedPkg() to supply the package path; the exact value depends on f.ResolvedPkg and f.Package.GetPath().
- Assumes f.Name is the function's simple name and f.Object, if present, is the receiver name.

*/
func (f *FunctionInfo) GetFullName() string {
	if f.Object != "" {
		return fmt.Sprintf("%s.%s.%s", f.GetResolvedPkg(), f.Object, f.Name)
	}

	return fmt.Sprintf("%s.%s", f.GetResolvedPkg(), f.Name)
}

/*
Summary: Returns the resolved package path for a FunctionInfo. If a value has been assigned to ResolvedPkg, that value is returned; otherwise the path from f.Package.GetPath() is returned.

Signature: func (f *FunctionInfo) GetResolvedPkg() string

Parameters:
- f: *FunctionInfo
  Role: receiver; the FunctionInfo instance on which GetResolvedPkg is invoked.
  Constraints: non-nil to avoid panic.

Returns:
- string: The resolved package path. If f.ResolvedPkg != "" then that value is returned; otherwise f.Package.GetPath().

Errors/Exceptions:
- Potential panic if the receiver is nil or if f.Package is nil; the function does not guard against these.

Side Effects:
- None.

Edge Cases & Assumptions:
- When f.ResolvedPkg is "" (empty), the method uses f.Package.GetPath().
- Assumes f and f.Package are non-nil for normal operation.

*/
func (f *FunctionInfo) GetResolvedPkg() string {
	if f.ResolvedPkg == "" {
		return f.Package.GetPath()
	}
	return f.ResolvedPkg
}

/*
Summary: Sets f.ResolvedPkg to the provided in value. Use this to record the resolved package for a FunctionInfo.
Signature: func (f *FunctionInfo) SetResolvedPkg(in string) error
Parameters:
  in: string — the value assigned to f.ResolvedPkg.
Returns:
  error: always nil.
Errors/Exceptions:
  If the receiver is nil, this method will panic when accessing f.ResolvedPkg.
Side Effects:
  Mutates f.ResolvedPkg of the receiver.
Edge Cases & Assumptions:
  No validation on in; no error cases; caller must ensure a non-nil receiver.

*/
func (f *FunctionInfo) SetResolvedPkg(in string) error {
	f.ResolvedPkg = in
	return nil
}

/*
Summary: Returns f.Object, with no error.
Signature: func (f *FunctionInfo) GetObjectName() (string, error)
Returns: (string, error) where string is f.Object and error is nil.
Side Effects: none

*/
func (f *FunctionInfo) GetObjectName() (string, error) {
	return f.Object, nil
}

/*

*/
/*

*/
func (f *FunctionInfo) SetObjectName(name string) error {
	f.Object = name
	return nil
}

/*
Summary:
GetFile returns the File field of FunctionInfo.

Signature:
func (f *FunctionInfo) GetFile() string

Returns:
string - the value of f.File

Edge Cases & Assumptions:
- The receiver f must be non-nil; calling this on a nil *FunctionInfo would panic.
- The returned string may be empty if f.File is not set.

*/
func (f *FunctionInfo) GetFile() string {
	return f.File
}

/*
Summary: SetFile assigns the given file string to f.File and returns nil.
When to use: To set or update the File field of a FunctionInfo instance.
Signature: func (f *FunctionInfo) SetFile(file string) error
Parameters:
  - file string: the value to assign to f.File
Returns:
  - error: always nil
Errors/Exceptions:
  - None
Side Effects:
  - Mutates f.File on the receiver
Edge Cases & Assumptions:
  - f must be non-nil; calling SetFile on a nil receiver will panic.
  - No validation is performed on file; any string is accepted.

*/
func (f *FunctionInfo) SetFile(file string) error {
	f.File = file
	return nil
}

/*

*/
/*

*/
func (f *FunctionInfo) SetDocumentation(docs string) error {
	f.Documentation = docs
	return nil
}

/*
Summary:
GetDocumentation returns the value of f.Documentation and a nil error.

Signature:
func (f *FunctionInfo) GetDocumentation() (string, error)

Returns:
(string, error) — string is f.Documentation; error is nil

Errors/Exceptions:
Always nil.

Side Effects:
None.

Edge Cases & Assumptions:
f.Documentation may be empty; this method never returns a non-nil error.

*/
func (f *FunctionInfo) GetDocumentation() (string, error) {
	return f.Documentation, nil
}

/*
Summary: GetHasDocumentation returns whether the FunctionInfo has documentation.
When to use: To query if a FunctionInfo instance contains documentation metadata.
Signature: func (f *FunctionInfo) GetHasDocumentation() bool
Parameters:
  - f: *FunctionInfo, the receiver instance.
Returns:
  - bool: true if the FunctionInfo has documentation; false otherwise.
Errors/Exceptions: none
Side Effects: none
Edge Cases & Assumptions: Assumes f != nil; calling on a nil receiver would panic when accessing f.HasDocumentation.

*/
func (f *FunctionInfo) GetHasDocumentation() bool {
	return f.HasDocumentation
}

/*

*/
/*
Summary: Sets the HasDocumentation flag on a FunctionInfo to the given value.

Use when you want to mark or clear that the function is documented.

Signature: func (f *FunctionInfo) SetHasDocumentation(t bool) error

Parameters:
- f *FunctionInfo: the receiver instance; must be non-nil.
- t bool: the desired value for HasDocumentation.

Returns:
- error: always nil.

Errors/Exceptions: None.

Side Effects:
- Mutates f.HasDocumentation.

Edge Cases & Assumptions:
- If called on a nil receiver, the call will panic.
- No validation beyond assigning the flag.

*/
func (f *FunctionInfo) SetHasDocumentation(t bool) error {
	f.HasDocumentation = t
	return nil
}

/*

*/
/*
Summary: Reports whether this FunctionInfo has been documented in this pass.
Use when you need to know if the FunctionInfo was marked as documented during the current processing pass.
Signature: func (f *FunctionInfo) GetDocumentedInThisPass() bool
Returns: bool — the value of f.DocumentedInThisPass.
Errors/Exceptions: May panic if f is nil (dereferencing a nil pointer).
Side Effects: none
Edge Cases & Assumptions: Assumes f != nil when called; nil receiver can cause a panic.

*/
func (f *FunctionInfo) GetDocumentedInThisPass() bool {
	return f.DocumentedInThisPass
}

/*
Summary: Sets the DocumentedInThisPass flag on the FunctionInfo receiver to the provided boolean t.
Use this to mark whether the FunctionInfo has been documented in this pass.
Signature: func (f *FunctionInfo) SetDocumentedInThisPass(t bool)
Parameters:
- t: bool, role: indicates whether the function is documented in this pass.
Returns: none
Errors/Exceptions: None. If f is nil, this method will panic due to dereferencing a nil pointer.
Side Effects: Mutates f.DocumentedInThisPass.
Edge Cases & Assumptions: Assumes the receiver f is non-nil; no other side effects beyond updating that field.

*/
func (f *FunctionInfo) SetDocumentedInThisPass(t bool) {
	f.DocumentedInThisPass = t
}

/*
Summary: GetIsAiAware returns the value of the IsAiAware field for this FunctionInfo.
Signature: func (f *FunctionInfo) GetIsAiAware() bool
Returns: bool - true if IsAiAware is true; otherwise false.
Edge Cases & Assumptions: Assumes f != nil; calling this method on a nil *FunctionInfo will cause a runtime panic due to nil dereference when accessing f.IsAiAware.

*/
func (f *FunctionInfo) GetIsAiAware() bool {
	return f.IsAiAware
}

/*
Summary: Sets the FunctionInfo.IsAiAware flag to the provided value.
Use: to update the AI-awareness state of a FunctionInfo instance.
Signature: func (f *FunctionInfo) SetIsAiAware(t bool) error
Parameters: t: bool — the new value for f.IsAiAware.
Returns: error — always nil.
Errors/Exceptions: none.
Side Effects: Mutates f.IsAiAware on the receiver.
Edge Cases & Assumptions: assumes f is a valid *FunctionInfo; always returns nil.

*/
func (f *FunctionInfo) SetIsAiAware(t bool) error {
	f.IsAiAware = t
	return nil
}

/*
Summary: GetDecl returns f.Declaration, the mst.FunctionDecl associated with this FunctionInfo.

Signature: func (f *FunctionInfo) GetDecl() mst.FunctionDecl

Parameters:
- none

Returns:
mst.FunctionDecl: the value of f.Declaration

Errors/Exceptions:
- none

Side Effects:
- none

Edge Cases & Assumptions:
- If f is nil, GetDecl will panic due to dereferencing a nil receiver.
- The returned value is a copy of f.Declaration (value semantics of mst.FunctionDecl).

*/
func (f *FunctionInfo) GetDecl() mst.FunctionDecl {
	return f.Declaration
}

/*
Summary: Sets f.Declaration to the provided d and returns nil. Use this to attach a FunctionDecl to a FunctionInfo instance.

Signature: func (f *FunctionInfo) SetDecl(d mst.FunctionDecl) error

Parameters:
- f: *FunctionInfo — receiver; the instance whose Declarations field will be updated; must be non-nil.
- d: mst.FunctionDecl — the function declaration to store in f.Declaration.

Returns:
- error — always nil.

Errors/Exceptions:
- None produced by the function itself. Note: if f is nil, this will panic due to attempting to set a field on a nil pointer.

Side Effects:
- Mutates f.Declaration by assigning d.

Edge Cases & Assumptions:
- No validation is performed on d; assumes d is a valid mst.FunctionDecl.
- Caller must ensure f is non-nil to avoid runtime panic.

*/
func (f *FunctionInfo) SetDecl(d mst.FunctionDecl) error {
	f.Declaration = d
	return nil
}

/*
Summary: GetCalls returns the []mst.FunctionCall associated with this FunctionInfo by delegating to f.Declaration.GetCalls().
Signature: func (f *FunctionInfo) GetCalls() []mst.FunctionCall
Returns: []mst.FunctionCall obtained from f.Declaration.GetCalls().
Side Effects: none
Edge Cases & Assumptions: Assumes f.Declaration != nil. If f.Declaration is nil, this will panic when calling f.Declaration.GetCalls().

*/
func (f *FunctionInfo) GetCalls() []mst.FunctionCall {
    if f.Declaration == nil {
        return nil
    }

    return f.Declaration.GetCalls()
}

/*
Summary: PrettyPrint formats a FunctionInfo for human-readable output.

Signature: func (f *FunctionInfo) PrettyPrint(string)

Parameters:
  - string: unnamed string parameter; type is string; purpose not defined in this snippet.

Returns: none

Side Effects: none

Edge Cases & Assumptions:
  - This is a stub; no behavior is defined in this snippet.

*/
func (f *FunctionInfo) PrettyPrint(string) {

}

/*
Summary: GetDocInsertLocation returns the doc insert location for this FunctionInfo by delegating to the associated FunctionDecl's GetDocInsertLocation.

Signature: func (f *FunctionInfo) GetDocInsertLocation() uint

Parameters:
- none

Returns:
uint: the doc insert location as provided by f.GetDecl().GetDocInsertLocation()

Errors/Exceptions:
- none

Side Effects:
- none

Edge Cases & Assumptions:
- If f is nil, GetDecl will panic due to dereferencing a nil receiver.
- The returned value is the result of GetDocInsertLocation() on the associated mst.FunctionDecl.

*/
func (f *FunctionInfo) GetDocInsertLocation() uint {
	return f.GetDecl().GetDocInsertLocation()
}

/*
Summary: Creates a string by wrapping the provided docs text in a C-style block comment and returning the result.
When to use: Use this to generate a literal block-comment string from documentation text.
Signature: func (f *FunctionInfo) CreateComment(docs string) string
Parameters:
- docs: string; the documentation text to wrap.
Returns:
- string: the result is a string containing a C-style block comment formed by a start sequence, the docs, and an end sequence.
Side Effects:
- None
Edge Cases & Assumptions:
- If docs contains the end-comment sequence, the resulting string may prematurely terminate when embedded in code.

*/
func (f *FunctionInfo) CreateComment(docs string) string {
	return "/*" + docs + "*/"
}

/*
Summary: CreateComment returns the result of f.GetInfo().CreateComment(docs), effectively delegating the comment creation to the FunctionInfo associated with this FunctionDecl.

Signature: func (f *FunctionDecl) CreateComment(docs string) string

Parameters:
- docs: string, the documentation content to incorporate into the comment (passed through to FunctionInfo.CreateComment)

Returns:
- string: the comment produced by FunctionInfo.CreateComment(docs)

Edge Cases & Assumptions:
- Assumes the receiver f is non-nil when called.
- Returns the result of f.GetInfo().CreateComment(docs).

*/
func (f *FunctionDecl) CreateComment(docs string) string {
	return f.GetInfo().CreateComment(docs)
}

/*
Summary: Returns the FunctionInfo stored in f.Info for this FunctionCall.
Signature: func (f *FunctionCall) CreateComment(docs string) string
Parameters:
  - docs: string - content to embed within the generated comment.
Returns:
  - string: the value returned by f.GetInfo().CreateComment(docs).
Errors/Exceptions: None.
Side Effects: None.
Edge Cases & Assumptions:
  - f must be non-nil; calling on a nil *FunctionCall will panic.
  - Assumes f.Info holds a valid mst.FunctionInfo value.

*/
func (f *FunctionCall) CreateComment(docs string) string {
	return f.GetInfo().CreateComment(docs)
}

/*

*/
/*
Summary: Sets the FunctionDecl's Info field to the provided i (mst.FunctionInfo).
Use SetInfo to update the metadata associated with a FunctionDecl.

Signature: func (f *FunctionDecl) SetInfo(i mst.FunctionInfo) error

Parameters:
- i: mst.FunctionInfo - information to assign to f.Info.

Returns:
- error: always nil in the current implementation.

Errors/Exceptions:
- None returned; potential panic if called with a nil receiver (f is nil) since f.Info is accessed.

Side Effects:
- Mutates f.Info by assigning i.

Edge Cases & Assumptions:
- Assumes f is non-nil; does not perform validation on i.

*/
func (f *FunctionDecl) SetInfo(i mst.FunctionInfo) error {
	f.Info = i
	return nil
}

/*
Summary: GetInfo returns the FunctionInfo stored in this FunctionDecl (i.e., f.Info).

Signature: func (f *FunctionDecl) GetInfo() mst.FunctionInfo

Parameters:
- f: *FunctionDecl, role: receiver, constraints: non-nil receiver when calling

Returns:
- mst.FunctionInfo: the FunctionInfo associated with this FunctionDecl (the value of f.Info)

Errors/Exceptions: none

Side Effects: none

Edge Cases & Assumptions:
- Assumes the receiver f is non-nil when called.
- Returns a copy of f.Info.

*/
func (f *FunctionDecl) GetInfo() mst.FunctionInfo {
	return f.Info
}

/*
Summary: GetName returns the name of the FunctionDecl by delegating to f.Info.GetName().

Signature: func (f *FunctionDecl) GetName() string

Parameters:
- f: *FunctionDecl; receiver; the instance on which GetName is invoked; must be non-nil.

Returns:
- string: the name as returned by f.Info.GetName().

Errors/Exceptions:
- None. This method does not return an error. A nil receiver or nil f.Info would cause a panic if invoked on a nil instance.

Side Effects:
- None.

Edge Cases & Assumptions:
- Precondition: f != nil and f.Info != nil.

*/
func (f *FunctionDecl) GetName() string {
	return f.Info.GetName()
}

/*
Summary: GetFullName returns the full name of this FunctionDecl by delegating to f.Info.GetFullName(). Use this when you need the canonical, fully-qualified name of the function declaration.
Signature: func (f *FunctionDecl) GetFullName() string
Parameters:
  f: *FunctionDecl - receiver; role: function declaration instance; constraints: none
Returns:
  string - the full name as produced by f.Info.GetFullName()
Errors/Exceptions: none
Side Effects: none
Edge Cases & Assumptions: assumes f.Info is non-nil; relies on f.Info.GetFullName() to provide the correct full name

*/
func (f *FunctionDecl) GetFullName() string {
	return f.Info.GetFullName()
}

/*
GetDocInsertLocation returns the DocInsertLocation field value of f.

*/
func (f *FunctionDecl) GetDocInsertLocation() uint {
	return f.DocInsertLocation
}

/*
func (f *FunctionDecl) GetNode() *ast.FuncDecl
func (f *FunctionDecl) SetNode() *ast.FuncDecl
*/

/*

*/
/*

*/
func (f *FunctionDecl) GetCalls() []mst.FunctionCall {
    if f == nil {
        return nil
    }

    return f.Calls
}

/*

*/
/*
Summary: Sets the Calls field of f to the provided []mst.FunctionCall.
Use this method to update the list of FunctionDecl calls.
Signature: func (f *FunctionDecl) SetCalls(c []mst.FunctionCall) error
Parameters:
  c: []mst.FunctionCall — the new calls sequence to assign to f.Calls.
Returns:
  error: nil (this method always returns nil).
Side Effects:
  Mutates f.Calls to equal c; no I/O, no concurrency.
Edge Cases & Assumptions:
  If c is nil, f.Calls is set to nil. The method does not validate elements within c.

*/
func (f *FunctionDecl) SetCalls(c []mst.FunctionCall) error {
	f.Calls = c
	return nil
}

/*
Summary: Adds c to f.Calls, initializing the slice if needed.
Use when you want to record that a FunctionDecl has a new FunctionCall.
Signature: func (f *FunctionDecl) AddCall(c mst.FunctionCall) error
Parameters:
  - c: mst.FunctionCall — the function call to add to f.Calls.
Returns:
  - error: always nil.
Errors/Exceptions:
  - none
Side Effects:
  - Mutates f.Calls by appending the provided call; may allocate memory if f.Calls is nil.
Edge Cases & Assumptions:
  - f must be non-nil when this method is called; otherwise it will panic.
  - No duplicate checks; duplicates are allowed.

*/
func (f *FunctionDecl) AddCall(c mst.FunctionCall) error {
	if f.Calls == nil {
		f.Calls = make([]mst.FunctionCall, 0)
	}

	f.Calls = append(f.Calls, c)
	return nil
}

/*
Summary: Return the start and end byte offsets of the given AST node within the file represented by the PackageNode's token.FileSet. Use these offsets to slice the source corresponding to the node.

Signature: func (p *PackageNode) FindStartEnd(n ast.Node) (int, int)

Parameters:
- n: ast.Node — the AST node to locate. Constraint: may be nil; if nil, the function returns -1, -1.

Returns:
- start: int — the byte offset of n.Pos() in the file.
- end: int — the byte offset of n.End() in the file.
  Note: End() is the position after the node; safe for slicing [start:end].

Errors/Exceptions:
- Returns -1, -1 if n is nil or if the file cannot be resolved (p.Fset.File(n.Pos()) == nil).

Side Effects:
- None beyond reading p.Fset and the node positions.

Edge Cases & Assumptions:
- If n is nil or its position cannot be resolved to a file, returns -1, -1.
- End() is the position after the node; safe for slicing [start:end].

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
Summary: ToStringForAi returns the string representation suitable for AI prompts
         of the FunctionInfo's associated FunctionDecl. It delegates to
         GetDecl().ToStringForAi().

Signature: func (f *FunctionInfo) ToStringForAi() (string, error)

Parameters:
- none

Returns:
- string: the AI-friendly string representation of the FunctionDecl.
- error: propagated from the underlying FunctionDecl.ToStringForAi() call.

Errors/Exceptions:
- error is non-nil if f.GetDecl().ToStringForAi() returns an error.
- If f is nil, this method will panic due to a nil receiver when calling GetDecl().

Side Effects:
- none

Edge Cases & Assumptions:
- The returned string is produced by ToStringForAi on the associated mst.FunctionDecl.
- The underlying ToStringForAi may impose its own formatting or constraints.

*/
func (f *FunctionInfo) ToStringForAi() (string, error) {
	return f.GetDecl().ToStringForAi()
}

/*
Summary: Returns the source text for this FunctionDecl, augmented with per-call documentation embedded as inline comments for AI consumption. It reads the source file, extracts the function text via f.FindStartEnd, and for each call with non-empty documentation, inserts a comment block containing the unescaped documentation just before the call's line.

Signature: func (f *FunctionDecl) ToStringForAi() (string, error)

Returns:
- string: the function's source text with inserted documentation comments
- error: non-nil if reading the file fails (wrapped as "read file: ...")

Errors/Exceptions:
- read file: error if os.ReadFile(f.Info.GetFile()) fails
- May panic if f.Info is not a *FunctionInfo, or its Package is not *PackageNode, or f.Node is nil (per underlying assumptions in the implementation)
- If FindStartEnd returns invalid bounds (e.g., -1, -1), subsequent slicing may panic or yield invalid output

Side Effects:
- Reads the file from disk
- Produces a modified in-memory string by injecting comment blocks for calls with documentation
- Does not mutate input arguments

Edge Cases & Assumptions:
- Assumes f.Info is a *FunctionInfo containing a *PackageNode and f.Node is non-nil
- For each call, if its documentation is empty or whitespace, no insertion occurs
- UnescapeCommonChars handles only a fixed set of escape sequences; others remain unchanged
- The inserted documentation is placed before the corresponding call line (as a formatted / * docs * / comment) and may affect line numbering in the returned text

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
		fd_text = fd_text[:fc_line_no] + " /* " + docs + " */\n " + fd_text[fc_line_no:]
	}

	return fd_text, nil
}

/*
Summary: UnescapeCommonChars converts certain backslash-escaped characters in s to their literal forms.
It replaces the following two-character sequences with their actual characters:
  - \\ -> \ (backslash)
  - \" -> " (double quote)
  - \' -> ' (single quote)
  - \n -> newline
  - \t -> tab
  - \r -> carriage return
The mappings are applied via strings.NewReplacer and replacer.Replace(s).
The function has no side effects and always returns the unescaped string.
Signature: func UnescapeCommonChars(s string) string
Parameters:
  name: s
  type: string
  role: input string to unescape
  constraints: none
Returns:
  value: string
  conditions: always succeeds; returned value is the unescaped string
Errors/Exceptions: none
Side Effects: none
Edge Cases & Assumptions: only the six escape sequences above are handled; other sequences are left unchanged; replacements occur in a single pass via NewReplacer.

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
Summary:
PrettyPrint is a method on FunctionDecl that accepts a string parameter and currently has an empty body (no-op).

Signature:
func (f *FunctionDecl) PrettyPrint(string)

Parameters:
- (unnamed) string — input parameter; the signature does not declare a parameter name.

Returns:
- none

Errors/Exceptions:
- none

Side Effects:
- none

Edge Cases & Assumptions:
- Behavior is undefined beyond the no-op implementation.

*/
func (f *FunctionDecl) PrettyPrint(string) {

}

/*
Summary: Sets f.Info to the provided mst.FunctionInfo. Use this setter to update the FunctionCall's associated function information.
Signature: func (f *FunctionCall) SetInfo(i mst.FunctionInfo) error
Parameters:
- f: *FunctionCall, receiver; the instance whose Info will be updated.
- i: mst.FunctionInfo, the new function info to assign.
Returns:
- error: always nil.
Side Effects:
- Mutates f.Info by assigning i.
Edge Conditions & Assumptions:
- No validation on i; assumes i is a valid mst.FunctionInfo.
- Caller must ensure f is non-nil before calling.

*/
func (f *FunctionCall) SetInfo(i mst.FunctionInfo) error {
	f.Info = i
	return nil
}

/*
Summary: Returns the FunctionInfo stored in f.Info for this FunctionCall.
Signature: func (f *FunctionCall) GetInfo() mst.FunctionInfo
Parameters:
  - f: *FunctionCall - receiver from which to retrieve Info; non-nil.
Returns:
  - mst.FunctionInfo: the value stored in f.Info.
Errors/Exceptions: None.
Side Effects: None.
Edge Cases & Assumptions:
  - f must be non-nil; calling on a nil *FunctionCall will panic.
  - Assumes f.Info holds a valid mst.FunctionInfo value.

*/
func (f *FunctionCall) GetInfo() mst.FunctionInfo {
	return f.Info
}

/*
Summary: GetName returns the name for this FunctionCall by delegating to f.Info.GetName().

Signature: func (f *FunctionCall) GetName() string

Parameters:
- f: *FunctionCall, receiver; the instance on which GetName is invoked.

Returns:
- string: the name obtained from f.Info.GetName().

Errors/Exceptions:
- None returned. May panic if f.Info is nil (nil pointer dereference).

Side Effects:
- None.

Edge Cases & Assumptions:
- Assumes f.Info != nil; otherwise a nil dereference occurs when calling f.Info.GetName().

*/
func (f *FunctionCall) GetName() string {
	return f.Info.GetName()
}

/*
GetFullName returns the full name for this FunctionCall by delegating to f.Info.GetFullName().

*/
func (f *FunctionCall) GetFullName() string {
	return f.Info.GetFullName()
}

/*
func (f *FunctionCall) GetNode() *ast.FuncDecl
func (f *FunctionCall) SetNode() *ast.FuncDecl
*/

/*

*/
/*
Summary: GetKind returns the Kind of this FunctionCall.
Signature: func (f *FunctionCall) GetKind() mst.FunctionCallKind
Parameters: none.
Returns: mst.FunctionCallKind — the Kind value stored in f.Kind.
Errors/Exceptions: none.
Side Effects: none.
Edge Cases & Assumptions: Requires f != nil when called; a nil receiver will panic when accessing f.Kind.

*/
func (f *FunctionCall) GetKind() mst.FunctionCallKind {
	return f.Kind
}

/*

*/
/*
Summary: Sets the FunctionCall.Kind to the provided mst.FunctionCallKind and returns nil. Use this when you need to assign or update the kind of a FunctionCall instance.

Signature: func (f *FunctionCall) SetKind(k mst.FunctionCallKind) error

Parameters:
- f: *FunctionCall — the receiver; must be non-nil.
- k: mst.FunctionCallKind — the new kind value to assign to f.Kind.

Returns:
- error — always nil in this implementation.

Errors/Exceptions:
- If f is nil, the call will panic when attempting to set f.Kind.

Side Effects:
- Mutates f.Kind on the receiver.

Edge Cases & Assumptions:
- No validation of k; assumes k is a valid FunctionCallKind.
- Caller must ensure f is non-nil before invocation.

*/
func (f *FunctionCall) SetKind(k mst.FunctionCallKind) error {
	f.Kind = k
	return nil
}

/*
Summary: PrettyPrint is a no-op placeholder method on *FunctionCall for formatting the instance into a string representation; currently, it performs no operation.
Signature: func (f *FunctionCall) PrettyPrint(string)
Parameters:
  - string: unnamed parameter of type string; unused in the current implementation
Returns: none
Errors/Exceptions: none
Side Effects: none
Edge Cases & Assumptions: The parameter is unused; behavior may be defined in a future implementation.

*/
func (f *FunctionCall) PrettyPrint(string) {

}

/*
Summary: GetDocInsertLocation returns the doc insert location for the function being called by this FunctionCall, delegating to f.GetDecl().GetDocInsertLocation().

Signature: func (f *FunctionCall) GetDocInsertLocation() uint

Returns: uint — the document insertion location for the function being called.

Side Effects: none

Edge Conditions & Assumptions: assumes f.GetDecl() returns a non-nil value whose GetDocInsertLocation() yields a valid uint.

*/
func (f *FunctionCall) GetDocInsertLocation() uint {
	return f.GetDecl().GetDocInsertLocation()
}

/*
Summary: Returns the absolute offset (in bytes) of the start of the line that contains f.Node within the PackageNode's file set. It delegates to f.Package.(*PackageNode).FindLineNo(f.Node).

Signature: func (f *FunctionDecl) FindLineNo() int

Returns:
  - int: the absolute offset of the start of the line containing f.Node, or -1 when f.Node is nil (as per underlying FindLineNo behavior).

Side Effects: reads f.Package, f.Node, and the FileSet to compute the offset.

Edge Cases & Assumptions:
  - Assumes f.Package is a *PackageNode; otherwise a runtime panic occurs due to the type assertion.
  - Assumes f.Node is within the FileSet associated with f.Package.

*/
func (f *FunctionDecl) FindLineNo() int {
	return f.Package.(*PackageNode).FindLineNo(f.Node)
}

/*
Summary: Returns the absolute offset (in bytes) of the start of the line that contains the FunctionCall's Node within the PackageNode's file set. This method delegates to PackageNode.FindLineNo for f.Node.
Signature: func (f *FunctionCall) FindLineNo() int
Parameters: none
Returns: int - the absolute offset of the start of the line containing f.Node, or -1 if f.Node is nil.
Errors/Exceptions: none
Side Effects: reads f.Package and its underlying *PackageNode to compute the offset via FindLineNo, and uses f.Node.
Edge Cases & Assumptions: If f.Node is nil, returns -1 per the delegated call. Assumes f.Node is within the PackageNode's FileSet (p.Fset).

*/
func (f *FunctionCall) FindLineNo() int {
	return f.Package.(*PackageNode).FindLineNo(f.Node)
}

/*
Summary: Return the start and end byte offsets of this FunctionDecl's AST node within the source file by delegating to the PackageNode.FindStartEnd logic using f.Node.

Signature: func (f *FunctionDecl) FindStartEnd() (int, int)

Returns:
- start: int — the byte offset of f.Node.Pos() in the file.
- end: int — the byte offset of f.Node.End() in the file.
  Note: End() is the position after the node; safe for slicing [start:end].

Errors/Exceptions:
- May panic if f.Info is not *FunctionInfo, or its Package is not *PackageNode, or f.Node is nil.
- Underlying behavior: if n is nil or the file cannot be resolved, the underlying FindStartEnd may return -1, -1.

Side Effects:
- None beyond reading f.Info, f.Node, and delegating to PackageNode.FindStartEnd.

Edge Cases & Assumptions:
- Assumes f.Info is a *FunctionInfo containing a *PackageNode and f.Node is non-nil; otherwise a runtime panic may occur.
- If f.Node is nil, the underlying call yields -1, -1 per the underlying contract.

*/
func (f *FunctionDecl) FindStartEnd() (int, int) {
	return f.Info.(*FunctionInfo).Package.(*PackageNode).FindStartEnd(f.Node)
}

/*
Summary: Return the start and end byte offsets of the FunctionCall's AST node within the file represented by its PackageNode's token.FileSet. This method delegates to PackageNode.FindStartEnd on f.Node and returns -1, -1 when the node is nil or its file cannot be resolved.

Signature: func (f *FunctionCall) FindStartEnd() (int, int)

Returns:
- start: int — the byte offset of f.Node.Pos() in the file.
- end: int — the byte offset of f.Node.End() in the file.
  Note: End() is the position after the node; safe for slicing [start:end].

Errors/Exceptions:
- Returns -1, -1 if f.Node is nil or if the file cannot be resolved (the PackageNode's Fset cannot resolve the position).

Side Effects:
- None beyond reading f.Info, f.Node, and the PackageNode's FileSet.

Edge Cases & Assumptions:
- If f.Node is nil or its position cannot be resolved to a file, returns -1, -1.
- End() is the position after the node; safe for slicing [start:end].

*/
func (f *FunctionCall) FindStartEnd() (int, int) {
	return f.Info.(*FunctionInfo).Package.(*PackageNode).FindStartEnd(f.Node)
}

/*
Summary: Returns the list of mst.FunctionCall values associated with this FunctionCall by delegating to f.Info.GetCalls().
Use when you need access to the underlying collection of FunctionCall entries for this FunctionCall.
Signature: func (f *FunctionCall) GetCalls() []mst.FunctionCall
Returns: []mst.FunctionCall — the value produced by f.Info.GetCalls().
Side Effects: none.
Edge Cases & Assumptions: This method delegates to f.Info.GetCalls(); behavior depends on the f.Info implementation.

*/
func (f *FunctionCall) GetCalls() []mst.FunctionCall {
    if f.Info == nil {
        return nil
    }

    return f.Info.GetCalls()
}

/*
Summary: GetDecl returns the declaration for the function being called by this FunctionCall, delegating to f.Info.GetDecl().

Signature: func (f *FunctionCall) GetDecl() mst.FunctionDecl

Returns: mst.FunctionDecl — the declaration of the function being called.

Side Effects: none

Edge Cases & Assumptions: assumes f and f.Info are non-nil; relies on f.Info.GetDecl() to provide the appropriate declaration.

*/
func (f *FunctionCall) GetDecl() mst.FunctionDecl {
	return f.Info.GetDecl()
}

/*
Summary: GetFile returns the file associated with this FunctionCall by delegating to f.Info.GetFile().

Signature: func (f *FunctionCall) GetFile() string

Parameters:
- f: *FunctionCall — the receiver; the FunctionCall instance on which the method is invoked.

Returns:
- string: the file path as reported by f.Info.GetFile().

Errors/Exceptions:
- None. This method does not return an error; it returns the result of f.Info.GetFile().

Side Effects:
- None.

Edge Cases & Assumptions:
- Assumes f.Info != nil; a nil Info would cause a runtime panic when invoking GetFile().
- Relies on the behavior of f.Info.GetFile(); this method simply forwards that result.

*/
func (f *FunctionCall) GetFile() string {
	return f.Info.GetFile()
}

/*
SetDocumentedInThisPass sets the DocumentedInThisPass flag on f.Info.
It forwards the value of in to f.Info.SetDocumentedInThisPass(in).

*/
func (f *FunctionCall) SetDocumentedInThisPass(in bool) {
	f.Info.SetDocumentedInThisPass(in)
}

/*
Summary: AddCall marks this FunctionCall as documented in this pass by calling f.Info.SetDocumentedInThisPass(in).
Signature: func (f *FunctionCall) AddCall(in bool)
Parameters:
  in: bool — indicates whether the FunctionCall should be considered documented in this pass.
Side Effects: Mutates f.Info by invoking SetDocumentedInThisPass on f.Info.
Edge Cases & Assumptions: Assumes f.Info is non-nil; relies on f.Info.SetDocumentedInThisPass to perform the update.

*/
func (f *FunctionCall) AddCall(in bool) {
	f.Info.SetDocumentedInThisPass(in)
}

/*

*/
/*
Summary: Returns an AI-friendly string representation of this FunctionCall's declaration by delegating to f.Info.GetDecl().ToStringForAi().
Use when you need an AI-ready textual form of the function's declaration.

Signature: func (f *FunctionCall) ToStringForAi() (string, error)

Parameters:
  receiver: f, type *FunctionCall, role: method receiver; constraints: non-nil when invoked.

Returns:
  string: AI-friendly representation of the declaration on success.
  error: non-nil if f.Info.GetDecl().ToStringForAi() fails.

Errors/Exceptions:
  Propagates the error returned by f.Info.GetDecl().ToStringForAi().

Side Effects:
  None beyond reading f.Info and the underlying declaration.

Edge Cases & Assumptions:
  Assumes f.Info != nil and f.Info.GetDecl() != nil; relies on the underlying ToStringForAi() implementation for actual formatting and error conditions.

*/
func (f *FunctionCall) ToStringForAi() (string, error) {
	return f.Info.GetDecl().ToStringForAi()
}

/*
Summary: Returns the receiver as an mst.FunctionDecl.
Use GetDecl to obtain the declaration value for a *FunctionDecl as the mst.FunctionDecl interface.
Signature: func (f *FunctionDecl) GetDecl() mst.FunctionDecl
Returns: mst.FunctionDecl representing the function declaration (the receiver).
Edge Cases & Assumptions: If called on a nil receiver, the returned interface has dynamic type *FunctionDecl with a nil value.

*/
func (f *FunctionDecl) GetDecl() mst.FunctionDecl {
	return f
}

/*
GetFile returns the file name associated with this FunctionDecl by delegating to f.Info.GetFile().

*/
func (f *FunctionDecl) GetFile() string {
	return f.Info.GetFile()
}

/*
Summary: Sets whether this FunctionDecl is documented in the current pass by delegating to f.Info.SetDocumentedInThisPass(in).

Signature: func (f *FunctionDecl) SetDocumentedInThisPass(in bool)

Parameters:
- in: bool. true to mark as documented in this pass; false to clear the flag.

Returns: None.

Errors/Exceptions: None (no return value).

Side Effects: Mutates f.Info's documented-in-this-pass state via f.Info.SetDocumentedInThisPass(in).

Edge Cases & Assumptions:
- Assumes f.Info != nil; if f.Info is nil, this will cause a runtime panic by dereferencing nil.

*/
func (f *FunctionDecl) SetDocumentedInThisPass(in bool) {
	f.Info.SetDocumentedInThisPass(in)
}

/*
GetName returns t.Name.

Signature: func (t *TypeDef) GetName() string
Returns: string - the TypeDef.Name value.

Preconditions: t != nil.
Side Effects: none.
Edge Cases & Assumptions: If t == nil, this method will panic due to nil pointer dereference.

*/
func (t *TypeDef) GetName() string {
	return t.Name
}
