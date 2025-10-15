# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

MCP SSH Wingman is an MCP (Model Context Protocol) Server written in Go that provides read-only access to Unix shell prompts via tmux. This allows AI assistants to safely observe terminal environments without executing commands.

## Documentation

- **[README.md](README.md)** - User-facing documentation including installation, usage, and configuration
- **[DEVELOPMENT.md](DEVELOPMENT.md)** - Developer documentation including build instructions, architecture, and contribution guidelines

## Quick Reference

### Prerequisites
- Go 1.21 or later
- tmux (for terminal session management)

### Common Commands
```bash
# Build the binary
make build

# Build and install to /usr/local/bin
make install

# Run tests
make test

# Format and vet code
make lint

# Clean build artifacts
go clean
```

### Running the Server
```bash
# Run with default session name (mcp-wingman)
./bin/mcp-ssh-wingman

# Run with custom tmux session name
./bin/mcp-ssh-wingman --session my-session

# Show version
./bin/mcp-ssh-wingman --version
```

## Project Structure

- `cmd/mcp-ssh-wingman/` - Main entry point
- `internal/mcp/` - MCP protocol types and definitions
- `internal/tmux/` - tmux session management
- `internal/server/` - MCP server implementation
- `examples/` - Configuration examples and test scripts

## Core Components

**MCP Server (internal/server/)**
- Implements JSON-RPC 2.0 protocol for MCP communication
- Handles initialize, tools/list, tools/call, resources/list, resources/read methods
- Uses stdin/stdout for communication with MCP clients

**tmux Manager (internal/tmux/)**
- Creates and manages tmux sessions for read-only terminal access
- Captures pane content and scrollback history
- Retrieves terminal metadata (dimensions, current path)

**MCP Protocol (internal/mcp/)**
- Type definitions for MCP protocol version 2024-11-05
- JSON-RPC request/response structures
- Tool and resource schemas

## Available Tools

1. `read_terminal` - Read current terminal content
2. `read_scrollback` - Read scrollback history (configurable line count)
3. `get_terminal_info` - Get terminal dimensions and metadata

## Available Resources

1. `terminal://current` - Current terminal content
2. `terminal://info` - Terminal information and metadata

## Claude Desktop Integration

Add to Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):
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

## License

BSD 3-Clause License (see LICENSE file)
