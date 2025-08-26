package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/BlankCanvasStudio/AutoScribe/pkg/config"
)

/*
*
 * Summary: Recursively searches the project directory for files matching the configured language file extension.
 * Use this function to obtain a list of relevant code files for processing.
 *
 * Signature: func FilterForCodeFiles(directory string) ([]string, error)
 *
 * Parameters:
 * - directory: string, the path to the project directory to search.
 *
 * Returns:
 * - []string: list of file paths matching the language extension.
 * - error: error encountered during directory traversal or if no language files are found.
 *
 * Errors/Exceptions:
 * - Returns an error if filepath.Walk encounters an issue.
 * - Returns an error if no matching files are found.
 *
 * Side Effects:
 * - Logs debug information about the filtering process.
 *
 * Edge Cases & Assumptions:
 * - Skips ".git" directories.
 * - Only includes files that pass the extension check.

*/
func FilterForCodeFiles(directory string) ([]string, error) {
	log.Debugf("Filtering for %v code files in: %v", config.LanguageFileExtension, config.ProjectDirectory)

	// Collect all the files in question
	var files []string

	err := filepath.Walk(config.ProjectDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			if ext, is := config.LanguageFileExtension.FileIsThisFormat(path); !is {
				log.Debugf("File %v with extension `%v` doens't pass language filter. Ignoring...\n", path, ext)
			} else {
				files = append(files, path)
			}
		} else {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
		}

		return nil
	})

	if err != nil {
		return files, fmt.Errorf("Failed to walk %v: %v", config.ProjectDirectory, err)
	}

	if len(files) == 0 {
		return files, fmt.Errorf("Cannot AutoScribe: language set to `%v` but none found.", config.LanguageFileExtension)
	}

	return files, nil
}

/*
*
 * Filters and collects build-related files within the specified project directory.
 *
 * @param directory string - The path of the directory to search in.
 * @return ([]string, error) - A slice of file paths matching build files and an error if the walk fails.
 *
 * The function searches through the project directory, ignoring ".git" directories,
 * and gathers paths containing "Makefile", "build.sh", "configure.sh", or "deps.sh".
 * Returns an empty slice and logs a debug message if no such files are found.

*/
func FilterForBuildFiles(directory string) ([]string, error) {
	// Collect all the files in question
	var files []string

	err := filepath.Walk(config.ProjectDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}

			return nil
		}

		if strings.Contains(path, "Makefile") {
			files = append(files, path)
		} else if strings.Contains(path, "build.sh") {
			files = append(files, path)
		} else if strings.Contains(path, "configure.sh") {
			files = append(files, path)
		} else if strings.Contains(path, "deps.sh") {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return files, fmt.Errorf("Failed to walk %v: %v", config.ProjectDirectory, err)
	}

	if len(files) == 0 {
		log.Debug("No build files found in %v", config.ProjectDirectory)
	}

	return files, nil
}
