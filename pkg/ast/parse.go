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
Returns the fully-qualified name of a FunctionInfo. If f.Object is non-empty,
the result is "Package.Object.Name"; otherwise, it is "Package.Name".

Signature
func (f *FunctionInfo) FullName() string

Parameters
- f: *FunctionInfo, receiver providing Package, Object, and Name fields; role: method receiver.
  Constraints: non-nil when invoked.
  Note: uses f.Package, f.Object, and f.Name to construct the result.

Returns
- string: the fully-qualified function name.
  Behavior:
  - if f.Object != "" => fmt.Sprintf("%s.%s.%s", f.Package, f.Object, f.Name)
  - else => fmt.Sprintf("%s.%s", f.Package, f.Name)

Errors/Exceptions
- none.

Side Effects
- none.

Edge Cases & Assumptions
- If f.Package or f.Name are empty, the returned string will contain empty segments (e.g., ".Name" or "Package."). No input validation is performed.

*/
func (f *FunctionInfo) FullName() string {
	if f.Object != "" {
		return fmt.Sprintf("%s.%s.%s", f.Package, f.Object, f.Name)
	}

	return fmt.Sprintf("%s.%s", f.Package, f.Name)
}

/*
Summary
Returns the fully-qualified name for the FunctionInfo associated with f by delegating to f.Info.FullName().

Signature
func (f *FunctionCall) FullName() string

Parameters
- f: *FunctionCall, receiver providing access to f.Info; role: method receiver.
  Constraints: non-nil when invoked.
  Note: delegates to f.Info.FullName() to compute the result.

Returns
- string: the fully-qualified function name, as produced by f.Info.FullName().

Errors/Exceptions
- none.

Side Effects
- none.

Edge Cases & Assumptions
- If f.Info is nil, this will panic at runtime. No nil-check or validation is performed.

*/
func (f *FunctionCall) FullName() string {
	return f.Info.FullName()
}

/*
Summary
Returns the fully-qualified name of a FunctionDecl by delegating to f.Info.FullName().

Signature
func (f *FunctionDecl) FullName() string

Parameters
- f: *FunctionDecl, receiver providing the Info field; role: method receiver.
  Constraints: non-nil when invoked.

Returns
- string: the fully-qualified function name as produced by f.Info.FullName().

Errors/Exceptions
- none.

Side Effects
- none.

Edge Cases & Assumptions
- If f.Info is nil, this will panic due to a nil pointer dereference when calling f.Info.FullName().

*/
func (f *FunctionDecl) FullName() string {
	return f.Info.FullName()
}

/*
Summary:
PrettyPrint prints a human-readable representation of the FunctionInfo instance and its nested call graph to standard output. It uses the provided prefix for indentation and renders the function identity, file, package, and optional documentation, then recursively prints any called functions.

Signature:
func (f *FunctionInfo) PrettyPrint(prefix string)

Parameters:
- prefix: string — indentation prefix prepended to each line of output. Used to nest nested calls visually (e.g., "\t").

Returns:
- None. The function prints directly to standard output and does not return values.

Errors/Exceptions:
- None. The function does not return errors.

Side Effects:
- Writes to standard output via fmt.Println and fmt.Printf.
- Recursively prints called functions when f.Declaration is non-nil.

Edge Cases & Assumptions:
- If f.Object == "" the function prints "<prefix> <Name>"; otherwise it prints "<prefix> <Object>.<Name>".
- Always prints "File: <f.File>" and "Package: <f.Package>".
- If f.Documentation != "" it prints "Documentation:" followed by the documentation content.
- If f.Declaration == nil, the function returns after printing the header.
- If f.Declaration.Calls contains cycles, this may lead to infinite recursion; no cycle detection is performed.
- Between sections, the function prints blank lines to improve readability; before each called function it inserts two blank lines and indented output.

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
Summary: PrettyPrint prints a human-readable representation of the FunctionDecl by delegating to its associated FunctionInfo.PrettyPrint(prefix) and thus prints the nested call graph.
Signature: func (f *FunctionDecl) PrettyPrint(prefix string)
Parameters:
- prefix: string — indentation prefix prepended to each line of output to visually nest nested calls (e.g., "\t").
Returns: none.
Errors/Exceptions: none returned; may panic if f.Info is nil since it directly calls f.Info.PrettyPrint(prefix).
Side Effects: Writes output to standard output via the delegated PrettyPrint call.
Edge Cases & Assumptions: Assumes f.Info is non-nil; behavior and formatting are determined by FunctionInfo.PrettyPrint. No cycle handling is performed by this wrapper.

*/
func (f *FunctionDecl) PrettyPrint(prefix string) {
	f.Info.PrettyPrint(prefix)
}

/*
Summary: Returns a GPT-friendly string representation of the FunctionDecl's source slice, annotated with per-call documentation. It reads the source file from f.Info.File, extracts the portion corresponding to the FunctionDecl, and injects any non-empty Documentation from each FunctionCall into the slice text at the line where that call is located. Use it when you need the function's source context with inline docs for GPT processing.

Signature: func (f *FunctionDecl) ToStringForGPT() (string, error)

Parameters:
  - name: f
    type: *FunctionDecl
    role: receiver

Returns:
  - string: The annotated source text for the FunctionDecl, with inlined documentation for calls.
  - error: Non-nil if reading the source file at f.Info.File fails (the error is wrapped as "read file: %w").

Errors/Exceptions:
  - read file: <err> if os.ReadFile(f.Info.File) fails.

Side Effects:
  - Reads the file f.Info.File. No mutation of external state occurs.

Edge Cases & Assumptions:
  - Assumes f.Info.File is a valid, readable path and f.FindStartEnd() returns valid offsets into that file. If offsets are invalid, behavior is undefined.
  - Assumes f.Node, f.Info.Package, and f.Calls are initialized; if a call has empty Documentation, nothing is injected for that call.
  - Insertion order is handled by iterating in reverse to preserve correct offsets when injecting docs as line-prefixed comments.

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
Summary: Returns a GPT-friendly string representation for this FunctionInfo by delegating to f.Declaration.ToStringForGPT(). Use when you need the textual GPT input corresponding to the function's declaration.
Signature: func (f *FunctionInfo) ToStringForGPT() (string, error)
Parameters:
  - name: f
    type: *FunctionInfo
    role: receiver
Returns:
  - string: The result of f.Declaration.ToStringForGPT().
  - error: Non-nil if the underlying call returns an error, or if f.Declaration is nil.
Errors/Exceptions:
  - "can't convery %v to string for gpt. no delcaration" when f.Declaration == nil (formatted with f.Name).
  - Propagates any error returned by f.Declaration.ToStringForGPT().
Side Effects:
  - None beyond possible evaluation of f.Declaration.ToStringForGPT().
Edge Cases & Assumptions:
  - Assumes f.Declaration may be nil; in that case, an error is returned.
  - Assumes f.Declaration.ToStringForGPT() handles its internal edge cases.

*/
func (f *FunctionInfo) ToStringForGPT() (string, error) {
	if f.Declaration == nil {
		return "", fmt.Errorf("can't convery %v to string for gpt. no delcaration", f.Name)
	}

	return f.Declaration.ToStringForGPT()
}

/*
Summary: Returns the documentation text for a FunctionInfo, preferring the pre-existing
Documentation and augmenting it with the AST node's doc comment when available.

Signature: func (f *FunctionInfo) GetDocumentation() (string, error)

Parameters:
- receiver: f *FunctionInfo — method receiver; operates on a FunctionInfo instance.
- explicit parameters: none.

Returns:
- (string, error): the assembled documentation text; error is always nil.

Errors/Exceptions:
- nil error is always returned.

Side Effects:
- Reads f.Documentation and f.Declaration.Node; does not modify input or global state.

Edge Cases & Assumptions:
- Assumes f.Declaration.Node is non-nil; otherwise a nil pointer dereference occurs when accessing fd.Doc.
- If f.Documentation is non-empty, its value is returned as the base docs.
- If f.Documentation is empty and fd.Doc != nil, each element el.Text in fd.Doc.List is appended with a trailing newline to form the docs.
- If both f.Documentation is empty and fd.Doc is nil, an empty string is returned.

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
Summary: SanityCheck performs a basic sanity check on a PackageNode.
It iterates over p.Errors to format an error message for each entry, but the
resulting error values are not used or returned. It then verifies that at least
one syntax tree is present.

Signature: func (p *PackageNode) SanityCheck() error

Parameters: none. This is a method on *PackageNode.

Returns: error. Non-nil if len(p.Syntax) == 0 (with message "No syntax trees in %v");
otherwise returns nil.

Errors/Exceptions: Returns fmt.Errorf("No syntax trees in %v", p.ID) when there are no
syntax trees. The error values created inside the loop are discarded and not returned.

Side Effects: None on p. The loop creates error values that are not used.

Edge Cases & Assumptions: If p.Syntax is nil or empty, the function returns an error.
The error message uses p.ID; if p.ID is empty, the message includes an empty identifier.

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
Summary: PopulatePackageInformation builds the package-wide data by processing each AST file.
For each element in p.Syntax, it sets p.CurrentFile to the corresponding p.CompiledGoFiles[i] and calls
AddToImportMap(syn_ast), AddToTypeDefinitions(syn_ast), and AddToFunctionDeclarations(syn_ast) to
collect imports, type definitions, and function declarations. Returns an error on failure.
Signature: func (p *PackageNode) PopulatePackageInformation() error
Parameters: none
Returns: error — non-nil if per-file processing fails; nil on success.
Errors/Exceptions: returns fmt.Errorf("failed to add to import map: %v", err) if AddToImportMap fails,
                  fmt.Errorf("failed to expand type definitions: %v", err) if AddToTypeDefinitions fails,
                  fmt.Errorf("failed to expand function definitions: %v", err) if AddToFunctionDeclarations fails.
Side Effects: mutates p.CurrentFile, p.Imports, p.TypeDefinitions, and p.FunctionDeclarations; logs progress.
Edge Cases & Assumptions: assumes aligned iteration between p.Syntax and p.CompiledGoFiles; relies on helper
functions to handle their internal edge cases; stops processing on first error.

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
Summary: Populates p.Imports from the imports in f_ast. For aliased imports, uses the alias as the key and the unquoted path as the value; for non-aliased imports, resolves the default import name with build.Import and uses that name as the key and the path as the value. Initializes p.Imports if nil and returns an error if import resolution fails.
Signature: func (p *PackageNode) AddToImportMap(f_ast *ast.File) error
Parameters:
- p *PackageNode: receiver; the node to populate.
- f_ast *ast.File: input AST file whose imports are processed.
Returns: error; non-nil if import resolution fails.
Errors/Exceptions: returns fmt.Errorf("failed to build imports: %v", err) when build.Import fails.
Side Effects: initializes p.Imports if nil; updates p.Imports with import name/path mappings derived from f_ast.Imports.
Edge Cases & Assumptions:
- If imp.Name != nil, p.Imports[imp.Name.Name] = strings.Trim(imp.Path.Value, "\"") is used.
- If imp.Name == nil, path is unquoted with strings.Trim(imp.Path.Value, "); build.Import(path, "", build.ImportComment) provides the default import name (pkg.Name), which becomes the key for path.
- On any error from build.Import, the function returns the wrapped error and stops processing further imports.

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
Summary: Walks the provided AST node f and collects all TypeSpec declarations by appending them to p.TypeDefinitions.

Signature: func (p *PackageNode) AddToTypeDefinitions(f ast.Node) error

Parameters:
- f: ast.Node — the root of the AST subtree to inspect for TypeSpec declarations (via ast.Inspect).

Returns:
- error — always nil in the current implementation.

Side Effects:
- Appends each found *ast.TypeSpec to p.TypeDefinitions.

Edge Cases & Assumptions:
- If f contains no TypeSpec nodes, p.TypeDefinitions remains unchanged.
- TypeSpec nodes are added in the order encountered during ast.Inspect traversal.
- p.TypeDefinitions may be nil prior to calls; append handles nil slices gracefully.

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
Summary:
Collects function declarations from an AST file and builds corresponding FunctionDecl entries in the PackageNode. For each function, it collects its function invocations, creates FunctionCall nodes, associates them with the function declaration, and appends the resulting FunctionDecl to p.FunctionDeclarations.

Signature:
func (p *PackageNode) AddToFunctionDeclarations(f *ast.File) error

Parameters:
- p: *PackageNode — receiver; the package node to populate. Mutates p.FunctionDeclarations.
- f: *ast.File — AST file to inspect for function declarations and their bodies. Must be non-nil.

Returns:
- error: non-nil if processing fails (e.g., GetFunctionInvocations fails for a declaration). Returns nil on success.

Errors/Exceptions:
- Returns an error if GetFunctionInvocations(decl) fails for any function declaration.
- May panic if f is nil (precondition not stated in code).

Side Effects:
- Initializes p.FunctionDeclarations if nil.
- Appends new FunctionDecl entries to p.FunctionDeclarations.
- May mutate internal state via CreateFunctionDecl and CreateFunctionCall (e.g., updating FunctionMap and FunctionInfo associations).
- Reads and relies on p.Syntax, and interacts with f.

Edge Cases & Assumptions:
- If no FuncDecls are present in f, the function performs no work.
- If a function has no invocations, the corresponding Calls slice is empty but attached to the FunctionDecl.
- CreateFunctionDecl may reuse an existing FunctionInfo from FunctionMap based on FunctionInfo.FullName().
- f is assumed to be a valid *ast.File; nil handling is not defined beyond typical usage.

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
Summary:
Creates a FunctionDecl for the given ast.FuncDecl. It determines the receiver's named type (if any) using MethodRecvNamed, builds a FunctionInfo, optionally replaces it with a previously-known entry from FunctionMap, and stores the resulting FunctionDecl in memory for later reference. Use when you need to register or retrieve metadata about a function declaration within a PackageNode.

Signature:
func (p *PackageNode) CreateFunctionDecl(f *ast.FuncDecl) *FunctionDecl

Parameters:
- f: *ast.FuncDecl — the function declaration to process; must be non-nil. The function reads f.Name, f.Doc, and may inspect the receiver to determine the associated named type.

Returns:
- *FunctionDecl — the created declaration object, with Info and Node populated, and Declaration set to the returned FunctionDecl. The function also populates or updates FunctionMap with the resulting FunctionInfo.

Errors/Exceptions:
- None returned. Note: if f is nil, the method will panic due to dereferencing f.

Side Effects:
- Mutates FunctionMap by inserting/updating an entry for fInfo.FullName().
- Mutates fInfo by assigning its Declaration to the created fDecl.
- May overwrite fInfo with a previously stored FunctionInfo from FunctionMap if a matching FullName() exists.
- Reads p.CurrentFile and f.Doc to populate documentation fields.

Edge Cases & Assumptions:
- f may omit a receiver; in that case, Object remains "".
- The receiver type resolution relies on MethodRecvNamed, and if a named receiver cannot be determined, Object remains "".
- If f contains documentation (f.Doc.Text()), WasDocumented is set accordingly; Documentation is preserved or overwritten when an existing FunctionInfo is reused.
- The function assumes f is a valid *ast.FuncDecl with non-nil fields accessed here (e.g., f.Name, f.Doc).

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
Creates a FunctionCall and its associated FunctionInfo from a Go AST CallExpr.
Handles internal, package, and object calls, and returns nil for non-call expressions
such as type casts or parenthesized expressions. Reuses or stores FunctionInfo in
a global FunctionMap when possible.
Signature
func (p *PackageNode) CreateFunctionCall(fun *ast.CallExpr) *FunctionCall
Parameters
- fun: *ast.CallExpr
    The call expression to analyze. Non-nil when invoked.
Returns
- *FunctionCall
    A FunctionCall with its Node set to fun and Info populated according to the call kind.
- nil
    Returned for ArrayType (type casts) and ParenExpr (parenthesized expressions).
Errors/Exceptions
- log.Fatalf on unexpected fun.Fun types: "Failed to switch types on function %T: %v"
Side Effects
- Mutates FunctionMap by assigning or reusing a FunctionInfo under the key fInfo.FullName()
- Sets fCall.Info to the derived fInfo
- May read and use p.TypesInfo, p.Imports, and p.ID
Edge Cases & Assumptions
- If FunctionInfo.FullName() yields a key that already exists in FunctionMap, the existing
  pointer is reused.
- If fInfo.Package or fInfo.Name are empty, FullName() may produce strings with empty segments.
- For package calls, ResolvedPkg is derived from p.Imports; for object calls, ResolvedPkg may be inferred
  from the object's package path or the type of the receiver if not a named package.

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
Summary: Clips cyclic graphs by pruning repeated function call entries from each FunctionDeclaration's FunctionInfo. It iterates over p.FunctionDeclarations and invokes ClipFunctionCycles with an empty callStack to prune cycles in place, returning an error if any call fails.
Signature: func (p *PackageNode) ClipCyclicGraphs() error
Parameters:
  - p *PackageNode: receiver; context for pruning. No additional parameters.
Returns:
  - error: nil on success; non-nil if a nested ClipFunctionCycles call fails (wrapped as "failed to clip function cycles: %v").
Errors/Exceptions:
  - none other than the returned error from ClipFunctionCycles; errors are wrapped and propagated.
Side Effects:
  - Mutates the in-memory data structures inside FunctionDeclarations via ClipFunctionCycles (specifically, FunctionInfo.Declaration.Calls is pruned in place).
Edge Cases & Assumptions:
  - If p.FunctionDeclarations is nil or empty, this function is a no-op.
  - Assumes each element of p.FunctionDeclarations provides Info suitable for ClipFunctionCycles; behavior depends on ClipFunctionCycles' handling of nil or missing fields.

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
ClipFunctionCycles removes repeated function call entries from a FunctionInfo by pruning calls in
f.Declaration.Calls whose fully-qualified names appear on the given callStack. It recursively
descends into non-cyclic calls, appending the current function's FullName() to the callStack.
The function uses the inline guide: "Remove the repeated node from the calls array, but don't descend it.
If you descend it, all the nodes above it will be removed as well (since they've already been included in the list)".
After traversal, it removes the collected indices from f.Declaration.Calls in reverse order.
Summary: Prunes cycles/repeats in f.Declaration.Calls by depth-first traversal, updating in-place.
Signature
func (p *PackageNode) ClipFunctionCycles(f *FunctionInfo, callStack []string) error
Parameters
- p *PackageNode: receiver; context for pruning.
- f *FunctionInfo: target function; if f.Declaration == nil, the call is a no-op.
- callStack []string: current path of fully-qualified names for cycle detection; used to identify repeats.
Returns
- error: always nil in this implementation.
Errors/Exceptions
- none.
Side Effects
- mutates f.Declaration.Calls by removing elements.
Edge Cases & Assumptions
- If f.Declaration is nil, the function returns immediately.
- Assumes each element of f.Declaration.Calls exposes FullName() and has an Info field; no nil checks
  are performed on nested structures.
- Removal is performed in reverse index order to preserve correct positions.

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
Summary: Returns the start and end byte offsets of the given ast.Node within the file associated with p.Fset. Use this when you need to slice source text corresponding to a node.
Signature: func (p *PackageNode) FindStartEnd(n ast.Node) (int, int)
Parameters:
  n: ast.Node — the node whose byte-range to locate; if n == nil, the function returns -1, -1.
Returns:
  start: int — file offset of n.Pos().
  end: int — file offset of n.End() (the end is the position after the node; suitable for slicing [start:end]).
Errors/Exceptions:
  None (no error return). Returns -1, -1 if n == nil or if the corresponding file cannot be found (file == nil).
Side Effects:
  Calls p.Fset.File(n.Pos()) to obtain the token.File and uses file.Offset; does not modify state.
Edge Cases & Assumptions:
  If p.Fset.File(n.Pos()) returns nil, returns -1, -1. Assumes p.Fset is initialized and n is a valid ast.Node. End() is treated as the position after the node for safe slicing.

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
Summary: Return the absolute byte offset of the start of the line that contains the given AST node n. If n is nil, returns -1.

Signature: func (p *PackageNode) FindLineNo(n ast.Node) int

Parameters:
- n: ast.Node — the node whose containing line's start offset is computed.

Returns:
- int — the absolute offset (in bytes) of the start of the line containing n.Pos().

Errors/Exceptions:
- None returned; -1 is returned when n is nil.

Side Effects:
- None.

Edge Cases & Assumptions:
- If n == nil, returns -1.
- Assumes p.Fset and its File(n.Pos()) provide valid position data; relies on File(n.Pos()) and LineStart to compute the offset.

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
Summary: Returns the start and end byte offsets of the FunctionDecl's Node within the file associated with f.Info.Package. Use this when you need to slice source text corresponding to a node.
Signature: func (f *FunctionDecl) FindStartEnd() (int, int)
Parameters:
  none
Returns:
  start: int — file offset of f.Node.Pos().
  end: int — file offset of f.Node.End() (the end is the position after the node; suitable for slicing [start:end]).
Errors/Exceptions:
  None (no error return). Returns -1, -1 if f.Node == nil or if the corresponding file cannot be found (file == nil).
Side Effects:
  No state modifications. Delegates to f.Info.Package.FindStartEnd(f.Node); does not modify state.
Edge Cases & Assumptions:
  Assumes f.Info.Package is initialized and f.Node is a valid ast.Node. If f.Node is nil or the underlying file lookup fails, the function returns -1, -1.

*/
func (f *FunctionDecl) FindStartEnd() (int, int) {
	return f.Info.Package.FindStartEnd(f.Node)
}

/*
Summary: Return the absolute byte offset of the start of the line that contains this FunctionDecl's Node.
          If the FunctionDecl's Node is nil, returns -1. This delegates to f.Info.Package.FindLineNo(f.Node).
Signature: func (f *FunctionDecl) FindLineNo() int
Returns: int — the absolute offset (in bytes) of the start of the line containing f.Node.
          May be -1 if f.Node is nil.
Errors/Exceptions: None.
Side Effects: None.
Edge Cases & Assumptions:
- If f.Node == nil, returns -1.
- Assumes f.Info.Package and its data provide valid position data; relies on File(n.Pos()) and LineStart to compute the offset.

*/
func (f *FunctionDecl) FindLineNo() int {
	return f.Info.Package.FindLineNo(f.Node)
}

/*
Summary: Returns the start and end byte offsets of the ast.Node stored in f.Node within the file associated with f.Info.Package. Use this when you need to slice source text corresponding to a node.
Signature: func (f *FunctionCall) FindStartEnd() (int, int)
Parameters: none
Returns:
  start: int — file offset of f.Node.Pos().
  end: int — file offset of f.Node.End() (the end is the position after the node; suitable for slicing [start:end])
Errors/Exceptions: None (no error return). Returns -1, -1 if f.Node == nil or if the corresponding file cannot be found (file == nil).
Side Effects: None. Delegates to f.Info.Package.FindStartEnd(f.Node); does not modify state.
Edge Cases & Assumptions: If f.Node is nil, returns -1, -1. Assumes f.Info and f.Node are initialized. End() is treated as the position after the node for safe slicing.

*/
func (f *FunctionCall) FindStartEnd() (int, int) {
	return f.Info.Package.FindStartEnd(f.Node)
}

/*
Summary: Return the absolute byte offset of the start of the line that contains the FunctionCall's AST node (f.Node). This method delegates to f.Info.Package.FindLineNo(f.Node).
Signature: func (f *FunctionCall) FindLineNo() int
Returns: int — the absolute offset (in bytes) of the start of the line containing f.Node.
Side Effects: None.
Edge Cases & Assumptions:
- If f.Node is nil, behavior comes from f.Info.Package.FindLineNo(f.Node).
- Assumes f.Info.Package provides valid position data and that Package.FindLineNo can compute the line start from f.Node.

*/
func (f *FunctionCall) FindLineNo() int {
	return f.Info.Package.FindLineNo(f.Node)
}

/*
Summary: Updates documentation strings for undocumented function declarations within the package file. It iterates p.FunctionDeclarations in reverse order, skipping any function that already has a non-empty doc comment, computes the start offset of the function node, and inserts f.Info.Documentation (followed by a newline) into the file at f.Info.File at that offset.

Signature: func (p *PackageNode) UpdateDocsInFile() error

Parameters:
  - none (method on *PackageNode)

Returns:
  - error: non-nil if updating the docs in any file fails

Errors/Exceptions:
  - Returns an error of the form "failed to update docs in file: %v" when insertIntoFile fails for any function
  - If a function already has a non-empty doc comment, it is skipped without error

Side Effects:
  - Modifies the content of files specified by f.Info.File by inserting documentation text at the function’s start position
  - May perform file I/O via insertIntoFile and may log debug output

Edge Cases & Assumptions:
  - If fd.Doc is non-nil and contains non-whitespace text, the function is skipped (no change)
  - start is obtained via p.FindStartEnd(fd) and used directly as the insertion offset
  - Assumes p.FindStartEnd(fd) returns a valid start offset and that f.Info.File is accessible for read/write
  - The inserted text consists of f.Info.Documentation followed by a newline
Notes:
  - The code comment indicates: We read in pre-existing docs.

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
Summary:
  Collects all function invocation expressions (*ast.CallExpr) found within the AST subtree rooted at f by traversing with ast.Inspect.

Signature:
  func GetFunctionInvocations(f ast.Node) ([]*ast.CallExpr, error)

Parameters:
  f: ast.Node - root node of the AST to inspect for function calls.

Returns:
  ([]*ast.CallExpr, error) - slice of pointers to ast.CallExpr. The error value is always nil.

Errors/Exceptions:
  None. This function always returns a nil error.

Side Effects:
  Allocates and populates a slice with found *ast.CallExpr values; does not mutate the input AST.

Edge Cases & Assumptions:
  - If f contains no function invocations, returns an empty slice.
  - The returned call expressions appear in the order encountered by ast.Inspect.
  - If f is nil, the result is an empty slice.

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
Summary: Returns the receiver's named type for a method declaration, first by inspecting the function object via go/types, and if that fails, by peeling the syntax and resolving through type information (info.Uses). Returns the named type and true when found; otherwise returns nil and false.

Signature: func MethodRecvNamed(fd *ast.FuncDecl, info *types.Info) (*types.Named, bool)

Parameters:
- fd: *ast.FuncDecl — the function declaration to inspect; must have a non-nil Recv with at least one element to be considered.
- info: *types.Info — type information used to resolve the receiver type; may be consulted via info.Defs and info.Uses.

Returns:
- *types.Named — the resolved named receiver type when found.
- bool — true if a named receiver type was found; false otherwise.

Errors/Exceptions: None. The function returns (nil, false) when a named receiver type cannot be determined.

Side Effects: None.

Edge Cases & Assumptions:
- If fd is nil, fd.Recv is nil, or fd.Recv.List is empty, the function returns (nil, false).
- The function first attempts to obtain the receiver type from the function object (info.Defs[fd.Name]); if that yields a signature with a receiver, it unwraps pointer types and returns a *types.Named when applicable.
- If the first path fails, it peels the receiver expression (handling *T, (T), T[...], etc.) and uses info.Uses to resolve either a *types.TypeName with a *types.Named.
- When the receiver is a pointer, it unwraps to the element type before checking for a named type.

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
Summary: ParsePackage locates Go source folders under foldername, loads their packages, builds []PackageNode entries, runs sanity checks, populates package information, and clips cyclic graphs. Use this to obtain a structured representation of packages in a directory tree for documentation or analysis.

Signature: func ParsePackage(foldername string) ([]PackageNode, error)

Parameters:
- foldername: string
  Path to the root directory to parse for Go packages.

Returns:
- []PackageNode
  A slice of PackageNode representing each discovered package.
- error
  Non-nil on failure to discover folders, load packages, validate syntax, populate information, or clip cycles.

Errors/Exceptions:
- "failed to get go directories: %v" if GetNestedFoldersWithGoFiles fails.
- "failed to load package %v: %v" if packages.Load fails.
- "failed to sanity check: %v" if PackageNode.SanityCheck fails.
- "failed to populate package information: %v" if PackageNode.PopulatePackageInformation fails.
- "failed to clip cyclic graphs: %v" if PackageNode.ClipCyclicGraphs fails.

Side Effects:
- Reads filesystem to locate Go source folders.
- Logs debug messages via log.Debugf.
- Mutates PackageNode fields (CurrentFile, Imports, TypeDefinitions, FunctionDeclarations).
- May deduplicate, populate, and clip data as part of processing.

Edge Cases & Assumptions:
- If no folders with Go files are found, returns an empty slice with nil error.
- Assumes alignment between p.Syntax and p.CompiledGoFiles within each PackageNode.
- Propagates the first encountered error; processing stops on error.

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
Summary
GetNestedFoldersWithGoFiles returns the set of folders under the given folder that contain at least one .go file. The result is sorted and deduplicated. Use this to identify Go source-containing directories within a tree.

Signature
func GetNestedFoldersWithGoFiles(folder string) ([]string, error)

Parameters
- folder: string
  - Path to the root directory to search for Go source files.

Returns
- []string
  - A sorted slice of absolute directory paths that contain at least one .go file. Each path is the directory containing a Go file.
- error
  - Non-nil if an error occurred while walking the directory tree or while resolving file paths.

Errors/Exceptions
- If filepath.WalkDir returns an error, the function returns an empty slice and an error formatted as:
  fmt.Errorf("failed to walk directory for go files: %v", err)
- If filepath.Abs(path) fails for any encountered Go file, that error is returned and the walk stops.

Side Effects
- Logs a debug message via log.Debugf("folders searching: %v", folder).
- Reads the filesystem to locate .go files.
- Populates and updates an internal seen map to deduplicate directories.
- Sorts the resulting list before returning.

Edge Cases & Assumptions
- If no .go files are found, the function returns an empty slice and nil error.
- Only files with the extension ".go" are considered; other file types are ignored.
- Directories are returned as absolute paths, derived from the directory of each Go file.
- The results are deduplicated using a map and then converted to a sorted slice.

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
