---
date: 2026-01-09T15:18:41-05:00
git_commit: 7355b52ab7f23664b905fe257bd15adfe57e113f
branch: main
repository: diffstory
topic: "Untracked Files Warning Dialog"
tags: [implementation, tui, diff-generation, user-experience]
last_updated: 2026-01-09
---

# Untracked Files Warning Dialog

## Summary

Added a warning dialog that appears when users select "Uncommitted changes" as their diff source and untracked files exist. Users can proceed without untracked files, stage all files and proceed, or cancel.

## Overview

When generating code reviews from "Uncommitted changes", the underlying `git diff HEAD` command only shows changes to tracked files. New files that haven't been staged are silently excluded, which could lead users to believe their new files were reviewed when they weren't. This implementation adds a warning dialog that intercepts the generation flow when untracked files are detected, giving users explicit control over how to handle them.

The dialog appears after the user enters their reviewer context message and presses Enter to generate. It displays up to 5 untracked filenames (with a count for additional files) and offers three options: proceed without the untracked files (Space), stage all files with `git add .` and include them (Enter), or return to edit the context message (Escape).

## Technical Details

### New UI State for Warning Dialog

A new state `GenerateUIStateUntrackedWarning` was added to the `GenerateUIState` enum to manage when the warning dialog is displayed:

```go
const (
	// ... existing states
	GenerateUIStateUntrackedWarning
)
```

The model tracks untracked files in a new field that gets populated when the check occurs:

```go
type Model struct {
	// ... existing fields
	untrackedFiles []string
}
```

### Detection of Untracked Files

The `getUntrackedFiles` function in `internal/tui/generate.go` uses `git ls-files --others --exclude-standard` to find files that exist in the working directory but aren't tracked by git:

```go
func getUntrackedFiles(ctx context.Context, workDir string) ([]string, error) {
	output, err := runCommand(ctx, workDir, []string{"git", "ls-files", "--others", "--exclude-standard"}, nil)
	if err != nil {
		return nil, err
	}

	output = strings.TrimSpace(output)
	if output == "" {
		return nil, nil
	}

	return strings.Split(output, "\n"), nil
}
```

This approach excludes files matched by `.gitignore` patterns, which is the expected behavior for determining what the user likely intends to include.

### Triggering the Check

The untracked files check is triggered specifically for "Uncommitted changes" sources. The `isUncommittedChangesSource` helper identifies this source type by examining the command structure:

```go
func isUncommittedChangesSource(source DiffSource) bool {
	return len(source.Command) >= 3 &&
		source.Command[0] == "git" &&
		source.Command[1] == "diff" &&
		source.Command[2] == "HEAD"
}
```

When the user submits the context input (presses Enter after editing the reviewer instructions), the code checks if the selected source is "Uncommitted changes" and triggers the async check if so (`internal/tui/generate_ui.go:292-295`).

### Asynchronous Message Flow

The check runs asynchronously using Bubble Tea's command pattern. The `CheckUntrackedMsg` delivers the result back to the update loop:

```go
type CheckUntrackedMsg struct {
	Files []string
	Err   error
}
```

If files are found, the UI transitions to the warning state. If the check fails (e.g., git command error), generation proceeds without blocking. If no files are found, generation starts immediately.

### User Actions in Warning Dialog

The `updateUntrackedWarning` function handles user input in the warning state. Pressing Space clears the untracked files list and starts generation. Pressing Enter triggers a `git add .` command followed by generation. Pressing Escape returns to the context input state.

When the user chooses to stage all files, the `StageCompleteMsg` signals that staging succeeded and generation should begin (`internal/tui/update.go:428-430`).

### View Rendering

The warning dialog renders a centered dialog box with the list of untracked files (truncated to 5 with a count of remaining) and keybinding hints. The existing `dialogStyle` provides visual consistency with other dialogs in the application (`internal/tui/generate_ui.go:433-462`).

## Git References

**Branch**: `main`

**Commit Range**: `f2033fdbd50d..0b9ed8f41174`

**Commits Documented**:

**0b9ed8f411745cf9d14fbd267c3eea599f871c3f** (2026-01-08)
Add untracked files warning before generation

When users generate a review using "Uncommitted changes", untracked files
won't be included in the diff (git diff HEAD only shows tracked files).
This adds a warning dialog that appears after the user edits the LLM message
and presses enter to generate.

The dialog shows up to 5 untracked filenames with "(and N more...)" for
longer lists. Users can:
- Space: proceed without staging untracked files
- Enter: stage all files with git add and proceed
- Escape: return to edit the LLM message

Implementation:
- Add GenerateUIStateUntrackedWarning state
- Add getUntrackedFiles() helper using git ls-files
- Add renderUntrackedWarning() and updateUntrackedWarning() for UI
- Check for untracked files after context input submit (not before)
- All three user actions work with proper state transitions
- Comprehensive tests for all code paths
