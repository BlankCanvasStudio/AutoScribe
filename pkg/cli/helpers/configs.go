package helpers

import (
	"fmt"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"

	"github.com/BlankCanvasStudio/AutoScribe/pkg/config"
)

/*
Summary
Determines which configuration files to load from the given Cobra command's flags. It inspects the boolean flags global, user, and local, and the string flag config. It returns the selected config file paths in a slice, in the order custom, local, user, global (if provided). If no scope flags are set, it returns a slice containing ProjectConfigFile as the default.

Signature
func GetConfigsFromFlags(cmd *cobra.Command) ([]string, error)

Parameters
- cmd: *cobra.Command — the command whose flags are consulted.

Returns
- []string — ordered list of configuration file paths to load. May be empty if an error occurs.
- error — non-nil if reading flags or preparing config files fails; otherwise nil.

Errors/Exceptions
- If GetConfigScopeFlags(cmd) returns an error, this function returns that error wrapped as
  "failed to load config scope flags: %v".
- When the custom config path is requested but reading the config flag fails, the function calls
  log.Fatalf, which terminates the process.
- If the custom config path is read as an empty string, returns an error "config file empty".
- If VerifyLocalConfigExists, VerifyUserConfigExists, or VerifyGlobalConfigExists fail, returns an error
  wrapped as "failed to create <scope> config file: %v".
- If no scope flags are provided, defaults to returning a slice containing ProjectConfigFile.

Side Effects
- Reads flags from cmd; may perform filesystem I/O to ensure config files exist (creating files/directories as needed).

Edge Cases & Assumptions
- Assumes flags "global", "user", "local" are boolean flags and "config" is a string flag.
- Custom scope takes precedence; when provided, its file is added first in the result.
- If multiple scopes are enabled, all corresponding files are appended in the order: custom, local, user, global.
- If none of the scopes are selected, the function returns [ProjectConfigFile].

*/
func GetConfigsFromFlags(cmd *cobra.Command) ([]string, error) {
	configFiles := make([]string, 0, 3)

	globalScope, userScope, localScope, customScope, err := GetConfigScopeFlags(cmd)
	if err != nil {
		return configFiles, fmt.Errorf("failed to load config scope flags: %v", err)
	}

	if customScope {

		configFile, err := cmd.Flags().GetString("config")
		if err != nil {
			log.Fatalf("failed to get config var: %v", err)
		}

		log.Debugf("working with custom scope %v", configFile)

		if configFile == "" {
			return nil, fmt.Errorf("config file empty")
		}

		configFiles = append(configFiles, configFile)
	}

	if localScope {
		log.Debugf("working with config in local scope")

		err := config.VerifyLocalConfigExists()
		if err != nil {
			return configFiles, fmt.Errorf("failed to create user config file: %v", err)
		}

		configFiles = append(configFiles, config.ProjectConfigFile)
	}

	if userScope {
		log.Debugf("working with config in user scope")

		err := config.VerifyUserConfigExists()
		if err != nil {
			return configFiles, fmt.Errorf("failed to create user config file: %v", err)
		}

		configFiles = append(configFiles, config.UserConfigFile)
	}

	if globalScope {
		log.Debugf("working with config in global scope")

		err := config.VerifyGlobalConfigExists()
		if err != nil {
			return configFiles, fmt.Errorf("failed to create global config file: %v", err)
		}

		configFiles = append(configFiles, config.GlobalConfigFile)
	}

	if len(configFiles) == 0 {
		configFiles = append(configFiles, config.ProjectConfigFile)
	}

	return configFiles, nil
}

/*
Summary
Determines the configuration scope flags from a Cobra command. Reads the boolean flags "global", "user", and "local", and the string flag "config". Returns the values as (global, user, local, custom, err), where custom is true when the "config" flag yields a non-empty string. If any flag retrieval fails, returns an error describing which flag could not be read alongside the current flag values.

Signature
func GetConfigScopeFlags(cmd *cobra.Command) (bool, bool, bool, bool, error)

Parameters
- cmd: *cobra.Command — the command whose flags are consulted.

Returns
- global: bool — value of the "global" flag.
- user: bool — value of the "user" flag.
- local: bool — value of the "local" flag.
- custom: bool — true if a non-empty string is provided by the "config" flag.
- err: error — non-nil if a flag read fails; otherwise nil.

Errors/Exceptions
- If cmd.Flags().GetBool("global") fails, returns (global, user, local, custom, fmt.Errorf("failed to get global var global: %v", err)).
- If cmd.Flags().GetBool("user") fails, returns (global, user, local, custom, fmt.Errorf("failed to get global var user: %v", err)).
- If cmd.Flags().GetBool("local") fails, returns (global, user, local, custom, fmt.Errorf("failed to get global var local: %v", err)).
- If cmd.Flags().GetString("config") fails, returns (global, user, local, custom, fmt.Errorf("failed to get global var local: %v", err)).

Side Effects
- Queries the provided command's flags; no other side effects.

Edge Cases & Assumptions
- Assumes flags "global", "user", "local" are boolean flags and "config" is a string flag.
- custom is derived solely from whether config is non-empty: custom = customFile != "".
- In case of an error, previously read flag values are returned alongside the error.

*/
func GetConfigScopeFlags(cmd *cobra.Command) (bool, bool, bool, bool, error) {
	var err error

	global := false
	user := false
	local := false
	custom := false

	global, err = cmd.Flags().GetBool("global")
	if err != nil {
		return global, user, local, custom, fmt.Errorf("failed to get global var global: %v", err)
	}

	user, err = cmd.Flags().GetBool("user")
	if err != nil {
		return global, user, local, custom, fmt.Errorf("failed to get global var user: %v", err)
	}

	local, err = cmd.Flags().GetBool("local")
	if err != nil {
		return global, user, local, custom, fmt.Errorf("failed to get global var local: %v", err)
	}

	customFile, err := cmd.Flags().GetString("config")
	if err != nil {
		return global, user, local, custom, fmt.Errorf("failed to get global var local: %v", err)
	}

	custom = customFile != ""

	return global, user, local, custom, nil
}
