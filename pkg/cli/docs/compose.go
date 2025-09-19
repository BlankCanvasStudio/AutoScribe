package docs

import (
    "github.com/spf13/cobra"

    log "github.com/sirupsen/logrus"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/docs"
)

func init() {
    // Cmd.AddCommand(CreateCmd)
}


var UndocCmd = &cobra.Command{
  Use:   "undoc [files / folders]",
  Short: "Remove documentation from files & folders",
  Long:  ``,
  Run: func(cmd *cobra.Command, args []string) {
    for _, arg := range args {
        err := docs.Undocument(arg, true)
        if err != nil {
            log.Fatalf("Failed to undocument %v: %v", arg, err)
        }
    }
  },
}

