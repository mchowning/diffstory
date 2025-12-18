---
date: 2025-12-18T14:09:22-05:00
git_commit: f4c986788f47d53effc089bad8400994714d5883
branch: main
repository: diffguide
topic: "Diffguide MVP Implementation"
tags: [plans, tui, bubble-tea, go, mvp]
status: complete
last_updated: 2025-12-18
---

# Diffguide MVP Implementation Plan

## Overview

Implement the minimum viable product for diffguide - a TUI application that receives code review data via HTTP and displays it in a two-pane lazygit-style interface. The MVP covers FR1-FR12 from the PRD, enabling developers to view AI-generated narrative explanations alongside syntax-highlighted diffs.

## Current State Analysis

**What exists now (after Phase 1):**
- PRD document defining requirements
- Research document with technology decisions
- Nix flake for reproducible dev environment
- Go module with Bubble Tea, Lipgloss, Chroma dependencies
- Domain types: Review, Section, Hunk
- TUI model with Init/Update/View implementing tea.Model
- Empty state view with ASCII art and instructions
- Quit handling (q, ctrl+c)
- Entry point (cmd/diffguide/main.go)
- Tests for helpers, update, and view

**Key decisions from research:**
- Bubble Tea for TUI framework
- Lipgloss for styling
- Chroma for syntax highlighting
- `tea.Send()` for HTTP → TUI message injection
- Bubbles viewport and list components
- Raw unified diff format within hunks

## Desired End State

A single Go binary that:
1. Launches and displays an empty state with ASCII art and instructions (showing actual bound port)
2. Runs an HTTP server on port 8765 (configurable, supports `-port=0` for ephemeral)
3. Accepts POST /review with JSON payload
4. Displays two-pane interface: sections on left, diffs on right
5. Supports j/k/arrow navigation between sections
6. Shows syntax-highlighted code with green/red diff coloring
7. Supports J/K scrolling in the diff pane
8. Quits with 'q' and gracefully shuts down HTTP server

**Verification:**
```bash
# Start the application
./diffguide

# In another terminal, send test data
curl -X POST http://localhost:8765/review \
  -H "Content-Type: application/json" \
  -d '{"title": "Test Review", "sections": [...]}'

# Verify: TUI displays the review data with syntax highlighting
```

## What We're NOT Doing

- FR13-FR21 (post-MVP features)
- MCP server wrapper
- Section filtering by importance
- External editor integration
- Git operations
- File watching or auto-refresh
- Persistence of reviews
- Configuration files (just CLI flags for MVP)

## Implementation Approach

We'll build incrementally using TDD, with each phase producing working, testable software. The testing strategy uses:

1. **Unit tests** for Update function state transitions and View output
2. **Integration tests** with mocked `Dispatcher` interface for HTTP server testing
3. **Table-driven tests** for comprehensive input coverage
4. **Race detector** (`go test -race ./...`) for concurrency safety

Key architectural decisions:
- **Dispatcher interface**: Decouples HTTP server from `tea.Program` for testability
- **Viewport from Phase 3**: Prevents layout breakage with long content
- **Graceful shutdown**: HTTP server stops cleanly when TUI exits

The phases are ordered to deliver a working end-to-end flow as early as possible (Phase 2), then incrementally add features.

---

## Phase 1: Project Foundation ✓ COMPLETE

### Overview

Set up the Go project structure, create a minimal Bubble Tea TUI that can quit, and display the empty state with ASCII art and instructions.

### Changes Required:

#### 1. Go Module Initialization

**File**: `go.mod`

```go
module github.com/mchowning/diffguide

go 1.21

require (
    github.com/charmbracelet/bubbletea v1.2.4
    github.com/charmbracelet/bubbles v0.20.0
    github.com/charmbracelet/lipgloss v1.0.0
    github.com/alecthomas/chroma/v2 v2.14.0
)
```

#### 2. Domain Types

**File**: `internal/model/review.go`

Define the data structures for review payloads:

```go
package model

type Review struct {
    Title    string    `json:"title"`
    Sections []Section `json:"sections"`
}

type Section struct {
    ID         string `json:"id"`
    Narrative  string `json:"narrative"`
    Importance string `json:"importance"`
    Hunks      []Hunk `json:"hunks"`
}

type Hunk struct {
    File      string `json:"file"`
    StartLine int    `json:"startLine"`
    Diff      string `json:"diff"`
}
```

#### 3. Helper Functions

**File**: `internal/tui/helpers.go`

```go
package tui

// truncate shortens a string to maxLen, adding "…" if truncated
func truncate(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    if maxLen <= 1 {
        return "…"
    }
    return s[:maxLen-1] + "…"
}
```

#### 4. Main TUI Model

**File**: `internal/tui/model.go`

```go
package tui

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/mchowning/diffguide/internal/model"
)

type Model struct {
    review   *model.Review
    selected int
    width    int
    height   int
    port     string  // Actual bound port for display
}

func NewModel(port string) Model {
    return Model{port: port}
}

func (m Model) Init() tea.Cmd {
    return nil
}
```

#### 5. Update Function

**File**: `internal/tui/update.go`

```go
package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return m, tea.Quit
        }
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
    }
    return m, nil
}
```

#### 6. View Function with Empty State

**File**: `internal/tui/view.go`

```go
package tui

func (m Model) View() string {
    if m.review == nil {
        return m.renderEmptyState()
    }
    return m.renderReviewState()
}

func (m Model) renderEmptyState() string {
    // ASCII art logo and instructions
    // Include: POST http://localhost:{m.port}/review (actual bound port)
    // Include: q: quit | ?: help
}

func (m Model) renderReviewState() string {
    return "" // Placeholder for Phase 3
}
```

#### 7. Entry Point

**File**: `cmd/diffguide/main.go`

```go
package main

import (
    "flag"
    "log"
    "os"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/mchowning/diffguide/internal/tui"
)

func main() {
    port := flag.String("port", "8765", "HTTP server port (use 0 for ephemeral)")
    verbose := flag.Bool("v", false, "Enable verbose logging")
    flag.Parse()

    // For Phase 1, just pass the flag value; Phase 2 will pass actual bound port
    m := tui.NewModel(*port)
    p := tea.NewProgram(m, tea.WithAltScreen())

    if _, err := p.Run(); err != nil {
        log.Printf("Error: %v", err)
        os.Exit(1)
    }
}
```

### Success Criteria:

#### Automated Verification:
- [x] `go build ./...` compiles without errors
- [x] `go test ./...` passes all tests
- [x] `go test -race ./...` passes (no race conditions)
- [x] Unit test: Update with 'q' key - execute returned cmd and assert it yields `tea.QuitMsg`
- [x] Unit test: Update with 'ctrl+c' - execute returned cmd and assert it yields `tea.QuitMsg`
- [x] Unit test: View() when review is nil contains "localhost" and port
- [x] Unit test: View() when review is nil contains "q: quit"
- [x] Unit test: truncate("hello", 3) returns "he…"
- [x] Unit test: truncate("hi", 10) returns "hi"

#### Manual Verification:
- [ ] Running `./diffguide` displays ASCII art and instructions
- [ ] Pressing 'q' exits the application cleanly

---

## Phase 2: HTTP Server and Data Flow

### Overview

Add an HTTP server running in a goroutine that accepts POST /review requests, parses JSON payloads, and sends them to the TUI via a `Dispatcher` interface. This completes the end-to-end data flow with proper architecture for testability.

### Changes Required:

#### 1. Message Types

**File**: `internal/tui/messages.go`

```go
package tui

import "github.com/mchowning/diffguide/internal/model"

// ReviewReceivedMsg is sent when the HTTP server receives a review
type ReviewReceivedMsg struct {
    Review model.Review
}

// ErrorMsg is sent when an error occurs
type ErrorMsg struct {
    Err error
}

// ClearStatusMsg clears the status bar message
type ClearStatusMsg struct{}

// PortBoundMsg is sent when the HTTP server binds to a port
type PortBoundMsg struct {
    Port string
}
```

#### 2. Dispatcher Interface

**File**: `internal/api/dispatcher.go`

```go
package api

import tea "github.com/charmbracelet/bubbletea"

// Dispatcher sends messages to the TUI. This interface enables
// testing the HTTP server without a real tea.Program.
type Dispatcher interface {
    Send(msg tea.Msg)
}

// ProgramDispatcher wraps a tea.Program to implement Dispatcher
type ProgramDispatcher struct {
    program *tea.Program
}

func NewProgramDispatcher(p *tea.Program) *ProgramDispatcher {
    return &ProgramDispatcher{program: p}
}

func (d *ProgramDispatcher) Send(msg tea.Msg) {
    d.program.Send(msg)
}
```

#### 3. Update Function - Handle ReviewReceivedMsg

**File**: `internal/tui/update.go`

Add cases for new messages:

```go
case ReviewReceivedMsg:
    m.review = &msg.Review
    m.selected = 0
    return m, nil

case PortBoundMsg:
    m.port = msg.Port
    return m, nil
```

#### 4. HTTP Server with Proper Structure

**File**: `internal/api/server.go`

```go
package api

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net"
    "net/http"
    "time"

    "github.com/mchowning/diffguide/internal/model"
    "github.com/mchowning/diffguide/internal/tui"
)

type Server struct {
    dispatcher Dispatcher
    server     *http.Server
    listener   net.Listener
    verbose    bool
}

func NewServer(dispatcher Dispatcher, port string, verbose bool) (*Server, error) {
    mux := http.NewServeMux()

    s := &Server{
        dispatcher: dispatcher,
        verbose:    verbose,
        server: &http.Server{
            Addr:              "127.0.0.1:" + port,
            Handler:           mux,
            ReadHeaderTimeout: 2 * time.Second,
            ReadTimeout:       5 * time.Second,
            WriteTimeout:      5 * time.Second,
            IdleTimeout:       30 * time.Second,
        },
    }

    mux.HandleFunc("/review", s.handleReview)

    // Bind to get actual port (supports port=0 for ephemeral)
    ln, err := net.Listen("tcp", s.server.Addr)
    if err != nil {
        return nil, fmt.Errorf("failed to bind: %w", err)
    }
    s.listener = ln

    return s, nil
}

// Port returns the actual bound port
func (s *Server) Port() string {
    addr := s.listener.Addr().(*net.TCPAddr)
    return fmt.Sprintf("%d", addr.Port)
}

func (s *Server) Start() {
    // Notify TUI of actual bound port
    s.dispatcher.Send(tui.PortBoundMsg{Port: s.Port()})

    if s.verbose {
        log.Printf("HTTP server listening on 127.0.0.1:%s", s.Port())
    }

    // Panic recovery to prevent terminal corruption
    go func() {
        defer func() {
            if r := recover(); r != nil {
                log.Printf("HTTP server panic: %v", r)
            }
        }()

        if err := s.server.Serve(s.listener); err != http.ErrServerClosed {
            log.Printf("HTTP server error: %v", err)
        }
    }()
}

func (s *Server) Shutdown(ctx context.Context) error {
    if s.verbose {
        log.Printf("HTTP server shutting down")
    }
    return s.server.Shutdown(ctx)
}

func (s *Server) handleReview(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var review model.Review
    if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
        if s.verbose {
            log.Printf("Invalid JSON: %v", err)
        }
        http.Error(w, err.Error(), http.StatusBadRequest)
        s.dispatcher.Send(tui.ErrorMsg{Err: fmt.Errorf("invalid JSON: %w", err)})
        return
    }

    if s.verbose {
        log.Printf("Received review: %s (%d sections)", review.Title, len(review.Sections))
    }

    s.dispatcher.Send(tui.ReviewReceivedMsg{Review: review})
    w.WriteHeader(http.StatusOK)
}
```

#### 5. Updated Entry Point with Graceful Shutdown

**File**: `cmd/diffguide/main.go`

```go
package main

import (
    "context"
    "flag"
    "log"
    "os"
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/mchowning/diffguide/internal/api"
    "github.com/mchowning/diffguide/internal/tui"
)

func main() {
    port := flag.String("port", "8765", "HTTP server port (use 0 for ephemeral)")
    verbose := flag.Bool("v", false, "Enable verbose logging")
    flag.Parse()

    // Create model with placeholder port (will be updated by PortBoundMsg)
    m := tui.NewModel(*port)
    p := tea.NewProgram(m, tea.WithAltScreen())

    // Create dispatcher and server
    dispatcher := api.NewProgramDispatcher(p)
    server, err := api.NewServer(dispatcher, *port, *verbose)
    if err != nil {
        log.Fatalf("Failed to create server: %v", err)
    }

    // Start HTTP server
    server.Start()

    // Run TUI (blocking)
    if _, err := p.Run(); err != nil {
        log.Printf("Error: %v", err)
        os.Exit(1)
    }

    // Graceful shutdown of HTTP server
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := server.Shutdown(ctx); err != nil {
        log.Printf("Shutdown error: %v", err)
    }
}
```

### Success Criteria:

#### Automated Verification:
- [ ] `go build ./...` compiles without errors
- [ ] `go test ./...` passes all tests
- [ ] `go test -race ./...` passes (no race conditions)
- [ ] Unit test: POST /review with valid JSON returns 200 OK (using mock Dispatcher)
- [ ] Unit test: POST /review with invalid JSON returns 400 Bad Request
- [ ] Unit test: GET /review returns 405 Method Not Allowed
- [ ] Unit test: JSON payload parses correctly into Review struct
- [ ] Unit test: handleReview calls dispatcher.Send with ReviewReceivedMsg
- [ ] Unit test: handleReview calls dispatcher.Send with ErrorMsg on invalid JSON
- [ ] Unit test: Update with ReviewReceivedMsg sets m.review
- [ ] Unit test: Update with ReviewReceivedMsg sets m.selected to 0
- [ ] Unit test: Update with PortBoundMsg updates m.port
- [ ] Unit test: Server.Port() returns actual bound port (test with port=0)
- [ ] Unit test: Server gracefully shuts down when Shutdown() called

#### Manual Verification:
- [ ] `curl -X POST localhost:8765/review -d '{"title":"Test"}'` returns 200
- [ ] TUI updates to show review title after POST
- [ ] With `-port=0`, empty state shows actual ephemeral port
- [ ] With `-v`, server logs requests to stderr
- [ ] Pressing 'q' cleanly exits (no hanging goroutines)

---

## Phase 3: Two-Pane Layout with Navigation and Viewport

### Overview

Implement the two-pane layout with section list on the left and diff content on the right. **Critically, this phase includes Viewport integration** to prevent layout breakage when content exceeds terminal height. Add j/k/arrow key navigation between sections.

### Changes Required:

#### 1. Update Model with Viewport

**File**: `internal/tui/model.go`

```go
package tui

import (
    "github.com/charmbracelet/bubbles/viewport"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/mchowning/diffguide/internal/model"
)

type Model struct {
    review    *model.Review
    selected  int
    width     int
    height    int
    port      string
    viewport  viewport.Model
    ready     bool  // viewport initialized after first WindowSizeMsg
}

func NewModel(port string) Model {
    return Model{port: port}
}

func (m Model) Init() tea.Cmd {
    return nil
}
```

#### 2. Styles

**File**: `internal/tui/styles.go`

```go
package tui

import "github.com/charmbracelet/lipgloss"

var (
    // Pane borders
    activeBorderStyle = lipgloss.NewStyle().
        BorderStyle(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("63"))

    inactiveBorderStyle = lipgloss.NewStyle().
        BorderStyle(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("240"))

    // Section list - use prefix glyph for accessibility
    selectedPrefix = "› "
    normalPrefix   = "  "

    selectedStyle = lipgloss.NewStyle().
        Background(lipgloss.Color("62")).
        Foreground(lipgloss.Color("230"))

    normalStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("252"))

    // Header
    headerStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("230"))
)
```

#### 3. Update Navigation with Viewport Reset

**File**: `internal/tui/update.go`

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "ctrl+c":
            return m, tea.Quit
        case "j", "down":
            if m.review != nil && m.selected < len(m.review.Sections)-1 {
                m.selected++
                m.viewport.GotoTop()  // Reset scroll on section change
                m.updateViewportContent()
            }
        case "k", "up":
            if m.review != nil && m.selected > 0 {
                m.selected--
                m.viewport.GotoTop()  // Reset scroll on section change
                m.updateViewportContent()
            }
        }
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height

        // Initialize or resize viewport
        rightWidth := m.width - (m.width / 3) - 4
        viewportHeight := m.height - 4  // header + footer

        if !m.ready {
            m.viewport = viewport.New(rightWidth, viewportHeight)
            m.viewport.Style = lipgloss.NewStyle()
            m.ready = true
        } else {
            m.viewport.Width = rightWidth
            m.viewport.Height = viewportHeight
        }
        m.updateViewportContent()

    case ReviewReceivedMsg:
        m.review = &msg.Review
        m.selected = 0
        m.viewport.GotoTop()
        m.updateViewportContent()
        return m, nil

    case PortBoundMsg:
        m.port = msg.Port
        return m, nil
    }
    return m, nil
}

// updateViewportContent sets the viewport content based on selected section
func (m *Model) updateViewportContent() {
    if m.review == nil || m.selected >= len(m.review.Sections) {
        m.viewport.SetContent("")
        return
    }

    section := m.review.Sections[m.selected]
    content := m.renderDiffContent(section)
    m.viewport.SetContent(content)
}
```

#### 4. Two-Pane View with Viewport

**File**: `internal/tui/view.go`

```go
func (m Model) renderReviewState() string {
    if !m.ready {
        return "Initializing..."
    }

    leftWidth := m.width / 3
    rightWidth := m.width - leftWidth - 4 // borders

    leftPane := m.renderSectionList(leftWidth, m.height-4)
    rightPane := m.renderDiffPane(rightWidth, m.height-4)

    header := headerStyle.Render("diffguide - " + m.review.Title)
    footer := "j/k: navigate | J/K: scroll | q: quit | ?: help"

    content := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

    return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}

func (m Model) renderSectionList(width, height int) string {
    var items []string
    for i, section := range m.review.Sections {
        style := normalStyle
        prefix := normalPrefix
        if i == m.selected {
            style = selectedStyle
            prefix = selectedPrefix
        }
        // Truncate narrative for display (account for prefix)
        text := prefix + truncate(section.Narrative, width-len(prefix)-4)
        items = append(items, style.Render(text))
    }
    // Join and apply border
    content := strings.Join(items, "\n")
    return activeBorderStyle.Width(width).Height(height).Render(content)
}

func (m Model) renderDiffPane(width, height int) string {
    // Use viewport for scrollable content
    return inactiveBorderStyle.Width(width).Height(height).Render(m.viewport.View())
}

func (m Model) renderDiffContent(section model.Section) string {
    var content strings.Builder

    for _, hunk := range section.Hunks {
        content.WriteString(hunk.File + "\n")
        content.WriteString(strings.Repeat("─", 40) + "\n")
        content.WriteString(hunk.Diff + "\n\n")
    }

    return content.String()
}
```

### Success Criteria:

#### Automated Verification:
- [ ] `go build ./...` compiles without errors
- [ ] `go test ./...` passes all tests
- [ ] `go test -race ./...` passes (no race conditions)
- [ ] Unit test: View with review shows section list in left pane
- [ ] Unit test: View with review shows hunks in right pane (via viewport)
- [ ] Unit test: 'j' key increments selected when not at end
- [ ] Unit test: 'j' key does not increment when at last section
- [ ] Unit test: 'j' key resets viewport to top (GotoTop called)
- [ ] Unit test: 'k' key decrements selected when not at start
- [ ] Unit test: 'k' key does not decrement when at first section
- [ ] Unit test: 'k' key resets viewport to top
- [ ] Unit test: Down arrow works same as 'j'
- [ ] Unit test: Up arrow works same as 'k'
- [ ] Unit test: Selected section has "› " prefix
- [ ] Unit test: Non-selected sections have "  " prefix
- [ ] Unit test: WindowSizeMsg initializes viewport on first call
- [ ] Unit test: WindowSizeMsg resizes viewport on subsequent calls
- [ ] Unit test: Long diff content doesn't break layout (viewport clips)

#### Manual Verification:
- [ ] Two panes visible with proper proportions (~1/3 left, ~2/3 right)
- [ ] Section list shows narrative text with selection prefix
- [ ] Selected section visually highlighted with "› " prefix
- [ ] j/k and arrows navigate between sections
- [ ] Right pane updates to show hunks for selected section
- [ ] Navigation resets scroll position to top
- [ ] Long diffs don't overflow the pane (viewport clips them)
- [ ] Terminal resize updates layout correctly

---

## Phase 4: Syntax Highlighting and Diff Colors

### Overview

Integrate Chroma for syntax highlighting based on file extension. Apply diff-specific coloring (green for additions, red for deletions).

### Changes Required:

#### 1. Syntax Highlighter

**File**: `internal/highlight/syntax.go`

```go
package highlight

import (
    "bytes"

    "github.com/alecthomas/chroma/v2"
    "github.com/alecthomas/chroma/v2/formatters"
    "github.com/alecthomas/chroma/v2/lexers"
    "github.com/alecthomas/chroma/v2/styles"
)

func HighlightCode(code, filename string) (string, error) {
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
        return code, err
    }

    var buf bytes.Buffer
    if err := formatter.Format(&buf, style, iterator); err != nil {
        return code, err
    }

    return buf.String(), nil
}
```

#### 2. Diff Colorizer

**File**: `internal/highlight/diff.go`

```go
package highlight

import (
    "strings"

    "github.com/charmbracelet/lipgloss"
)

var (
    additionStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#04B575"))

    deletionStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FF0000"))

    hunkHeaderStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#FFFF00")).
        Bold(true)

    contextStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("240"))
)

func ColorizeDiffLine(line string) string {
    if len(line) == 0 {
        return line
    }

    switch line[0] {
    case '+':
        return additionStyle.Render(line)
    case '-':
        return deletionStyle.Render(line)
    case '@':
        return hunkHeaderStyle.Render(line)
    default:
        return contextStyle.Render(line)
    }
}

func ColorizeDiff(diff string) string {
    lines := strings.Split(diff, "\n")
    var result []string
    for _, line := range lines {
        result = append(result, ColorizeDiffLine(line))
    }
    return strings.Join(result, "\n")
}
```

#### 3. Update Diff Content Rendering

**File**: `internal/tui/view.go`

Update `renderDiffContent` to use highlighting:

```go
func (m Model) renderDiffContent(section model.Section) string {
    var content strings.Builder

    for _, hunk := range section.Hunks {
        content.WriteString(hunk.File + "\n")
        content.WriteString(strings.Repeat("─", 40) + "\n")

        // Apply diff coloring
        coloredDiff := highlight.ColorizeDiff(hunk.Diff)
        content.WriteString(coloredDiff + "\n\n")
    }

    return content.String()
}
```

### Success Criteria:

#### Automated Verification:
- [ ] `go build ./...` compiles without errors
- [ ] `go test ./...` passes all tests
- [ ] `go test -race ./...` passes (no race conditions)
- [ ] Unit test: HighlightCode with "foo.go" returns string containing ANSI escape sequences
- [ ] Unit test: HighlightCode with "foo.py" returns string containing ANSI escape sequences
- [ ] Unit test: HighlightCode with unknown extension returns original code (no error)
- [ ] Unit test: ColorizeDiffLine with '+' prefix returns string containing ANSI green
- [ ] Unit test: ColorizeDiffLine with '-' prefix returns string containing ANSI red
- [ ] Unit test: ColorizeDiffLine with '@' prefix returns string containing ANSI yellow
- [ ] Unit test: ColorizeDiffLine with ' ' prefix returns string containing ANSI gray
- [ ] Unit test: ColorizeDiff colors multiple lines correctly

#### Manual Verification:
- [ ] Code in diffs is syntax highlighted based on file extension
- [ ] Added lines (+) appear in green
- [ ] Deleted lines (-) appear in red
- [ ] Hunk headers (@@) appear in yellow
- [ ] Context lines appear in gray
- [ ] Colors visible in 256-color terminal

---

## Phase 5: Scrolling and Polish

### Overview

Add J/K scrolling for the diff pane, implement help overlay with '?', and add error handling with status bar display.

### Changes Required:

#### 1. Update Model with Status and Help

**File**: `internal/tui/model.go`

Add fields:

```go
type Model struct {
    review    *model.Review
    selected  int
    width     int
    height    int
    port      string
    viewport  viewport.Model
    ready     bool
    showHelp  bool
    statusMsg string
}
```

#### 2. Update Function - Scrolling and Help

**File**: `internal/tui/update.go`

Add key handling:

```go
case "J":
    m.viewport.LineDown(1)
case "K":
    m.viewport.LineUp(1)
case "?":
    m.showHelp = !m.showHelp
```

Add message handling:

```go
case ErrorMsg:
    m.statusMsg = "Error: " + msg.Err.Error()
    return m, tea.Tick(3*time.Second, func(time.Time) tea.Msg {
        return ClearStatusMsg{}
    })

case ClearStatusMsg:
    m.statusMsg = ""
    return m, nil
```

#### 3. Help Overlay Style

**File**: `internal/tui/styles.go`

```go
var helpStyle = lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("63")).
    Padding(1, 2)

var statusStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("203"))
```

#### 4. Help Overlay and Status Bar

**File**: `internal/tui/view.go`

```go
func (m Model) View() string {
    if m.review == nil {
        return m.renderEmptyState()
    }

    base := m.renderReviewState()

    // Overlay help if toggled
    if m.showHelp {
        return m.renderHelpOverlay(base)
    }

    return base
}

func (m Model) renderReviewState() string {
    // ... existing code ...

    // Add status bar to footer if present
    footer := "j/k: navigate | J/K: scroll | q: quit | ?: help"
    if m.statusMsg != "" {
        footer = statusStyle.Render(m.statusMsg) + "  " + footer
    }

    // ... rest of function ...
}

func (m Model) renderHelpOverlay(base string) string {
    help := `Keybindings:

  j/k or ↑/↓    Navigate sections
  J/K           Scroll diff pane
  q             Quit
  ?             Toggle this help

HTTP API:

  POST /review  Send review data`

    overlay := helpStyle.Render(help)

    // Center the overlay on the base view
    return lipgloss.Place(
        m.width, m.height,
        lipgloss.Center, lipgloss.Center,
        overlay,
        lipgloss.WithWhitespaceChars(" "),
        lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
    )
}
```

### Success Criteria:

#### Automated Verification:
- [ ] `go build ./...` compiles without errors
- [ ] `go test ./...` passes all tests
- [ ] `go test -race ./...` passes (no race conditions)
- [ ] Unit test: 'J' key calls viewport.LineDown
- [ ] Unit test: 'K' key calls viewport.LineUp
- [ ] Unit test: '?' key toggles showHelp flag
- [ ] Unit test: View with showHelp=true contains "Keybindings"
- [ ] Unit test: ErrorMsg sets statusMsg field
- [ ] Unit test: ErrorMsg returns tea.Tick command for auto-clear
- [ ] Unit test: ClearStatusMsg clears statusMsg field
- [ ] Unit test: View with statusMsg shows error text
- [ ] Integration test: Full workflow - start, receive review, navigate, scroll, quit

#### Manual Verification:
- [ ] J/K scroll the diff content smoothly
- [ ] Scrolling respects content bounds (no over-scroll)
- [ ] '?' shows help overlay centered on screen
- [ ] '?' again hides help overlay
- [ ] 'q' works when help overlay is shown
- [ ] Invalid JSON POST shows error briefly in status bar
- [ ] Status bar error clears after ~3 seconds

---

## Testing Strategy

### Unit Tests

Located in `*_test.go` files alongside source:

1. **Helper tests** (`internal/tui/helpers_test.go`):
   - truncate function behavior

2. **Model tests** (`internal/tui/model_test.go`):
   - Initial state values
   - State after receiving review

3. **Update tests** (`internal/tui/update_test.go`):
   - Key handling (q, j, k, J, K, ?, arrows)
   - Message handling (ReviewReceivedMsg, ErrorMsg, WindowSizeMsg, PortBoundMsg, ClearStatusMsg)
   - Navigation bounds checking
   - Viewport reset on navigation
   - **Test pattern for quit**: Execute returned cmd and assert `tea.QuitMsg`

4. **View tests** (`internal/tui/view_test.go`):
   - Empty state contains expected text and port
   - Review state shows title, sections, hunks
   - Selected section has "› " prefix
   - Help overlay visible when toggled
   - Status bar shows error message

5. **Highlight tests** (`internal/highlight/syntax_test.go`, `diff_test.go`):
   - Lexer selection by extension
   - ANSI escape sequence presence (use `strings.Contains("\x1b[")`)
   - Diff line colorization by prefix
   - Full diff colorization

6. **API tests** (`internal/api/server_test.go`):
   - HTTP method handling
   - JSON parsing
   - Response codes
   - Mock Dispatcher receives correct messages
   - Graceful shutdown

### Testing the Dispatcher Interface

```go
// MockDispatcher for testing
type MockDispatcher struct {
    messages []tea.Msg
    mu       sync.Mutex
}

func (m *MockDispatcher) Send(msg tea.Msg) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.messages = append(m.messages, msg)
}

func (m *MockDispatcher) Messages() []tea.Msg {
    m.mu.Lock()
    defer m.mu.Unlock()
    return m.messages
}

// Example test
func TestHandleReview_ValidJSON(t *testing.T) {
    mock := &MockDispatcher{}
    server, _ := api.NewServer(mock, "0", false)

    req := httptest.NewRequest("POST", "/review",
        strings.NewReader(`{"title":"Test"}`))
    w := httptest.NewRecorder()

    server.handleReview(w, req)

    if w.Code != http.StatusOK {
        t.Errorf("expected 200, got %d", w.Code)
    }

    msgs := mock.Messages()
    if len(msgs) != 1 {
        t.Fatalf("expected 1 message, got %d", len(msgs))
    }

    if _, ok := msgs[0].(tui.ReviewReceivedMsg); !ok {
        t.Error("expected ReviewReceivedMsg")
    }
}
```

### Testing Quit Command Pattern

```go
func TestUpdate_QuitKey(t *testing.T) {
    m := tui.NewModel("8765")
    msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}

    _, cmd := m.Update(msg)

    if cmd == nil {
        t.Fatal("expected command, got nil")
    }

    // Execute the command and check result
    result := cmd()
    if _, ok := result.(tea.QuitMsg); !ok {
        t.Errorf("expected tea.QuitMsg, got %T", result)
    }
}
```

### Test Patterns

```go
// Table-driven test example with viewport reset verification
func TestNavigation(t *testing.T) {
    tests := []struct {
        name            string
        key             string
        initial         int
        numSections     int
        expectedSel     int
        expectGotoTop   bool
    }{
        {"j from first", "j", 0, 3, 1, true},
        {"j from last", "j", 2, 3, 2, false},  // no change, no reset
        {"k from middle", "k", 1, 3, 0, true},
        {"k from first", "k", 0, 3, 0, false}, // no change, no reset
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            m := modelWithSections(tt.numSections)
            m.selected = tt.initial
            // Set viewport offset to verify reset
            m.viewport.SetYOffset(10)

            msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
            updated, _ := m.Update(msg)
            result := updated.(tui.Model)

            if result.selected != tt.expectedSel {
                t.Errorf("selected: got %d, want %d", result.selected, tt.expectedSel)
            }

            if tt.expectGotoTop && result.viewport.YOffset() != 0 {
                t.Error("expected viewport to reset to top")
            }
        })
    }
}
```

---

## Performance Considerations

- **Chroma highlighting**: Applied per-hunk, should be fast for typical diff sizes
- **Viewport scrolling**: Uses Bubbles viewport which handles large content efficiently
- **HTTP response time**: JSON parsing is the only operation; should meet NFR3 (<100ms)
- **Startup time**: Minimal initialization; should meet NFR1 (<500ms)

If performance issues arise with large reviews (100 sections, 500 hunks per NFR2):
- Cache highlighted content instead of re-highlighting on each render
- Lazy-load hunks as user scrolls

---

## Project Structure

```
diffguide/
├── cmd/
│   └── diffguide/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── dispatcher.go
│   │   ├── server.go
│   │   └── server_test.go
│   ├── highlight/
│   │   ├── diff.go
│   │   ├── diff_test.go
│   │   ├── syntax.go
│   │   └── syntax_test.go
│   ├── model/
│   │   └── review.go
│   └── tui/
│       ├── helpers.go
│       ├── helpers_test.go
│       ├── messages.go
│       ├── model.go
│       ├── model_test.go
│       ├── styles.go
│       ├── update.go
│       ├── update_test.go
│       ├── view.go
│       └── view_test.go
├── go.mod
├── go.sum
├── prd.md
└── working-notes/
```

---

## References

- PRD: `prd.md`
- Research: `working-notes/2025-12-18_research_diffguide-implementation-technologies.md`
- Bubble Tea: https://github.com/charmbracelet/bubbletea
- Bubbles: https://github.com/charmbracelet/bubbles
- Lipgloss: https://github.com/charmbracelet/lipgloss
- Chroma: https://github.com/alecthomas/chroma
