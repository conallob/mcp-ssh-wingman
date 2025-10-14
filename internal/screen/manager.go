package screen

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	SessionPrefix     = "mcp-wingman"
	DefaultScrollback = 1000
)

// getScrollbackFromScreenrc reads the defscrollback setting from ~/.screenrc
// Returns the value and whether it was found in the file
func getScrollbackFromScreenrc() (int, bool) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return DefaultScrollback, false
	}

	screenrcPath := filepath.Join(homeDir, ".screenrc")
	file, err := os.Open(screenrcPath)
	if err != nil {
		return DefaultScrollback, false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "defscrollback ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				if scrollback, err := strconv.Atoi(parts[1]); err == nil {
					return scrollback, true
				}
			}
		}
	}

	return DefaultScrollback, false
}

// GetMaxScrollback returns the maximum scrollback lines configured
func GetMaxScrollback() int {
	scrollback, _ := getScrollbackFromScreenrc()
	return scrollback
}

// GetDefaultScrollback returns the default scrollback lines and the max limit
// If .screenrc has defscrollback, use that as default, otherwise use 1000
func GetDefaultScrollback() (defaultLines int, maxLines int) {
	configuredScrollback, found := getScrollbackFromScreenrc()
	if found {
		// User has configured defscrollback, use it as both default and max
		return configuredScrollback, configuredScrollback
	}
	// No defscrollback found, default to 1000 but allow up to 1000
	return DefaultScrollback, DefaultScrollback
}

// Manager handles screen session management
type Manager struct {
	sessionName string
	windowID    string
}

// NewManager creates a new screen manager
func NewManager(sessionName string) *Manager {
	if sessionName == "" {
		sessionName = SessionPrefix
	}
	return &Manager{
		sessionName: sessionName,
		windowID:    "", // Empty means current window
	}
}

// NewManagerWithWindow creates a new screen manager for a specific window
func NewManagerWithWindow(sessionName, windowID string) *Manager {
	if sessionName == "" {
		sessionName = SessionPrefix
	}
	return &Manager{
		sessionName: sessionName,
		windowID:    windowID,
	}
}

// EnsureSession ensures a screen session exists, creating it if necessary
func (m *Manager) EnsureSession() error {
	// Check if session exists
	exists, err := m.SessionExists()
	if err != nil {
		return fmt.Errorf("failed to check session: %w", err)
	}

	if !exists {
		// Create new session in detached mode
		cmd := exec.Command("screen", "-dmS", m.sessionName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to create screen session: %w", err)
		}
	}

	return nil
}

// SessionExists checks if the screen session exists
func (m *Manager) SessionExists() (bool, error) {
	sessions, err := ListSessions()
	if err != nil {
		return false, err
	}

	for _, session := range sessions {
		if session == m.sessionName {
			return true, nil
		}
	}
	return false, nil
}

// CapturePane captures the current window content
func (m *Manager) CapturePane() (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// Use screen's hardcopy command to capture content
	sessionName := m.sessionName
	var cmd *exec.Cmd

	if m.windowID != "" {
		// Create a temporary file for hardcopy output
		cmd = exec.Command("screen", "-S", sessionName, "-p", m.windowID, "-X", "hardcopy", "/tmp/screen_capture")
	} else {
		cmd = exec.Command("screen", "-S", sessionName, "-X", "hardcopy", "/tmp/screen_capture")
	}
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to capture screen content: %w (stderr: %s)", err, stderr.String())
	}

	// Read the captured content
	readCmd := exec.Command("cat", "/tmp/screen_capture")
	readCmd.Stdout = &stdout
	readCmd.Stderr = &stderr

	err = readCmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to read captured content: %w (stderr: %s)", err, stderr.String())
	}

	// Clean up temporary file
	exec.Command("rm", "/tmp/screen_capture").Run()

	return stdout.String(), nil
}

// GetPaneInfo returns information about the current window
func (m *Manager) GetPaneInfo() (map[string]string, error) {
	var stdout bytes.Buffer

	sessionTarget := m.sessionName
	if m.windowID != "" {
		sessionTarget = fmt.Sprintf("%s:%s", m.sessionName, m.windowID)
	}

	// Get window information using screen's display command
	// We'll use a combination of commands to get the information
	cmd := exec.Command("screen", "-S", sessionTarget, "-Q", "info")
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		// Fallback to basic info if screen doesn't support -Q info
		return map[string]string{
			"width":        "80", // Default values
			"height":       "24",
			"current_path": "unknown",
			"window_id":    m.windowID,
		}, nil
	}

	info := strings.TrimSpace(stdout.String())

	// Parse screen info output (format varies by screen version)
	// Basic implementation - can be enhanced based on actual screen output format
	return map[string]string{
		"width":        "80", // Screen doesn't easily expose dimensions
		"height":       "24",
		"current_path": "unknown", // Screen doesn't track current path like tmux
		"window_id":    m.windowID,
		"info":         info,
	}, nil
}

// GetScrollbackHistory gets the scrollback history from the window
func (m *Manager) GetScrollbackHistory(lines int) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	sessionName := m.sessionName
	var cmd *exec.Cmd

	if m.windowID != "" {
		cmd = exec.Command("screen", "-S", sessionName, "-p", m.windowID, "-X", "hardcopy", "-h", "/tmp/screen_scrollback")
	} else {
		cmd = exec.Command("screen", "-S", sessionName, "-X", "hardcopy", "-h", "/tmp/screen_scrollback")
	}
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to capture scrollback: %w (stderr: %s)", err, stderr.String())
	}

	// Read and limit to requested number of lines
	readCmd := exec.Command("tail", "-n", strconv.Itoa(lines), "/tmp/screen_scrollback")
	readCmd.Stdout = &stdout
	readCmd.Stderr = &stderr

	err = readCmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to read scrollback content: %w (stderr: %s)", err, stderr.String())
	}

	// Clean up temporary file
	exec.Command("rm", "/tmp/screen_scrollback").Run()

	return stdout.String(), nil
}

// ListSessions lists all screen sessions
func ListSessions() ([]string, error) {
	var stdout bytes.Buffer

	cmd := exec.Command("screen", "-ls")
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		// screen -ls returns exit code 1 when no sessions exist
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return []string{}, nil
			}
		}
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	output := stdout.String()
	lines := strings.Split(output, "\n")
	var sessions []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Parse screen -ls output format: "PID.sessionname	(Detached/Attached)"
		if strings.Contains(line, ".") && (strings.Contains(line, "Detached") || strings.Contains(line, "Attached")) {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				sessionPart := parts[0]
				if dotIndex := strings.Index(sessionPart, "."); dotIndex != -1 {
					sessionName := sessionPart[dotIndex+1:]
					sessions = append(sessions, sessionName)
				}
			}
		}
	}

	return sessions, nil
}

// ListWindows lists all windows in the current session
func (m *Manager) ListWindows() ([]map[string]string, error) {
	// For now, let's just use the original method that works but truncates
	// This is safer than methods that might interfere with the user's session
	// We can improve this later with a truly non-intrusive method
	return m.listWindowsOriginal()
}

// listWindowsOriginal is the original implementation that works but truncates
func (m *Manager) listWindowsOriginal() ([]map[string]string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// Use screen's windows command (may truncate with many windows)
	cmd := exec.Command("screen", "-S", m.sessionName, "-Q", "windows")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set a wide terminal width to avoid truncation of window list
	// Screen -Q windows output is limited by COLUMNS environment variable
	cmd.Env = append(os.Environ(), "COLUMNS=500", "LINES=50")

	err := cmd.Run()
	if err != nil {
		return m.listWindowsFallback()
	}

	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return m.listWindowsFallback()
	}

	// Parse the window list
	var windows []map[string]string

	// Split into fields (words)
	fields := strings.Fields(output)

	// Map to store window data: windowNum -> title
	windowData := make(map[int]string)
	currentWindow := -1

	// Process each field
	for _, field := range fields {
		// Check if field is a pure number (window ID)
		if windowNum, err := strconv.Atoi(field); err == nil {
			// This is a window number
			windowData[windowNum] = "" // Initialize with empty title
			currentWindow = windowNum
		} else {
			// This is a title/name for the current window
			if currentWindow >= 0 {
				// Handle indicators (* or -) in the field
				if strings.HasSuffix(field, "*") || strings.HasSuffix(field, "-") {
					indicator := field[len(field)-1:]
					title := field[:len(field)-1]
					if title != "" {
						windowData[currentWindow] = title + indicator
					} else {
						windowData[currentWindow] = indicator
					}
				} else {
					windowData[currentWindow] = field
				}
			}
		}
	}

	// Convert map to sorted slice of windows
	var windowNums []int
	for num := range windowData {
		windowNums = append(windowNums, num)
	}

	// Sort window numbers
	for i := 0; i < len(windowNums)-1; i++ {
		for j := i + 1; j < len(windowNums); j++ {
			if windowNums[i] > windowNums[j] {
				windowNums[i], windowNums[j] = windowNums[j], windowNums[i]
			}
		}
	}

	// Build the result
	for _, num := range windowNums {
		title := windowData[num]
		displayName := fmt.Sprintf("%d", num)
		if title != "" {
			displayName += " " + title
		}

		windows = append(windows, map[string]string{
			"id":   fmt.Sprintf("%d", num),
			"name": displayName,
		})
	}

	if len(windows) == 0 {
		return m.listWindowsFallback()
	}

	return windows, nil
}

// listWindowsFallback provides a fallback method to list windows
func (m *Manager) listWindowsFallback() ([]map[string]string, error) {
	// Basic fallback - assumes current window exists
	return []map[string]string{
		{
			"id":   "0",
			"name": "default",
		},
	}, nil
}

// KillSession kills the screen session
func (m *Manager) KillSession() error {
	cmd := exec.Command("screen", "-S", m.sessionName, "-X", "quit")
	return cmd.Run()
}

// SetWindow sets the window ID for this manager
func (m *Manager) SetWindow(windowID string) {
	m.windowID = windowID
}

// ListSessions lists all screen sessions (implements SessionLister interface)
func (m *Manager) ListSessions() ([]string, error) {
	return ListSessions()
}

// GetWindow returns the current window ID
func (m *Manager) GetWindow() string {
	return m.windowID
}
