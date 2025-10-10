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
Summary
Validates and normalizes a Config, applying defaults and ensuring each Directive
is in a usable, consistent state. Use before using a Config to guarantee proper
initialization and predictable behavior.

Signature
func (c *Config) SanityCheck() error

Parameters
- c: *Config — the receiver instance to validate and normalize.

Returns
- error: non-nil if validation or normalization fails; nil on success.

Errors/Exceptions
- non-nil error if a Directive fails its own SanityCheck; the error is returned with
  context: "failed to sanity check <name>: <err>".
- non-nil error when no API key is specified after considering Directive.ApiKey,
  Config.ApiKey, and Settings.ApiKey: "no api key specified in config for: %v. Perhaps you need to update /etc/autoscribe/conf.yml?".
- other errors may propagate from directive.SanityCheck() calls.

Side Effects
- Modifies c.Model to lowercase: c.Model = types.Model(strings.ToLower(string(c.Model)))
- Normalizes directive field values and defaults within c.Directives:
  - directive.Name is set to lowercase of the original key
  - directive.Kind defaults to types.DefaultDirective if NoneDirective, then lowercased
  - directive.ApiKey is filled from directive.ApiKey, c.ApiKey, or Settings.ApiKey
  - directive.Model is filled from directive.Model, c.Model, Settings.Model, or defaults to DefaultModel
  - directive.LocalDocs is filled from directive.LocalDocs, c.LocalDocs, Settings.LocalDocs, or defaults to DefaultLocalDocs
- Rebuilds c.Directives with lowercase keys by replacing each entry
- Invokes directive.SanityCheck() for each Directive

Edge Cases & Assumptions
- If a Directive has no ApiKey and no fallback ApiKey is available, SanityCheck()
  returns an error as described above.
- Defaults rely on predefined constants DefaultModel and DefaultLocalDocs; behavior
  assumes these constants are valid.
- Settings.* and global defaults are used as fallbacks when a field is not specified on the Directive
  or Config.
- If any Directive.SanityCheck() fails, SanityCheck() returns an error for that Directive.

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
Summary: PrettyPrint prints a human-readable representation of a Config to standard output for debugging and inspection. It outputs core fields and delegates to each Directive's PrettyPrint method.

Signature: func (c *Config) PrettyPrint()

Parameters: none.

Returns: none.

Errors/Exceptions: none.

Side Effects: writes to standard output via fmt.Println and fmt.Printf.

Behavior details:
- Prints an initial blank line.
- Prints:
  - "Configs: %v" with c.Files
  - "  ApiKey: %v" with c.ApiKey
  - "  Model: %v" with c.Model
  - "  LocalDocs: %v" with c.Model (note: this appears to display Model instead of LocalDocs)
- Prints an empty line.
- Iterates over c.Directives and for each element d, calls d.PrettyPrint("  ").
- Ends with an empty line.

Edge Cases & Assumptions:
- If c.Directives is nil or empty, the loop contributes no output.
- Each element of c.Directives exposes a PrettyPrint(prefix string) method.
- The LocalDocs field is printed using c.Model (likely a bug in the code).

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
Summary: Creates a new Config value with default initializations: an empty Files slice and an empty Directives map. Use this to obtain a fresh configuration object before populating its fields.
Signature: func NewConfig() Config
Returns: Config with Files: []string{} and Directives: make(map[string]types.Directive)
Side Effects: allocates a new slice and a new map; no I/O or external state.
Edge Cases & Assumptions: Assumes Config has the fields Files []string and Directives map[string]types.Directive; returns a new independent Config value and does not reuse or modify existing state.

*/
func NewConfig() Config {
	return Config{
		Files:      []string{},
		Directives: make(map[string]types.Directive),
	}
}
