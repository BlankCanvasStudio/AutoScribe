package ast;

import (
    "fmt"

    "bytes"
    "slices"
    "strings"

    "go/ast"
    "go/token"
    "go/types"
    "go/build"
    "go/printer"

    "golang.org/x/tools/go/packages"

    log "github.com/sirupsen/logrus"

    asTypes "github.com/BlankCanvasStudio/AutoScribe/pkg/types"
)


/*
*
*  Try handling variable functions from another package / module
*
*/

type FunctionKind string


var FunctionMap = map[string]*FunctionNode{}


const (
    ObjectCall     FunctionKind = "object"
    PackageCall    FunctionKind = "package"
    InternalCall   FunctionKind = "internal"
    FnDeclaration  FunctionKind = "declaration"
)


type FunctionNode struct {
    Kind          FunctionKind
    Name          string
    Package       string
    File          string
    Object        string
    Documented    bool
    AiAware      bool
    Documentation string
    Calls         []*FunctionNode
    Node          ast.Node
    Language      asTypes.SupportedFormat
}


/**
 * Summary: Returns the full name of the function, including package and object if present.
 * Signature: func (f *FunctionNode) FullName() string
 * Returns: A string representing the full name, formatted as "Package.Object.Name" if Object is non-empty; otherwise "Package.Name".
 * Side Effects: None.
 * Edge Cases & Assumptions: Assumes Package and Name are non-empty strings; Object may be empty.
 */
func (f *FunctionNode) FullName() string {
    if f.Object != "" {
        return fmt.Sprintf("%s.%s.%s", f.Package, f.Object, f.Name)
    }

    return fmt.Sprintf("%s.%s", f.Package, f.Name)
}

/**
 * PrettyPrint outputs a formatted representation of the FunctionNode and its calls to standard output.
 * Use it for debugging or visual inspection of function call structures.
 *
 * func (f *FunctionNode) PrettyPrint(prefix string)
 *
 * @param prefix string: indentation or prefix string for formatting output.
 */
func (f *FunctionNode) PrettyPrint(prefix string) {
    fmt.Println("")
    fmt.Println("")

    if f.Object == "" {
        fmt.Printf("%v %v (%v)\n", prefix, f.Name, f.Kind)
    } else {
        fmt.Printf("%v %v.%v (%v)\n", prefix, f.Object, f.Name, f.Kind)
    }

    fmt.Printf("%v File: %v\n", prefix, f.File)
    fmt.Printf("%v Package: %v\n", prefix, f.Package)
    fmt.Printf("%v Node: %+v\n", prefix, f.Node)
    if f.Documentation != "" {
        fmt.Printf("%v Documentation:\n%v\n", prefix, f.Documentation)
    }

    for _, called := range f.Calls {
        fmt.Println("")
        fmt.Println("")
        called.PrettyPrint(prefix + "\t")
    }
}


/**
 * Converts the FunctionNode's AST node to its source code representation suitable for GPT processing.
 * Uses the Go printer to generate a string from the node, assuming only one level of recursion.
 * 
 * @param f *FunctionNode: The function node to convert.
 * @return string: The source code string of the node.
 * @return error: Returns nil if conversion succeeds; otherwise, an error if printing fails.
 */
func (f *FunctionNode) ToStringForGPT() (string, error) {
    // This should only be one layer deep. We are using comments to avoid the recursion
    var buf bytes.Buffer
    printer.Fprint(&buf, token.NewFileSet(), f.Node)
    src := buf.String()

    return src, nil
}


type PackageNode struct {
    *packages.Package
    FunctionDeclarations []*FunctionNode
    TypeDefinitions      []*ast.TypeSpec
    Imports              map[string]string
    CurrentFile          string
}


/*
SanityCheck verifies the integrity of the PackageNode by checking for errors and the presence of syntax trees. Returns an error if issues are found; otherwise, returns nil.
*/
func (p *PackageNode) SanityCheck() error {
    for _, err := range p.Errors {
        return fmt.Errorf("Error in %v: %v", p.ID, err)
    }

    if len(p.Syntax) == 0 {
        return fmt.Errorf("No syntax trees in %v", p.ID)
    }

    return nil
}


/**
* Populates package information by processing each syntax AST in the PackageNode.
* Updates import map, type definitions, and function declarations for each AST.
* Returns an error if any step fails during processing.
* Sets p.CurrentFile to the corresponding compiled Go file for each AST.
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

        log.Debugf("Defines %v function(s)", len(p.FunctionDeclarations))
    }

    return nil
}


/**
Adds import declarations from an AST file to the PackageNode's Imports map.
Use when updating the PackageNode with new import statements.
@param f_ast *ast.File: AST of the Go source file containing import declarations.
@return error if importing a package fails; otherwise nil.
@side Effects: modifies p.Imports.
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
 * Adds all *ast.TypeSpec nodes within the provided ast.Node to the PackageNode's TypeDefinitions slice.
 * Use to accumulate type definitions contained in a given AST subtree.
 *
 * @param f ast.Node: AST node representing code to process (e.g., a file or package scope).
 * @return error: always returns nil.
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
 * Summary: Adds function declarations from the provided AST file to the PackageNode, creating corresponding FunctionNode instances and recording function calls.
 *
 * Signature:
 * func (p *PackageNode) AddToFunctionDeclarations(f *ast.File) error
 *
 * Parameters:
 *  - f: *ast.File - The AST file containing function declarations to add.
 *
 * Returns:
 *  - error: An error if retrieving function invocations fails.
 *
 * Errors/Exceptions:
 *  - Returns an error if GetFunctionInvocations encounters an issue.
 */
func (p *PackageNode) AddToFunctionDeclarations(f *ast.File) error {
    if p.FunctionDeclarations == nil {
        p.FunctionDeclarations = []*FunctionNode{}
    }

    // Get all the function Declarations from the AST
    funcs := make([]*ast.FuncDecl, 0, 10)

    ast.Inspect(f, func(n ast.Node) bool {
        fd, ok := n.(*ast.FuncDecl)
        if ok {
            funcs = append(funcs, fd)
        }

        return true
    })

    for _, node := range funcs {
        // Create a new function node for the newly declared function
        newFuncNode := p.CreateFunctionNodeFromDecl(node)
        // if newFuncNode.Name == "PrettyPrint" {
        //     continue
        // }

        // Properly save it to the global list without invalidating existing pointers
        possibleFuncNode, exists := FunctionMap[newFuncNode.FullName()]
        if exists { // If it exists in function map, something else has referenced it, but it hasn't been declared yet
            // Set the value at the pointer in the map to the full, updated value (only 1 delcaration possible)
            (*possibleFuncNode) = (*newFuncNode)
            // Set the newFuncNode pointer equal to the pointer in the dictionary
            newFuncNode = possibleFuncNode
        }

        // Only really does something when !exists
        FunctionMap[newFuncNode.FullName()] = newFuncNode

        // Save our newly declared function to the package object
        p.FunctionDeclarations = append(p.FunctionDeclarations, newFuncNode)

        // Get all the function invocations call in this function
        invocations, err := GetFunctionInvocations(node)
        if err != nil {
            return fmt.Errorf("failed to get function invocations: %v", err)
        }

        Calls := make([]*FunctionNode, len(invocations))

        for i, invoc := range invocations {
            // Create a new node from the call
            newNode := p.CreateFunctionNodeFromCall(invoc)
            // Check if the node exists in the map
            recordedNode, exists := FunctionMap[newNode.FullName()]

            if exists { // use recorded
                Calls[i] = recordedNode
            } else { // use new node and add to map
                Calls[i] = newNode
                FunctionMap[newNode.FullName()] = newNode
            }
        }

        newFuncNode.Calls = Calls
    }

    return nil
}


/**
 * Creates a new FunctionNode from a given *ast.FuncDecl, capturing relevant details.
 * Use this when converting a function declaration in the AST to a FunctionNode.
 *
 * @param f *ast.FuncDecl - the function declaration AST node.
 * @return *FunctionNode - the constructed function node with metadata and documentation.
 *
 * The created node includes function name, package ID, current file, associated object name (if method), AST node, and documentation.
 */
func (p *PackageNode) CreateFunctionNodeFromDecl(f *ast.FuncDecl) *FunctionNode {
    obj := ""

    typeName, found := MethodRecvNamed(f, p.TypesInfo)
    if found {
        obj = typeName.Obj().Id()
    }


    return &FunctionNode {
        Kind: FnDeclaration,
        Name: f.Name.String(),
        Package: p.ID,
        File: p.CurrentFile,
        Calls: []*FunctionNode{},
        Object: obj,
        Node: f,
        Documentation: f.Doc.Text(),
    }
}


// CreateFunctionNodeFromCall creates a FunctionNode representing the called function.
// It distinguishes between internal calls, package calls, and fallback conversions.
// Use it to generate a FunctionNode for a given ast.CallExpr based on call type.
func (p *PackageNode) CreateFunctionNodeFromCall(fun *ast.CallExpr) *FunctionNode {
    // Fast track for internal package function calls
    if sel, ok := fun.Fun.(*ast.Ident); ok {
        return &FunctionNode {
            Name: sel.Name,
            Package: p.ID,
            Kind: InternalCall,
            Node: fun,
        }
    }

    // If X is an identifier that resolves to a *types.PkgName, it's pkg.Sym.
    if sel, ok := fun.Fun.(*ast.SelectorExpr); ok {
        if id, ok := sel.X.(*ast.Ident); ok {
            if obj := p.TypesInfo.Uses[id]; obj != nil {
                if _, isPkg := obj.(*types.PkgName); isPkg {
                    pkg_name := types.ExprString(sel.X)

                    return &FunctionNode {
                            Name: sel.Sel.Name,
                            Package: p.Imports[pkg_name],
                            Kind: PackageCall,
                            Node: fun,
                    }
                }
            }
        }
    }

    // This must be an object function call / member variable
    node, _ := ConvertToFunctionNode(fun, p.Fset, p.TypesInfo, "")
    return node
}


/**
 * Cleans up cyclic graph references within the package's function declarations.
 * Iterates through each FunctionDeclaration and removes cycles by invoking ClipFunctionCycles.
 *
 * @receiver p *PackageNode: Pointer to the package node containing function declarations.
 * @return error: Returns an error if cycle clipping fails in any function.
 */
func (p *PackageNode) ClipCyclicGraphs() error {
    for _, decl := range p.FunctionDeclarations {
        callStack := []string{}
        err := p.ClipFunctionCycles(decl, callStack)
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
func (p *PackageNode) ClipFunctionCycles(f *FunctionNode, callStack []string) error {
    to_remove := []int{}
    for i, call := range f.Calls {
        // Remove the repeated node from the calls array, but don't descend it.
        // If you descend it, all the nodes above it will be removed as well (since they've 
        //   already been included in the list)
        if slices.Contains(callStack, call.FullName()) {
            to_remove = append(to_remove, i)
        } else {
            p.ClipFunctionCycles(call, append(callStack, f.FullName()))
        }
    }

    for i := len(to_remove) - 1 ; i >= 0 ; i-- {
        f.Calls = append(f.Calls[:to_remove[i]], f.Calls[to_remove[i] + 1:]...)
    }

    return nil
}

/**
 * FindStartEnd returns the start and end byte offsets of an AST node if it's a function declaration.
 * Use it to locate code positions within a source file.
 *
 * Signature:
 * func (p *PackageNode) FindStartEnd(n ast.Node) (int, int)
 *
 * Parameters:
 * - n: ast.Node; the node to evaluate; expected to be a *ast.FuncDecl
 *
 * Returns:
 * - start offset: int; position offset where the node begins
 * - end offset: int; position offset where the node ends
 *
 * Errors/Exceptions:
 * - Logs fatal error and terminates if the node is not a *ast.FuncDecl
 *
 * Side Effects:
 * - Logs an error message on invalid node type
 *
 * Edge Cases & Assumptions:
 * - Assumes the node is a function declaration; otherwise, the program terminates
 */
func (p *PackageNode) FindStartEnd(n ast.Node) (int, int) {
    if f, ok := n.(*ast.FuncDecl); ok {
        return p.Fset.Position(f.Pos()).Offset, p.Fset.Position(f.End()).Offset
    }

    log.Fatalf("tried to find start & end of unknown type")
    return -1, -1
}

/**
 * Updates documentation comments in the associated file for each function declaration in the PackageNode.
 * Skips functions that already have documentation comments.
 * Returns an error if a function declaration is not of type *ast.FuncDecl or if file update fails.
 */
func (p *PackageNode) UpdateDocsInFile() error {

    for i := len(p.FunctionDeclarations) - 1; i >= 0; i-- {
        f := p.FunctionDeclarations[i]

        fd, ok := f.Node.(*ast.FuncDecl)
        if !ok {
            return fmt.Errorf("p.FunctionDeclarations top level object not *ast.FuncDecl")
        }

        // We read in pre-existing docs
        if fd.Doc != nil {
            continue
        }

        start, _ := p.FindStartEnd(fd)

        toAdd := fmt.Sprintf("%v\n", f.Documentation)

        err := insertIntoFile(f.File, start, toAdd)
        if err != nil {
            return fmt.Errorf("failed to update docs in file: %v", err)
        }
    }

    return nil
}




/**
 * GetFunctionInvocations traverses an AST node to find all function call expressions.
 * Use it to collect all function invocations within the given AST node.
 *
 * func GetFunctionInvocations(f ast.Node) ([]*ast.CallExpr, error)
 *
 * @param f ast.Node - the root AST node to inspect.
 * 
 * @return slice of *ast.CallExpr - all call expressions found; empty if none.
 * @return error - always nil; provided for interface compatibility.
 *
 * No side effects or errors are thrown.
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
 * Converts an *ast.CallExpr to a *FunctionNode, extracting function name, object type, package path, and related info.
 * Use when transforming AST call expressions into a structured function node representation.
 *
 * @param fun Pointer to ast.CallExpr representing the function call.
 * @param fset Pointer to token.FileSet, used for position info (not directly used here).
 * @param info Pointer to types.Info, containing type information for AST nodes.
 * @param file String representing the filename of the source code.
 * @return Pointer to FunctionNode with populated fields, or an error if conversion fails.
 */
func ConvertToFunctionNode(fun *ast.CallExpr, fset *token.FileSet, info *types.Info, file string) (*FunctionNode, error) {
    sel, ok := fun.Fun.(*ast.SelectorExpr)
    if !ok {
        return nil, fmt.Errorf("failed to parse object call: %v", fun)
    }
    
    selInfo := info.Selections[sel]
    recv := selInfo.Recv()
    if p, ok := recv.(*types.Pointer); ok {
        recv = p.Elem()
    }

    var typeName, pkgPath string
    if n, ok := recv.(*types.Named); ok {
    typeName = n.Obj().Name()
    if n.Obj().Pkg() != nil {
        pkgPath = n.Obj().Pkg().Path()
    }
    } else {
        typeName = types.TypeString(recv, func(*types.Package) string { return "" })
    }

    obj := selInfo.Obj()
    funcName := obj.Name()
    if pkgPath == "" && obj.Pkg() != nil {
        pkgPath = obj.Pkg().Path()
    }

    return &FunctionNode {
        Name: funcName,
        Object: typeName,
        Package: pkgPath,
        File: file,
        Kind: ObjectCall,
        Node: fun,
    }, nil
}


/*
func GetTypeDefinitionPositions(n *ast.TypeSpec, fset *token.FileSet) (int, int) {
    return fset.Position(n.Pos()).Offset - 5, fset.Position(n.End()).Offset
}
*/


/**
 * Extracts the named type associated with a receiver of a function declaration.
 * Returns the named type and true if found; otherwise, nil and false.
 *
 * @param fd Pointer to ast.FuncDecl representing the function.
 * @param info Pointer to types.Info containing type information.
 * @return *types.Named and bool indicating if the receiver's type was identified.
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
// ParsePackage loads and processes Go packages in the specified folder, returning a slice of PackageNode with populated information.
// It uses the Go 'packages' package with a comprehensive configuration to load all relevant package data.
// If loading fails, returns an error indicating the failure.
// Iterates over loaded packages to initialize PackageNode instances, performs sanity checks, and populates package info.
// Clips cyclic graphs within each PackageNode; returns error if clipping fails.
// Returns a slice of fully processed PackageNode objects.
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

    for _, pkg := range pkgs {
        pkgNode := PackageNode{
            Package: pkg,
            FunctionDeclarations: []*FunctionNode{},
            TypeDefinitions:      []*ast.TypeSpec{},
            Imports:              make(map[string]string),
        }

        pkgNode.SanityCheck()

        pkgNode.PopulatePackageInformation()

        /*
        for _, value := range pkgNode.FunctionDeclarations {
            log.Infof("%+v", value)
        }
        */

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

