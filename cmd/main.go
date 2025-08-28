package main

import (
	// "fmt"
	log "github.com/sirupsen/logrus"

	// gast "go/ast"

	// "github.com/BlankCanvasStudio/AutoScribe/pkg/ast"
	"github.com/BlankCanvasStudio/AutoScribe/pkg/cli"
	"github.com/BlankCanvasStudio/AutoScribe/pkg/config"
	// "github.com/BlankCanvasStudio/AutoScribe/pkg/openai/calls"
)

/*
*
* Main function orchestrates configuration loading, CLI parsing, documentation and help menu generation, and AST processing.
*
* Usage:
* - Loads configuration settings from files and environment.
* - Parses command-line interface arguments.
* - Generates README, help menu implementation, or help menu text based on configuration flags.
* - Parses and optionally documents or undocument the AST file specified in configuration.
*
* Parameters: none
* Returns: none

*/
func main() {
	err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

        cli.Execute()

        /*
	err = config.ParseCli()
	if err != nil {
		log.Fatalf("Failed to parse cli: %v", err)
	}

	if config.MakeReadme {
		log.Infof("Making README.md for %v", config.ProjectDirectory)

		err := calls.CreateReadme(config.LanguageFileExtension)
		if err != nil {
			log.Fatalf("Failed to create a README: %v", err)
		}
	}

	if config.MakeHelpMenuImpl {
		log.Infof("Making Help Menu for %v", config.ProjectDirectory)

		impl, err := calls.CreateHelpMenuImplementation(config.LanguageFileExtension)
		if err != nil {
			log.Fatalf("Failed to create a help menu implementation: %v", err)
		}

		log.Infof("Help menu:\n\n%v\n\n", impl)
	}

	if config.MakeHelpMenuText {
		log.Infof("Making Help Menu for %v", config.ProjectDirectory)

		text, err := calls.CreateHelpMenuText(config.LanguageFileExtension)
		if err != nil {
			log.Fatalf("Failed to create the text for a help menu: %v", err)
		}

		log.Infof("Help menu:\n\n%v\n\n", text)
	}

	log.Infof("ast file name: %v", config.AstFileName)
	if config.AstFileName != "" {
		if config.UndocumentAst {
			ast.UndocumentDir(config.AstFileName, true)
			return
		}

		pkgNodes, err := ast.ParsePackage(config.AstFileName)
		if err != nil {
			log.Fatalf("Failed to parse package: %v", err)
		}

		for _, pkg := range pkgNodes {
			if config.DocumentAst {
				for _, f := range pkg.FunctionDeclarations {
					log.Infof("Documenting %v...", f.Info.Name)
					err := ast.DocumentFunctions(f.Info)
					if err != nil {
						log.Errorf("Failed to document %v: %v", f.Info.Name, err)
					}
				}

				err := pkg.UpdateDocsInFile()
				if err != nil {
					log.Fatalf("failed to update doc in file: %v", err)
				}
			} else {
				for _, decl := range pkg.FunctionDeclarations {
					decl.PrettyPrint("")
				}
			}
		}

	}
        */

	log.Info("AutoScribe-d successfully!")
}
