---
date: 2025-12-18T14:09:22-05:00
git_commit: f4c986788f47d53effc089bad8400994714d5883
branch: main
repository: diffguide
topic: "Diffguide MVP Implementation"
tags: [plans, tui, bubble-tea, go, mvp, server-viewer]
status: in_progress
last_updated: 2025-12-18
last_updated_note: "Added robust payload testing requirements with testdata fixtures"
---

# Diffguide MVP Implementation Plan

## Overview

Implement the minimum viable product for diffguide using a **server + viewer architecture**. This enables running multiple viewer instances across different project directories simultaneously (matching the lazygit multi-tab workflow), while maintaining a single HTTP endpoint for MCP integration.

**Architecture**:
- **Server** (`diffguide server`): Single HTTP endpoint, writes reviews to per-directory files
- **Viewer** (`diffguide`): Watches for updates to its working directory's review file

The MVP covers FR1-FR12 from the PRD, enabling developers to view AI-generated narrative explanations alongside syntax-highlighted diffs.

## Current State Analysis

**What exists now (after Phase 2):**
- PRD document defining requirements
- Research document with technology decisions
- Nix flake for reproducible dev environment
- Go module with Bubble Tea, Lipgloss, Chroma dependencies
- Domain types: Review (with WorkingDirectory), Section, Hunk
- TUI model with Init/Update/View implementing tea.Model
- Empty state view with ASCII art and instructions (shows working directory)
- Quit handling (q, ctrl+c)
- Entry point with subcommand routing (cmd/diffguide/main.go)
- Server mode (cmd/diffguide/server.go)
- Storage package: NormalizePath, HashDirectory, Store with atomic Write/Read
- Server package: HTTP server with POST /review endpoint, validation, body limits
- Tests for helpers, update, view, storage, and server

**Key decisions from research:**
- Bubble Tea for TUI framework
- Lipgloss for styling
- Chroma for syntax highlighting
- `tea.Send()` for HTTP → TUI message injection
- Bubbles viewport and list components
- Raw unified diff format within hunks

## Desired End State

A single Go binary with two modes:

### Server Mode (`diffguide server`)
1. Runs HTTP server on port 8765 (configurable)
2. Accepts POST /review with JSON payload including `workingDirectory`
3. Writes review data to `~/.diffguide/reviews/{dir-hash}.json`
4. Responds with 200 OK on success
5. Logs activity when `-v` flag is set

### Viewer Mode (`diffguide` or `diffguide view`)
1. Launches and displays empty state with instructions
2. Watches `~/.diffguide/reviews/{hash-of-pwd}.json` for changes
3. When review file appears/changes, displays two-pane interface
4. Supports j/k/arrow navigation between sections
5. Shows syntax-highlighted code with green/red diff coloring
6. Supports J/K scrolling in the diff pane
7. Quits with 'q'

**Verification:**
```bash
# Terminal 1: Start the server (once)
./diffguide server

# Terminal 2: Start a viewer in project A
cd ~/code/project-a
./diffguide

# Terminal 3: Start a viewer in project B
cd ~/code/project-b
./diffguide

# Terminal 4: Send test data for project A
curl -X POST http://localhost:8765/review \
  -H "Content-Type: application/json" \
  -d '{"workingDirectory": "/Users/you/code/project-a", "title": "Test Review", "sections": [...]}'

# Verify: Only the viewer in Terminal 2 updates
```

## What We're NOT Doing

- FR13-FR21 (post-MVP features)
- MCP server wrapper (post-MVP, but architecture supports it)
- Section filtering by importance
- External editor integration
- Git operations
- Server auto-start from viewer (user starts server manually for MVP)
- Review history/persistence beyond latest per directory
- Configuration files (just CLI flags for MVP)

## Implementation Approach

We'll build incrementally using TDD, with each phase producing working, testable software. The testing strategy uses:

1. **Unit tests** for Update function state transitions and View output
2. **Integration tests** for HTTP server and file watcher
3. **Table-driven tests** for comprehensive input coverage
4. **Race detector** (`go test -race ./...`) for concurrency safety

Key architectural decisions:
- **Server + Viewer separation**: Server handles HTTP, viewer handles display
- **File-based communication**: Server writes JSON files, viewers watch with fsnotify
- **Directory hashing**: `~/.diffguide/reviews/{sha256(abs-path)}.json` for per-directory isolation
- **Viewport from Phase 3**: Prevents layout breakage with long content

The phases are ordered to:
1. ✅ Phase 1: Foundation (TUI skeleton) - COMPLETE
2. ✅ Phase 2: Server mode (HTTP → file writing) - COMPLETE
3. ✅ Phase 3: Viewer mode (file watching → TUI display) - COMPLETE
4. Phase 4: Two-pane layout with navigation
5. Phase 5: Syntax highlighting and diff colors
6. Phase 6: Scrolling and polish

---

## Phase 1: Project Foundation ✓ COMPLETE

### Overview

Set up the Go project structure, create a minimal Bubble Tea TUI that can quit, and display the empty state with ASCII art and instructions.

### Changes Required:

#### 1. Go Module Initialization

**File**: `go.mod`

```go
module github.com/mchowning/diffguide

go 1.24

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

// Truncate shortens a string to maxLen, adding "…" if truncated
func Truncate(s string, maxLen int) string {
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
    workDir  string  // Working directory this viewer is watching
}

func NewModel(workDir string) Model {
    return Model{workDir: workDir}
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
    // Include: "Watching: {m.workDir}"
    // Include: "Start server: diffguide server"
    // Include: "POST http://localhost:8765/review"
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
- [x] Unit test: View() when review is nil contains working directory path
- [x] Unit test: View() when review is nil contains "q: quit"
- [x] Unit test: Truncate("hello", 3) returns "he…"
- [x] Unit test: Truncate("hi", 10) returns "hi"

#### Manual Verification:
- [x] Running `./diffguide` displays ASCII art and instructions
- [x] Pressing 'q' exits the application cleanly

---

## Phase 2: Server Mode ✓ COMPLETE

### Overview

Implement the server mode that accepts HTTP requests and writes review data to per-directory files. The server is a headless process (no TUI) that runs continuously.

### Changes Required:

#### 1. Update Domain Types with WorkingDirectory

**File**: `internal/model/review.go`

```go
package model

type Review struct {
    WorkingDirectory string    `json:"workingDirectory"`
    Title            string    `json:"title"`
    Sections         []Section `json:"sections"`
}

// (Section and Hunk unchanged)
```

#### 2. Review Storage

**File**: `internal/storage/store.go`

```go
package storage

import (
    "crypto/sha256"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"

    "github.com/mchowning/diffguide/internal/model"
)

// Store handles persisting reviews to disk
type Store struct {
    baseDir string
}

// NewStore creates a store with the default base directory (~/.diffguide/reviews)
func NewStore() (*Store, error) {
    home, err := os.UserHomeDir()
    if err != nil {
        return nil, err
    }
    baseDir := filepath.Join(home, ".diffguide", "reviews")
    if err := os.MkdirAll(baseDir, 0755); err != nil {
        return nil, err
    }
    return &Store{baseDir: baseDir}, nil
}

// NewStoreWithDir creates a store with a custom base directory (for testing)
func NewStoreWithDir(baseDir string) (*Store, error) {
    if err := os.MkdirAll(baseDir, 0755); err != nil {
        return nil, err
    }
    return &Store{baseDir: baseDir}, nil
}

// NormalizePath returns a canonical absolute path for consistent hashing.
// Applies filepath.Abs, filepath.Clean, and filepath.EvalSymlinks to handle
// trailing slashes, relative paths, symlinks, and case variations on
// case-insensitive filesystems (macOS, Windows).
func NormalizePath(dir string) (string, error) {
    abs, err := filepath.Abs(dir)
    if err != nil {
        return "", fmt.Errorf("failed to get absolute path: %w", err)
    }
    cleaned := filepath.Clean(abs)

    // EvalSymlinks resolves symlinks AND canonicalizes case on macOS/Windows.
    // This ensures "/users/foo" and "/Users/Foo" produce the same hash.
    resolved, err := filepath.EvalSymlinks(cleaned)
    if err != nil {
        // If path doesn't exist yet, fall back to cleaned path
        // (server may receive paths before directories are created)
        if os.IsNotExist(err) {
            return cleaned, nil
        }
        return "", fmt.Errorf("failed to resolve path: %w", err)
    }
    return resolved, nil
}

// HashDirectory returns the SHA256 hash of a normalized directory path
func HashDirectory(dir string) string {
    // Note: caller should normalize path first for consistency
    hash := sha256.Sum256([]byte(dir))
    return hex.EncodeToString(hash[:])
}

// PathForDirectory returns the file path for a given working directory.
// The directory path is normalized before hashing.
func (s *Store) PathForDirectory(dir string) (string, error) {
    normalized, err := NormalizePath(dir)
    if err != nil {
        return "", err
    }
    return filepath.Join(s.baseDir, HashDirectory(normalized)+".json"), nil
}

// BaseDir returns the base directory for review files (for watcher setup)
func (s *Store) BaseDir() string {
    return s.baseDir
}

// Write persists a review to disk using atomic write (temp file + rename)
// to prevent partial reads by file watchers.
func (s *Store) Write(review model.Review) error {
    path, err := s.PathForDirectory(review.WorkingDirectory)
    if err != nil {
        return err
    }

    data, err := json.MarshalIndent(review, "", "  ")
    if err != nil {
        return err
    }

    // Atomic write: write to temp file, then rename
    tempPath := path + ".tmp"
    if err := os.WriteFile(tempPath, data, 0644); err != nil {
        return fmt.Errorf("failed to write temp file: %w", err)
    }

    if err := os.Rename(tempPath, path); err != nil {
        os.Remove(tempPath) // Clean up temp file on rename failure
        return fmt.Errorf("failed to rename temp file: %w", err)
    }

    return nil
}

// Read loads a review from disk for a given directory
func (s *Store) Read(dir string) (*model.Review, error) {
    path, err := s.PathForDirectory(dir)
    if err != nil {
        return nil, err
    }
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var review model.Review
    if err := json.Unmarshal(data, &review); err != nil {
        return nil, err
    }
    return &review, nil
}
```

#### 3. HTTP Server (Headless)

**File**: `internal/server/server.go`

```go
package server

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "log"
    "net"
    "net/http"
    "time"

    "github.com/mchowning/diffguide/internal/model"
    "github.com/mchowning/diffguide/internal/storage"
)

type Server struct {
    store    *storage.Store
    server   *http.Server
    listener net.Listener
    verbose  bool
}

func New(store *storage.Store, port string, verbose bool) (*Server, error) {
    mux := http.NewServeMux()

    s := &Server{
        store:   store,
        verbose: verbose,
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

    ln, err := net.Listen("tcp", s.server.Addr)
    if err != nil {
        return nil, fmt.Errorf("failed to bind: %w", err)
    }
    s.listener = ln

    return s, nil
}

func (s *Server) Port() string {
    addr := s.listener.Addr().(*net.TCPAddr)
    return fmt.Sprintf("%d", addr.Port)
}

func (s *Server) Run() error {
    if s.verbose {
        log.Printf("Server listening on http://127.0.0.1:%s", s.Port())
    }
    return s.server.Serve(s.listener)
}

func (s *Server) Shutdown(ctx context.Context) error {
    return s.server.Shutdown(ctx)
}

// maxRequestBody limits request body size to 10MB to prevent DoS
const maxRequestBody = 10 * 1024 * 1024

func (s *Server) handleReview(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Limit request body size to prevent memory exhaustion
    r.Body = http.MaxBytesReader(w, r.Body, maxRequestBody)
    defer r.Body.Close()

    var review model.Review
    if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
        // Check for oversized body using proper type assertion
        var maxBytesErr *http.MaxBytesError
        if errors.As(err, &maxBytesErr) {
            http.Error(w, "Request body too large (max 10MB)", http.StatusRequestEntityTooLarge)
            return
        }
        http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
        return
    }

    if review.WorkingDirectory == "" {
        http.Error(w, "Missing workingDirectory field", http.StatusBadRequest)
        return
    }

    // Normalize the working directory path for consistent hashing
    normalized, err := storage.NormalizePath(review.WorkingDirectory)
    if err != nil {
        http.Error(w, "Invalid workingDirectory: "+err.Error(), http.StatusBadRequest)
        return
    }
    review.WorkingDirectory = normalized

    if err := s.store.Write(review); err != nil {
        http.Error(w, "Failed to store review: "+err.Error(), http.StatusInternalServerError)
        return
    }

    if s.verbose {
        log.Printf("Stored review for %s: %s (%d sections)",
            review.WorkingDirectory, review.Title, len(review.Sections))
    }

    w.WriteHeader(http.StatusOK)
}
```

#### 4. Server Command Entry Point

**File**: `cmd/diffguide/server.go`

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/mchowning/diffguide/internal/server"
    "github.com/mchowning/diffguide/internal/storage"
)

func runServer(port string, verbose bool) {
    store, err := storage.NewStore()
    if err != nil {
        log.Fatalf("Failed to create store: %v", err)
    }

    srv, err := server.New(store, port, verbose)
    if err != nil {
        log.Fatalf("Failed to create server: %v", err)
    }

    // Handle shutdown signals
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-stop
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        srv.Shutdown(ctx)
    }()

    if err := srv.Run(); err != nil && err != http.ErrServerClosed {
        log.Fatalf("Server error: %v", err)
    }
}
```

#### 5. Updated Main Entry Point with Subcommands

**File**: `cmd/diffguide/main.go`

```go
package main

import (
    "flag"
    "fmt"
    "log"
    "os"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/mchowning/diffguide/internal/tui"
)

func main() {
    if len(os.Args) > 1 && os.Args[1] == "server" {
        // Server mode
        serverCmd := flag.NewFlagSet("server", flag.ExitOnError)
        port := serverCmd.String("port", "8765", "HTTP server port")
        verbose := serverCmd.Bool("v", false, "Enable verbose logging")
        serverCmd.Parse(os.Args[2:])
        runServer(*port, *verbose)
        return
    }

    // Viewer mode (default)
    flag.Parse()
    runViewer()
}

func runViewer() {
    // Get current working directory
    cwd, err := os.Getwd()
    if err != nil {
        log.Fatalf("Failed to get working directory: %v", err)
    }

    m := tui.NewModel(cwd)
    p := tea.NewProgram(m, tea.WithAltScreen())

    if _, err := p.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

### Success Criteria:

#### Automated Verification:
- [x] `go build ./...` compiles without errors
- [x] `go test ./...` passes all tests
- [x] `go test -race ./...` passes (no race conditions)
- [x] Unit test: NormalizePath handles trailing slashes consistently
- [x] Unit test: NormalizePath handles relative paths
- [x] Unit test: NormalizePath resolves symlinks
- [x] Unit test: NormalizePath canonicalizes case on case-insensitive filesystems
- [x] Unit test: HashDirectory returns consistent hash for same input
- [x] Unit test: HashDirectory returns different hash for different inputs
- [x] Unit test: Store.Write creates file at expected path
- [x] Unit test: Store.Write uses atomic write (temp file + rename)
- [x] Unit test: Store.Read returns written review
- [x] Unit test: POST /review with valid JSON returns 200 OK
- [x] Unit test: POST /review without workingDirectory returns 400
- [x] Unit test: POST /review with invalid JSON returns 400
- [x] Unit test: POST /review normalizes workingDirectory path
- [x] Unit test: POST /review with oversized body returns 413
- [x] Unit test: GET /review returns 405 Method Not Allowed
- [x] Unit test: Review file is created after POST
- [x] Unit test: Server gracefully shuts down on SIGTERM

**Payload Integration Tests** (using testdata fixtures):
- [x] Integration test: POST `simple_review.json` - all fields round-trip correctly
- [x] Integration test: POST `multi_section_review.json` - all sections/hunks preserved
- [x] Integration test: POST `realistic_claude_review.json` - full Claude-style payload works
- [x] Integration test: POST `unicode_content.json` - unicode in narratives/diffs preserved
- [x] Integration test: POST `special_characters.json` - quotes, backslashes, newlines preserved
- [x] Integration test: POST `empty_arrays.json` - empty sections array handled correctly
- [x] Integration test: POST `large_review.json` - 100 sections/500 hunks within NFR limits
- [x] Integration test: Diff content with hunk headers (`@@`) preserved exactly
- [x] Integration test: Diff content with context lines (space prefix) preserved exactly
- [x] Integration test: Multi-line narrative text preserved with exact whitespace

#### Manual Verification:
- [x] `./diffguide server` starts and listens on port 8765
- [x] `./diffguide server -v` logs incoming requests
- [x] `curl -X POST localhost:8765/review -d '{"workingDirectory":"/tmp/test","title":"Test"}'` returns 200
- [x] File exists at `~/.diffguide/reviews/{hash}.json` after POST
- [x] No .tmp files left in reviews directory after POST
- [x] Ctrl+C gracefully shuts down server

---

## Phase 3: Viewer Mode with File Watching ✓ COMPLETE

### Overview

Implement the viewer mode that watches for review file changes using fsnotify. When the server writes a review file for this viewer's working directory, the TUI updates to display it.

### Changes Required:

#### 1. Add fsnotify Dependency

**File**: `go.mod`

```
require github.com/fsnotify/fsnotify v1.7.0
```

#### 2. TUI Messages for File Watching

**File**: `internal/tui/messages.go`

```go
package tui

import "github.com/mchowning/diffguide/internal/model"

// ReviewReceivedMsg is sent when a review file is created/updated
type ReviewReceivedMsg struct {
    Review model.Review
}

// ReviewClearedMsg is sent when the review file is deleted
type ReviewClearedMsg struct{}

// WatchErrorMsg is sent when file watching fails
type WatchErrorMsg struct {
    Err error
}
```

#### 3. File Watcher

**File**: `internal/watcher/watcher.go`

```go
package watcher

import (
    "encoding/json"
    "os"
    "path/filepath"

    "github.com/fsnotify/fsnotify"
    "github.com/mchowning/diffguide/internal/model"
    "github.com/mchowning/diffguide/internal/storage"
)

// Watcher watches for review file changes for a specific directory
type Watcher struct {
    workDir    string
    reviewPath string
    reviewDir  string
    fsWatcher  *fsnotify.Watcher
    Reviews    chan model.Review  // Buffered to prevent deadlock on initial send
    Cleared    chan struct{}      // Sent when review file is deleted
    Errors     chan error         // Buffered to prevent deadlock
    done       chan struct{}
}

// New creates a watcher for the given working directory.
// Uses the default storage location (~/.diffguide/reviews).
func New(workDir string) (*Watcher, error) {
    store, err := storage.NewStore()
    if err != nil {
        return nil, err
    }
    return NewWithStore(workDir, store)
}

// NewWithStore creates a watcher using a custom store (for testing).
// This enables tests to use t.TempDir() for isolation.
func NewWithStore(workDir string, store *storage.Store) (*Watcher, error) {
    // Normalize the working directory for consistent hashing
    normalized, err := storage.NormalizePath(workDir)
    if err != nil {
        return nil, err
    }

    fsWatcher, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, err
    }

    reviewPath, err := store.PathForDirectory(normalized)
    if err != nil {
        fsWatcher.Close()
        return nil, err
    }
    reviewDir := store.BaseDir()

    // Ensure reviews directory exists and watch it
    if err := os.MkdirAll(reviewDir, 0755); err != nil {
        fsWatcher.Close()
        return nil, err
    }
    if err := fsWatcher.Add(reviewDir); err != nil {
        fsWatcher.Close()
        return nil, err
    }

    return &Watcher{
        workDir:    normalized,
        reviewPath: reviewPath,
        reviewDir:  reviewDir,
        fsWatcher:  fsWatcher,
        Reviews:    make(chan model.Review, 1),  // Buffered to avoid deadlock
        Cleared:    make(chan struct{}, 1),      // Buffered to avoid deadlock
        Errors:     make(chan error, 1),         // Buffered to avoid deadlock
        done:       make(chan struct{}),
    }, nil
}

// Start begins watching for file changes.
// Loads any existing review file asynchronously to avoid blocking.
func (w *Watcher) Start() {
    // Load existing review asynchronously to avoid deadlock
    go func() {
        if review, err := w.loadReview(); err == nil {
            select {
            case w.Reviews <- *review:
            case <-w.done:
            }
        }
    }()

    go w.watch()
}

func (w *Watcher) watch() {
    for {
        select {
        case <-w.done:
            return
        case event, ok := <-w.fsWatcher.Events:
            if !ok {
                return
            }
            // Only care about our specific file
            if event.Name != w.reviewPath {
                continue
            }

            // Handle file deletion - transition TUI back to empty state
            if event.Has(fsnotify.Remove) {
                select {
                case w.Cleared <- struct{}{}:
                case <-w.done:
                    return
                }
                continue
            }

            // Handle Write, Create, and Rename events (Rename for atomic writes)
            if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Rename) {
                review, err := w.loadReview()
                if err != nil {
                    // File might have been deleted between event and read
                    if os.IsNotExist(err) {
                        select {
                        case w.Cleared <- struct{}{}:
                        case <-w.done:
                            return
                        }
                        continue
                    }
                    select {
                    case w.Errors <- err:
                    case <-w.done:
                        return
                    }
                    continue
                }
                select {
                case w.Reviews <- *review:
                case <-w.done:
                    return
                }
            }
        case err, ok := <-w.fsWatcher.Errors:
            if !ok {
                return
            }
            select {
            case w.Errors <- err:
            case <-w.done:
                return
            }
        }
    }
}

func (w *Watcher) loadReview() (*model.Review, error) {
    data, err := os.ReadFile(w.reviewPath)
    if err != nil {
        return nil, err
    }
    var review model.Review
    if err := json.Unmarshal(data, &review); err != nil {
        return nil, err
    }
    return &review, nil
}

// ReviewPath returns the path being watched (for testing)
func (w *Watcher) ReviewPath() string {
    return w.reviewPath
}

// Close stops the watcher
func (w *Watcher) Close() error {
    close(w.done)
    return w.fsWatcher.Close()
}
```

#### 4. Update TUI Model

**File**: `internal/tui/model.go`

Update to store working directory instead of port:

```go
type Model struct {
    review   *model.Review
    selected int
    width    int
    height   int
    workDir  string  // Working directory this viewer is watching
}

func NewModel(workDir string) Model {
    return Model{workDir: workDir}
}
```

#### 5. Update View for New Empty State

**File**: `internal/tui/view.go`

Update empty state to show working directory info:

```go
func (m Model) renderEmptyState() string {
    // ASCII art logo and instructions
    // Show: "Watching for reviews in: {workDir}"
    // Show: "Start server: diffguide server"
    // Show: "Send review: POST http://localhost:8765/review"
}
```

#### 6. Update Function - Handle File Watcher Messages

**File**: `internal/tui/update.go`

```go
case ReviewReceivedMsg:
    m.review = &msg.Review
    m.selected = 0
    return m, nil

case ReviewClearedMsg:
    // Review file was deleted - return to empty state
    m.review = nil
    m.selected = 0
    return m, nil

case WatchErrorMsg:
    // Could show in status bar; for MVP just log
    return m, nil
```

#### 7. Bubble Tea Command for File Watching

**File**: `internal/tui/commands.go`

```go
package tui

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/mchowning/diffguide/internal/watcher"
)

// WatchForReviews returns a command that listens for review updates
func WatchForReviews(w *watcher.Watcher) tea.Cmd {
    return func() tea.Msg {
        select {
        case review := <-w.Reviews:
            return ReviewReceivedMsg{Review: review}
        case err := <-w.Errors:
            return WatchErrorMsg{Err: err}
        }
    }
}
```

#### 8. Update Main Entry Point

**File**: `cmd/diffguide/main.go`

```go
func runViewer() {
    cwd, err := os.Getwd()
    if err != nil {
        log.Fatalf("Failed to get working directory: %v", err)
    }

    w, err := watcher.New(cwd)
    if err != nil {
        log.Fatalf("Failed to create watcher: %v", err)
    }
    defer w.Close()

    w.Start()

    m := tui.NewModel(cwd)
    p := tea.NewProgram(m, tea.WithAltScreen())

    // Continuously pump watcher events to TUI
    go func() {
        for {
            select {
            case review := <-w.Reviews:
                p.Send(tui.ReviewReceivedMsg{Review: review})
            case <-w.Cleared:
                p.Send(tui.ReviewClearedMsg{})
            case err := <-w.Errors:
                p.Send(tui.WatchErrorMsg{Err: err})
            }
        }
    }()

    if _, err := p.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

### Success Criteria:

#### Automated Verification:
- [x] `go build ./...` compiles without errors
- [x] `go test ./...` passes all tests
- [x] `go test -race ./...` passes (no race conditions)
- [x] Unit test: Watcher.New creates watcher for given directory
- [x] Unit test: Watcher.NewWithStore accepts custom store for test isolation
- [x] Unit test: Watcher normalizes working directory path
- [x] Unit test: Watcher watches correct file path based on directory hash
- [x] Unit test: Watcher sends ReviewReceivedMsg when file is created
- [x] Unit test: Watcher sends ReviewReceivedMsg when file is modified
- [x] Unit test: Watcher sends ReviewReceivedMsg on file rename (atomic writes)
- [x] Unit test: Watcher sends on Cleared channel when file is deleted
- [x] Unit test: Watcher ignores changes to other files in reviews directory
- [x] Unit test: Watcher channels are buffered (no deadlock on initial send)
- [x] Unit test: Update with ReviewReceivedMsg sets m.review
- [x] Unit test: Update with ReviewReceivedMsg sets m.selected to 0
- [x] Unit test: Update with ReviewClearedMsg sets m.review to nil
- [x] Unit test: View in empty state shows working directory

#### Manual Verification:
- [x] Start viewer: `./diffguide` shows empty state with working directory
- [x] Start server in another terminal: `./diffguide server`
- [x] Send review via curl - viewer updates immediately
- [x] Send another review - viewer updates with new content
- [x] Viewer in different directory doesn't see updates for other directories

---

## Phase 4: Two-Pane Layout with Navigation and Viewport

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
    workDir   string
    viewport  viewport.Model
    ready     bool  // viewport initialized after first WindowSizeMsg
}

func NewModel(workDir string) Model {
    return Model{workDir: workDir}
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
        text := prefix + Truncate(section.Narrative, width-len(prefix)-4)
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

## Phase 5: Syntax Highlighting and Diff Colors

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

## Phase 6: Scrolling and Polish

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
    workDir   string
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

### Realistic Payload Testing

To avoid regressions and ensure the system handles real-world Claude-generated reviews correctly, we maintain a set of testdata fixtures and comprehensive integration tests.

#### Testdata Fixtures

**Directory**: `internal/testdata/`

| File | Purpose |
|------|---------|
| `simple_review.json` | Minimal valid review: 1 section, 1 hunk, basic diff |
| `multi_section_review.json` | 5 sections with varying numbers of hunks (1-4 each) |
| `realistic_claude_review.json` | Full Claude-style payload with narrative explanations, multiple file types (.go, .ts, .md), real unified diff format |
| `unicode_content.json` | Unicode in narratives (emoji, CJK, accented chars) and file paths |
| `special_characters.json` | Quotes, backslashes, tabs, embedded newlines in diff content |
| `empty_arrays.json` | Valid review with `sections: []` |
| `large_review.json` | Near NFR2 limits: 100 sections, ~500 hunks total |

#### Realistic Unified Diff Content

All diff content in fixtures should use proper unified diff format:

```diff
@@ -10,7 +10,9 @@ func processFile(path string) error {
     if err != nil {
         return fmt.Errorf("failed to open: %w", err)
     }
+    defer f.Close()
+
     data, err := io.ReadAll(f)
-    f.Close()
     return nil
 }
```

Key elements to include:
- Hunk headers with line numbers (`@@ -start,count +start,count @@`)
- Optional function context after hunk header
- Context lines (space prefix)
- Addition lines (`+` prefix)
- Deletion lines (`-` prefix)
- Multiple hunks per file
- Various file extensions for syntax highlighting coverage

#### Integration Test Requirements

**File**: `internal/integration/payload_test.go`

These tests verify the full round-trip from HTTP POST to stored file:

1. **Full Field Preservation Test**
   - POST each testdata fixture via HTTP
   - Read the stored JSON file
   - Verify ALL fields match exactly:
     - `workingDirectory` (normalized)
     - `title`
     - Each section: `id`, `narrative`, `importance`
     - Each hunk: `file`, `startLine`, `diff`

2. **Diff Content Integrity Test**
   - POST payload with multi-line unified diff content
   - Verify stored diff preserves:
     - Exact line breaks (`\n`)
     - Leading whitespace (context line indentation)
     - Special characters (no escaping corruption)
     - Unicode characters

3. **Multi-Section Navigation Test** (Phase 4+)
   - Load `multi_section_review.json` into TUI
   - Verify all sections appear in list
   - Navigate to each section, verify correct hunks display

4. **Boundary Test**
   - POST `large_review.json` (100 sections, ~500 hunks)
   - Verify 200 OK response within NFR3 time limit (<100ms)
   - Verify file is stored correctly
   - Verify TUI can load and navigate without lag

#### Fixture Creation Guidelines

When creating test fixtures:

1. **Use realistic narratives** - Not "test narrative" but actual explanatory text like Claude would generate:
   ```json
   {
     "narrative": "This change adds proper resource cleanup by deferring the file close immediately after opening. Previously, the file handle could leak if an error occurred between Open and the manual Close call."
   }
   ```

2. **Use real file paths** - Not `/test/file.txt` but realistic paths:
   ```json
   {
     "file": "internal/server/handler.go",
     "startLine": 45
   }
   ```

3. **Include various importance levels** - Mix of "high", "medium", "low"

4. **Cover multiple languages** - `.go`, `.ts`, `.py`, `.rs`, `.md`, etc.

#### Example: realistic_claude_review.json

```json
{
  "workingDirectory": "/Users/dev/projects/myapp",
  "title": "Add user authentication middleware",
  "sections": [
    {
      "id": "sec-1",
      "narrative": "This change introduces JWT-based authentication middleware. The middleware extracts the Bearer token from the Authorization header, validates it against the signing key, and attaches the decoded user claims to the request context for downstream handlers.",
      "importance": "high",
      "hunks": [
        {
          "file": "internal/middleware/auth.go",
          "startLine": 1,
          "diff": "@@ -0,0 +1,45 @@\n+package middleware\n+\n+import (\n+\t\"context\"\n+\t\"net/http\"\n+\t\"strings\"\n+\n+\t\"github.com/golang-jwt/jwt/v5\"\n+)\n+\n+type contextKey string\n+\n+const UserClaimsKey contextKey = \"userClaims\"\n+\n+func AuthMiddleware(signingKey []byte) func(http.Handler) http.Handler {\n+\treturn func(next http.Handler) http.Handler {\n+\t\treturn http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {\n+\t\t\tauthHeader := r.Header.Get(\"Authorization\")\n+\t\t\tif !strings.HasPrefix(authHeader, \"Bearer \") {\n+\t\t\t\thttp.Error(w, \"Missing or invalid Authorization header\", http.StatusUnauthorized)\n+\t\t\t\treturn\n+\t\t\t}\n+\n+\t\t\ttokenStr := strings.TrimPrefix(authHeader, \"Bearer \")\n+\t\t\ttoken, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {\n+\t\t\t\treturn signingKey, nil\n+\t\t\t})\n+\t\t\tif err != nil || !token.Valid {\n+\t\t\t\thttp.Error(w, \"Invalid token\", http.StatusUnauthorized)\n+\t\t\t\treturn\n+\t\t\t}\n+\n+\t\t\tctx := context.WithValue(r.Context(), UserClaimsKey, token.Claims)\n+\t\t\tnext.ServeHTTP(w, r.WithContext(ctx))\n+\t\t})\n+\t}\n+}"
        }
      ]
    },
    {
      "id": "sec-2",
      "narrative": "Updated the router to apply the auth middleware to protected routes. Public endpoints like /health and /login remain accessible without authentication.",
      "importance": "medium",
      "hunks": [
        {
          "file": "cmd/server/main.go",
          "startLine": 25,
          "diff": "@@ -25,6 +25,8 @@ func main() {\n \trouter := http.NewServeMux()\n \n \t// Public routes\n \trouter.HandleFunc(\"/health\", handlers.Health)\n \trouter.HandleFunc(\"/login\", handlers.Login)\n+\n+\t// Protected routes (with auth middleware)\n+\tprotected := middleware.AuthMiddleware([]byte(os.Getenv(\"JWT_SECRET\")))\n+\trouter.Handle(\"/api/\", protected(http.StripPrefix(\"/api\", apiRouter)))\n \n \tlog.Fatal(http.ListenAndServe(\":8080\", router))\n }"
        }
      ]
    },
    {
      "id": "sec-3",
      "narrative": "Added comprehensive tests for the auth middleware covering valid tokens, expired tokens, malformed tokens, and missing headers.",
      "importance": "medium",
      "hunks": [
        {
          "file": "internal/middleware/auth_test.go",
          "startLine": 1,
          "diff": "@@ -0,0 +1,89 @@\n+package middleware_test\n+\n+import (\n+\t\"net/http\"\n+\t\"net/http/httptest\"\n+\t\"testing\"\n+\t\"time\"\n+\n+\t\"github.com/golang-jwt/jwt/v5\"\n+\t\"github.com/myapp/internal/middleware\"\n+)\n+\n+var testSigningKey = []byte(\"test-secret-key\")\n+\n+func createTestToken(t *testing.T, exp time.Time) string {\n+\tt.Helper()\n+\ttoken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{\n+\t\t\"sub\": \"user123\",\n+\t\t\"exp\": exp.Unix(),\n+\t})\n+\ttokenStr, err := token.SignedString(testSigningKey)\n+\tif err != nil {\n+\t\tt.Fatalf(\"failed to sign token: %v\", err)\n+\t}\n+\treturn tokenStr\n+}\n+\n+func TestAuthMiddleware_ValidToken(t *testing.T) {\n+\ttoken := createTestToken(t, time.Now().Add(time.Hour))\n+\t// ... test implementation\n+}"
        }
      ]
    }
  ]
}
```

### Unit Tests

Located in `*_test.go` files alongside source:

1. **Helper tests** (`internal/tui/helpers_test.go`):
   - Truncate function behavior

2. **Model tests** (`internal/tui/model_test.go`):
   - Initial state values
   - State after receiving review

3. **Update tests** (`internal/tui/update_test.go`):
   - Key handling (q, j, k, J, K, ?, arrows)
   - Message handling (ReviewReceivedMsg, WatchErrorMsg, WindowSizeMsg, ClearStatusMsg)
   - Navigation bounds checking
   - Viewport reset on navigation
   - **Test pattern for quit**: Execute returned cmd and assert `tea.QuitMsg`

4. **View tests** (`internal/tui/view_test.go`):
   - Empty state contains expected text and working directory
   - Review state shows title, sections, hunks
   - Selected section has "› " prefix
   - Help overlay visible when toggled
   - Status bar shows error message

5. **Highlight tests** (`internal/highlight/syntax_test.go`, `diff_test.go`):
   - Lexer selection by extension
   - ANSI escape sequence presence (use `strings.Contains("\x1b[")`)
   - Diff line colorization by prefix
   - Full diff colorization

6. **Storage tests** (`internal/storage/store_test.go`):
   - HashDirectory consistency
   - PathForDirectory correctness
   - Write creates file at expected path
   - Read returns written review
   - Round-trip (Write then Read) - **must verify ALL fields**:
     - `workingDirectory`, `title`
     - Each section: `id`, `narrative`, `importance`
     - Each hunk: `file`, `startLine`, `diff` (exact content)

7. **Server tests** (`internal/server/server_test.go`):
   - HTTP method handling
   - JSON parsing and validation
   - workingDirectory required
   - Response codes
   - File created after POST

8. **Watcher tests** (`internal/watcher/watcher_test.go`):
   - Watches correct file based on directory hash
   - Sends review when file created
   - Sends review when file modified
   - Ignores other files in reviews directory

### Testing Storage

```go
func TestStore_RoundTrip(t *testing.T) {
    dir := t.TempDir()
    store, _ := storage.NewStoreWithDir(dir)

    review := model.Review{
        WorkingDirectory: "/test/project",
        Title:            "Test Review",
        Sections:         []model.Section{{ID: "1", Narrative: "Test"}},
    }

    err := store.Write(review)
    if err != nil {
        t.Fatalf("Write failed: %v", err)
    }

    loaded, err := store.Read("/test/project")
    if err != nil {
        t.Fatalf("Read failed: %v", err)
    }

    if loaded.Title != review.Title {
        t.Errorf("Title mismatch: got %q, want %q", loaded.Title, review.Title)
    }
}
```

### Testing Quit Command Pattern

```go
func TestUpdate_QuitKey(t *testing.T) {
    m := tui.NewModel("/test/project")
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
│       ├── main.go         # Entry point with subcommand routing
│       └── server.go       # Server mode implementation
├── internal/
│   ├── highlight/
│   │   ├── diff.go
│   │   ├── diff_test.go
│   │   ├── syntax.go
│   │   └── syntax_test.go
│   ├── integration/
│   │   └── payload_test.go # Full round-trip integration tests
│   ├── model/
│   │   └── review.go       # Domain types (Review, Section, Hunk)
│   ├── server/
│   │   ├── server.go       # HTTP server for server mode
│   │   └── server_test.go
│   ├── storage/
│   │   ├── store.go        # File-based review storage
│   │   └── store_test.go
│   ├── testdata/           # Realistic test fixtures
│   │   ├── simple_review.json
│   │   ├── multi_section_review.json
│   │   ├── realistic_claude_review.json
│   │   ├── unicode_content.json
│   │   ├── special_characters.json
│   │   ├── empty_arrays.json
│   │   └── large_review.json
│   ├── tui/
│   │   ├── commands.go     # Bubble Tea commands
│   │   ├── helpers.go
│   │   ├── helpers_test.go
│   │   ├── messages.go     # Message types
│   │   ├── model.go
│   │   ├── model_test.go
│   │   ├── styles.go
│   │   ├── update.go
│   │   ├── update_test.go
│   │   ├── view.go
│   │   └── view_test.go
│   └── watcher/
│       ├── watcher.go      # File watcher for viewer mode
│       └── watcher_test.go
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
