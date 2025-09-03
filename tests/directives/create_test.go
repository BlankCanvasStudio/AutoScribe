package main;

import (
    "os"
    "testing"
    log "github.com/sirupsen/logrus"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/config"
    // "github.com/BlankCanvasStudio/AutoScribe/tests/helpers"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/cli/directives"
)

func TestSanityCheck(t *testing.T) {
    err := config.LoadConfigFile("./tests/configs/ymls/valid.yml")
    if err != nil {
            log.Fatalf("Failed to load config: %v", err)
    }

    err = config.Settings.SanityCheck()
    if err != nil {
        log.Fatalf("failed to sanity check configs: %v", err)
    }
}


func TestAddDirective(t *testing.T) {
    name := "someDirective"
    prompt_file := "/tmp/random"
    configs := []string{"/tmp/testing"}

    os.WriteFile(prompt_file, []byte{}, 0644)

    err := directives.CreateNewDirective(name, prompt_file, configs)
    if err != nil {
        log.Fatalf("failed to create directive: %v", err)
    }

    config.PushLoadedConfig()

    err = config.LoadConfigFile(configs[0])
    if err != nil {
        log.Fatalf("failed to load %v: %v", prompt_file, err)
    }

    if _, ok := config.Settings.Directives[name]; !ok {
        log.Fatalf("failed to load directive %v we saved to %v", name, prompt_file)
    }

    config.PopLoadedConfig()

    // leave no trace
    os.Remove(prompt_file)
}

func TestAddDirectiveFailures(t *testing.T) {
    configs := []string{"/tmp/testing"}

    prompt_file := "/tmp/random"

    os.Remove(prompt_file)

    err := directives.CreateNewDirective("someDirective", prompt_file, configs)
    if err == nil {
        log.Fatalf("created directive without valid prompt file")
    }
}

