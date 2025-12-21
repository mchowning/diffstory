---
date: 2025-12-21 15:52:20 EST
git_commit: 70574bfd3f96519fb8388e4e389063f99fc1852b
branch: main
repository: diffguide
topic: "Three-Panel UI Polish: Borders, File Selection, and Diff Deduplication"
tags: [implementation, tui, lipgloss, borders, diff-view]
last_updated: 2025-12-21
---

# Three-Panel UI Polish: Borders, File Selection, and Diff Deduplication

## Summary

Completed visual polish for the three-panel lazygit-style UI: replaced partial Lipgloss borders with full box-drawing character borders featuring embedded titles, changed file selection to always control diff content regardless of panel focus, and deduplicated file headings when multiple hunks exist for the same file.

## Overview

This work addresses the final polish items from the three-panel UI implementation plan. The changes fall into three categories:

1. **Border rendering overhaul** - Lipgloss border styles could not render titles embedded in the top border (`╭─[1] Sections──╮`). A custom border renderer now uses Unicode box-drawing characters directly to achieve the lazygit aesthetic.

2. **File selection behavior change** - Previously, switching focus from Files panel to Sections panel would show all files in the diff, even though a specific file was still selected. This was confusing because the visual selection state didn't match the diff content. Now the selected file always controls what the diff displays.

3. **Diff view improvements** - When a section had multiple hunks for the same file, the file heading appeared repeatedly before each hunk. The heading now appears only once per file, with spacing between hunks.

## Technical Details

### Custom Border Rendering

Lipgloss border styles (`lipgloss.RoundedBorder()`) don't support embedding titles in the border line. The previous implementation used Lipgloss borders and manually prepended the title as content, resulting in misaligned panels when content contained ANSI color codes.

The new `renderBorderedPanel()` function builds borders character by character:

```go
func renderBorderedPanel(title, content string, width, height int, isActive bool) string {
	color := inactiveBorderColor
	if isActive {
		color = activeBorderColor
	}

	colorStyle := lipgloss.NewStyle().Foreground(color)

	innerWidth := width - 2 // account for left and right borders

	// Build top border with embedded title
	topBorder := buildTopBorder(title, innerWidth, colorStyle)

	// Build bottom border
	bottomBorder := colorStyle.Render(borderBottomLeft + strings.Repeat(borderHorizontal, innerWidth) + borderBottomRight)

	// Build content lines
	contentLines := strings.Split(content, "\n")
	contentHeight := height - 2 // account for top and bottom borders
	// ...
}
```

The key insight is using `lipgloss.Width()` instead of `len()` when measuring content width, because `lipgloss.Width()` correctly ignores ANSI escape sequences that would otherwise throw off the padding calculations (`internal/tui/border.go:71`).

### File Selection Always Controls Diff

The original plan (Phase 3) specified that diff content should change based on panel focus: show all files when Sections focused, show selected file when Files focused. In practice, this was confusing—you'd select a file, switch panels to navigate sections, and suddenly see different diff content even though your file selection hadn't changed.

The fix removes the focus check from `updateViewportContent()`:

```go
// Before (in model.go)
if m.focusedPanel == PanelFiles && m.flattenedFiles != nil && m.selectedFile < len(m.flattenedFiles) {
    selectedNode := m.flattenedFiles[m.selectedFile]
    // ...
}

// After
if m.flattenedFiles != nil && m.selectedFile < len(m.flattenedFiles) {
    selectedNode := m.flattenedFiles[m.selectedFile]
    // ...
}
```

The diff context header was also updated to always show the selected file/directory path regardless of focus (`internal/tui/view.go:166-176`).

### Deduplicated File Headings in Diff

When multiple hunks existed for the same file, the file path and separator line appeared before each hunk. Now file headings appear only once, with three empty lines separating hunks:

```go
func (m Model) renderDiffContent(section model.Section) string {
	var content strings.Builder
	var lastFile string

	for _, hunk := range section.Hunks {
		if hunk.File != lastFile {
			if lastFile != "" {
				content.WriteString("\n\n\n")
			}
			content.WriteString(hunk.File + "\n")
			content.WriteString(strings.Repeat("─", 40) + "\n")
			lastFile = hunk.File
		} else {
			content.WriteString("\n\n\n")
		}
		coloredDiff := highlight.ColorizeDiff(hunk.Diff)
		content.WriteString(coloredDiff + "\n")
	}

	return content.String()
}
```

The same pattern was applied to `renderDiffForFile()` and `renderDiffForDirectory()`.

### PRD Cleanup

Removed NFR6 (debounce requirement for diff panel re-renders) from the PRD since it was marked as deferred and not part of the current implementation scope.

## Git References

**Branch**: `main`

**Commit Range**: `6347614909b0...70574bfd3f96`

**Commits Documented**:

**a151daf24b27e4b6cf71ffc0c7b60fd24858e84a** (2025-12-21T07:54:15-05:00)
Remove deferred debounce requirement from PRD
NFR6 about debouncing diff panel re-renders was noted as deferred
and is not part of the current implementation scope. Cleaned up PRD
to reflect actual requirements.

**a81ce6d4ac02802c55d5dc2a37f6c3dc1aa283a2** (2025-12-21T09:23:11-05:00)
File selection always controls diff content, regardless of panel focus
Previously, switching focus from Files panel to Sections panel would show
all files again, even though a file was still selected. This was confusing
because the visual state didn't match the internal selection.

Now the diff always filters based on the selected file/directory in the
Files panel, regardless of which panel currently has focus. The context
header also always reflects the selected path.

**be43a1a837a56c8ec3ce1d228949850289bdd36c** (2025-12-21T09:23:35-05:00)
Deduplicate file headings in diff view with improved spacing
When multiple hunks exist for the same file, the file heading now appears
only once instead of repeating for each hunk. This reduces visual clutter
and makes the diffs easier to scan.

Spacing improvements:
- Three empty lines separate hunks within the same file
- Three empty lines separate different files in the diff view

Tests updated to verify:
- File headings appear only once for multiple hunks
- File selection controls diff regardless of panel focus
- Context header always shows selected file/directory

**70574bfd3f96519fb8388e4e389063f99fc1852b** (2025-12-21T12:33:58-05:00)
Add lazygit-style full box borders to panels
Replace partial borders with full box rendering using Unicode box-drawing characters.
Title is now embedded in the top border: ╭─[1] Sections───────╮

- New renderBorderedPanel() handles all border rendering
- Uses lipgloss.Width() to properly handle ANSI escape sequences in styled content
- Active/inactive panel state reflected in border color (cyan/gray)
- All three panels (Sections, Files, Diff) now have complete borders
- Removed unused legacy border styles and rendering functions
- Added comprehensive test coverage for border rendering

Fixes misaligned right borders when content has ANSI color codes.
