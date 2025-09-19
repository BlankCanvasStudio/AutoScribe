package mst;

import (
    "os"
    "fmt"
    "strings"

    log "github.com/sirupsen/logrus"
)

type FunctionCallKind string

const (
	ObjectCall   FunctionCallKind = "object"
	PackageCall  FunctionCallKind = "package"
	InternalCall FunctionCallKind = "internal"
)

type MST interface {
    GetPackages() []PackageNode
    SetPackages([]PackageNode) error

    AddPackage(PackageNode) error

    AddToFunctionMap(string, FunctionInfo) error

    GetFromFunctionMap(string) (FunctionInfo, bool, error)
    // GetFunctionMap() map[string]FunctionInfo

    Populate(folder []string) error

    HandleCyclicDependencies() error

    PrettyPrint(string)
}


type PackageNode interface {
    GetMST() MST

    GetFunctionDecls() []FunctionDecl
    SetFunctionDecls([]FunctionDecl) error

    GetTypeDefs() []TypeDefinition
    SetTypeDefs([]TypeDefinition) error

    GetImports() map[string]string
    AddToImports(string, string) error

    GetPath() string
    SetPath(string) error

    ClipFunctionCycles(FunctionInfo, []string) error   

    // TypeDefinitions      []*ast.TypeSpec
    PrettyPrint(string)
}


type FunctionDetails interface {
    GetFullName() string

    GetName() string

    GetInfo() FunctionInfo

    GetDecl() FunctionDecl

    GetFile() string


    GetDocInsertLocation() uint

    SetDocumentedInThisPass(bool)

    CreateComment(string) string

    GetCalls() []FunctionCall

    ToStringForAi() (string, error)
}

type FunctionNode interface {
    FunctionDetails

    FindLineNo() int

    ToStringForAi() (string, error)
    FindStartEnd() (int, int)
}


type FunctionInfo interface {
    FunctionDetails

    GetPackage() PackageNode

    GetName() string
    SetName(string) error

    GetResolvedPkg() string
    SetResolvedPkg(string) error

    GetObjectName() (string, error)
    SetObjectName(string) error

    GetFile() string
    SetFile(string) error

    SetDocumentation(string) error
    GetDocumentation() (string, error)

    SetDocumentedInThisPass(bool)

    GetHasDocumentation() bool
    SetHasDocumentation(bool) error

    GetDocumentedInThisPass() bool

    GetIsAiAware() bool
    SetIsAiAware(bool) error

    // GetDeclaration() (FunctionDecl, error)
    SetDecl(FunctionDecl) error

    PrettyPrint(string)
}


type FunctionDecl interface {
    FunctionNode

    SetInfo(FunctionInfo) error
    GetInfo() FunctionInfo

    /*
    GetNode() *ast.FuncDecl
    SetNode() *ast.FuncDecl
    */

    SetCalls([]FunctionCall) error
    AddCall(FunctionCall) error

    PrettyPrint(string)
}


type FunctionCall interface {
    FunctionNode

    SetInfo(FunctionInfo) error
    GetInfo() FunctionInfo

    /*
    GetNode() *ast.FuncDecl
    SetNode() *ast.FuncDecl
    */

    GetKind() FunctionCallKind
    SetKind(FunctionCallKind) error

    PrettyPrint(string)
}

type TypeDefinition interface {
    GetName() string
}

func GetDocumentation(f FunctionDetails) (string, error) {
    fi := f.GetInfo()
    if fi == nil {
        return "", fmt.Errorf("function info isn't in function node")
    }

    return fi.GetDocumentation()
}

func GetFullName(f FunctionDetails) (string, error) {
    fi := f.GetInfo()
    if fi == nil {
        return "", fmt.Errorf("function info isn't in function node")
    }

    return fi.GetFullName(), nil
}

func DocumentedThisPass(f FunctionDetails) bool {
    fi := f.GetInfo()
    if fi == nil {
        return false
    }

    return fi.GetDocumentedInThisPass()
}

func HasDocumentation(f FunctionDetails) bool {
    fi := f.GetInfo()
    if fi == nil {
        return false
    }

    return fi.GetHasDocumentation()
}

func IsAiAware(f FunctionDetails) bool {
    fi := f.GetInfo()
    if fi == nil {
        return false
    }

    return fi.GetIsAiAware()
}

func GetFuncDecl(f FunctionDetails) (FunctionDecl, error) {
    fi := f.GetInfo()
    if fi == nil {
        return nil, fmt.Errorf("FunctionDecl was nil in FunctionInfo")
    }

    return fi.GetDecl(), nil
}


func UpdateDocumentation(f FunctionDetails) error {
    // Don't need to do anything if its AI aware or documented in a previous pass
    if !DocumentedThisPass(f) {
        return nil
    }

    path := f.GetFile()

    data, err := os.ReadFile(path)
    if err != nil {
        return fmt.Errorf("failed to load file %v: %v", path, err)
    }

    offset := f.GetDocInsertLocation()

    if offset > uint(len(data)) {
        return fmt.Errorf("DocInsertLocation (%v) is out of range of len data (%v)", offset, len(data))
    }

    toAdd, err := GetDocumentation(f)
    if err != nil {
        return fmt.Errorf("failed to get documentation: %v", f.GetName())
    }

    out := append(append([]byte{}, data[:offset]...), append([]byte(toAdd), data[offset:]...)...)

    return os.WriteFile(path, out, 0644)
}


func ToStringForAi(f FunctionNode) (string, error) {

	// // This should only be one layer deep. We are using comments to avoid the recursion

	raw, err := os.ReadFile(f.GetFile())
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}

	fd_start, fd_end := f.FindStartEnd()
	fd_text := string(raw)[fd_start:fd_end]

        calls := f.GetCalls()

	for i := len(calls) - 1; i >= 0; i-- {
		fc_line_no := calls[i].FindLineNo()
		fc_start, fc_end := calls[i].FindStartEnd()

		fc_start -= fd_start
		fc_end -= fd_start
		fc_line_no -= fd_start

		docs, err := calls[i].GetInfo().GetDocumentation()
                if err != nil {
                    return "", fmt.Errorf("failed to get docs: %v", err)
                }

		if strings.TrimSpace(docs) == "" {
			continue
		}
		fd_text = fd_text[:fc_line_no] + f.CreateComment(docs) + fd_text[fc_line_no:]
	}

	return fd_text, nil
}

func DocumentMST (m MST) (MST, error) {
    err := m.HandleCyclicDependencies()
    if err != nil {
        return m, fmt.Errorf("failed to handle cyclic dependencies: %v", err)
    }

    for _, pkg := range m.GetPackages() {
        for _, fd := range pkg.GetFunctionDecls() {

            err := Document(fd)
            if err != nil {
                return m, fmt.Errorf("failed to document func %v: %v", fd.GetName(), err)
            }
        }
    }

    return m, nil
}

func Document(f FunctionDetails) error {
	// Consider how gpt aware gets loaded
        /*
	if IsAiAware(f) || IsDocumented(f) || WasDocumented(f) {
		return nil

	}
        */

        /*
	if !DocumentedThisPass(f) {
		return nil

	}
        */

        decl, err := GetFuncDecl(f)
        if err != nil {
            return fmt.Errorf("failed to get function declaration from %v: %v", (f).GetFullName(), err)
        }

        // Undecided about this
	if decl == nil {
		log.Debugf("No declaration for `%v`. Assuming its defined in another package...", (f).GetName())
		return nil
	}

        // This is almost certainly an issue
        for _, call := range decl.GetCalls() {
            tInfo := call.GetInfo()
            if !HasDocumentation(tInfo) && !IsAiAware(tInfo) {
                // Recursively document if we need to
                err := Document(tInfo)
                // err := DocumentFunctions(*tInfo)
                if err != nil {
                    return fmt.Errorf("failed to document call %v in %v: %v", tInfo.GetName(), (f).GetName(), err)
                }
            }
        }

        // By this point all nodes are either GPT aware or documented
        NodeAsAiText, err := f.ToStringForAi()
        if err != nil {
                return fmt.Errorf("failed to convert FunctionNode to AI string: %v", err)
        }

        log.Infof("FunctionNode: %+v", f)
        log.Infof("Node %v as AI text:\n%v", f.GetName(), NodeAsAiText)

        /*
        DocumentationString, err := calls.QueryFromFile(GPT_41_Nano, config.DocsPrompt, NodeAsAiText)
        if err != nil {
            return fmt.Errorf("failed to query from file: %v", err)
        }

        log.Debugf("result from gpt: `%v`", DocumentationString)

        DocumentedComment, err := formatting.FormatAsGoComment(DocumentationString)
        if err != nil {
                return fmt.Errorf("failed to parse for comments: %v", err)
        }

        f.SetDocumentation(DocumentedComment)
        */

        f.SetDocumentedInThisPass(true)

	return nil
}


