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
Summary: Creates a set of custom Cobra commands based on directives found in config.Settings.Directives. For each directive, a parent command is created with a Use that includes the directive name and a fixed set of subcommands (run/init/export and per-field commands). The function wires in handler creators (CreateCustomRunHandler, CreateCustomInitHandler, CreateCustomExportHandler, AddFieldUpdates, AddArrayUpdates) to build the directive-specific CLI.

Signature: func CreateCustomCommands() ([]*cobra.Command, error)

Returns:
- []*cobra.Command: slice of constructed parent commands (one per directive).
- error: non-nil if any per-directive handler creation fails; otherwise nil.

Errors/Exceptions:
- Returns an error if CreateCustomRunHandler, CreateCustomInitHandler, or CreateCustomExportHandler fail to create a command for a directive.
- AddFieldUpdates or AddArrayUpdates failures cause an error to be returned.

Side Effects:
- Reads config.Settings.Directives and constructs in-memory Cobra Command trees; does not persist or mutate configuration by itself.
- Logs debug information during directive discovery and command creation.

Edge Cases & Assumptions:
- If config.Settings.Directives is empty, returns an empty slice with nil error.
- Uses init_options "[init|add|ignore|model|docs|prompt|prompt-text|run]" to compose Use for each directive.
- Assumes CreateCustomRunHandler, CreateCustomInitHandler, CreateCustomExportHandler, AddFieldUpdates, and AddArrayUpdates are defined elsewhere and return the appropriate command objects or errors.

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
Summary: Creates a Cobra command named "run" that executes the provided directive when invoked and saves its output. Use this to expose a per-directive CLI run entry point.
Signature: func CreateCustomRunHandler(directive types.Directive) (*cobra.Command, error)
Parameters:
  - directive: types.Directive - the directive to run; its Name is used for logging and description, and Output labels the saved result.
Returns:
  - *cobra.Command: a configured Cobra command with Use="run", Short/Long set to a description derived from directive.
  - error: nil in the current implementation.
Errors/Exceptions:
  - On execution failure, the Run function logs a fatal error via log.Fatalf and terminates the process.
Side Effects:
  - May invoke directive.Execute() within Run, which can perform I/O, network calls, or state mutations; may terminate the process on error.
Edge Cases & Assumptions:
  - The command enforces exactly 0 arguments (cobra.ExactArgs(0)).
  - No validation is performed on directive before wiring it into the command.
  - The constructed command description reflects directive.Name and directive.Output.

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
Summary: Creates a Cobra command that initializes a given directive by loading configuration
from the active scope (global/user/local or a custom config), then applies the directive in the
local project configuration and persists changes.
Use this to seed or constrain a directive within the local project config using existing configs.

Signature: func CreateCustomInitHandler(directive types.Directive) (*cobra.Command, error)

Parameters:
- directive: types.Directive — the directive to initialize in the local scope. Its Name is used to
  describe the command.

Returns:
- *cobra.Command: a configured command with Use "init [files]", Short/Long descriptions, and a Run
  implementation that performs the initialization.
- error: non-nil if the command cannot be constructed.

Errors/Exceptions:
- The Run implementation may call log.Fatalf on failures to load config scope flags or to resolve
  config files, terminating the process.
- If initialization in the local scope fails (InitDirectiveInLocalScope returns an error), Run logs
  and exits with a non-nil error.

Side Effects:
- Creates a command object and, when executed, reads flags, loads config files, and mutates the
  local configuration via InitDirectiveInLocalScope, persisting changes to config.ProjectConfigFile.
- Emits debug/info logs during command creation and execution.

Edge Cases & Assumptions:
- If the directive is already initialized in the initial ProjectConfigFile, InitDirectiveInLocalScope may
  return nil without changes (per its documented behavior).
- Assumes package-level identifiers (config.ProjectConfigFile, etc.) and helper functions are defined
  elsewhere and accessible from this context.

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
Summary: Creates a Cobra command that exports the provided directive to configuration files chosen by scope flags. The command description uses directive.Name and exports across global, user, local, or a custom config as determined at runtime.
Signature: func CreateCustomExportHandler(directive types.Directive) (*cobra.Command, error)
Parameters:
- directive: types.Directive — the directive to export; its name is used for command description.
Returns:
- *cobra.Command: the constructed export command.
- error: returned as nil; the function always returns a command and nil error; Run-time errors are handled by log.Fatalf.
Errors/Exceptions:
- The Run function will call log.Fatalf on failures when loading config scope flags, obtaining config files, or exporting the directive.
Side Effects:
- Creates and initializes a Cobra command.
- Reads and interprets scope flags via helpers.GetConfigScopeFlags(cmd).
- Determines config files via helpers.GetConfigsFromFlags(cmd) and may default to appropriate project config paths.
- Exports the directive to the selected config files via ExportCustomDirective(directive, configFiles), mutating config.Settings and persisting YAML files.
- May modify Cobra flags (e.g., defaulting to the user scope with cmd.Flags().Set("user", "true")).
Edge Cases & Assumptions:
- If no scopes are explicitly selected, the command defaults to exporting to the user scope.
- The helper functions may terminate the process via log.Fatalf on errors; this function does not return errors to its caller.
- Assumes directive.Name is a valid key for the target config maps; behavior for missing keys relies on the underlying ExportCustomDirective implementation.

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
AddFieldUpdates builds a set of cobra.Command entries to update string fields on a given
Directive and persist the changes to configuration files loaded at runtime.

Summary: Creates a dedicated command for each field defined in StringFieldsToUpdate to
update that field on the provided directive and save the changes to all relevant config
files. Use when you want CLI-driven, per-field updates of a Directive that are written back
to configuration files.

Signature: func AddFieldUpdates(directive types.Directive) ([]*cobra.Command, error)

Parameters:
- directive: types.Directive — the target directive whose fields may be updated via the generated commands.
  The directive's name is used for logging and traceability during updates.

Returns:
- []*cobra.Command — a slice of commands, one per field in StringFieldsToUpdate, each configured with
  Use, Short, Long, and an ExactArgs(1) constraint.
- error — currently nil for this function; any run-time errors are surfaced via fatal logs within the
  command execution path (Run) when interacting with config loading or updating.

Errors/Exceptions:
- The function itself does not return non-nil errors. Run-time failures may terminate the process via
  log.Fatalf inside the command's Run handler (e.g., failures in GetConfigsFromFlags or updating the directive).

Side Effects:
- Registers new cobra.Command instances that mutate a Directive and persist changes to configFiles via
  UpdateDirectiveFieldInConfigs. May cause log output and creation/verification of config files when executed.
- May terminate the process if errors occur while resolving configuration files (through log.Fatalf in Run).

Edge Cases & Assumptions:
- Iterates over StringFieldsToUpdate; for each field, a command is created using field.Use, field.Short, and
  field.Long, with Args enforced as cobra.ExactArgs(1).
- The Run closure captures field and directive; it loads config files via helpers.GetConfigsFromFlags(cmd) and
  updates the corresponding field on the directive when invoked.
- If StringFieldsToUpdate is empty, the function returns an empty slice without errors.

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
Summary: Build and return a list of Cobra commands, one per element of ArrayFieldsToUpdate, to update array/slice fields on the provided directive. Each command loads configuration file paths from flags and updates the corresponding directive array across config files (if any), or just in memory when no files are provided.
Signature: func AddArrayUpdates(directive types.Directive) ([]*cobra.Command, error)
Parameters:
- directive: types.Directive — the directive to update; each generated command uses directive in its Run closure.
Returns:
- []*cobra.Command: a list of commands, one per field in ArrayFieldsToUpdate.
- error: nil in the current implementation; errors encountered at runtime are handled via log.Fatalf inside the Run closures.
Errors/Exceptions:
- The function itself does not return a non-nil error. Runtime failures are surfaced via log.Fatalf within the Run callbacks (e.g., failures loading config scopes or updating directive arrays).
Side Effects:
- Creates and returns Cobra commands; each command, when run, may read config scope flags, determine configFiles, and update config.Settings.Directives[directive.Name] via UpdateDirectiveArrayInConfigs. May log debug messages and terminate the process on error.
- May mutate persistent configuration files on disk and the in-memory config.Settings.
Edge Cases & Assumptions:
- If ArrayFieldsToUpdate is empty, returns an empty slice of commands.
- Run closures capture loop variables (field, directive); without per-iteration copies, they may reference the last field value at execution time.
- Assumes ArrayFieldsToUpdate elements expose Use, Short, Long, and Name fields; uses field.Name for the target directive field and for logging.
- Assumes helpers.GetConfigsFromFlags(cmd) yields configFiles or handles failures via a fatal log; if configFiles is empty, updates apply only to the in-memory config.Settings.Directives and do not write to disk.

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
