---
date: 2025-12-24 14:13:55 EST
git_commit: aa4624fda16ffa134128bd8e5f7cf415676f3b3c
branch: main
repository: diffguide
topic: "Scroll Support for Sections and Files Panels"
tags: [implementation, tui, scrolling, bubbletea]
last_updated: 2025-12-24
---

# Scroll Support for Sections and Files Panels

## Summary

Added scroll offset tracking and lazygit-style scrollbar rendering to the Sections and Files panels. Previously, when items exceeded the visible area, they were clipped and the selected item could become invisible. Now both panels track scroll position, automatically scroll to keep the selected item visible, and display a visual scrollbar indicator.

## Overview

The TUI has a three-panel layout: Sections (top-left), Files (bottom-left), and Diff (right). The Diff panel already had scroll support via Bubble Tea's viewport component, but the Sections and Files panels rendered all items and relied on the border panel to clip overflow. This meant users could navigate to items that were off-screen with no visual feedback.

The implementation adds:
1. Scroll offset state tracking for both panels
2. Scroll calculation logic that keeps the selected item visible
3. Rendering that only displays items within the visible range
4. A lazygit-style scrollbar (`▐` character) on the right border when content overflows

A key design challenge was handling variable-height items in the Sections panel, where narratives can wrap to multiple lines. The solution uses two different heuristics: a conservative estimate (6 lines/section) for scroll triggering and a generous estimate (4 lines/section) for rendering. This ensures scrolling kicks in before items go off-screen while still filling available space.

## Technical Details

### Scroll Calculation Logic

A new `scroll.go` file contains pure functions for scroll calculations, making them easy to test independently.

The core scroll offset calculation keeps the selected item visible:

```go
func CalculateScrollOffset(currentOffset, selectedIndex, totalItems, visibleCount int) int {
	// If selection is above the visible range, scroll up
	if selectedIndex < currentOffset {
		return selectedIndex
	}
	// If selection is below the visible range, scroll down
	if selectedIndex >= currentOffset+visibleCount {
		return selectedIndex - visibleCount + 1
	}
	return currentOffset
}
```

The scrollbar position and size calculation follows lazygit's approach:

```go
func CalcScrollbar(totalItems, visibleCount, scrollOffset, scrollAreaHeight int) (start, height int) {
	if visibleCount >= totalItems {
		return 0, scrollAreaHeight
	}
	height = (visibleCount * scrollAreaHeight) / totalItems
	if height < 1 {
		height = 1
	}
	maxOffset := totalItems - visibleCount
	if scrollOffset >= maxOffset {
		return scrollAreaHeight - height, height
	}
	start = (scrollOffset * (scrollAreaHeight - height)) / maxOffset
	return start, height
}
```

For variable-height sections, two estimation functions serve different purposes:
- `EstimateSectionVisibleCount` (6 lines/section): Conservative estimate for scroll triggering
- `EstimateSectionRenderCount` (4 lines/section): Generous estimate for rendering to fill space

The files panel uses a simpler calculation since each item (file or directory) occupies one line, with an adjustment for the position indicator at the bottom (`internal/tui/scroll.go:36-45`). Note that the scrollbar reflects all items in the flattened tree, while the position indicator shows files-only count.

### Model State Changes

The Model struct in `internal/tui/model.go` gains two new fields for tracking scroll position:

```go
// Scroll state for panels
sectionScrollOffset int
filesScrollOffset   int
```

Helper methods calculate panel heights based on window dimensions (`sectionPanelHeight()` and `filesPanelHeight()` at `internal/tui/model.go:188-199`). The `updateFileTree()` method resets `filesScrollOffset` to 0 when the section changes, since the file list is rebuilt.

### Navigation Handler Updates

Every navigation handler in `update.go` that changes selection now calls `CalculateScrollOffset` to update the appropriate scroll offset. This includes:
- `j`/`k` and arrow keys for both panels (`internal/tui/update.go:51-108`)
- `ctrl+j`/`ctrl+k` for file navigation regardless of focus (`internal/tui/update.go:113-136`)
- `<`/`>` for jumping to first/last item (`internal/tui/update.go:211-255`)
- `,`/`.` for page up/down (`internal/tui/update.go:258-314`)

The `ReviewReceivedMsg` handler resets both scroll offsets to 0 when a new review is loaded (`internal/tui/update.go:333-335`).

### Rendering Changes

The `renderSectionPane` function now calculates a visible range and only renders sections within that range:

```go
renderCount := EstimateSectionRenderCount(height)
startIdx := m.sectionScrollOffset
endIdx := startIdx + renderCount
if endIdx > len(m.review.Sections) {
	endIdx = len(m.review.Sections)
}

for i := startIdx; i < endIdx; i++ {
	section := m.review.Sections[i]
	// ... render section
}
```

The scrollbar is calculated using the conservative visible count and passed to the border renderer (`internal/tui/view.go:146-153`).

Similar changes apply to `renderFilesContent` (`internal/tui/view.go:193-209`).

### Scrollbar Border Rendering

The `border.go` file adds a `ScrollbarInfo` struct and a new `renderBorderedPanelWithScrollbar` function. The original `renderBorderedPanel` becomes a wrapper that passes `nil` for the scrollbar.

```go
type ScrollbarInfo struct {
	Start  int // Line position where scrollbar starts (0-indexed from content top)
	Height int // Number of lines the scrollbar occupies
}
```

When rendering content lines, the function checks if the current line falls within the scrollbar range and uses `▐` instead of `│` for the right border:

```go
showScrollbar := scrollbar != nil && i >= scrollbar.Start && i < scrollbar.Start+scrollbar.Height
lines = append(lines, buildContentLineWithScrollbar(lineContent, innerWidth, colorStyle, showScrollbar))
```

The `buildContentLineWithScrollbar` function (`internal/tui/border.go:80-104`) renders the appropriate right border character.

## Git References

**Branch**: `main`

**Status**: Uncommitted changes

**Files Changed**:
- `internal/tui/scroll.go` (new) - Scroll calculation functions
- `internal/tui/scroll_test.go` (new) - 12 tests for scroll logic
- `internal/tui/border.go` - Scrollbar rendering support
- `internal/tui/border_test.go` - Scrollbar rendering tests
- `internal/tui/model.go` - Scroll offset state and helpers
- `internal/tui/update.go` - Navigation handler scroll updates
- `internal/tui/update_test.go` - Scroll offset behavior tests
- `internal/tui/view.go` - Rendering with scroll offsets
- `internal/tui/view_test.go` - Scroll rendering tests
