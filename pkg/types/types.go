package types

import (
	"bufio"
	"os"
	"path/filepath"
	"slices"
	"strings"
	// log "github.com/sirupsen/logrus"
)

type ConcatenatedFileContents string

type SupportedFormat string

const (
	Shell    SupportedFormat = "sh"
	Golang   SupportedFormat = "go"
	Bash     SupportedFormat = "sh"
	Text     SupportedFormat = "txt"
	MarkDown SupportedFormat = "md"
)

var SupportedFormats = []SupportedFormat{
	Shell,
	Golang,
	Bash,
	Text,
	MarkDown,
}

/*
*
 * Checks if the given format string is supported.
 *
 * @param val {string} - The format string to verify.
 * @returns {bool} - True if the format is supported; otherwise, false.

*/
func IsSupportedFormat(val string) bool {
	return slices.Contains(SupportedFormats, SupportedFormat(val))
}


/*
*
 * Summary: Checks whether the file at the given path matches the supported format, based on extension or shebang line. Use this for verifying if a file conforms to a specified SupportedFormat.
 * Signature: func (f SupportedFormat) FileIsThisFormat(path string) (string, bool)
 * Parameters:
 *   - path: string, the file system path to the file to be checked.
 * Returns:
 *   - string: the file extension (without leading '.') if matched, or an empty string if not.
 *   - bool: true if the file matches the format; false otherwise.
 * Errors/Exceptions: Returns false if there's an error reading the file head or if shebang verification fails.
 * Side Effects: Opens the file for reading when extension is absent.
 * Edge Cases & Assumptions: Assumes file exists and is readable when no extension is present; ignores errors during file opening.

*/
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

/*
*
 * Generates a list of shebang lines for the specified supported format.
 *
 * @param program SupportedFormat - the format of the program (e.g., 'python', 'bash') to include in the shebang lines.
 * @return []string - a slice containing two shebang lines: one using '/usr/bin/env' and one using '/bin/'.

*/
func GenerateShebangs(program SupportedFormat) []string {
	return []string{
		"#!/usr/bin/env " + string(program),
		"#!/bin/" + string(program),
	}
}
