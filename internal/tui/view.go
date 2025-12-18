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

    Waiting for review data...

    Send a POST request to:
    http://localhost:` + m.port + `/review

    Example:
    curl -X POST http://localhost:` + m.port + `/review \
      -H "Content-Type: application/json" \
      -d '{"title": "My Review", "sections": [...]}'

    q: quit | ?: help
`
}

func (m Model) renderReviewState() string {
	return ""
}
