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
// func CreateReadme(data types.ConcatenatedFileContents, fileFormat types.SupportedFormat) error {
func CreateReadme(fileFormat types.SupportedFormat) error {
    data, err := files.FormatCodeFilesForContext()

    buildData, err := files.FormatBuildFilesForContext()

    data += buildData

    readmePrompt := fmt.Sprintf(
`You are an expert technical writer and software engineer.

ONLY OUTPUT A SINGLE FILE CALLED 'README.md' — **do not include any other text, commentary, or explanation**.

The README should:
- Clearly describe the purpose of the project.
- Explain how to install, configure, and run it.
- Include examples of usage.
- Document any important dependencies or architectures in the codebase.
- Use proper Markdown formatting.
- If there is a Makefile or other build system, include the configuration, building, and installation steps using that system
- If the installation and build process includes multiple steps, you can add extra documentation for those steps, but don't be too verbose
- Do not include *End of README.* or any similar stort of annotations

**IMPORTANT:**
- Use the build and install commands given by the build system. Do not write your own shell script to install the code, unless no build system is provided
- For build process make sure to include: installing dependencies, building code, installing necessary parts of package. **Use existing build system wherever possible; don't write your own code if you don't have to**

Here are the project files:

%v`, data)



    log.Info("Querying ai for output...")
    readmeText, err := Query4_1Nano(readmePrompt)
    if err != nil {
        return fmt.Errorf("failed to query 4.1 Nano: %v", err)
    }

    inputFile := config.EditFile
    if inputFile == "" {
        inputFile = "README.md"
    }

    ReadmePath := fmt.Sprintf("%v/%v", config.OutputDirectory, inputFile)

    os.WriteFile(ReadmePath, []byte(readmeText), 0644)

    return nil
}

