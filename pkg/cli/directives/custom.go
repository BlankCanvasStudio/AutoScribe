package directives;

import (
    "fmt"
    "strings"
    "github.com/spf13/cobra"

    log "github.com/sirupsen/logrus"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/ai"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/config"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/cli/helpers"
)



func CreateCustomCommands() ([]*cobra.Command, error) {
    log.Infof("custom directives found: %v", len(config.Settings.Directives))

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

        log.Infof("Creating custom handler for %v", name)

        // If I ever have to add to this, make it a for loop with function pointers plz
        //      vim makes copy paste too easy and I'm trying to move fast

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


// We can honestly also probably do some fuck shit with introspection & spread operator 
//      to write this automatically. Too lazy. Copy, paste, find, and replace too EZ
func CreateCustomKind(directive config.Directive) (*cobra.Command, error) {

    return &cobra.Command{
        Use: "kind <text argument>",
        Short: "set the directive kind (file base or recursion base)",
        Long:  "set the directive kind (file base or recursion base)",
        Args:  cobra.ExactArgs(1),

        Run: func(cmd *cobra.Command, args []string) {
            log.Debugf("%v docs triggered", directive)

            configFiles, err := helpers.GetConfigsFromFlags(cmd)
            if err != nil {
                log.Fatalf("failed to get the correct config files: %v", err)
            }

            log.Infof("Saving to config files: %v", configFiles)

            directive.Kind = config.DirectiveType(strings.ToLower(args[0]))

            config.Settings.Directives[directive.Name] = directive

            savedConfig := config.Settings

            for _, configFile := range configFiles {
                log.Infof("Updating settings in: %v", configFile)
                config.Settings = config.NewConfig()

                config.LoadConfigFile(configFile)

                d := config.Settings.Directives[directive.Name]

                d.Kind = config.DirectiveType(strings.ToLower(args[0]))

                config.Settings.Directives[directive.Name] = d

                err := config.SaveConfigFile(configFile, config.Settings)
                if err != nil {
                    log.Fatalf("Failed to create new directive: %v", err)
                }
            }

            config.Settings = savedConfig
        },

    }, nil
}


func CreateCustomDocs(directive config.Directive) (*cobra.Command, error) {

    return &cobra.Command{
        Use: "description <text argument>",
        Short: "set the full description of a custom directive",
        Long:  "set the full description of a custom directive",
        Args:  cobra.ExactArgs(1),

        Run: func(cmd *cobra.Command, args []string) {
            log.Debugf("%v docs triggered", directive)

            configFiles, err := helpers.GetConfigsFromFlags(cmd)
            if err != nil {
                log.Fatalf("failed to get the correct config files: %v", err)
            }

            log.Infof("Saving to config files: %v", configFiles)

            directive.Description = args[0]

            config.Settings.Directives[directive.Name] = directive

            savedConfig := config.Settings

            for _, configFile := range configFiles {
                log.Infof("Updating settings in: %v", configFile)
                config.Settings = config.NewConfig()

                config.LoadConfigFile(configFile)

                d := config.Settings.Directives[directive.Name]

                d.Description = args[0]

                config.Settings.Directives[directive.Name] = d

                err := config.SaveConfigFile(configFile, config.Settings)
                if err != nil {
                    log.Fatalf("Failed to create new directive: %v", err)
                }
            }

            config.Settings = savedConfig
        },

    }, nil
}


func CreateCustomShort(directive config.Directive) (*cobra.Command, error) {

    return &cobra.Command{
        Use: "description <text argument>",
        Short: "set the full description of a custom directive",
        Long:  "set the full description of a custom directive",
        Args:  cobra.ExactArgs(1),

        Run: func(cmd *cobra.Command, args []string) {
            log.Debugf("%v docs triggered", directive)

            configFiles, err := helpers.GetConfigsFromFlags(cmd)
            if err != nil {
                log.Fatalf("failed to get the correct config files: %v", err)
            }

            log.Infof("Saving to config files: %v", configFiles)

            directive.Short = args[0]

            config.Settings.Directives[directive.Name] = directive

            savedConfig := config.Settings

            for _, configFile := range configFiles {
                log.Infof("Updating settings in: %v", configFile)
                config.Settings = config.NewConfig()

                config.LoadConfigFile(configFile)

                d := config.Settings.Directives[directive.Name]

                d.Short = args[0]

                config.Settings.Directives[directive.Name] = d

                err := config.SaveConfigFile(configFile, config.Settings)
                if err != nil {
                    log.Fatalf("Failed to create new directive: %v", err)
                }
            }

            config.Settings = savedConfig
        },

    }, nil
}


func CreateCustomPrompt(directive config.Directive) (*cobra.Command, error) {

    return &cobra.Command{
        Use: "prompt <file path>",
        Short: "set the prompt file for a directive",
        Long:  "set the prompt file for a directive",
        Args:  cobra.ExactArgs(1),

        Run: func(cmd *cobra.Command, args []string) {
            log.Debugf("%v docs triggered", directive)

            configFiles, err := helpers.GetConfigsFromFlags(cmd)
            if err != nil {
                log.Fatalf("failed to get the correct config files: %v", err)
            }

            log.Infof("Saving to config files: %v", configFiles)

            directive.Prompt = args[0]

            config.Settings.Directives[directive.Name] = directive

            savedConfig := config.Settings

            for _, configFile := range configFiles {
                log.Infof("Updating settings in: %v", configFile)
                config.Settings = config.NewConfig()

                config.LoadConfigFile(configFile)

                d := config.Settings.Directives[directive.Name]

                d.Prompt = args[0]

                config.Settings.Directives[directive.Name] = d

                err := config.SaveConfigFile(configFile, config.Settings)
                if err != nil {
                    log.Fatalf("Failed to create new directive: %v", err)
                }
            }

            config.Settings = savedConfig
        },

    }, nil
}


func CreateCustomPromptText(directive config.Directive) (*cobra.Command, error) {

    return &cobra.Command{
        Use: "prompt-text <text argument>",
        Short: "set the prompt text for a directive (override the prompt file)",
        Long:  "set the prompt text for a directive (override the prompt file)",
        Args:  cobra.ExactArgs(1),

        Run: func(cmd *cobra.Command, args []string) {
            log.Debugf("%v docs triggered", directive)

            configFiles, err := helpers.GetConfigsFromFlags(cmd)
            if err != nil {
                log.Fatalf("failed to get the correct config files: %v", err)
            }

            log.Infof("Saving to config files: %v", configFiles)

            directive.PromptText = args[0]

            config.Settings.Directives[directive.Name] = directive

            savedConfig := config.Settings

            for _, configFile := range configFiles {
                log.Infof("Updating settings in: %v", configFile)
                config.Settings = config.NewConfig()

                config.LoadConfigFile(configFile)

                d := config.Settings.Directives[directive.Name]

                d.PromptText = args[0]

                config.Settings.Directives[directive.Name] = d

                err := config.SaveConfigFile(configFile, config.Settings)
                if err != nil {
                    log.Fatalf("Failed to create new directive: %v", err)
                }
            }

            config.Settings = savedConfig
        },

    }, nil
}


func CreateCustomFocus(directive config.Directive) (*cobra.Command, error) {

    return &cobra.Command{
        Use: "add <text arguments>",
        Short: "add functions or files for autoscibe to add documentation to / focus on",
        Long:  "add functions or files for autoscibe to add documentation to / focus on",

        Run: func(cmd *cobra.Command, args []string) {
            log.Debugf("%v files triggered", directive)

            configFiles, err := helpers.GetConfigsFromFlags(cmd)
            if err != nil {
                log.Fatalf("failed to get the correct config files: %v", err)
            }

            log.Infof("Saving to config files: %v", configFiles)

            for _, file := range args {
                directive.Focus = append(directive.Focus, file)
            }

            config.Settings.Directives[directive.Name] = directive

            savedConfig := config.Settings

            for _, configFile := range configFiles {
                log.Infof("Updating settings in: %v", configFile)
                config.Settings = config.NewConfig()

                config.LoadConfigFile(configFile)

                d := config.Settings.Directives[directive.Name]

                for _, file := range args {
                    directive.Focus = append(directive.Focus, file)
                }

                config.Settings.Directives[directive.Name] = d

                err := config.SaveConfigFile(configFile, config.Settings)
                if err != nil {
                    log.Fatalf("Failed to create new directive: %v", err)
                }
            }

            config.Settings = savedConfig
        },

    }, nil
}


func CreateCustomIgnore(directive config.Directive) (*cobra.Command, error) {

    return &cobra.Command{
        Use: "ignore <text arguments>",
        Short: "ignore functions or files",
        Long:  "ignore functions or files",

        Run: func(cmd *cobra.Command, args []string) {
            log.Debugf("%v files triggered", directive)

            configFiles, err := helpers.GetConfigsFromFlags(cmd)
            if err != nil {
                log.Fatalf("failed to get the correct config files: %v", err)
            }

            log.Infof("Saving to config files: %v", configFiles)

            for _, file := range args {
                directive.Ignore = append(directive.Ignore, file)
            }

            config.Settings.Directives[directive.Name] = directive

            savedConfig := config.Settings

            for _, configFile := range configFiles {
                log.Infof("Updating settings in: %v", configFile)
                config.Settings = config.NewConfig()

                config.LoadConfigFile(configFile)

                d := config.Settings.Directives[directive.Name]

                for _, file := range args {
                    directive.Ignore = append(directive.Ignore, file)
                }

                config.Settings.Directives[directive.Name] = d

                err := config.SaveConfigFile(configFile, config.Settings)
                if err != nil {
                    log.Fatalf("Failed to create new directive: %v", err)
                }
            }

            config.Settings = savedConfig
        },

    }, nil
}


func CreateCustomModel(directive config.Directive) (*cobra.Command, error) {

    return &cobra.Command{
        Use: "model <text argument>",
        Short: "set the model for a particular directive ",
        Long:  "set the model for a particular directive ",
        Args:  cobra.ExactArgs(1),

        Run: func(cmd *cobra.Command, args []string) {
            log.Debugf("%v docs triggered", directive)

            configFiles, err := helpers.GetConfigsFromFlags(cmd)
            if err != nil {
                log.Fatalf("failed to get the correct config files: %v", err)
            }

            log.Infof("Saving to config files: %v", configFiles)

            directive.Model = ai.Model(strings.ToLower(args[0]))

            config.Settings.Directives[directive.Name] = directive

            savedConfig := config.Settings

            for _, configFile := range configFiles {
                log.Infof("Updating settings in: %v", configFile)
                config.Settings = config.NewConfig()

                config.LoadConfigFile(configFile)

                d := config.Settings.Directives[directive.Name]

                d.Model = ai.Model(strings.ToLower(args[0]))

                config.Settings.Directives[directive.Name] = d

                err := config.SaveConfigFile(configFile, config.Settings)
                if err != nil {
                    log.Fatalf("Failed to create new directive: %v", err)
                }
            }

            config.Settings = savedConfig
        },

    }, nil
}

func CreateCustomOutput(directive config.Directive) (*cobra.Command, error) {

    return &cobra.Command{
        Use: "output <text argument>",
        Short: "set the output file for a directive's result",
        Long:  "set the output file for a directive's result",
        Args:  cobra.ExactArgs(1),

        Run: func(cmd *cobra.Command, args []string) {
            log.Debugf("%v docs triggered", directive)

            configFiles, err := helpers.GetConfigsFromFlags(cmd)
            if err != nil {
                log.Fatalf("failed to get the correct config files: %v", err)
            }

            log.Infof("Saving to config files: %v", configFiles)

            directive.Output = args[0]

            config.Settings.Directives[directive.Name] = directive

            savedConfig := config.Settings

            for _, configFile := range configFiles {
                log.Infof("Updating settings in: %v", configFile)
                config.Settings = config.NewConfig()

                config.LoadConfigFile(configFile)

                d := config.Settings.Directives[directive.Name]

                d.Output = args[0]

                config.Settings.Directives[directive.Name] = d

                err := config.SaveConfigFile(configFile, config.Settings)
                if err != nil {
                    log.Fatalf("Failed to create new directive: %v", err)
                }
            }

            config.Settings = savedConfig
        },

    }, nil
}

func CreateCustomApiKey(directive config.Directive) (*cobra.Command, error) {

    return &cobra.Command{
        Use: "apikey <text argument>",
        Short: "set a directive's api key",
        Long:  "set a directive's api key",
        Args:  cobra.ExactArgs(1),

        Run: func(cmd *cobra.Command, args []string) {
            log.Debugf("%v docs triggered", directive)

            configFiles, err := helpers.GetConfigsFromFlags(cmd)
            if err != nil {
                log.Fatalf("failed to get the correct config files: %v", err)
            }

            log.Infof("Saving to config files: %v", configFiles)

            directive.ApiKey = args[0]

            config.Settings.Directives[directive.Name] = directive

            savedConfig := config.Settings

            for _, configFile := range configFiles {
                log.Infof("Updating settings in: %v", configFile)
                config.Settings = config.NewConfig()

                config.LoadConfigFile(configFile)

                d := config.Settings.Directives[directive.Name]

                d.ApiKey = args[0]

                config.Settings.Directives[directive.Name] = d

                err := config.SaveConfigFile(configFile, config.Settings)
                if err != nil {
                    log.Fatalf("Failed to create new directive: %v", err)
                }
            }

            config.Settings = savedConfig
        },

    }, nil
}

func CreateCustomLocalDocs(directive config.Directive) (*cobra.Command, error) {

    return &cobra.Command{
        Use: "local-docs <path>",
        Short: "set where a directive can locally source documentation not written in files",
        Long:  "set where a directive can locally source documentation not written in files",
        Args:  cobra.ExactArgs(1),

        Run: func(cmd *cobra.Command, args []string) {
            log.Debugf("%v docs triggered", directive)

            configFiles, err := helpers.GetConfigsFromFlags(cmd)
            if err != nil {
                log.Fatalf("failed to get the correct config files: %v", err)
            }

            log.Infof("Saving to config files: %v", configFiles)

            directive.LocalDocs = args[0]

            config.Settings.Directives[directive.Name] = directive

            savedConfig := config.Settings

            for _, configFile := range configFiles {
                log.Infof("Updating settings in: %v", configFile)
                config.Settings = config.NewConfig()

                config.LoadConfigFile(configFile)

                d := config.Settings.Directives[directive.Name]

                d.LocalDocs = args[0]

                config.Settings.Directives[directive.Name] = d

                err := config.SaveConfigFile(configFile, config.Settings)
                if err != nil {
                    log.Fatalf("Failed to create new directive: %v", err)
                }
            }

            config.Settings = savedConfig
        },

    }, nil
}


func CreateCustomServers(directive config.Directive) (*cobra.Command, error) {

    return &cobra.Command{
        Use: "add <text arguments>",
        Short: "set where a directive sources documentation from if it can't be found locally (godocs is implicit)",
        Long:  "set where a directive sources documentation from if it can't be found locally (godocs is implicit)",

        Run: func(cmd *cobra.Command, args []string) {
            log.Debugf("%v files triggered", directive)

            configFiles, err := helpers.GetConfigsFromFlags(cmd)
            if err != nil {
                log.Fatalf("failed to get the correct config files: %v", err)
            }

            log.Infof("Saving to config files: %v", configFiles)

            for _, file := range args {
                directive.Servers = append(directive.Focus, file)
            }

            config.Settings.Directives[directive.Name] = directive

            savedConfig := config.Settings

            for _, configFile := range configFiles {
                log.Infof("Updating settings in: %v", configFile)
                config.Settings = config.NewConfig()

                config.LoadConfigFile(configFile)

                d := config.Settings.Directives[directive.Name]

                for _, file := range args {
                    directive.Servers = append(directive.Focus, file)
                }

                config.Settings.Directives[directive.Name] = d

                err := config.SaveConfigFile(configFile, config.Settings)
                if err != nil {
                    log.Fatalf("Failed to create new directive: %v", err)
                }
            }

            config.Settings = savedConfig
        },

    }, nil
}

