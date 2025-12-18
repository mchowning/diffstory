package tui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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
