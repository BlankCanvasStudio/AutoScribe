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


func CreateCustomCommands() ([]*cobra.Command, error) {
    log.Debugf("custom directives found: %v", len(config.Settings.Directives))

    var customCmds = make([]*cobra.Command, 0, len(config.Settings.Directives));

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

        // If I ever have to add to this, make it a for loop with function pointers plz
        //      vim makes copy paste too easy and I'm trying to move fast

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


        Kind, err := CreateCustomKind(directive)
        if err != nil {
            return nil, fmt.Errorf("failed to create custom kind handler for %v: %v", name, err)
        }
        createCmd.AddCommand(Kind)


        Docs, err := CreateCustomDocs(directive)
        if err != nil {
            return nil, fmt.Errorf("failed to create custom docs handler for %v: %v", name, err)
        }
        createCmd.AddCommand(Docs)


        Short, err := CreateCustomShort(directive)
        if err != nil {
            return nil, fmt.Errorf("failed to create custom short handler for %v: %v", name, err)
        }
        createCmd.AddCommand(Short)


        Prompt, err := CreateCustomPrompt(directive)
        if err != nil {
            return nil, fmt.Errorf("failed to create custom prompt handler for %v: %v", name, err)
        }
        createCmd.AddCommand(Prompt)


        PromptText, err := CreateCustomPrompt(directive)
        if err != nil {
            return nil, fmt.Errorf("failed to create custom prompt-text handler for %v: %v", name, err)
        }
        createCmd.AddCommand(PromptText)


        Focus, err := CreateCustomFocus(directive)
        if err != nil {
            return nil, fmt.Errorf("failed to create custom focus handler for %v: %v", name, err)
        }
        createCmd.AddCommand(Focus)


        Ignore, err := CreateCustomIgnore(directive)
        if err != nil {
            return nil, fmt.Errorf("failed to create custom ignore handler for %v: %v", name, err)
        }
        createCmd.AddCommand(Ignore)


        Model, err := CreateCustomModel(directive)
        if err != nil {
            return nil, fmt.Errorf("failed to create custom model handler for %v: %v", name, err)
        }
        createCmd.AddCommand(Model)


        Output, err := CreateCustomModel(directive)
        if err != nil {
            return nil, fmt.Errorf("failed to create custom output handler for %v: %v", name, err)
        }
        createCmd.AddCommand(Output)


        ApiKey, err := CreateCustomModel(directive)
        if err != nil {
            return nil, fmt.Errorf("failed to create custom api key handler for %v: %v", name, err)
        }
        createCmd.AddCommand(ApiKey)


        LocalDocs, err := CreateCustomModel(directive)
        if err != nil {
            return nil, fmt.Errorf("failed to create custom local-docs handler for %v: %v", name, err)
        }
        createCmd.AddCommand(LocalDocs)


        Servers, err := CreateCustomModel(directive)
        if err != nil {
            return nil, fmt.Errorf("failed to create custom servers handler for %v: %v", name, err)
        }
        createCmd.AddCommand(Servers)




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






