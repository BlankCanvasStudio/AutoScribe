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
Summary: UndocumentDir recursively searches the provided folder for Go source files and removes
documentation comments from top-level function declarations, rewriting the files in place.
If includeTests is false, test files ending with _test.go are skipped. Directories are discovered
via GetNestedFoldersWithGoFiles and only directories containing at least one .go file are processed.

Signature: func UndocumentDir(dir string, includeTests bool) error

Parameters:
- dir string: root folder to search for Go source files.
- includeTests bool: if true, include _test.go files in parsing; if false, skip test files.

Returns:
- error: non-nil if processing fails at any stage; otherwise nil.

Errors/Exceptions:
- non-nil error when GetNestedFoldersWithGoFiles fails to return directories;
- non-nil error if parsing a directory with parser.ParseDir fails;
- non-nil error if formatting a file with format.Node fails;
- non-nil error if writing a modified file with os.WriteFile fails.

Side Effects:
- Mutates source files on disk by removing function doc comments and their associated file.Comments.
- Overwrites each touched file with the dedented, formatted version.

Edge Cases & Assumptions:
- If a Go file has no function doc comments, the file is unchanged.
- If no directories contain Go files under dir, the function completes with nil.
- All returned directories are absolute paths; the function relies on GetNestedFoldersWithGoFiles for discovery.
- The filter used when parsing includes only .go files and optionally excludes _test.go based on includeTests.

Notes:
- The implementation parses each directory’s AST, clears Doc from top-level function declarations, prunes corresponding comment groups, formats the AST, and writes back to disk.

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
Summary
UndocumentFile removes GoDoc comments from the Go source file identified by filename, preserving inline comments. It deletes package-level documentation and documentation associated with declarations (functions and general declarations), then reformats and writes the updated file back to disk.

Signature
func UndocumentFile(filename string) error

Parameters
- filename: string — path to the Go source file to process.

Returns
- error: non-nil if parsing, formatting, or writing fails; nil on success.

Errors/Exceptions
- parse <filename>: error if the Go file cannot be parsed (including syntax errors).
- format <filename>: error if the AST could not be formatted.
- write <filename>: error if the updated content cannot be written to disk.

Side Effects
- Mutates the file at filename by removing GoDoc comments (package and declaration docs) and writing the formatted result back to disk.

Edge Cases & Assumptions
- Removes only GoDoc groups collected from file.Doc and from d.Doc for *ast.FuncDecl and *ast.GenDecl; inline/comments preserved where not part of a GoDoc group.
- If there are no GoDoc sections detected, the file may be reformatted but otherwise unchanged.
- Assumes filename points to a readable/writable, valid Go source file.

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
Summary: Choose between undocumenting a directory or a single file. If path is a directory, it delegates to UndocumentDir with includeTests; otherwise, it delegates to UndocumentFile.

Signature: func Undocument(path string, includeTests bool) error

Parameters:
- path string: path to a file or directory to process.
- includeTests bool: when path is a directory, controls whether _test.go files are included in parsing; ignored when path is a file.

Returns:
- error: non-nil if stat fails or if the delegated UndocumentDir/UndocumentFile returns an error; nil on success.

Errors/Exceptions:
- non-nil error if os.Stat(path) fails.
- non-nil error if UndocumentDir(path, includeTests) fails (when path is a directory).
- non-nil error if UndocumentFile(path) fails (when path is a file).

Side Effects:
- May mutate the filesystem by updating Go source files via UndocumentDir or UndocumentFile.
- Writes modified files back to disk as part of the undocumentation process.

Edge Cases & Assumptions:
- If path denotes a directory, the function delegates to UndocumentDir; if not, it delegates to UndocumentFile.
- The behavior and errors of UndocumentDir/UndocumentFile are propagated to the caller.
- If path does not exist or is inaccessible, an error from os.Stat is returned.

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
