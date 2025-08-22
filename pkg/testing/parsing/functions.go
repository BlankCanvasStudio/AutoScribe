package parsing;

/*
*
*   This package exists to run the parsing on and test the output. Please don't use it
*
*/



import (
    "fmt"

    "go/token"
    "go/parser"

    log "github.com/sirupsen/logrus"

    "github.com/openai/openai-go/v2"
    "github.com/openai/openai-go/v2/option"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/ast"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/config"
)

type Testing struct {

}

/**
 * Run logs the input string using the info log level and returns nil.
 * Use this method to record the provided string for testing purposes.
 * 
 * func (t Testing) Run(p string) error
 * 
 * @param p string - the input string to be logged
 * @return error - always returns nil
 * @errors - never returns an error
 * @side effects - logs the input string
 * @edge cases - none
 */
func (t Testing) Run(p string) error {
    log.Infof("input: %s", p)
    return nil
}

type Testing2 string;

/**
* Parses a Go source file, logs function definitions, initializes an OpenAI client, and calls internal functions.
* Use when processing and analyzing Go files with associated external API interaction.
*
* @param filename string - Path to the file to parse.
* @return error - Error if parsing or AST extraction fails; otherwise nil.
* @sideEffects - Logs function definitions and client info; calls internal functions.
* @edgeCases - Handles parse errors, function extraction errors, and nil values gracefully.
*/
func ParseFile(filename string) error {
    fset := token.NewFileSet()

    // Will need to handle this case
    t := Testing{}
    t.Run()

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

