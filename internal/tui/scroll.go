package tui

// EstimateSectionVisibleCount estimates how many sections fit in the panel.
// Each section/chapter header takes 1 line (narratives are in Description panel).
// Used for scroll triggering.
func EstimateSectionVisibleCount(panelHeight int) int {
	contentHeight := panelHeight - 2 // account for borders
	if contentHeight < 1 {
		return 1
	}
	return contentHeight
}

// EstimateSectionRenderCount estimates how many sections to render.
// Each section/chapter header takes 1 line (narratives are in Description panel).
func EstimateSectionRenderCount(panelHeight int) int {
	contentHeight := panelHeight - 2 // account for borders
	if contentHeight < 1 {
		return 1
	}
	return contentHeight
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
