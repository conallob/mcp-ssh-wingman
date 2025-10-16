package server

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/conall-obrien/mcp-ssh-wingman/internal/mcp"
	"github.com/conall-obrien/mcp-ssh-wingman/internal/tmux"
)

const (
	ProtocolVersion = "2024-11-05"
	ServerName      = "mcp-ssh-wingman"
)

var (
	// ServerVersion is set via ldflags during build (e.g., -ldflags "-X github.com/conall-obrien/mcp-ssh-wingman/internal/server.Version=v1.0.0")
	ServerVersion = "dev"
)

// Server represents the MCP server
type Server struct {
	tmuxManager *tmux.Manager
	reader      io.Reader
	writer      io.Writer
}

// NewServer creates a new MCP server instance
func NewServer(sessionName string, reader io.Reader, writer io.Writer) *Server {
	return &Server{
		tmuxManager: tmux.NewManager(sessionName),
		reader:      reader,
		writer:      writer,
	}
}

// Start begins the server message loop
func (s *Server) Start() error {
	// Ensure tmux session exists
	if err := s.tmuxManager.EnsureSession(); err != nil {
		// Send a proper JSON-RPC error response before returning
		encoder := json.NewEncoder(s.writer)
		errorResponse := &mcp.JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      nil, // No request ID yet
			Error: &mcp.JSONRPCError{
				Code:    -32603, // Internal error
				Message: fmt.Sprintf("Failed to setup tmux session: %s. Please ensure tmux is installed and the specified session exists or can be created.", err.Error()),
			},
		}
		// Best-effort attempt to send error response
		_ = encoder.Encode(errorResponse)
		return fmt.Errorf("failed to setup tmux session: %w", err)
	}

	decoder := json.NewDecoder(s.reader)
	encoder := json.NewEncoder(s.writer)

	for {
		var request mcp.JSONRPCRequest
		if err := decoder.Decode(&request); err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("failed to decode request: %w", err)
		}

		response := s.handleRequest(&request)
		if err := encoder.Encode(response); err != nil {
			return fmt.Errorf("failed to encode response: %w", err)
		}
	}
}

func (s *Server) handleRequest(request *mcp.JSONRPCRequest) *mcp.JSONRPCResponse {
	response := &mcp.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
	}

	switch request.Method {
	case "initialize":
		result, err := s.handleInitialize(request)
		if err != nil {
			response.Error = &mcp.JSONRPCError{
				Code:    -32603,
				Message: err.Error(),
			}
		} else {
			response.Result = result
		}

	case "tools/list":
		response.Result = s.listTools()

	case "tools/call":
		result, err := s.callTool(request)
		if err != nil {
			response.Error = &mcp.JSONRPCError{
				Code:    -32603,
				Message: err.Error(),
			}
		} else {
			response.Result = result
		}

	case "resources/list":
		response.Result = s.listResources()

	case "resources/read":
		result, err := s.readResource(request)
		if err != nil {
			response.Error = &mcp.JSONRPCError{
				Code:    -32603,
				Message: err.Error(),
			}
		} else {
			response.Result = result
		}

	default:
		response.Error = &mcp.JSONRPCError{
			Code:    -32601,
			Message: fmt.Sprintf("Method not found: %s", request.Method),
		}
	}

	return response
}

func (s *Server) handleInitialize(request *mcp.JSONRPCRequest) (*mcp.InitializeResult, error) {
	return &mcp.InitializeResult{
		ProtocolVersion: ProtocolVersion,
		Capabilities: mcp.ServerCapabilities{
			Tools: &mcp.ToolsCapability{
				ListChanged: false,
			},
			Resources: &mcp.ResourcesCapability{
				Subscribe:   false,
				ListChanged: false,
			},
		},
		ServerInfo: mcp.ServerInfo{
			Name:    ServerName,
			Version: ServerVersion,
		},
	}, nil
}

func (s *Server) listTools() *mcp.ListToolsResult {
	return &mcp.ListToolsResult{
		Tools: []mcp.Tool{
			{
				Name:        "read_terminal",
				Description: "Read the current terminal content from the tmux session",
				InputSchema: mcp.InputSchema{
					Type:       "object",
					Properties: map[string]mcp.Property{},
					Required:   []string{},
				},
			},
			{
				Name:        "read_scrollback",
				Description: "Read scrollback history from the tmux session",
				InputSchema: mcp.InputSchema{
					Type: "object",
					Properties: map[string]mcp.Property{
						"lines": {
							Type:        "number",
							Description: "Number of lines of scrollback history to retrieve (default: 100)",
						},
					},
					Required: []string{},
				},
			},
			{
				Name:        "get_terminal_info",
				Description: "Get information about the terminal (dimensions, current path, etc.)",
				InputSchema: mcp.InputSchema{
					Type:       "object",
					Properties: map[string]mcp.Property{},
					Required:   []string{},
				},
			},
		},
	}
}

func (s *Server) callTool(request *mcp.JSONRPCRequest) (*mcp.CallToolResult, error) {
	paramsBytes, err := json.Marshal(request.Params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}

	var toolRequest mcp.CallToolRequest
	if err := json.Unmarshal(paramsBytes, &toolRequest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tool request: %w", err)
	}

	switch toolRequest.Name {
	case "read_terminal":
		content, err := s.tmuxManager.CapturePane()
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Error: %s", err)}},
				IsError: true,
			}, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{{Type: "text", Text: content}},
		}, nil

	case "read_scrollback":
		lines := 100 // default
		if linesVal, ok := toolRequest.Arguments["lines"]; ok {
			switch v := linesVal.(type) {
			case float64:
				lines = int(v)
			case int:
				lines = v
			}
		}

		content, err := s.tmuxManager.GetScrollbackHistory(lines)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Error: %s", err)}},
				IsError: true,
			}, nil
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{{Type: "text", Text: content}},
		}, nil

	case "get_terminal_info":
		info, err := s.tmuxManager.GetPaneInfo()
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Error: %s", err)}},
				IsError: true,
			}, nil
		}

		infoText := fmt.Sprintf("Terminal Info:\n- Width: %s\n- Height: %s\n- Current Path: %s\n- Pane Index: %s",
			info["width"], info["height"], info["current_path"], info["pane_index"])

		return &mcp.CallToolResult{
			Content: []mcp.Content{{Type: "text", Text: infoText}},
		}, nil

	default:
		return nil, fmt.Errorf("unknown tool: %s", toolRequest.Name)
	}
}

func (s *Server) listResources() *mcp.ListResourcesResult {
	return &mcp.ListResourcesResult{
		Resources: []mcp.Resource{
			{
				URI:         "terminal://current",
				Name:        "Current Terminal",
				Description: "Current terminal content",
				MimeType:    "text/plain",
			},
			{
				URI:         "terminal://info",
				Name:        "Terminal Information",
				Description: "Terminal dimensions and metadata",
				MimeType:    "text/plain",
			},
		},
	}
}

func (s *Server) readResource(request *mcp.JSONRPCRequest) (*mcp.ReadResourceResult, error) {
	paramsBytes, err := json.Marshal(request.Params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}

	var resourceRequest mcp.ReadResourceRequest
	if err := json.Unmarshal(paramsBytes, &resourceRequest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal resource request: %w", err)
	}

	switch resourceRequest.URI {
	case "terminal://current":
		content, err := s.tmuxManager.CapturePane()
		if err != nil {
			return nil, err
		}
		return &mcp.ReadResourceResult{
			Contents: []mcp.ResourceContent{
				{
					URI:      resourceRequest.URI,
					MimeType: "text/plain",
					Text:     content,
				},
			},
		}, nil

	case "terminal://info":
		info, err := s.tmuxManager.GetPaneInfo()
		if err != nil {
			return nil, err
		}
		infoText := fmt.Sprintf("Terminal Information:\n\nDimensions: %sx%s\nCurrent Path: %s\nPane Index: %s",
			info["width"], info["height"], info["current_path"], info["pane_index"])

		return &mcp.ReadResourceResult{
			Contents: []mcp.ResourceContent{
				{
					URI:      resourceRequest.URI,
					MimeType: "text/plain",
					Text:     infoText,
				},
			},
		}, nil

	default:
		return nil, fmt.Errorf("unknown resource: %s", resourceRequest.URI)
	}
}
