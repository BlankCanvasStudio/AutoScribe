package formatting

import (
	"fmt"
	"os"
        "path/filepath"

	// "github.com/BlankCanvasStudio/AutoScribe/pkg/types"
	// "github.com/BlankCanvasStudio/AutoScribe/pkg/config"
)

/*
*
 * Combines multiple files into a single ConcatenatedFileContents string,
 * appending each file's name and contents.
 *
 * Signature:
 * func CombineFilesForContext(files []string) (types.ConcatenatedFileContents, error)
 *
 * Parameters:
 * - files: []string - a list of file paths to read and concatenate.
 *
 * Returns:
 * - types.ConcatenatedFileContents: concatenated content with file headers.
 * - error: non-nil if reading any file fails.
 *
 * Errors/Exceptions:
 * - Returns an error if reading a file fails, including the filename and underlying error.
 *
 * Side Effects:
 * - Reads files from the filesystem.

*/
func CombineFilesForContext(focus []string, ignore []string) (string, error) {

    fmt.Printf("Would ignore these files, but that's not implemented yet: %v\n", ignore) 

    data := ""

    for _, file := range focus {
        info, err := os.Stat(file)
        if err != nil {
            return "", err
        }

        if info.IsDir() {
            more_files, err := os.ReadDir(file)
            if err != nil {
                return "", fmt.Errorf("failed to read directory %v: %v", file, err)
            }

            for _, f := range more_files {
                path := filepath.Join(file, f.Name())
                tmp, err := CombineFilesForContext([]string{path}, ignore)
                if err != nil {
                    return "", fmt.Errorf("Failed to append %v data: %v", path, err)
                }

                data += tmp
            }
            
        } else {
            tmp, err := AppendFileToData(file, data)
            if err != nil {
                return "", fmt.Errorf("Failed to append %v data: %v", file, err)
            }

            data += tmp
        }

    }

    return data, nil
}


func AppendFileToData(filename, data string) (string, error) {
    content, err := os.ReadFile(filename)
    if err != nil {
            return "", fmt.Errorf("failed to read file %v: %v", filename, err)
    }

    data += fmt.Sprintf("File:\n%v\nContents:\n%v\n\n", filename, string(content))

    return data, nil
}


/*
*
 * Summary: Recursively searches the project directory for files matching the configured language file extension and concatenates their contents for context.
 *
 * Signature: func FormatCodeFilesForContext() (types.ConcatenatedFileContents, error)
 *
 * Returns:
 * - types.ConcatenatedFileContents: concatenated contents of the filtered code files.
 * - error: error encountered during filtering or file reading.
 *
 * Errors/Exceptions:
 * - Returns an error if filtering for code files fails.
 * - Returns an error if reading any file fails.
 *
 * Side Effects:
 * - Reads files from the filesystem.
 *
 * Edge Cases & Assumptions:
 * - Skips ".git" directories.
 * - Only includes files that pass the language extension check.

*/
/*
func FormatCodeFilesForContext() (types.ConcatenatedFileContents, error) {
	files, err := FilterForCodeFiles(config.ProjectDirectory)
	if err != nil {
		return types.ConcatenatedFileContents(""), fmt.Errorf("Failed to filter for code files: %v", config.ProjectDirectory, err)
	}

	data := ""

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return types.ConcatenatedFileContents(""), fmt.Errorf("failed to read file %v: %v", file, err)
		}

		data += fmt.Sprintf("File:\n%v\nContents:\n%v\n\n", file, string(content))
	}

	return types.ConcatenatedFileContents(data), nil
}

*/
/*
*
Summary:
Formats and concatenates the contents of build-related files within the project directory for easy inspection or processing.

Signature:
func FormatBuildFilesForContext() (types.ConcatenatedFileContents, error)

Returns:
- types.ConcatenatedFileContents: concatenated string containing filenames and their contents
- error: error encountered during file filtering or reading

Errors/Exceptions:
Returns an error if filtering files fails or if reading any individual file fails.

*/
/*
func FormatBuildFilesForContext() (types.ConcatenatedFileContents, error) {
	files, err := FilterForBuildFiles(config.ProjectDirectory)
	if err != nil {
		return types.ConcatenatedFileContents(""), fmt.Errorf("Failed to filter for code files: %v", config.ProjectDirectory, err)
	}

	data := ""

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return types.ConcatenatedFileContents(""), fmt.Errorf("failed to read file %v: %v", file, err)
		}

		data += fmt.Sprintf("File:\n%v\nContents:\n%v\n\n", file, string(content))
	}

	return types.ConcatenatedFileContents(data), nil
}
*/
