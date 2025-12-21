package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mchowning/diffguide/internal/model"
)

type Panel int

const (
	PanelDiff    Panel = 0
	PanelSection Panel = 1
	PanelFiles   Panel = 2
)

type Model struct {
	review       *model.Review
	selected     int
	width        int
	height       int
	workDir      string
	viewport     viewport.Model
	ready        bool
	showHelp     bool
	statusMsg    string
	focusedPanel Panel

	// Files panel state
	fileTree       *FileNode
	collapsedPaths CollapsedPaths
	selectedFile   int
	flattenedFiles []*FileNode
}

func NewModel(workDir string) Model {
	return Model{
		workDir:      workDir,
		focusedPanel: PanelSection,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Width() int {
	return m.width
}

func (m Model) Height() int {
	return m.height
}

func (m Model) WorkDir() string {
	return m.workDir
}

func (m Model) Review() *model.Review {
	return m.review
}

func (m Model) Selected() int {
	return m.selected
}

func (m Model) Ready() bool {
	return m.ready
}

func (m Model) ViewportYOffset() int {
	return m.viewport.YOffset
}

func (m Model) ShowHelp() bool {
	return m.showHelp
}

func (m Model) StatusMsg() string {
	return m.statusMsg
}

func (m Model) FocusedPanel() Panel {
	return m.focusedPanel
}

func (m Model) SelectedFile() int {
	return m.selectedFile
}

func (m Model) FlattenedFilesCount() int {
	if m.flattenedFiles == nil {
		return 0
	}
	return len(m.flattenedFiles)
}

func (m *Model) updateViewportContent() {
	if m.review == nil || m.selected >= len(m.review.Sections) {
		m.viewport.SetContent("")
		return
	}

	section := m.review.Sections[m.selected]
	var content string

	if m.flattenedFiles != nil && m.selectedFile < len(m.flattenedFiles) {
		selectedNode := m.flattenedFiles[m.selectedFile]
		if selectedNode.IsDir {
			content = m.renderDiffForDirectory(section, selectedNode.FullPath)
		} else {
			content = m.renderDiffForFile(section, selectedNode.FullPath)
		}
	} else {
		content = m.renderDiffContent(section)
	}

	m.viewport.SetContent(content)
}

func (m *Model) updateFileTree() {
	if m.review == nil || m.selected >= len(m.review.Sections) {
		m.fileTree = nil
		m.flattenedFiles = nil
		return
	}

	section := m.review.Sections[m.selected]
	paths := extractFilePaths(section)
	m.fileTree = BuildFileTree(paths)
	m.collapsedPaths = make(CollapsedPaths)
	m.flattenedFiles = Flatten(m.fileTree, m.collapsedPaths)
	m.selectedFile = 0
}

func extractFilePaths(section model.Section) []string {
	seen := make(map[string]bool)
	var paths []string
	for _, hunk := range section.Hunks {
		if !seen[hunk.File] {
			seen[hunk.File] = true
			paths = append(paths, hunk.File)
		}
	}
	return paths
}
