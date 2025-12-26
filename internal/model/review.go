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
	ID       string    `json:"id" jsonschema_description:"Unique identifier for this chapter"`
	Title    string    `json:"title" jsonschema_description:"Short chapter title (~20-30 characters)"`
	Sections []Section `json:"sections" jsonschema_description:"Sections belonging to this chapter"`
}

type Section struct {
	ID        string `json:"id" jsonschema_description:"Unique identifier for this section"`
	Title     string `json:"title" jsonschema_description:"Short title for list display (~30-40 characters)"`
	Narrative string `json:"narrative" jsonschema_description:"Summary explaining what changed and why - should be understandable on its own AND connect smoothly to adjacent sections, building a coherent narrative arc"`
	Hunks     []Hunk `json:"hunks" jsonschema_description:"Code changes belonging to this section"`
}

type Hunk struct {
	File       string `json:"file" jsonschema_description:"File path relative to working directory"`
	StartLine  int    `json:"startLine" jsonschema_description:"Starting line number of the hunk"`
	Diff       string `json:"diff" jsonschema_description:"Complete unified diff content - include ALL lines, do not truncate or summarize"`
	Importance string `json:"importance" jsonschema:"enum=high,enum=medium,enum=low" jsonschema_description:"Importance level: high (critical changes), medium (significant changes), or low (minor changes)"`
	IsTest     *bool  `json:"isTest,omitempty" jsonschema_description:"True if this hunk contains test code changes, false for production code"`
}
