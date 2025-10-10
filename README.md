# AutoScribe

AutoScribe is a command-line tool designed to automate documentation and directive management within projects. It allows users to create, update, export, and manage custom directives and documentation scopes across multiple configuration files and project directories.

---

## Features

- Create and initialize custom directives from CLI commands
- Update directive properties and focus/ignore lists
- Export directives to various configuration files
- Remove documentation from files and folders
- Run directives to generate outputs
- Support for global, user, and local configuration scopes
- Extensible architecture for custom directive support

---

## Dependencies

- Go 1.23.11 or later
- Modules:
  - github.com/spf13/cobra
  - github.com/sirupsen/logrus
  - gopkg.in/yaml.v3
  - github.com/openai/openai-go/v2 (for OpenAI API integrations)

---

## Installation

### Build from Source

Ensure you have Go installed (version 1.23.11 or newer).

Clone the repository:

```bash
git clone https://github.com/BlankCanvasStudio/AutoScribe.git
cd AutoScribe
```

Build the project:

```bash
make
```

This creates the binary `build/autoscribe`.

### Install System-wide

```bash
sudo make install
```

This copies the binary to `/usr/local/bin/autoscribe` and sets up default configuration directories.

---

## Usage

### Basic Command Structure

```bash
autoscribe [command] [flags] [args]
```

### Global Flags

| Flag             | Short | Description                                              |
|------------------|--------|----------------------------------------------------------|
| `--debug`        | `-d`   | Enable verbose debug logging                             |
| `--global`       | `-g`   | Use global configuration scope                          |
| `--user`         | `-u`   | Use user-level configuration scope                        |
| `--local`        | `-l`   | Use local (project) configuration scope                 |
| `--config <file>`| `-c`   | Specify a custom configuration file                     |
| `--prompt <text>`| `-p`   | Add extra context to directives prompts                |

### Commands

#### Run All Scoped Directives

```bash
autoscribe run
```

Runs all directives within the specified scope.

#### Create a New Directive

```bash
autoscribe directive create <name> <prompt_file>
```

Creates a new directive with a given name and prompt file.

#### Initialize an Existing Directive

```bash
autoscribe directive init <directive_name> [files]
```

Initializes the directive into the current project, optionally specifying configuration files.

#### Export a Directive

```bash
autoscribe directive export <directive_name> [files]
```

Exports a directive to specified configuration files.

#### Update Directive Properties

Examples:

```bash
autoscribe directive kind <directive_name> <text>
autoscribe directive description <directive_name> <text>
autoscribe directive scope <directive_name> <text>
autoscribe directive model <directive_name> <text>
autoscribe directive output <directive_name> <text>
autoscribe directive apikey <directive_name> <text>
autoscribe directive local-docs <directive_name> <path>
```

Updates the respective property of a directive.

#### Add Focus or Ignore Items

```bash
autoscribe directive add focus <directive_name> <items>
autoscribe directive ignore <directive_name> <items>
```

Adds items to the focus or ignore list of a directive.

#### Manage Server Sources

```bash
autoscribe directive server <directive_name> <items>
```

Sets documentation source URLs for a directive.

#### Remove Documentation

```bash
autoscribe undoc [files / folders]
```

Removes documentation from specified files or folders.

---

## Example Usage

Create a new directive with a prompt file:

```bash
autoscribe directive create summarize ./prompts/summarize.txt
```

Initialize an existing directive:

```bash
autoscribe directive init summarize --config myconfig.yml
```

Export a directive to config files:

```bash
autoscribe directive export summarize --config global.yml
```

Update directive description:

```bash
autoscribe directive description summarize "This directive summarizes text content"
```

Run all directives in scope:

```bash
autoscribe run --global
```

Remove documentation from files:

```bash
autoscribe undoc ./src ./docs
```

---

## Customization

You can extend functionality by adding new directives in the `config/` directory or by creating custom command handlers through the CLI interface. The architecture is designed to facilitate easy integration and augmentation.

---

## License

This project is licensed under the MIT License. See the LICENSE file for details.

---

## Support & Contributions

Contributions are welcome! Please open an issue or pull request on the GitHub repository. For questions, contact the maintainer via the GitHub issues page.

---

## Notes

- The configuration system supports multiple layers: global, user, and local.
- The CLI commands are built with cobra and support subcommands and flags for flexible operation.
- Ensure your configuration files are properly formatted YAML files matching the expected schema.

---

*Happy documenting!*