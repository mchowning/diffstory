package tui_test

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mchowning/diffguide/internal/model"
	"github.com/mchowning/diffguide/internal/tui"
)

func TestView_EmptyStateContainsWorkingDirectory(t *testing.T) {
	m := tui.NewModel("/test/project")
	view := m.View()

	if !strings.Contains(view, "/test/project") {
		t.Error("empty state view should contain working directory")
	}
}

func TestView_EmptyStateContainsQuitInstruction(t *testing.T) {
	m := tui.NewModel("/test/project")
	view := m.View()

	if !strings.Contains(view, "q: quit") {
		t.Error("empty state view should contain 'q: quit'")
	}
}

func TestView_EmptyStateContainsServerInstructions(t *testing.T) {
	m := tui.NewModel("/test/project")
	view := m.View()

	if !strings.Contains(view, "diffguide server") {
		t.Error("empty state view should contain server start instructions")
	}
	if !strings.Contains(view, "POST") {
		t.Error("empty state view should contain POST instruction")
	}
}

func modelWithReviewAndSize(numSections int) tui.Model {
	m := tui.NewModel("/test/project")

	// Initialize viewport via WindowSizeMsg
	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// Set review
	sections := make([]model.Section, numSections)
	for i := range numSections {
		sections[i] = model.Section{
			ID:        string(rune('1' + i)),
			Narrative: "Section " + string(rune('A'+i)),
			Hunks: []model.Hunk{
				{
					File:      "file" + string(rune('1'+i)) + ".go",
					StartLine: 10 + i,
					Diff:      "+added line\n-removed line",
				},
			},
		}
	}
	review := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test Review Title",
		Sections:         sections,
	}
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	return updated.(tui.Model)
}

func TestView_ReviewStateShowsTitle(t *testing.T) {
	m := modelWithReviewAndSize(2)
	view := m.View()

	if !strings.Contains(view, "Test Review Title") {
		t.Error("review state view should contain the title")
	}
}

func TestView_ReviewStateShowsSectionList(t *testing.T) {
	m := modelWithReviewAndSize(2)
	view := m.View()

	if !strings.Contains(view, "Section A") {
		t.Error("review state view should contain section A narrative")
	}
	if !strings.Contains(view, "Section B") {
		t.Error("review state view should contain section B narrative")
	}
}

func TestView_SelectedSectionHasPrefix(t *testing.T) {
	m := modelWithReviewAndSize(2)
	view := m.View()

	// First section should be selected with "› " prefix
	if !strings.Contains(view, "› Section A") {
		t.Error("selected section should have '› ' prefix")
	}
}

func TestView_NonSelectedSectionHasSpacePrefix(t *testing.T) {
	m := modelWithReviewAndSize(2)
	view := m.View()

	// Second section should not be selected, should have "  " prefix
	if !strings.Contains(view, "  Section B") {
		t.Error("non-selected section should have '  ' prefix")
	}
}

func TestView_ReviewStateShowsHunkContent(t *testing.T) {
	m := modelWithReviewAndSize(1)
	view := m.View()

	// Should show file name
	if !strings.Contains(view, "file1.go") {
		t.Error("review state view should contain hunk file name")
	}
}

func TestView_NotReadyShowsInitializing(t *testing.T) {
	m := tui.NewModel("/test/project")
	// Set a review but don't send WindowSizeMsg (so viewport not initialized)
	review := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test",
		Sections:         []model.Section{{ID: "1", Narrative: "Section"}},
	}
	updated, _ := m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	view := m.View()

	if !strings.Contains(view, "Initializing") {
		t.Error("view should show 'Initializing' when viewport not ready")
	}
}

func TestView_SectionListDoesNotTruncateText(t *testing.T) {
	m := tui.NewModel("/test/project")

	// Initialize viewport with narrow width to force wrapping
	sizeMsg := tea.WindowSizeMsg{Width: 80, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// Create a review with a long narrative
	longNarrative := "This is a very long narrative that should wrap instead of being truncated with an ellipsis character"
	review := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test",
		Sections: []model.Section{
			{ID: "1", Narrative: longNarrative},
		},
	}
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	view := m.View()

	// Should NOT contain truncation ellipsis
	if strings.Contains(view, "…") {
		t.Error("section list should wrap text, not truncate with ellipsis")
	}
	// Should contain the full text (or at least the ending words)
	if !strings.Contains(view, "ellipsis character") {
		t.Error("section list should contain full narrative text")
	}
}

func TestView_StatusBarShowsErrorText(t *testing.T) {
	m := tui.NewModel("/test/project")

	// Initialize viewport
	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// Set a review
	review := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test",
		Sections:         []model.Section{{ID: "1", Narrative: "Section"}},
	}
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	// Set an error
	errMsg := tui.ErrorMsg{Err: errors.New("test error")}
	updated, _ = m.Update(errMsg)
	m = updated.(tui.Model)

	view := m.View()

	if !strings.Contains(view, "Error: test error") {
		t.Error("view should contain error message in status bar")
	}
}

func TestView_HelpOverlayContainsKeybindings(t *testing.T) {
	m := tui.NewModel("/test/project")

	// Initialize viewport
	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// Set a review
	review := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test",
		Sections:         []model.Section{{ID: "1", Narrative: "Section"}},
	}
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	// Toggle help on
	helpMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")}
	updated, _ = m.Update(helpMsg)
	m = updated.(tui.Model)

	view := m.View()

	if !strings.Contains(view, "Keybindings") {
		t.Error("help overlay should contain 'Keybindings'")
	}
}

func TestView_SectionListHasSpacingBetweenSections(t *testing.T) {
	m := modelWithReviewAndSize(2)
	view := m.View()

	// Find positions of both sections
	posA := strings.Index(view, "Section A")
	posB := strings.Index(view, "Section B")

	if posA == -1 || posB == -1 {
		t.Fatal("could not find both sections in view")
	}

	// Extract text between sections and check for blank line
	between := view[posA+len("Section A") : posB]

	// Should have at least 2 newlines (indicating a blank line between sections)
	newlineCount := strings.Count(between, "\n")
	if newlineCount < 2 {
		t.Errorf("expected at least 2 newlines between sections for spacing, got %d", newlineCount)
	}
}
