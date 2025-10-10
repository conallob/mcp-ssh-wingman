package tmux

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

const (
	SessionPrefix = "mcp-wingman"
)

// Manager handles tmux session management
type Manager struct {
	sessionName string
}

// NewManager creates a new tmux manager
func NewManager(sessionName string) *Manager {
	if sessionName == "" {
		sessionName = SessionPrefix
	}
	return &Manager{
		sessionName: sessionName,
	}
}

// EnsureSession ensures a tmux session exists, creating it if necessary
func (m *Manager) EnsureSession() error {
	// Check if session exists
	exists, err := m.SessionExists()
	if err != nil {
		return fmt.Errorf("failed to check session: %w", err)
	}

	if !exists {
		// Create new session in detached mode
		cmd := exec.Command("tmux", "new-session", "-d", "-s", m.sessionName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create tmux session: %w", err)
		}
	}

	return nil
}

// SessionExists checks if the tmux session exists
func (m *Manager) SessionExists() (bool, error) {
	cmd := exec.Command("tmux", "has-session", "-t", m.sessionName)
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Exit code 1 means session doesn't exist
			if exitErr.ExitCode() == 1 {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

// CapturePane captures the current pane content
func (m *Manager) CapturePane() (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command("tmux", "capture-pane", "-t", m.sessionName, "-p", "-S", "-")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to capture pane: %w (stderr: %s)", err, stderr.String())
	}

	return stdout.String(), nil
}

// GetPaneInfo returns information about the current pane
func (m *Manager) GetPaneInfo() (map[string]string, error) {
	var stdout bytes.Buffer

	// Get pane format info: width, height, current path, pane index
	cmd := exec.Command("tmux", "display-message",
		"-t", m.sessionName,
		"-p", "#{pane_width},#{pane_height},#{pane_current_path},#{pane_index}")
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to get pane info: %w", err)
	}

	parts := strings.Split(strings.TrimSpace(stdout.String()), ",")
	if len(parts) < 4 {
		return nil, fmt.Errorf("unexpected pane info format: %s", stdout.String())
	}

	return map[string]string{
		"width":        parts[0],
		"height":       parts[1],
		"current_path": parts[2],
		"pane_index":   parts[3],
	}, nil
}

// GetScrollbackHistory gets the scrollback history from the pane
func (m *Manager) GetScrollbackHistory(lines int) (string, error) {
	var stdout bytes.Buffer

	linesArg := fmt.Sprintf("-%d", lines)
	cmd := exec.Command("tmux", "capture-pane", "-t", m.sessionName, "-p", "-S", linesArg)
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to capture scrollback: %w", err)
	}

	return stdout.String(), nil
}

// ListSessions lists all tmux sessions
func ListSessions() ([]string, error) {
	var stdout bytes.Buffer

	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}")
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Exit code 1 with "no server running" is expected when no sessions exist
			if exitErr.ExitCode() == 1 {
				return []string{}, nil
			}
		}
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	sessions := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(sessions) == 1 && sessions[0] == "" {
		return []string{}, nil
	}

	return sessions, nil
}

// KillSession kills the tmux session
func (m *Manager) KillSession() error {
	cmd := exec.Command("tmux", "kill-session", "-t", m.sessionName)
	return cmd.Run()
}
