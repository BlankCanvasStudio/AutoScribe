package ast

import (
    "os"
    "fmt"
    "go/ast"
    "go/token"
    "go/parser"
    "go/printer"
)

func Undocument(filename string) error {
    fset := token.NewFileSet()
    file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
    if err != nil {
        return fmt.Errorf("failed to parse file: %v", err)
    }

    // Walk functions and clear their leading doc comments
    for _, decl := range file.Decls {
        if fn, ok := decl.(*ast.FuncDecl); ok {
            fn.Doc = nil
        }
    }

    // Print back the code without those comments
    printer.Fprint(os.Stdout, fset, file)

    return nil
}

