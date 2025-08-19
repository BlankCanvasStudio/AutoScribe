package files;

import (
    "os"
    "fmt"
    "strings"
    "path/filepath"

    log "github.com/sirupsen/logrus"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/config"
)


func FilterForCodeFiles(directory string) ([]string, error) {
    log.Debugf("Filtering for %v code files in: %v", config.LanguageFileExtension, config.ProjectDirectory)

    // Collect all the files in question
    var files []string

    err := filepath.Walk(config.ProjectDirectory, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if ! info.IsDir() {
            if ext, is := config.LanguageFileExtension.FileIsThisFormat(path); ! is {
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

