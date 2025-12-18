package tui

func (m Model) View() string {
	if m.review == nil {
		return m.renderEmptyState()
	}
	return m.renderReviewState()
}

func (m Model) renderEmptyState() string {
	return `
    ╔═══════════════════════════════════════════╗
    ║                                           ║
    ║             d i f f g u i d e             ║
    ║                                           ║
    ╚═══════════════════════════════════════════╝

    Watching for reviews in:
    ` + m.workDir + `

    Start server: diffguide server
    Send review:  POST http://localhost:8765/review

    q: quit | ?: help
`
}

func (m Model) renderReviewState() string {
	return ""
}
