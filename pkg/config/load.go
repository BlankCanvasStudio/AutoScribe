package config

import (
	"os"
	"fmt"
	// "flag"
        "path/filepath"
	"gopkg.in/yaml.v3"

	log "github.com/sirupsen/logrus"

	// "github.com/BlankCanvasStudio/AutoScribe/pkg/ai"
	// "github.com/BlankCanvasStudio/AutoScribe/pkg/types"
)


var Settings Config = Config {
    Files: []string{},
    Directives: make(map[string]Directive),
}


func LoadConfig() error {
    // Make sure we have absolute paths to everything. Go freaks out with a ~
    err := ExpandPaths()
    if err != nil {
        return fmt.Errorf("failed to resolve user path: %v", err)
    }

    // Source global configs
    err = LoadConfigFile(GlobalConfigFile)
    if err != nil {
        return err
    }

    // Grab user configs (preferred over global)
    err = LoadConfigFile(UserConfigFile)
    if err != nil {
        return err
    }

    // Prefer local configs
    err = LoadConfigFile(ProjectConfigFile)
    if err != nil {
        return err
    }

    err = Settings.SanityCheck()
    if err != nil {
        return fmt.Errorf("failed to sanity check configs: %v", err)
    }

    return nil
}


func LoadConfigFile(filename string) error {
	_, err := os.Stat(filename)

	if os.IsNotExist(err) {
            return nil

	} else if err != nil {
		return fmt.Errorf("failed to find config %v: %v", filename, err)
	}

	log.Debugf("Loading config from %v", filename)

	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading config file: %v", err)
	}

        cfg := Config{ Files: []string{ filename } }
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("error parsing yaml: %v", err)
	}

        if Settings.Files == nil {
            Settings.Files = []string{ filename }
        } else {
            Settings.Files = append(Settings.Files, filename)
        }

        // Always prefer "more local values" if specified
        if cfg.ApiKey != "" {
            Settings.ApiKey = cfg.ApiKey
        }

        if cfg.Model != "" {
            Settings.Model = cfg.Model
        }

        for name, directive := range cfg.Directives {
            Settings.Directives[name] = directive
        }

	return nil
}

func SaveConfigFile(filename string, cfg Config) error {
    cfg.Files = nil
    	data, err := yaml.Marshal(&cfg)
	if err != nil {
            return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
             return fmt.Errorf("failed to write %v: %v", filename, err)
	}

        return nil
}

func VerifyLocalConfigExists() error {
    // check file
    if _, err := os.Stat(ProjectConfigFile); os.IsNotExist(err) {
            if err := os.WriteFile(ProjectConfigFile, []byte{}, 0644); err != nil {
                return fmt.Errorf("failed to write to file: %v")
            }
    } else if err == nil {
        return nil
    } else {
        return err
    }

    return nil
}

func VerifyUserConfigExists() error {
    // ensure parent dir exists
    if err := os.MkdirAll(filepath.Dir(UserConfigFile), 0755); err != nil {
        return fmt.Errorf("failed to make directories: %v", err)
    }

    // check file
    if _, err := os.Stat(UserConfigFile); os.IsNotExist(err) {
            if err := os.WriteFile(UserConfigFile, []byte{}, 0644); err != nil {
                return fmt.Errorf("failed to write to file: %v")
            }
    } else if err == nil {
        return nil
    } else {
        return err
    }

    return nil
}

func VerifyGlobalConfigExists() error {
    // ensure parent dir exists
    if err := os.MkdirAll(filepath.Dir(GlobalConfigFile), 0755); err != nil {
        return fmt.Errorf("failed to make directories: %v", err)
    }

    // check file
    if _, err := os.Stat(GlobalConfigFile); os.IsNotExist(err) {
            if err := os.WriteFile(GlobalConfigFile, []byte{}, 0644); err != nil {
                return fmt.Errorf("failed to write to file: %v")
            }
    } else if err != nil {
        return err
    }

    return nil
}

var ConfigStack = []Config{}


func PushLoadedConfig() error {
    ConfigStack = append(ConfigStack, Settings)
    Settings = NewConfig()

    return nil
}

func PopLoadedConfig() error {
    if len(ConfigStack) == 0 {
        return fmt.Errorf("can't pop empty config stack")
    }

    Settings = ConfigStack[len(ConfigStack) - 1]
    ConfigStack = ConfigStack[:len(ConfigStack) - 1]

    return nil
}

