package diff

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestParse_EmptyDiff(t *testing.T) {
	hunks, err := Parse("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hunks) != 0 {
		t.Errorf("expected 0 hunks, got %d", len(hunks))
	}
}

func TestParse_SingleFileWithSingleHunk(t *testing.T) {
	diff := `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -10,6 +10,7 @@ func main() {
 	fmt.Println("hello")
+	fmt.Println("world")
 }
`
	hunks, err := Parse(diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(hunks))
	}

	hunk := hunks[0]
	if hunk.File != "main.go" {
		t.Errorf("expected file 'main.go', got %q", hunk.File)
	}
	if hunk.StartLine != 10 {
		t.Errorf("expected start line 10, got %d", hunk.StartLine)
	}
	if hunk.ID != "main.go::10" {
		t.Errorf("expected ID 'main.go::10', got %q", hunk.ID)
	}
	if !strings.Contains(hunk.Diff, "@@ -10,6 +10,7 @@") {
		t.Errorf("expected diff to contain hunk header, got %q", hunk.Diff)
	}
}

func TestParse_SingleFileWithMultipleHunks(t *testing.T) {
	diff := `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -10,6 +10,7 @@ func main() {
 	fmt.Println("hello")
+	fmt.Println("world")
 }
@@ -50,3 +51,4 @@ func helper() {
 	return nil
+	// added comment
 }
`
	hunks, err := Parse(diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hunks) != 2 {
		t.Fatalf("expected 2 hunks, got %d", len(hunks))
	}

	if hunks[0].ID != "main.go::10" {
		t.Errorf("expected first hunk ID 'main.go::10', got %q", hunks[0].ID)
	}
	if hunks[0].StartLine != 10 {
		t.Errorf("expected first hunk start line 10, got %d", hunks[0].StartLine)
	}

	if hunks[1].ID != "main.go::51" {
		t.Errorf("expected second hunk ID 'main.go::51', got %q", hunks[1].ID)
	}
	if hunks[1].StartLine != 51 {
		t.Errorf("expected second hunk start line 51, got %d", hunks[1].StartLine)
	}
}

func TestParse_MultipleFiles(t *testing.T) {
	diff := `diff --git a/foo.go b/foo.go
index 1234567..abcdefg 100644
--- a/foo.go
+++ b/foo.go
@@ -1,3 +1,4 @@
 package foo
+// comment

diff --git a/bar.go b/bar.go
index 1234567..abcdefg 100644
--- a/bar.go
+++ b/bar.go
@@ -5,2 +5,3 @@ func Bar() {
 	return
+	// another comment
 }
`
	hunks, err := Parse(diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hunks) != 2 {
		t.Fatalf("expected 2 hunks, got %d", len(hunks))
	}

	if hunks[0].File != "foo.go" {
		t.Errorf("expected first file 'foo.go', got %q", hunks[0].File)
	}
	if hunks[1].File != "bar.go" {
		t.Errorf("expected second file 'bar.go', got %q", hunks[1].File)
	}
}

func TestParse_BinaryFilesAreSkipped(t *testing.T) {
	diff := `diff --git a/image.png b/image.png
Binary files a/image.png and b/image.png differ
diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main
+// comment
`
	hunks, err := Parse(diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hunks) != 1 {
		t.Fatalf("expected 1 hunk (binary skipped), got %d", len(hunks))
	}
	if hunks[0].File != "main.go" {
		t.Errorf("expected file 'main.go', got %q", hunks[0].File)
	}
}

func TestParse_DuplicateIDCollisionHandling(t *testing.T) {
	// Contrived case: two hunks with same file and start line
	// This shouldn't happen in practice but we handle it for safety
	diff := `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -10,3 +10,4 @@ func main() {
 	fmt.Println("hello")
+	fmt.Println("world")
 }
@@ -10,3 +10,4 @@ func other() {
 	fmt.Println("foo")
+	fmt.Println("bar")
 }
`
	hunks, err := Parse(diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hunks) != 2 {
		t.Fatalf("expected 2 hunks, got %d", len(hunks))
	}

	// First hunk keeps original ID
	if hunks[0].ID != "main.go::10" {
		t.Errorf("expected first hunk ID 'main.go::10', got %q", hunks[0].ID)
	}
	// Second hunk gets disambiguated ID
	if hunks[1].ID != "main.go::10#2" {
		t.Errorf("expected second hunk ID 'main.go::10#2', got %q", hunks[1].ID)
	}
}

func TestParse_NewFile(t *testing.T) {
	diff := `diff --git a/newfile.go b/newfile.go
new file mode 100644
index 0000000..1234567
--- /dev/null
+++ b/newfile.go
@@ -0,0 +1,5 @@
+package newfile
+
+func New() {
+	// new function
+}
`
	hunks, err := Parse(diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(hunks))
	}
	if hunks[0].File != "newfile.go" {
		t.Errorf("expected file 'newfile.go', got %q", hunks[0].File)
	}
	if hunks[0].StartLine != 1 {
		t.Errorf("expected start line 1, got %d", hunks[0].StartLine)
	}
}

func TestParse_RenamedFile(t *testing.T) {
	diff := `diff --git a/old.go b/new.go
similarity index 95%
rename from old.go
rename to new.go
index 1234567..abcdefg 100644
--- a/old.go
+++ b/new.go
@@ -1,3 +1,4 @@
 package pkg
+// added
`
	hunks, err := Parse(diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(hunks))
	}
	// Should use the new name (from b/ path)
	if hunks[0].File != "new.go" {
		t.Errorf("expected file 'new.go' (renamed to), got %q", hunks[0].File)
	}
}

func TestParse_StartLineFromPlusSide(t *testing.T) {
	// Verify we use the + (new file) line number, not - (old file)
	diff := `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -100,5 +200,6 @@ func shifted() {
 	// This function moved
+	// added line
 }
`
	hunks, err := Parse(diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(hunks))
	}
	// Start line should be 200 (from +200), not 100 (from -100)
	if hunks[0].StartLine != 200 {
		t.Errorf("expected start line 200 (from + side), got %d", hunks[0].StartLine)
	}
	if hunks[0].ID != "main.go::200" {
		t.Errorf("expected ID 'main.go::200', got %q", hunks[0].ID)
	}
}

func TestParse_NestedFilePaths(t *testing.T) {
	diff := `diff --git a/internal/pkg/deep/file.go b/internal/pkg/deep/file.go
index 1234567..abcdefg 100644
--- a/internal/pkg/deep/file.go
+++ b/internal/pkg/deep/file.go
@@ -1,3 +1,4 @@
 package deep
+// comment
`
	hunks, err := Parse(diff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(hunks))
	}
	if hunks[0].File != "internal/pkg/deep/file.go" {
		t.Errorf("expected nested path, got %q", hunks[0].File)
	}
	if hunks[0].ID != "internal/pkg/deep/file.go::1" {
		t.Errorf("expected ID with nested path, got %q", hunks[0].ID)
	}
}

func TestParse_Performance_LargeDiff(t *testing.T) {
	// Generate a diff with 150+ hunks across multiple files
	var sb strings.Builder
	for fileNum := 0; fileNum < 15; fileNum++ {
		sb.WriteString(fmt.Sprintf("diff --git a/file%d.go b/file%d.go\n", fileNum, fileNum))
		sb.WriteString("index 1234567..abcdefg 100644\n")
		sb.WriteString(fmt.Sprintf("--- a/file%d.go\n", fileNum))
		sb.WriteString(fmt.Sprintf("+++ b/file%d.go\n", fileNum))

		for hunkNum := 0; hunkNum < 10; hunkNum++ {
			startLine := hunkNum*100 + 1
			sb.WriteString(fmt.Sprintf("@@ -%d,5 +%d,6 @@ func example%d() {\n", startLine, startLine, hunkNum))
			sb.WriteString(" 	existing line\n")
			sb.WriteString("+	added line\n")
			sb.WriteString(" 	another line\n")
		}
	}

	diff := sb.String()

	start := time.Now()
	hunks, err := Parse(diff)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hunks) != 150 {
		t.Errorf("expected 150 hunks, got %d", len(hunks))
	}
	if elapsed > time.Second {
		t.Errorf("parsing took %v, expected < 1 second", elapsed)
	}
}
