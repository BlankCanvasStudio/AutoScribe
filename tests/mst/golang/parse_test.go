package main;

import (
    // "os"
    "fmt"
    "slices"
    "strings"
    "testing"
    log "github.com/sirupsen/logrus"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/types/mst/golang"
)

func TestPopulateMst(t *testing.T) {
    gMst := golang.MST{}
    gMst.Populate([]string {"./examples/simple"})

    numPackages := 1
    numFunctions := 4
    numImports := 1
    numNestedCalls := 3


    if len(gMst.PackageNodes) != numPackages {
        log.Fatalf("simple package parsed into %v packages, not %v", len(gMst.PackageNodes), numPackages)
    }

    if len(gMst.FunctionMap) != numFunctions {
        log.Fatalf("simple package didn't parse %v functions, parsed %v", numFunctions, len(gMst.FunctionMap))
    }

    if len(gMst.PackageNodes[0].GetImports()) != numImports {
        log.Fatalf("simple package didn't parse %v import, parsed %v", numImports, len(gMst.PackageNodes[0].GetImports()))
    }

    nestedName := "github.com/BlankCanvasStudio/AutoScribe/tests/mst/golang/examples/simple.NestingFunctionCalls"

    nestedInfo, ok := gMst.FunctionMap[nestedName]
    if !ok {
        log.Fatalf("didn't find %v in FunctionMap", nestedName)
    }

    nestedCalls := nestedInfo.GetDecl().GetCalls()

    if len(nestedCalls) != numNestedCalls {
        log.Fatalf("incorrect number of calls in %v: %v", nestedName, len(nestedInfo.GetDecl().GetCalls()))
    }

    nestedCallNames := make([]string, 0)

    for _, val := range nestedCalls {
        nestedCallNames = append(nestedCallNames, val.GetFullName())
    }

    if !slices.Contains(nestedCallNames, "github.com/BlankCanvasStudio/AutoScribe/tests/mst/golang/examples/simple.aRegularFunctionCall") {
        log.Fatalf("didn't parse internal nested calls correctly")
    }

    if !slices.Contains(nestedCallNames, "fmt.Println") {
        log.Fatalf("didn't parse package calls correctly")
    }

    if !strings.Contains(gMst.FunctionMap["github.com/BlankCanvasStudio/AutoScribe/tests/mst/golang/examples/simple.aRegularFunctionCall"].GetFile(), "simple.go") {
        log.Fatalf("file name not populated correctly in function definition")
    }
}


func TestPopulatingDocumentation(t *testing.T) {
    gMst := golang.MST{}
    gMst.Populate([]string {"./examples/simple"})

    for name, info := range gMst.FunctionMap {
        info.SetDocumentation(name)
    }

    for _, pkg := range gMst.GetPackages() {
        for _, decl := range pkg.GetFunctionDecls() {
            NodeAsText, err := decl.ToStringForAi()
            if err != nil {
                log.Fatalf("failed to turn decl into text for ai: %v", err)
            }

            if decl.GetName() != "NestingFunctionCalls" { continue }

            ExpectedText := "func NestingFunctionCalls() error {\n /* github.com/BlankCanvasStudio/AutoScribe/tests/mst/golang/examples/simple.aRegularFunctionCall */\n     aRegularFunctionCall(2, \"else\")\n /* fmt.Println */\n     fmt.Println(\"some print statement\")\n\n    a := ANewStructure{}\n /* github.com/BlankCanvasStudio/AutoScribe/tests/mst/golang/examples/simple.ANewStructure.SomeCall */\n     a.SomeCall()\n\n    return nil\n}"

            if NodeAsText != ExpectedText {
                log.Fatalf("Didn't format function text with documentation as expected")
            }
        }
    }
}

func TestPopulateAcrossPackages(t *testing.T) {
    gMst := golang.MST{}
    // Order matters here
    gMst.Populate([]string {"./examples/importing", "./examples/simple"})

    function_we_import := "github.com/BlankCanvasStudio/AutoScribe/tests/mst/golang/examples/simple.NestingFunctionCalls"

    val, ok := gMst.FunctionMap[function_we_import]

    if !ok {
        log.Fatalf("function %v isn't in function map, despite being declared and imported", function_we_import)
    }

    if val.GetPackage().GetPath() != "github.com/BlankCanvasStudio/AutoScribe/tests/mst/golang/examples/simple" {
        fmt.Printf("\n\n%v\n\n", val.GetPackage().GetPath())
        log.Fatalf(fmt.Sprintf("package for function info not set correctly"))
    }


}



