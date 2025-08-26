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
 * summary: Removes documentation comments from function declarations within Go source files
 *          located in the specified directory and its subdirectories (optionally excluding test files).
 *          It updates the source files by stripping out function documentation comments and saving the changes.
 *
 * signature: func UndocumentDir(dir string, includeTests bool) error
 *
 * parameters:
 * - dir: string; the root directory to scan for Go source files.
 * - includeTests: bool; whether to include test files ('_test.go') in processing.
 *
 * returns:
 * - error: non-nil if an error occurs during directory traversal, parsing, formatting, or writing files.
 *
 * errors/exceptions:
 * - Returns an error if directory traversal, parsing, formatting, or file writing fails.
 *
 * side effects:
 * - Modifies source files by removing function documentation comments.
 * - Writes updated source files back to disk.
 *
 * edge cases & assumptions:
 * - Assumes input 'dir' exists and is accessible.
 * - Only processes files ending with ".go" (and optionally excluding "_test.go" files).

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

	dirs, err := GetNestedFoldersWithGoFiles(dir)
	if err != nil {
		return fmt.Errorf("failed to get all the go directories from %v: %v", dir, err)
	}

	for _, found_dir := range dirs {
		pkgs, err := parser.ParseDir(fset, found_dir, filter, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("failed to parse dir %q: %v", found_dir, err)
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
	}

	return nil
}
