package directives

import (
    // "fmt"
    "github.com/spf13/cobra"

    log "github.com/sirupsen/logrus"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/types"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/config"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/cli/helpers"
)

var createCmd = &cobra.Command{
    Use:   "create [name] [prompt file]",
    Short: "",
    Long:  ``,
    Args:  cobra.ExactArgs(2),

    Run: func(cmd *cobra.Command, args []string) {
        log.Debugf("Create directive triggered")

        configFiles, err := helpers.GetConfigsFromFlags(cmd)
        if err != nil {
            log.Fatalf("failed to get the correct config files: %v", err)
        }

        log.Debugf("Saving to config files: %v", configFiles)

        savedConfig := config.Settings

        directive := args[0]
        prompt := args[1]

        newDirective, err := types.NewDirective(directive, prompt)
        if err != nil {
            log.Fatalf("failed to create new directive %v: %v", directive, err)
        }

        for _, configFile := range configFiles {
            log.Debugf("Updating settings in: %v", configFile)
            config.Settings = config.NewConfig()

            config.LoadConfigFile(configFile)

            config.Settings.Directives[directive] = *newDirective

            err := config.SaveConfigFile(configFile, config.Settings)
            if err != nil {
                log.Fatalf("Failed to create new directive: %v", err)
            }
        }

        config.Settings = savedConfig
    },
}

