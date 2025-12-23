package tui

import (
	"testing"
)

func TestExtractReviewJSON_CleanOutput(t *testing.T) {
	input := `{"title": "Test Review", "sections": []}`

	review, err := extractReviewJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if review.Title != "Test Review" {
		t.Errorf("expected title 'Test Review', got %q", review.Title)
	}
}

func TestExtractReviewJSON_MarkdownFences(t *testing.T) {
	input := "```json\n{\"title\": \"Fenced Review\", \"sections\": []}\n```"

	review, err := extractReviewJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if review.Title != "Fenced Review" {
		t.Errorf("expected title 'Fenced Review', got %q", review.Title)
	}
}

func TestExtractReviewJSON_MarkdownFencesNoLang(t *testing.T) {
	input := "```\n{\"title\": \"No Lang\", \"sections\": []}\n```"

	review, err := extractReviewJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if review.Title != "No Lang" {
		t.Errorf("expected title 'No Lang', got %q", review.Title)
	}
}

func TestExtractReviewJSON_SurroundingText(t *testing.T) {
	input := `Here is the review:

{"title": "Surrounded", "sections": []}

I hope this helps!`

	review, err := extractReviewJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if review.Title != "Surrounded" {
		t.Errorf("expected title 'Surrounded', got %q", review.Title)
	}
}

func TestExtractReviewJSON_NestedBraces(t *testing.T) {
	input := `{
		"title": "Nested",
		"sections": [
			{
				"id": "1",
				"narrative": "Test",
				"importance": "high",
				"hunks": [{"file": "test.go", "startLine": 1, "diff": "{}"}]
			}
		]
	}`

	review, err := extractReviewJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if review.Title != "Nested" {
		t.Errorf("expected title 'Nested', got %q", review.Title)
	}
	if len(review.Sections) != 1 {
		t.Errorf("expected 1 section, got %d", len(review.Sections))
	}
}

func TestExtractReviewJSON_NoJSON(t *testing.T) {
	input := "This is just plain text with no JSON at all."

	_, err := extractReviewJSON(input)
	if err == nil {
		t.Fatal("expected error for missing JSON")
	}
}

func TestExtractReviewJSON_UnclosedBrace(t *testing.T) {
	input := `{"title": "Unclosed", "sections": [`

	_, err := extractReviewJSON(input)
	if err == nil {
		t.Fatal("expected error for unclosed brace")
	}
}

func TestExtractReviewJSON_InvalidJSON(t *testing.T) {
	input := `{"title": invalid}`

	_, err := extractReviewJSON(input)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestExtractReviewJSON_UnbalancedBracesInStringValues(t *testing.T) {
	// This simulates real LLM output where diff content contains unbalanced braces
	input := `{
		"title": "Test",
		"sections": [
			{
				"id": "1",
				"narrative": "Added a function",
				"importance": "high",
				"hunks": [{"file": "main.go", "startLine": 1, "diff": "func main() {"}]
			}
		]
	}`

	review, err := extractReviewJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if review.Title != "Test" {
		t.Errorf("expected title 'Test', got %q", review.Title)
	}
}
