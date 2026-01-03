package tui_test

import (
	"testing"

	"github.com/mchowning/diffstory/internal/tui"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "truncates long string",
			input:    "hello",
			maxLen:   3,
			expected: "he…",
		},
		{
			name:     "no truncation needed",
			input:    "hi",
			maxLen:   10,
			expected: "hi",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   5,
			expected: "",
		},
		{
			name:     "maxLen of 1 returns ellipsis",
			input:    "abc",
			maxLen:   1,
			expected: "…",
		},
		{
			name:     "exact length no truncation",
			input:    "abc",
			maxLen:   3,
			expected: "abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tui.Truncate(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}
