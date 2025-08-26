package ast

import (
	"os"
	"fmt"
        "time"
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
 * Summary:
 * Returns the full name of the function, including package and object if present.
 *
 * Signature:
 * func (f *FunctionInfo) FullName() string
 *
 * Parameters:
 * f - pointer to FunctionInfo struct
 *
 * Returns:
 * The full name as a string, formatted as "Package.Object.Name" if Object is non-empty,
 * otherwise "Package.Name".

*/
func (f *FunctionInfo) FullName() string {
	if f.Object != "" {
		return fmt.Sprintf("%s.%s.%s", f.Package, f.Object, f.Name)
	}

	return fmt.Sprintf("%s.%s", f.Package, f.Name)
}

/*
*
 * Summary: Returns the full name of the function, including package and object if present.
 * Signature: func (f *FunctionCall) FullName() string
 * Parameters:
 *   f - pointer to FunctionCall instance
 * Returns:
 *   The full name as a string, formatted as "Package.Object.Name" if Object is non-empty, otherwise "Package.Name".

*/
func (f *FunctionCall) FullName() string {
	return f.Info.FullName()
}

/*
*
 * Summary: Returns the full name of the function, including package and object (if any).
 *
 * Signature:
 * func (f *FunctionDecl) FullName() string
 *
 * Parameters:
 *   f - pointer to FunctionDecl instance
 *
 * Returns:
 *   The full name as a string, formatted as "Package.Object.Name" if Object is non-empty, otherwise "Package.Name".
 *
 * Errors/Exceptions:
 *   None
 *
 * Side Effects:
 *   None
 *
 * Edge Cases & Assumptions:
 *   Assumes f.Info.FullName() correctly formats the name.

*/
func (f *FunctionDecl) FullName() string {
	return f.Info.FullName()
}

/*
*
 * Prints detailed information about the FunctionInfo instance, including object, name, file, package, documentation, and called functions.
 *
 * @param prefix string to prepend to each line for indentation.

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
 * Prints detailed information about the FunctionDecl instance, including object, name, file, package, documentation, and called functions.
 *
 * @param prefix string to prepend to each line for indentation.

*/
func (f *FunctionDecl) PrettyPrint(prefix string) {
	f.Info.PrettyPrint(prefix)
}

/*
*
 * Summarizes the function's purpose: generates a string representation of a FunctionDecl, incorporating documentation comments from its calls.
 *
 * Signature:
 * func (f *FunctionDecl) ToStringForGPT() (string, error)
 *
 * Parameters:
 *   - f: *FunctionDecl, the function declaration node to convert to string.
 *
 * Returns:
 *   - string: the source code text of the function, with embedded documentation comments.
 *   - error: error if the file cannot be read.
 *
 * Errors/Exceptions:
 *   - Returns an error if reading the file fails.
 *
 * Side Effects:
 *   - Reads the file specified in f.Info.File.
 *   - Modifies the fd_text string with documentation comments from calls.
 *
 * Edge Cases & Assumptions:
 *   - Assumes f.FindStartEnd() correctly sets fd_start and fd_end.
 *   - Assumes f.Calls[i].FindStartEnd() accurately finds start/end offsets for each call.
 *   - Assumes documentation comments do not contain newlines or special formatting requiring more complex handling.

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
		fc_start, fc_end := f.Calls[i].FindStartEnd()

		fc_start -= fd_start
		fc_end -= fd_start

		docs := f.Calls[i].Info.Documentation
		if strings.TrimSpace(docs) == "" {
			continue
		}
		fd_text = fd_text[:fc_start] + " /* " + strings.ReplaceAll(docs, "\n", "|") + " */ " + fd_text[fc_start:]
	}

	return fd_text, nil
}

/*
*
 * Summarizes the function's purpose: generates a string representation of a FunctionDecl, incorporating documentation comments from its calls.
 *
 * Signature:
 * func (f *FunctionDecl) ToStringForGPT() (string, error)
 *
 * Parameters:
 *   - f: *FunctionDecl, the function declaration node to convert to string.
 *
 * Returns:
 *   - string: the source code text of the function, with embedded documentation comments.
 *   - error: error if the declaration is nil.
 *
 * Errors/Exceptions:
 *   - Returns an error if f.Declaration is nil.
 *
 * Side Effects:
 *   - Reads the file specified in f.Info.File.
 *   - Potentially modifies the string with documentation comments from calls.
 *
 * Edge Cases & Assumptions:
 *   - Assumes f.FindStartEnd() correctly sets start and end positions.
 *   - Assumes f.Calls[i].FindStartEnd() accurately finds call boundaries.
 *   - Assumes documentation comments are in a format that does not require special parsing.

*/
func (f *FunctionInfo) ToStringForGPT() (string, error) {
	if f.Declaration == nil {
		return "", fmt.Errorf("can't convery %v to string for gpt. no delcaration", f.Name)
	}

	return f.Declaration.ToStringForGPT()
}

/*
*
 * GetDocumentation retrieves the documentation associated with the function.
 *
 * Signature:
 * func (f *FunctionInfo) GetDocumentation() (string, error)
 *
 * Parameters:
 * None
 *
 * Returns:
 * - string: the accumulated documentation text.
 * - error: nil if successful.
 *
 * Side Effects:
 * None
 *
 * Edge Cases & Assumptions:
 * If the function's Declaration Node has associated documentation comments (fd.Doc), they are appended to the existing documentation string if it is empty.

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
 * Performs a sanity check on the PackageNode to verify internal consistency.
 * Reports errors contained in the node and ensures syntax trees are present.
 * Should be invoked to validate the integrity of the PackageNode before further use.
 *
 * Signature:
 * func (p *PackageNode) SanityCheck() error
 *
 * Parameters:
 * - p: *PackageNode, the package node to validate.
 *
 * Returns:
 * - error: nil if no issues are found; otherwise, an error describing the problem.
 *
 * Errors/Exceptions:
 * - Returns an error if no syntax trees are present in the node.
 *
 * Side Effects:
 * - None.
 *
 * Edge Cases & Assumptions:
 * - Assumes p.Errors contains error messages to be logged; the function formats but does not output them.

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
 * Summary: Populates package information by processing syntax trees, updating import mappings,
 * adding type definitions, and extracting function declarations with their calls.
 * Use this function to initialize comprehensive package metadata from syntax files.
 *
 * Signature: func (p *PackageNode) PopulatePackageInformation() error
 *
 * Side Effects:
 * - Updates `p.CurrentFile` for each syntax file.
 * - Mutates `p` by adding import paths, type definitions, and function declarations.
 * - Logs information about processing.
 *
 * Errors/Exceptions:
 * - Returns an error if adding to the import map, expanding type definitions, or function declarations fails.
 *
 * Edge Cases & Assumptions:
 * - Assumes `p.Syntax` and `p.CompiledGoFiles` are aligned and contain valid data.
 * - Assumes each syntax tree is properly parsed and valid.

*/
func (p *PackageNode) PopulatePackageInformation() error {

	for i, syn_ast := range p.Syntax {
		p.CurrentFile = p.CompiledGoFiles[i]
		log.Infof("Stripping ASTs from %v: ", p.CurrentFile)

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
 * Adds import paths and their alias names to the PackageNode's import map.
 *
 * @param f_ast *ast.File - The syntax tree of the file containing import declarations.
 * @return error - An error if importing any package fails; otherwise, nil.

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
 * Adds all *ast.TypeSpec nodes found in the provided AST node to the PackageNode's TypeDefinitions slice.
 *
 * Signature: func (p *PackageNode) AddToTypeDefinitions(f ast.Node) error
 *
 * Parameters:
 * - f: ast.Node - The AST node to inspect for *ast.TypeSpec nodes.
 *
 * Returns:
 * - error: always nil.
 *
 * Side Effects:
 * - Appends found *ast.TypeSpec nodes to p.TypeDefinitions.

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
 * Summary: Adds function declarations from the AST file `f` to the `PackageNode`, extracting function call information within each declaration.
 * Signature: func (p *PackageNode) AddToFunctionDeclarations(f *ast.File) error
 * Parameters:
 *   - f: *ast.File — The AST file containing function declarations to process.
 * Returns:
 *   - error: Returns an error if processing fails; otherwise, nil.
 * Errors/Exceptions:
 *   - Returns an error if retrieving function invocations fails for any declaration.
 * Side Effects:
 *   - Mutates `p.FunctionDeclarations` by appending new `FunctionDecl` objects.
 *   - Populates `Calls` in each `FunctionDecl` with function call expressions.
 * Edge Cases & Assumptions:
 *   - Assumes `f` contains valid AST nodes.
 *   - If no function declarations are present, `p.FunctionDeclarations` is initialized as an empty slice.

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

		log.Infof("Declaration later: %v", FunctionMap[newFuncNode.FullName()].Declaration)
	}

	return nil
}


/*
*
 * Creates a new FunctionDecl object based on the provided ast.FuncDecl.
 *
 * Determines the receiver's named type if present and initializes FunctionInfo with relevant details.
 * Checks for existing function information in the FunctionMap and updates accordingly.
 * Associates the FunctionDecl with its FunctionInfo.
 *
 * @param f Pointer to ast.FuncDecl representing the function declaration.
 * @return Pointer to the created FunctionDecl.

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
 * Creates a *FunctionCall object representing a call expression.
 *
 * @param fun The *ast.CallExpr representing the function call.
 * @return The constructed *FunctionCall with populated *FunctionInfo.

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

	default:
		log.Fatalf("failed to switch types on function %T: %v", fun.Fun, fun.Fun)
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
 * Summary: Removes function calls that create cycles within each function declaration in the PackageNode.
 * Signature: func (p *PackageNode) ClipCyclicGraphs() error
 * Parameters:
 *   p - pointer to PackageNode; the package containing function declarations to process.
 * Returns:
 *   error if cycle clipping fails; otherwise, nil.
 * Errors/Exceptions:
 *   Returns an error if ClipFunctionCycles encounters an issue processing any function declaration.
 * Side Effects:
 *   Mutates each FunctionDeclaration's Calls by removing cyclic calls.

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
 * Summary: Removes function calls from the declaration that create cycles based on the call stack.
 *
 * Signature: func (p *PackageNode) ClipFunctionCycles(f *FunctionInfo, callStack []string) error
 *
 * Parameters:
 *   f - pointer to FunctionInfo; the function whose calls are to be checked and potentially removed
 *   callStack - slice of strings; the current call hierarchy to detect cycles
 *
 * Returns:
 *   nil on successful removal; otherwise, an error if applicable
 *
 * Side Effects:
 *   Mutates f.Declaration.Calls by removing calls that form cycles

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
 * FindStartEnd returns the start and end byte offsets of the given AST node within the source file.
 *
 * Signature:
 * func (p *PackageNode) FindStartEnd(n ast.Node) (int, int)
 *
 * Parameters:
 *   - n: ast.Node, the code node to locate, can be nil.
 *
 * Returns:
 *   - start: int, the byte offset where the node begins; -1 if n is nil or the file is unavailable.
 *   - end: int, the byte offset where the node ends; -1 if n is nil or the file is unavailable.
 *
 * Errors/Exceptions:
 *   - None explicitly; returns -1, -1 if input is invalid or file info is unavailable.

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
 * Summary:
 * Finds the start and end byte offsets of the given AST node within the source file.
 *
 * Signature:
 * func (f *FunctionDecl) FindStartEnd() (int, int)
 *
 * Parameters:
 *   - f: *FunctionDecl, the function declaration node to locate.
 *
 * Returns:
 *   - start: int, the byte offset where the node begins; -1 if f or its file info is unavailable.
 *   - end: int, the byte offset where the node ends; -1 if f or its file info is unavailable.
 *
 * Errors/Exceptions:
 *   - None explicitly; returns -1, -1 if input is invalid or file info is unavailable.
 *
 * Side Effects:
 *   - None.
 *
 * Edge Cases & Assumptions:
 *   - Assumes f.Info.Package.FindStartEnd properly retrieves position info for f.Node.

*/
func (f *FunctionDecl) FindStartEnd() (int, int) {
	return f.Info.Package.FindStartEnd(f.Node)
}

/*
*
 * FindStartEnd returns the start and end byte offsets of the given AST node within the source file.
 *
 * Signature:
 *   func (p *PackageNode) FindStartEnd(n ast.Node) (int, int)
 *
 * Parameters:
 *   - n: ast.Node, the code node to locate, can be nil.
 *
 * Returns:
 *   - start: int, the byte offset where the node begins; -1 if n is nil or the file is unavailable.
 *   - end: int, the byte offset where the node ends; -1 if n is nil or the file is unavailable.
 *
 * Errors/Exceptions:
 *   - None explicitly; returns -1, -1 if input is invalid or file info is unavailable.

*/
func (f *FunctionCall) FindStartEnd() (int, int) {
	return f.Info.Package.FindStartEnd(f.Node)
}

/*
*
 * Updates missing documentation comments in function declarations within a package.
 * For each function declaration lacking a doc comment, inserts the associated documentation into the source file at the function's position.
 *
 * Signature:
 * func (p *PackageNode) UpdateDocsInFile() error
 *
 * Side Effects:
 * - Modifies source files by inserting documentation comments.
 *
 * Errors/Exceptions:
 * - Returns an error if inserting documentation into any file fails.

*/
func (p *PackageNode) UpdateDocsInFile() error {

	for i := len(p.FunctionDeclarations) - 1; i >= 0; i-- {
		f := p.FunctionDeclarations[i]

		fd := f.Node

		// We read in pre-existing docs
		if fd.Doc != nil {
			continue
		}

		start, _ := p.FindStartEnd(fd)

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
 * Summary: Collects all function call expressions (`*ast.CallExpr`) within the provided AST node.
 *
 * Signature: func GetFunctionInvocations(f ast.Node) ([]*ast.CallExpr, error)
 *
 * Parameters:
 * - f: ast.Node — the AST node to inspect for function call expressions.
 *
 * Returns:
 * - slice of *ast.CallExpr — all function call expressions found in the AST node.
 * - error — always nil in current implementation.
 *
 * Errors/Exceptions: None explicitly, error is always nil.
 *
 * Side Effects: None.
 *
 * Edge Cases & Assumptions:
 * - Returns an empty slice if no function call expressions are found.
 * - Assumes `f` is a valid AST node.

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
 * Summary: Extracts the named type of a method receiver from a function declaration, returning it along with a boolean indicating success.
 *
 * Signature: func MethodRecvNamed(fd *ast.FuncDecl, info *types.Info) (*types.Named, bool)
 *
 * Parameters:
 *   - fd: *ast.FuncDecl
 *       The function declaration to analyze.
 *   - info: *types.Info
 *       Type information mapped from the AST.
 *
 * Returns:
 *   - *types.Named: The named type of the receiver if found; nil otherwise.
 *   - bool: true if the receiver's named type was successfully identified; false otherwise.
 *
 * Errors/Exceptions: None documented; method returns nil, false on failure.
 *
 * Side Effects: None.

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
Summary: Loads and processes Go package syntax files from the specified folder, constructing a slice of PackageNode objects with validated and populated package data, including function declarations, type definitions, and import mappings. Performs cycle clipping on function call graphs.
Signature: func ParsePackage(foldername string) ([]PackageNode, error)
Parameters:
  - foldername: string; the directory path containing the package's source files.
Returns:
  - []PackageNode: a slice of processed package nodes.
  - error: non-nil if loading, validation, population, or cycle clipping fails.
Errors/Exceptions:
  - Returns an error if package loading, sanity checking, populating, or cycle clipping fails.
Side Effects:
  - Performs package loading, syntax parsing, validation, population, and cycle clipping.
  - Mutates PackageNode fields such as FunctionDeclarations, TypeDefinitions, Imports, and calls within FunctionDeclarations.
Edge Cases & Assumptions:
  - Assumes package files are valid and syntax is correctly parsed.
  - Assumes aligned syntax and compiled Go files.
  - Assumes cycle clipping handles cyclic call graphs appropriately.

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

        folders, err := GetNestedFoldersWithGoFiles(foldername)
        if err != nil {
            return nil, fmt.Errorf("failed to get go directories: %v", err)
        }

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

func GetNestedFoldersWithGoFiles(folder string) ([]string, error) {
        seen := make(map[string]struct{})

	err := filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".go" {
			dir := filepath.Dir(path)
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

        log.Infof("keys: %v", keys)

        time.Sleep(time.Duration(1000000000000000000000000) * time.Second)

        return keys, nil
}

