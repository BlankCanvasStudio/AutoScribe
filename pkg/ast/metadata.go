package ast

import (
	"fmt"
	"os"

	"golang.org/x/mod/modfile"
)

/*
*
 * Gets the module name from a go.mod file located in the specified folder.
 *
 * @param folder: string - the directory containing the go.mod file.
 * @return: string - the module path extracted from go.mod.
 * @return: error - if reading or parsing go.mod fails.
 *
 * Errors are returned if the go.mod file cannot be read or parsed successfully.

*/
func GetModuleName(folder string) (string, error) {
	goMod := fmt.Sprintf("%v/go.mod", folder)

	data, err := os.ReadFile(goMod)
	if err != nil {
		return "", fmt.Errorf("failed to read %v: %v", goMod, err)
	}

	file, err := modfile.Parse(goMod, data, nil)
	if err != nil {
		return "", fmt.Errorf("failed to parse %v: %v", goMod, err)
	}

	return file.Module.Mod.Path, nil
}
