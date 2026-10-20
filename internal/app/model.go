package app

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/vivekananda-2201/ascii-font-generator-tui/internal/fonts"
)

// State represents the current active screen/view of the TUI.
type State int

const (
	StateBrowse State = iota
	StateSearch
	StateVariant
)

// fontNameSource implements fuzzy.Source for fuzzy searching font entries.
type fontNameSource []fonts.FontEntry

func (f fontNameSource) String(i int) string { return f[i].Name }
func (f fontNameSource) Len() int           { return len(f) }

// Model represents the complete Bubble Tea application state.
type Model struct {
	State State

	// Fonts & Rendering
	renderer      *fonts.Renderer
	allFonts      []fonts.FontEntry
	filteredFonts []fonts.FontEntry
	cursor        int
	listOffset    int
	xOffset       int // Horizontal preview scroll offset

	// Sub-components
	textInput   textinput.Model
	searchInput textinput.Model
	viewport    viewport.Model

	// State tracking for Search
	prevCursor   int
	prevFontName string

	// State tracking for Variant Modal
	selectedFont   fonts.FontEntry
	variants       []fonts.VariantOption
	varCursor      int
	varOffset      int
	varXOffset     int
	varYOffset     int
	varPreviewText string
	varViewport    viewport.Model

	// Dimensions
	width       int
	height      int
	ready       bool
	focusInput  bool // true = textInput focused, false = font list focused

	// Cached Art
	renderedArt   string
	artWidth      int
	renderedLines int

	// Flags and output
	ClipEnabled   bool
	Confirmed     bool
	Output        string
	StatusMessage string
}

// NewModel initializes the root TUI model with discovered fonts.
func NewModel(fontList []fonts.FontEntry, clip bool) Model {
	ti := textinput.New()
	ti.Placeholder = "Type something..."
	ti.Focus()
	ti.CharLimit = 120
	ti.Width = 50

	si := textinput.New()
	si.Placeholder = "Filter fonts (fuzzy)..."
	si.CharLimit = 50
	si.Width = 40

	renderer := fonts.NewRenderer(fontList)

	m := Model{
		State:         StateBrowse,
		renderer:      renderer,
		allFonts:      fontList,
		filteredFonts: fontList,
		cursor:        0,
		listOffset:    0,
		xOffset:       0,
		textInput:     ti,
		searchInput:   si,
		focusInput:    true,
		ClipEnabled:   clip,
	}

	// Render initial preview with placeholder
	if len(fontList) > 0 {
		m.updatePreview()
	}

	return m
}

// Init initializes the Bubble Tea program.
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// visibleVariantSlots determines how many variant items fit in the modal dynamically.
func (m Model) visibleVariantSlots() int {
	modalOuterH := m.height - 2
	if modalOuterH > 28 {
		modalOuterH = 28
	}
	contentH := modalOuterH - 4
	if contentH >= 22 {
		return 5
	} else if contentH < 14 && contentH >= 10 {
		return 3
	} else if contentH < 10 {
		return 2
	}
	return 4
}

