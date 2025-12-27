---
date: 2025-12-26 22:48:53 EST
git_commit: 70e525f66ab40509f693c5fb2caa9152edbbcb2c
branch: main
repository: diffguide
version: 1.0
---

# PRD: Separate Description Panel for Section Narratives

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

## How to Read This PRD

This document describes a UI change to the diffguide TUI application. The change involves adding a new panel and modifying an existing panel's behavior. Start with the Problem Statement to understand why this change is needed, then review the Design Considerations for the visual layout, and finally look at Functional Requirements for specific implementation details.

---

## 1. Summary

This feature separates the narrative text from the sections outline into a dedicated "Description" panel. Currently, when a section is selected, its narrative expands inline within the sections panel, disrupting the visual structure and causing the user to lose their place. The new design places the narrative in a fixed panel above the diff on the right side, keeping the sections outline stable and making the narrative easier to read.

---

## 2. Problem Statement

The current inline expansion of narrative text within the sections panel causes two problems:

1. **Loss of visual structure**: When a section is selected, its narrative text expands below the title, pushing other sections down. This makes it difficult to see the overall outline structure and understand where you are in the review.

2. **Disorientation during navigation**: When navigating between sections with `j`/`k`, the expansion/collapse of narratives causes the outline to "jump around," making it easy to lose your place.

These issues undermine the goal of providing a quick overview while keeping the narrative accessible.

---

## 3. Goals & Objectives

| Objective | Measure of Success |
|-----------|-------------------|
| Maintain spatial stability in the outline | User does not lose their place when navigating between sections |
| Improve narrative readability | Narrative text is easier to read than the current inline display |
| Preserve quick overview capability | User can see more of the outline structure at a glance |

---

## 4. User Stories

1. **As a reviewer**, I want the sections outline to remain stable as I navigate, so that I don't lose my place in the review structure.

2. **As a reviewer**, I want to see the narrative for the selected section in a dedicated space, so that I can read it without it disrupting the outline.

3. **As a reviewer**, I want to see the overall chapter/section structure at a glance, so that I understand where I am in the review.

---

## 5. Functional Requirements

### Must Have

- **FR1**: Add a new "Description" panel on the right side, positioned above the diff panel
- **FR2**: The Description panel displays the narrative text of the currently selected section
- **FR3**: The sections panel no longer expands to show narrative text inline - it shows only chapter headers and section titles
- **FR4**: The Description panel updates immediately when the user navigates to a different section
- **FR5**: The Description panel has a border matching the style of other panels, with the title "Description"
- **FR6**: The Description panel has a minimum height of 1 line
- **FR7**: The Description panel height grows dynamically to fit the narrative content (no maximum)
- **FR8**: When there is no narrative (e.g., chapter header selected, section has no narrative), the Description panel displays empty space
- **FR9**: The Description panel is not focusable - it is display-only with no keyboard interaction

### Should Have

- None identified

### Could Have

- None identified

### Won't Have

- **FR-NOT-1**: Scrolling within the Description panel (narratives are expected to be short)
- **FR-NOT-2**: Toggle to hide/show the Description panel
- **FR-NOT-3**: Maximum height constraint on the Description panel

---

## 6. Non-Functional Requirements

- **NFR1**: Panel rendering must remain responsive - no perceptible lag when navigating between sections
- **NFR2**: Text wrapping in the Description panel should respect the panel width

---

## 7. Assumptions

- Narrative text for sections is generally short (a few sentences to a paragraph)
- Users will primarily interact with the Sections and Diff panels; the Description panel is passive
- The reduction in diff panel height is an acceptable trade-off for improved outline stability

---

## 8. Dependencies

- Existing panel rendering infrastructure in `internal/tui/view.go`
- Border rendering utilities in `internal/tui/border.go`
- Styles defined in `internal/tui/styles.go`

---

## 9. Non-Goals (Out of Scope)

- Moving the Description panel to the left side (documented as an alternative approach below)
- Dynamic panel resizing based on which panel is active
- Any keyboard shortcuts for the Description panel
- Collapsing/hiding the Description panel

---

## 10. Design Considerations

### Current Layout

```
┌─────────────────────┬────────────────────────────────────────┐
│  [1] Sections       │  [0] Diff                              │
│  ▼ Chapter A        │  Viewing: {file}                       │
│  › Section 1        │  ─────────────────────────────────     │
│    │ Narrative...   │  {diff content}                        │
│    Section 2        │                                        │
├─────────────────────┤                                        │
│  [2] Files          │                                        │
│  › file.go          │                                        │
│                     │                                        │
├─────────────────────┴────────────────────────────────────────┤
│ Filter bar (full width)                                      │
├──────────────────────────────────────────────────────────────┤
│ Help/keybindings bar (full width)                            │
└──────────────────────────────────────────────────────────────┘
```

### New Layout (Implemented)

```
┌─────────────────────┬────────────────────────────────────────┐
│  [1] Sections       │  Description                           │
│  ▼ Chapter A        │  {narrative text for selected section} │
│  › Section 1        ├────────────────────────────────────────┤
│    Section 2        │  [0] Diff                              │
├─────────────────────┤  Viewing: {file}                       │
│  [2] Files          │  ─────────────────────────────────     │
│  › file.go          │  {diff content}                        │
│                     │                                        │
├─────────────────────┴────────────────────────────────────────┤
│ Filter bar (full width)                                      │
├──────────────────────────────────────────────────────────────┤
│ Help/keybindings bar (full width)                            │
└──────────────────────────────────────────────────────────────┘
```

### Key Visual Changes

1. **Sections panel**: No longer shows narrative text inline; only chapter headers and section titles
2. **Description panel**: New panel above the diff, shows narrative for selected section
3. **Diff panel**: Slightly shorter due to Description panel above it

### Alternative Layout (Not Implemented)

A third vertical panel on the left side was considered:

```
┌─────────────────────┬────────────────────────────────────────┐
│  [1] Sections       │  [0] Diff                              │
│  ▼ Chapter A        │  Viewing: {file}                       │
│  › Section 1        │  ─────────────────────────────────     │
├─────────────────────┤  {diff content}                        │
│  Description        │                                        │
│  {narrative text}   │                                        │
├─────────────────────┤                                        │
│  [2] Files          │                                        │
│  › file.go          │                                        │
├─────────────────────┴────────────────────────────────────────┤
│ Filter bar                                                   │
└──────────────────────────────────────────────────────────────┘
```

**Trade-offs:**

| Approach | Pros | Cons |
|----------|------|------|
| Right side (above diff) | More horizontal space for text; simpler implementation; dynamic height works naturally | Reduces diff visibility |
| Left side (third panel) | Keeps diff at full height | Needs active-panel-grows logic for usability; narrower text area; more complex implementation |

The right-side approach was chosen for simplicity. If diff space becomes a problem in practice, the left-side approach can be revisited.

---

## 11. Technical Considerations

### Panel Rendering Changes

1. **`renderSectionPane`**: Remove the logic that renders narrative text below selected sections. Only render chapter headers and section titles.

2. **New `renderDescriptionPane`**: Create a new function to render the Description panel. It should:
   - Get the narrative from the currently selected section
   - Wrap text to fit the panel width
   - Calculate height based on wrapped content (minimum 1 line)
   - Render with standard panel border and "Description" title

3. **`View` function**: Update the layout composition to:
   - Calculate Description panel height based on content
   - Render Description panel above Diff panel
   - Reduce Diff panel height accordingly

### Height Calculation

The Description panel height should be:
```
height = max(1, lines_needed_for_wrapped_narrative) + 2  // +2 for top/bottom border
```

### Panel Focus

The Description panel should not be added to the panel focus cycle. The existing `0`, `1`, `2` key bindings remain unchanged.

---

## 12. Technical Skills Required

- Go programming language
- Bubbletea TUI framework
- Lipgloss styling library
- Understanding of the existing diffguide panel rendering system

---

## 13. Risks & Mitigations

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Diff panel becomes too short with long narratives | High - diff is the most important panel | Medium | If this becomes a problem, implement the left-side alternative layout |
| Layout breaks on small terminal sizes | Medium | Low | Test with minimum reasonable terminal size; existing layout already has minimum size considerations |
