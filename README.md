# AutoScribe

AutoScribe is an automation tool designed to analyze, document, and generate project artifacts such as README files and help menus. It leverages OpenAI's API to assist in code and documentation generation, providing a flexible command-line interface to customize its behavior.

## Features

- Generate `README.md` files for projects
- Create help menus and menu implementations
- Document functions within a code package
- Remove documentation comments from code
- Display the Abstract Syntax Tree (AST) of files
- Customize target file extensions and directories
- Configure via YAML config files or environment variables
- Add custom prompts for AI processing

## Dependencies

- Go (version 1.16+ recommended)
- `gopkg.in/yaml.v3`
- `github.com/sirupsen/logrus`
- Supported file formats as specified in `pkg/types`

## Installation

### Prerequisites

- Install Go from [https://golang.org/dl/](https://golang.org/dl/)

### Building the Application

Clone this repository and build the binary:

```bash
git clone <repository_url>
cd <repository_directory>
go build -o autoscribe ./cmd/autoscribe
```

*Note: Replace `<repository_url>` and `<repository_directory>` with the actual repository URL and directory.*

### Configuration

Create or edit the configuration file at `/etc/autoscribe/autoscribe.conf` with content similar to:

```yaml
OPENAI_API_KEY: your-openai-api-key
```

Alternatively, set the environment variable:

```bash
export OPENAI_API_KEY=your-openai-api-key
```

## Usage

### Basic Commands

```bash
./autoscribe [options] [project_directory]
```

### Common Options

| Option | Description | Example |
| --- | --- | --- |
| `-r` | Generate a `README.md` for the project | `./autoscribe -r` |
| `-m` | Create a help menu implementation | `./autoscribe -m` |
| `-mt` | Write help menu text | `./autoscribe -mt` |
| `-a` | Display the AST of a specific file | `./autoscribe -a src/main.go` |
| `-d` | Set the project directory (default: `./`) | `./autoscribe -d ./myproject` |
| `-o` | Output directory/file for generated artifacts | `./autoscribe -o ./output` |
| `-e` | Specify a file to edit with new content | `./autoscribe -e src/main.go` |
| `-l` | Target file extension (default: `sh`) | `./autoscribe -l py` |
| `-debug` | Enable debug logging | `./autoscribe -debug` |
| `-c` | Path to configuration file | `./autoscribe -c ./config.yml` |
| `-p` | Additional instructions prompt | `./autoscribe -p "Write detailed docs"` |
| `-docs` | Document functions within a package | `./autoscribe -docs` |
| `-undocs` | Remove all comments from package | `./autoscribe -undocs` |

### Example Usage

Generate a README for a project in the current directory:

```bash
./autoscribe -r
```

Display the AST of a specific file:

```bash
./autoscribe -a src/utils.go
```

Create a help menu implementation:

```bash
./autoscribe -m
```

## Architecture & Code Structure

This project is built using Go modules, with core configuration logic in `pkg/config/load.go`. The main command-line interface parses arguments and invokes appropriate functionalities, which interact with OpenAI's API for code analysis and generation.

## License

This project is licensed under the MIT License. See `LICENSE` for more details.