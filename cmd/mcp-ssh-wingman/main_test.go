package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestVersionFlag builds the binary and verifies the --version flag
// prints version information and exits without starting the server.
func TestVersionFlag(t *testing.T) {
	binPath := filepath.Join(t.TempDir(), "mcp-ssh-wingman-test")

	build := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, out)
	}

	cmd := exec.Command(binPath, "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("running --version failed: %v\n%s", err, out)
	}

	output := string(out)
	if !strings.Contains(output, "mcp-ssh-wingman") {
		t.Errorf("--version output = %q, want it to mention mcp-ssh-wingman", output)
	}
	if !strings.Contains(output, "commit:") {
		t.Errorf("--version output = %q, want it to mention commit", output)
	}
	if !strings.Contains(output, "built:") {
		t.Errorf("--version output = %q, want it to mention built", output)
	}
}

// TestSessionFlagDefault verifies the --help output documents the
// default tmux session name flag.
func TestSessionFlagDefault(t *testing.T) {
	binPath := filepath.Join(t.TempDir(), "mcp-ssh-wingman-test-help")

	build := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, out)
	}

	cmd := exec.Command(binPath, "--help")
	out, _ := cmd.CombinedOutput() // flag.Parse() exits non-zero on -help, that's expected

	output := string(out)
	if !strings.Contains(output, "session") {
		t.Errorf("--help output = %q, want it to document the -session flag", output)
	}
	if !strings.Contains(output, "mcp-wingman") {
		t.Errorf("--help output = %q, want it to document the default session name", output)
	}
}
