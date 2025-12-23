package tui

import (
	"strings"
	"testing"
)

func TestDefaultDiffSources_HasExpectedCount(t *testing.T) {
	sources := DefaultDiffSources()
	if len(sources) != 5 {
		t.Errorf("expected 5 diff sources, got %d", len(sources))
	}
}

func TestDefaultDiffSources_AllHaveLabels(t *testing.T) {
	sources := DefaultDiffSources()
	for i, source := range sources {
		if source.Label == "" {
			t.Errorf("source %d has empty label", i)
		}
	}
}

func TestDefaultDiffSources_CommandSourcesHaveCommands(t *testing.T) {
	sources := DefaultDiffSources()
	for i, source := range sources {
		if !source.NeedsCommit && !source.NeedsCommitRange && len(source.Command) == 0 {
			t.Errorf("source %d (%s) has no command but doesn't need commit selection", i, source.Label)
		}
	}
}

func TestDefaultDiffSources_CommandSourcesHaveCommandHint(t *testing.T) {
	sources := DefaultDiffSources()
	for i, source := range sources {
		if source.NeedsCommit || source.NeedsCommitRange {
			continue // These don't have commands to show
		}
		if !strings.HasPrefix(source.CommandHint, "git ") {
			t.Errorf("source %d (%s) should have CommandHint starting with 'git '", i, source.Label)
		}
	}
}

func TestDefaultDiffSources_DoesNotIncludeUnstagedOnly(t *testing.T) {
	sources := DefaultDiffSources()
	for _, source := range sources {
		if strings.Contains(strings.ToLower(source.Label), "unstaged changes only") {
			t.Error("should not include 'Unstaged changes only' option")
		}
	}
}

func TestDefaultDiffSources_UncommittedChangesIsFirst(t *testing.T) {
	sources := DefaultDiffSources()
	if len(sources) == 0 {
		t.Fatal("expected at least one diff source")
	}
	if !strings.HasPrefix(sources[0].Label, "Uncommitted changes") {
		t.Errorf("first source should be 'Uncommitted changes', got %q", sources[0].Label)
	}
}
