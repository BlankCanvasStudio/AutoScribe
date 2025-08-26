package ast

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

/*
*
Removes documentation comments from functions and code comments in all Go source files within the specified directory. Use this to strip function doc comments and associated comments from Go files, optionally including test files.

Signature:
func UndocumentDir(dir string, includeTests bool) error

Parameters:
- dir: string - the directory path containing Go files to process.
- includeTests: bool - whether to include test files (`*_test.go`) in processing.

Returns:
- error: non-nil if parsing, formatting, or writing files fails.

Errors/Exceptions:
- Returns an error if directory parsing, file formatting, or writing fails.

Side Effects:
- Modifies source files by removing function documentation comments and associated comments.

*/
func UndocumentDir(dir string, includeTests bool) error {
	fset := token.NewFileSet()

	filter := func(fi os.FileInfo) bool {
		name := fi.Name()
		if !strings.HasSuffix(name, ".go") {
			return false
		}
		if !includeTests && strings.HasSuffix(name, "_test.go") {
			return false
		}
		return true
	}

	pkgs, err := parser.ParseDir(fset, dir, filter, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse dir %q: %v", dir, err)
	}

	for _, pkg := range pkgs {
		for filename, file := range pkg.Files {
			// Collect func doc comment groups to remove and clear fn.Doc
			rm := map[*ast.CommentGroup]struct{}{}
			for _, decl := range file.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok && fn.Doc != nil {
					rm[fn.Doc] = struct{}{}
					fn.Doc = nil
				}
			}

			// Also drop those groups from file.Comments so they don't get printed
			if len(rm) > 0 && len(file.Comments) > 0 {
				filtered := make([]*ast.CommentGroup, 0, len(file.Comments))
				for _, cg := range file.Comments {
					if _, drop := rm[cg]; !drop {
						filtered = append(filtered, cg)
					}
				}
				file.Comments = filtered
			}

			var buf bytes.Buffer
			if err := format.Node(&buf, fset, file); err != nil {
				return fmt.Errorf("format %q: %v", filename, err)
			}
			if err := os.WriteFile(filename, buf.Bytes(), 0o644); err != nil {
				return fmt.Errorf("write %q: %v", filename, err)
			}
		}
	}
	return nil
}
