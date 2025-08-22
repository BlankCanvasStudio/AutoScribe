# AutoScribe

AutoScribe is a command-line tool designed to analyze, document, and enhance your codebase automatically. It leverages OpenAI's language models to generate project documentation, help menus, and detailed code annotations, streamlining technical writing and code understanding for developers.

## Features
- Generate comprehensive `README.md` documentation reflecting project purpose, setup, and usage.
- Create help menus and usage texts based on your code and project structure.
- Parse Go source files to extract function definitions and generate in-code documentation.
- Analyze code to update existing documentation comments within source files.
- Detect and document function structures and handle cyclic dependency graphs in code.

## Dependencies
- Go (version >= 1.16 recommended)
- [github.com/sirupsen/logrus](https://github.com/sirupsen/logrus) for logging
- [golang.org/x/tools/go/packages](https://pkg.go.dev/golang.org/x/tools/go/packages) for package analysis
- [github.com/openai/openai-go/v2](https://github.com/openai/openai-go) for OpenAI API integration
- YAML support via [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3)

Ensure your environment variables contain your OpenAI API key (`OPENAI_API_KEY`) or set it within your config file.

## Installation

```bash
# Clone the repository
git clone https://github.com/BlankCanvasStudio/AutoScribe.git
cd AutoScribe

# Build the project
make
```

This will produce an executable at `build/autoscribe`.

## Configuration

- Configuration file: `/etc/autoscribe/autoscribe.conf`
- Environment variable: `OPENAI_API_KEY` (recommended if not using config file)

You can specify the project directory, output directory, and other options via CLI flags.

## Usage

### Generate README.md

```bash
./build/autoscribe -r -d /path/to/your/project
```

- `-r` or `--makeReadme`: Generate a `README.md` for the specified project.
- `-d` or `--directory`: Path to your project directory. Defaults to `./`.

### Create Help Menu Implementation

```bash
./build/autoscribe -m -d /path/to/your/project
```

- `-m` or `--makeHelpMenuImpl`: Generate a help menu implementation based on your code.

### Generate Help Menu Text

```bash
./build/autoscribe -mt -d /path/to/your/project
```

- `-mt` or `--makeHelpMenuText`: Generate a textual help menu output.

### Parse and Document a Single Source File

```bash
./build/autoscribe -a path/to/file.go
```

- `-a` or `--ast`: Parse the Go source file, extract functions, and generate documentation comments.

## CLI Flags Summary

| Flag | Description | Default | Example |
|-------|--------------|---------|---------|
| `-r` / `--makeReadme` | Generate README.md | false | `-r` |
| `-m` / `--makeHelpMenuImpl` | Generate help menu implementation | false | `-m` |
| `-mt` / `--makeHelpMenuText` | Generate help menu text | false | `-mt` |
| `-d` / `--directory` | Project directory | `./` | `-d /path/to/project` |
| `-a` / `--ast` | Parse specific Go source file for documentation | | `-a file.go` |
| `-c` | Config file path | `/etc/autoscribe/autoscribe.conf` | `-c ./myconfig.yaml` |
| `-p` | Additional prompt instructions for OpenAI | | `-p "Explain modules"` |
| `--debug` | Enable debug logging | false | `--debug` |

## Building from Source
Use the provided Makefile:

```bash
make
```

This builds the `autoscribe` binary in the `build/` directory.

You can install it system-wide:

```bash
sudo make install
```

which copies the binary to `/usr/local/bin/autoscribe` and the default config to `/etc/autoscribe/autoscribe.conf`.

---

For further customization or plugin development, refer to the source code in `pkg/` and `cmd/`.
