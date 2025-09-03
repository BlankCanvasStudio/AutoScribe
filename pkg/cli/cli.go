package cli

import (
    "fmt"
    "slices"
    "github.com/spf13/cobra"

    log "github.com/sirupsen/logrus"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/config"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/cli/helpers"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/cli/directives"
)

var CfgFile     string
var GlobalLevel bool
var UserLevel   bool

var AdditionalPrompt string

var debug bool

var rootCmd = &cobra.Command{
    Use:   "asb",
    Short: "",
    Long: ``,
    Run: func(cmd *cobra.Command, args []string) {
        // Do Stuff Here
        log.Infof("hugo ran")
    },
    PersistentPreRun: func(cmd *cobra.Command, args []string) {
            if debug {
                    log.SetLevel(log.DebugLevel)
                    log.Debug("Debug logging enabled")
            } else {
                    log.SetLevel(log.InfoLevel)
            }
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

var runCmd = &cobra.Command{
    Use:   "run",
    Short: "Run all the directives initialized in this repository",
    Long:  `Run all the directives initialized in this repository`,
    Run: func(cmd *cobra.Command, args []string) {
        configScopes, err := helpers.GetConfigsFromFlags(cmd)
        if err != nil {
            log.Fatalf("failed to get configs from flags: %v", err)
        }

        for name, directive := range config.Settings.Directives {
            if !slices.Contains(configScopes, directive.Scope) {
                continue
            }

            err := directive.Execute()
            if err != nil {
                log.Fatalf("failed to execte directive %v: %v", name, err)
            }
        }
    },
}

func Execute() {
    rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging")
    rootCmd.PersistentFlags().BoolP("global", "g", false, "Utilize global settings (otherwise local folder)")
    rootCmd.PersistentFlags().BoolP("user",   "u", false, "Utilize user settings (otherwise local folder)")
    rootCmd.PersistentFlags().BoolP("local",  "l", false, "Utilize current folder settings (default)")
    rootCmd.PersistentFlags().StringVarP(&CfgFile, "prompt", "p", "", "Add additional context to your directive's prompt")

    rootCmd.PersistentFlags().StringVarP(&CfgFile, "config", "c", "", "Specify config file to use")

    rootCmd.AddCommand(runCmd)
    rootCmd.AddCommand(versionCmd)
    rootCmd.AddCommand(directives.Cmd)

    customCmds, err := directives.CreateCustomCommands()
    if err != nil {
        log.Fatalf("Couldn't build custom directives: %v", customCmds)
    }

    for _, cmd := range customCmds {
        rootCmd.AddCommand(cmd)
    }


    if err := rootCmd.Execute(); err != nil {
        log.Fatalf("%v", err)
    }
}

