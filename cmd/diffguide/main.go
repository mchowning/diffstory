package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mchowning/diffguide/internal/tui"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "server" {
		// Server mode
		serverCmd := flag.NewFlagSet("server", flag.ExitOnError)
		port := serverCmd.String("port", "8765", "HTTP server port")
		verbose := serverCmd.Bool("v", false, "Enable verbose logging")
		serverCmd.Parse(os.Args[2:])
		runServer(*port, *verbose)
		return
	}

	// Viewer mode (default)
	flag.Parse()
	runViewer()
}

func runViewer() {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	m := tui.NewModel(cwd)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
