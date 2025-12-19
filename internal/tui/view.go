package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mchowning/diffguide/internal/model"
)

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
	if !m.ready {
		return "Initializing..."
	}

	leftWidth := m.width / 3
	rightWidth := m.width - leftWidth - 4 // borders

	leftPane := m.renderSectionList(leftWidth, m.height-4)
	rightPane := m.renderDiffPane(rightWidth, m.height-4)

	header := headerStyle.Render("diffguide - " + m.review.Title)
	footer := "j/k: navigate | J/K: scroll | q: quit | ?: help"

	content := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}

func (m Model) renderSectionList(width, height int) string {
	var items []string
	for i, section := range m.review.Sections {
		style := normalStyle
		prefix := normalPrefix
		if i == m.selected {
			style = selectedStyle
			prefix = selectedPrefix
		}
		// Truncate narrative for display (account for prefix)
		text := prefix + Truncate(section.Narrative, width-len(prefix)-4)
		items = append(items, style.Render(text))
	}
	content := strings.Join(items, "\n")
	return activeBorderStyle.Width(width).Height(height).Render(content)
}

func (m Model) renderDiffPane(width, height int) string {
	return inactiveBorderStyle.Width(width).Height(height).Render(m.viewport.View())
}

func (m Model) renderDiffContent(section model.Section) string {
	var content strings.Builder

	for _, hunk := range section.Hunks {
		content.WriteString(hunk.File + "\n")
		content.WriteString(strings.Repeat("─", 40) + "\n")
		content.WriteString(hunk.Diff + "\n\n")
	}

	return content.String()
}
