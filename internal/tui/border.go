package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

const (
	borderTopLeft     = "╭"
	borderTopRight    = "╮"
	borderBottomLeft  = "╰"
	borderBottomRight = "╯"
	borderHorizontal  = "─"
	borderVertical    = "│"
	scrollbarChar     = "▐"
)

// ScrollbarInfo contains the position and size of a scrollbar
type ScrollbarInfo struct {
	Start  int // Line position where scrollbar starts (0-indexed from content top)
	Height int // Number of lines the scrollbar occupies
}

func renderBorderedPanel(title, content string, width, height int, isActive bool) string {
	return renderBorderedPanelWithScrollbar(title, content, width, height, isActive, nil)
}

func renderBorderedPanelWithScrollbar(title, content string, width, height int, isActive bool, scrollbar *ScrollbarInfo) string {
	color := inactiveBorderColor
	if isActive {
		color = activeBorderColor
	}

	colorStyle := lipgloss.NewStyle().Foreground(color)

	innerWidth := width - 2 // account for left and right borders

	// Build top border with embedded title
	topBorder := buildTopBorder(title, innerWidth, colorStyle)

	// Build bottom border
	bottomBorder := colorStyle.Render(borderBottomLeft + strings.Repeat(borderHorizontal, innerWidth) + borderBottomRight)

	// Build content lines
	contentLines := strings.Split(content, "\n")
	contentHeight := height - 2 // account for top and bottom borders

	var lines []string
	lines = append(lines, topBorder)

	for i := 0; i < contentHeight; i++ {
		var lineContent string
		if i < len(contentLines) {
			lineContent = contentLines[i]
		}
		showScrollbar := scrollbar != nil && i >= scrollbar.Start && i < scrollbar.Start+scrollbar.Height
		lines = append(lines, buildContentLineWithScrollbar(lineContent, innerWidth, colorStyle, showScrollbar))
	}

	lines = append(lines, bottomBorder)

	return strings.Join(lines, "\n")
}

func buildTopBorder(title string, innerWidth int, colorStyle lipgloss.Style) string {
	titleWidth := runewidth.StringWidth(title)

	// Format: ╭─Title────────╮
	// We need: 1 horizontal char before title + title + remaining horizontal chars
	remainingWidth := innerWidth - 1 - titleWidth
	if remainingWidth < 0 {
		remainingWidth = 0
	}

	return colorStyle.Render(borderTopLeft + borderHorizontal + title + strings.Repeat(borderHorizontal, remainingWidth) + borderTopRight)
}

func buildContentLineWithScrollbar(content string, innerWidth int, colorStyle lipgloss.Style, showScrollbar bool) string {
	// Use lipgloss.Width which properly handles ANSI escape sequences
	contentWidth := lipgloss.Width(content)

	if contentWidth > innerWidth {
		// Truncate content to fit
		content = truncateToWidth(content, innerWidth)
		contentWidth = lipgloss.Width(content)
	}

	padding := innerWidth - contentWidth
	if padding < 0 {
		padding = 0
	}

	leftBorder := colorStyle.Render(borderVertical)
	var rightBorder string
	if showScrollbar {
		rightBorder = colorStyle.Render(scrollbarChar)
	} else {
		rightBorder = colorStyle.Render(borderVertical)
	}

	return leftBorder + content + strings.Repeat(" ", padding) + rightBorder
}

func truncateToWidth(s string, maxWidth int) string {
	var result strings.Builder
	currentWidth := 0

	for _, r := range s {
		w := runewidth.RuneWidth(r)
		if currentWidth+w > maxWidth {
			break
		}
		result.WriteRune(r)
		currentWidth += w
	}

	return result.String()
}
