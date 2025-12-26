---
date: 2025-12-26 16:36:10 EST
git_commit: 31ef2ee6246e03f061778c690054ba1d856c9db5
branch: main
repository: diffguide
topic: "Phase 2: Chapter Hierarchy Implementation"
tags: [implementation, tui, data-model, llm-prompt]
last_updated: 2025-12-26
---

# Phase 2: Chapter Hierarchy Implementation

## Summary

Implemented Phase 2 of the enhanced review sections feature, adding a chapter hierarchy to organize related review sections. This includes data model changes (Chapter and Section.Title), updated LLM prompt to generate chapters, and TUI rendering changes that display chapter headers with section titles that expand to show narratives only when selected.

## Overview

The diffguide TUI now organizes code review sections into logical chapters. Previously, reviews contained a flat list of sections, each displaying its full narrative. The new structure groups related sections under chapter headers, with non-selected sections showing only their short title and the selected section expanding to show both title and narrative.

The implementation spans three layers: the data model (adding Chapter struct and Section.Title field), the LLM prompt (instructing the model to output chapters containing sections), and the TUI rendering (displaying chapter headers as visual dividers with sections underneath). Navigation continues to work on flat section indices, making chapters purely visual groupings that don't affect keyboard navigation.

## Technical Details

### Data Model Changes

The core change replaces `Review.Sections []Section` with `Review.Chapters []Chapter`, where each chapter contains sections:

```go
type Chapter struct {
	ID       string    `json:"id" jsonschema_description:"Unique identifier for this chapter"`
	Title    string    `json:"title" jsonschema_description:"Short chapter title (~20-30 characters)"`
	Sections []Section `json:"sections" jsonschema_description:"Sections belonging to this chapter"`
}

type Section struct {
	ID        string `json:"id" jsonschema_description:"Unique identifier for this section"`
	Title     string `json:"title" jsonschema_description:"Short title for list display (~30-40 characters)"`
	Narrative string `json:"narrative" jsonschema_description:"Summary explaining what changed and why"`
	Hunks     []Hunk `json:"hunks" jsonschema_description:"Code changes belonging to this section"`
}
```

Helper methods on `Review` provide backward compatibility with code that expects flat section lists:

```go
func (r Review) AllSections() []Section {
	var sections []Section
	for _, ch := range r.Chapters {
		sections = append(sections, ch.Sections...)
	}
	return sections
}

func (r Review) SectionCount() int {
	count := 0
	for _, ch := range r.Chapters {
		count += len(ch.Sections)
	}
	return count
}
```

A convenience constructor `NewReviewWithSections()` wraps sections in a default "Changes" chapter, primarily used in tests:

```go
func NewReviewWithSections(workDir, title string, sections []Section) Review {
	return Review{
		WorkingDirectory: workDir,
		Title:            title,
		Chapters: []Chapter{
			{ID: "default", Title: "Changes", Sections: sections},
		},
	}
}
```

### LLM Prompt Updates

The prompt template in `internal/tui/generate.go` was updated to request chapters:

```go
const classificationPromptTemplate = `You are a code review assistant. Classify diff hunks into logical chapters and sections.
...
{
  "title": "Brief title for this review",
  "chapters": [
    {
      "id": "chapter-identifier",
      "title": "Short chapter title",
      "sections": [
        {
          "id": "section-identifier",
          "title": "Short section title",
          "narrative": "Concise explanation of what and why...",
          "hunks": [...]
        }
      ]
    }
  ]
}

Guidelines:
- Organize hunks into chapters, each containing related sections.
- Chapter title: ~20-30 characters, describes the theme (e.g., "Authentication", "Database Schema").
- Section title: ~30-40 characters, describes the specific change (e.g., "Add login endpoint handler").
...
```

### TUI Rendering Changes

The sections panel rendering in `internal/tui/view.go` was rewritten to iterate through chapters and render them as visual dividers:

```go
func (m Model) renderSectionPane(width, height int) string {
	// Track flat section index as we iterate through chapters
	flatIdx := 0
	startIdx := m.sectionScrollOffset
	...
	for _, chapter := range m.review.Chapters {
		// Render chapter header
		if flatIdx >= startIdx || (flatIdx < startIdx && startIdx < chapterEndIdx) {
			chapterHeader := chapterStyle.Width(contentWidth).Render(chapterPrefix + chapter.Title)
			items = append(items, chapterHeader)
		}

		// Render sections in this chapter
		for _, section := range chapter.Sections {
			if flatIdx >= startIdx && rendered < renderCount {
				items = append(items, m.renderSection(section, flatIdx, contentWidth))
				rendered++
			}
			flatIdx++
		}
	}
	...
}
```

Section rendering differentiates between selected and non-selected:

```go
func (m Model) renderSection(section model.Section, flatIdx, contentWidth int) string {
	isSelected := flatIdx == m.selected
	title := section.Title
	if title == "" {
		title = section.Narrative  // Fallback for sections without title
	}

	if isSelected {
		// Selected: show title with prefix, then narrative indented below
		titleLine := selectedStyle.Width(contentWidth).Render(selectedPrefix + title)
		if section.Narrative != "" && section.Narrative != title {
			narrativeLines := wrapText(section.Narrative, contentWidth-len(narrativePrefix))
			// ... render narrative lines with indent
		}
		return titleLine + "\n" + strings.Join(narrativeRendered, "\n")
	}

	// Non-selected: show only title
	return normalStyle.Width(contentWidth).Render(normalPrefix + title)
}
```

New styles for chapter headers in `internal/tui/styles.go`:

```go
chapterPrefix = "▼ "
chapterStyle  = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("81")) // Cyan

narrativePrefix = "  │ "
```

### Validation Changes

The LLM response validation in `internal/tui/validation.go` was updated to iterate through the new nested structure:

```go
type LLMResponse struct {
	Title    string       `json:"title"`
	Chapters []LLMChapter `json:"chapters"`
}

type LLMChapter struct {
	ID       string       `json:"id"`
	Title    string       `json:"title"`
	Sections []LLMSection `json:"sections"`
}

func validateClassification(inputHunks []diff.ParsedHunk, response LLMResponse) ValidationResult {
	// Triple-nested loop: chapters -> sections -> hunks
	for _, chapter := range response.Chapters {
		for _, section := range chapter.Sections {
			for _, hunk := range section.Hunks {
				outputIDs[hunk.ID]++
				// ...validate importance
			}
		}
	}
}
```

### Navigation Behavior

Navigation (`j`/`k` keys) operates on flat section indices via `m.selected`. Since chapters are visual-only dividers not tracked in selection, navigation automatically "skips" chapter headers. The section count indicator `[3/8]` counts only sections, not chapters.

### Test Updates

All test files were updated to use the chapter structure. The `NewReviewWithSections()` helper simplified many tests. Key files updated:
- `internal/tui/update_test.go` - Navigation and update tests
- `internal/tui/view_test.go` - New tests for chapter headers, title display
- `internal/tui/validation_test.go` - LLM response validation
- `internal/tui/generate_logic_test.go` - Review assembly tests
- Multiple other test files using chapter structure in fixtures

### Bug Fix

Fixed a nil pointer dereference in `updateViewportContent()` where `m.review.AllSections()` was called before checking if `m.review` was nil.

## Git References

**Branch**: `main`

**Commit Range**: These are uncommitted changes on top of `31ef2ee`

**Commits Documented**:

(Uncommitted changes - to be committed)

**Files Changed**: 27 files
- `internal/model/review.go` - Chapter struct, Section.Title, helper methods
- `internal/tui/generate.go` - LLM prompt and review assembly
- `internal/tui/view.go` - Section pane rendering with chapters
- `internal/tui/styles.go` - Chapter header styles
- `internal/tui/model.go` - AllSections() usage, nil check fix
- `internal/tui/validation.go` - LLMChapter type, nested validation
- `internal/tui/update.go` - SectionCount() usage
- `internal/review/service.go` - Nested validation loop
- `internal/server/server.go` - SectionCount() for logging
- `internal/mcpserver/mcpserver.go` - Chapters in SubmitReviewInput
- Multiple test files updated for chapter structure
- Multiple test fixture JSON files updated
