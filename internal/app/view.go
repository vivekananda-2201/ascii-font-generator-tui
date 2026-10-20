package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// View renders the TUI layout based on the current state and dimensions.
func (m Model) View() string {
	if !m.ready || m.width < 10 || m.height < 5 {
		return "agen: terminal too small or initializing..."
	}

	switch m.State {
	case StateVariant:
		modal := m.renderVariantModal()
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal)

	case StateSearch:
		return m.renderSearchLayout()

	default:
		return m.renderMainLayout()
	}
}

func (m Model) renderMainLayout() string {
	topBar := m.renderTopBar()
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

	leftPanel := m.renderFontList(leftWidth, availableHeight)
	rightPanel := m.renderPreviewPanel(rightWidth, availableHeight)

	mainPanels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
	statusBar := m.renderStatusBar()

	return lipgloss.JoinVertical(lipgloss.Left, topBar, mainPanels, statusBar)
}

func (m Model) renderSearchLayout() string {
	searchBar := m.renderSearchBar()
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

	leftPanel := m.renderFontList(leftWidth, availableHeight)
	rightPanel := m.renderPreviewPanel(rightWidth, availableHeight)

	mainPanels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
	statusBar := m.renderStatusBar()

	return lipgloss.JoinVertical(lipgloss.Left, searchBar, mainPanels, statusBar)
}

func (m Model) renderTopBar() string {
	boxStyle := inputBoxUnfocused
	label := titleStyle.Render("Text Input:")
	if m.focusInput {
		boxStyle = inputBoxFocused
		label = activeTitleStyle.Render("Text Input (Focused):")
	}

	content := fmt.Sprintf("%s %s", label, m.textInput.View())
	// In Lip Gloss, Width(w) sets INNER width. Border adds 2.
	return boxStyle.Width(m.width - 2).Height(1).Render(content)
}

func (m Model) renderSearchBar() string {
	label := activeTitleStyle.Render("🔍 Search Font:")
	countInfo := lipgloss.NewStyle().Foreground(dimColor).Render(
		fmt.Sprintf(" [%d matches]", len(m.filteredFonts)),
	)
	content := fmt.Sprintf("%s %s%s", label, m.searchInput.View(), countInfo)
	return searchBoxStyle.Width(m.width - 2).Height(1).Render(content)
}

func (m Model) renderFontList(outerWidth, outerHeight int) string {
	innerWidth := outerWidth - 2
	innerHeight := outerHeight - 2
	if innerWidth < 10 {
		innerWidth = 10
	}
	if innerHeight < 2 {
		innerHeight = 2
	}

	// 1 line for header, remaining for items
	itemSlots := innerHeight - 1
	if itemSlots < 1 {
		itemSlots = 1
	}

	temp := m
	temp.adjustListOffset(itemSlots)
	offset := temp.listOffset

	headerText := fmt.Sprintf("Fonts (%d)", len(m.filteredFonts))
	maxHeaderLen := innerWidth - 2
	if len([]rune(headerText)) > maxHeaderLen && maxHeaderLen > 3 {
		headerText = string([]rune(headerText)[:maxHeaderLen-1]) + "…"
	}
	header := titleStyle.Render(headerText)

	lines := make([]string, 0, innerHeight)
	lines = append(lines, header)

	if len(m.filteredFonts) == 0 {
		lines = append(lines, itemNormalStyle.Render("No fonts found"))
	} else {
		end := offset + itemSlots
		if end > len(m.filteredFonts) {
			end = len(m.filteredFonts)
		}

		for i := offset; i < end; i++ {
			font := m.filteredFonts[i]
			displayName := font.Name

			// Truncate plain text BEFORE styling! Never slice styled strings!
			maxNameLen := innerWidth - 4
			if len([]rune(displayName)) > maxNameLen && maxNameLen > 3 {
				displayName = string([]rune(displayName)[:maxNameLen-1]) + "…"
			}

			if i == m.cursor {
				lines = append(lines, itemSelectedStyle.Render("▸ "+displayName))
			} else {
				lines = append(lines, itemNormalStyle.Render("  "+displayName))
			}
		}
	}

	// Pad with empty lines if needed to match exact inner height
	for len(lines) < innerHeight {
		lines = append(lines, "")
	}
	if len(lines) > innerHeight {
		lines = lines[:innerHeight]
	}

	body := strings.Join(lines, "\n")

	style := panelUnfocused
	if !m.focusInput && m.State == StateBrowse {
		style = panelFocused
	}

	return style.Width(innerWidth).Height(innerHeight).Render(body)
}

func (m Model) renderPreviewPanel(outerWidth, outerHeight int) string {
	innerWidth := outerWidth - 2
	innerHeight := outerHeight - 2
	if innerWidth < 10 {
		innerWidth = 10
	}
	if innerHeight < 2 {
		innerHeight = 2
	}

	fontName := "None"
	if len(m.filteredFonts) > 0 && m.cursor < len(m.filteredFonts) {
		fontName = m.filteredFonts[m.cursor].Name
	}

	title := fmt.Sprintf("Live Preview — [%s]", fontName)
	if m.artWidth > innerWidth {
		title += fmt.Sprintf(" [W:%d|←/→ pan:%d]", m.artWidth, m.xOffset)
	}
	if m.renderedLines > innerHeight-1 {
		title += fmt.Sprintf(" [H:%d|PgUp/Dn]", m.renderedLines)
	}

	// Truncate title plain text BEFORE styling
	maxTitleLen := innerWidth - 2
	if len([]rune(title)) > maxTitleLen && maxTitleLen > 3 {
		title = string([]rune(title)[:maxTitleLen-1]) + "…"
	}

	var titleHeader string
	if !m.focusInput {
		titleHeader = activeTitleStyle.Render(title)
	} else {
		titleHeader = titleStyle.Render(title)
	}

	vpHeight := innerHeight - 1
	if vpHeight < 1 {
		vpHeight = 1
	}

	vpContent := m.viewport.View()
	vpLines := strings.Split(vpContent, "\n")

	// Calculate vertical breathing room / padding when art is smaller than viewport height
	// so the ASCII art sits elegantly centered in the large preview pane rather than glued to the ceiling!
	topPad := 0
	if m.viewport.YOffset == 0 && m.renderedLines > 0 && m.renderedLines < vpHeight {
		topPad = (vpHeight - m.renderedLines) / 3
	}

	lines := make([]string, 0, innerHeight)
	lines = append(lines, titleHeader)

	// Add vertical padding lines if art is shorter than pane
	for p := 0; p < topPad && len(lines) < innerHeight; p++ {
		lines = append(lines, "")
	}

	for i := 0; i < len(vpLines) && len(lines) < innerHeight; i++ {
		lines = append(lines, vpLines[i])
	}

	// Fill remaining space to match exact innerHeight
	for len(lines) < innerHeight {
		lines = append(lines, "")
	}
	if len(lines) > innerHeight {
		lines = lines[:innerHeight]
	}

	body := strings.Join(lines[:innerHeight], "\n")

	panelStyle := panelUnfocused
	if !m.focusInput {
		panelStyle = panelFocused
	}

	return panelStyle.Width(innerWidth).Height(innerHeight).Render(body)
}

func (m Model) renderStatusBar() string {
	var keys []string

	if m.State == StateSearch {
		keys = append(keys,
			statusKeyStyle.Render("Enter")+" "+statusDescStyle.Render("Select"),
			statusKeyStyle.Render("Esc")+" "+statusDescStyle.Render("Cancel"),
			statusKeyStyle.Render("↑/↓")+" "+statusDescStyle.Render("Navigate"),
		)
	} else {
		if m.width >= 90 {
			keys = append(keys,
				statusKeyStyle.Render("↑/↓")+" "+statusDescStyle.Render("Navigate"),
				statusKeyStyle.Render("Tab")+" "+statusDescStyle.Render("Focus"),
				statusKeyStyle.Render("^S")+" "+statusDescStyle.Render("Search"),
				statusKeyStyle.Render("Enter")+" "+statusDescStyle.Render("Variants"),
				statusKeyStyle.Render("^Y")+" "+statusDescStyle.Render("Copy"),
				statusKeyStyle.Render("Esc")+" "+statusDescStyle.Render("Quit"),
			)
		} else if m.width >= 70 {
			keys = append(keys,
				statusKeyStyle.Render("↑/↓")+" "+statusDescStyle.Render("Fonts"),
				statusKeyStyle.Render("Tab")+" "+statusDescStyle.Render("Focus"),
				statusKeyStyle.Render("^S")+" "+statusDescStyle.Render("Search"),
				statusKeyStyle.Render("Enter")+" "+statusDescStyle.Render("Select"),
				statusKeyStyle.Render("Esc")+" "+statusDescStyle.Render("Quit"),
			)
		} else {
			keys = append(keys,
				statusKeyStyle.Render("↑/↓")+" "+statusDescStyle.Render("Fonts"),
				statusKeyStyle.Render("^S")+" "+statusDescStyle.Render("Search"),
				statusKeyStyle.Render("Enter")+" "+statusDescStyle.Render("OK"),
			)
		}
	}

	left := strings.Join(keys, statusSepStyle.String())

	var right string
	if m.StatusMessage != "" {
		right = statusKeyStyle.Render(m.StatusMessage)
	} else if m.width >= 60 {
		right = statusDescStyle.Render("agen v1.0")
	}

	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	gap := m.width - leftWidth - rightWidth
	if gap < 0 {
		gap = 0
		right = ""
	}

	content := left + strings.Repeat(" ", gap) + right
	return statusBarStyle.Width(m.width).Height(1).Render(content)
}

func (m Model) renderVariantModal() string {
	// Responsive modal sizing:
	modalOuterW := int(float64(m.width) * 0.75)
	if modalOuterW > 100 {
		modalOuterW = 100
	}
	if modalOuterW < 60 {
		modalOuterW = 60
	}
	if modalOuterW > m.width-2 {
		modalOuterW = m.width - 2
	}
	if modalOuterW < 20 {
		modalOuterW = 20
	}

	// Modal height: scale with terminal height
	modalOuterH := m.height - 2
	if modalOuterH > 28 {
		modalOuterH = 28
	}
	if modalOuterH > m.height {
		modalOuterH = m.height
	}
	if modalOuterH < 10 {
		modalOuterH = 10
	}

	// modalBoxStyle decoration:
	// Border: doubleBorder (2 horizontal, 2 vertical)
	// Padding: (1, 2) -> (4 horizontal, 2 vertical)
	// Available content width and height inside modalBoxStyle:
	contentW := modalOuterW - 6
	if contentW < 14 {
		contentW = 14
	}
	contentH := modalOuterH - 4
	if contentH < 6 {
		contentH = 6
	}

	maxVisibleVars := m.visibleVariantSlots()

	// Non-preview rows: title (1) + subtitle (1) + variants (maxVisibleVars) + divider (1) + footer (1)
	nonPreviewH := 1 + 1 + maxVisibleVars + 1 + 1
	previewOuterH := contentH - nonPreviewH
	if previewOuterH < 3 {
		previewOuterH = 3
	}

	// Re-adjust contentH so contentH ALWAYS exactly equals nonPreviewH + previewOuterH
	contentH = nonPreviewH + previewOuterH
	modalOuterH = contentH + 4
	if modalOuterH > m.height && m.height >= 8 {
		modalOuterH = m.height
		contentH = modalOuterH - 4
		if contentH < 4 {
			contentH = 4
		}
		previewOuterH = contentH - nonPreviewH
		if previewOuterH < 1 {
			previewOuterH = 1
		}
	}

	// Title
	titleText := "═══ Select Size / Style Variant ═══"
	title := modalTitleStyle.Width(contentW).Render(titleText)

	// Subtitle
	subText := fmt.Sprintf("Family: %s • Enter: Confirm, Esc: Back", m.selectedFont.Family)
	if runewidth.StringWidth(subText) > contentW {
		subText = sliceLineByWidth(subText, 0, contentW-1) + "…"
	}
	subtitle := lipgloss.NewStyle().Foreground(dimColor).Align(lipgloss.Center).Width(contentW).Render(subText)

	// Variants list
	startVar := m.varOffset
	endVar := startVar + maxVisibleVars
	if endVar > len(m.variants) {
		endVar = len(m.variants)
	}

	var listLines []string
	for i := startVar; i < endVar; i++ {
		v := m.variants[i]
		label := v.Label
		if v.Description != "" {
			label = fmt.Sprintf("%-24s (%s)", v.Label, v.Description)
		}
		maxLabelW := contentW - 4
		if runewidth.StringWidth(label) > maxLabelW && maxLabelW > 3 {
			label = sliceLineByWidth(label, 0, maxLabelW-1) + "…"
		}

		if i == m.varCursor {
			listLines = append(listLines, modalItemActive.Width(contentW).Render("▸ "+label))
		} else {
			listLines = append(listLines, modalItemNormal.Width(contentW).Render("  "+label))
		}
	}
	for len(listLines) < maxVisibleVars {
		listLines = append(listLines, lipgloss.NewStyle().Width(contentW).Render(""))
	}

	variantListBlock := strings.Join(listLines, "\n")

	// Divider with scroll/pan hints if applicable
	divText := "─── Live Variant Preview ───"
	if m.varXOffset > 0 {
		divText = fmt.Sprintf("─── Live Preview [Pan: %d | ←/→] ───", m.varXOffset)
	}
	previewDivider := lipgloss.NewStyle().Foreground(dimColor).Align(lipgloss.Center).Width(contentW).Render(divText)

	// Preview Box
	previewInnerW := contentW - 2
	if previewInnerW < 10 {
		previewInnerW = 10
	}
	previewInnerH := previewOuterH - 2
	if previewInnerH < 1 {
		previewInnerH = 1
	}

	rawLines := strings.Split(strings.TrimRight(m.varPreviewText, "\n"), "\n")

	// Apply varYOffset if any
	startLineIdx := m.varYOffset
	if startLineIdx >= len(rawLines) {
		startLineIdx = len(rawLines) - 1
	}
	if startLineIdx < 0 {
		startLineIdx = 0
	}

	var visibleArtLines []string
	if startLineIdx < len(rawLines) {
		visibleArtLines = rawLines[startLineIdx:]
	}

	previewLines := make([]string, 0, previewInnerH)

	// Vertical centering if art is shorter than previewInnerH and yOffset == 0
	topPad := 0
	if m.varYOffset == 0 && len(visibleArtLines) > 0 && len(visibleArtLines) < previewInnerH {
		topPad = (previewInnerH - len(visibleArtLines)) / 2
	}
	for p := 0; p < topPad; p++ {
		previewLines = append(previewLines, "")
	}

	for i := 0; i < len(visibleArtLines) && len(previewLines) < previewInnerH; i++ {
		previewLines = append(previewLines, sliceLineByWidth(visibleArtLines[i], m.varXOffset, previewInnerW))
	}
	for len(previewLines) < previewInnerH {
		previewLines = append(previewLines, "")
	}

	previewBlock := lipgloss.NewStyle().
		Border(roundedBorder).
		BorderForeground(accentColor).
		Width(previewInnerW).
		Height(previewInnerH).
		Render(strings.Join(previewLines, "\n"))

	// Footer
	footerText := "↑/↓: Choose  •  Enter: Finalize & Print  •  Esc: Cancel"
	if runewidth.StringWidth(footerText) > contentW {
		footerText = "↑/↓: Choose • Enter: OK • Esc: Cancel"
	}
	footer := lipgloss.NewStyle().
		Foreground(dimColor).
		Align(lipgloss.Center).
		Width(contentW).
		Render(footerText)

	modalContent := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		subtitle,
		variantListBlock,
		previewDivider,
		previewBlock,
		footer,
	)

	return modalBoxStyle.Width(modalOuterW - 2).Height(modalInnerH(modalOuterH)).Render(modalContent)
}

func modalInnerH(outerH int) int {
	h := outerH - 2
	if h < 1 {
		return 1
	}
	return h
}

