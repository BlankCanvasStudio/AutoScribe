package files;

import (
    "fmt"
    "sort"
    "io/fs"
    "path/filepath"

    log "github.com/sirupsen/logrus"
)

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
