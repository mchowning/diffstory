package main

import (
	"flag"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mchowning/diffguide/internal/tui"
)

func main() {
	port := flag.String("port", "8765", "HTTP server port (use 0 for ephemeral)")
	flag.Bool("v", false, "Enable verbose logging")
	flag.Parse()

	m := tui.NewModel(*port)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
