package calls

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/BlankCanvasStudio/AutoScribe/pkg/files"
	"github.com/BlankCanvasStudio/AutoScribe/pkg/types"
	"github.com/BlankCanvasStudio/AutoScribe/pkg/config"
	ai "github.com/BlankCanvasStudio/AutoScribe/pkg/openai"
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

        helpmenuText, err := QueryFromFile(ai.GPT_41_Nano, config.HelpMenuPromptCodeUpdate, data)
        if err != nil {
            return "", fmt.Errorf("failed to query from file: %v", err)
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

        helpmenuText, err := QueryFromFile(ai.GPT_41_Nano, config.HelpMenuPromptCodeExample, data)
        if err != nil {
            return "", fmt.Errorf("failed to query from file: %v", err)
        }

	os.WriteFile(config.EditFile, []byte(helpmenuText), 0644)

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

        helpmenuText, err := QueryFromFile(ai.GPT_41_Nano, config.HelpMenuPromptText, data)
        if err != nil {
            return "", fmt.Errorf("failed to query from file: %v", err)
        }

	os.WriteFile(config.EditFile, []byte(helpmenuText), 0644)

	return helpmenuText, nil
}
