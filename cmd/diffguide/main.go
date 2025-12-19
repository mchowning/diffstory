package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mchowning/diffguide/internal/tui"
	"github.com/mchowning/diffguide/internal/watcher"
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

	if len(os.Args) > 1 && os.Args[1] == "mcp" {
		// MCP server mode
		mcpCmd := flag.NewFlagSet("mcp", flag.ExitOnError)
		verbose := mcpCmd.Bool("v", false, "Enable verbose logging (logs to stderr)")
		mcpCmd.Parse(os.Args[2:])
		runMCP(*verbose)
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

	w, err := watcher.New(cwd)
	if err != nil {
		log.Fatalf("Failed to create watcher: %v", err)
	}
	defer w.Close()

	w.Start()

	m := tui.NewModel(cwd)
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Pump watcher events to TUI
	go func() {
		for {
			select {
			case review := <-w.Reviews:
				p.Send(tui.ReviewReceivedMsg{Review: review})
			case <-w.Cleared:
				p.Send(tui.ReviewClearedMsg{})
			case err := <-w.Errors:
				p.Send(tui.WatchErrorMsg{Err: err})
			}
		}
	}()

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
