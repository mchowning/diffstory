package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	LLMCommand          []string `json:"llmCommand"`
	DiffCommand         []string `json:"diffCommand"`
	DebugLoggingEnabled bool     `json:"debugLoggingEnabled"`
	DefaultFilterLevel  string   `json:"defaultFilterLevel"`
}

// Load reads config from XDG_CONFIG_HOME or ~/.config
func Load() (*Config, error) {
	for _, dir := range configDirs() {
		jsonPath := filepath.Join(dir, "config.json")
		jsoncPath := filepath.Join(dir, "config.jsonc")

		jsonExists := fileExists(jsonPath)
		jsoncExists := fileExists(jsoncPath)

		if jsonExists && jsoncExists {
			return nil, fmt.Errorf("both config.json and config.jsonc exist in %s; please use only one", dir)
		}

		var path string
		if jsonExists {
			path = jsonPath
		} else if jsoncExists {
			path = jsoncPath
		} else {
			continue
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		var cfg Config
		cleanedData := stripComments(data)
		if err := json.Unmarshal(cleanedData, &cfg); err != nil {
			return nil, fmt.Errorf("invalid config at %s: %w", path, err)
		}

		// Apply defaults
		if len(cfg.DiffCommand) == 0 {
			cfg.DiffCommand = []string{"git", "diff", "HEAD"}
		}
		if cfg.DefaultFilterLevel == "" {
			cfg.DefaultFilterLevel = "medium"
		}

		return &cfg, nil
	}

	return nil, nil // No config file found (not an error)
}

func configDirs() []string {
	var dirs []string

	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		dirs = append(dirs, filepath.Join(xdg, "diffstory"))
	}

	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, ".config", "diffstory"))
	}

	return dirs
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func stripComments(data []byte) []byte {
	var result strings.Builder
	inString := false
	i := 0
	for i < len(data) {
		if data[i] == '"' && (i == 0 || data[i-1] != '\\') {
			inString = !inString
			result.WriteByte(data[i])
			i++
		} else if !inString && i+1 < len(data) && data[i] == '/' && data[i+1] == '/' {
			// Single-line comment: skip until end of line
			for i < len(data) && data[i] != '\n' {
				i++
			}
		} else if !inString && i+1 < len(data) && data[i] == '/' && data[i+1] == '*' {
			// Block comment: skip until */
			i += 2
			for i+1 < len(data) && !(data[i] == '*' && data[i+1] == '/') {
				i++
			}
			if i+1 < len(data) {
				i += 2 // Skip the closing */
			}
		} else {
			result.WriteByte(data[i])
			i++
		}
	}
	return stripTrailingCommas([]byte(result.String()))
}

func stripTrailingCommas(data []byte) []byte {
	var result strings.Builder
	inString := false
	i := 0
	for i < len(data) {
		if data[i] == '"' && (i == 0 || data[i-1] != '\\') {
			inString = !inString
			result.WriteByte(data[i])
			i++
		} else if !inString && data[i] == ',' {
			// Look ahead for ] or } (skipping whitespace)
			j := i + 1
			for j < len(data) && (data[j] == ' ' || data[j] == '\t' || data[j] == '\n' || data[j] == '\r') {
				j++
			}
			if j < len(data) && (data[j] == ']' || data[j] == '}') {
				// Skip the trailing comma
				i++
			} else {
				result.WriteByte(data[i])
				i++
			}
		} else {
			result.WriteByte(data[i])
			i++
		}
	}
	return []byte(result.String())
}
