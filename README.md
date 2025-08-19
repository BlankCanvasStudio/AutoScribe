# AutoScribe

AutoScribe is a command-line tool designed to analyze your project files and generate helpful documentation and code components automatically by leveraging AI language models. It can create README files, help menus, and implementation templates tailored for your codebase, streamlining the documentation and development process.

## Purpose

The primary goal of AutoScribe is to assist developers in documenting and enhancing their projects by programmatically generating README files, help menus, and code snippets that fit seamlessly with the existing code structure. It utilizes OpenAI’s models to understand project contents and produce relevant outputs.

## Dependencies

- Go (version 1.18+ recommended)
- Make (for building the project)
- External Go packages:
  - github.com/sirupsen/logrus
  - github.com/openai/openai-go/v2
  - github.com/BlankCanvasStudio/AutoScribe/pkg/types

## Architecture Overview

- **cmd/main.go**: Entry point for the application, managing command-line arguments and orchestrating actions.
- **pkg/config/load.go**: Loads configuration from file or environment.
- **pkg/files/formatting.go**: Handles file content concatenation and formatting.
- **pkg/files/selection.go**: Filters project files based on language or build criteria.
- **pkg/openai/calls/helpmenu.go**: Contains functions to generate help menu code and text via AI.
- **pkg/openai/calls/readme.go**: Creates a comprehensive README file using AI.
- **pkg/types/types.go**: Defines supported formats and utilities for file extension handling.
- **Makefile**: Provides build and install commands.

## Installation

### Prerequisites

- Ensure Go is installed:
```bash
https://golang.org/dl/
```

### Building from Source

Clone the repository:

```bash
git clone https://github.com/BlankCanvasStudio/AutoScribe.git
cd AutoScribe
```

Build the project:

```bash
make
```

This will compile the executable `autoscribe` inside the `build/` directory.

### Installing

To install the autoScribe binary system-wide:

```bash
sudo make install
```

This copies the executable to `/usr/local/bin/autoscribe` and sets up the default configuration file in `/etc/autoscribe/autoscribe.conf`.

## Configuration

1. Edit the configuration file `/etc/autoscribe/autoscribe.conf` to set your OpenAI API key:

```yaml
OPENAI_API_KEY: "your-openai-api-key"
```

2. Alternatively, set the environment variable:

```bash
export OPENAI_API_KEY="your-openai-api-key"
```

3. Customize other options by passing command-line flags (see Usage below).

## Usage

Run the tool with desired options:

```bash
# Generate a README.md for the current project
autoscibe -r

# Generate a help menu implementation
autoscibe -m

# Generate only help menu text
autoscibe -mt

# Specify project directory
autoscibe -d ./myproject

# Specify output directory
autoscibe -o ./output

# Specify file to edit or generate
autoscibe -e ./file.go

# Set language file extension (default is 'sh')
autoscibe -l go

# Enable debug logging
autoscibe -debug
```

### Examples

- Generate a README based on current directory project files:

```bash
autoscibe -r
```

- Generate a help menu implementation for the project:

```bash
autoscibe -m
```

- Update a file with a new help menu implementation

```bash
autoscibe -m -e ./main.go
```

- Generate only the help menu text that describes commands and options:

```bash
autoscibe -mt
```

- Output files to a specified directory:

```bash
autoscibe -d ./myproject -o ./docs
```

- Set the language format to Go:

```bash
autoscibe -l go
```

## Important Notes

- The tool relies on the OpenAI API; ensure your API key has sufficient quota.
- It processes project files based on the specified or default language extension.
- Generated contents are based on AI understanding and may require manual review.

## License

This project is licensed under the MIT License. See `LICENSE` for details.

## Contact

For issues or contributions, please open an issue on the GitHub repository:
[https://github.com/BlankCanvasStudio/AutoScribe](https://github.com/BlankCanvasStudio/AutoScribe)
