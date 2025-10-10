package directives

import (
	"fmt"

	"github.com/BlankCanvasStudio/AutoScribe/pkg/config"
	"github.com/BlankCanvasStudio/AutoScribe/pkg/types"
	log "github.com/sirupsen/logrus"
)

/*
Summary: CreateNewDirective creates a new Directive from the given name and prompt, then updates each provided config file by pushing the current loaded config, loading the file, assigning the new directive to Settings.Directives[name], and saving the updated config to disk. A final PopLoadedConfig reverts the config stack.
Signature: func CreateNewDirective(name string, prompt string, configFiles []string) error
Parameters:
- name string: the directive's Name.
- prompt string: filesystem path to the prompt; passed to types.NewDirective.
- configFiles []string: list of config file paths to update with the new directive.
Returns:
- error: nil on success; non-nil if directive creation or any config save fails.
Errors/Exceptions:
- error when creating the new directive: "failed to create new directive %v: %v"
- error when saving a config file: "failed to save new directive: %v"
Side Effects:
- Mutates Settings through Load/Save of each config file.
- Pushes and pops the loaded config stack around per-file operations.
- Sets config.Settings.Directives[name] to *newDirective for each file.
Edge Cases & Assumptions:
- If configFiles is empty, no per-file updates occur and the function returns nil after creating the directive.
- PopLoadedConfig is called after processing all files; its return value is ignored.

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
