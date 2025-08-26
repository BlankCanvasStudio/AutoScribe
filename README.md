# AutoScribe

AutoScribe is a command-line tool designed to facilitate automatic documentation generation and management for Go projects. It can generate comprehensive README files, produce help menu implementations or texts, and parse or undocument Abstract Syntax Trees (ASTs) within your codebase. By integrating with your code, AutoScribe streamlines documentation workflows and enhances project clarity.

## Features

- Generate README.md based on project files
- Create help menu implementations or plain text
- Parse AST files for documentation or undocumentation
- Document functions within parsed packages
- Update documentation directly within source files

## Prerequisites

- Go 1.16 or higher
- Modules dependencies specified below

## Dependencies

The project uses the following Go modules:

- `github.com/sirupsen/logrus` for logging

Make sure to initialize your environment with all dependencies:

```bash
go mod tidy
```

## Installation

Clone the repository and build the binary:

```bash
git clone <repository-url>
cd <repository-directory>
go build -o autoscribe ./cmd
```

This will produce an executable named `autoscribe`.

## Usage

```bash
./autoscribe [flags]
```

### Configuration Flags

- `--make-readme`  
  Generate a `README.md` file for the project.

- `--make-help-menu-impl`  
  Create a Help Menu implementation file.

- `--make-help-menu-text`  
  Generate a plain text version of the Help Menu.

- `--ast-file-name <path>`  
  Specify the AST file to parse or modify.

- `--undocument-ast`  
  Remove documentation from the specified AST files.

- `--document-ast`  
  Document functions within the AST files.

### Example

Generate README and document AST:

```bash
./autoscribe --make-readme --ast-file-name path/to/ast.go --document-ast
```

Generate Help Menu implementation:

```bash
./autoscribe --make-help-menu-impl
```

## Architecture & Data Flow

- Loads configuration from files and environment variables.
- Parses CLI arguments to determine actions.
- Based on flags, generates documentation resources or processes AST files.
- Can document or remove documentation from code, updating source files directly.
- Utilizes internal packages for AST parsing (`pkg/ast`), configuration management (`pkg/config`), and OpenAI interactions (`pkg/openai/calls`).

## Contributing

Contributions are welcome! Please fork the repository and submit pull requests.

## License

This project is licensed under the MIT License.