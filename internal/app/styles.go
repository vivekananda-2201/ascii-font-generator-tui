package app

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Palette
	primaryColor   = lipgloss.Color("#7D56F4") // Electric Purple
	accentColor    = lipgloss.Color("#04B575") // Mint Green
	secondaryColor = lipgloss.Color("#FF79C6") // Vibrant Pink
	textColor      = lipgloss.Color("#F8F8F2") // Off-white
	dimColor       = lipgloss.Color("#6272A4") // Muted Slate
	bgDarkColor    = lipgloss.Color("#1E1F29") // Deep Gray
	highlightBg    = lipgloss.Color("#353749") // Highlight background
	warningColor   = lipgloss.Color("#FFB86C") // Orange

	// Borders
	roundedBorder = lipgloss.RoundedBorder()
	thickBorder   = lipgloss.ThickBorder()
	doubleBorder  = lipgloss.DoubleBorder()

	// Title styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Padding(0, 1)

	activeTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(accentColor).
				Padding(0, 1)

	// Top Input Container Style
	inputBoxFocused = lipgloss.NewStyle().
			Border(roundedBorder).
			BorderForeground(accentColor).
			Padding(0, 1)

	inputBoxUnfocused = lipgloss.NewStyle().
				Border(roundedBorder).
				BorderForeground(dimColor).
				Padding(0, 1)

	// Panel Styles
	panelFocused = lipgloss.NewStyle().
			Border(roundedBorder).
			BorderForeground(accentColor)

	panelUnfocused = lipgloss.NewStyle().
			Border(roundedBorder).
			BorderForeground(dimColor)

	// Font List item styles
	itemSelectedStyle = lipgloss.NewStyle().
				Background(highlightBg).
				Foreground(accentColor).
				Bold(true).
				Padding(0, 1)

	itemNormalStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Padding(0, 1)

	itemFamilyTag = lipgloss.NewStyle().
			Foreground(dimColor).
			Italic(true)

	// Status Bar Style
	statusBarStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Background(bgDarkColor)

	statusKeyStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Bold(true)

	statusDescStyle = lipgloss.NewStyle().
			Foreground(dimColor)

	statusSepStyle = lipgloss.NewStyle().
			Foreground(dimColor).
			SetString(" • ")

	// Search bar overlay style
	searchBoxStyle = lipgloss.NewStyle().
			Border(thickBorder).
			BorderForeground(secondaryColor).
			Padding(0, 1)

	// Modal Overlay Style
	modalBoxStyle = lipgloss.NewStyle().
			Border(doubleBorder).
			BorderForeground(primaryColor).
			Background(bgDarkColor).
			Padding(1, 2)

	modalTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor).
			Align(lipgloss.Center)

	modalItemActive = lipgloss.NewStyle().
			Background(highlightBg).
			Foreground(secondaryColor).
			Bold(true).
			Padding(0, 1)

	modalItemNormal = lipgloss.NewStyle().
			Foreground(textColor).
			Padding(0, 1)

	warningStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Italic(true)
)
