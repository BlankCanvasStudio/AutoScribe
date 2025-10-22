package helpers

import (
	"fmt"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"

	"github.com/BlankCanvasStudio/AutoScribe/pkg/config"
)

/*
Summary: Builds the list of configuration file paths to load based on flags on the provided Cobra command. It reads "global", "user", "local", and a custom config via "config", and returns the config file paths in a defined order. If no scopes are selected, it defaults to the ProjectConfigFile.

Signature: func GetConfigsFromFlags(cmd *cobra.Command) ([]string, error)

Parameters:
- cmd: *cobra.Command — command from which to read flags. Supported flags: "global" (bool), "user" (bool), "local" (bool), "config" (string).

Returns:
- []string: ordered list of config file paths to load. If custom scope is used, includes the provided config file; otherwise includes file paths for local (ProjectConfigFile), user (UserConfigFile), and/or global (GlobalConfigFile) as requested. If no scopes are selected, returns []string{ProjectConfigFile}.
- error: non-nil on failure.

Errors/Exceptions:
- On failure to read scope flags from GetConfigScopeFlags(cmd): returns an error wrapped as "failed to load config scope flags: %v".
- If custom scope is active and GetString("config") fails: the function calls log.Fatalf, terminating the process.
- If the custom config file path is empty: returns error "config file empty".
- If VerifyLocalConfigExists() fails: wraps as "failed to create user config file: %v".
- If VerifyUserConfigExists() fails: wraps as "failed to create user config file: %v".
- If VerifyGlobalConfigExists() fails: wraps as "failed to create global config file: %v".

Side Effects:
- May create or verify the existence of configuration files and their parent directories via VerifyLocalConfigExists, VerifyUserConfigExists, and VerifyGlobalConfigExists.
- May terminate the process via log.Fatalf when reading the "config" flag.

Edge Cases & Assumptions:
- custom is processed before local, user, and global scopes; local, user, and global are appended if their respective scopes are true.
- If no scopes are selected, the function defaults to ProjectConfigFile.
- Assumes the identifiers ProjectConfigFile, UserConfigFile, GlobalConfigFile, and the config package functions are defined elsewhere.

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
Summary: Determines which configuration scope is active by reading the flags "global", "user", "local" and whether a custom config file is provided via "config" on the given Cobra command. Returns four booleans (global, user, local, custom) and an error.
Signature: func GetConfigScopeFlags(cmd *cobra.Command) (bool, bool, bool, bool, error)
Parameters:
- cmd: *cobra.Command — command from which to read flags. Supported flags: "global" (bool), "user" (bool), "local" (bool), "config" (string).
Returns:
- bool global: value of the "global" flag.
- bool user: value of the "user" flag.
- bool local: value of the "local" flag.
- bool custom: true if "config" is non-empty, false otherwise.
- error: non-nil if any flag retrieval fails.
Errors/Exceptions:
- If cmd.Flags().GetBool("global") fails, returns (global, user, local, custom, error) with message "failed to get global var global: %v".
- If cmd.Flags().GetBool("user") fails, returns (global, user, local, custom, error) with message "failed to get global var user: %v".
- If cmd.Flags().GetBool("local") fails, returns (global, user, local, custom, error) with message "failed to get global var local: %v".
- If cmd.Flags().GetString("config") fails, returns (global, user, local, custom, error) with message "failed to get global var local: %v".
Side Effects:
- Reads flag values from cmd. No other I/O or state mutation.
Edge Cases & Assumptions:
- custom is true only when the "config" flag yields a non-empty string; otherwise custom is false.
- On success, all four booleans are set according to the flag values, and error is nil.

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
