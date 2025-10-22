package ast

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"go/ast"
	"go/build"
	"go/types"

	"golang.org/x/tools/go/packages"

	log "github.com/sirupsen/logrus"
	// asTypes "github.com/BlankCanvasStudio/AutoScribe/pkg/types"
)

type FunctionKind string

var FunctionMap = map[string]*FunctionInfo{}

const (
	ObjectCall   FunctionKind = "object"
	PackageCall  FunctionKind = "package"
	InternalCall FunctionKind = "internal"
)

type FunctionDecl struct {
	Info *FunctionInfo
	Node *ast.FuncDecl

	Calls []*FunctionCall
}

type FunctionCall struct {
	Info *FunctionInfo
	Node *ast.CallExpr

	Kind FunctionKind
}

type FunctionInfo struct {
	Package *PackageNode

	// Language asTypes.SupportedFormat

	Name        string
	ResolvedPkg string
	Object      string
	File        string

	Documentation string

	WasDocumented bool // Did we find documentation for it
	Documented    bool // Did we write documentation for it
	AiAware       bool

	Declaration *FunctionDecl // Where the function is declared & all that jazz
}

/*
Summary
Returns the fully-qualified name for this FunctionInfo.
If f.Object != "", the result is formatted as "Package.Object.Name"; otherwise, "Package.Name".
Signature
func (f *FunctionInfo) FullName() string
Parameters
- f: *FunctionInfo (receiver). The instance used to build the fully-qualified name.
Returns
- string: the fully-qualified name, either "Package.Object.Name" or "Package.Name".
Errors/Exceptions
- None explicitly returned. May panic if the receiver (f) is nil due to nil dereference.
Side Effects
- None.
Edge Cases & Assumptions
- Assumes the receiver f is non-nil. A nil receiver will cause a panic when accessing f.Package, f.Object, or f.Name.
- If f.Object is non-empty, the result includes f.Object; otherwise the result omits it.
- If f.Name is empty, the returned string may end with a trailing dot (e.g., "Package.Name" becomes "Package." when Name is empty).

*/
func (f *FunctionInfo) FullName() string {
	if f.Object != "" {
		return fmt.Sprintf("%s.%s.%s", f.Package, f.Object, f.Name)
	}

	return fmt.Sprintf("%s.%s", f.Package, f.Name)
}

/*
Summary
Returns the fully-qualified name for the FunctionCall by delegating to the associated FunctionInfo's FullName().

Signature
func (f *FunctionCall) FullName() string

Parameters
- f: *FunctionCall (receiver). The instance used to obtain the fully-qualified name.

Returns
- string: the fully-qualified name, as produced by FunctionInfo.FullName().

Errors/Exceptions
- None explicitly returned. A nil receiver will panic when invoked.

Side Effects
- None.

Edge Cases & Assumptions
- Assumes f is non-nil. If f.Info is nil, calling f.Info.FullName() will panic.
- The returned value reflects the logic of FunctionInfo.FullName(): either "Package.Object.Name" (when Object is non-empty) or "Package.Name" (when Object is empty).

*/
func (f *FunctionCall) FullName() string {
	return f.Info.FullName()
}

/*
Summary
Returns the fully-qualified name for this FunctionDecl by delegating to f.Info.FullName().

Signature
func (f *FunctionDecl) FullName() string

Parameters
- f: *FunctionDecl (receiver). The instance used to obtain the fully-qualified name.

Returns
- string: the fully-qualified name, either "Package.Object.Name" or "Package.Name" as produced by f.Info.FullName().

Errors/Exceptions
- None explicitly returned. May panic if the receiver (f) is nil or if f.Info is nil due to a nil dereference.

Side Effects
- None.

Edge Cases & Assumptions
- Assumes the receiver f is non-nil and f.Info is non-nil. If f.Object is non-empty, the result includes f.Object; otherwise the result omits it. If f.Info.Name is empty, the result may end with a trailing dot (consistent with FunctionInfo.FullName()).

*/
func (f *FunctionDecl) FullName() string {
	return f.Info.FullName()
}

/*
Summary: PrettyPrint prints a human-readable representation of a FunctionInfo and its recursive call graph to stdout.
Use it to inspect a FunctionInfo's metadata (Name, Object, File, Package) and its Documentation, expanding
into the called functions via f.Declaration.Calls with indentation controlled by prefix.

Signature: func (f *FunctionInfo) PrettyPrint(prefix string)
Parameters:
- prefix: string. Role: indentation and prefix for each line of output; is augmented with a tab ("\t") for nested calls.
           Constraints: can be any string; initial lines use this value directly.
Returns: none. The function has a side effect of writing formatted information to stdout.
Side Effects: writes to standard output using fmt.Println and fmt.Printf; may produce a deeply nested, indented
              printout of the function call graph through recursive calls.
Errors/Exceptions: none handled by the function. Assumes f.Declaration and its Calls (and each called.Info) are non-nil
                  where accessed. If f.Declaration is nil, the function returns after printing header information.

Edge Cases & Assumptions:
- If f.Object is empty, the first line prints "<prefix> <Name>"; otherwise prints "<prefix> <Object>.<Name>".
- If f.Documentation is non-empty, it is printed under "Documentation:".
- If f.Declaration is nil, the function returns after header output; otherwise it iterates over f.Declaration.Calls.
- Each element of f.Declaration.Calls is expected to have a non-nil called.Info with PrettyPrint; nil pointers would panic.
- The function performs a depth-first traversal and can recurse infinitely on cyclic call graphs; it assumes an acyclic,
  well-formed structure.

*/
func (f *FunctionInfo) PrettyPrint(prefix string) {
	fmt.Println("")
	fmt.Println("")

	if f.Object == "" {
		fmt.Printf("%v %v\n", prefix, f.Name)
	} else {
		fmt.Printf("%v %v.%v\n", prefix, f.Object, f.Name)
	}

	fmt.Printf("%v File: %v\n", prefix, f.File)
	fmt.Printf("%v Package: %v\n", prefix, f.Package)

	if f.Documentation != "" {
		fmt.Printf("%v Documentation:\n%v\n", prefix, f.Documentation)
	}

	if f.Declaration == nil {
		return
	}

	for _, called := range f.Declaration.Calls {
		fmt.Println("")
		fmt.Println("")
		called.Info.PrettyPrint(prefix + "\t")
	}
}

/*
Summary: PrettyPrint on FunctionDecl delegates to its embedded FunctionInfo to print a human-readable
representation of the function and its recursive call graph to stdout. Use this to inspect a FunctionDecl's
metadata via its Info and to render the call graph.

Signature: func (f *FunctionDecl) PrettyPrint(prefix string)

Parameters:
- prefix: string. Role: indentation and prefix for each line of output; initial lines use this value directly.
  The underlying PrettyPrint augments the prefix with a tab for nested calls.

Returns: none. The function prints to standard output through the underlying FunctionInfo.PrettyPrint.

Errors/Exceptions: none explicitly handled. Assumes f.Info is non-nil; if f.Info is nil, this will panic at runtime.

Side Effects: writes to standard output; may produce a deeply nested, indented representation of the function's
  metadata and its call graph.

Edge Cases & Assumptions:
- If f.Info is nil, the function will panic.
- The underlying print traverses the call graph depth-first and may recurse infinitely on cyclic graphs; the
  structure is assumed to be acyclic or well-formed to avoid this.
- Output formatting and header details come from the embedded FunctionInfo.PrettyPrint implementation.

*/
func (f *FunctionDecl) PrettyPrint(prefix string) {
	f.Info.PrettyPrint(prefix)
}

/*
Summary: Returns the source text for this FunctionDecl's containing file with
its FunctionCall docs injected as comments above the corresponding lines.
Use when you need a GPT-friendly representation of a function declaration
annotated with the calls' Documentation comments.

Signature: func (f *FunctionDecl) ToStringForGPT() (string, error)

Parameters: none

Returns:
 - string: the source text of the function's file with inserted documentation
           comments for non-empty f.Calls[i].Info.Documentation.
 - error: non-nil if the source file cannot be read; otherwise nil.

Errors/Exceptions:
 - read file: error if os.ReadFile(f.Info.File) fails.
 - Potential runtime panic if FindStartEnd() returns an invalid range (assumes
   a valid [fd_start, fd_end]).

Side Effects: reads the file at f.Info.File; builds and returns a new string
              with inserted docs (does not modify the file on disk).

Edge Cases & Assumptions:
 - If f.Calls[i].Info.Documentation is whitespace-only, that call is skipped.
 - If there are no calls with documentation, the original file text is returned.
 - Offsets (fd_start, fd_end, fc_start, fc_end, fc_line_no) are assumed to be
   valid and aligned to the same underlying file text.
 - Documentation is inserted as a line containing a C-style comment: " /* <docs> */"
   placed immediately before the line containing the FunctionCall.

*/
func (f *FunctionDecl) ToStringForGPT() (string, error) {

	// // This should only be one layer deep. We are using comments to avoid the recursion

	raw, err := os.ReadFile(f.Info.File)
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

		docs := f.Calls[i].Info.Documentation
		if strings.TrimSpace(docs) == "" {
			continue
		}
		fd_text = fd_text[:fc_line_no] + " /* " + docs + " */\n " + fd_text[fc_line_no:]
	}

	return fd_text, nil
}

/*
Summary: Returns a GPT-friendly string representation by delegating to f.Declaration.ToStringForGPT(); returns an error if f.Declaration is nil.
Signature: func (f *FunctionInfo) ToStringForGPT() (string, error)
Parameters: none
Returns:
  - string: the value returned by f.Declaration.ToStringForGPT().
  - error: non-nil if f.Declaration is nil or if the underlying ToStringForGPT() returns an error.
Errors/Exceptions:
  - error when f.Declaration == nil: same as produced by the implementation.
  - propagation of any error from f.Declaration.ToStringForGPT().
Side Effects: none (no I/O or mutation; simply returns a value or an error).
Edge Cases & Assumptions:
  - If f.Declaration == nil, ToStringForGPT() returns an error.
  - Assumes f.Declaration is a valid, initialized object with ToStringForGPT().

*/
func (f *FunctionInfo) ToStringForGPT() (string, error) {
	if f.Declaration == nil {
		return "", fmt.Errorf("can't convery %v to string for gpt. no delcaration", f.Name)
	}

	return f.Declaration.ToStringForGPT()
}

/*
Summary: Returns the documentation string for a FunctionInfo by using its own Documentation field and, if that is empty, appending the Text from fd.Doc.List (where fd := f.Declaration.Node).
Signature: func (f *FunctionInfo) GetDocumentation() (string, error)
Receiver: f *FunctionInfo
Parameters: none (uses the receiver; no explicit parameters)
Returns: (string, error) – the accumulated documentation string and nil
Errors/Exceptions: nil (the function always returns a nil error in current implementation)
Side Effects: reads f.Documentation and f.Declaration.Node.Doc (and, if applicable, fd.Doc.List to build the result)
Edge Cases & Assumptions: assumes f.Declaration and f.Declaration.Node are non-nil; if f.Documentation != "" the returned string is that value; if f.Declaration.Node.Doc is nil or f.Documentation != "" no additional lines are appended

*/
func (f *FunctionInfo) GetDocumentation() (string, error) {

	docs := f.Documentation

	fd := f.Declaration.Node

	// We read in pre-existing docs
	if fd.Doc != nil && f.Documentation == "" {
		for _, el := range fd.Doc.List {
			docs += el.Text + "\n"
		}
	}

	return docs, nil
}

type PackageNode struct {
	*packages.Package
	FunctionDeclarations []*FunctionDecl
	TypeDefinitions      []*ast.TypeSpec
	Imports              map[string]string
	CurrentFile          string
}

/*
Summary: SanityCheck performs a basic, non-mutating validity check on a PackageNode by ensuring at least one syntax tree exists; it does not propagate or report errors from p.Errors.
Signature: func (p *PackageNode) SanityCheck() error
Parameters: none
Returns: error - non-nil if len(p.Syntax) == 0, with message "No syntax trees in %v" using p.ID; otherwise nil.
Errors/Exceptions: "No syntax trees in %v" when there are no syntax trees.
Side Effects: none (the loop over p.Errors constructs errors that are discarded; no mutation or I/O occurs).
Edge Cases & Assumptions: If p.ID is empty, the message will reflect that; existing p.Errors are iterated but not propagated.

*/
func (p *PackageNode) SanityCheck() error {
	for _, err := range p.Errors {
		fmt.Errorf("Error in %v: %v", p.ID, err)
	}

	if len(p.Syntax) == 0 {
		return fmt.Errorf("No syntax trees in %v", p.ID)
	}

	return nil
}

/*
PopulatePackageInformation initializes the PackageNode by collecting imports,
type definitions, and function declarations from the package's syntax trees.
It iterates over p.Syntax, setting p.CurrentFile to the corresponding
p.CompiledGoFiles[i], and delegates to AddToImportMap, AddToTypeDefinitions, and
AddToFunctionDeclarations for each file. Use after parsing to prepare the
in-memory representation of the package.

Signature
func (p *PackageNode) PopulatePackageInformation() error

Parameters
- p: *PackageNode — the receiver; used to mutate import mappings, type definitions, and
  function declarations.

Returns
- error: non-nil if any of the processing steps fail while analyzing the syntax trees.

Errors/Exceptions
- Propagates errors from:
  - AddToImportMap: e.g., "failed to add to import map: %v"
  - AddToTypeDefinitions: e.g., "failed to expand type definitions: %v"
  - AddToFunctionDeclarations: e.g., "failed to expand function definitions: %v"

Side Effects
- Mutates p.Imports (initializes if nil), p.TypeDefinitions, and p.FunctionDeclarations.
- Updates p.CurrentFile during processing and emits debug logs.
- May allocate temporaries and rely on helper functions that inspect and transform ASTs.

Edge Cases & Assumptions
- If p.Syntax is empty, the function performs no mutations.
- p.Imports is initialized by AddToImportMap if nil.
- Each p.Syntax[i] corresponds to p.CompiledGoFiles[i].
- Assumes valid internal helpers behave as documented and that f or ast nodes are well-formed.
- This function returns nil only when all per-file processing steps succeed.

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

	for _, decl := range p.FunctionDeclarations {
		log.Debugf("defined function: %v", decl.Info.Name)
	}

	return nil
}

/*
Summary:
Populate the PackageNode.Imports map with import path mappings from f_ast. For aliased imports,
use the alias as the map key and the import path as the value. For non-aliased imports, determine
the default local package name via build.Import and store the path under that name. Initialize
p.Imports if it is nil. Return an error if resolving any import path via the build system fails.

Signature:
func (p *PackageNode) AddToImportMap(f_ast *ast.File) error

Parameters:
- p: *PackageNode — the receiver whose Imports map will be populated.
- f_ast: *ast.File — the AST of a Go source file containing import declarations.

Returns:
- error: non-nil if a required import path cannot be resolved by build.Import; nil otherwise.

Errors/Exceptions:
- Returns fmt.Errorf("failed to build imports: %v", err) when build.Import fails for a non-aliased import.

Side Effects:
- Mutates p.Imports by inserting entries for each import found in f_ast.

Edge Cases & Assumptions:
- If p.Imports is nil, it is initialized to an empty map.
- Aliased imports (imp.Name != nil) use the alias name as the key; otherwise, the default package
  name resolved by build.Import is used as the key.
- Path strings are trimmed of surrounding quotes when read from f_ast.Imports.

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
Summary: Collects all *ast.TypeSpec declarations from the given AST node and appends them to p.TypeDefinitions.
Use when you need to gather type definitions from a parsed file or subtree.
Signature: func (p *PackageNode) AddToTypeDefinitions(f ast.Node) error
Parameters:
  f: ast.Node - the AST node to traverse for TypeSpec declarations.
Returns:
  error - nil in all cases (no error is produced by this function).
Side Effects:
  Appends discovered *ast.TypeSpec pointers to p.TypeDefinitions.
Edge Cases & Assumptions:
  - If f contains no *ast.TypeSpec nodes, p.TypeDefinitions remains unchanged.
  - If f is nil, the function is a no-op.
  - All found TypeSpec nodes within the subtree of f are collected.

*/
func (p *PackageNode) AddToTypeDefinitions(f ast.Node) error {
	ast.Inspect(f, func(n ast.Node) bool {
		fd, ok := n.(*ast.TypeSpec)
		if ok {
			p.TypeDefinitions = append(p.TypeDefinitions, fd)
		}

		return true
	})

	return nil
}

/*
Summary
Scans the provided AST file f for function declarations, constructs a FunctionDecl for each,
collects the function invocations within each declaration, converts those invocations into
FunctionCall nodes, and stores the resulting FunctionDecls in p.FunctionDeclarations. Initializes
the slice if needed and uses the package's Syntax length as initial capacity.

Signature
func (p *PackageNode) AddToFunctionDeclarations(f *ast.File) error

Parameters
- p *PackageNode: package context; holds Syntax, FunctionDeclarations, and related state used during processing.
- f *ast.File: AST of the file to analyze for function declarations and their calls.

Returns
- error: non-nil if obtaining function invocations for any declaration fails; otherwise nil.

Errors/Exceptions
- Returns an error if GetFunctionInvocations(decl) fails for any function declaration.
- May panic or skip if f is nil or if internal helpers encounter unexpected input (as per helper behavior).

Side Effects
- Mutates p.FunctionDeclarations by initializing (if nil) and appending new FunctionDecls.
- For each FunctionDecl, sets its Calls field to the slice of FunctionCall nodes derived from the
  function's invocations.
- May allocate temporary slices (decl_funcs, invocations, Calls) during processing.
- Invokes GetFunctionInvocations and CreateFunctionDecl/CreateFunctionCall, which may mutate
  internal caches or state.

Edge Cases & Assumptions
- If f contains no function declarations, the method performs no additions.
- decl_funcs is populated via ast.Inspect, so all function declarations in f are discovered.
- If a function declaration yields no callable invocations, its Calls will be empty.
- CreateFunctionCall may return nil for non-call expressions (e.g., type casts); such invocations are skipped.
- Initializes p.FunctionDeclarations with capacity len(p.Syntax) to optimize allocations; if p.Syntax
  is nil or empty, capacity may be zero.
- Assumes f is a valid *ast.File; nil f will cause a runtime panic when dereferenced (as per Go rules).

*/
func (p *PackageNode) AddToFunctionDeclarations(f *ast.File) error {
	if p.FunctionDeclarations == nil {
		p.FunctionDeclarations = make([]*FunctionDecl, 0, len(p.Syntax))
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
		invocations, err := GetFunctionInvocations(decl)
		if err != nil {
			return fmt.Errorf("failed to get function invocations: %v", err)
		}

		Calls := make([]*FunctionCall, 0, len(invocations)) // Not necessarily full length. Typecasts

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

		newFuncNode.Calls = Calls

		// Save our newly declared function to the package object
		p.FunctionDeclarations = append(p.FunctionDeclarations, newFuncNode)
	}

	return nil
}

/*
Summary
Creates a FunctionDecl for the given *ast.FuncDecl on a PackageNode, populating
FunctionInfo with receiver information (if available), documentation, and
linking the result into the FunctionMap. Prefers go/types-based receiver
resolution when possible and falls back to syntactic resolution otherwise.
Use when you need an internal representation (FunctionDecl) for a method
declaration within a package.
Signature
func (p *PackageNode) CreateFunctionDecl(f *ast.FuncDecl) *FunctionDecl
Parameters
- p *PackageNode: the package context containing typing information, file context,
  and maps used to build and register the FunctionDecl.
- f *ast.FuncDecl: the function declaration to convert; expected to represent a
  method with a receiver. Must be non-nil.
Returns
- *FunctionDecl: the created declaration node, with Info.Declaration pointing back
  to the created FunctionDecl.
Errors/Exceptions
- This function does not return an error. It may return a FunctionDecl with incomplete
  information if receiver resolution fails (e.g., unnamed or non-named receivers).
Side Effects
- Updates or inserts an entry in FunctionMap keyed by fInfo.FullName().
- Mutates fInfo and fDecl, including setting fInfo.Declaration and linking to the
  created FunctionDecl. May update fInfo.Documentation and fInfo.File based on
  f.Doc and the current file context in p.CurrentFile.
Edge Cases & Assumptions
- Assumes f != nil and represents a method with a receiver.
- If MethodRecvNamed(f, p.TypesInfo) yields a named receiver, the receiver's
  name is stored in fInfo.Object; otherwise Object remains empty.
- If a pre-existing FunctionInfo exists in FunctionMap for the computed FullName,
its fields (Documentation, File, WasDocumented) may overwrite the new values
  accordingly.

*/
func (p *PackageNode) CreateFunctionDecl(f *ast.FuncDecl) *FunctionDecl {
	obj := ""

	typeName, found := MethodRecvNamed(f, p.TypesInfo)
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

		WasDocumented: f.Doc.Text() != "",
		Documented:    false,
		AiAware:       false,
	}

	// Check if we have a function definition set up in the map.
	//   This would happen if call was experienced before we saw definition
	if possibleInfo, exists := FunctionMap[fInfo.FullName()]; exists {
		docs := f.Doc.Text()
		fInfo = possibleInfo
		fInfo.Documentation = docs
		fInfo.File = p.CurrentFile
		fInfo.WasDocumented = docs != ""
	}

	fDecl := &FunctionDecl{
		Info: fInfo,
		Node: f,

		Calls: []*FunctionCall{},
	}

	fInfo.Declaration = fDecl

	FunctionMap[fInfo.FullName()] = fInfo

	return fDecl
}

/*
Summary
Constructs a FunctionCall representation for a given ast.CallExpr within a PackageNode. It determines whether the call is InternalCall, PackageCall, or ObjectCall, populates a corresponding FunctionInfo with resolved names and package/object context, and optionally caches the FunctionInfo by its fully-qualified name. Returns the created *FunctionCall, or nil if the expression is not a resolvable function call (e.g., type casts or parenthesized expressions).

Signature
func (p *PackageNode) CreateFunctionCall(fun *ast.CallExpr) *FunctionCall

Parameters
- p *PackageNode: the package context used to resolve imports, types, and cache FunctionInfo.
- fun *ast.CallExpr: the AST node representing the function call to process. May be of kinds such as an identifier or a selector expression; may also represent non-call expressions (e.g., type casts or paren-wrapped expressions) which cause nil to be returned.

Returns
- *FunctionCall: a new FunctionCall with Node: fun and Info populated to describe the call (Name, ResolvedPkg, Object, etc.). Returns nil if the expression is a type cast or a parenthesized expression that does not correspond to a function call.

Errors/Exceptions
- May terminate the process via log.Fatalf in the default switch case when fun.Fun has an unsupported type.

Side Effects
- May mutate FunctionMap by caching fInfo.FullName() -> fInfo.
- Sets fCall.Info = fInfo before returning.

Edge Cases & Assumptions
- Assumes fun is non-nil; passing nil would panic when accessing fun.Fun.
- For ast.SelectorExpr, distinguishes package calls from object calls by inspecting the underlying object via p.TypesInfo.Uses and related type information.
- For object calls, attempts to resolve the receiver and its type/name, and fills ResolvedPkg when available.
- For type casts (ast.ArrayType) orParenExpr cases, returns nil to indicate non-call expressions.

*/
func (p *PackageNode) CreateFunctionCall(fun *ast.CallExpr) *FunctionCall {
	// Make FCall and FInfo
	fInfo := &FunctionInfo{Package: p}
	fCall := &FunctionCall{Node: fun}

	// Populate FInfo and FCall so we can look things up
	switch fd := fun.Fun.(type) {
	case *ast.Ident:
		fCall.Kind = InternalCall

		fInfo.Name = fd.Name
		fInfo.ResolvedPkg = p.ID

	case *ast.SelectorExpr:
		sel, _ := fd.X.(*ast.Ident)
		obj := p.TypesInfo.Uses[sel]

		if _, isPkg := obj.(*types.PkgName); isPkg { // Then its a package call
			fCall.Kind = PackageCall

			pkg_name := types.ExprString(fd.X)

			fInfo.Name = fd.Sel.Name
			fInfo.ResolvedPkg = p.Imports[pkg_name]

		} else { // Its an object call
			fCall.Kind = ObjectCall

			selInfo := p.TypesInfo.Selections[fd]
			obj := selInfo.Obj()

			fInfo.Name = obj.Name()

			recv := selInfo.Recv()
			if p, ok := recv.(*types.Pointer); ok {
				recv = p.Elem()
			}

			if n, ok := recv.(*types.Named); ok {
				fInfo.Object = n.Obj().Name()
				if n.Obj().Pkg() != nil {
					fInfo.ResolvedPkg = n.Obj().Pkg().Path()
				}
			} else {
				fInfo.Object = types.TypeString(recv, func(*types.Package) string { return "" })
			}

			if fInfo.ResolvedPkg == "" && obj.Pkg() != nil {
				fInfo.ResolvedPkg = obj.Pkg().Path()
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
	possibleInfo, exists := FunctionMap[fInfo.FullName()]
	if exists {
		fInfo = possibleInfo
	} else {
		FunctionMap[fInfo.FullName()] = fInfo
	}

	fCall.Info = fInfo

	return fCall
}

/*
Summary
ClipCyclicGraphs iterates p.FunctionDeclarations and prunes cyclical function calls by delegating to p.ClipFunctionCycles for each function's FunctionInfo, using a fresh callStack per declaration. It mutates the underlying function-call graph via ClipFunctionCycles and returns any error encountered.

Signature
func (p *PackageNode) ClipCyclicGraphs() error

Returns
- error: non-nil if any call to ClipFunctionCycles fails for a declaration; otherwise nil.

Errors/Exceptions
- None explicitly returned by this function except those propagated from ClipFunctionCycles (wrapped with context on failure).
- Potential panic if a declaration’s Info is nil and ClipFunctionCycles dereferences it.

Side Effects
- Mutates p.FunctionDeclarations[*].Info.Calls by removing cycle-inducing entries through ClipFunctionCycles (in place modification).

Edge Cases & Assumptions
- If p.FunctionDeclarations is nil or empty, the method is a no-op.
- Assumes p != nil when called; otherwise dereferencing p would panic.
- Assumes decl.Info (FunctionInfo) is non-nil for safe operation; nil decl.Info may lead to a panic when passed to ClipFunctionCycles.
- A new callStack []string is created for each declaration to isolate cycle detection per function.

*/
func (p *PackageNode) ClipCyclicGraphs() error {
	for _, decl := range p.FunctionDeclarations {
		callStack := []string{}
		err := p.ClipFunctionCycles(decl.Info, callStack)
		if err != nil {
			return fmt.Errorf("failed to clip function cycles: %v", err)
		}
	}

	return nil
}

/*
Summary
Prunes cyclical function calls within f.Declaration.Calls by tracking a call path in callStack, removing calls that would introduce a cycle, and recursing into callee FunctionInfo for deeper pruning. The function mutates f.Declaration.Calls by removing identified cycle-inducing entries after scanning.

Signature
func (p *PackageNode) ClipFunctionCycles(f *FunctionInfo, callStack []string) error

Parameters
- p *PackageNode: the package node whose function graph is being processed.
- f *FunctionInfo: the function to prune; its Declaration must be non-nil for processing to occur.
- callStack []string: the current path of fully-qualified function names; used to detect cycles.

Returns
- error: always nil in current implementation (no error paths).

Errors/Exceptions
- None explicitly returned. If f is nil, a nil dereference will occur when accessing f.Declaration.

Side Effects
- Mutates f.Declaration.Calls by removing elements identified as cycles.

Edge Cases & Assumptions
- If f.Declaration == nil, the function returns immediately (no-op).
- Assumes f != nil when invoked; nil f would trigger a panic on f.Declaration access.
- After identifying non-cyclic calls, the function recurses with an extended callStack: append(f.FullName()).
- The removal of cycle-causing calls is performed after the loop to avoid disturbing iteration.
- Quoted behavior: "Remove the repeated node from the calls array, but don't descend it. If you descend it, all the nodes above it will be removed as well (since they've already been included in the list)"

*/
func (p *PackageNode) ClipFunctionCycles(f *FunctionInfo, callStack []string) error {
	if f.Declaration == nil {
		return nil
	}

	to_remove := []int{}
	for i, call := range f.Declaration.Calls {
		// Remove the repeated node from the calls array, but don't descend it.
		// If you descend it, all the nodes above it will be removed as well (since they've
		//   already been included in the list)
		if slices.Contains(callStack, call.FullName()) {
			to_remove = append(to_remove, i)
		} else {
			p.ClipFunctionCycles(call.Info, append(callStack, f.FullName()))
		}
	}

	for i := len(to_remove) - 1; i >= 0; i-- {
		f.Declaration.Calls = append(f.Declaration.Calls[:to_remove[i]], f.Declaration.Calls[to_remove[i]+1:]...)
	}

	return nil
}

/*
Summary:
Returns the byte-range [start, end) in the source file for the given ast.Node, using p.Fset to map positions to file offsets. If n is nil or the corresponding file cannot be found, returns -1, -1.

Signature:
func (p *PackageNode) FindStartEnd(n ast.Node) (int, int)

Parameters:
- n: ast.Node to locate within the file associated with p.Fset.

Returns:
- start int: byte offset of the node's start position within the file (inclusive).
- end int: byte offset of the node's end position within the file (exclusive).
- If the input is nil or the file cannot be found, both start and end are -1.

Errors/Exceptions:
- None emitted; failure reported via -1, -1.

Side Effects:
- None.

Edge Cases & Assumptions:
- If file := p.Fset.File(n.Pos()) is nil, returns -1, -1.
- Note: End() is the position *after* the node; safe for slicing [start:end].

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
Summary: Return the absolute byte offset of the start of the line containing the given ast.Node within the PackageNode's FileSet. Use when you need the line start position for a node.
Signature: func (p *PackageNode) FindLineNo(n ast.Node) int
Parameters:
  n: ast.Node - the AST node whose line start offset is sought; if n is nil, the function returns -1.
Returns:
  int - the absolute offset (in bytes) of the start of the line containing n.Pos(), or -1 when n is nil.
Errors/Exceptions: none (no error return). Behavior follows token.FileSet semantics.
Side Effects: none.
Edge Cases & Assumptions: assumes p.Fset is initialized and that n.Pos() can be resolved by p.Fset; if n is nil, -1 is returned immediately.

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

	// return fset.File(n.Pop.Fset.Position(n.Pos()).Line
}

/*
Summary: Returns the byte-range [start, end) in the source file for this FunctionDecl's AST node by delegating to f.Info.Package.FindStartEnd(f.Node). Use when you need the exact file offsets of a function declaration.
Signature: func (f *FunctionDecl) FindStartEnd() (int, int)
Returns:
  start int: byte offset (inclusive) of the node's start position within the file.
  end   int: byte offset (exclusive) of the node's end position within the file.
  If the node or the associated file cannot be found, both start and end are -1.
Errors/Exceptions: None emitted; failure indicated by -1, -1.
Side Effects: None.
Edge Cases & Assumptions: Assumes f.Node is non-nil and associated with a file via f.Info.Package; if not, returns -1, -1.

*/
func (f *FunctionDecl) FindStartEnd() (int, int) {
	return f.Info.Package.FindStartEnd(f.Node)
}

/*
Summary: Return the absolute byte offset of the start of the line containing this FunctionDecl's ast.Node, using the FileSet in f.Info.Package. Use when you need the line-start position for a function node.
Signature: func (f *FunctionDecl) FindLineNo() int
Parameters:
  - none
Returns:
  int - the absolute offset (in bytes) of the start of the line containing f.Node; -1 if f.Node is nil.
Errors/Exceptions:
  none (no error return)
Side Effects:
  none
Edge Cases & Assumptions:
  assumes f.Info and f.Info.Package are initialized; if f.Node is nil, the result follows the underlying FileSet behavior and may be -1.

*/
func (f *FunctionDecl) FindLineNo() int {
	return f.Info.Package.FindLineNo(f.Node)
}

/*
Summary:
Returns the byte-range [start, end) in the source file for the FunctionCall's Node by delegating to f.Info.Package.FindStartEnd(f.Node). If the node or its file cannot be resolved, returns -1, -1.

Signature:
func (f *FunctionCall) FindStartEnd() (int, int)

Parameters:
- None.

Returns:
- start int: byte offset of the node's start position within the file (inclusive).
- end int: byte offset of the node's end position within the file (exclusive).
- If the node or the corresponding file cannot be found, both start and end are -1.

Errors/Exceptions:
- None emitted; failure indicated by -1, -1.

Side Effects:
- None.

Edge Cases & Assumptions:
- If the underlying file lookup fails or f.Node is nil, returns -1, -1.
- Note: End() is the position *after* the node; safe for slicing [start:end].

*/
func (f *FunctionCall) FindStartEnd() (int, int) {
	return f.Info.Package.FindStartEnd(f.Node)
}

/*
Summary: Return the absolute byte offset of the start of the line containing this FunctionCall's Node within the PackageNode's FileSet. Use when you need the line start position for a node.
Signature: func (f *FunctionCall) FindLineNo() int
Parameters: none
Returns: int - the absolute offset (in bytes) of the start of the line containing f.Node, or -1 if f.Node is nil. Delegates to f.Info.Package.FindLineNo(f.Node).
Errors/Exceptions: none (no error return). Behavior follows token.FileSet semantics.
Side Effects: none.
Edge Cases & Assumptions: assumes f.Info.Package is initialized; if f.Node is nil, -1 is returned; assumes f.Node.Pos() can be resolved by the FileSet.

*/
func (f *FunctionCall) FindLineNo() int {
	return f.Info.Package.FindLineNo(f.Node)
}

/*
Summary:
Updates source files by appending documentation strings for function declarations that lack explicit docs. It processes FunctionDeclarations in reverse order, skips those with existing documentation, and inserts the declaration's Documentation string into its source file at the position determined by FindStartEnd.

Signature:
func (p *PackageNode) UpdateDocsInFile() error

Parameters:
- p *PackageNode: receiver; the package node whose function declarations are scanned and updated.

Returns:
- error: nil on success; non-nil if an update operation fails for any function declaration.

Errors/Exceptions:
- Returns a non-nil error if insertIntoFile fails while updating a file, with context "failed to update docs in file: %v".

Side Effects:
- Mutates the files identified by f.Info.File by inserting the documentation text at a computed byte offset within the file.

Edge Cases & Assumptions:
- Only FunctionDeclarations without existing documentation (fd.Doc == nil or empty) are updated.
- Insertion position is obtained via start, _ := p.FindStartEnd(fd); insertion may fail if start is -1.
- The inserted content is f.Info.Documentation followed by a newline.
- Large files or unusual file permissions may affect behavior via insertIntoFile.

*/
func (p *PackageNode) UpdateDocsInFile() error {

	for i := len(p.FunctionDeclarations) - 1; i >= 0; i-- {
		f := p.FunctionDeclarations[i]

		fd := f.Node

		// We read in pre-existing docs
		if fd.Doc != nil && strings.TrimSpace(fd.Doc.Text()) != "" {
			continue
		}

		start, _ := p.FindStartEnd(fd)

		log.Debugf("Documentation string: %v", f.Info.Documentation)

		toAdd := fmt.Sprintf("%v\n", f.Info.Documentation)

		err := insertIntoFile(f.Info.File, start, toAdd)
		if err != nil {
			return fmt.Errorf("failed to update docs in file: %v", err)
		}
	}

	return nil
}

/*
Summary: Returns all function invocation expressions (ast.CallExpr) found within the subtree rooted at f by traversing the AST with ast.Inspect.
Use when you need to analyze every function call within a given AST node.
Signature: func GetFunctionInvocations(f ast.Node) ([]*ast.CallExpr, error)
Parameters:
  f: ast.Node — root of the AST subtree to search; input
Returns:
  []*ast.CallExpr — all found function invocation expressions, in traversal order
  error — always nil for this implementation
Errors/Exceptions: None (function does not currently produce a non-nil error)
Side Effects: None beyond memory allocation for the result slice; does not modify f or other state
Edge Cases & Assumptions:
  - If no function invocations are present, returns an empty []*ast.CallExpr.
  - Traversal is performed via ast.Inspect over the subtree rooted at f.

*/
func GetFunctionInvocations(f ast.Node) ([]*ast.CallExpr, error) {
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
Summary: Determines and returns the receiver type as a *types.Named for a given method
declaration, if such a named receiver exists. Prefers go/types information when available
and falls back to syntax-based resolution if needed. Use when you need the concrete
named receiver type of a method.

The function inspects the method declaration fd and uses info to locate the receiver's
type. It first tries the go/types-based path, then falls back to syntactic resolution,
peeling common wrappers around the receiver expression (pointer, parens, index expressions)
to obtain a named type.

Signature: func MethodRecvNamed(fd *ast.FuncDecl, info *types.Info) (*types.Named, bool)

Parameters:
- fd: *ast.FuncDecl - the function declaration to inspect; must represent a method with a receiver.
- info: *types.Info - type information used to resolve the receiver; must be non-nil.

Returns:
- *types.Named: the receiver type if it is a named type; otherwise nil.
- bool: true if a named receiver type was found; false otherwise.

Errors/Exceptions: none returned; on failure to determine a named receiver, returns nil, false.

Side Effects: none.

Edge Cases & Assumptions:
- If fd == nil || fd.Recv == nil || len(fd.Recv.List) == 0, returns nil, false.
- If the receiver is a pointer, the pointer is unwrapped to its Elem before checking for a Named.
- If the receiver cannot be resolved to a named type (e.g., unnamed or primitive), returns nil, false.
- When info.Defs[fd.Name] provides a Func with a valid Signature, that path is preferred.
- If the preferred path fails, the function attempts to resolve via info.Uses by inspecting the
  receiver expression after peeling wrappers: *ast.StarExpr, *ast.ParenExpr, *ast.IndexExpr,
  *ast.IndexListExpr.

*/
func MethodRecvNamed(fd *ast.FuncDecl, info *types.Info) (*types.Named, bool) {
	if fd == nil || fd.Recv == nil || len(fd.Recv.List) == 0 {
		return nil, false
	}

	// Preferred: use go/types from the function object.
	if obj, ok := info.Defs[fd.Name].(*types.Func); ok && obj != nil {
		if sig, ok := obj.Type().(*types.Signature); ok && sig.Recv() != nil {
			t := sig.Recv().Type()
			if p, ok := t.(*types.Pointer); ok {
				t = p.Elem()
			}
			if n, ok := t.(*types.Named); ok {
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
		if tn, ok := info.Uses[e].(*types.TypeName); ok {
			if n, ok := tn.Type().(*types.Named); ok {
				return n, true
			}
		}
	case *ast.SelectorExpr:
		if tn, ok := info.Uses[e.Sel].(*types.TypeName); ok {
			if n, ok := tn.Type().(*types.Named); ok {
				return n, true
			}
		}
	}

	return nil, false
}

/*
ParsePackage builds in-memory representations of all Go packages under foldername.

It locates folders containing Go source files, loads package data, validates and enriches each PackageNode,
and prunes cyclic function-call graphs. Use ParsePackage when you need a complete set of PackageNode
metadata suitable for documentation or static analysis.

Signature:
func ParsePackage(foldername string) ([]PackageNode, error)

Parameters:
- foldername string: root folder to scan for Go packages. Must be accessible; errors if inaccessible.

Returns:
- []PackageNode: slice of PackageNode, one per discovered package, in processing order.
- error: non-nil if an error occurs during directory discovery, package loading, or graph processing. Some steps may return partial results along with an error (notably after SanityCheck or during PopulatePackageInformation or ClipCyclicGraphs).

Errors/Exceptions:
- error if GetNestedFoldersWithGoFiles fails to identify folders.
- error if packages.Load fails to load any package data.
- error on SanityCheck failure (returns partial pkgNodes and non-nil error).
- error on PopulatePackageInformation failure (returns partial pkgNodes and non-nil error).
- error on ClipCyclicGraphs failure (returns nil and non-nil error).

Side Effects:
- Emits debug logs via log.Debugf.
- Reads the filesystem to discover folders and load packages.
- Mutates and returns new pkgNodes; does not modify external state beyond local variables.

Edge Cases & Assumptions:
- If no Go-containing folders exist, returns an empty []PackageNode with nil error.
- If a package has no syntax trees, SanityCheck reports an error.
- Each PackageNode is populated in the same order as discovered packages and assumes a one-to-one correspondence between Syntax and CompiledGoFiles.

*/
func ParsePackage(foldername string) ([]PackageNode, error) {
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

	log.Debugf("Documenting folder: %v", foldername)

	folders, err := GetNestedFoldersWithGoFiles(foldername)
	if err != nil {
		return nil, fmt.Errorf("failed to get go directories: %v", err)
	}

	log.Debugf("Found folders with go files: %v", folders)

	pkgs, err := packages.Load(cfg, folders...)
	if err != nil {
		return nil, fmt.Errorf("failed to load package %v: %v", foldername, err)
	}

	pkgNodes := []PackageNode{}

	log.Debugf("# packages seen: %v", len(pkgs))

	for _, pkg := range pkgs {
		pkgNode := PackageNode{
			Package:              pkg,
			FunctionDeclarations: []*FunctionDecl{},
			TypeDefinitions:      []*ast.TypeSpec{},
			Imports:              make(map[string]string),
		}

		err := pkgNode.SanityCheck()
		if err != nil {
			return pkgNodes, fmt.Errorf("failed to sanity check: %v", err)
		}

		log.Debugf("Sanity check passed for `%v`", pkgNode.Name)

		err = pkgNode.PopulatePackageInformation()
		if err != nil {
			return pkgNodes, fmt.Errorf("failed to populate package information: %v", err)
		}

		log.Debugf("%v functions declared", len(pkgNode.FunctionDeclarations))

		pkgNodes = append(pkgNodes, pkgNode)
	}

	// Function call stacks can be cyclic graphs. We clip those cyclic graphs here
	for _, pkgNode := range pkgNodes {
		err = pkgNode.ClipCyclicGraphs()
		if err != nil {
			return nil, fmt.Errorf("failed to clip cyclic graphs: %v", err)
		}
	}

	return pkgNodes, nil
}

/*
Summary:
GetNestedFoldersWithGoFiles walks the given folder recursively and returns the set of unique directories that contain at least one .go file. The results are sorted lexicographically.

Signature:
func GetNestedFoldersWithGoFiles(folder string) ([]string, error)

Parameters:
- folder: string
  Root folder to search from.

Returns:
- ([]string, error)
  - []string: absolute directory paths that contain at least one .go file, sorted lexicographically.
  - error: non-nil if an error occurs during traversal or path resolution.

Errors/Exceptions:
- error if filepath.WalkDir encounters a traversal error.
- error if obtaining the absolute path for a discovered .go file fails.

Side Effects:
- Logs a debug message via log.Debugf("folders searching: %v", folder).
- Reads the filesystem; uses a seen map to deduplicate directories.

Edge Cases & Assumptions:
- Only files with extension ".go" are considered; other files are ignored.
- Directories containing multiple .go files are reported once.
- Returned paths are absolute; the final slice is sorted before returning.
- If no .go files are found, an empty slice is returned.
- If folder does not exist or is inaccessible, an error is returned from WalkDir.

*/
func GetNestedFoldersWithGoFiles(folder string) ([]string, error) {
	seen := make(map[string]struct{})

	log.Debugf("folders searching: %v", folder)

	err := filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {

		if !d.IsDir() && filepath.Ext(path) == ".go" {
			abs, err := filepath.Abs(path)
			if err != nil {
				return err
			}

			dir := filepath.Dir(abs)

			if _, ok := seen[dir]; !ok {
				seen[dir] = struct{}{}
			}
		}

		return nil
	})

	if err != nil {
		return []string{}, fmt.Errorf("failed to walk directory for go files: %v", err)
	}

	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys, nil
}
