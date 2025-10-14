package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/conall-obrien/mcp-ssh-wingman/internal/server"
)

var (
	// Build-time variables set by GoReleaser
	version = "dev"
	commit  = "none"
	date    = "unknown"

	sessionName  = flag.String("session", "mcp-wingman", "terminal session name to attach to")
	terminalType = flag.String("terminal", "tmux", "terminal multiplexer type: tmux or screen")
	windowID     = flag.String("window", "", "specific window/pane ID to attach to (optional)")
	versionFlag  = flag.Bool("version", false, "print version and exit")
)

func main() {
	flag.Parse()

	if *versionFlag {
		fmt.Printf("mcp-ssh-wingman %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
		os.Exit(0)
	}

	log.SetPrefix("[mcp-ssh-wingman] ")
	log.SetFlags(log.Ldate | log.Ltime)

	// Log to stderr so it doesn't interfere with JSON-RPC on stdout
	log.SetOutput(os.Stderr)

	// Validate terminal type
	if *terminalType != "tmux" && *terminalType != "screen" {
		log.Fatalf("Invalid terminal type: %s. Must be 'tmux' or 'screen'", *terminalType)
	}

	log.Printf("Starting MCP server for %s session: %s", *terminalType, *sessionName)
	if *windowID != "" {
		log.Printf("Targeting specific window/pane: %s", *windowID)
	}

	srv := server.NewServer(*terminalType, *sessionName, *windowID, os.Stdin, os.Stdout)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
