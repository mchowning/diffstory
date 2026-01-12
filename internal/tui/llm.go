package tui

import (
	"fmt"
	"os/exec"

	"github.com/mchowning/diffstory/internal/config"
)

// DefaultLookPath uses exec.LookPath to find commands on PATH
func DefaultLookPath(file string) (string, error) {
	return exec.LookPath(file)
}

// LookPathFunc is a function type for looking up commands on PATH
type LookPathFunc func(file string) (string, error)

// LLMCommandResult represents the result of resolving the LLM command
type LLMCommandResult struct {
	Command []string
	Error   string
}

// ResolveLLMCommand determines which LLM command to use.
// If config has an explicit llmCommand, it validates that command exists.
// If no llmCommand is configured, it checks for claude on PATH.
func ResolveLLMCommand(cfg *config.Config, lookPath LookPathFunc) LLMCommandResult {
	// Case 1: Explicit llmCommand configured
	if cfg != nil && len(cfg.LLMCommand) > 0 {
		cmd := cfg.LLMCommand[0]
		if _, err := lookPath(cmd); err != nil {
			return LLMCommandResult{
				Error: fmt.Sprintf("LLM command not found: %q\n\nCheck that the command is installed and on your PATH.", cmd),
			}
		}
		return LLMCommandResult{Command: cfg.LLMCommand}
	}

	// Case 2: No llmCommand configured - try claude as default
	if _, err := lookPath("claude"); err != nil {
		configPath := "~/.config/diffstory/config.jsonc"
		return LLMCommandResult{
			Error: fmt.Sprintf("Claude Code not found on PATH.\n\nInstall Claude Code, or configure a different LLM in %s:\n\n  {\n    \"llmCommand\": [\"your-llm-command\", \"args\"]\n  }", configPath),
		}
	}

	return LLMCommandResult{Command: []string{"claude", "-p"}}
}
