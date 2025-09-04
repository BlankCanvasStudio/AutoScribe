package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/BlankCanvasStudio/AutoScribe/pkg/cli"
	"github.com/BlankCanvasStudio/AutoScribe/pkg/config"
)


func main() {
	err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

        log.Debugf("Loaded config!")

        cli.Execute()

        /*
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
