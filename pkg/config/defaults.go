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
Summary: ExpandPaths expands user home directory shortcuts in UserConfigFile and UserDatabaseDir by replacing a leading "~/"
with the actual user home directory. If either path starts with "~/", it will be converted to an absolute path; otherwise the value is left unchanged.

Signature: func ExpandPaths() error

Returns:
- error: non-nil if obtaining the user home directory fails (os.UserHomeDir returns an error). On success, returns nil.

Errors/Exceptions:
- If os.UserHomeDir() returns an error while expanding either UserConfigFile or UserDatabaseDir, that error is returned immediately.

Side Effects:
- Mutates UserConfigFile or UserDatabaseDir by replacing a leading "~/\" with the user's home directory path.

Edge Cases & Assumptions:
- If UserConfigFile or UserDatabaseDir start with "~/" but lack a path component after it (e.g., "~/"), the code expands to the home directory.
- Only the specific "~/\" prefix is expanded; other forms like "~user" are not handled.
- Assumes UserConfigFile and UserDatabaseDir are in scope and of type string.

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
Summary: expandPath expands a path that begins with ~/ to the current user’s home directory; if the input does not start with ~/, it is returned unchanged.
Use when you need to resolve user-home-relative paths in a cross-platform way.

Signature: func expandPath(filename string) (string, error)

Parameters:
- filename: string — the path to potentially expand; if it starts with "~/", it will be expanded to the user’s home directory; otherwise it is returned unchanged.

Returns:
- string: the expanded path if expansion was performed, or the original filename if no expansion was needed.
- error: non-nil if os.UserHomeDir() fails during expansion.

Errors/Exceptions:
- error is non-nil only when os.UserHomeDir() fails to determine the current user’s home directory; in this case the function returns (filename, err).

Side Effects:
- Calls os.UserHomeDir to obtain the current user’s home directory.
- Uses filepath.Join to construct the expanded path.
- No file I/O or external mutations beyond retrieving the home directory.

Edge Cases & Assumptions:
- If filename does not start with \"~/\", the function returns filename, nil.
- If filename starts with \"~/\" but the home directory cannot be determined, returns filename, err.
- Handles OS-specific path separators via filepath.Join.

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
