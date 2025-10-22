package docs

import (
	"fmt"
	"os"
	"sync"
	// "strings"
	"encoding/json"
	"math/rand"

	log "github.com/sirupsen/logrus"

	"github.com/BlankCanvasStudio/AutoScribe/pkg/ai/call"
	"github.com/BlankCanvasStudio/AutoScribe/pkg/types/mst"
	"github.com/BlankCanvasStudio/AutoScribe/pkg/types/mst/golang"
)

type RagContext struct {
	Title      string  `json:"title"`
	Text       string  `json:"text"`
	Score      float64 `json:"score"`
	TitleScore float64 `json:"title_score"`
	PassageId  int     `json:"passage_id"`
}

type RagEntry struct {
	Dataset  string   `json:"dataset"`
	Question string   `json:"question"`
	Answers  []string `json:"answers"`

	PositiveCtxs     []RagContext `json:"positive_ctxs"`
	NegativeCtxs     []RagContext `json:"negative_ctxs"`
	HardNegativeCtxs []RagContext `json:"hard_negative_ctxs"`
}

type QA struct {
	Question string `json:"question"`
	Answer   string `josn:"answer"`
}

/*
Summary:
CreateRagChunkCounts builds an MST from inputPackages, documents the MST using a fixed prompt loaded from /etc/autoscribe/prompts/docs.txt via an API key, and then reports per-function character lengths and the overall average. Use it to prepare and audit AI-generated documentation statistics for a set of Go packages.

Signature:
func CreateRagChunkCounts(inputPackages []string, ApiKey string) error

Parameters:
- inputPackages: []string; filesystem paths to Go package roots to include in the MST.
- ApiKey: string; API key used by DocumentMST to generate documentation.

Returns:
- error: non-nil if any step fails (MST population, prompt read, or MST documentation).

Errors/Exceptions:
- Non-nil error if MST population fails for the processed packages.
- Non-nil error if the prompt document cannot be read from /etc/autoscribe/prompts/docs.txt.
- Non-nil error if MST documentation via mst.DocumentMST fails.

Side Effects:
- Reads /etc/autoscribe/prompts/docs.txt.
- Logs per-function AI string lengths and the resulting average.
- Mutates the local gMst state during Populate and DocumentMST (no external state exposure).

Edge Cases & Assumptions:
- If no packages or no function declarations exist, the computed average may cause a division-by-zero panic.
- The function relies on the fixed prompt at /etc/autoscribe/prompts/docs.txt and on ApiKey for documentation.
- Errors within package population or documentation are surfaced as returned errors.

*/
func CreateRagChunkCounts(inputPackages []string, ApiKey string) error {
	gMst := golang.MST{}

	// Build the MST for all the important packages
	err := gMst.Populate(inputPackages)
	if err != nil {
		return fmt.Errorf("failed to populate packages: %v", err)
	}

	documentFilename := "/etc/autoscribe/prompts/docs.txt"

	promptBytes, err := os.ReadFile(documentFilename)
	if err != nil {
		return fmt.Errorf("failed to read %v: %v", documentFilename, err)
	}
	prompt := string(promptBytes)

	// Make sure the entire MST is documented
	_, err = mst.DocumentMST(&gMst, prompt, ApiKey)
	if err != nil {
		return fmt.Errorf("failed to document MST: %v", err)
	}

	sum := 0
	count := 0

	for _, pkg := range gMst.GetPackages() {
		for _, d := range pkg.GetFunctionDecls() {
			str, err := d.ToStringForAi()
			if err != nil {
				log.Errorf("failed to turn %v into AI string: %v", d.GetFullName(), err)
				continue
			}

			length := len(str)

			sum += length
			count += 1

			log.Infof("%v is %v chars long", d.GetFullName(), length)
		}
	}

	log.Infof("average is %v chars long", sum/count)

	return nil
}

/*
Summary: Creates a RAG training dataset for all function declarations found under inputPackages by building a Package MST, documenting it, computing overlapping/non-overlapping relationships, and generating per-function QA data. The resulting rag-data.jsonl is written to /tmp, and prompts are read from fixed paths.
"Build the MST for all the important packages" and "PopulatePackageInformation to derive imports and function declarations" are used to construct PackageNodes.
The function executes package population, MST/documentation, overlap graph generation, and per-definition QA data generation in parallel where appropriate.
The implementation relies on fixed prompt files at /etc/autoscribe/prompts/docs.txt and /etc/autoscribe/prompts/rag-data.txt and uses ApiKey for external queries.
Signature: func CreateRagData(outputFolder string, numQuestions int, inputPackages []string, ApiKey string) error
Parameters:
- outputFolder string: (currently unused by implementation; present for compatibility).
- numQuestions int: maximum number of questions to generate per function.
- inputPackages []string: filesystem paths used as package roots to load and process.
- ApiKey string: API key used for external Question/Answer queries during RAG data generation.
Returns:
- error: non-nil on failure to populate packages, read prompts, document the MST, generate overlapping data, or create/write the rag data file. On success, returns nil.
Errors/Exceptions:
- "failed to populate packages: %v" if gMst.Populate(inputPackages) fails.
- "failed to read %v: %v" if /etc/autoscribe/prompts/docs.txt cannot be read.
- "failed to document MST: %v" if mst.DocumentMST fails.
- "failed to read %v: %v" if /etc/autoscribe/prompts/rag-data.txt cannot be read.
- "Failed to generate overlapping database: %v" if GenerateOverlappingDatabase(definitions) fails.
- "failed to create %v: %v" if /tmp/rag-data.jsonl cannot be created.
Side Effects:
- Reads fixed prompt files; mutates and uses the MST to produce definitions; spawns goroutines to process definitions concurrently; writes rag data to /tmp/rag-data.jsonl; logs progress and errors.
Edge Cases & Assumptions:
- Assumes inputPackages contains valid package roots and that packages.Load (indirectly) yields packages for processing.
- If any per-definition processing fails, that definition is skipped with an error log; overall function may still complete.
- outputFolder is not currently used, and behavior depends on the hardcoded prompt paths and /tmp/rag-data.jsonl path.
- Concurrency is bounded by the provided definitions and synchronized access to ragData via ragDataMux.

*/
func CreateRagData(outputFolder string, numQuestions int, inputPackages []string, ApiKey string) ([]RagEntry, error) {
	gMst := golang.MST{}

	// Build the MST for all the important packages
	err := gMst.Populate(inputPackages)
	if err != nil {
		return nil, fmt.Errorf("failed to populate packages: %v", err)
	}

	documentFilename := "/etc/autoscribe/prompts/docs.txt"

	promptBytes, err := os.ReadFile(documentFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to read %v: %v", documentFilename, err)
	}
	prompt := string(promptBytes)

	// Make sure the entire MST is documented
	_, err = mst.DocumentMST(&gMst, prompt, ApiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to document MST: %v", err)
	}

	// For every function declaration in this package, generate RAG question data for it
	//  So we can associate questions with function definitions

	// fd := os.Open(fmt.Sprintf("%v/training-data.txt"))

	ragFilename := "/etc/autoscribe/prompts/rag-data.txt"

	promptBytes, err = os.ReadFile(ragFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to read %v: %v", ragFilename, err)
	}
	prompt = string(promptBytes)

	definitions := []mst.FunctionDetails{}

	ragData := []RagEntry{}

	for _, pkg := range gMst.GetPackages() {
		var decls []mst.FunctionDetails

		for _, d := range pkg.GetFunctionDecls() {
			decls = append(decls, d) // implicit interface conversion per element
		}

		definitions = append(definitions, decls...)
	}

	err = GenerateOverlappingDatabase(definitions)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate overlapping database: %v", err)
	}

	var wg sync.WaitGroup
	ragDataMux := sync.Mutex{}

	for _, decl := range definitions {
		wg.Add(1)

		go func(decl mst.FunctionDetails) {
			defer wg.Done()

			log.Infof("Handling definition %v", decl.GetFullName())

			NodeAsAiText, err := decl.ToStringForAi()
			if err != nil {
				log.Errorf("failed to generate ai string: %v", err)
				return
			}

			fullPrompt := fmt.Sprintf("%v\nCRITICAL: Only generate max %v questions\n---- BEGIN CODE ----\n%v\n---- END CODE ----\n", prompt, numQuestions, NodeAsAiText)

			var qas []QA

			// Get questions & answers for rag data
			for {
				QuestionString, err := call.Query5Nano(fullPrompt, ApiKey, nil)
				if err != nil {
					log.Errorf("failed to query 5 Nano: %v", err)
					continue
				}

				if err := json.Unmarshal([]byte(QuestionString), &qas); err == nil {
					break
				}
			}

			// Get non-related data for rag
			nonRelated, err := GetNonRelated(decl, definitions)
			if err != nil {
				log.Errorf("failed to get non-related functions: %v", err)
				return
			}

			nonRelatedCode := []string{}

			for _, el := range nonRelated {
				tmp, err := el.ToStringForAi()
				if err != nil {
					log.Errorf("failed to convert %v to Ai string: %v", el.GetFullName(), err)
					return
				}

				nonRelatedCode = append(nonRelatedCode, tmp)
			}

			declStr, err := decl.ToStringForAi()
			if err != nil {
				log.Errorf("failed to convert %v to Ai string: %v", decl.GetFullName(), err)
				return
			}

			// Create positive RAG context
			pCtx := []RagContext{
				RagContext{
					Title:      decl.GetFile(),
					Text:       declStr,
					Score:      0.0,
					TitleScore: 0.0,
					PassageId:  0,
				},
			}

			// Generate negative and hard negative content
			nCtx := make([]RagContext, len(nonRelated))
			hnCtx := make([]RagContext, len(nonRelated))
			for i, el := range nonRelated {
				elStr, err := el.ToStringForAi()
				if err != nil {
					log.Errorf("failed to convert %v to string for ai: %v", el.GetFullName(), err)
				}

				nCtx[i] = RagContext{
					Title:      el.GetFile(),
					Text:       elStr,
					Score:      0.0,
					TitleScore: 0.0,
					PassageId:  0,
				}

				hnCtx[i] = RagContext{
					Title:      el.GetFile(),
					Text:       elStr,
					Score:      0.0,
					TitleScore: 0.0,
					PassageId:  0,
				}
			}

			for _, qa := range qas {
				newData := RagEntry{
					Dataset:  "",
					Question: qa.Question,
					Answers:  []string{qa.Answer},

					PositiveCtxs:     pCtx,
					NegativeCtxs:     nCtx,
					HardNegativeCtxs: hnCtx,
				}

				ragDataMux.Lock()
				ragData = append(ragData, newData)
				ragDataMux.Unlock()
			}
		}(decl)
	}

	wg.Wait()

	ragDataFile := "/tmp/rag-data.jsonl"

	fd, err := os.Create(ragDataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create %v: %v", ragDataFile, err)
	}

	encoder := json.NewEncoder(fd)

	if err := encoder.Encode(ragData); err != nil {
		log.Errorf("failed to encode rag data: %v", err)
	}

	// for _, data := range ragData {
	//     if err := encoder.Encode(data); err != nil {
	//         log.Errorf("failed to encode rag data: %v", err)
	//     }
	// }

	return ragData, nil
}

// var OverLappingDatabase = make(map[mst.FunctionDetails]map[mst.FunctionDetails]bool)
var OverLappingDatabase = make(map[string]map[string]bool)

/*
Summary: Returns up to 5 non-related FunctionDetails for a given function declaration by querying precomputed non-overlapping results and selecting a subset (randomly when more than 5 are available).

Signature: func GetNonRelated(decl mst.FunctionDetails, definitions []mst.FunctionDetails) ([]mst.FunctionDetails, error)

Parameters:
- decl: mst.FunctionDetails — the function declaration to retrieve non-related definitions for.
- definitions: []mst.FunctionDetails — candidate definitions (provided for API compatibility; not used by this function).

Returns:
- []mst.FunctionDetails — up to 5 non-related function details for decl.
- error — non-nil if retrieving non-overlapping results fails; nil otherwise.

Errors/Exceptions:
- Propagates error from GetNonOverlapping as: "failed to get non overlapping: %v".

Side Effects:
- Calls GetNonOverlapping to obtain precomputed non-overlapping definitions.

Edge Cases & Assumptions:
- If fewer than 5 non-overlapping definitions exist, all are returned.
- If none exist, returns an empty slice.
- Assumes decl.GetFullName() yields a stable, unique key used to fetch results from the precomputed source.

*/
func GetNonRelated(decl mst.FunctionDetails, definitions []mst.FunctionDetails) ([]mst.FunctionDetails, error) {
	numNonOverlapping := 5

	NonRelated := make([]mst.FunctionDetails, 0, numNonOverlapping)

	// return GetNonOverlapping(decl, definitions)

	nonOverlapping, err := GetNonOverlapping(decl, definitions)
	if err != nil {
		return nil, fmt.Errorf("failed to get non overlapping: %v", err)
	}

	indicies := make([]int, min(numNonOverlapping, len(nonOverlapping)))

	// Get random indicies for our data
	if len(nonOverlapping) < numNonOverlapping {
		for i := range len(nonOverlapping) {
			indicies[i] = i
		}
	} else {
		for i := range numNonOverlapping {
			indicies[i] = rand.Intn(numNonOverlapping)
		}
	}

	for _, i := range indicies {
		NonRelated = append(NonRelated, nonOverlapping[i])
	}

	return NonRelated, nil
}

/*
Summary: Retrieves the precomputed non-overlapping FunctionDetails for a given function declaration by querying NonOverlappingDatabase with decl.GetFullName(), and logs the count found.
When to use: when you need the non-overlapping variants for a function without recomputing them.

Signature: func GetNonOverlapping(decl mst.FunctionDetails, definitions []mst.FunctionDetails) ([]mst.FunctionDetails, error)

Parameters:
- decl: mst.FunctionDetails — the function declaration to query for non-overlapping definitions.
- definitions: []mst.FunctionDetails — input candidate definitions (present for API compatibility; not used by this function).

Returns:
- []mst.FunctionDetails — the precomputed non-overlapping definitions for decl, or nil if none are found.
- error — always nil in the current implementation.

Errors/Exceptions: None (the function returns nil error).

Side Effects:
- Reads NonOverlappingDatabase.
- Emits a log via log.Infof indicating the number of non-overlapping results found for decl.

Edge Cases & Assumptions:
- If decl.GetFullName() is not present in NonOverlappingDatabase, result is nil.
- Assumes decl.GetFullName() provides a stable, unique key for the function.

*/
func GetNonOverlapping(decl mst.FunctionDetails, definitions []mst.FunctionDetails) ([]mst.FunctionDetails, error) {
	result := NonOverlappingDatabase[decl.GetFullName()]
	log.Infof("Found %v non overlapping for %v", len(result), decl.GetFullName())
	return result, nil
}

/*
Summary:
Initializes or updates OverLappingDatabase with the provided FunctionDetails and its transitive call graph, creating edges from each function to the functions it calls (and to their calls, recursively).

Signature:
func AddToOverlappingDatabase(f mst.FunctionDetails) error

Parameters:
- f mst.FunctionDetails: the function details to index in OverLappingDatabase. Identified by GetFullName().

Returns:
- error: always nil. Non-nil errors are not propagated; recursive calls are logged but ignored by this function.

Errors/Exceptions:
- If a recursive call to AddToOverlappingDatabase(c) returns an error, it is logged with log.Errorf but not returned to the caller.

Side Effects:
- Mutates the global OverLappingDatabase by:
  - creating an entry for f.GetFullName() if absent
  - recording direct edges from f to its calls
  - propagating edges from each callee to the caller (transitive closure)

Edge Cases & Assumptions:
- If f.GetFullName() already exists in OverLappingDatabase, the function returns immediately.
- For each c in f.GetCalls(), the function ensures c is indexed and then adds edges f -> c and, for every n in OverLappingDatabase[c.GetFullName()], adds f -> n.
- Cycles in the call graph are handled safely due to the initial existence check, preventing infinite recursion.
- Assumes f.GetFullName(), f.GetCalls(), and the methods on the nested calls are valid and non-nil.

*/
func AddToOverlappingDatabase(f mst.FunctionDetails) error {
	if _, ok := OverLappingDatabase[f.GetFullName()]; ok {
		return nil
	}

	OverLappingDatabase[f.GetFullName()] = make(map[string]bool)

	for _, c := range f.GetCalls() {
		err := AddToOverlappingDatabase(c)
		if err != nil {
			log.Errorf("Failed to add %v to overlapping database: %v", c.GetFullName(), err)
		}
	}

	for _, c := range f.GetCalls() {
		OverLappingDatabase[f.GetFullName()][c.GetFullName()] = true
		for n := range OverLappingDatabase[c.GetFullName()] {
			OverLappingDatabase[f.GetFullName()][n] = true
		}
	}

	return nil
}

var NonOverlappingDatabase = make(map[string][]mst.FunctionDetails)

/*
Summary:
Builds and caches the overlapping function graph and its corresponding non-overlapping
candidate lists from the provided definitions. It first populates OverLappingDatabase
by iterating definitions and invoking AddToOverlappingDatabase for each function detail,
then precomputes NonOverlappingDatabase entries by excluding overlapping functions per
declaration.
Signature:
func GenerateOverlappingDatabase(definitions []mst.FunctionDetails) error
Parameters:
- definitions []mst.FunctionDetails: input function details to index. Each item should expose
  GetFullName() and GetCalls() for graph construction.
Returns:
- error: non-nil if any AddToOverlappingDatabase call fails for a function; nil on success.
Errors/Exceptions:
- Propagates errors from AddToOverlappingDatabase with contextual message; processing continues
  for other definitions, but a non-nil error will be returned after logging.
Side Effects:
- Mutates global OverLappingDatabase by adding entries and edges (including transitive edges).
- Mutates global NonOverlappingDatabase by storing, for each declaration, the list of
  non-overlapping FunctionDetails.
Edge Cases & Assumptions:
- If a function already exists in OverLappingDatabase, AddToOverlappingDatabase handles it gracefully.
- Assumes definitions provide non-nil GetFullName(), GetCalls(), and that nested calls expose these
  methods similarly.
- Requires OverLappingDatabase and NonOverlappingDatabase to be initialized prior to invocation.
- Cycles in the call graph are handled by AddToOverlappingDatabase; this function relies on that behavior.

*/
func GenerateOverlappingDatabase(definitions []mst.FunctionDetails) error {
	// Build overlapping database first
	for _, f := range definitions {
		log.Infof("Adding %v to overlapping database", f.GetFullName())
		err := AddToOverlappingDatabase(f)
		if err != nil {
			return fmt.Errorf("failed to add %v to overlapping database: %v", f.GetFullName(), err)
		}
	}

	// Pre-compute non-overlapping lists
	for _, decl := range definitions {
		declName := decl.GetFullName()
		overlapping := OverLappingDatabase[declName]
		nonOverlapping := make([]mst.FunctionDetails, 0, len(definitions))

		for _, def := range definitions {
			defName := def.GetFullName()
			if defName == declName {
				continue
			}
			if !overlapping[defName] {
				nonOverlapping = append(nonOverlapping, def)
			}
		}

		NonOverlappingDatabase[declName] = nonOverlapping
		log.Infof("Pre-computed %v non-overlapping for %v", len(nonOverlapping), declName)
	}

	return nil
}


func CreateRagDataV2(outputFolder string, numQuestions int, inputPackages []string, ApiKey string) ([]RagData, error) {
	gMst := golang.MST{}

	// Build the MST for all the important packages
	err := gMst.Populate(inputPackages)
	if err != nil {
		return nil, fmt.Errorf("failed to populate packages: %v", err)
	}

	documentFilename := "/etc/autoscribe/prompts/docs.txt"

	promptBytes, err := os.ReadFile(documentFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to read %v: %v", documentFilename, err)
	}
	prompt := string(promptBytes)

	// Make sure the entire MST is documented
	_, err = mst.DocumentMST(&gMst, prompt, ApiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to document MST: %v", err)
	}

	// For every function declaration in this package, generate RAG question data for it
	//  So we can associate questions with function definitions

	// fd := os.Open(fmt.Sprintf("%v/training-data.txt"))

	ragFilename := "/etc/autoscribe/prompts/rag-data-v2.txt"

	promptBytes, err = os.ReadFile(ragFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to read %v: %v", ragFilename, err)
	}
	prompt = string(promptBytes)

	definitions := []mst.FunctionDetails{}

	ragData := []RagEntry{}

	for _, pkg := range gMst.GetPackages() {
		var decls []mst.FunctionDetails

		for _, d := range pkg.GetFunctionDecls() {
			decls = append(decls, d) // implicit interface conversion per element
		}

		definitions = append(definitions, decls...)
	}

	err = GenerateOverlappingDatabase(definitions)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate overlapping database: %v", err)
	}

	var wg sync.WaitGroup
	ragDataMux := sync.Mutex{}

	for _, decl := range definitions {
		wg.Add(1)

		go func(decl mst.FunctionDetails) {
			defer wg.Done()

			log.Infof("Handling definition %v", decl.GetFullName())

			NodeAsAiText, err := decl.ToStringForAi()
			if err != nil {
				log.Errorf("failed to generate ai string: %v", err)
				return
			}

			fullPrompt := fmt.Sprintf("%v\nCRITICAL: Only generate max %v questions\n---- BEGIN CODE ----\n%v\n---- END CODE ----\n", prompt, numQuestions, NodeAsAiText)

			var qas []QA

			// Get questions & answers for rag data
			for {
				QuestionString, err := call.Query5Nano(fullPrompt, ApiKey, nil)
				if err != nil {
					log.Errorf("failed to query 5 Nano: %v", err)
					continue
				}

				if err := json.Unmarshal([]byte(QuestionString), &qas); err == nil {
					break
				}
			}

			// Get non-related data for rag
			nonRelated, err := GetNonRelated(decl, definitions)
			if err != nil {
				log.Errorf("failed to get non-related functions: %v", err)
				return
			}

			nonRelatedCode := []string{}

			for _, el := range nonRelated {
				tmp, err := el.ToStringForAi()
				if err != nil {
					log.Errorf("failed to convert %v to Ai string: %v", el.GetFullName(), err)
					return
				}

				nonRelatedCode = append(nonRelatedCode, tmp)
			}

			declStr, err := decl.ToStringForAi()
			if err != nil {
				log.Errorf("failed to convert %v to Ai string: %v", decl.GetFullName(), err)
				return
			}

			// Create positive RAG context
			pCtx := []RagContext{
				RagContext{
					Title:      decl.GetFile(),
					Text:       declStr,
					Score:      0.0,
					TitleScore: 0.0,
					PassageId:  0,
				},
			}

			// Generate negative and hard negative content
			nCtx := make([]RagContext, len(nonRelated))
			hnCtx := make([]RagContext, len(nonRelated))
			for i, el := range nonRelated {
				elStr, err := el.ToStringForAi()
				if err != nil {
					log.Errorf("failed to convert %v to string for ai: %v", el.GetFullName(), err)
				}

				nCtx[i] = RagContext{
					Title:      el.GetFile(),
					Text:       elStr,
					Score:      0.0,
					TitleScore: 0.0,
					PassageId:  0,
				}

				hnCtx[i] = RagContext{
					Title:      el.GetFile(),
					Text:       elStr,
					Score:      0.0,
					TitleScore: 0.0,
					PassageId:  0,
				}
			}

			for _, qa := range qas {
				newData := RagEntry{
					Dataset:  "",
					Question: qa.Question,
					Answers:  []string{qa.Answer},

					PositiveCtxs:     pCtx,
					NegativeCtxs:     nCtx,
					HardNegativeCtxs: hnCtx,
				}

				ragDataMux.Lock()
				ragData = append(ragData, newData)
				ragDataMux.Unlock()
			}
		}(decl)
	}

	wg.Wait()

	ragDataFile := "/tmp/rag-data.jsonl"

	fd, err := os.Create(ragDataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create %v: %v", ragDataFile, err)
	}

	encoder := json.NewEncoder(fd)

	if err := encoder.Encode(ragData); err != nil {
		log.Errorf("failed to encode rag data: %v", err)
	}

	// for _, data := range ragData {
	//     if err := encoder.Encode(data); err != nil {
	//         log.Errorf("failed to encode rag data: %v", err)
	//     }
	// }

	return ragData, nil
}


