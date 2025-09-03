package main;

import (
    "os"
    "testing"
    log "github.com/sirupsen/logrus"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/types"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/config"
    // "github.com/BlankCanvasStudio/AutoScribe/tests/helpers"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/cli/directives"
)

func TestInitHandler(t *testing.T) {
    // Create sample config file in local directory
    name := "someDirective"
    prompt_file := "/tmp/random"
    saveToConfig := "/tmp/testing"
    configs := []string{ saveToConfig }


    os.WriteFile(prompt_file, []byte{}, 0644)

    // Write the random directive to that config
    err := directives.CreateNewDirective(name, prompt_file, configs)
    if err != nil {
        log.Fatalf("failed to create directive: %v", err)
    }


    // Load the default scope
    config.PushLoadedConfig()
    err = config.LoadConfig()
    if err != nil {
        log.Fatalf("Failed to load configs: %v", err)
    }

    directive, err := types.NewDirective(name, prompt_file)
    if err != nil {
        log.Fatalf("failed to make new directive %v: %v", name, err)
    }

    // Initialize Directive we just saved into project scope
    err = directives.InitDirectiveInLocalScope([]string{saveToConfig}, *directive, []string{"fileOne", "fileTwo"})
    if err != nil {
        log.Fatalf("Failed to init directive in local scope: %v", err)
    }


    {
        // Load the default scope
        config.PushLoadedConfig()
        err = config.LoadConfig()
        if err != nil {
            log.Fatalf("Failed to load configs: %v", err)
        }

        if _, ok := config.Settings.Directives[name]; ok {
            log.Fatalf("Loaded %v from non-local scope (it shouldn't be saved there)", name)
        }

        // Load config we just made
        err = config.LoadConfigFile(types.ProjectDirectory + "asb.yml")
        if err != nil {
            log.Fatalf("Failed to load configs: %v", err)
        }

        if _, ok := config.Settings.Directives[name]; !ok {
            log.Fatalf("Failed to load %v from local scope (which we previously saved)", name)
        }
    }

    os.Remove(saveToConfig)
    os.Remove(types.ProjectDirectory + "asb.yml")

    config.PopLoadedConfig()
}

/*
func TestExportHandler(t *testing.T) {
    // Create sample config file in local directory
    name := "someDirective"
    prompt_file := "/tmp/random"
    saveToConfig := "/tmp/testing"
    configs := []string{types.ProjectDirectory}


    os.WriteFile(prompt_file, []byte{}, 0644)

    err := directives.CreateNewDirective(name, prompt_file, configs)
    if err != nil {
        log.Fatalf("failed to create directive: %v", err)
    }


    // Load the local config
    config.PushLoadedConfig()
    err = config.LoadConfigFile(configs[0])
    if err != nil {
        log.Fatalf("failed to load %v: %v", prompt_file, err)
    }


    err = directives.CreateCustomInitHandler()


    // Undo local config loading
    config.PopLoadedConfig()

    // leave no trace
    os.Remove(saveToConfig)
    os.Remove("./asb.yml")
}
*/

