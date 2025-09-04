# AutoScribe

AutoScribe is a command-line tool designed for automating code documentation and project management tasks. It supports defining custom directives to generate, update, or export documentation and configuration snippets, making it ideal for streamlining development workflows.

---

## Features

- Load and manage multiple configuration scopes (global, user, local, custom)
- Define and create custom directives for automation
- Generate README files based on code annotations
- Create and document project ASTs
- Support for dynamic CLI directives with flexible scope management
- Extensible architecture with plugin-like directives

---

## Requirements

- Go 1.23.11 or higher
- Dependencies:
  - `github.com/spf13/cobra`
  - `github.com/sirupsen/logrus`
  - `github.com/openai/openai-go/v2`
  - `gopkg.in/yaml.v3`

## Dependencies Installation

```bash
# Ensure dependencies are downloaded
go mod tidy
```

## Building the Application

```bash
# Build the executable
make
```

This will generate a binary named `build/autoscribe`.

## Installation

```bash
# Copy the binary to your PATH
sudo cp build/autoscribe /usr/local/bin/autoscribe
```

You can now run `autoscribe` from any directory.

---

## Usage

### Basic CLI Commands

```bash
# Run all directives in scope
autoscribe run
```

### Common Options

- `--debug`: Enable verbose debug logging
- `-g`, `--global`: Use global configuration scope
- `-u`, `--user`: Use user configuration scope
- `-l`, `--local`: Use local (current folder) configuration scope (default)
- `--config <file>`: Specify a custom configuration file

### Example: Running all directives

```bash
autoscribe --debug run
```

### Example: Creating a new directive

```bash
autoscribe directive create myDirective "/path/to/prompt.txt"
```

This saves a new custom directive named `myDirective` with the specified prompt file in the selected configuration scope.

### Example: Export a directive

```bash
autoscribe directive export myDirective
```

Exports the directive to specified configuration files based on scope flags.

### Example: Generate Readme / Help Menus

The tool can generate README, help menu implementation, or help text based on configuration flags, during main execution (see code comments). Uncomment relevant sections in `main.go` to activate.

---

## Extending with Custom Directives

Create custom directives using the `directive create` command, then customize behavior with flags such as:

- `kind <text>`: Set directive kind (e.g., "file", "recursion")
- `description <text>`: Set directive description
- `short <text>`: Set short description
- `prompt <file>`: Set prompt file path
- `model <model>`: Specify ML model
- `output <file>`: Save output to specified file
- `apikey <key>`: Set API key
- `local-docs <path>`: Path for local documentation sources

These can be combined with `add`, `ignore`, and `server` array commands for targeted modifications.

---

## Maintenance & Contribution

- Fork and clone the repository
- Implement features or fixes
- Submit pull requests
- Support issues via GitHub

---

## License

MIT License. See `LICENSE` for details.

---

For detailed instructions and advanced usages, refer to the source code and inline comments within the project files.

---
**Enjoy automating your documentation with AutoScribe!**