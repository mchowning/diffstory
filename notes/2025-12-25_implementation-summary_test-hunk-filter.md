---
date: 2025-12-25 08:16:22 EST
git_commit: 60a8c07f63339667e5089aa27036f998c842d469
branch: main
repository: diffguide
topic: "Test Hunk Filter Implementation"
tags: [implementation-summary, tui, filtering]
status: complete
related_plan: working-notes/2025-12-24_plan_test-hunk-filter.md
---

# Test Hunk Filter - Implementation Summary

## Overview

This implementation adds the ability to filter code review hunks by test vs production code in the diffguide TUI. Users cycle through three filter states (All, Excluding Tests, Only Tests) using the `t` key. The filter combines with the existing importance filter for compound filtering.

## Key Features

- **Test classification**: LLM categorizes each hunk with `isTest` boolean during review generation
- **Three filter states**: All → Excluding Tests → Only Tests (cycles with `t` key)
- **Compound filtering**: Hunks must pass both importance AND test filter to display
- **Backward compatible**: Existing reviews without `isTest` field continue to work (nil passes all filters)
- **Contextual indicators**: "Diff (filtered)" header appears only when current view has hidden content

## Files Changed

### Data Model

**`internal/model/review.go`**
- Added `IsTest *bool` field to `Hunk` struct with JSON tag `json:"isTest,omitempty"`

### Filter Logic

**`internal/tui/filter.go`**
- Added `TestFilter` type with three constants: `TestFilterAll`, `TestFilterExcluding`, `TestFilterOnly`
- `String()` method returns display labels ("", "Excluding Tests", "Only Tests")
- `Next()` method cycles through states
- `PassesFilter(isTest *bool)` returns true if hunk passes (nil always passes)

**`internal/tui/filter_test.go`**
- Added 5 tests for `TestFilter` type behavior

### TUI Model

**`internal/tui/model.go`**
- Added `testFilter TestFilter` field to Model struct
- Added `TestFilter()` accessor method
- Added `hunkPassesFilters(hunk)` helper combining both filters
- Updated `extractFilteredFilePaths()` to use compound filtering

### Key Handling

**`internal/tui/keybindings_init.go`**
- Registered `t` keybinding with description "Cycle test filter"

**`internal/tui/update.go`**
- Added `t` key handler that cycles `testFilter` and refreshes UI

**`internal/tui/update_test.go`**
- Added 2 tests for `t` key behavior

### View Rendering

**`internal/tui/view.go`**
- Updated footer hint: `f: importance filter | t: test filter`
- Updated `renderFilterIndicator()` to show compound display: "Diff filter: Medium | Excluding tests"
- Added `renderTestFilterIndicator()` helper
- Added `currentViewHasFilteredContent()` to determine if current view has hidden content
- Added `hunkInCurrentView()` to check if hunk belongs to current file/directory view
- Updated `renderDiffPaneWithTitle()` to show "Diff (filtered)" when content hidden
- Removed "(filtered)" suffix from section rendering (cleaner UI)
- Updated all 5 filter application points to use `hunkPassesFilters()`

**`internal/tui/view_test.go`**
- Updated `TestView_ShowsFilteredIndicatorWhenAllHunksHidden` to check diff header
- Added 5 Phase 2 tests for filter indicator refactor
- Added 6 Phase 3 tests for test filter view behavior

### LLM Integration

**`internal/tui/validation.go`**
- Added `IsTest *bool` field to `LLMHunkRef` struct

**`internal/tui/generate.go`**
- Updated prompt template with `isTest` field in JSON example
- Added guidelines for test code classification
- Updated `assembleReview()` to copy `IsTest` from LLM response

**`internal/tui/generate_logic_test.go`**
- Added 3 tests for `IsTest` field handling in review assembly

## Filter Application Points

The compound filter (`hunkPassesFilters`) is applied at 5 locations:

1. `view.go:sectionHasVisibleHunks()` - Section visibility checks
2. `view.go:currentViewHasFilteredContent()` - Diff header "(filtered)" indicator
3. `view.go:renderDiffContent()` - All-files diff view
4. `view.go:renderDiffForFile()` - Single-file diff view
5. `view.go:renderDiffForDirectory()` - Directory diff view
6. `model.go:extractFilteredFilePaths()` - File tree construction

## UI Behavior

### Filter Indicator Display

The filter indicator shows both filters when test filter is active:
- `Diff filter: Low (all)` (default, test filter at All)
- `Diff filter: High only | Excluding tests` (both filters active)
- `Diff filter: Medium | Tests only` (compound state)

### Diff Header

The diff panel header shows "(filtered)" only when the current view has hidden content:
- `[0] Diff` - No hidden content in current view
- `[0] Diff (filtered)` - Some hunks hidden by filters in current view

### Footer Shortcuts

Updated footer shows all filter shortcuts:
```
j/k: navigate | J/K: scroll | h/l: panels | f: importance filter | t: test filter | q: quit | ?: help
```

## Backward Compatibility

- `*bool` pointer type allows distinguishing nil (legacy) from false (production code)
- `TestFilter.PassesFilter(nil)` returns true for all filter states
- Existing review files without `isTest` field continue to work without modification

## Test Coverage

Total new tests added: 16
- 5 unit tests for `TestFilter` type
- 2 integration tests for `t` key handling
- 6 view tests for filter display behavior
- 3 tests for LLM `IsTest` field handling

All tests pass: `nix develop -c go test ./...`
