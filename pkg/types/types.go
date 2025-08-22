package types

import (
    "os"
    "bufio"
    "slices"
    "strings"
    "path/filepath"

    // log "github.com/sirupsen/logrus"
)

type ConcatenatedFileContents string

type SupportedFormat string
const (
    Shell SupportedFormat  = "sh"
    Golang SupportedFormat = "go"
    Bash SupportedFormat   = "sh"
    Text SupportedFormat   = "txt"
)

var SupportedFormats = []SupportedFormat {
    Shell,
    Golang,
    Bash,
    Text,
};

func IsSupportedFormat(val string) (bool) {
    return slices.Contains(SupportedFormats, SupportedFormat(val))
}

func (f SupportedFormat) FileIsThisFormat(path string) (string, bool) {
    ext := filepath.Ext(path)

    // If there's no file extension, verify it with the shebang
    if len(ext) == 0 {
        fd, _ := os.Open(path)
        first_line, err := bufio.NewReader(fd).ReadString('\n')
        if err != nil {
            return "", false
        }

        if slices.Contains(GenerateShebangs(f), strings.TrimSpace(first_line)) {
            return string(f), true
        }

        return "", false
    }

    if ext[0] == '.' {
        ext = ext[1:]
    }

    if SupportedFormat(ext) != f {
        return ext, false
    } 

    return ext, true
}

func GenerateShebangs(program SupportedFormat) []string {
    return []string {
        "#!/usr/bin/env " + string(program), 
        "#!/bin/" + string(program), 
    }
}




