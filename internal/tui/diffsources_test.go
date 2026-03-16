package tui

import (
	"fmt"
	"strings"
	"testing"
)

func TestDefaultDiffSources_HasExpectedCount(t *testing.T) {
	sources := DefaultDiffSources("main")
	if len(sources) != 5 {
		t.Errorf("expected 5 diff sources, got %d", len(sources))
	}
}

func TestDefaultDiffSources_AllHaveLabels(t *testing.T) {
	sources := DefaultDiffSources("main")
	for i, source := range sources {
		if source.Label == "" {
			t.Errorf("source %d has empty label", i)
		}
	}
}

func TestDefaultDiffSources_CommandSourcesHaveCommands(t *testing.T) {
	sources := DefaultDiffSources("main")
	for i, source := range sources {
		if !source.NeedsCommit && !source.NeedsCommitRange && len(source.Command) == 0 {
			t.Errorf("source %d (%s) has no command but doesn't need commit selection", i, source.Label)
		}
	}
}

func TestDefaultDiffSources_CommandSourcesHaveCommandHint(t *testing.T) {
	sources := DefaultDiffSources("main")
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
	sources := DefaultDiffSources("main")
	for _, source := range sources {
		if strings.Contains(strings.ToLower(source.Label), "unstaged changes only") {
			t.Error("should not include 'Unstaged changes only' option")
		}
	}
}

func TestDefaultDiffSources_UncommittedChangesIsFirst(t *testing.T) {
	sources := DefaultDiffSources("main")
	if len(sources) == 0 {
		t.Fatal("expected at least one diff source")
	}
	if !strings.HasPrefix(sources[0].Label, "Uncommitted changes") {
		t.Errorf("first source should be 'Uncommitted changes', got %q", sources[0].Label)
	}
}

func TestDefaultDiffSources_UsesBaseBranchInLabel(t *testing.T) {
	sources := DefaultDiffSources("develop")
	found := false
	for _, source := range sources {
		if source.Label == "Changes since develop" {
			found = true
			if source.CommandHint != "git diff develop...HEAD" {
				t.Errorf("expected command hint 'git diff develop...HEAD', got %q", source.CommandHint)
			}
			expectedCmd := []string{"git", "diff", "develop...HEAD", "--no-color", "--no-ext-diff"}
			if len(source.Command) != len(expectedCmd) {
				t.Fatalf("expected command %v, got %v", expectedCmd, source.Command)
			}
			for i, v := range expectedCmd {
				if source.Command[i] != v {
					t.Errorf("command[%d] = %q, want %q", i, source.Command[i], v)
				}
			}
			break
		}
	}
	if !found {
		t.Error("expected a 'Changes since develop' diff source")
	}
}

func TestDefaultDiffSources_UsesMainBranchInLabel(t *testing.T) {
	sources := DefaultDiffSources("main")
	found := false
	for _, source := range sources {
		if source.Label == "Changes since main" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected a 'Changes since main' diff source")
	}
}

// mockGitRunner creates a gitRunner that returns predefined results based on args.
// Each entry maps a key like "symbolic-ref" or "rev-parse" to {output, error}.
type mockResult struct {
	output string
	err    error
}

func mockRunner(results map[string]mockResult) gitRunner {
	return func(workDir string, args ...string) (string, error) {
		// Build a key from args for lookup
		key := strings.Join(args, " ")
		if r, ok := results[key]; ok {
			return r.output, r.err
		}
		return "", fmt.Errorf("command not found: git %s", key)
	}
}

func TestDetectBaseBranch_UsesOriginHEAD(t *testing.T) {
	run := mockRunner(map[string]mockResult{
		"symbolic-ref refs/remotes/origin/HEAD": {output: "refs/remotes/origin/develop"},
	})

	branch := detectBaseBranchWith("", run)
	if branch != "develop" {
		t.Errorf("expected 'develop', got %q", branch)
	}
}

func TestDetectBaseBranch_FallsBackToLocalMain(t *testing.T) {
	run := mockRunner(map[string]mockResult{
		"symbolic-ref refs/remotes/origin/HEAD": {err: fmt.Errorf("not found")},
		"rev-parse --verify main":               {output: "abc123"},
	})

	branch := detectBaseBranchWith("", run)
	if branch != "main" {
		t.Errorf("expected 'main', got %q", branch)
	}
}

func TestDetectBaseBranch_FallsBackToLocalMaster(t *testing.T) {
	run := mockRunner(map[string]mockResult{
		"symbolic-ref refs/remotes/origin/HEAD": {err: fmt.Errorf("not found")},
		"rev-parse --verify main":               {err: fmt.Errorf("not found")},
		"rev-parse --verify master":             {output: "abc123"},
	})

	branch := detectBaseBranchWith("", run)
	if branch != "master" {
		t.Errorf("expected 'master', got %q", branch)
	}
}

func TestDetectBaseBranch_FallsBackToLocalDevelop(t *testing.T) {
	run := mockRunner(map[string]mockResult{
		"symbolic-ref refs/remotes/origin/HEAD": {err: fmt.Errorf("not found")},
		"rev-parse --verify main":               {err: fmt.Errorf("not found")},
		"rev-parse --verify master":             {err: fmt.Errorf("not found")},
		"rev-parse --verify develop":            {output: "abc123"},
	})

	branch := detectBaseBranchWith("", run)
	if branch != "develop" {
		t.Errorf("expected 'develop', got %q", branch)
	}
}

func TestDetectBaseBranch_PrefersOriginHEADOverLocalBranches(t *testing.T) {
	run := mockRunner(map[string]mockResult{
		"symbolic-ref refs/remotes/origin/HEAD": {output: "refs/remotes/origin/develop"},
		"rev-parse --verify main":               {output: "abc123"}, // exists but should not be used
	})

	branch := detectBaseBranchWith("", run)
	if branch != "develop" {
		t.Errorf("expected 'develop' (from origin/HEAD), got %q", branch)
	}
}

func TestDetectBaseBranch_FallsBackToMainWhenNothingFound(t *testing.T) {
	run := mockRunner(map[string]mockResult{
		"symbolic-ref refs/remotes/origin/HEAD": {err: fmt.Errorf("not found")},
		"rev-parse --verify main":               {err: fmt.Errorf("not found")},
		"rev-parse --verify master":             {err: fmt.Errorf("not found")},
		"rev-parse --verify develop":            {err: fmt.Errorf("not found")},
	})

	branch := detectBaseBranchWith("", run)
	if branch != "main" {
		t.Errorf("expected fallback 'main', got %q", branch)
	}
}
