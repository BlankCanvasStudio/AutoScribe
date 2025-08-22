package ast;

import (
    "fmt"

    "slices"
    "strings"

    "go/ast"
    "go/token"
    "go/types"
    "go/build"

    "golang.org/x/tools/go/packages"

    log "github.com/sirupsen/logrus"
)


/*
*
* YOU AREN'T HANDLING RECURSION PROPERLY!
*   Try handling variable functions from another package / module
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
    GPTAware      bool
    Documentation string
    Calls         []*FunctionNode
    Node          ast.Node
}


func (f *FunctionNode) FullName() string {
    if f.Object != "" {
        return fmt.Sprintf("%s.%s.%s", f.Package, f.Object, f.Name)
    }

    return fmt.Sprintf("%s.%s", f.Package, f.Name)
}

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

    for _, called := range f.Calls {
        fmt.Println("")
        fmt.Println("")
        called.PrettyPrint(prefix + "\t")
    }
}


type PackageNode struct {
    *packages.Package
    FunctionDeclarations []*FunctionNode
    TypeDefinitions      []*ast.TypeSpec
    Imports              map[string]string
    CurrentFile          string
}


func (p *PackageNode) SanityCheck() error {
    for _, err := range p.Errors {
        return fmt.Errorf("Error in %v: %v", p.ID, err)
    }

    if len(p.Syntax) == 0 {
        return fmt.Errorf("No syntax trees in %v", p.ID)
    }

    return nil
}


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


func (p *PackageNode) CreateFunctionNodeFromDecl(f *ast.FuncDecl) *FunctionNode {
    log.Infof("func decl %v type: %+v", f.Name, f.Body)

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
    }
}


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
              packages.NeedDeps,
    }

    pkgs, err := packages.Load(cfg, "./cmd", "./pkg/ast")
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

        for _, value := range pkgNode.FunctionDeclarations {
            log.Infof("%+v", value)
        }

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

