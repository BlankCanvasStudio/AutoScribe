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
