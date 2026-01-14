package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Border colors
	activeBorderColor   = lipgloss.Color("63")
	inactiveBorderColor = lipgloss.Color("240")

	// Section list - indented to align with chapter title text
	selectedPrefix = "  "
	normalPrefix   = "  "

	// Chapter headers in section pane
	chapterPrefix = "▼ "
	chapterStyle  = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("81")) // Cyan for visual distinction

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("62")).
			Foreground(lipgloss.Color("230"))

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	// Header
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230"))

	// Help overlay
	helpStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2)

	// Status bar
	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("203"))

	// Timestamp line
	timestampStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Italic(true)

	// Description pane labels (WHAT/WHY)
	descriptionLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("183")) // Soft lavender
)
