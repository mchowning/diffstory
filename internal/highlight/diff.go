package highlight

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	additionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	deletionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	hunkHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")).
			Bold(true)

	contextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)

// ColorizeDiffLine applies color styling to a diff line based on its prefix.
func ColorizeDiffLine(line string) string {
	if len(line) == 0 {
		return line
	}

	switch line[0] {
	case '+':
		return additionStyle.Render(line)
	case '-':
		return deletionStyle.Render(line)
	case '@':
		return hunkHeaderStyle.Render(line)
	default:
		return contextStyle.Render(line)
	}
}

// ColorizeDiff applies color styling to all lines in a diff.
func ColorizeDiff(diff string) string {
	if diff == "" {
		return ""
	}

	lines := strings.Split(diff, "\n")
	result := make([]string, len(lines))
	for i, line := range lines {
		result[i] = ColorizeDiffLine(line)
	}
	return strings.Join(result, "\n")
}
