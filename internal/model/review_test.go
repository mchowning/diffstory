package model

import (
	"encoding/json"
	"testing"
	"time"
)

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

func TestReview_CreatedAt_Serialization(t *testing.T) {
	createdAt := time.Date(2024, 12, 24, 10, 30, 0, 0, time.UTC)
	review := Review{
		WorkingDirectory: "/test",
		Title:            "Test Review",
		Sections:         []Section{},
		CreatedAt:        createdAt,
	}

	data, err := json.Marshal(review)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled Review
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if !unmarshaled.CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt = %v, want %v", unmarshaled.CreatedAt, createdAt)
	}
}

func TestReview_LegacyJSON_WithoutCreatedAt(t *testing.T) {
	legacyJSON := `{
		"workingDirectory": "/test",
		"title": "Legacy Review",
		"sections": []
	}`

	var review Review
	if err := json.Unmarshal([]byte(legacyJSON), &review); err != nil {
		t.Fatalf("failed to unmarshal legacy JSON: %v", err)
	}

	if !review.CreatedAt.IsZero() {
		t.Errorf("CreatedAt should be zero time for legacy reviews, got %v", review.CreatedAt)
	}
}
