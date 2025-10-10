package manipulate

// import (
//     "fmt"
//
//     log "github.com/sirupsen/logrus"
//
//     "github.com/BlankCanvasStudio/AutoScribe/pkg/types/mst"
//     // "github.com/BlankCanvasStudio/AutoScribe/pkg/config"
//
// )
//
// func DocumentMST (m mst.MST) (mst.MST, error) {
//     for _, pkg := range m.GetPackages() {
//         for _, fd := range pkg.GetFunctionDecls() {
//             err := Document(fd)
//             if err != nil {
//                 return m, fmt.Errorf("failed to document func %v: %v", fd.GetName(), err)
//             }
//         }
//     }
//
//     return m, nil
// }
//
// func Document(f mst.FunctionDetails) error {
// 	// Consider how gpt aware gets loaded
//         /*
// 	if mst.IsAiAware(f) || mst.IsDocumented(f) || mst.WasDocumented(f) {
// 		return nil
//
// 	}
//         */
//
// 	if !mst.DocumentedThisPass(f) {
// 		return nil
//
// 	}
//
//
//         decl, err := mst.GetFuncDecl(f)
//         if err != nil {
//             return fmt.Errorf("failed to get function declaration from %v: %v", (f).GetFullName(), err)
//         }
//
//         // Undecided about this
// 	if decl == nil {
// 		log.Debugf("No declaration for `%v`. Assuming its defined in another package...", (f).GetName())
// 		return nil
// 	}
//
//         // This is almost certainly an issue
//         for _, call := range decl.GetCalls() {
//             tInfo := call.GetInfo()
//             if !mst.HasDocumentation(tInfo) && !mst.IsAiAware(tInfo) {
//                 // Recursively document if we need to
//                 err := Document(tInfo)
//                 // err := DocumentFunctions(*tInfo)
//                 if err != nil {
//                     return fmt.Errorf("failed to document call %v in %v: %v", tInfo.GetName(), (f).GetName(), err)
//                 }
//             }
//         }
//
//         /*
//         // By this point all nodes are either GPT aware or documented
//         NodeAsAiText, err := f.ToStringForAi()
//         if err != nil {
//                 return fmt.Errorf("failed to convert FunctionNode to AI string: %v", err)
//         }
//
//         DocumentationString, err := calls.QueryFromFile(mst.GPT_41_Nano, config.DocsPrompt, NodeAsAiText)
//         if err != nil {
//             return fmt.Errorf("failed to query from file: %v", err)
//         }
//
//         log.Debugf("result from gpt: `%v`", DocumentationString)
//
//         DocumentedComment, err := formatting.FormatAsGoComment(DocumentationString)
//         if err != nil {
//                 return fmt.Errorf("failed to parse for comments: %v", err)
//         }
//
//         f.SetDocumentation(DocumentedComment)
//         */
//
//         f.SetDocumentedInThisPass(true)
//
// 	return nil
// }
