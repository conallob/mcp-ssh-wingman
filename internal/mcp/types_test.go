package mcp

import (
	"encoding/json"
	"testing"
)

func TestJSONRPCRequest_Marshal(t *testing.T) {
	tests := []struct {
		name    string
		request JSONRPCRequest
		wantErr bool
	}{
		{
			name: "basic request",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      1,
				Method:  "initialize",
			},
			wantErr: false,
		},
		{
			name: "request with params",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				ID:      "abc123",
				Method:  "tools/call",
				Params: map[string]interface{}{
					"name": "read_terminal",
				},
			},
			wantErr: false,
		},
		{
			name: "request without ID",
			request: JSONRPCRequest{
				JSONRPC: "2.0",
				Method:  "notification",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(data) == 0 {
				t.Errorf("json.Marshal() returned empty data")
			}

			// Verify we can unmarshal it back
			var decoded JSONRPCRequest
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Errorf("json.Unmarshal() error = %v", err)
			}
			if decoded.JSONRPC != tt.request.JSONRPC {
				t.Errorf("JSONRPC mismatch: got %v, want %v", decoded.JSONRPC, tt.request.JSONRPC)
			}
			if decoded.Method != tt.request.Method {
				t.Errorf("Method mismatch: got %v, want %v", decoded.Method, tt.request.Method)
			}
		})
	}
}

func TestJSONRPCResponse_Marshal(t *testing.T) {
	tests := []struct {
		name     string
		response JSONRPCResponse
		wantErr  bool
	}{
		{
			name: "successful response",
			response: JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      1,
				Result:  map[string]string{"status": "ok"},
			},
			wantErr: false,
		},
		{
			name: "error response",
			response: JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      2,
				Error: &JSONRPCError{
					Code:    -32600,
					Message: "Invalid Request",
				},
			},
			wantErr: false,
		},
		{
			name: "response with both result and error",
			response: JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      3,
				Result:  "something",
				Error: &JSONRPCError{
					Code:    -32603,
					Message: "Internal error",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(data) == 0 {
				t.Errorf("json.Marshal() returned empty data")
			}

			// Verify we can unmarshal it back
			var decoded JSONRPCResponse
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Errorf("json.Unmarshal() error = %v", err)
			}
			if decoded.JSONRPC != tt.response.JSONRPC {
				t.Errorf("JSONRPC mismatch: got %v, want %v", decoded.JSONRPC, tt.response.JSONRPC)
			}
		})
	}
}

func TestJSONRPCError_Marshal(t *testing.T) {
	tests := []struct {
		name    string
		err     JSONRPCError
		wantErr bool
	}{
		{
			name: "basic error",
			err: JSONRPCError{
				Code:    -32601,
				Message: "Method not found",
			},
			wantErr: false,
		},
		{
			name: "error with data",
			err: JSONRPCError{
				Code:    -32603,
				Message: "Internal error",
				Data:    map[string]string{"detail": "something went wrong"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.err)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			var decoded JSONRPCError
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Errorf("json.Unmarshal() error = %v", err)
			}
			if decoded.Code != tt.err.Code {
				t.Errorf("Code mismatch: got %v, want %v", decoded.Code, tt.err.Code)
			}
			if decoded.Message != tt.err.Message {
				t.Errorf("Message mismatch: got %v, want %v", decoded.Message, tt.err.Message)
			}
		})
	}
}

func TestInitializeRequest_Marshal(t *testing.T) {
	req := InitializeRequest{
		ProtocolVersion: "2024-11-05",
		Capabilities: map[string]interface{}{
			"experimental": map[string]interface{}{},
		},
		ClientInfo: ClientInfo{
			Name:    "test-client",
			Version: "1.0.0",
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded InitializeRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.ProtocolVersion != req.ProtocolVersion {
		t.Errorf("ProtocolVersion mismatch: got %v, want %v", decoded.ProtocolVersion, req.ProtocolVersion)
	}
	if decoded.ClientInfo.Name != req.ClientInfo.Name {
		t.Errorf("ClientInfo.Name mismatch: got %v, want %v", decoded.ClientInfo.Name, req.ClientInfo.Name)
	}
	if decoded.ClientInfo.Version != req.ClientInfo.Version {
		t.Errorf("ClientInfo.Version mismatch: got %v, want %v", decoded.ClientInfo.Version, req.ClientInfo.Version)
	}
}

func TestInitializeResult_Marshal(t *testing.T) {
	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: ServerCapabilities{
			Tools: &ToolsCapability{
				ListChanged: true,
			},
			Resources: &ResourcesCapability{
				Subscribe:   true,
				ListChanged: false,
			},
		},
		ServerInfo: ServerInfo{
			Name:    "test-server",
			Version: "1.0.0",
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded InitializeResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.ProtocolVersion != result.ProtocolVersion {
		t.Errorf("ProtocolVersion mismatch: got %v, want %v", decoded.ProtocolVersion, result.ProtocolVersion)
	}
	if decoded.ServerInfo.Name != result.ServerInfo.Name {
		t.Errorf("ServerInfo.Name mismatch: got %v, want %v", decoded.ServerInfo.Name, result.ServerInfo.Name)
	}
	if decoded.Capabilities.Tools == nil {
		t.Error("Tools capability is nil")
	}
	if decoded.Capabilities.Resources == nil {
		t.Error("Resources capability is nil")
	}
}

func TestListToolsResult_Marshal(t *testing.T) {
	result := ListToolsResult{
		Tools: []Tool{
			{
				Name:        "test_tool",
				Description: "A test tool",
				InputSchema: InputSchema{
					Type: "object",
					Properties: map[string]Property{
						"param1": {
							Type:        "string",
							Description: "First parameter",
						},
					},
					Required: []string{"param1"},
				},
			},
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded ListToolsResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if len(decoded.Tools) != 1 {
		t.Fatalf("Tools length mismatch: got %v, want 1", len(decoded.Tools))
	}
	if decoded.Tools[0].Name != result.Tools[0].Name {
		t.Errorf("Tool name mismatch: got %v, want %v", decoded.Tools[0].Name, result.Tools[0].Name)
	}
}

func TestCallToolRequest_Marshal(t *testing.T) {
	req := CallToolRequest{
		Name: "read_terminal",
		Arguments: map[string]interface{}{
			"lines": 100,
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded CallToolRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.Name != req.Name {
		t.Errorf("Name mismatch: got %v, want %v", decoded.Name, req.Name)
	}
	if decoded.Arguments == nil {
		t.Error("Arguments is nil")
	}
}

func TestCallToolResult_Marshal(t *testing.T) {
	tests := []struct {
		name   string
		result CallToolResult
	}{
		{
			name: "success result",
			result: CallToolResult{
				Content: []Content{
					{Type: "text", Text: "Hello, world!"},
				},
				IsError: false,
			},
		},
		{
			name: "error result",
			result: CallToolResult{
				Content: []Content{
					{Type: "text", Text: "Error: something went wrong"},
				},
				IsError: true,
			},
		},
		{
			name: "multiple contents",
			result: CallToolResult{
				Content: []Content{
					{Type: "text", Text: "First part"},
					{Type: "text", Text: "Second part"},
				},
				IsError: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.result)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			var decoded CallToolResult
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if decoded.IsError != tt.result.IsError {
				t.Errorf("IsError mismatch: got %v, want %v", decoded.IsError, tt.result.IsError)
			}
			if len(decoded.Content) != len(tt.result.Content) {
				t.Errorf("Content length mismatch: got %v, want %v", len(decoded.Content), len(tt.result.Content))
			}
		})
	}
}

func TestListResourcesResult_Marshal(t *testing.T) {
	result := ListResourcesResult{
		Resources: []Resource{
			{
				URI:         "terminal://current",
				Name:        "Current Terminal",
				Description: "Current terminal content",
				MimeType:    "text/plain",
			},
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded ListResourcesResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if len(decoded.Resources) != 1 {
		t.Fatalf("Resources length mismatch: got %v, want 1", len(decoded.Resources))
	}
	if decoded.Resources[0].URI != result.Resources[0].URI {
		t.Errorf("Resource URI mismatch: got %v, want %v", decoded.Resources[0].URI, result.Resources[0].URI)
	}
}

func TestReadResourceRequest_Marshal(t *testing.T) {
	req := ReadResourceRequest{
		URI: "terminal://current",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded ReadResourceRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.URI != req.URI {
		t.Errorf("URI mismatch: got %v, want %v", decoded.URI, req.URI)
	}
}

func TestReadResourceResult_Marshal(t *testing.T) {
	result := ReadResourceResult{
		Contents: []ResourceContent{
			{
				URI:      "terminal://current",
				MimeType: "text/plain",
				Text:     "Terminal content",
			},
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded ReadResourceResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if len(decoded.Contents) != 1 {
		t.Fatalf("Contents length mismatch: got %v, want 1", len(decoded.Contents))
	}
	if decoded.Contents[0].URI != result.Contents[0].URI {
		t.Errorf("Content URI mismatch: got %v, want %v", decoded.Contents[0].URI, result.Contents[0].URI)
	}
}
