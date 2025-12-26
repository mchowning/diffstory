---
date: 2025-12-26 09:29:52 EST
git_commit: 31ef2ee6246e03f061778c690054ba1d856c9db5
branch: main
repository: diffguide
version: 1.0
---

# PRD: Enhanced Review Sections

## How to Read This PRD

This document describes a two-phase enhancement to how diffguide generates and displays code review sections. Phase 1 focuses on prompt changes to improve section granularity and narrative quality. Phase 2 adds a chapter hierarchy to organize related sections. Each phase can be implemented and tested independently.

## Table of Contents

1. [Summary](#1-summary)
2. [Problem Statement](#2-problem-statement)
3. [Goals & Objectives](#3-goals--objectives)
4. [User Stories](#4-user-stories)
5. [Functional Requirements](#5-functional-requirements)
6. [Non-Functional Requirements](#6-non-functional-requirements)
7. [Assumptions](#7-assumptions)
8. [Dependencies](#8-dependencies)
9. [Non-Goals (Out of Scope)](#9-non-goals-out-of-scope)
10. [Design Considerations](#10-design-considerations)
11. [Technical Considerations](#11-technical-considerations)
12. [Technical Skills Required](#12-technical-skills-required)
13. [Risks & Mitigations](#13-risks--mitigations)

---

## 1. Summary

Enhance diffguide's LLM-generated review sections to be more granular and easier to understand. Phase 1 updates the LLM prompt to encourage smaller sections with clearer what/why narratives and explicit mention of decisions and alternatives. Phase 2 adds a chapter hierarchy that groups related sections, with an improved TUI that shows section titles inline and expands narratives only for the selected section.

## 2. Problem Statement

Currently, LLM-generated review sections sometimes combine two or more related but distinct changes into a single section. While the changes are related, they would be clearer if presented as separate sections that build on each other. Additionally, important decisions and considered alternatives are not consistently highlighted in the narrative, making it harder to understand the reasoning behind implementation choices.

The result is that reviewers occasionally need to re-read sections to fully understand them, and the logical progression of changes is not always clear.

## 3. Goals & Objectives

| Objective | Measure |
|-----------|---------|
| Reviews feel easier to follow | Subjective improvement in comprehension when reviewing diffs |
| Less re-reading required | Reduced time spent going back to re-read sections |
| Decisions are visible | Important implementation decisions and alternatives are surfaced in narratives |
| Clear section hierarchy | Related sections are visually grouped under chapter headers |

## 4. User Stories

1. **As a code reviewer**, I want each section to cover the smallest reasonable chunk of work so that I can understand one concept at a time without mental context-switching.

2. **As a code reviewer**, I want to see what decision was made and why when alternatives existed so that I understand the reasoning behind implementation choices.

3. **As a code reviewer**, I want related sections grouped under a common chapter heading so that I can see how sections relate to each other at a glance.

4. **As a code reviewer**, I want to scan section titles quickly and only expand the full narrative for the section I'm focused on so that the UI doesn't feel cluttered.

## 5. Functional Requirements

### Phase 1: Prompt Enhancements ✅ COMPLETE

Implemented in commit `31ef2ee` on 2025-12-26.

- ✅ FR1.1: Prompt encourages smaller, more granular sections
- ✅ FR1.2: Prompt requests sections that "build on each other"
- ✅ FR1.3: Narratives include both "what" and "why"
- ✅ FR1.4: Decisions/alternatives mentioned inline
- ✅ FR1.5: Narrative guidance allows 1-2 sentences
- ⏭️ FR1.6: Skipped - will revisit if LLM over-explains obvious changes

### Phase 2: Chapter Hierarchy

**Must Have:**

- FR2.1: Add `Chapter` concept to the data model with `id` and `title` fields
- FR2.2: Add `title` field to `Section` (short, ~30-40 characters, distinct from `narrative`)
- FR2.3: Update LLM prompt to generate chapters that group related sections
- FR2.4: Chapter titles must be short (~20-30 characters)
- FR2.5: TUI sections panel displays chapter headers as non-selectable divider rows
- FR2.6: Non-selected sections display as single-line showing only their `title`
- FR2.7: Selected section displays its `title` plus expanded `narrative` below it
- FR2.8: Keyboard navigation (`j`/`k`) skips chapter headers and moves only between sections

**Should Have:**

- FR2.9: Visual distinction between chapter headers and section rows (e.g., different color, bold, or indentation)

**Won't Have:**

- FR2.10: Collapsible chapters (may add later if needed)
- FR2.11: Selectable chapter headers (no clear action for selection)
- FR2.12: Backwards compatibility with reviews generated before this change

## 6. Non-Functional Requirements

- NFR1: LLM prompt changes must not significantly increase generation time
- NFR2: TUI must remain responsive with the new display logic
- NFR3: Chapter/section structure must be valid JSON that passes existing validation patterns

## 7. Assumptions

- The LLM (Claude or similar) can reliably follow updated prompt guidance for section granularity and narrative content
- Users are willing to regenerate existing reviews to get the new format (no migration needed)
- Terminal width of ~80+ characters is available for displaying section titles

## 8. Dependencies

- Phase 2 depends on Phase 1 being complete (the prompt changes establish the foundation for smaller sections)
- No external dependencies beyond the existing LLM integration

## 9. Non-Goals (Out of Scope)

- Three-panel TUI layout (explored but rejected as too complex)
- Separate outline panel for chapters
- Full tree hierarchy (nested chapters within chapters)
- Backwards compatibility with existing saved reviews
- Collapsible chapter groups
- Quantitative metrics or analytics on section sizes

## 10. Design Considerations

### Phase 1: No UI Changes

Phase 1 is purely prompt changes. The existing TUI displays sections the same way, but the content will be different (smaller sections, richer narratives).

### Phase 2: Sections Panel Layout

Current sections panel (simplified):
```
┌─[1] Sections [2/5]───────┐
│ ► Add login endpoint     │
│   Implements POST /login │
│   with bcrypt hashing... │
│   Add login tests        │
│   Tests the login...     │
└──────────────────────────┘
```

New sections panel with chapters:
```
┌─[1] Sections [3/8]───────┐
│ ▼ Authentication         │  ← chapter header (styled differently)
│   Define login types     │  ← section title (not selected)
│ ► Implement login handler│  ← SELECTED section title
│   │ Adds POST /login     │  ← narrative (indented, expanded)
│   │ endpoint. Uses       │
│   │ bcrypt over argon2   │
│   │ for wider support.   │
│   Add login tests        │  ← section title (not selected)
│ ▼ Database               │  ← chapter header
│   Add user migration     │
└──────────────────────────┘
```

Key visual elements:
- Chapter headers use a marker (e.g., `▼`) and possibly different styling
- Non-selected sections show only their short title on one line
- Selected section shows title on one line, narrative indented below
- Section indicator `[3/8]` counts only sections, not chapter headers

## 11. Technical Considerations

### Data Model Changes (Phase 2)

Current `LLMSection` in `validation.go`:
```go
type LLMSection struct {
    ID        string       `json:"id"`
    Narrative string       `json:"narrative"`
    Hunks     []LLMHunkRef `json:"hunks"`
}
```

New structure:
```go
type LLMChapter struct {
    ID       string       `json:"id"`
    Title    string       `json:"title"`
    Sections []LLMSection `json:"sections"`
}

type LLMSection struct {
    ID        string       `json:"id"`
    Title     string       `json:"title"`     // NEW: short title for list display
    Narrative string       `json:"narrative"`
    Hunks     []LLMHunkRef `json:"hunks"`
}
```

The top-level response changes from `sections: []` to `chapters: []`.

### Prompt Template Changes

Location: `/Volumes/workspace/diffguide/internal/tui/generate.go`

Phase 1 changes to the prompt guidelines section.

Phase 2 changes the expected JSON response format to include chapters.

### TUI Rendering Changes

Location: `/Volumes/workspace/diffguide/internal/tui/view.go`

The `renderSectionPane` function needs to:
1. Iterate through chapters, rendering each chapter header
2. For each section in a chapter, render either just the title or title + narrative based on selection state
3. Track which items are selectable (sections) vs display-only (chapters)
4. Adjust scroll offset calculations to account for variable row heights

### Navigation Changes

Location: `/Volumes/workspace/diffguide/internal/tui/update.go`

The `j`/`k` navigation logic needs to skip chapter header rows when moving selection.

## 12. Technical Skills Required

- Go programming
- Bubble Tea TUI framework (lipgloss for styling)
- LLM prompt engineering
- JSON schema design

## 13. Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| LLM doesn't follow granularity guidance consistently | Medium | Medium | Iterate on prompt wording; add examples in prompt if needed |
| LLM generates too many tiny sections | Low | Low | Prompt says "smallest reasonable chunk" not "smallest possible"; adjust if observed |
| Variable row heights make scroll calculation complex | Medium | Medium | Reference lazygit's scroll handling for similar patterns |
| Chapter groupings feel arbitrary or unhelpful | Low | Medium | Prompt guidance emphasizes logical grouping; can adjust based on feedback |
