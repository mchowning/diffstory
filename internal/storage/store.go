package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mchowning/diffstory/internal/model"
)

// Store handles persisting reviews to disk
type Store struct {
	baseDir string
}

// NewStore creates a store with the default base directory (~/.diffstory/reviews)
func NewStore() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	baseDir := filepath.Join(home, ".diffstory", "reviews")
	return NewStoreWithDir(baseDir)
}

// NewStoreWithDir creates a store with a custom base directory (for testing)
func NewStoreWithDir(baseDir string) (*Store, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}
	return &Store{baseDir: baseDir}, nil
}

// NormalizePath returns a canonical absolute path for consistent hashing.
// Applies filepath.Abs, filepath.Clean, and filepath.EvalSymlinks to handle
// trailing slashes, relative paths, symlinks, and case variations on
// case-insensitive filesystems (macOS, Windows).
func NormalizePath(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}
	cleaned := filepath.Clean(abs)

	// EvalSymlinks resolves symlinks AND canonicalizes case on macOS/Windows.
	// This ensures "/users/foo" and "/Users/Foo" produce the same hash.
	resolved, err := filepath.EvalSymlinks(cleaned)
	if err != nil {
		// If path doesn't exist yet, fall back to cleaned path
		// (server may receive paths before directories are created)
		if os.IsNotExist(err) {
			return cleaned, nil
		}
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}
	return resolved, nil
}

// HashDirectory returns the SHA256 hash of a directory path
func HashDirectory(dir string) string {
	hash := sha256.Sum256([]byte(dir))
	return hex.EncodeToString(hash[:])
}

// PathForDirectory returns the file path for a given working directory.
// The directory path is normalized before hashing.
func (s *Store) PathForDirectory(dir string) (string, error) {
	normalized, err := NormalizePath(dir)
	if err != nil {
		return "", err
	}
	return filepath.Join(s.baseDir, HashDirectory(normalized)+".json"), nil
}

// BaseDir returns the base directory for review files (for watcher setup)
func (s *Store) BaseDir() string {
	return s.baseDir
}

// Write persists a review to disk using atomic write (temp file + rename)
// to prevent partial reads by file watchers.
func (s *Store) Write(review model.Review) error {
	path, err := s.PathForDirectory(review.WorkingDirectory)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(review, "", "  ")
	if err != nil {
		return err
	}

	// Atomic write: write to temp file, then rename
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath) // Clean up temp file on rename failure
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// Read loads a review from disk for a given directory
func (s *Store) Read(dir string) (*model.Review, error) {
	path, err := s.PathForDirectory(dir)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var review model.Review
	if err := json.Unmarshal(data, &review); err != nil {
		return nil, err
	}
	return &review, nil
}
