package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_FromXDGConfigHome(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "diffstory")
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
	configDir := filepath.Join(tmpDir, ".config", "diffstory")
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
	configDir := filepath.Join(tmpDir, "diffstory")
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
	configDir := filepath.Join(tmpDir, "diffstory")
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
	configDir := filepath.Join(tmpDir, "diffstory")
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
	configDir := filepath.Join(tmpDir, "diffstory")
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

func TestLoad_SupportsJSONCWithSingleLineComments(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "diffstory")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(configDir, "config.json")
	configContent := `{
	// This is a comment explaining the LLM command
	"llmCommand": ["claude", "-p"],
	// Enable debug logging for troubleshooting
	"debugLoggingEnabled": true
}`
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
	if len(cfg.LLMCommand) != 2 || cfg.LLMCommand[0] != "claude" {
		t.Errorf("unexpected LLMCommand: %v", cfg.LLMCommand)
	}
	if !cfg.DebugLoggingEnabled {
		t.Error("expected DebugLoggingEnabled to be true")
	}
}

func TestLoad_SupportsJSONCWithInlineComments(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "diffstory")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(configDir, "config.json")
	configContent := `{
	"llmCommand": ["claude", "-p"], // The command to invoke LLM
	"debugLoggingEnabled": true // Enable for troubleshooting
}`
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
	if len(cfg.LLMCommand) != 2 || cfg.LLMCommand[0] != "claude" {
		t.Errorf("unexpected LLMCommand: %v", cfg.LLMCommand)
	}
	if !cfg.DebugLoggingEnabled {
		t.Error("expected DebugLoggingEnabled to be true")
	}
}

func TestLoad_PreservesDoubleSlashInsideStrings(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "diffstory")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(configDir, "config.json")
	// URL contains // which should NOT be treated as a comment
	configContent := `{
	"llmCommand": ["curl", "https://api.example.com/v1"]
}`
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
	if len(cfg.LLMCommand) != 2 {
		t.Fatalf("expected 2 elements in LLMCommand, got: %v", cfg.LLMCommand)
	}
	if cfg.LLMCommand[1] != "https://api.example.com/v1" {
		t.Errorf("URL was corrupted: got %q", cfg.LLMCommand[1])
	}
}

func TestLoad_SupportsJSONCWithBlockComments(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "diffstory")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(configDir, "config.json")
	configContent := `{
	/* This is a block comment
	   that spans multiple lines */
	"llmCommand": ["claude", "-p"],
	"debugLoggingEnabled": /* inline block comment */ true
}`
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
	if len(cfg.LLMCommand) != 2 || cfg.LLMCommand[0] != "claude" {
		t.Errorf("unexpected LLMCommand: %v", cfg.LLMCommand)
	}
	if !cfg.DebugLoggingEnabled {
		t.Error("expected DebugLoggingEnabled to be true")
	}
}

func TestLoad_SupportsJSONCWithTrailingCommas(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "diffstory")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(configDir, "config.json")
	configContent := `{
	"llmCommand": ["claude", "-p",],
	"debugLoggingEnabled": true,
}`
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
	if len(cfg.LLMCommand) != 2 || cfg.LLMCommand[0] != "claude" {
		t.Errorf("unexpected LLMCommand: %v", cfg.LLMCommand)
	}
	if !cfg.DebugLoggingEnabled {
		t.Error("expected DebugLoggingEnabled to be true")
	}
}

func TestLoad_LoadsFromJSONCExtension(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "diffstory")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Use .jsonc extension instead of .json
	configPath := filepath.Join(configDir, "config.jsonc")
	configContent := `{
	// This file uses .jsonc extension
	"llmCommand": ["claude", "-p"],
	"debugLoggingEnabled": true
}`
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
	if len(cfg.LLMCommand) != 2 || cfg.LLMCommand[0] != "claude" {
		t.Errorf("unexpected LLMCommand: %v", cfg.LLMCommand)
	}
	if !cfg.DebugLoggingEnabled {
		t.Error("expected DebugLoggingEnabled to be true")
	}
}

func TestLoad_ErrorsWhenBothJSONAndJSONCExist(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "diffstory")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create both .json and .jsonc files
	jsonPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(jsonPath, []byte(`{"llmCommand": ["from-json"]}`), 0644); err != nil {
		t.Fatal(err)
	}

	jsoncPath := filepath.Join(configDir, "config.jsonc")
	if err := os.WriteFile(jsoncPath, []byte(`{"llmCommand": ["from-jsonc"]}`), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg, err := Load()
	if err == nil {
		t.Fatal("expected error when both config.json and config.jsonc exist")
	}
	if cfg != nil {
		t.Errorf("expected nil config on error, got: %+v", cfg)
	}
}

func TestLoad_DefaultFilterLevelDefaultsToMedium(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "diffstory")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	configPath := filepath.Join(configDir, "config.json")
	// Config without defaultFilterLevel
	configContent := `{"llmCommand": ["claude"]}`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.DefaultFilterLevel != "medium" {
		t.Errorf("expected defaultFilterLevel to be 'medium', got %q", cfg.DefaultFilterLevel)
	}
}

func TestLoad_ParsesDefaultFilterLevel(t *testing.T) {
	tests := []struct {
		configValue string
		expected    string
	}{
		{"low", "low"},
		{"medium", "medium"},
		{"high", "high"},
	}

	for _, tt := range tests {
		t.Run(tt.configValue, func(t *testing.T) {
			tmpDir := t.TempDir()
			configDir := filepath.Join(tmpDir, "diffstory")
			if err := os.MkdirAll(configDir, 0755); err != nil {
				t.Fatal(err)
			}

			configPath := filepath.Join(configDir, "config.json")
			configContent := `{"defaultFilterLevel": "` + tt.configValue + `"}`
			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				t.Fatal(err)
			}

			t.Setenv("XDG_CONFIG_HOME", tmpDir)

			cfg, err := Load()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cfg.DefaultFilterLevel != tt.expected {
				t.Errorf("expected defaultFilterLevel %q, got %q", tt.expected, cfg.DefaultFilterLevel)
			}
		})
	}
}
