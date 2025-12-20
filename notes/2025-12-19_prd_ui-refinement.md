---
date: 2025-12-19 19:04:01 EST
git_commit: 2144fc99116738d6d7b795f84486faae7aa967fa
branch: main
repository: diffguide
version: 1.1
---

# PRD: Three-Panel Lazygit-Style UI Redesign

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

This document describes a UI redesign for diffguide, transforming the current two-panel layout into a three-panel layout inspired by lazygit. Start with the Summary and Problem Statement to understand the "why," then review the Design Considerations section for visual mockups. The Functional Requirements section contains the detailed specifications needed for implementation.

---

## 1. Summary

Redesign diffguide's TUI from a two-panel layout (section list + diff) to a three-panel layout inspired by lazygit. The new layout features a Section panel (top-left), a collapsible File tree panel (bottom-left), and a Diff panel (right). This design makes better use of screen real estate, provides file-level navigation for reviewing individual changes, and follows familiar lazygit UI patterns that users already know.

---

## 2. Problem Statement

The current two-panel layout has usability issues:

1. **Wasted space**: When section narratives are short, the left panel appears mostly empty
2. **Constrained space**: When narratives are long, they compete with diff content for horizontal space
3. **No file-level navigation**: Users cannot easily focus on individual files within a section—they must scroll through all diffs
4. **Unfamiliar patterns**: The current navigation doesn't follow established TUI conventions

Users need a layout that efficiently uses terminal space, provides granular file-level navigation, and follows familiar patterns from tools like lazygit.

---

## 3. Goals & Objectives

| Objective | Measure of Success |
|-----------|-------------------|
| Improve space utilization | Left column is useful at all times (no empty space) |
| Enable file-level review | Users can view diff for a single selected file |
| Follow lazygit conventions | Navigation matches lazygit patterns (j/k, h/l, number keys) |
| Maintain context awareness | Users always know which section (N of M) and file they're viewing |

---

## 4. User Stories

1. **As a code reviewer**, I want to see which files are changed in the current section so that I can understand the scope of changes at a glance.

2. **As a code reviewer**, I want to view the diff for a single file so that I can focus on one change at a time without scrolling through unrelated diffs.

3. **As a code reviewer**, I want to expand and collapse directories in the file tree so that I can navigate large changesets efficiently.

4. **As a lazygit user**, I want navigation to work the same way (j/k, h/l, number keys) so that I don't have to learn new keybindings.

5. **As a code reviewer**, I want to see my position in lists (e.g., "2 of 5") so that I know how much content remains.

---

## 5. Functional Requirements

### 5.1 Panel Layout

#### Must Have

- **FR1**: Display three panels: Section (top-left), Files (bottom-left), Diff (right)
- **FR2**: Panel titles must appear in the border using lazygit style: `─[1]─Section─`
- **FR3**: Active panel must have a colored border (cyan/blue); inactive panels must have gray borders
- **FR4**: Each panel must have a unique number: Section = `[1]`, Files = `[2]`, Diff = `[0]`
- **FR5**: Left panels (Section + Files) must share approximately 1/3 of terminal width; Diff panel gets remaining 2/3

#### Should Have

- **FR6**: Panel proportions should be adjustable based on terminal size

### 5.2 Section Panel ([1])

#### Must Have

- **FR7**: Display current section's title and narrative text
- **FR8**: Display section position indicator in format `[N/M]` (e.g., `[2/5]`)
- **FR9**: When focused, `j` moves to next section; `k` moves to previous section
- **FR10**: When section changes, file selection in Files panel must reset to first file
- **FR11**: Section panel must wrap text that exceeds panel width (no truncation)

### 5.3 Files Panel ([2])

#### Must Have

- **FR12**: Display files from current section grouped by directory structure (tree view)
- **FR13**: Directories must be collapsible/expandable using `Enter` key
- **FR14**: Display expand/collapse indicators: `▼` for expanded, `▶` for collapsed
- **FR14a**: All directories must be expanded by default when a section loads
- **FR14b**: Only display directories that contain files with diffs (no empty directories)
- **FR14c**: Truncate deeply nested paths with ellipsis in the middle (e.g., `src/.../deep/file.ts`) when path exceeds panel width
- **FR15**: Display position indicator (e.g., `2 of 5`) showing current file position
- **FR16**: Display scrollbar on right edge indicating scroll position within the file list
- **FR17**: When focused, `j` moves selection down; `k` moves selection up
- **FR18**: Display Files panel even when section contains only one file
- **FR19**: Selected file must be visually highlighted (background color + prefix indicator)

#### Should Have

- **FR20**: Scrollbar should be a visual indicator (like lazygit's) not just text

### 5.4 Diff Panel ([0])

#### Must Have

- **FR21**: When Section panel is focused, display diffs for ALL files in current section
- **FR22**: When Files panel is focused, display diff for ONLY the selected file
- **FR22a**: Display a context header at the top of the Diff panel indicating current view mode:
  - When showing all files: `Viewing: All files`
  - When showing single file: `Viewing: src/auth/middleware.ts (2 of 5)`
- **FR23**: When Diff panel is focused, `j` scrolls down; `k` scrolls up
- **FR24**: `J` and `K` must ALWAYS scroll the diff panel regardless of which panel is focused
- **FR25**: Display scrollbar on right edge indicating scroll position within diff content
- **FR26**: Maintain existing diff syntax highlighting (green for additions, red for deletions, yellow for hunk headers)
- **FR27**: Display file path header above each file's diff content
- **FR28**: Display separator line between file path and diff content

#### Won't Have (This Version)

- **FR29**: Line numbers in diff display (deferred to future version)

### 5.5 Navigation

#### Must Have

- **FR30**: `j` / `k` navigate within the currently focused panel
- **FR31**: `h` / `l` cycle focus between left panels only (Section ↔ Files)
- **FR32**: `1` jumps focus directly to Section panel
- **FR33**: `2` jumps focus directly to Files panel
- **FR34**: `0` jumps focus directly to Diff panel
- **FR35**: `J` / `K` scroll the diff panel regardless of current focus
- **FR36**: `<` jumps to top of current panel's content
- **FR37**: `>` jumps to bottom of current panel's content
- **FR37a**: `,` pages up in the current panel
- **FR37b**: `.` pages down in the current panel
- **FR38**: `?` toggles help overlay (existing functionality)
- **FR39**: `q` quits the application (existing functionality)

#### Should Have

- **FR40**: Arrow keys should work as alternatives to vim-style keys (↑↓ for j/k, ←→ for h/l)

### 5.6 Visual Styling

#### Must Have

- **FR41**: Match lazygit's visual style: rounded borders, panel numbers in title bar
- **FR42**: Active panel border color: cyan/blue (consistent with lazygit)
- **FR43**: Inactive panel border color: gray
- **FR44**: Selected item styling: background highlight + prefix indicator (`›`)

---

## 6. Non-Functional Requirements

- **NFR1**: UI must remain responsive with up to 100 files in a section
- **NFR2**: Panel switching must feel instantaneous (< 16ms response time)
- **NFR3**: File tree expansion/collapse must be instantaneous
- **NFR4**: UI must gracefully handle terminal resize events
- **NFR5**: Minimum supported terminal size: 80 columns × 24 rows
- **NFR6**: Debounce diff panel re-renders during rapid file navigation (50-100ms) to prevent flicker while maintaining responsiveness

---

## 7. Assumptions

1. Users are familiar with vim-style navigation (j/k/h/l)
2. Users may have used lazygit and expect similar behavior
3. Review data structure (sections with hunks containing file paths) remains unchanged
4. Terminal supports Unicode characters for tree indicators (▼, ▶) and box-drawing

---

## 8. Dependencies

1. **Bubble Tea framework**: Existing dependency, no changes needed
2. **Lip Gloss**: Existing dependency for styling, no changes needed
3. **Review data model**: Current `model/review.go` structure must support extracting file paths from hunks

---

## 9. Non-Goals (Out of Scope)

1. **Line numbers in diff**: Explicitly deferred to future version
2. **Mouse support**: Not implementing click-to-focus or scroll
3. **Configurable keybindings**: Using fixed lazygit-style bindings
4. **Configurable colors/themes**: Using fixed color scheme
5. **Panel resize with keyboard**: Fixed proportions only
6. **Search/filter functionality**: Not in this version
7. **File status indicators**: Not showing modified/added/deleted status icons

---

## 10. Design Considerations

### 10.1 Panel Layout Mockup

**Section panel focused (showing all diffs):**

```
┌─[1]─Section────────────────────┬─[0]─Diff──────────────────────────────────┐
│ [2/5] Auth Token Validation    │ Viewing: All files                        │
│                                │ ──────────────────────────────────────────│
│ The auth flow needs to         │ src/auth/validate.ts                      │
│ validate tokens before         │ ──────────────────────────────────────────│
│ allowing access. This adds     │ + export function validateToken(token) {  │
│ a helper function.             │ +   if (!token) return false;             │
│                                │ +   return jwt.verify(token, SECRET_KEY); │
│                                │ + }                                       │
├─[2]─Files──────────────────────┤                                           │
│ ▼ src/auth                     │ src/auth/middleware.ts                    │
│     validate.ts                │ ──────────────────────────────────────────│
│     middleware.ts              │ - // TODO: add validation                 │
│     index.ts                   │ + const isValid = validateToken(req...);  │
│                                │                                           │
│                       1 of 3 ░ │                                        ░  │
└────────────────────────────────┴───────────────────────────────────────────┘
│ j/k: navigate | h/l: panels | 0-2: jump | J/K: scroll diff | </>: bounds   │
```

**Files panel focused (showing single file diff):**

```
┌─[1]─Section────────────────────┬─[0]─Diff──────────────────────────────────┐
│ [2/5] Auth Token Validation    │ Viewing: src/auth/middleware.ts (2 of 3)  │
│                                │ ──────────────────────────────────────────│
│ The auth flow needs to         │ @@ -15,7 +15,8 @@                         │
│ validate tokens before         │  import { getToken } from './store';      │
│ allowing access.               │                                           │
│                                │ - // TODO: add validation                 │
├─[2]─Files──────────────────────┤ + const isValid = validateToken(          │
│ ▼ src/auth                     │ +   req.headers.authorization             │
│     validate.ts                │ + );                                      │
│   › middleware.ts              │                                           │
│     index.ts                   │                                           │
│                                │                                           │
│                       2 of 3 ░ │                                        ░  │
└────────────────────────────────┴───────────────────────────────────────────┘
│ j/k: navigate | h/l: panels | 0-2: jump | J/K: scroll diff | </>: bounds   │
```

### 10.2 File Tree Structure

Files are grouped by directory structure. All directories are **expanded by default**, and only directories containing files with diffs are displayed:

```
▼ src/auth
    validate.ts
  › middleware.ts      <- selected
    index.ts
▼ tests
    auth.test.ts
```

User can press `Enter` to collapse a directory:

```
▼ src/auth
    validate.ts
  › middleware.ts      <- selected
    index.ts
▶ tests                <- collapsed by user
```

For deeply nested paths, truncate the middle when space is limited:

```
▼ src/.../components
    Button.tsx
    Modal.tsx
```

### 10.3 Keybinding Reference

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down / scroll down in focused panel |
| `k` / `↑` | Move up / scroll up in focused panel |
| `h` / `←` | Move focus to panel above (Section ↔ Files) |
| `l` / `→` | Move focus to panel below (Section ↔ Files) |
| `0` | Jump to Diff panel |
| `1` | Jump to Section panel |
| `2` | Jump to Files panel |
| `J` | Scroll diff down (always) |
| `K` | Scroll diff up (always) |
| `<` | Jump to top of current panel |
| `>` | Jump to bottom of current panel |
| `,` | Page up in current panel |
| `.` | Page down in current panel |
| `Enter` | Expand/collapse directory (in Files panel) |
| `?` | Toggle help overlay |
| `q` | Quit |

---

## 11. Technical Considerations

### 11.1 File Tree Construction

The current review model stores hunks with file paths:

```go
type Hunk struct {
    File      string
    StartLine int
    Diff      string
}
```

A new component will need to:
1. Extract unique file paths from section hunks
2. Parse paths into directory structure
3. Build a tree model with expand/collapse state
4. Render the tree with proper indentation and indicators

### 11.2 Panel State Management

The model will need to track:
- Currently focused panel (0, 1, or 2)
- Section panel: current section index, scroll offset
- Files panel: selected file index, expand/collapse state per directory, scroll offset
- Diff panel: scroll offset (via viewport)

### 11.3 Diff Content Switching

When focus changes between Section and Files panels, the diff content must update:
- Section focused → concatenate all hunks for the section
- Files focused → filter hunks to only the selected file

### 11.4 Suggested Implementation Order

1. Refactor layout to three panels with proper borders/titles
2. Implement panel focus switching (h/l, number keys)
3. Build file tree data structure and rendering
4. Implement file tree navigation (j/k, expand/collapse)
5. Implement context-sensitive diff display
6. Add position indicators and scrollbars
7. Add `<` / `>` jump-to-bounds navigation
8. Add `,` / `.` page up/down navigation

---

## 12. Technical Skills Required

- **Go**: Primary implementation language
- **Bubble Tea**: TUI framework (model-update-view pattern)
- **Lip Gloss**: Terminal styling library
- **Tree data structures**: For file tree implementation

---

## 13. Risks & Mitigations

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| File tree complexity for deep paths | Medium | Medium | Limit visible depth, truncate long paths |
| Panel sizing on narrow terminals | High | Medium | Implement minimum widths, graceful degradation |
| Performance with many files | Medium | Low | Lazy rendering, only render visible items |
| Breaking existing keybindings | High | Low | Keep `J/K` for diff scroll, `j/k` behavior changes only when new panels focused |
