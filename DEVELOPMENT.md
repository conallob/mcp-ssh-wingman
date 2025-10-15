# Development Guide

This document provides guidance for developers working on MCP SSH Wingman.

## Prerequisites

- Go 1.21 or later
- tmux
- make (optional, but recommended)

## Getting Started

### Clone the repository

```bash
git clone https://github.com/conallob/mcp-ssh-wingman.git
cd mcp-ssh-wingman
```

### Build from source

```bash
# Build the binary (outputs to ./bin/mcp-ssh-wingman)
make build

# Build and install to /usr/local/bin
make install

# Or build directly with Go
go build -o bin/mcp-ssh-wingman ./cmd/mcp-ssh-wingman
```

## Development Workflow

### Running tests

```bash
# Run all tests
make test

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

### Code quality

```bash
# Format code
make fmt

# Run go vet
make vet

# Run both formatting and vetting
make lint
```

### Clean build artifacts

```bash
# Remove build artifacts
make clean

# Or using Go directly
go clean
```

## Project Structure

```
mcp-ssh-wingman/
├── cmd/
│   └── mcp-ssh-wingman/     # Main application entry point
│       └── main.go
├── internal/
│   ├── mcp/                 # MCP protocol type definitions
│   │   └── types.go         # JSON-RPC and MCP protocol types
│   ├── tmux/                # tmux session management
│   │   └── manager.go       # Session creation, content capture
│   └── server/              # MCP server implementation
│       └── server.go        # Request handling, tool/resource registration
├── examples/                # Configuration examples and test scripts
├── Makefile                 # Build automation
├── go.mod                   # Go module definition
├── go.sum                   # Go module checksums
├── README.md                # User-facing documentation
├── DEVELOPMENT.md           # This file
├── CLAUDE.md                # Claude Code guidance
└── LICENSE                  # BSD 3-Clause License
```

## Architecture

### Core Components

#### MCP Server (`internal/server/`)

The server implements the Model Context Protocol (MCP) using JSON-RPC 2.0 over stdin/stdout.

**Key responsibilities:**
- Protocol negotiation and initialization
- Tool and resource registration
- Request routing and handling
- Error handling and logging

**Supported MCP methods:**
- `initialize` - Protocol version negotiation
- `tools/list` - List available tools
- `tools/call` - Execute a tool
- `resources/list` - List available resources
- `resources/read` - Read a resource

#### tmux Manager (`internal/tmux/`)

Manages tmux session lifecycle and provides read-only access to terminal content.

**Key responsibilities:**
- Session creation and attachment
- Pane content capture
- Scrollback buffer reading
- Terminal metadata extraction

**Core functions:**
- `NewManager(sessionName)` - Create a new tmux manager
- `EnsureSession()` - Create or attach to a session
- `CapturePane()` - Read visible terminal content
- `CaptureScrollback(lines)` - Read scrollback history
- `GetTerminalInfo()` - Get terminal dimensions and metadata

#### MCP Protocol Types (`internal/mcp/`)

Type definitions for the MCP protocol (version 2024-11-05).

**Key types:**
- `Request` - JSON-RPC request structure
- `Response` - JSON-RPC response structure
- `Tool` - Tool definition with schema
- `Resource` - Resource definition
- `ToolResult` - Tool execution result

### Available Tools

#### `read_terminal`
Reads the current visible content from the tmux pane.

**Implementation:** Calls `tmux capture-pane -p`

#### `read_scrollback`
Reads historical terminal output from the scrollback buffer.

**Parameters:**
- `lines` (number): Number of lines to retrieve (default: 100)

**Implementation:** Calls `tmux capture-pane -p -S -<lines>`

#### `get_terminal_info`
Retrieves terminal metadata including dimensions and current working directory.

**Returns:**
- Width and height (columns/rows)
- Current working directory (via `tmux display-message -p '#{pane_current_path}'`)
- Pane ID and session info

### Available Resources

#### `terminal://current`
Provides current terminal content as a text resource.

**MIME type:** `text/plain`

#### `terminal://info`
Provides terminal metadata as JSON.

**MIME type:** `application/json`

## Building and Releasing

### Version Management

Version information is embedded at build time using Go's `-ldflags`:

```bash
go build -ldflags "-X main.version=v1.0.0" ./cmd/mcp-ssh-wingman
```

The Makefile automatically sets the version from git tags or defaults to `dev`.

### Release Process

Releases are automated using [GoReleaser](https://goreleaser.com/).

**Local release (dry-run):**
```bash
goreleaser release --snapshot --clean
```

**GitHub Actions:**
The project uses GitHub Actions to automatically build and publish releases when a new tag is pushed:

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

This will:
1. Build binaries for multiple platforms (macOS, Linux, FreeBSD)
2. Create a GitHub release with changelog
3. Update the Homebrew tap

## Testing

### Manual Testing

You can test the server manually using the MCP protocol:

```bash
# Start the server
./bin/mcp-ssh-wingman

# Send an initialize request (JSON-RPC)
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}

# List available tools
{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}

# Call read_terminal tool
{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"read_terminal","arguments":{}}}
```

### Integration Testing

Test with Claude Desktop by:
1. Building the binary: `make build`
2. Adding to Claude Desktop config
3. Restarting Claude Desktop
4. Using the MCP tools in a conversation

## Code Style and Conventions

- Follow standard Go formatting (`gofmt`)
- Use meaningful variable and function names
- Document exported functions and types
- Keep functions focused and small
- Handle errors explicitly
- Use structured logging where appropriate

## Common Development Tasks

### Adding a new tool

1. Define the tool schema in `internal/server/server.go`
2. Implement the tool handler function
3. Register the tool in the `handleToolsList` method
4. Add the handler to `handleToolsCall`
5. Update documentation

### Adding a new resource

1. Define the resource URI and metadata
2. Implement the resource reader function
3. Register in `handleResourcesList`
4. Add the reader to `handleResourcesRead`
5. Update documentation

### Debugging

Enable verbose logging by examining tmux command output:

```go
// In internal/tmux/manager.go
cmd := exec.Command("tmux", args...)
cmd.Stderr = os.Stderr  // Show tmux errors
```

## Contributing

When contributing:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests and linting (`make test lint`)
5. Commit your changes with clear commit messages
6. Push to your fork
7. Open a Pull Request

### Commit Message Guidelines

- Use the imperative mood ("Add feature" not "Added feature")
- Keep the first line under 72 characters
- Reference issues and PRs where appropriate
- Provide context in the commit body for non-trivial changes

## Resources

- [Model Context Protocol Specification](https://spec.modelcontextprotocol.io/)
- [Go Documentation](https://go.dev/doc/)
- [tmux Manual](https://man.openbsd.org/tmux)
- [JSON-RPC 2.0 Specification](https://www.jsonrpc.org/specification)

## License

BSD 3-Clause License - see [LICENSE](LICENSE) file for details.
