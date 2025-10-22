package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"slices"

	log "github.com/sirupsen/logrus"

	"github.com/BlankCanvasStudio/AutoScribe/pkg/cli/directives"
	"github.com/BlankCanvasStudio/AutoScribe/pkg/cli/docs"
	"github.com/BlankCanvasStudio/AutoScribe/pkg/cli/helpers"
	"github.com/BlankCanvasStudio/AutoScribe/pkg/config"
	// "github.com/BlankCanvasStudio/AutoScribe/pkg/cli/document"
)

var CfgFile string
var GlobalLevel bool
var UserLevel bool

var AdditionalPrompt string

var debug bool

/*
Summary: Configures the global logger based on the package-level debug flag. If debug is true, enables DebugLevel and logs a Debug message; otherwise uses InfoLevel.
Signature: func SetDebug()
Side Effects: mutates the global logger level via log.SetLevel; may emit a debug log entry when enabling.
Edge Cases & Assumptions: relies on a package-level variable debug (bool) and on an initialized log logger. No return values.

*/
func SetDebug() {
	if debug {
		log.SetLevel(log.DebugLevel)
		log.Debug("Debug logging enabled")
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

var rootCmd = &cobra.Command{
	Use:   "asb",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
		log.Infof("hugo ran")
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		SetDebug()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of AutoScribe",
	Long:  `All software has versions. This is AutoScribe's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("AutoScribe version 0.1")
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run all the directives initialized in this repository",
	Long:  `Run all the directives initialized in this repository`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debugf("Running all directives in scope")

		configScopes, err := helpers.GetConfigsFromFlags(cmd)
		if err != nil {
			log.Fatalf("failed to get configs from flags: %v", err)
		}

		for name, directive := range config.Settings.Directives {
			if !slices.Contains(configScopes, directive.Scope) {
				log.Debugf("skipping directive %v; not in scope", name)
				continue
			}

			err := directive.Execute()
			if err != nil {
				log.Fatalf("failed to execte directive %v: %v", name, err)
			}
		}
	},
}

var ragDataCmd = &cobra.Command{
	Use:   "rag",
	Short: "generate rag data",
	Long:  `generate rag data`,
	Run: func(cmd *cobra.Command, args []string) {

		_, err := docs.CreateRagData("/tmp", 5, args, config.Settings.ApiKey)
		if err != nil {
			log.Errorf("Failed to generate RAG data: %v", err)
		}
	},
}

var ragCountChunkCmd = &cobra.Command{
	Use:   "chunk-count",
	Short: "show length of every chunk pre-rag",
	Long:  `show length of every chunk pre-rag`,
	Run: func(cmd *cobra.Command, args []string) {

		err := docs.CreateRagChunkCounts(args, config.Settings.ApiKey)
		if err != nil {
			log.Errorf("Failed to generate RAG data: %v", err)
		}
	},
}

/*
Summary: Adds persistent flags to the given Cobra command and registers top-level subcommands used by the CLI. This configures debug, config file, and directive-related options, and wires in core commands for execution.
Signature: func AddFlagsToCmd(in *cobra.Command)
Parameters:
- in: *cobra.Command to which persistent flags are attached and top-level subcommands are added.
Returns:
- none
Errors/Exceptions:
- none
Side Effects:
- mutates in by defining persistent flags and adding subcommands: runCmd, versionCmd, directives.Cmd, docs.UndocCmd, ragDataCmd, ragCountChunkCmd.
- Logs a debug message when initializing flags.
Edge Cases & Assumptions:
- Assumes runCmd, versionCmd, directives.Cmd, docs.UndocCmd, ragDataCmd, ragCountChunkCmd are defined elsewhere.
- Assumes debug, AdditionalPrompt, CfgFile are accessible package-level variables.

*/
func AddFlagsToCmd(in *cobra.Command) {
	log.Debugf("Initializing flags...")

	in.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging")
	in.PersistentFlags().BoolP("global", "g", false, "Utilize global settings (otherwise local folder)")
	in.PersistentFlags().BoolP("user", "u", false, "Utilize user settings (otherwise local folder)")
	in.PersistentFlags().BoolP("local", "l", false, "Utilize current folder settings (default)")
	in.PersistentFlags().StringVarP(&AdditionalPrompt, "prompt", "p", "", "Add additional context to your directive's prompt")

	in.PersistentFlags().StringVarP(&CfgFile, "config", "c", "", "Specify config file to use")

	in.AddCommand(runCmd)
	in.AddCommand(versionCmd)
	in.AddCommand(directives.Cmd)

	customCmds, err := directives.CreateCustomCommands()
	if err != nil {
		log.Fatalf("Couldn't build custom directives: %v", customCmds)
	}

	for _, cmd := range customCmds {
		in.AddCommand(cmd)
	}

	in.AddCommand(docs.UndocCmd)

	in.AddCommand(ragDataCmd)
	in.AddCommand(ragCountChunkCmd)
}

/*
Summary: Initializes the CLI by attaching persistent flags and top-level subcommands to rootCmd and then executes the Cobra command tree. Use at program start to configure and run the CLI.

Signature: func Execute()

Parameters: none

Returns: none

Errors/Exceptions: If rootCmd.Execute() returns an error, the function terminates the process with log.Fatalf("%v", err).

Side Effects: Logs a debug message, mutates rootCmd by adding persistent flags and subcommands (e.g., runCmd, versionCmd, directives.Cmd, docs.UndocCmd, ragDataCmd, ragCountChunkCmd), and executes the command tree.

Edge Cases & Assumptions: Assumes rootCmd is defined and accessible; assumes AddFlagsToCmd is defined elsewhere and wires in the necessary subcommands and flags; may rely on package-level variables such as AdditionalPrompt and CfgFile for configuration.

*/
func Execute() {
	log.Debugf("Adding flags to cmd...")
	AddFlagsToCmd(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("%v", err)
	}
}
