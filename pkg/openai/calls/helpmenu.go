package calls

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/BlankCanvasStudio/AutoScribe/pkg/config"
	"github.com/BlankCanvasStudio/AutoScribe/pkg/files"
	"github.com/BlankCanvasStudio/AutoScribe/pkg/types"
)


/*
*
 * Summary: Generates a help menu implementation code snippet based on the contents of code files formatted for context, and provides an example of how to integrate it into the existing codebase.
 * Signature: func CreateHelpMenuImplementationSample(fileFormat types.SupportedFormat) (string, error)
 * Parameters:
 * - fileFormat: types.SupportedFormat — The format of the supported files (currently unused in implementation).
 * Returns:
 * - string: The generated help menu implementation code.
 * - error: Error encountered during processing, if any.
 * Errors/Exceptions:
 * - Returns an error if the file formatting or API query fails.
 * Side Effects:
 * - Reads and processes code files for context.
 * - Performs an API call to generate help menu code.
 * Edge Cases & Assumptions:
 * - Assumes file formatting and API responses are successful.
 * - The `data` variable correctly contains formatted code files for context.

*/
func CreateHelpMenuImplementation(fileFormat types.SupportedFormat) (string, error) {
	log.Debugf("Input file for CreateHelpMenuImplementation: %v", config.EditFile)

	if config.EditFile == "" {
		return CreateHelpMenuImplementationSample(fileFormat)
	}

	return CreateHelpMenuAndUpdateImplementation(fileFormat)
}

/*
*
 * Creates a help menu implementation within the specified supported file format.
 * Reads all relevant files, generates code that prints a help/usage menu, and updates the file content.
 * Ensures all original functionality remains intact and appends only the help menu code.
 *
 * Signature:
 * func CreateHelpMenuAndUpdateImplementation(fileFormat types.SupportedFormat) (string, error)
 *
 * Parameters:
 * - fileFormat: types.SupportedFormat
 *   The target file format to identify and update the code.
 *
 * Returns:
 * - string: The updated file content with the help menu implemented.
 * - error: An error if the process fails.
 *
 * Errors/Exceptions:
 * - Returns an error if file reading or writing fails.
 *
 * Side Effects:
 * - Reads and writes to the file specified in config.EditFile.

*/
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

/*
*
Summary:
Generates a help menu implementation code snippet based on the contents of code files formatted for context, and provides an example of how to integrate it into the existing codebase.

Signature:
func CreateHelpMenuImplementationSample(fileFormat types.SupportedFormat) (string, error)

Parameters:
- fileFormat (types.SupportedFormat): The format of the supported files (currently unused in implementation).

Returns:
- string: The generated help menu implementation code.
- error: Error encountered during processing, if any.

Errors/Exceptions:
- Returns an error if the file formatting or API query fails.

Side Effects:
- Reads and processes code files for context.
- Performs an API call to generate help menu code.

Edge Cases & Assumptions:
- Assumes file formatting and API responses are successful.
- The `data` variable correctly contains formatted code files for context.

*/
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

/*
*
 * Summary:
 * Generates help menu text based on the provided file format by analyzing relevant files and summarizing available commands, flags, functions, and configurations.
 *
 * Signature:
 * func CreateHelpMenuText(fileFormat types.SupportedFormat) (string, error)
 *
 * Parameters:
 * - fileFormat: types.SupportedFormat, the file format to determine how to process files.
 *
 * Returns:
 * - string: the generated help menu text.
 * - error: an error if the process fails.
 *
 * Errors/Exceptions:
 * - Returns an error if the file formatting or query fails.

*/
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
