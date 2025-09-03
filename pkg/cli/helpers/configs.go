package helpers

import (
    "fmt"
    "github.com/spf13/cobra"

    log "github.com/sirupsen/logrus"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/config"
)

func GetConfigsFromFlags(cmd *cobra.Command) ([]string, error) {
    configFiles := make([]string, 0, 3)

    globalScope, userScope, localScope, customScope, err := GetConfigScopeFlags(cmd)
    if err != nil {
        return configFiles, fmt.Errorf("failed to load config scope flags: %v", err)
    }

    if customScope {

        configFile, err := cmd.Flags().GetString("config")
        if err != nil {
            log.Fatalf("failed to get config var: %v", err)
        }

        log.Debugf("working with custom scope %v", configFile)

        if configFile == "" {
            return nil, fmt.Errorf("config file empty")
        }

        configFiles = append(configFiles, configFile)        
    }

    if localScope {
        log.Debugf("working with config in local scope")

        err := config.VerifyLocalConfigExists()
        if err != nil {
            return configFiles, fmt.Errorf("failed to create user config file: %v", err)
        }

        configFiles = append(configFiles, config.ProjectConfigFile)
    }

    if userScope {
        log.Debugf("working with config in user scope")

        err := config.VerifyUserConfigExists()
        if err != nil {
            return configFiles, fmt.Errorf("failed to create user config file: %v", err)
        }

        configFiles = append(configFiles, config.UserConfigFile)
    }

    if globalScope {
        log.Debugf("working with config in global scope")

        err := config.VerifyGlobalConfigExists()
        if err != nil {
            return configFiles, fmt.Errorf("failed to create global config file: %v", err)
        }

        configFiles = append(configFiles, config.GlobalConfigFile)
    }

    if len(configFiles) == 0 {
        configFiles = append(configFiles, config.ProjectConfigFile)
    }

    return configFiles, nil
}


func GetConfigScopeFlags(cmd *cobra.Command) (bool, bool, bool, bool, error) {
    var err error

    global := false
    user := false
    local := false
    custom := false

    global, err = cmd.Flags().GetBool("global")
    if err != nil {
        return global, user, local, custom, fmt.Errorf("failed to get global var global: %v", err)
    }

    user, err = cmd.Flags().GetBool("user")
    if err != nil {
        return global, user, local, custom, fmt.Errorf("failed to get global var user: %v", err)
    }

    local, err = cmd.Flags().GetBool("local")
    if err != nil {
        return global, user, local, custom, fmt.Errorf("failed to get global var local: %v", err)
    }

    customFile, err := cmd.Flags().GetString("config")
    if err != nil {
        return global, user, local, custom, fmt.Errorf("failed to get global var local: %v", err)
    }

    custom = customFile != ""

    return global, user, local, custom, nil
}

