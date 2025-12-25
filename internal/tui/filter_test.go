package tui

import (
	"testing"

	"github.com/mchowning/diffguide/internal/model"
)

func TestFilterLevel_String_ReturnsDisplayName(t *testing.T) {
	tests := []struct {
		level    FilterLevel
		expected string
	}{
		{FilterLevelLow, "Low (all)"},
		{FilterLevelMedium, "Medium"},
		{FilterLevelHigh, "High only"},
	}

	for _, tt := range tests {
		result := tt.level.String()
		if result != tt.expected {
			t.Errorf("FilterLevel(%d).String() = %q, want %q", tt.level, result, tt.expected)
		}
	}
}

func TestFilterLevel_Next_CyclesThroughLevels(t *testing.T) {
	// Low -> Medium -> High -> Low
	if FilterLevelLow.Next() != FilterLevelMedium {
		t.Error("Low.Next() should return Medium")
	}
	if FilterLevelMedium.Next() != FilterLevelHigh {
		t.Error("Medium.Next() should return High")
	}
	if FilterLevelHigh.Next() != FilterLevelLow {
		t.Error("High.Next() should return Low")
	}
}

func TestFilterLevel_PassesFilter_LowPassesAll(t *testing.T) {
	level := FilterLevelLow

	if !level.PassesFilter(model.ImportanceLow) {
		t.Error("Low filter should pass low importance")
	}
	if !level.PassesFilter(model.ImportanceMedium) {
		t.Error("Low filter should pass medium importance")
	}
	if !level.PassesFilter(model.ImportanceHigh) {
		t.Error("Low filter should pass high importance")
	}
}

func TestFilterLevel_PassesFilter_MediumPassesMediumAndHigh(t *testing.T) {
	level := FilterLevelMedium

	if level.PassesFilter(model.ImportanceLow) {
		t.Error("Medium filter should NOT pass low importance")
	}
	if !level.PassesFilter(model.ImportanceMedium) {
		t.Error("Medium filter should pass medium importance")
	}
	if !level.PassesFilter(model.ImportanceHigh) {
		t.Error("Medium filter should pass high importance")
	}
}

func TestFilterLevel_PassesFilter_HighPassesOnlyHigh(t *testing.T) {
	level := FilterLevelHigh

	if level.PassesFilter(model.ImportanceLow) {
		t.Error("High filter should NOT pass low importance")
	}
	if level.PassesFilter(model.ImportanceMedium) {
		t.Error("High filter should NOT pass medium importance")
	}
	if !level.PassesFilter(model.ImportanceHigh) {
		t.Error("High filter should pass high importance")
	}
}

func TestFilterLevel_PassesFilter_EmptyImportanceAlwaysPasses(t *testing.T) {
	// Empty importance should always pass for backward compatibility
	if !FilterLevelLow.PassesFilter("") {
		t.Error("Low filter should pass empty importance")
	}
	if !FilterLevelMedium.PassesFilter("") {
		t.Error("Medium filter should pass empty importance")
	}
	if !FilterLevelHigh.PassesFilter("") {
		t.Error("High filter should pass empty importance")
	}
}
