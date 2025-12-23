package watcher_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mchowning/diffguide/internal/model"
	"github.com/mchowning/diffguide/internal/storage"
	"github.com/mchowning/diffguide/internal/watcher"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestWatcher_NewCreatesWatcherForDirectory(t *testing.T) {
	dir := t.TempDir()
	store := createTestStore(t, dir)
	workDir := filepath.Join(dir, "project")
	os.MkdirAll(workDir, 0755)

	w, err := watcher.NewWithStore(workDir, store, discardLogger())
	if err != nil {
		t.Fatalf("NewWithStore failed: %v", err)
	}
	defer w.Close()

	if w == nil {
		t.Fatal("expected watcher, got nil")
	}
}

func TestWatcher_NewWithStoreAcceptsCustomStore(t *testing.T) {
	dir := t.TempDir()
	store := createTestStore(t, dir)
	workDir := filepath.Join(dir, "project")
	os.MkdirAll(workDir, 0755)

	w, err := watcher.NewWithStore(workDir, store, discardLogger())
	if err != nil {
		t.Fatalf("NewWithStore failed: %v", err)
	}
	defer w.Close()

	// Verify it uses the custom store by checking the review path
	expectedPath, _ := store.PathForDirectory(workDir)
	if w.ReviewPath() != expectedPath {
		t.Errorf("ReviewPath() = %q, want %q", w.ReviewPath(), expectedPath)
	}
}

func TestWatcher_NormalizesWorkingDirectoryPath(t *testing.T) {
	dir := t.TempDir()
	store := createTestStore(t, dir)
	workDir := filepath.Join(dir, "project")
	os.MkdirAll(workDir, 0755)

	// Create watcher with trailing slash - should normalize
	w, err := watcher.NewWithStore(workDir+"/", store, discardLogger())
	if err != nil {
		t.Fatalf("NewWithStore failed: %v", err)
	}
	defer w.Close()

	// Create watcher without trailing slash
	w2, err := watcher.NewWithStore(workDir, store, discardLogger())
	if err != nil {
		t.Fatalf("NewWithStore failed: %v", err)
	}
	defer w2.Close()

	// Both should watch the same file
	if w.ReviewPath() != w2.ReviewPath() {
		t.Errorf("paths differ: %q vs %q", w.ReviewPath(), w2.ReviewPath())
	}
}

func TestWatcher_WatchesCorrectFileBasedOnDirectoryHash(t *testing.T) {
	dir := t.TempDir()
	store := createTestStore(t, dir)
	workDir := filepath.Join(dir, "project")
	os.MkdirAll(workDir, 0755)

	w, err := watcher.NewWithStore(workDir, store, discardLogger())
	if err != nil {
		t.Fatalf("NewWithStore failed: %v", err)
	}
	defer w.Close()

	expectedPath, _ := store.PathForDirectory(workDir)
	if w.ReviewPath() != expectedPath {
		t.Errorf("ReviewPath() = %q, want %q", w.ReviewPath(), expectedPath)
	}
}

func TestWatcher_SendsReviewWhenFileCreated(t *testing.T) {
	dir := t.TempDir()
	store := createTestStore(t, dir)
	workDir := filepath.Join(dir, "project")
	os.MkdirAll(workDir, 0755)

	w, err := watcher.NewWithStore(workDir, store, discardLogger())
	if err != nil {
		t.Fatalf("NewWithStore failed: %v", err)
	}
	defer w.Close()

	w.Start()

	// Write a review file
	review := model.Review{
		WorkingDirectory: workDir,
		Title:            "Test Review",
		Sections: []model.Section{
			{ID: "1", Narrative: "Test narrative"},
		},
	}
	if err := store.Write(review); err != nil {
		t.Fatalf("store.Write failed: %v", err)
	}

	// Wait for the review to be received
	select {
	case received := <-w.Reviews:
		if received.Title != review.Title {
			t.Errorf("Title = %q, want %q", received.Title, review.Title)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for review")
	}
}

func TestWatcher_SendsReviewWhenFileModified(t *testing.T) {
	dir := t.TempDir()
	store := createTestStore(t, dir)
	workDir := filepath.Join(dir, "project")
	os.MkdirAll(workDir, 0755)

	// Create initial review file before starting watcher
	initialReview := model.Review{
		WorkingDirectory: workDir,
		Title:            "Initial Review",
	}
	if err := store.Write(initialReview); err != nil {
		t.Fatalf("store.Write failed: %v", err)
	}

	w, err := watcher.NewWithStore(workDir, store, discardLogger())
	if err != nil {
		t.Fatalf("NewWithStore failed: %v", err)
	}
	defer w.Close()

	w.Start()

	// Drain the initial review load
	select {
	case <-w.Reviews:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for initial review load")
	}

	// Modify the review file
	modifiedReview := model.Review{
		WorkingDirectory: workDir,
		Title:            "Modified Review",
	}
	if err := store.Write(modifiedReview); err != nil {
		t.Fatalf("store.Write failed: %v", err)
	}

	// Wait for the modified review
	select {
	case received := <-w.Reviews:
		if received.Title != "Modified Review" {
			t.Errorf("Title = %q, want %q", received.Title, "Modified Review")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for modified review")
	}
}

func TestWatcher_SendsReviewOnRenameEvent(t *testing.T) {
	// Atomic writes use rename, so this tests that path
	dir := t.TempDir()
	store := createTestStore(t, dir)
	workDir := filepath.Join(dir, "project")
	os.MkdirAll(workDir, 0755)

	w, err := watcher.NewWithStore(workDir, store, discardLogger())
	if err != nil {
		t.Fatalf("NewWithStore failed: %v", err)
	}
	defer w.Close()

	w.Start()

	// Storage.Write uses atomic rename internally
	review := model.Review{
		WorkingDirectory: workDir,
		Title:            "Atomic Write Review",
	}
	if err := store.Write(review); err != nil {
		t.Fatalf("store.Write failed: %v", err)
	}

	select {
	case received := <-w.Reviews:
		if received.Title != review.Title {
			t.Errorf("Title = %q, want %q", received.Title, review.Title)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for review from atomic write")
	}
}

func TestWatcher_SendsClearedWhenFileDeleted(t *testing.T) {
	dir := t.TempDir()
	store := createTestStore(t, dir)
	workDir := filepath.Join(dir, "project")
	os.MkdirAll(workDir, 0755)

	// Create initial review file
	review := model.Review{
		WorkingDirectory: workDir,
		Title:            "To Be Deleted",
	}
	if err := store.Write(review); err != nil {
		t.Fatalf("store.Write failed: %v", err)
	}

	w, err := watcher.NewWithStore(workDir, store, discardLogger())
	if err != nil {
		t.Fatalf("NewWithStore failed: %v", err)
	}
	defer w.Close()

	w.Start()

	// Drain initial load
	select {
	case <-w.Reviews:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for initial review")
	}

	// Delete the review file
	reviewPath := w.ReviewPath()
	if err := os.Remove(reviewPath); err != nil {
		t.Fatalf("failed to remove review file: %v", err)
	}

	// Wait for cleared signal
	select {
	case <-w.Cleared:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for cleared signal")
	}
}

func TestWatcher_IgnoresOtherFilesInReviewsDirectory(t *testing.T) {
	dir := t.TempDir()
	store := createTestStore(t, dir)
	workDir := filepath.Join(dir, "project")
	os.MkdirAll(workDir, 0755)

	w, err := watcher.NewWithStore(workDir, store, discardLogger())
	if err != nil {
		t.Fatalf("NewWithStore failed: %v", err)
	}
	defer w.Close()

	w.Start()

	// Write a different file to the reviews directory
	otherFile := filepath.Join(store.BaseDir(), "other-file.json")
	if err := os.WriteFile(otherFile, []byte(`{"title": "other"}`), 0644); err != nil {
		t.Fatalf("failed to write other file: %v", err)
	}

	// Give fsnotify time to process
	time.Sleep(100 * time.Millisecond)

	// Should not receive anything on Reviews channel
	select {
	case review := <-w.Reviews:
		t.Errorf("unexpected review received: %+v", review)
	default:
		// Success - no review received
	}
}

func TestWatcher_ChannelsAreBuffered(t *testing.T) {
	dir := t.TempDir()
	store := createTestStore(t, dir)
	workDir := filepath.Join(dir, "project")
	os.MkdirAll(workDir, 0755)

	// Create existing review file before watcher starts
	review := model.Review{
		WorkingDirectory: workDir,
		Title:            "Existing Review",
	}
	if err := store.Write(review); err != nil {
		t.Fatalf("store.Write failed: %v", err)
	}

	w, err := watcher.NewWithStore(workDir, store, discardLogger())
	if err != nil {
		t.Fatalf("NewWithStore failed: %v", err)
	}
	defer w.Close()

	// Start without reading from channels - should not deadlock
	w.Start()

	// Give time for initial load to happen
	time.Sleep(100 * time.Millisecond)

	// Now read - should have the review buffered
	select {
	case received := <-w.Reviews:
		if received.Title != review.Title {
			t.Errorf("Title = %q, want %q", received.Title, review.Title)
		}
	case <-time.After(time.Second):
		t.Fatal("expected buffered review, got nothing")
	}
}

func TestWatcher_LoadsExistingReviewFileOnStart(t *testing.T) {
	dir := t.TempDir()
	store := createTestStore(t, dir)
	workDir := filepath.Join(dir, "project")
	os.MkdirAll(workDir, 0755)

	// Create review file before watcher starts
	review := model.Review{
		WorkingDirectory: workDir,
		Title:            "Pre-existing Review",
		Sections: []model.Section{
			{ID: "1", Narrative: "Pre-existing section"},
		},
	}
	if err := store.Write(review); err != nil {
		t.Fatalf("store.Write failed: %v", err)
	}

	w, err := watcher.NewWithStore(workDir, store, discardLogger())
	if err != nil {
		t.Fatalf("NewWithStore failed: %v", err)
	}
	defer w.Close()

	w.Start()

	// Should receive the existing review
	select {
	case received := <-w.Reviews:
		if received.Title != review.Title {
			t.Errorf("Title = %q, want %q", received.Title, review.Title)
		}
		if len(received.Sections) != 1 {
			t.Errorf("len(Sections) = %d, want 1", len(received.Sections))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for existing review")
	}
}

func TestWatcher_CloseStopsWatching(t *testing.T) {
	dir := t.TempDir()
	store := createTestStore(t, dir)
	workDir := filepath.Join(dir, "project")
	os.MkdirAll(workDir, 0755)

	w, err := watcher.NewWithStore(workDir, store, discardLogger())
	if err != nil {
		t.Fatalf("NewWithStore failed: %v", err)
	}

	w.Start()

	// Close the watcher
	if err := w.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Write a review after close
	review := model.Review{
		WorkingDirectory: workDir,
		Title:            "After Close",
	}
	if err := store.Write(review); err != nil {
		t.Fatalf("store.Write failed: %v", err)
	}

	// Should not receive anything
	select {
	case <-w.Reviews:
		t.Error("received review after close")
	case <-time.After(200 * time.Millisecond):
		// Success - no review received after close
	}
}

// Helper to create a test store with isolated temp directory
func createTestStore(t *testing.T, baseDir string) *storage.Store {
	t.Helper()
	reviewsDir := filepath.Join(baseDir, "reviews")
	store, err := storage.NewStoreWithDir(reviewsDir)
	if err != nil {
		t.Fatalf("NewStoreWithDir failed: %v", err)
	}
	return store
}

// writeReviewDirectly writes a review JSON file directly (not through storage)
// to test watcher without coupling to storage implementation
func writeReviewDirectly(t *testing.T, path string, review model.Review) {
	t.Helper()
	data, err := json.MarshalIndent(review, "", "  ")
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
}
