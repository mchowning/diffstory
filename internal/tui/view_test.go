package tui_test

import (
	"strings"
	"testing"

	"github.com/mchowning/diffguide/internal/tui"
)

func TestView_EmptyStateContainsPort(t *testing.T) {
	m := tui.NewModel("8765")
	view := m.View()

	if !strings.Contains(view, "localhost") {
		t.Error("empty state view should contain 'localhost'")
	}
	if !strings.Contains(view, "8765") {
		t.Error("empty state view should contain port number")
	}
}

func TestView_EmptyStateContainsQuitInstruction(t *testing.T) {
	m := tui.NewModel("8765")
	view := m.View()

	if !strings.Contains(view, "q: quit") {
		t.Error("empty state view should contain 'q: quit'")
	}
}
