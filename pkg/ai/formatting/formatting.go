package formatting

import (
	"fmt"
	"os"
        "path/filepath"

	log "github.com/sirupsen/logrus"
)


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


func AppendFileToData(filename, data string) (string, error) {
    content, err := os.ReadFile(filename)
    if err != nil {
            return "", fmt.Errorf("failed to read file %v: %v", filename, err)
    }

    data += fmt.Sprintf("File:\n%v\nContents:\n%v\n\n", filename, string(content))

    return data, nil
}

