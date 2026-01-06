---
date: 2026-01-06T08:24:44-05:00
git_commit: 61632cd78043126beaed558d33128ead49c8edc2
branch: main
repository: diffstory
topic: "Move Description Panel Above Diff with Responsive Padding"
tags: [implementation, tui, layout, ui]
last_updated: 2026-01-06
---

# Move Description Panel Above Diff with Responsive Padding

## Summary

Reorganized the TUI layout to move the Description panel from the left column to the right column, placing it above the Diff panel. Added responsive horizontal padding and fixed vertical padding to the Description content for improved readability.

## Overview

The previous layout positioned the Description panel in the left column between Sections and Files, which crowded the left side and left the Description competing for limited vertical space. The new layout moves Description to the right column above the Diff panel, giving it more horizontal space and placing it closer to the diff content it describes.

The layout transformation:
- **Before**: Left column (Sections + Description + Files) | Right column (Diff)
- **After**: Left column (Sections + Files) | Right column (Description + Diff)

This change also introduced responsive padding to the Description content. The horizontal padding scales with panel width (approximately 5% per side) but caps at 8 characters to prevent excessive whitespace on very wide terminals. Fixed vertical padding of one empty line above and below the text provides visual breathing room.

## Technical Details

### Responsive Padding Calculation

A new helper function `CalcDescriptionPadding` was added to calculate horizontal padding based on panel width. The function uses linear scaling (width / 20, approximately 5% per side) with bounds of 1-8 characters:

```go
func CalcDescriptionPadding(width int) int {
	padding := width / 20
	if padding < 1 {
		padding = 1
	}
	if padding > 8 {
		padding = 8
	}
	return padding
}
```

The function was placed in `helpers.go` and exported for testability. Three tests verify the behavior at narrow widths (minimum padding), moderate widths (growth), and wide terminals (maximum cap).

### Description Pane Rendering

The `renderDescriptionPane` function was updated to apply both horizontal and vertical padding (`internal/tui/view.go:296-328`):

```go
hPadding := CalcDescriptionPadding(width)
contentWidth := width - 2 - (hPadding * 2)
if contentWidth < 10 {
	contentWidth = 10
}

var content string
if narrative != "" {
	lines := wrapText(narrative, contentWidth)
	paddedLines := make([]string, 0, len(lines)+2)
	paddedLines = append(paddedLines, "") // vertical padding top
	padding := strings.Repeat(" ", hPadding)
	for _, line := range lines {
		paddedLines = append(paddedLines, padding+line)
	}
	paddedLines = append(paddedLines, "") // vertical padding bottom
	content = strings.Join(paddedLines, "\n")
}
```

The content width calculation now accounts for borders (2 characters) plus horizontal padding on both sides. A minimum content width of 10 characters prevents text from becoming unreadable at extreme narrow widths.

### Height Calculation Updates

The `descriptionPaneHeight` function was updated to match the new padding scheme (`internal/tui/view.go:331-365`). The minimum height increased from 3 to 5 to account for 2 border lines + 2 vertical padding lines + 1 content line. The height formula changed from `len(lines) + 2` to `len(lines) + 4` to include the vertical padding lines.

### Layout Restructuring

The main layout in `renderReviewState` was restructured to create two columns with different panel compositions (`internal/tui/view.go:81-136`):

- Left column: Sections panel + Files panel (joined vertically)
- Right column: Description panel + Diff panel (joined vertically)

The Description panel now uses the right column width for sizing, which provides more horizontal space for text content. The Diff panel height is calculated as the remaining space after Description, rather than taking the full content height.

## Git References

**Branch**: `main`

**Commit Range**: Single commit

**Commits Documented**:

**61632cd78043126beaed558d33128ead49c8edc2** (2026-01-06)
Move Description panel to right column above Diff with responsive padding

Reorganizes the three-panel layout to place Description in the right column
above Diff, and simplifies the left column to just Sections + Files.

Key changes:
- Add CalcDescriptionPadding helper for responsive horizontal padding
  (5% of width, capped 1-8 characters)
- Update renderDescriptionPane to apply horizontal and vertical padding
- Update descriptionPaneHeight to account for padding lines
- Restructure renderReviewState to place Description in right column

Layout transformation:
Left  (1/3): Sections + Description + Files   →  Sections + Files
Right (2/3): Diff                            →  Description + Diff

All tests passing.
