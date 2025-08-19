# AutoScribe

AutoScribe is a command-line tool designed to automate the generation of project documentation and help menus by leveraging AI capabilities. It analyzes your project's codebase, build scripts, and related files to create comprehensive README files and help menu implementations, streamlining documentation efforts for developers.

## Features
- Automatically generates a detailed README.md describing project setup, dependencies, and usage.
- Creates help menu code implementations compatible with your project's programming language.
- Produces help menu text snippets suitable for command-line interfaces.
- Supports multiple scripting and programming formats, including shell scripts and Go files.

## Dependencies
- Go 1.16+ (for compiling and running the application)
- Git (for walking through the project directory)
- External Go modules:
  - [`github.com/sirupsen/logrus`](https://github.com/sirupsen/logrus)
  - [`gopkg.in/yaml.v3`](https://gopkg.in/yaml.v3)
  - [`github.com/openai/openai-go/v2`](https://github.com/openai/openai-go)

## Architecture
- `cmd/main.go`: Entry point for the application, orchestrates command execution based on CLI flags.
- `pkg/config`: Handles configuration loading from file or environment variables.
- `pkg/files`: Reads and formats project files for AI analysis.
- `pkg/files/selection`: Filters project files for code and build scripts.
- `pkg/openai/calls`: Contains functions for querying OpenAI's API to generate help menus and documentation.
- `pkg/types`: Defines supported formats and related utilities.
- `Makefile`: Builds the project and installs the executable.

## Installation

### Prerequisites
Ensure you have Go installed:
```sh
go version
```

### Build the Application
Clone the repository and build:
```sh
git clone <repository_url>
cd <repository_directory>
make
```

This will generate the `build/autoscribe` executable.

### Install
To install the executable system-wide:
```sh
sudo make install
```

It copies the binary to `/usr/local/bin/autoscribe` and ensures configuration files are in place.

## Configuration

### API Key
Set your OpenAI API key through environment variable:
```sh
export OPENAI_API_KEY="your-openai-api-key"
```

Alternatively, place your API key in `/etc/autoscribe/autoscribe.conf`:
```yaml
OPENAI_API_KEY: "your-openai-api-key"
```

### CLI Flags
You can customize behavior via command-line options:

| Flag | Description | Default | Example |
|---|---|---|---|
| `-r` | Generate a README.md for the project | false | `-r` |
| `-m` | Generate a help menu implementation | false | `-m` |
| `-mt` | Generate help menu text | true | `-mt` |
| `-d` | Specify the project directory | `./` | `-d ./myproject` |
| `-o` | Specify output directory | `./` | `-o ./output` |
| `-l` | Set file extension target (`sh`, `go`, etc.) | `sh` | `-l go` |
| `-debug` | Enable debug logging | false | `-debug` |
| `-c` | Path to config file | `/etc/autoscribe/autoscribe.conf` | `-c ./config.yaml` |

## Usage

### Generate README.md
```sh
./autoscribe -r -d /path/to/project
```

### Generate Help Menu Implementation
```sh
./autoscribe -m -d /path/to/project
```

### Generate Help Menu Text
```sh
./autoscribe -mt -d /path/to/project
```

## Examples

### Creating a README
```sh
./autoscribe -r -d ./myproject
```

### Generating Help Menu Code (e.g., in Go)
```sh
./autoscribe -m -d ./myproject -l go
```

### Generating Help Menu Text for Shell Scripts
```sh
./autoscribe -mt -d ./scripts -l sh
```

## Building and Installing
```sh
make
sudo make install
```

This will compile the application and copy the binary to `/usr/local/bin/autoscribe`.

## License
This project is open source and available under the MIT License.

---
**Note:** Replace `<repository_url>` and `<repository_directory>` with the actual URLs when deploying or using this project.
