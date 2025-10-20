package docs

import (
    "os"
    "fmt"
    "sync"
    // "strings"
    "math/rand"
    "encoding/json"

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
    Dataset           string       `json:"dataset"`
    Question          string       `json:"question"`
    Answers           []string     `json:"answers"`

    PositiveCtxs      []RagContext `json:"positive_ctxs"`
    NegativeCtxs      []RagContext `json:"negative_ctxs"`
    HardNegativeCtxs  []RagContext `json:"hard_negative_ctxs"`
}

type QA struct {
    Question string `json:"question"`
    Answer   string `josn:"answer"`
}

func CreateRagData(outputFolder string, numQuestions int, inputPackages []string, ApiKey string) error {
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


    // For every function declaration in this package, generate RAG question data for it
    //  So we can associate questions with function definitions

    // fd := os.Open(fmt.Sprintf("%v/training-data.txt"))

    ragFilename := "/etc/autoscribe/prompts/rag-data.txt"

    promptBytes, err = os.ReadFile(ragFilename)
    if err != nil {
            return fmt.Errorf("failed to read %v: %v", ragFilename, err)
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
        return fmt.Errorf("Failed to generate overlapping database: %v", err)
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
                    Title: decl.GetFile(),
                    Text: declStr,
                    Score: 0.0,
                    TitleScore: 0.0,
                    PassageId: 0,
                },
            }

            // Generate negative and hard negative content
            nCtx  := make([]RagContext, len(nonRelated))
            hnCtx := make([]RagContext, len(nonRelated))
            for i, el := range nonRelated {
                elStr, err := el.ToStringForAi()
                if err != nil {
                    log.Errorf("failed to convert %v to string for ai: %v", el.GetFullName(), err)
                }

                nCtx[i] = RagContext{
                    Title: el.GetFile(),
                    Text: elStr,
                    Score: 0.0,
                    TitleScore: 0.0,
                    PassageId: 0,
                }

                hnCtx[i] = RagContext{
                    Title: el.GetFile(),
                    Text: elStr,
                    Score: 0.0,
                    TitleScore: 0.0,
                    PassageId: 0,
                }
            }

            for _, qa := range qas {
                newData := RagEntry{
                    Dataset: "",
                    Question: qa.Question,
                    Answers: []string{qa.Answer},

                    PositiveCtxs: pCtx,
                    NegativeCtxs: nCtx,
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
        return fmt.Errorf("failed to create %v: %v", ragDataFile, err)
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

    return nil
}

// var OverLappingDatabase = make(map[mst.FunctionDetails]map[mst.FunctionDetails]bool)
var OverLappingDatabase = make(map[string]map[string]bool)

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


func GetNonOverlapping(decl mst.FunctionDetails, definitions []mst.FunctionDetails) ([]mst.FunctionDetails, error) {
    result := NonOverlappingDatabase[decl.GetFullName()]
    log.Infof("Found %v non overlapping for %v", len(result), decl.GetFullName())
    return result, nil
}

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

