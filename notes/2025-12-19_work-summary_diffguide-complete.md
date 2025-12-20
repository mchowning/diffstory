---
date: 2025-12-19
git_commit: c39e79b
branch: main
repository: diffguide
topic: "Complete Work Summary: diffguide Implementation"
tags: [summary, tui, mcp, bubble-tea, go, implementation]
last_updated: 2025-12-19
---

# Complete Work Summary: diffguide Implementation

## Summary

diffguide is a terminal UI (TUI) application for viewing code reviews, built from scratch in Go over approximately two days of development. The project evolved from initial PRD through a complete MVP implementation, including HTTP server, file watching, syntax highlighting, and MCP (Model Context Protocol) integration for use with Claude Code.

## Project Overview

### What diffguide Does

diffguide displays code reviews in an interactive terminal interface. It receives structured review data (sections with narrative explanations and code diffs) and presents them in a two-pane layout:

- **Left pane**: List of narrative sections explaining the changes
- **Right pane**: Syntax-highlighted diffs for the selected section

The tool integrates with Claude Code via MCP, allowing Claude to submit code reviews that appear instantly in the viewer.

### Architecture

The project uses a **server + viewer architecture**:

```
┌─────────────────┐      POST /review        ┌──────────────────────┐
│   Claude Code   │ ──────────────────────── │   diffguide server   │
│  (MCP client)   │                          │   (single instance)  │
└─────────────────┘                          └──────────┬───────────┘
                                                        │
                                         Writes to ~/.diffguide/reviews/
                                                        │
                    ┌───────────────────────────────────┼───────────────────┐
                    │                                   │                   │
                    ▼                                   ▼                   ▼
          ┌─────────────────┐                ┌─────────────────┐  ┌─────────────────┐
          │ diffguide view  │                │ diffguide view  │  │ diffguide view  │
          │  (project A)    │                │  (project B)    │  │  (project C)    │
          └─────────────────┘                └─────────────────┘  └─────────────────┘
```

## Development Timeline

### Phase 1: Foundation (Initial PRD - December 18)

**Commit**: 4b1986e - Initial PRD

Created the product requirements document defining:
- Problem statement: disconnect between AI explanations and code diffs
- Two-pane lazygit-inspired interface design
- HTTP API specification for receiving reviews
- Keybinding conventions (j/k, J/K, q, ?)
- Non-functional requirements (startup <500ms, handle 100 sections)

### Phase 2: Research (December 18)

**Commit**: f4c9867 - Add research document

Researched and documented technology decisions:
- **Bubble Tea**: TUI framework with Elm-style Model-Update-View architecture
- **Lipgloss**: Terminal styling (colors, borders, layouts)
- **Chroma**: Syntax highlighting
- **fsnotify**: File watching for live updates
- Server+viewer architecture (vs. single HTTP server per instance)

Key insight: `tea.Send()` enables thread-safe message injection from HTTP handlers.

### Phase 3: MVP Planning (December 18)

**Commit**: 5905008 - Add MVP implementation plan

Created detailed implementation plan with 6 phases:
1. Project foundation with TUI skeleton
2. Server mode (HTTP → file writing)
3. Viewer mode (file watching → TUI display)
4. Two-pane layout with navigation
5. Syntax highlighting and diff colors
6. Scrolling and polish

Each phase included specific code changes, test requirements, and success criteria.

### Phase 4: TUI Skeleton Implementation (December 18)

**Commit**: c793c6a - Implement Phase 1: Project foundation with TUI skeleton

Implemented:
- Domain types (`Review`, `Section`, `Hunk`)
- Bubble Tea model with Init/Update/View
- ASCII art empty state display
- Quit handling (q, ctrl+c)
- Helper functions with tests

### Phase 5: Server Mode (December 18)

**Commits**: 6db93d1, 599d4b2, ad3e50d

Updated architecture to server+viewer model:
- HTTP server accepting POST /review with JSON payload
- Storage layer with SHA256-based directory hashing
- Atomic writes (temp file + rename) for safety
- Path normalization for consistent hashing
- Request body limits (10MB) and validation

### Phase 6: File Watching (December 18)

**Commit**: 4fd1c93 - Implement Phase 3: File watching for viewer mode

Implemented:
- fsnotify-based file watcher
- Watches `~/.diffguide/reviews/{dir-hash}.json`
- Handles file creation, modification, and deletion
- Channel-based communication with TUI
- Buffered channels to prevent deadlock

### Phase 7: Two-Pane Layout (December 18)

**Commit**: 6edad40 - Implement Phase 4: Two-pane layout with navigation

Implemented:
- Section list (left pane) with selection highlighting
- Diff content (right pane) using Bubbles viewport
- j/k and arrow key navigation
- Viewport scroll reset on section change
- Responsive layout based on terminal size

### Phase 8: Syntax Highlighting (December 18)

**Commits**: 6d75f80, e54662c

Implemented:
- Chroma-based syntax highlighting by file extension
- Diff colorization (green additions, red deletions, yellow headers)
- Text wrapping in section list (replaced truncation)
- Spacing between sections for readability

### Phase 9: Scrolling and Polish (December 18-19)

**Commits**: 3643109, 8b958a1, ae2633f

Implemented:
- J/K scrolling for diff pane
- Help overlay (? key)
- Status bar for error display
- Integration tests for full TUI workflow

### Phase 10: MCP Server (December 19)

**Commits**: d57ba63, c09550b

Research and implementation of MCP server:
- Created `internal/review/service.go` extracting shared business logic
- Refactored HTTP server to use review service
- Built MCP server using official Go MCP SDK
- Added `diffguide mcp` subcommand
- Created README with usage documentation

Key architectural decision: Both HTTP and MCP servers delegate to the same review service for consistent behavior.

### Phase 11: Bug Fix - Diff Truncation (December 19)

**Commit**: 84cfa5f, e62d015

Discovered and fixed issue where Claude Code was summarizing/truncating diffs:
- Root cause: Tool description didn't request complete diffs
- Fix: Updated tool description and JSON schema field descriptions
- Added explicit instructions: "Include COMPLETE diff content - do not truncate"

### Phase 12: Future Planning (December 19)

**Commit**: c39e79b - Add PRD for three-panel lazygit-style UI redesign

Created PRD for next major feature:
- Three-panel layout (Section, Files, Diff)
- File tree navigation with expand/collapse
- Context-sensitive diff display (all files vs. selected file)
- Lazygit-style panel switching (h/l, number keys)

## Technical Implementation Details

### Project Structure

```
diffguide/
├── cmd/diffguide/
│   ├── main.go      # Entry point with subcommand routing
│   ├── server.go    # HTTP server runner
│   └── mcp.go       # MCP server runner
├── internal/
│   ├── model/       # Review, Section, Hunk types
│   ├── storage/     # File-based persistence with hashing
│   ├── review/      # Shared business logic
│   ├── server/      # HTTP server implementation
│   ├── mcpserver/   # MCP server implementation
│   ├── watcher/     # fsnotify file watching
│   ├── tui/         # Bubble Tea TUI
│   │   ├── model.go
│   │   ├── update.go
│   │   ├── view.go
│   │   ├── styles.go
│   │   ├── helpers.go
│   │   └── messages.go
│   ├── highlight/   # Chroma syntax highlighting
│   └── integration/ # Integration tests
├── notes/           # Implementation summaries
└── working-notes/   # Research and plans
```

### Key Technologies

| Technology | Purpose |
|------------|---------|
| Go 1.24+ | Implementation language |
| Bubble Tea | TUI framework (Elm architecture) |
| Lipgloss | Terminal styling |
| Chroma | Syntax highlighting |
| fsnotify | File watching |
| MCP Go SDK | Claude Code integration |

### Data Flow

1. **HTTP/MCP Input** → Review service validates and normalizes
2. **Storage** → Writes to `~/.diffguide/reviews/{sha256(path)}.json`
3. **File Watcher** → Detects change, loads review
4. **TUI Update** → Receives `ReviewReceivedMsg`, updates model
5. **View** → Renders two-pane layout with syntax highlighting

### Testing Strategy

- **Unit tests**: Every package has comprehensive tests
- **Integration tests**: Full round-trip HTTP → storage → viewer
- **Table-driven tests**: Extensive input coverage
- **Race detection**: All tests pass with `-race` flag

## Commits Summary

| Commit | Description |
|--------|-------------|
| 4b1986e | Initial PRD |
| f4c9867 | Add research document |
| 5905008 | Add MVP implementation plan |
| c793c6a | Implement Phase 1: Project foundation |
| 6db93d1 | Update plan for server+viewer architecture |
| 599d4b2 | Implement Phase 2: Server Mode |
| ad3e50d | Add integration tests for Phase 2 |
| 4fd1c93 | Implement Phase 3: File watching |
| 6edad40 | Implement Phase 4: Two-pane layout |
| 6d75f80 | Implement Phase 5: Syntax highlighting |
| e54662c | Integrate diff colorization |
| 3643109 | Implement Phase 6: Scrolling and Polish |
| 8b958a1 | Mark Phase 6 complete |
| ae2633f | Add integration test for full TUI workflow |
| d57ba63 | Add MCP server for Claude Code integration |
| c09550b | Add README with usage documentation |
| 84cfa5f | Fix MCP tool diff truncation |
| e62d015 | Document MCP tool diff truncation fix |
| 2144fc9 | Add install location to CLAUDE.md |
| 9b9c6ab | Remove working-notes from version control |
| 09761da | Add lazygit as reference implementation |
| c39e79b | Add PRD for three-panel UI redesign |

## Current State

### Completed (MVP)

- TUI viewer with two-pane layout
- HTTP server for receiving reviews
- MCP server for Claude Code integration
- File watching for live updates
- Syntax highlighting for code
- Diff colorization (green/red/yellow)
- j/k navigation, J/K scrolling
- Help overlay
- Comprehensive test suite

### Planned (Next Version)

- Three-panel layout (Section, Files, Diff)
- File tree navigation
- Context-sensitive diff display
- Lazygit-style panel switching
- Scrollbars and position indicators

## Documentation

| Document | Purpose |
|----------|---------|
| `notes/2025-12-18_prd_initial.md` | Original product requirements |
| `notes/2025-12-19_prd_ui-refinement.md` | Three-panel UI PRD |
| `notes/2025-12-19_mcp-tool-diff-truncation-fix.md` | Bug fix documentation |
| `working-notes/2025-12-18_research_*.md` | Technology research |
| `working-notes/2025-12-18_plan_*.md` | MVP implementation plan |
| `working-notes/2025-12-19_*.md` | MCP implementation docs |
| `README.md` | User documentation |
| `CLAUDE.md` | Development instructions |
