package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mchowning/diffguide/internal/highlight"
	"github.com/mchowning/diffguide/internal/model"
)

func (m Model) View() string {
	if m.review == nil {
		return m.renderEmptyState()
	}

	base := m.renderReviewState()

	if m.showHelp {
		return m.renderHelpOverlay(base)
	}

	return base
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

	// Three-panel layout:
	// Left column (1/3 width): Section (top) + Files (bottom)
	// Right column (2/3 width): Diff
	leftWidth := m.width / 3
	rightWidth := m.width - leftWidth - 2 // account for borders

	contentHeight := m.height - 4 // header + footer
	sectionHeight := contentHeight / 2
	filesHeight := contentHeight - sectionHeight

	sectionPane := m.renderSectionPane(leftWidth, sectionHeight)
	filesPane := m.renderFilesPane(leftWidth, filesHeight)

	// Join Section and Files vertically to create left column
	leftColumn := lipgloss.JoinVertical(lipgloss.Left, sectionPane, filesPane)

	diffPane := m.renderDiffPaneWithTitle(rightWidth, contentHeight)

	header := headerStyle.Render("diffguide - " + m.review.Title)
	footer := "j/k: navigate | J/K: scroll | h/l: panels | q: quit | ?: help"
	if m.statusMsg != "" {
		footer = statusStyle.Render(m.statusMsg) + "  " + footer
	}

	// Join left column with diff horizontally
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, diffPane)

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

func (m Model) renderSectionPane(width, height int) string {
	var items []string
	contentWidth := width - 4
	for i, section := range m.review.Sections {
		style := normalStyle
		prefix := normalPrefix
		if i == m.selected {
			style = selectedStyle
			prefix = selectedPrefix
		}
		wrappedStyle := style.Width(contentWidth)
		text := prefix + section.Narrative
		items = append(items, wrappedStyle.Render(text))
	}
	content := strings.Join(items, "\n\n")

	// Build title with position indicator
	title := "[1] Sections"
	if len(m.review.Sections) > 0 {
		title = fmt.Sprintf("[1] Sections [%d/%d]", m.selected+1, len(m.review.Sections))
	}

	borderStyle := inactiveBorderStyle
	if m.focusedPanel == PanelSection {
		borderStyle = activeBorderStyle
	}
	return borderStyle.Width(width).Height(height).BorderTop(true).
		BorderTopForeground(borderStyle.GetBorderTopForeground()).
		Render(title + "\n" + content)
}

func (m Model) renderFilesPane(width, height int) string {
	content := m.renderFilesContent(width - 4)

	// Add position indicator
	if m.flattenedFiles != nil && len(m.flattenedFiles) > 0 {
		totalFiles := countFiles(m.flattenedFiles)
		currentPosition := m.selectedFile + 1
		positionStr := fmt.Sprintf("%d of %d", currentPosition, totalFiles)
		content += "\n" + positionStr
	}

	borderStyle := inactiveBorderStyle
	if m.focusedPanel == PanelFiles {
		borderStyle = activeBorderStyle
	}
	return borderStyle.Width(width).Height(height).BorderTop(true).
		BorderTopForeground(borderStyle.GetBorderTopForeground()).
		Render("[2] Files\n" + content)
}

func countFiles(nodes []*FileNode) int {
	count := 0
	for _, n := range nodes {
		if !n.IsDir {
			count++
		}
	}
	return count
}

func (m Model) renderFilesContent(width int) string {
	if m.flattenedFiles == nil || len(m.flattenedFiles) == 0 {
		return "(no files)"
	}

	var lines []string
	for i, node := range m.flattenedFiles {
		indent := strings.Count(node.FullPath, "/")
		indentStr := strings.Repeat("  ", indent)

		prefix := "  "
		if i == m.selectedFile {
			prefix = "› "
		}

		var indicator string
		if node.IsDir {
			if m.collapsedPaths[node.FullPath] {
				indicator = "▶ "
			} else {
				indicator = "▼ "
			}
		} else {
			indicator = "  "
		}

		line := prefix + indentStr + indicator + node.Name

		if i == m.selectedFile {
			line = selectedStyle.Width(width).Render(line)
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderDiffPaneWithTitle(width, height int) string {
	borderStyle := inactiveBorderStyle
	if m.focusedPanel == PanelDiff {
		borderStyle = activeBorderStyle
	}

	// Build context header - always show selected file/directory
	var contextHeader string
	if m.flattenedFiles != nil && m.selectedFile < len(m.flattenedFiles) {
		selectedNode := m.flattenedFiles[m.selectedFile]
		if selectedNode.IsDir {
			contextHeader = "Viewing: " + selectedNode.FullPath + "/"
		} else {
			contextHeader = "Viewing: " + selectedNode.FullPath
		}
	} else {
		contextHeader = "Viewing: All files"
	}

	content := contextHeader + "\n" + strings.Repeat("─", width-6) + "\n" + m.viewport.View()

	return borderStyle.Width(width).Height(height).BorderTop(true).
		BorderTopForeground(borderStyle.GetBorderTopForeground()).
		Render("[0] Diff\n" + content)
}

func (m Model) renderDiffContent(section model.Section) string {
	var content strings.Builder
	var lastFile string

	for _, hunk := range section.Hunks {
		if hunk.File != lastFile {
			if lastFile != "" {
				content.WriteString("\n\n\n")
			}
			content.WriteString(hunk.File + "\n")
			content.WriteString(strings.Repeat("─", 40) + "\n")
			lastFile = hunk.File
		} else {
			content.WriteString("\n\n\n")
		}
		coloredDiff := highlight.ColorizeDiff(hunk.Diff)
		content.WriteString(coloredDiff + "\n")
	}

	return content.String()
}

func (m Model) renderDiffForFile(section model.Section, filePath string) string {
	var content strings.Builder
	first := true

	for _, hunk := range section.Hunks {
		if hunk.File == filePath {
			if first {
				content.WriteString(hunk.File + "\n")
				content.WriteString(strings.Repeat("─", 40) + "\n")
				first = false
			} else {
				content.WriteString("\n\n\n")
			}
			coloredDiff := highlight.ColorizeDiff(hunk.Diff)
			content.WriteString(coloredDiff + "\n")
		}
	}

	return content.String()
}

func (m Model) renderDiffForDirectory(section model.Section, dirPath string) string {
	var content strings.Builder
	var lastFile string

	for _, hunk := range section.Hunks {
		if strings.HasPrefix(hunk.File, dirPath+"/") || hunk.File == dirPath {
			if hunk.File != lastFile {
				if lastFile != "" {
					content.WriteString("\n\n\n")
				}
				content.WriteString(hunk.File + "\n")
				content.WriteString(strings.Repeat("─", 40) + "\n")
				lastFile = hunk.File
			} else {
				content.WriteString("\n\n\n")
			}
			coloredDiff := highlight.ColorizeDiff(hunk.Diff)
			content.WriteString(coloredDiff + "\n")
		}
	}

	return content.String()
}

func (m Model) renderHelpOverlay(base string) string {
	help := `Keybindings:

  j/k or ↑/↓    Navigate within focused panel
  h/l or ←/→    Cycle focus: Section ↔ Files
  0             Focus Diff panel
  1             Focus Section panel
  2             Focus Files panel

  J/K           Scroll diff (always)
  </>           Jump to top/bottom
  ,/.           Page up/down

  Enter         Expand/collapse directory

  q             Quit
  ?             Toggle this help

HTTP API:

  POST /review  Send review data`

	overlay := helpStyle.Render(help)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		overlay,
	)
}
