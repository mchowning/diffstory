package tui

import (
	"strings"
	"testing"
)

func TestRenderBorderedPanel_Structure(t *testing.T) {
	result := renderBorderedPanel("Title", "content", 20, 5, false)
	lines := strings.Split(result, "\n")

	// Should have exactly 5 lines (height)
	if len(lines) != 5 {
		t.Errorf("expected 5 lines, got %d", len(lines))
	}

	// Top border should contain title
	if !strings.Contains(lines[0], "Title") {
		t.Errorf("top border should contain title, got: %s", lines[0])
	}

	// Top border should start with rounded corner
	if !strings.HasPrefix(lines[0], "╭") {
		t.Errorf("top border should start with ╭, got: %s", lines[0])
	}

	// Top border should end with rounded corner
	if !strings.HasSuffix(lines[0], "╮") {
		t.Errorf("top border should end with ╮, got: %s", lines[0])
	}

	// Bottom border should have correct corners
	lastLine := lines[len(lines)-1]
	if !strings.HasPrefix(lastLine, "╰") {
		t.Errorf("bottom border should start with ╰, got: %s", lastLine)
	}
	if !strings.HasSuffix(lastLine, "╯") {
		t.Errorf("bottom border should end with ╯, got: %s", lastLine)
	}

	// Middle lines should have vertical borders
	for i := 1; i < len(lines)-1; i++ {
		if !strings.HasPrefix(lines[i], "│") {
			t.Errorf("line %d should start with │, got: %s", i, lines[i])
		}
		if !strings.HasSuffix(lines[i], "│") {
			t.Errorf("line %d should end with │, got: %s", i, lines[i])
		}
	}
}

func TestRenderBorderedPanel_ContentVisible(t *testing.T) {
	result := renderBorderedPanel("Test", "hello world", 30, 5, false)

	if !strings.Contains(result, "hello world") {
		t.Errorf("content should be visible in result, got: %s", result)
	}
}

func TestRenderBorderedPanel_MultilineContent(t *testing.T) {
	content := "line one\nline two\nline three"
	result := renderBorderedPanel("Multi", content, 25, 6, false)
	lines := strings.Split(result, "\n")

	// Should have 6 lines total
	if len(lines) != 6 {
		t.Errorf("expected 6 lines, got %d", len(lines))
	}

	// Content lines should be present
	found := 0
	for _, line := range lines {
		if strings.Contains(line, "line one") {
			found++
		}
		if strings.Contains(line, "line two") {
			found++
		}
		if strings.Contains(line, "line three") {
			found++
		}
	}
	if found != 3 {
		t.Errorf("expected all 3 content lines to be present, found %d", found)
	}
}

func TestRenderBorderedPanel_EmptyContent(t *testing.T) {
	result := renderBorderedPanel("Empty", "", 20, 4, false)
	lines := strings.Split(result, "\n")

	// Should still have correct structure
	if len(lines) != 4 {
		t.Errorf("expected 4 lines, got %d", len(lines))
	}

	// Should still have borders
	if !strings.HasPrefix(lines[0], "╭") {
		t.Errorf("empty panel should still have top border")
	}
}

func TestRenderBorderedPanel_WithScrollbar(t *testing.T) {
	// Panel height 10, content height 8 (10 - 2 borders)
	// Scrollbar starts at line 2, height 3
	scrollbar := &ScrollbarInfo{Start: 2, Height: 3}
	result := renderBorderedPanelWithScrollbar("Test", "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8", 30, 10, false, scrollbar)
	lines := strings.Split(result, "\n")

	// Lines 0 is top border, 1-8 are content, 9 is bottom border
	// Scrollbar should appear at content lines 2, 3, 4 (indices 3, 4, 5 in the full array)
	// Those lines should end with ▐ instead of │

	// Line 3 (content line 2) should have scrollbar
	if !strings.HasSuffix(lines[3], "▐") {
		t.Errorf("line 3 should have scrollbar ▐, got: %s", lines[3])
	}

	// Line 4 (content line 3) should have scrollbar
	if !strings.HasSuffix(lines[4], "▐") {
		t.Errorf("line 4 should have scrollbar ▐, got: %s", lines[4])
	}

	// Line 5 (content line 4) should have scrollbar
	if !strings.HasSuffix(lines[5], "▐") {
		t.Errorf("line 5 should have scrollbar ▐, got: %s", lines[5])
	}

	// Line 2 (content line 1) should NOT have scrollbar
	if strings.HasSuffix(lines[2], "▐") {
		t.Errorf("line 2 should NOT have scrollbar ▐, got: %s", lines[2])
	}

	// Line 6 (content line 5) should NOT have scrollbar
	if strings.HasSuffix(lines[6], "▐") {
		t.Errorf("line 6 should NOT have scrollbar ▐, got: %s", lines[6])
	}
}

func TestRenderBorderedPanel_NoScrollbar(t *testing.T) {
	// When scrollbar is nil, should work like before
	result := renderBorderedPanelWithScrollbar("Test", "content", 20, 5, false, nil)
	lines := strings.Split(result, "\n")

	// All content lines should end with regular border
	for i := 1; i < len(lines)-1; i++ {
		if !strings.HasSuffix(lines[i], "│") {
			t.Errorf("line %d should end with │ when no scrollbar, got: %s", i, lines[i])
		}
	}
}
