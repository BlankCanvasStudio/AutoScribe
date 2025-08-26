# AutoScribe

AutoScribe is a command-line tool designed to assist developers with code documentation generation and project scaffolding. It automates the creation of README files, help menus, and manages Abstract Syntax Tree (AST) documentation for Go projects, streamlining your development workflow.

## Features
- Generate a comprehensive `README.md` for your project.
- Create help menu implementation code and text.
- Parse and document Go source files' AST, including updating doc comments.
- Undocument directories or files based on AST structure.

## Dependencies
- Go (version 1.16+ recommended)
- [logrus](https://github.com/sirupsen/logrus) for logging

## Installation

### Prerequisites
Ensure you have Go installed. You can download it from [golang.org](https://golang.org/dl/).

### Clone the repository
```bash
git clone https://github.com/BlankCanvasStudio/AutoScribe.git
cd AutoScribe
```

### Build the project
```bash
go build -o autoscribe ./cmd
```

The binary `autoscribe` will be created in your directory.

## Usage

### Basic command
```bash
./autoscribe [flags]
```

### Configuration
AutoScribe loads configuration from environment variables, configuration files, or CLI flags. Below are the available command-line flags:

| Flag                        | Description                                                      | Default                         |
|------------------------------|------------------------------------------------------------------|---------------------------------|
| `--make-readme`             | Generate `README.md` for the project.                             | false                           |
| `--make-help-menu-impl`     | Generate help menu implementation code.                          | false                           |
| `--make-help-menu-text`     | Generate help menu text.                                         | false                           |
| `--ast-file`                | Path to the AST file or project directory to process.           | ""                              |
| `--undocument-ast`          | Remove documentation from specified AST directory/file.         | false                           |
| `--document-ast`            | Generate documentation from AST.                                  | false                           |
| `--language-file-ext`       | File extension associated with the language (e.g., `.go`).     | `.go`                           |
| `--project-dir`             | Path to the project directory for README and other files.       | Current directory (`./`)        |

### Example commands

- Generate README and documentation for a Go project:
  ```bash
  ./autoscribe --make-readme --document-ast --ast-file ./myproject
  ```

- Remove documentation from a directory:
  ```bash
  ./autoscribe --undocument-ast --ast-file ./myproject
  ```

- Generate help menu:
  ```bash
  ./autoscribe --make-help-menu-text
  ```

## Configuration

- The tool automatically loads configuration settings, which can also be overridden via CLI flags.
- Environment variables can be used to set configurations (`AUTO_DIR`, `AUTO_EXT`, etc.).

## Architecture

- **Main Orchestrator**: Parses configuration, CLI args, and routes tasks accordingly.
- **AST Processing**: Parses Go package files, updates documentation or removes it.
- **Documentation Generation**: Creates README files, help menu code, and help menu text via the `calls` package.

## Notes
- Make sure the `AstFileName` points to the correct source files or directories.
- The tool is designed to work primarily with Go source files with `.go` extensions.
- Error handling is verbose; review logs for troubleshooting.

## License
This project is licensed under the MIT License.

---

*For more details, refer to the source code or visit the project's repository.*