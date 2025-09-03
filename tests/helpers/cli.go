package helpers;

import (
    "github.com/spf13/cobra"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/cli"
)

func NewCobraConfig() *cobra.Command {

    ret := &cobra.Command{}

    ret.PersistentFlags().Set("global", "false")
    ret.PersistentFlags().Set("user",   "false")
    ret.PersistentFlags().Set("config", "")

    cli.AddFlagsToCmd(ret)

    return ret
}
