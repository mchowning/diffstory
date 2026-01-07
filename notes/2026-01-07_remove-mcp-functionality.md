---
date: 2026-01-07T15:52:32-05:00
git_commit: 6f8df70f1cdbdc6fd4ae2fff431dc7b65a385235
branch: main
repository: diffstory
topic: "Remove MCP Server Functionality"
tags: [implementation, mcp, cleanup, refactoring]
last_updated: 2026-01-07
---

# Remove MCP Server Functionality

## Summary

Removed the MCP (Model Context Protocol) server from diffstory, making HTTP the sole integration method for Claude Code. This cleanup eliminated approximately 350 lines of code while maintaining all core TUI viewer and HTTP server functionality.

## Overview

The diffstory application previously supported two integration methods: an HTTP server and an MCP server (using the Model Context Protocol SDK). Both methods wrote to the same storage location, allowing the TUI viewer to display reviews from either source. The decision was made to simplify the codebase by removing MCP support, as the HTTP server provides equivalent functionality with a simpler integration model.

The MCP server was architected as a completely isolated component - it had its own package (`internal/mcpserver/`), its own entry point (`cmd/diffstory/mcp.go`), and only depended on shared infrastructure (the `review.Service` and `storage.Store`). Nothing else in the codebase depended on it, making removal low-risk.

## Technical Details

### MCP Server Removal

The MCP server package and command entry point were deleted entirely:

- `internal/mcpserver/mcpserver.go` - The MCP server implementation that wrapped the shared `review.Service`
- `internal/mcpserver/mcpserver_test.go` - Unit tests for the MCP server
- `internal/mcpserver/spike_test.go` - SDK exploration tests
- `cmd/diffstory/mcp.go` - The `runMCP()` function that bootstrapped the MCP server

The command handling in `cmd/diffstory/main.go` was simplified by removing the MCP command branch:

```go
// Removed from main.go
if len(os.Args) > 1 && os.Args[1] == "mcp" {
    mcpCmd := flag.NewFlagSet("mcp", flag.ExitOnError)
    verbose := mcpCmd.Bool("v", false, "Enable verbose logging (logs to stderr)")
    mcpCmd.Parse(os.Args[2:])
    runMCP(*verbose)
    return
}
```

### Dependency Cleanup

The MCP SDK dependency was removed from `go.mod`, and `go mod tidy` cleaned up the indirect dependencies that were only used by the MCP SDK:

- `github.com/modelcontextprotocol/go-sdk` - The MCP SDK itself
- `github.com/google/jsonschema-go` - Used by MCP SDK for schema generation
- `github.com/yosida95/uritemplate/v3` - Used by MCP SDK

The `jsonrepair` dependency was preserved because it is used by the TUI's review generation feature (`internal/tui/generate.go`), not just the MCP server.

### Model Struct Tag Cleanup

The model structs (`Chapter`, `Section`, `Hunk`) contained `jsonschema_description` struct tags that were used by the MCP SDK to generate JSON schemas for tool input validation. With MCP removed, these tags became dead code:

```go
// Before
type Hunk struct {
    File       string `json:"file" jsonschema_description:"File path relative to working directory"`
    StartLine  int    `json:"startLine" jsonschema_description:"Starting line number of the hunk"`
    Diff       string `json:"diff" jsonschema_description:"Complete unified diff content - include ALL lines, do not truncate or summarize"`
    Importance string `json:"importance" jsonschema:"enum=high,enum=medium,enum=low" jsonschema_description:"Importance level: high (critical changes), medium (significant changes), or low (minor changes)"`
    IsTest     *bool  `json:"isTest,omitempty" jsonschema_description:"True if this hunk contains test code changes, false for production code"`
}

// After
type Hunk struct {
    File       string `json:"file"`
    StartLine  int    `json:"startLine"`
    Diff       string `json:"diff"`
    Importance string `json:"importance"`
    IsTest     *bool  `json:"isTest,omitempty"`
}
```

The useful field guidance from these tags was moved to the README's "Review Data Format" section, where it serves as documentation for Claude when generating reviews via the HTTP API.

### Documentation Updates

The README was updated to reflect HTTP-only integration:

- Description changed from "HTTP or MCP" to "HTTP"
- Setup instructions simplified to just starting the HTTP server
- "MCP Server" section removed
- Architecture diagram updated to remove `mcp.go` and `mcpserver/`
- "How It Works" section updated to remove multi-source reference
- Field guidance added to the Review Data Format section

The Justfile `mcp` target was also removed.

## Git References

**Branch**: `main`

**Commit Range**: `cec98df...6f8df70`

**Commits Documented**:

**6f8df70f1cdbdc6fd4ae2fff431dc7b65a385235** (2026-01-07)
Remove MCP server functionality

The HTTP server is now the sole integration method for Claude Code.
This cleanup removes the MCP (Model Context Protocol) server entry point
and package, eliminating ~350 lines of code while maintaining all core
functionality.

Changes:
- Delete MCP server package (internal/mcpserver/)
- Remove MCP command entry point (cmd/diffstory/mcp.go)
- Remove MCP command handling from main.go
- Remove MCP SDK dependency from go.mod
- Remove jsonschema struct tags from model (MCP SDK dead code)
- Update README to document HTTP-only integration
- Remove MCP target from Justfile
- Add field guidance to Review Data Format section

All tests pass, no behavioral changes to HTTP server or TUI viewer.
