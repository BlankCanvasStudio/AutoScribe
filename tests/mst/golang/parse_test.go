package main;

import (
    // "os"
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


