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

type TestFilter int

const (
	TestFilterAll       TestFilter = iota // Show all hunks
	TestFilterExcluding                   // Exclude test hunks
	TestFilterOnly                        // Show only test hunks
)

func (f TestFilter) String() string {
	switch f {
	case TestFilterAll:
		return ""
	case TestFilterExcluding:
		return "Excluding Tests"
	case TestFilterOnly:
		return "Only Tests"
	default:
		return ""
	}
}

func (f TestFilter) Next() TestFilter {
	switch f {
	case TestFilterAll:
		return TestFilterExcluding
	case TestFilterExcluding:
		return TestFilterOnly
	default:
		return TestFilterAll
	}
}

func (f TestFilter) PassesFilter(isTest *bool) bool {
	// nil always passes (backward compatibility)
	if isTest == nil {
		return true
	}
	switch f {
	case TestFilterExcluding:
		return !*isTest
	case TestFilterOnly:
		return *isTest
	default: // TestFilterAll
		return true
	}
}
