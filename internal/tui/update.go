package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "j", "down":
			if m.review != nil && m.selected < len(m.review.Sections)-1 {
				m.selected++
				m.viewport.GotoTop()
				m.updateViewportContent()
			}
		case "k", "up":
			if m.review != nil && m.selected > 0 {
				m.selected--
				m.viewport.GotoTop()
				m.updateViewportContent()
			}
		case "J":
			m.viewport.LineDown(1)
		case "K":
			m.viewport.LineUp(1)
		case "?":
			m.showHelp = !m.showHelp
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Calculate viewport dimensions (right pane is 2/3 width minus borders)
		rightWidth := msg.Width - (msg.Width / 3) - 4
		viewportHeight := msg.Height - 4 // header + footer

		if !m.ready {
			m.viewport = viewport.New(rightWidth, viewportHeight)
			m.ready = true
		} else {
			m.viewport.Width = rightWidth
			m.viewport.Height = viewportHeight
		}
		m.updateViewportContent()
	case ReviewReceivedMsg:
		m.review = &msg.Review
		m.selected = 0
		m.viewport.GotoTop()
		m.updateViewportContent()
		return m, nil
	case ReviewClearedMsg:
		m.review = nil
		m.selected = 0
		return m, nil
	case ErrorMsg:
		m.statusMsg = "Error: " + msg.Err.Error()
		return m, tea.Tick(3*time.Second, func(time.Time) tea.Msg {
			return ClearStatusMsg{}
		})
	case ClearStatusMsg:
		m.statusMsg = ""
		return m, nil
	}
	return m, nil
}
