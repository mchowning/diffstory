---
date: 2026-01-07 14:25:05 EST
git_commit: 7b74d7fc57874f53882d876880753f3498edaeb0
branch: main
repository: diffstory
topic: "Editable Reviewer Instructions"
tags: [implementation, tui, generate-flow]
last_updated: 2026-01-07
---

# Editable Reviewer Instructions

## Summary

Enhanced the generate flow's context input screen to pre-fill the textarea with default reviewer instructions, giving users transparency into and control over the style guidance sent to the LLM.

## Overview

Previously, the generate flow showed a blank textarea labeled "Additional context (optional)" while hiding the actual style instructions that were hardcoded in the LLM prompt template. Users had no visibility into what guidance the LLM was receiving and could only add to it, not modify it.

This change moves the editable style instructions from the hardcoded prompt template into the visible textarea, pre-populated with defaults. The header now reads "Instructions for reviewer (editable)" to clearly indicate users can modify these instructions. Additional UI improvements include dynamic textarea sizing, removal of visual clutter (line numbers, cursor line highlighting), and a corrected help text (Alt+Enter for newlines, since terminals don't distinguish Shift+Enter).

## Technical Details

### Model Initialization (`internal/tui/model.go`)

The core change introduces a `DefaultReviewerInstructions` constant containing the style guidance that was previously hardcoded in the prompt template:

```go
const DefaultReviewerInstructions = "Prefer many small, focused sections over fewer large ones. Keep section narratives to 1-2 sentences explaining what and why. "
```

The trailing space allows users to immediately continue typing after the period without needing to add a space themselves.

The textarea initialization in `NewModel` was updated to pre-fill this content and configure the visual presentation:

```go
ctx.SetValue(DefaultReviewerInstructions)
ctx.SetHeight(calcTextareaHeight(ctx.Value(), 60))

// Remove cursor line background highlight to avoid highlighting entire wrapped text
focusedStyle, blurredStyle := textarea.DefaultStyles()
focusedStyle.CursorLine = lipgloss.NewStyle()
ctx.FocusedStyle = focusedStyle
ctx.BlurredStyle = blurredStyle
ctx.ShowLineNumbers = false
```

The cursor line highlight removal addresses an issue where the Bubble Tea textarea's default `CursorLine` style applies a background to the entire logical line. Since the default instructions are one line that wraps visually across multiple rows, this caused the entire text block to appear highlighted.

A helper function calculates the visual height needed for the textarea content:

```go
func calcTextareaHeight(text string, width int) int {
	if text == "" {
		return 1
	}
	lines := strings.Split(text, "\n")
	height := 0
	for _, line := range lines {
		if len(line) == 0 {
			height++
		} else {
			height += (len(line) + width - 1) / width
		}
	}
	return max(height, 1)
}
```

This accounts for soft-wrapping by dividing each line's character count by the textarea width.

### Context Input UI (`internal/tui/generate_ui.go`)

The header text was updated from "Additional context (optional)" to "Instructions for reviewer (editable)" at `internal/tui/generate_ui.go:125`.

The help text was corrected from "Shift+Enter" to "Alt+Enter" at `internal/tui/generate_ui.go:128`. This reflects a terminal limitation where Shift+Enter and Enter produce identical key codes, making them indistinguishable. The code already checked `msg.Alt` for Alt+Enter support.

Dynamic textarea resizing was added to `updateContextInput` so the textarea grows and shrinks as users type:

```go
m.contextInput.SetHeight(calcTextareaHeight(m.contextInput.Value(), 60))
```

### Prompt Template (`internal/tui/generate.go`)

Two lines were removed from the Guidelines section of `classificationPromptTemplate` since they now come through user context:

- `- Section narrative: 1-2 sentences explaining "what" and "why". Mention key decisions or alternatives if relevant.`
- `- Prefer multiple granular sections over fewer large ones.`

This avoids duplicate instructions when users keep the defaults, and allows the instructions to be modified or removed entirely.

The textarea content flows to the LLM via `GenerateParams.Context`, which is appended to the prompt as `\nUser context: <content>` at `internal/tui/generate.go:138-141`. On retry, the last context is preserved in `m.lastContext` and reused automatically.

## Git References

**Branch**: `main`

**Commit Range**: Single commit

**Commits Documented**:

**7b74d7fc57874f53882d876880753f3498edaeb0** (2026-01-07)
Enhance generate flow context input with editable reviewer instructions

- Pre-fill textarea with default instructions instead of blank placeholder
- Make header label "Instructions for reviewer (editable)" to indicate editability
- Dynamically size textarea to fit content (grow/shrink as user types)
- Remove cursor line background highlight to prevent visual distraction
- Hide line numbers for cleaner appearance
- Fix help text to show Alt+Enter (not Shift+Enter which terminals don't distinguish)
- Remove duplicate style instructions from prompt template since they now come via user context
