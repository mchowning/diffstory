package tui

// EstimateSectionVisibleCount estimates how many sections fit in the panel.
// Uses a conservative heuristic of ~6 lines per section to account for
// text wrapping in narratives plus the blank line separator.
// Used for scroll triggering - we want scrolling to kick in early.
func EstimateSectionVisibleCount(panelHeight int) int {
	contentHeight := panelHeight - 2 // account for borders
	if contentHeight < 1 {
		return 1
	}
	linesPerSection := 6 // conservative: allows for wrapped narratives + blank line
	count := contentHeight / linesPerSection
	if count < 1 {
		return 1
	}
	return count
}

// EstimateSectionRenderCount estimates how many sections to render.
// Uses a generous heuristic of ~4 lines per section to fill available space.
// The panel will clip any overflow.
func EstimateSectionRenderCount(panelHeight int) int {
	contentHeight := panelHeight - 2 // account for borders
	if contentHeight < 1 {
		return 1
	}
	linesPerSection := 4 // generous: fill available space, let panel clip overflow
	count := contentHeight / linesPerSection
	if count < 1 {
		return 1
	}
	return count
}

// EstimateFilesVisibleCount estimates how many files fit in the panel.
// Each file takes 1 line. Subtract 1 for the position indicator at the bottom.
func EstimateFilesVisibleCount(panelHeight int) int {
	contentHeight := panelHeight - 2 // account for borders
	contentHeight -= 1               // account for position indicator
	if contentHeight < 1 {
		return 1
	}
	return contentHeight
}

// CalculateScrollOffset returns the scroll offset needed to keep selectedIndex visible.
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

// CalcScrollbar returns scrollbar start position and height.
func CalcScrollbar(totalItems, visibleCount, scrollOffset, scrollAreaHeight int) (start, height int) {
	// If all items fit, scrollbar fills entire area
	if visibleCount >= totalItems {
		return 0, scrollAreaHeight
	}

	// Height is proportional to visibleCount/totalItems
	height = (visibleCount * scrollAreaHeight) / totalItems
	if height < 1 {
		height = 1 // Minimum height of 1
	}

	// Position is based on scrollOffset relative to max offset
	maxOffset := totalItems - visibleCount
	if maxOffset <= 0 {
		return 0, height
	}

	// At max scroll, scrollbar should be at the bottom
	if scrollOffset >= maxOffset {
		return scrollAreaHeight - height, height
	}

	// Calculate proportional position
	start = (scrollOffset * (scrollAreaHeight - height)) / maxOffset
	return start, height
}
