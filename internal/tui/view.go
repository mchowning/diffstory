package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mchowning/diffstory/internal/highlight"
	"github.com/mchowning/diffstory/internal/model"
	"github.com/mchowning/diffstory/internal/timeutil"
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
    ║             d i f f s t o r y             ║
    ║                                           ║
    ╚═══════════════════════════════════════════╝

    Watching for reviews in:
    ` + m.workDir + `

    Start server: diffstory server
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

	// Layout:
	// Left column (1/3 width): Sections + Files (2 panels)
	// Right column (2/3 width): Description + Diff (2 panels stacked)
	leftWidth := m.width / 3
	rightWidth := m.width - leftWidth - 2 // account for borders

	contentHeight := m.height - 5 // header + footer + filter line
	timestampLine := m.renderTimestamp()
	if timestampLine != "" {
		contentHeight-- // account for timestamp line
	}

	// Left column: Sections panel wraps to content, capped at half screen
	sectionHeight := m.sectionPaneHeight(contentHeight / 2)

	// Left column: Files panel takes remaining space
	filesHeight := contentHeight - sectionHeight

	// Right column: Description panel wraps to content (using right column width)
	descriptionHeight := m.descriptionPaneHeight(rightWidth, contentHeight/2)

	// Right column: Diff panel takes remaining space
	diffHeight := contentHeight - descriptionHeight

	sectionPane := m.renderSectionPane(leftWidth, sectionHeight)
	filesPane := m.renderFilesPane(leftWidth, filesHeight)

	descriptionPane := m.renderDescriptionPane(rightWidth, descriptionHeight)
	diffPane := m.renderDiffPaneWithTitle(rightWidth, diffHeight)

	// Join Sections and Files vertically to create left column
	leftColumn := lipgloss.JoinVertical(lipgloss.Left, sectionPane, filesPane)

	// Join Description and Diff vertically to create right column
	rightColumn := lipgloss.JoinVertical(lipgloss.Left, descriptionPane, diffPane)

	header := headerStyle.Render("diffstory - " + m.review.Title)
	filterLine := m.renderFilterIndicator()
	footer := "j/k: navigate | J/K: scroll | h/l: panels | f: importance filter | t: test filter | q: quit | ?: help"
	if m.statusMsg != "" {
		footer = statusStyle.Render(m.statusMsg) + "  " + footer
	}

	// Join left column with right column horizontally
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn)

	if timestampLine != "" {
		return lipgloss.JoinVertical(lipgloss.Left, header, timestampLine, content, filterLine, footer)
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, content, filterLine, footer)
}

func (m Model) renderFilterIndicator() string {
	parts := []string{"Diff filter: " + m.filterLevel.String()}
	if m.testFilter != TestFilterAll {
		parts = append(parts, m.renderTestFilterIndicator())
	}
	return strings.Join(parts, " | ")
}

func (m Model) renderTestFilterIndicator() string {
	switch m.testFilter {
	case TestFilterExcluding:
		return "Excluding tests"
	case TestFilterOnly:
		return "Tests only"
	default:
		return ""
	}
}

func (m Model) renderTimestamp() string {
	if m.review == nil || m.review.CreatedAt.IsZero() {
		return ""
	}
	relative := timeutil.FormatRelative(m.review.CreatedAt, time.Now())
	return timestampStyle.Render("Review generated " + relative)
}

func (m Model) sectionHasVisibleHunks(section model.Section) bool {
	for _, hunk := range section.Hunks {
		if m.hunkPassesFilters(hunk) {
			return true
		}
	}
	return false
}

// currentViewHasFilteredContent returns true if any hunks in the current view
// are being filtered out by the active filters
func (m Model) currentViewHasFilteredContent() bool {
	if m.review == nil {
		return false
	}
	sections := m.review.AllSections()
	if m.selected >= len(sections) {
		return false
	}

	section := sections[m.selected]

	// Determine which hunks are in current view based on file selection
	for _, hunk := range section.Hunks {
		if m.hunkInCurrentView(hunk) {
			if !m.hunkPassesFilters(hunk) {
				return true
			}
		}
	}
	return false
}

// hunkInCurrentView returns true if the hunk belongs to the current view
// (all files, selected file, or selected directory)
func (m Model) hunkInCurrentView(hunk model.Hunk) bool {
	// If no file tree or invalid selection, show all hunks
	if m.flattenedFiles == nil || m.selectedFile >= len(m.flattenedFiles) {
		return true
	}

	selectedNode := m.flattenedFiles[m.selectedFile]
	if selectedNode.IsDir {
		// Directory view - check if hunk is in this directory
		return strings.HasPrefix(hunk.File, selectedNode.FullPath+"/") || hunk.File == selectedNode.FullPath
	}
	// File view - check if hunk matches this file
	return hunk.File == selectedNode.FullPath
}

func (m Model) renderSectionPane(width, height int) string {
	var items []string
	contentWidth := width - 4
	sectionCount := m.review.SectionCount()

	// Track flat section index as we iterate through chapters
	flatIdx := 0
	startIdx := m.sectionScrollOffset
	renderCount := EstimateSectionRenderCount(height)
	rendered := 0

	for _, chapter := range m.review.Chapters {
		chapterStartIdx := flatIdx

		// Check if any section in this chapter is in the visible range
		chapterEndIdx := chapterStartIdx + len(chapter.Sections)
		if chapterEndIdx <= startIdx {
			// Skip chapters entirely before scroll offset
			flatIdx = chapterEndIdx
			continue
		}

		// Render chapter header if we're at or past its first section
		if flatIdx >= startIdx || (flatIdx < startIdx && startIdx < chapterEndIdx) {
			chapterHeader := chapterStyle.Width(contentWidth).Render(chapterPrefix + chapter.Title)
			items = append(items, chapterHeader)
		}

		// Render sections in this chapter
		for _, section := range chapter.Sections {
			if flatIdx >= startIdx && rendered < renderCount {
				items = append(items, m.renderSection(section, flatIdx, contentWidth))
				rendered++
			}
			flatIdx++

			if rendered >= renderCount {
				break
			}
		}

		if rendered >= renderCount {
			break
		}
	}

	content := strings.Join(items, "\n")

	title := "[1] Sections"
	if sectionCount > 0 {
		title = fmt.Sprintf("[1] Sections [%d/%d]", m.selected+1, sectionCount)
	}

	// Use conservative estimate for scrollbar (matches scroll triggering)
	visibleCount := EstimateSectionVisibleCount(height)
	var scrollbar *ScrollbarInfo
	contentHeight := height - 2 // account for borders
	if sectionCount > visibleCount {
		start, sbHeight := CalcScrollbar(sectionCount, visibleCount, m.sectionScrollOffset, contentHeight)
		scrollbar = &ScrollbarInfo{Start: start, Height: sbHeight}
	}

	return renderBorderedPanelWithScrollbar(title, content, width, height, m.focusedPanel == PanelSection, scrollbar)
}

func (m Model) renderSection(section model.Section, flatIdx, contentWidth int) string {
	isSelected := flatIdx == m.selected

	// Determine title to show (use Narrative as fallback if Title is empty)
	title := section.Title
	if title == "" {
		title = section.Narrative
	}

	if isSelected {
		return selectedStyle.Width(contentWidth).Render(selectedPrefix + title)
	}

	return normalStyle.Width(contentWidth).Render(normalPrefix + title)
}

func (m Model) renderDescriptionPane(width, height int) string {
	var narrative string
	if m.review != nil {
		sections := m.review.AllSections()
		if m.selected < len(sections) {
			narrative = sections[m.selected].Narrative
		}
	}

	// Calculate responsive horizontal padding
	hPadding := CalcDescriptionPadding(width)

	// Content width accounts for borders and horizontal padding on both sides
	contentWidth := width - 2 - (hPadding * 2)
	if contentWidth < 10 {
		contentWidth = 10
	}

	var content string
	if narrative != "" {
		lines := wrapText(narrative, contentWidth)
		// Apply horizontal padding to each line
		paddedLines := make([]string, 0, len(lines)+2)
		paddedLines = append(paddedLines, "") // vertical padding top
		padding := strings.Repeat(" ", hPadding)
		for _, line := range lines {
			paddedLines = append(paddedLines, padding+line)
		}
		paddedLines = append(paddedLines, "") // vertical padding bottom
		content = strings.Join(paddedLines, "\n")
	}

	return renderBorderedPanel("Description", content, width, height, false)
}

func (m Model) descriptionPaneHeight(width, maxHeight int) int {
	const minHeight = 5 // 2 border lines + 2 vertical padding lines + 1 content line minimum

	if m.review == nil {
		return minHeight
	}

	sections := m.review.AllSections()
	if m.selected >= len(sections) {
		return minHeight
	}

	narrative := sections[m.selected].Narrative
	if narrative == "" {
		return minHeight
	}

	// Calculate content width accounting for borders and horizontal padding
	hPadding := CalcDescriptionPadding(width)
	contentWidth := width - 2 - (hPadding * 2)
	if contentWidth < 10 {
		contentWidth = 10
	}

	lines := wrapText(narrative, contentWidth)

	// Height = wrapped lines + 2 borders + 2 vertical padding lines
	height := len(lines) + 4
	if height < minHeight {
		height = minHeight
	}
	if height > maxHeight {
		height = maxHeight
	}
	return height
}

func (m Model) sectionPaneHeight(maxHeight int) int {
	const minHeight = 3 // 1 content line + 2 border lines

	if m.review == nil {
		return minHeight
	}

	// Calculate lines needed: 1 per chapter header + 1 per section + 2 for borders
	contentLines := len(m.review.Chapters) + m.review.SectionCount()
	height := contentLines + 2

	if height < minHeight {
		height = minHeight
	}
	if height > maxHeight {
		height = maxHeight
	}
	return height
}

func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)
	var currentLine string

	for _, word := range words {
		if currentLine == "" {
			currentLine = word
		} else if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
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

	// Calculate scrollbar
	var scrollbar *ScrollbarInfo
	visibleCount := EstimateFilesVisibleCount(height)
	contentHeight := height - 2 // account for borders
	if m.flattenedFiles != nil && len(m.flattenedFiles) > visibleCount {
		start, sbHeight := CalcScrollbar(len(m.flattenedFiles), visibleCount, m.filesScrollOffset, contentHeight)
		scrollbar = &ScrollbarInfo{Start: start, Height: sbHeight}
	}

	return renderBorderedPanelWithScrollbar("[2] Files", content, width, height, m.focusedPanel == PanelFiles, scrollbar)
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

	// Calculate visible range based on scroll offset
	visibleCount := EstimateFilesVisibleCount(m.filesPanelHeight())
	startIdx := m.filesScrollOffset
	endIdx := startIdx + visibleCount
	if endIdx > len(m.flattenedFiles) {
		endIdx = len(m.flattenedFiles)
	}

	var lines []string
	for i := startIdx; i < endIdx; i++ {
		node := m.flattenedFiles[i]
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

	// Add "(filtered)" indicator when current view has hidden content
	title := "[0] Diff"
	if m.currentViewHasFilteredContent() {
		title = "[0] Diff (filtered)"
	}

	return renderBorderedPanel(title, content, width, height, m.focusedPanel == PanelDiff)
}

func (m Model) renderDiffContent(section model.Section) string {
	var content strings.Builder
	var lastFile string

	for _, hunk := range section.Hunks {
		if !m.hunkPassesFilters(hunk) {
			continue
		}
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

	if content.Len() == 0 {
		return "(all hunks filtered)"
	}

	return content.String()
}

func (m Model) renderDiffForFile(section model.Section, filePath string) string {
	var content strings.Builder
	first := true

	for _, hunk := range section.Hunks {
		if hunk.File == filePath && m.hunkPassesFilters(hunk) {
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

	if content.Len() == 0 {
		return "(all hunks filtered)"
	}

	return content.String()
}

func (m Model) renderDiffForDirectory(section model.Section, dirPath string) string {
	var content strings.Builder
	var lastFile string

	for _, hunk := range section.Hunks {
		inDir := strings.HasPrefix(hunk.File, dirPath+"/") || hunk.File == dirPath
		if inDir && m.hunkPassesFilters(hunk) {
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

	if content.Len() == 0 {
		return "(all hunks filtered)"
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
