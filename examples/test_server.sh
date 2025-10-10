#!/bin/bash

# Example script to test the MCP server manually
# This script sends JSON-RPC requests to test the server functionality

SESSION_NAME=${1:-"mcp-wingman"}

echo "Starting MCP SSH Wingman test..."
echo "Using tmux session: $SESSION_NAME"
echo ""

# Build the server
echo "Building server..."
make build

# Start the server in the background
echo "Starting server..."
bin/mcp-ssh-wingman --session "$SESSION_NAME" &
SERVER_PID=$!

# Give it time to start
sleep 1

# Test initialize
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0.0"}}}' | bin/mcp-ssh-wingman --session "$SESSION_NAME" 2>/dev/null &

# Cleanup
trap "kill $SERVER_PID 2>/dev/null" EXIT

echo ""
echo "Server running with PID: $SERVER_PID"
echo "Press Ctrl+C to stop"

wait $SERVER_PID
