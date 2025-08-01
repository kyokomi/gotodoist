# gotodoist

[![CI](https://github.com/kyokomi/gotodoist/actions/workflows/ci.yml/badge.svg)](https://github.com/kyokomi/gotodoist/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/kyokomi/gotodoist)](https://goreportcard.com/report/github.com/kyokomi/gotodoist)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Reference](https://pkg.go.dev/badge/github.com/kyokomi/gotodoist.svg)](https://pkg.go.dev/github.com/kyokomi/gotodoist)

A powerful command-line interface tool for managing your Todoist tasks, built with Go.

[日本語版 README はこちら](README_ja.md)

## Features

- 📝 **Task Management**: List, add, update, and delete tasks
- 📁 **Project Management**: Organize tasks with projects
- ⚙️ **Configuration Management**: Easy setup and configuration
- 🌍 **Multi-language Support**: English and Japanese
- 🚀 **Fast & Lightweight**: Built with Go for optimal performance
- 🔒 **Secure**: Local configuration with API token protection

## Installation

### Download Binary

Download the latest release from the [releases page](https://github.com/kyokomi/gotodoist/releases).

### Build from Source

```bash
git clone https://github.com/kyokomi/gotodoist.git
cd gotodoist
make build
```

### Using Go Install

```bash
go install github.com/kyokomi/gotodoist@latest
```

## Quick Start

### 1. Get Your API Token

1. Go to [Todoist Integrations](https://todoist.com/prefs/integrations)
2. Copy your API token

### 2. Configuration

#### Option A: Environment Variable
```bash
export TODOIST_API_TOKEN="your-api-token-here"
```

#### Option B: Configuration File
```bash
gotodoist config init
```

This will create a configuration file at `~/.config/gotodoist/config.yaml`.

### 3. Start Using

```bash
# List all tasks
gotodoist task list

# Add a new task
gotodoist task add "Buy groceries"

# List all projects
gotodoist project list

# Add a new project
gotodoist project add "Work Projects"
```

## Usage

### Task Commands

```bash
# List all tasks
gotodoist task list

# Add a new task
gotodoist task add "Task content"

# Update a task
gotodoist task update <task-id> --content "New content"

# Delete a task
gotodoist task delete <task-id>

# Complete a task
gotodoist task complete <task-id>
```

### Project Commands

```bash
# List all projects
gotodoist project list

# Add a new project
gotodoist project add "Project name"

# Update a project
gotodoist project update <project-id> --name "New name"

# Delete a project
gotodoist project delete <project-id>
```

### Configuration Commands

```bash
# Initialize configuration
gotodoist config init

# Show current configuration
gotodoist config show

# Set language preference
gotodoist config set language en  # or ja
```

### Global Options

```bash
# Enable verbose output
gotodoist --verbose task list

# Enable debug mode
gotodoist --debug task list

# Set language for single command
gotodoist --lang ja task list
```

## Configuration

### Configuration File Location

- **Linux/macOS**: `~/.config/gotodoist/config.yaml`
- **Windows**: `%APPDATA%\gotodoist\config.yaml`

### Configuration Options

```yaml
api_token: "your-todoist-api-token"
base_url: "https://api.todoist.com/rest/v2"
language: "en"  # en or ja
```

### Environment Variables

- `TODOIST_API_TOKEN`: Your Todoist API token
- `GOTODOIST_LANG`: Language preference (en/ja)

## Development

### Prerequisites

- Go 1.24 or later
- Make (optional, for using Makefile commands)

### Building

```bash
# Build the application
make build

# Run tests
make test

# Run tests with coverage
make coverage

# Format code
make fmt

# Run linter
make lint

# Check for vulnerabilities
make vuln
```

### Available Make Commands

Run `make help` to see all available commands:

```bash
make help
```

### Project Structure

```
gotodoist/
├── cmd/           # CLI command definitions
├── internal/      # Internal packages
│   ├── api/       # Todoist API client
│   ├── config/    # Configuration management
│   └── i18n/      # Internationalization
├── locales/       # Translation files
├── .github/       # GitHub Actions workflows
├── Makefile       # Build automation
├── go.mod         # Go module definition
└── main.go        # Application entry point
```

## Contributing

We welcome contributions! Please feel free to submit a Pull Request.

### Development Process

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/#issue-number-description`
3. Make your changes
4. Run tests: `make test`
5. Run linter: `make lint`
6. Commit your changes with a descriptive message
7. Push to your fork and submit a Pull Request

### Commit Message Format

```
<type>: <description> (#<issue-number>)

Types: feat, fix, docs, refactor, test, chore
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Links

- [Todoist API Documentation](https://developer.todoist.com/rest/v2/)
- [Issue Tracker](https://github.com/kyokomi/gotodoist/issues)
- [Releases](https://github.com/kyokomi/gotodoist/releases)

## Author

[@kyokomi](https://github.com/kyokomi)