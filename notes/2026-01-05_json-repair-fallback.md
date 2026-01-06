---
date: 2026-01-06T06:53:57+00:00
git_commit: 1e22f1064f41bf78c5282c4233ad86ed7e7fb49c
branch: main
repository: diffstory
topic: "JSON Repair Fallback for LLM Output"
tags: [implementation, llm, json, reliability]
last_updated: 2026-01-06
---

# JSON Repair Fallback for LLM Output

## Summary

Added `github.com/kaptinlin/jsonrepair` as a fallback mechanism when the LLM returns malformed JSON, addressing a bug where missing closing brackets caused hard failures during diff review generation.

## Overview

The diffstory tool calls an LLM to generate structured JSON responses describing code changes. A bug occurred where the LLM returned JSON with mismatched brackets (29 `[` but only 28 `]`), causing a parse failure. Rather than retry the expensive LLM call or switch output formats, this implementation adds a repair step that attempts to fix common JSON syntax errors before parsing.

The repair fallback is transparent to the happy path: valid JSON parses directly without modification. Only when the initial parse fails does the repair library attempt to fix the JSON. Successful repairs are logged for monitoring purposes.

## Technical Details

### extractLLMResponse Function

The core change modifies `extractLLMResponse()` to accept a logger and implement a two-stage parsing approach:

```go
func extractLLMResponse(output string, logger *slog.Logger) (*LLMResponse, error) {
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
				"originalLength", len(jsonStr),
				"repairedLength", len(repaired))
		}
	}

	return &response, nil
}
```

The function signature changed to include the logger parameter (`internal/tui/generate.go:234`), and the call site was updated to pass the logger (`internal/tui/generate.go:173`).

### Dependency Addition

The `kaptinlin/jsonrepair` library was added to `go.mod`. This library handles common LLM JSON errors including:

- Missing closing brackets (the original bug)
- Trailing commas
- Unquoted values
- Single quotes instead of double quotes

### Test Changes

The test suite at `internal/tui/generate_logic_test.go` was updated in two ways:

1. All existing `extractLLMResponse` calls now pass `nil` for the logger parameter

2. Tests for malformed JSON were converted from expecting errors to expecting successful repair:

```go
func TestExtractLLMResponse_UnclosedBrace(t *testing.T) {
	input := `{"title": "Unclosed", "chapters": [`

	response, err := extractLLMResponse(input, nil)
	if err != nil {
		t.Fatalf("expected repair to succeed, got: %v", err)
	}
	if response.Title != "Unclosed" {
		t.Errorf("expected title 'Unclosed', got %q", response.Title)
	}
}
```

3. New tests were added for trailing comma repair and unrepairable JSON scenarios (`internal/tui/generate_logic_test.go:158-178`).

## Git References

**Branch**: `main`

**Commit Range**: `e5e394afd172b237c60938c5a4c36b18f377162f...1e22f1064f41bf78c5282c4233ad86ed7e7fb49c`

**Commits Documented**:

**1e22f1064f41bf78c5282c4233ad86ed7e7fb49c** (2026-01-05)
Add JSON repair fallback for LLM output reliability

- Add kaptinlin/jsonrepair dependency as fallback when LLM returns malformed JSON
- Update extractLLMResponse to accept logger and attempt repair on parse failure
- Log when repair is applied for monitoring and debugging
- Update all extractLLMResponse tests to pass logger parameter
- Add new tests for JSON repair scenarios (missing brackets, trailing commas, unrepairable)
- Modified existing error tests to expect successful repair instead of failure

This addresses issues where LLM can return syntactically invalid JSON (e.g., unmatched brackets)
by automatically repairing common errors like missing closing brackets, trailing commas, and
unquoted values before parsing.
