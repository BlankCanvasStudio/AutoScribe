package formatting

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

/*
Summary: Recursively collect and concatenate the contents of the files specified in focus into a single string, traversing directories and using AppendFileToData to format each file's data. Use when you need a unified contextual blob representing a set of files.

Signature: func CombineFilesForContext(focus []string, ignore []string) (string, error)

Parameters:
- focus: []string, paths to include. Each entry may be a file or a directory. Directories are traversed recursively.
- ignore: []string, file paths to ignore. Currently not implemented; provided for future use. The function logs that ignoring is not yet implemented.

Returns:
- string: The accumulated data produced by combining each file's formatted block, as produced by AppendFileToData for each processed file.
- error: Non-nil if any stat/read/recursion/appending operation fails.

Errors/Exceptions:
- non-nil if os.Stat(file) returns an error for any path in focus.
- non-nil if os.ReadDir(file) returns an error when processing directories.
- non-nil if CombineFilesForContext recursively called on a path returns an error.
- non-nil if AppendFileToData(file, data) returns an error for any file.

Side Effects:
- Reads files and directories from disk; does not write to disk.
- Emits a debug log about ignore handling: "Would ignore these files, but that's not implemented yet: %v"

Edge Cases & Assumptions:
- ignore currently has no effect; future behavior may filter paths before processing.
- The order of produced data follows the order of focus, and, for directories, the order returned by os.ReadDir.
- If a path cannot be read or stat fails, the function returns an error and does not partially modify the result.

*/
func CombineFilesForContext(focus []string, ignore []string) (string, error) {

	log.Debugf("Would ignore these files, but that's not implemented yet: %v\n", ignore)

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

/*
Summary: Appends a labeled representation of the file specified by filename to the given data string.
Use when you want to accumulate the contents of a file alongside existing data.
Signature: func AppendFileToData(filename, data string) (string, error)
Parameters:
- filename: string, path to the file to read.
- data: string, existing data to append to.
Returns:
- string: the updated data containing the appended file block in the form:
  "File:\n<filename>\nContents:\n<contents>\n\n"
- error: non-nil if reading the file fails.
Errors/Exceptions:
- returns an error: "failed to read file %v: %v" on read failure, with the filename and underlying error.
Side Effects:
- Reads the file from disk. No write operations are performed.
Edge Cases & Assumptions:
- If the file cannot be read, the function returns an error and does not modify data.
- The file contents are appended as a string; binary data will be included as its string representation.

*/
func AppendFileToData(filename, data string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read file %v: %v", filename, err)
	}

	data += fmt.Sprintf("File:\n%v\nContents:\n%v\n\n", filename, string(content))

	return data, nil
}
