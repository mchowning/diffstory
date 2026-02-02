package tui

import (
	"testing"
)

func TestCalculateScrollOffset_SelectionInView_NoChange(t *testing.T) {
	// When selection is within visible range, offset should not change
	currentOffset := 0
	selectedIndex := 2
	totalItems := 10
	visibleCount := 5

	result := CalculateScrollOffset(currentOffset, selectedIndex, totalItems, visibleCount)

	if result != 0 {
		t.Errorf("expected offset 0, got %d", result)
	}
}

func TestCalculateScrollOffset_SelectionBelowView_ScrollsDown(t *testing.T) {
	// When selection moves below visible range, offset should increase
	// Items 0-4 are visible, selecting item 6 should scroll to show it
	currentOffset := 0
	selectedIndex := 6
	totalItems := 10
	visibleCount := 5

	result := CalculateScrollOffset(currentOffset, selectedIndex, totalItems, visibleCount)

	// Expected: offset = 6 - 5 + 1 = 2 (items 2-6 visible)
	if result != 2 {
		t.Errorf("expected offset 2, got %d", result)
	}
}

func TestCalculateScrollOffset_SelectionAboveView_ScrollsUp(t *testing.T) {
	// When selection moves above visible range, offset should decrease
	// Items 5-9 are visible (offset=5), selecting item 2 should scroll up
	currentOffset := 5
	selectedIndex := 2
	totalItems := 10
	visibleCount := 5

	result := CalculateScrollOffset(currentOffset, selectedIndex, totalItems, visibleCount)

	// Expected: offset = 2 (items 2-6 visible)
	if result != 2 {
		t.Errorf("expected offset 2, got %d", result)
	}
}

func TestCalculateScrollOffset_AllItemsFit_OffsetAlwaysZero(t *testing.T) {
	// When all items fit in view, offset should always be 0
	currentOffset := 0
	selectedIndex := 2
	totalItems := 3
	visibleCount := 5 // More visible space than items

	result := CalculateScrollOffset(currentOffset, selectedIndex, totalItems, visibleCount)

	if result != 0 {
		t.Errorf("expected offset 0, got %d", result)
	}
}

func TestCalculateScrollOffset_ClampsToMaxOffset(t *testing.T) {
	// Offset should not exceed totalItems - visibleCount
	// With 10 items and 5 visible, max offset is 5
	currentOffset := 0
	selectedIndex := 9 // Last item
	totalItems := 10
	visibleCount := 5

	result := CalculateScrollOffset(currentOffset, selectedIndex, totalItems, visibleCount)

	// Expected: offset = 5 (items 5-9 visible, max valid offset)
	if result != 5 {
		t.Errorf("expected offset 5, got %d", result)
	}
}

// Tests for CalcScrollbar

func TestCalcScrollbar_AllItemsFit_FullHeight(t *testing.T) {
	// When all items fit, scrollbar fills entire area (no scrolling needed)
	totalItems := 3
	visibleCount := 5
	scrollOffset := 0
	scrollAreaHeight := 10

	start, height := CalcScrollbar(totalItems, visibleCount, scrollOffset, scrollAreaHeight)

	if start != 0 {
		t.Errorf("expected start 0, got %d", start)
	}
	if height != scrollAreaHeight {
		t.Errorf("expected height %d, got %d", scrollAreaHeight, height)
	}
}

func TestCalcScrollbar_ProportionalHeight(t *testing.T) {
	// Height should be proportional to visibleCount/totalItems
	// 5 visible out of 10 items = 50% of area height
	totalItems := 10
	visibleCount := 5
	scrollOffset := 0
	scrollAreaHeight := 10

	_, height := CalcScrollbar(totalItems, visibleCount, scrollOffset, scrollAreaHeight)

	// Expected: (5/10) * 10 = 5
	if height != 5 {
		t.Errorf("expected height 5, got %d", height)
	}
}

func TestCalcScrollbar_PositionAtTop(t *testing.T) {
	// When scrollOffset is 0, scrollbar should be at top
	totalItems := 10
	visibleCount := 5
	scrollOffset := 0
	scrollAreaHeight := 10

	start, _ := CalcScrollbar(totalItems, visibleCount, scrollOffset, scrollAreaHeight)

	if start != 0 {
		t.Errorf("expected start 0, got %d", start)
	}
}

func TestCalcScrollbar_PositionAtBottom(t *testing.T) {
	// When at max scroll, scrollbar should be at bottom
	totalItems := 10
	visibleCount := 5
	scrollOffset := 5 // Max offset (totalItems - visibleCount)
	scrollAreaHeight := 10

	start, height := CalcScrollbar(totalItems, visibleCount, scrollOffset, scrollAreaHeight)

	// Expected: start should position scrollbar at bottom
	// scrollAreaHeight - height = 10 - 5 = 5
	expectedStart := scrollAreaHeight - height
	if start != expectedStart {
		t.Errorf("expected start %d, got %d", expectedStart, start)
	}
}

// Tests for EstimateSectionVisibleCount

func TestEstimateSectionVisibleCount_OneLinePerItem(t *testing.T) {
	// With narratives in Description panel, each section/chapter header takes 1 line
	// Panel height of 20 = 18 content lines (after borders) = 18 visible items
	panelHeight := 20
	result := EstimateSectionVisibleCount(panelHeight)

	expected := 18 // contentHeight = 20 - 2 borders
	if result != expected {
		t.Errorf("expected %d sections for height 20, got %d", expected, result)
	}
}

func TestEstimateSectionVisibleCount_SmallPanel(t *testing.T) {
	// Even a small panel should show at least 1 section
	panelHeight := 8
	result := EstimateSectionVisibleCount(panelHeight)

	if result < 1 {
		t.Errorf("expected at least 1 section, got %d", result)
	}
}

func TestEstimateSectionRenderCount_MatchesVisibleCount(t *testing.T) {
	// With 1 line per item, render count equals visible count
	panelHeight := 20

	renderCount := EstimateSectionRenderCount(panelHeight)
	visibleCount := EstimateSectionVisibleCount(panelHeight)

	if renderCount != visibleCount {
		t.Errorf("renderCount (%d) should equal visibleCount (%d)", renderCount, visibleCount)
	}
}

// Tests for ScrollOffset helper (used for mouse scrolling)

func TestScrollOffset_ScrollDownFromTop(t *testing.T) {
	result := ScrollOffset(0, 3, 20, 10)
	if result != 3 {
		t.Errorf("ScrollOffset(0, 3, 20, 10) = %d, want 3", result)
	}
}

func TestScrollOffset_ScrollUpFromMiddle(t *testing.T) {
	result := ScrollOffset(5, -3, 20, 10)
	if result != 2 {
		t.Errorf("ScrollOffset(5, -3, 20, 10) = %d, want 2", result)
	}
}

func TestScrollOffset_ClampAtTop(t *testing.T) {
	// Trying to scroll up past the beginning should clamp to 0
	result := ScrollOffset(1, -5, 20, 10)
	if result != 0 {
		t.Errorf("ScrollOffset(1, -5, 20, 10) = %d, want 0", result)
	}
}

func TestScrollOffset_ClampAtBottom(t *testing.T) {
	// Trying to scroll past the end should clamp to maxOffset
	// maxOffset = totalItems - visibleCount = 20 - 10 = 10
	result := ScrollOffset(8, 5, 20, 10)
	if result != 10 {
		t.Errorf("ScrollOffset(8, 5, 20, 10) = %d, want 10", result)
	}
}

func TestScrollOffset_NoScrollWhenAllVisible(t *testing.T) {
	// When all items fit (totalItems <= visibleCount), offset should be 0
	result := ScrollOffset(0, 3, 5, 10)
	if result != 0 {
		t.Errorf("ScrollOffset(0, 3, 5, 10) = %d, want 0 (all items visible)", result)
	}
}
