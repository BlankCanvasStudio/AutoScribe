package types

import (
	"context"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"

	"github.com/BlankCanvasStudio/AutoScribe/pkg/types/mst"
	"github.com/BlankCanvasStudio/AutoScribe/pkg/types/mst/golang"

	"github.com/BlankCanvasStudio/AutoScribe/pkg/ai/formatting"
)

type DirectiveType string

const (
	NoneDirective      DirectiveType = ""
	TextDirective      DirectiveType = "text"
	RecursiveDirective DirectiveType = "recursive"
)

type DirectiveLanguage string

const (
	NoLang DirectiveLanguage = ""
	GoLang DirectiveLanguage = "golang"
)

type Directive struct {
	Name        string            `yaml:"name,omitempty"`
	Kind        DirectiveType     `yaml:"kind,omitempty"`
	Language    DirectiveLanguage `yaml:"language,omitempty"`
	Description string            `yaml:"description,omitempty"`
	Short       string            `yaml:"short,omitempty"`
	Prompt      string            `yaml:"prompt,omitempty"`
	PromptText  string            `yaml:"prompt_text,omitempty"`
	Focus       []string          `yaml:"focus,omitempty"`
	Ignore      []string          `yaml:"ignore,omitempty"`
	Model       Model             `yaml:"model,omitempty"`
	Output      string            `yaml:"output,omitempty"`
	ApiKey      string            `yaml:"api_key,omitempty"`
	LocalDocs   string            `yaml:"local_docs,omitempty"`
	Servers     []string          `yaml:"servers,omitempty"`
	Scope       string            `yaml:"-"`
}

/*
Summary: Validates and normalizes a Directive instance before use. Ensures required fields are set and fills in defaults; optionally verifies the existence of the prompt file when provided.

Signature: func (d *Directive) SanityCheck() error

Parameters:
- d: *Directive — the receiver being validated and normalized.

Returns:
- error: non-nil if validation fails (see Errors/Exceptions); nil if validation and normalization succeed.

Errors/Exceptions:
- "no directive name specified for: %+v" if d.Name == ""
- "no directive kind specified for %v" if d.Kind == NoneDirective
- "no prompt file or text specified for directive %v" if d.Prompt == "" && d.PromptText == ""
- "prompt file %v doesn't exist" if d.Prompt != "" and the file does not exist

Side Effects:
- Mutates d.F ocus to an empty []string when nil
- Mutates d.Ignore to an empty []string when nil
- If d.Model == NoModel, sets d.Model = DefaultModel
- If d.LocalDocs == "", sets d.LocalDocs = DefaultLocalDocs
- If d.Servers == nil, sets d.Servers = make([]string, 0)

Edge Cases & Assumptions:
- Requires a non-empty d.Name and a non-None directive kind.
- Requires at least a prompt file path (Prompt) or prompt text (PromptText); PromptText is accepted without file validation.
- When Prompt is provided, its file must exist; the check uses os.Stat.
- Defaulting behavior depends on constants DefaultModel, DefaultLocalDocs, and NoModel / NoneDirective.

*/
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
		_, err := os.Stat(d.Prompt)
		if err != nil {
			return fmt.Errorf("prompt file %v doesn't exist", d.Prompt)
		}
	}

	return nil
}

/*
Summary: Validates that a Directive has enough information to be saved and that any referenced prompt file exists. Used before persisting a Directive.
Signature: func (d *Directive) CheckForSave() error
Returns: error — non-nil when validation fails; nil if the Directive is ready to save.
Errors/Exceptions:
  - "no directive name specified for: %+v" if d.Name is the empty string.
  - "no prompt file or text specified for directive %v" if both d.Prompt and d.PromptText are empty.
  - "prompt file %v doesn't exist" if d.Prompt is non-empty but the file does not exist.
Side Effects: When d.Prompt != "", performs os.Stat(d.Prompt) to verify the file exists.
Edge Cases & Assumptions:
  - The check for d.Kind == NoneDirective is present in comments and not executed (disabled).
  - If both Prompt and PromptText are provided, only the existence of Prompt is checked; PromptText is not validated in this path.

*/
func (d *Directive) CheckForSave() error {
	if d.Name == "" {
		return fmt.Errorf("no directive name specified for: %+v", d)
	}

	/*
	   if d.Kind == NoneDirective {
	       return fmt.Errorf("no directive kind specified for %v", d.Name)
	   }
	*/

	if d.Prompt == "" && d.PromptText == "" {
		return fmt.Errorf("no prompt file or text specified for directive %v", d.Name)
	}

	if d.Prompt != "" {
		_, err := os.Stat(d.Prompt)
		if err != nil {
			return fmt.Errorf("prompt file %v doesn't exist", d.Prompt)
		}
	}

	return nil
}

/*
Summary: Pretty-prints the Directive to standard output with a given prefix, displaying its Name, Kind, ApiKey, Model, LocalDocs (as the Model value), Prompt or PromptText, and the lists in Focus, Ignore, and Servers for debugging/inspection.
Signature: func (d *Directive) PrettyPrint(prefix string)
Parameters:
  - d: *Directive, receiver (implicit) — the directive instance to print
  - prefix: string — indentation prefix applied to all output lines
Returns: none
Errors/Exceptions: none
Side Effects: writes formatted output to standard output via fmt.Printf and fmt.Println
Edge Cases & Assumptions:
  - If d.PromptText != "", the PromptText block is printed; otherwise the Prompt value is printed.
  - Headers for Focus, Ignore, and Servers are always printed; if the corresponding slices are empty, no items follow the header.
  - LocalDocs is displayed using d.Model (i.e., the code prints "LocalDocs: <Model>"), which may reflect a implementation quirk.

*/
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

/*
Summary: Executes the Directive by routing to the appropriate specialized executor based on d.Kind. Delegates to ExecuteTextDirective() for NoneDirective or TextDirective, or to ExecuteRecursiveDirective() for RecursiveDirective; otherwise returns an error indicating the kind is not implemented.
Signature: func (d *Directive) Execute() error
Parameters:
  - d *Directive - receiver; uses d.Kind and related fields to determine the execution path.
Returns:
  - error - non-nil if the selected execution path fails or if the directive kind is not implemented; nil on success.
Errors/Exceptions:
  - "directive kind %v not implemented" when d.Kind is not one of the handled kinds.
  - Any error returned by d.ExecuteTextDirective() or d.ExecuteRecursiveDirective().
Side Effects:
  - May invoke d.ExecuteTextDirective() or d.ExecuteRecursiveDirective(), potentially performing I/O, network calls, or state mutations via those paths.
Edge Cases & Assumptions:
  - If d.Kind == NoneDirective or TextDirective, the text execution path is chosen; if RecursiveDirective, the recursive path is chosen; otherwise an error is returned.
  - Relies on downstream methods to perform their own validation and error reporting.

*/
func (d *Directive) Execute() error {
	if d.Kind == NoneDirective || d.Kind == TextDirective {
		return d.ExecuteTextDirective()
	}

	if d.Kind == RecursiveDirective {
		return d.ExecuteRecursiveDirective()
	}

	return fmt.Errorf("directive kind %v not implemented", d.Kind)
}

/*
Summary: Executes a text directive by dispatching to the model-specific prompt generation and AI query path (currently GPT_41_Nano), returning the assistant's reply or writing it to a file.
Use when you want to process a Directive into AI-generated text and either print it or persist it to disk.
Signature: func (d *Directive) ExecuteTextDirective() error
Parameters:
  d *Directive - receiver; uses d.ApiKey, d.Model, d.PromptText, d.Prompt, d.Focus, d.Ignore, and potentially d.Output.
Returns:
  error - non-nil if prompt generation, API query, or file I/O fails; nil on success.
Errors/Exceptions:
  - "failed to query 41Nano: %v" if d.Query41Nano() returns an error.
  - "model %v doesn't exist" if d.Model is unrecognized.
Side Effects:
  - May create an OpenAI client with d.ApiKey and perform a network request via d.Query41Nano().
  - GetFullPrompt (invoked within the query path) may read prompt templates/files during prompt assembly.
  - If d.Output is non-empty, writes the AI output to the path d.Output; otherwise prints the output to stdout.
Edge Cases & Assumptions:
  - For GPT_41_Nano: GetFullPrompt must succeed and return a non-empty prompt; assumes at least one chatCompletion.Choices entry is available (accessed as chatCompletion.Choices[0].Message.Content); uses context.TODO() for the API call.

*/
func (d *Directive) ExecuteTextDirective() error {
	aiResult := ""
	var err error

	switch d.Model {
	case GPT_41_Nano:
		aiResult, err = d.Query41Nano()
		if err != nil {
			return fmt.Errorf("failed to query 41Nano: %v", err)
		}

	default:
		return fmt.Errorf("model %v doesn't exist", d.Model)
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

/*
Summary: Executes a recursive directive for the current Directive by loading Go packages from the specified folders, constructing PackageNode entries, populating their metadata (imports, function declarations, etc.), and generating documentation via MST. This path is implemented for d.Language equal to NoLang or GoLang; other languages return an error.

Signature: func (d *Directive) ExecuteRecursiveDirective() error

Parameters: none (uses fields on the Directive: d.Focus, d.PromptText, d.Prompt, d.ApiKey)

Returns:
- error: non-nil if any step fails, or if the language is unsupported.

Errors/Exceptions:
- "failed to populate packages: %v" if gMst.Populate(d.Focus) returns an error.
- "failed to read %v: %v" if reading the prompt file fails when d.PromptText is empty.
- "failed to document MST: %v" if mst.DocumentMST(&gMst, prompt, d.ApiKey) returns an error.
- "cannot parse language %v. not implemented" if d.Language is not NoLang or GoLang.
- Note: Errors from packages.Load are assigned but not surfaced here.

Side Effects:
- Mutates gMst via Populate(d.Focus).
- Each PackageNode is mutated by PopulatePackageInformation(), updating its Imports and FunctionDecls (and related metadata).
- Calls UpdateDocsInFile() on each discovered package to refresh its documentation.
- Logs progress and errors related to per-package documentation.

Edge Cases & Assumptions:
- Assumes folders contains valid Go package roots and that packages.Load returns packages for each folder.
- If no packages are found, the function returns nil with m.PackageNodes left empty or partially populated.
- If d.PromptText is "", a valid file must be readable from d.Prompt to provide the prompt content.
- PackageNode.PopulatePackageInformation handles its own internal processing and error reporting.

*/
func (d *Directive) ExecuteRecursiveDirective() error {
	if d.Language == NoLang || d.Language == GoLang {
		gMst := golang.MST{}

		err := gMst.Populate(d.Focus)
		if err != nil {
			return fmt.Errorf("failed to populate packages: %v", err)
		}

		prompt := d.PromptText

		if d.PromptText == "" {
			promptBytes, err := os.ReadFile(d.Prompt)
			if err != nil {
				return fmt.Errorf("failed to read %v: %v", d.Prompt, err)
			}
			prompt = string(promptBytes)
		}

		_, err = mst.DocumentMST(&gMst, prompt, d.ApiKey)
		if err != nil {
			return fmt.Errorf("failed to document MST: %v", err)
		}

		for _, pkg := range gMst.GetPackages() {
			err = pkg.UpdateDocsInFile()
			if err != nil {
				log.Errorf("Failed to update documentation for %v: %v", pkg.GetPath(), err)
			}
		}

	} else {
		return fmt.Errorf("cannot parse language %v. not implemented", d.Language)
	}

	return nil
}

/*
Summary: Obtain the full prompt via GetFullPrompt and send it to the OpenAI API using model GPT4_1Nano as a chat completion; return the assistant's reply content.

Signature: func (d *Directive) Query41Nano() (string, error)

Parameters:
  d *Directive - receiver; uses d.ApiKey and the fields utilized by GetFullPrompt (d.PromptText, d.Prompt, d.Focus, d.Ignore).

Returns:
  string - the assistant's reply content from the OpenAI model.
  error  - non-nil if prompt generation or the OpenAI query fails.

Errors/Exceptions:
  - "failed to generate full prompt: %v" if GetFullPrompt returns an error.
  - "failed to query 4.1 nano : %v" if the OpenAI API call fails.

Side Effects:
  Creates an OpenAI client with d.ApiKey and performs a network request to request a chat completion; GetFullPrompt may read prompt templates/files as part of prompt assembly.

Edge Cases & Assumptions:
  Assumes GetFullPrompt succeeds and returns a non-empty prompt; assumes at least one choice is returned in chatCompletion (access via chatCompletion.Choices[0].Message.Content); uses context.TODO() for the API call.

*/
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

/*
Summary: Build the full prompt by substituting a data blob into the prompt template. If d.PromptText is non-empty, it is used as the template; otherwise, the contents of the file at d.Prompt are read and used. The data blob is produced by formatting.CombineFilesForContext(d.Focus, d.Ignore) and injected via fmt.Sprintf. The result is logged at debug level and returned.
Signature: func (d *Directive) GetFullPrompt() (string, error)
Parameters:
  d *Directive - receiver; uses d.PromptText, d.Prompt, d.Focus, d.Ignore.
Returns:
  string - the fully composed prompt text ready for sending to a model.
  error  - non-nil if reading the prompt template or combining files for context fails.
Errors/Exceptions:
  - "failed to read %v: %v" if reading the prompt template file fails.
  - "failed to combine files for context: %v" if formatting.CombineFilesForContext fails.
Side Effects:
  Reads the filesystem (os.ReadFile or equivalent inside CombineFilesForContext); logs the full prompt via log.Debugf.
Edge Cases & Assumptions:
  If d.PromptText != "", it is treated as the prompt template; otherwise the template is read from the file at d.Prompt.
  The template is treated as a format string for a single data argument produced by formatting.CombineFilesForContext(d.Focus, d.Ignore).
  Directories listed in d.Focus are traversed by CombineFilesForContext in its own logic; files are appended in traversal order.

*/
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

/*
Summary:
Updates the receiver d by filling in unset fields from the provided Directive u. Only overwrites fields that are currently zero-valued or nil; existing values are preserved. Use when you want to apply a default or fallback Directive to an existing one without clobbering already-specified settings.

Signature:
func (d *Directive) Update(u Directive) error

Parameters:
- u: Directive — source of values to apply to d when corresponding fields are unset.

Returns:
- error: always nil; this function does not produce an error.

Details:
- If d.Kind == NoneDirective, sets d.Kind = u.Kind.
- If d.Description == "", sets d.Description = u.Description.
- If d.Short == "", sets d.Short = u.Short.
- If d.Prompt == "", sets d.Prompt = u.Prompt.
- If d.PromptText == "", sets d.PromptText = u.PromptText.
- If d.Focus == nil or len(d.Focus) == 0, sets d.Focus = u.Focus.
- If d.Ignore == nil or len(d.Ignore) == 0, sets d.Ignore = u.Ignore.
- If d.Model == NoModel, sets d.Model = u.Model.
- If d.Output == "", sets d.Output = u.Output.
- If d.ApiKey == "", sets d.ApiKey = u.ApiKey.
- If d.LocalDocs == "", sets d.LocalDocs = u.LocalDocs.
- If d.Servers == nil or len(d.Servers) == 0, sets d.Servers = u.Servers.

Side Effects:
- Mutates the receiver d by assigning values from u to previously unset fields.

Edge Cases & Assumptions:
- Fields use zero-value checks for strings and slices; for Focus/Ignore/Servers, nil or empty slices trigger updates.
- If u provides zero-values, no change occurs for the corresponding field.
- The function path does not fail and always returns nil.

*/
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
		d.ApiKey = u.ApiKey
	}

	if d.LocalDocs == "" {
		d.LocalDocs = u.LocalDocs
	}

	if d.Servers == nil || len(d.Servers) == 0 {
		d.Servers = u.Servers
	}

	return nil
}

/*
Summary:
NewDirective validates that the provided prompt path exists and, if so, returns a new Directive with the given name and Prompt fields set. Use it to initialize a Directive that relies on an existing prompt file.

Signature:
func NewDirective(name, prompt string) (*Directive, error)

Parameters:
- name: string — the directive name.
- prompt: string — path to the prompt file; must exist on the filesystem.

Returns:
- *Directive: a pointer to a Directive with Name = name and Prompt = prompt.
- error: non-nil if the prompt file does not exist.

Errors/Exceptions:
- error if the prompt file does not exist: "prompt file %v doesn't exist"

Side Effects:
- Checks file existence via os.Stat(prompt).

Edge Cases & Assumptions:
- If os.Stat(prompt) returns an error other than non-existence, the function proceeds to return a Directive (no error returned) because the only explicit error check is os.IsNotExist(err).
- Existence is checked, not the file type; a directory at prompt will still be accepted as valid.

*/
func NewDirective(name, prompt string) (*Directive, error) {
	_, err := os.Stat(prompt)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("prompt file %v doesn't exist", prompt)
	}

	return &Directive{
		Name:   name,
		Prompt: prompt,
	}, nil
}
