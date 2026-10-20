package app

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/vivekananda-2201/ascii-font-generator-tui/internal/fonts"
)

func TestVariantModalAllDiscoveredFonts(t *testing.T) {
	discovered, err := fonts.DiscoverFonts("")
	if err != nil {
		t.Fatalf("discover failed: %v", err)
	}

	termW := 120
	termH := 30
	var badModals []string

	for _, f := range discovered {
		m := NewModel([]fonts.FontEntry{f}, false)
		m.width = termW
		m.height = termH
		m.textInput.SetValue("GTR")
		m.selectedFont = f
		m.variants = fonts.GetVariants(f, []fonts.FontEntry{f})
		m.varCursor = 0
		m.varOffset = 0
		m.State = StateVariant

		for vIdx := 0; vIdx < len(m.variants); vIdx++ {
			m.varCursor = vIdx
			m.updateVariantPreview()

			rendered := m.renderVariantModal()
			modalW := lipgloss.Width(rendered)
			modalH := lipgloss.Height(rendered)

			if modalH > termH {
				badModals = append(badModals, fmt.Sprintf("%s (var %d): modal height %d > term height %d", f.Name, vIdx, modalH, termH))
				break
			}
			if modalW > termW {
				badModals = append(badModals, fmt.Sprintf("%s (var %d): modal width %d > term width %d", f.Name, vIdx, modalW, termW))
				break
			}

			lines := strings.Split(rendered, "\n")
			if len(lines) != modalH {
				badModals = append(badModals, fmt.Sprintf("%s line count mismatch: len(lines)=%d != modalH=%d", f.Name, len(lines), modalH))
				break
			}

			for lineIdx, line := range lines {
				w := runewidth.StringWidth(line)
				if w != modalW {
					badModals = append(badModals, fmt.Sprintf("%s line %d width %d != modalW %d: %q", f.Name, lineIdx, w, modalW, line))
					break
				}
			}
		}
	}

	if len(badModals) > 0 {
		t.Fatalf("Found %d broken modals across discovered fonts: %v", len(badModals), badModals[:5])
	}
}

func TestVariantModalMultiResolution(t *testing.T) {
	discovered, err := fonts.DiscoverFonts("")
	if err != nil {
		t.Fatalf("discover failed: %v", err)
	}

	testSizes := [][2]int{
		{160, 45}, // Large widescreen
		{120, 30}, // Standard modern
		{80, 24},  // Classic VT100
		{60, 18},  // Small window
	}

	testInputs := []string{
		"AGEN",
		"GTR",
		"",
		"Testing Overflow In Variant Preview Modal",
	}

	sampleFonts := []string{
		"Delta Corps Priest 1",
		"Standard",
		"Big",
		"Banner",
		"Slant",
		"Small",
		"Alpha",
		"Block",
		"Digital",
	}

	for _, size := range testSizes {
		termW, termH := size[0], size[1]
		for _, input := range testInputs {
			for _, fontName := range sampleFonts {
				var found *fonts.FontEntry
				for i := range discovered {
					if discovered[i].Name == fontName {
						found = &discovered[i]
						break
					}
				}
				if found == nil {
					continue
				}

				m := NewModel([]fonts.FontEntry{*found}, false)
				m.width = termW
				m.height = termH
				m.textInput.SetValue(input)
				m.selectedFont = *found
				m.variants = fonts.GetVariants(*found, discovered)
				m.varCursor = 0
				m.varOffset = 0
				m.State = StateVariant
				m.updateVariantPreview()

				rendered := m.renderVariantModal()
				modalW := lipgloss.Width(rendered)
				modalH := lipgloss.Height(rendered)

				if modalH > termH {
					t.Fatalf("Size %dx%d, input %q, font %s: modalH %d > termH %d", termW, termH, input, fontName, modalH, termH)
				}
				if modalW > termW {
					t.Fatalf("Size %dx%d, input %q, font %s: modalW %d > termW %d", termW, termH, input, fontName, modalW, termW)
				}

				lines := strings.Split(rendered, "\n")
				if len(lines) != modalH {
					t.Fatalf("Size %dx%d, font %s: len(lines)=%d != modalH=%d", termW, termH, fontName, len(lines), modalH)
				}

				for lineIdx, line := range lines {
					w := runewidth.StringWidth(line)
					if w != modalW {
						t.Fatalf("Size %dx%d, input %q, font %s: line %d width %d != modalW %d", termW, termH, input, fontName, lineIdx, w, modalW)
					}
				}
			}
		}
	}
}

func TestVariantModalNavigationAndPanning(t *testing.T) {
	fontEntry := fonts.FontEntry{
		Name:   "Delta Corps Priest 1",
		Path:   "/home/vicky/.local/share/figlet-fonts/Delta Corps Priest 1.flf",
		Family: "Delta Corps Priest 1",
	}

	m := NewModel([]fonts.FontEntry{fontEntry}, false)
	nm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 26})
	m = nm.(Model)
	m.textInput.SetValue("LONG_TEXT_FOR_PREVIEW")

	// Enter variant state
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = nm.(Model)
	if m.State != StateVariant {
		t.Fatalf("expected StateVariant, got %v", m.State)
	}

	// Navigate down
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = nm.(Model)
	if m.varCursor != 1 {
		t.Fatalf("expected varCursor 1, got %d", m.varCursor)
	}

	// Pan right
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = nm.(Model)
	if m.varXOffset != 4 {
		t.Fatalf("expected varXOffset 4, got %d", m.varXOffset)
	}

	// Pan left
	nm, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = nm.(Model)
	if m.varXOffset != 0 {
		t.Fatalf("expected varXOffset 0, got %d", m.varXOffset)
	}

	// Check rendered modal does not panic or wrap
	rendered := m.renderVariantModal()
	if lipgloss.Height(rendered) > 26 {
		t.Fatalf("modal overflowed terminal height: %d > 26", lipgloss.Height(rendered))
	}
}
