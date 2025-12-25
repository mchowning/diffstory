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
	Sections         []Section `json:"sections"`
	CreatedAt        time.Time `json:"createdAt,omitempty"`
}

type Section struct {
	ID        string `json:"id" jsonschema_description:"Unique identifier for this section"`
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
