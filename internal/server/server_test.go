package server

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/conall-obrien/mcp-ssh-wingman/internal/mcp"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		name        string
		sessionName string
	}{
		{
			name:        "with session name",
			sessionName: "test-session",
		},
		{
			name:        "with empty session name",
			sessionName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &bytes.Buffer{}
			writer := &bytes.Buffer{}
			srv := NewServer(tt.sessionName, reader, writer)

			if srv == nil {
				t.Fatal("NewServer() returned nil")
			}
			if srv.tmuxManager == nil {
				t.Error("NewServer() tmuxManager is nil")
			}
			if srv.reader == nil {
				t.Error("NewServer() reader is nil")
			}
			if srv.writer == nil {
				t.Error("NewServer() writer is nil")
			}
		})
	}
}

func TestServer_handleRequest_Initialize(t *testing.T) {
	srv := NewServer("test-session", &bytes.Buffer{}, &bytes.Buffer{})

	request := &mcp.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	response := srv.handleRequest(request)

	if response == nil {
		t.Fatal("handleRequest() returned nil")
	}
	if response.JSONRPC != "2.0" {
		t.Errorf("response.JSONRPC = %v, want 2.0", response.JSONRPC)
	}
	if response.ID != request.ID {
		t.Errorf("response.ID = %v, want %v", response.ID, request.ID)
	}
	if response.Error != nil {
		t.Errorf("response.Error = %v, want nil", response.Error)
	}
	if response.Result == nil {
		t.Fatal("response.Result is nil")
	}

	// Verify the result can be marshaled to InitializeResult
	resultBytes, err := json.Marshal(response.Result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	var initResult mcp.InitializeResult
	if err := json.Unmarshal(resultBytes, &initResult); err != nil {
		t.Fatalf("Failed to unmarshal InitializeResult: %v", err)
	}

	if initResult.ProtocolVersion != ProtocolVersion {
		t.Errorf("initResult.ProtocolVersion = %v, want %v", initResult.ProtocolVersion, ProtocolVersion)
	}
	if initResult.ServerInfo.Name != ServerName {
		t.Errorf("initResult.ServerInfo.Name = %v, want %v", initResult.ServerInfo.Name, ServerName)
	}
}

func TestServer_handleRequest_ToolsList(t *testing.T) {
	srv := NewServer("test-session", &bytes.Buffer{}, &bytes.Buffer{})

	request := &mcp.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}

	response := srv.handleRequest(request)

	if response == nil {
		t.Fatal("handleRequest() returned nil")
	}
	if response.Error != nil {
		t.Errorf("response.Error = %v, want nil", response.Error)
	}
	if response.Result == nil {
		t.Fatal("response.Result is nil")
	}

	// Verify the result structure
	resultBytes, err := json.Marshal(response.Result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	var toolsResult mcp.ListToolsResult
	if err := json.Unmarshal(resultBytes, &toolsResult); err != nil {
		t.Fatalf("Failed to unmarshal ListToolsResult: %v", err)
	}

	if len(toolsResult.Tools) == 0 {
		t.Error("toolsResult.Tools is empty, expected at least one tool")
	}

	// Verify expected tools are present
	expectedTools := map[string]bool{
		"read_terminal":     false,
		"read_scrollback":   false,
		"get_terminal_info": false,
	}

	for _, tool := range toolsResult.Tools {
		if _, ok := expectedTools[tool.Name]; ok {
			expectedTools[tool.Name] = true
		}
	}

	for toolName, found := range expectedTools {
		if !found {
			t.Errorf("Expected tool %q not found in tools list", toolName)
		}
	}
}

func TestServer_handleRequest_ResourcesList(t *testing.T) {
	srv := NewServer("test-session", &bytes.Buffer{}, &bytes.Buffer{})

	request := &mcp.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "resources/list",
	}

	response := srv.handleRequest(request)

	if response == nil {
		t.Fatal("handleRequest() returned nil")
	}
	if response.Error != nil {
		t.Errorf("response.Error = %v, want nil", response.Error)
	}
	if response.Result == nil {
		t.Fatal("response.Result is nil")
	}

	// Verify the result structure
	resultBytes, err := json.Marshal(response.Result)
	if err != nil {
		t.Fatalf("Failed to marshal result: %v", err)
	}

	var resourcesResult mcp.ListResourcesResult
	if err := json.Unmarshal(resultBytes, &resourcesResult); err != nil {
		t.Fatalf("Failed to unmarshal ListResourcesResult: %v", err)
	}

	if len(resourcesResult.Resources) == 0 {
		t.Error("resourcesResult.Resources is empty, expected at least one resource")
	}

	// Verify expected resources are present
	expectedResources := map[string]bool{
		"terminal://current": false,
		"terminal://info":    false,
	}

	for _, resource := range resourcesResult.Resources {
		if _, ok := expectedResources[resource.URI]; ok {
			expectedResources[resource.URI] = true
		}
	}

	for uri, found := range expectedResources {
		if !found {
			t.Errorf("Expected resource %q not found in resources list", uri)
		}
	}
}

func TestServer_handleRequest_UnknownMethod(t *testing.T) {
	srv := NewServer("test-session", &bytes.Buffer{}, &bytes.Buffer{})

	request := &mcp.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      4,
		Method:  "unknown/method",
	}

	response := srv.handleRequest(request)

	if response == nil {
		t.Fatal("handleRequest() returned nil")
	}
	if response.Result != nil {
		t.Error("response.Result should be nil for unknown method")
	}
	if response.Error == nil {
		t.Fatal("response.Error is nil, expected error for unknown method")
	}
	if response.Error.Code != -32601 {
		t.Errorf("response.Error.Code = %v, want -32601 (Method not found)", response.Error.Code)
	}
	if !strings.Contains(response.Error.Message, "unknown/method") {
		t.Errorf("response.Error.Message = %v, should contain method name", response.Error.Message)
	}
}

func TestServer_callTool_ReadTerminal(t *testing.T) {
	srv := NewServer("test-session", &bytes.Buffer{}, &bytes.Buffer{})

	request := &mcp.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      5,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      "read_terminal",
			"arguments": map[string]interface{}{},
		},
	}

	response := srv.handleRequest(request)

	if response == nil {
		t.Fatal("handleRequest() returned nil")
	}

	// The response might have an error if tmux session doesn't exist
	// but we should still get a valid response structure
	if response.JSONRPC != "2.0" {
		t.Errorf("response.JSONRPC = %v, want 2.0", response.JSONRPC)
	}

	// If there's a result, verify it's a CallToolResult
	if response.Result != nil {
		resultBytes, err := json.Marshal(response.Result)
		if err != nil {
			t.Fatalf("Failed to marshal result: %v", err)
		}

		var toolResult mcp.CallToolResult
		if err := json.Unmarshal(resultBytes, &toolResult); err != nil {
			t.Fatalf("Failed to unmarshal CallToolResult: %v", err)
		}

		if len(toolResult.Content) == 0 {
			t.Error("toolResult.Content is empty")
		}
	}
}

func TestServer_callTool_ReadScrollback(t *testing.T) {
	srv := NewServer("test-session", &bytes.Buffer{}, &bytes.Buffer{})

	tests := []struct {
		name      string
		arguments map[string]interface{}
	}{
		{
			name:      "with default lines",
			arguments: map[string]interface{}{},
		},
		{
			name: "with specific lines (float64)",
			arguments: map[string]interface{}{
				"lines": float64(50),
			},
		},
		{
			name: "with specific lines (int)",
			arguments: map[string]interface{}{
				"lines": 75,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &mcp.JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      6,
				Method:  "tools/call",
				Params: map[string]interface{}{
					"name":      "read_scrollback",
					"arguments": tt.arguments,
				},
			}

			response := srv.handleRequest(request)

			if response == nil {
				t.Fatal("handleRequest() returned nil")
			}

			// Should get a valid response even if session doesn't exist
			if response.JSONRPC != "2.0" {
				t.Errorf("response.JSONRPC = %v, want 2.0", response.JSONRPC)
			}
		})
	}
}

func TestServer_callTool_GetTerminalInfo(t *testing.T) {
	srv := NewServer("test-session", &bytes.Buffer{}, &bytes.Buffer{})

	request := &mcp.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      7,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      "get_terminal_info",
			"arguments": map[string]interface{}{},
		},
	}

	response := srv.handleRequest(request)

	if response == nil {
		t.Fatal("handleRequest() returned nil")
	}

	// Should get a valid response even if session doesn't exist
	if response.JSONRPC != "2.0" {
		t.Errorf("response.JSONRPC = %v, want 2.0", response.JSONRPC)
	}
}

func TestServer_callTool_UnknownTool(t *testing.T) {
	srv := NewServer("test-session", &bytes.Buffer{}, &bytes.Buffer{})

	request := &mcp.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      8,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      "unknown_tool",
			"arguments": map[string]interface{}{},
		},
	}

	response := srv.handleRequest(request)

	if response == nil {
		t.Fatal("handleRequest() returned nil")
	}
	if response.Result != nil {
		t.Error("response.Result should be nil for unknown tool")
	}
	if response.Error == nil {
		t.Fatal("response.Error is nil, expected error for unknown tool")
	}
	if !strings.Contains(response.Error.Message, "unknown tool") {
		t.Errorf("response.Error.Message = %v, should mention unknown tool", response.Error.Message)
	}
}

func TestServer_readResource_Current(t *testing.T) {
	srv := NewServer("test-session", &bytes.Buffer{}, &bytes.Buffer{})

	request := &mcp.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      9,
		Method:  "resources/read",
		Params: map[string]interface{}{
			"uri": "terminal://current",
		},
	}

	response := srv.handleRequest(request)

	if response == nil {
		t.Fatal("handleRequest() returned nil")
	}

	// Should get a response even if session doesn't exist
	if response.JSONRPC != "2.0" {
		t.Errorf("response.JSONRPC = %v, want 2.0", response.JSONRPC)
	}
}

func TestServer_readResource_Info(t *testing.T) {
	srv := NewServer("test-session", &bytes.Buffer{}, &bytes.Buffer{})

	request := &mcp.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      10,
		Method:  "resources/read",
		Params: map[string]interface{}{
			"uri": "terminal://info",
		},
	}

	response := srv.handleRequest(request)

	if response == nil {
		t.Fatal("handleRequest() returned nil")
	}

	// Should get a response even if session doesn't exist
	if response.JSONRPC != "2.0" {
		t.Errorf("response.JSONRPC = %v, want 2.0", response.JSONRPC)
	}
}

func TestServer_readResource_UnknownURI(t *testing.T) {
	srv := NewServer("test-session", &bytes.Buffer{}, &bytes.Buffer{})

	request := &mcp.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      11,
		Method:  "resources/read",
		Params: map[string]interface{}{
			"uri": "terminal://unknown",
		},
	}

	response := srv.handleRequest(request)

	if response == nil {
		t.Fatal("handleRequest() returned nil")
	}
	if response.Result != nil {
		t.Error("response.Result should be nil for unknown resource")
	}
	if response.Error == nil {
		t.Fatal("response.Error is nil, expected error for unknown resource")
	}
	if !strings.Contains(response.Error.Message, "unknown resource") {
		t.Errorf("response.Error.Message = %v, should mention unknown resource", response.Error.Message)
	}
}

func TestServer_Start_EOF(t *testing.T) {
	// Test that Start() returns nil on EOF
	reader := &bytes.Buffer{} // Empty buffer will return EOF
	writer := &bytes.Buffer{}
	srv := NewServer("test-session-eof", reader, writer)

	// Start will try to ensure session exists, which may fail if tmux is not installed
	// But we're mainly testing the EOF handling in the message loop
	err := srv.Start()

	// If tmux is not installed, we'll get an error about that
	// Otherwise, we should get nil (EOF is not an error)
	if err != nil && !strings.Contains(err.Error(), "tmux") {
		t.Errorf("Start() error = %v, want nil or tmux-related error", err)
	}
}

func TestServer_Start_InvalidJSON(t *testing.T) {
	// Test that Start() handles invalid JSON
	reader := strings.NewReader("invalid json\n")
	writer := &bytes.Buffer{}
	srv := NewServer("test-session-invalid", reader, writer)

	err := srv.Start()

	// Should get an error about decoding
	if err == nil {
		t.Error("Start() should return error for invalid JSON")
	}
	if err != nil && !strings.Contains(err.Error(), "decode") && !strings.Contains(err.Error(), "tmux") {
		t.Errorf("Start() error = %v, should mention decode failure or tmux", err)
	}
}

func TestServer_Start_ValidRequest(t *testing.T) {
	// Test that Start() processes a valid request
	request := mcp.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	reader := bytes.NewReader(requestJSON)
	writer := &bytes.Buffer{}
	srv := NewServer("test-session-valid", reader, writer)

	err = srv.Start()

	// Will fail if tmux session setup fails, which is expected in test environment
	// We're mainly testing that the JSON processing works
	if err != nil && !strings.Contains(err.Error(), "tmux") {
		t.Logf("Start() error = %v (expected if tmux is not available)", err)
	}

	// Check if response was written
	if writer.Len() > 0 {
		var response mcp.JSONRPCResponse
		if err := json.Unmarshal(writer.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}
	}
}

func TestServer_handleInitialize(t *testing.T) {
	srv := NewServer("test-session", &bytes.Buffer{}, &bytes.Buffer{})

	request := &mcp.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
	}

	result, err := srv.handleInitialize(request)

	if err != nil {
		t.Fatalf("handleInitialize() error = %v, want nil", err)
	}
	if result == nil {
		t.Fatal("handleInitialize() returned nil result")
	}
	if result.ProtocolVersion != ProtocolVersion {
		t.Errorf("result.ProtocolVersion = %v, want %v", result.ProtocolVersion, ProtocolVersion)
	}
	if result.ServerInfo.Name != ServerName {
		t.Errorf("result.ServerInfo.Name = %v, want %v", result.ServerInfo.Name, ServerName)
	}
	if result.ServerInfo.Version != ServerVersion {
		t.Errorf("result.ServerInfo.Version = %v, want %v", result.ServerInfo.Version, ServerVersion)
	}
	if result.Capabilities.Tools == nil {
		t.Error("result.Capabilities.Tools is nil")
	}
	if result.Capabilities.Resources == nil {
		t.Error("result.Capabilities.Resources is nil")
	}
}

func TestServer_listTools(t *testing.T) {
	srv := NewServer("test-session", &bytes.Buffer{}, &bytes.Buffer{})

	result := srv.listTools()

	if result == nil {
		t.Fatal("listTools() returned nil")
	}
	if len(result.Tools) == 0 {
		t.Error("listTools() returned empty tools list")
	}

	// Verify each tool has required fields
	for _, tool := range result.Tools {
		if tool.Name == "" {
			t.Error("Tool has empty name")
		}
		if tool.Description == "" {
			t.Errorf("Tool %q has empty description", tool.Name)
		}
		if tool.InputSchema.Type != "object" {
			t.Errorf("Tool %q InputSchema.Type = %v, want object", tool.Name, tool.InputSchema.Type)
		}
	}
}

func TestServer_listResources(t *testing.T) {
	srv := NewServer("test-session", &bytes.Buffer{}, &bytes.Buffer{})

	result := srv.listResources()

	if result == nil {
		t.Fatal("listResources() returned nil")
	}
	if len(result.Resources) == 0 {
		t.Error("listResources() returned empty resources list")
	}

	// Verify each resource has required fields
	for _, resource := range result.Resources {
		if resource.URI == "" {
			t.Error("Resource has empty URI")
		}
		if resource.Name == "" {
			t.Errorf("Resource %q has empty name", resource.URI)
		}
		if !strings.HasPrefix(resource.URI, "terminal://") {
			t.Errorf("Resource URI %q does not start with terminal://", resource.URI)
		}
	}
}

func TestServer_callTool_InvalidParams(t *testing.T) {
	srv := NewServer("test-session", &bytes.Buffer{}, &bytes.Buffer{})

	request := &mcp.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      12,
		Method:  "tools/call",
		Params:  "invalid params", // String instead of object
	}

	response := srv.handleRequest(request)

	if response == nil {
		t.Fatal("handleRequest() returned nil")
	}
	if response.Error == nil {
		t.Error("response.Error should not be nil for invalid params")
	}
}

func TestServer_readResource_InvalidParams(t *testing.T) {
	srv := NewServer("test-session", &bytes.Buffer{}, &bytes.Buffer{})

	request := &mcp.JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      13,
		Method:  "resources/read",
		Params:  "invalid params", // String instead of object
	}

	response := srv.handleRequest(request)

	if response == nil {
		t.Fatal("handleRequest() returned nil")
	}
	if response.Error == nil {
		t.Error("response.Error should not be nil for invalid params")
	}
}

// Mock reader that returns error after first read
type errorReader struct {
	count int
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	r.count++
	if r.count > 1 {
		return 0, io.ErrUnexpectedEOF
	}
	// Return a valid JSON-RPC request on first read
	request := `{"jsonrpc":"2.0","id":1,"method":"initialize"}`
	copy(p, request)
	return len(request), nil
}

func TestServer_Start_ReadError(t *testing.T) {
	reader := &errorReader{}
	writer := &bytes.Buffer{}
	srv := NewServer("test-session-error", reader, writer)

	err := srv.Start()

	// Should get an error (either from tmux setup or from the reader)
	if err == nil {
		t.Error("Start() should return error when reader fails")
	}
}
