package tui

import (
	"github.com/mchowning/diffstory/internal/diff"
	"github.com/mchowning/diffstory/internal/model"
)

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

// ErrorMsg is sent when an error should be displayed in the status bar
type ErrorMsg struct {
	Err error
}

// ClearStatusMsg is sent to clear the status bar message
type ClearStatusMsg struct{}

// GenerateSuccessMsg signals that LLM generation completed successfully.
// The review is written to disk and will be delivered via the watcher.
type GenerateSuccessMsg struct{}

// GenerateErrorMsg indicates LLM generation failed
type GenerateErrorMsg struct {
	Err error
}

// GenerateCancelledMsg indicates user cancelled generation
type GenerateCancelledMsg struct{}

// CommitListMsg delivers the list of recent commits
type CommitListMsg struct {
	Commits []CommitInfo
}

// CommitListErrorMsg indicates failure to load commit list
type CommitListErrorMsg struct {
	Err error
}

// GenerateNeedsRetryMsg indicates validation failed and retry is needed
type GenerateNeedsRetryMsg struct {
	Hunks      []diff.ParsedHunk
	MissingIDs []string
	Context    string
}

// GenerateValidationFailedMsg indicates validation failed after retry
type GenerateValidationFailedMsg struct {
	Hunks      []diff.ParsedHunk
	Missing    []string
	Duplicates []string
	Invalid    []string
	Response   *LLMResponse // The partial response for "proceed with partial" option
}

// CheckUntrackedMsg delivers the result of checking for untracked files
type CheckUntrackedMsg struct {
	Files []string
	Err   error
}

// StageCompleteMsg signals that git add completed successfully
type StageCompleteMsg struct{}
