package directives

import (
	"fmt"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"

	"github.com/BlankCanvasStudio/AutoScribe/pkg/cli/helpers"
	"github.com/BlankCanvasStudio/AutoScribe/pkg/config"
	"github.com/BlankCanvasStudio/AutoScribe/pkg/types"
)

var Cmd = &cobra.Command{
	Use:   "directive",
	Short: "Update information about & add custom directives",
	Long:  ``,
	Run:   func(cmd *cobra.Command, args []string) {},
}

/*
Summary: During package initialization, this init function registers CreateCmd as a subcommand by calling Cmd.AddCommand(CreateCmd).

Signature: func init()

Parameters: none

Returns: none

Errors/Exceptions: None documented; relies on the behavior of Cmd.AddCommand for any runtime errors.

Side Effects: Mutates Cmd by adding CreateCmd to its subcommands.

Edge Cases & Assumptions: Assumes Cmd and CreateCmd are defined and initialized in the package before init runs.

*/
func init() {
	Cmd.AddCommand(CreateCmd)
}

var CreateCmd = &cobra.Command{
	Use:   "create [name] [prompt file]",
	Short: "",
	Long:  ``,
	Args:  cobra.ExactArgs(2),

	Run: func(cmd *cobra.Command, args []string) {
		configFiles, err := helpers.GetConfigsFromFlags(cmd)
		if err != nil {
			log.Fatalf("failed to get the correct config files: %v", err)
		}

		name := args[0]
		prompt := args[1]

		err = CreateNewDirective(name, prompt, configFiles)
		if err != nil {
			log.Fatalf("failed to create new directive: %v", err)
		}
	},
}

/*
var RemoveCmd = &cobra.Command{
    Use:   "remove [directive] [parameter] <values>",
    Short: "",
    Long:  ``,

    Run: func(cmd *cobra.Command, args []string) {
        configFiles, err := helpers.GetConfigsFromFlags(cmd)
        if err != nil {
            log.Fatalf("failed to get the correct config files: %v", err)
        }

        name := args[0]
        field := args[1]

        err = CreateNewDirective(name, prompt, configFiles)
        if err != nil {
            log.Fatalf("failed to create new directive: %v", err)
        }
    },
}f
*/

/*
Summary: Builds a slice of Cobra commands for each custom directive defined in config.Settings.Directives. For every directive, it creates a parent command with Use: "<name> [init|add|ignore|model|docs|prompt|prompt-text|run] <targets>" and attaches subcommands for Run, Init, Export, and per-field updates (string and array updates) derived from helper functions.
Signature: func CreateCustomCommands() ([]*cobra.Command, error)
Parameters: none
Returns:
  - []*cobra.Command: the constructed parent commands, one per directive.
  - error: non-nil if any helper fails to create a subcommand for a directive.
Errors/Exceptions: If CreateCustomRunHandler, CreateCustomInitHandler, CreateCustomExportHandler, AddFieldUpdates, or AddArrayUpdates return an error, the function returns nil and that error.
Side Effects: Logs a debug message about the number of directives; may perform allocation and wiring of commands; may invoke helper functions to create subcommands.
Edge Cases & Assumptions:
  - Assumes config.Settings.Directives is a map-like collection of directives with keys as names and values exposing Short and Description.
  - If a directive's subcommand creation fails, the error is propagated and the function aborts.
  - If config.Settings.Directives is empty, returns an empty slice without errors.

*/
func CreateCustomCommands() ([]*cobra.Command, error) {
	log.Debugf("custom directives found: %v", len(config.Settings.Directives))

	var customCmds = make([]*cobra.Command, 0, len(config.Settings.Directives))

	// We should auto generate these from lists now
	init_options := "[init|add|ignore|model|docs|prompt|prompt-text|run]"

	for name, directive := range config.Settings.Directives {

		var createCmd = &cobra.Command{
			Use:   fmt.Sprintf("%v %v <targets>", name, init_options),
			Short: directive.Short,
			Long:  directive.Description,
			// Args:  cobra.ExactArgs(2),

			Run: func(cmd *cobra.Command, args []string) {},
		}

		log.Debugf("Creating custom handler for %v", name)

		Run, err := CreateCustomRunHandler(directive)
		if err != nil {
			return nil, fmt.Errorf("failed to create custom kind handler for %v: %v", name, err)
		}
		createCmd.AddCommand(Run)

		Init, err := CreateCustomInitHandler(directive)
		if err != nil {
			return nil, fmt.Errorf("failed to create custom kind handler for %v: %v", name, err)
		}
		createCmd.AddCommand(Init)

		Export, err := CreateCustomExportHandler(directive)
		if err != nil {
			return nil, fmt.Errorf("failed to create custom kind handler for %v: %v", name, err)
		}
		createCmd.AddCommand(Export)

		// If I ever have to add to this, make it a for loop with function pointers plz
		//      vim makes copy paste too easy and I'm trying to move fast
		// UPDATE: I did it
		stringCmds, err := AddFieldUpdates(directive)
		if err != nil {
			log.Fatalf("failed to add a string field to %v: %v", directive.Name, err)
		}

		for _, stringCmd := range stringCmds {
			createCmd.AddCommand(stringCmd)
		}

		arrayCmds, err := AddArrayUpdates(directive)
		if err != nil {
			log.Fatalf("failed to add a string field to %v: %v", directive.Name, err)
		}

		for _, arrayCmd := range arrayCmds {
			log.Debugf("Adding array command: %+v", arrayCmd)
			createCmd.AddCommand(arrayCmd)
		}

		customCmds = append(customCmds, createCmd)

	}

	return customCmds, nil
}

/*
Summary: Constructs a Cobra command named "run" that executes the provided directive by invoking directive.Execute() when run, and includes a descriptive text derived from directive.Name and directive.Output.

Signature: func CreateCustomRunHandler(directive types.Directive) (*cobra.Command, error)

Parameters:
- directive: types.Directive — the directive to be executed when the command is invoked. The command description uses directive.Name and directive.Output.

Returns:
- *cobra.Command — a ready-to-use command configured with Use: "run", zero-argument requirement, and a Run handler that executes the directive.
- error — always nil in this implementation.

Errors/Exceptions:
- The Run function calls directive.Execute(); if it returns an error, log.Fatalf is invoked, terminating the process with the error.

Side Effects:
- Logs a debug message at creation time: "Executing %v custom run handler".
- May perform I/O through directive.Execute(); on error, the process exits via log.Fatalf.

Edge Cases & Assumptions:
- Assumes directive.Execute() handles the directive's actual execution logic across possible kinds.
- Uses directive.Name for the log message and directive.Output for the command description; empty fields may yield empty descriptions.
- The command enforces zero arguments (Args: cobra.ExactArgs(0)).

*/
func CreateCustomRunHandler(directive types.Directive) (*cobra.Command, error) {
	log.Debugf("Executing %v custom run handler", directive.Name)
	desc := fmt.Sprintf("Run the %v directive and save to `%v`", directive.Name, directive.Output)

	return &cobra.Command{
		Use:   "run",
		Short: desc,
		Long:  desc,
		Args:  cobra.ExactArgs(0),

		Run: func(cmd *cobra.Command, args []string) {
			err := directive.Execute()
			if err != nil {
				log.Fatalf("failed to execte directive %v: %v", directive.Name, err)
			}
		},
	}, nil
}

/*
Summary: Creates a Cobra command that initializes the given directive by loading and applying
the directive's configuration from the project's config files into the local scope. Use this
to bootstrap or rehydrate a directive from existing configuration during CLI usage.
Signature: func CreateCustomInitHandler(directive types.Directive) (*cobra.Command, error)
Parameters:
- directive: types.Directive — the directive to initialize within the local project configuration.
Returns:
- *cobra.Command — the initialized command with Use: "init [files]".

- error — always nil per the implementation; any errors during execution are handled inside the command
  via log.Fatalf in Run.
Errors/Exceptions:
- No error return from the function itself; internal failures during Run are reported by log.Fatalf (terminating
  the process).
Side Effects:
- Logs a debug message indicating the custom init handler execution.
- Reads and potentially mutates CLI flags (global, user, local) to determine config scope.
- Retrieves configuration files via helpers.GetConfigsFromFlags and may initialize directive data via
  InitDirectiveInLocalScope, which can write updated configuration to disk.
Edge Cases & Assumptions:
- Assumes flags "global", "user", "local" are boolean and "config" is a string flag used by helpers.
- If no scopes are active, the code enables global and user scopes by setting their flags to true.
- The directive is expected to be a valid types.Directive; if InitDirectiveInLocalScope fails, execution
  is terminated with log.Fatalf.

*/
func CreateCustomInitHandler(directive types.Directive) (*cobra.Command, error) {
	log.Debugf("Executing %v custom init handler", directive.Name)

	desc := fmt.Sprintf("Initialize the %v directive saved in global configs into this project", directive.Name)

	return &cobra.Command{
		Use:   "init [files]",
		Short: desc,
		Long:  desc,

		Run: func(cmd *cobra.Command, args []string) {
			// Check if user specified what config to draw from
			globalScope, userScope, _, _, err := helpers.GetConfigScopeFlags(cmd)
			if err != nil {
				log.Fatalf("failed to load config scope flags: %v", err)
			}

			// If they didn't we are drawing from user & global files
			if !globalScope && !userScope {
				cmd.Flags().Set("global", "true")
				cmd.Flags().Set("user", "true")
			}

			// Get the config files
			configFiles, err := helpers.GetConfigsFromFlags(cmd)
			if err != nil {
				log.Fatalf("failed to get the correct config files: %v", err)
			}

			err = InitDirectiveInLocalScope(directive, args, configFiles)
			if err != nil {
				log.Fatalf("failed to init %v locally: %v", directive.Name, err)
			}

		},
	}, nil
}

/*
Summary
Creates a Cobra command that exports a given directive to configured files. The command is named "export" and uses the directive's Name to describe its purpose. It initializes export behavior and delegates to ExportCustomDirective to persist the directive across selected config files.

Signature
func CreateCustomExportHandler(directive types.Directive) (*cobra.Command, error)

Parameters
- directive: types.Directive — the directive to export; its Name identifies the entry in Settings.Directives to update.

Returns
- *cobra.Command — a ready-to-use Cobra command named "export" for the directive.
- error — always nil in this implementation (the function constructs and returns the command or an error during construction, if any).

Errors/Exceptions
- Construction errors during command initialization would be returned; runtime errors inside Run are handled via log.Fatalf and terminate the process.

Side Effects
- Creates a new Cobra command instance.
- On execution, reads and validates configuration scope and config files, then updates persistent directive state across files.
- May terminate the process via log.Fatalf on runtime failures.

Edge Cases & Assumptions
- Assumes directive.Name is a valid key in the settings for export.
- If config file collection yields no paths, ExportCustomDirective is effectively a no-op and returns nil.
- Runtime errors during flag retrieval or directive export are handled by terminating the process with a logged error.

*/
func CreateCustomExportHandler(directive types.Directive) (*cobra.Command, error) {
	log.Debugf("Executing %v custom init handler", directive.Name)

	desc := fmt.Sprintf("Export the %v directive to various configs", directive.Name)

	return &cobra.Command{
		Use:   "export [files]",
		Short: desc,
		Long:  desc,

		Run: func(cmd *cobra.Command, args []string) {

			// Check if user specified what config to draw from
			globalScope, _, _, customScope, err := helpers.GetConfigScopeFlags(cmd)
			if err != nil {
				log.Fatalf("Failed to load config scope flags: %v", err)
			}

			// Export to user scope by default
			if !globalScope && !customScope {
				cmd.Flags().Set("user", "true")
			}

			// Get the config files
			configFiles, err := helpers.GetConfigsFromFlags(cmd)
			if err != nil {
				log.Fatalf("Failed to get the correct config files: %v", err)
			}

			err = ExportCustomDirective(directive, configFiles)
			if err != nil {
				log.Fatalf("Failed to export custom directive: %v", err)
			}
		},
	}, nil
}

type Field struct {
	Name  string
	Use   string
	Short string
	Long  string
}

var StringFieldsToUpdate = []Field{
	Field{
		Name:  "Kind",
		Use:   "kind <text argument>",
		Short: "set the directive kind (file base or recursion base)",
		Long:  "set the directive kind (file base or recursion base)",
	},

	Field{
		Name:  "Description",
		Use:   "description <text argument>",
		Short: "set the full description of a custom directive",
		Long:  "set the full description of a custom directive",
	},

	Field{
		Name:  "Short",
		Use:   "short <text argument>",
		Short: "set the short description of a custom directive",
		Long:  "set the short description of a custom directive",
	},

	Field{
		Name:  "Prompt",
		Use:   "prompt <text argument>",
		Short: "set the prompt file for a directive",
		Long:  "set the prompt file for a directive",
	},

	Field{
		Name:  "PromptText",
		Use:   "prompt-text <text argument>",
		Short: "set the prompt text for a directive (override the prompt file)",
		Long:  "set the prompt text for a directive (override the prompt file)",
	},

	Field{
		Name:  "PromptText",
		Use:   "prompt-text <text argument>",
		Short: "set the prompt text for a directive (override the prompt file)",
		Long:  "set the prompt text for a directive (override the prompt file)",
	},

	Field{
		Name:  "Model",
		Use:   "model <text argument>",
		Short: "set the model for a particular directive",
		Long:  "set the model for a particular directive",
	},

	Field{
		Name:  "Output",
		Use:   "output <text argument>",
		Short: "set the output file for a directive's result",
		Long:  "set the output file for a directive's result",
	},

	Field{
		Name:  "ApiKey",
		Use:   "apikey <text argument>",
		Short: "set a directive's api key",
		Long:  "set a directive's api key",
	},

	Field{
		Name:  "LocalDocs",
		Use:   "local-docs <path>",
		Short: "set where a directive can locally source documentation not written in files",
		Long:  "set where a directive can locally source documentation not written in files",
	},
}

/*
Summary:
AddFieldUpdates creates a set of Cobra commands, one for each field listed in StringFieldsToUpdate, to update string fields on a given directive. Each generated command accepts exactly one argument and, when run, updates the corresponding field on the provided directive across configuration files selected via command flags.

Signature:
func AddFieldUpdates(directive types.Directive) ([]*cobra.Command, error)

Parameters:
- directive: types.Directive — the directive to update when any generated command runs.

Returns:
- []*cobra.Command — the created commands for updating each string field.
- error — always nil under current implementation (no error is returned by this function).

Errors/Exceptions:
- This function itself does not return errors. Errors encountered during command execution are surfaced via log.Fatalf within the Run closures (e.g., failures in GetConfigsFromFlags or UpdateDirectiveFieldInConfigs).

Side Effects:
- Constructs and returns new []*cobra.Command objects.
- Execution of each command may read flags from cmd, derive configFiles, and update and persist directive changes to config files.

Edge Cases & Assumptions:
- Assumes StringFieldsToUpdate is populated with entries containing Use, Short, Long, and Name.
- Run closures rely on the loop variable field; without capturing a local copy inside the loop, all commands may reference the final field value due to closure semantics.
- If config file updates fail during Run, the process terminates via log.Fatalf (no error is returned to the caller).

*/
func AddFieldUpdates(directive types.Directive) ([]*cobra.Command, error) {
	ret := make([]*cobra.Command, 0, len(StringFieldsToUpdate))

	for _, field := range StringFieldsToUpdate {
		ret = append(ret, &cobra.Command{
			Use:   field.Use,
			Short: field.Short,
			Long:  field.Long,
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				log.Debugf("%v %v triggered", field.Name, directive)

				configFiles, err := helpers.GetConfigsFromFlags(cmd)
				if err != nil {
					log.Fatalf("failed to get the correct config files: %v", err)
				}

				err = UpdateDirectiveFieldInConfigs(directive, field.Name, args[0], configFiles)
				if err != nil {
					log.Fatalf("Failed to update directive field %v: %v", field.Name, err)
				}
			},
		})
	}

	return ret, nil
}

var ArrayFieldsToUpdate = []Field{
	Field{
		Name:  "Focus",
		Use:   "add <text arguments>",
		Short: "add functions or files for autoscibe to add documentation to / focus on",
		Long:  "add functions or files for autoscibe to add documentation to / focus on",
	},

	Field{
		Name:  "Ignore",
		Use:   "ignore <text arguments>",
		Short: "ignore functions or files",
		Long:  "ignore functions or files",
	},

	Field{
		Name:  "Servers",
		Use:   "server <text arguments>",
		Short: "set where a directive sources documentation from if it can't be found locally (godocs is implicit)",
		Long:  "set where a directive sources documentation from if it can't be found locally (godocs is implicit)",
	},
}

/*
Summary
Creates a set of Cobra commands to update exported array/slice fields on the provided directive.
Each command corresponds to an entry in ArrayFieldsToUpdate and, when executed, loads config file
paths from flags and updates the specified field across those config files.

Signature
func AddArrayUpdates(directive types.Directive) ([]*cobra.Command, error)

Parameters
- directive: types.Directive — the directive whose array fields will be updated.

Returns
- []*cobra.Command — commands for updating each array field; the slice length equals len(ArrayFieldsToUpdate).
- error — always nil for this function; runtime errors are surfaced by the Run handlers via log.Fatalf.

Errors/Exceptions
- The function itself does not return an error; failures during flag resolution or config updates cause
  the Run handlers to log and terminate the process via log.Fatalf.

Side Effects
- Allocates and returns a slice of *cobra.Command with Run callbacks that perform flag parsing, config loading,
  and directive array updates on disk.

Edge Cases & Assumptions
- If ArrayFieldsToUpdate is empty, returns an empty slice.
- Each command uses field.Use, field.Short, field.Long, and field.Name from the corresponding ArrayFieldsToUpdate entry.
- Assumes helpers.GetConfigsFromFlags(cmd) and UpdateDirectiveArrayInConfigs(...) behave as documented elsewhere.

*/
func AddArrayUpdates(directive types.Directive) ([]*cobra.Command, error) {
	ret := make([]*cobra.Command, 0, len(StringFieldsToUpdate))

	for _, field := range ArrayFieldsToUpdate {
		ret = append(ret, &cobra.Command{
			Use:   field.Use,
			Short: field.Short,
			Long:  field.Long,

			Run: func(cmd *cobra.Command, args []string) {
				log.Debugf("%v triggered: %v", field.Name, directive)

				configFiles, err := helpers.GetConfigsFromFlags(cmd)
				if err != nil {
					log.Fatalf("failed to get the correct config files: %v", err)
				}

				err = UpdateDirectiveArrayInConfigs(directive, field.Name, args, configFiles)
				if err != nil {
					log.Fatalf("Failed to update directive array %v: %v", directive.Name, err)
				}
			},
		})
	}

	return ret, nil
}
