package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_FromXDGConfigHome(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "diffguide")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(configDir, "config.json")
	configContent := `{"llmCommand": ["claude", "-p"]}`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected config, got nil")
	}
	if len(cfg.LLMCommand) != 2 || cfg.LLMCommand[0] != "claude" || cfg.LLMCommand[1] != "-p" {
		t.Errorf("unexpected LLMCommand: %v", cfg.LLMCommand)
	}
}

func TestLoad_FallbackToHomeConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".config", "diffguide")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(configDir, "config.json")
	configContent := `{"llmCommand": ["llm", "prompt"]}`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Clear XDG_CONFIG_HOME so it falls back to ~/.config
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected config, got nil")
	}
	if len(cfg.LLMCommand) != 2 || cfg.LLMCommand[0] != "llm" {
		t.Errorf("unexpected LLMCommand: %v", cfg.LLMCommand)
	}
}

func TestLoad_MissingFileReturnsNil(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	t.Setenv("HOME", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg != nil {
		t.Errorf("expected nil config for missing file, got: %+v", cfg)
	}
}

func TestLoad_InvalidJSONReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "diffguide")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, []byte("not valid json"), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if cfg != nil {
		t.Errorf("expected nil config on error, got: %+v", cfg)
	}
}

func TestLoad_AppliesDefaultDiffCommand(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "diffguide")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(configDir, "config.json")
	// Config with llmCommand but no diffCommand
	configContent := `{"llmCommand": ["claude", "-p"]}`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedDiff := []string{"git", "diff", "HEAD"}
	if len(cfg.DiffCommand) != 3 {
		t.Fatalf("expected default diffCommand, got: %v", cfg.DiffCommand)
	}
	for i, v := range expectedDiff {
		if cfg.DiffCommand[i] != v {
			t.Errorf("diffCommand[%d] = %q, want %q", i, cfg.DiffCommand[i], v)
		}
	}
}

func TestLoad_PreservesCustomDiffCommand(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "diffguide")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(configDir, "config.json")
	configContent := `{"llmCommand": ["claude"], "diffCommand": ["git", "diff", "--cached"]}`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.DiffCommand) != 3 || cfg.DiffCommand[2] != "--cached" {
		t.Errorf("expected custom diffCommand preserved, got: %v", cfg.DiffCommand)
	}
}

func TestLoad_ParsesDebugLoggingEnabled(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "diffguide")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(configDir, "config.json")
	configContent := `{"debugLoggingEnabled": true}`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.DebugLoggingEnabled {
		t.Error("expected DebugLoggingEnabled to be true")
	}
}
