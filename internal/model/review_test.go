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
		Chapters:         []Chapter{},
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
		"chapters": []
	}`

	var review Review
	if err := json.Unmarshal([]byte(legacyJSON), &review); err != nil {
		t.Fatalf("failed to unmarshal legacy JSON: %v", err)
	}

	if !review.CreatedAt.IsZero() {
		t.Errorf("CreatedAt should be zero time for legacy reviews, got %v", review.CreatedAt)
	}
}

func TestSection_Title_Serialization(t *testing.T) {
	section := Section{
		ID:        "section-1",
		Title:     "Add login handler",
		What: "Implements the login endpoint with bcrypt hashing.",
		Hunks:     []Hunk{},
	}

	data, err := json.Marshal(section)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled Section
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.Title != "Add login handler" {
		t.Errorf("Title = %q, want %q", unmarshaled.Title, "Add login handler")
	}
}

func TestChapter_Serialization(t *testing.T) {
	chapter := Chapter{
		ID:    "auth-chapter",
		Title: "Authentication",
		Sections: []Section{
			{
				ID:        "section-1",
				Title:     "Add login types",
				What: "Defines types for login.",
				Hunks:     []Hunk{},
			},
		},
	}

	data, err := json.Marshal(chapter)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled Chapter
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.ID != "auth-chapter" {
		t.Errorf("ID = %q, want %q", unmarshaled.ID, "auth-chapter")
	}
	if unmarshaled.Title != "Authentication" {
		t.Errorf("Title = %q, want %q", unmarshaled.Title, "Authentication")
	}
	if len(unmarshaled.Sections) != 1 {
		t.Fatalf("len(Sections) = %d, want 1", len(unmarshaled.Sections))
	}
}

func TestReview_WithChapters(t *testing.T) {
	review := Review{
		WorkingDirectory: "/test",
		Title:            "Test Review",
		Chapters: []Chapter{
			{
				ID:    "auth",
				Title: "Authentication",
				Sections: []Section{
					{ID: "s1", Title: "Login types", What: "Adds types", Hunks: []Hunk{}},
					{ID: "s2", Title: "Login handler", What: "Adds handler", Hunks: []Hunk{}},
				},
			},
			{
				ID:    "db",
				Title: "Database",
				Sections: []Section{
					{ID: "s3", Title: "User migration", What: "Adds migration", Hunks: []Hunk{}},
				},
			},
		},
	}

	data, err := json.Marshal(review)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled Review
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(unmarshaled.Chapters) != 2 {
		t.Fatalf("len(Chapters) = %d, want 2", len(unmarshaled.Chapters))
	}
	if unmarshaled.Chapters[0].Title != "Authentication" {
		t.Errorf("Chapters[0].Title = %q, want %q", unmarshaled.Chapters[0].Title, "Authentication")
	}
}

func TestReview_AllSections(t *testing.T) {
	review := Review{
		Chapters: []Chapter{
			{
				ID:    "ch1",
				Title: "Chapter 1",
				Sections: []Section{
					{ID: "s1", Title: "Section 1"},
					{ID: "s2", Title: "Section 2"},
				},
			},
			{
				ID:    "ch2",
				Title: "Chapter 2",
				Sections: []Section{
					{ID: "s3", Title: "Section 3"},
				},
			},
		},
	}

	sections := review.AllSections()

	if len(sections) != 3 {
		t.Fatalf("len(AllSections()) = %d, want 3", len(sections))
	}
	if sections[0].ID != "s1" {
		t.Errorf("sections[0].ID = %q, want %q", sections[0].ID, "s1")
	}
	if sections[2].ID != "s3" {
		t.Errorf("sections[2].ID = %q, want %q", sections[2].ID, "s3")
	}
}

func TestSection_WhatWhy_Serialization(t *testing.T) {
	section := Section{
		ID:    "section-1",
		Title: "Add login handler",
		What:  "Added rate limiting to login endpoint",
		Why:   "Prevents brute-force attacks by limiting failed attempts",
		Hunks: []Hunk{},
	}

	data, err := json.Marshal(section)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled Section
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.What != "Added rate limiting to login endpoint" {
		t.Errorf("What = %q, want %q", unmarshaled.What, "Added rate limiting to login endpoint")
	}
	if unmarshaled.Why != "Prevents brute-force attacks by limiting failed attempts" {
		t.Errorf("Why = %q, want %q", unmarshaled.Why, "Prevents brute-force attacks by limiting failed attempts")
	}
}
