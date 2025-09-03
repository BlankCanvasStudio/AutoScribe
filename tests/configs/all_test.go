package main;

import (
    "testing"
    log "github.com/sirupsen/logrus"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/config"
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

    config.Settings.PrettyPrint()
}


func TestAddDirective(t *testing.T) {
    output_config := "/tmp/sample.config"    

    directives.CreateCmd.Run(CreateCmd, []string{"testing", "/tmp/testing")
}
