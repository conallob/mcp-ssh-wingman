# MCP SSH Wingman

A Model Context Protocol (MCP) server that provides read-only access to Unix shell prompts via tmux or GNU screen. This enables AI assistants like Claude to safely observe terminal environments without executing commands.

## Features

- 🔒 **Read-only access** - Observe terminal content without execution risks
- 🖥️ **tmux & screen integration** - Leverages tmux or GNU screen session management for reliable terminal access
- 📺 **Multiple window support** - Access different windows/panes within your terminal sessions
- 📜 **Scrollback history** - Access historical terminal output
- 📊 **Terminal metadata** - Retrieve dimensions, current path, and session info
- 🔌 **MCP protocol** - Standard protocol for AI assistant integration

## Prerequisites

- Go 1.21 or later
- tmux (for tmux support)
- GNU screen (for screen support)

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap conallob/tap
brew install mcp-ssh-wingman
```

### Pre-built binaries

Download pre-built binaries from the [releases page](https://github.com/conallob/mcp-ssh-wingman/releases).

Available for:
- macOS (arm64/amd64)
- Linux (arm64/amd64)
- FreeBSD (arm64/amd64)

### Build from source

```bash
# Clone the repository
git clone https://github.com/conallob/mcp-ssh-wingman.git
cd mcp-ssh-wingman

# Build
make build

# Install to /usr/local/bin (optional)
make install
```

## Usage

### Running the server

```bash
# Start with default tmux session name (mcp-wingman)
./bin/mcp-ssh-wingman

# Use tmux with a custom session name
./bin/mcp-ssh-wingman --terminal tmux --session my-session

# Use GNU screen with default session name
./bin/mcp-ssh-wingman --terminal screen

# Use GNU screen with custom session name and specific window
./bin/mcp-ssh-wingman --terminal screen --session my-screen --window 2
```

### Integration with Claude Desktop

Add the server to your Claude Desktop configuration file:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`

**For tmux:**
```json
{
  "mcpServers": {
    "ssh-wingman": {
      "command": "/usr/local/bin/mcp-ssh-wingman",
      "args": ["--terminal", "tmux", "--session", "mcp-wingman"]
    }
  }
}
```

**For GNU screen:**
```json
{
  "mcpServers": {
    "ssh-wingman": {
      "command": "/usr/local/bin/mcp-ssh-wingman",
      "args": ["--terminal", "screen", "--session", "mcp-wingman"]
    }
  }
}
```

**For screen with specific window:**
```json
{
  "mcpServers": {
    "ssh-wingman": {
      "command": "/usr/local/bin/mcp-ssh-wingman",
      "args": ["--terminal", "screen", "--session", "my-screen", "--window", "2"]
    }
  }
}
```

Restart Claude Desktop after updating the configuration.

## Available Tools

The server exposes the following MCP tools:

### `read_terminal`
Read the current terminal content from the tmux/screen session.

```json
{
  "name": "read_terminal"
}
```

### `read_scrollback`
Read scrollback history from the tmux/screen session.

```json
{
  "name": "read_scrollback",
  "arguments": {
    "lines": 100
  }
}
```

### `get_terminal_info`
Get information about the terminal (dimensions, current path, etc.).

```json
{
  "name": "get_terminal_info"
}
```

### `list_windows`
List all windows/panes in the current session.

```json
{
  "name": "list_windows"
}
```

### `set_window`
Set the active window/pane for subsequent operations.

```json
{
  "name": "set_window",
  "arguments": {
    "window_id": "2"
  }
}
```

## Available Resources

### `terminal://current`
Current terminal content as a text resource.

### `terminal://info`
Terminal metadata and information.

## How It Works

The server creates or attaches to a tmux/screen session and uses their built-in commands to safely read terminal content:

1. **Session Management**: Creates/attaches to a detached tmux or screen session
2. **Content Capture**: Uses `tmux capture-pane` or `screen hardcopy` to read visible content
3. **Multiple Windows**: Can switch between different windows/panes within the session
4. **Read-Only**: Never sends keystrokes or commands to the session
5. **MCP Protocol**: Exposes terminal content via standard MCP tools and resources

### Screen-Specific Features

For GNU screen users, the implementation provides:
- **Existing Session Support**: Attach to your existing screen session with all your windows
- **Window Navigation**: List and switch between different screen windows
- **Backscroll Access**: Access your screen's scrollback buffer history
- **Multi-Window Workflow**: Perfect for users who run local screen with multiple remote connections

## Development

```bash
# Run tests
make test

# Format code
make fmt

# Lint code
make vet

# Clean build artifacts
make clean
```

## Architecture

- `cmd/mcp-ssh-wingman/` - Main application entry point
- `internal/mcp/` - MCP protocol type definitions
- `internal/tmux/` - tmux session management
- `internal/server/` - MCP server implementation
- `examples/` - Configuration examples and test scripts


## Features to Add

- [zellij](https://github.com/zellij-org/zellij/) , once https://github.com/zellij-org/zellij/issues/4348 is implemented

## License

BSD 3-Clause License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.
