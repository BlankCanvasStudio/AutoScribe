package main;

import (
    "testing"
    log "github.com/sirupsen/logrus"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/config"
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

    config.Settings.PrettyPrint()
}

