package config

import (
    "fmt"
    "strings"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/ai"
)

type DirectiveType string

const (
    NoneDirective DirectiveType = ""
    TextDirective DirectiveType = "text"
    DocsDirective DirectiveType = "docs"
)


type Directive struct {
    Name       string       `yaml:"name"`
    Kind       DirectiveType `yaml:"kind"`
    Prompt     string       `yaml:"prompt"`
    PromptText string       `yaml:"prompt_text"`
    Focus      []string     `yaml:"focus"`
    Ignore     []string     `yaml:"ignore"`
    Output     string       `yaml:"output"`
    ApiKey     string       `yaml:"api_key"`
    Model      ai.Model     `yaml:"model"`
    LocalDocs  string       `yaml:"local_docs"`
    Servers    []string     `yaml:"servers"`
}

type Config struct {
    Files      []string
    Directives map[string]Directive `yaml:"Directives"`
    ApiKey     string               `yaml:"ApiKey"`
    Model      ai.Model             `yaml:"model"`
    LocalDocs  string               `yaml:"local_docs"`
}

func (d *Directive) SanityCheck() error {
    if d.Name == "" {
        return fmt.Errorf("no directive name specified for: %+v", d)
    }

    if d.Kind == NoneDirective {
        return fmt.Errorf("no directive kind specified for %v", d.Name)
    }

    if d.Prompt == "" && d.PromptText == "" {
        return fmt.Errorf("no prompt file or text specified for directive %v", d.Name)
    }

    if d.Focus == nil {
        d.Focus = make([]string, 0)
    }

    if d.Ignore == nil {
        d.Ignore = make([]string, 0)
    }

    if d.ApiKey == "" {
        return fmt.Errorf("no api key specified for directive %v", d.Name)
    }

    if d.Model == ai.NoModel {
        d.Model = DefaultModel
    }

    if d.LocalDocs == "" {
        d.LocalDocs = DefaultLocalDocs
    }

    if d.Servers == nil {
        d.Servers = make([]string, 0)
    }

    return nil
}

func (d *Directive) PrettyPrint(prefix string) {
    fmt.Printf("%vDirective: %v\n", prefix, d.Name)
    fmt.Printf("%v  Kind: %v\n", prefix, d.Kind)
    fmt.Printf("%v  ApiKey: %v\n", prefix, d.ApiKey)
    fmt.Printf("%v  Model: %v\n", prefix, d.Model)
    fmt.Printf("%v  LocalDocs: %v\n", prefix, d.Model)
    fmt.Println("")

    if d.PromptText != "" {
        fmt.Printf("%v  Prompt:\n%v\n", prefix, d.PromptText)
    } else {
        fmt.Printf("%v  Prompt: %v\n", prefix, d.Prompt)
    }

    fmt.Printf("%v  Focus:\n", prefix)
    for _, f := range d.Focus {
        fmt.Printf("%v  - %v\n", prefix, f)
    }

    fmt.Printf("%v  Ignore:\n", prefix)
    for _, f := range d.Ignore {
        fmt.Printf("%v  - %v\n", prefix, f)
    }
    fmt.Printf("%v  Servers:\n", prefix)
    for _, f := range d.Servers {
        fmt.Printf("%v  - %v\n", prefix, f)
    }

    fmt.Println("")
}


func (c *Config) SanityCheck() error {
    // Make sure model is lower case
    c.Model = ai.Model(strings.ToLower(string(c.Model)))

    // Set default values if not present
    for name, directive := range c.Directives {
        directive.Name = strings.ToLower(name)

        if directive.Kind == NoneDirective {
            directive.Kind = DefaultDirective
        }

        directive.Kind = DirectiveType(strings.ToLower(string(directive.Kind)))

        if directive.ApiKey == "" {                 // Value not set in directive itself
            if c.ApiKey != "" {                     // Does this config specify it?
                directive.ApiKey = c.ApiKey
            } else if Settings.ApiKey != "" {       // Do one of the previous configs specify it?
                 directive.ApiKey = Settings.ApiKey
            } else {                                // Fall through to default
                return fmt.Errorf("no api key specified for: %v", name)
            }
       }

        if directive.Model == "" {
            if c.Model != "" {
                directive.Model = c.Model
            } else if Settings.Model != "" {
                 directive.Model = Settings.Model
            } else {
                directive.Model = DefaultModel
            }
        } else { // Verify model is lower case
            directive.Model = ai.Model(strings.ToLower(string(directive.Model)))
        }

        if directive.LocalDocs == "" {
            if c.LocalDocs != "" {
                directive.LocalDocs = c.LocalDocs
            } else if Settings.LocalDocs != "" {
                 directive.LocalDocs = Settings.LocalDocs
            } else {
                directive.LocalDocs = DefaultLocalDocs
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



