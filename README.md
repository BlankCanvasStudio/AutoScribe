# AutoScribe

AutoScribe is an automated code documentation and assistance tool designed to analyze your project files, generate comprehensive README files, help menus, and documentation snippets using OpenAI's GPT models. It streamlines understanding and documenting complex codebases with minimal manual effort.

## Features

- Generate detailed `README.md` files explaining project setup, usage, dependencies, and architecture.
- Create implementation code for help menus within your projects.
- Generate concise help menu text akin to `--help` outputs.
- Supports multiple programming languages and formats, including shell scripts and Go.

## Dependencies

- Go 1.16+  
- External Go modules:
  - [sirupsen/logrus](https://github.com/sirupsen/logrus)
  - [openai/openai-go](https://github.com/openai/openai-go)
  - [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3)

### Installing Dependencies

```bash
go mod tidy
```

## Installation

Clone the repository and build the project:

```bash
git clone <repository-url> autoScribe
cd autoScribe
make
```

### Makefile Commands

- **build/autoscribe**: Builds the AutoScribe binary.
- **install**: Installs the binary and configuration files (requires root privileges).

## Usage

### Basic Command-Line Options

```bash
./build/autoscribe [options] [project_directory]
```

### Options

| Flag               | Description                                                                                              | Effect                                                       |
|--------------------|----------------------------------------------------------------------------------------------------------|--------------------------------------------------------------|
| `-r`               | Generate a `README.md` for the project                                                                   | Creates a detailed README.md                                |
| `-m`               | Generate a help menu implementation (`Help Menu`) for the project                                        | Produces code that implements a help menu                  |
| `-mt`              | Generate only the help menu text (summary like `--help`)                                                 | Produces help menu text only                              |
| `-d <dir>`         | Specify the project directory to source files from                                                       | Defaults to `./`                                           |
| `-o <file/dir>`    | Specify output directory or filename for generated files                                                  | Defaults to `./`                                           |
| `-e <file>`        | Edit a specific file with new content (used internally)                                                  | Optional                                                   |
| `-l <format>`      | Target specific file format (`sh`, `go`, etc.)                                                            | Defaults to `sh`                                           |
| `-c <config file>` | Path to the config file for AutoScribe                                                                   | Defaults to `/etc/autoscribe/autoscribe.conf`             |
| `-p <prompt>`      | Additional instructions added to AI prompt                                                              | Useful for customizing behavior                            |
| `-debug`           | Set log level to debug                                                                                    | Useful for troubleshooting                                 |

### Examples

Generate a README for your current directory:

```bash
./build/autoscribe -r
```

Generate a help menu implementation and display it:

```bash
./build/autoscribe -m -d ./myproject
```

Generate only the help menu text:

```bash
./build/autoscribe -mt
```

### Environment Variables

- `OPENAI_API_KEY`: Set your OpenAI API key as an environment variable if not specified in the config file:

```bash
export OPENAI_API_KEY="your-api-key-here"
```

## Architecture & Codebase

- **Main Entry**: `cmd/main.go` manages command-line parsing and orchestrates actions.
- **Configuration**: Loaded from a YAML config file or environment variables in `pkg/config/load.go`.
- **AI Calls**: Made through OpenAI's API in `pkg/openai/calls/`, supporting functions for generating help menus, README, and prompts.
- **File Handling**: Files are processed and formatted in `pkg/files/formatting.go` with filtering mechanisms in `pkg/files/selection.go`.
- **Supported Formats**: Defined in `pkg/types/types.go`, supporting scripting languages like shell and Go.

## Building and Installing

```bash
make
sudo make install
```

This compiles the binary and copies it to `/usr/local/bin/autoscribe`.

## Customization

- To adjust prompts or add support for other languages, modify the relevant prompt templates or extend the supported formats in `pkg/types/types.go`.
- For advanced usage, customize configuration files specified in `/etc/autoscribe/autoscribe.conf`.

---

For further details or contributions, please refer to the repository documentation or contact the maintainers.