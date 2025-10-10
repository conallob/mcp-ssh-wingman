# MCP SSH Wingman

A Model Context Protocol (MCP) server that provides read-only access to Unix shell prompts via tmux. This enables AI assistants like Claude to safely observe terminal environments without executing commands.

## Features

- 🔒 **Read-only access** - Observe terminal content without execution risks
- 🖥️ **tmux integration** - Leverages tmux's session management for reliable terminal access
- 📜 **Scrollback history** - Access historical terminal output
- 📊 **Terminal metadata** - Retrieve dimensions, current path, and session info
- 🔌 **MCP protocol** - Standard protocol for AI assistant integration

## Prerequisites

- Go 1.21 or later
- tmux

## Installation

### Build from source

```bash
# Clone the repository
git clone https://github.com/conall-obrien/mcp-ssh-wingman.git
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

# Use a custom tmux session name
./bin/mcp-ssh-wingman --session my-session
```

### Integration with Claude Desktop

Add the server to your Claude Desktop configuration file:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "ssh-wingman": {
      "command": "/usr/local/bin/mcp-ssh-wingman",
      "args": ["--session", "mcp-wingman"]
    }
  }
}
```

Restart Claude Desktop after updating the configuration.

## Available Tools

The server exposes the following MCP tools:

### `read_terminal`
Read the current terminal content from the tmux session.

```json
{
  "name": "read_terminal"
}
```

### `read_scrollback`
Read scrollback history from the tmux session.

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

## Available Resources

### `terminal://current`
Current terminal content as a text resource.

### `terminal://info`
Terminal metadata and information.

## How It Works

The server creates or attaches to a tmux session and uses tmux's built-in commands to safely read terminal content:

1. **Session Management**: Creates/attaches to a detached tmux session
2. **Content Capture**: Uses `tmux capture-pane` to read visible content
3. **Read-Only**: Never sends keystrokes or commands to the session
4. **MCP Protocol**: Exposes terminal content via standard MCP tools and resources

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

## License

BSD 3-Clause License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.
