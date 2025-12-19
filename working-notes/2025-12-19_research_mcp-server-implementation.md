---
date: 2025-12-19T11:19:34-05:00
git_commit: ae2633f258f656090e17def71825a57472bae7b5
branch: main
repository: diffguide
topic: "MCP Server Implementation for diffguide"
tags: [research, mcp, go-sdk, claude-code, integration]
last_updated: 2025-12-19
last_updated_note: "Added design decisions for error reporting, logging, and shutdown"
---

# Research: MCP Server Implementation for diffguide

**Date**: 2025-12-19T11:19:34-05:00
**Git Commit**: ae2633f258f656090e17def71825a57472bae7b5
**Branch**: main
**Repository**: diffguide

## Research Question

How to create an MCP server that wraps the diffguide HTTP server functionality so Claude Code can communicate with it easily?

## Summary

An MCP (Model Context Protocol) server can be added to diffguide as a new subcommand (`diffguide mcp`) that exposes a `submit_review` tool. The implementation will use the official Go MCP SDK (`github.com/modelcontextprotocol/go-sdk/mcp`) and interact directly with the existing storage layer, avoiding the need to run a separate HTTP server.

Key findings:
1. **Official Go SDK available**: The `github.com/modelcontextprotocol/go-sdk` package provides a clean API for creating MCP servers
2. **Storage layer is ideal**: The existing `internal/storage` package handles all persistence logic and can be reused directly
3. **Stdio transport**: MCP servers for local CLI tools use stdio transport, which is simple and requires no network configuration
4. **Single tool needed**: Only `submit_review` is required based on user requirements

## Detailed Findings

### Existing diffguide Architecture

The codebase already has well-structured components that can be reused:

**Storage Layer** (`internal/storage/store.go`):
- `Store` struct with `Write(review model.Review)` method
- Handles path normalization via `NormalizePath()`
- Uses SHA256 hashing for directory-based file names
- Atomic writes (temp file + rename) for safety
- Reviews stored at `~/.diffguide/reviews/{hash}.json`

**Data Model** (`internal/model/review.go`):
```go
type Review struct {
    WorkingDirectory string    `json:"workingDirectory"`
    Title            string    `json:"title"`
    Sections         []Section `json:"sections"`
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

**Subcommand Pattern** (`cmd/diffguide/main.go`):
- Uses simple `os.Args` check for subcommands
- Pattern: `if len(os.Args) > 1 && os.Args[1] == "server" { ... }`
- Same pattern will work for `mcp` subcommand

### MCP Server Implementation

**Official Go SDK**: `github.com/modelcontextprotocol/go-sdk/mcp` (v1.1.0)

The SDK provides:
- `mcp.NewServer()` - Creates server instance
- `mcp.AddTool()` - Registers tools with JSON schema
- `mcp.StdioTransport{}` - Stdio transport for CLI integration
- Automatic JSON-RPC 2.0 protocol handling

**Basic Server Pattern**:
```go
server := mcp.NewServer(&mcp.Implementation{
    Name:    "diffguide",
    Version: "1.0.0",
}, nil)

mcp.AddTool(server, &mcp.Tool{
    Name:        "submit_review",
    Description: "Submit a code review to diffguide for display",
}, handleSubmitReview)

server.Run(context.Background(), &mcp.StdioTransport{})
```

**Tool Handler Signature**:
```go
func handleSubmitReview(
    ctx context.Context,
    req *mcp.CallToolRequest,
    input SubmitReviewInput,
) (*mcp.CallToolResult, SubmitReviewOutput, error)
```

### Proposed Design

**Package Structure**:
```
internal/
├── mcpserver/
│   ├── mcpserver.go      # MCP server implementation
│   └── mcpserver_test.go # Tests
```

**Tool: `submit_review`**

Input (mirrors `model.Review`):
```go
type SubmitReviewInput struct {
    WorkingDirectory string          `json:"workingDirectory" jsonschema:"required,description=Absolute path to the project directory"`
    Title            string          `json:"title" jsonschema:"description=Title for the review"`
    Sections         []model.Section `json:"sections" jsonschema:"description=Review sections with narratives and code hunks"`
}
```

Output:
```go
type SubmitReviewOutput struct {
    Success  bool   `json:"success"`
    FilePath string `json:"filePath,omitempty" jsonschema:"description=Path where review was stored"`
    Error    string `json:"error,omitempty"`
}
```

**Implementation Flow**:
1. MCP server receives `submit_review` call via stdio
2. Handler validates input (workingDirectory required)
3. Creates `model.Review` from input
4. Calls `store.Write(review)` to persist
5. Returns success with file path

### Integration with Claude Code

Claude Code discovers MCP servers via configuration. Users would add to their `.claude/settings.json`:

```json
{
  "mcpServers": {
    "diffguide": {
      "command": "diffguide",
      "args": ["mcp"]
    }
  }
}
```

Once configured, Claude Code can use the `submit_review` tool to send code reviews directly to diffguide.

## Code References

- `internal/storage/store.go:85-108` - `Write()` method that persists reviews
- `internal/storage/store.go:41-59` - `NormalizePath()` for consistent path handling
- `internal/model/review.go:1-21` - Data model definitions
- `cmd/diffguide/main.go:14-23` - Subcommand routing pattern
- `cmd/diffguide/server.go:16-45` - Example of subcommand implementation

## Architecture Insights

1. **Direct storage access is simpler**: Rather than making HTTP calls to the existing server, the MCP server can use the storage layer directly. This:
   - Eliminates need to run two processes
   - Reduces latency
   - Reuses tested code
   - Matches user preference for simplicity

2. **Stdio transport is appropriate**: For local CLI tools like diffguide, stdio is the standard MCP transport:
   - No port configuration needed
   - Works across all platforms
   - Simpler than HTTP-based transports

3. **Single tool is sufficient**: The user specified "submit only" - no need for read/list/delete operations. This keeps the MCP surface minimal.

## Historical Context

From `working-notes/2025-12-18_plan_diffguide-mvp.md`:
- MCP server wrapper was explicitly listed as "post-MVP" feature
- The architecture was designed to support this: "Single HTTP endpoint for MCP integration"
- This research validates the pre-planned approach

## Dependencies

**New dependency added**:
```
github.com/modelcontextprotocol/go-sdk v1.1.0
```

This brings in:
- `github.com/google/jsonschema-go` - JSON schema generation
- `github.com/yosida95/uritemplate/v3` - URI template handling
- `golang.org/x/oauth2` - OAuth support (not used for stdio)

## Implementation Checklist

- [ ] Create `internal/mcpserver/mcpserver.go` with server and tool implementation
- [ ] Create `internal/mcpserver/mcpserver_test.go` with TDD tests
- [ ] Add `mcp` subcommand handling in `cmd/diffguide/main.go`
- [ ] Create `cmd/diffguide/mcp.go` with `runMCP()` function
- [ ] Verify tests pass and server works end-to-end
- [ ] Document MCP configuration in README or separate doc

## Design Decisions

### 1. Error Reporting Strategy

**Question**: Should validation errors be returned as tool errors or as structured error responses?

**Option A: Return Go error (tool-level error)**
```go
return nil, SubmitReviewOutput{}, fmt.Errorf("workingDirectory is required")
```

| Pros | Cons |
|------|------|
| Simple, idiomatic Go | MCP marks entire tool call as failed |
| Clear failure signal | Claude may retry unnecessarily |
| Less boilerplate | No partial success possible |
| SDK handles error formatting | Less control over error presentation |

**Option B: Structured response with success/error fields**
```go
return nil, SubmitReviewOutput{Success: false, Error: "workingDirectory is required"}, nil
```

| Pros | Cons |
|------|------|
| Tool call "succeeds" - error is in payload | More verbose response type |
| Claude can interpret error semantically | Requires checking `success` field |
| Matches HTTP API pattern (400 vs 500) | Slightly more complex handler logic |
| Actionable guidance in error message | |

**Decision**: Use **Option B (structured response)** for validation/business errors, reserving Go errors for unexpected failures (storage I/O errors, panics). This matches MCP best practices: "Provide actionable guidance, not just error flags."

### 2. Logging Support

**Decision**: Yes, support verbose logging via `-v` flag, consistent with the HTTP server.

Implementation:
```go
// cmd/diffguide/mcp.go
func runMCP(verbose bool) {
    // Pass verbose flag to MCP server
}
```

Logging will go to stderr (not stdout) since stdout is reserved for MCP JSON-RPC communication.

### 3. Graceful Shutdown

**Decision**: Use context cancellation when stdin closes.

The MCP SDK's `StdioTransport` handles this automatically - when the client (Claude Code) disconnects, the transport closes and `server.Run()` returns. No special signal handling needed beyond what the SDK provides.

## Related Research

- `working-notes/2025-12-18_research_diffguide-implementation-technologies.md` - Original technology decisions
- `working-notes/2025-12-18_plan_diffguide-mvp.md` - MVP implementation plan

## Sources

- [Official Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk)
- [MCP Specification](https://modelcontextprotocol.io/specification/2025-06-18)
- [MCP Best Practices](https://modelcontextprotocol.info/docs/best-practices/)
