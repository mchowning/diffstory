package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mchowning/diffguide/internal/highlight"
	"github.com/mchowning/diffguide/internal/model"
)

func (m Model) View() string {
	// Generate UI states take priority
	switch m.generateUIState {
	case GenerateUIStateSourcePicker:
		return m.renderSourcePicker()
	case GenerateUIStateCommitSelector, GenerateUIStateCommitRangeStart, GenerateUIStateCommitRangeEnd:
		return m.renderCommitSelector()
	case GenerateUIStateContextInput:
		return m.renderContextInput()
	case GenerateUIStateValidationError:
		return m.renderValidationError()
	}

	// Cancel confirmation prompt
	if m.showCancelPrompt {
		prompt := helpStyle.Render("Cancel review generation? (y/n)")
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, prompt)
	}

	// Loading state
	if m.isGenerating {
		elapsed := time.Since(m.generateStartTime).Truncate(time.Second)
		line1 := m.spinner.View() + " Generating review..."
		line2 := elapsed.String()
		// Center both lines
		centered := lipgloss.NewStyle().Align(lipgloss.Center).Render(line1 + "\n\n" + line2)
		content := helpStyle.Render(centered)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}

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
	status := ""
	if m.statusMsg != "" {
		status = "\n    " + statusStyle.Render(m.statusMsg) + "\n"
	}

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
    Generate:     Press Shift+G
` + status + `
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

	title := "[1] Sections"
	if len(m.review.Sections) > 0 {
		title = fmt.Sprintf("[1] Sections [%d/%d]", m.selected+1, len(m.review.Sections))
	}

	return renderBorderedPanel(title, content, width, height, m.focusedPanel == PanelSection)
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

	return renderBorderedPanel("[2] Files", content, width, height, m.focusedPanel == PanelFiles)
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

	return renderBorderedPanel("[0] Diff", content, width, height, m.focusedPanel == PanelDiff)
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
	var sb strings.Builder
	sb.WriteString("Keybindings:\n\n")

	// Group by context in display order
	contexts := []struct {
		name  string
		title string
	}{
		{"global", "Global"},
		{"navigation", "Navigation"},
		{"files", "Files"},
	}

	for _, ctx := range contexts {
		bindings := m.keybindings.GetByContext(ctx.name)
		if len(bindings) == 0 {
			continue
		}
		sb.WriteString(ctx.title + "\n")
		for _, b := range bindings {
			sb.WriteString(fmt.Sprintf("  %-12s %s\n", b.Key, b.Description))
		}
		sb.WriteString("\n")
	}

	overlay := helpStyle.Render(strings.TrimSuffix(sb.String(), "\n"))

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		overlay,
	)
}
