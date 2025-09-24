package mst;

import (
    "os"
    "io"
    "fmt"
    "bufio"
    "bytes"
    "errors"
    "strings"
    "path/filepath"
    "crypto/sha256"

    log "github.com/sirupsen/logrus"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/ai/call"
)

var IsDocumentedDb string = fmt.Sprintf("/etc/autoscribe/db/is-documented.txt")
var DocumentationDb string = "/etc/autoscribe/db/documentation.txt"

var IsAiAwareDb string = "/etc/autoscribe/db/is-ai-aware.txt"
var NotAiAwareDb string = "/etc/autoscribe/db/not-ai-aware.txt"

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

    GetResolvedPackageName() string
    SetResolvedPackageName(string)

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
    
    SetDocumentation(string) error
    GetDocumentation() (string, error)

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

func GetFullNameHash(f FunctionDetails) [32]byte {
    return sha256.Sum256([]byte(f.GetFullName()))
}

var IsAiAwarePrompt string = `
Do you already know what the function '%v' does?  
Answer "Yes" if you’ve seen it before and can describe its purpose,  
or "No" if you can’t.  
Don't respond with anything else
`

func IsAiAware(f FunctionDetails, ApiKey string) bool {
    hash := GetFullNameHash(f)

    isThere, _, err := CheckFileForHash(IsAiAwareDb, hash)
    if err != nil {
        log.Errorf("Failed to check AI Aware db for hash: %v", err)
    }

    if isThere {
        return true;
    }

    isThere, _, err = CheckFileForHash(NotAiAwareDb, hash)
    if err != nil {
        log.Errorf("Failed to check Not AI Aware db for hash: %v", err)
    }

    if isThere {
        return false
    }

    // Check if function definition is from the current package. If it is, then its not AI aware
    if strings.Contains(f.GetFullName(), f.GetInfo().GetPackage().GetResolvedPackageName()) {
        err := InsertHash(NotAiAwareDb, hash)
        if err != nil {
            log.Errorf("failed to insert into %v: %v", NotAiAwareDb, err)
        }

        return false
    }

    // Actually check via AI
    fullMsg := fmt.Sprintf(IsAiAwarePrompt, f.GetFullName())
    yesOrNo, err := call.Query41Nano(fullMsg, ApiKey, []string{"Yes", "No"})
    if err != nil { // Not sure I like this
        log.Errorf("failed to query 4.1 Nano: %v", err)
        return false
    }

    // Add it to the proper database so we aren't constantly looking things up
    if yesOrNo == "Yes" {
        InsertHash(IsAiAwareDb, hash)
        if err != nil {
            log.Errorf("failed to insert into %v: %v", IsAiAwareDb, err)
        }

        return true
    }

    InsertHash(NotAiAwareDb, hash)
    if err != nil {
        log.Errorf("failed to insert into %v: %v", NotAiAwareDb, err)
    }

    return false
}

func IsDocumented(f FunctionDetails) bool {
    hash := GetFullNameHash(f)

    isThere, index, err := CheckFileForHash(IsDocumentedDb, hash)
    if err != nil {
        log.Errorf("Failed to check AI Aware db for hash: %v", err)
    }
    
    // Not in db
    if ! isThere {
        return false
    }

    // in db and docs loaded
    if docs, _ := f.GetDocumentation(); docs != "" {
        return true;
    }

    // load the docs into the function details
    docs, err := ReadLine(DocumentationDb, int(index))

    f.SetDocumentation(docs)

    return true
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

func DocumentMST (m MST, ApiKey string) (MST, error) {
    err := m.HandleCyclicDependencies()
    if err != nil {
        return m, fmt.Errorf("failed to handle cyclic dependencies: %v", err)
    }

    for _, pkg := range m.GetPackages() {
        for _, fd := range pkg.GetFunctionDecls() {

            err := Document(fd, ApiKey)
            if err != nil {
                return m, fmt.Errorf("failed to document func %v: %v", fd.GetName(), err)
            }
        }
    }

    return m, nil
}

func Document(f FunctionDetails, ApiKey string) error {
	// Consider how gpt aware gets loaded
	// if IsAiAware(f) || IsDocumented(f) || WasDocumented(f) {
	if IsDocumented(f) || IsAiAware(f, ApiKey) {
		return nil
	}

        decl, err := GetFuncDecl(f)
        if err != nil {
            return fmt.Errorf("failed to get function declaration from %v: %v", (f).GetFullName(), err)
        }

        // Undecided about this
	if decl == nil {
		log.Debugf("No declaration for `%v`. Assuming its defined in another package...", (f).GetName())
		return nil
	}

        // Make sure we have all the nested documentation
        for _, call := range decl.GetCalls() {
            tInfo := call.GetInfo()
            if !IsDocumented(tInfo) && !IsAiAware(tInfo, ApiKey) {
                // Recursively document if we need to
                err := Document(tInfo, ApiKey)
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

        _ = NodeAsAiText
        /*
        log.Infof("FunctionNode: %+v", f)
        log.Infof("Node %v as AI text:\n%v", f.GetName(), NodeAsAiText)
        */

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

        f.SetDocumentation(f.GetName())

        f.SetDocumentedInThisPass(true)

	return nil
}


func CheckFileForHash(filename string, hash [32]byte) (exists bool, index int64, err error) {
    _ = filename
    _ = hash

    info, err := os.Stat(filename)
    if err != nil {
        if errors.Is(err, os.ErrNotExist) {
            return false, -1, nil
        }

        return false, -1, fmt.Errorf("failed to stat %v: %v", filename, err)
    }

    // hash plus new line
    const entry_len int64 = 32 + 1

    // No entries
    if info.Size() < entry_len {
        return false, -1, nil
    }

    // Get the number of hashes in the database
    entries := info.Size() / entry_len // hashes are 32 bytes long

    top := entries - 1;
    bottom := int64(0);

    f, err := os.Open(filename)
    if err != nil {
        return false, -1, fmt.Errorf("failed to open %v: %v", filename, err)
    }
    defer f.Close()

    var val = make([]byte, 32);

    for bottom <= top {
        mid := (top + bottom) / 2

        offset := mid * entry_len

        if _, err := f.ReadAt(val[:], offset); err != nil {
	    return false, -1, fmt.Errorf("read 32 bytes at %d: %w", offset, err)
	}

        cmp := bytes.Compare(val[:], hash[:])
        switch {
        case cmp == 0:
            return true, mid, nil
        case cmp < 0:
            bottom = mid + 1
        default:
            if mid == 0 {
                    return false, -1, nil
            }
            top = mid - 1
        }
    }

    return false, -1, nil
}

// InsertHash keeps file sorted by lexicographic order of 32-byte records (plus newline).
func InsertHash(filename string, hash [32]byte) error {
	const entryLen int64 = 32 + 1

	f, err := os.Open(filename)
        if err != nil {
            if os.IsNotExist(err) {
                if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
                    return fmt.Errorf("mkdir %q: %w", filepath.Dir(filename), err)
                }
                if err := os.WriteFile(filename, append(hash[:], '\n'), 0644); err != nil {
                    return fmt.Errorf("create %q: %w", filename, err)
                }
                return nil
            }
            return fmt.Errorf("open %q: %w", filename, err)
        }
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat %q: %w", filename, err)
	}
	entries := info.Size() / entryLen

	// Binary search for insert position
	var val [32]byte
	lo, hi := int64(0), entries-1
	insertIdx := entries // default: append at end
	for lo <= hi {
		mid := (lo + hi) / 2
		offset := mid * entryLen
		if _, err := f.ReadAt(val[:], offset); err != nil {
			return fmt.Errorf("read at %d: %w", offset, err)
		}
		cmp := bytes.Compare(val[:], hash[:])
		if cmp == 0 {
			// Already exists → nothing to do
			return nil
		} else if cmp < 0 {
			lo = mid + 1
		} else {
			insertIdx = mid
			if mid == 0 {
				break
			}
			hi = mid - 1
		}
	}

	// Temp file for rewrite
	tmpName := filename + ".tmp"
	tf, err := os.Create(tmpName)
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}

	// Copy up to insertion point
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		tf.Close()
		return err
	}
	if _, err := io.CopyN(tf, f, insertIdx*entryLen); err != nil && err != io.EOF {
		tf.Close()
		return fmt.Errorf("copy head: %w", err)
	}

	// Write new entry
	if _, err := tf.Write(hash[:]); err != nil {
		tf.Close()
		return fmt.Errorf("write hash: %w", err)
	}
	if _, err := tf.Write([]byte{'\n'}); err != nil {
		tf.Close()
		return fmt.Errorf("write newline: %w", err)
	}

	// Copy tail
	if _, err := f.Seek(insertIdx*entryLen, io.SeekStart); err != nil {
		tf.Close()
		return err
	}
	if _, err := io.Copy(tf, f); err != nil {
		tf.Close()
		return fmt.Errorf("copy tail: %w", err)
	}

	if err := tf.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, filename)
}


func ReadLine(filename string, lineNum int) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	cur := 1
	for scanner.Scan() {
		if cur == lineNum {
			return scanner.Text(), nil
		}
		cur++
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("line %d not found", lineNum)
}
