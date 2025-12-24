---
date: 2025-12-24 16:43:08 EST
git_commit: 2bba3119c5c03733c9d6ea5d88d0f6c79b907b13
branch: main
repository: diffguide
topic: "Review Generation Timestamp Display"
tags: [implementation, tui, model, timeutil]
last_updated: 2025-12-24
---

# Review Generation Timestamp Display

## Summary

Added a timestamp display to the TUI that shows when a review was generated, using dynamic relative time formatting (e.g., "Review generated 2 hours ago"). The timestamp appears as a status line below the header and updates on each view render.

## Overview

Reviews now track when they were created via a new `CreatedAt` field in the Review model. This timestamp is set automatically when reviews are created through either the server Submit API or the TUI generation flow. The TUI displays this timestamp in a subtle, italicized style below the header, formatted as a relative time that dynamically adjusts based on elapsed time (seconds, minutes, hours, days, weeks, months, years).

The implementation maintains backward compatibility with existing reviews that don't have timestamps—they simply don't display a timestamp line.

## Technical Details

### Model Layer

The `Review` struct was extended with a `CreatedAt` field using the `omitempty` JSON tag to ensure backward compatibility:

```go
type Review struct {
	WorkingDirectory string    `json:"workingDirectory"`
	Title            string    `json:"title"`
	Sections         []Section `json:"sections"`
	CreatedAt        time.Time `json:"createdAt,omitempty"`
}
```

The `omitempty` tag ensures that existing reviews stored without this field will deserialize correctly with a zero time value, and new reviews will include the timestamp in their JSON representation.

### Time Formatting Utility

A new `timeutil` package was created with a `FormatRelative` function that converts timestamps to human-readable relative times. The function is pure (takes `now` as a parameter) for testability:

```go
func FormatRelative(t time.Time, now time.Time) string {
	if t.IsZero() {
		return "unknown"
	}

	duration := now.Sub(t)
	seconds := int(duration.Seconds())

	if seconds < 60 {
		return pluralize(seconds, "second")
	}
	// ... continues through minutes, hours, days, weeks, months, years
}
```

The function handles time buckets progressively: seconds → minutes → hours → days → weeks → months → years. A helper `pluralize` function handles singular/plural formatting ("1 hour ago" vs "2 hours ago").

### Timestamp Setting

Timestamps are set in two creation paths:

1. **Server Submit path** (`internal/review/service.go:42-44`): The `Submit` function sets `CreatedAt` to the current time if not already provided, preserving any timestamp that was explicitly set by the caller.

2. **TUI generation path** (`internal/tui/generate.go:248`): The `assembleReview` function sets `CreatedAt` when assembling a review from LLM classification results.

### TUI Display

The timestamp is rendered as a status line between the header and content panels:

```go
func (m Model) renderTimestamp() string {
	if m.review == nil || m.review.CreatedAt.IsZero() {
		return ""
	}
	relative := timeutil.FormatRelative(m.review.CreatedAt, time.Now())
	return timestampStyle.Render("Review generated " + relative)
}
```

The `renderReviewState` function was modified to:
1. Calculate the timestamp line first
2. Reduce content height by 1 when a timestamp is present (to maintain proper panel sizing)
3. Include the timestamp line in the vertical layout between header and content

The timestamp uses a subtle style (`internal/tui/styles.go:36-39`): color 244 (dim gray) with italic formatting.

## Git References

**Branch**: `main`

**Commit Range**: Single commit

**Commits Documented**:

**2bba3119c5c03733c9d6ea5d88d0f6c79b907b13** (2025-12-24)
Add review generation timestamp display

Show when a review was generated with dynamic relative time formatting
(e.g., "Review generated 2 hours ago") in a status line below the header.

- Add CreatedAt field to Review model (backward compatible with omitempty)
- Add timeutil.FormatRelative() for human-readable durations
- Set timestamp when reviews are created via Submit() or TUI generation
- Display timestamp with subtle styling, updates on each view render
