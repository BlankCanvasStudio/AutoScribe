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
LoadConfig expands user-relative paths, loads configuration from global, user, and project YAML config files (if present),
validates and normalizes the resulting configuration, and resolves database path settings by preferring global
database directories when writable and falling back to user directories otherwise.

Signature:
func LoadConfig() error

Parameters:
- None

Returns:
- error: non-nil on failure (e.g., path expansion failures, config file load errors, or failed sanity checks); nil on success.

Errors/Exceptions:
- "failed to resolve user path: %v" if ExpandPaths() fails.
- Propagates errors from LoadConfigFile(GlobalConfigFile), LoadConfigFile(UserConfigFile), and LoadConfigFile(ProjectConfigFile).
- "failed to sanity check configs: %v" if Settings.SanityCheck() fails.
- "Failed to expand path: %v" or "Failed to expand user path: %v" if path expansion of database bases fails.

Side Effects:
- Mutates global mst fields (IsDocumentedDb, DocumentationDb, IsAiAwareDb, NotAiAwareDb) based on expanded and writability-checked paths.
- May log debug messages or errors during path expansion.
- Invokes ExpandPaths, LoadConfigFile for multiple files, and Settings.SanityCheck() which may modify internal state.

Edge Cases & Assumptions:
- ExpandPaths() expands leading "~/" in relevant paths; expandPath handles only paths starting with "~/".
- LoadConfigFile may be a no-op if a config file is absent; errors from loading present files are surfaced.
- The function assumes GlobalDatabaseDir, UserDatabaseDir, and the various base constants (IsDocumentedDbBase, DocumentationDbBase, IsAiAwareDbBase, NotAiAwareDbBase) are defined in the package and used to construct candidate database paths.
- If none of the candidate paths are writable, corresponding mst fields remain unchanged.

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
LoadConfigFile loads and applies configuration from a YAML file specified by filename.
If the file does not exist, the function performs no action. If stat, read, or parse errors occur,
it returns a non-nil error.

Signature: func LoadConfigFile(filename string) error

Parameters:
  filename string - path to the YAML config file. If the file does not exist, no action is taken.

Returns:
  error - non-nil on failure reading or parsing the file; nil if the file was absent or successfully applied.

Errors/Exceptions:
  - If os.Stat(filename) yields an error other than NotExist, returns: "failed to find config %v: %v".
  - If reading the file fails, returns: "error reading config file: %v".
  - If YAML unmarshalling fails, returns: "error parsing yaml: %v".

Side Effects:
  - Updates global Settings: appends the filename to Settings.Files, and to per-file directive scope via Settings.Directives.
  - Prefers and applies values from cfg (ApiKey, Model) to Settings when provided.
  - For each entry in cfg.Directives, merges with any existing directive in Settings.Directives by calling directive.Update(d),
    then sets directive.Scope to filename and stores the directive in Settings.Directives[name].
  - Logs a debug message: "Loading config from %v" when a file is loaded.

Edge Cases & Assumptions:
  - If cfg.Directives includes a directive that already exists in Settings.Directives, the existing directive is used to fill
    in missing fields of the loaded directive via directive.Update(d) before storing.
  - The function assumes Config and related types (Config, Directive, Settings) are defined in the package.
  - If cfg.ApiKey or cfg.Model are empty, Settings.ApiKey and Settings.Model are not modified.

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
Summary: Writes the given Config to a YAML file, clearing the Files field beforehand to exclude it from serialization. Use when you need to persist a Config as YAML to disk and ensure the Files data is omitted.

Signature: func SaveConfigFile(filename string, cfg Config) error

Parameters:
- filename string: path to the output file. The function writes/overwrites this path with 0644 permissions.
- cfg Config: configuration to save. The function sets cfg.Files = nil prior to marshaling to exclude Files from the YAML output.

Returns:
- error: non-nil if marshaling or file writing fails. On success, returns nil.

Errors/Exceptions:
- "failed to marshal config: %v" if yaml.Marshal fails.
- "failed to write %v: %v" if os.WriteFile fails.

Side Effects:
- cfg.Files is set to nil (mutates the provided cfg).
- Writes YAML data to the specified filename on disk.

Edge Cases & Assumptions:
- Assumes Config can be marshaled by yaml.Marshal.
- The Files field is intentionally cleared to influence the serialized output.
- Uses 0644 permissions for the written file; may be affected by OS umask.

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
Summary: VerifyLocalConfigExists ensures that the local project config file exists on disk. If the file is missing, it creates an empty file at ProjectConfigFile with permissions 0644; if the file already exists, it returns nil without modification. Use during startup or test setup to guarantee a config file is present.
Signature: func VerifyLocalConfigExists() error
Returns: error indicating a failure to stat or write the file; otherwise nil.
Errors/Exceptions:
- os.Stat returns an error other than os.IsNotExist: the error is returned.
- The case where the file is missing and os.WriteFile fails: returns a wrapped error with message "failed to write to file: %v", ProjectConfigFile.
Side Effects:
- Performs filesystem I/O: Stat on ProjectConfigFile and potentially WriteFile to create the file.
- May create a new file on disk when it is missing.
Edge Cases & Assumptions:
- Assumes ProjectConfigFile is a valid path string.
- Not synchronized for concurrent calls; a race may occur between Stat and WriteFile if the file is created by another process.
- If the file exists, function returns nil regardless of current permissions; does not verify writability.
- If ProjectConfigFile points to a directory or an inaccessible path, os.Stat will return an error which is propagated.

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
Summary: VerifyUserConfigExists ensures the user config file exists by creating its parent directory
and the file itself if necessary. Use this before reading or writing UserConfigFile.
Signature: func VerifyUserConfigExists() error
Parameters: none
Returns: error
Errors/Exceptions: returns a non-nil error if the parent directory cannot be created
or if writing the file fails; otherwise returns nil.
Side Effects: Creates the parent directory (via os.MkdirAll) and may create UserConfigFile
as an empty file (via os.WriteFile) when it does not already exist.
Edge Cases & Assumptions: If os.Stat returns an error other than IsNotExist, that error is returned.
Assumes UserConfigFile is a path to the configuration file; parent directory will be created with 0755.
Quotes from code: "ensure parent dir exists" and "check file"

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
Summary: Guarantees the presence of the global configuration path by ensuring the parent directory exists and the file itself exists. It creates the parent directory if needed and, if the file is missing, creates an empty file at GlobalConfigFile.

Signature: func VerifyGlobalConfigExists() error

Returns: error
  - nil if the required directory and file exist or are created successfully.
  - non-nil if directory creation or file creation fails, or if os.Stat on GlobalConfigFile returns an error other than IsNotExist.

Errors/Exceptions:
  - "failed to make directories: %v" when MkdirAll fails.
  - "failed to write to file: %v" when WriteFile fails.
  - any non-nil error returned by os.Stat when the path exists but an error occurs.

Side Effects:
  - Creates directories for the parent of GlobalConfigFile with mode 0755.
  - Creates an empty GlobalConfigFile with mode 0644 if it does not already exist.

Edge Cases & Assumptions:
  - If GlobalConfigFile already exists, the function returns nil without modifying the file.
  - If a non-regular file or an invalid path exists at GlobalConfigFile, errors from the underlying filesystem calls propagate.

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
Summary: PushLoadedConfig saves the current Settings by appending it to ConfigStack, then resets Settings to a new default configuration by calling NewConfig().
Signature: func PushLoadedConfig() error
Returns: error - always nil
Side Effects: Mutates global state by: 1) ConfigStack is appended with the current Settings; 2) Settings is replaced with the value returned by NewConfig(); may allocate memory for a new Config.
Edge Cases & Assumptions: Assumes ConfigStack and Settings are defined at package scope and that NewConfig() returns a distinct Config value. After the call, the previous Settings remains in ConfigStack and current Settings points to a fresh configuration.
Errors/Exceptions: None; the function returns nil unconditionally.

*/
func PushLoadedConfig() error {
	ConfigStack = append(ConfigStack, Settings)
	Settings = NewConfig()

	return nil
}

/*
Summary: Pops the most recently loaded configuration from ConfigStack and applies it to Settings. Returns an error if the stack is empty.

Signature: func PopLoadedConfig() error

Parameters: none

Returns:
  - error: nil on success; non-nil on failure
  - On failure, the error is fmt.Errorf("can't pop empty config stack")

Errors/Exceptions:
  - can't pop empty config stack

Side Effects:
  - Sets Settings to the popped configuration
  - Truncates ConfigStack by removing the top element

Edge Cases & Assumptions:
  - If ConfigStack is empty, the function returns an error and does not modify Settings.
  - Assumes ConfigStack is a slice where the last element is the top of the stack.

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
Summary:
CanWriteFile reports whether the current process can write to the file at the given path. It attempts to open the file with write access (creating it if it does not exist) and returns true on success, or false if the open operation fails.

Signature:
CanWriteFile(path string) bool

Parameters:
  path string - path to the file to test writability. Must be a valid filesystem path.

Returns:
  bool - true if the file can be opened for writing (os.O_WRONLY|os.O_CREATE) without error; false otherwise.

Errors/Exceptions:
  This function does not return errors. All failures are indicated by returning false.

Side Effects:
  Attempts to open the file with write permission; may create the file if it does not exist; closes the file immediately after opening.

Edge Cases & Assumptions:
  - If the file already exists but is not writable, this returns false.
  - There is a potential race condition between opening and subsequent writes if the file is altered by another process.
  - The created file uses permission 0644 when created.
  - Empty or invalid paths will cause os.OpenFile to fail and return false.

*/
func CanWriteFile(path string) bool {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return false
	}
	f.Close()
	return true
}
