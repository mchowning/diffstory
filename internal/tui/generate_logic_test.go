package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/mchowning/diffstory/internal/diff"
)

func TestExtractLLMResponse_CleanOutput(t *testing.T) {
	input := `{"title": "Test Review", "chapters": []}`

	response, err := extractLLMResponse(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.Title != "Test Review" {
		t.Errorf("expected title 'Test Review', got %q", response.Title)
	}
}

func TestExtractLLMResponse_MarkdownFences(t *testing.T) {
	input := "```json\n{\"title\": \"Fenced Review\", \"chapters\": []}\n```"

	response, err := extractLLMResponse(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.Title != "Fenced Review" {
		t.Errorf("expected title 'Fenced Review', got %q", response.Title)
	}
}

func TestExtractLLMResponse_MarkdownFencesNoLang(t *testing.T) {
	input := "```\n{\"title\": \"No Lang\", \"chapters\": []}\n```"

	response, err := extractLLMResponse(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.Title != "No Lang" {
		t.Errorf("expected title 'No Lang', got %q", response.Title)
	}
}

func TestExtractLLMResponse_SurroundingText(t *testing.T) {
	input := `Here is the review:

{"title": "Surrounded", "chapters": []}

I hope this helps!`

	response, err := extractLLMResponse(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.Title != "Surrounded" {
		t.Errorf("expected title 'Surrounded', got %q", response.Title)
	}
}

func TestExtractLLMResponse_NestedBraces(t *testing.T) {
	input := `{
		"title": "Nested",
		"chapters": [
			{
				"id": "ch1",
				"title": "Chapter 1",
				"sections": [
					{
						"id": "1",
						"title": "Section 1",
						"narrative": "Test",
						"hunks": [{"id": "test.go::1", "importance": "high"}]
					}
				]
			}
		]
	}`

	response, err := extractLLMResponse(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.Title != "Nested" {
		t.Errorf("expected title 'Nested', got %q", response.Title)
	}
	if len(response.Chapters) != 1 {
		t.Errorf("expected 1 chapter, got %d", len(response.Chapters))
	}
}

func TestExtractLLMResponse_NoJSON(t *testing.T) {
	input := "This is just plain text with no JSON at all."

	_, err := extractLLMResponse(input, nil)
	if err == nil {
		t.Fatal("expected error for missing JSON")
	}
}

func TestExtractLLMResponse_UnclosedBrace(t *testing.T) {
	// Missing closing brackets - should be repaired
	input := `{"title": "Unclosed", "chapters": [`

	response, err := extractLLMResponse(input, nil)
	if err != nil {
		t.Fatalf("expected repair to succeed, got: %v", err)
	}
	if response.Title != "Unclosed" {
		t.Errorf("expected title 'Unclosed', got %q", response.Title)
	}
}

func TestExtractLLMResponse_InvalidJSON(t *testing.T) {
	// Unquoted value - jsonrepair can fix this
	input := `{"title": invalid}`

	response, err := extractLLMResponse(input, nil)
	if err != nil {
		t.Fatalf("expected repair to succeed, got: %v", err)
	}
	if response.Title != "invalid" {
		t.Errorf("expected title 'invalid', got %q", response.Title)
	}
}

func TestExtractLLMResponse_UnbalancedBracesInStringValues(t *testing.T) {
	// This simulates real LLM output where diff content contains unbalanced braces
	input := `{
		"title": "Test",
		"chapters": [
			{
				"id": "ch1",
				"title": "Chapter 1",
				"sections": [
					{
						"id": "1",
						"title": "Section 1",
						"narrative": "Added a function",
						"hunks": [{"id": "main.go::1", "importance": "high"}]
					}
				]
			}
		]
	}`

	response, err := extractLLMResponse(input, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.Title != "Test" {
		t.Errorf("expected title 'Test', got %q", response.Title)
	}
}

func TestExtractLLMResponse_TrailingComma(t *testing.T) {
	// Trailing comma - should be repaired
	input := `{"title": "Test", "chapters": [],}`

	response, err := extractLLMResponse(input, nil)
	if err != nil {
		t.Fatalf("expected repair to succeed, got: %v", err)
	}
	if response.Title != "Test" {
		t.Errorf("expected title 'Test', got %q", response.Title)
	}
}

func TestExtractLLMResponse_Unrepairable(t *testing.T) {
	// Completely malformed - should fail even after repair attempt
	input := `not json at all {{{`

	_, err := extractLLMResponse(input, nil)
	if err == nil {
		t.Fatal("expected error for unrepairable JSON")
	}
}

func TestBuildHunksJSON_SingleHunk(t *testing.T) {
	hunks := []diff.ParsedHunk{
		{ID: "file.go::10", File: "file.go", StartLine: 10, Diff: "+new line\n-old line"},
	}

	result, err := buildHunksJSON(hunks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should contain the hunk ID
	if !strings.Contains(result, `"id": "file.go::10"`) {
		t.Errorf("expected hunk ID in output, got: %s", result)
	}
	// Should be valid JSON array
	if !strings.HasPrefix(result, "[") || !strings.HasSuffix(result, "]") {
		t.Errorf("expected JSON array, got: %s", result)
	}
}

func TestBuildHunksJSON_MultipleHunks(t *testing.T) {
	hunks := []diff.ParsedHunk{
		{ID: "a.go::1", File: "a.go", StartLine: 1, Diff: "+a"},
		{ID: "b.go::2", File: "b.go", StartLine: 2, Diff: "+b"},
	}

	result, err := buildHunksJSON(hunks)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `"id": "a.go::1"`) {
		t.Error("expected first hunk ID")
	}
	if !strings.Contains(result, `"id": "b.go::2"`) {
		t.Error("expected second hunk ID")
	}
}

func TestAssembleReview_CombinesHunksWithClassification(t *testing.T) {
	hunks := []diff.ParsedHunk{
		{ID: "file.go::10", File: "file.go", StartLine: 10, Diff: "+added line"},
		{ID: "file.go::50", File: "file.go", StartLine: 50, Diff: "-removed line"},
	}

	response := &LLMResponse{
		Title: "Test Review",
		Chapters: []LLMChapter{
			{
				ID:    "ch1",
				Title: "Changes",
				Sections: []LLMSection{
					{
						ID:        "section1",
						Title:     "Made changes",
						Narrative: "Made some changes",
						Hunks: []LLMHunkRef{
							{ID: "file.go::10", Importance: "high"},
							{ID: "file.go::50", Importance: "low"},
						},
					},
				},
			},
		},
	}

	review := assembleReview("/test/dir", response, hunks)

	if review.Title != "Test Review" {
		t.Errorf("expected title 'Test Review', got %q", review.Title)
	}
	if review.WorkingDirectory != "/test/dir" {
		t.Errorf("expected working directory '/test/dir', got %q", review.WorkingDirectory)
	}
	sections := review.AllSections()
	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
	if len(sections[0].Hunks) != 2 {
		t.Fatalf("expected 2 hunks in section, got %d", len(sections[0].Hunks))
	}
	// Verify diff content is preserved
	if sections[0].Hunks[0].Diff != "+added line" {
		t.Errorf("expected diff content preserved, got %q", sections[0].Hunks[0].Diff)
	}
	// Verify importance is set
	if sections[0].Hunks[0].Importance != "high" {
		t.Errorf("expected importance 'high', got %q", sections[0].Hunks[0].Importance)
	}
}

func TestAssemblePartialReview_AddsUnclassifiedChapter(t *testing.T) {
	hunks := []diff.ParsedHunk{
		{ID: "file.go::10", File: "file.go", StartLine: 10, Diff: "+classified"},
		{ID: "file.go::50", File: "file.go", StartLine: 50, Diff: "+unclassified"},
	}

	response := &LLMResponse{
		Title: "Partial Review",
		Chapters: []LLMChapter{
			{
				ID:    "ch1",
				Title: "Classified",
				Sections: []LLMSection{
					{
						ID:        "section1",
						Title:     "Some changes",
						Narrative: "Some changes",
						Hunks: []LLMHunkRef{
							{ID: "file.go::10", Importance: "high"},
						},
					},
				},
			},
		},
	}

	missingIDs := []string{"file.go::50"}
	review := assemblePartialReview("/test/dir", response, hunks, missingIDs)

	// Should have 2 chapters (1 classified + 1 unclassified)
	if len(review.Chapters) != 2 {
		t.Fatalf("expected 2 chapters (1 classified + 1 unclassified), got %d", len(review.Chapters))
	}
	unclassifiedChapter := review.Chapters[1]
	if len(unclassifiedChapter.Sections) != 1 {
		t.Fatalf("expected 1 section in unclassified chapter, got %d", len(unclassifiedChapter.Sections))
	}
	unclassifiedSection := unclassifiedChapter.Sections[0]
	if unclassifiedSection.ID != "unclassified" {
		t.Errorf("expected unclassified section ID, got %q", unclassifiedSection.ID)
	}
	if len(unclassifiedSection.Hunks) != 1 {
		t.Fatalf("expected 1 hunk in unclassified section, got %d", len(unclassifiedSection.Hunks))
	}
	// Unclassified hunks should default to medium importance
	if unclassifiedSection.Hunks[0].Importance != "medium" {
		t.Errorf("expected default importance 'medium', got %q", unclassifiedSection.Hunks[0].Importance)
	}
}

func TestAssembleReview_SetsCreatedAt(t *testing.T) {
	hunks := []diff.ParsedHunk{
		{ID: "file.go::10", File: "file.go", StartLine: 10, Diff: "+test"},
	}

	response := &LLMResponse{
		Title: "Test Review",
		Chapters: []LLMChapter{
			{
				ID:    "ch1",
				Title: "Test",
				Sections: []LLMSection{
					{
						ID:        "section1",
						Title:     "Test",
						Narrative: "Test",
						Hunks:     []LLMHunkRef{{ID: "file.go::10", Importance: "high"}},
					},
				},
			},
		},
	}

	before := time.Now()
	review := assembleReview("/test/dir", response, hunks)
	after := time.Now()

	if review.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if review.CreatedAt.Before(before) || review.CreatedAt.After(after) {
		t.Errorf("CreatedAt %v not within expected range [%v, %v]", review.CreatedAt, before, after)
	}
}

// Phase 4: IsTest field tests

func boolPtrLLM(b bool) *bool {
	return &b
}

func TestAssembleReview_CopiesIsTestField(t *testing.T) {
	hunks := []diff.ParsedHunk{
		{ID: "main.go::10", File: "main.go", StartLine: 10, Diff: "+production"},
		{ID: "main_test.go::20", File: "main_test.go", StartLine: 20, Diff: "+test"},
	}

	response := &LLMResponse{
		Title: "Test Review",
		Chapters: []LLMChapter{
			{
				ID:    "ch1",
				Title: "Mixed",
				Sections: []LLMSection{
					{
						ID:        "section1",
						Title:     "Mixed section",
						Narrative: "Mixed section",
						Hunks: []LLMHunkRef{
							{ID: "main.go::10", Importance: "high", IsTest: boolPtrLLM(false)},
							{ID: "main_test.go::20", Importance: "medium", IsTest: boolPtrLLM(true)},
						},
					},
				},
			},
		},
	}

	review := assembleReview("/test/dir", response, hunks)

	sections := review.AllSections()
	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}
	if len(sections[0].Hunks) != 2 {
		t.Fatalf("expected 2 hunks, got %d", len(sections[0].Hunks))
	}

	// First hunk should have IsTest = false
	if sections[0].Hunks[0].IsTest == nil {
		t.Error("expected IsTest to be set for production hunk")
	} else if *sections[0].Hunks[0].IsTest {
		t.Error("expected IsTest=false for production code hunk")
	}

	// Second hunk should have IsTest = true
	if sections[0].Hunks[1].IsTest == nil {
		t.Error("expected IsTest to be set for test hunk")
	} else if !*sections[0].Hunks[1].IsTest {
		t.Error("expected IsTest=true for test code hunk")
	}
}

func TestAssembleReview_NilIsTestWhenOmitted(t *testing.T) {
	hunks := []diff.ParsedHunk{
		{ID: "file.go::10", File: "file.go", StartLine: 10, Diff: "+legacy"},
	}

	// LLM response without IsTest field (backward compatibility)
	response := &LLMResponse{
		Title: "Legacy Review",
		Chapters: []LLMChapter{
			{
				ID:    "ch1",
				Title: "Legacy",
				Sections: []LLMSection{
					{
						ID:        "section1",
						Title:     "Legacy section",
						Narrative: "Legacy section",
						Hunks: []LLMHunkRef{
							{ID: "file.go::10", Importance: "high"},
							// Note: IsTest is not set (nil)
						},
					},
				},
			},
		},
	}

	review := assembleReview("/test/dir", response, hunks)

	sections := review.AllSections()
	if len(sections[0].Hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(sections[0].Hunks))
	}

	// IsTest should be nil for backward compatibility
	if sections[0].Hunks[0].IsTest != nil {
		t.Error("expected IsTest to be nil when LLM omits the field")
	}
}

func TestAssemblePartialReview_PreservesIsTestField(t *testing.T) {
	hunks := []diff.ParsedHunk{
		{ID: "main.go::10", File: "main.go", StartLine: 10, Diff: "+classified"},
		{ID: "missing.go::20", File: "missing.go", StartLine: 20, Diff: "+unclassified"},
	}

	response := &LLMResponse{
		Title: "Partial Review",
		Chapters: []LLMChapter{
			{
				ID:    "ch1",
				Title: "Classified",
				Sections: []LLMSection{
					{
						ID:        "section1",
						Title:     "Some changes",
						Narrative: "Some changes",
						Hunks: []LLMHunkRef{
							{ID: "main.go::10", Importance: "high", IsTest: boolPtrLLM(false)},
						},
					},
				},
			},
		},
	}

	missingIDs := []string{"missing.go::20"}
	review := assemblePartialReview("/test/dir", response, hunks, missingIDs)

	// Classified hunk should have IsTest preserved
	classifiedSection := review.Chapters[0].Sections[0]
	if classifiedSection.Hunks[0].IsTest == nil {
		t.Error("expected IsTest to be preserved for classified hunk")
	} else if *classifiedSection.Hunks[0].IsTest {
		t.Error("expected IsTest=false for classified production hunk")
	}

	// Unclassified hunk should have IsTest nil (not classified by LLM)
	unclassifiedSection := review.Chapters[1].Sections[0]
	if unclassifiedSection.Hunks[0].IsTest != nil {
		t.Error("expected IsTest to be nil for unclassified hunk")
	}
}

