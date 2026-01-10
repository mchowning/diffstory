package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReviewFromFile_ValidReview(t *testing.T) {
	// Use existing test fixture
	path := filepath.Join("..", "..", "internal", "testdata", "realistic_claude_review.json")

	review, err := loadReviewFromFile(path)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if review == nil {
		t.Fatal("expected review to be non-nil")
	}
	if review.Title != "Add user authentication middleware" {
		t.Errorf("unexpected title: %s", review.Title)
	}
	if len(review.Chapters) != 2 {
		t.Errorf("expected 2 chapters, got %d", len(review.Chapters))
	}
}

func TestLoadReviewFromFile_MissingFile(t *testing.T) {
	path := "/nonexistent/path/to/review.json"

	review, err := loadReviewFromFile(path)

	if err == nil {
		t.Fatal("expected error for missing file")
	}
	if review != nil {
		t.Fatal("expected review to be nil on error")
	}
}

func TestLoadReviewFromFile_InvalidJSON(t *testing.T) {
	// Create temp file with invalid JSON
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "invalid.json")
	err := os.WriteFile(path, []byte(`{invalid json`), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	review, err := loadReviewFromFile(path)

	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if review != nil {
		t.Fatal("expected review to be nil on error")
	}
}

func TestLoadReviewFromFile_InvalidImportance(t *testing.T) {
	// Create temp file with invalid importance value
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "invalid_importance.json")
	reviewJSON := `{
		"workingDirectory": "/test",
		"title": "Test Review",
		"chapters": [{
			"id": "ch-1",
			"title": "Chapter 1",
			"sections": [{
				"id": "sec-1",
				"narrative": "Test section",
				"hunks": [{
					"file": "test.go",
					"startLine": 1,
					"diff": "@@ -1 +1 @@\n-old\n+new",
					"importance": "critical"
				}]
			}]
		}]
	}`
	err := os.WriteFile(path, []byte(reviewJSON), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	review, err := loadReviewFromFile(path)

	if err == nil {
		t.Fatal("expected error for invalid importance")
	}
	if review != nil {
		t.Fatal("expected review to be nil on error")
	}
}

func TestLoadReviewFromFile_EmptyImportance(t *testing.T) {
	// Create temp file with empty importance (should be valid)
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "empty_importance.json")
	reviewJSON := `{
		"workingDirectory": "/test",
		"title": "Test Review",
		"chapters": [{
			"id": "ch-1",
			"title": "Chapter 1",
			"sections": [{
				"id": "sec-1",
				"narrative": "Test section",
				"hunks": [{
					"file": "test.go",
					"startLine": 1,
					"diff": "@@ -1 +1 @@\n-old\n+new",
					"importance": ""
				}]
			}]
		}]
	}`
	err := os.WriteFile(path, []byte(reviewJSON), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	review, err := loadReviewFromFile(path)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if review == nil {
		t.Fatal("expected review to be non-nil")
	}
}
