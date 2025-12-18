---
date: 2025-12-18
version: 1.0
---

# diffguide - Product Requirements Document

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

This document describes `diffguide`, a terminal-based tool for reviewing code changes presented as a narrative story.

- **Sections 1-3** explain the why and what at a high level
- **Section 4** provides concrete user scenarios
- **Section 5** lists specific features to implement (start here for implementation)
- **Sections 6-8** cover quality attributes, assumptions, and dependencies
- **Section 9** explicitly states what is NOT being built
- **Sections 10-13** provide technical guidance and risk awareness

Requirements are numbered (FR1, FR2, etc.) for easy reference during implementation and testing.

---

## 1. Summary

`diffguide` is a terminal user interface (TUI) application that presents code changes as a coherent narrative story. Instead of viewing diffs in arbitrary file order, users see a curated sequence of narrative sections, each associated with relevant code hunks. The tool receives review data via an HTTP endpoint (typically sent from Claude Code via a slash command) and displays it in a two-column lazygit-inspired interface: narrative sections on the left, associated diffs on the right.

---

## 2. Problem Statement

When using AI coding assistants like Claude Code, the assistant generates code changes and provides a high-level overview of what was done. However, there is a disconnect between this overview and the actual code changes. Users must manually correlate the AI's explanation with specific diffs, typically by switching between the AI assistant in one terminal panel and a tool like lazygit in another.

This manual correlation is tedious and error-prone. Users lose context switching between windows and must mentally map narrative descriptions to scattered code changes across multiple files.

---

## 3. Goals & Objectives

| Objective | Success Metric |
|-----------|----------------|
| Reduce context switching when reviewing AI-generated changes | User can view narrative and associated code in a single interface |
| Present changes as a coherent story | Changes are grouped by narrative section, not by file |
| Provide familiar UX for lazygit users | Keyboard shortcuts match lazygit conventions |
| Enable quick integration with Claude Code | HTTP endpoint accepts JSON payload; slash command can send data |

---

## 4. User Stories

**US1**: As a developer using Claude Code, I want to see the AI's explanation alongside the exact code it changed, so that I can quickly understand and verify the changes without switching windows.

**US2**: As a developer reviewing changes, I want to navigate through narrative sections using familiar lazygit keybindings, so that I don't have to learn a new interface.

**US3**: As a developer, I want to see syntax-highlighted diffs with clear addition/deletion markers, so that I can easily read the code changes.

**US4**: As a developer, I want to launch the tool and have it wait for review data, so that I can send reviews to it whenever I'm ready.

---

## 5. Functional Requirements

### Must Have (MVP)

| ID | Requirement |
|----|-------------|
| FR1 | Application launches and displays an empty state with instructions for how to send review data |
| FR2 | Application runs an HTTP server on a configurable port (default: 8765) that accepts POST requests with review data |
| FR3 | HTTP endpoint accepts JSON payload containing a title, and an array of sections, where each section has an id, narrative text, importance level, and array of code hunks |
| FR4 | Each code hunk contains a file path, start line number, and diff content |
| FR5 | A single hunk may appear in multiple sections |
| FR6 | TUI displays two columns: left column shows list of narrative sections, right column shows code hunks for the selected section |
| FR7 | User can navigate between sections using keyboard (j/k or arrow keys, matching lazygit) |
| FR8 | Code hunks display with syntax highlighting appropriate to the file type |
| FR9 | Diff content displays with visual differentiation for added lines (green) and removed lines (red) |
| FR10 | When new review data is received via HTTP, the current view is replaced immediately |
| FR11 | User can scroll the diff pane (right) independently using J/K keys |
| FR12 | User can quit the application using q key (matching lazygit) |

### Should Have (Post-MVP Priority)

| ID | Requirement |
|----|-------------|
| FR13 | MCP server wrapper around HTTP endpoint for self-describing Claude Code integration |
| FR14 | Toggle to filter sections, showing only "key" importance sections vs all sections |
| FR15 | Jump to file in external editor at the relevant line (matching lazygit behavior) |

### Could Have (Future Enhancements)

| ID | Requirement |
|----|-------------|
| FR16 | Stage/unstage individual hunks directly from the tool |
| FR17 | Collapse/expand code blocks within a section |
| FR18 | User-selectable git state (staged, unstaged, branch comparison) |
| FR19 | Web UI in addition to TUI |
| FR20 | Smooth animations for transitions (section selection, pane scrolling, content loading) |

### Won't Have (Explicitly Out of Scope)

| ID | Requirement |
|----|-------------|
| FR21 | Reviewing others' code / pull request integration |
| FR22 | Multi-user / collaborative features |
| FR23 | Generating the narrative (this is done externally by Claude Code) |

---

## 6. Non-Functional Requirements

| ID | Requirement |
|----|-------------|
| NFR1 | Application must start and display UI within 500ms |
| NFR2 | Application must handle review data with up to 100 sections and 500 total hunks without noticeable lag |
| NFR3 | HTTP endpoint must respond within 100ms |
| NFR4 | Application must be distributable as a single binary (no runtime dependencies) |
| NFR5 | Keyboard navigation must feel responsive (no perceptible delay) |

---

## 7. Assumptions

1. Claude Code (or another tool) will generate the narrative sections and associate them with code hunks before sending to diffguide
2. The user has a terminal that supports 256 colors for syntax highlighting
3. The user is familiar with lazygit or similar TUI navigation patterns
4. Review data will be sent as valid JSON matching the expected schema
5. The HTTP endpoint only needs to handle one client (local use)

---

## 8. Dependencies

| Dependency | Purpose |
|------------|---------|
| Go 1.21+ | Programming language |
| Bubble Tea | TUI framework |
| Lipgloss | TUI styling |
| Chroma | Syntax highlighting |
| Standard library `net/http` | HTTP server |

---

## 9. Non-Goals (Out of Scope)

The following are explicitly NOT part of this project:

1. **Generating narratives** - diffguide only displays narratives; generation is done by Claude Code or another LLM
2. **Git operations** - MVP does not read from git directly; it receives pre-processed data
3. **Pull request integration** - No GitHub/GitLab integration
4. **Remote/collaborative use** - Single user, local only
5. **Persisting reviews** - No database or file storage of past reviews
6. **Authentication** - HTTP endpoint is localhost only, no auth needed

---

## 10. Design Considerations

### UI Layout

```
+----------------------------------+----------------------------------------+
|  diffguide - [Review Title]      |                                        |
+----------------------------------+----------------------------------------+
|                                  |                                        |
|  > 1. Add JWT validation         |  src/middleware/auth.ts                |
|    2. Create user model          |  ─────────────────────────────────────  |
|    3. Update dependencies        |  @@ -15,6 +15,14 @@                    |
|                                  |   import { Request } from 'express';   |
|                                  |  +import { verify } from 'jsonwebtoken';|
|                                  |  +                                      |
|                                  |  +export const validateToken = (req) => {|
|                                  |  +  const token = req.headers.auth;     |
|                                  |  +  return verify(token, SECRET);       |
|                                  |  +};                                    |
|                                  |                                        |
+----------------------------------+----------------------------------------+
|  j/k: navigate | q: quit | ?: help                                       |
+-----------------------------------------------------------------------------+
```

### Key UI Principles

1. **Lazygit-inspired**: Two-pane layout with list on left, detail on right
2. **Keyboard-first**: All navigation via keyboard, no mouse required
3. **Clear visual hierarchy**: Selected section highlighted, diff colors for +/-

### Keybindings (lazygit-compatible)

| Key | Action |
|-----|--------|
| j / Down | Move to next section (left pane) |
| k / Up | Move to previous section (left pane) |
| J | Scroll diff down (right pane) |
| K | Scroll diff up (right pane) |
| q | Quit |
| ? | Show help |

---

## 11. Technical Considerations

### HTTP API

**Endpoint**: `POST /review`

**Request Body**:
```json
{
  "title": "Add user authentication",
  "sections": [
    {
      "id": "1",
      "narrative": "Add JWT token validation middleware to protect API routes. This middleware extracts the token from the Authorization header and verifies it against our secret.",
      "importance": "key",
      "hunks": [
        {
          "file": "src/middleware/auth.ts",
          "startLine": 15,
          "diff": "@@ -15,6 +15,14 @@\n import { Request } from 'express';\n+import { verify } from 'jsonwebtoken';\n+\n+export const validateToken = (req) => {\n+  const token = req.headers.auth;\n+  return verify(token, SECRET);\n+};"
        }
      ]
    },
    {
      "id": "2",
      "narrative": "Update package.json to include jsonwebtoken dependency.",
      "importance": "supporting",
      "hunks": [
        {
          "file": "package.json",
          "startLine": 12,
          "diff": "@@ -12,6 +12,7 @@\n   \"dependencies\": {\n     \"express\": \"^4.18.0\",\n+    \"jsonwebtoken\": \"^9.0.0\",\n   }"
        }
      ]
    }
  ]
}
```

**Response**: `200 OK` on success, `400 Bad Request` with error message on invalid JSON

### Architecture

```
┌─────────────────┐     HTTP POST      ┌─────────────────┐
│   Claude Code   │ ─────────────────> │    diffguide    │
│  (slash cmd)    │    /review         │   (TUI + HTTP)  │
└─────────────────┘                    └─────────────────┘
```

### Project Structure (suggested)

```
diffguide/
├── main.go              # Entry point, starts HTTP server and TUI
├── internal/
│   ├── api/
│   │   └── server.go    # HTTP server and handlers
│   ├── model/
│   │   └── review.go    # Data structures for review, section, hunk
│   ├── tui/
│   │   ├── app.go       # Main Bubble Tea model
│   │   ├── sections.go  # Left pane component
│   │   ├── diff.go      # Right pane component
│   │   └── styles.go    # Lipgloss styles
│   └── highlight/
│       └── syntax.go    # Chroma syntax highlighting wrapper
├── go.mod
└── go.sum
```

---

## 12. Technical Skills Required

- **Go programming**: Intermediate level, understanding of goroutines for concurrent HTTP server + TUI
- **Bubble Tea framework**: Basic familiarity with the Elm architecture (Model, Update, View)
- **HTTP APIs**: Basic understanding of REST endpoints and JSON handling
- **Terminal UI concepts**: Understanding of terminal rendering, ANSI colors

---

## 13. Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Bubble Tea learning curve slows development | Medium | Medium | Start with official examples; reference lazygit source for patterns |
| Syntax highlighting performance on large diffs | Low | Medium | Use Chroma's fast lexers; consider lazy rendering for off-screen content |
| HTTP server conflicts with TUI rendering | Low | High | Run HTTP server in separate goroutine; use channels to communicate with TUI |
| Diff format variations cause parsing issues | Medium | Low | Accept raw diff strings; don't parse internally for MVP |
| Port 8765 already in use | Low | Low | Make port configurable via flag or environment variable |
