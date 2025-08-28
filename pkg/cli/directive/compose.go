package directives

import (
    "fmt"
    "github.com/spf13/cobra"

    log "github.com/sirupsen/logrus"

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

        globalScope, err := cmd.Flags().GetBool("global")
        if err != nil {
            log.Fatalf("failed to get global var: %v", err)
        }

        userScope, err := cmd.Flags().GetBool("user")
        if err != nil {
            log.Fatalf("failed to get global var: %v", err)
        }

        configFile, err := cmd.Flags().GetString("config")
        if err != nil {
            log.Fatalf("failed to get config var: %v", err)
        }

        configFiles := []string{}

        if configFile != "" {
            configFiles = append(configFiles, configFile)        
        } else {
            log.Infof("config file empty")
        } 

        if globalScope {
            log.Infof("adjusting global scope")

            err := config.VerifyGlobalConfigExists()
            if err != nil {
                log.Fatalf("failed to create global config file: %v", err)
            }

            configFiles = append(configFiles, config.GlobalConfigFile)
        }

        if userScope {
            log.Infof("adjusting user scope")

            err := config.VerifyUserConfigExists()
            if err != nil {
                log.Fatalf("failed to create user config file: %v", err)
            }

            configFiles = append(configFiles, config.UserConfigFile)
        }

        if len(configFiles) == 0 {
            configFiles = append(configFiles, config.ProjectConfigFile)
        }

        savedConfig := config.Settings

        log.Infof("Saved config files: %v", configFiles)

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

