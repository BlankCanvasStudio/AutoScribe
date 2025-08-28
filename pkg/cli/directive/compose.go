package directives

import (
    "fmt"
    "github.com/spf13/cobra"

    log "github.com/sirupsen/logrus"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/cli"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/config"
)

var Cmd = &cobra.Command{
  Use:   "directive",
  Short: "Update information about & add custom directives",
  Long:  ``,
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("AutoScribe version 0.1")
  },
}

func init() {
    Cmd.AddCommand(createCmd)
}

var createCmd = &cobra.Command{
    Use:   "create [name] [prompt file]",
    Short: "",
    Long:  ``,
    Args:  cobra.ExactArgs(2),

    Run: func(cmd *cobra.Command, args []string) {
        log.Debugf("Create directive triggered")

        configFiles := cli.GetConfigsFromFlags(cmd)

        log.Infof("Saving to config files: %v", configFiles)

        savedConfig := config.Settings

        directive := args[0]
        prompt := args[1]

        newDirective, err := config.NewDirective(directive, prompt)
        if err != nil {
            log.Fatalf("failed to create new directive %v: %v", directive, err)
        }

        for _, configFile := range configFiles {
            log.Infof("Updating settings in: %v", configFile)
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

