package tui

import (
	"github.com/mchowning/diffguide/internal/diff"
	"github.com/mchowning/diffguide/internal/model"
)

// LLMResponse is the classification returned by the LLM
type LLMResponse struct {
	Title    string       `json:"title"`
	Sections []LLMSection `json:"sections"`
}

// LLMSection represents a classified section from the LLM
type LLMSection struct {
	ID        string       `json:"id"`
	Narrative string       `json:"narrative"`
	Hunks     []LLMHunkRef `json:"hunks"`
}

// LLMHunkRef references a hunk by ID with its classified importance
type LLMHunkRef struct {
	ID         string `json:"id"`
	Importance string `json:"importance"`
}

// ValidationResult holds the results of classification validation
type ValidationResult struct {
	Valid             bool
	MissingIDs        []string
	DuplicateIDs      []string
	InvalidImportance []string
}

// validateClassification checks that all input hunks are classified exactly once
// with valid importance values
func validateClassification(inputHunks []diff.ParsedHunk, response LLMResponse) ValidationResult {
	result := ValidationResult{Valid: true}

	// Build set of input IDs
	inputIDs := make(map[string]bool)
	for _, h := range inputHunks {
		inputIDs[h.ID] = true
	}

	// Track output IDs and their counts
	outputIDs := make(map[string]int)
	for _, section := range response.Sections {
		for _, hunk := range section.Hunks {
			outputIDs[hunk.ID]++

			// Check importance validity (using normalization)
			normalized := model.NormalizeImportance(hunk.Importance)
			if normalized == "" {
				result.InvalidImportance = append(result.InvalidImportance, hunk.ID)
				result.Valid = false
			}
		}
	}

	// Find missing IDs (in input but not in output)
	for id := range inputIDs {
		if outputIDs[id] == 0 {
			result.MissingIDs = append(result.MissingIDs, id)
			result.Valid = false
		}
	}

	// Find duplicates (appearing more than once in output)
	for id, count := range outputIDs {
		if count > 1 {
			result.DuplicateIDs = append(result.DuplicateIDs, id)
			result.Valid = false
		}
	}

	return result
}
