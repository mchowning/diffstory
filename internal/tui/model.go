package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mchowning/diffguide/internal/model"
)

type Model struct {
	review   *model.Review
	selected int
	width    int
	height   int
	workDir  string
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
