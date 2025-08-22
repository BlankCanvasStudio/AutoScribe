package main;

import (
    // "fmt"
    log "github.com/sirupsen/logrus"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/ast"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/config"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/openai/calls"
)


func main() {
    err := config.LoadConfig()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    err = config.ParseCli()
    if err != nil {
        log.Fatalf("Failed to parse cli: %v", err)
    }


    if config.MakeReadme {
        log.Infof("Making README.md for %v", config.ProjectDirectory)

        // err := calls.CreateReadme(formattedFileContents, config.LanguageFileExtension)
        err := calls.CreateReadme(config.LanguageFileExtension)
        if err != nil {
            log.Fatalf("Failed to create a README: %v", err)
        }
    }


    if config.MakeHelpMenuImpl {
        log.Infof("Making Help Menu for %v", config.ProjectDirectory)

        // impl, err := calls.CreateHelpMenuImplementation(formattedFileContents, config.LanguageFileExtension)
        impl, err := calls.CreateHelpMenuImplementation(config.LanguageFileExtension)
        if err != nil {
            log.Fatalf("Failed to create a help menu implementation: %v", err)
        }

        log.Infof("Help menu:\n\n%v\n\n", impl)
    }


    if config.MakeHelpMenuText {
        log.Infof("Making Help Menu for %v", config.ProjectDirectory)

        // text, err := calls.CreateHelpMenuText(formattedFileContents, config.LanguageFileExtension)
        text, err := calls.CreateHelpMenuText(config.LanguageFileExtension)
        if err != nil {
            log.Fatalf("Failed to create the text for a help menu: %v", err)
        }

        log.Infof("Help menu:\n\n%v\n\n", text)
    }

    if config.AstFileName != "" {
        pkgNodes, err := ast.ParsePackage(config.AstFileName)
        if err != nil {
            log.Fatalf("failed to parse package: %v", err)
        }

        for _, pkg := range pkgNodes {
            if config.DocumentAst {
                for _, f := range pkg.FunctionDeclarations {
                    log.Infof("Documenting %v...", f.Name)
                    ast.DocumentFunctions(f)
                }
                // for _, f := range pkg.FunctionDeclarations {
                //     f.PrettyPrint("")
                // }
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



    log.Info("AutoScribe-d successfully!")
}

