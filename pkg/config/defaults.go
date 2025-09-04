package config

import (
    "os"
    "strings"
    "path/filepath"
)

var GlobalConfigFile    string = "/etc/autoscribe/conf.yml"
var UserConfigFile      string = "~/.config/autoscribe/conf.yml"
var ProjectConfigFile   string = "./asb.yml"

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

