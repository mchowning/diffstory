package tui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kaptinlin/jsonrepair"
	"github.com/mchowning/diffstory/internal/diff"
	"github.com/mchowning/diffstory/internal/model"
	"github.com/mchowning/diffstory/internal/storage"
)

const classificationPromptTemplate = `You are a code review assistant. Classify diff hunks into logical chapters and sections.

IMPORTANT: You MUST classify ALL hunks. Every hunk ID must appear exactly once in your response.

Read the input hunks from this JSON file: %s

The file contains a JSON array of hunk objects with fields: id, file, startLine, diff.

Respond with JSON in this exact format (no markdown fences, no explanation text):
{
  "title": "Brief title for this review",
  "chapters": [
    {
      "id": "chapter-identifier",
      "title": "Short chapter title",
      "sections": [
        {
          "id": "section-identifier",
          "title": "Short section title",
          "narrative": "Concise explanation of what and why. Mention key decisions if relevant.",
          "hunks": [
            {"id": "file/path.go::45", "importance": "high", "isTest": false},
            {"id": "file/path_test.go::120", "importance": "medium", "isTest": true}
          ]
        }
      ]
    }
  ]
}

## Grouping Philosophy

Group hunks by FUNCTIONAL PURPOSE, not by file path. Hunks that work together to achieve a goal belong in the same section, even if they span multiple files.

DO NOT:
- Group by file path. "Changes to auth.go" is never a good chapter title.
- Separate documentation into its own chapter. Docs belong with their related code.

## Format Requirements

- Each chapter contains one or more sections.
- Chapter title: ~20-30 characters, describes the theme (e.g., "Authentication", "Database Schema").
- Section title: ~30-40 characters, describes the specific change (e.g., "Add login endpoint handler").
- Each hunk must have importance: "high", "medium", or "low"
  - high: Critical changes (security, core logic, breaking changes)
  - medium: Important changes (new features, significant refactors)
  - low: Minor changes (formatting, comments, trivial fixes)
- Each hunk must have isTest: true if the hunk is test code, false if production code
  - Test code includes: unit tests, integration tests, test fixtures, test utilities, mocks
%s`

const retryPromptAddendum = `

CRITICAL: The previous response was incomplete. These hunk IDs were missing:
%s

You MUST include ALL hunk IDs in your response, including the ones listed above.`

// GenerateParams holds parameters for review generation
type GenerateParams struct {
	DiffCommand []string
	LLMCommand  []string // Resolved LLM command to use
	Context     string
	IsRetry     bool
	MissingIDs  []string
	ParsedHunks []diff.ParsedHunk // Set on retry to avoid re-parsing
}

// generateReviewCmd returns a command that runs the LLM generation with
// deterministic diff parsing and classification validation.
func generateReviewCmd(ctx context.Context, workDir string, store *storage.Store, logger *slog.Logger, params GenerateParams) tea.Cmd {
	return func() tea.Msg {
		var parsedHunks []diff.ParsedHunk

		// Use cached hunks on retry, otherwise parse fresh
		if params.IsRetry && len(params.ParsedHunks) > 0 {
			parsedHunks = params.ParsedHunks
			if logger != nil {
				logger.Info("using cached hunks for retry", "count", len(parsedHunks))
			}
		} else {
			// Step 1: Run diff command
			if logger != nil {
				logger.Info("running diff command", "command", params.DiffCommand)
			}
			diffOutput, err := runCommand(ctx, workDir, params.DiffCommand, nil)
			if err != nil {
				return GenerateErrorMsg{Err: fmt.Errorf("diff command failed: %w", err)}
			}

			if strings.TrimSpace(diffOutput) == "" {
				return GenerateErrorMsg{Err: fmt.Errorf("no changes found")}
			}

			// Step 2: Parse diff into hunks
			parsedHunks, err = diff.Parse(diffOutput)
			if err != nil {
				return GenerateErrorMsg{Err: fmt.Errorf("failed to parse diff: %w", err)}
			}
			if len(parsedHunks) == 0 {
				return GenerateErrorMsg{Err: fmt.Errorf("no hunks found in diff")}
			}
			if logger != nil {
				logger.Info("parsed diff into hunks", "count", len(parsedHunks))
			}
		}

		// Step 3: Write hunks to file in working directory for LLM to read
		hunksJSON, err := buildHunksJSON(parsedHunks)
		if err != nil {
			return GenerateErrorMsg{Err: fmt.Errorf("failed to build hunks JSON: %w", err)}
		}
		inputPath := filepath.Join(workDir, ".diffstory-input.json")
		if err := os.WriteFile(inputPath, []byte(hunksJSON), 0600); err != nil {
			return GenerateErrorMsg{Err: fmt.Errorf("failed to write input file: %w", err)}
		}
		defer os.Remove(inputPath)
		if logger != nil {
			logger.Info("wrote hunks to input file", "path", inputPath, "bytes", len(hunksJSON))
		}

		// Step 4: Build LLM prompt
		contextAddendum := ""
		if params.Context != "" {
			contextAddendum = fmt.Sprintf("\nUser context: %s", params.Context)
		}

		prompt := fmt.Sprintf(classificationPromptTemplate, inputPath, contextAddendum)

		// Add retry addendum if this is a retry
		if params.IsRetry && len(params.MissingIDs) > 0 {
			prompt += fmt.Sprintf(retryPromptAddendum, strings.Join(params.MissingIDs, ", "))
		}

		// Step 5: Call LLM
		if ctx.Err() != nil {
			return GenerateCancelledMsg{}
		}

		llmArgs := append([]string{}, params.LLMCommand[1:]...)
		llmArgs = append(llmArgs, prompt)
		llmCmd := append([]string{params.LLMCommand[0]}, llmArgs...)
		if logger != nil {
			logger.Info("calling LLM", "fullCommand", llmCmd, "prompt", prompt)
		}
		output, err := runCommand(ctx, workDir, llmCmd, nil)
		if err != nil {
			if ctx.Err() != nil {
				return GenerateCancelledMsg{}
			}
			return GenerateErrorMsg{Err: fmt.Errorf("LLM failed: %w", err)}
		}
		if logger != nil {
			logger.Info("LLM returned", "outputLength", len(output))
		}

		// Step 6: Parse LLM response
		response, err := extractLLMResponse(output, logger)
		if err != nil {
			if logger != nil {
				logger.Error("LLM response parse failed", "output", output, "error", err)
			}
			return GenerateErrorMsg{Err: fmt.Errorf("failed to parse LLM response: %w", err)}
		}

		// Step 7: Validate classification
		validation := validateClassification(parsedHunks, *response)
		if !validation.Valid {
			if params.IsRetry {
				// Second failure - return for user decision
				return GenerateValidationFailedMsg{
					Hunks:      parsedHunks,
					Missing:    validation.MissingIDs,
					Duplicates: validation.DuplicateIDs,
					Invalid:    validation.InvalidImportance,
					Response:   response,
				}
			}
			// First failure - auto retry
			return GenerateNeedsRetryMsg{
				Hunks:      parsedHunks,
				MissingIDs: validation.MissingIDs,
				Context:    params.Context,
			}
		}

		// Step 8: Assemble final review
		review := assembleReview(workDir, response, parsedHunks)

		// Step 9: Write to storage
		if err := store.Write(review); err != nil {
			return GenerateErrorMsg{Err: fmt.Errorf("failed to save review: %w", err)}
		}

		return GenerateSuccessMsg{}
	}
}

// buildHunksJSON creates a JSON representation of hunks for the LLM prompt
func buildHunksJSON(hunks []diff.ParsedHunk) (string, error) {
	var sb strings.Builder
	sb.WriteString("[\n")
	for i, h := range hunks {
		if i > 0 {
			sb.WriteString(",\n")
		}
		// Use JSON encoding for the diff content to handle special characters
		diffBytes, err := json.Marshal(h.Diff)
		if err != nil {
			return "", fmt.Errorf("failed to marshal diff for %s: %w", h.ID, err)
		}
		sb.WriteString(fmt.Sprintf(`  {"id": %q, "file": %q, "startLine": %d, "diff": %s}`,
			h.ID, h.File, h.StartLine, string(diffBytes)))
	}
	sb.WriteString("\n]")
	return sb.String(), nil
}

// extractLLMResponse parses the LLM output to find JSON response
func extractLLMResponse(output string, logger *slog.Logger) (*LLMResponse, error) {
	// Find the first '{' character (LLM may include preamble)
	start := strings.Index(output, "{")
	if start == -1 {
		return nil, fmt.Errorf("no JSON object found in response")
	}

	jsonStr := output[start:]

	// Try parsing as-is first
	decoder := json.NewDecoder(strings.NewReader(jsonStr))
	var response LLMResponse
	if err := decoder.Decode(&response); err != nil {
		// Attempt repair
		repaired, repairErr := jsonrepair.JSONRepair(jsonStr)
		if repairErr != nil {
			return nil, fmt.Errorf("failed to decode JSON: %w (repair also failed: %v)", err, repairErr)
		}

		// Try parsing repaired JSON
		decoder = json.NewDecoder(strings.NewReader(repaired))
		if err := decoder.Decode(&response); err != nil {
			return nil, fmt.Errorf("failed to decode repaired JSON: %w", err)
		}

		// Log that repair was needed
		if logger != nil {
			logger.Info("JSON repair applied to LLM output",
				"originalLength", len(jsonStr),
				"repairedLength", len(repaired))
		}
	}

	return &response, nil
}

// assembleReview combines LLM classification with parsed hunk data
func assembleReview(workDir string, response *LLMResponse, hunks []diff.ParsedHunk) model.Review {
	hunkMap := make(map[string]diff.ParsedHunk)
	for _, h := range hunks {
		hunkMap[h.ID] = h
	}

	review := model.Review{
		WorkingDirectory: workDir,
		Title:            response.Title,
		CreatedAt:        time.Now(),
	}

	// Build chapters from the response
	for _, ch := range response.Chapters {
		chapter := model.Chapter{
			ID:    ch.ID,
			Title: ch.Title,
		}
		for _, s := range ch.Sections {
			section := model.Section{
				ID:        s.ID,
				Title:     s.Title,
				Narrative: s.Narrative,
			}
			for _, href := range s.Hunks {
				if h, ok := hunkMap[href.ID]; ok {
					section.Hunks = append(section.Hunks, model.Hunk{
						File:       h.File,
						StartLine:  h.StartLine,
						Diff:       h.Diff,
						Importance: model.NormalizeImportance(href.Importance),
						IsTest:     href.IsTest,
					})
				}
			}
			chapter.Sections = append(chapter.Sections, section)
		}
		review.Chapters = append(review.Chapters, chapter)
	}

	return review
}

// assemblePartialReview creates a review with unclassified hunks in a separate chapter
func assemblePartialReview(workDir string, response *LLMResponse, hunks []diff.ParsedHunk, missingIDs []string) model.Review {
	review := assembleReview(workDir, response, hunks)

	// Create "Unclassified" chapter for missing hunks
	hunkMap := make(map[string]diff.ParsedHunk)
	for _, h := range hunks {
		hunkMap[h.ID] = h
	}

	var unclassifiedHunks []model.Hunk
	for _, id := range missingIDs {
		if h, ok := hunkMap[id]; ok {
			unclassifiedHunks = append(unclassifiedHunks, model.Hunk{
				File:       h.File,
				StartLine:  h.StartLine,
				Diff:       h.Diff,
				Importance: model.ImportanceMedium, // Default to medium
			})
		}
	}

	if len(unclassifiedHunks) > 0 {
		review.Chapters = append(review.Chapters, model.Chapter{
			ID:    "unclassified-chapter",
			Title: "Unclassified",
			Sections: []model.Section{
				{
					ID:        "unclassified",
					Title:     "Unclassified changes",
					Narrative: "The following changes could not be automatically classified.",
					Hunks:     unclassifiedHunks,
				},
			},
		})
	}

	return review
}

func getUntrackedFiles(ctx context.Context, workDir string) ([]string, error) {
	output, err := runCommand(ctx, workDir, []string{"git", "ls-files", "--others", "--exclude-standard"}, nil)
	if err != nil {
		return nil, err
	}

	output = strings.TrimSpace(output)
	if output == "" {
		return nil, nil
	}

	return strings.Split(output, "\n"), nil
}

func runCommand(ctx context.Context, workDir string, args []string, stdin []byte) (string, error) {
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = workDir
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%w: %s", err, stderr.String())
	}

	return stdout.String(), nil
}
