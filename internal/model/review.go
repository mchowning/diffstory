package model

import (
	"strings"
	"time"
)

const (
	ImportanceHigh   = "high"
	ImportanceMedium = "medium"
	ImportanceLow    = "low"
)

func ValidImportance(s string) bool {
	switch s {
	case ImportanceHigh, ImportanceMedium, ImportanceLow:
		return true
	default:
		return false
	}
}

func NormalizeImportance(s string) string {
	switch strings.ToLower(s) {
	case "high", "critical", "important":
		return ImportanceHigh
	case "medium", "moderate", "normal":
		return ImportanceMedium
	case "low", "minor", "trivial":
		return ImportanceLow
	default:
		return ""
	}
}

type Review struct {
	WorkingDirectory string    `json:"workingDirectory"`
	Title            string    `json:"title"`
	Chapters         []Chapter `json:"chapters"`
	CreatedAt        time.Time `json:"createdAt,omitempty"`
}

// AllSections returns a flattened list of all sections across all chapters.
func (r Review) AllSections() []Section {
	var sections []Section
	for _, ch := range r.Chapters {
		sections = append(sections, ch.Sections...)
	}
	return sections
}

// SectionCount returns the total number of sections across all chapters.
func (r Review) SectionCount() int {
	count := 0
	for _, ch := range r.Chapters {
		count += len(ch.Sections)
	}
	return count
}

// NewReviewWithSections creates a Review with a single default chapter containing the given sections.
// This is a convenience function primarily for testing and migration purposes.
func NewReviewWithSections(workDir, title string, sections []Section) Review {
	return Review{
		WorkingDirectory: workDir,
		Title:            title,
		Chapters: []Chapter{
			{
				ID:       "default",
				Title:    "Changes",
				Sections: sections,
			},
		},
	}
}

type Chapter struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	Sections []Section `json:"sections"`
}

type Section struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Narrative string `json:"narrative"`
	Hunks     []Hunk `json:"hunks"`
}

type Hunk struct {
	File       string `json:"file"`
	StartLine  int    `json:"startLine"`
	Diff       string `json:"diff"`
	Importance string `json:"importance"`
	IsTest     *bool  `json:"isTest,omitempty"`
}
