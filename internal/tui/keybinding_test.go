package tui_test

import (
	"testing"

	"github.com/mchowning/diffstory/internal/tui"
)

func TestKeybindingRegistry_EmptyRegistry(t *testing.T) {
	registry := tui.NewKeybindingRegistry()

	bindings := registry.GetAll()

	if len(bindings) != 0 {
		t.Errorf("expected empty slice, got %d bindings", len(bindings))
	}
}

func TestKeybindingRegistry_RegisterAndRetrieveSingle(t *testing.T) {
	registry := tui.NewKeybindingRegistry()
	binding := tui.Keybinding{
		Key:         "q",
		Description: "Quit",
		Context:     "global",
	}

	registry.Register(binding)
	bindings := registry.GetAll()

	if len(bindings) != 1 {
		t.Fatalf("expected 1 binding, got %d", len(bindings))
	}
	if bindings[0] != binding {
		t.Errorf("expected %v, got %v", binding, bindings[0])
	}
}

func TestKeybindingRegistry_RegisterAndRetrieveMultiple(t *testing.T) {
	registry := tui.NewKeybindingRegistry()
	binding1 := tui.Keybinding{Key: "q", Description: "Quit", Context: "global"}
	binding2 := tui.Keybinding{Key: "?", Description: "Toggle help", Context: "global"}
	binding3 := tui.Keybinding{Key: "j/k", Description: "Navigate up/down", Context: "navigation"}

	registry.Register(binding1)
	registry.Register(binding2)
	registry.Register(binding3)
	bindings := registry.GetAll()

	if len(bindings) != 3 {
		t.Fatalf("expected 3 bindings, got %d", len(bindings))
	}
	if bindings[0] != binding1 {
		t.Errorf("expected %v at index 0, got %v", binding1, bindings[0])
	}
	if bindings[1] != binding2 {
		t.Errorf("expected %v at index 1, got %v", binding2, bindings[1])
	}
	if bindings[2] != binding3 {
		t.Errorf("expected %v at index 2, got %v", binding3, bindings[2])
	}
}

func TestKeybindingRegistry_GetByContext(t *testing.T) {
	registry := tui.NewKeybindingRegistry()
	registry.Register(tui.Keybinding{Key: "q", Description: "Quit", Context: "global"})
	registry.Register(tui.Keybinding{Key: "?", Description: "Toggle help", Context: "global"})
	registry.Register(tui.Keybinding{Key: "j/k", Description: "Navigate up/down", Context: "navigation"})
	registry.Register(tui.Keybinding{Key: "enter", Description: "Toggle directory", Context: "files"})

	globalBindings := registry.GetByContext("global")
	navBindings := registry.GetByContext("navigation")
	filesBindings := registry.GetByContext("files")

	if len(globalBindings) != 2 {
		t.Fatalf("expected 2 global bindings, got %d", len(globalBindings))
	}
	if globalBindings[0].Key != "q" {
		t.Errorf("expected first global binding key 'q', got %q", globalBindings[0].Key)
	}
	if globalBindings[1].Key != "?" {
		t.Errorf("expected second global binding key '?', got %q", globalBindings[1].Key)
	}

	if len(navBindings) != 1 {
		t.Fatalf("expected 1 navigation binding, got %d", len(navBindings))
	}
	if navBindings[0].Key != "j/k" {
		t.Errorf("expected navigation binding key 'j/k', got %q", navBindings[0].Key)
	}

	if len(filesBindings) != 1 {
		t.Fatalf("expected 1 files binding, got %d", len(filesBindings))
	}
	if filesBindings[0].Key != "enter" {
		t.Errorf("expected files binding key 'enter', got %q", filesBindings[0].Key)
	}
}

func TestKeybindingRegistry_GetByContextNoMatches(t *testing.T) {
	registry := tui.NewKeybindingRegistry()
	registry.Register(tui.Keybinding{Key: "q", Description: "Quit", Context: "global"})

	bindings := registry.GetByContext("nonexistent")

	if len(bindings) != 0 {
		t.Errorf("expected empty slice for nonexistent context, got %d bindings", len(bindings))
	}
}
