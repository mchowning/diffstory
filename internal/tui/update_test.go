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
