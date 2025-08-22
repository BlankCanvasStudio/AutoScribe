package ast;

import (
    "os"
    "fmt"

    "golang.org/x/mod/modfile"
)

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

