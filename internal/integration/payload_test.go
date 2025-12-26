package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mchowning/diffguide/internal/model"
	"github.com/mchowning/diffguide/internal/server"
	"github.com/mchowning/diffguide/internal/storage"
)

// testdataDir returns the path to the testdata directory
func testdataDir() string {
	return filepath.Join("..", "testdata")
}

// setupTestServer creates a server with an ephemeral port and temp storage
func setupTestServer(t *testing.T) (*server.Server, *storage.Store, func()) {
	t.Helper()
	baseDir := t.TempDir()
	store, err := storage.NewStoreWithDir(baseDir)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	srv, err := server.New(store, "0", false)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	go srv.Run()
	time.Sleep(10 * time.Millisecond) // give server time to start

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}

	return srv, store, cleanup
}

// loadFixture reads a JSON fixture file and returns its contents
func loadFixture(t *testing.T, filename string) []byte {
	t.Helper()
	path := filepath.Join(testdataDir(), filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", filename, err)
	}
	return data
}

// postReview sends a review payload to the server and returns the response
func postReview(t *testing.T, srv *server.Server, payload []byte) *http.Response {
	t.Helper()
	url := "http://127.0.0.1:" + srv.Port() + "/review"
	resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	return resp
}

// readStoredReview reads the stored review for a given working directory
func readStoredReview(t *testing.T, store *storage.Store, workingDir string) model.Review {
	t.Helper()
	review, err := store.Read(workingDir)
	if err != nil {
		t.Fatalf("failed to read stored review: %v", err)
	}
	return *review
}

func TestPayloadRoundTrip_SimpleReview(t *testing.T) {
	srv, store, cleanup := setupTestServer(t)
	defer cleanup()

	payload := loadFixture(t, "simple_review.json")

	resp := postReview(t, srv, payload)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// Parse original payload
	var original model.Review
	if err := json.Unmarshal(payload, &original); err != nil {
		t.Fatalf("failed to parse original: %v", err)
	}

	// Read stored review
	stored := readStoredReview(t, store, original.WorkingDirectory)

	// Verify all fields match (workingDirectory will be normalized)
	if stored.Title != original.Title {
		t.Errorf("Title mismatch: got %q, want %q", stored.Title, original.Title)
	}

	storedSections := stored.AllSections()
	originalSections := original.AllSections()
	if len(storedSections) != len(originalSections) {
		t.Fatalf("Sections count mismatch: got %d, want %d", len(storedSections), len(originalSections))
	}

	for i, section := range storedSections {
		origSection := originalSections[i]
		if section.ID != origSection.ID {
			t.Errorf("Section[%d].ID mismatch: got %q, want %q", i, section.ID, origSection.ID)
		}
		if section.Narrative != origSection.Narrative {
			t.Errorf("Section[%d].Narrative mismatch: got %q, want %q", i, section.Narrative, origSection.Narrative)
		}
		if len(section.Hunks) != len(origSection.Hunks) {
			t.Fatalf("Section[%d].Hunks count mismatch: got %d, want %d", i, len(section.Hunks), len(origSection.Hunks))
		}

		for j, hunk := range section.Hunks {
			origHunk := origSection.Hunks[j]
			if hunk.File != origHunk.File {
				t.Errorf("Section[%d].Hunk[%d].File mismatch: got %q, want %q", i, j, hunk.File, origHunk.File)
			}
			if hunk.StartLine != origHunk.StartLine {
				t.Errorf("Section[%d].Hunk[%d].StartLine mismatch: got %d, want %d", i, j, hunk.StartLine, origHunk.StartLine)
			}
			if hunk.Diff != origHunk.Diff {
				t.Errorf("Section[%d].Hunk[%d].Diff mismatch:\ngot:  %q\nwant: %q", i, j, hunk.Diff, origHunk.Diff)
			}
		}
	}
}

func TestPayloadRoundTrip_MultiSectionReview(t *testing.T) {
	srv, store, cleanup := setupTestServer(t)
	defer cleanup()

	payload := loadFixture(t, "multi_section_review.json")

	resp := postReview(t, srv, payload)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var original model.Review
	if err := json.Unmarshal(payload, &original); err != nil {
		t.Fatalf("failed to parse original: %v", err)
	}

	stored := readStoredReview(t, store, original.WorkingDirectory)

	storedSections := stored.AllSections()
	originalSections := original.AllSections()
	// Verify we have all 5 sections
	if len(storedSections) != 5 {
		t.Fatalf("expected 5 sections, got %d", len(storedSections))
	}

	// Count total hunks
	totalHunks := 0
	for _, section := range storedSections {
		totalHunks += len(section.Hunks)
	}

	originalHunks := 0
	for _, section := range originalSections {
		originalHunks += len(section.Hunks)
	}

	if totalHunks != originalHunks {
		t.Errorf("Total hunks mismatch: got %d, want %d", totalHunks, originalHunks)
	}

	// Verify all section IDs preserved
	for i, section := range storedSections {
		if section.ID != originalSections[i].ID {
			t.Errorf("Section[%d].ID mismatch: got %q, want %q", i, section.ID, originalSections[i].ID)
		}
	}
}

func TestPayloadRoundTrip_RealisticClaudeReview(t *testing.T) {
	srv, store, cleanup := setupTestServer(t)
	defer cleanup()

	payload := loadFixture(t, "realistic_claude_review.json")

	resp := postReview(t, srv, payload)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var original model.Review
	if err := json.Unmarshal(payload, &original); err != nil {
		t.Fatalf("failed to parse original: %v", err)
	}

	stored := readStoredReview(t, store, original.WorkingDirectory)

	// Verify title
	if stored.Title != original.Title {
		t.Errorf("Title mismatch: got %q, want %q", stored.Title, original.Title)
	}

	storedSections := stored.AllSections()
	originalSections := original.AllSections()
	// Verify all diff content preserved exactly
	for i, section := range storedSections {
		for j, hunk := range section.Hunks {
			origHunk := originalSections[i].Hunks[j]
			if hunk.Diff != origHunk.Diff {
				t.Errorf("Section[%d].Hunk[%d].Diff not preserved exactly", i, j)
			}
		}
	}
}

func TestPayloadRoundTrip_UnicodeContent(t *testing.T) {
	srv, store, cleanup := setupTestServer(t)
	defer cleanup()

	payload := loadFixture(t, "unicode_content.json")

	resp := postReview(t, srv, payload)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var original model.Review
	if err := json.Unmarshal(payload, &original); err != nil {
		t.Fatalf("failed to parse original: %v", err)
	}

	stored := readStoredReview(t, store, original.WorkingDirectory)

	// Verify title with emoji preserved
	if stored.Title != original.Title {
		t.Errorf("Title with emoji not preserved: got %q, want %q", stored.Title, original.Title)
	}

	storedSections := stored.AllSections()
	originalSections := original.AllSections()
	// Verify narrative with unicode preserved
	for i, section := range storedSections {
		if section.Narrative != originalSections[i].Narrative {
			t.Errorf("Section[%d].Narrative unicode not preserved:\ngot:  %q\nwant: %q",
				i, section.Narrative, originalSections[i].Narrative)
		}
	}

	// Verify diff content with unicode preserved
	for i, section := range storedSections {
		for j, hunk := range section.Hunks {
			origHunk := originalSections[i].Hunks[j]
			if hunk.Diff != origHunk.Diff {
				t.Errorf("Section[%d].Hunk[%d].Diff unicode not preserved", i, j)
			}
		}
	}
}

func TestPayloadRoundTrip_SpecialCharacters(t *testing.T) {
	srv, store, cleanup := setupTestServer(t)
	defer cleanup()

	payload := loadFixture(t, "special_characters.json")

	resp := postReview(t, srv, payload)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var original model.Review
	if err := json.Unmarshal(payload, &original); err != nil {
		t.Fatalf("failed to parse original: %v", err)
	}

	stored := readStoredReview(t, store, original.WorkingDirectory)

	storedSections := stored.AllSections()
	originalSections := original.AllSections()
	// Verify all diff content with special characters preserved exactly
	for i, section := range storedSections {
		for j, hunk := range section.Hunks {
			origHunk := originalSections[i].Hunks[j]
			if hunk.Diff != origHunk.Diff {
				t.Errorf("Section[%d].Hunk[%d].Diff special chars not preserved:\ngot:  %q\nwant: %q",
					i, j, hunk.Diff, origHunk.Diff)
			}
		}
	}
}

func TestPayloadRoundTrip_EmptyArrays(t *testing.T) {
	srv, store, cleanup := setupTestServer(t)
	defer cleanup()

	payload := loadFixture(t, "empty_arrays.json")

	resp := postReview(t, srv, payload)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var original model.Review
	if err := json.Unmarshal(payload, &original); err != nil {
		t.Fatalf("failed to parse original: %v", err)
	}

	stored := readStoredReview(t, store, original.WorkingDirectory)

	// Verify empty sections array is preserved
	if stored.SectionCount() != 0 {
		t.Errorf("expected 0 sections, got %d", stored.SectionCount())
	}

	if stored.Title != original.Title {
		t.Errorf("Title mismatch: got %q, want %q", stored.Title, original.Title)
	}
}

func TestPayloadRoundTrip_LargeReview(t *testing.T) {
	srv, store, cleanup := setupTestServer(t)
	defer cleanup()

	payload := loadFixture(t, "large_review.json")

	start := time.Now()
	resp := postReview(t, srv, payload)
	elapsed := time.Since(start)
	defer resp.Body.Close()

	// Verify 200 OK
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// NFR3: Response should be within 100ms
	if elapsed > 100*time.Millisecond {
		t.Errorf("POST took %v, should be under 100ms per NFR3", elapsed)
	}

	var original model.Review
	if err := json.Unmarshal(payload, &original); err != nil {
		t.Fatalf("failed to parse original: %v", err)
	}

	stored := readStoredReview(t, store, original.WorkingDirectory)

	storedSections := stored.AllSections()
	// Verify we have all 100 sections
	if len(storedSections) != 100 {
		t.Errorf("expected 100 sections, got %d", len(storedSections))
	}

	// Count total hunks
	totalHunks := 0
	for _, section := range storedSections {
		totalHunks += len(section.Hunks)
	}

	// Should have approximately 450-500 hunks
	if totalHunks < 400 || totalHunks > 550 {
		t.Errorf("expected ~450-500 hunks, got %d", totalHunks)
	}
}

func TestDiffContentIntegrity_HunkHeaders(t *testing.T) {
	srv, store, cleanup := setupTestServer(t)
	defer cleanup()

	// Create a review with hunk headers that must be preserved exactly
	review := model.Review{
		WorkingDirectory: "/test/hunk-headers",
		Title:            "Test hunk headers",
		Chapters: []model.Chapter{
			{
				ID:    "ch-1",
				Title: "Test",
				Sections: []model.Section{
					{
						ID:        "sec-1",
						Narrative: "Test",
						Hunks: []model.Hunk{
							{
								File:       "test.go",
								StartLine:  10,
								Diff:       "@@ -10,7 +10,9 @@ func TestFunc() {\n context line\n-removed\n+added\n context\n }",
								Importance: "medium",
							},
						},
					},
				},
			},
		},
	}

	payload, _ := json.Marshal(review)
	resp := postReview(t, srv, payload)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	stored := readStoredReview(t, store, review.WorkingDirectory)

	// Verify hunk header preserved exactly
	expectedDiff := "@@ -10,7 +10,9 @@ func TestFunc() {\n context line\n-removed\n+added\n context\n }"
	storedSections := stored.AllSections()
	if storedSections[0].Hunks[0].Diff != expectedDiff {
		t.Errorf("Hunk header not preserved:\ngot:  %q\nwant: %q",
			storedSections[0].Hunks[0].Diff, expectedDiff)
	}
}

func TestDiffContentIntegrity_ContextLines(t *testing.T) {
	srv, store, cleanup := setupTestServer(t)
	defer cleanup()

	// Create a review with context lines (space prefix) that must be preserved
	review := model.Review{
		WorkingDirectory: "/test/context-lines",
		Title:            "Test context lines",
		Chapters: []model.Chapter{
			{
				ID:    "ch-1",
				Title: "Test",
				Sections: []model.Section{
					{
						ID:        "sec-1",
						Narrative: "Test",
						Hunks: []model.Hunk{
							{
								File:       "test.go",
								StartLine:  1,
								Diff:       " func foo() {\n     indented context\n-    old line\n+    new line\n     more indented context\n }",
								Importance: "medium",
							},
						},
					},
				},
			},
		},
	}

	payload, _ := json.Marshal(review)
	resp := postReview(t, srv, payload)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	stored := readStoredReview(t, store, review.WorkingDirectory)

	storedSections := stored.AllSections()
	reviewSections := review.AllSections()
	// Verify context lines with leading spaces preserved exactly
	if storedSections[0].Hunks[0].Diff != reviewSections[0].Hunks[0].Diff {
		t.Errorf("Context lines not preserved:\ngot:  %q\nwant: %q",
			storedSections[0].Hunks[0].Diff, reviewSections[0].Hunks[0].Diff)
	}
}

func TestDiffContentIntegrity_MultilineNarrative(t *testing.T) {
	srv, store, cleanup := setupTestServer(t)
	defer cleanup()

	// Create a review with multi-line narrative
	multilineNarrative := `This is a multi-line narrative.

It has multiple paragraphs with blank lines between them.

    And some indented content too.

Final paragraph.`

	review := model.Review{
		WorkingDirectory: "/test/multiline-narrative",
		Title:            "Test multiline narrative",
		Chapters: []model.Chapter{
			{
				ID:    "ch-1",
				Title: "Test",
				Sections: []model.Section{
					{
						ID:        "sec-1",
						Narrative: multilineNarrative,
						Hunks: []model.Hunk{
							{
								File:       "test.go",
								StartLine:  1,
								Diff:       "+new line",
								Importance: "low",
							},
						},
					},
				},
			},
		},
	}

	payload, _ := json.Marshal(review)
	resp := postReview(t, srv, payload)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	stored := readStoredReview(t, store, review.WorkingDirectory)

	storedSections := stored.AllSections()
	// Verify multi-line narrative preserved exactly
	if storedSections[0].Narrative != multilineNarrative {
		t.Errorf("Multiline narrative not preserved:\ngot:  %q\nwant: %q",
			storedSections[0].Narrative, multilineNarrative)
	}
}
