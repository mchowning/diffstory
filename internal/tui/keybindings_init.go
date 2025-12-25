package tui

// initKeybindings registers all keybindings at startup
func initKeybindings() *KeybindingRegistry {
	r := NewKeybindingRegistry()

	// Global
	r.Register(Keybinding{Key: "q", Description: "Quit", Context: "global"})
	r.Register(Keybinding{Key: "?/esc", Description: "Toggle/close help", Context: "global"})
	r.Register(Keybinding{Key: "0", Description: "Focus Diff panel", Context: "global"})
	r.Register(Keybinding{Key: "1", Description: "Focus Section panel", Context: "global"})
	r.Register(Keybinding{Key: "2", Description: "Focus Files panel", Context: "global"})
	r.Register(Keybinding{Key: "h/l", Description: "Cycle panel focus", Context: "global"})
	r.Register(Keybinding{Key: "f", Description: "Cycle importance filter", Context: "global"})
	r.Register(Keybinding{Key: "G", Description: "Generate review (LLM)", Context: "global"})

	// Navigation
	r.Register(Keybinding{Key: "j/k", Description: "Navigate up/down", Context: "navigation"})
	r.Register(Keybinding{Key: "J/K", Description: "Scroll diff", Context: "navigation"})
	r.Register(Keybinding{Key: "C-j/k", Description: "Navigate files", Context: "navigation"})
	r.Register(Keybinding{Key: "</>", Description: "Jump to first/last", Context: "navigation"})
	r.Register(Keybinding{Key: ",/.", Description: "Page up/down", Context: "navigation"})

	// Files panel
	r.Register(Keybinding{Key: "enter", Description: "Expand/collapse directory", Context: "files"})

	return r
}
