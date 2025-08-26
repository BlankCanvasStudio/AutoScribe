package ast;

import (
    "os"
    "fmt"

    "golang.org/x/mod/modfile"
)

/**
 * Retrieves the module name from the go.mod file in the specified folder.
 *
 * @param folder string: the path to the folder containing the go.mod file.
 * @return string: the module's path.
 * @return error: an error if reading or parsing go.mod fails.
 * 
 * Errors occur if the go.mod file cannot be read or the content is invalid.
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

