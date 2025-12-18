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
	port     string
}

func NewModel(port string) Model {
	return Model{port: port}
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

func (m Model) Port() string {
	return m.port
}
