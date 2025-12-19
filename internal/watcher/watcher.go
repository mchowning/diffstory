package watcher

import (
	"encoding/json"
	"os"

	"github.com/fsnotify/fsnotify"
	"github.com/mchowning/diffguide/internal/model"
	"github.com/mchowning/diffguide/internal/storage"
)

// Watcher watches for review file changes for a specific directory
type Watcher struct {
	workDir    string
	reviewPath string
	reviewDir  string
	fsWatcher  *fsnotify.Watcher
	Reviews    chan model.Review
	Cleared    chan struct{}
	Errors     chan error
	done       chan struct{}
}

// New creates a watcher for the given working directory.
// Uses the default storage location (~/.diffguide/reviews).
func New(workDir string) (*Watcher, error) {
	store, err := storage.NewStore()
	if err != nil {
		return nil, err
	}
	return NewWithStore(workDir, store)
}

// NewWithStore creates a watcher using a custom store (for testing).
// This enables tests to use t.TempDir() for isolation.
func NewWithStore(workDir string, store *storage.Store) (*Watcher, error) {
	normalized, err := storage.NormalizePath(workDir)
	if err != nil {
		return nil, err
	}

	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	reviewPath, err := store.PathForDirectory(normalized)
	if err != nil {
		fsWatcher.Close()
		return nil, err
	}
	reviewDir := store.BaseDir()

	if err := os.MkdirAll(reviewDir, 0755); err != nil {
		fsWatcher.Close()
		return nil, err
	}
	if err := fsWatcher.Add(reviewDir); err != nil {
		fsWatcher.Close()
		return nil, err
	}

	return &Watcher{
		workDir:    normalized,
		reviewPath: reviewPath,
		reviewDir:  reviewDir,
		fsWatcher:  fsWatcher,
		Reviews:    make(chan model.Review, 1),
		Cleared:    make(chan struct{}, 1),
		Errors:     make(chan error, 1),
		done:       make(chan struct{}),
	}, nil
}

// Start begins watching for file changes.
// Loads any existing review file asynchronously to avoid blocking.
func (w *Watcher) Start() {
	go func() {
		if review, err := w.loadReview(); err == nil {
			select {
			case w.Reviews <- *review:
			case <-w.done:
			}
		}
	}()

	go w.watch()
}

func (w *Watcher) watch() {
	for {
		select {
		case <-w.done:
			return
		case event, ok := <-w.fsWatcher.Events:
			if !ok {
				return
			}
			if event.Name != w.reviewPath {
				continue
			}

			if event.Has(fsnotify.Remove) {
				select {
				case w.Cleared <- struct{}{}:
				case <-w.done:
					return
				}
				continue
			}

			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) || event.Has(fsnotify.Rename) {
				review, err := w.loadReview()
				if err != nil {
					if os.IsNotExist(err) {
						select {
						case w.Cleared <- struct{}{}:
						case <-w.done:
							return
						}
						continue
					}
					select {
					case w.Errors <- err:
					case <-w.done:
						return
					}
					continue
				}
				select {
				case w.Reviews <- *review:
				case <-w.done:
					return
				}
			}
		case err, ok := <-w.fsWatcher.Errors:
			if !ok {
				return
			}
			select {
			case w.Errors <- err:
			case <-w.done:
				return
			}
		}
	}
}

func (w *Watcher) loadReview() (*model.Review, error) {
	data, err := os.ReadFile(w.reviewPath)
	if err != nil {
		return nil, err
	}
	var review model.Review
	if err := json.Unmarshal(data, &review); err != nil {
		return nil, err
	}
	return &review, nil
}

// ReviewPath returns the path being watched (for testing)
func (w *Watcher) ReviewPath() string {
	return w.reviewPath
}

// Close stops the watcher
func (w *Watcher) Close() error {
	close(w.done)
	return w.fsWatcher.Close()
}
