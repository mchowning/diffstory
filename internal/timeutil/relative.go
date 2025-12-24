package timeutil

import (
	"fmt"
	"time"
)

// FormatRelative returns a human-readable relative time string.
// Returns "unknown" for zero time.
func FormatRelative(t time.Time, now time.Time) string {
	if t.IsZero() {
		return "unknown"
	}

	duration := now.Sub(t)
	seconds := int(duration.Seconds())

	if seconds < 60 {
		return pluralize(seconds, "second")
	}

	minutes := seconds / 60
	if minutes < 60 {
		return pluralize(minutes, "minute")
	}

	hours := minutes / 60
	if hours < 24 {
		return pluralize(hours, "hour")
	}

	days := hours / 24
	if days < 7 {
		return pluralize(days, "day")
	}

	weeks := days / 7
	if weeks < 4 {
		return pluralize(weeks, "week")
	}

	months := days / 30
	if months < 12 {
		return pluralize(months, "month")
	}

	years := days / 365
	return pluralize(years, "year")
}

func pluralize(count int, unit string) string {
	if count == 1 {
		return fmt.Sprintf("1 %s ago", unit)
	}
	return fmt.Sprintf("%d %ss ago", count, unit)
}
