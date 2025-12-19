package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case ReviewReceivedMsg:
		m.review = &msg.Review
		m.selected = 0
		return m, nil
	case ReviewClearedMsg:
		m.review = nil
		m.selected = 0
		return m, nil
	}
	return m, nil
}
