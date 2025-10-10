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
Summary: Validates and normalizes a Directive, ensuring required fields are set, defaults are applied, and the prompt resource exists when provided. Use this before using a Directive to guarantee it is in a usable, consistent state.

Signature:
func (d *Directive) SanityCheck() error

Parameters:
- d: *Directive — the receiver instance to validate and normalize.

Returns:
- error: non-nil if validation fails; nil if validation and normalization succeed.

Errors/Exceptions:
- error when Name is empty ("no directive name specified for: %+v").
- error when Kind is NoneDirective ("no directive kind specified for %v", d.Name).
- error when both Prompt and PromptText are empty ("no prompt file or text specified for directive %v", d.Name).
- error when Prompt is set but the file does not exist ("prompt file %v doesn't exist").
- error when ApiKey is empty ("no api key specified for directive %v", d.Name).

Side Effects:
- Sets d.Focus to an empty []string if nil.
- Sets d.Ignore to an empty []string if nil.
- Replaces d.Model with DefaultModel when it is NoModel.
- Replaces d.LocalDocs with DefaultLocalDocs when empty.
- Sets d.Servers to an empty []string if nil.
- When Prompt is provided, verifies the prompt file exists via os.Stat.

Edge Cases & Assumptions:
- If Prompt is provided, its file path must exist; otherwise an error is returned.
- PromptText is considered only insofar as providing a prompt when Prompt is empty; both empty triggers an error.
- Defaults rely on predefined constants DefaultModel and DefaultLocalDocs; behavior assumes these are valid defaults.

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
CheckForSave validates that a Directive is ready to be saved.

Summary: Ensures a non-empty Name and that a prompt source is provided (Prompt or PromptText).
If a Prompt file is specified, verifies that the prompt file exists on disk.
Signature: func (d *Directive) CheckForSave() error
Returns: error (nil if valid; non-nil on validation failure or missing prompt file)
Errors/Exceptions:
- "no directive name specified for: %+v" when d.Name == ""
- "no prompt file or text specified for directive %v" when both d.Prompt and d.PromptText are empty
- "prompt file %v doesn't exist" when d.Prompt != "" and the file does not exist
Side Effects: Reads the filesystem via os.Stat when d.Prompt is non-empty
Edge Cases & Assumptions:
- If d.Prompt is non-empty, the path must refer to an existing file.
- The Kind check is present in comments but not enforced by this function.
- The function does not mutate d; it solely validates readiness for saving.

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
Summary: PrettyPrint outputs a human-readable representation of a Directive to standard output,
using the provided prefix to indent lines. Use for debugging or inspecting a Directive's fields.
Signature: func (d *Directive) PrettyPrint(prefix string)
Parameters:
- prefix: string, role: formatting prefix prepended to each output line.
Returns: none.
Errors/Exceptions: none.
Side Effects: writes to standard output via fmt.Printf and fmt.Println.
Behavior details:
- Prints:
  - "Directive: <d.Name>", "  Kind: <d.Kind>", "  ApiKey: <d.ApiKey>", "  Model: <d.Model>", "  LocalDocs: <d.Model>".
- If d.PromptText != "" then prints "  Prompt:" followed by the contents of d.PromptText;
  otherwise prints "  Prompt: <d.Prompt>".
- Prints "  Focus:" followed by each element in d.Focus as "- <value>".
- Prints "  Ignore:" followed by each element in d.Ignore as "- <value>".
- Prints "  Servers:" followed by each element in d.Servers as "- <value>".
- Ends with an empty line.
Edge Cases & Assumptions:
- If d.PromptText is non-empty, it takes precedence over d.Prompt for the Prompt section.
- Nil slices for Focus, Ignore, or Servers are handled gracefully (no output for those sections when nil).

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
Summary: Dispatches execution based on the Directive.Kind. For NoneDirective and TextDirective, it delegates to d.ExecuteTextDirective(); for RecursiveDirective, it delegates to d.ExecuteRecursiveDirective(); any other kind results in an error.
Use when you want to run the appropriate directive handler without invoking the specific handlers directly.
Signature: func (d *Directive) Execute() error
Parameters: none (method on Directive)
Returns: error. Nil on success; non-nil on failure. May be an error from ExecuteTextDirective or ExecuteRecursiveDirective, or a "directive kind %v not implemented" error for unsupported kinds.
Errors/Exceptions:
- "directive kind %v not implemented" if d.Kind is not one of NoneDirective, TextDirective, or RecursiveDirective
- Propagates errors from ExecuteTextDirective() and ExecuteRecursiveDirective() as they are invoked
Side Effects: May perform I/O via the delegated handlers (e.g., text execution, recursive processing), and may mutate internal state through those calls.
Edge Cases & Assumptions:
- NoneDirective is treated the same as TextDirective for execution purposes.
- Only NoneDirective, TextDirective, and RecursiveDirective are supported; others will error.
- The actual behavior and error messages depend on the implementations of ExecuteTextDirective() and ExecuteRecursiveDirective().

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
Summary: Executes a text directive by selecting the configured language model. For GPT_41_Nano, it delegates to d.Query41Nano() to obtain a model response, then either prints the result to stdout or writes it to the file specified by d.Output.
Signature: func (d *Directive) ExecuteTextDirective() error
Parameters: none (method on Directive)
Returns: error. Nil on success; non-nil on failure.
Errors/Exceptions:
  - returns an error if an unsupported model is configured ("model %v doesn't exist")
  - propagates errors from Query41Nano as "failed to query 41Nano: %v"
  - returns an error if writing ai output to d.Output fails ("failed to write ai output to %v: %v")
Side Effects: may perform a network request to the OpenAI API via Query41Nano; may read prompt data from disk via GetFullPrompt; may print to stdout or write to a file.
Edge Cases & Assumptions:
  - only GPT_41_Nano is implemented in this path; other models yield an error
  - when d.Output is non-empty, output is written to that path with 0644 permissions; otherwise, the result is printed
  - relies on Query41Nano to return the first response content as aiResult and to handle its own internal errors

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
Summary: Executes the recursive directive to load Go packages from the folders specified in Directive.Focus, build PackageNode entries, enrich them with PopulatePackageInformation, and generate and persist documentation for the MST. When Language is NoLang or GoLang it processes Go packages; for other languages it returns an error.
Signature: func (d *Directive) ExecuteRecursiveDirective() error
Parameters:
  - none explicit; uses Directive fields:
      - Language: controls language handling (NoLang, GoLang supported)
      - Focus: []string, folders to load packages from
      - PromptText: string, optional inline prompt
      - Prompt: string, path to prompt file if PromptText is empty
      - ApiKey: string, API key for DocumentMST
Returns:
  - error: non-nil on failure; nil on success
Errors/Exceptions:
  - "failed to populate packages: %v" wraps errors from gMst.Populate / PopulatePackageInformation
  - "failed to read %v: %v" if PromptText is empty and reading Prompt file fails
  - "failed to document MST: %v" if mst.DocumentMST fails
  - "cannot parse language %v. not implemented" if Language is not NoLang or GoLang
Side Effects:
  - Mutates gMst.PackageNodes by appending one PackageNode per loaded package
  - Creates new PackageNode instances with MST, Package, FunctionDecls, Imports initialized
  - Per-package documentation is generated by DocumentMST and subsequently updated via UpdateDocsInFile
  - Logs the number of functions declared per package
Edge Cases & Assumptions:
  - If no folders/packages are found, the PackageNodes slice remains unchanged
  - package loading errors are not surfaced beyond the wrap when populating & documenting MST
  - Each package's PopulatePackageInformation is responsible for its own internal errors

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
Summary: Builds a single, ready-to-send prompt by invoking GetFullPrompt, then sends that prompt as a user message to the OpenAI GPT-4.1 Nano model via the OpenAI API, returning the content of the first response choice.
Use when you need a complete prompt that incorporates contextual data from files and then obtain the model's reply.
Signature: func (d *Directive) Query41Nano() (string, error)
Parameters: none (method on Directive)
Returns: (string, error) - the model's reply content and any error encountered.
Errors/Exceptions: returns an error if GetFullPrompt fails ("failed to generate full prompt: ..."), or if the OpenAI API call fails ("failed to query 4.1 nano : ...").
Side Effects: creates an OpenAI client using d.ApiKey; makes a network request to the OpenAI API; may read prompt data from disk via GetFullPrompt.
Edge Cases & Assumptions: assumes GetFullPrompt returns a valid prompt string on success; assumes the API returns at least one Choice and uses Choices[0].Message.Content; may panic if no choices are returned. The behavior of GetFullPrompt depends on d.PromptText, d.Prompt, d.Focus, and d.Ignore.

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
Summary: Returns the complete prompt string by using d.PromptText directly when set, or loading the file at d.Prompt and using its contents as the prompt text; then injects the loaded text as a format string into the data produced by formatting.CombineFilesForContext(d.Focus, d.Ignore). Use when you need a single, ready-to-send prompt that includes contextual file data.
Signature: func (d *Directive) GetFullPrompt() (string, error)
Returns: a full prompt string (string) and a non-nil error if reading the prompt source or combining file context fails.
Errors/Exceptions: returns an error if reading d.Prompt fails (when PromptText == ""), or if formatting.CombineFilesForContext(d.Focus, d.Ignore) returns an error.
Side Effects: reads files from disk; emits a debug log with the final prompt via log.Debugf.
Edge Cases & Assumptions: if d.PromptText is non-empty, its value is used as the prompt template; otherwise the contents of the file at d.Prompt are used. The final prompt is produced by applying fmt.Sprintf to the prompt text with the data from CombineFilesForContext. The behavior of CombineFilesForContext with d.Focus and d.Ignore governs the embedded data content.

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
Summary: Populate the receiver d with values from u for any fields that are not already set, without overwriting existing values.
Use this to apply a default or fallback Directive onto an existing one.
Signature: func (d *Directive) Update(u Directive) error
Parameters:
  d: *Directive - the directive to update (mutated in place).
  u: Directive - the source directive providing fallback values.
Returns:
  error - always nil; this function does not produce an error.
Errors/Exceptions:
  None.
Side Effects:
  Mutates the receiver d by filling in missing fields from u according to specific conditions.
Edge Cases & Assumptions:
  - If d.Kind == NoneDirective, it is set to u.Kind.
  - If d.Description, d.Short, d.Prompt, or d.PromptText are empty, they are set from u.
  - If d.Focus is nil or empty (len == 0), it is set from u.Focus.
  - If d.Ignore is nil or empty (len == 0), it is set from u.Ignore.
  - If d.Model == NoModel, it is set from u.Model.
  - If d.Output, d.ApiKey, or d.LocalDocs are empty, they are set from u.
  - If d.Servers is nil or empty (len == 0), it is set from u.Servers.

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
Summary: NewDirective creates a new Directive with the given name and prompt path after validating the prompt file exists. Use it to instantiate a Directive that references a prompt file by its filesystem path.
Signature: func NewDirective(name, prompt string) (*Directive, error)
Parameters:
  - name string: the directive's Name.
  - prompt string: filesystem path to the prompt file; must exist.
Returns:
  - (*Directive, error): on success, a pointer to Directive with Name: name and Prompt: prompt and nil error; on failure, nil and a descriptive error.
Errors/Exceptions:
  - error when the prompt file does not exist: "prompt file %v doesn't exist".
Side Effects:
  - Reads filesystem via os.Stat; does not modify inputs or global state.
Edge Cases & Assumptions:
  - If os.Stat(prompt) returns an error other than NotExist, the function proceeds and returns a Directive without error.
  - The function only enforces existence for NotExist; it does not verify that prompt is a regular file (directories or symlinks at prompt are treated as existing).
  - No validation on name (e.g., empty string) is performed.

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
