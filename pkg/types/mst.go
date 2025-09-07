package types;

import (
    "fmt"

)

type FunctionCallKind string

const (
	ObjectCall   FunctionCallKind = "object"
	PackageCall  FunctionCallKind = "package"
	InternalCall FunctionCallKind = "internal"
)

type MST interface {
    GetPackages() []*PackageNode
    SetPackages([]*PackageNode) error

    AddPackage(*PackageNode) error

    AddToFunctionMap(string, *FunctionInfo) error

    GetFromFunctionMap(string) (*FunctionInfo, error)

    UpdateFunctionMap(string, *FunctionInfo) error

    PrettyPrint(string)
}


type PackageNode interface {
    GetMST() *MST

    GetFunctionDecl() []*FunctionDecl
    SetFunctionDecl([]*FunctionDecl) error

    GetImports() map[string]string
    AddToImports(string, string) error

    GetCurrentFile() string
    SetCurrentFile(string) error

    // TypeDefinitions      []*ast.TypeSpec
    PrettyPrint(string)
}


type FunctionNode interface {
    FullName() string

    ToStringForAi() string

    GetName() string

    GetInfo() *FunctionInfo

    GetDecl() *FunctionDecl
}


type FunctionInfo interface {
    FunctionNode

    GetPackage() *PackageNode

    SetName(string) error

    GetResolvedPkg() string
    SetResolvedPkg(string) error

    GetObject() (string, error) // Not sure about this one
    SetObject(string) error

    GetFile() string
    SetFile(string) error

    GetDocumentation() (string, error)
    SetDocumentation(string) error

    WasDocumented() bool
    SetWasDocumented(bool) error

    DocumentedInThisPass() bool
    SetDocumentedInThisPass(bool) error

    IsAiAware() bool
    SetIsAiAware(bool) error

    GetDeclaration() (*FunctionDecl, error)
    SetDeclaration(*FunctionDecl) error

    PrettyPrint(string)
}


type FunctionDecl interface {
    FunctionNode

    SetInfo(*FunctionInfo) error
    GetInfo() *FunctionInfo

    /*
    GetNode() *ast.FuncDecl
    SetNode() *ast.FuncDecl
    */

    GetCalls() []*FunctionCall
    SetCalls([]*FunctionCall) error
    AddCall(*FunctionCall) error

    PrettyPrint(string)
}


type FunctionCall interface {
    FunctionNode

    SetInfo(*FunctionInfo) error
    GetInfo() *FunctionInfo

    /*
    GetNode() *ast.FuncDecl
    SetNode() *ast.FuncDecl
    */

    GetKind() FunctionCallKind
    SetKind(FunctionCallKind) error

    PrettyPrint(string)
}

func GetDocumentation(f FunctionNode) (string, error) {
    fi := f.GetInfo()
    if fi == nil {
        return "", fmt.Errorf("function info isn't in function node")
    }

    return (*fi).GetDocumentation()
}

func FullName(f FunctionNode) (string, error) {
    fi := f.GetInfo()
    if fi == nil {
        return "", fmt.Errorf("function info isn't in function node")
    }

    return (*fi).FullName(), nil
}

func DocumentedThisPass(f FunctionNode) bool {
    fi := f.GetInfo()
    if fi == nil {
        return false
    }

    return (*fi).DocumentedInThisPass()
}

func WasDocumented(f FunctionNode) bool {
    fi := f.GetInfo()
    if fi == nil {
        return false
    }

    return (*fi).WasDocumented()
}

func IsAiAware(f FunctionNode) bool {
    fi := f.GetInfo()
    if fi == nil {
        return false
    }

    return (*fi).IsAiAware()
}

func GetFuncDecl(f FunctionNode) (*FunctionDecl, error) {
    fi := f.GetInfo()
    if fi == nil {
        return nil, fmt.Errorf("FunctionDecl was nil in FunctionInfo")
    }

    return (*fi).GetDeclaration()
}


