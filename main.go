package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"
	"github.com/vivekananda-2201/ascii-font-generator-tui/internal/app"
	"github.com/vivekananda-2201/ascii-font-generator-tui/internal/clipboard"
	"github.com/vivekananda-2201/ascii-font-generator-tui/internal/fonts"
)

const version = "1.0.0"

func main() {
	clipFlag := flag.Bool("clip", false, "Copy generated ASCII art to clipboard on exit")
	shortClipFlag := flag.Bool("c", false, "Copy generated ASCII art to clipboard on exit (shorthand)")
	fontDirFlag := flag.String("font-dir", "", "Additional directory to scan for .flf FIGlet fonts")
	shortFontDirFlag := flag.String("d", "", "Additional directory to scan for .flf FIGlet fonts (shorthand)")
	versionFlag := flag.Bool("version", false, "Print version information and exit")
	shortVersionFlag := flag.Bool("v", false, "Print version information and exit (shorthand)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "agen (ASCII Generator) - Interactive Terminal FIGlet Art Generator v%s\n\n", version)
		fmt.Fprintf(os.Stderr, "Usage: agen [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -c, --clip       Copy output directly to clipboard upon exit\n")
		fmt.Fprintf(os.Stderr, "  -d, --font-dir   Scan additional directory for .flf fonts\n")
		fmt.Fprintf(os.Stderr, "  -v, --version    Print version and exit\n")
		fmt.Fprintf(os.Stderr, "  -h, --help       Show this help message\n\n")
		fmt.Fprintf(os.Stderr, "Keybindings:\n")
		fmt.Fprintf(os.Stderr, "  ↑ / ↓            Navigate font list\n")
		fmt.Fprintf(os.Stderr, "  Tab              Toggle focus between Text Input and Font List\n")
		fmt.Fprintf(os.Stderr, "  Ctrl+S           Fuzzy search/filter fonts\n")
		fmt.Fprintf(os.Stderr, "  Enter            Advance to Size/Variant selector & finalize\n")
		fmt.Fprintf(os.Stderr, "  Ctrl+Y           Copy current preview to clipboard\n")
		fmt.Fprintf(os.Stderr, "  Esc / Ctrl+C     Cancel and exit without output\n")
	}

	flag.Parse()

	if *versionFlag || *shortVersionFlag {
		fmt.Printf("agen version %s\n", version)
		os.Exit(0)
	}

	clip := *clipFlag || *shortClipFlag
	extraDir := *fontDirFlag
	if extraDir == "" {
		extraDir = *shortFontDirFlag
	}

	// Discover fonts
	fontList, err := fonts.DiscoverFonts(extraDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error discovering fonts: %v\n", err)
		os.Exit(1)
	}

	if len(fontList) == 0 {
		fmt.Fprintf(os.Stderr, "No FIGlet fonts found. Please install figlet fonts or pass --font-dir\n")
		os.Exit(1)
	}

	if !term.IsTerminal(os.Stdin.Fd()) {
		fmt.Fprintf(os.Stderr, "agen: interactive TUI requires a terminal (TTY).\n")
		os.Exit(1)
	}

	// Initialize Bubble Tea program with alternate screen buffer
	model := app.NewModel(fontList, clip)
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running agen: %v\n", err)
		os.Exit(1)
	}

	m, ok := finalModel.(app.Model)
	if !ok {
		os.Exit(0)
	}

	// If confirmed, cleanly emit the generated ASCII art to stdout
	if m.Confirmed && m.Output != "" {
		fmt.Print(m.Output)

		// If clip enabled, copy to clipboard
		if clip {
			_ = clipboard.Copy(m.Output)
		}
	}
}
