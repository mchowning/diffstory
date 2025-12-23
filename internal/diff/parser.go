package diff

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ParsedHunk represents a single hunk from a unified diff
type ParsedHunk struct {
	ID        string // format: "file/path.go::lineNumber"
	File      string
	StartLine int
	Diff      string // includes @@ header and content
}

var (
	diffGitRegex    = regexp.MustCompile(`^diff --git a/.+ b/(.+)$`)
	hunkHeaderRegex = regexp.MustCompile(`^@@ -\d+(?:,\d+)? \+(\d+)(?:,\d+)? @@`)
)

// Parse splits unified diff output into individual hunks
func Parse(diffOutput string) ([]ParsedHunk, error) {
	if diffOutput == "" {
		return nil, nil
	}

	var hunks []ParsedHunk
	fileDiffs := splitOnFileBoundaries(diffOutput)

	for _, fileDiff := range fileDiffs {
		if strings.Contains(fileDiff, "Binary files") {
			continue
		}

		filePath := extractFilePath(fileDiff)
		if filePath == "" {
			continue
		}

		fileHunks := splitIntoHunks(fileDiff, filePath)
		hunks = append(hunks, fileHunks...)
	}

	hunks = makeUniqueIDs(hunks)
	return hunks, nil
}

func splitOnFileBoundaries(diff string) []string {
	var files []string
	var current strings.Builder

	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "diff --git") && current.Len() > 0 {
			files = append(files, current.String())
			current.Reset()
		}
		current.WriteString(line)
		current.WriteString("\n")
	}
	if current.Len() > 0 {
		files = append(files, current.String())
	}

	return files
}

func extractFilePath(fileDiff string) string {
	for _, line := range strings.Split(fileDiff, "\n") {
		if matches := diffGitRegex.FindStringSubmatch(line); len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

func splitIntoHunks(fileDiff string, filePath string) []ParsedHunk {
	var hunks []ParsedHunk
	var currentHunk strings.Builder
	var currentStartLine int
	inHunk := false

	for _, line := range strings.Split(fileDiff, "\n") {
		if matches := hunkHeaderRegex.FindStringSubmatch(line); len(matches) > 1 {
			if inHunk && currentHunk.Len() > 0 {
				hunks = append(hunks, ParsedHunk{
					ID:        fmt.Sprintf("%s::%d", filePath, currentStartLine),
					File:      filePath,
					StartLine: currentStartLine,
					Diff:      strings.TrimSuffix(currentHunk.String(), "\n"),
				})
			}

			currentHunk.Reset()
			currentStartLine, _ = strconv.Atoi(matches[1])
			inHunk = true
		}

		if inHunk {
			currentHunk.WriteString(line)
			currentHunk.WriteString("\n")
		}
	}

	if inHunk && currentHunk.Len() > 0 {
		hunks = append(hunks, ParsedHunk{
			ID:        fmt.Sprintf("%s::%d", filePath, currentStartLine),
			File:      filePath,
			StartLine: currentStartLine,
			Diff:      strings.TrimSuffix(currentHunk.String(), "\n"),
		})
	}

	return hunks
}

func makeUniqueIDs(hunks []ParsedHunk) []ParsedHunk {
	seen := make(map[string]int)
	for i := range hunks {
		id := hunks[i].ID
		if count, exists := seen[id]; exists {
			hunks[i].ID = fmt.Sprintf("%s#%d", id, count+1)
		}
		seen[id]++
	}
	return hunks
}
