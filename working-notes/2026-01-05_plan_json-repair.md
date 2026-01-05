---
date: 2026-01-05T07:32:15+00:00
git_commit: e5e394afd172b237c60938c5a4c36b18f377162f
branch: main
repository: diffstory
topic: "Add JSON Repair for LLM Output Reliability"
tags: [plans, llm, json, reliability]
status: complete
last_updated: 2026-01-05
---

# Add JSON Repair for LLM Output Reliability

## Checklist

- [x] Phase 1: Add jsonrepair dependency
- [x] Phase 2: Implement repair fallback in extractLLMResponse
- [x] Phase 3: Add tests for JSON repair behavior

## Overview

Add `github.com/kaptinlin/jsonrepair` as a fallback when LLM output contains malformed JSON. This addresses the bug documented in `working-notes/2026-01-01_research_llm-json-output-reliability.md` where the LLM returned JSON with a missing closing bracket (29 `[` but only 28 `]`).

## Current State Analysis

The `extractLLMResponse()` function at `internal/tui/generate.go:233-249` currently:
1. Finds the first `{` character in LLM output
2. Uses `json.NewDecoder` to parse the JSON
3. Returns a wrapped error on parse failure, which becomes `GenerateErrorMsg`

**Problem**: No handling for JSON syntax errors. Malformed JSON causes a hard failure with no recovery attempt.

## Desired End State

When the LLM returns malformed JSON:
1. Initial parse attempt fails
2. `jsonrepair.JSONRepair()` attempts to fix common issues (missing brackets, trailing commas, etc.)
3. Repaired JSON is parsed
4. Success: Log that repair was applied (for monitoring)
5. Failure: Return error indicating both original and repair failures

### Verification

- Unit tests cover: valid JSON (no repair), repairable JSON (repair succeeds), unrepairable JSON (error)
- Manual test: Run diffstory on a real diff and verify generation works
- Check `/tmp/diffstory.log` shows repair logging when applicable

## What We're NOT Doing

- Switching to YAML or other formats (research showed lack of repair tools makes this riskier)
- Adding retry logic for JSON syntax errors (repair is more efficient than re-calling LLM)
- Migrating to Claude API structured outputs (requires separate API billing)

## Implementation Approach

Minimal change: wrap the existing parse logic with a repair fallback. The happy path (valid JSON) is unchanged. Only on parse failure do we attempt repair.

---

## Phase 1: Add jsonrepair dependency

### Overview

Add the `kaptinlin/jsonrepair` Go library to the project dependencies.

### Changes Required:

#### 1. Update go.mod

**Command**: `go get github.com/kaptinlin/jsonrepair`

This will add the dependency and update go.sum.

### Success Criteria:

#### Automated Verification:
- [x] Dependency added: `grep jsonrepair go.mod` shows the import
- [x] Build succeeds: `nix develop -c go build ./...`

---

## Phase 2: Implement repair fallback in extractLLMResponse

### Overview

Modify `extractLLMResponse()` to attempt JSON repair when initial parsing fails.

### Changes Required:

#### 1. Update imports in generate.go

**File**: `internal/tui/generate.go`

Add import:
```go
"github.com/kaptinlin/jsonrepair"
```

#### 2. Modify extractLLMResponse function

**File**: `internal/tui/generate.go`
**Function**: `extractLLMResponse` (lines 233-249)

Replace the current implementation with:

```go
// extractLLMResponse parses the LLM output to find JSON response
func extractLLMResponse(output string, logger *slog.Logger) (*LLMResponse, error) {
	// Find the first '{' character (LLM may include preamble)
	start := strings.Index(output, "{")
	if start == -1 {
		return nil, fmt.Errorf("no JSON object found in response")
	}

	jsonStr := output[start:]

	// Try parsing as-is first
	decoder := json.NewDecoder(strings.NewReader(jsonStr))
	var response LLMResponse
	if err := decoder.Decode(&response); err != nil {
		// Attempt repair
		repaired, repairErr := jsonrepair.JSONRepair(jsonStr)
		if repairErr != nil {
			return nil, fmt.Errorf("failed to decode JSON: %w (repair also failed: %v)", err, repairErr)
		}

		// Try parsing repaired JSON
		decoder = json.NewDecoder(strings.NewReader(repaired))
		if err := decoder.Decode(&response); err != nil {
			return nil, fmt.Errorf("failed to decode repaired JSON: %w", err)
		}

		// Log that repair was needed
		if logger != nil {
			logger.Info("JSON repair applied to LLM output",
				"originalError", err.Error(),
				"originalLength", len(jsonStr),
				"repairedLength", len(repaired))
		}
	}

	return &response, nil
}
```

#### 3. Update call site to pass logger

**File**: `internal/tui/generate.go`
**Line**: 172

Change from:
```go
response, err := extractLLMResponse(output)
```

To:
```go
response, err := extractLLMResponse(output, logger)
```

### Success Criteria:

#### Automated Verification:
- [x] Build succeeds: `nix develop -c go build ./...`
- [x] Linting passes: `nix develop -c golangci-lint run` (no new issues in changed files)
- [x] Type checking passes (implicit in build)

#### Manual Verification:
- [ ] Run diffstory on a real diff, verify generation completes
- [ ] Check `/tmp/diffstory.log` for any repair messages

---

## Phase 3: Add tests for JSON repair behavior

### Overview

Add unit tests to verify the repair fallback works correctly for valid, repairable, and unrepairable JSON.

### Changes Required:

#### 1. Create or update test file

**File**: `internal/tui/generate_test.go`

Add tests:

```go
func TestExtractLLMResponse_ValidJSON(t *testing.T) {
	input := `{"title": "Test", "chapters": []}`
	response, err := extractLLMResponse(input, nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if response.Title != "Test" {
		t.Errorf("expected title 'Test', got '%s'", response.Title)
	}
}

func TestExtractLLMResponse_MissingBracket(t *testing.T) {
	// Missing closing bracket - should be repaired
	input := `{"title": "Test", "chapters": [}`
	response, err := extractLLMResponse(input, nil)
	if err != nil {
		t.Fatalf("expected repair to succeed, got: %v", err)
	}
	if response.Title != "Test" {
		t.Errorf("expected title 'Test', got '%s'", response.Title)
	}
}

func TestExtractLLMResponse_TrailingComma(t *testing.T) {
	// Trailing comma - should be repaired
	input := `{"title": "Test", "chapters": [],}`
	response, err := extractLLMResponse(input, nil)
	if err != nil {
		t.Fatalf("expected repair to succeed, got: %v", err)
	}
	if response.Title != "Test" {
		t.Errorf("expected title 'Test', got '%s'", response.Title)
	}
}

func TestExtractLLMResponse_Unrepairable(t *testing.T) {
	// Completely malformed - should fail
	input := `not json at all {{{`
	_, err := extractLLMResponse(input, nil)
	if err == nil {
		t.Fatal("expected error for unrepairable JSON")
	}
}

func TestExtractLLMResponse_WithPreamble(t *testing.T) {
	// LLM sometimes includes text before JSON
	input := `Here is the JSON response:
{"title": "Test", "chapters": []}`
	response, err := extractLLMResponse(input, nil)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if response.Title != "Test" {
		t.Errorf("expected title 'Test', got '%s'", response.Title)
	}
}
```

### Success Criteria:

#### Automated Verification:
- [x] All tests pass: `nix develop -c go test ./internal/tui/...`
- [x] Build succeeds: `nix develop -c go build ./...`

---

## Testing Strategy

### Unit Tests:
- Valid JSON parses without repair
- JSON with missing brackets triggers repair and succeeds
- JSON with trailing commas triggers repair and succeeds
- Completely malformed input returns error
- JSON with preamble text is found and parsed

### Manual Testing Steps:
1. Run `diffstory` on a repository with staged changes
2. Verify the review generates successfully
3. Enable debug logging and check `/tmp/diffstory.log`
4. Confirm no unexpected repair messages for normal operation

## References

- Research document: `working-notes/2026-01-01_research_llm-json-output-reliability.md`
- jsonrepair library: https://github.com/kaptinlin/jsonrepair
- Current implementation: `internal/tui/generate.go:233-249`
