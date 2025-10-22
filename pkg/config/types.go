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

/*
Summary: Validates and normalizes a Config by applying defaults and ensuring required fields for each directive. It also normalizes casing for Model and directive fields, assigns ApiKey from global defaults when missing, and rebuilds the c.Directives map with lower-case directive names. Delegates per-directive validation to Directive.SanityCheck().

Signature: func (c *Config) SanityCheck() error

Parameters:
- c: *Config — the receiver being validated.

Returns:
- error — non-nil if validation or normalization fails; nil otherwise.

Errors/Exceptions:
- Propagates errors from directive.SanityCheck() with context: "failed to sanity check %v: %v"
- If a directive has no ApiKey and no ApiKey provided via c.ApiKey or Settings.ApiKey, returns an error message: "no api key specified in config for: %v. Perhaps you need to update /etc/autoscribe/conf.yml?"
- Other per-directive validation errors are surfaced similarly.

Side Effects:
- Mutates c.Model to lower-case.
- For each directive:
  - directive.Name = strings.ToLower(name)
  - If directive.Kind == NoneDirective, set to DefaultDirective
  - directive.Kind = types.DirectiveType(strings.ToLower(string(directive.Kind)))
  - If directive.ApiKey is "", inherits from c.ApiKey or Settings.ApiKey if present; otherwise error.
  - If directive.Model == "", inherits from c.Model or Settings.Model; otherwise sets to DefaultModel
  - If directive.LocalDocs == "", inherits from c.LocalDocs or Settings.LocalDocs; otherwise sets to DefaultLocalDocs
- Rebuilds c.Directives by removing the old key and inserting the updated directive with directive.Name as the key.
- May return an error from directive.SanityCheck().

Edge Cases & Assumptions:
- If c.Directives is nil or empty, the method completes with nil.
- If two directives collapse to the same lower-case name, the latter overwrites the former.
- Defaulting behavior depends on DefaultModel, DefaultLocalDocs, NoModel, and NoneDirective constants.
- Prompt validation and file existence are handled by Directive.SanityCheck(); PromptText is accepted without file validation.

*/
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

		if directive.ApiKey == "" { // Value not set in directive itself
			if c.ApiKey != "" { // Does this config specify it?
				directive.ApiKey = c.ApiKey
			} else if Settings.ApiKey != "" { // Do one of the previous configs specify it?
				directive.ApiKey = Settings.ApiKey
			} else { // Fall through to default
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

/*
Summary: Pretty-prints the Config to standard output for debugging, displaying top-level fields and the details of each Directive.
Use this to inspect the current configuration state and its directives during development or troubleshooting.

Signature: func (c *Config) PrettyPrint()

Parameters: none
Returns: none

Errors/Exceptions: none

Side Effects: writes formatted output to standard output via fmt.Println and fmt.Printf; calls d.PrettyPrint("  ") for each directive in c.Directives

Edge Cases & Assumptions:
 - If c.Directives is nil or empty, no directive output is produced.
 - LocalDocs is printed using c.Model due to the code (i.e., "LocalDocs: %v" with c.Model), which may reflect a bug or intentional quirk.

*/
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

/*
Summary: NewConfig returns a Config initialized with default empty values, preparing it for population.
Signature: func NewConfig() Config
Returns: Config — a new Config where Files is []string{} and Directives is make(map[string]types.Directive).
Side Effects: allocates a new Config instance and initializes its fields.
Edge Cases & Assumptions: none; always returns a valid Config with default empty collections.

*/
func NewConfig() Config {
	return Config{
		Files:      []string{},
		Directives: make(map[string]types.Directive),
	}
}
