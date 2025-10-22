package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/BlankCanvasStudio/AutoScribe/pkg/types"
	"github.com/BlankCanvasStudio/AutoScribe/pkg/types/mst"
)

var Settings Config = Config{
	Files:      []string{},
	Directives: make(map[string]types.Directive),
}

/*
Summary:
LoadConfig orchestrates configuration initialization. It expands tilde-prefixed paths to absolute locations, loads global, user, and project configuration files in precedence order, validates the combined configuration, and resolves database paths by probing writability and expanding base paths. The function mutates global state and returns an error on failure to perform any step.

Signature:
func LoadConfig() error

Parameters:
- none

Returns:
- error: non-nil if any step fails (path expansion, config loading, or validation); nil on success.

Errors/Exceptions:
- "failed to resolve user path: %v" if ExpandPaths() fails.
- Propagated errors from LoadConfigFile(GlobalConfigFile), LoadConfigFile(UserConfigFile), or LoadConfigFile(ProjectConfigFile).
- "failed to sanity check configs: %v" if Settings.SanityCheck() fails.
- "Failed to expand path: %v" or "Failed to expand user path: %v" if path expansion during database base resolution fails.
- Other internal expansion errors may propagate similarly.

Side Effects:
- Mutates global Settings via LoadConfigFile calls.
- Sets mst.IsDocumentedDb, mst.DocumentationDb, mst.IsAiAwareDb, mst.NotAiAwareDb based on expanded paths and writability checks.
- May log errors during path resolution.

Edge Cases & Assumptions:
- Precedence: global config loaded first, then user config overrides, then project config, if files exist.
- expandPath expands only "~/" prefixes; other tilde forms are not handled.
- If a writable database path cannot be created, alternatives are attempted and may still set an alternate mst field.
- If the configuration is incomplete, existing Settings and mst values may remain unchanged for those fields.

*/
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

	// Set up where we are reading & writing our database from
	// Man, I hate this. We will definitely need to re-visit the configuration scheme
	base := IsDocumentedDbBase
	fullConfig, err := expandPath(fmt.Sprintf("%v/%v", GlobalDatabaseDir, base))
	if err != nil {
		return fmt.Errorf("Failed to expand path: %v", err)
	}
	if CanWriteFile(fullConfig) {
		mst.IsDocumentedDb = fullConfig
	} else {
		mst.IsDocumentedDb, err = expandPath(fmt.Sprintf("%v/%v", UserDatabaseDir, base))
		if err != nil {
			return fmt.Errorf("Failed to expand user path: %v", err)
		}
	}

	base = DocumentationDbBase
	fullConfig, err = expandPath(fmt.Sprintf("%v/%v", GlobalDatabaseDir, base))
	if err != nil {
		log.Errorf("Failed to expand path: %v", err)
	}
	if CanWriteFile(fullConfig) {
		mst.DocumentationDb = fullConfig
	} else {
		mst.DocumentationDb, err = expandPath(fmt.Sprintf("%v/%v", UserDatabaseDir, base))
		if err != nil {
			return fmt.Errorf("Failed to expand user path: %v", err)
		}
	}

	base = IsAiAwareDbBase
	fullConfig, err = expandPath(fmt.Sprintf("%v/%v", GlobalDatabaseDir, base))
	if err != nil {
		log.Errorf("Failed to expand path: %v", err)
	}
	if CanWriteFile(fullConfig) {
		mst.IsAiAwareDb = fullConfig
	} else {
		mst.IsAiAwareDb, err = expandPath(fmt.Sprintf("%v/%v", UserDatabaseDir, base))
		if err != nil {
			return fmt.Errorf("Failed to expand user path: %v", err)
		}
	}

	base = NotAiAwareDbBase
	fullConfig, err = expandPath(fmt.Sprintf("%v/%v", GlobalDatabaseDir, base))
	if err != nil {
		log.Errorf("Failed to expand path: %v", err)
	}
	if CanWriteFile(fullConfig) {
		mst.NotAiAwareDb = fullConfig
	} else {
		mst.NotAiAwareDb, err = expandPath(fmt.Sprintf("%v/%v", UserDatabaseDir, base))
		if err != nil {
			return fmt.Errorf("Failed to expand user path: %v", err)
		}
	}

	return nil
}

/*
Summary:
LoadConfigFile reads a YAML configuration file at filename and merges its values into the global Settings.
If the file does not exist, it is a no-op. Use this to initialize or augment runtime configuration from disk.
Signature:
func LoadConfigFile(filename string) error
Parameters:
- filename: string — path to the YAML config file to load.
Returns:
- error: non-nil if reading, parsing, or processing the file fails; nil otherwise.
  - nil when the file does not exist (no-op).
  - non-nil for read errors, YAML parse errors, or other failures during loading.
Errors/Exceptions:
- "failed to find config %v: %v" when os.Stat returns an error other than NotExist.
- "error reading config file: %v" if os.ReadFile fails.
- "error parsing yaml: %v" if yaml.Unmarshal fails.
Side Effects:
- Mutates the global Settings (Files, ApiKey, Model, Directives) based on the loaded config.
- Extends Settings.Directives with directives from the file, setting directive.Scope to filename.
- Logs a debug message: "Loading config from %v" when a file is loaded.
Edge Cases & Assumptions:
- If cfg.ApiKey or cfg.Model are non-empty, they override corresponding Settings values.
- For cfg.Directives, existing directives in Settings are updated (via Directive.Update) when a matching name exists; otherwise new directives are added with Scope set to filename.
- If the file exists but is empty or missing expected fields, existing Settings values may remain unchanged.

*/
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

	cfg := Config{Files: []string{filename}}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("error parsing yaml: %v", err)
	}

	if Settings.Files == nil {
		Settings.Files = []string{filename}
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
		d, exists := Settings.Directives[name]
		if exists {
			directive.Update(d)
		}

		directive.Scope = filename

		Settings.Directives[name] = directive
	}

	return nil
}

/*
Summary: Persists a Config to a YAML file at filename, excluding the Files field from the output.
This is achieved by setting cfg.Files = nil on a local copy before marshaling and writing.
Use when you need to persist configuration state without the associated Files data.

Signature: func SaveConfigFile(filename string, cfg Config) error

Parameters:
- filename: string. Path to the YAML file to write.
- cfg: Config. The value is serialized to YAML after cfg.Files is cleared locally; the Files field is not saved.

Returns:
- error. Non-nil on marshal or write failure; nil on success.

Errors/Exceptions:
- "failed to marshal config: %v" if YAML marshaling fails.
- "failed to write %v: %v" if writing to disk fails.

Side Effects:
- Writes the YAML representation of cfg to the filesystem at filename with permissions 0644.
- cfg.Files is set to nil for the duration of the function (local copy only; does not affect caller's cfg).

Edge Cases & Assumptions:
- If filename is empty or the path is unwritable, the function returns an error.
- All other Config fields are serialized according to the YAML library's rules; Files is intentionally omitted.

*/
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

/*
Summary: Verifies that the local configuration file exists at ProjectConfigFile. If the file is missing, it creates an empty file with mode 0644. If the path already exists, it returns nil. If an error occurs while inspecting the path, it returns that error.
Signature: func VerifyLocalConfigExists() error
Parameters: none
Returns: error — nil on success; non-nil on failure. If the file was missing, a new empty file is created and the function returns nil unless creation fails, in which case an error is returned.
Errors/Exceptions: - If os.Stat reports an error other than IsNotExist, that error is returned. - If creating the file fails, returns fmt.Errorf("failed to write to file: %v", ProjectConfigFile). - If the path exists, the function returns nil (even if it is a directory); it does not verify the path is a regular file.
Side Effects: May create the file at ProjectConfigFile with permissions 0644; may write to disk.
Edge Cases & Assumptions: Assumes ProjectConfigFile is a valid path. The function is idempotent when the file already exists. If ProjectConfigFile exists but is a directory, the function returns nil.

*/
func VerifyLocalConfigExists() error {
	// check file
	if _, err := os.Stat(ProjectConfigFile); os.IsNotExist(err) {
		if err := os.WriteFile(ProjectConfigFile, []byte{}, 0644); err != nil {
			return fmt.Errorf("failed to write to file: %v", ProjectConfigFile)
		}
	} else if err == nil {
		return nil
	} else {
		return err
	}

	return nil
}

/*
VerifyUserConfigExists ensures that the user configuration file and its parent directory exist.
It guarantees the parent directory via os.MkdirAll(filepath.Dir(UserConfigFile), 0755) and,
if the UserConfigFile does not exist, creates an empty file at UserConfigFile with 0644 permissions.
If the file already exists, the function returns nil. Any filesystem error encountered is returned.

Signature: func VerifyUserConfigExists() error

Parameters: none

Returns: error — nil on success; non-nil on failure

Errors/Exceptions:
- "failed to make directories: %v" when directory creation fails
- "failed to write to file: %v" when creating the empty UserConfigFile fails
- any error returned by os.Stat other than os.IsNotExist

Side Effects:
- Creates the parent directory of UserConfigFile if needed
- Creates an empty file at UserConfigFile if it does not exist

Edge Cases & Assumptions:
- If UserConfigFile already exists, the function is a no-op and returns nil
- If UserConfigFile's parent directory cannot be created or the file cannot be written, an error is returned
- Relies on the global/UserConfigFile and its filesystem permissions; behavior depends on the environment

*/
func VerifyUserConfigExists() error {
	// ensure parent dir exists
	if err := os.MkdirAll(filepath.Dir(UserConfigFile), 0755); err != nil {
		return fmt.Errorf("failed to make directories: %v", err)
	}

	// check file
	if _, err := os.Stat(UserConfigFile); os.IsNotExist(err) {
		if err := os.WriteFile(UserConfigFile, []byte{}, 0644); err != nil {
			return fmt.Errorf("failed to write to file: %v", UserConfigFile)
		}
	} else if err == nil {
		return nil
	} else {
		return err
	}

	return nil
}

/*
Summary: Ensures the global configuration file exists by creating its parent directory if missing
and creating an empty file when absent. Use this before reading or writing the global config to
guarantee a writable path.
Signature: func VerifyGlobalConfigExists() error
Parameters: none
Returns: error — non-nil if the operation fails; nil if the parent directory was created and/or the
file exists or was created successfully.
Errors/Exceptions:
  - non-nil from the underlying failure to create directories: "failed to make directories: %v"
  - non-nil from failing to write the file: "failed to write to file: %v" (with the path GlobalConfigFile)
  - any non-nil error returned by os.Stat other than IsNotExist
Side Effects: Creates the parent directory for GlobalConfigFile (with 0755 permissions) and,
  if the file does not exist, creates GlobalConfigFile with empty content (mode 0644).
Edge Cases & Assumptions:
  - If the parent directory already exists, MkdirAll is a no-op.
  - If GlobalConfigFile already exists, the function leaves it unchanged and returns nil.
  - The function relies on GlobalConfigFile being a valid path defined elsewhere.
  - The in-code comments indicate behavior: "ensure parent dir exists" and "check file".

*/
func VerifyGlobalConfigExists() error {
	// ensure parent dir exists
	if err := os.MkdirAll(filepath.Dir(GlobalConfigFile), 0755); err != nil {
		return fmt.Errorf("failed to make directories: %v", err)
	}

	// check file
	if _, err := os.Stat(GlobalConfigFile); os.IsNotExist(err) {
		if err := os.WriteFile(GlobalConfigFile, []byte{}, 0644); err != nil {
			return fmt.Errorf("failed to write to file: %v", GlobalConfigFile)
		}
	} else if err != nil {
		return err
	}

	return nil
}

var ConfigStack = []Config{}

/*
Summary: Save the current Settings by appending it to ConfigStack, then reset Settings to a new, empty Config via NewConfig.

Signature: func PushLoadedConfig() error

Parameters: none

Returns: error — always nil

Errors/Exceptions: none

Side Effects: mutates package-level state by updating ConfigStack and Settings

Edge Cases & Assumptions: assumes ConfigStack and Settings are package-level variables; relies on NewConfig() to produce a Config with default empty fields; no error handling is performed

*/
func PushLoadedConfig() error {
	ConfigStack = append(ConfigStack, Settings)
	Settings = NewConfig()

	return nil
}

/*
Summary:
Pops the most recently loaded configuration from the global ConfigStack and assigns it to Settings, restoring the previous configuration. Returns an error if the stack is empty.

Signature:
func PopLoadedConfig() error

Parameters:
- none

Returns:
- error: nil on success; non-nil on failure

Errors/Exceptions:
- error when ConfigStack is empty: "can't pop empty config stack"

Side Effects:
- Mutates Settings by restoring the top element of ConfigStack
- Mutates ConfigStack by removing the top element

Edge Cases & Assumptions:
- If ConfigStack is empty, no changes to Settings or ConfigStack occur; the error is returned.
- Assumes ConfigStack and Settings are package-global identifiers available to this function.

*/
func PopLoadedConfig() error {
	if len(ConfigStack) == 0 {
		return fmt.Errorf("can't pop empty config stack")
	}

	Settings = ConfigStack[len(ConfigStack)-1]
	ConfigStack = ConfigStack[:len(ConfigStack)-1]

	return nil
}

/*
Summary: CanWriteFile reports whether the current process can create or write to the specified path. It attempts to open path with os.O_WRONLY|os.O_CREATE and file mode 0644; if successful, it closes the file and returns true; otherwise it returns false.

Signature: func CanWriteFile(path string) bool

Parameters:
- path: string
  - Role: target file path to test writability.
  - Constraints: may be non-empty; parent directory must exist; creating the file is allowed.

Returns:
- bool
  - true: the file can be opened for writing (or created) with 0644 permissions.
  - false: an error occurred while opening the file for writing, or the path is not writable.

Errors/Exceptions:
- None exposed; internal errors cause a false return value.

Side Effects:
- May create the file at path if it does not exist (due to os.O_CREATE).
- Opens the file for writing briefly and then closes it.

Edge Cases & Assumptions:
- If path refers to a directory, or parent directory does not exist, or write permission is denied, returns false.
- Existing file is opened for writing without truncation; no data is written by this function.
- Behavior may vary by OS with respect to file permissions and open semantics.

*/
func CanWriteFile(path string) bool {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return false
	}
	f.Close()
	return true
}
