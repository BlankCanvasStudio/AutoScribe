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
Summary: DocumentFunctions prepares a FunctionInfo for documentation by ensuring all referenced
functions within f.Declaration are either documented or AI-aware, then marks f as documented.
If the function is already AI-aware, already documented, or has no Declaration, it returns nil
without changes.

Signature: func DocumentFunctions(f *FunctionInfo) error

Parameters:
  - f: *FunctionInfo — the function to document; may be updated in place. If f.Declaration is nil,
    a debug message is logged and the function returns nil, assuming it is defined in another package.

Returns:
  - error: non-nil if a recursive documentation call fails; otherwise nil.

Errors/Exceptions:
  - "failed to document call %v in %v: %v" if a recursive DocumentFunctions call returns an error.

Side Effects:
  - Sets f.Documented = true on success.
  - May recursively call DocumentFunctions on f.Declaration.Calls[i].Info for each undocumented, non-AI-aware call.

Edge Cases & Assumptions:
  - If f.AiAware || f.Documented || f.WasDocumented, the function returns nil immediately.
  - If f.Declaration == nil, logs a debug message and returns nil (assumes definition elsewhere).
  - The GPT-based documentation logic is present but currently commented out.

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
Summary: Inserts insertion into the file at the given byte offset.
It reads the file at path, validates offset, and writes the modified content back
to path with insertion inserted at offset.
Use when you need to programmatically inject text into a file at a specific byte position.

Signature: func insertIntoFile(path string, offset int, insertion string) error

Parameters:
- path string: path to the target file.
- offset int: zero-based byte offset where insertion occurs; must satisfy 0 <= offset <= len(data).
- insertion string: content to insert at offset.

Returns:
- error: non-nil if reading, offset validation, or writing fails.

Errors/Exceptions:
- if os.ReadFile(path) fails, the error is returned.
- if offset < 0 || offset > len(data), returns fmt.Errorf("offset out of range").
- if os.WriteFile fails, the error is returned.

Side Effects:
- Reads the file at path; mutates its content by inserting insertion at offset.
- Writes the updated content back to path with permissions 0644.

Edge Cases & Assumptions:
- offset == 0 inserts at the start; offset == len(data) appends.
- insertion may be empty; file content remains unchanged except for the insertion.
- path must be a writable regular file; no directory creation is performed.
- Reads entire file into memory; not suitable for very large files.

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
