package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/BlankCanvasStudio/AutoScribe/pkg/cli"
	"github.com/BlankCanvasStudio/AutoScribe/pkg/config"

	"github.com/BlankCanvasStudio/AutoScribe/pkg/types/mst"
)


/*
Summary:
Main entry point for the application. It coordinates configuration loading and the execution of the root command. Specifically:
- Calls config.LoadConfig() to load and validate configuration.
- Logs the loaded configuration state.
- Executes the root command via cli.Execute().
- Logs final success.

Note: The code contains optional, currently disabled steps for creating database directories and for AST/documentation processing. These are present as commented blocks and are not executed in this version.

Signature:
func main()

Parameters:
- None

Returns:
- None

Errors/Exceptions:
- If config.LoadConfig() returns an error, the program terminates with log.Fatalf("Failed to load config: %v", err).

Side Effects:
- Mutates global state via config.LoadConfig() as needed by the application.
- Produces log output at various levels (Debug, Info).
- May mutate program state via cli.Execute() and potential global variables.

Edge Cases & Assumptions:
- Assumes config loading succeeds for normal startup.
- The AST and database directory creation paths are not active in this version.
- The behavior depends on external packages (config, log, mst, cli) being initialized elsewhere.

*/
func main() {
	err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

        log.Debugf("Loaded config!")

        /*
        // Make the data bases path, if we can
        if err := os.MkdirAll(config.GlobalDatabaseDir, 0755); err != nil {
            log.Errorf("failed to make %v: %v", config.GlobalDatabaseDir, err)
        }

        if err := os.MkdirAll(config.UserDatabaseDir, 0755); err != nil {
            log.Errorf("failed to make %v: %v", config.UserDatabaseDir, err)
        }
        */

        log.Infof("Using documentation database: %v", mst.DocumentationDb)

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
