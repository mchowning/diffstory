package tui

import "github.com/mchowning/diffguide/internal/model"

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
