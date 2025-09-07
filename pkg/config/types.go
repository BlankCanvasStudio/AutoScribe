package config

import (
    "fmt"
    "strings"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/types"
)

type Config struct {
    ApiKey     string                     `yaml:"apikey,omitempty"`
    Model      types.Model                `yaml:"model,omitempty"`
    LocalDocs  string                     `yaml:"local_docs,omitempty"`
    Directives map[string]types.Directive `yaml:"directives,omitempty"`
    Files      []string                   `yaml:"files,omitempty"`
}


func (c *Config) SanityCheck() error {
    // Make sure model is lower case
    c.Model = types.Model(strings.ToLower(string(c.Model)))

    // Set default values if not present
    for name, directive := range c.Directives {
        directive.Name = strings.ToLower(name)

        if directive.Kind == types.NoneDirective {
            directive.Kind = types.DefaultDirective
        }

        directive.Kind = types.DirectiveType(strings.ToLower(string(directive.Kind)))

        if directive.ApiKey == "" {                 // Value not set in directive itself
            if c.ApiKey != "" {                     // Does this config specify it?
                directive.ApiKey = c.ApiKey
            } else if Settings.ApiKey != "" {       // Do one of the previous configs specify it?
                 directive.ApiKey = Settings.ApiKey
            } else {                                // Fall through to default
                return fmt.Errorf("no api key specified in config for: %v. Perhaps you need to update /etc/autoscribe/conf.yml?", name)
            }
       }

        if directive.Model == "" {
            if c.Model != "" {
                directive.Model = c.Model
            } else if Settings.Model != "" {
                 directive.Model = Settings.Model
            } else {
                directive.Model = types.DefaultModel
            }
        } else { // Verify model is lower case
            directive.Model = types.Model(strings.ToLower(string(directive.Model)))
        }

        if directive.LocalDocs == "" {
            if c.LocalDocs != "" {
                directive.LocalDocs = c.LocalDocs
            } else if Settings.LocalDocs != "" {
                 directive.LocalDocs = Settings.LocalDocs
            } else {
                directive.LocalDocs = types.DefaultLocalDocs
            }
        }

        err := directive.SanityCheck()
        if err != nil {
            return fmt.Errorf("failed to sanity check %v: %v", name, err)
        }

        // Convert all directive names to lower case
        delete(c.Directives, name)
        c.Directives[directive.Name] = directive
    }

    return nil
}


func (c *Config) PrettyPrint() {
    fmt.Println("")

    fmt.Printf("Configs: %v\n", c.Files)
    fmt.Printf("  ApiKey: %v\n", c.ApiKey)
    fmt.Printf("  Model: %v\n", c.Model)
    fmt.Printf("  LocalDocs: %v\n", c.Model)

    fmt.Println("")

    for _, d := range c.Directives {
        d.PrettyPrint("  ")
    }

    fmt.Println("")
}


func NewConfig() Config {
    return Config {
        Files: []string{},
        Directives: make(map[string]types.Directive),
    }
}


