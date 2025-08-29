package ast

import (
	"os"
	"fmt"
	"sort"
	"io/fs"
	"slices"
	"strings"
	"path/filepath"

	"go/ast"
	"go/build"
	"go/types"

	"golang.org/x/tools/go/packages"

	log "github.com/sirupsen/logrus"

	asTypes "github.com/BlankCanvasStudio/AutoScribe/pkg/types"
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

	Language asTypes.SupportedFormat

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
*
 * Summary: Returns the full name of the function, including package and object names if available.
 * Signature: func (f *FunctionInfo) FullName() string
 * Returns: A string representing the full name, with format "package.object.name" if Object is non-empty, otherwise "package.name".

*/
func (f *FunctionInfo) FullName() string {
	if f.Object != "" {
		return fmt.Sprintf("%s.%s.%s", f.Package, f.Object, f.Name)
	}

	return fmt.Sprintf("%s.%s", f.Package, f.Name)
}

/*
*
 * Summary: Returns the full name of the function, including package and object names if available.
 * Signature: func (f *FunctionCall) FullName() string
 * Returns: A string representing the full name, with format "package.object.name" if Object is non-empty, otherwise "package.name".

*/
func (f *FunctionCall) FullName() string {
	return f.Info.FullName()
}

/*
*
 * Summary: Returns the complete name of the function, including package and object names if present.
 * Signature: func (f *FunctionDecl) FullName() string
 * Returns: A string representing the function's full name, formatted as "package.object.name" if Object is non-empty, otherwise "package.name".

*/
func (f *FunctionDecl) FullName() string {
	return f.Info.FullName()
}

/*
*
 * Summary: Prints detailed information about the FunctionInfo instance, including its name, object, file, package, documentation, and called functions.
 * Signature: func (f *FunctionInfo) PrettyPrint(prefix string)
 * Parameters:
 *   - prefix: string, a string to prepend to each printed line for indentation.
 * Returns: none
 * Errors/Exceptions: none
 * Side Effects: prints information to standard output
 * Edge Cases & Assumptions: handles empty Object or Documentation fields appropriately; does nothing if Declaration is nil.

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
*
 * Summary: Prints detailed information about the FunctionDecl instance, including its associated FunctionInfo, with optional prefix indentation.
 * Signature: func (f *FunctionDecl) PrettyPrint(prefix string)
 * Parameters:
 *   - prefix: string, a string to prepend to each printed line for indentation.
 * Returns: none
 * Errors/Exceptions: none
 * Side Effects: prints information to standard output
 * Edge Cases & Assumptions: handles empty Object or Documentation fields appropriately; does nothing if Declaration is nil.

*/
func (f *FunctionDecl) PrettyPrint(prefix string) {
	f.Info.PrettyPrint(prefix)
}

/*
*
 * Summary: Generates a source code string with inline documentation comments for the specified FunctionDecl, suitable for GPT processing. It inserts documentation from its Calls into the raw file content.
 * Signature: func (f *FunctionDecl) ToStringForGPT() (string, error)
 * Parameters:
 *   - f: *FunctionDecl, the function declaration object containing file info and calls.
 * Returns:
 *   - string: the modified source code with embedded documentation comments.
 *   - error: if reading the source file fails.
 * Errors/Exceptions:
 *   - Returns an error if the source file cannot be read.
 * Side Effects:
 *   - Reads the source file; modifies and returns a string with inserted comments.
 *   - Calls FindStartEnd on the FunctionDecl and its Calls to locate code segments.
 * Edge Cases & Assumptions:
 *   - Assumes the raw file can be read successfully.
 *   - Assumes the offset calculations and insertions are valid and do not corrupt the source code structure.

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
*
Summary: Generates a source code string with inline documentation comments for the FunctionInfo's declaration, suitable for GPT processing. It inserts documentation from its Calls into the raw file content.
Signature: func (f *FunctionInfo) ToStringForGPT() (string, error)
Parameters:
  - f: *FunctionInfo, the object containing the function declaration and related calls.
Returns:
  - string: the modified source code with embedded documentation comments.
  - error: if the declaration is nil or reading the source file fails.
Errors/Exceptions:
  - Returns an error if the declaration is nil.
Side Effects:
  - Reads the source file; modifies and returns a string with inserted comments.
  - Calls FindStartEnd on the FunctionInfo and its Calls to locate code segments.
Edge Cases & Assumptions:
  - Assumes the raw file can be read successfully.
  - Assumes the offset calculations and insertions are valid and do not corrupt the source code structure.

*/
func (f *FunctionInfo) ToStringForGPT() (string, error) {
	if f.Declaration == nil {
		return "", fmt.Errorf("can't convery %v to string for gpt. no delcaration", f.Name)
	}

	return f.Declaration.ToStringForGPT()
}

/*
*
 * GetDocumentation returns the documentation string for the FunctionInfo.
 *
 * Signature: func (f *FunctionInfo) GetDocumentation() (string, error)
 *
 * Parameters:
 *   - f: The FunctionInfo instance to retrieve documentation from.
 *
 * Returns:
 *   - A string containing the documentation text.
 *   - An error, if any occurs during execution.
 *
 * Notes:
 *   - If Documentation is empty and the associated declaration node has doc comments, those are appended.

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
*
 * Validates the PackageNode by checking for existing errors and syntax trees.
 * Returns an error if no syntax trees are present or if other issues are detected.
 *
 * Signature:
 * func (p *PackageNode) SanityCheck() error
 *
 * Parameters:
 * - p: PackageNode - the package node to validate.
 *
 * Returns:
 * - error: an error indicating missing syntax trees or other validation issues.
 *
 * Side Effects:
 * - Records errors to standard output (via fmt.Errorf in the loop).
 *
 * Edge Cases & Assumptions:
 * - Assumes p.Errors is a list of error messages relevant to the package.
 * - Assumes p.Syntax is a collection of syntax trees; emptiness triggers an error.

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
*
 * Populates the PackageNode with package information including import statements,
 * type definitions, and function declarations from its associated syntax ASTs.
 *
 * @return error if any of the processing steps fail.

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
*
 * Adds import statements from an *ast.File to the PackageNode's Imports map.
 *
 * @param f_ast *ast.File The abstract syntax tree representing the Go source file.
 * @return error If building import information fails.
 *
 * The function populates the PackageNode's Imports map with package alias or default names
 * mapped to their import paths. It initializes the map if nil and handles aliasing.

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
*
 * Adds all *ast.TypeSpec nodes within the given AST node to the PackageNode's TypeDefinitions slice.
 * Use this to collect type definitions from a parsed AST subtree.
 *
 * Signature:
 * func (p *PackageNode) AddToTypeDefinitions(f ast.Node) error
 *
 * Parameters:
 * - f: ast.Node; the AST subtree to inspect for type definitions.
 *
 * Returns:
 * - error: always nil; placeholder for future error handling.
 *
 * Side Effects:
 * - Mutates p.TypeDefinitions by appending found *ast.TypeSpec nodes.

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
*
 * Adds function declarations and their call expressions from the provided AST file to the PackageNode.
 * Extracts all function declarations, collects their call expressions, creates FunctionCall objects,
 * and appends new FunctionDecls with associated calls to the PackageNode.
 *
 * @param f *ast.File - The AST file to process.
 * @return error - Returns an error if the process fails.

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
*
Summary: Creates a FunctionDecl object from an ast.FuncDecl, extracting type information and updating the function map.
Signature: func (p *PackageNode) CreateFunctionDecl(f *ast.FuncDecl) *FunctionDecl
Parameters:
  - f: *ast.FuncDecl, the AST node representing the function declaration.
Returns: *FunctionDecl, the created function declaration object.
Errors/Exceptions: None.
Side Effects: Adds or updates entries in the FunctionMap.
Edge Cases & Assumptions: Handles functions with receiver types, including pointer receivers, and manages existing entries in the function map.

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
*
 * Creates a FunctionCall object for the provided ast.CallExpr.
 * Populates FunctionInfo with function name, package, and object details based on the call expression.
 * Reuses existing FunctionInfo if available.
 * Returns nil for type casts and parenthetical expressions.
 *
 * @param fun *ast.CallExpr - The call expression node.
 * @return *FunctionCall - The constructed FunctionCall object with associated FunctionInfo.

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
*
 * Summary: Removes cyclic function calls from each FunctionDeclaration in the PackageNode to prevent infinite recursion.
 *
 * Signature: func (p *PackageNode) ClipCyclicGraphs() error
 *
 * Parameters:
 *   - p: *PackageNode - the package node containing function declarations to process.
 *
 * Returns:
 *   - error: nil if all function cycles are successfully clipped; otherwise, an error indicating failure.
 *
 * Errors/Exceptions:
 *   - Returns an error if ClipFunctionCycles fails on any function declaration.
 *
 * Side Effects:
 *   - Alters the Calls within each FunctionInfo's Declaration by removing calls that form cycles.

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
*
 * Summary: Removes cyclic function calls from a PackageNode's FunctionInfo declaration based on call stack to prevent infinite loops.
 *
 * Signature: func (p *PackageNode) ClipFunctionCycles(f *FunctionInfo, callStack []string) error
 *
 * Parameters:
 *   - f: *FunctionInfo - the function information to process.
 *   - callStack: []string - stack of function names leading to current call, used to detect cycles.
 *
 * Returns:
 *   - error: nil if processing completes successfully.
 *
 * Errors/Exceptions:
 *   - None expected.
 *
 * Side Effects:
 *   - Mutates f.Declaration.Calls by removing calls that create cycles.
 *
 * Edge Cases & Assumptions:
 *   - If f.Declaration is nil, the function exits early.
 *   - Calls are removed only if their full name appears in the callStack, indicating a cycle.

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
*
 * Finds the start and end offsets of an AST node within the source file.
 *
 * Signature: func (p *PackageNode) FindStartEnd(n ast.Node) (int, int)
 *
 * Parameters:
 * - n: ast.Node, the node to locate; if nil, returns (-1, -1).
 *
 * Returns:
 * - start: int, the file offset where the node begins.
 * - end: int, the file offset immediately after the node ends.
 *
 * Errors/Exceptions:
 * - Returns (-1, -1) if the node is nil or if the file cannot be found.
 *
 * Side Effects: none

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
*
 * FindLineNo returns the offset in the source file where the given AST node starts.
 * Use it to determine the position of a node within the source code.
 *
 * Signature:
 * func (p *PackageNode) FindLineNo(n ast.Node) int
 *
 * Parameters:
 * - n: ast.Node, the node for which to find the starting offset; if nil, returns -1.
 *
 * Returns:
 * - int: the offset of the start of the node within the source file.

*/
func (p *PackageNode) FindLineNo(n ast.Node) (int) {
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
*
 * Finds the start and end offsets of an AST node within the source file.
 *
 * Signature:
 *   func (p *PackageNode) FindStartEnd(n ast.Node) (int, int)
 *
 * Parameters:
 *   - n: ast.Node, the node to locate; if nil, returns (-1, -1).
 *
 * Returns:
 *   - start: int, the file offset where the node begins.
 *   - end: int, the file offset immediately after the node ends.
 *
 * Errors/Exceptions:
 *   - Returns (-1, -1) if the node is nil or if the file cannot be found.
 *
 * Side Effects:
 *   - None.

*/
func (f *FunctionDecl) FindStartEnd() (int, int) {
	return f.Info.Package.FindStartEnd(f.Node)
}


/*
*
 * Summary: Returns the absolute offset of the start of the line containing the given AST node within the file, or -1 if the node is nil.
 *
 * Signature: func (f *FunctionDecl) FindLineNo() int
 *
 * Parameters:
 * - None
 *
 * Returns:
 * - int; the absolute offset position of the line start in the file.
 *
 * Errors/Exceptions:
 * - None explicitly; returns -1 when the node is nil.
 *
 * Side Effects:
 * - Reads from the associated token file set and file.
 *
 * Edge Cases & Assumptions:
 * - Assumes f.Node.Pos() provides a valid position within the file set.
 * - If f is nil or its position is invalid, behavior depends on underlying implementation.

*/
func (f *FunctionDecl) FindLineNo() (int) {
	return f.Info.Package.FindLineNo(f.Node)
}

/*
*
 * Finds the start and end offsets of an AST node within the source file.
 * Returns (-1, -1) if the node is nil or if the file cannot be found.
 *
 * Signature: func (p *PackageNode) FindStartEnd(n ast.Node) (int, int)
 *
 * Parameters:
 * - n: ast.Node, the node to locate; if nil, returns (-1, -1).
 *
 * Returns:
 * - start: int, the file offset where the node begins.
 * - end: int, the file offset immediately after the node ends.
 *
 * Errors/Exceptions:
 * - Returns (-1, -1) if the node is nil or if the file cannot be found.
 *
 * Side Effects: none

*/
func (f *FunctionCall) FindStartEnd() (int, int) {
	return f.Info.Package.FindStartEnd(f.Node)
}


/*
*
 * Summary: Returns the absolute offset of the start of the line containing the given AST node within the file, or -1 if the node is nil.
 * Signature: func (f *FunctionCall) FindLineNo() int
 * Parameters:
 * - f: *FunctionCall; the function call instance containing the node and info.
 * Returns:
 * - int; the absolute offset position of the line start in the file.
 * Errors/Exceptions:
 * - None explicitly; returns -1 when f.Node is nil.
 * Side Effects:
 * - Reads from the associated token file set and file.
 * Edge Cases & Assumptions:
 * - Assumes f.Node.Pos() provides a valid position within the file set.
 * - If f.Node is nil, returns -1.

*/
func (f *FunctionCall) FindLineNo() (int) {
	return f.Info.Package.FindLineNo(f.Node)
}

/*
*
 * Summary:
 * Updates documentation comments for function declarations in a package node that lack existing docs.
 * Reads existing docs and inserts new documentation at the appropriate position in the source file.
 *
 * Signature:
 * func (p *PackageNode) UpdateDocsInFile() error
 *
 * Parameters:
 * None
 *
 * Returns:
 * error: nil if all updates succeed; otherwise, an error indicating failure.
 *
 * Errors/Exceptions:
 * - Returns an error if inserting documentation into the file fails for any function declaration.

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
*
 * Summary:
 * Collects and returns all function call expressions (`*ast.CallExpr`) within the given AST node.
 *
 * Signature:
 * func GetFunctionInvocations(f ast.Node) ([]*ast.CallExpr, error)
 *
 * Parameters:
 * - f: ast.Node - The AST node to inspect for function call expressions.
 *
 * Returns:
 * - Slice of pointers to ast.CallExpr containing all found function calls.
 * - error: always returns nil; included for compatibility or future error handling.
 *
 * Errors/Exceptions:
 * - None currently; always returns nil.
 *
 * Side Effects:
 * - None.
 *
 * Edge Cases & Assumptions:
 * - Handles nested AST nodes.
 * - Returns an empty slice if no function calls are found.

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
*
 * Summary: Retrieves the named type of the receiver of a given function declaration, if available.
 * Signature: func MethodRecvNamed(fd *ast.FuncDecl, info *types.Info) (*types.Named, bool)
 * Parameters:
 *   - fd: *ast.FuncDecl, the function declaration to analyze.
 *   - info: *types.Info, type information returned from the type checker.
 * Returns:
 *   - *types.Named: the named type of the receiver if found.
 *   - bool: true if the receiver's named type was successfully identified; false otherwise.
 * Errors/Exceptions: None.
 * Side Effects: None.
 * Edge Cases & Assumptions: Handles pointer receivers and various syntax forms of receiver expressions.

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
*
 * Summary: Parses the specified folder to identify Go packages, validate their syntax, populate package information, and remove cyclic function call graphs.
 *
 * Signature: func ParsePackage(foldername string) ([]PackageNode, error)
 *
 * Parameters:
 * - foldername: string; the root directory to scan for Go source files and packages.
 *
 * Returns:
 * - []PackageNode: list of package nodes representing parsed packages with populated info.
 * - error: non-nil if an error occurs during directory scanning, package loading, validation, or cyclic graph clipping.
 *
 * Errors/Exceptions:
 * - Returns an error if directory traversal, package loading, validation, or cyclic graph clipping fails.
 *
 * Side Effects:
 * - Performs directory traversal and package processing.
 * - Modifies PackageNode objects during validation, population, and cyclic graph clipping.
 *
 * Edge Cases & Assumptions:
 * - Assumes folder contains valid Go source files and packages.
 * - Assumes GetNestedFoldersWithGoFiles handles directory traversal errors.
 * - Assumes PackageNode.SanityCheck, PopulatePackageInformation, and ClipCyclicGraphs modify internal state accordingly.

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
*
 * Summary: Collects and returns a sorted list of unique nested folder paths (relative to the input folder) containing '.go' files.
 * Use this function to identify all directories with Go source files within a specified folder hierarchy.
 *
 * Signature: func GetNestedFoldersWithGoFiles(folder string) ([]string, error)
 *
 * Parameters:
 * - folder: string; the root directory to scan for Go files.
 *
 * Returns:
 * - []string: sorted list of unique folder paths (relative to 'folder') containing '.go' files.
 * - error: non-nil if an error occurs during directory traversal.
 *
 * Errors/Exceptions:
 * - Returns an error if WalkDir encounters an error during traversal.
 * - Defaults to absolute directory path if filepath.Rel fails.

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
