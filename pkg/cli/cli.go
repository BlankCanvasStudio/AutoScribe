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
Summary: SetDebug configures the global logging level based on the package-level boolean debug. If debug is true, it enables debug-level logging and logs a "Debug logging enabled" message; otherwise it sets the logging level to info.
Signature: func SetDebug()
Parameters: none
Returns: none
Errors/Exceptions: none
Side Effects: modifies the global log level via log.SetLevel; may emit a debug log message when enabling debug
Edge Cases & Assumptions:
- relies on a package-scoped bool variable named debug
- relies on the log package exposing SetLevel, DebugLevel, InfoLevel, and Debug

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

            err := docs.CreateRagData("/tmp", 5, args, config.Settings.ApiKey)
            if err != nil {
                log.Errorf("Failed to generate RAG data: %v", err)
            }
	},
}

/*
Summary: Initializes and wires up flags and subcommands on the provided Cobra command. It sets core persistent flags, registers standard commands, generates and attaches custom directive commands, and exposes a documentation command.

Signature: func AddFlagsToCmd(in *cobra.Command)

Parameters:
- in: *cobra.Command — the command to augment with flags and subcommands. The function mutates in in place.

Returns: none

Errors/Exceptions:
- This function does not return errors. If creating custom commands fails, it calls log.Fatalf, terminating the process.

Side Effects:
- Adds persistent flags to in:
  - --debug (bool, default false)
  - --global (-g) (bool)
  - --user (-u) (bool)
  - --local (-l) (bool)
  - --prompt (-p) (string) stored in AdditionalPrompt
  - --config (-c) (string) stored in CfgFile
- Adds core commands: runCmd, versionCmd, directives.Cmd
- Builds and attaches custom directive commands via directives.CreateCustomCommands() and adds them to in
- Registers docs.UndocCmd
- May terminate the process via log.Fatalf on error

Edge Cases & Assumptions:
- in must be a non-nil pointer to a cobra.Command.
- directives.CreateCustomCommands() may return an error; in that case the function will terminate the process.
- The function relies on global variables: debug, AdditionalPrompt, CfgFile, and the presence of runCmd, versionCmd, directives.Cmd, docs.UndocCmd.

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
}

/*
Summary: Executes the application's root command after configuring flags and subcommands. It first logs a debug message, calls AddFlagsToCmd(rootCmd) to set up flags and subcommands, then runs rootCmd.Execute(). If execution fails, it terminates the process via log.Fatalf with the encountered error.

Signature: func Execute()

Parameters: none

Returns: none

Errors/Exceptions: On error from rootCmd.Execute(), the function terminates the process using log.Fatalf.

Side Effects: May mutate global state via AddFlagsToCmd(rootCmd); writes to logs; may terminate the process via log.Fatalf on error.

Edge Cases & Assumptions:
- in must be a non-nil pointer to a cobra.Command. (Note: the function relies on rootCmd and AddFlagsToCmd being available.)
- rootCmd.Execute() may return an error; such errors cause the process to exit via log.Fatalf.

*/
func Execute() {
	log.Debugf("Adding flags to cmd...")
	AddFlagsToCmd(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("%v", err)
	}
}
