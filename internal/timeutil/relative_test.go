package timeutil

import (
	"testing"
	"time"
)

func TestFormatRelative(t *testing.T) {
	now := time.Date(2024, 12, 24, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		t        time.Time
		expected string
	}{
		// Zero time
		{"zero time", time.Time{}, "unknown"},

		// Seconds
		{"1 second ago", now.Add(-1 * time.Second), "1 second ago"},
		{"30 seconds ago", now.Add(-30 * time.Second), "30 seconds ago"},
		{"59 seconds ago", now.Add(-59 * time.Second), "59 seconds ago"},

		// Minutes
		{"1 minute ago", now.Add(-60 * time.Second), "1 minute ago"},
		{"2 minutes ago", now.Add(-2 * time.Minute), "2 minutes ago"},
		{"30 minutes ago", now.Add(-30 * time.Minute), "30 minutes ago"},
		{"59 minutes ago", now.Add(-59 * time.Minute), "59 minutes ago"},

		// Hours
		{"1 hour ago", now.Add(-60 * time.Minute), "1 hour ago"},
		{"2 hours ago", now.Add(-2 * time.Hour), "2 hours ago"},
		{"23 hours ago", now.Add(-23 * time.Hour), "23 hours ago"},

		// Days
		{"1 day ago", now.Add(-24 * time.Hour), "1 day ago"},
		{"2 days ago", now.Add(-48 * time.Hour), "2 days ago"},
		{"6 days ago", now.Add(-6 * 24 * time.Hour), "6 days ago"},

		// Weeks
		{"1 week ago", now.Add(-7 * 24 * time.Hour), "1 week ago"},
		{"2 weeks ago", now.Add(-14 * 24 * time.Hour), "2 weeks ago"},
		{"3 weeks ago", now.Add(-21 * 24 * time.Hour), "3 weeks ago"},

		// Months
		{"1 month ago", now.Add(-30 * 24 * time.Hour), "1 month ago"},
		{"2 months ago", now.Add(-60 * 24 * time.Hour), "2 months ago"},
		{"11 months ago", now.Add(-330 * 24 * time.Hour), "11 months ago"},

		// Years
		{"1 year ago", now.Add(-365 * 24 * time.Hour), "1 year ago"},
		{"2 years ago", now.Add(-730 * 24 * time.Hour), "2 years ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatRelative(tt.t, now)
			if result != tt.expected {
				t.Errorf("FormatRelative(%v, %v) = %q, want %q", tt.t, now, result, tt.expected)
			}
		})
	}
}
