package app

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vivekananda-2201/ascii-font-generator-tui/internal/fonts"
)

func createTestModel() Model {
	fontList := []fonts.FontEntry{
		{Name: "standard", Family: "Standard", IsBuiltin: true},
		{Name: "larry3d", Family: "Larry", IsBuiltin: true},
		{Name: "ANSI Shadow", Family: "Ansi", Path: ""},
		{Name: "ANSI Regular", Family: "Ansi", Path: ""},
	}
	m := NewModel(fontList, false)
	// Simulate window size message
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	return updated.(Model)
}

func TestModelInitialization(t *testing.T) {
	m := createTestModel()

	if m.State != StateBrowse {
		t.Errorf("Expected initial state StateBrowse, got %v", m.State)
	}
	if !m.focusInput {
		t.Errorf("Expected focusInput to be true initially")
	}
	if len(m.filteredFonts) != 4 {
		t.Errorf("Expected 4 fonts, got %d", len(m.filteredFonts))
	}
	if m.cursor != 0 {
		t.Errorf("Expected cursor at 0, got %d", m.cursor)
	}

	view := m.View()
	if !strings.Contains(view, "Text Input") {
		t.Errorf("Expected view to contain 'Text Input'")
	}
	if !strings.Contains(view, "Live Preview") {
		t.Errorf("Expected view to contain 'Live Preview'")
	}
}

func TestFocusToggle(t *testing.T) {
	m := createTestModel()

	// Press Tab
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.focusInput {
		t.Errorf("Expected focusInput to be false after Tab")
	}

	// Press Tab again
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if !m.focusInput {
		t.Errorf("Expected focusInput to be true after second Tab")
	}
}

func TestFontNavigation(t *testing.T) {
	m := createTestModel()

	// Navigate down
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	if m.cursor != 1 {
		t.Errorf("Expected cursor at 1, got %d", m.cursor)
	}

	// Navigate down again
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	if m.cursor != 2 {
		t.Errorf("Expected cursor at 2, got %d", m.cursor)
	}

	// Navigate up
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(Model)
	if m.cursor != 1 {
		t.Errorf("Expected cursor back at 1, got %d", m.cursor)
	}
}

func TestSearchMode(t *testing.T) {
	m := createTestModel()

	// Press Ctrl+S
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = updated.(Model)
	if m.State != StateSearch {
		t.Errorf("Expected StateSearch after Ctrl+S, got %v", m.State)
	}

	// View should render search bar
	view := m.View()
	if !strings.Contains(view, "Search Font") {
		t.Errorf("Expected view to display 'Search Font'")
	}

	// Type search query "larry"
	for _, r := range "larry" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(Model)
	}

	if len(m.filteredFonts) == 0 {
		t.Fatalf("Expected fuzzy matches for 'larry', got 0")
	}
	if m.filteredFonts[0].Name != "larry3d" {
		t.Errorf("Expected top match to be 'larry3d', got %s", m.filteredFonts[0].Name)
	}

	// Press Enter to confirm search selection
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if m.State != StateBrowse {
		t.Errorf("Expected return to StateBrowse after Enter in search, got %v", m.State)
	}

	// Test Esc cancels search and restores all fonts
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = updated.(Model)
	if m.State != StateBrowse {
		t.Errorf("Expected StateBrowse after Esc in search")
	}
	if len(m.filteredFonts) != len(m.allFonts) {
		t.Errorf("Expected all fonts restored after search cancel")
	}
}

func TestVariantModalAndConfirmation(t *testing.T) {
	m := createTestModel()

	// Move cursor to "ANSI Shadow"
	m.cursor = 2
	m.updatePreview()

	// Press Enter to open variant modal
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)

	if m.State != StateVariant {
		t.Errorf("Expected StateVariant after Enter, got %v", m.State)
	}
	if len(m.variants) == 0 {
		t.Fatalf("Expected variants to be populated, got 0")
	}

	view := m.View()
	if !strings.Contains(view, "Select Size / Style Variant") {
		t.Errorf("Expected modal title in view")
	}

	// Navigate down in variant list
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(Model)
	if m.varCursor != 1 {
		t.Errorf("Expected varCursor at 1, got %d", m.varCursor)
	}

	// Finalize selection with Enter
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(Model)
	if !m.Confirmed {
		t.Errorf("Expected Confirmed to be true upon final Enter")
	}
	if m.Output == "" {
		t.Errorf("Expected non-empty output upon final confirmation")
	}
	if cmd == nil {
		t.Errorf("Expected tea.Quit command upon final confirmation")
	}
}

func TestHorizontalScroll(t *testing.T) {
	m := createTestModel()

	if m.xOffset != 0 {
		t.Errorf("Initial xOffset should be 0")
	}

	// Press Right arrow
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = updated.(Model)
	if m.xOffset <= 0 {
		t.Errorf("Expected xOffset to increase after Right key, got %d", m.xOffset)
	}

	// Press Left arrow
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = updated.(Model)
	if m.xOffset != 0 {
		t.Errorf("Expected xOffset to return to 0 after Left key, got %d", m.xOffset)
	}
}

func TestTerminalBounds(t *testing.T) {
	fontList, err := fonts.DiscoverFonts("")
	if err != nil {
		t.Fatalf("DiscoverFonts error: %v", err)
	}

	termH := 24
	termW := 80

	m := NewModel(fontList, false)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: termW, Height: termH})
	m = updated.(Model)

	for i := 0; i < len(fontList); i++ {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = updated.(Model)
		v := m.View()
		h := lipgloss.Height(v)
		w := lipgloss.Width(v)
		if h > termH {
			lines := strings.Split(v, "\n"); fmt.Printf("DEBUG: total lines=%d, top=%d\n", len(lines), h); t.Errorf("Font [%s] caused view height %d > terminal height %d", fontList[i].Name, h, termH)
			break
		}
		if w > termW {
			t.Errorf("Font [%s] caused view width %d > terminal width %d", fontList[i].Name, w, termW)
			break
		}
	}
}

func TestComponentHeights(t *testing.T) {
	fontList, _ := fonts.DiscoverFonts("")
	m := NewModel(fontList, false)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(Model)

	topBar := m.renderTopBar()
	tH := lipgloss.Height(topBar)
	tW := lipgloss.Width(topBar)
	fmt.Printf("DEBUG: topBar: h=%d, w=%d\n", tH, tW)

	leftPanel := m.renderFontList(22, 20)
	lH := lipgloss.Height(leftPanel)
	lW := lipgloss.Width(leftPanel)
	fmt.Printf("DEBUG: leftPanel: h=%d, w=%d\n", lH, lW)

	rightPanel := m.renderPreviewPanel(58, 20)
	rH := lipgloss.Height(rightPanel)
	rW := lipgloss.Width(rightPanel)
	fmt.Printf("DEBUG: rightPanel: h=%d, w=%d\n", rH, rW)

	mainPanels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
	mH := lipgloss.Height(mainPanels)
	mW := lipgloss.Width(mainPanels)
	fmt.Printf("DEBUG: mainPanels: h=%d, w=%d\n", mH, mW)

	statusBar := m.renderStatusBar()
	sH := lipgloss.Height(statusBar)
	sW := lipgloss.Width(statusBar)
	fmt.Printf("DEBUG: statusBar: h=%d, w=%d, content=%q\n", sH, sW, statusBar)

	total := lipgloss.JoinVertical(lipgloss.Left, topBar, mainPanels, statusBar)
	totH := lipgloss.Height(total)
	totW := lipgloss.Width(total)
	fmt.Printf("DEBUG: TOTAL: h=%d, w=%d (target: h=24, w=80)\n", totH, totW)
	fmt.Printf("DEBUG: Lines in total: %d\n", len(strings.Split(total, "\n")))
}

func TestTerminalBoundsWithLongTextAndVariousSizes(t *testing.T) {
	fontList, err := fonts.DiscoverFonts("")
	if err != nil {
		t.Fatalf("DiscoverFonts error: %v", err)
	}

	testSizes := [][2]int{
		{80, 24},
		{120, 35},
		{60, 16},
		{160, 45},
	}

	testTexts := []string{
		"",
		"A",
		"Hello World",
		"The quick brown fox jumps over the lazy dog 1234567890!@#$%^&*()",
		"LONG TEXT WITH MANY CHARACTERS THAT WOULD NORMALLY WRAP AND BREAK EVERYTHING IN THE TUI",
	}

	for _, size := range testSizes {
		termW, termH := size[0], size[1]
		m := NewModel(fontList, false)
		updated, _ := m.Update(tea.WindowSizeMsg{Width: termW, Height: termH})
		m = updated.(Model)

		for _, txt := range testTexts {
			m.textInput.SetValue(txt)
			m.updatePreview()

			// Check first 15 fonts
			for i := 0; i < 15 && i < len(fontList); i++ {
				updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
				m = updated.(Model)
				v := m.View()
				h := lipgloss.Height(v)
				w := lipgloss.Width(v)
				if h > termH {
					t.Fatalf("Text %q on font [%s] (term %dx%d) caused view height %d > %d",
						txt, fontList[i].Name, termW, termH, h, termH)
				}
				if w > termW {
					t.Fatalf("Text %q on font [%s] (term %dx%d) caused view width %d > %d",
						txt, fontList[i].Name, termW, termH, w, termW)
				}
			}
		}
	}
}

func TestNoBrokenANSIInViews(t *testing.T) {
	fontList, err := fonts.DiscoverFonts("")
	if err != nil {
		t.Fatalf("DiscoverFonts error: %v", err)
	}

	m := NewModel(fontList, false)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(Model)

	for i := 0; i < len(fontList); i++ {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = updated.(Model)

		v := m.View()
		for lineIdx, line := range strings.Split(v, "\n") {
			// Check for raw broken ANSI pattern like "[m" or "[0m" not preceded by ESC
			brokenAnsiRegex := regexp.MustCompile(`[^\x1b]\[[0-9;]*m(?:[^a-zA-Z]|$)`)
			if brokenAnsiRegex.MatchString(line) {
				t.Fatalf("Font [%s] has raw broken ANSI on line %d: %q", fontList[i].Name, lineIdx, line)
			}
			// Check for dangling ESC at end of line
			if strings.HasSuffix(line, "\x1b") {
				t.Fatalf("Font [%s] has dangling ESC on line %d: %q", fontList[i].Name, lineIdx, line)
			}
		}
	}
}
