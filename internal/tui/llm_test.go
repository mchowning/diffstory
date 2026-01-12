package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/mchowning/diffstory/internal/config"
)

func TestResolveLLMCommand_ExplicitCommandFound(t *testing.T) {
	cfg := &config.Config{LLMCommand: []string{"my-llm", "--flag"}}
	lookPath := func(cmd string) (string, error) {
		if cmd == "my-llm" {
			return "/usr/bin/my-llm", nil
		}
		return "", errors.New("not found")
	}

	result := ResolveLLMCommand(cfg, lookPath)

	if result.Error != "" {
		t.Errorf("expected no error, got %q", result.Error)
	}
	if len(result.Command) != 2 || result.Command[0] != "my-llm" || result.Command[1] != "--flag" {
		t.Errorf("expected [my-llm --flag], got %v", result.Command)
	}
}

func TestResolveLLMCommand_ExplicitCommandNotFound(t *testing.T) {
	cfg := &config.Config{LLMCommand: []string{"nonexistent-llm", "--flag"}}
	lookPath := func(cmd string) (string, error) {
		return "", errors.New("not found")
	}

	result := ResolveLLMCommand(cfg, lookPath)

	if result.Command != nil {
		t.Errorf("expected no command, got %v", result.Command)
	}
	if !strings.Contains(result.Error, "nonexistent-llm") {
		t.Errorf("expected error to mention command name, got %q", result.Error)
	}
	if !strings.Contains(result.Error, "not found") {
		t.Errorf("expected error to mention 'not found', got %q", result.Error)
	}
}

func TestResolveLLMCommand_NoConfigClaudeFound(t *testing.T) {
	lookPath := func(cmd string) (string, error) {
		if cmd == "claude" {
			return "/usr/bin/claude", nil
		}
		return "", errors.New("not found")
	}

	result := ResolveLLMCommand(nil, lookPath)

	if result.Error != "" {
		t.Errorf("expected no error, got %q", result.Error)
	}
	if len(result.Command) != 2 || result.Command[0] != "claude" || result.Command[1] != "-p" {
		t.Errorf("expected [claude -p], got %v", result.Command)
	}
}

func TestResolveLLMCommand_NoConfigClaudeNotFound(t *testing.T) {
	lookPath := func(cmd string) (string, error) {
		return "", errors.New("not found")
	}

	result := ResolveLLMCommand(nil, lookPath)

	if result.Command != nil {
		t.Errorf("expected no command, got %v", result.Command)
	}
	if !strings.Contains(result.Error, "Claude Code not found") {
		t.Errorf("expected error to mention Claude Code, got %q", result.Error)
	}
	if !strings.Contains(result.Error, "config") {
		t.Errorf("expected error to mention config, got %q", result.Error)
	}
}

func TestResolveLLMCommand_EmptyLLMCommandClaudeFound(t *testing.T) {
	cfg := &config.Config{LLMCommand: []string{}}
	lookPath := func(cmd string) (string, error) {
		if cmd == "claude" {
			return "/usr/bin/claude", nil
		}
		return "", errors.New("not found")
	}

	result := ResolveLLMCommand(cfg, lookPath)

	if result.Error != "" {
		t.Errorf("expected no error, got %q", result.Error)
	}
	if len(result.Command) != 2 || result.Command[0] != "claude" || result.Command[1] != "-p" {
		t.Errorf("expected [claude -p], got %v", result.Command)
	}
}

func TestResolveLLMCommand_EmptyLLMCommandClaudeNotFound(t *testing.T) {
	cfg := &config.Config{LLMCommand: []string{}}
	lookPath := func(cmd string) (string, error) {
		return "", errors.New("not found")
	}

	result := ResolveLLMCommand(cfg, lookPath)

	if result.Command != nil {
		t.Errorf("expected no command, got %v", result.Command)
	}
	if !strings.Contains(result.Error, "Claude Code not found") {
		t.Errorf("expected error to mention Claude Code, got %q", result.Error)
	}
}
