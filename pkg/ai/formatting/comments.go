package formatting

import (
	"regexp"
	"go/token"
	"go/parser"

	log "github.com/sirupsen/logrus"
)


func FormatAsGoComment(input string) (string, error) {
	re := regexp.MustCompile("`[`]*[a-zA-Z0-9]*")
	input = re.ReplaceAllString(input, "")

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "stdin.go", input, parser.ParseComments)
	if err != nil {
		// Try adding a main package statement so it doesn't fail
		input = "package main;\n" + input
		file, err = parser.ParseFile(fset, "stdin.go", input, parser.ParseComments)
		if err != nil {
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

