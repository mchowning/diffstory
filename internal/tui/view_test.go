package tui_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mchowning/diffstory/internal/config"
	"github.com/mchowning/diffstory/internal/model"
	"github.com/mchowning/diffstory/internal/tui"
)

func TestView_EmptyStateContainsWorkingDirectory(t *testing.T) {
	m := tui.NewModel("/test/project", nil, nil, nil)
	view := m.View()

	if !strings.Contains(view, "/test/project") {
		t.Error("empty state view should contain working directory")
	}
}

func TestView_EmptyStateContainsQuitInstruction(t *testing.T) {
	m := tui.NewModel("/test/project", nil, nil, nil)
	view := m.View()

	if !strings.Contains(view, "q: quit") {
		t.Error("empty state view should contain 'q: quit'")
	}
}

func TestView_EmptyStateContainsServerInstructions(t *testing.T) {
	m := tui.NewModel("/test/project", nil, nil, nil)
	view := m.View()

	if !strings.Contains(view, "diffstory server") {
		t.Error("empty state view should contain server start instructions")
	}
	if !strings.Contains(view, "POST") {
		t.Error("empty state view should contain POST instruction")
	}
}

func modelWithReviewAndSize(numSections int) tui.Model {
	m := tui.NewModel("/test/project", nil, nil, nil)

	// Initialize viewport via WindowSizeMsg
	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// Set review
	sections := make([]model.Section, numSections)
	for i := range numSections {
		sections[i] = model.Section{
			ID:        string(rune('1' + i)),
			Title:     "Title " + string(rune('A'+i)),
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
	review := model.NewReviewWithSections("/test/project", "Test Review Title", sections)
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	return updated.(tui.Model)
}

func modelWithChapters() tui.Model {
	m := tui.NewModel("/test/project", nil, nil, nil)

	// Initialize viewport via WindowSizeMsg
	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// Create review with explicit chapters
	review := model.Review{
		WorkingDirectory: "/test/project",
		Title:            "Test Review",
		Chapters: []model.Chapter{
			{
				ID:    "ch1",
				Title: "Authentication",
				Sections: []model.Section{
					{
						ID:        "s1",
						Title:     "Add login endpoint",
						Narrative: "Implements POST /login with bcrypt hashing.",
						Hunks:     []model.Hunk{{File: "auth.go", Diff: "+code"}},
					},
					{
						ID:        "s2",
						Title:     "Add login tests",
						Narrative: "Tests the login endpoint.",
						Hunks:     []model.Hunk{{File: "auth_test.go", Diff: "+test"}},
					},
				},
			},
			{
				ID:    "ch2",
				Title: "Database",
				Sections: []model.Section{
					{
						ID:        "s3",
						Title:     "Add user migration",
						Narrative: "Creates users table with email column.",
						Hunks:     []model.Hunk{{File: "migrate.go", Diff: "+sql"}},
					},
				},
			},
		},
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

	// Section titles should be visible
	if !strings.Contains(view, "Title A") {
		t.Error("review state view should contain section A title")
	}
	if !strings.Contains(view, "Title B") {
		t.Error("review state view should contain section B title")
	}
}

func TestView_SelectedSectionHasPrefix(t *testing.T) {
	m := modelWithReviewAndSize(2)
	view := m.View()

	// First section should be selected with "  " prefix (showing title)
	if !strings.Contains(view, "  Title A") {
		t.Error("selected section should have '  ' prefix with title")
	}
}

func TestView_NonSelectedSectionHasSpacePrefix(t *testing.T) {
	m := modelWithReviewAndSize(2)
	view := m.View()

	// Second section should not be selected, should have "  " prefix (showing title)
	if !strings.Contains(view, "  Title B") {
		t.Error("non-selected section should have '  ' prefix with title")
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

func TestView_ChapterHeadersAppearInSectionPane(t *testing.T) {
	m := modelWithChapters()
	view := m.View()

	// Chapter headers should appear in the view
	if !strings.Contains(view, "Authentication") {
		t.Error("view should contain chapter header 'Authentication'")
	}
	if !strings.Contains(view, "Database") {
		t.Error("view should contain chapter header 'Database'")
	}
}

func TestView_NonSelectedSectionShowsOnlyTitle(t *testing.T) {
	m := modelWithChapters()
	view := m.View()

	// Non-selected sections should show title but NOT narrative
	// Section "Add login tests" (s2) is not selected (s1 is selected by default)
	if !strings.Contains(view, "Add login tests") {
		t.Error("view should contain non-selected section title 'Add login tests'")
	}
	if strings.Contains(view, "Tests the login endpoint") {
		t.Error("view should NOT contain narrative for non-selected section")
	}
}

func TestView_SelectedSectionDoesNotShowNarrativeInline(t *testing.T) {
	m := modelWithChapters()
	view := m.View()

	// The narrative prefix "  │ " should NOT appear in the sections panel
	// because narratives are now displayed in the Description panel instead
	if strings.Contains(view, "  │ ") {
		t.Error("sections panel should NOT show narrative prefix '  │ ' - narratives belong in Description panel")
	}
}

func TestView_DescriptionPanelShowsNarrative(t *testing.T) {
	m := modelWithChapters()
	view := m.View()

	// Description panel should exist and show the narrative for the selected section
	if !strings.Contains(view, "Description") {
		t.Error("view should contain Description panel title")
	}

	// The first section is selected by default, its narrative should appear
	// in the Description panel
	if !strings.Contains(view, "Implements POST /login") {
		t.Error("Description panel should contain narrative for selected section")
	}
}

func TestView_SelectedSectionShowsTitleAndNarrative(t *testing.T) {
	m := modelWithChapters()
	view := m.View()

	// First section is selected by default - should show both title and narrative
	if !strings.Contains(view, "Add login endpoint") {
		t.Error("view should contain selected section title 'Add login endpoint'")
	}
	if !strings.Contains(view, "Implements POST /login") {
		t.Error("view should contain narrative for selected section")
	}
}

func TestView_NotReadyShowsInitializing(t *testing.T) {
	m := tui.NewModel("/test/project", nil, nil, nil)
	// Set a review but don't send WindowSizeMsg (so viewport not initialized)
	review := model.NewReviewWithSections("/test/project", "Test", []model.Section{{ID: "1", Narrative: "Section"}})
	updated, _ := m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	view := m.View()

	if !strings.Contains(view, "Initializing") {
		t.Error("view should show 'Initializing' when viewport not ready")
	}
}

func TestView_SectionListDoesNotTruncateText(t *testing.T) {
	m := tui.NewModel("/test/project", nil, nil, nil)

	// Initialize viewport - width 120 gives section pane ~30 chars for wrapping test
	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// Create a review with a long narrative
	longNarrative := "This is a very long narrative that should wrap instead of being truncated with an ellipsis character"
	review := model.NewReviewWithSections("/test/project", "Test", []model.Section{
		{ID: "1", Narrative: longNarrative},
	})
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
	m := tui.NewModel("/test/project", nil, nil, nil)

	// Initialize viewport
	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// Set a review
	review := model.NewReviewWithSections("/test/project", "Test", []model.Section{{ID: "1", Narrative: "Section"}})
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
	m := tui.NewModel("/test/project", nil, nil, nil)

	// Initialize viewport
	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// Set a review
	review := model.NewReviewWithSections("/test/project", "Test", []model.Section{{ID: "1", Narrative: "Section"}})
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

	// Find positions of both section titles
	posA := strings.Index(view, "Title A")
	posB := strings.Index(view, "Title B")

	if posA == -1 || posB == -1 {
		t.Fatal("could not find both section titles in view")
	}

	// Extract text between sections - should include narrative for selected section
	// plus at least one newline before the next section
	between := view[posA+len("Title A") : posB]

	// Should have at least 1 newline (separating sections)
	newlineCount := strings.Count(between, "\n")
	if newlineCount < 1 {
		t.Errorf("expected at least 1 newline between sections, got %d", newlineCount)
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
	m := tui.NewModel("/test/project", nil, nil, nil)

	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	review := model.NewReviewWithSections("/test/project", "Test", []model.Section{
		{
			ID:        "1",
			Narrative: "Test section",
			Hunks: []model.Hunk{
				{File: "src/main.go", Diff: "+added"},
				{File: "src/util.go", Diff: "+added"},
				{File: "pkg/lib.go", Diff: "+added"},
			},
		},
	})
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
	m := tui.NewModel("/test/project", nil, nil, nil)

	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	review := model.NewReviewWithSections("/test/project", "Test", []model.Section{
		{
			ID:        "1",
			Narrative: "Test section",
			Hunks: []model.Hunk{
				{File: "src/main.go", Diff: "+added in main"},
				{File: "src/util.go", Diff: "+added in util"},
				{File: "pkg/lib.go", Diff: "+added in lib"},
			},
		},
	})
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

func TestView_DiffContextHeaderShowsSelectedFile(t *testing.T) {
	m := modelWithMultipleFilesForDiff()

	// Section panel focused by default, but header still shows selected file/dir
	// First item in flattened files is "pkg" directory
	view := m.View()

	if !strings.Contains(view, "Viewing: pkg/") {
		t.Error("diff header should show selected file/directory regardless of focus")
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

func TestView_FileSelectionControlsDiffRegardlessOfFocus(t *testing.T) {
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

	// Diff content should show the selected file regardless of focus
	// The selected file is pkg/lib.go, so both views should show "added in lib"
	if !strings.Contains(viewFilesPanel, "added in lib") {
		t.Error("diff should show selected file content when files panel focused")
	}
	if !strings.Contains(viewSectionPanel, "added in lib") {
		t.Error("diff should still show selected file content when section panel focused")
	}

	// Neither view should show content from other files
	if strings.Contains(viewSectionPanel, "added in main") {
		t.Error("diff should not show other files when a specific file is selected")
	}
	if strings.Contains(viewSectionPanel, "added in util") {
		t.Error("diff should not show other files when a specific file is selected")
	}
}

// Phase 4: Help Overlay Tests

func TestView_HelpOverlayShowsNewKeybindings(t *testing.T) {
	m := tui.NewModel("/test/project", nil, nil, nil)

	// Initialize viewport
	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// Set a review
	review := model.NewReviewWithSections("/test/project", "Test", []model.Section{{ID: "1", Narrative: "Section"}})
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
	m := tui.NewModel("/test/project", nil, nil, nil)

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
	review := model.NewReviewWithSections("/test/project", "Test", sections)
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

func TestView_SectionPaneScrollsWithOffset(t *testing.T) {
	m := tui.NewModel("/test/project", nil, nil, nil)

	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// Create review with sections that have unique titles and narratives
	// Titles appear in Sections panel, narratives appear in Description panel
	sections := make([]model.Section, 10)
	for i := range 10 {
		sections[i] = model.Section{
			ID:        string(rune('0' + i)),
			Title:     "SectionTitle" + string(rune('A'+i)),
			Narrative: "Narrative for section " + string(rune('A'+i)),
			Hunks:     []model.Hunk{{File: "file.go", Diff: "+added"}},
		}
	}
	review := model.NewReviewWithSections("/test/project", "Test", sections)
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	// Set scroll offset to 5 (skip first 5 sections)
	m = m.SetSectionScrollOffset(5)

	view := m.View()

	// Section F (index 5) should be visible in sections panel since scroll offset is 5
	if !strings.Contains(view, "SectionTitleF") {
		t.Error("section title at scroll offset (F) should be visible in sections panel")
	}

	// Section A's title (index 0) should NOT be visible in sections panel
	// since it's before the scroll offset
	if strings.Contains(view, "SectionTitleA") {
		t.Error("section title before scroll offset (A) should NOT be visible in sections panel")
	}

	// However, section A's narrative SHOULD be visible in Description panel
	// because section 0 is still selected (just scrolled off in sections panel)
	if !strings.Contains(view, "Narrative for section A") {
		t.Error("narrative for selected section (A) should be visible in Description panel")
	}
}

func TestView_FilesPaneScrollsWithOffset(t *testing.T) {
	m := tui.NewModel("/test/project", nil, nil, nil)

	// Use small height so not all files fit
	// Files panel height = (16 - 4) / 2 = 6, so only ~4 files visible
	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 16}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// Create review with many files
	hunks := make([]model.Hunk, 10)
	for i := range 10 {
		hunks[i] = model.Hunk{
			File: "scrollfile" + string(rune('a'+i)) + ".go",
			Diff: "+added",
		}
	}
	review := model.NewReviewWithSections("/test/project", "Test", []model.Section{
		{ID: "1", Narrative: "Test section", Hunks: hunks},
	})
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	// Switch to files panel and set scroll offset
	switchMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")}
	updated, _ = m.Update(switchMsg)
	m = updated.(tui.Model)

	m = m.SetFilesScrollOffset(5)

	view := m.View()

	// File at scroll offset (f) should be visible in the files panel
	// The file name appears with its .go extension in the files list
	if !strings.Contains(view, "scrollfilef.go") {
		t.Error("file at scroll offset (f) should be visible")
	}

	// File before scroll offset (a) appears in diff pane:
	// 1. In header "Viewing: scrollfilea.go"
	// 2. In diff content as file heading
	// But it should NOT appear in the files panel with the list prefix
	// The files panel uses "› " prefix for selected and "  " for others
	// If we see the file with the files panel prefix, scrolling is broken
	if strings.Contains(view, "  scrollfilea.go") || strings.Contains(view, "› scrollfilea.go") {
		t.Error("file before scroll offset (a) should NOT be visible in files panel")
	}
}

func TestView_FilesPanelShowsPositionIndicator(t *testing.T) {
	m := tui.NewModel("/test/project", nil, nil, nil)

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
	review := model.NewReviewWithSections("/test/project", "Test", []model.Section{{ID: "1", Narrative: "Section", Hunks: hunks}})
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

func TestView_FileHeadingAppearsOnceForMultipleHunks(t *testing.T) {
	m := tui.NewModel("/test/project", nil, nil, nil)

	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// Create review with multiple hunks for the same file
	review := model.NewReviewWithSections("/test/project", "Test", []model.Section{
		{
			ID:        "1",
			Narrative: "Test section",
			Hunks: []model.Hunk{
				{File: "src/main.go", Diff: "+first hunk"},
				{File: "src/main.go", Diff: "+second hunk"},
				{File: "src/main.go", Diff: "+third hunk"},
			},
		},
	})
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	view := m.View()

	// Count occurrences of the file name "src/main.go"
	count := strings.Count(view, "src/main.go")

	// File heading should appear only once, not three times
	if count != 1 {
		t.Errorf("file heading 'src/main.go' should appear once, but appeared %d times", count)
	}

	// All hunk content should still be present
	if !strings.Contains(view, "first hunk") {
		t.Error("diff should contain first hunk content")
	}
	if !strings.Contains(view, "second hunk") {
		t.Error("diff should contain second hunk content")
	}
	if !strings.Contains(view, "third hunk") {
		t.Error("diff should contain third hunk content")
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

// Timestamp Display Tests

func TestView_ReviewStateShowsTimestamp(t *testing.T) {
	m := tui.NewModel("/test/project", nil, nil, nil)

	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// Create review with CreatedAt set to 2 hours ago
	review := model.NewReviewWithSections("/test/project", "Test Review", []model.Section{{ID: "1", Narrative: "Section"}})
	review.CreatedAt = time.Now().Add(-2 * time.Hour)
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	view := m.View()

	if !strings.Contains(view, "Review generated") {
		t.Error("review view should contain 'Review generated' timestamp line")
	}
	if !strings.Contains(view, "hours ago") {
		t.Error("review view should show relative time like 'X hours ago'")
	}
}

func TestView_ReviewWithoutTimestampShowsNoCreatedLine(t *testing.T) {
	m := tui.NewModel("/test/project", nil, nil, nil)

	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// Create review WITHOUT CreatedAt (zero time)
	review := model.NewReviewWithSections("/test/project", "Test Review", []model.Section{{ID: "1", Narrative: "Section"}})
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	view := m.View()

	if strings.Contains(view, "Review generated") {
		t.Error("review view should NOT show 'Review generated' line when timestamp is zero")
	}
}

// Filter Indicator Tests

func TestView_ReviewStateShowsFilterIndicator(t *testing.T) {
	m := tui.NewModel("/test/project", nil, nil, nil)

	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	review := model.NewReviewWithSections("/test/project", "Test Review", []model.Section{{ID: "1", Narrative: "Section"}})
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	view := m.View()

	if !strings.Contains(view, "Diff filter:") {
		t.Error("review view should contain 'Diff filter:' indicator")
	}
}

func TestView_FilterIndicatorShowsCurrentLevel(t *testing.T) {
	cfg := &config.Config{DefaultFilterLevel: "high"}
	m := tui.NewModel("/test/project", cfg, nil, nil)

	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	review := model.NewReviewWithSections("/test/project", "Test Review", []model.Section{{ID: "1", Narrative: "Section"}})
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	view := m.View()

	// Filter indicator should show only current level
	if !strings.Contains(view, "Diff filter: High only") {
		t.Error("filter indicator should show 'Diff filter: High only'")
	}
}

func TestView_FooterShowsFilterShortcut(t *testing.T) {
	m := tui.NewModel("/test/project", nil, nil, nil)

	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	review := model.NewReviewWithSections("/test/project", "Test Review", []model.Section{{ID: "1", Narrative: "Section"}})
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	view := m.View()

	// Footer should show filter shortcut
	if !strings.Contains(view, "f: importance filter") {
		t.Error("footer should show 'f: importance filter' shortcut")
	}
}

// Helper to create a model with specific filter level
// Uses a single file with multiple hunks of different importance levels
func modelWithFilterLevel(filterLevel string) tui.Model {
	cfg := &config.Config{DefaultFilterLevel: filterLevel}
	m := tui.NewModel("/test/project", cfg, nil, nil)

	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// Use a single file with multiple hunks of different importance
	// This ensures the view shows all hunks for that file
	review := model.NewReviewWithSections("/test/project", "Test Review", []model.Section{
		{
			ID:        "1",
			Narrative: "Section with mixed importance",
			Hunks: []model.Hunk{
				{File: "main.go", Diff: "+low importance change", Importance: "low"},
				{File: "main.go", Diff: "+medium importance change", Importance: "medium"},
				{File: "main.go", Diff: "+high importance change", Importance: "high"},
			},
		},
	})
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	return updated.(tui.Model)
}

func TestView_LowFilterShowsAllHunks(t *testing.T) {
	m := modelWithFilterLevel("low")
	view := m.View()

	if !strings.Contains(view, "low importance change") {
		t.Error("low filter should show low importance hunks")
	}
	if !strings.Contains(view, "medium importance change") {
		t.Error("low filter should show medium importance hunks")
	}
	if !strings.Contains(view, "high importance change") {
		t.Error("low filter should show high importance hunks")
	}
}

func TestView_MediumFilterHidesLowImportance(t *testing.T) {
	m := modelWithFilterLevel("medium")
	view := m.View()

	if strings.Contains(view, "low importance change") {
		t.Error("medium filter should NOT show low importance hunks")
	}
	if !strings.Contains(view, "medium importance change") {
		t.Error("medium filter should show medium importance hunks")
	}
	if !strings.Contains(view, "high importance change") {
		t.Error("medium filter should show high importance hunks")
	}
}

func TestView_HighFilterShowsOnlyHigh(t *testing.T) {
	m := modelWithFilterLevel("high")
	view := m.View()

	if strings.Contains(view, "low importance change") {
		t.Error("high filter should NOT show low importance hunks")
	}
	if strings.Contains(view, "medium importance change") {
		t.Error("high filter should NOT show medium importance hunks")
	}
	if !strings.Contains(view, "high importance change") {
		t.Error("high filter should show high importance hunks")
	}
}

func TestView_ShowsFilteredIndicatorWhenAllHunksHidden(t *testing.T) {
	cfg := &config.Config{DefaultFilterLevel: "high"}
	m := tui.NewModel("/test/project", cfg, nil, nil)

	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// Section with only low importance hunks
	review := model.NewReviewWithSections("/test/project", "Test Review", []model.Section{
		{
			ID:        "1",
			Narrative: "All low importance section",
			Hunks: []model.Hunk{
				{File: "low.go", Diff: "+low", Importance: "low"},
			},
		},
	})
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	view := m.View()

	// Diff header should show (filtered) indicator when content is hidden
	// (Previously this was shown in section list, now shown in diff header)
	if !strings.Contains(view, "Diff (filtered)") {
		t.Error("diff header should show 'Diff (filtered)' when content is hidden by filter")
	}
}

func TestView_FilesPanelHidesFilesWithNoVisibleHunks(t *testing.T) {
	cfg := &config.Config{DefaultFilterLevel: "high"}
	m := tui.NewModel("/test/project", cfg, nil, nil)

	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	review := model.NewReviewWithSections("/test/project", "Test Review", []model.Section{
		{
			ID:        "1",
			Narrative: "Mixed section",
			Hunks: []model.Hunk{
				{File: "low.go", Diff: "+low", Importance: "low"},
				{File: "high.go", Diff: "+high", Importance: "high"},
			},
		},
	})
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	view := m.View()

	// With high filter, low.go should NOT appear in files panel
	// (But it might appear in diff pane context header)
	// We check the files panel specifically by looking for the file indicator pattern
	if strings.Contains(view, "› low.go") || strings.Contains(view, "  low.go") {
		t.Error("files panel should not show low.go when high filter is active")
	}
	// high.go should still be visible
	if !strings.Contains(view, "high.go") {
		t.Error("files panel should show high.go when high filter is active")
	}
}

// Phase 2: Filter Indicator Refactor Tests

func TestView_SectionListDoesNotShowFilteredIndicator(t *testing.T) {
	cfg := &config.Config{DefaultFilterLevel: "high"}
	m := tui.NewModel("/test/project", cfg, nil, nil)

	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// Section with only low importance hunks - all will be filtered
	review := model.NewReviewWithSections("/test/project", "Test Review", []model.Section{
		{
			ID:        "1",
			Narrative: "All low importance section",
			Hunks: []model.Hunk{
				{File: "low.go", Diff: "+low", Importance: "low"},
			},
		},
	})
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	view := m.View()

	// Section narrative should be visible but WITHOUT "(filtered)" suffix
	// The "(filtered)" indicator was removed from sections for cleaner UI
	if !strings.Contains(view, "All low importance section") {
		t.Error("section narrative should still be visible")
	}
	// Check that "(filtered)" does NOT appear directly after the section narrative
	// in the section list (it might appear elsewhere, e.g., diff header)
	if strings.Contains(view, "section (filtered)") {
		t.Error("section list should NOT show '(filtered)' suffix on section narratives")
	}
}

func TestView_DiffHeaderShowsFilteredWhenContentHidden(t *testing.T) {
	cfg := &config.Config{DefaultFilterLevel: "high"}
	m := tui.NewModel("/test/project", cfg, nil, nil)

	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// Create review with mixed importance - some will be hidden
	review := model.NewReviewWithSections("/test/project", "Test Review", []model.Section{
		{
			ID:        "1",
			Narrative: "Mixed section",
			Hunks: []model.Hunk{
				{File: "main.go", Diff: "+low change", Importance: "low"},
				{File: "main.go", Diff: "+high change", Importance: "high"},
			},
		},
	})
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	view := m.View()

	// Diff header should show "(filtered)" because low importance hunk is hidden
	if !strings.Contains(view, "Diff (filtered)") {
		t.Error("diff header should show 'Diff (filtered)' when content is hidden by filter")
	}
}

func TestView_DiffHeaderNoFilteredWhenNoContentHidden(t *testing.T) {
	cfg := &config.Config{DefaultFilterLevel: "low"} // Show all
	m := tui.NewModel("/test/project", cfg, nil, nil)

	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	review := model.NewReviewWithSections("/test/project", "Test Review", []model.Section{
		{
			ID:        "1",
			Narrative: "Test section",
			Hunks: []model.Hunk{
				{File: "main.go", Diff: "+change", Importance: "low"},
			},
		},
	})
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	view := m.View()

	// Should show "[0] Diff" but NOT "Diff (filtered)"
	if strings.Contains(view, "Diff (filtered)") {
		t.Error("diff header should NOT show '(filtered)' when no content is hidden")
	}
	if !strings.Contains(view, "[0] Diff") {
		t.Error("diff header should show '[0] Diff'")
	}
}

func TestView_DiffHeaderNoFilteredWhenFilterActiveButAllPassFilter(t *testing.T) {
	cfg := &config.Config{DefaultFilterLevel: "high"}
	m := tui.NewModel("/test/project", cfg, nil, nil)

	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// All hunks are high importance - none will be hidden
	review := model.NewReviewWithSections("/test/project", "Test Review", []model.Section{
		{
			ID:        "1",
			Narrative: "High importance section",
			Hunks: []model.Hunk{
				{File: "main.go", Diff: "+high1", Importance: "high"},
				{File: "main.go", Diff: "+high2", Importance: "high"},
			},
		},
	})
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	view := m.View()

	// Filter is active but no content is hidden, so no "(filtered)" indicator
	if strings.Contains(view, "Diff (filtered)") {
		t.Error("diff header should NOT show '(filtered)' when filter active but no content hidden")
	}
}

func TestView_DiffHeaderFilteredUpdatesWithFileNavigation(t *testing.T) {
	cfg := &config.Config{DefaultFilterLevel: "high"}
	m := tui.NewModel("/test/project", cfg, nil, nil)

	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// File with mixed importance vs file with only high
	review := model.NewReviewWithSections("/test/project", "Test Review", []model.Section{
		{
			ID:        "1",
			Narrative: "Test section",
			Hunks: []model.Hunk{
				{File: "all_high.go", Diff: "+high", Importance: "high"},
				{File: "mixed.go", Diff: "+low", Importance: "low"},
				{File: "mixed.go", Diff: "+high", Importance: "high"},
			},
		},
	})
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	// Focus files panel and navigate to all_high.go
	focusMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")}
	updated, _ = m.Update(focusMsg)
	m = updated.(tui.Model)

	// First file should be all_high.go (sorted alphabetically)
	view := m.View()

	// When viewing all_high.go, no content is hidden
	if strings.Contains(view, "Viewing: all_high.go") && strings.Contains(view, "Diff (filtered)") {
		t.Error("diff header should NOT show '(filtered)' when viewing file with no hidden content")
	}

	// Navigate to mixed.go
	jMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	updated, _ = m.Update(jMsg)
	m = updated.(tui.Model)

	view = m.View()

	// When viewing mixed.go, low importance content is hidden
	if strings.Contains(view, "Viewing: mixed.go") && !strings.Contains(view, "Diff (filtered)") {
		t.Error("diff header should show '(filtered)' when viewing file with hidden content")
	}
}

// Phase 3: Test Filter View Tests

// boolPtr returns a pointer to the given bool value
func boolPtr(b bool) *bool {
	return &b
}

func TestView_TestFilterHidesTestHunks(t *testing.T) {
	m := tui.NewModel("/test/project", nil, nil, nil)

	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	// Use same file for both hunks so they show in the same view
	review := model.NewReviewWithSections("/test/project", "Test Review", []model.Section{
		{
			ID:        "1",
			Narrative: "Section with tests",
			Hunks: []model.Hunk{
				{File: "main.go", Diff: "+production code change", IsTest: boolPtr(false)},
				{File: "main.go", Diff: "+test code change", IsTest: boolPtr(true)},
			},
		},
	})
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	// By default, TestFilter is All - both should be visible
	view := m.View()
	if !strings.Contains(view, "production code change") {
		t.Error("with TestFilterAll, production code should be visible")
	}
	if !strings.Contains(view, "test code change") {
		t.Error("with TestFilterAll, test code should be visible")
	}

	// Press t to switch to Excluding (hide tests)
	tMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")}
	updated, _ = m.Update(tMsg)
	m = updated.(tui.Model)

	view = m.View()
	if !strings.Contains(view, "production code change") {
		t.Error("with TestFilterExcluding, production code should still be visible")
	}
	if strings.Contains(view, "test code change") {
		t.Error("with TestFilterExcluding, test code should be hidden")
	}

	// Press t again to switch to Only (show only tests)
	updated, _ = m.Update(tMsg)
	m = updated.(tui.Model)

	view = m.View()
	if strings.Contains(view, "production code change") {
		t.Error("with TestFilterOnly, production code should be hidden")
	}
	if !strings.Contains(view, "test code change") {
		t.Error("with TestFilterOnly, test code should be visible")
	}
}

func TestView_FooterShowsTestFilterShortcut(t *testing.T) {
	m := tui.NewModel("/test/project", nil, nil, nil)

	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	review := model.NewReviewWithSections("/test/project", "Test Review", []model.Section{{ID: "1", Narrative: "Section"}})
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	view := m.View()

	// Footer should show test filter shortcut
	if !strings.Contains(view, "t: test filter") {
		t.Error("footer should show 't: test filter' shortcut")
	}
}

func TestView_FilterIndicatorShowsTestFilterWhenExcluding(t *testing.T) {
	m := tui.NewModel("/test/project", nil, nil, nil)

	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	review := model.NewReviewWithSections("/test/project", "Test Review", []model.Section{{ID: "1", Narrative: "Section"}})
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	// Set test filter to Excluding
	tMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")}
	updated, _ = m.Update(tMsg)
	m = updated.(tui.Model)

	view := m.View()

	// Filter indicator should show test filter status
	if !strings.Contains(view, "Excluding tests") {
		t.Error("filter indicator should show 'Excluding tests' when test filter is active")
	}
}

func TestView_FilterIndicatorShowsTestFilterWhenOnly(t *testing.T) {
	m := tui.NewModel("/test/project", nil, nil, nil)

	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	review := model.NewReviewWithSections("/test/project", "Test Review", []model.Section{{ID: "1", Narrative: "Section"}})
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	// Set test filter to Only (press t twice: All -> Excluding -> Only)
	tMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")}
	updated, _ = m.Update(tMsg)
	m = updated.(tui.Model)
	updated, _ = m.Update(tMsg)
	m = updated.(tui.Model)

	view := m.View()

	// Filter indicator should show test filter status
	if !strings.Contains(view, "Tests only") {
		t.Error("filter indicator should show 'Tests only' when test filter is set to Only")
	}
}

func TestView_CompoundFilteringAppliesBothFilters(t *testing.T) {
	cfg := &config.Config{DefaultFilterLevel: "high"}
	m := tui.NewModel("/test/project", cfg, nil, nil)

	sizeMsg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updated, _ := m.Update(sizeMsg)
	m = updated.(tui.Model)

	review := model.NewReviewWithSections("/test/project", "Test Review", []model.Section{
		{
			ID:        "1",
			Narrative: "Section",
			Hunks: []model.Hunk{
				{File: "main.go", Diff: "+high prod", Importance: "high", IsTest: boolPtr(false)},
				{File: "main.go", Diff: "+low prod", Importance: "low", IsTest: boolPtr(false)},
				{File: "main_test.go", Diff: "+high test", Importance: "high", IsTest: boolPtr(true)},
				{File: "main_test.go", Diff: "+low test", Importance: "low", IsTest: boolPtr(true)},
			},
		},
	})
	updated, _ = m.Update(tui.ReviewReceivedMsg{Review: review})
	m = updated.(tui.Model)

	// With high importance filter + test filter excluding:
	// Only high importance, non-test hunks should show
	tMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")}
	updated, _ = m.Update(tMsg)
	m = updated.(tui.Model)

	view := m.View()

	if !strings.Contains(view, "high prod") {
		t.Error("compound filter should show high importance production code")
	}
	if strings.Contains(view, "low prod") {
		t.Error("compound filter should hide low importance production code")
	}
	if strings.Contains(view, "high test") {
		t.Error("compound filter should hide test code (even high importance)")
	}
	if strings.Contains(view, "low test") {
		t.Error("compound filter should hide low importance test code")
	}
}
