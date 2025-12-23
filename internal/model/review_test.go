package model

import "testing"

func TestValidImportance(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"high", true},
		{"medium", true},
		{"low", true},
		{"invalid", false},
		{"", false},
		{"High", false}, // case-sensitive
		{"HIGH", false},
		{"critical", false}, // synonyms are not valid, only canonical values
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ValidImportance(tt.input)
			if result != tt.expected {
				t.Errorf("ValidImportance(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeImportance(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// High variants
		{"high", "high"},
		{"High", "high"},
		{"HIGH", "high"},
		{"critical", "high"},
		{"Critical", "high"},
		{"important", "high"},

		// Medium variants
		{"medium", "medium"},
		{"Medium", "medium"},
		{"moderate", "medium"},
		{"normal", "medium"},

		// Low variants
		{"low", "low"},
		{"Low", "low"},
		{"minor", "low"},
		{"trivial", "low"},

		// Invalid returns empty string
		{"invalid", ""},
		{"", ""},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeImportance(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeImportance(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
