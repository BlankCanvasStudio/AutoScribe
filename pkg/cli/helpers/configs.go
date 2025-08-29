package helpers

import (
    "fmt"
    "github.com/spf13/cobra"

    log "github.com/sirupsen/logrus"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/config"
)

func GetConfigsFromFlags(cmd *cobra.Command) ([]string, error) {
    configFiles := make([]string, 0, 3)

    globalScope, err := cmd.Flags().GetBool("global")
    if err != nil {
        return configFiles, fmt.Errorf("failed to get global var: %v", err)
    }

    userScope, err := cmd.Flags().GetBool("user")
    if err != nil {
        return configFiles, fmt.Errorf("failed to get global var: %v", err)
    }

    configFile, err := cmd.Flags().GetString("config")
    if err != nil {
        log.Fatalf("failed to get config var: %v", err)
    }

    if configFile != "" {
        configFiles = append(configFiles, configFile)        
    } else {
        log.Infof("config file empty")
    } 

    if globalScope {
        log.Infof("working with config in global scope")

        err := config.VerifyGlobalConfigExists()
        if err != nil {
            return configFiles, fmt.Errorf("failed to create global config file: %v", err)
        }

        configFiles = append(configFiles, config.GlobalConfigFile)
    }

    if userScope {
        log.Infof("working with config in user scope")

        err := config.VerifyUserConfigExists()
        if err != nil {
            return configFiles, fmt.Errorf("failed to create user config file: %v", err)
        }

        configFiles = append(configFiles, config.UserConfigFile)
    }

    if len(configFiles) == 0 {
        configFiles = append(configFiles, config.ProjectConfigFile)
    }

    return configFiles, nil
}

