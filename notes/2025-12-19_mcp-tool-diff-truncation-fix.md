---
date: 2025-12-19 17:27:53 EST
git_commit: 84cfa5fe2a985f0972269973a0e74143a6e51e06
branch: main
repository: diffguide
topic: "MCP Tool Diff Truncation Fix"
tags: [implementation, mcp-server, tool-schema]
last_updated: 2025-12-19
---

# MCP Tool Diff Truncation Fix

## Summary

Updated the MCP server's `submit_review` tool description and JSON schema to explicitly request complete diff content and narrative summaries, addressing a bug where Claude Code was summarizing/truncating diffs instead of including complete content.

## Overview

When using the diffguide MCP server, Claude Code was generating truncated and summarized diff content rather than complete diffs. Investigation revealed the root cause: the tool description and schema provided no guidance about what content to include, leading Claude to interpret "review" as a summarization task.

The fix adds explicit instructions to both the tool description and JSON schema field descriptions. The tool description now requests a narrative structure with complete diffs, and each schema field has a description tag explaining the expected content.

## Technical Details

### Tool Description Enhancement

The original tool description was minimal:

```
Submit a code review for display in the diffguide TUI viewer
```

This provided no guidance about content expectations. The updated description explicitly requests:

1. A narrative structure that tells the story of the changes
2. Section summaries that are understandable in sequence
3. Complete diff content without truncation

```go
Description: "Submit a code review for display in the diffguide TUI viewer. " +
	"Structure the review as a narrative that tells the story of the changes - " +
	"someone reading just the section summaries in order should understand what changed and why. " +
	"Each section groups related changes with a narrative explaining the intent and context. " +
	"IMPORTANT: Include COMPLETE diff content for each hunk - do not summarize or truncate diffs.",
```

### JSON Schema Field Descriptions

The `Section` and `Hunk` types in `internal/model/review.go:10-21` received `jsonschema:"description=..."` tags. Two fields are particularly important for preventing truncation:

```go
Narrative  string `json:"narrative" jsonschema:"description=Summary explaining what changed and why - should be understandable without reading the diff"`
```

```go
Diff      string `json:"diff" jsonschema:"description=Complete unified diff content - include ALL lines, do not truncate or summarize"`
```

The Narrative field description establishes that narratives should be self-contained summaries, while the Diff field description explicitly prohibits truncation or summarization.

## Git References

**Branch**: `main`

**Commit Range**: Single commit

**Commits Documented**:

**84cfa5fe2a985f0972269973a0e74143a6e51e06** (2025-12-19T16:53:18-05:00)
Improve MCP tool guidance to prevent diff truncation

Claude Code was summarizing/truncating diffs when calling the submit_review
tool because the tool description didn't specify what content to include.

- Expand tool description to request narrative structure and complete diffs
- Add jsonschema description tags to Section and Hunk fields
- Narrative field: "should be understandable without reading the diff"
- Diff field: "include ALL lines, do not truncate or summarize"

[GitHub permalink](https://github.com/mchowning/diffguide/commit/84cfa5fe2a985f0972269973a0e74143a6e51e06)
