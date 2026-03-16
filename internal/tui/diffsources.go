package tui

import (
	"fmt"
	"os/exec"
	"strings"
)

// DiffSource represents a source of diff content for review generation
type DiffSource struct {
	Label            string
	CommandHint      string   // Git command shown in parentheses (styled dimmer)
	Command          []string
	NeedsCommit      bool // true for "Specific commit"
	NeedsCommitRange bool // true for "Commit range"
}

// DefaultDiffSources returns the standard set of diff sources.
// baseBranch is the branch name to use for the "Changes since ..." option.
func DefaultDiffSources(baseBranch string) []DiffSource {
	diffRef := baseBranch + "...HEAD"
	return []DiffSource{
		{Label: "Uncommitted changes", CommandHint: "git diff HEAD", Command: []string{"git", "diff", "HEAD", "--no-color", "--no-ext-diff"}},
		{Label: "Staged changes", CommandHint: "git diff --cached", Command: []string{"git", "diff", "--cached", "--no-color", "--no-ext-diff"}},
		{Label: fmt.Sprintf("Changes since %s", baseBranch), CommandHint: fmt.Sprintf("git diff %s", diffRef), Command: []string{"git", "diff", diffRef, "--no-color", "--no-ext-diff"}},
		{Label: "Specific commit...", NeedsCommit: true},
		{Label: "Commit range...", NeedsCommitRange: true},
	}
}

// gitRunner executes a git command in a directory and returns stdout.
type gitRunner func(workDir string, args ...string) (string, error)

func defaultGitRunner(workDir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = workDir
	output, err := cmd.Output()
	return strings.TrimSpace(string(output)), err
}

// DetectBaseBranch attempts to detect the repository's default branch.
// It tries the following strategies in order:
//  1. git symbolic-ref refs/remotes/origin/HEAD (set by clone or fetch)
//  2. Check if common branch names exist locally: main, master, develop
//  3. Fall back to "main"
func DetectBaseBranch(workDir string) string {
	return detectBaseBranchWith(workDir, defaultGitRunner)
}

func detectBaseBranchWith(workDir string, run gitRunner) string {
	// Strategy 1: Check the remote HEAD reference
	if output, err := run(workDir, "symbolic-ref", "refs/remotes/origin/HEAD"); err == nil {
		// output looks like "refs/remotes/origin/main"
		if parts := strings.Split(output, "/"); len(parts) > 0 {
			branch := parts[len(parts)-1]
			if branch != "" {
				return branch
			}
		}
	}

	// Strategy 2: Check for common branch names locally
	for _, candidate := range []string{"main", "master", "develop"} {
		if _, err := run(workDir, "rev-parse", "--verify", candidate); err == nil {
			return candidate
		}
	}

	// Strategy 3: Default fallback
	return "main"
}

// CommitInfo holds metadata about a git commit
type CommitInfo struct {
	Hash    string
	Subject string
	Age     string
}
