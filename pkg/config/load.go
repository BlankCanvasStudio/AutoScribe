package config

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/BlankCanvasStudio/AutoScribe/pkg/types"
)

type Config struct {
	OPENAI_API_KEY             string `yaml:"OPENAI_API_KEY"`
        DocsPrompt                 string `yaml:"/etc/autoscribe/prompts/docs.txt"`
        ReadmePrompt               string `yaml:"/etc/autoscribe/prompts/readme.txt"`
        HelpMenuPromptText         string `yaml:"/etc/autoscribe/prompts/helpmenu/generate-text.txt"`
        HelpMenuPromptCodeExample  string `yaml:"/etc/autoscribe/prompts/helpmenu/print-implementation.txt"`
        HelpMenuPromptCodeUpdate   string `yaml:"/etc/autoscribe/prompts/helpmenu/update-implementation.txt"`
}

var ConfigFile string = "/etc/autoscribe/conf.yaml"

var OpenAIKey string

var ProjectDirectory string = "./"
var OutputDirectory string = "./"
var EditFile string = ""
var LanguageFileExtension types.SupportedFormat = "sh"
var MakeReadme bool = false
var MakeHelpMenuImpl bool = false
var MakeHelpMenuText bool = false
var AstFileName string = ""
var DocumentAst bool = false
var UndocumentAst bool = false

var LogLevelDebug bool = false

var AdditionalPrompt string = ""


// Prompt files
var DocsPrompt                 string = "/etc/autoscribe/prompts/docs.txt"
var ReadmePrompt               string = "/etc/autoscribe/prompts/readme.txt"
var HelpMenuPromptText         string = "/etc/autoscribe/prompts/helpmenu/generate-text.txt"
var HelpMenuPromptCodeExample  string = "/etc/autoscribe/prompts/helpmenu/print-implementation.txt"
var HelpMenuPromptCodeUpdate   string = "/etc/autoscribe/prompts/helpmenu/update-implementation.txt"



/*
*
 * LoadConfig checks for the existence of a configuration file and loads its contents.
 * If the file does not exist, it attempts to retrieve the OpenAI API key from environment variables.
 * Returns an error if reading or parsing the config fails or if the environment variable is missing.
*
 * Parameters:
 * - None
 *
 * Returns:
 * - error: Indicates success (nil) or failure with a description.
 *
 * Errors/Exceptions:
 * - Fails if reading the config file or parsing its YAML content.
 * - Fails if the environment variable "OPENAI_API_KEY" is missing when the config file does not exist.

*/
func LoadConfig() error {
    // Source global configs
    err := LoadConfigFile(ConfigFile)
    if err != nil {
        return err
    }

    // Prefer local configs
    err = LoadConfigFile("go.asb")
    if err != nil {
        return err
    }

    // Grab OpenAI key from env if it doesn't exist
    if OpenAIKey == "" {
	var exists bool
	OpenAIKey, exists = os.LookupEnv("OPENAI_API_KEY")
	if !exists {
	    return fmt.Errorf("failed to strip OpenAI API key out of the env. doesn't exist")
	}
    }

    return nil
}


func LoadConfigFile(filename string) error {
	_, err := os.Stat(filename)

	if os.IsNotExist(err) {
            return nil

	} else if err != nil {
		return fmt.Errorf("failed to check for config %v: %v", filename, err)
	}

	log.Infof("Loading config from %v", filename)

	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading config file: %v", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("error parsing yaml: %v", err)
	}

	OpenAIKey = cfg.OPENAI_API_KEY

        if cfg.DocsPrompt != "" {
            DocsPrompt = cfg.DocsPrompt
        }

        if cfg.ReadmePrompt != "" {
            ReadmePrompt = cfg.ReadmePrompt
        }

        if cfg.HelpMenuPromptText != "" {
            HelpMenuPromptText = cfg.HelpMenuPromptText
        }

        if cfg.HelpMenuPromptCodeExample != "" {
            HelpMenuPromptCodeExample = cfg.HelpMenuPromptCodeExample
        }

        if cfg.HelpMenuPromptCodeUpdate != "" {
            HelpMenuPromptCodeUpdate = cfg.HelpMenuPromptCodeUpdate
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
