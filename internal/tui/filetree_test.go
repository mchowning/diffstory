package tui_test

import (
	"testing"

	"github.com/mchowning/diffstory/internal/tui"
)

func TestBuildFileTree_SingleFile(t *testing.T) {
	paths := []string{"src/main.go"}

	tree := tui.BuildFileTree(paths)

	if tree == nil {
		t.Fatal("expected tree, got nil")
	}

	// Root should have one child (src directory)
	if len(tree.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(tree.Children))
	}

	srcDir := tree.Children[0]
	if srcDir.Name != "src" {
		t.Errorf("expected 'src', got %q", srcDir.Name)
	}
	if !srcDir.IsDir {
		t.Error("expected src to be a directory")
	}

	// src should have one child (main.go file)
	if len(srcDir.Children) != 1 {
		t.Fatalf("expected 1 child in src, got %d", len(srcDir.Children))
	}

	mainFile := srcDir.Children[0]
	if mainFile.Name != "main.go" {
		t.Errorf("expected 'main.go', got %q", mainFile.Name)
	}
	if mainFile.IsDir {
		t.Error("expected main.go to be a file, not directory")
	}
}

func TestBuildFileTree_MultipleFilesInSameDir(t *testing.T) {
	paths := []string{"src/a.go", "src/b.go", "src/c.go"}

	tree := tui.BuildFileTree(paths)

	if len(tree.Children) != 1 {
		t.Fatalf("expected 1 child (src), got %d", len(tree.Children))
	}

	srcDir := tree.Children[0]
	if len(srcDir.Children) != 3 {
		t.Fatalf("expected 3 children in src, got %d", len(srcDir.Children))
	}
}

func TestBuildFileTree_NestedDirectories(t *testing.T) {
	paths := []string{"src/pkg/util/helpers.go"}

	tree := tui.BuildFileTree(paths)

	// Navigate: root -> src -> pkg -> util -> helpers.go
	if len(tree.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(tree.Children))
	}

	src := tree.Children[0]
	if src.Name != "src" || !src.IsDir {
		t.Errorf("expected src directory, got %q (isDir=%v)", src.Name, src.IsDir)
	}

	if len(src.Children) != 1 {
		t.Fatalf("expected 1 child in src, got %d", len(src.Children))
	}

	pkg := src.Children[0]
	if pkg.Name != "pkg" || !pkg.IsDir {
		t.Errorf("expected pkg directory, got %q (isDir=%v)", pkg.Name, pkg.IsDir)
	}

	if len(pkg.Children) != 1 {
		t.Fatalf("expected 1 child in pkg, got %d", len(pkg.Children))
	}

	util := pkg.Children[0]
	if util.Name != "util" || !util.IsDir {
		t.Errorf("expected util directory, got %q (isDir=%v)", util.Name, util.IsDir)
	}

	if len(util.Children) != 1 {
		t.Fatalf("expected 1 child in util, got %d", len(util.Children))
	}

	file := util.Children[0]
	if file.Name != "helpers.go" || file.IsDir {
		t.Errorf("expected helpers.go file, got %q (isDir=%v)", file.Name, file.IsDir)
	}
}

func TestBuildFileTree_SortsDirsFirst(t *testing.T) {
	paths := []string{"a.go", "src/b.go", "z.go"}

	tree := tui.BuildFileTree(paths)

	if len(tree.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(tree.Children))
	}

	// First child should be the directory (src)
	if !tree.Children[0].IsDir {
		t.Error("expected first child to be a directory")
	}
	if tree.Children[0].Name != "src" {
		t.Errorf("expected first child to be 'src', got %q", tree.Children[0].Name)
	}

	// Remaining should be files
	if tree.Children[1].IsDir || tree.Children[2].IsDir {
		t.Error("expected remaining children to be files")
	}
}

func TestBuildFileTree_SortsAlphabetically(t *testing.T) {
	paths := []string{"z.go", "a.go", "m.go"}

	tree := tui.BuildFileTree(paths)

	if len(tree.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(tree.Children))
	}

	expected := []string{"a.go", "m.go", "z.go"}
	for i, name := range expected {
		if tree.Children[i].Name != name {
			t.Errorf("expected child %d to be %q, got %q", i, name, tree.Children[i].Name)
		}
	}
}

func TestFlattenTree_AllExpanded(t *testing.T) {
	paths := []string{"src/a.go", "src/b.go"}
	tree := tui.BuildFileTree(paths)
	collapsed := tui.CollapsedPaths{}

	flat := tui.Flatten(tree, collapsed)

	// Should have: src, a.go, b.go
	if len(flat) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(flat))
	}

	if flat[0].Name != "src" {
		t.Errorf("expected first node to be 'src', got %q", flat[0].Name)
	}
	if flat[1].Name != "a.go" {
		t.Errorf("expected second node to be 'a.go', got %q", flat[1].Name)
	}
	if flat[2].Name != "b.go" {
		t.Errorf("expected third node to be 'b.go', got %q", flat[2].Name)
	}
}

func TestFlattenTree_CollapsedDirectory(t *testing.T) {
	paths := []string{"src/a.go", "src/b.go"}
	tree := tui.BuildFileTree(paths)
	collapsed := tui.CollapsedPaths{"src": true}

	flat := tui.Flatten(tree, collapsed)

	// Should only have: src (children hidden)
	if len(flat) != 1 {
		t.Fatalf("expected 1 node when src collapsed, got %d", len(flat))
	}

	if flat[0].Name != "src" {
		t.Errorf("expected 'src', got %q", flat[0].Name)
	}
}

func TestToggleCollapse_ExpandsCollapsed(t *testing.T) {
	collapsed := tui.CollapsedPaths{"src": true}

	tui.ToggleCollapse(collapsed, "src")

	if collapsed["src"] {
		t.Error("expected 'src' to be expanded after toggle")
	}
}

func TestToggleCollapse_CollapsesExpanded(t *testing.T) {
	collapsed := tui.CollapsedPaths{}

	tui.ToggleCollapse(collapsed, "src")

	if !collapsed["src"] {
		t.Error("expected 'src' to be collapsed after toggle")
	}
}
