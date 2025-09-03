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


//
//
// Summary: Generates inline documentation comments for the specified FunctionInfo by leveraging GPT-based processing. It ensures that all called functions are documented or flagged as AI-aware before creating the documentation.
// Signature: func (f *FunctionInfo) DocumentFunctions() error
// Parameters:
//   - f: *FunctionInfo — the function information object containing declaration and call details.
// Returns:
//   - error: if the function declaration is nil, if the GPT string conversion fails, or if the documentation query or formatting encounters an error.
// Errors/Exceptions:
//   - Returns an error if the declaration is nil or if any step in generating or formatting the documentation fails.
// Side Effects:
//   - Reads the source file content, potentially alters the FunctionInfo's Documentation field, and updates the documented status.
// Edge Cases & Assumptions:
//   - Assumes the source code can be successfully read and modified.
//   - Assumes all offset calculations for code segments are valid.
//   - Assumes that the related call functions are accessible and properly linked.
//   - Handles recursive documentation for unmarked call functions.

//
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

//
//
// Inserts a string into a file at a specified byte offset.
//
// Signature: func insertIntoFile(path string, offset int, insertion string) error
//
// Parameters:
// - path: string; the file path to modify.
// - offset: int; the byte position in the file where the insertion should occur.
// - insertion: string; the content to insert into the file.
//
// Returns:
// - error: nil if successful; otherwise, an error indicating the failure reason.
//
// Errors/Exceptions:
// - Returns an error if reading or writing the file fails.
// - Returns an error if the offset is outside the valid range (less than 0 or greater than file size).
//
// Side Effects:
// - Modifies the file at the specified path by inserting the given string at the specified offset.
//
// Edge Cases & Assumptions:
// - Assumes the file at 'path' exists and is accessible.
// - Handles insertion at the start (offset 0) and end (offset == file size).

//
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

