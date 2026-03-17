package tui

import (
	"testing"

	"github.com/mchowning/diffstory/internal/model"
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

// Tests for ClickYToSectionIndex

func TestClickYToSectionIndex_FirstSectionInFirstChapter(t *testing.T) {
	// Review with one chapter containing 3 sections
	// Visual layout (with border offset 1):
	// Line 0: border
	// Line 1: chapter header "Changes"
	// Line 2: section 0
	// Line 3: section 1
	// Line 4: section 2
	review := &model.Review{
		Chapters: []model.Chapter{
			{
				Title:    "Changes",
				Sections: []model.Section{{ID: "1"}, {ID: "2"}, {ID: "3"}},
			},
		},
	}

	// Click on line 2 (first section, after border and chapter header)
	result := ClickYToSectionIndex(review, 0, 2)
	if result != 0 {
		t.Errorf("ClickYToSectionIndex for line 2 = %d, want 0", result)
	}
}

func TestClickYToSectionIndex_SecondSectionInFirstChapter(t *testing.T) {
	review := &model.Review{
		Chapters: []model.Chapter{
			{
				Title:    "Changes",
				Sections: []model.Section{{ID: "1"}, {ID: "2"}, {ID: "3"}},
			},
		},
	}

	// Click on line 3 (second section)
	result := ClickYToSectionIndex(review, 0, 3)
	if result != 1 {
		t.Errorf("ClickYToSectionIndex for line 3 = %d, want 1", result)
	}
}

func TestClickYToSectionIndex_ClickOnChapterHeader(t *testing.T) {
	review := &model.Review{
		Chapters: []model.Chapter{
			{
				Title:    "Changes",
				Sections: []model.Section{{ID: "1"}, {ID: "2"}},
			},
		},
	}

	// Click on line 1 (chapter header) - should return -1
	result := ClickYToSectionIndex(review, 0, 1)
	if result != -1 {
		t.Errorf("ClickYToSectionIndex for chapter header = %d, want -1", result)
	}
}

func TestClickYToSectionIndex_WithScrollOffset(t *testing.T) {
	// Review with multiple sections
	review := &model.Review{
		Chapters: []model.Chapter{
			{
				Title: "Changes",
				Sections: []model.Section{
					{ID: "1"}, {ID: "2"}, {ID: "3"}, {ID: "4"}, {ID: "5"},
				},
			},
		},
	}

	// With scroll offset 2, the chapter header is STILL shown (because sections from this chapter are visible)
	// Visual layout when scrolled:
	// Line 0: border
	// Line 1: chapter header "Changes" (still shown!)
	// Line 2: section 2 (index 2)
	// Line 3: section 3 (index 3)
	// etc.
	result := ClickYToSectionIndex(review, 2, 2)
	if result != 2 {
		t.Errorf("ClickYToSectionIndex with scroll offset 2, line 2 = %d, want 2", result)
	}
}

func TestClickYToSectionIndex_MultipleChapters(t *testing.T) {
	// Review with two chapters
	// Visual layout:
	// Line 0: border
	// Line 1: chapter header "Chapter 1"
	// Line 2: section 0
	// Line 3: section 1
	// Line 4: chapter header "Chapter 2"
	// Line 5: section 2
	review := &model.Review{
		Chapters: []model.Chapter{
			{
				Title:    "Chapter 1",
				Sections: []model.Section{{ID: "1"}, {ID: "2"}},
			},
			{
				Title:    "Chapter 2",
				Sections: []model.Section{{ID: "3"}},
			},
		},
	}

	// Click on line 5 (section in second chapter, flat index 2)
	result := ClickYToSectionIndex(review, 0, 5)
	if result != 2 {
		t.Errorf("ClickYToSectionIndex for line 5 = %d, want 2", result)
	}
}

func TestClickYToSectionIndex_ClickOnSecondChapterHeader(t *testing.T) {
	review := &model.Review{
		Chapters: []model.Chapter{
			{
				Title:    "Chapter 1",
				Sections: []model.Section{{ID: "1"}, {ID: "2"}},
			},
			{
				Title:    "Chapter 2",
				Sections: []model.Section{{ID: "3"}},
			},
		},
	}

	// Click on line 4 (second chapter header) - should return -1
	result := ClickYToSectionIndex(review, 0, 4)
	if result != -1 {
		t.Errorf("ClickYToSectionIndex for second chapter header = %d, want -1", result)
	}
}

func TestClickYToSectionIndex_OutOfBounds(t *testing.T) {
	review := &model.Review{
		Chapters: []model.Chapter{
			{
				Title:    "Changes",
				Sections: []model.Section{{ID: "1"}, {ID: "2"}},
			},
		},
	}

	// Click on line 10 (beyond all content)
	result := ClickYToSectionIndex(review, 0, 10)
	if result != -1 {
		t.Errorf("ClickYToSectionIndex out of bounds = %d, want -1", result)
	}
}

func TestClickYToSectionIndex_NilReview(t *testing.T) {
	result := ClickYToSectionIndex(nil, 0, 2)
	if result != -1 {
		t.Errorf("ClickYToSectionIndex with nil review = %d, want -1", result)
	}
}

// Tests for ClickYToFileIndex

func TestClickYToFileIndex_FirstFile(t *testing.T) {
	// Files pane layout (localY is already relative to files pane):
	// Line 0: top border
	// Line 1: file 0
	// Line 2: file 1
	// etc.

	// Click on line 1 (first file)
	result := ClickYToFileIndex(0, 1, 10)
	if result != 0 {
		t.Errorf("ClickYToFileIndex for line 1 = %d, want 0", result)
	}
}

func TestClickYToFileIndex_SecondFile(t *testing.T) {
	// Click on line 2 (second file)
	result := ClickYToFileIndex(0, 2, 10)
	if result != 1 {
		t.Errorf("ClickYToFileIndex for line 2 = %d, want 1", result)
	}
}

func TestClickYToFileIndex_WithScrollOffset(t *testing.T) {
	// With scroll offset 3, line 1 shows file index 3
	result := ClickYToFileIndex(3, 1, 10)
	if result != 3 {
		t.Errorf("ClickYToFileIndex with scroll offset 3, line 1 = %d, want 3", result)
	}
}

func TestClickYToFileIndex_OutOfBounds(t *testing.T) {
	// Click beyond the file list
	result := ClickYToFileIndex(0, 20, 5)
	if result != -1 {
		t.Errorf("ClickYToFileIndex out of bounds = %d, want -1", result)
	}
}

func TestClickYToFileIndex_OnBorder(t *testing.T) {
	// Click on border (line 0)
	result := ClickYToFileIndex(0, 0, 10)
	if result != -1 {
		t.Errorf("ClickYToFileIndex on border = %d, want -1", result)
	}
}

func TestClickYToFileIndex_EmptyList(t *testing.T) {
	// No files
	result := ClickYToFileIndex(0, 1, 0)
	if result != -1 {
		t.Errorf("ClickYToFileIndex with empty list = %d, want -1", result)
	}
}
