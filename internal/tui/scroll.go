package tui

import "github.com/mchowning/diffstory/internal/model"

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

// ScrollOffset adjusts an offset by delta, clamping to valid range.
// Used for mouse scroll handling in section and files panels.
func ScrollOffset(current, delta, totalItems, visibleCount int) int {
	maxOffset := totalItems - visibleCount
	if maxOffset < 0 {
		maxOffset = 0
	}
	newOffset := current + delta
	if newOffset < 0 {
		return 0
	}
	if newOffset > maxOffset {
		return maxOffset
	}
	return newOffset
}

// ClickYToSectionIndex converts a mouse click Y coordinate to a section index.
// It walks the chapter structure to account for chapter headers in the visual layout.
// Returns -1 if click is on a chapter header, out of bounds, or review is nil.
//
// Visual layout:
//
//	Line 0: top border
//	Line 1: first chapter header (if scrollOffset=0)
//	Line 2: first section
//	...
func ClickYToSectionIndex(review *model.Review, scrollOffset, clickY int) int {
	if review == nil {
		return -1
	}

	// Account for top border
	borderOffset := 1
	visualLine := clickY - borderOffset

	if visualLine < 0 {
		return -1
	}

	// Walk through chapters and sections to find which section the visual line corresponds to
	currentLine := 0
	flatSectionIdx := 0

	for _, chapter := range review.Chapters {
		chapterStartSection := flatSectionIdx
		chapterEndSection := chapterStartSection + len(chapter.Sections)

		// Check if this chapter's content is in the visible range
		if chapterEndSection <= scrollOffset {
			// Skip chapters entirely before scroll offset
			flatSectionIdx = chapterEndSection
			continue
		}

		// If we haven't scrolled past this chapter's sections, render the chapter header
		if flatSectionIdx >= scrollOffset || (flatSectionIdx < scrollOffset && scrollOffset < chapterEndSection) {
			// This chapter header is visible
			if visualLine == currentLine {
				// Clicked on chapter header
				return -1
			}
			currentLine++
		}

		// Render sections in this chapter
		for i := range chapter.Sections {
			sectionFlatIdx := chapterStartSection + i
			if sectionFlatIdx >= scrollOffset {
				if visualLine == currentLine {
					return sectionFlatIdx
				}
				currentLine++
			}
			flatSectionIdx = sectionFlatIdx + 1
		}
	}

	// Click was beyond all content
	return -1
}

// ClickYToFileIndex converts a mouse click Y coordinate (relative to files pane) to a file index.
// The localY is expected to be relative to the files pane top edge.
// Returns -1 if click is on border or out of bounds.
func ClickYToFileIndex(scrollOffset, localY, totalFiles int) int {
	// Account for top border
	borderOffset := 1
	visualLine := localY - borderOffset

	if visualLine < 0 {
		return -1
	}

	fileIdx := scrollOffset + visualLine
	if fileIdx < 0 || fileIdx >= totalFiles {
		return -1
	}

	return fileIdx
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
