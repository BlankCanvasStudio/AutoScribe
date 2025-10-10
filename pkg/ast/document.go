package ast

import (
	"fmt"
	"os"

	// "go/ast"

	log "github.com/sirupsen/logrus"
	// "github.com/BlankCanvasStudio/AutoScribe/pkg/types"
	// "github.com/BlankCanvasStudio/AutoScribe/pkg/config"
	// "github.com/BlankCanvasStudio/AutoScribe/pkg/ai/calls"
	// "github.com/BlankCanvasStudio/AutoScribe/pkg/ai/formatting"
)

/*
Summary:
DocumentFunctions recursively documents a FunctionInfo and its dependencies. It skips already-documented or AI-aware items, traverses declarations' Calls to document their Info as needed, and marks the input as documented upon completion. Use this to ensure complete, consistent documentation coverage for a function node and its call graph.

Signature:
func DocumentFunctions(f *FunctionInfo) error

Parameters:
- f: *FunctionInfo — the function node to document.

Returns:
- error — non-nil if a nested documentation step fails; otherwise nil.

Errors/Exceptions:
- Returns a formatted error when a recursive call fails: "failed to document call %v in %v: %v".
- May return nil early in cases where documentation is not needed (AiAware, already documented, or WasDocumented, or Declaration is nil).

Side Effects:
- May mutate f.Documented = true.
- May recursively mutate other FunctionInfo instances via DocumentFunctions(f.Declaration.Calls[i].Info).

Edge Cases & Assumptions:
- If f.AiAware || f.Documented || f.WasDocumented, the function returns nil without changes.
- If f.Declaration is nil, logs a debug message and returns nil (assumes the function is defined in another package).
- The function assumes valid, non-nil structures for Declaration and Calls when proceeding with recursion; if a nested call fails, the error propagates up.

*/
func DocumentFunctions(f *FunctionInfo) error {
	// Consider how gpt aware gets loaded
	if f.AiAware || f.Documented || f.WasDocumented {
		return nil
	}

	if f.Declaration == nil {
		log.Debugf("No declaration for `%v`. Assuming its defined in another package...", f.Name)
		return nil
	}

	for i := range len(f.Declaration.Calls) {
		if !f.Declaration.Calls[i].Info.Documented && !f.Declaration.Calls[i].Info.AiAware {
			// Recursively document if we need to
			err := DocumentFunctions(f.Declaration.Calls[i].Info)
			if err != nil {
				return fmt.Errorf("failed to document call %v in %v: %v", f.Declaration.Calls[i].Info.Name, f.Name, err)
			}
		}
	}

	/*
			// By this point all nodes are either GPT aware or documented
			NodeAsAiText, err := f.ToStringForGPT()
			if err != nil {
				return fmt.Errorf("failed to convert FunctionNode to GPT string: %v", err)
			}
		        NodeAsAiText = ""

		        DocumentationString, err := calls.QueryFromFile(types.GPT_41_Nano, config.DocsPrompt, f.Language, NodeAsAiText)
		        if err != nil {
		            return fmt.Errorf("failed to query from file: %v", err)
		        }

		        log.Debugf("result from gpt: `%v`", DocumentationString)

			DocumentedComment, err := formatting.FormatAsGoComment(DocumentationString)
			if err != nil {
				return fmt.Errorf("failed to parse for comments: %v", err)
			}

			f.Documentation = DocumentedComment
	*/

	f.Documented = true

	return nil
}

/*
Summary: Inserts the string insertion into the file at path at the given byte offset,
         replacing the content after the offset with the insertion, and writes back
         to the same path. Use when you need to inject text at a specific byte offset.
Signature: func insertIntoFile(path string, offset int, insertion string) error
Parameters:
  path: string - path to the target file.
  offset: int - byte offset where insertion occurs; must satisfy 0 <= offset <= len(data).
  insertion: string - text to insert at the offset.
Returns:
  error - non-nil if reading or writing the file fails, or if offset is out of range.
            Specifically, offset out of range occurs when offset < 0 || offset > len(data).
Errors/Exceptions:
  - offset out of range when offset is outside [0, len(data)].
  - errors from os.ReadFile(path) are propagated as-is.
  - errors from os.WriteFile(path, out, 0644) are propagated as-is.
Side Effects:
  - Reads the entire file content from path.
  - Writes the modified content back to path with permissions 0644.
  - May allocate memory proportional to the file size.
Edge Cases & Assumptions:
  - offset can be 0 (insert at start) or len(data) (append at end).
  - insertion may be empty (no change to content).
  - path must be accessible for reading and writing; errors are returned if not.

*/
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
