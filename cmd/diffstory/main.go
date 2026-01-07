package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mchowning/diffstory/internal/config"
	"github.com/mchowning/diffstory/internal/logging"
	"github.com/mchowning/diffstory/internal/storage"
	"github.com/mchowning/diffstory/internal/tui"
	"github.com/mchowning/diffstory/internal/watcher"
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
	debug := flag.Bool("debug", false, "Enable debug logging to /tmp/diffstory.log")
	flag.Parse()
	runViewer(*debug)
}

func runViewer(debug bool) {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Warning: config error: %v", err)
	}

	// Enable logging if flag or config enables it
	debugEnabled := debug || (cfg != nil && cfg.DebugLoggingEnabled)
	logger := logging.Setup(debugEnabled)
	logger.Info("diffstory starting", "debug_flag", debug, "config_enabled", cfg != nil && cfg.DebugLoggingEnabled)

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	// Create shared storage for both TUI persistence and watcher
	store, err := storage.NewStore()
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}

	w, err := watcher.NewWithStore(cwd, store, logger)
	if err != nil {
		log.Fatalf("Failed to create watcher: %v", err)
	}
	defer w.Close()

	w.Start()

	m := tui.NewModel(cwd, cfg, store, logger)
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
