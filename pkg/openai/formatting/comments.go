package formatting

import (
	// "fmt"

	// "go/ast"
	"go/parser"
	"go/token"
	"regexp"
)



/*
*
 * Summary: Parses the input Go code string, extracts comments, removes code block delimiters, and formats the comments as a Go-style block comment.
 *
 * Signature: func FormatAsGoComment(input string) (string, error)
 *
 * Parameters:
 * - input: string - the Go code to extract comments from.
 *
 * Returns:
 * - string: the comments formatted as a Go comment block.
 * - error: if parsing fails after attempting to prepend a package statement.
 *
 * Errors/Exceptions:
 * - Returns an error if the code cannot be parsed even after modification.
 *
 * Side Effects:
 * - None.
 *
 * Edge Cases & Assumptions:
 * - Assumes input is valid Go code or can be parsed when a package statement is added.

*/
func FormatAsGoComment(input string) (string, error) {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "stdin.go", input, parser.ParseComments)
	if err != nil {
		// Try adding a main package statement so it doesn't fail
		input = "package main;\n" + input
		file, err = parser.ParseFile(fset, "stdin.go", input, parser.ParseComments)
		if err != nil {
			return "", err
		}
	}

	re := regexp.MustCompile("```[a-zA-Z0-9]*")

	response := ""

	for _, cg := range file.Comments {
		response += cg.Text()
	}

	response = re.ReplaceAllString(response, "")

	return "/*\n" + response + "\n*/", nil
}
