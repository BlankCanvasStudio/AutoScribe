package formatting

import (
	// "fmt"

	// "go/ast"
	"go/parser"
	"go/token"
	"regexp"
)

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
