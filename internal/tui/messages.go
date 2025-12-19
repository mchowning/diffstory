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
