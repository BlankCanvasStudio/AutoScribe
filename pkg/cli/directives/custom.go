package directives;

import (
    // "os"
    "fmt"
    "strings"
    "github.com/spf13/cobra"

    log "github.com/sirupsen/logrus"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/types"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/config"
    // "github.com/BlankCanvasStudio/AutoScribe/pkg/ai/calls"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/cli/helpers"
    // "github.com/BlankCanvasStudio/AutoScribe/pkg/ai/formatting"
)


func InitDirectiveInLocalScope(directive types.Directive, toFocus []string, configFiles []string) error {
    log.Debugf("Looking for directive: %v", directive.Name)
    log.Debugf("Looking at configs: %v", configFiles)

    config.PushLoadedConfig()

    {

        config.LoadConfigFile(config.ProjectConfigFile)

        _, exists := config.Settings.Directives[directive.Name]
        if exists {
            log.Infof("Directive %v already initialized", directive.Name)
            return nil
        }

    }

    err := config.PopLoadedConfig()
    if err != nil {
        return fmt.Errorf("Failed to PopLoadedConfig: %v", err)
    }


    // Iterate to move "up" in scope
    for i := 0; i < len(configFiles); i++ {
        config.PushLoadedConfig()

        {

            // Load those settings cofigs
            config.LoadConfigFile(configFiles[i])

            // Verify that directive exists
            toLoad, exists := config.Settings.Directives[directive.Name]
            if !exists {
                log.Debugf("Directive %v not defined in %v", directive.Name, configFiles[i])
                err := config.PopLoadedConfig()
                if err != nil {
                    return fmt.Errorf("Failed to PopLoadedConfig: %v", err)
                }

                continue
            }

            config.PushLoadedConfig()

            {

                // Load the config to append to
                config.LoadConfigFile(config.ProjectConfigFile)

                // load things to focus for ease of use
                for _, arg := range toFocus {
                    toLoad.Focus = append(toLoad.Focus, arg)
                }

                config.Settings.Directives[directive.Name] = toLoad

                err = config.SaveConfigFile(config.ProjectConfigFile, config.Settings)
                if err != nil {
                    return fmt.Errorf("Failed to save to config %v: %v", configFiles[i], err)
                }

            }

            err := config.PopLoadedConfig()
            if err != nil {
                return fmt.Errorf("Failed to PopLoadedConfig: %v", err)
            }

        }

        err := config.PopLoadedConfig()
        if err != nil {
            return fmt.Errorf("Failed to PopLoadedConfig: %v", err)
        }

        log.Infof("Initialized %v from %v", directive.Name, configFiles[i])

        return nil
    }

    return fmt.Errorf("Couldn't load directive %v not found in configs: %v", directive, configFiles)
}


func ExportCustomDirective(directive types.Directive, configFiles []string) error {

    toAdd := config.Settings.Directives[directive.Name]

    // Idk if keeping the focus & ignore is the play
    toAdd.Focus = nil
    toAdd.Ignore = nil
    toAdd.Model = types.NoModel
    toAdd.ApiKey = ""
    toAdd.Scope = ""

    // Iterate to move "up" in scope
    for _, configFile := range configFiles {
        config.PushLoadedConfig()
        {

            config.LoadConfigFile(configFile)

            config.Settings.Directives[toAdd.Name] = toAdd

            err := config.SaveConfigFile(configFile, config.Settings)
            if err != nil {
                return fmt.Errorf("Failed to create new directive: %v", err)
            }
        }
    }

    return nil
}

// We can honestly also probably do some fuck shit with introspection & spread operator 
//      to write this automatically. Too lazy. Copy, paste, find, and replace too EZ
func CreateCustomKind(directive types.Directive) (*cobra.Command, error) {

    return &cobra.Command{
        Use: "kind <text argument>",
        Short: "set the directive kind (file base or recursion base)",
        Long:  "set the directive kind (file base or recursion base)",
        Args:  cobra.ExactArgs(1),

        Run: func(cmd *cobra.Command, args []string) {
            log.Debugf("%v kind triggered", directive)

            configFiles, err := helpers.GetConfigsFromFlags(cmd)
            if err != nil {
                log.Fatalf("failed to get the correct config files: %v", err)
            }

            log.Debugf("Saving to config files: %v", configFiles)

            directive.Kind = types.DirectiveType(strings.ToLower(args[0]))

            err = directive.SanityCheck()
            if err != nil {
                log.Fatalf("can't update directive: %v", err)
            }

            config.Settings.Directives[directive.Name] = directive


            for _, configFile := range configFiles {
                log.Debugf("Updating settings in: %v", configFile)
                config.PushLoadedConfig()
                {

                    config.LoadConfigFile(configFile)

                    d := config.Settings.Directives[directive.Name]

                    d.Kind = types.DirectiveType(strings.ToLower(args[0]))

                    config.Settings.Directives[directive.Name] = d

                    err := config.SaveConfigFile(configFile, config.Settings)
                    if err != nil {
                        log.Fatalf("Failed to create new directive: %v", err)
                    }
                }
                config.PopLoadedConfig()
            }
        },

    }, nil
}


func CreateCustomDocs(directive types.Directive) (*cobra.Command, error) {

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

            log.Debugf("Saving to config files: %v", configFiles)

            directive.Description = args[0]

            err = directive.SanityCheck()
            if err != nil {
                log.Fatalf("can't update directive: %v", err)
            }

            config.Settings.Directives[directive.Name] = directive

            savedConfig := config.Settings

            for _, configFile := range configFiles {
                log.Debugf("Updating settings in: %v", configFile)
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


func CreateCustomShort(directive types.Directive) (*cobra.Command, error) {

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

            log.Debugf("Saving to config files: %v", configFiles)

            directive.Short = args[0]

            err = directive.SanityCheck()
            if err != nil {
                log.Fatalf("can't update directive: %v", err)
            }

            config.Settings.Directives[directive.Name] = directive

            savedConfig := config.Settings

            for _, configFile := range configFiles {
                log.Debugf("Updating settings in: %v", configFile)
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


func CreateCustomPrompt(directive types.Directive) (*cobra.Command, error) {

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

            log.Debugf("Saving to config files: %v", configFiles)

            directive.Prompt = args[0]

            err = directive.SanityCheck()
            if err != nil {
                log.Fatalf("can't update directive: %v", err)
            }

            config.Settings.Directives[directive.Name] = directive

            savedConfig := config.Settings

            for _, configFile := range configFiles {
                log.Debugf("Updating settings in: %v", configFile)
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


func CreateCustomPromptText(directive types.Directive) (*cobra.Command, error) {

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

            log.Debugf("Saving to config files: %v", configFiles)

            directive.PromptText = args[0]

            err = directive.SanityCheck()
            if err != nil {
                log.Fatalf("can't update directive: %v", err)
            }

            config.Settings.Directives[directive.Name] = directive

            savedConfig := config.Settings

            for _, configFile := range configFiles {
                log.Debugf("Updating settings in: %v", configFile)
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


func CreateCustomFocus(directive types.Directive) (*cobra.Command, error) {

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

            log.Debugf("Saving to config files: %v", configFiles)

            for _, file := range args {
                directive.Focus = append(directive.Focus, file)
            }

            err = directive.SanityCheck()
            if err != nil {
                log.Fatalf("can't update directive: %v", err)
            }

            config.Settings.Directives[directive.Name] = directive

            savedConfig := config.Settings

            for _, configFile := range configFiles {
                log.Debugf("Updating settings in: %v", configFile)
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


func CreateCustomIgnore(directive types.Directive) (*cobra.Command, error) {

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

            log.Debugf("Saving to config files: %v", configFiles)

            for _, file := range args {
                directive.Ignore = append(directive.Ignore, file)
            }

            err = directive.SanityCheck()
            if err != nil {
                log.Fatalf("can't update directive: %v", err)
            }

            config.Settings.Directives[directive.Name] = directive

            savedConfig := config.Settings

            for _, configFile := range configFiles {
                log.Debugf("Updating settings in: %v", configFile)
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


func CreateCustomModel(directive types.Directive) (*cobra.Command, error) {

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

            log.Debugf("Saving to config files: %v", configFiles)

            directive.Model = types.Model(strings.ToLower(args[0]))

            err = directive.SanityCheck()
            if err != nil {
                log.Fatalf("can't update directive: %v", err)
            }

            config.Settings.Directives[directive.Name] = directive

            savedConfig := config.Settings

            for _, configFile := range configFiles {
                log.Debugf("Updating settings in: %v", configFile)
                config.Settings = config.NewConfig()

                config.LoadConfigFile(configFile)

                d := config.Settings.Directives[directive.Name]

                d.Model = types.Model(strings.ToLower(args[0]))

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

func CreateCustomOutput(directive types.Directive) (*cobra.Command, error) {

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

            log.Debugf("Saving to config files: %v", configFiles)

            directive.Output = args[0]

            err = directive.SanityCheck()
            if err != nil {
                log.Fatalf("can't update directive: %v", err)
            }

            config.Settings.Directives[directive.Name] = directive

            savedConfig := config.Settings

            for _, configFile := range configFiles {
                log.Debugf("Updating settings in: %v", configFile)
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

func CreateCustomApiKey(directive types.Directive) (*cobra.Command, error) {

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

            log.Debugf("Saving to config files: %v", configFiles)

            directive.ApiKey = args[0]

            err = directive.SanityCheck()
            if err != nil {
                log.Fatalf("can't update directive: %v", err)
            }

            config.Settings.Directives[directive.Name] = directive

            savedConfig := config.Settings

            for _, configFile := range configFiles {
                log.Debugf("Updating settings in: %v", configFile)
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

func CreateCustomLocalDocs(directive types.Directive) (*cobra.Command, error) {

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

            log.Debugf("Saving to config files: %v", configFiles)

            directive.LocalDocs = args[0]

            err = directive.SanityCheck()
            if err != nil {
                log.Fatalf("can't update directive: %v", err)
            }

            config.Settings.Directives[directive.Name] = directive

            savedConfig := config.Settings

            for _, configFile := range configFiles {
                log.Debugf("Updating settings in: %v", configFile)
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


func CreateCustomServers(directive types.Directive) (*cobra.Command, error) {

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

            log.Debugf("Saving to config files: %v", configFiles)

            for _, file := range args {
                directive.Servers = append(directive.Focus, file)
            }

            err = directive.SanityCheck()
            if err != nil {
                log.Fatalf("can't update directive: %v", err)
            }

            config.Settings.Directives[directive.Name] = directive

            savedConfig := config.Settings

            for _, configFile := range configFiles {
                log.Debugf("Updating settings in: %v", configFile)
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

