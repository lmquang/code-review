# Code Review Tool

This is a command-line tool for automated code review using OpenAI's GPT model. It analyzes git diffs and provides AI-powered feedback on your code changes.

## Features

- Automated code review using OpenAI's GPT model
- Git integration for analyzing code changes
- Configurable ignore patterns for files and extensions
- Easy setup and configuration of OpenAI API key and model

## Installation

1. Ensure you have Go 1.20 or later installed on your system.
2. Clone the repository:
   ```
   git clone https://github.com/lmquang/code-review.git
   cd code-review
   ```
3. Set up your `GOBIN` environment variable if you haven't already:
   ```
   export GOBIN=$GOPATH/bin
   ```
4. Build the project:
   ```
   make build
   ```

## Usage

Before using the tool, make sure to set up your OpenAI API key. This key is required for the tool to function. You can set it up in two ways:

1. Set the `OPENAI_API_KEY` environment variable:
   ```
   export OPENAI_API_KEY='your-api-key'
   ```

2. Use the `set` command to configure the API key:
   ```
   code-review set -openai-api-key YOUR_API_KEY
   ```

To run a code review:

```
code-review review
```

Or use the shorthand:

```
code-review r
```

## Configuration

You can configure the OpenAI API key and model using the `set` command:

```
code-review set -openai-api-key YOUR_API_KEY
code-review set -openai-model MODEL_NAME
```

The configuration is stored in `~/.code-review.yaml`.

## Commands

- `set` or `s`: Set the OpenAI API Key and/or model
  - Flags:
    - `-openai-api-key`: Set the OpenAI API Key
    - `-openai-model`: Set the OpenAI Model

- `review` or `r`: Run the code review process
  - Flags:
    - `-ignore`: Comma-separated list of files or extensions to ignore (e.g., '*.yaml,*.json,docs.go')

## Project Structure

The project is organized as follows:

- `cmd/code-review/`: Contains the main application code
- `pkg/`: Contains the core packages used by the application
  - `diff/`: Handles diff formatting and processing
  - `git/`: Manages Git operations
  - `gpt/`: Interfaces with the OpenAI GPT model
- `Makefile`: Defines common development commands
- `go.mod` and `go.sum`: Go module files for dependency management

## Development

To run tests:

```
make test
```

To clean build files:

```
make clean
```

## License

[Insert appropriate license information here]

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.