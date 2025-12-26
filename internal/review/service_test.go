package review_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mchowning/diffguide/internal/model"
	"github.com/mchowning/diffguide/internal/review"
	"github.com/mchowning/diffguide/internal/storage"
)

func setupTestService(t *testing.T) (*review.Service, *storage.Store) {
	t.Helper()
	baseDir := t.TempDir()
	store, err := storage.NewStoreWithDir(baseDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	svc := review.NewService(store)
	return svc, store
}

func TestService_SubmitWithValidReview(t *testing.T) {
	svc, store := setupTestService(t)
	ctx := context.Background()

	input := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test Review",
	}

	result, err := svc.Submit(ctx, input)
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	if result.FilePath == "" {
		t.Error("expected FilePath to be set")
	}

	// Verify stored
	stored, err := store.Read("/test/project")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if stored.Title != input.Title {
		t.Errorf("Title = %q, want %q", stored.Title, input.Title)
	}
}

func TestService_SubmitMissingWorkingDirectory(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	input := model.Review{
		Title: "Test Review",
		// WorkingDirectory intentionally omitted
	}

	_, err := svc.Submit(ctx, input)
	if !errors.Is(err, review.ErrMissingWorkingDirectory) {
		t.Errorf("expected ErrMissingWorkingDirectory, got %v", err)
	}
}

func TestService_SubmitInvalidWorkingDirectory(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	input := model.Review{
		WorkingDirectory: "\x00invalid", // null byte is invalid in paths
		Title:            "Test Review",
	}

	_, err := svc.Submit(ctx, input)
	if !errors.Is(err, review.ErrInvalidWorkingDirectory) {
		t.Errorf("expected ErrInvalidWorkingDirectory, got %v", err)
	}
}

func TestService_SubmitNormalizesPath(t *testing.T) {
	svc, store := setupTestService(t)
	ctx := context.Background()

	// Submit with trailing slash
	input := model.Review{
		WorkingDirectory: "/test/project/",
		Title:            "Test Review",
	}

	_, err := svc.Submit(ctx, input)
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	// Should be readable without trailing slash (normalized)
	stored, err := store.Read("/test/project")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if stored.Title != input.Title {
		t.Errorf("Title = %q, want %q", stored.Title, input.Title)
	}
}

func TestService_SubmitPreservesAllFields(t *testing.T) {
	svc, store := setupTestService(t)
	ctx := context.Background()

	input := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Full Review",
		Chapters: []model.Chapter{
			{
				ID:    "chapter-1",
				Title: "Changes",
				Sections: []model.Section{
					{
						ID:        "section-1",
						Narrative: "This is the first section",
						Hunks: []model.Hunk{
							{
								File:       "main.go",
								StartLine:  10,
								Diff:       "@@ -10,3 +10,5 @@\n func main() {\n+    fmt.Println(\"hello\")\n }",
								Importance: "high",
							},
						},
					},
					{
						ID:        "section-2",
						Narrative: "Second section",
						Hunks:     []model.Hunk{},
					},
				},
			},
		},
	}

	_, err := svc.Submit(ctx, input)
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	stored, err := store.Read("/test/project")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if stored.Title != input.Title {
		t.Errorf("Title = %q, want %q", stored.Title, input.Title)
	}

	sections := stored.AllSections()
	if len(sections) != 2 {
		t.Fatalf("Sections count = %d, want 2", len(sections))
	}

	// Check first section
	s1 := sections[0]
	if s1.ID != "section-1" {
		t.Errorf("Section[0].ID = %q, want %q", s1.ID, "section-1")
	}
	if s1.Narrative != "This is the first section" {
		t.Errorf("Section[0].Narrative = %q, want %q", s1.Narrative, "This is the first section")
	}
	if len(s1.Hunks) != 1 {
		t.Fatalf("Section[0].Hunks count = %d, want 1", len(s1.Hunks))
	}

	h := s1.Hunks[0]
	if h.File != "main.go" {
		t.Errorf("Hunk.File = %q, want %q", h.File, "main.go")
	}
	if h.StartLine != 10 {
		t.Errorf("Hunk.StartLine = %d, want %d", h.StartLine, 10)
	}
	if h.Diff != input.Chapters[0].Sections[0].Hunks[0].Diff {
		t.Errorf("Hunk.Diff mismatch")
	}
	if h.Importance != "high" {
		t.Errorf("Hunk.Importance = %q, want %q", h.Importance, "high")
	}

	// Check second section
	s2 := sections[1]
	if s2.ID != "section-2" {
		t.Errorf("Section[1].ID = %q, want %q", s2.ID, "section-2")
	}
}

func TestService_SubmitRejectsInvalidImportance(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	input := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test Review",
		Chapters: []model.Chapter{
			{
				ID:    "ch-1",
				Title: "Test",
				Sections: []model.Section{
					{
						ID:        "section-1",
						Narrative: "Test",
						Hunks: []model.Hunk{
							{
								File:       "main.go",
								StartLine:  10,
								Diff:       "+test",
								Importance: "invalid", // invalid importance
							},
						},
					},
				},
			},
		},
	}

	_, err := svc.Submit(ctx, input)
	if !errors.Is(err, review.ErrInvalidHunkImportance) {
		t.Errorf("expected ErrInvalidHunkImportance, got %v", err)
	}
}

func TestService_SubmitRejectsMissingImportance(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	input := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test Review",
		Chapters: []model.Chapter{
			{
				ID:    "ch-1",
				Title: "Test",
				Sections: []model.Section{
					{
						ID:        "section-1",
						Narrative: "Test",
						Hunks: []model.Hunk{
							{
								File:       "main.go",
								StartLine:  10,
								Diff:       "+test",
								Importance: "", // missing importance
							},
						},
					},
				},
			},
		},
	}

	_, err := svc.Submit(ctx, input)
	if !errors.Is(err, review.ErrInvalidHunkImportance) {
		t.Errorf("expected ErrInvalidHunkImportance, got %v", err)
	}
}

func TestService_SubmitAcceptsValidImportance(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	validImportances := []string{"high", "medium", "low"}

	for _, imp := range validImportances {
		input := model.Review{
			WorkingDirectory: "/test/project",
			Title:            "Test Review",
			Chapters: []model.Chapter{
				{
					ID:    "ch-1",
					Title: "Test",
					Sections: []model.Section{
						{
							ID:        "section-1",
							Narrative: "Test",
							Hunks: []model.Hunk{
								{
									File:       "main.go",
									StartLine:  10,
									Diff:       "+test",
									Importance: imp,
								},
							},
						},
					},
				},
			},
		}

		_, err := svc.Submit(ctx, input)
		if err != nil {
			t.Errorf("Submit with importance=%q failed: %v", imp, err)
		}
	}
}

func TestService_SubmitSetsCreatedAtWhenNotProvided(t *testing.T) {
	svc, store := setupTestService(t)
	ctx := context.Background()

	before := time.Now()
	input := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test Review",
		// CreatedAt intentionally not set
	}

	_, err := svc.Submit(ctx, input)
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}
	after := time.Now()

	stored, err := store.Read("/test/project")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if stored.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
	if stored.CreatedAt.Before(before) || stored.CreatedAt.After(after) {
		t.Errorf("CreatedAt %v not within expected range [%v, %v]", stored.CreatedAt, before, after)
	}
}

func TestService_SubmitPreservesProvidedCreatedAt(t *testing.T) {
	svc, store := setupTestService(t)
	ctx := context.Background()

	providedTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	input := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test Review",
		CreatedAt:        providedTime,
	}

	_, err := svc.Submit(ctx, input)
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	stored, err := store.Read("/test/project")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if !stored.CreatedAt.Equal(providedTime) {
		t.Errorf("CreatedAt = %v, want %v", stored.CreatedAt, providedTime)
	}
}
