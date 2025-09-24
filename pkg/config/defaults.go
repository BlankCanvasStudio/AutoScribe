package config

import (
    "os"
    "strings"
    "path/filepath"
)

var GlobalConfigFile    string = "/etc/autoscribe/conf.yml"
var UserConfigFile      string = "~/.config/autoscribe/conf.yml"
var ProjectConfigFile   string = "./asb.yml"

var GlobalDatabaseDir   string = "/etc/autoscribe/db"
var UserDatabaseDir     string = "~/.config/autoscribe/db"

var IsDocumentedDbBase  string = "is-documented.txt"
var DocumentationDbBase string = "documentation.txt"

var IsAiAwareDbBase     string = "is-ai-aware.txt"
var NotAiAwareDbBase    string = "not-ai-aware.txt"



func ExpandPaths() (error) {
	if strings.HasPrefix(UserConfigFile, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
                UserConfigFile = filepath.Join(home, UserConfigFile[2:])
                return nil
	}

	if strings.HasPrefix(UserDatabaseDir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
                UserDatabaseDir = filepath.Join(home, UserDatabaseDir[2:])
                return nil
	}

	return nil
}

func expandPath(filename string) (string, error) {
    if !strings.HasPrefix(filename, "~/") {
        return filename, nil
    }

    home, err := os.UserHomeDir()
    if err != nil {
        return filename, err
    }

    return filepath.Join(home, filename[2:]), nil
}

