package directives

import (
    // "fmt"
    "github.com/spf13/cobra"

    // log "github.com/sirupsen/logrus"

    // "github.com/BlankCanvasStudio/AutoScribe/pkg/config"
    // "github.com/BlankCanvasStudio/AutoScribe/pkg/cli/helpers"
)

var Cmd = &cobra.Command{
  Use:   "directive",
  Short: "Update information about & add custom directives",
  Long:  ``,
  Run: func(cmd *cobra.Command, args []string) {},
}

func init() {
    Cmd.AddCommand(createCmd)
}


