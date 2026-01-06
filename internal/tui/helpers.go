package tui

import "strings"

// CalcDescriptionPadding returns the horizontal padding for the description pane.
// Uses capped linear scaling: padding grows with width but caps at a maximum.
func CalcDescriptionPadding(width int) int {
	// Approximately 5% of width on each side, capped at 16 characters
	padding := width / 20
	if padding < 1 {
		padding = 1
	}
	if padding > 16 {
		padding = 16
	}
	return padding
}

// Truncate shortens a string to maxLen, adding "…" if truncated
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 1 {
		return "…"
	}
	return s[:maxLen-1] + "…"
}

// TruncatePathMiddle truncates a path by replacing middle directories with "..."
// Example: "src/components/auth/middleware/validate.ts" -> "src/.../validate.ts"
func TruncatePathMiddle(path string, maxWidth int) string {
	if len(path) <= maxWidth {
		return path
	}

	parts := strings.Split(path, "/")
	if len(parts) <= 2 {
		return Truncate(path, maxWidth)
	}

	first := parts[0]
	last := parts[len(parts)-1]
	ellipsis := "..."

	truncated := first + "/" + ellipsis + "/" + last

	if len(truncated) <= maxWidth {
		return truncated
	}

	return Truncate(last, maxWidth)
}
