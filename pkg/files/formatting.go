package files

import (
    "os"
    "fmt"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/types"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/config"
)

func CombineFilesForContext(files []string) (types.ConcatenatedFileContents, error) {
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

