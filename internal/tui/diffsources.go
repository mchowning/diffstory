package tui

// DiffSource represents a source of diff content for review generation
type DiffSource struct {
	Label            string
	CommandHint      string   // Git command shown in parentheses (styled dimmer)
	Command          []string
	NeedsCommit      bool // true for "Specific commit"
	NeedsCommitRange bool // true for "Commit range"
}

// DefaultDiffSources returns the standard set of diff sources
func DefaultDiffSources() []DiffSource {
	return []DiffSource{
		{Label: "Uncommitted changes", CommandHint: "git diff HEAD", Command: []string{"git", "diff", "HEAD", "--no-color", "--no-ext-diff"}},
		{Label: "Staged changes", CommandHint: "git diff --cached", Command: []string{"git", "diff", "--cached", "--no-color", "--no-ext-diff"}},
		{Label: "Changes since main", CommandHint: "git diff main...HEAD", Command: []string{"git", "diff", "main...HEAD", "--no-color", "--no-ext-diff"}},
		{Label: "Specific commit...", NeedsCommit: true},
		{Label: "Commit range...", NeedsCommitRange: true},
	}
}

// CommitInfo holds metadata about a git commit
type CommitInfo struct {
	Hash    string
	Subject string
	Age     string
}
