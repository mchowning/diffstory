package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	LLMCommand          []string `json:"llmCommand"`
	DiffCommand         []string `json:"diffCommand"`
	DebugLoggingEnabled bool     `json:"debugLoggingEnabled"`
}

// Load reads config from XDG_CONFIG_HOME or ~/.config
func Load() (*Config, error) {
	paths := configPaths()

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}

		var cfg Config
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("invalid config at %s: %w", path, err)
		}

		// Apply defaults
		if len(cfg.DiffCommand) == 0 {
			cfg.DiffCommand = []string{"git", "diff", "HEAD"}
		}

		return &cfg, nil
	}

	return nil, nil // No config file found (not an error)
}

func configPaths() []string {
	var paths []string

	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		paths = append(paths, filepath.Join(xdg, "diffguide", "config.json"))
	}

	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, filepath.Join(home, ".config", "diffguide", "config.json"))
	}

	return paths
}
