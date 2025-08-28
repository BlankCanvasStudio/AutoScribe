package config

import (
	"os"
	"fmt"
	// "flag"
	"gopkg.in/yaml.v3"

	log "github.com/sirupsen/logrus"

	// "github.com/BlankCanvasStudio/AutoScribe/pkg/ai"
	// "github.com/BlankCanvasStudio/AutoScribe/pkg/types"
)


var Settings Config = Config {
    Files: []string{},
    Directives: make(map[string]Directive),
}


func LoadConfig() error {
    // Source global configs
    err := LoadConfigFile(GlobalConfigFile)
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

    return nil
}


func LoadConfigFile(filename string) error {
	_, err := os.Stat(filename)

	if os.IsNotExist(err) {
            return nil

	} else if err != nil {
		return fmt.Errorf("failed to find config %v: %v", filename, err)
	}

	log.Infof("Loading config from %v", filename)

	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading config file: %v", err)
	}

        cfg := Config{ Files: []string{ filename } }
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("error parsing yaml: %v", err)
	}

        if Settings.Files == nil {
            Settings.Files = []string{ filename }
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
            Settings.Directives[name] = directive
        }

	return nil
}

/*
*
 * Parses command-line arguments to configure AutoScribe behavior.
 *
 * Sets flags for displaying AST, creating README, generating help menu, specifying project and output directories, editing files, targeting file extensions, setting log level, providing configuration files, adding prompts, and managing documentation.
 *
 * Validates the file extension format and updates the project directory if provided as a positional argument.
 *
 * If the debug flag is set, adjusts the log level to debug.
 *
 * @return error if the file extension is unsupported, otherwise nil.

*/
/*
func ParseCli() error {
	// Set the flags
	flag.StringVar(&AstFileName, "a", "", "Display the AST of a file")

	flag.BoolVar(&MakeReadme, "r", false, "Make a README.md for a project")

	flag.BoolVar(&MakeHelpMenuImpl, "m", false, "Make a help 'Menu' implementation for a project")
	flag.BoolVar(&MakeHelpMenuText, "mt", false, "Write the text of a help 'Menu' for a project")

	flag.StringVar(&ProjectDirectory, "d", "./", "Project directory to source files from")

	flag.StringVar(&OutputDirectory, "o", "./", "Project directory / file to save output into")

	flag.StringVar(&EditFile, "e", "", "Set the file you'd like AutoScribe to edit with new content")

	extPtr := flag.String("l", "sh", "Set the file extensions we should be targetting")

	flag.BoolVar(&LogLevelDebug, "debug", false, "Set log level to debug")

	flag.StringVar(&ConfigFile, "c", "/etc/autoscribe/autoscribe.conf", "Set the config file for AutoScribe")

	flag.StringVar(&AdditionalPrompt, "p", "", "Add additional instructions to the prompt generating your output")

	flag.BoolVar(&DocumentAst, "docs", false, "Document the functions of a package")

	flag.BoolVar(&UndocumentAst, "undocs", false, "Remove all the comments in a package")

	flag.Parse()

	if !types.IsSupportedFormat(*extPtr) {
		return fmt.Errorf("unsupported language format %v", *extPtr)
	}

	LanguageFileExtension = types.SupportedFormat(*extPtr)

	if len(flag.Args()) > 0 && ProjectDirectory == "./" {
		ProjectDirectory = flag.Arg(0)
	}

	if LogLevelDebug == true {
		log.SetLevel(log.DebugLevel)
	}

	return nil
}
*/

