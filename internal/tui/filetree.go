package tui

import (
	"path/filepath"
	"sort"
	"strings"
)

type FileNode struct {
	Name     string
	FullPath string
	IsDir    bool
	Children []*FileNode
}

type CollapsedPaths map[string]bool

func BuildFileTree(paths []string) *FileNode {
	root := &FileNode{Name: "", IsDir: true}

	for _, path := range paths {
		parts := strings.Split(path, string(filepath.Separator))
		current := root

		for i, part := range parts {
			isLast := i == len(parts)-1
			child := findChild(current, part)

			if child == nil {
				fullPath := strings.Join(parts[:i+1], string(filepath.Separator))
				child = &FileNode{
					Name:     part,
					FullPath: fullPath,
					IsDir:    !isLast,
				}
				current.Children = append(current.Children, child)
			}
			current = child
		}
	}

	sortTree(root)
	return root
}

func findChild(parent *FileNode, name string) *FileNode {
	for _, child := range parent.Children {
		if child.Name == name {
			return child
		}
	}
	return nil
}

func sortTree(node *FileNode) {
	sort.Slice(node.Children, func(i, j int) bool {
		if node.Children[i].IsDir != node.Children[j].IsDir {
			return node.Children[i].IsDir
		}
		return node.Children[i].Name < node.Children[j].Name
	})

	for _, child := range node.Children {
		sortTree(child)
	}
}

func Flatten(root *FileNode, collapsed CollapsedPaths) []*FileNode {
	var result []*FileNode
	flattenRecursive(root, collapsed, &result)
	return result
}

func flattenRecursive(node *FileNode, collapsed CollapsedPaths, result *[]*FileNode) {
	for _, child := range node.Children {
		*result = append(*result, child)
		if child.IsDir && !collapsed[child.FullPath] {
			flattenRecursive(child, collapsed, result)
		}
	}
}

func ToggleCollapse(collapsed CollapsedPaths, path string) {
	if collapsed[path] {
		delete(collapsed, path)
	} else {
		collapsed[path] = true
	}
}
