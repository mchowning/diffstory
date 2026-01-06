package tui_test

import (
	"testing"

	"github.com/mchowning/diffstory/internal/tui"
)

func TestCalcDescriptionPadding_MinimumPaddingForNarrowWidth(t *testing.T) {
	// Even for very narrow widths, padding should be at least 1
	result := tui.CalcDescriptionPadding(10)
	if result < 1 {
		t.Errorf("CalcDescriptionPadding(10) = %d, want at least 1", result)
	}
}

func TestCalcDescriptionPadding_GrowsWithWidth(t *testing.T) {
	// Padding should increase as width increases
	narrowPadding := tui.CalcDescriptionPadding(40)
	widePadding := tui.CalcDescriptionPadding(100)

	if widePadding <= narrowPadding {
		t.Errorf("CalcDescriptionPadding(100) = %d should be greater than CalcDescriptionPadding(40) = %d",
			widePadding, narrowPadding)
	}
}

func TestCalcDescriptionPadding_CapsAtMaximum(t *testing.T) {
	// Padding should cap at 8 for very wide terminals
	result := tui.CalcDescriptionPadding(500)
	if result != 8 {
		t.Errorf("CalcDescriptionPadding(500) = %d, want 8 (maximum)", result)
	}
}

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
