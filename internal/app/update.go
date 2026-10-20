package app

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
	"github.com/sahilm/fuzzy"
	"github.com/vivekananda-2201/ascii-font-generator-tui/internal/clipboard"
	"github.com/vivekananda-2201/ascii-font-generator-tui/internal/fonts"
)

// Update handles incoming messages and events in the Elm architecture.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Strict layout arithmetic ensuring total height == m.height
		topBarHeight := 3
		statusBarHeight := 1
		availableHeight := m.height - topBarHeight - statusBarHeight
		if availableHeight < 4 {
			availableHeight = 4
		}

		leftWidth := int(float64(m.width) * 0.25)
		if leftWidth < 24 {
			leftWidth = 24
		}
		if leftWidth > 46 {
			leftWidth = 46
		}
		if leftWidth > m.width-30 && m.width > 35 {
			leftWidth = m.width - 30
		}

		rightWidth := m.width - leftWidth
		if rightWidth < 10 {
			rightWidth = 10
		}

		// Inner viewport bounds
		innerViewportHeight := availableHeight - 3
		if innerViewportHeight < 1 {
			innerViewportHeight = 1
		}
		innerViewportWidth := rightWidth - 2
		if innerViewportWidth < 10 {
			innerViewportWidth = 10
		}

		if !m.ready {
			m.viewport = viewport.New(innerViewportWidth, innerViewportHeight)
			m.varViewport = viewport.New(60, 6)
			m.ready = true
		} else {
			m.viewport.Width = innerViewportWidth
			m.viewport.Height = innerViewportHeight
		}

		m.textInput.Width = m.width - 28
		m.updatePreview()

	case tea.KeyMsg:
		// Global hotkeys
		switch msg.String() {
		case "ctrl+c":
			m.Confirmed = false
			return m, tea.Quit

		case "ctrl+y":
			// Copy active preview to clipboard
			artToCopy := m.renderedArt
			if m.State == StateVariant && m.varPreviewText != "" {
				artToCopy = m.varPreviewText
			}
			if err := clipboard.Copy(artToCopy); err == nil {
				m.StatusMessage = "✓ Copied to clipboard!"
			} else {
				m.StatusMessage = "✗ Clipboard copy failed"
			}
			return m, nil
		}

		// State-specific key handling
		switch m.State {
		case StateBrowse:
			return m.updateBrowse(msg)
		case StateSearch:
			return m.updateSearch(msg)
		case StateVariant:
			return m.updateVariant(msg)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) updateBrowse(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.Confirmed = false
		return m, tea.Quit

	case "tab":
		m.focusInput = !m.focusInput
		if m.focusInput {
			m.textInput.Focus()
		} else {
			m.textInput.Blur()
		}
		return m, nil

	case "ctrl+s":
		// Enter search mode
		m.State = StateSearch
		m.prevCursor = m.cursor
		if len(m.filteredFonts) > m.cursor {
			m.prevFontName = m.filteredFonts[m.cursor].Name
		}
		m.searchInput.Reset()
		m.searchInput.Focus()
		return m, nil

	case "enter":
		if len(m.filteredFonts) == 0 {
			return m, nil
		}
		selected := m.filteredFonts[m.cursor]
		m.selectedFont = selected
		m.variants = fonts.GetVariants(selected, m.allFonts)
		m.varCursor = 0
		m.varOffset = 0
		m.varXOffset = 0
		m.varYOffset = 0
		m.State = StateVariant
		m.updateVariantPreview()
		return m, nil

	case "up":
		if m.cursor > 0 {
			m.cursor--
			m.xOffset = 0
			m.viewport.SetYOffset(0)
			m.updatePreview()
		}
		return m, nil

	case "down":
		if m.cursor < len(m.filteredFonts)-1 {
			m.cursor++
			m.xOffset = 0
			m.viewport.SetYOffset(0)
			m.updatePreview()
		}
		return m, nil

	case "pgup":
		if !m.focusInput {
			// Scroll viewport up when font list is focused
			m.viewport.LineUp(3)
		} else {
			m.cursor -= 10
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.xOffset = 0
			m.viewport.SetYOffset(0)
			m.updatePreview()
		}
		return m, nil

	case "pgdown":
		if !m.focusInput {
			// Scroll viewport down when font list is focused
			m.viewport.LineDown(3)
		} else {
			m.cursor += 10
			if m.cursor >= len(m.filteredFonts) {
				m.cursor = len(m.filteredFonts) - 1
			}
			m.xOffset = 0
			m.viewport.SetYOffset(0)
			m.updatePreview()
		}
		return m, nil

	case "[":
		// Scroll preview art up
		m.viewport.LineUp(2)
		return m, nil

	case "]":
		// Scroll preview art down
		m.viewport.LineDown(2)
		return m, nil

	case "left":
		if m.xOffset > 0 {
			m.xOffset = max(0, m.xOffset-4)
			m.refreshViewportContent()
		}
		return m, nil

	case "right":
		m.xOffset += 4
		m.refreshViewportContent()
		return m, nil

	default:
		if m.focusInput {
			var cmd tea.Cmd
			prevVal := m.textInput.Value()
			m.textInput, cmd = m.textInput.Update(msg)
			if m.textInput.Value() != prevVal {
				m.xOffset = 0
				m.viewport.SetYOffset(0)
				m.updatePreview()
			}
			return m, cmd
		}
	}

	return m, nil
}

func (m Model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel search and restore previous state
		m.State = StateBrowse
		m.filteredFonts = m.allFonts
		m.cursor = m.prevCursor
		if m.cursor >= len(m.filteredFonts) {
			m.cursor = 0
		}
		m.xOffset = 0
		m.viewport.SetYOffset(0)
		m.updatePreview()
		if m.focusInput {
			m.textInput.Focus()
		}
		return m, nil

	case "enter":
		// Confirm top match
		m.State = StateBrowse
		if m.cursor >= len(m.filteredFonts) {
			m.cursor = 0
		}
		m.xOffset = 0
		m.viewport.SetYOffset(0)
		m.updatePreview()
		if m.focusInput {
			m.textInput.Focus()
		}
		return m, nil

	case "up":
		if m.cursor > 0 {
			m.cursor--
			m.xOffset = 0
			m.viewport.SetYOffset(0)
			m.updatePreview()
		}
		return m, nil

	case "down":
		if m.cursor < len(m.filteredFonts)-1 {
			m.cursor++
			m.xOffset = 0
			m.viewport.SetYOffset(0)
			m.updatePreview()
		}
		return m, nil

	default:
		var cmd tea.Cmd
		prevVal := m.searchInput.Value()
		m.searchInput, cmd = m.searchInput.Update(msg)
		newVal := m.searchInput.Value()

		if newVal != prevVal {
			if strings.TrimSpace(newVal) == "" {
				m.filteredFonts = m.allFonts
				m.cursor = 0
			} else {
				matches := fuzzy.FindFrom(newVal, fontNameSource(m.allFonts))
				m.filteredFonts = make([]fonts.FontEntry, len(matches))
				for i, match := range matches {
					m.filteredFonts[i] = m.allFonts[match.Index]
				}
				m.cursor = 0
			}
			m.xOffset = 0
			m.viewport.SetYOffset(0)
			m.updatePreview()
		}
		return m, cmd
	}
}

func (m Model) updateVariant(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	slots := m.visibleVariantSlots()
	switch msg.String() {
	case "esc":
		// Return to browse mode
		m.State = StateBrowse
		m.updatePreview()
		return m, nil

	case "up", "k":
		if m.varCursor > 0 {
			m.varCursor--
			m.varXOffset = 0
			m.varYOffset = 0
			m.adjustVarOffset(slots)
			m.updateVariantPreview()
		}
		return m, nil

	case "down", "j":
		if m.varCursor < len(m.variants)-1 {
			m.varCursor++
			m.varXOffset = 0
			m.varYOffset = 0
			m.adjustVarOffset(slots)
			m.updateVariantPreview()
		}
		return m, nil

	case "left", "h":
		if m.varXOffset > 0 {
			m.varXOffset -= 4
			if m.varXOffset < 0 {
				m.varXOffset = 0
			}
		}
		return m, nil

	case "right", "l":
		m.varXOffset += 4
		return m, nil

	case "pgup":
		if m.varYOffset > 0 {
			m.varYOffset -= 2
			if m.varYOffset < 0 {
				m.varYOffset = 0
			}
		}
		return m, nil

	case "pgdown":
		m.varYOffset += 2
		return m, nil

	case "ctrl+y":
		if m.varPreviewText != "" {
			err := clipboard.Copy(m.varPreviewText)
			if err != nil {
				m.StatusMessage = "Copy failed: " + err.Error()
			} else {
				m.StatusMessage = "Variant art copied to clipboard!"
			}
		}
		return m, nil

	case "enter":
		// Final confirmation!
		m.Confirmed = true
		m.Output = m.varPreviewText
		return m, tea.Quit
	}

	return m, nil
}

func (m *Model) updatePreview() {
	if len(m.filteredFonts) == 0 {
		m.renderedArt = ""
		m.artWidth = 0
		m.renderedLines = 0
		m.viewport.SetContent("No matching fonts.")
		return
	}

	if m.cursor >= len(m.filteredFonts) {
		m.cursor = len(m.filteredFonts) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}

	font := m.filteredFonts[m.cursor]
	m.renderer.EnsureFontLoaded(font.Path, font.Name)

	out, err := m.renderer.Render(font.Name, m.textInput.Value(), fonts.LayoutDefault)
	if err != nil {
		out = "Error rendering: " + err.Error()
	}

	m.renderedArt = out
	m.artWidth = computeMaxLineWidth(out)
	m.renderedLines = len(strings.Split(strings.TrimRight(out, "\n"), "\n"))
	m.refreshViewportContent()
}

func (m *Model) refreshViewportContent() {
	maxW := m.viewport.Width
	if maxW <= 0 {
		leftW := int(float64(m.width) * 0.25)
		if leftW < 24 {
			leftW = 24
		}
		if leftW > 46 {
			leftW = 46
		}
		maxW = m.width - leftW - 4
		if maxW <= 0 {
			maxW = 40
		}
	}

	lines := strings.Split(m.renderedArt, "\n")
	sliced := make([]string, len(lines))
	for i, line := range lines {
		// Slice strictly by display width so lines NEVER wrap in Lip Gloss
		sliced[i] = sliceLineByWidth(line, m.xOffset, maxW)
	}

	m.viewport.SetContent(strings.Join(sliced, "\n"))
}

func (m *Model) updateVariantPreview() {
	if len(m.variants) == 0 {
		m.varPreviewText = ""
		return
	}
	if m.varCursor >= len(m.variants) {
		m.varCursor = len(m.variants) - 1
	}
	v := m.variants[m.varCursor]
	m.renderer.EnsureFontLoaded(v.FontPath, v.FontName)

	out, err := m.renderer.Render(v.FontName, m.textInput.Value(), v.Layout)
	if err != nil {
		out = "Error rendering: " + err.Error()
	}
	m.varPreviewText = out
}

func (m *Model) adjustListOffset(visibleHeight int) {
	if visibleHeight <= 0 {
		return
	}
	if m.cursor < m.listOffset {
		m.listOffset = m.cursor
	} else if m.cursor >= m.listOffset+visibleHeight {
		m.listOffset = m.cursor - visibleHeight + 1
	}
	if m.listOffset < 0 {
		m.listOffset = 0
	}
}

func (m *Model) adjustVarOffset(visibleCount int) {
	if visibleCount <= 0 {
		return
	}
	if m.varCursor < m.varOffset {
		m.varOffset = m.varCursor
	} else if m.varCursor >= m.varOffset+visibleCount {
		m.varOffset = m.varCursor - visibleCount + 1
	}
	if m.varOffset < 0 {
		m.varOffset = 0
	}
}

func computeMaxLineWidth(s string) int {
	maxW := 0
	for _, line := range strings.Split(s, "\n") {
		w := runewidth.StringWidth(line)
		if w > maxW {
			maxW = w
		}
	}
	return maxW
}

// sliceLineByWidth slices a string starting from visual column xOffset up to maxWidth columns.
// This guarantees that the returned string will NEVER wrap inside a box of maxWidth visual width.
func sliceLineByWidth(line string, xOffset int, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	runes := []rune(line)
	if len(runes) == 0 {
		return ""
	}

	// Advance through xOffset visual columns
	curCol := 0
	startIdx := len(runes)
	for i, r := range runes {
		w := runewidth.RuneWidth(r)
		if curCol >= xOffset {
			startIdx = i
			break
		}
		curCol += w
	}

	if startIdx >= len(runes) {
		return ""
	}

	// Collect up to maxWidth visual columns
	curWidth := 0
	endIdx := startIdx
	for i := startIdx; i < len(runes); i++ {
		w := runewidth.RuneWidth(runes[i])
		if curWidth+w > maxWidth {
			break
		}
		curWidth += w
		endIdx = i + 1
	}

	return string(runes[startIdx:endIdx])
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
