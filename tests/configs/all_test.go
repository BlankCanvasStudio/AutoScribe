package main;

import (
    "testing"
    log "github.com/sirupsen/logrus"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/config"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/cli/directives"
    "github.com/BlankCanvasStudio/AutoScribe/tests/helpers"
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

    // config.Settings.PrettyPrint()
}


func TestAddDirective(t *testing.T) {
    // output_config := "/tmp/sample.config"    

    cnf := helpers.NewCobraConfig()

    directives.CreateCmd.Run(cnf, []string{"testing", "/tmp/testing"})
}
