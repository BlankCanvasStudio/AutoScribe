package ast

import (
	"fmt"
	"os"

	"golang.org/x/mod/modfile"
)

/*
*
 * Retrieves the module name from a Go project's go.mod file located in the specified folder.
 *
 * Signature:
 * func GetModuleName(folder string) (string, error)
 *
 * Parameters:
 *   folder (string): Path to the directory containing the go.mod file.
 *
 * Returns:
 *   string: The module's import path as specified in go.mod.
 *   error: An error if reading or parsing the go.mod file fails.
 *
 * Errors/Exceptions:
 *   Returns an error if reading the go.mod file or parsing its contents fails.
 *
 * Side Effects:
 *   Reads the contents of the go.mod file.

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
