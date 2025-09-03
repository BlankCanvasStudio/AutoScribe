package types;

import (
    "os"
    "fmt"
    "context"

    log "github.com/sirupsen/logrus"

    "github.com/openai/openai-go/v2"
    "github.com/openai/openai-go/v2/option"


    "github.com/BlankCanvasStudio/AutoScribe/pkg/ai/formatting"
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
    Model       Model         `yaml:"model,omitempty"`
    Output      string        `yaml:"output,omitempty"`
    ApiKey      string        `yaml:"api_key,omitempty"`
    LocalDocs   string        `yaml:"local_docs,omitempty"`
    Servers     []string      `yaml:"servers,omitempty"`
    Scope       string        `yaml:"-"`
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

    if d.Model == NoModel {
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
            return fmt.Errorf("prompt file %v doesn't exist", d.Prompt)
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
    aiResult := "" 
    var err error

    switch d.Model {
        case GPT_41_Nano:
            return nil
            // return Query4_1Nano(directive, full_prompt)
        default:
            return nil
            // return "", fmt.Errorf("model %v doesn't exist", directive.Model)
    }

    if d.Output == "" {
        fmt.Printf("Output:\n\n%v\n", aiResult)
        return nil
    }

    err = os.WriteFile(d.Output, []byte(aiResult), 0644)
    if err != nil {
        return fmt.Errorf("failed to write ai output to %v: %v", d.Output, err)
    }
    
    return nil
}

func (d *Directive) Query41Nano() (string, error) {

    msg, err := d.GetFullPrompt()
    if err != nil {
        return "", fmt.Errorf("failed to generate full prompt: %v", err)
    }

    // Load API key
    client := openai.NewClient(
            option.WithAPIKey(d.ApiKey),
    )

    chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
            Messages: []openai.ChatCompletionMessageParamUnion{
                    openai.UserMessage(msg),
            },
            Model: openai.ChatModelGPT4_1Nano,
    })

    if err != nil {
            return "", fmt.Errorf("failed to query 4.1 nano : %v", err)
    }

    return chatCompletion.Choices[0].Message.Content, nil
}


func (d *Directive) GetFullPrompt() (string, error) {
    param_prompt := d.PromptText

    if d.PromptText == "" {
        param_prompt_b, err := os.ReadFile(d.Prompt)
        if err != nil {
            return "", fmt.Errorf("failed to read %v: %v", d.Prompt, err)
        }

        param_prompt = string(param_prompt_b)
    }

    data, err := formatting.CombineFilesForContext(d.Focus, d.Ignore)
    if err != nil {
        return "", fmt.Errorf("failed to combine files for context: %v", err)
    }

    full_prompt := fmt.Sprintf(string(param_prompt), data)

    log.Debugf("Full gpt prompt:\n%v\n", full_prompt)

    return full_prompt, nil
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
        d.ApiKey  = u.ApiKey
    }

    if d.LocalDocs == "" {
        d.LocalDocs = u.LocalDocs
    }

    if d.Servers == nil || len(d.Servers) == 0 {
        d.Servers = u.Servers
    }

    return nil
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

