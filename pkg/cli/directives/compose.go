package directives

import (
    // "fmt"
    "github.com/spf13/cobra"

    log "github.com/sirupsen/logrus"

    // "github.com/BlankCanvasStudio/AutoScribe/pkg/config"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/cli/helpers"
)

var Cmd = &cobra.Command{
  Use:   "directive",
  Short: "Update information about & add custom directives",
  Long:  ``,
  Run: func(cmd *cobra.Command, args []string) {},
}


func init() {
    Cmd.AddCommand(CreateCmd)
}


var CreateCmd = &cobra.Command{
    Use:   "create [name] [prompt file]",
    Short: "",
    Long:  ``,
    Args:  cobra.ExactArgs(2),

    Run: func(cmd *cobra.Command, args []string) {
        configFiles, err := helpers.GetConfigsFromFlags(cmd)
        if err != nil {
            log.Fatalf("failed to get the correct config files: %v", err)
        }

        name := args[0]
        prompt := args[1]

        err = CreateNewDirective(name, prompt, configFiles)
        if err != nil {
            log.Fatalf("failed to create new directive: %v", err)
        }
    },
}


