package tui_test

import (
	"strings"
	"testing"

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
