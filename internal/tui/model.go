package tui

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mchowning/diffguide/internal/model"
)

type Model struct {
	review    *model.Review
	selected  int
	width     int
	height    int
	workDir   string
	viewport  viewport.Model
	ready     bool
	showHelp  bool
	statusMsg string
}

func NewModel(workDir string) Model {
	return Model{workDir: workDir}
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

func (m *Model) updateViewportContent() {
	if m.review == nil || m.selected >= len(m.review.Sections) {
		m.viewport.SetContent("")
		return
	}

	section := m.review.Sections[m.selected]
	content := m.renderDiffContent(section)
	m.viewport.SetContent(content)
}
