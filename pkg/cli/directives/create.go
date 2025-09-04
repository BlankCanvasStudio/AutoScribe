package directives

import (
    "fmt"

    log "github.com/sirupsen/logrus"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/types"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/config"
)

func CreateNewDirective(name string, prompt string, configFiles []string) error {
    log.Debugf("Create directive triggered")

    log.Debugf("Saving to config files: %v", configFiles)

    newDirective, err := types.NewDirective(name, prompt)
    if err != nil {
        return fmt.Errorf("failed to create new directive %v: %v", name, err)
    }

    for _, configFile := range configFiles {
        config.PushLoadedConfig()

        log.Debugf("Updating settings in: %v", configFile)

        config.Settings = config.NewConfig()

        config.LoadConfigFile(configFile)

        config.Settings.Directives[name] = *newDirective

        err := config.SaveConfigFile(configFile, config.Settings)
        if err != nil {
            return fmt.Errorf("failed to save new directive: %v", err)
        }

        config.PopLoadedConfig()
    }

    return nil
}

