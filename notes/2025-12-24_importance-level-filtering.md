---
date: 2025-12-24 21:33:42 EST
git_commit: f14842dd964256b1ab579516e7d8d2acb8d781af
branch: main
repository: diffguide
topic: "Importance Level Filtering"
tags: [implementation, tui, filtering]
last_updated: 2025-12-24
---

# Importance Level Filtering

## Summary

Added filtering by importance level to the diffguide TUI, allowing users to show only hunks at or above a threshold (Low/Medium/High). The filter cycles with the `f` key and displays the current level in a status line.

## Overview

The diffguide TUI displays code review hunks organized by sections. Each hunk has an importance level assigned by the LLM (low, medium, or high). This implementation adds the ability to filter displayed hunks by importance threshold, reducing noise when reviewing large diffs.

Three filter levels are available:
- **Low (all)**: Shows all hunks regardless of importance
- **Medium**: Shows only medium and high importance hunks
- **High only**: Shows only high importance hunks

The current filter level is displayed in a status line above the footer. Pressing `f` cycles through the levels. The default filter level can be configured in the config file and defaults to "medium".

When filtering hides all hunks in a section, the section remains visible with a "(filtered)" indicator. Files with no visible hunks after filtering are hidden from the files panel entirely.

## Technical Details

### FilterLevel Type

A new `FilterLevel` type encapsulates the filtering logic in `internal/tui/filter.go`. The type provides three methods:

```go
type FilterLevel int

const (
	FilterLevelLow    FilterLevel = iota // Show all hunks
	FilterLevelMedium                    // Show medium + high importance
	FilterLevelHigh                      // Show only high importance
)

func (f FilterLevel) PassesFilter(importance string) bool {
	// Empty importance always passes (backward compatibility)
	if importance == "" {
		return true
	}
	switch f {
	case FilterLevelHigh:
		return importance == model.ImportanceHigh
	case FilterLevelMedium:
		return importance == model.ImportanceHigh || importance == model.ImportanceMedium
	default: // FilterLevelLow
		return true
	}
}
```

The `PassesFilter` method determines if a hunk should be shown given its importance string. Empty importance values always pass to maintain backward compatibility with existing review data that may not have importance set.

### Configuration Support

The `Config` struct in `internal/config/config.go` gained a `DefaultFilterLevel` field:

```go
type Config struct {
	LLMCommand          []string `json:"llmCommand"`
	DiffCommand         []string `json:"diffCommand"`
	DebugLoggingEnabled bool     `json:"debugLoggingEnabled"`
	DefaultFilterLevel  string   `json:"defaultFilterLevel"`
}
```

The config loader applies a default of "medium" when the field is empty (`internal/config/config.go:55`).

### Model Integration

The TUI Model stores the current filter level and initializes it from config in `NewModel()` (`internal/tui/model.go:119-130`). The filter level affects two operations:

1. **File tree building**: `extractFilteredFilePaths()` only includes files that have at least one hunk passing the filter (`internal/tui/model.go:325-335`)

2. **Panel height calculations**: Updated to account for the new filter indicator line by using `m.height - 5` instead of `m.height - 4` (`internal/tui/model.go:205,211`)

### View Rendering

The view layer applies filtering in multiple places:

**Filter indicator** (`internal/tui/view.go:124-126`):
```go
func (m Model) renderFilterIndicator() string {
	return "Diff filter: " + m.filterLevel.String()
}
```

**Section pane** shows "(filtered)" when all hunks are hidden (`internal/tui/view.go:164-168`):
```go
if !m.sectionHasVisibleHunks(section) {
	text += " (filtered)"
}
```

**Diff rendering** filters hunks in all three render methods:
- `renderDiffContent()` - full section view (`internal/tui/view.go:295`)
- `renderDiffForFile()` - single file view (`internal/tui/view.go:324`)
- `renderDiffForDirectory()` - directory view (`internal/tui/view.go:349`)

Each returns "(all hunks filtered)" when no hunks pass the filter.

### Keyboard Handling

The `f` key cycles the filter level (`internal/tui/update.go:147-152`):

```go
case "f":
	if m.review != nil {
		m.filterLevel = m.filterLevel.Next()
		m.updateFileTree()
		m.updateViewportContent()
	}
```

The handler rebuilds the file tree and updates the viewport content after changing the filter. The filter only operates when a review is loaded.

### Keybinding Registration

The `f` keybinding is registered in `internal/tui/keybindings_init.go:14` and appears in both the help overlay and the footer shortcut hints.

## Git References

**Branch**: `main`

**Status**: Uncommitted changes

**Files Changed**:
- `internal/tui/filter.go` (new)
- `internal/tui/filter_test.go` (new)
- `internal/config/config.go`
- `internal/config/config_test.go`
- `internal/tui/model.go`
- `internal/tui/view.go`
- `internal/tui/view_test.go`
- `internal/tui/update.go`
- `internal/tui/update_test.go`
- `internal/tui/keybindings_init.go`
