package parse;

import (
    "github.com/BlankCanvasStudio/AutoScribe/pkg/types"

)

func DocumentFunctions(f *FuncitonNode) error {
	// Consider how gpt aware gets loaded
	if types.IsAiAware(f) || types.IsDocumented(f) || types.WasDocumented(f) {
		return nil

	}

        decl, err := types.GetFuncDecl(f)
        if err != nil {
            return fmt.Errorf("failed to get function declaration from %v: %v", f.FullName(), err)
        }

        // Undecided about this
	if decl == nil {
		log.Debugf("No declaration for `%v`. Assuming its defined in another package...", f.Name)
		return nil
	}

        // info := types.GetInfo(f)

        for _, call := range decl.Calls {
            tInfo := types.GetInfo(call)
            if !types.WasDocumented(tInfo) && !types.IsAiAware(iInfo) {
                // Recursively document if we need to
                err := DocumentFunctions(tInfo)
                if err != nil {
                    return fmt.Errorf("failed to document call %v in %v: %v", tInfo.GetName(), f.GetName(), err)
                }
            }
        }

        /*
	for i := range len(f.Declaration.Calls) {
		if !f.Declaration.Calls[i].Info.Documented && !f.Declaration.Calls[i].Info.AiAware {
			// Recursively document if we need to
			err := DocumentFunctions(f.Declaration.Calls[i].Info)
			if err != nil {
				return fmt.Errorf("failed to document call %v in %v: %v", f.Declaration.Calls[i].Info.Name, f.Name, err)
			}
		}
	}
        */

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









