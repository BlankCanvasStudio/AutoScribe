package directives

import (
    "fmt"
    "github.com/spf13/cobra"

    log "github.com/sirupsen/logrus"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/types"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/config"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/cli/helpers"
)

var Cmd = &cobra.Command{
  Use:   "directive",
  Short: "Update information about & add custom directives",
  Long:  ``,
  Run: func(cmd *cobra.Command, args []string) {},
}


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


func CreateCustomCommands() ([]*cobra.Command, error) {
    log.Debugf("custom directives found: %v", len(config.Settings.Directives))

    var customCmds = make([]*cobra.Command, 0, len(config.Settings.Directives));

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

        customCmds = append(customCmds,  createCmd)

    }

    return customCmds, nil
}


func CreateCustomRunHandler(directive types.Directive) (*cobra.Command, error) {
    log.Debugf("Executing %v custom run handler", directive.Name)
    desc := fmt.Sprintf("Run the %v directive and save to `%v`", directive.Name, directive.Output)

    return &cobra.Command{
        Use: "run",
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


func CreateCustomInitHandler(directive types.Directive) (*cobra.Command, error) {
    log.Debugf("Executing %v custom init handler", directive.Name)

    desc := fmt.Sprintf("Initialize the %v directive saved in global configs into this project", directive.Name)

    return &cobra.Command{
        Use: "init [files]",
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


func CreateCustomExportHandler(directive types.Directive) (*cobra.Command, error) {
    log.Debugf("Executing %v custom init handler", directive.Name)

    desc := fmt.Sprintf("Export the %v directive to various configs", directive.Name)

    return &cobra.Command{
        Use: "export [files]",
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
        Name: "Kind",
        Use: "kind <text argument>",
        Short: "set the directive kind (file base or recursion base)",
        Long:  "set the directive kind (file base or recursion base)",
    },

    Field{
        Name: "Description",
        Use: "description <text argument>",
        Short: "set the full description of a custom directive",
        Long:  "set the full description of a custom directive",
    },

    Field{
        Name: "Short",
        Use: "short <text argument>",
        Short: "set the short description of a custom directive",
        Long:  "set the short description of a custom directive",
    },

    Field{
        Name: "Prompt",
        Use: "prompt <text argument>",
        Short: "set the prompt file for a directive",
        Long:  "set the prompt file for a directive",
    },

    Field{
        Name: "PromptText",
        Use: "prompt-text <text argument>",
        Short: "set the prompt text for a directive (override the prompt file)",
        Long:  "set the prompt text for a directive (override the prompt file)",
    },

    Field{
        Name: "PromptText",
        Use: "prompt-text <text argument>",
        Short: "set the prompt text for a directive (override the prompt file)",
        Long:  "set the prompt text for a directive (override the prompt file)",
    },

    Field{
        Name: "Model",
        Use: "model <text argument>",
        Short: "set the model for a particular directive",
        Long:  "set the model for a particular directive",
    },

    Field{
        Name: "Output",
        Use: "output <text argument>",
        Short: "set the output file for a directive's result",
        Long:  "set the output file for a directive's result",
    },

    Field{
        Name: "ApiKey",
        Use: "apikey <text argument>",
        Short: "set a directive's api key",
        Long:  "set a directive's api key",
    },

    Field{
        Name: "LocalDocs",
        Use: "local-docs <path>",
        Short: "set where a directive can locally source documentation not written in files",
        Long:  "set where a directive can locally source documentation not written in files",
    },
}


func AddFieldUpdates(directive types.Directive) ([]*cobra.Command, error){
    ret := make([]*cobra.Command, 0, len(StringFieldsToUpdate))

    for _, field := range StringFieldsToUpdate {
        ret = append(ret, &cobra.Command{
            Use: field.Use,
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
        Name: "Focus",
        Use: "add <text arguments>",
        Short: "add functions or files for autoscibe to add documentation to / focus on",
        Long:  "add functions or files for autoscibe to add documentation to / focus on",
    },

    Field{
        Name: "Ignore",
        Use: "ignore <text arguments>",
        Short: "ignore functions or files",
        Long:  "ignore functions or files",
    },

    Field{
        Name: "Servers",
        Use: "server <text arguments>",
        Short: "set where a directive sources documentation from if it can't be found locally (godocs is implicit)",
        Long:  "set where a directive sources documentation from if it can't be found locally (godocs is implicit)",
    },
}


func AddArrayUpdates(directive types.Directive) ([]*cobra.Command, error) {
    ret := make([]*cobra.Command, 0, len(StringFieldsToUpdate))

    for _, field := range ArrayFieldsToUpdate {
        ret = append(ret,  &cobra.Command{
            Use: field.Use,
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

