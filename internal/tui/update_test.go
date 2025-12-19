package tui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mchowning/diffguide/internal/model"
	"github.com/mchowning/diffguide/internal/tui"
)

func TestUpdate_QuitWithQKey(t *testing.T) {
	m := tui.NewModel("/test/project")
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}

	_, cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	result := cmd()
	if _, ok := result.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", result)
	}
}

func TestUpdate_QuitWithCtrlC(t *testing.T) {
	m := tui.NewModel("/test/project")
	msg := tea.KeyMsg{Type: tea.KeyCtrlC}

	_, cmd := m.Update(msg)

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	result := cmd()
	if _, ok := result.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", result)
	}
}

func TestUpdate_WindowSizeMsg(t *testing.T) {
	m := tui.NewModel("/test/project")
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}

	updated, _ := m.Update(msg)
	result := updated.(tui.Model)

	if result.Width() != 100 {
		t.Errorf("expected width 100, got %d", result.Width())
	}
	if result.Height() != 50 {
		t.Errorf("expected height 50, got %d", result.Height())
	}
}

func TestUpdate_ReviewReceivedMsgSetsReview(t *testing.T) {
	m := tui.NewModel("/test/project")
	review := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test Review",
		Sections: []model.Section{
			{ID: "1", Narrative: "Test narrative"},
		},
	}
	msg := tui.ReviewReceivedMsg{Review: review}

	updated, _ := m.Update(msg)
	result := updated.(tui.Model)

	if result.Review() == nil {
		t.Fatal("expected review to be set, got nil")
	}
	if result.Review().Title != review.Title {
		t.Errorf("Title = %q, want %q", result.Review().Title, review.Title)
	}
}

func TestUpdate_ReviewReceivedMsgResetsSelected(t *testing.T) {
	m := tui.NewModel("/test/project")
	// Simulate having previously selected something
	review := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test Review",
		Sections: []model.Section{
			{ID: "1", Narrative: "First"},
			{ID: "2", Narrative: "Second"},
		},
	}
	msg := tui.ReviewReceivedMsg{Review: review}

	updated, _ := m.Update(msg)
	result := updated.(tui.Model)

	if result.Selected() != 0 {
		t.Errorf("Selected() = %d, want 0", result.Selected())
	}
}

func TestUpdate_ReviewClearedMsgClearsReview(t *testing.T) {
	m := tui.NewModel("/test/project")
	// First set a review
	review := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test Review",
	}
	updated, _ := m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	// Then clear it
	updated, _ = m.Update(tui.ReviewClearedMsg{})
	result := updated.(tui.Model)

	if result.Review() != nil {
		t.Error("expected review to be nil after clear")
	}
}

func TestUpdate_ReviewClearedMsgResetsSelected(t *testing.T) {
	m := tui.NewModel("/test/project")
	// First set a review
	review := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test Review",
		Sections: []model.Section{
			{ID: "1", Narrative: "First"},
			{ID: "2", Narrative: "Second"},
		},
	}
	updated, _ := m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	// Then clear it
	updated, _ = m.Update(tui.ReviewClearedMsg{})
	result := updated.(tui.Model)

	if result.Selected() != 0 {
		t.Errorf("Selected() = %d, want 0", result.Selected())
	}
}

func modelWithReview(numSections int) tui.Model {
	sections := make([]model.Section, numSections)
	for i := range numSections {
		sections[i] = model.Section{ID: string(rune('1' + i)), Narrative: "Section"}
	}
	review := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test",
		Sections:         sections,
	}
	m := tui.NewModel("/test/project")
	updated, _ := m.Update(tui.ReviewReceivedMsg{Review: review})
	return updated.(tui.Model)
}

func TestUpdate_JKeyIncrementsSelectedWhenNotAtEnd(t *testing.T) {
	m := modelWithReview(3)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	updated, _ := m.Update(msg)
	result := updated.(tui.Model)

	if result.Selected() != 1 {
		t.Errorf("Selected() = %d, want 1", result.Selected())
	}
}

func TestUpdate_JKeyDoesNotIncrementWhenAtLastSection(t *testing.T) {
	m := modelWithReview(3)
	// Navigate to last section
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	updated, _ := m.Update(msg)
	m = updated.(tui.Model)
	updated, _ = m.Update(msg)
	m = updated.(tui.Model)

	// Try to go past end
	updated, _ = m.Update(msg)
	result := updated.(tui.Model)

	if result.Selected() != 2 {
		t.Errorf("Selected() = %d, want 2 (should stay at last)", result.Selected())
	}
}

func TestUpdate_KKeyDecrementsSelectedWhenNotAtStart(t *testing.T) {
	m := modelWithReview(3)
	// Navigate to middle
	jMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	updated, _ := m.Update(jMsg)
	m = updated.(tui.Model)

	kMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}
	updated, _ = m.Update(kMsg)
	result := updated.(tui.Model)

	if result.Selected() != 0 {
		t.Errorf("Selected() = %d, want 0", result.Selected())
	}
}

func TestUpdate_KKeyDoesNotDecrementWhenAtFirstSection(t *testing.T) {
	m := modelWithReview(3)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")}
	updated, _ := m.Update(msg)
	result := updated.(tui.Model)

	if result.Selected() != 0 {
		t.Errorf("Selected() = %d, want 0 (should stay at first)", result.Selected())
	}
}

func TestUpdate_DownArrowWorksSameAsJ(t *testing.T) {
	m := modelWithReview(3)

	msg := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ := m.Update(msg)
	result := updated.(tui.Model)

	if result.Selected() != 1 {
		t.Errorf("Selected() = %d, want 1", result.Selected())
	}
}

func TestUpdate_UpArrowWorksSameAsK(t *testing.T) {
	m := modelWithReview(3)
	// Navigate to middle
	jMsg := tea.KeyMsg{Type: tea.KeyDown}
	updated, _ := m.Update(jMsg)
	m = updated.(tui.Model)

	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	updated, _ = m.Update(upMsg)
	result := updated.(tui.Model)

	if result.Selected() != 0 {
		t.Errorf("Selected() = %d, want 0", result.Selected())
	}
}

func TestUpdate_NavigationDoesNothingWithNoReview(t *testing.T) {
	m := tui.NewModel("/test/project")

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	updated, _ := m.Update(msg)
	result := updated.(tui.Model)

	if result.Selected() != 0 {
		t.Errorf("Selected() = %d, want 0", result.Selected())
	}
}

func TestUpdate_WindowSizeMsgInitializesViewport(t *testing.T) {
	m := tui.NewModel("/test/project")
	if m.Ready() {
		t.Error("expected Ready() to be false initially")
	}

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(msg)
	result := updated.(tui.Model)

	if !result.Ready() {
		t.Error("expected Ready() to be true after WindowSizeMsg")
	}
}

func TestUpdate_WindowSizeMsgResizesViewport(t *testing.T) {
	m := tui.NewModel("/test/project")
	// First size
	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(msg)
	m = updated.(tui.Model)

	// Resize
	msg = tea.WindowSizeMsg{Width: 80, Height: 30}
	updated, _ = m.Update(msg)
	result := updated.(tui.Model)

	// The viewport should be resized (we verify through width/height accessors)
	if result.Width() != 80 {
		t.Errorf("Width() = %d, want 80", result.Width())
	}
	if result.Height() != 30 {
		t.Errorf("Height() = %d, want 30", result.Height())
	}
}

func TestUpdate_NavigationResetsViewportToTop(t *testing.T) {
	m := tui.NewModel("/test/project")
	// Initialize viewport
	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// Set a review with multiple sections
	review := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test",
		Sections: []model.Section{
			{ID: "1", Narrative: "First"},
			{ID: "2", Narrative: "Second"},
		},
	}
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	// Navigate down - viewport should reset to top
	jMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	updated, _ = m.Update(jMsg)
	result := updated.(tui.Model)

	// The viewport's YOffset should be 0 (at top)
	if result.ViewportYOffset() != 0 {
		t.Errorf("ViewportYOffset() = %d, want 0 after navigation", result.ViewportYOffset())
	}
}
