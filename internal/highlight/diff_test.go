package highlight_test

import (
	"os"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/mchowning/diffguide/internal/highlight"
	"github.com/muesli/termenv"
)

func TestMain(m *testing.M) {
	// Force color output in tests (lipgloss doesn't render colors without TTY)
	lipgloss.SetColorProfile(termenv.TrueColor)
	os.Exit(m.Run())
}

func TestColorizeDiffLine_AdditionLine(t *testing.T) {
	line := "+func NewModel() Model {"
	result := highlight.ColorizeDiffLine(line)

	if !strings.Contains(result, "\x1b[") {
		t.Error("expected ANSI escape sequences in output")
	}
	if !strings.Contains(result, line) {
		t.Errorf("expected result to contain original line %q", line)
	}
}

func TestColorizeDiffLine_DeletionLine(t *testing.T) {
	line := "-func OldModel() Model {"
	result := highlight.ColorizeDiffLine(line)

	if !strings.Contains(result, "\x1b[") {
		t.Error("expected ANSI escape sequences in output")
	}
	if !strings.Contains(result, line) {
		t.Errorf("expected result to contain original line %q", line)
	}
}

func TestColorizeDiffLine_HunkHeader(t *testing.T) {
	line := "@@ -10,7 +10,9 @@ func processFile(path string) error {"
	result := highlight.ColorizeDiffLine(line)

	if !strings.Contains(result, "\x1b[") {
		t.Error("expected ANSI escape sequences in output")
	}
	if !strings.Contains(result, line) {
		t.Errorf("expected result to contain original line %q", line)
	}
}

func TestColorizeDiffLine_ContextLine(t *testing.T) {
	line := " 	return nil"
	result := highlight.ColorizeDiffLine(line)

	if !strings.Contains(result, "\x1b[") {
		t.Error("expected ANSI escape sequences in output")
	}
	if !strings.Contains(result, "return nil") {
		t.Error("expected result to contain original content")
	}
}

func TestColorizeDiffLine_EmptyLine(t *testing.T) {
	result := highlight.ColorizeDiffLine("")

	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestColorizeDiffLine_OtherPrefix(t *testing.T) {
	line := "diff --git a/file.go b/file.go"
	result := highlight.ColorizeDiffLine(line)

	if !strings.Contains(result, "\x1b[") {
		t.Error("expected ANSI escape sequences in output")
	}
	if !strings.Contains(result, line) {
		t.Errorf("expected result to contain original line %q", line)
	}
}

func TestColorizeDiff_MultipleLines(t *testing.T) {
	diff := `@@ -10,7 +10,9 @@ func processFile(path string) error {
     if err != nil {
         return fmt.Errorf("failed: %w", err)
     }
+    defer f.Close()
-    f.Close()
     return nil
 }`
	result := highlight.ColorizeDiff(diff)

	// Each line should be colorized
	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		if !strings.Contains(line, "\x1b[") {
			t.Errorf("expected ANSI escape sequences in line %q", line)
		}
	}
}

func TestColorizeDiff_EmptyDiff(t *testing.T) {
	result := highlight.ColorizeDiff("")

	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}
