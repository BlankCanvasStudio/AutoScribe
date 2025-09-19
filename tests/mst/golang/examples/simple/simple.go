package simple;

import (
    "fmt"
)

func aRegularFunctionCall(a int, b string) error {
    a = a + 1
    b += "testing"
    return nil
}

type ANewStructure struct {
    Value string
}

func (a ANewStructure) SomeCall() {
    a.Value = "test"
}

func NestingFunctionCalls() error {
    aRegularFunctionCall(2, "else")
    fmt.Println("some print statement")

    a := ANewStructure{}
    a.SomeCall()

    return nil
}

