package tui

import "github.com/mchowning/diffguide/internal/model"

// FilterLevel represents the importance filter threshold
type FilterLevel int

const (
	FilterLevelLow    FilterLevel = iota // Show all hunks
	FilterLevelMedium                    // Show medium + high importance
	FilterLevelHigh                      // Show only high importance
)

func (f FilterLevel) String() string {
	switch f {
	case FilterLevelLow:
		return "Low (all)"
	case FilterLevelMedium:
		return "Medium"
	case FilterLevelHigh:
		return "High only"
	default:
		return "Unknown"
	}
}

func (f FilterLevel) Next() FilterLevel {
	switch f {
	case FilterLevelLow:
		return FilterLevelMedium
	case FilterLevelMedium:
		return FilterLevelHigh
	default:
		return FilterLevelLow
	}
}

func (f FilterLevel) PassesFilter(importance string) bool {
	// Empty importance always passes (backward compatibility)
	if importance == "" {
		return true
	}
	switch f {
	case FilterLevelHigh:
		return importance == model.ImportanceHigh
	case FilterLevelMedium:
		return importance == model.ImportanceHigh || importance == model.ImportanceMedium
	default: // FilterLevelLow
		return true
	}
}
