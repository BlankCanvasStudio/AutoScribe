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

func (t Testing) Run(p string) error {
    log.Infof("input: %s", p)
    return nil
}

type Testing2 string;

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

