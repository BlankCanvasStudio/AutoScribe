package config

import (
    "github.com/BlankCanvasStudio/AutoScribe/pkg/ai"
)

var GlobalConfigFile    string = "/etc/autoscribe/conf.yml"
var UserConfigFile      string = "~/.config/autoscribe/conf.yml"
var ProjectConfigFile   string = "./asb.yml"

// Default values
var DefaultModel ai.Model = ai.GPT_41_Nano

var DefaultDirective = TextDirective

var DefaultLocalDocs = "/opt/autoscribe/docs-database"

