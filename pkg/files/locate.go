package files

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"

	log "github.com/sirupsen/logrus"
)

/*
Summary: Recursively search the provided folder for Go source files and return a
sorted, deduplicated list of absolute directories that contain at least one .go file.
The function logs the search and collects the directory of each .go file found.

Signature: func GetNestedFoldersWithGoFiles(folder string) ([]string, error)

Parameters:
- folder string: root folder to search for .go files.

Returns:
- []string: sorted, deduplicated absolute directory paths containing at least one .go file.
- error: non-nil if the directory walk fails; the error is wrapped with context.

Errors/Exceptions:
- non-nil error if filepath.WalkDir returns an error (e.g., permission denied, missing path).
- error during filepath.Abs(path) propagation causes a non-nil error return.

Side Effects:
- Logs: log.Debugf("folders searching: %v", folder).
- Allocates memory for the result and the internal seen map.

Edge Cases & Assumptions:
- If no .go files are found, returns []string{} with nil error.
- Returned paths are absolute; directories are deduplicated and sorted.
- .go files in nested subdirectories are accounted for via their parent directories.

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
