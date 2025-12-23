---
date: 2025-12-23 08:16:28 EST
git_commit: 29cb5ae36fc467b130e34f246b43256410e8fbe9
branch: main
repository: diffguide
topic: "Standalone LLM Review Generation"
tags: [implementation, tui, llm, config, logging, async]
last_updated: 2025-12-23
---

# Standalone LLM Review Generation

## Summary

This commit implements the ability to generate code reviews directly from the diffguide TUI by pressing `Shift+G`, which pipes a git diff to a configured LLM command and displays the resulting review. The feature includes a loading spinner, cancellation support, and configuration via JSON file.

## Overview

The standalone LLM review generation feature enables diffguide to work without requiring MCP server integration. Users configure an LLM command (such as Claude Code or OpenCode) in a JSON config file, then press `Shift+G` to generate a review of their current changes. The LLM output is parsed as JSON and displayed through the existing watcher infrastructure, ensuring generated reviews behave identically to those submitted via MCP.

Key capabilities added:
- Config file loading from `$XDG_CONFIG_HOME/diffguide/config.json` or `~/.config/diffguide/config.json`
- Async command execution with animated spinner during generation
- Cancellation flow via `Escape` → confirmation prompt → `y/n`
- Debug logging to `/tmp/diffguide.log` (enabled via `--debug` flag or config)
- JSON extraction from LLM output using `json.Decoder` (handles unbalanced braces in diff strings)

## Technical Details

### Config Package

A new `internal/config` package handles reading configuration from the XDG config directory hierarchy. The config file specifies the LLM command to invoke and optionally a custom diff command.

```go
type Config struct {
	LLMCommand          []string `json:"llmCommand"`
	DiffCommand         []string `json:"diffCommand"`
	DebugLoggingEnabled bool     `json:"debugLoggingEnabled"`
}
```

The `Load()` function checks `$XDG_CONFIG_HOME/diffguide/config.json` first, falling back to `~/.config/diffguide/config.json`. If no config file exists, it returns `nil` without error—the feature is simply unavailable. If `diffCommand` is not specified, it defaults to `["git", "diff", "HEAD"]`.

Example configuration (`~/.config/diffguide/config.json`):

```json
{
  "llmCommand": ["claude", "-p"],
  "diffCommand": ["git", "diff", "HEAD"],
  "debugLoggingEnabled": false
}
```

### Logging Infrastructure

The `internal/logging` package provides structured logging via `slog`. Logging is enabled by either the `--debug` command-line flag or the `debugLoggingEnabled` config option. When disabled, logs are discarded; when enabled, they write to `/tmp/diffguide.log`.

```go
func Setup(enabled bool) *slog.Logger {
	if !enabled {
		return slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	file, err := os.OpenFile(LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	// ...
	return slog.New(slog.NewTextHandler(file, nil))
}
```

### Generation Command

The `internal/tui/generate.go` file contains the LLM execution logic. When `Shift+G` is pressed, the TUI spawns an async command via `tea.Cmd` that:

1. Runs the configured diff command to get the current changes
2. Checks if the diff is empty (returns error if so)
3. Appends a prompt template to the LLM command arguments
4. Executes the LLM command with the diff as the final argument
5. Extracts JSON from the LLM response using `json.Decoder`
6. Writes the review to disk via the shared `storage.Store`

The JSON extraction uses `json.Decoder` instead of manual brace counting. This correctly handles diffs containing unbalanced braces in string values (e.g., `"diff": "func main() {"`):

```go
func extractReviewJSON(output string) (*model.Review, error) {
	start := strings.Index(output, "{")
	if start == -1 {
		return nil, fmt.Errorf("no JSON object found in response")
	}

	decoder := json.NewDecoder(strings.NewReader(output[start:]))
	var review model.Review
	if err := decoder.Decode(&review); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return &review, nil
}
```

### TUI State Management

The Model struct gains several fields to track generation state:

- `config *config.Config`: Loaded configuration
- `store *storage.Store`: Shared storage for persistence
- `isGenerating bool`: Whether generation is in progress
- `cancelGenerate context.CancelFunc`: For cancellation support
- `showCancelPrompt bool`: Whether the cancel confirmation is visible
- `spinner spinner.Model`: Animated loading indicator
- `logger *slog.Logger`: For debug logging

The cancellation flow uses `context.WithCancel`. When the user confirms cancellation, calling `cancelGenerate()` terminates the running command via `exec.CommandContext`.

### Watcher Integration

Rather than delivering the generated review directly via message, the generation command writes to disk through the shared `storage.Store`. The existing file watcher detects this change and delivers the review to the TUI via `ReviewReceivedMsg`. This ensures generated reviews follow the same flow as MCP-submitted reviews.

The watcher was updated to accept a store instance via `NewWithStore()`, and a new `WatchErrorMsg` type was added for error propagation.

### New Message Types

Three new message types support the generation lifecycle:

- `GenerateSuccessMsg`: Signals completion (review delivered via watcher)
- `GenerateErrorMsg`: Contains the error when generation fails
- `GenerateCancelledMsg`: Indicates user cancelled the operation

### Keybinding Registration

The `Shift+G` keybinding was added to the registry in `keybindings_init.go`:

```go
r.Register(Keybinding{Key: "G", Description: "Generate review (LLM)", Context: "global"})
```

This ensures the keybinding appears in the help overlay (`?`).

## Git References

**Branch**: `main`

**Commit Range**: `fbcf7e4c1e8f62163222b14663e87758445fb5ee...29cb5ae36fc467b130e34f246b43256410e8fbe9`

**Commits Documented**:

**29cb5ae36fc467b130e34f246b43256410e8fbe9** (2025-12-23T08:13:26-05:00)
Add standalone LLM review generation (Shift+G)

New feature: Generate code reviews directly in the TUI by pressing G.
The LLM generates a structured review which is persisted to disk and
displayed via the existing watcher infrastructure.

Key changes:
- internal/tui/generate.go: LLM command execution with JSON extraction
  - Uses json.Decoder instead of brace counting (fixes parsing of diffs
    containing unbalanced braces in string values)
  - Writes to disk via storage.Store; watcher delivers to TUI
- internal/config/: Config loading for llmCommand, diffCommand
- internal/logging/: Debug logging to /tmp/diffguide.log (--debug flag)
- TUI: Spinner during generation, cancellation with Esc→y/n
- Watcher: Now accepts shared store instance; handles WatchErrorMsg
- Tests for generation flow, cancellation, and error handling

**Pull Request**: Not yet created
