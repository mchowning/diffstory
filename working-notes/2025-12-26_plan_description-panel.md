---
date: 2025-12-26 22:58:35 EST
git_commit: 8dbcc8f48b763eb886d1c84e5e226eeca599883a
branch: main
repository: diffguide
topic: "Description Panel for Section Narratives"
tags: [plans, tui, ui]
status: complete
last_updated: 2025-12-26
---

# Description Panel Implementation Plan

## Overview

Add a new "Description" panel positioned above the Diff panel on the right side. This panel displays the narrative text for the currently selected section, replacing the current inline expansion behavior in the Sections panel. This improves outline stability during navigation and makes narratives easier to read.

## Current State Analysis

The current implementation shows narrative text inline within the Sections panel when a section is selected:

- `view.go:266-293` - `renderSection()` renders selected sections with narrative text below the title
- `view.go:279-286` - Conditional logic wraps and displays narrative with `narrativePrefix` ("  │ ")
- This causes the outline to "jump" when navigating between sections

### Key Discoveries:

- `view.go:81-122` - `renderReviewState()` composes the three-panel layout
- `view.go:295-319` - `wrapText()` helper already exists for text wrapping
- `border.go:26-28` - `renderBorderedPanel()` creates bordered panels with optional active state
- `model.go:284-290` - Pattern for accessing selected section: `m.review.AllSections()[m.selected]`
- The panel does NOT need to be in the focus cycle (display-only per PRD)

## Desired End State

After implementation:
1. The Sections panel shows only chapter headers and section titles (no inline narrative)
2. A new "Description" panel appears above the Diff panel on the right side
3. The Description panel shows the narrative for the currently selected section
4. The Description panel height grows dynamically based on content (minimum 1 line)
5. Navigation between sections updates the Description panel immediately
6. The outline remains stable during navigation (no jumping)

### Verification:
- Navigate with `j`/`k` - sections outline should not expand/collapse
- Description panel should update to show current section's narrative
- Description panel should wrap text appropriately
- When no narrative exists, Description panel shows empty space

## What We're NOT Doing

- Scrolling within the Description panel (narratives are expected to be short)
- Toggle to hide/show the Description panel
- Making the Description panel focusable
- Adding keyboard shortcuts for the Description panel

## Safety Constraints

- **Minimum Diff height**: The Diff panel must always have at least 6 lines (to remain usable)
- **Description height cap**: If Description panel would leave less than 6 lines for Diff, cap the Description height

## Implementation Approach

Three phases that can be implemented and tested independently:
1. Remove inline narrative from Sections panel (simplifies outline)
2. Create Description panel rendering function
3. Integrate Description panel into layout composition

---

## Phase 1: Remove Inline Narrative from Sections Panel

### Overview

Simplify `renderSection()` to only render section titles, removing the conditional narrative expansion logic. This makes the outline stable during navigation.

### Changes Required:

#### 1. Simplify renderSection function

**File**: `internal/tui/view.go`
**Lines**: 266-293
**Changes**: Remove the narrative rendering logic (lines 279-286), keeping only title rendering

```go
func (m Model) renderSection(section model.Section, flatIdx, contentWidth int) string {
	isSelected := flatIdx == m.selected

	// Determine title to show (use Narrative as fallback if Title is empty)
	title := section.Title
	if title == "" {
		title = section.Narrative
	}

	if isSelected {
		return selectedStyle.Width(contentWidth).Render(selectedPrefix + title)
	}

	return normalStyle.Width(contentWidth).Render(normalPrefix + title)
}
```

### Success Criteria:

#### Automated Verification:
- [x] All tests pass: `nix develop -c go test ./...`
- [x] No linting errors: `nix develop -c golangci-lint run`
- [x] Build succeeds: `nix develop -c go build ./...`

#### Manual Verification:
- [ ] Sections panel shows only titles (no narrative expansion)
- [ ] Navigation with j/k does not cause outline to jump
- [ ] Selected section is still highlighted correctly

---

## Phase 2: Create Description Panel Rendering

### Overview

Add a new `renderDescriptionPane()` function that renders a bordered panel containing the narrative text for the currently selected section. The panel height is dynamic based on content.

### Changes Required:

#### 1. Add renderDescriptionPane function

**File**: `internal/tui/view.go`
**Location**: After `renderSection()` (around line 293)
**Changes**: Add new function to render the Description panel

```go
func (m Model) renderDescriptionPane(width, height int) string {
	var narrative string
	if m.review != nil {
		sections := m.review.AllSections()
		if m.selected < len(sections) {
			narrative = sections[m.selected].Narrative
		}
	}

	// Wrap narrative text to fit panel width (accounting for borders)
	contentWidth := width - 2
	var content string
	if narrative != "" {
		lines := wrapText(narrative, contentWidth)
		content = strings.Join(lines, "\n")
	}

	return renderBorderedPanel("Description", content, width, height, false)
}
```

#### 2. Add helper to calculate description panel height

**File**: `internal/tui/view.go`
**Location**: After `renderDescriptionPane()`
**Changes**: Add helper function to calculate required height

```go
func (m Model) descriptionPaneHeight(width, maxHeight int) int {
	const minHeight = 3 // 1 content line + 2 border lines

	if m.review == nil {
		return minHeight
	}

	sections := m.review.AllSections()
	if m.selected >= len(sections) {
		return minHeight
	}

	narrative := sections[m.selected].Narrative
	if narrative == "" {
		return minHeight
	}

	// Calculate wrapped line count
	contentWidth := width - 2 // account for borders
	lines := wrapText(narrative, contentWidth)

	// Height = wrapped lines + 2 for borders, capped at maxHeight
	height := len(lines) + 2
	if height < minHeight {
		height = minHeight
	}
	if height > maxHeight {
		height = maxHeight
	}
	return height
}
```

### Success Criteria:

#### Automated Verification:
- [x] All tests pass: `nix develop -c go test ./...`
- [x] No linting errors: `nix develop -c golangci-lint run`
- [x] Build succeeds: `nix develop -c go build ./...`

#### Manual Verification:
- [ ] (Deferred to Phase 3 - function not yet integrated into layout)

---

## Phase 3: Update Layout Composition

### Overview

Modify `renderReviewState()` to include the Description panel above the Diff panel. The right column becomes a vertical stack of Description + Diff panels.

### Changes Required:

#### 1. Update renderReviewState layout

**File**: `internal/tui/view.go`
**Lines**: 81-122
**Changes**: Calculate Description panel height, create right column with Description + Diff

```go
func (m Model) renderReviewState() string {
	if !m.ready {
		return "Initializing..."
	}

	// Three-panel layout:
	// Left column (1/3 width): Section (top) + Files (bottom)
	// Right column (2/3 width): Description (top) + Diff (bottom)
	leftWidth := m.width / 3
	rightWidth := m.width - leftWidth - 2 // account for borders

	contentHeight := m.height - 5 // header + footer + filter line
	timestampLine := m.renderTimestamp()
	if timestampLine != "" {
		contentHeight-- // account for timestamp line
	}
	sectionHeight := contentHeight / 2
	filesHeight := contentHeight - sectionHeight

	// Calculate description panel height (dynamic based on content)
	// Cap at (contentHeight - minDiffHeight) to ensure Diff panel remains usable
	const minDiffHeight = 6
	maxDescriptionHeight := contentHeight - minDiffHeight
	descriptionHeight := m.descriptionPaneHeight(rightWidth, maxDescriptionHeight)
	diffHeight := contentHeight - descriptionHeight

	sectionPane := m.renderSectionPane(leftWidth, sectionHeight)
	filesPane := m.renderFilesPane(leftWidth, filesHeight)

	// Join Section and Files vertically to create left column
	leftColumn := lipgloss.JoinVertical(lipgloss.Left, sectionPane, filesPane)

	// Create right column: Description (top) + Diff (bottom)
	descriptionPane := m.renderDescriptionPane(rightWidth, descriptionHeight)
	diffPane := m.renderDiffPaneWithTitle(rightWidth, diffHeight)
	rightColumn := lipgloss.JoinVertical(lipgloss.Left, descriptionPane, diffPane)

	header := headerStyle.Render("diffguide - " + m.review.Title)
	filterLine := m.renderFilterIndicator()
	footer := "j/k: navigate | J/K: scroll | h/l: panels | f: importance filter | t: test filter | q: quit | ?: help"
	if m.statusMsg != "" {
		footer = statusStyle.Render(m.statusMsg) + "  " + footer
	}

	// Join left column with right column horizontally
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn)

	if timestampLine != "" {
		return lipgloss.JoinVertical(lipgloss.Left, header, timestampLine, content, filterLine, footer)
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, content, filterLine, footer)
}
```

### Success Criteria:

#### Automated Verification:
- [x] All tests pass: `nix develop -c go test ./...`
- [x] No linting errors: `nix develop -c golangci-lint run`
- [x] Build succeeds: `nix develop -c go build ./...`

#### Manual Verification:
- [ ] Description panel appears above Diff panel on right side
- [ ] Description panel shows narrative for selected section
- [ ] Description panel updates when navigating with j/k
- [ ] Description panel wraps text appropriately to panel width
- [ ] Empty narrative shows empty Description panel (not collapsed)
- [ ] Diff panel remains usable (scrolling with J/K works)
- [ ] Layout looks correct at various terminal sizes

---

## Testing Strategy

### TDD Approach

Per project standards, TDD will be followed during implementation. Tests will be written before each code change. This section describes the verification approach, not the order of implementation.

### Existing Tests

- `TestView_SelectedSectionShowsTitleAndNarrative` (`view_test.go:202-213`) - This test checks that the narrative appears in the view. It will continue to pass because the narrative is still rendered (now in Description panel instead of inline).
- Existing navigation and scroll tests should continue to pass unchanged.

### New Tests to Write

During implementation, write behavior-driven tests verifying:

1. **Description panel content**: View output contains "Description" border title and narrative text for selected section
2. **Sections panel simplification**: Selected section in view does NOT contain `narrativePrefix` ("  │ ") lines
3. **Height calculation**: `descriptionPaneHeight` respects maxHeight parameter

### Integration Tests:

- None required - changes are display-only

### Manual Testing Steps:

1. Start diffguide with a review: `diffguide server` + load a review
2. Verify Description panel appears above Diff panel
3. Navigate between sections with `j`/`k` - verify Description updates
4. Verify Sections panel outline does not jump during navigation
5. Test with sections that have:
   - Short narrative (1 line)
   - Long narrative (multiple wrapped lines)
   - No narrative (empty Description panel)
6. Test with small terminal window to verify layout doesn't break
7. Verify Diff scrolling (J/K) still works correctly

## Performance Considerations

- `wrapText()` is called on every render for the Description panel
- This is the same function already used for inline narrative rendering
- No performance concerns - narratives are expected to be short

## Migration Notes

No migration required - this is a display-only change with no data model changes.

## References

- Original PRD: `working-notes/2025-12-26_prd_description-panel.md`
- Section rendering: `internal/tui/view.go:266-293`
- Layout composition: `internal/tui/view.go:81-122`
- Border utilities: `internal/tui/border.go:26-28`
- Text wrapping: `internal/tui/view.go:295-319`
