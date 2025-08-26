package ast;

import (
    "os"
    "fmt"

    "slices"
    "strings"

    "go/ast"
    "go/types"
    "go/build"

    "golang.org/x/tools/go/packages"

    log "github.com/sirupsen/logrus"

    asTypes "github.com/BlankCanvasStudio/AutoScribe/pkg/types"
)


type FunctionKind string


var FunctionMap = map[string]*FunctionInfo{}


const (
    ObjectCall     FunctionKind = "object"
    PackageCall    FunctionKind = "package"
    InternalCall   FunctionKind = "internal"
)


type FunctionDecl struct {
    Info          *FunctionInfo
    Node          *ast.FuncDecl
    
    Calls         []*FunctionCall
}

type FunctionCall struct {
    Info          *FunctionInfo
    Node          *ast.CallExpr

    Kind          FunctionKind
}

type FunctionInfo struct {
    Package       *PackageNode

    Language      asTypes.SupportedFormat

    Name          string
    ResolvedPkg   string
    Object        string
    File          string

    Documentation string

    WasDocumented bool // Did we find documentation for it
    Documented    bool // Did we write documentation for it
    AiAware       bool

    Declaration   *FunctionDecl // Where the function is declared & all that jazz
}


/*
*
 * Returns the full name of the function, including package and object if present.
 *
 * Signature:
 * func (f *FunctionInfo) FullName() string
 *
 * Parameters:
 *   - f: Pointer to FunctionInfo struct.
 *
 * Returns:
 *   - A string representing the function's full name in the format "package.object.name" if Object is set,
 *     otherwise "package.name".

*/
func (f *FunctionInfo) FullName() string {
    if f.Object != "" {
        return fmt.Sprintf("%s.%s.%s", f.Package, f.Object, f.Name)
    }

    return fmt.Sprintf("%s.%s", f.Package, f.Name)
}


/*
*
 * Returns the full name of the function, including package and object if present.
 *
 * Signature:
 * func (f *FunctionCall) FullName() string
 *
 * Parameters:
 *   - f: Pointer to FunctionCall instance.
 *
 * Returns:
 *   - A string representing the function's full name in the format "package.object.name" if Object is set, otherwise "package.name".
 *
 * Errors/Exceptions:
 *   - None.
 *
 * Side Effects:
 *   - None.
 *
 * Edge Cases & Assumptions:
 *   - Assumes f.Info.FullName() correctly constructs the name based on Object presence.
 *   - Does not handle nil f or f.Info; expects valid receiver.

*/
func (f *FunctionCall) FullName() string {
    return f.Info.FullName()
}


/*
*
 * Returns the full name of the function, including package and object if present.
 *
 * Signature:
 * func (f *FunctionDecl) FullName() string
 *
 * Parameters:
 *   - f: Pointer to FunctionDecl instance.
 *
 * Returns:
 *   - A string representing the function's full name in the format "package.object.name" if Object is set, otherwise "package.name".

*/
func (f *FunctionDecl) FullName() string {
    return f.Info.FullName()
}


/*
*
 * Summarizes and displays detailed information about a FunctionInfo instance.
 *
 * @param prefix String prefix used for indentation.
 *
 * Prints the function's qualified name, file, package, and optional documentation.
 * If the Declaration field is present, recursively prints information about each called function.

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


/**
 * Summarizes detailed information about the FunctionDecl instance.
 *
 * Signature:
 *   func (f *FunctionDecl) PrettyPrint(prefix string)
 *
 * Parameters:
 *   prefix: string - a string to prepend to each line for formatting purposes.
 *
 * Notes:
 *   - If the Object associated with the FunctionDecl is empty, prints the function name.
 *   - Otherwise, prints Object.Name.
 *   - Includes information about the file, package, documentation, and recursively about called functions.
 */
func (f *FunctionDecl) PrettyPrint(prefix string) {
    f.Info.PrettyPrint(prefix)
}


/**
 * Converts the FunctionDecl to a string suitable for GPT analysis, embedding documentation comments.
 *
 * Signature:
 *   func (f *FunctionDecl) ToStringForGPT() (string, error)
 *
 * Parameters:
 *   - f: *FunctionDecl; the function declaration to process.
 *
 * Returns:
 *   - string: the source code of the function with embedded documentation comments.
 *   - error: if reading the source file fails.
 *
 * Errors/Exceptions:
 *   - Returns an error if os.ReadFile fails.
 *
 * Side Effects:
 *   - Reads the source file specified in f.Info.File.
 *   - Modifies the source code string by embedding documentation comments.
 *
 * Edge Cases & Assumptions:
 *   - Assumes f.FindStartEnd() correctly identifies the function's position in the source file.
 *   - Assumes f.Calls contains relevant call nodes, and their start/end can be adjusted and annotated.
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
        fc_end  -= fd_start

        docs := f.Calls[i].Info.Documentation
        if strings.TrimSpace(docs) == "" { continue }
        fd_text = fd_text[:fc_start] + " /* " + strings.ReplaceAll(docs, "\n", "|") + " */ " + fd_text[fc_start:]
    }

    return fd_text, nil
}

/**
 * Converts the FunctionInfo's Declaration to a string suitable for GPT analysis.
 *
 * Signature:
 *   func (f *FunctionInfo) ToStringForGPT() (string, error)
 *
 * Parameters:
 *   - f: *FunctionInfo; the function info containing the declaration.
 *
 * Returns:
 *   - string: the source code of the function with embedded documentation comments.
 *   - error: if the declaration is nil.
 *
 * Errors/Exceptions:
 *   - Returns an error if f.Declaration is nil.
 *
 * Side Effects:
 *   - Calls f.Declaration.ToStringForGPT() to perform conversion.
 */
func (f *FunctionInfo) ToStringForGPT() (string, error) {
    if f.Declaration == nil {
        return "", fmt.Errorf("can't convery %v to string for gpt. no delcaration", f.Name)
    }

    return f.Declaration.ToStringForGPT()
}


/**
 * Returns the documentation string for the FunctionInfo object. If no documentation is explicitly set,it retrieves any associated comments from the function declaration node.
 *
 * Signature:
 *   func (f *FunctionInfo) GetDocumentation() (string, error)
 *
 * Parameters:
 *   - f: pointer to a FunctionInfo instance.
 *
 * Returns:
 *   - string: the combined documentation text.
 *   - error: always nil.
 * 
 * Side Effects:
 *   - reads the FunctionInfo's Documentation field.
 *   - reads comment nodes from the declaration node if Documentation is empty.
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


/**
 * Performs a sanity check on the PackageNode by processing its errors and verifying the presence of syntax trees.
 *
 * Signature:
 * func (p *PackageNode) SanityCheck() error
 *
 * Parameters:
 * p - *PackageNode: the package node to validate.
 *
 * Returns:
 * error - an error indicating issues such as missing syntax trees or encountered errors.
 *
 * Errors/Exceptions:
 * Returns an error if p.Syntax is empty, indicating no syntax trees are present.
 *
 * Side Effects:
 * None.
 *
 * Edge Cases & Assumptions:
 * Assumes p.Errors may contain errors to be processed; the method currently creates error objects but does not handle or log them.
 * If p.Syntax is empty, it reports this as an error.
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


/**
Summary:
Populates package information by processing each syntax AST file, updating import map, type definitions, and function declarations accordingly.

Signature:
func (p *PackageNode) PopulatePackageInformation() error

Errors/Exceptions:
Returns an error if adding to import map, expanding type definitions, or expanding function declarations fails.
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


/**
 * Adds the imports from an AST file to the PackageNode's import map.
 * For each import, it uses aliasing if present; otherwise, it determines the package name via the build system.
 *
 * Signature: func (p *PackageNode) AddToImportMap(f_ast *ast.File) error
 *
 * Parameters:
 * - f_ast: *ast.File - The abstract syntax tree of a Go source file containing import declarations.
 *
 * Returns:
 * - error: Returns an error if the build system fails to import a package.
 *
 * Side Effects:
 * - Mutates the PackageNode's Imports map by adding new entries.
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


/**
 * Adds type definitions found within the given AST node to the PackageNode's TypeDefinitions slice.
 *
 * Signature:
 * func (p *PackageNode) AddToTypeDefinitions(f ast.Node) error
 *
 * Parameters:
 * - f: ast.Node - the AST node to inspect for type specifications.
 *
 * Returns:
 * - error: always nil.
 *
 * Side Effects:
 * - Updates p.TypeDefinitions by appending type specifications found within f.
 */

/**
 * Adds all *ast.TypeSpec nodes found within the provided AST node `f` to the PackageNode's TypeDefinitions slice.
 *
 * @param f ast.Node - the AST node to inspect for type specifications.
 * @return error - always returns nil.
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


/**
 * Adds function declarations from an AST file to the PackageNode, extracting call information within each function and creating corresponding FunctionDecl nodes.
 *
 * Signature: func (p *PackageNode) AddToFunctionDeclarations(f *ast.File) error
 *
 * Parameters:
 *   - f: *ast.File - The parsed AST file containing function declarations.
 *
 * Returns:
 *   - error: Returns an error if extraction of function invocations fails.
 *
 * Side Effects:
 *   - Modifies the PackageNode's FunctionDeclarations slice.
 *   - Populates FunctionDecls with new FunctionDecl objects linked to function call info.
 *   - Updates FunctionMap via CreateFunctionCall and CreateFunctionDecl.
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


/**
 * Summary: Creates a FunctionDecl object from a function declaration, establishing associated metadata and linking to existing function information if available.
 * Signature: func (p *PackageNode) CreateFunctionNodeFromDecl(f *ast.FuncDecl) *FunctionDecl
 * Parameters:
 *   - f: *ast.FuncDecl - The function declaration to process.
 *   - p: *PackageNode - The package context containing type info and current file.
 * Returns:
 *   - *FunctionDecl: The created FunctionDecl node with linked info.
 * Errors/Exceptions: None.
 * Side Effects: Updates FunctionMap with the new function info if it doesn't already exist.
 * Edge Cases & Assumptions: Assumes 'f' is a valid AST function declaration; 'TypesInfo' is properly populated. 
 */
func (p *PackageNode) CreateFunctionNodeFromDecl(f *ast.FuncDecl) *FunctionDecl {
    obj := ""

    typeName, found := MethodRecvNamed(f, p.TypesInfo)
    if found {
        obj = typeName.Obj().Id()
    }

    // Create a POSSIBLE object
    fInfo := &FunctionInfo{
        Package: p,

        Name: f.Name.String(),
        ResolvedPkg: p.ID,
        Object: obj,
        File: p.CurrentFile,

        Documentation: f.Doc.Text(),
        
        WasDocumented: f.Doc.Text() == "",
        Documented: false,
        AiAware: false,
    }

    // Check if we have a function definition set up in the map.
    //   This would happen if call was experienced before we saw definition
    if possibleInfo, exists := FunctionMap[fInfo.FullName()]; exists {
        fInfo = possibleInfo
    } else {
        FunctionMap[fInfo.FullName()] = fInfo
    }

    fDecl := &FunctionDecl {
        Info: fInfo,
        Node: f,

        Calls: []*FunctionCall{},
    }

    fInfo.Declaration = fDecl 

    return fDecl
}


/**
 * Creates a new FunctionDecl for the given ast.FuncDecl.
 * Initializes FunctionInfo with package, name, documentation, and file info.
 * Checks for existing FunctionInfo in FunctionMap and updates if found.
 * Associates the FunctionDecl with its FunctionInfo.
 * 
 * @param f: *ast.FuncDecl - the function declaration to process.
 * @return: *FunctionDecl - the constructed FunctionDecl with linked FunctionInfo.
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

        Name: f.Name.String(),
        ResolvedPkg: p.ID,
        Object: obj,
        File: p.CurrentFile,

        Documentation: f.Doc.Text(),
        
        WasDocumented: f.Doc.Text() != "",
        Documented: false,
        AiAware: false,
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

    fDecl := &FunctionDecl {
        Info: fInfo,
        Node: f,

        Calls: []*FunctionCall{},
    }

    fInfo.Declaration = fDecl 

    FunctionMap[fInfo.FullName()] = fInfo

    return fDecl
}


/**
Summary: Creates a FunctionCall object from a given ast.CallExpr, populating function information based on call expression type.

Signature: func (p *PackageNode) CreateFunctionCall(fun *ast.CallExpr) *FunctionCall

Parameters:
- fun: *ast.CallExpr - The call expression node to process.

Returns:
- *FunctionCall - The constructed FunctionCall with populated Node and Info.

Errors/Exceptions:
- Logs fatal error if fun.Fun is an unexpected type.

Side Effects:
- May modify the global FunctionMap with new FunctionInfo entries.

Edge Cases & Assumptions:
- Returns nil if the call expression represents a type cast (ArrayType).
- Assumes valid type switches on fun.Fun.
*/
func (p *PackageNode) CreateFunctionCall(fun *ast.CallExpr) *FunctionCall {
    // Make FCall and FInfo
    fInfo := &FunctionInfo { Package: p }
    fCall := &FunctionCall { Node: fun }

    // Populate FInfo and FCall so we can look things up
    switch fd := fun.Fun.(type) {
        case *ast.Ident:
            fCall.Kind = InternalCall;

            fInfo.Name = fd.Name;
            fInfo.ResolvedPkg = p.ID;

        case *ast.SelectorExpr:
            sel, _ := fd.X.(*ast.Ident);
            obj := p.TypesInfo.Uses[sel]

            if _, isPkg := obj.(*types.PkgName); isPkg { // Then its a package call
                fCall.Kind = PackageCall;

                pkg_name := types.ExprString(fd.X)

                fInfo.Name = fd.Sel.Name;
                fInfo.ResolvedPkg = p.Imports[pkg_name];


            } else { // Its an object call
                fCall.Kind = ObjectCall;

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


/**
 * Summary: Removes cyclic graphs within the package's function declarations by recursively clipping function call cycles.
 *
 * Signature: func (p *PackageNode) ClipCyclicGraphs() error
 *
 * Parameters:
 * - p: *PackageNode — the package node containing function declarations to process.
 *
 * Returns:
 * - error — returns an error if cycle clipping fails for any function; otherwise, nil.
 *
 * Errors/Exceptions:
 * - Returns an error with message "failed to clip function cycles" if an error occurs during cycle clipping of any function.
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


// I'm going to assume we can write docs for "deepest" node in call stack without the above.
// Empirical data will lmk if that's wrong.
// Generally, programmers should avoid this pattern, unless its recursive, and my strat 
//   works for the recursive case.
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

    for i := len(to_remove) - 1 ; i >= 0 ; i-- {
        f.Declaration.Calls = append(f.Declaration.Calls[:to_remove[i]], f.Declaration.Calls[to_remove[i] + 1:]...)
    }

    return nil
}

/**
 * Finds the start and end positions of an AST node within the source file.
 *
 * Signature:
 * func (p *PackageNode) FindStartEnd(n ast.Node) (int, int)
 *
 * Parameters:
 * - n: ast.Node; the node whose positions are to be determined.
 *
 * Returns:
 * - start: int; the file offset of the node's start position, or -1 if invalid.
 * - end: int; the file offset of the node's end position, or -1 if invalid.
 *
 * Errors/Exceptions:
 * - Returns (-1, -1) if the node is nil or its position cannot be determined.
 */

/**
 * Retrieves the start and end offsets of an AST node within the source file.
 * Returns (-1, -1) if the node is nil or its position cannot be determined.
 *
 * Signature: (p *PackageNode) FindStartEnd(n ast.Node) (int, int)
 *
 * Parameters:
 *   - n: ast.Node; the node to find offsets for. If nil, returns (-1, -1).
 *
 * Returns:
 *   - start: int; the offset where the node begins.
 *   - end: int; the offset where the node ends.
 *
 * Errors/Exceptions:
 *   - Returns (-1, -1) if n is nil or if the file cannot be found.
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

// Finds the start and end positions of the AST node within the source file.
// Signature:
// func (p *PackageNode) FindStartEnd(n ast.Node) (int, int)
// Parameters:
// - n: ast.Node; the node whose positions are to be determined.
// Returns:
// - start: int; the file offset of the node's start position, or -1 if invalid.
// - end: int; the file offset of the node's end position, or -1 if invalid.
// Errors/Exceptions:
// - Returns (-1, -1) if the node is nil or its position cannot be determined.
func (f *FunctionDecl) FindStartEnd() (int, int) {
    return f.Info.Package.FindStartEnd(f.Node)
}

/**
 * Finds the start and end positions of the AST node within the source file.
 *
 * Signature:
 *   func (p *PackageNode) FindStartEnd(n ast.Node) (int, int)
 *
 * Parameters:
 *   - n: ast.Node; the node whose positions are to be determined.
 *
 * Returns:
 *   - start: int; the file offset of the node's start position, or -1 if invalid.
 *   - end: int; the file offset of the node's end position, or -1 if invalid.
 *
 * Errors/Exceptions:
 *   - Returns (-1, -1) if the node is nil or its position cannot be determined.
 */

/**
 * Retrieves the start and end offsets of an AST node within the source file.
 * Returns (-1, -1) if the node is nil or its position cannot be determined.
 *
 * Signature:
 *   (p *PackageNode) FindStartEnd(n ast.Node) (int, int)
 *
 * Parameters:
 *   - n: ast.Node; the node to find offsets for. If nil, returns (-1, -1).
 *
 * Returns:
 *   - start: int; the offset where the node begins.
 *   - end: int; the offset where the node ends.
 *
 * Errors/Exceptions:
 *   - Returns (-1, -1) if n is nil or if the file cannot be found.
 */


/*
*
 * Retrieves the start and end offsets of the FunctionCall's associated AST node within the source file.
 * Returns (-1, -1) if the node is nil or if the position cannot be determined.
 *
 * Signature: (f *FunctionCall) FindStartEnd() (int, int)
 *
 * Parameters:
 *   - f: *FunctionCall; the object containing the AST node and package information.
 *
 * Returns:
 *   - start: int; the offset where the node begins.
 *   - end: int; the offset where the node ends.
 *
 * Errors/Exceptions:
 *   - Returns (-1, -1) if f.Node is nil or if the source position cannot be found.

*/
func (f *FunctionCall) FindStartEnd() (int, int) {
    return f.Info.Package.FindStartEnd(f.Node)
}


/**
 * Updates missing documentation comments for function declarations within the package node.
 *
 * Iterates through the package's function declarations in reverse order.
 * For each function without existing documentation, it calculates the start position
 * of the function's AST node and inserts the associated documentation string into the file at that position.
 *
 * Parameters:
 *   - p: *PackageNode; the package node containing function declarations.
 *
 * Returns:
 *   - error: nil if all updates succeed; otherwise, an error indicating failure.
 *
 * Errors/Exceptions:
 *   - Returns an error if reading or writing a file fails, or if insertion position is invalid.
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




/**
 * GetFunctionInvocations extracts all function call expressions from the provided AST node.
 *
 * @param f ast.Node: The AST node to inspect.
 * @return funcs []*ast.CallExpr: Slice of all call expressions found.
 *         error: Always nil.
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


/**
 * Finds and returns the named type of the receiver in a method declaration.
 *
 * This function inspects the method declaration `fd` and, if available, uses type
 * information from `info` to determine the receiver's named type. It first attempts
 * to extract the receiver type via the method's signature; if unsuccessful, it
 * falls back to syntactic analysis and type info lookups.
 *
 * Parameters:
 *   - fd: *ast.FuncDecl - the function declaration to analyze.
 *   - info: *types.Info - type information for resolving identifiers.
 *
 * Returns:
 *   - *types.Named: the receiver's named type if found; otherwise, nil.
 *   - bool: true if the named type was successfully identified; false otherwise.
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


/**
Summary:
Parses a Go package located in the specified folder, performing sanity checks, populating package information, and clipping cyclic function call graphs.

Signature:
func ParsePackage(foldername string) ([]PackageNode, error)

Parameters:
- foldername: string — the path to the package folder to parse.

Returns:
- []PackageNode: a slice of parsed PackageNode objects with populated function declarations, type definitions, and import mappings.
- error: an error if package loading, sanity checking, populating information, or cycle clipping fails.

Errors/Exceptions:
- Returns an error if loading the package fails.
- Returns an error if sanity checks or population of package information fail.
- Returns an error if clipping cyclic graphs within function call graphs fails.

Side Effects:
- Loads package data and updates PackageNode fields.
- Performs in-place cycle clipping on function call graphs.

Edge Cases & Assumptions:
- Assumes the specified folder contains a valid Go package.
- The function handles errors from each processing step explicitly.
*/
func ParsePackage(foldername string) ([]PackageNode, error) {
    cfg := &packages.Config{
        Mode: packages.NeedName            |
              packages.NeedFiles           | 
              packages.NeedSyntax          | 
              packages.NeedCompiledGoFiles |
              packages.NeedSyntax          |
              packages.NeedTypes           |
              packages.NeedTypesInfo       |
              packages.NeedImports         |
              packages.NeedDeps            ,
    }

    pkgs, err := packages.Load(cfg, foldername)
    if err != nil {
        return nil, fmt.Errorf("failed to load package %v: %v", foldername, err)
    }

    pkgNodes := []PackageNode{}

    log.Debugf("# packages seen: %v", len(pkgs))

    for _, pkg := range pkgs {
        pkgNode := PackageNode{
            Package: pkg,
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

