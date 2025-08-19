package calls

import (
    "os"
    "fmt"

    log "github.com/sirupsen/logrus"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/files"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/types"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/config"
)


// Maybe this is bytes
// func CreateHelpMenuImplementation(data types.ConcatenatedFileContents, fileFormat types.SupportedFormat) (string, error) {
func CreateHelpMenuImplementation(fileFormat types.SupportedFormat) (string, error) {
    log.Debugf("Input file for CreateHelpMenuImplementation: %v", config.EditFile)

    if config.EditFile == "" {
        return CreateHelpMenuImplementationSample(fileFormat)
    }

    return CreateHelpMenuAndUpdateImplementation(fileFormat)
}


func CreateHelpMenuAndUpdateImplementation(fileFormat types.SupportedFormat) (string, error) {
    data, err := files.FormatCodeFilesForContext()

    helpmenuPrompt := fmt.Sprintf(
`Your task is to:
1. Read all files and understand their purpose and functionality.
2. Based on their contents, generate a **help menu implementation** (i.e. code that, when run, will print a help/usage menu) using the same programming language and packages used in the provided files.
3. Return an updated version of %v with the help menu implemented. Make sure all the original functionality is still implemented

⚠️ IMPORTANT — When responding:
- Only output the **exact rewrite of the file**
- Do **not** include explanations, summaries, or any additional commentary.
- Do **not** delete any functionality to the file. You are only allowed to **add** to the code
- Do **not** write or adjust any code that isn't related to the help menu
- Do **not** adjust the spacing in the file
- Do **not** adjust the number of line breaks or their placement in the file
- Do **not** adjust any of the other cli parameters. You will break software dependencies
- Mimic the spacing patterns used in the file
- Mimic the layout, placement, and functionality of the snippet provided when appending to the code

Here are the files:

%v`, config.EditFile, data)

    log.Info("Querying ai for output...")
    helpmenuText, err := Query4_1Nano(helpmenuPrompt)
    if err != nil {
        return "", fmt.Errorf("failed to query 4.1 Nano: %v", err)
    }

    os.WriteFile(config.EditFile, []byte(helpmenuText), 0644)

    return helpmenuText, nil
}


func CreateHelpMenuImplementationSample(fileFormat types.SupportedFormat) (string, error) {
    data, err := files.FormatCodeFilesForContext()

    helpmenuPrompt := fmt.Sprintf(
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

