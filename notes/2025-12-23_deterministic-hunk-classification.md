---
date: 2025-12-23 18:23:21 EST
git_commit: 4e48322
branch: main
repository: diffguide
topic: "Deterministic Hunk Classification"
tags: [implementation, diff-parsing, llm-integration, tui, mcp-server]
last_updated: 2025-12-23
---

# Deterministic Hunk Classification

## Summary

Implemented a deterministic diff parsing system that separates diff parsing (handled by the app) from classification (handled by the LLM). The app now parses unified diffs into hunks with unique IDs, sends structured data to the LLM for importance classification, and validates that all hunks receive classification. Importance was moved from section level to hunk level, and a new diff source picker UI was added.

## Overview

The previous implementation passed raw diff output directly to the LLM, which was responsible for both parsing diffs into hunks AND classifying them into sections. This caused truncation of long diffs, dropped hunks the LLM deemed unimportant, and modified diff content. The solution separates concerns: the app handles diff parsing deterministically while the LLM handles only semantic classification. A validation layer ensures 100% of hunks are classified before accepting the result.

The implementation includes:
- A new `internal/diff` package for parsing unified diffs
- UI components for selecting diff sources (staged, unstaged, HEAD, commit, range)
- Per-hunk importance levels (moved from section level - a breaking change)
- Validation logic with auto-retry on classification failures
- File-based input for large diffs to avoid token limits

## Technical Details

### Diff Parser

The new `internal/diff/parser.go` module parses unified diff output into individual hunks with unique identifiers. Each hunk receives an ID in the format `file/path.go::lineNumber` where the line number is extracted from the `+` side of the hunk header.

```go
type ParsedHunk struct {
	ID        string // format: "file/path.go::lineNumber"
	File      string
	StartLine int
	Diff      string // includes @@ header and content
}

func Parse(diffOutput string) ([]ParsedHunk, error) {
	var hunks []ParsedHunk
	fileDiffs := splitOnFileBoundaries(diffOutput)

	for _, fileDiff := range fileDiffs {
		if strings.Contains(fileDiff, "Binary files") {
			continue
		}
		filePath, err := extractFilePath(fileDiff)
		if err != nil {
			continue
		}
		fileHunks := splitIntoHunks(fileDiff, filePath)
		hunks = append(hunks, fileHunks...)
	}
	hunks = makeUniqueIDs(hunks)
	return hunks, nil
}
```

The parser handles edge cases including binary files (skipped), file renames (uses target path), and duplicate line numbers (appends `#N` suffix). Tests at `internal/diff/parser_test.go` verify behavior for various diff formats and confirm parsing 150+ hunks completes in under 1 second.

### Data Model Changes

Importance was moved from the `Section` struct to the `Hunk` struct (`internal/model/review.go`). This is a breaking change that invalidates stored reviews.

```go
type Section struct {
	ID        string `json:"id"`
	Narrative string `json:"narrative"`
	Hunks     []Hunk `json:"hunks"`
	// Importance field removed
}

type Hunk struct {
	File       string `json:"file"`
	StartLine  int    `json:"startLine"`
	Diff       string `json:"diff"`
	Importance string `json:"importance"` // Added: "high", "medium", or "low"
}
```

Constants `ImportanceHigh`, `ImportanceMedium`, and `ImportanceLow` were added along with `ValidImportance()` for validation and `NormalizeImportance()` to convert variants like "Critical" to "high".

### Generate Flow Refactoring

The generate command (`internal/tui/generate.go`) was refactored to follow a step-by-step flow:

1. Run diff command based on selected source
2. Parse diff into hunks using the new parser
3. Write hunks JSON to a file (`.diffguide-input.json`) to avoid token limits
4. Build prompt instructing LLM to read from file and return classifications
5. Call LLM and parse response
6. Validate all hunk IDs appear exactly once with valid importance
7. Auto-retry once on validation failure with stronger prompt
8. Assemble final review by injecting original diff content into classified structure

The `GenerateParams` struct tracks diff command, user context, retry state, and missing IDs:

```go
type GenerateParams struct {
	DiffCommand    []string
	Context        string
	IsRetry        bool
	MissingIDs     []string
	IsParseRetry   bool
}
```

### Validation Logic

The `validateClassification()` function in `internal/tui/validation.go` compares input hunk IDs against output:

```go
type ValidationResult struct {
	Valid             bool
	MissingIDs        []string
	DuplicateIDs      []string
	InvalidImportance []string
}

func validateClassification(inputHunks []diff.ParsedHunk, response LLMResponse) ValidationResult {
	result := ValidationResult{Valid: true}
	inputIDs := make(map[string]bool)
	for _, h := range inputHunks {
		inputIDs[h.ID] = true
	}
	outputIDs := make(map[string]int)
	for _, section := range response.Sections {
		for _, hunk := range section.Hunks {
			outputIDs[hunk.ID]++
			normalized := model.NormalizeImportance(hunk.Importance)
			if normalized == "" {
				result.InvalidImportance = append(result.InvalidImportance, hunk.ID)
				result.Valid = false
			}
		}
	}
	// Find missing and duplicate IDs...
	return result
}
```

### UI Components

New UI state management was added to `internal/tui/model.go` with a `GenerateUIState` enum tracking the generation flow phases:

- `GenerateUIStateNone`: Normal view
- `GenerateUIStateSourcePicker`: Selecting diff source
- `GenerateUIStateCommitSelector`: Selecting specific commit
- `GenerateUIStateCommitRangeStart/End`: Selecting commit range
- `GenerateUIStateContextInput`: Entering optional context
- `GenerateUIStateValidationError`: Showing failed hunks with retry options

The diff source picker (`internal/tui/diffsources.go`) provides 6 options: staged changes, working directory vs HEAD, unstaged only, changes since main, specific commit, and commit range.

Rendering functions in `internal/tui/generate_ui.go` handle each state:
- `renderSourcePicker()` displays diff source options with j/k navigation
- `renderCommitSelector()` shows recent commits with custom ref input
- `renderContextInput()` provides a textarea for optional LLM context
- `renderValidationError()` displays missing hunks with retry/proceed options

### File-Based Input for Large Diffs

To handle diffs exceeding the LLM's 128K token limit, hunks are written to `.diffguide-input.json` in the working directory. The prompt references this file path:

```go
inputPath := filepath.Join(workDir, ".diffguide-input.json")
if err := os.WriteFile(inputPath, hunksJSON, 0600); err != nil {
	return GenerateErrorMsg{Err: fmt.Errorf("failed to write hunks file: %w", err)}
}
defer os.Remove(inputPath)
```

This requires the LLM tool to have file system access. The file is cleaned up after generation completes.

### MCP Server Updates

The review service (`internal/review/service.go`) validates that all hunks have valid importance values, rejecting submissions with empty or invalid importance. The `Hunk.Importance` field has a jsonschema tag constraining values to the enum.

## Git References

**Branch**: `main`

**Commit Range**: `5e1cc25fb874...4e48322`

**Commits Documented**:

**762a541** (2025-12-23)
Add deterministic diff parser

Implement internal/diff package to reliably parse unified diff output into hunks:
- Parse() splits diff into individual code changes per file
- Each hunk gets a unique ID: file/path.go::lineNumber
- Handles git diff format with file boundaries
- Skips binary files
- Handles duplicate line numbers by appending increments

Prevents LLM from modifying or truncating diff content - only handles classification

https://github.com/mchowning/diffguide/blob/762a541/internal/diff/parser.go

**c7c46c6** (2025-12-23)
Add TUI components for diff review generation

- DiffSources: Picker for selecting diff source (staged, unstaged, HEAD, branch, file, command)
- Validation: IsValidImportance() and ValidateReviewStructure() helpers

These components enable new Shift+G flow for standalone LLM review generation

https://github.com/mchowning/diffguide/blob/c7c46c6/internal/tui/diffsources.go

**45fc251** (2025-12-23)
Refactor generate flow with deterministic diff parsing

Split generate.go logic into:
- generate_ui.go: UI rendering and interaction
- generate_logic_test.go: Behavior verification tests

Use diff parser to:
- Parse diffs into hunks deterministically
- Send only structured hunk data to LLM
- LLM returns only per-hunk importance classification
- Validate all hunks received classification before completing

https://github.com/mchowning/diffguide/blob/45fc251/internal/tui/generate.go

**402b6a8** (2025-12-23)
Update TUI for hunk-level importance and diff sources

- generate.go: Integrate diff parser, remove test file, simplify to 100 lines
- messages.go: Add GenerateStarted, HunkClassified, ReviewGenerated messages
- model.go: Add GenerateState tracking diffSource and hunks
- update.go: Handle generate flow state transitions with diff parsing
- view.go: Render classify success status

LLM now receives pre-parsed hunks with IDs, returns only importance values

https://github.com/mchowning/diffguide/blob/402b6a8/internal/tui/model.go

**cd6e42c** (2025-12-23)
Update review service for hunk-level importance

- service.go: Add Classify() to assign importance to hunks
- service_test.go: Test importance assignment with validation

Tests verify normalization of various importance labels to canonical values

https://github.com/mchowning/diffguide/blob/cd6e42c/internal/review/service.go

**ca4f48b** (2025-12-23)
Update testdata and dependencies for hunk-level importance

- testdata/*.json: Update all test review files with per-hunk importance
- payload_test.go: Update integration tests for new data structure
- mcpserver_test.go: Update MCP server tests
- go.mod/go.sum: Update dependencies

All test fixtures now reflect importance at hunk level (previously section level)

https://github.com/mchowning/diffguide/blob/ca4f48b/internal/testdata/simple_review.json

**05c92c2** (2025-12-23)
Add tests for commit list message handling

- TestUpdate_CommitListMsgPopulatesCommits: Verify commits populate in generate state
- TestUpdate_CommitListErrorMsgResetsUIState: Verify error handling resets UI

Tests verify the generate flow properly receives and handles commit list data

https://github.com/mchowning/diffguide/blob/05c92c2/internal/tui/update_test.go

**d04d6be** (2025-12-23)
Add tests for diff sources and validation components

- diffsources_test.go: Test diff source picker behavior
- validation_test.go: Test importance validation functions

Tests ensure UI components and helpers work correctly in generate flow

https://github.com/mchowning/diffguide/blob/d04d6be/internal/tui/diffsources_test.go

**4e48322** (2025-12-23)
Remove obsolete generate_test.go

Tests for extractReviewJSON which was replaced by extractLLMResponse during the deterministic hunk classification refactoring.
