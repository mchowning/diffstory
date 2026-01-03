package logging

import (
	"io"
	"log/slog"
	"os"
)

const LogPath = "/tmp/diffstory.log"

func Setup(enabled bool) *slog.Logger {
	if !enabled {
		return slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	file, err := os.OpenFile(LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// Fall back to discarding if we can't open the log file
		return slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	return slog.New(slog.NewTextHandler(file, nil))
}
