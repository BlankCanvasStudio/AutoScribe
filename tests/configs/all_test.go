package main;

import (
    "os"
    "testing"
    log "github.com/sirupsen/logrus"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/types"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/config"
)


func TestSanityCheck(t *testing.T) {
    configFile := "./ymls/valid.yml"

    _, err := os.Stat(configFile)
    if os.IsNotExist(err) {
        log.Fatalf("can't load %v; file doesn't exist", configFile)
    }

    err = config.LoadConfigFile(configFile)
    if err != nil {
            log.Fatalf("Failed to load config: %v", err)
    }

    err = config.Settings.SanityCheck()
    if err != nil {
        log.Fatalf("failed to sanity check configs: %v", err)
    }

    if config.Settings.ApiKey == "" {
        log.Fatalf("No api key loaded!")
    }

    directives := []string{"readme", "helpmenu", "helpimplementation", "helpupdate", "docs"}
    for _, name := range directives {
        if _, ok := config.Settings.Directives[name]; !ok {
            log.Infof("%v", config.Settings.Directives)
            log.Fatalf("Directive %v wasn't loaded in config %v", name, configFile)
        }
    }

    // Verify README model loaded correctly
    if d := config.Settings.Directives["readme"]; d.Model != types.GPT_41_Nano {
        log.Fatalf("loaded readme model wasn't %v, its %v", types.GPT_41_Nano, d.Model)
    }

    for _, name := range directives[1:] {
        if d := config.Settings.Directives[name]; d.Model != types.DefaultModel {
            log.Fatalf("loaded %v model wasn't %v, its %v", name, types.DefaultModel, d.Model)
        }
    }

    for _, name := range directives {
        if name == "docs" { continue }

        if d := config.Settings.Directives[name]; d.Kind != types.TextDirective {
            log.Fatalf("loaded %v kind wasn't %v, its %v", name, types.TextDirective, d.Kind)
        }
    }
}

