package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mchowning/diffstory/internal/model"
)

func loadReviewFromFile(path string) (*model.Review, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var review model.Review
	if err := json.Unmarshal(data, &review); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	for _, chapter := range review.Chapters {
		for _, section := range chapter.Sections {
			for _, hunk := range section.Hunks {
				if hunk.Importance != "" && !model.ValidImportance(hunk.Importance) {
					return nil, fmt.Errorf("invalid importance %q in file %s", hunk.Importance, hunk.File)
				}
			}
		}
	}

	return &review, nil
}
