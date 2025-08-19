package config;

import (
    "os"
    "fmt"
    "flag"
    "gopkg.in/yaml.v3"

    log "github.com/sirupsen/logrus"

    "github.com/BlankCanvasStudio/AutoScribe/pkg/types"
)

type Config struct {
    OPENAI_API_KEY string `yaml:"OPENAI_API_KEY"`
}

var ConfigFile            string = "/etc/autoscribe/autoscribe.conf"

var OpenAIKey             string

var ProjectDirectory      string                = "./"
var OutputDirectory       string                = "./"
var EditFile             string                = ""
var LanguageFileExtension types.SupportedFormat = "sh"
var MakeReadme            bool                  = false
var MakeHelpMenuImpl      bool                  = true
var MakeHelpMenuText      bool                  = true

var LogLevelDebug         bool                  = false

var AdditionalPrompt      string                = ""


func LoadConfig() error {

    _, err := os.Stat(ConfigFile)
        
    if os.IsNotExist(err) {
        var exists bool
        OpenAIKey, exists = os.LookupEnv("OPENAI_API_KEY")                
        if ! exists {
            return fmt.Errorf("failed to strip OpenAI API key out of the env. doesn't exist")
        }
        return nil

    } else if err != nil {
        return fmt.Errorf("failed to check for config %v: %v", ConfigFile, err)
    }

    log.Infof("Loading config from %v", ConfigFile)

    data, err := os.ReadFile(ConfigFile)
    if err != nil {
        return fmt.Errorf("error reading config file: %v", err)
    }

    var cfg Config
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return fmt.Errorf("error parsing yaml: %v", err)
    }

    OpenAIKey = cfg.OPENAI_API_KEY

    return nil
}


func ParseCli() error {
    // Set the flags
    flag.BoolVar(&MakeReadme, "r", false, "Make a README.md for a project")

    flag.BoolVar(&MakeHelpMenuImpl, "m", false,  "Make a help 'Menu' implementation for a project")
    flag.BoolVar(&MakeHelpMenuText, "mt", false, "Write the text of a help 'Menu' for a project")


    flag.StringVar(&ProjectDirectory, "d", "./", "Project directory to source files from")

    flag.StringVar(&OutputDirectory, "o", "./", "Project directory / file to save output into")

    flag.StringVar(&EditFile, "e", "", "Set the file you'd like AutoScribe to edit with new content")

    extPtr := flag.String("l", "sh", "Set the file extensions we should be targetting")

    flag.BoolVar(&LogLevelDebug, "debug", false, "Set log level to debug")

    flag.StringVar(&ConfigFile, "c", "/etc/autoscribe/autoscribe.conf", "Set the config file for AutoScribe")

    flag.StringVar(&AdditionalPrompt, "p", "", "Add additional instructions to the prompt generating your output")

    flag.Parse()

    if ! types.IsSupportedFormat(*extPtr) {
        return fmt.Errorf("unsupported language format %v", *extPtr)
    }

    LanguageFileExtension = types.SupportedFormat(*extPtr)

    if len(flag.Args()) > 0 && ProjectDirectory == "./" {
        ProjectDirectory = flag.Arg(0)
    }

    if LogLevelDebug == true { 
        log.SetLevel(log.DebugLevel); 
    }

    return nil
}
