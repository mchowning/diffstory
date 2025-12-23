package tui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mchowning/diffguide/internal/config"
	"github.com/mchowning/diffguide/internal/model"
	"github.com/mchowning/diffguide/internal/storage"
)

const promptTemplate = `Summarize the following diff, explaining what changed and why.

Guidelines:
- Structure the summary into logical sections, grouping related changes together
- Each section's narrative should explain what changed and why
- The narratives should work independently but also form a cohesive story when read in sequence
- This is a summary, not a code review. Explain the changes; do not critique them.

Return ONLY a JSON object (no markdown fences, no explanation text) matching this schema:
{
  "title": "<brief summary title>",
  "sections": [
    {
      "id": "<unique section id>",
      "narrative": "<what changed and why>",
      "importance": "<high|medium|low>",
      "hunks": [
        {
          "file": "<relative file path>",
          "startLine": <line number>,
          "diff": "<complete unified diff content>"
        }
      ]
    }
  ]
}

Diff:
`

// generateReviewCmd returns a command that runs the LLM generation.
// On success, it writes the review to disk via the store; the watcher
// will then deliver it to the TUI.
func generateReviewCmd(ctx context.Context, cfg *config.Config, workDir string, store *storage.Store, logger *slog.Logger) tea.Cmd {
	return func() tea.Msg {
		// Run diff command
		diffOutput, err := runCommand(ctx, workDir, cfg.DiffCommand, nil)
		if err != nil {
			return GenerateErrorMsg{Err: fmt.Errorf("diff command failed: %w", err)}
		}

		if strings.TrimSpace(diffOutput) == "" {
			return GenerateErrorMsg{Err: fmt.Errorf("no changes to review")}
		}

		// Build prompt with diff
		prompt := promptTemplate + diffOutput

		// Run LLM command with prompt as final argument
		llmArgs := append([]string{}, cfg.LLMCommand[1:]...)
		llmArgs = append(llmArgs, prompt)

		output, err := runCommand(ctx, workDir, append([]string{cfg.LLMCommand[0]}, llmArgs...), nil)
		if err != nil {
			if ctx.Err() == context.Canceled {
				return GenerateCancelledMsg{}
			}
			return GenerateErrorMsg{Err: fmt.Errorf("LLM command failed: %w", err)}
		}

		// Extract and parse JSON
		review, err := extractReviewJSON(output)
		if err != nil {
			if logger != nil {
				logger.Error("failed to extract JSON from LLM response",
					"error", err,
					"llm_output", output,
				)
			}
			return GenerateErrorMsg{Err: err}
		}

		// Set working directory
		review.WorkingDirectory = workDir

		// Write to disk - watcher will pick up the change and deliver to TUI
		if err := store.Write(*review); err != nil {
			return GenerateErrorMsg{Err: fmt.Errorf("failed to save review: %w", err)}
		}

		return GenerateSuccessMsg{}
	}
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

func extractReviewJSON(output string) (*model.Review, error) {
	// Find first { - json.Decoder will handle parsing from there
	start := strings.Index(output, "{")
	if start == -1 {
		return nil, fmt.Errorf("no JSON object found in response")
	}

	// Use json.Decoder to parse exactly one complete JSON object
	// This correctly handles braces inside string values
	decoder := json.NewDecoder(strings.NewReader(output[start:]))
	var review model.Review
	if err := decoder.Decode(&review); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return &review, nil
}
