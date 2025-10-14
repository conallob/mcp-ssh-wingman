package terminal

// Manager defines the interface for terminal session managers (tmux, screen, etc.)
type Manager interface {
	// EnsureSession ensures a terminal session exists, creating it if necessary
	EnsureSession() error

	// SessionExists checks if the terminal session exists
	SessionExists() (bool, error)

	// CapturePane captures the current pane/window content
	CapturePane() (string, error)

	// GetPaneInfo returns information about the current pane/window
	GetPaneInfo() (map[string]string, error)

	// GetScrollbackHistory gets the scrollback history
	GetScrollbackHistory(lines int) (string, error)

	// KillSession kills the terminal session
	KillSession() error
}

// WindowManager extends Manager with window/pane selection capabilities
type WindowManager interface {
	Manager

	// ListWindows lists all windows/panes in the session
	ListWindows() ([]map[string]string, error)

	// SetWindow sets the active window/pane
	SetWindow(windowID string)

	// GetWindow returns the current window/pane ID
	GetWindow() string
}

// SessionLister provides session listing capabilities
type SessionLister interface {
	// ListSessions lists all available sessions
	ListSessions() ([]string, error)
}
