package server

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/conall-obrien/mcp-ssh-wingman/internal/mcp"
	"github.com/conall-obrien/mcp-ssh-wingman/internal/screen"
	"github.com/conall-obrien/mcp-ssh-wingman/internal/terminal"
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
	terminalManager terminal.WindowManager
	terminalType    string
	reader          io.Reader
	writer          io.Writer
}

// NewServer creates a new MCP server instance
func NewServer(terminalType, sessionName, windowID string, reader io.Reader, writer io.Writer) *Server {
	var manager terminal.WindowManager

	switch terminalType {
	case "screen":
		screenManager := screen.NewManager(sessionName)
		if windowID != "" {
			screenManager.SetWindow(windowID)
		}
		manager = screenManager
	case "tmux":
		fallthrough
	default:
		tmuxManager := tmux.NewManager(sessionName)
		if windowID != "" {
			tmuxManager.SetWindow(windowID)
		}
		manager = tmuxManager
	}

	return &Server{
		terminalManager: manager,
		terminalType:    terminalType,
		reader:          reader,
		writer:          writer,
	}
}

// Start begins the server message loop
func (s *Server) Start() error {
	// Ensure terminal session exists
	if err := s.terminalManager.EnsureSession(); err != nil {
		return fmt.Errorf("failed to setup terminal session: %w", err)
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
	tools := []mcp.Tool{
		{
			Name:        "read_terminal",
			Description: fmt.Sprintf("Read the current terminal content from the %s session", s.terminalType),
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
				Required:   []string{},
			},
		},
		{
			Name:        "read_scrollback",
			Description: fmt.Sprintf("Read scrollback history from the %s session", s.terminalType),
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
		{
			Name:        "list_windows",
			Description: fmt.Sprintf("List all windows/panes in the %s session", s.terminalType),
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{},
				Required:   []string{},
			},
		},
		{
			Name:        "set_window",
			Description: fmt.Sprintf("Set the active window/pane in the %s session", s.terminalType),
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"window_id": {
						Type:        "string",
						Description: "The window/pane ID to switch to",
					},
				},
				Required: []string{"window_id"},
			},
		},
	}

	return &mcp.ListToolsResult{
		Tools: tools,
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
		content, err := s.terminalManager.CapturePane()
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

		content, err := s.terminalManager.GetScrollbackHistory(lines)
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
		info, err := s.terminalManager.GetPaneInfo()
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Error: %s", err)}},
				IsError: true,
			}, nil
		}

		infoText := fmt.Sprintf("Terminal Info (%s):\n- Width: %s\n- Height: %s\n- Current Path: %s\n- Window/Pane ID: %s",
			s.terminalType, info["width"], info["height"], info["current_path"], s.terminalManager.GetWindow())

		return &mcp.CallToolResult{
			Content: []mcp.Content{{Type: "text", Text: infoText}},
		}, nil

	case "list_windows":
		windows, err := s.terminalManager.ListWindows()
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Error: %s", err)}},
				IsError: true,
			}, nil
		}

		var windowList strings.Builder
		windowList.WriteString(fmt.Sprintf("Available windows/panes in %s session:\n", s.terminalType))
		for _, window := range windows {
			windowList.WriteString(fmt.Sprintf("- ID: %s, Name: %s\n", window["id"], window["name"]))
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{{Type: "text", Text: windowList.String()}},
		}, nil

	case "set_window":
		windowID, ok := toolRequest.Arguments["window_id"].(string)
		if !ok {
			return &mcp.CallToolResult{
				Content: []mcp.Content{{Type: "text", Text: "Error: window_id must be a string"}},
				IsError: true,
			}, nil
		}

		s.terminalManager.SetWindow(windowID)
		return &mcp.CallToolResult{
			Content: []mcp.Content{{Type: "text", Text: fmt.Sprintf("Switched to window/pane: %s", windowID)}},
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
		content, err := s.terminalManager.CapturePane()
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
		info, err := s.terminalManager.GetPaneInfo()
		if err != nil {
			return nil, err
		}
		infoText := fmt.Sprintf("Terminal Information (%s):\n\nDimensions: %sx%s\nCurrent Path: %s\nWindow/Pane ID: %s",
			s.terminalType, info["width"], info["height"], info["current_path"], s.terminalManager.GetWindow())

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
