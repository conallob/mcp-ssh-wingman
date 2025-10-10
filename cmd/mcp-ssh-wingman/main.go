package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/conall-obrien/mcp-ssh-wingman/internal/server"
)

var (
	sessionName = flag.String("session", "mcp-wingman", "tmux session name to attach to")
	version     = flag.Bool("version", false, "print version and exit")
)

func main() {
	flag.Parse()

	if *version {
		fmt.Println("mcp-ssh-wingman v0.1.0")
		os.Exit(0)
	}

	log.SetPrefix("[mcp-ssh-wingman] ")
	log.SetFlags(log.Ldate | log.Ltime)

	// Log to stderr so it doesn't interfere with JSON-RPC on stdout
	log.SetOutput(os.Stderr)

	log.Printf("Starting MCP server for tmux session: %s", *sessionName)

	srv := server.NewServer(*sessionName, os.Stdin, os.Stdout)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
