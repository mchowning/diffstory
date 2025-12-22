---
date: 2025-12-22 14:39:18 EST
git_commit: 8d154762d21001fd6c87c3b75572b7289ffc46cc
branch: main
repository: diffguide
topic: "Keybinding Registry and Help Overlay Refactor"
tags: [implementation, tui, keybindings, help-overlay]
last_updated: 2025-12-22
---

# Keybinding Registry and Help Overlay Refactor

## Summary

Implemented Phase 1 of the standalone LLM review generation plan: a keybinding registry that auto-populates the help overlay, eliminating manual synchronization between key handlers and help text. Also added escape key support for closing the help overlay and improved MCP tool descriptions.

## Overview

The diffguide TUI previously had hardcoded help text in `view.go` that required manual updates whenever keybindings changed. This created a maintenance burden and risk of help text becoming stale. The solution introduces a `KeybindingRegistry` that serves as the single source of truth for all keyboard shortcuts, with the help overlay dynamically generated from this registry.

Additionally, the escape key was added as an intuitive way to dismiss the help overlay (in addition to the `?` toggle), and the MCP tool descriptions were enhanced to provide better guidance for LLM-based review generation.

## Technical Details

### Keybinding Registry

The registry provides a centralized store for keyboard shortcut metadata. Each keybinding has three properties: the key combination, a description, and a context (global, navigation, or files) used for grouping in the help display.

```go
// Keybinding represents a single keyboard shortcut
type Keybinding struct {
	Key         string
	Description string
	Context     string // "global", "navigation", "files"
}

// KeybindingRegistry holds all registered keybindings
type KeybindingRegistry struct {
	bindings []Keybinding
}
```

The registry supports registration and retrieval by context:

```go
// GetByContext returns keybindings for a specific context
func (r *KeybindingRegistry) GetByContext(ctx string) []Keybinding {
	var result []Keybinding
	for _, b := range r.bindings {
		if b.Context == ctx {
			result = append(result, b)
		}
	}
	return result
}
```

### Keybinding Initialization

All keybindings are registered at startup in `keybindings_init.go`. This centralizes the keybinding definitions and makes it straightforward to add new shortcuts:

```go
func initKeybindings() *KeybindingRegistry {
	r := NewKeybindingRegistry()

	// Global
	r.Register(Keybinding{Key: "q", Description: "Quit", Context: "global"})
	r.Register(Keybinding{Key: "?/esc", Description: "Toggle/close help", Context: "global"})
	// ... additional bindings

	return r
}
```

### Model Integration

The `Model` struct now holds a pointer to the registry, initialized in `NewModel`:

```go
type Model struct {
	// ... existing fields ...
	keybindings *KeybindingRegistry
}

func NewModel(workDir string) Model {
	return Model{
		workDir:      workDir,
		focusedPanel: PanelSection,
		keybindings:  initKeybindings(),
	}
}
```

### Dynamic Help Overlay

The `renderHelpOverlay` function was refactored to iterate over the registry and group keybindings by context, replacing the previous hardcoded string:

```go
func (m Model) renderHelpOverlay(base string) string {
	var sb strings.Builder
	sb.WriteString("Keybindings:\n\n")

	contexts := []struct {
		name  string
		title string
	}{
		{"global", "Global"},
		{"navigation", "Navigation"},
		{"files", "Files"},
	}

	for _, ctx := range contexts {
		bindings := m.keybindings.GetByContext(ctx.name)
		if len(bindings) == 0 {
			continue
		}
		sb.WriteString(ctx.title + "\n")
		for _, b := range bindings {
			sb.WriteString(fmt.Sprintf("  %-12s %s\n", b.Key, b.Description))
		}
		sb.WriteString("\n")
	}

	overlay := helpStyle.Render(strings.TrimSuffix(sb.String(), "\n"))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, overlay)
}
```

### Escape Key Support

The escape key was added as an additional way to close the help overlay. This provides a more intuitive UX since escape commonly dismisses modal overlays:

```go
case "esc":
	if m.showHelp {
		m.showHelp = false
	}
```

The keybinding registry reflects this with `"?/esc"` as the key for toggling/closing help.

### MCP Tool Description Improvements

The `submit_review` MCP tool description was enhanced to provide better guidance for structuring reviews. The changes emphasize narrative flow between sections and completeness of diff coverage:

- Sections should flow naturally into each other, building context progressively
- Order sections to tell a coherent story (setup before usage, core before peripheral)
- Include ALL changes - every diff hunk must appear in exactly one section

The `Narrative` field's jsonschema description was also updated to emphasize that narratives should connect smoothly to adjacent sections (`internal/model/review.go:11`).

## Git References

**Branch**: `main`

**Commit Range**: Uncommitted changes based on `8d154762d21001fd6c87c3b75572b7289ffc46cc`

**Files Changed**:

New files:
- `internal/tui/keybinding.go` - Keybinding struct and KeybindingRegistry
- `internal/tui/keybindings_init.go` - Keybinding initialization
- `internal/tui/keybinding_test.go` - Registry tests

Modified files:
- `internal/tui/model.go` - Added keybindings field to Model
- `internal/tui/view.go` - Refactored renderHelpOverlay to use registry
- `internal/tui/update.go` - Added escape key handler
- `internal/tui/update_test.go` - Added test for escape closing help
- `internal/mcpserver/mcpserver.go` - Improved tool description
- `internal/model/review.go` - Improved Narrative jsonschema description

---

*End of Implementation Summary*
