package helpers;

import (
    "github.com/spf13/cobra"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/cli"
)

func NewCobraConfig() *cobra.Command {

    ret := &cobra.Command{}

    cli.AddFlagsToCmd(ret)

    _ = ret.PersistentFlags().Set("global", "false")
    _ = ret.PersistentFlags().Set("user",   "false")
    _ = ret.PersistentFlags().Set("local",  "false")
    _ = ret.PersistentFlags().Set("config", "")    

    return ret
}
