package tui

import (
	"encoding/json"
	"testing"

	"github.com/mchowning/diffstory/internal/diff"
)

func TestLLMChapter_UnmarshalJSON(t *testing.T) {
	jsonData := `{
		"id": "auth-chapter",
		"title": "Authentication",
		"sections": [
			{
				"id": "section-1",
				"title": "Add login types",
				"what": "Added types for login request and response",
				"why": "Needed for type-safe API communication",
				"hunks": [{"id": "auth.go::10", "importance": "high"}]
			}
		]
	}`

	var chapter LLMChapter
	err := json.Unmarshal([]byte(jsonData), &chapter)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if chapter.ID != "auth-chapter" {
		t.Errorf("ID = %q, want %q", chapter.ID, "auth-chapter")
	}
	if chapter.Title != "Authentication" {
		t.Errorf("Title = %q, want %q", chapter.Title, "Authentication")
	}
	if len(chapter.Sections) != 1 {
		t.Fatalf("len(Sections) = %d, want 1", len(chapter.Sections))
	}
	if chapter.Sections[0].Title != "Add login types" {
		t.Errorf("Section.Title = %q, want %q", chapter.Sections[0].Title, "Add login types")
	}
}

func TestLLMSection_UnmarshalJSON_WhatWhy(t *testing.T) {
	jsonData := `{
		"id": "section-1",
		"title": "Add rate limiting",
		"what": "Added rate limiting to login endpoint",
		"why": "Prevents brute-force attacks by limiting failed attempts",
		"hunks": [{"id": "auth.go::10", "importance": "high"}]
	}`

	var section LLMSection
	err := json.Unmarshal([]byte(jsonData), &section)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if section.What != "Added rate limiting to login endpoint" {
		t.Errorf("What = %q, want %q", section.What, "Added rate limiting to login endpoint")
	}
	if section.Why != "Prevents brute-force attacks by limiting failed attempts" {
		t.Errorf("Why = %q, want %q", section.Why, "Prevents brute-force attacks by limiting failed attempts")
	}
}

func TestValidateClassification_AllHunksPresent(t *testing.T) {
	inputHunks := []diff.ParsedHunk{
		{ID: "file.go::10", File: "file.go", StartLine: 10},
		{ID: "file.go::50", File: "file.go", StartLine: 50},
	}

	response := LLMResponse{
		Title: "Test Review",
		Chapters: []LLMChapter{
			{
				ID:    "ch1",
				Title: "Test Chapter",
				Sections: []LLMSection{
					{
						ID:        "section1",
						Title:     "Test section",
						What: "Test narrative",
						Hunks: []LLMHunkRef{
							{ID: "file.go::10", Importance: "high"},
							{ID: "file.go::50", Importance: "medium"},
						},
					},
				},
			},
		},
	}

	result := validateClassification(inputHunks, response)

	if !result.Valid {
		t.Errorf("expected Valid to be true, got false")
	}
	if len(result.MissingIDs) != 0 {
		t.Errorf("expected no missing IDs, got %v", result.MissingIDs)
	}
	if len(result.DuplicateIDs) != 0 {
		t.Errorf("expected no duplicate IDs, got %v", result.DuplicateIDs)
	}
	if len(result.InvalidImportance) != 0 {
		t.Errorf("expected no invalid importance, got %v", result.InvalidImportance)
	}
}

func TestValidateClassification_MissingHunks(t *testing.T) {
	inputHunks := []diff.ParsedHunk{
		{ID: "file.go::10", File: "file.go", StartLine: 10},
		{ID: "file.go::50", File: "file.go", StartLine: 50},
		{ID: "file.go::100", File: "file.go", StartLine: 100},
	}

	response := LLMResponse{
		Title: "Test Review",
		Chapters: []LLMChapter{
			{
				ID:    "ch1",
				Title: "Test Chapter",
				Sections: []LLMSection{
					{
						ID:        "section1",
						Title:     "Test section",
						What: "Test narrative",
						Hunks: []LLMHunkRef{
							{ID: "file.go::10", Importance: "high"},
							// file.go::50 and file.go::100 are missing
						},
					},
				},
			},
		},
	}

	result := validateClassification(inputHunks, response)

	if result.Valid {
		t.Errorf("expected Valid to be false")
	}
	if len(result.MissingIDs) != 2 {
		t.Errorf("expected 2 missing IDs, got %d: %v", len(result.MissingIDs), result.MissingIDs)
	}
}

func TestValidateClassification_DuplicateHunks(t *testing.T) {
	inputHunks := []diff.ParsedHunk{
		{ID: "file.go::10", File: "file.go", StartLine: 10},
	}

	response := LLMResponse{
		Title: "Test Review",
		Chapters: []LLMChapter{
			{
				ID:    "ch1",
				Title: "Test Chapter",
				Sections: []LLMSection{
					{
						ID:        "section1",
						Title:     "First section",
						What: "First section",
						Hunks:     []LLMHunkRef{{ID: "file.go::10", Importance: "high"}},
					},
					{
						ID:        "section2",
						Title:     "Second section",
						What: "Second section",
						Hunks:     []LLMHunkRef{{ID: "file.go::10", Importance: "medium"}}, // duplicate
					},
				},
			},
		},
	}

	result := validateClassification(inputHunks, response)

	if result.Valid {
		t.Errorf("expected Valid to be false due to duplicate")
	}
	if len(result.DuplicateIDs) != 1 {
		t.Errorf("expected 1 duplicate ID, got %d: %v", len(result.DuplicateIDs), result.DuplicateIDs)
	}
}

func TestValidateClassification_InvalidImportance(t *testing.T) {
	inputHunks := []diff.ParsedHunk{
		{ID: "file.go::10", File: "file.go", StartLine: 10},
		{ID: "file.go::50", File: "file.go", StartLine: 50},
	}

	response := LLMResponse{
		Title: "Test Review",
		Chapters: []LLMChapter{
			{
				ID:    "ch1",
				Title: "Test Chapter",
				Sections: []LLMSection{
					{
						ID:        "section1",
						Title:     "Test section",
						What: "Test narrative",
						Hunks: []LLMHunkRef{
							{ID: "file.go::10", Importance: "high"},
							{ID: "file.go::50", Importance: "invalid"}, // invalid importance
						},
					},
				},
			},
		},
	}

	result := validateClassification(inputHunks, response)

	if result.Valid {
		t.Errorf("expected Valid to be false due to invalid importance")
	}
	if len(result.InvalidImportance) != 1 {
		t.Errorf("expected 1 invalid importance, got %d: %v", len(result.InvalidImportance), result.InvalidImportance)
	}
}

func TestValidateClassification_NormalizesImportance(t *testing.T) {
	inputHunks := []diff.ParsedHunk{
		{ID: "file.go::10", File: "file.go", StartLine: 10},
	}

	response := LLMResponse{
		Title: "Test Review",
		Chapters: []LLMChapter{
			{
				ID:    "ch1",
				Title: "Test Chapter",
				Sections: []LLMSection{
					{
						ID:        "section1",
						Title:     "Test section",
						What: "Test narrative",
						Hunks: []LLMHunkRef{
							{ID: "file.go::10", Importance: "Critical"}, // should normalize to "high"
						},
					},
				},
			},
		},
	}

	result := validateClassification(inputHunks, response)

	if !result.Valid {
		t.Errorf("expected Valid to be true (Critical normalizes to high)")
	}
}

func TestValidateClassification_EmptyInput(t *testing.T) {
	inputHunks := []diff.ParsedHunk{}
	response := LLMResponse{Title: "Empty"}

	result := validateClassification(inputHunks, response)

	if !result.Valid {
		t.Errorf("expected Valid to be true for empty input")
	}
}
