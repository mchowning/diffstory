package storage_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mchowning/diffstory/internal/model"
	"github.com/mchowning/diffstory/internal/storage"
)

func TestNormalizePath_HandlesTrailingSlash(t *testing.T) {
	// Create a real directory to test with
	dir := t.TempDir()

	withSlash, err := storage.NormalizePath(dir + "/")
	if err != nil {
		t.Fatalf("NormalizePath failed: %v", err)
	}

	withoutSlash, err := storage.NormalizePath(dir)
	if err != nil {
		t.Fatalf("NormalizePath failed: %v", err)
	}

	if withSlash != withoutSlash {
		t.Errorf("trailing slash should not affect result: %q vs %q", withSlash, withoutSlash)
	}
}

func TestNormalizePath_HandlesRelativePaths(t *testing.T) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	result, err := storage.NormalizePath(".")
	if err != nil {
		t.Fatalf("NormalizePath failed: %v", err)
	}

	// Result should be absolute and match cwd (after resolving symlinks)
	expected, _ := filepath.EvalSymlinks(cwd)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestNormalizePath_ResolvesSymlinks(t *testing.T) {
	// Create a real directory and a symlink to it
	realDir := t.TempDir()
	symlinkDir := t.TempDir()
	symlink := filepath.Join(symlinkDir, "link")

	if err := os.Symlink(realDir, symlink); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	realResult, err := storage.NormalizePath(realDir)
	if err != nil {
		t.Fatalf("NormalizePath(realDir) failed: %v", err)
	}

	symlinkResult, err := storage.NormalizePath(symlink)
	if err != nil {
		t.Fatalf("NormalizePath(symlink) failed: %v", err)
	}

	if realResult != symlinkResult {
		t.Errorf("symlink should resolve to same path: real=%q, symlink=%q", realResult, symlinkResult)
	}
}

func TestHashDirectory_ConsistentForSameInput(t *testing.T) {
	hash1 := storage.HashDirectory("/test/project")
	hash2 := storage.HashDirectory("/test/project")

	if hash1 != hash2 {
		t.Errorf("same input should produce same hash: %q vs %q", hash1, hash2)
	}
}

func TestHashDirectory_DifferentForDifferentInputs(t *testing.T) {
	hash1 := storage.HashDirectory("/test/project-a")
	hash2 := storage.HashDirectory("/test/project-b")

	if hash1 == hash2 {
		t.Error("different inputs should produce different hashes")
	}
}

func TestStore_WriteCreatesFileAtExpectedPath(t *testing.T) {
	baseDir := t.TempDir()
	store, err := storage.NewStoreWithDir(baseDir)
	if err != nil {
		t.Fatalf("NewStoreWithDir failed: %v", err)
	}

	review := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test Review",
	}

	if err := store.Write(review); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Verify file exists at expected path
	expectedPath, _ := store.PathForDirectory("/test/project")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("expected file at %q to exist", expectedPath)
	}
}

func TestStore_WriteUsesAtomicWrite(t *testing.T) {
	baseDir := t.TempDir()
	store, err := storage.NewStoreWithDir(baseDir)
	if err != nil {
		t.Fatalf("NewStoreWithDir failed: %v", err)
	}

	review := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test Review",
	}

	if err := store.Write(review); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Verify no .tmp files remain
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".tmp" {
			t.Errorf("temp file should not remain: %s", entry.Name())
		}
	}
}

func TestStore_RoundTrip(t *testing.T) {
	baseDir := t.TempDir()
	store, err := storage.NewStoreWithDir(baseDir)
	if err != nil {
		t.Fatalf("NewStoreWithDir failed: %v", err)
	}

	review := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test Review",
		Chapters: []model.Chapter{
			{
				ID:    "ch-1",
				Title: "Changes",
				Sections: []model.Section{
					{
						ID:        "1",
						What: "Test narrative",
						Hunks: []model.Hunk{
							{File: "main.go", StartLine: 10, Diff: "+added line"},
						},
					},
				},
			},
		},
	}

	if err := store.Write(review); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	loaded, err := store.Read("/test/project")
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if loaded.Title != review.Title {
		t.Errorf("Title mismatch: got %q, want %q", loaded.Title, review.Title)
	}
	if loaded.WorkingDirectory != review.WorkingDirectory {
		t.Errorf("WorkingDirectory mismatch: got %q, want %q", loaded.WorkingDirectory, review.WorkingDirectory)
	}
	if loaded.SectionCount() != review.SectionCount() {
		t.Errorf("Sections count mismatch: got %d, want %d", loaded.SectionCount(), review.SectionCount())
	}
}

func TestStore_ReadNonExistentFileReturnsError(t *testing.T) {
	baseDir := t.TempDir()
	store, err := storage.NewStoreWithDir(baseDir)
	if err != nil {
		t.Fatalf("NewStoreWithDir failed: %v", err)
	}

	_, err = store.Read("/nonexistent/project")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestNewStore_UsesXDGCacheHomeWhenSet(t *testing.T) {
	tmpDir := t.TempDir()
	originalXDG := os.Getenv("XDG_CACHE_HOME")
	t.Cleanup(func() {
		os.Setenv("XDG_CACHE_HOME", originalXDG)
	})

	os.Setenv("XDG_CACHE_HOME", tmpDir)

	store, err := storage.NewStore()
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	expected := filepath.Join(tmpDir, "diffstory")
	if store.BaseDir() != expected {
		t.Errorf("expected BaseDir %q, got %q", expected, store.BaseDir())
	}
}

func TestNewStore_FallsBackToDotCacheWhenXDGUnset(t *testing.T) {
	originalXDG := os.Getenv("XDG_CACHE_HOME")
	t.Cleanup(func() {
		os.Setenv("XDG_CACHE_HOME", originalXDG)
	})

	os.Unsetenv("XDG_CACHE_HOME")

	store, err := storage.NewStore()
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir failed: %v", err)
	}

	expected := filepath.Join(home, ".cache", "diffstory")
	if store.BaseDir() != expected {
		t.Errorf("expected BaseDir %q, got %q", expected, store.BaseDir())
	}
}
