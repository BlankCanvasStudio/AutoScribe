package formatting

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

/*
Summary: Build a single string by recursively traversing the paths in focus, reading files, and appending their contents using AppendFileToData. The result represents a cohesive context blob for the given inputs. The ignore parameter is accepted but currently not implemented (a debug log notes it).
Signature: func CombineFilesForContext(focus []string, ignore []string) (string, error)
Parameters:
  focus []string - file or directory paths to include; directories are traversed recursively.
  ignore []string - paths to ignore; currently not implemented.
Returns:
  string - concatenated data containing per-file blocks (via AppendFileToData).
  error - non-nil if a file cannot be stat'ed, a directory cannot be read, or AppendFileToData fails for a file.
Errors/Exceptions:
  Returns an error when os.Stat(file) fails, os.ReadDir(file) fails, or AppendFileToData(file, data) returns an error.
Side Effects:
  Reads filesystem contents; does not modify input files or global state.
Edge Cases & Assumptions:
  Directories are processed in the order returned by os.ReadDir; files are appended in that traversal order.
  File contents are included as strings (binary content is represented as string(bytes)); data is extended for each processed file.
  If data is initially empty, the File/Contents block is still appended for each file.

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
Summary: Reads the file named by filename and appends a formatted block containing the file name and its contents to data, returning the updated string.
Signature: func AppendFileToData(filename, data string) (string, error)
Parameters:
  filename string - path to the file to read.
  data string - existing data to which the file block will be appended.
Returns:
  string - updated data with the appended File and Contents block.
  error - non-nil on read failure.
Errors/Exceptions:
  If reading the file fails, returns "" and an error formatted as "failed to read file %v: %v" including filename and the underlying error.
Side Effects:
  Reads from the filesystem; does not modify the input file or the filesystem; returns a new string.
Edge Cases & Assumptions:
  File contents are included as string(content); binary content will be represented as a string via string(content).
  If data is empty, the File/Contents block is still appended.
  The function assumes os.ReadFile and fmt.Sprintf behave as documented; on error it returns a non-nil error.

*/
func AppendFileToData(filename, data string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read file %v: %v", filename, err)
	}

	data += fmt.Sprintf("File:\n%v\nContents:\n%v\n\n", filename, string(content))

	return data, nil
}
