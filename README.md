# MCP SSH Wingman

[![Test](https://github.com/conallob/mcp-ssh-wingman/workflows/Test/badge.svg)](https://github.com/conallob/mcp-ssh-wingman/actions)
[![codecov](https://codecov.io/gh/conallob/mcp-ssh-wingman/branch/main/graph/badge.svg)](https://codecov.io/gh/conallob/mcp-ssh-wingman)
[![Go Version](https://img.shields.io/badge/Go-1.24%2B-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/License-BSD%203--Clause-blue.svg)](LICENSE)
[![Buy Me A Coffee](https://img.shields.io/badge/Buy%20Me%20A%20Coffee-support-yellow.svg?style=flat&logo=buy-me-a-coffee)](https://www.buymeacoffee.com/conallob)

A Model Context Protocol (MCP) server that provides read-only access to Unix shell prompts via `tmux`. This enables AI assistants like Claude to safely observe terminal environments without executing commands.

## Features

- 🔒 **Read-only access** - Observe terminal content without execution risks
- 🖥️ **tmux integration** - Leverages tmux's session management for reliable terminal access
- 📜 **Scrollback history** - Access historical terminal output
- 📊 **Terminal metadata** - Retrieve dimensions, current path, and session info
- 🔌 **MCP protocol** - Standard protocol for AI assistant integration

## FAQ

### Will you be adding [GNU `screen`](https://www.gnu.org/software/screen/) support?

Not in the immediate future. Although `screen` can be used for pair programming/debugging, it does not have any mechanism to enforce read only mode.

Since the read-only safeguard is the main difference between this MCP and other [MCPs with ssh functionality](https://www.gnu.org/software/screen/) , screen will need to support read-only sessions before it can be integrated

### Will you be adding [zellij](https://github.com/zellij-org/zellij/) support?

In the future, possibly. https://github.com/zellij-org/zellij/issues/4348 needs to be implemented first, to ensure `zellij` has read-only functionality, just like `tmux`

## Prerequisites

- tmux (for terminal session management)
- (Optional) Go 1.21 or later (only needed for building from source)

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

See [DEVELOPMENT.md](DEVELOPMENT.md) for build instructions.

## Usage

### Running the server

```bash
# Start with default tmux session name (mcp-wingman)
mcp-ssh-wingman

# Use a custom tmux session name
mcp-ssh-wingman --session my-session

# Show version
mcp-ssh-wingman --version
```

### Integration with Claude Desktop

Add the server to your Claude Desktop configuration file:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Linux**: `~/.config/Claude/claude_desktop_config.json`

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

**Example:**
```json
{
  "name": "read_terminal"
}
```

### `read_scrollback`

Read scrollback history from the tmux session.

**Parameters:**
- `lines` (number): Number of lines to retrieve from scrollback buffer

**Example:**
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

**Example:**
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

This approach ensures that the AI assistant can observe terminal activity without any risk of accidentally executing commands or modifying the terminal state.

## Use Cases

- Monitor long-running processes and build outputs
- Observe terminal-based application behavior
- Debug issues by reviewing terminal history
- Provide context-aware assistance based on current terminal state

## Security Considerations

MCP SSH Wingman is designed with security in mind:

- **Read-only**: The server never sends input to the terminal
- **Local access**: Operates on local tmux sessions only
- **No command execution**: Cannot execute shell commands
- **Isolated sessions**: Each session is independent and sandboxed by tmux

## Troubleshooting

### Server won't start
- Ensure tmux is installed: `tmux -V`
- Check that the binary has execute permissions: `chmod +x /usr/local/bin/mcp-ssh-wingman`

### Claude Desktop can't connect
- Verify the path to the binary in your configuration file
- Check Claude Desktop logs for error messages
- Ensure the binary runs successfully from command line first

### Can't see terminal content
- Verify the tmux session exists: `tmux list-sessions`
- Ensure the session name matches the one specified in configuration
- Check that the tmux session has active panes


## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

For development guidelines, see [DEVELOPMENT.md](DEVELOPMENT.md).
