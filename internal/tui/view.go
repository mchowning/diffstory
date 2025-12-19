package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mchowning/diffguide/internal/highlight"
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
	// Account for border padding
	contentWidth := width - 4
	for i, section := range m.review.Sections {
		style := normalStyle
		prefix := normalPrefix
		if i == m.selected {
			style = selectedStyle
			prefix = selectedPrefix
		}
		// Use lipgloss width to wrap text instead of truncating
		wrappedStyle := style.Width(contentWidth)
		text := prefix + section.Narrative
		items = append(items, wrappedStyle.Render(text))
	}
	// Join with blank line between sections for spacing
	content := strings.Join(items, "\n\n")
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
		coloredDiff := highlight.ColorizeDiff(hunk.Diff)
		content.WriteString(coloredDiff + "\n\n")
	}

	return content.String()
}
