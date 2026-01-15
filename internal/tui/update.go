package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Route to generate UI handlers when in a generate state
		switch m.generateUIState {
		case GenerateUIStateSourcePicker:
			return m.updateSourcePicker(msg)
		case GenerateUIStateCommitSelector, GenerateUIStateCommitRangeStart, GenerateUIStateCommitRangeEnd:
			return m.updateCommitSelector(msg)
		case GenerateUIStateContextInput:
			return m.updateContextInput(msg)
		case GenerateUIStateValidationError:
			return m.updateValidationError(msg)
		case GenerateUIStateUntrackedWarning:
			return m.updateUntrackedWarning(msg)
		}

		// Handle arrow keys for panel focus cycling
		switch msg.Type {
		case tea.KeyLeft:
			if m.focusedPanel == PanelSection {
				m.focusedPanel = PanelFiles
				m.updateViewportContent()
			} else if m.focusedPanel == PanelFiles {
				m.focusedPanel = PanelSection
				m.updateViewportContent()
			}
		case tea.KeyRight:
			if m.focusedPanel == PanelSection {
				m.focusedPanel = PanelFiles
				m.updateViewportContent()
			} else if m.focusedPanel == PanelFiles {
				m.focusedPanel = PanelSection
				m.updateViewportContent()
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "j", "down":
			switch m.focusedPanel {
			case PanelSection:
				if m.review != nil && m.selected < m.review.SectionCount()-1 {
					m.selected++
					m.sectionScrollOffset = CalculateScrollOffset(
						m.sectionScrollOffset,
						m.selected,
						m.review.SectionCount(),
						EstimateSectionVisibleCount(m.sectionPanelHeight()),
					)
					m.viewport.GotoTop()
					m.updateFileTree()
					m.updateViewportContent()
				}
			case PanelFiles:
				if m.flattenedFiles != nil && m.selectedFile < len(m.flattenedFiles)-1 {
					m.selectedFile++
					m.filesScrollOffset = CalculateScrollOffset(
						m.filesScrollOffset,
						m.selectedFile,
						len(m.flattenedFiles),
						EstimateFilesVisibleCount(m.filesPanelHeight()),
					)
					m.updateViewportContent()
					m.viewport.GotoTop()
				}
			case PanelDiff:
				m.viewport.LineDown(1)
			}
		case "k", "up":
			switch m.focusedPanel {
			case PanelSection:
				if m.review != nil && m.selected > 0 {
					m.selected--
					m.sectionScrollOffset = CalculateScrollOffset(
						m.sectionScrollOffset,
						m.selected,
						m.review.SectionCount(),
						EstimateSectionVisibleCount(m.sectionPanelHeight()),
					)
					m.viewport.GotoTop()
					m.updateFileTree()
					m.updateViewportContent()
				}
			case PanelFiles:
				if m.selectedFile > 0 {
					m.selectedFile--
					m.filesScrollOffset = CalculateScrollOffset(
						m.filesScrollOffset,
						m.selectedFile,
						len(m.flattenedFiles),
						EstimateFilesVisibleCount(m.filesPanelHeight()),
					)
					m.updateViewportContent()
					m.viewport.GotoTop()
				}
			case PanelDiff:
				m.viewport.LineUp(1)
			}
		case "J":
			m.viewport.LineDown(1)
		case "K":
			m.viewport.LineUp(1)
		case "]":
			m.viewport.HalfViewDown()
		case "[":
			m.viewport.HalfViewUp()
		case "ctrl+j":
			if m.flattenedFiles != nil && m.selectedFile < len(m.flattenedFiles)-1 {
				m.selectedFile++
				m.filesScrollOffset = CalculateScrollOffset(
					m.filesScrollOffset,
					m.selectedFile,
					len(m.flattenedFiles),
					EstimateFilesVisibleCount(m.filesPanelHeight()),
				)
				m.updateViewportContent()
				m.viewport.GotoTop()
			}
		case "ctrl+k":
			if m.selectedFile > 0 {
				m.selectedFile--
				m.filesScrollOffset = CalculateScrollOffset(
					m.filesScrollOffset,
					m.selectedFile,
					len(m.flattenedFiles),
					EstimateFilesVisibleCount(m.filesPanelHeight()),
				)
				m.updateViewportContent()
				m.viewport.GotoTop()
			}
		case "enter":
			if m.focusedPanel == PanelFiles && m.flattenedFiles != nil && m.selectedFile < len(m.flattenedFiles) {
				node := m.flattenedFiles[m.selectedFile]
				if node.IsDir {
					ToggleCollapse(m.collapsedPaths, node.FullPath)
					m.flattenedFiles = Flatten(m.fileTree, m.collapsedPaths)
				}
			}
		case "?":
			m.showHelp = !m.showHelp
		case "f":
			if m.review != nil {
				m.filterLevel = m.filterLevel.Next()
				m.updateFileTree()
				m.updateViewportContent()
			}
		case "t":
			if m.review != nil {
				m.testFilter = m.testFilter.Next()
				m.updateFileTree()
				m.updateViewportContent()
			}
		case "esc":
			if m.showCancelPrompt {
				m.showCancelPrompt = false
			} else if m.isGenerating {
				m.showCancelPrompt = true
			} else if m.showHelp {
				m.showHelp = false
			}
		case "G":
			if m.isGenerating {
				return m, nil // Ignore if already generating
			}
			result := ResolveLLMCommand(m.config, m.lookPath)
			if result.Error != "" {
				m.statusMsg = result.Error
				return m, nil
			}
			if m.store == nil {
				m.statusMsg = "Storage not initialized"
				return m, nil
			}
			m.resolvedLLMCommand = result.Command
			m.generateUIState = GenerateUIStateSourcePicker
			m.diffSourceSelected = 0
			return m, nil
		case "y":
			if m.showCancelPrompt && m.cancelGenerate != nil {
				m.cancelGenerate()
				m.cancelGenerate = nil
				m.isGenerating = false
				m.showCancelPrompt = false
				m.statusMsg = "Generation cancelled"
				return m, tea.Tick(3*time.Second, func(time.Time) tea.Msg {
					return ClearStatusMsg{}
				})
			}
		case "n":
			if m.showCancelPrompt {
				m.showCancelPrompt = false
			}
		case "0":
			m.focusedPanel = PanelDiff
			m.updateViewportContent()
		case "1":
			m.focusedPanel = PanelSection
			m.updateViewportContent()
		case "2":
			m.focusedPanel = PanelFiles
			m.updateViewportContent()
		case "h":
			if m.focusedPanel == PanelSection {
				m.focusedPanel = PanelFiles
				m.updateViewportContent()
			} else if m.focusedPanel == PanelFiles {
				m.focusedPanel = PanelSection
				m.updateViewportContent()
			}
		case "l":
			if m.focusedPanel == PanelSection {
				m.focusedPanel = PanelFiles
				m.updateViewportContent()
			} else if m.focusedPanel == PanelFiles {
				m.focusedPanel = PanelSection
				m.updateViewportContent()
			}
		case "<":
			switch m.focusedPanel {
			case PanelSection:
				m.selected = 0
				m.sectionScrollOffset = 0
				m.updateFileTree()
				m.viewport.GotoTop()
				m.updateViewportContent()
			case PanelFiles:
				m.selectedFile = 0
				m.filesScrollOffset = 0
				m.updateViewportContent()
				m.viewport.GotoTop()
			case PanelDiff:
				m.viewport.GotoTop()
			}
		case ">":
			switch m.focusedPanel {
			case PanelSection:
				if m.review != nil {
					m.selected = m.review.SectionCount() - 1
					m.sectionScrollOffset = CalculateScrollOffset(
						m.sectionScrollOffset,
						m.selected,
						m.review.SectionCount(),
						EstimateSectionVisibleCount(m.sectionPanelHeight()),
					)
					m.updateFileTree()
					m.viewport.GotoTop()
					m.updateViewportContent()
				}
			case PanelFiles:
				if m.flattenedFiles != nil {
					m.selectedFile = len(m.flattenedFiles) - 1
					m.filesScrollOffset = CalculateScrollOffset(
						m.filesScrollOffset,
						m.selectedFile,
						len(m.flattenedFiles),
						EstimateFilesVisibleCount(m.filesPanelHeight()),
					)
					m.updateViewportContent()
					m.viewport.GotoTop()
				}
			case PanelDiff:
				m.viewport.GotoBottom()
			}
		case ",":
			pageSize := max(1, (m.height-4)/4)
			switch m.focusedPanel {
			case PanelSection:
				m.selected = max(0, m.selected-pageSize)
				m.sectionScrollOffset = CalculateScrollOffset(
					m.sectionScrollOffset,
					m.selected,
					m.review.SectionCount(),
					EstimateSectionVisibleCount(m.sectionPanelHeight()),
				)
				m.updateFileTree()
				m.viewport.GotoTop()
				m.updateViewportContent()
			case PanelFiles:
				m.selectedFile = max(0, m.selectedFile-pageSize)
				m.filesScrollOffset = CalculateScrollOffset(
					m.filesScrollOffset,
					m.selectedFile,
					len(m.flattenedFiles),
					EstimateFilesVisibleCount(m.filesPanelHeight()),
				)
				m.updateViewportContent()
				m.viewport.GotoTop()
			case PanelDiff:
				m.viewport.HalfViewUp()
			}
		case ".":
			pageSize := max(1, (m.height-4)/4)
			switch m.focusedPanel {
			case PanelSection:
				if m.review != nil {
					m.selected = min(m.review.SectionCount()-1, m.selected+pageSize)
					m.sectionScrollOffset = CalculateScrollOffset(
						m.sectionScrollOffset,
						m.selected,
						m.review.SectionCount(),
						EstimateSectionVisibleCount(m.sectionPanelHeight()),
					)
					m.updateFileTree()
					m.viewport.GotoTop()
					m.updateViewportContent()
				}
			case PanelFiles:
				if m.flattenedFiles != nil {
					m.selectedFile = min(len(m.flattenedFiles)-1, m.selectedFile+pageSize)
					m.filesScrollOffset = CalculateScrollOffset(
						m.filesScrollOffset,
						m.selectedFile,
						len(m.flattenedFiles),
						EstimateFilesVisibleCount(m.filesPanelHeight()),
					)
					m.updateViewportContent()
					m.viewport.GotoTop()
				}
			case PanelDiff:
				m.viewport.HalfViewDown()
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Calculate viewport dimensions (right pane is 2/3 width minus borders)
		rightWidth := msg.Width - (msg.Width / 3) - 4
		viewportHeight := msg.Height - 4 // header + footer

		if !m.ready {
			m.viewport = viewport.New(rightWidth, viewportHeight)
			m.ready = true
			// Initialize file tree if review was pre-loaded (e.g., via --review flag)
			if m.review != nil {
				m.updateFileTree()
			}
		} else {
			m.viewport.Width = rightWidth
			m.viewport.Height = viewportHeight
		}
		m.updateViewportContent()
	case ReviewReceivedMsg:
		m.review = &msg.Review
		m.selected = 0
		m.sectionScrollOffset = 0
		m.filesScrollOffset = 0
		m.viewport.GotoTop()
		m.updateFileTree()
		m.updateViewportContent()
		return m, nil
	case ReviewClearedMsg:
		m.review = nil
		m.selected = 0
		return m, nil
	case WatchErrorMsg:
		if m.logger != nil {
			m.logger.Error("watch error", "error", msg.Err)
		}
		m.statusMsg = "Watch error: " + msg.Err.Error()
		return m, tea.Tick(5*time.Second, func(time.Time) tea.Msg {
			return ClearStatusMsg{}
		})
	case ErrorMsg:
		if m.logger != nil {
			m.logger.Error("error received", "error", msg.Err)
		}
		m.statusMsg = "Error: " + msg.Err.Error()
		return m, tea.Tick(3*time.Second, func(time.Time) tea.Msg {
			return ClearStatusMsg{}
		})
	case ClearStatusMsg:
		m.statusMsg = ""
		return m, nil
	case GenerateSuccessMsg:
		// Just stop the spinner - the watcher will deliver the review
		m.isGenerating = false
		m.cancelGenerate = nil
		return m, nil
	case GenerateErrorMsg:
		if m.logger != nil {
			m.logger.Error("generation failed", "error", msg.Err)
		}
		m.isGenerating = false
		m.cancelGenerate = nil
		m.review = nil // Clear stale review
		m.statusMsg = "Error: " + msg.Err.Error()
		return m, tea.Tick(5*time.Second, func(time.Time) tea.Msg {
			return ClearStatusMsg{}
		})
	case GenerateCancelledMsg:
		m.isGenerating = false
		m.cancelGenerate = nil
		m.statusMsg = "Generation cancelled"
		return m, tea.Tick(3*time.Second, func(time.Time) tea.Msg {
			return ClearStatusMsg{}
		})
	case CommitListMsg:
		m.commits = msg.Commits
		m.commitSelected = 0
		return m, nil
	case CommitListErrorMsg:
		m.generateUIState = GenerateUIStateNone
		m.statusMsg = "Failed to load commits: " + msg.Err.Error()
		return m, tea.Tick(5*time.Second, func(time.Time) tea.Msg {
			return ClearStatusMsg{}
		})
	case spinner.TickMsg:
		if m.isGenerating {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	case CheckUntrackedMsg:
		if msg.Err != nil {
			// If we can't check, just proceed with generation
			m.generateUIState = GenerateUIStateNone
			return m, m.startGeneration()
		}
		if len(msg.Files) > 0 {
			m.untrackedFiles = msg.Files
			m.generateUIState = GenerateUIStateUntrackedWarning
			return m, nil
		}
		// No untracked files, proceed with generation
		m.generateUIState = GenerateUIStateNone
		return m, m.startGeneration()
	case StageCompleteMsg:
		// Staging complete, start generation
		m.generateUIState = GenerateUIStateNone
		return m, m.startGeneration()
	}
	return m, nil
}
