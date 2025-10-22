package files

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"

	log "github.com/sirupsen/logrus"
)

/*
Summary:
GetNestedFoldersWithGoFiles recursively scans the provided folder for Go source files (*.go) and returns the unique set of directories that contain at least one such file. The returned paths are absolute and sorted ascending.

Signature:
GetNestedFoldersWithGoFiles(folder string) ([]string, error)

Parameters:
- folder: string. Path to the root directory to search recursively for Go files.

Returns:
- []string: absolute directory paths containing at least one .go file, sorted ascending. Empty slice if none are found.
- error: non-nil if the directory walk or path resolution fails.

Errors/Exceptions:
- If filepath.Abs(path) fails for a matched .go file, the error is returned.
- If filepath.WalkDir encounters an error, the function returns fmt.Errorf("failed to walk directory for go files: %v", err).

Side Effects:
- Logs a debug message: log.Debugf("folders searching: %v", folder).
- Traverses the filesystem and builds a deduplicated set of directories containing Go files.

Edge Cases & Assumptions:
- Only files with the exact .go extension are considered; the check is case-sensitive.
- If no .go files exist, the function returns an empty slice and nil error.
- Directories are collected from the absolute path of each discovered .go file; duplicates are eliminated via a seen map, and results are sorted before returning.

*/
func GetNestedFoldersWithGoFiles(folder string) ([]string, error) {
	seen := make(map[string]struct{})

	log.Debugf("folders searching: %v", folder)

	err := filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {

		if !d.IsDir() && filepath.Ext(path) == ".go" {
			abs, err := filepath.Abs(path)
			if err != nil {
				return err
			}

			dir := filepath.Dir(abs)

			if _, ok := seen[dir]; !ok {
				seen[dir] = struct{}{}
			}
		}

		return nil
	})

	if err != nil {
		return []string{}, fmt.Errorf("failed to walk directory for go files: %v", err)
	}

	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys, nil
}
