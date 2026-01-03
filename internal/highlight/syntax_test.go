package highlight_test

import (
	"strings"
	"testing"

	"github.com/mchowning/diffstory/internal/highlight"
)

func TestHighlightCode_GoFile(t *testing.T) {
	code := `func main() {
	fmt.Println("Hello")
}`
	result, err := highlight.HighlightCode(code, "main.go")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "\x1b[") {
		t.Error("expected ANSI escape sequences in output")
	}
	if !strings.Contains(result, "func") {
		t.Error("expected result to contain original code")
	}
}

func TestHighlightCode_PythonFile(t *testing.T) {
	code := `def hello():
    print("Hello")`
	result, err := highlight.HighlightCode(code, "script.py")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "\x1b[") {
		t.Error("expected ANSI escape sequences in output")
	}
	if !strings.Contains(result, "def") {
		t.Error("expected result to contain original code")
	}
}

func TestHighlightCode_UnknownExtension(t *testing.T) {
	code := "some plain text content"
	result, err := highlight.HighlightCode(code, "file.unknown")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "some plain text content") {
		t.Error("expected result to contain original code")
	}
}

func TestHighlightCode_EmptyCode(t *testing.T) {
	result, err := highlight.HighlightCode("", "main.go")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}
