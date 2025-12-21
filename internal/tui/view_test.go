package tui_test

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mchowning/diffguide/internal/model"
	"github.com/mchowning/diffguide/internal/tui"
)

func TestView_EmptyStateContainsWorkingDirectory(t *testing.T) {
	m := tui.NewModel("/test/project")
	view := m.View()

	if !strings.Contains(view, "/test/project") {
		t.Error("empty state view should contain working directory")
	}
}

func TestView_EmptyStateContainsQuitInstruction(t *testing.T) {
	m := tui.NewModel("/test/project")
	view := m.View()

	if !strings.Contains(view, "q: quit") {
		t.Error("empty state view should contain 'q: quit'")
	}
}

func TestView_EmptyStateContainsServerInstructions(t *testing.T) {
	m := tui.NewModel("/test/project")
	view := m.View()

	if !strings.Contains(view, "diffguide server") {
		t.Error("empty state view should contain server start instructions")
	}
	if !strings.Contains(view, "POST") {
		t.Error("empty state view should contain POST instruction")
	}
}

func modelWithReviewAndSize(numSections int) tui.Model {
	m := tui.NewModel("/test/project")

	// Initialize viewport via WindowSizeMsg
	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// Set review
	sections := make([]model.Section, numSections)
	for i := range numSections {
		sections[i] = model.Section{
			ID:        string(rune('1' + i)),
			Narrative: "Section " + string(rune('A'+i)),
			Hunks: []model.Hunk{
				{
					File:      "file" + string(rune('1'+i)) + ".go",
					StartLine: 10 + i,
					Diff:      "+added line\n-removed line",
				},
			},
		}
	}
	review := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test Review Title",
		Sections:         sections,
	}
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	return updated.(tui.Model)
}

func TestView_ReviewStateShowsTitle(t *testing.T) {
	m := modelWithReviewAndSize(2)
	view := m.View()

	if !strings.Contains(view, "Test Review Title") {
		t.Error("review state view should contain the title")
	}
}

func TestView_ReviewStateShowsSectionList(t *testing.T) {
	m := modelWithReviewAndSize(2)
	view := m.View()

	if !strings.Contains(view, "Section A") {
		t.Error("review state view should contain section A narrative")
	}
	if !strings.Contains(view, "Section B") {
		t.Error("review state view should contain section B narrative")
	}
}

func TestView_SelectedSectionHasPrefix(t *testing.T) {
	m := modelWithReviewAndSize(2)
	view := m.View()

	// First section should be selected with "› " prefix
	if !strings.Contains(view, "› Section A") {
		t.Error("selected section should have '› ' prefix")
	}
}

func TestView_NonSelectedSectionHasSpacePrefix(t *testing.T) {
	m := modelWithReviewAndSize(2)
	view := m.View()

	// Second section should not be selected, should have "  " prefix
	if !strings.Contains(view, "  Section B") {
		t.Error("non-selected section should have '  ' prefix")
	}
}

func TestView_ReviewStateShowsHunkContent(t *testing.T) {
	m := modelWithReviewAndSize(1)
	view := m.View()

	// Should show file name
	if !strings.Contains(view, "file1.go") {
		t.Error("review state view should contain hunk file name")
	}
}

func TestView_NotReadyShowsInitializing(t *testing.T) {
	m := tui.NewModel("/test/project")
	// Set a review but don't send WindowSizeMsg (so viewport not initialized)
	review := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test",
		Sections:         []model.Section{{ID: "1", Narrative: "Section"}},
	}
	updated, _ := m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	view := m.View()

	if !strings.Contains(view, "Initializing") {
		t.Error("view should show 'Initializing' when viewport not ready")
	}
}

func TestView_SectionListDoesNotTruncateText(t *testing.T) {
	m := tui.NewModel("/test/project")

	// Initialize viewport - width 120 gives section pane ~30 chars for wrapping test
	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// Create a review with a long narrative
	longNarrative := "This is a very long narrative that should wrap instead of being truncated with an ellipsis character"
	review := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test",
		Sections: []model.Section{
			{ID: "1", Narrative: longNarrative},
		},
	}
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	view := m.View()

	// Should NOT contain truncation ellipsis
	if strings.Contains(view, "…") {
		t.Error("section list should wrap text, not truncate with ellipsis")
	}
	// Should contain the full text (or at least the ending words)
	if !strings.Contains(view, "ellipsis character") {
		t.Error("section list should contain full narrative text")
	}
}

func TestView_StatusBarShowsErrorText(t *testing.T) {
	m := tui.NewModel("/test/project")

	// Initialize viewport
	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// Set a review
	review := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test",
		Sections:         []model.Section{{ID: "1", Narrative: "Section"}},
	}
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	// Set an error
	errMsg := tui.ErrorMsg{Err: errors.New("test error")}
	updated, _ = m.Update(errMsg)
	m = updated.(tui.Model)

	view := m.View()

	if !strings.Contains(view, "Error: test error") {
		t.Error("view should contain error message in status bar")
	}
}

func TestView_HelpOverlayContainsKeybindings(t *testing.T) {
	m := tui.NewModel("/test/project")

	// Initialize viewport
	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// Set a review
	review := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test",
		Sections:         []model.Section{{ID: "1", Narrative: "Section"}},
	}
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	// Toggle help on
	helpMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")}
	updated, _ = m.Update(helpMsg)
	m = updated.(tui.Model)

	view := m.View()

	if !strings.Contains(view, "Keybindings") {
		t.Error("help overlay should contain 'Keybindings'")
	}
}

func TestView_SectionListHasSpacingBetweenSections(t *testing.T) {
	m := modelWithReviewAndSize(2)
	view := m.View()

	// Find positions of both sections
	posA := strings.Index(view, "Section A")
	posB := strings.Index(view, "Section B")

	if posA == -1 || posB == -1 {
		t.Fatal("could not find both sections in view")
	}

	// Extract text between sections and check for blank line
	between := view[posA+len("Section A") : posB]

	// Should have at least 2 newlines (indicating a blank line between sections)
	newlineCount := strings.Count(between, "\n")
	if newlineCount < 2 {
		t.Errorf("expected at least 2 newlines between sections for spacing, got %d", newlineCount)
	}
}

// Three-Panel Layout Tests

func TestView_ThreePanelsRendered(t *testing.T) {
	m := modelWithReviewAndSize(2)
	view := m.View()

	// All three panels should have their numbered titles
	if !strings.Contains(view, "[1]") {
		t.Error("view should contain section panel title [1]")
	}
	if !strings.Contains(view, "[2]") {
		t.Error("view should contain files panel title [2]")
	}
	if !strings.Contains(view, "[0]") {
		t.Error("view should contain diff panel title [0]")
	}
}

func TestView_SectionPanelShowsTitle1(t *testing.T) {
	m := modelWithReviewAndSize(1)
	view := m.View()

	if !strings.Contains(view, "[1]") {
		t.Error("section panel should show [1] in title")
	}
}

func TestView_FilesPanelShowsTitle2(t *testing.T) {
	m := modelWithReviewAndSize(1)
	view := m.View()

	if !strings.Contains(view, "[2]") {
		t.Error("files panel should show [2] in title")
	}
}

func TestView_DiffPanelShowsTitle0(t *testing.T) {
	m := modelWithReviewAndSize(1)
	view := m.View()

	if !strings.Contains(view, "[0]") {
		t.Error("diff panel should show [0] in title")
	}
}

// Files Panel View Tests

func modelWithFilesForView() tui.Model {
	m := tui.NewModel("/test/project")

	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	review := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test",
		Sections: []model.Section{
			{
				ID:        "1",
				Narrative: "Test section",
				Hunks: []model.Hunk{
					{File: "src/main.go", Diff: "+added"},
					{File: "src/util.go", Diff: "+added"},
					{File: "pkg/lib.go", Diff: "+added"},
				},
			},
		},
	}
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	return updated.(tui.Model)
}

func TestView_FilesPanelShowsFileTree(t *testing.T) {
	m := modelWithFilesForView()
	view := m.View()

	// Should show file names from hunks
	if !strings.Contains(view, "main.go") {
		t.Error("files panel should show main.go")
	}
	if !strings.Contains(view, "util.go") {
		t.Error("files panel should show util.go")
	}
	if !strings.Contains(view, "lib.go") {
		t.Error("files panel should show lib.go")
	}
}

func TestView_FilesPanelShowsDirectories(t *testing.T) {
	m := modelWithFilesForView()
	view := m.View()

	// Should show directory names
	if !strings.Contains(view, "src") {
		t.Error("files panel should show src directory")
	}
	if !strings.Contains(view, "pkg") {
		t.Error("files panel should show pkg directory")
	}
}

func TestView_ExpandedDirHasDownArrow(t *testing.T) {
	m := modelWithFilesForView()
	view := m.View()

	// Expanded directories should show ▼
	if !strings.Contains(view, "▼") {
		t.Error("expanded directories should show ▼ indicator")
	}
}

func TestView_CollapsedDirHasRightArrow(t *testing.T) {
	m := modelWithFilesForView()

	// Focus files panel
	focusMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")}
	updated, _ := m.Update(focusMsg)
	m = updated.(tui.Model)

	// Collapse first directory with Enter
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	updated, _ = m.Update(enterMsg)
	m = updated.(tui.Model)

	view := m.View()

	// Collapsed directory should show ▶
	if !strings.Contains(view, "▶") {
		t.Error("collapsed directories should show ▶ indicator")
	}
}

// Context-Sensitive Diff Display Tests

func modelWithMultipleFilesForDiff() tui.Model {
	m := tui.NewModel("/test/project")

	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	review := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test",
		Sections: []model.Section{
			{
				ID:        "1",
				Narrative: "Test section",
				Hunks: []model.Hunk{
					{File: "src/main.go", Diff: "+added in main"},
					{File: "src/util.go", Diff: "+added in util"},
					{File: "pkg/lib.go", Diff: "+added in lib"},
				},
			},
		},
	}
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	return updated.(tui.Model)
}

func TestView_DiffShowsAllFilesWhenSectionFocused(t *testing.T) {
	m := modelWithMultipleFilesForDiff()

	// Default focus is on Section panel
	view := m.View()

	// Should show all files' content
	if !strings.Contains(view, "main.go") {
		t.Error("diff should show main.go when section focused")
	}
	if !strings.Contains(view, "util.go") {
		t.Error("diff should show util.go when section focused")
	}
	if !strings.Contains(view, "lib.go") {
		t.Error("diff should show lib.go when section focused")
	}
}

func TestView_DiffShowsSingleFileWhenFilesFocused(t *testing.T) {
	m := modelWithMultipleFilesForDiff()

	// Focus files panel
	focusMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")}
	updated, _ := m.Update(focusMsg)
	m = updated.(tui.Model)

	// Navigate to a file (skip directory, go to first file)
	jMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	updated, _ = m.Update(jMsg)
	m = updated.(tui.Model)

	view := m.View()

	// Should show content from selected file
	// The first file after pkg directory should be lib.go
	if !strings.Contains(view, "added in lib") {
		t.Error("diff should show selected file's content when files focused")
	}
}

func TestView_DiffShowsAllFilesWhenDirectorySelected(t *testing.T) {
	m := modelWithMultipleFilesForDiff()

	// Focus files panel
	focusMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")}
	updated, _ := m.Update(focusMsg)
	m = updated.(tui.Model)

	// First item is a directory (pkg or src)
	view := m.View()

	// When directory is selected, should show all files under that directory
	// Since first directory is "pkg", should show lib.go content
	if !strings.Contains(view, "lib.go") {
		t.Error("diff should show files under selected directory")
	}
}

func TestView_DiffContextHeaderShowsAllFiles(t *testing.T) {
	m := modelWithMultipleFilesForDiff()

	// Section panel focused by default
	view := m.View()

	if !strings.Contains(view, "All files") {
		t.Error("diff header should show 'All files' when section focused")
	}
}

func TestView_DiffContextHeaderShowsFileName(t *testing.T) {
	m := modelWithMultipleFilesForDiff()

	// Focus files panel
	focusMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")}
	updated, _ := m.Update(focusMsg)
	m = updated.(tui.Model)

	// Navigate to a file (not directory)
	jMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	updated, _ = m.Update(jMsg)
	m = updated.(tui.Model)

	view := m.View()

	// Should show the file path in header
	if !strings.Contains(view, "pkg/lib.go") {
		t.Error("diff header should show file path when file is selected")
	}
}

func TestUpdate_FileSelectionUpdatesDiffContent(t *testing.T) {
	m := modelWithMultipleFilesForDiff()

	// Focus files panel
	focusMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")}
	updated, _ := m.Update(focusMsg)
	m = updated.(tui.Model)

	// Navigate to first file
	jMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	updated, _ = m.Update(jMsg)
	m = updated.(tui.Model)

	view1 := m.View()

	// Navigate to next item
	updated, _ = m.Update(jMsg)
	m = updated.(tui.Model)

	view2 := m.View()

	// Views should be different (different file selected)
	if view1 == view2 {
		t.Error("diff content should change when file selection changes")
	}
}

func TestUpdate_FocusSwitchUpdatesDiffContent(t *testing.T) {
	m := modelWithMultipleFilesForDiff()

	// Focus files panel and navigate to a file
	focusMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")}
	updated, _ := m.Update(focusMsg)
	m = updated.(tui.Model)

	jMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	updated, _ = m.Update(jMsg)
	m = updated.(tui.Model)

	viewFilesPanel := m.View()

	// Switch to section panel
	focusMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")}
	updated, _ = m.Update(focusMsg)
	m = updated.(tui.Model)

	viewSectionPanel := m.View()

	// Views should be different (all files vs single file)
	if viewFilesPanel == viewSectionPanel {
		t.Error("diff content should change when focus switches between panels")
	}
}

// Phase 4: Help Overlay Tests

func TestView_HelpOverlayShowsNewKeybindings(t *testing.T) {
	m := tui.NewModel("/test/project")

	// Initialize viewport
	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// Set a review
	review := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test",
		Sections:         []model.Section{{ID: "1", Narrative: "Section"}},
	}
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	// Toggle help on
	helpMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")}
	updated, _ = m.Update(helpMsg)
	m = updated.(tui.Model)

	view := m.View()

	// Check for panel focus keys
	if !strings.Contains(view, "h/l") {
		t.Error("help overlay should show h/l for panel cycling")
	}

	// Check for number keys
	if !strings.Contains(view, "0") && !strings.Contains(view, "1") && !strings.Contains(view, "2") {
		t.Error("help overlay should show 0-2 for panel jumping")
	}

	// Check for bounds navigation
	if !strings.Contains(view, "<") || !strings.Contains(view, ">") {
		t.Error("help overlay should show </> for bounds navigation")
	}

	// Check for page navigation
	if !strings.Contains(view, ",") || !strings.Contains(view, ".") {
		t.Error("help overlay should show ,/. for page navigation")
	}
}

// Phase 5: Position Indicators and Visual Polish Tests

func TestView_SectionPanelShowsPositionIndicator(t *testing.T) {
	m := tui.NewModel("/test/project")

	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// Create review with 5 sections
	sections := make([]model.Section, 5)
	for i := range 5 {
		sections[i] = model.Section{
			ID:        string(rune('1' + i)),
			Narrative: "Section " + string(rune('A'+i)),
			Hunks:     []model.Hunk{{File: "file.go", Diff: "+added"}},
		}
	}
	review := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test",
		Sections:         sections,
	}
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	// Navigate to section 2 (index 1)
	jMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	updated, _ = m.Update(jMsg)
	m = updated.(tui.Model)

	view := m.View()

	// Should show [2/5] position indicator
	if !strings.Contains(view, "[2/5]") {
		t.Error("section panel should show [2/5] position indicator")
	}
}

func TestView_FilesPanelShowsPositionIndicator(t *testing.T) {
	m := tui.NewModel("/test/project")

	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// Create review with multiple files
	hunks := make([]model.Hunk, 5)
	for i := range 5 {
		hunks[i] = model.Hunk{
			File: "file" + string(rune('a'+i)) + ".go",
			Diff: "+added",
		}
	}
	review := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test",
		Sections:         []model.Section{{ID: "1", Narrative: "Section", Hunks: hunks}},
	}
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	// Focus files panel
	focusMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")}
	updated, _ = m.Update(focusMsg)
	m = updated.(tui.Model)

	// Navigate to file 2
	jMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	updated, _ = m.Update(jMsg)
	m = updated.(tui.Model)

	view := m.View()

	// Should show "2 of 5" position indicator
	if !strings.Contains(view, "2 of 5") {
		t.Error("files panel should show '2 of 5' position indicator")
	}
}

func TestTruncatePath_ShortPathUnchanged(t *testing.T) {
	shortPath := "src/main.go"
	result := tui.TruncatePathMiddle(shortPath, 50)

	if result != shortPath {
		t.Errorf("TruncatePathMiddle(%q, 50) = %q, want %q", shortPath, result, shortPath)
	}
}

func TestTruncatePath_LongPathTruncatesMiddle(t *testing.T) {
	longPath := "src/components/auth/middleware/validators/token.go"
	result := tui.TruncatePathMiddle(longPath, 25)

	// Should contain ellipsis and preserve first and last parts
	if !strings.Contains(result, "...") {
		t.Errorf("TruncatePathMiddle(%q, 25) = %q, expected to contain '...'", longPath, result)
	}
	if !strings.Contains(result, "src") {
		t.Errorf("TruncatePathMiddle(%q, 25) = %q, expected to contain 'src'", longPath, result)
	}
	if !strings.Contains(result, "token.go") {
		t.Errorf("TruncatePathMiddle(%q, 25) = %q, expected to contain 'token.go'", longPath, result)
	}
}
