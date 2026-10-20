package fonts

import (
	"bytes"
	"os"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/mbndr/figlet4go"
	"golang.org/x/text/encoding/charmap"
)

// Renderer encapsulates FIGlet rendering with normalized font loading, sanitization, and layout options.
type Renderer struct {
	ascii       *figlet4go.AsciiRender
	loadedFonts map[string]bool
	fontPaths   map[string]string
	mu          sync.RWMutex
}

// NewRenderer initializes a new Renderer preloaded with the discovered fonts.
func NewRenderer(fonts []FontEntry) *Renderer {
	r := &Renderer{
		ascii:       figlet4go.NewAsciiRender(),
		loadedFonts: make(map[string]bool),
		fontPaths:   make(map[string]string),
	}

	for _, f := range fonts {
		if f.Path != "" {
			r.fontPaths[strings.ToLower(f.Name)] = f.Path
		}
	}

	return r
}

// EnsureFontLoaded cleans and registers the font with figlet4go if not already loaded.
func (r *Renderer) EnsureFontLoaded(fontPath, fontName string) {
	lowerName := strings.ToLower(fontName)
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.loadedFonts[lowerName] {
		return
	}

	targetPath := fontPath
	if targetPath == "" {
		targetPath = r.fontPaths[lowerName]
	}
	if targetPath == "" {
		return
	}

	_ = r.loadFontClean(targetPath, fontName)
	r.loadedFonts[lowerName] = true
}

// loadFontClean loads a .flf font file, normalizes CRLF and headers,
// and registers it with figlet4go's binary loader.
// This eliminates the critical bug where \r in header caused strconv.Atoi to fail,
// shifting characters by the comment line count and displaying descenders at the top!
func (r *Renderer) loadFontClean(path, name string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// 1. Normalize line endings (CRLF and stray CR to LF)
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	data = bytes.ReplaceAll(data, []byte("\r"), []byte("\n"))

	// 2. Normalize header line to single spaces between tokens so strings.Split always succeeds
	idx := bytes.IndexByte(data, '\n')
	if idx > 0 {
		headerLine := string(data[:idx])
		fields := strings.Fields(headerLine)
		if len(fields) >= 6 {
			cleanHeader := strings.Join(fields, " ")
			var buf bytes.Buffer
			buf.WriteString(cleanHeader)
			buf.WriteByte('\n')
			buf.Write(data[idx+1:])
			data = buf.Bytes()
		}
	}

	return r.ascii.LoadBindataFont(data, name)
}

// Render renders the provided text using the specified font and layout mode.
// If text is empty, the placeholder "AGEN" is used.
// Output is sanitized to guarantee valid UTF-8 and clean terminal control sequences.
func (r *Renderer) Render(fontName, text string, layout LayoutMode) (string, error) {
	r.EnsureFontLoaded("", fontName)

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Default placeholder if text is empty
	displayText := text
	if displayText == "" {
		displayText = "AGEN"
	}

	// Apply layout spacing to characters if requested
	switch layout {
	case LayoutWide:
		var b strings.Builder
		for i, runeChar := range displayText {
			if i > 0 {
				b.WriteString(" ")
			}
			b.WriteRune(runeChar)
		}
		displayText = b.String()
	case LayoutExtraWide:
		var b strings.Builder
		for i, runeChar := range displayText {
			if i > 0 {
				b.WriteString("  ")
			}
			b.WriteRune(runeChar)
		}
		displayText = b.String()
	}

	opts := figlet4go.NewRenderOptions()
	opts.FontName = fontName

	out, err := r.ascii.RenderOpts(displayText, opts)
	if err != nil || len(out) == 0 {
		// Fallback to standard font
		opts.FontName = "standard"
		out, err = r.ascii.RenderOpts(displayText, opts)
	}

	out = SanitizeASCIIArt(out)

	if layout == LayoutCompact {
		out = trimBlankLines(out)
	}

	return out, err
}

// SanitizeASCIIArt cleans raw FIGlet output:
// 1. Converts legacy CP437/ISO-8859-1 byte sequences into valid UTF-8 block runes.
// 2. Strips terminal-corrupting control characters and stray carriage returns.
// 3. Normalizes CRLF to LF.
func SanitizeASCIIArt(s string) string {
	s = strings.ReplaceAll(s, "\r", "")

	// Check for invalid UTF-8 (common in legacy MS-DOS FIGlet fonts with CP437 high bytes)
	if !utf8.ValidString(s) {
		// Decode from CP437 (handles 0xDB -> █, 0xB0 -> ░, 0xB1 -> ▒, etc.)
		decoded, err := charmap.CodePage437.NewDecoder().String(s)
		if err == nil && utf8.ValidString(decoded) {
			s = decoded
		} else {
			s = strings.ToValidUTF8(s, "")
		}
	}

	// Filter out non-printable control characters that could corrupt terminal emulator state
	var clean strings.Builder
	clean.Grow(len(s))
	for _, r := range s {
		if r == '\n' || r == '\t' || (r >= 32 && r != 127) {
			clean.WriteRune(r)
		} else if r < 32 {
			clean.WriteRune(' ') // Replace rogue control char with space
		}
	}
	return clean.String()
}

// trimBlankLines removes empty leading and trailing lines from the ASCII output.
func trimBlankLines(s string) string {
	lines := strings.Split(s, "\n")
	start := 0
	end := len(lines)

	for start < end && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}

	if start >= end {
		return s
	}

	return strings.Join(lines[start:end], "\n") + "\n"
}
