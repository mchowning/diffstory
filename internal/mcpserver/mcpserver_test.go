package mcpserver_test

import (
	"context"
	"testing"

	"github.com/mchowning/diffguide/internal/mcpserver"
	"github.com/mchowning/diffguide/internal/model"
	"github.com/mchowning/diffguide/internal/storage"
)

func setupTestServer(t *testing.T) (*mcpserver.Server, *storage.Store) {
	t.Helper()
	baseDir := t.TempDir()
	store, err := storage.NewStoreWithDir(baseDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	srv := mcpserver.New(store)
	return srv, store
}

func TestServer_SubmitReviewSuccess(t *testing.T) {
	srv, _ := setupTestServer(t)
	ctx := context.Background()

	input := mcpserver.SubmitReviewInput{
		WorkingDirectory: "/test/project",
		Title:            "Test Review",
		Chapters:         []model.Chapter{},
	}

	result, err := srv.SubmitReview(ctx, input)
	if err != nil {
		t.Fatalf("SubmitReview returned error: %v", err)
	}

	if !result.Success {
		t.Errorf("expected Success=true, got false with error: %s", result.Error)
	}

	if result.FilePath == "" {
		t.Error("expected FilePath to be set")
	}
}

func TestServer_SubmitReviewMissingWorkingDirectory(t *testing.T) {
	srv, _ := setupTestServer(t)
	ctx := context.Background()

	input := mcpserver.SubmitReviewInput{
		Title: "Test Review",
		// WorkingDirectory intentionally omitted
	}

	result, err := srv.SubmitReview(ctx, input)
	if err != nil {
		t.Fatalf("SubmitReview returned Go error: %v (expected structured error)", err)
	}

	if result.Success {
		t.Error("expected Success=false for missing workingDirectory")
	}

	if result.Error == "" {
		t.Error("expected Error message to be set")
	}
}

func TestServer_SubmitReviewInvalidWorkingDirectory(t *testing.T) {
	srv, _ := setupTestServer(t)
	ctx := context.Background()

	input := mcpserver.SubmitReviewInput{
		WorkingDirectory: "\x00invalid", // null byte is invalid
		Title:            "Test Review",
	}

	result, err := srv.SubmitReview(ctx, input)
	if err != nil {
		t.Fatalf("SubmitReview returned Go error: %v (expected structured error)", err)
	}

	if result.Success {
		t.Error("expected Success=false for invalid workingDirectory")
	}

	if result.Error == "" {
		t.Error("expected Error message to be set")
	}
}

func TestServer_SubmitReviewStoresCorrectly(t *testing.T) {
	srv, store := setupTestServer(t)
	ctx := context.Background()

	input := mcpserver.SubmitReviewInput{
		WorkingDirectory: "/test/project",
		Title:            "Full Review",
		Chapters: []model.Chapter{
			{
				ID:    "chapter-1",
				Title: "Changes",
				Sections: []model.Section{
					{
						ID:        "section-1",
						Narrative: "First section narrative",
						Hunks: []model.Hunk{
							{
								File:       "main.go",
								StartLine:  10,
								Diff:       "@@ -10,3 +10,5 @@",
								Importance: "high",
							},
						},
					},
				},
			},
		},
	}

	result, err := srv.SubmitReview(ctx, input)
	if err != nil {
		t.Fatalf("SubmitReview returned error: %v", err)
	}

	if !result.Success {
		t.Fatalf("SubmitReview failed: %s", result.Error)
	}

	// Verify stored data
	stored, err := store.Read("/test/project")
	if err != nil {
		t.Fatalf("failed to read stored review: %v", err)
	}

	if stored.Title != input.Title {
		t.Errorf("Title = %q, want %q", stored.Title, input.Title)
	}

	sections := stored.AllSections()
	if len(sections) != 1 {
		t.Fatalf("Sections count = %d, want 1", len(sections))
	}

	s := sections[0]
	if s.ID != "section-1" {
		t.Errorf("Section.ID = %q, want %q", s.ID, "section-1")
	}
	if s.Narrative != "First section narrative" {
		t.Errorf("Section.Narrative = %q, want %q", s.Narrative, "First section narrative")
	}
	if len(s.Hunks) != 1 {
		t.Fatalf("Hunks count = %d, want 1", len(s.Hunks))
	}
	if s.Hunks[0].File != "main.go" {
		t.Errorf("Hunk.File = %q, want %q", s.Hunks[0].File, "main.go")
	}
}
