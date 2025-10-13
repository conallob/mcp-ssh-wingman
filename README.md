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

### Integration with Gemini CLI

Add this section to your Gemini config file (`~/.gemini/settings.json`):

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

Run the `gemini` CLI and make sure it can see the MCP server:

```shell
> /mcp list

Configured MCP servers:

🟢 ssh-wingman - Ready (3 tools)
  Tools:
  - get_terminal_info
  - read_scrollback
  - read_terminal
```

Test with a prompt:

```shell
> using ssh-wingman MCP Server, what do you see in my session?

 ╭────────────────────────────────────────────────────────────────────────────────────────╮
 │ ?  read_terminal (ssh-wingman MCP Server) {} ←                                         │
 │                                                                                        │
 │   MCP Server: ssh-wingman                                                              │
 │   Tool: read_terminal                                                                  │
 │                                                                                        │
 │ Allow execution of MCP tool "read_terminal" from server "ssh-wingma…                   │
 │                                                                                        │
 │ ● 1. Yes, allow once                                                                   │
 │   2. Yes, always allow tool "read_terminal" from server "ssh-wingma…                   │
 │   3. Yes, always allow all tools from server "ssh-wingman"                             │
 │   4. No, suggest changes (esc)                                                         │
 │                                                                                        │
 ╰────────────────────────────────────────────────────────────────────────────────────────╯
⠏ Waiting for user confirmation...
```

You might as well select option 3 there.

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
