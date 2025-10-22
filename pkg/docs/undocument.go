package docs

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	// "go/printer"
	"go/token"
	"os"
	"strings"

	"github.com/BlankCanvasStudio/AutoScribe/pkg/files"
)

/*
Summary:
UndocumentDir removes Go doc comments from all functions in Go source files under dir, recursively. It also clears the corresponding function Doc and related file-level comment groups, then rewrites and formats the modified files. If includeTests is false, _test.go files are skipped.

Signature:
UndocumentDir(dir string, includeTests bool) error

Parameters:
- dir: string. Root directory to process recursively for Go files.
- includeTests: bool. If false, files matching *_test.go are ignored.

Returns:
- error. Non-nil on failure. Causes include:
  - failure to obtain Go-containing directories,
  - failure to parse a directory's Go files,
  - failure to format a file via format.Node,
  - failure to write a modified file with os.WriteFile.

Errors/Exceptions:
- Propagates descriptive errors from internal steps, such as:
  - "failed to get all the go directories from %v: %v"
  - "failed to parse dir %q: %v"
  - "format %q: %v"
  - "write %q: %v"

Side Effects:
- Mutates source files on disk by removing function doc comments and related comment groups, then rewrites and saves the updated files (permissions 0644).

Edge Cases & Assumptions:
- Only files with the exact .go extension are considered; the check is case-sensitive.
- If no Go files are found, the function returns nil without modifications.
- Only function documentation comments (fn.Doc) are removed; other non-function comments remain unless tied to a function declaration.
- Files are reformatted after modification to ensure consistent formatting.

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

	dirs, err := files.GetNestedFoldersWithGoFiles(dir)
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

/*
Summary: UndocumentFile removes GoDoc documentation comments from a Go source file and rewrites it without those docs. It targets the package doc and top-level declaration docs, preserving inline comments.
Signature: func UndocumentFile(filename string) error
Parameters:
  - filename: string — path to the Go source file to process; constraints: must be an existing, readable and writable Go source file.
Returns: error — non-nil if parsing, formatting, or writing fails; nil on success.
Errors/Exceptions:
  - parse error: returned as fmt.Errorf("parse %q: %v", filename, err)
  - format error: returned as fmt.Errorf("format %q: %v", filename, err)
  - write error: returned as fmt.Errorf("write %q: %v", filename, err)
Side Effects:
  - Mutates the file on disk by overwriting it with a version that has GoDoc comments removed (package doc and declaration docs); preserves non-doc inline comments and formatting is normalized.
Edge Cases & Assumptions:
  - Assumes the input is valid Go source; if there are no GoDoc comments, the file may still be reformatted.
  - Inline comments are kept; only GoDoc comment groups (package/decl docs) are removed.

*/
func UndocumentFile(filename string) error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse %q: %v", filename, err)
	}

	// Collect only GoDoc comment groups (package/decl docs) to remove
	rm := map[*ast.CommentGroup]struct{}{}

	// Package doc (e.g., file header GoDoc)
	if file.Doc != nil {
		rm[file.Doc] = struct{}{}
		file.Doc = nil
	}

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Doc != nil {
				rm[d.Doc] = struct{}{}
				d.Doc = nil
			}
		case *ast.GenDecl: // const/var/type declarations
			if d.Doc != nil {
				rm[d.Doc] = struct{}{}
				d.Doc = nil
			}
		}
	}

	// Filter out only those doc groups from file.Comments; keep inline comments
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
	return nil
}

/*
Summary:
Undokument processes the path by removing Go doc comments. If path is a directory, it delegates to UndocumentDir(path, includeTests) to recursively strip documentation from Go sources and rewrite files. If path is a file, it delegates to UndocumentFile(path) to remove package and declaration GoDoc comments from that file. When path is a directory, includeTests controls whether *_test.go files are included; when path is a file, includeTests is ignored.

Signature:
func Undocument(path string, includeTests bool) error

Parameters:
- path: string — Path to a directory or Go source file to process.
- includeTests: bool — If path is a directory, include or exclude *_test.go files during processing.

Returns:
- error — Non-nil on failure. Errors may originate from os.Stat, or from UndocumentDir or UndocumentFile.

Errors/Exceptions:
- Propagates errors from:
  - "failed to stat %v: %v" if stat fails
  - errors returned by UndocumentDir or UndocumentFile

Side Effects:
- Mutates on-disk Go source files by removing GoDoc comments, then rewrites and formats the modified files (permissions as in the called helpers).

Edge Cases & Assumptions:
- If path is a directory, includeTests governs test file handling; if a file, includeTests is ignored.
- Only the presence of a directory or file is checked; actual Go source parsing/formatting is performed by UndocumentDir/UndocumentFile.

*/
func Undocument(path string, includeTests bool) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat %v: %v", path, err)
	}

	if info.IsDir() {
		return UndocumentDir(path, includeTests)
	}

	return UndocumentFile(path)
}
