package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mchowning/diffstory/internal/config"
	"github.com/mchowning/diffstory/internal/logging"
	"github.com/mchowning/diffstory/internal/model"
	"github.com/mchowning/diffstory/internal/storage"
	"github.com/mchowning/diffstory/internal/tui"
	"github.com/mchowning/diffstory/internal/watcher"
	"github.com/muesli/termenv"
)

func main() {
	// Handle --version anywhere in args
	for _, arg := range os.Args[1:] {
		if arg == "--version" || arg == "-version" {
			fmt.Printf("diffstory %s\n", Version)
			return
		}
	}

	// Handle help for main command
	if len(os.Args) > 1 && (os.Args[1] == "--help" || os.Args[1] == "-help" || os.Args[1] == "-h") {
		printUsage()
		return
	}

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
	reviewPath := flag.String("review", "", "Load review from JSON file (bypasses watcher)")
	flag.Parse()
	runViewer(*debug, *reviewPath)
}

func printUsage() {
	fmt.Print(`diffstory - A terminal UI viewer for code reviews

Usage:
  diffstory [flags]         Start the TUI viewer
  diffstory server [flags]  Start the HTTP server

Viewer flags:
  -debug    Enable debug logging to /tmp/diffstory.log
  -review   Load review from JSON file (bypasses watcher)

Server flags:
  -port     HTTP server port (default: 8765)
  -v        Enable verbose logging

See README.md for configuration options and keybindings.
`)
}

func runViewer(debug bool, reviewPath string) {
	// Force TrueColor for consistent rendering in headless environments (e.g., VHS recordings)
	lipgloss.SetColorProfile(termenv.TrueColor)

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

	// Load review from file if provided (bypasses watcher)
	var initialReview *model.Review
	if reviewPath != "" {
		review, err := loadReviewFromFile(reviewPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading review: %v\n", err)
			os.Exit(1)
		}
		initialReview = review
	}

	// Create shared storage for both TUI persistence and watcher
	// (not needed when loading from file, but doesn't hurt)
	store, err := storage.NewStore()
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}

	var opts []tui.ModelOption
	if initialReview != nil {
		opts = append(opts, tui.WithInitialReview(initialReview))
	}

	m := tui.NewModel(cwd, cfg, store, logger, opts...)
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Only create watcher if not in direct review mode
	if initialReview == nil {
		w, err := watcher.NewWithStore(cwd, store, logger)
		if err != nil {
			log.Fatalf("Failed to create watcher: %v", err)
		}
		defer w.Close()

		w.Start()

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
	}

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
