package calls

import (
    // "os"
    "fmt"

    log "github.com/sirupsen/logrus"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/types"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/files"
)


// Maybe this is bytes
// func CreateHelpMenuImplementation(data types.ConcatenatedFileContents, fileFormat types.SupportedFormat) (string, error) {
func CreateHelpMenuImplementation(fileFormat types.SupportedFormat) (string, error) {
    data, err := files.FormatCodeFilesForContext()

    helpmenuPrompt := fmt.Sprintf(
`Your task is to:
1. Read all files and understand their purpose and functionality.
2. Based on their contents, generate a **help menu implementation** (i.e., code that, when run, will print a help/usage menu) using the same programming language and packages used in the provided files.
3. The implementation should expose commands, flags, functions, or configuration options inferred from the files.

Ensure that the output:
- Is syntactically correct for the language and integrates with existing code consistently
- Uses the same framework or libraries found in the project (e.g., argparse vs click, commander.js vs yargs, etc.)
- Reflects the capabilities across all provided files
- Is production-ready and minimal
- Do not provide details on how the code work; only provide the code

Here are the files:

%v`, data)


    helpmenuPrompt = fmt.Sprintf(
`Your task is to:
1. Read all files and understand their purpose and functionality.
2. Based on their contents, generate a **help menu implementation** (i.e. code that, when run, will print a help/usage menu) using the same programming language and packages used in the provided files.

⚠️ IMPORTANT — When responding:
- Only output the **exact code** that should be added to implement the help menu.
- Also output a **small code example snippet** that demonstrates how to *hook the help menu into the existing codebase*.
- After the code, include a **very small section** indicating *where* the code should be inserted.
- Do **not** include explanations, summaries, or any additional commentary.

Here are the files:

%v`, data)

    log.Info("Querying ai for output...")
    helpmenuText, err := Query4_1Nano(helpmenuPrompt)
    if err != nil {
        return "", fmt.Errorf("failed to query 4.1 Nano: %v", err)
    }

    // ReadmePath := fmt.Sprintf("%v/README.md", config.OutputDirectory)

    // os.WriteFile(helpmenuText, []byte(readmeText), 0644)

    return helpmenuText, nil
}

// func CreateHelpMenuText(data types.ConcatenatedFileContents, fileFormat types.SupportedFormat) (string, error) {
func CreateHelpMenuText(fileFormat types.SupportedFormat) (string, error) {
    data, err := files.FormatCodeFilesForContext()

    helpmenuPrompt := fmt.Sprintf(
`Your task is to:
1. Read all files and understand their purpose and functionality.
2. Based on their contents, generate only the **help menu text output** (as if the user ran the program with '--help'), summarizing commands, flags, functions, and configuration options derived from the files.
3. Do *not* generate implementation code — only the help text a user would see.

Ensure that the output:
- Matches the style and conventions of the programming language and libraries used in the files
- Is clear, concise, and developer-friendly
- Reflects the functionality available across all files provided

Here are the files:

%v`, data)



    log.Info("Querying ai for output...")
    helpmenuText, err := Query4_1Nano(helpmenuPrompt)
    if err != nil {
        return "", fmt.Errorf("failed to query 4.1 Nano: %v", err)
    }

    // os.WriteFile(helpmenuText, []byte(readmeText), 0644)

    return helpmenuText, nil
}

