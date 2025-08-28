package cli

import (
    "fmt"
    "github.com/spf13/cobra"

    log "github.com/sirupsen/logrus"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/cli/directive"
)

var CfgFile     string
var GlobalLevel bool
var UserLevel   bool

var rootCmd = &cobra.Command{
  Use:   "asb",
  Short: "",
  Long: ``,
  Run: func(cmd *cobra.Command, args []string) {
    // Do Stuff Here
    log.Infof("hugo ran")
  },
}

var versionCmd = &cobra.Command{
  Use:   "version",
  Short: "Print the version number of AutoScribe",
  Long:  `All software has versions. This is AutoScribe's`,
  Run: func(cmd *cobra.Command, args []string) {
    fmt.Println("AutoScribe version 0.1")
  },
}

func Execute() {
    rootCmd.PersistentFlags().BoolP("global", "g", false, "Update global settings (otherwise local folder)")
    rootCmd.PersistentFlags().BoolP("user",   "u", false, "Update user settings (otherwise local folder)")
    rootCmd.PersistentFlags().BoolP("local",  "l", false, "Update current folder settings (default)")

    rootCmd.PersistentFlags().StringVarP(&CfgFile, "config", "c", "", "Specify config file to use")

    rootCmd.AddCommand(versionCmd)
    rootCmd.AddCommand(directives.Cmd)

    if err := rootCmd.Execute(); err != nil {
        log.Fatalf("%v", err)
    }

    log.Infof("Adding cfg: %v", CfgFile)
}
