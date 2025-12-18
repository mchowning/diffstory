---
date: 2025-12-18T12:55:32-05:00
git_commit: 4b1986eccbdbd317315b62404a0444e9ece6abde
branch: main
repository: diffguide
topic: "Diffguide Implementation Technologies Research"
tags: [research, bubble-tea, lipgloss, chroma, lazygit, tui, go]
last_updated: 2025-12-18
last_updated_note: "All open questions resolved"
---

# Research: Diffguide Implementation Technologies

**Date**: 2025-12-18T12:55:32-05:00
**Git Commit**: 4b1986eccbdbd317315b62404a0444e9ece6abde
**Branch**: main
**Repository**: diffguide

## Research Question

What technologies and patterns should we understand before implementing diffguide - a TUI application that receives diffs via HTTP and displays them in a lazygit-like two-pane interface with syntax highlighting?

## Summary

The implementation will use:
- **Bubble Tea** for the TUI framework (Elm-style architecture with Model-Update-View)
- **Lipgloss** for styling (colors, borders, layouts)
- **Chroma** for syntax highlighting (with TTY256/TTY16m formatters for terminal output)
- **Lazygit-style UX** patterns (j/k navigation, two-pane layout, vim-inspired keybindings)

Key architectural insight: Bubble Tea's `tea.Send()` method enables thread-safe message injection from HTTP handlers, making the HTTP+TUI integration straightforward.

## Detailed Findings

### 1. Bubble Tea Framework

**Source**: [github.com/charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea), [pkg.go.dev/github.com/charmbracelet/bubbletea](https://pkg.go.dev/github.com/charmbracelet/bubbletea)

#### Model-Update-View Architecture

Bubble Tea implements The Elm Architecture:

```go
type model struct {
    sections []Section     // Left pane: list of narrative sections
    selected int           // Currently selected section index
    // Right pane content is derived from sections[selected]
}

func (m model) Init() tea.Cmd {
    return nil  // No initial I/O
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Handle keyboard input
    case DiffReceivedMsg:
        // Handle incoming diff from HTTP
    }
    return m, nil
}

func (m model) View() string {
    return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
}
```

#### Keyboard Input Handling

```go
    case tea.KeyMsg:
        switch msg.String() {
        case "ctrl+c", "q":
            return m, tea.Quit
        case "up", "k":
            if m.selected > 0 {
                m.selected--
            }
        case "down", "j":
            if m.selected < len(m.sections)-1 {
                m.selected++
            }
        case "enter", " ":
            // Expand details or focus right pane
        }
```

#### External Message Injection (Critical for HTTP Integration)

The `tea.Send()` method enables thread-safe message injection from external goroutines:

```go
// Create custom message type for diffs
type DiffReceivedMsg struct {
    Sections []Section
}

// In main - store program reference
p := tea.NewProgram(initialModel())

// In HTTP handler - send message to TUI
func handleDiff(w http.ResponseWriter, r *http.Request) {
    var payload ReviewPayload
    json.NewDecoder(r.Body).Decode(&payload)

    // Thread-safe injection into Bubble Tea
    p.Send(DiffReceivedMsg{Sections: payload.Sections})
}
```

Key properties of `tea.Send()`:
- Thread-safe for concurrent access
- Blocking if called before program starts
- No-op if called after program exits
- Primary API for external communication

#### Bubbles Components

Use the **viewport** component for scrollable diff content:
- High-performance mode for alternate screen buffer
- Standard pager keybindings (PgUp/PgDown)
- Mouse wheel support

Use the **list** component for file selection:
- Built-in pagination
- Fuzzy filtering capability
- Auto-generated help

### 2. Lipgloss Styling

**Source**: [github.com/charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss)

#### Diff Colors

```go
// Diff-specific styles
var (
    additionStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#04B575")).  // Green
        Bold(true)

    deletionStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FF0000")).  // Red
        Bold(true)

    contextStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("240"))  // Gray

    hunkHeaderStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FFFF00")).  // Yellow
        Bold(true)
)

// Adaptive colors for light/dark terminals
adaptiveGreen := lipgloss.AdaptiveColor{
    Light: "#008000",  // Darker for light backgrounds
    Dark:  "#04B575",  // Brighter for dark backgrounds
}
```

#### Two-Pane Layout

```go
// Border styles
var (
    activeBorderStyle = lipgloss.NewStyle().
        BorderStyle(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("63"))  // Blue when focused

    inactiveBorderStyle = lipgloss.NewStyle().
        BorderStyle(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("240"))  // Gray when unfocused
)

// Layout composition
func (m model) View() string {
    termWidth, termHeight, _ := lipgloss.TerminalSize(os.Stdout)

    leftWidth := termWidth / 3
    rightWidth := termWidth - leftWidth - 2  // Account for borders

    leftPane := m.renderSectionList(leftWidth, termHeight)
    rightPane := m.renderSectionContent(rightWidth, termHeight)

    return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)
}
```

#### Responsive Sizing

```go
termWidth, termHeight, _ := lipgloss.TerminalSize(os.Stdout)

style := lipgloss.NewStyle().
    Width(termWidth - 4).   // Dynamic width
    MaxWidth(100).          // But cap at 100
    Height(termHeight - 2)  // Full height minus margins
```

### 3. Chroma Syntax Highlighting

**Source**: [github.com/alecthomas/chroma](https://github.com/alecthomas/chroma)

#### Lexer Detection by File Extension

```go
import (
    "github.com/alecthomas/chroma/v2/lexers"
    "github.com/alecthomas/chroma/v2/formatters"
    "github.com/alecthomas/chroma/v2/styles"
)

// Detect lexer from filename
lexer := lexers.Match("foo.go")
if lexer == nil {
    lexer = lexers.Fallback  // Plain text fallback
}

// Or by language name
lexer := lexers.Get("python")
```

#### Terminal Output Formatters

```go
// Available formatters:
// - "terminal"    (8-color)
// - "terminal16"  (16-color)
// - "terminal256" (256-color)
// - "terminal16m" (true-color/24-bit)

formatter := formatters.Get("terminal256")  // Good default
// OR for modern terminals:
formatter := formatters.Get("terminal16m")  // True color
```

#### Performance Optimization

```go
// Coalesce runs of identical tokens to reduce overhead
lexer = chroma.Coalesce(lexer)
```

#### Complete Highlighting Example

```go
func highlightCode(code, filename string) (string, error) {
    lexer := lexers.Match(filename)
    if lexer == nil {
        lexer = lexers.Fallback
    }
    lexer = chroma.Coalesce(lexer)

    style := styles.Get("monokai")
    if style == nil {
        style = styles.Fallback
    }

    formatter := formatters.Get("terminal256")

    iterator, err := lexer.Tokenise(nil, code)
    if err != nil {
        return "", err
    }

    var buf bytes.Buffer
    err = formatter.Format(&buf, style, iterator)
    return buf.String(), err
}
```

### 4. HTTP + TUI Concurrency Pattern

#### Architecture

```go
func main() {
    // Create the model and program
    m := initialModel()
    p := tea.NewProgram(m, tea.WithAltScreen())

    // Start HTTP server in goroutine
    go startHTTPServer(p)

    // Run TUI (blocking)
    if _, err := p.Run(); err != nil {
        log.Fatal(err)
    }
}

func startHTTPServer(p *tea.Program) {
    http.HandleFunc("/diff", func(w http.ResponseWriter, r *http.Request) {
        var payload DiffPayload
        if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        // Send to TUI - thread-safe!
        p.Send(DiffReceivedMsg{Payload: payload})

        w.WriteHeader(http.StatusOK)
    })

    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

#### Thread Safety

- `tea.Send()` is thread-safe - no mutex needed
- Messages are queued and processed in order by the Update function
- The TUI remains responsive during HTTP handling

### 5. Lazygit UX Patterns

**Source**: [github.com/jesseduffield/lazygit](https://github.com/jesseduffield/lazygit)

#### Keybinding Conventions

| Key | Action |
|-----|--------|
| `j` / `k` or `↑` / `↓` | Move up/down in list |
| `h` / `l` or `←` / `→` | Switch between panes |
| `PgUp` / `PgDn` | Page scroll in content pane |
| `Shift+J` / `Shift+K` | Scroll content pane (vim-style) |
| `<` / `>` | Jump to top/bottom |
| `q` | Quit |
| `?` | Show help |

#### Visual Design Principles

1. **Active border highlighting**: Green/bold for focused pane, gray for unfocused
2. **Rounded borders**: `lipgloss.RoundedBorder()` for modern look
3. **Selection highlighting**: Background color change for selected items
4. **Status line**: Bottom line showing context-sensitive help

#### Layout Configuration

```go
// Lazygit-style proportions
sidePanelWidth := 0.3333  // ~1/3 of screen for file list
mainPanelWidth := 0.6667  // ~2/3 for diff content
```

## Architecture Insights

### Recommended Project Structure

```
diffguide/
├── cmd/
│   └── diffguide/
│       └── main.go           # Entry point, HTTP + TUI setup
├── internal/
│   ├── tui/
│   │   ├── model.go          # Main Bubble Tea model
│   │   ├── update.go         # Update function (key handling, messages)
│   │   ├── view.go           # View function (rendering)
│   │   ├── styles.go         # Lipgloss styles
│   │   └── components/
│   │       ├── sectionlist.go # Left pane component
│   │       └── contentview.go # Right pane component
│   ├── api/
│   │   └── server.go          # HTTP server handling
│   ├── diff/
│   │   └── parser.go          # Parse incoming diff payloads
├── go.mod
└── go.sum
```

### Key Design Decisions

1. **Component Composition**: Embed Bubbles components (viewport, list) in main model and delegate Update/View calls
2. **Message Types**: Define clear message types for each external event (DiffReceived, FileSelected, etc.)
3. **Style Organization**: Group styles by component/state (focused/unfocused) in a dedicated styles.go
4. **Graceful Degradation**: Use adaptive colors for light/dark terminals, TTY256 formatter for broad compatibility

## Code References

- Bubble Tea basics: [github.com/charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)
- Bubbles components: [github.com/charmbracelet/bubbles](https://github.com/charmbracelet/bubbles)
- Lipgloss styling: [github.com/charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss)
- Chroma highlighting: [github.com/alecthomas/chroma](https://github.com/alecthomas/chroma)
- Lazygit keybindings: [Keybindings_en.md](https://github.com/jesseduffield/lazygit/blob/master/docs/keybindings/Keybindings_en.md)
- Lazygit config: [Config.md](https://github.com/jesseduffield/lazygit/blob/master/docs/Config.md)

## Decisions Made

### 1. Diff Format Within Hunks: Raw Unified Diff

**Decision**: The `diff` field within each hunk uses raw unified diff format (as already specified in the PRD).

**Rationale**:
- Minimizes burden on Claude Code - can use raw `git diff` output per hunk
- Standard format that's well-understood
- The TUI can use [`go-gitdiff`](https://github.com/bluekeyes/go-gitdiff) library to parse if needed
- GitHub and GitLab APIs use the same approach (raw patch strings)

**Example** (from PRD):
```json
{
  "file": "src/middleware/auth.ts",
  "startLine": 15,
  "diff": "@@ -15,6 +15,14 @@\n import { Request } from 'express';\n+import { verify } from 'jsonwebtoken';\n+\n+export const validateToken = (req) => {...}"
}
```

The overall payload structure (with narrative sections, importance levels, etc.) remains as defined in PRD section 11.

### 2. Syntax Highlighting: Built-in Chroma

**Decision**: Use `Chroma` library directly within the application.

**Rationale**:
- Zero dependencies for the user (single binary)
- More robust than managing external process pipes
- Chroma is Go-native and performant
- Avoids configuration complexity
- Consistent with PRD dependencies

**Implementation approach**:
- Parse the diff content to identify language
- Use Chroma lexers to tokenize code lines
- Apply diff-specific coloring (red/green backgrounds or prefixes) on top of syntax highlighting
- Render to the viewport buffer

### 3. Viewport Scrolling: Full Content

**Decision**: Load entire diff content into the Bubbles viewport component. No virtual scrolling or pagination.

**Rationale**:
- Bubbles viewport is designed for scrollable content and handles this well
- NFR2 scope (500 total hunks) means individual hunks are typically small
- External highlighter is the performance bottleneck, not rendering
- Lazygit uses this same approach
- Simpler implementation; can optimize later if profiling shows issues

### 4. Focus Management: Both h/l and Tab

**Decision**: Support both h/l (vim-style) and Tab/Shift+Tab for switching between panes.

**Rationale**:
- Lazygit uses h/l - maintains compatibility for existing users
- Tab is discoverable for non-vim users
- No conflict - Tab isn't used for anything else
- Low implementation cost - just additional key bindings

**Key bindings for pane focus**:
- `h` or `Shift+Tab`: Focus left pane (section list)
- `l` or `Tab`: Focus right pane (diff content)

### 5. Empty State: ASCII Art + Instructions

**Decision**: Display a branded ASCII art logo with instructional text when no review data has been received.

**Rationale**:
- More memorable/fun user experience
- Still satisfies FR1 (instructions for sending data)
- Shows personality for the tool

**Example layout**:
```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│       ╔╦╗╦╔═╗╔═╗╔═╗╦ ╦╦╔╦╗╔═╗                              │
│        ║║║╠╣ ╠╣ ║ ╦║ ║║ ║║║╣                               │
│       ═╩╝╩╚  ╚  ╚═╝╚═╝╩═╩╝╚═╝                              │
│                                                             │
│            Waiting for review data...                       │
│                                                             │
│       POST http://localhost:8765/review                     │
│                                                             │
│       curl -X POST http://localhost:8765/review \           │
│         -H "Content-Type: application/json" \               │
│         -d '{"title": "...", "sections": [...]}'            │
│                                                             │
│                    q: quit | ?: help                        │
└─────────────────────────────────────────────────────────────┘
```

### 6. Error Handling: Hybrid Approach

**Decision**: Use status bar for recoverable errors, modal for critical setup issues.

**Error handling by type**:

| Error Type | Display | HTTP Response |
|------------|---------|---------------|
| Malformed JSON | Status bar: "Error: Invalid JSON" | 400 Bad Request |
| Schema violation | Status bar: "Error: Missing 'title' field" | 400 Bad Request |
| Highlighter not found | Modal popup (blocking) | N/A (startup) |

**Rationale**:
- Non-blocking for recoverable errors - bad JSON doesn't interrupt workflow
- Blocking for critical setup issues - user must fix highlighter before proceeding
- Status bar messages clear on next valid request
- Follows lazygit pattern (status bar for transient, modals for important)

## Open Questions (All Resolved)

1. ~~**Diff parsing**~~: Resolved - using raw unified diff format within hunks.

2. ~~**Syntax highlighting scope**~~: Resolved - using internal Chroma library for zero-dependency distribution.

3. ~~**Viewport scrolling strategy**~~: Resolved - full content in viewport; optimize later if needed.

4. ~~**Focus management**~~: Resolved - both h/l and Tab/Shift+Tab switch panes.

5. ~~**Empty state**~~: Resolved - ASCII art logo + instructions (have some fun!).

6. ~~**Error handling**~~: Resolved - hybrid approach (status bar for recoverable, modal for critical).
