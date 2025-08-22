package ast;

import (
    "os"
    "fmt"

    "go/ast"

    // log "github.com/sirupsen/logrus"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/openai/calls"
)

var AiDocumentPromptV1 string = `
You are a precise code documenter. Read the function below and produce clear, concise documentation.

**IMPORTANT: All of these files are written in the Language: %v. Repond accordingly**

Constraints:
- Keep less than 250 words unless the function is unusually complex.
- Do not speculate about code not shown; note external dependencies without guessing behavior.
- Use exact identifier names and types from the code.
- Prefer active voice and plain English.
- **VERY IMPORTANT:** If you can clearly understand the code without your documentation, don't return anything
- Ouptput the result in a Language Docstring/Comment of the language used: provide an idiomatic doc block (e.g., GoDoc for golang, JSDoc for js, Python docstring for python) for the function.
- **NEVER RETURN THE SOURCE CODE. ONLY COMMENTS**
- **DO NOT RETURN ANYTHING OUTSIDE THE COMMENT BLOCK
- **Respond only with comments, no markdown formatting or code fences.**

Output format (omit any that don’t apply or can be trivially seen):
1) Summary: what the function does and when to use it.
2) Signature: An abveriation of the function signature. Do not return code
3) Parameters: list with name, type, meaning, important constraints.
4) Returns: what is returned and under which conditions.
5) Errors/Exceptions: when/why failures occur (panics, error values, thrown exceptions).
6) Side Effects: state mutation, I/O, network calls, concurrency (locks/goroutines/threads).
7) Edge Cases & Assumptions: inputs that change behavior; pre/postconditions.

--- BEGIN CODE ---
%v
--- END CODE ---
`

var AiDocumentPrompt string = `
You are a precise code documenter. Read the function below and produce clear, concise documentation in the file’s language (Language: %v).

Rules
- CRITICAL: Respond only when necessary. If the function is obvious from its name/signature, return nothing.
- Use exact identifier names and types from the code.
- Prefer active voice and plain English.
- Keep responses short; expand only if truly needed to clarify function behavior.
- Do not speculate about unseen code.
- Output only a language-idiomatic doc block (e.g. GoDoc, JSDoc, Python docstring).
- Do not wrap in markdown/code fences.
- Assume the code is valid.
- Omit any section that is empty, trivial, or you would put "none" in.
- If a comment in the code helps to explain the code, quote it directly, do not paraphrase it

Output format (omit irrelevant/trivial):
- Summary: what the function does and when to use it.
- Signature: accurate language syntax.
- Parameters: name, type, role, constraints.
- Returns: values and conditions.
- Errors/Exceptions: failure cases.
- Side Effects: mutations, I/O, concurrency.
- Edge Cases & Assumptions: unusual inputs, pre/postconditions.

--- BEGIN CODE ---
%v
--- END CODE ---`


// Should add parsing to this to drop anything that isn't a comment
func DocumentFunctions(f *FunctionNode) error {
    // Consider how gpt aware gets loaded
    if f.AiAware || f.Documented {
        return nil
    }

    for i := range(len(f.Calls)) {
        if !f.Calls[i].Documented && !f.Calls[i].AiAware {
            // log.Infof("%v: %v", f.Calls[i].Name, f.Calls[i].Documented)
            // Recursively document if we need to
            err := DocumentFunctions(f.Calls[i])
            if err != nil {
                return fmt.Errorf("failed to document call %v in %v: %v", f.Calls[i].Name, f.Name, err)
            }
        }
    }

    // We only want to document function declarations
    if f.Kind != FnDeclaration {
        return nil
    }

    if fd, ok := f.Node.(*ast.FuncDecl); !ok || fd.Doc != nil {
        return nil
    }


    // By this point all nodes are either GPT aware or documented
    NodeAsAiText, err := f.ToStringForGPT()    
    if err != nil {
        return fmt.Errorf("failed to convert FunctionNode to GPT string: %v", err)
    }

    FullDocumentationQuery := fmt.Sprintf(AiDocumentPrompt, f.Language, NodeAsAiText)

    DocumentationString, err := calls.Query4_1Nano(FullDocumentationQuery)
    if err != nil {
        return fmt.Errorf("failed to query 4.1 Nano: %v", err)
    }

    f.Documentation = DocumentationString

    // Actually moving this outside the loop. That way we can tell if we need to update docs or 
    //  not based on init presence of goDoc
    /*
    fd, ok := f.Node.(*ast.FuncDecl)
    if ok {
        // Add the comment in properly
        fd.Doc = &ast.CommentGroup{
            List: []*ast.Comment{
                {Text: DocumentationString},
            },
        }
    }
    */

    f.Documented = true;

    // log.Infof("GPT Query: \n%v\n", FullDocumentationQuery)
    // FullDocumentationQuery += "1"

    return nil
}


func insertIntoFile(path string, offset int, insertion string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return err
    }
    if offset < 0 || offset > len(data) {
        return fmt.Errorf("offset out of range")
    }
    out := append(append([]byte{}, data[:offset]...), append([]byte(insertion), data[offset:]...)...)
    return os.WriteFile(path, out, 0644)
}

