# AutoScribe

AutoScribe is a command-line tool designed to facilitate automation and management of code documentation, directives, and configurations. It provides functionalities to create, initialize, update, export, and document code directives and manage configuration files across different scopes (global, user, local).

---

## Purpose

The project enables developers and teams to:
- Manage custom directives with various attributes (kind, description, prompt, model, etc.).
- Generate and update documentation across codebases.
- Export and initialize directives based on existing configurations.
- Remove documentation from files and folders.
- Handle configurations across global, user, and local scopes.
- Seamlessly integrate documentation workflows into development processes.

---

## Installation

### Build from Source
Ensure you have Go 1.23.11 or later installed.

```bash
git clone https://github.com/BlankCanvasStudio/AutoScribe.git
cd AutoScribe
make
```

### Install via Makefile
```bash
make all
```

This will generate the `autoscribe` binary in the `build/` directory.

### System-wide Installation
Once built, copy the binary to a directory in your `$PATH`:

```bash
sudo cp build/autoscribe /usr/local/bin/
```

### Dependencies
- Go modules:
  - github.com/spf13/cobra
  - github.com/sirupsen/logrus
  - gopkg.in/yaml.v3
  - github.com/openai/openai-go/v2

---

## Usage

Run `autoscribe` with the desired command and subcommand options.

```bash
autoscribe --help
```

### Basic Commands

- **Run all directives in scope:**
  ```bash
  autoscribe run
  ```

- **Display version:**
  ```bash
  autoscribe version
  ```

### Example: Create a New Directive

```bash
autoscribe directive create MyDirective prompts/my_prompt.txt
```

### Example: Initialize Directive in Current Project

```bash
autoscribe directive init MyDirective
```

### Example: Export a Directive

```bash
autoscribe directive export MyDirective
```

### Example: Document Files

```bash
autoscribe undoc ./src
```

---

## Command-Line Options

The tool provides several persistent flags to control scope and configuration behavior:

| Option | Shortcut | Description | Default |
|---------|------------|----------------|---------|
| `--debug` | `-d` | Enable debug logging | false |
| `--global` | `-g` | Use global configuration scope | false |
| `--user` | `-u` | Use user configuration scope | false |
| `--local` | `-l` | Use local (current folder) scope | false |
| `--prompt` | `-p` | Add additional context to directive prompts | "" |
| `--config` | `-c` | Specify custom configuration file | "" |

These flags influence how configuration files are selected or created, and how directives operate.

---

## Architecture & Dependencies

The project is organized into several key packages:

- `pkg/cli` — Command-line interface setup and command execution.
- `pkg/cli/directives` — Management of custom directives including creation, initialization, export, and updates.
- `pkg/cli/docs` — Documentation management tasks like documenting and removing documentation.
- `pkg/config` — Handling configuration files across different scopes (global, user, local) including loading and saving configurations.
- `pkg/types` — Definitions for directives, configuration structures, and related types.

**Dependencies:**

- `cobra` for CLI parsing and command management.
- `logrus` for structured logging.
- `yaml.v3` for configuration serialization.
- `openai` for AI-driven documentation generation (implied).

---

## Usage Examples

### Running with Debugging Enabled:
```bash
autoscribe --debug run
```

### Creating a New Directive:
```bash
autoscribe directive create EncryptFunction prompts/encrypt_prompt.txt
```

### Initializing a Directive from Current Config:
```bash
autoscribe directive init EncryptFunction
```

### Exporting a Directive to Configs:
```bash
autoscribe directive export EncryptFunction
```

### Removing Documentation from Files:
```bash
autoscribe undoc ./src
```

---

## Notes

- Commands are flexible and can be customized with flags.
- Directives support attributes that can be programmatically updated via subcommands.
- Configuration management ensures that settings are scoped appropriately and persisted across different files.

---

## Contribution

Contributions are welcome. Please fork the repository, implement enhancements or fixes, and submit pull requests.

---

## License

This project is licensed under the MIT License.

---

**Happy documenting!**