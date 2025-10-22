package directives

import (
	"fmt"

	"github.com/BlankCanvasStudio/AutoScribe/pkg/config"
	"github.com/BlankCanvasStudio/AutoScribe/pkg/types"
	log "github.com/sirupsen/logrus"
)

/*
Summary:
CreateNewDirective orchestrates the creation of a new Directive using name and prompt, then applies it to each path in configFiles. It relies on types.NewDirective(name, prompt) to validate the prompt path, and for each config file it preserves the current configuration, resets to a fresh one, loads the target file, inserts the new directive, and persists the updated configuration.

Signature:
func CreateNewDirective(name string, prompt string, configFiles []string) error

Parameters:
- name: string — the directive name.
- prompt: string — path to the prompt file; must exist on the filesystem. This validation is performed by types.NewDirective.
- configFiles: []string — list of config file paths to update.

Returns:
- error — non-nil on failure; nil on success.
  - Non-nil when creating the new directive fails: "failed to create new directive %v: %v".
  - Non-nil when saving any updated config file fails: "failed to save new directive: %v".

Errors/Exceptions:
- error from NewDirective if the prompt path is invalid: "failed to create new directive %v: %v".
- error from config.SaveConfigFile for any configFile: propagated as "failed to save new directive: %v".

Side Effects:
- Mutates global config state via config.PushLoadedConfig, config.Settings, config.LoadConfigFile, and config.SaveConfigFile.
- For each configFile, updates Settings.Directives[name] to the new directive and persists to disk.
- Logs/debug instrumentation occurs (e.g., directive creation and config updates).

Edge Cases & Assumptions:
- If configFiles is empty, the function performs no updates and returns nil.
- Assumes ConfigStack and Settings are package-global; uses PushLoadedConfig/PopLoadedConfig to manage state across each file.
- If the prompt file does not exist, the error from NewDirective is propagated.

*/
func CreateNewDirective(name string, prompt string, configFiles []string) error {
	log.Debugf("Create directive triggered")

	log.Debugf("Saving to config files: %v", configFiles)

	newDirective, err := types.NewDirective(name, prompt)
	if err != nil {
		return fmt.Errorf("failed to create new directive %v: %v", name, err)
	}

	for _, configFile := range configFiles {
		config.PushLoadedConfig()

		log.Debugf("Updating settings in: %v", configFile)

		config.Settings = config.NewConfig()

		config.LoadConfigFile(configFile)

		config.Settings.Directives[name] = *newDirective

		err := config.SaveConfigFile(configFile, config.Settings)
		if err != nil {
			return fmt.Errorf("failed to save new directive: %v", err)
		}

		config.PopLoadedConfig()
	}

	return nil
}
