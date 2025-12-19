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
