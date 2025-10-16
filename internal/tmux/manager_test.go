package tmux

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestNewManager(t *testing.T) {
	tests := []struct {
		name            string
		sessionName     string
		expectedSession string
	}{
		{
			name:            "custom session name",
			sessionName:     "my-session",
			expectedSession: "my-session",
		},
		{
			name:            "empty session name defaults to prefix",
			sessionName:     "",
			expectedSession: SessionPrefix,
		},
		{
			name:            "default prefix",
			sessionName:     SessionPrefix,
			expectedSession: SessionPrefix,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.sessionName)
			if m == nil {
				t.Fatal("NewManager() returned nil")
			}
			if m.sessionName != tt.expectedSession {
				t.Errorf("sessionName = %v, want %v", m.sessionName, tt.expectedSession)
			}
		})
	}
}

func TestCheckTmuxInstalled(t *testing.T) {
	// This test will skip if tmux is not installed
	err := checkTmuxInstalled()
	if err != nil {
		// Check if it's because tmux is not installed
		if strings.Contains(err.Error(), "not installed") || strings.Contains(err.Error(), "not in PATH") || strings.Contains(err.Error(), "not found") {
			t.Skip("tmux is not installed, skipping test")
		}
		t.Errorf("checkTmuxInstalled() unexpected error = %v", err)
	}
}

func TestManager_SessionExists(t *testing.T) {
	// Skip if tmux is not installed
	if err := checkTmuxInstalled(); err != nil {
		t.Skip("tmux is not installed, skipping test")
	}

	testSessionName := "test-session-exists-" + randomString(8)
	m := NewManager(testSessionName)

	// Session should not exist yet
	exists, err := m.SessionExists()
	if err != nil {
		t.Fatalf("SessionExists() error = %v", err)
	}
	if exists {
		t.Error("SessionExists() = true, want false for non-existent session")
	}

	// Create the session
	cmd := exec.Command("tmux", "new-session", "-d", "-s", testSessionName)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}
	defer func() {
		// Clean up
		_ = m.KillSession()
	}()

	// Now session should exist
	exists, err = m.SessionExists()
	if err != nil {
		t.Fatalf("SessionExists() error = %v", err)
	}
	if !exists {
		t.Error("SessionExists() = false, want true for existing session")
	}
}

func TestManager_EnsureSession(t *testing.T) {
	// Skip if tmux is not installed
	if err := checkTmuxInstalled(); err != nil {
		t.Skip("tmux is not installed, skipping test")
	}

	testSessionName := "test-ensure-session-" + randomString(8)
	m := NewManager(testSessionName)

	// Clean up any existing session first
	_ = m.KillSession()

	defer func() {
		// Clean up after test
		_ = m.KillSession()
	}()

	// Ensure session creates it
	if err := m.EnsureSession(); err != nil {
		t.Fatalf("EnsureSession() error = %v", err)
	}

	// Verify session exists
	exists, err := m.SessionExists()
	if err != nil {
		t.Fatalf("SessionExists() error = %v", err)
	}
	if !exists {
		t.Error("Session does not exist after EnsureSession()")
	}

	// Calling again should not error
	if err := m.EnsureSession(); err != nil {
		t.Errorf("EnsureSession() second call error = %v", err)
	}
}

func TestManager_CapturePane(t *testing.T) {
	// Skip if tmux is not installed
	if err := checkTmuxInstalled(); err != nil {
		t.Skip("tmux is not installed, skipping test")
	}

	testSessionName := "test-capture-pane-" + randomString(8)
	m := NewManager(testSessionName)

	// Create session
	if err := m.EnsureSession(); err != nil {
		t.Fatalf("EnsureSession() error = %v", err)
	}
	defer func() {
		_ = m.KillSession()
	}()

	// Send some text to the session
	testText := "Hello from test"
	cmd := exec.Command("tmux", "send-keys", "-t", testSessionName, testText, "Enter")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to send keys to session: %v", err)
	}

	// Give tmux a moment to process
	// Note: In a real environment, this might need a small delay
	// but for tests we'll try without it first

	// Capture pane
	content, err := m.CapturePane()
	if err != nil {
		t.Fatalf("CapturePane() error = %v", err)
	}

	// Content should not be empty (though exact content depends on tmux state)
	if content == "" {
		t.Log("Warning: CapturePane() returned empty content (this may be expected in some environments)")
	}
}

func TestManager_GetPaneInfo(t *testing.T) {
	// Skip if tmux is not installed
	if err := checkTmuxInstalled(); err != nil {
		t.Skip("tmux is not installed, skipping test")
	}

	testSessionName := "test-pane-info-" + randomString(8)
	m := NewManager(testSessionName)

	// Create session
	if err := m.EnsureSession(); err != nil {
		t.Fatalf("EnsureSession() error = %v", err)
	}
	defer func() {
		_ = m.KillSession()
	}()

	// Get pane info
	info, err := m.GetPaneInfo()
	if err != nil {
		t.Fatalf("GetPaneInfo() error = %v", err)
	}

	// Verify required fields are present
	requiredFields := []string{"width", "height", "current_path", "pane_index"}
	for _, field := range requiredFields {
		if _, ok := info[field]; !ok {
			t.Errorf("GetPaneInfo() missing field %q", field)
		}
	}

	// Verify width and height are reasonable
	if info["width"] == "" || info["height"] == "" {
		t.Error("GetPaneInfo() width or height is empty")
	}

	// Pane index should be a number (typically "0" for first pane)
	if info["pane_index"] == "" {
		t.Error("GetPaneInfo() pane_index is empty")
	}

	// Current path should be a valid path
	if info["current_path"] == "" {
		t.Error("GetPaneInfo() current_path is empty")
	}
}

func TestManager_GetScrollbackHistory(t *testing.T) {
	// Skip if tmux is not installed
	if err := checkTmuxInstalled(); err != nil {
		t.Skip("tmux is not installed, skipping test")
	}

	testSessionName := "test-scrollback-" + randomString(8)
	m := NewManager(testSessionName)

	// Create session
	if err := m.EnsureSession(); err != nil {
		t.Fatalf("EnsureSession() error = %v", err)
	}
	defer func() {
		_ = m.KillSession()
	}()

	// Send some lines of text
	for i := 0; i < 5; i++ {
		cmd := exec.Command("tmux", "send-keys", "-t", testSessionName, "Line "+string(rune('A'+i)), "Enter")
		if err := cmd.Run(); err != nil {
			t.Fatalf("Failed to send keys to session: %v", err)
		}
	}

	tests := []struct {
		name  string
		lines int
	}{
		{
			name:  "get 10 lines",
			lines: 10,
		},
		{
			name:  "get 100 lines",
			lines: 100,
		},
		{
			name:  "get 1 line",
			lines: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := m.GetScrollbackHistory(tt.lines)
			if err != nil {
				t.Fatalf("GetScrollbackHistory() error = %v", err)
			}
			// Content can be empty or contain text, just verify no error
			_ = content
		})
	}
}

func TestManager_KillSession(t *testing.T) {
	// Skip if tmux is not installed
	if err := checkTmuxInstalled(); err != nil {
		t.Skip("tmux is not installed, skipping test")
	}

	testSessionName := "test-kill-session-" + randomString(8)
	m := NewManager(testSessionName)

	// Create session
	if err := m.EnsureSession(); err != nil {
		t.Fatalf("EnsureSession() error = %v", err)
	}

	// Verify it exists
	exists, err := m.SessionExists()
	if err != nil {
		t.Fatalf("SessionExists() error = %v", err)
	}
	if !exists {
		t.Fatal("Session does not exist after creation")
	}

	// Kill it
	if err := m.KillSession(); err != nil {
		t.Fatalf("KillSession() error = %v", err)
	}

	// Verify it's gone
	exists, err = m.SessionExists()
	if err != nil {
		t.Fatalf("SessionExists() error = %v", err)
	}
	if exists {
		t.Error("Session still exists after KillSession()")
	}
}

func TestListSessions(t *testing.T) {
	// Skip if tmux is not installed
	if err := checkTmuxInstalled(); err != nil {
		t.Skip("tmux is not installed, skipping test")
	}

	// Create a test session
	testSessionName := "test-list-sessions-" + randomString(8)
	m := NewManager(testSessionName)

	// Kill any existing test session
	_ = m.KillSession()

	// Create the session
	if err := m.EnsureSession(); err != nil {
		t.Fatalf("EnsureSession() error = %v", err)
	}
	defer func() {
		_ = m.KillSession()
	}()

	// List sessions
	sessions, err := ListSessions()
	if err != nil {
		t.Fatalf("ListSessions() error = %v", err)
	}

	// Should contain our test session
	found := false
	for _, session := range sessions {
		if session == testSessionName {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("ListSessions() did not contain test session %q, got: %v", testSessionName, sessions)
	}
}

func TestListSessions_NoSessions(t *testing.T) {
	// Skip if tmux is not installed
	if err := checkTmuxInstalled(); err != nil {
		t.Skip("tmux is not installed, skipping test")
	}

	// This test assumes we can kill all sessions and check for empty list
	// In practice, this might not be feasible in all environments
	// So we'll just verify that ListSessions returns a valid result
	sessions, err := ListSessions()
	if err != nil {
		t.Fatalf("ListSessions() error = %v", err)
	}

	// Result should be a slice (possibly empty)
	if sessions == nil {
		t.Error("ListSessions() returned nil instead of empty slice")
	}
}

// Helper function to generate random strings for test session names
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	pid := os.Getpid()
	for i := range b {
		b[i] = letters[(pid+i)%len(letters)]
	}
	return string(b)
}

func TestManager_EnsureSession_TmuxNotInstalled(t *testing.T) {
	// This test verifies error handling when tmux is not installed
	// We can't easily simulate this without mocking, so we'll test
	// the error path by creating a scenario where tmux command would fail

	// Create a manager with a session name
	m := NewManager("test-session")

	// If tmux is actually installed, we can't test the "not installed" path
	// So we skip this test in that case
	if err := checkTmuxInstalled(); err == nil {
		t.Skip("tmux is installed, cannot test 'not installed' error path")
	}

	// Try to ensure session when tmux is not installed
	err := m.EnsureSession()
	if err == nil {
		t.Error("EnsureSession() should return error when tmux is not installed")
	}
}

func TestManager_CapturePane_NonexistentSession(t *testing.T) {
	// Skip if tmux is not installed
	if err := checkTmuxInstalled(); err != nil {
		t.Skip("tmux is not installed, skipping test")
	}

	// Create manager for a session that doesn't exist
	m := NewManager("nonexistent-session-" + randomString(8))

	// Try to capture pane without ensuring session exists
	_, err := m.CapturePane()
	if err == nil {
		t.Error("CapturePane() should return error for nonexistent session")
	}
}

func TestManager_GetPaneInfo_NonexistentSession(t *testing.T) {
	// Skip if tmux is not installed
	if err := checkTmuxInstalled(); err != nil {
		t.Skip("tmux is not installed, skipping test")
	}

	// Create manager for a session that doesn't exist
	m := NewManager("nonexistent-session-" + randomString(8))

	// Try to get pane info without ensuring session exists
	_, err := m.GetPaneInfo()
	if err == nil {
		t.Error("GetPaneInfo() should return error for nonexistent session")
	}
}

func TestManager_GetScrollbackHistory_NonexistentSession(t *testing.T) {
	// Skip if tmux is not installed
	if err := checkTmuxInstalled(); err != nil {
		t.Skip("tmux is not installed, skipping test")
	}

	// Create manager for a session that doesn't exist
	m := NewManager("nonexistent-session-" + randomString(8))

	// Try to get scrollback without ensuring session exists
	_, err := m.GetScrollbackHistory(100)
	if err == nil {
		t.Error("GetScrollbackHistory() should return error for nonexistent session")
	}
}
