package config

import (
    "os"
    "strings"
    "path/filepath"
    "github.com/BlankCanvasStudio/AutoScribe/pkg/ai"
)

var GlobalConfigFile    string = "/etc/autoscribe/conf.yml"
var UserConfigFile      string = "~/.config/autoscribe/conf.yml"
var ProjectConfigFile   string = "./asb.yml"

// Default values
var DefaultModel ai.Model = ai.GPT_41_Nano

var DefaultDirective = TextDirective

var DefaultLocalDocs = "/opt/autoscribe/docs-database"

func ExpandPaths() (error) {
	if strings.HasPrefix(UserConfigFile, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
                UserConfigFile = filepath.Join(home, UserConfigFile[2:])
                return nil
	}
	return nil
}

