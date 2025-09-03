package config

import (
    "os"
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
    Name        string        `yaml:"name,omitempty"`
    Kind        DirectiveType `yaml:"kind,omitempty"`
    Description string        `yaml:"description,omitempty"`
    Short       string        `yaml:"short,omitempty"`
    Prompt      string        `yaml:"prompt,omitempty"`
    PromptText  string        `yaml:"prompt_text,omitempty"`
    Focus       []string      `yaml:"focus,omitempty"`
    Ignore      []string      `yaml:"ignore,omitempty"`
    Model       ai.Model      `yaml:"model,omitempty"`
    Output      string        `yaml:"output,omitempty"`
    ApiKey      string        `yaml:"api_key,omitempty"`
    LocalDocs   string        `yaml:"local_docs,omitempty"`
    Servers     []string      `yaml:"servers,omitempty"`
    Scope       string        `yaml:"scope,omit"`
}

type Config struct {
    ApiKey     string               `yaml:"apikey,omitempty"`
    Model      ai.Model             `yaml:"model,omitempty"`
    LocalDocs  string               `yaml:"local_docs,omitempty"`
    Directives map[string]Directive `yaml:"directives,omitempty"`
    Files      []string             `yaml:"files,omitempty"`
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

    if d.Prompt != "" {
        _, err := os.Stat(d.Prompt); 
        if err != nil {
            return fmt.Errorf("prompt file %v doesn't exist")
        }    
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


func (d *Directive) Execute() error {
    data, err := formatting.CombineFilesForContext(directive.Focus, directive.Ignore)
    if err != nil {
        return fmt.Errorf("failed to combine files for context: %v", err)
    }

    aiResult, err := calls.QueryFromDirective(directive, data)

    if directive.Output == "" {
        fmt.Printf("Output:\n\n%v\n", aiResult)
        return nil
    }

    err = os.WriteFile(directive.Output, []byte(aiResult), 0644)
    if err != nil {
        return fmt.Errorf("failed to write ai output to %v: %v", directive.Output, err)
    }
    
    return nil
}


func (d *Directive) Update(u Directive) error {
    if d.Kind == NoneDirective {
        d.Kind = u.Kind
    }

    if d.Description == "" {
        d.Description = u.Description
    }

    if d.Short == "" {
        d.Short = u.Short
    }
    
    if d.Prompt == "" {
        d.Prompt = u.Prompt
    }

    if d.PromptText == "" {
        d.PromptText = u.PromptText
    }

    if d.Focus == nil || len(d.Focus) == 0 {
        d.Focus = u.Focus
    }

    if d.Ignore == nil || len(d.Ignore) == 0 {
        d.Ignore = u.Ignore
    }

    if d.Model == NoModel {
        d.Model = u.Model
    }

    if d.Output == "" {
        d.Output = u.Output
    }

    if d.ApiKey == "" {
        d.== = u..==
    }

    if d.LocalDocs == "" {
        d.LocalDocs = u.LocalDocs
    }

    if d.Server == nil || len(d.Server) == 0 {
        d.Server = u.Server
    }

    return nil
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


func NewConfig() Config {
    return Config {
        Files: []string{},
        Directives: make(map[string]Directive),
    }
}

func NewDirective(name, prompt string) (*Directive, error) {
    _, err := os.Stat(prompt);
    if os.IsNotExist(err)  {
        return nil, fmt.Errorf("prompt file %v doesn't exist", prompt)
    }

    return &Directive {
        Name: name,
        Prompt: prompt,
    }, nil
}

