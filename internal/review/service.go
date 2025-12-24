package review

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mchowning/diffguide/internal/model"
	"github.com/mchowning/diffguide/internal/storage"
)

var (
	ErrMissingWorkingDirectory = errors.New("workingDirectory is required")
	ErrInvalidWorkingDirectory = errors.New("invalid workingDirectory")
	ErrInvalidHunkImportance   = errors.New("invalid hunk importance")
)

type SubmitResult struct {
	FilePath string
}

type Service struct {
	store *storage.Store
}

func NewService(store *storage.Store) *Service {
	return &Service{store: store}
}

func (s *Service) Submit(ctx context.Context, review model.Review) (SubmitResult, error) {
	if review.WorkingDirectory == "" {
		return SubmitResult{}, ErrMissingWorkingDirectory
	}

	normalized, err := storage.NormalizePath(review.WorkingDirectory)
	if err != nil {
		return SubmitResult{}, fmt.Errorf("%w: %v", ErrInvalidWorkingDirectory, err)
	}
	review.WorkingDirectory = normalized

	if review.CreatedAt.IsZero() {
		review.CreatedAt = time.Now()
	}

	// Validate all hunks have valid importance
	for i, section := range review.Sections {
		for j, hunk := range section.Hunks {
			if !model.ValidImportance(hunk.Importance) {
				return SubmitResult{}, fmt.Errorf("%w: section[%d].hunks[%d] has importance %q",
					ErrInvalidHunkImportance, i, j, hunk.Importance)
			}
		}
	}

	if err := s.store.Write(review); err != nil {
		return SubmitResult{}, err
	}

	filePath, _ := s.store.PathForDirectory(normalized)
	return SubmitResult{FilePath: filePath}, nil
}
