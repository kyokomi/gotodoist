# gotodoist

[![CI](https://github.com/kyokomi/gotodoist/actions/workflows/ci.yml/badge.svg)](https://github.com/kyokomi/gotodoist/actions/workflows/ci.yml)
[![codecov](https://codecov.io/github/kyokomi/gotodoist/graph/badge.svg?token=cGdi7YkLjv)](https://codecov.io/github/kyokomi/gotodoist)
[![Go Report Card](https://goreportcard.com/badge/github.com/kyokomi/gotodoist)](https://goreportcard.com/report/github.com/kyokomi/gotodoist)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Reference](https://pkg.go.dev/badge/github.com/kyokomi/gotodoist.svg)](https://pkg.go.dev/github.com/kyokomi/gotodoist)

A powerful command-line interface tool for managing your Todoist tasks, built with Go.

[日本語版 README はこちら](README_ja.md)

## Features

- 📝 **Task Management**: List, add, update, complete, and delete tasks
- 📁 **Project Management**: Organize tasks with projects
- 🏷️ **Label Support**: Categorize tasks with labels
- 📅 **Due Date Management**: Set and manage task deadlines
- 🔄 **Offline Support**: Work offline with local sync
- 🚀 **Fast & Lightweight**: Optimized for speed and efficiency

## Installation

### Download Binary

Download the latest binary for your platform from the [releases page](https://github.com/kyokomi/gotodoist/releases).

### Using Go Install

```bash
go install github.com/kyokomi/gotodoist@latest
```

## Quick Start

### 1. Get Your API Token

1. Visit [Todoist Integrations](https://todoist.com/prefs/integrations)
2. Copy your API token from the "API token" section

### 2. Set Up Authentication

```bash
# Option 1: Environment variable (recommended)
export TODOIST_API_TOKEN="your-api-token-here"

# Option 2: Configuration file
gotodoist config init
```

### 3. Start Using

```bash
# Sync with Todoist (recommended for first use)
gotodoist sync

# List all tasks
gotodoist task list

# Add a new task
gotodoist task add "Buy groceries"

# Add task to a specific project
gotodoist task add "Write report" -p "Work"
```

## Core Commands

### Task Management

```bash
# List tasks
gotodoist task list                          # All active tasks
gotodoist task list -p "Work"                # Tasks in "Work" project
gotodoist task list -f "p1"                  # Priority 1 tasks
gotodoist task list -f "@important"          # Tasks with "important" label
gotodoist task list -a                       # All tasks (including completed)

# Add tasks
gotodoist task add "Task content"
gotodoist task add "Important task" -P 1     # With priority (1-4)
gotodoist task add "Meeting" -d "tomorrow"   # With due date
gotodoist task add "Call client" -p "Work" -l "urgent,calls"  # With project and labels

# Update tasks
gotodoist task update <task-id> -c "New content"
gotodoist task update <task-id> -P 2         # Change priority
gotodoist task update <task-id> -d "next monday"  # Change due date

# Complete/Uncomplete tasks
gotodoist task complete <task-id>
gotodoist task uncomplete <task-id>

# Delete tasks
gotodoist task delete <task-id>
gotodoist task delete <task-id> -f           # Skip confirmation
```

### Project Management

```bash
# List projects
gotodoist project list
gotodoist project list -v                    # Verbose (with IDs)

# Add projects
gotodoist project add "New Project"

# Update projects
gotodoist project update <project-id> --name "Updated Name"

# Delete projects
gotodoist project delete <project-id>
gotodoist project delete <project-id> -f     # Skip confirmation
```

### Synchronization

```bash
# Sync with Todoist
gotodoist sync                               # Sync all data
gotodoist sync init                          # Initial full sync
gotodoist sync status                        # Check sync status
gotodoist sync reset -f                      # Reset local data
```

### Configuration

```bash
# Initialize config
gotodoist config init

# View current config
gotodoist config show

```

## Configuration Options

Configuration file location:
- **Linux/macOS**: `~/.config/gotodoist/config.yaml`
- **Windows**: `%APPDATA%\gotodoist\config.yaml`

### Environment Variables

- `TODOIST_API_TOKEN`: Your Todoist API token

## Tips and Examples

### Filter Tasks by Priority
```bash
# High priority tasks
gotodoist task list -f "p1"

# Medium and low priority
gotodoist task list -f "p3 | p4"
```

### Work with Labels
```bash
# Add task with multiple labels
gotodoist task add "Review PR" -l "code-review,urgent"

# Filter by label
gotodoist task list -f "@code-review"
```

### Due Date Examples
```bash
# Natural language dates
gotodoist task add "Submit report" -d "next friday"
gotodoist task add "Team meeting" -d "every monday"

# Specific dates
gotodoist task add "Birthday party" -d "2024-12-25"
```

### Combining Filters
```bash
# High priority tasks in Work project
gotodoist task list -p "Work" -f "p1"

# Tasks due today with urgent label
gotodoist task list -f "today & @urgent"
```

## Troubleshooting

### Common Issues

1. **"API token not found" error**
   - Ensure `TODOIST_API_TOKEN` is set or run `gotodoist config init`

2. **"Project not found" error**
   - Use `gotodoist project list` to see available projects
   - Project names are case-sensitive

3. **Sync issues**
   - Run `gotodoist sync` to refresh local data
   - Use `gotodoist sync reset -f` if data seems corrupted

## Contributing

Contributions are welcome! Please check the [issues page](https://github.com/kyokomi/gotodoist/issues) for areas where you can help.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Links

- [Todoist API Documentation](https://developer.todoist.com/rest/v2/)
- [Report Issues](https://github.com/kyokomi/gotodoist/issues)
- [Releases](https://github.com/kyokomi/gotodoist/releases)

## Author

[@kyokomi](https://github.com/kyokomi)
