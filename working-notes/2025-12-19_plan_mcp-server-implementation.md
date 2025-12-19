---
date: 2025-12-19T12:25:33-05:00
git_commit: ae2633f258f656090e17def71825a57472bae7b5
branch: main
repository: diffguide
topic: "MCP Server Implementation for diffguide"
tags: [plans, mcp, go-sdk, claude-code, refactoring]
status: complete
last_updated: 2025-12-19
---

# MCP Server Implementation Plan

## Overview

Add an MCP (Model Context Protocol) server to diffguide as a new subcommand (`diffguide mcp`) that exposes a `submit_review` tool. This enables Claude Code to submit code reviews directly to diffguide without requiring HTTP server setup.

To maximize code reuse, we'll first extract the shared business logic (validation, normalization, storage) into a review service that both the HTTP and MCP servers will use.

## Current State Analysis

The HTTP server (`internal/server/server.go:92-108`) contains business logic that will be duplicated by the MCP server:
- Validation: `workingDirectory` required check
- Path normalization: `storage.NormalizePath()`
- Storage: `store.Write(review)`

This logic should be extracted into a shared service.

### Key Discoveries:

- HTTP handler at `internal/server/server.go:70-116` mixes HTTP concerns with business logic
- Storage layer at `internal/storage/store.go` is already well-factored
- Go MCP SDK dependency needs to be added to `go.mod`

## Desired End State

After implementation:
1. Shared `internal/review/service.go` handles validation, normalization, and storage
2. HTTP server is a thin adapter that delegates to the review service
3. MCP server is a thin adapter that delegates to the same review service
4. Running `diffguide mcp` starts an MCP server on stdio
5. Reviews submitted via either HTTP or MCP are stored identically

**Verification**: Both servers use identical business logic; submitting via MCP produces the same stored result as submitting via HTTP.

## What We're NOT Doing

- No read/list/delete tools (submit only per requirements)
- No HTTP transport for MCP (stdio only)
- No authentication (local tool, no network exposure)
- No changes to TUI viewer

## Implementation Approach

1. Install and verify MCP SDK dependency
2. Extract shared business logic into `internal/review/service.go`
3. Refactor HTTP server to use the service (behavior unchanged)
4. Build MCP server using the same service
5. Wire up CLI integration

## Important Notes

### Stdout/Stderr Separation

**Critical**: MCP uses stdio for protocol communication. All logging MUST go to stderr, never stdout.

- `log.Println` and `log.Printf` default to stderr (safe)
- Never use `fmt.Println` or `fmt.Printf` in MCP code path
- Document this constraint in code comments

---

## Phase 0: SDK Setup & Verification

### Overview

Install the MCP SDK dependency and verify the actual API matches our planned usage. The code snippets in subsequent phases are based on SDK documentation and may need adjustment.

### Changes Required:

#### 1. Install Dependency

```bash
nix develop -c go get github.com/modelcontextprotocol/go-sdk
nix develop -c go mod tidy
```

#### 2. Create SDK Spike

Create a minimal test to verify SDK API surface:

**File**: `internal/mcpserver/spike_test.go` (temporary, delete after verification)

```go
//go:build spike

package mcpserver

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestSDKSpike verifies the MCP SDK API matches our assumptions.
// Run with: go test -tags=spike ./internal/mcpserver/...
func TestSDKSpike(t *testing.T) {
	// Verify server creation
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "test",
		Version: "1.0.0",
	}, nil)

	if server == nil {
		t.Fatal("NewServer returned nil")
	}

	// Verify tool registration compiles
	// Adjust signatures based on actual SDK
	t.Log("SDK API verified - update plan if signatures differ")
}
```

#### 3. Verify and Update Plan

After running the spike:
- Confirm `mcp.NewServer` signature
- Confirm `mcp.AddTool` signature and handler type
- Confirm transport creation (`mcp.NewStdioTransport` or similar)
- Confirm how tool results are returned (typed struct vs content array)
- Check if `jsonschema` struct tags are supported for schema generation

**Update subsequent phases if API differs from plan.**

### Success Criteria:

#### Automated Verification:
- [x] Dependency added: `nix develop -c go mod tidy` succeeds
- [x] Spike compiles: `nix develop -c go test -tags=spike ./internal/mcpserver/...`

#### Manual Verification:
- [x] Review SDK API and update Phase 3 code snippets if needed (jsonschema tags use `"description"` not `"key=value"` format; StdioTransport is `&mcp.StdioTransport{}` not `NewStdioTransport()`)

---

## Phase 1: Extract Review Service

### Overview

Extract the validation, normalization, and storage logic from the HTTP server into a reusable service.

### Changes Required:

#### 1. Create Review Service Package

**File**: `internal/review/service.go`

```go
package review

import (
	"context"
	"errors"
	"fmt"

	"github.com/mchowning/diffguide/internal/model"
	"github.com/mchowning/diffguide/internal/storage"
)

// Sentinel errors for error handling with errors.Is
var (
	ErrMissingWorkingDirectory = errors.New("workingDirectory is required")
	ErrInvalidWorkingDirectory = errors.New("invalid workingDirectory")
)

// SubmitResult contains the result of a successful review submission
type SubmitResult struct {
	FilePath string
}

// Service handles review submission business logic
type Service struct {
	store *storage.Store
}

// NewService creates a new review service
func NewService(store *storage.Store) *Service {
	return &Service{store: store}
}

// Submit validates, normalizes, and stores a review
func (s *Service) Submit(ctx context.Context, review model.Review) (SubmitResult, error) {
	if review.WorkingDirectory == "" {
		return SubmitResult{}, ErrMissingWorkingDirectory
	}

	normalized, err := storage.NormalizePath(review.WorkingDirectory)
	if err != nil {
		return SubmitResult{}, fmt.Errorf("%w: %v", ErrInvalidWorkingDirectory, err)
	}
	review.WorkingDirectory = normalized

	if err := s.store.Write(review); err != nil {
		return SubmitResult{}, err
	}

	filePath, _ := s.store.PathForDirectory(normalized)
	return SubmitResult{FilePath: filePath}, nil
}
```

#### 2. Create Service Tests

**File**: `internal/review/service_test.go`

Test cases (TDD - write these first):

1. `TestService_SubmitWithValidReview` - Happy path: stores review, returns file path
2. `TestService_SubmitMissingWorkingDirectory` - Returns `ErrMissingWorkingDirectory`
3. `TestService_SubmitInvalidWorkingDirectory` - Returns `ErrInvalidWorkingDirectory`
4. `TestService_SubmitNormalizesPath` - Trailing slashes and relative paths handled
5. `TestService_SubmitPreservesAllFields` - Title, sections, hunks all stored

```go
package review_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mchowning/diffguide/internal/model"
	"github.com/mchowning/diffguide/internal/review"
	"github.com/mchowning/diffguide/internal/storage"
)

func setupTestService(t *testing.T) (*review.Service, *storage.Store) {
	t.Helper()
	baseDir := t.TempDir()
	store, err := storage.NewStoreWithDir(baseDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	svc := review.NewService(store)
	return svc, store
}

func TestService_SubmitWithValidReview(t *testing.T) {
	svc, store := setupTestService(t)
	ctx := context.Background()

	input := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test Review",
	}

	result, err := svc.Submit(ctx, input)
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	if result.FilePath == "" {
		t.Error("expected FilePath to be set")
	}

	// Verify stored
	stored, err := store.Read("/test/project")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if stored.Title != input.Title {
		t.Errorf("Title = %q, want %q", stored.Title, input.Title)
	}
}

func TestService_SubmitMissingWorkingDirectory(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	input := model.Review{
		Title: "Test Review",
		// WorkingDirectory intentionally omitted
	}

	_, err := svc.Submit(ctx, input)
	if !errors.Is(err, review.ErrMissingWorkingDirectory) {
		t.Errorf("expected ErrMissingWorkingDirectory, got %v", err)
	}
}

func TestService_SubmitInvalidWorkingDirectory(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	input := model.Review{
		WorkingDirectory: "\x00invalid", // null byte is invalid in paths
		Title:            "Test Review",
	}

	_, err := svc.Submit(ctx, input)
	if !errors.Is(err, review.ErrInvalidWorkingDirectory) {
		t.Errorf("expected ErrInvalidWorkingDirectory, got %v", err)
	}
}
```

### Success Criteria:

#### Automated Verification:
- [x] All tests pass: `nix develop -c go test ./internal/review/...`
- [x] Package builds: `nix develop -c go build ./internal/review/`

#### Manual Verification:
- [x] None for this phase (unit tests cover behavior)

---

## Phase 2: Refactor HTTP Server

### Overview

Update the HTTP server to use the new review service. Behavior should be unchanged; this is a pure refactoring.

### Changes Required:

#### 1. Update HTTP Server

**File**: `internal/server/server.go`

Update the `Server` struct to include the review service:

```go
import (
	// ... existing imports ...
	"github.com/mchowning/diffguide/internal/review"
)

type Server struct {
	reviewService *review.Service
	server        *http.Server
	listener      net.Listener
	verbose       bool
}

func New(store *storage.Store, port string, verbose bool) (*Server, error) {
	mux := http.NewServeMux()

	s := &Server{
		reviewService: review.NewService(store),
		verbose:       verbose,
		server: &http.Server{
			// ... unchanged ...
		},
	}
	// ... rest unchanged ...
}
```

Update `handleReview` to delegate to the service with proper error handling:

```go
func (s *Server) handleReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBody)
	defer r.Body.Close()

	var reviewData model.Review
	if err := json.NewDecoder(r.Body).Decode(&reviewData); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "Request body too large (max 10MB)", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	_, err := s.reviewService.Submit(r.Context(), reviewData)
	if err != nil {
		// Use errors.Is for type-safe error checking
		if errors.Is(err, review.ErrMissingWorkingDirectory) {
			http.Error(w, "Missing workingDirectory field", http.StatusBadRequest)
			return
		}
		if errors.Is(err, review.ErrInvalidWorkingDirectory) {
			http.Error(w, "Invalid workingDirectory: "+err.Error(), http.StatusBadRequest)
			return
		}
		// Storage errors
		http.Error(w, "Failed to store review: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if s.verbose {
		log.Printf("Stored review for %s: %s (%d sections)",
			reviewData.WorkingDirectory, reviewData.Title, len(reviewData.Sections))
	}

	w.WriteHeader(http.StatusOK)
}
```

### Success Criteria:

#### Automated Verification:
- [x] All existing server tests pass: `nix develop -c go test ./internal/server/...`
- [x] All integration tests pass: `nix develop -c go test ./internal/integration/...`
- [x] All tests pass: `nix develop -c go test ./...`

#### Manual Verification:
- [x] HTTP server still works (covered by integration tests)

---

## Phase 3: MCP Server Package

### Overview

Create the MCP server package that uses the shared review service.

**Note**: Code snippets below are based on SDK documentation. Verify against actual SDK API in Phase 0 and adjust as needed.

### Changes Required:

#### 1. Create MCP Server Package

**File**: `internal/mcpserver/mcpserver.go`

```go
package mcpserver

import (
	"context"
	"errors"
	"io"

	"github.com/mchowning/diffguide/internal/model"
	"github.com/mchowning/diffguide/internal/review"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server wraps the MCP server with diffguide-specific functionality
type Server struct {
	server        *mcp.Server
	reviewService *review.Service
}

// SubmitReviewInput is the input schema for the submit_review tool
type SubmitReviewInput struct {
	WorkingDirectory string          `json:"workingDirectory" jsonschema:"required,description=Absolute path to the project directory"`
	Title            string          `json:"title" jsonschema:"description=Title for the review"`
	Sections         []model.Section `json:"sections" jsonschema:"description=Review sections with narratives and code hunks"`
}

// SubmitReviewOutput is the response from submit_review tool
type SubmitReviewOutput struct {
	Success  bool   `json:"success"`
	FilePath string `json:"filePath,omitempty" jsonschema:"description=Path where review was stored"`
	Error    string `json:"error,omitempty"`
}

// New creates a new MCP server with the given review service
func New(reviewService *review.Service) *Server {
	s := &Server{reviewService: reviewService}

	s.server = mcp.NewServer(&mcp.Implementation{
		Name:    "diffguide",
		Version: "1.0.0",
	}, nil)

	mcp.AddTool(s.server, &mcp.Tool{
		Name:        "submit_review",
		Description: "Submit a code review to diffguide for display in the TUI viewer",
	}, s.handleSubmitReview)

	return s
}

// Run starts the MCP server with the given reader/writer
// Note: stdout is reserved for MCP protocol; all logging must use stderr
func (s *Server) Run(ctx context.Context, reader io.Reader, writer io.Writer) error {
	transport := mcp.NewStdioTransport(reader, writer)
	return s.server.Run(ctx, transport)
}

// handleSubmitReview processes submit_review tool calls
func (s *Server) handleSubmitReview(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input SubmitReviewInput,
) (*mcp.CallToolResult, SubmitReviewOutput, error) {
	reviewData := model.Review{
		WorkingDirectory: input.WorkingDirectory,
		Title:            input.Title,
		Sections:         input.Sections,
	}

	result, err := s.reviewService.Submit(ctx, reviewData)
	if err != nil {
		// Validation/normalization errors: return structured response
		if errors.Is(err, review.ErrMissingWorkingDirectory) {
			return nil, SubmitReviewOutput{
				Success: false,
				Error:   "workingDirectory is required",
			}, nil
		}
		if errors.Is(err, review.ErrInvalidWorkingDirectory) {
			return nil, SubmitReviewOutput{
				Success: false,
				Error:   err.Error(),
			}, nil
		}
		// Storage errors: return structured response (not Go error)
		return nil, SubmitReviewOutput{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return nil, SubmitReviewOutput{
		Success:  true,
		FilePath: result.FilePath,
	}, nil
}
```

#### 2. Create MCP Server Tests

**File**: `internal/mcpserver/mcpserver_test.go`

```go
package mcpserver_test

import (
	"context"
	"testing"

	"github.com/mchowning/diffguide/internal/mcpserver"
	"github.com/mchowning/diffguide/internal/review"
	"github.com/mchowning/diffguide/internal/storage"
)

func setupTestServer(t *testing.T) (*mcpserver.Server, *storage.Store) {
	t.Helper()
	baseDir := t.TempDir()
	store, err := storage.NewStoreWithDir(baseDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	reviewService := review.NewService(store)
	srv := mcpserver.New(reviewService)
	return srv, store
}

// Tests verify MCP server correctly delegates to review service
// and translates results to MCP responses
```

Test cases:
1. `TestServer_SubmitReviewSuccess` - Valid input returns success with file path
2. `TestServer_SubmitReviewMissingWorkingDirectory` - Returns structured error (not Go error)
3. `TestServer_SubmitReviewInvalidWorkingDirectory` - Returns structured error
4. `TestServer_SubmitReviewStoresCorrectly` - Verify data persisted via storage

### Success Criteria:

#### Automated Verification:
- [x] All tests pass: `nix develop -c go test ./internal/mcpserver/...`
- [x] Package builds: `nix develop -c go build ./internal/mcpserver/`

#### Manual Verification:
- [x] None for this phase (unit tests cover behavior)

---

## Phase 4: CLI Integration

### Overview

Add the `mcp` subcommand to the CLI.

### Changes Required:

#### 1. Update Main Entry Point

**File**: `cmd/diffguide/main.go`

Add MCP subcommand detection after the existing server check (around line 23):

```go
if len(os.Args) > 1 && os.Args[1] == "mcp" {
	mcpCmd := flag.NewFlagSet("mcp", flag.ExitOnError)
	verbose := mcpCmd.Bool("v", false, "Enable verbose logging (logs to stderr)")
	mcpCmd.Parse(os.Args[2:])
	runMCP(*verbose)
	return
}
```

#### 2. Create MCP Subcommand Handler

**File**: `cmd/diffguide/mcp.go`

```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mchowning/diffguide/internal/mcpserver"
	"github.com/mchowning/diffguide/internal/review"
	"github.com/mchowning/diffguide/internal/storage"
)

// runMCP starts the MCP server on stdio.
// Important: stdout is reserved for MCP protocol communication.
// All logging goes to stderr via the standard log package.
func runMCP(verbose bool) {
	store, err := storage.NewStore()
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	reviewService := review.NewService(store)
	srv := mcpserver.New(reviewService)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-stop
		if verbose {
			log.Println("Shutting down MCP server...")
		}
		cancel()
	}()

	if verbose {
		log.Println("Starting MCP server on stdio...")
	}

	if err := srv.Run(ctx, os.Stdin, os.Stdout); err != nil && err != context.Canceled {
		log.Fatalf("MCP server error: %v", err)
	}
}
```

### Success Criteria:

#### Automated Verification:
- [x] Build succeeds: `nix develop -c go build ./cmd/diffguide/`
- [x] All tests pass: `nix develop -c go test ./...`

#### Manual Verification:
- [x] Running `diffguide mcp` starts without error
- [x] Ctrl+C cleanly shuts down

---

## Phase 5: End-to-End Verification

### Overview

Verify the MCP server works correctly with Claude Code.

### Success Criteria:

#### Automated Verification:
- [x] All tests pass: `nix develop -c go test ./...`
- [x] Binary builds: `nix develop -c go build ./cmd/diffguide/`

#### Manual Verification:
- [x] Configure Claude Code with MCP server:
  ```bash
  claude mcp add --transport stdio diffguide --scope user -- ~/dotfiles/bin/diffguide mcp
  ```
- [x] MCP server can discover and call `submit_review` tool (verified via stdio test)
- [ ] Submitted review appears in TUI viewer (`diffguide`)
- [x] Review data is correctly persisted (title, sections, hunks)

---

## Testing Strategy

### Unit Tests:
- **Review Service**: Validation, normalization, storage (source of truth)
- **HTTP Server**: Request parsing, error translation, delegates to service
- **MCP Server**: Input translation, error translation, delegates to service

### Integration Tests:
- HTTP round-trip (existing tests)
- MCP protocol message round-trip (if feasible with SDK)

### Manual Testing Steps:
1. Build: `nix develop -c go build ./cmd/diffguide/`
2. Test HTTP: `curl -X POST http://localhost:8765/review -d '{"workingDirectory":"/tmp/test","title":"Test"}'`
3. Test MCP: Configure in Claude Code, submit review
4. Verify both: `./diffguide` shows the review

## Architecture Summary

```
┌─────────────────┐     ┌─────────────────┐
│   HTTP Server   │     │   MCP Server    │
│ (server.go)     │     │ (mcpserver.go)  │
└────────┬────────┘     └────────┬────────┘
         │                       │
         │  delegates to         │  delegates to
         ▼                       ▼
    ┌─────────────────────────────────┐
    │        Review Service           │
    │     (review/service.go)         │
    │  - Validation                   │
    │  - Path normalization           │
    │  - Storage coordination         │
    └────────────────┬────────────────┘
                     │
                     ▼
    ┌─────────────────────────────────┐
    │          Storage Layer          │
    │     (storage/store.go)          │
    └─────────────────────────────────┘
```

## References

- Research document: `working-notes/2025-12-19_research_mcp-server-implementation.md`
- MVP plan: `working-notes/2025-12-18_plan_diffguide-mvp.md`
- HTTP server: `internal/server/server.go:70-116`
- Storage layer: `internal/storage/store.go`
- Official Go MCP SDK: https://github.com/modelcontextprotocol/go-sdk
