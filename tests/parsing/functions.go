package parsing

/*
*
*   This package exists to run the parsing on and test the output. Please don't use it
*
 */

import (
	"fmt"

	"go/parser"
	"go/token"

	log "github.com/sirupsen/logrus"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"

	"github.com/BlankCanvasStudio/AutoScribe/pkg/ast"
	"github.com/BlankCanvasStudio/AutoScribe/pkg/config"
)

type Testing struct {
}

/*
*
 * Run executes the testing process with the specified input string.
 *
 * Parameters:
 *   p - string: the input to be processed during testing.
 *
 * Returns:
 *   error: nil if the process completes successfully; otherwise, an error.
 *
 * Side Effects:
 *   Logs the input string at info level.
Commnet #1
Comment #2
And its multi-line

*/
func (t Testing) Run(p string) error {
	// Commnet #1
	log.Infof("input: %s", p)
	// Comment #2
	// And its multi-line
	return nil
}

type Testing2 string

/*
*
 * Parses the specified Go source file, extracts function definitions, and logs their names.
 * Initializes an OpenAI client with the provided API key.
 * Executes a test run with a hardcoded input string.
 *
 * Signature:
 * func ParseFile(filename string) error
 *
 * Parameters:
 * - filename: string, the path to the Go source file to parse.
 *
 * Returns:
 * - error: nil if parsing and processing succeed; otherwise, an error describing the failure.
 *
 * Errors/Exceptions:
 * - Returns an error if file parsing or function extraction fails.
 *
 * Side Effects:
 * - Logs function names and client information.
 * - Performs a test run with hardcoded input.
 *
 * Edge Cases & Assumptions:
 * - Assumes the file exists and is a valid Go source file.
 * - Does not handle specific errors within the test run or client creation beyond reporting.

*/
func ParseFile(filename string) error {
	fset := token.NewFileSet()

	// Will need to handle this case
	t := Testing{}
	t.Run("some input")

	f, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
	if err != nil {
		return fmt.Errorf("failed to parse %v: %+v", filename, err)
	}

	funcs, err := ast.GetFunctionDefinitions(f)
	if err != nil {
		return fmt.Errorf("failed to get function definitions: %v", err)
	}

	for _, fun := range funcs {
		log.Infof("function definition for: %+v", fun.Name)
	}

	client := openai.NewClient(
		option.WithAPIKey(config.OpenAIKey),
	)

	log.Infof("client: %v", client)

	AnInternalFunction()

	tmp := Testing2("some value")
	log.Infof("has been typecast: %v", tmp)

	return nil
}
