package tui

import (
	"context"
	"log/slog"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mchowning/diffstory/internal/config"
	"github.com/mchowning/diffstory/internal/diff"
	"github.com/mchowning/diffstory/internal/model"
	"github.com/mchowning/diffstory/internal/storage"
)

type Panel int

const (
	PanelDiff    Panel = 0
	PanelSection Panel = 1
	PanelFiles   Panel = 2
)

// GenerateUIState represents the current state of the generate flow
type GenerateUIState int

const (
	GenerateUIStateNone GenerateUIState = iota
	GenerateUIStateSourcePicker
	GenerateUIStateCommitSelector
	GenerateUIStateCommitRangeStart
	GenerateUIStateCommitRangeEnd
	GenerateUIStateContextInput
	GenerateUIStateValidationError
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
	filterLevel  FilterLevel
	testFilter   TestFilter

	// Scroll state for panels
	sectionScrollOffset int
	filesScrollOffset   int

	// Files panel state
	fileTree       *FileNode
	collapsedPaths CollapsedPaths
	selectedFile   int
	flattenedFiles []*FileNode

	// Keybinding registry for help display
	keybindings *KeybindingRegistry

	// LLM generation state
	config             *config.Config
	store              *storage.Store
	isGenerating       bool
	generateStartTime  time.Time
	cancelGenerate     context.CancelFunc
	showCancelPrompt   bool
	spinner            spinner.Model

	// Generate UI state
	generateUIState    GenerateUIState
	diffSources        []DiffSource
	diffSourceSelected int
	selectedDiffSource *DiffSource // The source selected for generation

	// Commit selector state
	commits           []CommitInfo
	commitSelected    int
	commitInput       textinput.Model
	commitInputActive bool
	rangeStartCommit  string // For commit range selection

	// Context input state
	contextInput textarea.Model
	lastContext  string // Preserved for retry

	// Validation error state
	parsedHunks     []diff.ParsedHunk
	missingHunkIDs  []string
	lastLLMResponse *LLMResponse // Cached for "proceed with partial" option

	// Logging
	logger *slog.Logger
}

func NewModel(workDir string, cfg *config.Config, store *storage.Store, logger *slog.Logger) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))

	// Initialize commit input
	ci := textinput.New()
	ci.Placeholder = "commit hash or ref"
	ci.CharLimit = 64

	// Initialize context textarea
	ctx := textarea.New()
	ctx.Placeholder = "Additional context for the reviewer (optional)..."
	ctx.CharLimit = 2000
	ctx.SetWidth(60)
	ctx.SetHeight(5)

	// Initialize filter level from config
	filterLevel := FilterLevelLow // default
	if cfg != nil {
		switch cfg.DefaultFilterLevel {
		case "medium":
			filterLevel = FilterLevelMedium
		case "high":
			filterLevel = FilterLevelHigh
		default:
			filterLevel = FilterLevelLow
		}
	}

	return Model{
		workDir:      workDir,
		focusedPanel: PanelSection,
		filterLevel:  filterLevel,
		keybindings:  initKeybindings(),
		config:       cfg,
		store:        store,
		spinner:      s,
		logger:       logger,
		diffSources:  DefaultDiffSources(),
		commitInput:  ci,
		contextInput: ctx,
	}
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
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

func (m Model) SectionScrollOffset() int {
	return m.sectionScrollOffset
}

func (m Model) FilesScrollOffset() int {
	return m.filesScrollOffset
}

// sectionPanelHeight calculates the height of the section panel.
func (m Model) sectionPanelHeight() int {
	contentHeight := m.height - 5 // header + footer + filter line
	return contentHeight / 2
}

// filesPanelHeight calculates the height of the files panel.
func (m Model) filesPanelHeight() int {
	contentHeight := m.height - 5 // header + footer + filter line
	sectionHeight := contentHeight / 2
	return contentHeight - sectionHeight
}

func (m Model) FlattenedFilesCount() int {
	if m.flattenedFiles == nil {
		return 0
	}
	return len(m.flattenedFiles)
}

func (m Model) IsGenerating() bool {
	return m.isGenerating
}

func (m Model) GenerateUIState() GenerateUIState {
	return m.generateUIState
}

func (m Model) FilterLevel() FilterLevel {
	return m.filterLevel
}

func (m Model) TestFilter() TestFilter {
	return m.testFilter
}

// SetSectionScrollOffset is a test helper to set the section scroll offset
func (m Model) SetSectionScrollOffset(offset int) Model {
	m.sectionScrollOffset = offset
	return m
}

// SetFilesScrollOffset is a test helper to set the files scroll offset
func (m Model) SetFilesScrollOffset(offset int) Model {
	m.filesScrollOffset = offset
	return m
}

// SetGenerating is a test helper to set the generating state
func (m Model) SetGenerating(generating bool) Model {
	m.isGenerating = generating
	return m
}

// SetShowCancelPrompt is a test helper to set the cancel prompt state
func (m Model) SetShowCancelPrompt(show bool) Model {
	m.showCancelPrompt = show
	return m
}

// SetCancelFunc is a test helper to set a cancel function
func (m Model) SetCancelFunc(cancel func()) Model {
	m.cancelGenerate = cancel
	return m
}

func (m Model) ShowCancelPrompt() bool {
	return m.showCancelPrompt
}

// Commits returns the loaded commit list (for testing)
func (m Model) Commits() []CommitInfo {
	return m.commits
}

func (m *Model) updateViewportContent() {
	if m.review == nil {
		m.viewport.SetContent("")
		return
	}
	sections := m.review.AllSections()
	if m.selected >= len(sections) {
		m.viewport.SetContent("")
		return
	}

	section := sections[m.selected]
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
	sections := m.review.AllSections()
	if m.review == nil || m.selected >= len(sections) {
		m.fileTree = nil
		m.flattenedFiles = nil
		return
	}

	section := sections[m.selected]
	paths := m.extractFilteredFilePaths(section)
	m.fileTree = BuildFileTree(paths)
	m.collapsedPaths = make(CollapsedPaths)
	m.flattenedFiles = Flatten(m.fileTree, m.collapsedPaths)
	m.selectedFile = 0
	m.filesScrollOffset = 0
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

func (m Model) extractFilteredFilePaths(section model.Section) []string {
	seen := make(map[string]bool)
	var paths []string
	for _, hunk := range section.Hunks {
		if m.hunkPassesFilters(hunk) && !seen[hunk.File] {
			seen[hunk.File] = true
			paths = append(paths, hunk.File)
		}
	}
	return paths
}

// hunkPassesFilters returns true if the hunk passes both importance and test filters
func (m Model) hunkPassesFilters(hunk model.Hunk) bool {
	return m.filterLevel.PassesFilter(hunk.Importance) && m.testFilter.PassesFilter(hunk.IsTest)
}
