# Example Claude Desktop Configuration for Screen Users

This directory contains example configuration files for different screen setups.

## Basic Screen Configuration

For users with a single screen session:

```json
{
  "mcpServers": {
    "ssh-wingman": {
      "command": "/usr/local/bin/mcp-ssh-wingman",
      "args": ["--terminal", "screen", "--session", "main"]
    }
  }
}
```

## Multi-Window Screen Configuration

For users who want to target a specific window in their screen session:

```json
{
  "mcpServers": {
    "ssh-wingman": {
      "command": "/usr/local/bin/mcp-ssh-wingman",
      "args": ["--terminal", "screen", "--session", "main", "--window", "1"]
    }
  }
}
```

## Usage Tips for Screen Users

### 1. Find Your Screen Session Name
```bash
screen -ls
```

### 2. List Windows in Your Session
Once connected via MCP, use the `list_windows` tool to see all available windows.

### 3. Switch Between Windows
Use the `set_window` tool to switch to different windows:
```json
{
  "name": "set_window",
  "arguments": {
    "window_id": "2"
  }
}
```

### 4. Access Scrollback History
Use the `read_scrollback` tool to access your screen's backscroll:
```json
{
  "name": "read_scrollback",
  "arguments": {
    "lines": 1000
  }
}
```

## Screen Session Setup

If you don't have a screen session running, you can create one:

```bash
# Create a new detached screen session
screen -dmS main

# Or attach to create windows
screen -S main
# Then use Ctrl-A c to create new windows
# Use Ctrl-A d to detach
```

## Benefits for Screen Users

- **Preserve Your Workflow**: Use your existing screen setup without changes
- **Multiple Remote Connections**: Perfect for users who connect to multiple servers through screen windows
- **Rich Scrollback**: Access all your historical output stored in screen's backscroll buffer
- **Window Management**: Easily navigate between different terminal environments
