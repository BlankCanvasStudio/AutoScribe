package golang;

import (
    "os"
    "fmt"
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
    Package     mst.PackageNode

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
    Info mst.FunctionInfo
    Calls []mst.FunctionCall

    DocInsertLocation uint

    Package     mst.PackageNode

    Node *ast.FuncDecl
}

type FunctionCall struct {
    Info mst.FunctionInfo
    Kind mst.FunctionCallKind

    Package     mst.PackageNode

    Node *ast.CallExpr
}

type TypeDef struct {
    Name string
}


func (m *MST) GetPackages() []mst.PackageNode {
    return m.PackageNodes
}


func (m *MST) SetPackages(pkgs []mst.PackageNode) error {
    m.PackageNodes = pkgs
    return nil
}


func (m *MST) AddPackage(n mst.PackageNode) error {
    if m.PackageNodes == nil {
        m.PackageNodes = make([]mst.PackageNode, 0)
    }

    m.PackageNodes = append(m.PackageNodes, n)

    return nil
}


func (m *MST) AddToFunctionMap(name string, info mst.FunctionInfo) error {
    if m.FunctionMap == nil {
        m.FunctionMap = make(map[string]mst.FunctionInfo)
    }

    m.FunctionMap[name] = info

    return nil
}


func (m *MST) GetFromFunctionMap(name string) (mst.FunctionInfo, bool, error) {
    if m.FunctionMap == nil {
        m.FunctionMap = make(map[string]mst.FunctionInfo)
    }

    val, ok := m.FunctionMap[name]

    return val, ok, nil
}


func (m *MST) PrettyPrint(string) {

}

func(m *MST) Populate(folders []string) error {
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

    pkgNodes := []mst.PackageNode{}

    for _, f := range folders {
        pkgs, err := packages.Load(cfg, f)

        for _, pkg := range pkgs {
            pkgNode := &PackageNode {
                MST: m,
                Package: pkg,
                FunctionDecls: []mst.FunctionDecl{},
                // TypeDefinitions:      []*ast.TypeSpec{},
                Imports:              make(map[string]string),
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


func(m *MST) HandleCyclicDependencies() error {
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


func (p *PackageNode) GetMST() mst.MST {
    return p.MST
}


func (p *PackageNode) GetFunctionDecls() []mst.FunctionDecl {
    return p.FunctionDecls
}


func (p *PackageNode) SetFunctionDecls(decls []mst.FunctionDecl) error {
    p.FunctionDecls = decls
    return nil
}


func (p *PackageNode) GetImports() map[string]string {
    return p.Imports
}


func (p *PackageNode) AddToImports(short string, fqn string) error {
    if p.Imports == nil {
        p.Imports = make(map[string]string)
    }

    p.Imports[short] = fqn

    return nil
}


func (p *PackageNode) GetTypeDefs() []mst.TypeDefinition {
    return p.TypeDefs
}


func (p *PackageNode) SetTypeDefs(in []mst.TypeDefinition) error {
    p.TypeDefs = in
    return nil
}


func (p *PackageNode) GetPath() string {
    return p.PkgPath
}

func (p *PackageNode) SetPath(s string) error {
    return fmt.Errorf("do not set path %v on golang packages", s)
}


func (p *PackageNode) SetCurrentFile(file string) error {
    return fmt.Errorf("don't set path on go packages; I got it")
}


// func (*PackageNode) TypeDefinitions      []*ast.TypeSpec
func (p *PackageNode) PrettyPrint(string) {

}


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
                    td := &TypeDef{ Name: fd.Name.Name }
		    p.TypeDefs = append(p.TypeDefs, td)
		}

		return true
	})

	return nil
}



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

		HasDocumentation:         f.Doc.Text() != "",
		DocumentedInThisPass:    false,
		IsAiAware:               false,
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
		Info: fInfo,
		Node: f,
                Package: p,

		Calls: []mst.FunctionCall{},
	}

	fInfo.Declaration = fDecl

	p.GetMST().AddToFunctionMap(fInfo.GetFullName(), fInfo)

        fInfo.SetFile(p.CurrentFile)

	return fDecl
}


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


func (f *FunctionInfo) GetInfo() mst.FunctionInfo {
    return f
}

func (f *FunctionInfo) GetPackage() mst.PackageNode {
    return f.Package
}

func (f *FunctionInfo) GetName() string {
    return f.Name
}

func (f *FunctionInfo) SetName(name string) error {
    f.Name = name
    return nil
}

func (f *FunctionInfo) GetFullName() string {
	if f.Object != "" {
		return fmt.Sprintf("%s.%s.%s", f.GetResolvedPkg(), f.Object, f.Name)
	}

	return fmt.Sprintf("%s.%s", f.GetResolvedPkg(), f.Name)
}

func (f *FunctionInfo) GetResolvedPkg() string {
    if f.ResolvedPkg == "" {
        return f.Package.GetPath()
    }
    return f.ResolvedPkg
}

func (f *FunctionInfo) SetResolvedPkg(in string) error {
    f.ResolvedPkg = in
    return nil
}

func (f *FunctionInfo) GetObjectName() (string, error) {
    return f.Object, nil
}

func (f *FunctionInfo) SetObjectName(name string) error {
    f.Object = name
    return nil
}

func (f *FunctionInfo) GetFile() string {
    return f.File
}

func (f *FunctionInfo) SetFile(file string) error {
    f.File = file
    return nil
}

func (f *FunctionInfo) SetDocumentation(docs string) error {
    f.Documentation = docs
    return nil
}

func (f *FunctionInfo) GetDocumentation() (string, error) {
    return f.Documentation, nil
}

func (f *FunctionInfo) GetHasDocumentation() bool {
    return f.HasDocumentation
}

func (f *FunctionInfo) SetHasDocumentation(t bool) error {
    f.HasDocumentation = t
    return nil
}

func (f *FunctionInfo) GetDocumentedInThisPass() bool {
    return f.DocumentedInThisPass
}

func (f *FunctionInfo) SetDocumentedInThisPass(t bool) {
    f.DocumentedInThisPass = t
}

func (f *FunctionInfo) GetIsAiAware() bool {
    return f.IsAiAware
}

func (f *FunctionInfo) SetIsAiAware(t bool) error {
    f.IsAiAware = t
    return nil
}

func (f *FunctionInfo) GetDecl() mst.FunctionDecl {
    return f.Declaration
}

func (f *FunctionInfo) SetDecl(d mst.FunctionDecl) error {
    f.Declaration = d
    return nil
}

func (f *FunctionInfo) GetCalls() []mst.FunctionCall {
    return f.Declaration.GetCalls()
}

func (f *FunctionInfo) PrettyPrint(string) {

}

func (f *FunctionInfo) GetDocInsertLocation() uint {
    return f.GetDecl().GetDocInsertLocation()
}

func (f *FunctionInfo) CreateComment(docs string) string {
    return "/*" + docs + "*/"
}

func (f *FunctionDecl) CreateComment(docs string) string {
    return f.GetInfo().CreateComment(docs)
}

func (f *FunctionCall) CreateComment(docs string) string {
    return f.GetInfo().CreateComment(docs)
}

func (f *FunctionDecl) SetInfo(i mst.FunctionInfo) error {
    f.Info = i
    return nil
}

func (f *FunctionDecl) GetInfo() mst.FunctionInfo {
    return f.Info
}

func (f *FunctionDecl) GetName() string {
    return f.Info.GetName()
}

func (f *FunctionDecl) GetFullName() string {
	return f.Info.GetFullName()
}


func (f *FunctionDecl) GetDocInsertLocation() uint {
    return f.DocInsertLocation
}


/*
func (f *FunctionDecl) GetNode() *ast.FuncDecl
func (f *FunctionDecl) SetNode() *ast.FuncDecl
*/

func (f *FunctionDecl) GetCalls() []mst.FunctionCall {
    return f.Calls
}

func (f *FunctionDecl) SetCalls(c []mst.FunctionCall) error {
    f.Calls = c
    return nil
}

func (f *FunctionDecl) AddCall(c mst.FunctionCall) error {
    if f.Calls == nil {
        f.Calls = make([]mst.FunctionCall, 0)
    }

    f.Calls = append(f.Calls, c)
    return nil
}


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

func (f *FunctionInfo) ToStringForAi() (string, error) {
    return f.GetDecl().ToStringForAi()
}

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
		fd_text = fd_text[:fc_line_no] + " /* " + docs + " */\n " + fd_text[fc_line_no:]
	}

	return fd_text, nil
}

func (f *FunctionDecl) PrettyPrint(string) {

}


func (f *FunctionCall) SetInfo(i mst.FunctionInfo) error {
    f.Info = i
    return nil
}

func (f *FunctionCall) GetInfo() mst.FunctionInfo {
    return f.Info
}

func (f *FunctionCall) GetName() string {
    return f.Info.GetName()
}

func (f *FunctionCall) GetFullName() string {
	return f.Info.GetFullName()
}

/*
func (f *FunctionCall) GetNode() *ast.FuncDecl
func (f *FunctionCall) SetNode() *ast.FuncDecl
*/

func (f *FunctionCall) GetKind() mst.FunctionCallKind {
    return f.Kind
}

func (f *FunctionCall) SetKind(k mst.FunctionCallKind) error {
    f.Kind = k
    return nil
}

func (f *FunctionCall) PrettyPrint(string) {

}

func (f *FunctionCall) GetDocInsertLocation() uint {
    return f.GetDecl().GetDocInsertLocation()
}

func (f *FunctionDecl) FindLineNo() int {
    return f.Package.(*PackageNode).FindLineNo(f.Node)
}

func (f *FunctionCall) FindLineNo() int {
    return f.Package.(*PackageNode).FindLineNo(f.Node)
}

func (f *FunctionDecl) FindStartEnd() (int, int) {
    return f.Info.(*FunctionInfo).Package.(*PackageNode).FindStartEnd(f.Node)
}

func (f *FunctionCall) FindStartEnd() (int, int) {
    return f.Info.(*FunctionInfo).Package.(*PackageNode).FindStartEnd(f.Node)
}

func (f *FunctionCall) GetCalls() ([]mst.FunctionCall) {
    return f.Info.GetCalls()
}

func (f *FunctionCall) GetDecl() mst.FunctionDecl {
    return f.Info.GetDecl()
}

func (f *FunctionCall) GetFile() string {
    return f.Info.GetFile()
}

func (f *FunctionCall) SetDocumentedInThisPass(in bool) {
    f.Info.SetDocumentedInThisPass(in)
}

func (f *FunctionCall) AddCall(in bool) {
    f.Info.SetDocumentedInThisPass(in)
}

func (f *FunctionCall) ToStringForAi() (string, error) {
    return f.Info.GetDecl().ToStringForAi()
}

func (f *FunctionDecl) GetDecl() mst.FunctionDecl {
    return f
}

func (f *FunctionDecl) GetFile() string {
    return f.Info.GetFile()
}

func (f *FunctionDecl) SetDocumentedInThisPass(in bool) {
    f.Info.SetDocumentedInThisPass(in)
}

func (t *TypeDef) GetName() string {
    return t.Name
}

