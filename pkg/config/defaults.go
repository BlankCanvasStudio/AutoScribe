package config

import (
	"os"
	"path/filepath"
	"strings"
)

var GlobalConfigFile string = "/etc/autoscribe/conf.yml"
var UserConfigFile string = "~/.config/autoscribe/conf.yml"
var ProjectConfigFile string = "./asb.yml"

var GlobalDatabaseDir string = "/etc/autoscribe/db"
var UserDatabaseDir string = "~/.config/autoscribe/db"

var IsDocumentedDbBase string = "is-documented.txt"
var DocumentationDbBase string = "documentation.txt"

var IsAiAwareDbBase string = "is-ai-aware.txt"
var NotAiAwareDbBase string = "not-ai-aware.txt"

/*
ExpandPaths expands tilde-prefixed paths to absolute paths using the current user's home directory.
It modifies global variables in place.
If UserConfigFile begins with \"~/\", it is replaced with the user's home directory plus the remainder of the path.
Otherwise, if UserDatabaseDir begins with \"~/\", it is replaced similarly.
Returns an error if the home directory cannot be determined; otherwise returns nil.
Behavior notes: expansion stops after the first successful expansion (UserConfigFile takes precedence); if neither path requires expansion, returns nil.

Signature: func ExpandPaths() error

Parameters: none

Returns: error - non-nil if os.UserHomeDir() fails during expansion; nil on success or if no expansion is needed.

Errors/Exceptions:
- If os.UserHomeDir() returns an error during expansion of UserConfigFile or UserDatabaseDir, that error is returned.
- No other errors are produced; paths are mutated in place when expanded.

Side Effects: Mutates UserConfigFile and/or UserDatabaseDir when expanding \"~/\" prefixes.

Edge Cases & Assumptions:
- Only prefixes exactly matching \"~/\" are expanded; paths starting with other tilde forms (e.g., \"~user\") are not expanded.
- If both UserConfigFile and UserDatabaseDir start with \"~/\", only UserConfigFile is expanded and the function returns immediately afterward.

*/
func ExpandPaths() error {
	if strings.HasPrefix(UserConfigFile, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		UserConfigFile = filepath.Join(home, UserConfigFile[2:])
		return nil
	}

	if strings.HasPrefix(UserDatabaseDir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		UserDatabaseDir = filepath.Join(home, UserDatabaseDir[2:])
		return nil
	}

	return nil
}

/*
Summary: expandPath expands a path that starts with "~/" to the current user's home directory; otherwise it returns the input unchanged.

Signature: func expandPath(filename string) (string, error)

Parameters:
- filename: string. The input path. If it begins with "~/", it will be expanded to the user's home directory; otherwise the value is returned as-is.

Returns:
- string: the resulting path, with "~/" expanded when applicable.
- error: non-nil if the user's home directory cannot be determined.

Errors/Exceptions:
- error from os.UserHomeDir() when the home directory cannot be determined.

Side Effects:
- Reads the current user's home directory via os.UserHomeDir().

Edge Cases & Assumptions:
- Only the "~/" prefix is expanded; other forms like "~user" are not supported.
- If filename == "~/", the result is the home directory path as produced by filepath.Join(home, filename[2:]).

*/
func expandPath(filename string) (string, error) {
	if !strings.HasPrefix(filename, "~/") {
		return filename, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return filename, err
	}

	return filepath.Join(home, filename[2:]), nil
}
