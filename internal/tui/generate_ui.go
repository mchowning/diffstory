package tui

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// dialogStyle creates a bordered dialog box
var dialogStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("240")).
	Padding(1, 2)

// dimStyle is used for less prominent text like command hints
var dimStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

// renderSourcePicker renders the diff source selection UI
func (m Model) renderSourcePicker() string {
	var sb strings.Builder
	sb.WriteString("Select diff source\n\n")

	// Calculate max width for consistent formatting
	maxWidth := 0
	for _, source := range m.diffSources {
		line := "  " + source.Label + commandHintSuffix(source.CommandHint)
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}

	for i, source := range m.diffSources {
		prefix := "  "
		style := normalStyle
		if i == m.diffSourceSelected {
			prefix = "› "
			style = selectedStyle
		}
		sb.WriteString(style.Render(prefix+source.Label) + dimStyle.Render(commandHintSuffix(source.CommandHint)))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	helpBox := helpStyle.Width(maxWidth).Render("j/k  navigate\nEnter  select\nEsc  cancel")
	sb.WriteString(helpBox)

	dialog := dialogStyle.Render(sb.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog)
}

// commandHintSuffix returns the formatted command hint or empty string
func commandHintSuffix(hint string) string {
	if hint == "" {
		return ""
	}
	return " (" + hint + ")"
}

// renderCommitSelector renders the commit selection UI
func (m Model) renderCommitSelector() string {
	var sb strings.Builder

	title := "Select commit"
	if m.generateUIState == GenerateUIStateCommitRangeStart {
		title = "Select start commit (older)"
	} else if m.generateUIState == GenerateUIStateCommitRangeEnd {
		title = "Select end commit (newer)"
	}
	sb.WriteString(title + "\n\n")

	// Calculate available width for commit messages
	// Dialog padding (2 each side) + border (1 each side) + prefix (2) + hash (7) + spaces (2) + age (~15) = ~30 overhead
	dialogWidth := min(m.width-4, 100) // Leave margin, cap at 100
	subjectWidth := max(dialogWidth-35, 30)

	// Calculate how many commits we can show based on terminal height
	// Reserve: title (2 lines) + help (4 lines) + custom input (3 lines) + padding (4) = ~13 lines overhead
	maxDisplay := max(min((m.height-13), 25), 5)

	// Show commit list
	for i, commit := range m.commits {
		if i >= maxDisplay {
			sb.WriteString(fmt.Sprintf("  ... and %d more commits\n", len(m.commits)-maxDisplay))
			break
		}
		prefix := "  "
		style := normalStyle
		if i == m.commitSelected && !m.commitInputActive {
			prefix = "› "
			style = selectedStyle
		}
		hash := commit.Hash[:min(7, len(commit.Hash))]
		subject := truncate(commit.Subject, subjectWidth)
		line := fmt.Sprintf("%s %s (%s)", hash, subject, commit.Age)
		sb.WriteString(style.Render(prefix + line))
		sb.WriteString("\n")
	}

	// Show custom input
	sb.WriteString("\n")
	inputPrefix := "  "
	if m.commitInputActive {
		inputPrefix = "› "
	}
	sb.WriteString(inputPrefix + "Custom: " + m.commitInput.View())
	sb.WriteString("\n\n")
	sb.WriteString(helpStyle.Render("j/k  navigate\nTab  custom input\nEnter  select\nEsc  cancel"))

	// Use wider dialog style
	wideDialogStyle := dialogStyle.Width(dialogWidth)
	dialog := wideDialogStyle.Render(sb.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog)
}

// renderContextInput renders the context input UI
func (m Model) renderContextInput() string {
	var sb strings.Builder
	sb.WriteString("Instructions for reviewer (editable)\n\n")
	sb.WriteString(m.contextInput.View())
	sb.WriteString("\n\n")
	sb.WriteString(helpStyle.Render("Enter  generate\nAlt+Enter  new line\nEsc  cancel"))

	dialog := dialogStyle.Render(sb.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog)
}

// renderValidationError renders the validation error state with retry options
func (m Model) renderValidationError() string {
	var sb strings.Builder
	sb.WriteString("Classification incomplete\n\n")
	sb.WriteString("Missing hunks:\n")
	maxDisplay := 10
	for i, id := range m.missingHunkIDs {
		if i >= maxDisplay {
			sb.WriteString(fmt.Sprintf("  ... and %d more\n", len(m.missingHunkIDs)-maxDisplay))
			break
		}
		sb.WriteString(fmt.Sprintf("  • %s\n", id))
	}
	sb.WriteString("\n")
	sb.WriteString(helpStyle.Render("r  retry\np  proceed with partial\nEsc  cancel"))

	dialog := dialogStyle.Render(sb.String())
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog)
}

// updateSourcePicker handles key events in the source picker state
func (m Model) updateSourcePicker(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.diffSourceSelected < len(m.diffSources)-1 {
			m.diffSourceSelected++
		}
	case "k", "up":
		if m.diffSourceSelected > 0 {
			m.diffSourceSelected--
		}
	case "enter":
		source := m.diffSources[m.diffSourceSelected]
		m.selectedDiffSource = &source
		if source.NeedsCommit {
			m.generateUIState = GenerateUIStateCommitSelector
			m.commitSelected = 0
			m.commitInputActive = false
			return m, loadCommitList()
		} else if source.NeedsCommitRange {
			m.generateUIState = GenerateUIStateCommitRangeStart
			m.commitSelected = 0
			m.commitInputActive = false
			return m, loadCommitList()
		}
		m.generateUIState = GenerateUIStateContextInput
		m.contextInput.Focus()
		return m, textarea.Blink
	case "esc":
		m.generateUIState = GenerateUIStateNone
	}
	return m, nil
}

// updateCommitSelector handles key events in the commit selector state
func (m Model) updateCommitSelector(msg tea.KeyMsg) (Model, tea.Cmd) {
	if m.commitInputActive {
		switch msg.String() {
		case "tab":
			m.commitInputActive = false
			m.commitInput.Blur()
		case "enter":
			return m.selectCommit(m.commitInput.Value())
		case "esc":
			m.generateUIState = GenerateUIStateSourcePicker
			m.commitInput.SetValue("")
			m.commitInput.Blur()
			return m, nil
		default:
			var cmd tea.Cmd
			m.commitInput, cmd = m.commitInput.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	switch msg.String() {
	case "j", "down":
		if m.commitSelected < len(m.commits)-1 {
			m.commitSelected++
		}
	case "k", "up":
		if m.commitSelected > 0 {
			m.commitSelected--
		}
	case "tab":
		m.commitInputActive = true
		m.commitInput.Focus()
		return m, textinput.Blink
	case "enter":
		if len(m.commits) > 0 && m.commitSelected < len(m.commits) {
			return m.selectCommit(m.commits[m.commitSelected].Hash)
		}
	case "esc":
		m.generateUIState = GenerateUIStateSourcePicker
	}
	return m, nil
}

// selectCommit handles commit selection for both single commit and range modes
func (m Model) selectCommit(commitRef string) (Model, tea.Cmd) {
	if m.generateUIState == GenerateUIStateCommitRangeStart {
		m.rangeStartCommit = commitRef
		m.generateUIState = GenerateUIStateCommitRangeEnd
		m.commitSelected = 0
		m.commitInputActive = false
		m.commitInput.SetValue("")
		return m, nil
	} else if m.generateUIState == GenerateUIStateCommitRangeEnd {
		// Build range command
		m.selectedDiffSource = &DiffSource{
			Label:   fmt.Sprintf("Range: %s..%s", m.rangeStartCommit, commitRef),
			Command: []string{"git", "diff", m.rangeStartCommit + ".." + commitRef, "--no-color", "--no-ext-diff"},
		}
		m.generateUIState = GenerateUIStateContextInput
		m.contextInput.Focus()
		return m, textarea.Blink
	}

	// Single commit mode
	m.selectedDiffSource = &DiffSource{
		Label:   fmt.Sprintf("Commit: %s", commitRef),
		Command: []string{"git", "show", commitRef, "--no-color", "--no-ext-diff", "--format="},
	}
	m.generateUIState = GenerateUIStateContextInput
	m.contextInput.Focus()
	return m, textarea.Blink
}

// updateContextInput handles key events in the context input state
func (m Model) updateContextInput(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		if msg.Alt {
			// Alt+Enter: insert newline
			m.contextInput.InsertString("\n")
			m.contextInput.SetHeight(calcTextareaHeight(m.contextInput.Value(), 60))
			return m, nil
		}
		// Enter: submit and generate
		m.lastContext = m.contextInput.Value()
		m.generateUIState = GenerateUIStateNone
		return m, m.startGeneration()
	case tea.KeyEsc:
		m.generateUIState = GenerateUIStateSourcePicker
		m.contextInput.Blur()
		return m, nil
	default:
		var cmd tea.Cmd
		m.contextInput, cmd = m.contextInput.Update(msg)
		m.contextInput.SetHeight(calcTextareaHeight(m.contextInput.Value(), 60))
		return m, cmd
	}
}

// updateValidationError handles key events in the validation error state
func (m Model) updateValidationError(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "r":
		// Retry generation
		m.generateUIState = GenerateUIStateNone
		return m, m.startRetryGeneration()
	case "p":
		// Proceed with partial results
		m.generateUIState = GenerateUIStateNone
		return m, m.proceedWithPartial()
	case "esc":
		m.generateUIState = GenerateUIStateNone
		m.parsedHunks = nil
		m.missingHunkIDs = nil
	}
	return m, nil
}

// loadCommitList returns a command that loads the recent commit list
func loadCommitList() tea.Cmd {
	return func() tea.Msg {
		output, err := exec.Command("git", "log", "--oneline", "-n", "50", "--format=%h|%s|%cr").Output()
		if err != nil {
			return CommitListErrorMsg{Err: err}
		}

		var commits []CommitInfo
		for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
			if line == "" {
				continue
			}
			parts := strings.SplitN(line, "|", 3)
			if len(parts) == 3 {
				commits = append(commits, CommitInfo{
					Hash:    parts[0],
					Subject: parts[1],
					Age:     parts[2],
				})
			}
		}
		return CommitListMsg{Commits: commits}
	}
}

// startGeneration begins the LLM generation with the selected source
func (m *Model) startGeneration() tea.Cmd {
	if m.selectedDiffSource == nil {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.cancelGenerate = cancel
	m.isGenerating = true
	m.generateStartTime = time.Now()

	params := GenerateParams{
		DiffCommand: m.selectedDiffSource.Command,
		Context:     m.lastContext,
		IsRetry:     false,
	}

	return tea.Batch(
		m.spinner.Tick,
		generateReviewCmd(ctx, m.config, m.workDir, m.store, m.logger, params),
	)
}

// startRetryGeneration retries generation with the preserved context
func (m *Model) startRetryGeneration() tea.Cmd {
	if m.selectedDiffSource == nil {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.cancelGenerate = cancel
	m.isGenerating = true
	m.generateStartTime = time.Now()

	params := GenerateParams{
		DiffCommand: m.selectedDiffSource.Command,
		Context:     m.lastContext,
		IsRetry:     true,
		MissingIDs:  m.missingHunkIDs,
		ParsedHunks: m.parsedHunks,
	}

	return tea.Batch(
		m.spinner.Tick,
		generateReviewCmd(ctx, m.config, m.workDir, m.store, m.logger, params),
	)
}

// proceedWithPartial saves the review with unclassified hunks
func (m *Model) proceedWithPartial() tea.Cmd {
	if m.lastLLMResponse == nil {
		return func() tea.Msg {
			return GenerateErrorMsg{Err: fmt.Errorf("no cached response available")}
		}
	}

	response := m.lastLLMResponse
	hunks := m.parsedHunks
	missingIDs := m.missingHunkIDs
	workDir := m.workDir
	store := m.store

	return func() tea.Msg {
		review := assemblePartialReview(workDir, response, hunks, missingIDs)
		if err := store.Write(review); err != nil {
			return GenerateErrorMsg{Err: fmt.Errorf("failed to save partial review: %w", err)}
		}
		return GenerateSuccessMsg{}
	}
}

// truncate shortens a string to maxLen, adding "..." if truncated
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
