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
FormatAsGoComment takes an input string, normalizes and extracts Go-style comments, and returns them as a single Go-style block comment. It removes markdown wrappers, fixes comment delimiters, parses the input as Go code, and, if needed, prepends a package statement to improve parsing. The function then collects all top-level comments from the parsed file and returns them wrapped as a single comment block.

Signature:
func FormatAsGoComment(input string) (string, error)

Parameters:
- input: string
  role: the source text to convert into a Go comment block.
  constraints: arbitrary string; markdown wrappers and inline code markers will be stripped.

Returns:
- string: the concatenated comments extracted from the input, formatted as a Go-style block comment.
- error: non-nil if parsing or extraction fails.

Errors/Exceptions:
- Returns an error if parsing fails even after attempting to salvage by adding "package main;\n".
- May log details via log.Errorf and log.Debugf when errors occur.

Side Effects:
- Reads and mutates a local copy of input during processing.
- May emit log messages.

Edge Cases & Assumptions:
- If input already contains valid Go comments, those are returned.
- If initial parsing fails, the function attempts to salvage by prepending a minimal package declaration.
- The output is all the comments found in file.Comments, concatenated in visitation order.

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
