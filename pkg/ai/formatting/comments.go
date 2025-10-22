package formatting

import (
	"go/parser"
	"go/token"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

/*
Summary:
FormatAsGoComment formats an input string into a Go block comment by parsing the input as Go code after normalizing markdown wrappers and extracting the comments from the AST. If parsing fails, it prepends a synthetic "package main;" declaration and retries. The function returns the collected comments wrapped as a single Go block comment, or an error if parsing ultimately fails.

Signature:
func FormatAsGoComment(input string) (string, error)

Parameters:
- input: string. The text to convert. May contain markdown wrappers, code fences, and inline Go comments. The function normalizes the input to valid Go syntax for parsing.

Returns:
- string: a Go block comment containing the extracted comments.
- error: non-nil if both parsing attempts fail.

Errors/Exceptions:
- If the initial parse fails, the function tries again with a synthetic "package main;" prefix; if that also fails, it returns the encountered error.
- If no comments are found, the returned block represents an empty comment.

Side Effects:
- Logs errors and debug information via log.Errorf and log.Debugf when parsing fails.

Edge Cases & Assumptions:
- Assumes input can become valid Go syntax after normalization; otherwise the function returns an error.
- Output is a single Go-style block comment that can be embedded directly in Go source.
- If the input contains no Go comments, the result is an empty comment block.

*/
func FormatAsGoComment(input string) (string, error) {
	// Remove markdown wrapper
	re := regexp.MustCompile("`[`]*[a-zA-Z0-9]*")
	input = re.ReplaceAllString(input, "")

	// Fix opening comment - remove spaces between / and *
	re2 := regexp.MustCompile(`(?m)^[ ]*/[ ]*\*+`)
	input = re2.ReplaceAllString(input, "/*")

	// Fix closing comment - remove spaces between * and /
	re3 := regexp.MustCompile(`(?m)^[ ]*\*+[ ]*/`)
	input = re3.ReplaceAllString(input, "*/")

	input = strings.TrimSpace(input)

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "stdin.go", input, parser.ParseComments)
	if err != nil {
		// Try adding a main package statement so it doesn't fail
		input = "package main;\n" + input
		file, err = parser.ParseFile(fset, "stdin.go", input, parser.ParseComments)
		if err != nil {
			log.Errorf("failed with: %v", input)
			log.Debugf("error information: %+v", err)
			return "", err
		}
	}

	response := ""

	for _, cg := range file.Comments {
		response += cg.Text()
	}

	return "/*\n" + response + "\n*/", nil
}
