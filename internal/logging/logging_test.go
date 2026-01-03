package logging

import (
	"os"
	"strings"
	"testing"
)

func TestSetup_EnabledWritesToLogFile(t *testing.T) {
	// Clean up any existing log file
	logPath := "/tmp/diffstory.log"
	os.Remove(logPath)

	logger := Setup(true)
	logger.Info("test message")

	// Read the log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "test message") {
		t.Errorf("expected log file to contain 'test message', got: %s", content)
	}

	// Clean up
	os.Remove(logPath)
}

func TestSetup_DisabledDoesNotWriteLogFile(t *testing.T) {
	logPath := "/tmp/diffstory.log"
	os.Remove(logPath)

	logger := Setup(false)
	logger.Info("test message")

	// Log file should not exist
	if _, err := os.Stat(logPath); !os.IsNotExist(err) {
		t.Error("expected log file to not exist when logging is disabled")
		os.Remove(logPath)
	}
}
