package fonts

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FontEntry represents an indexed FIGlet font.
type FontEntry struct {
	Name      string // Font display name (filename without .flf, or builtin name)
	Path      string // Path to the .flf file (empty if built-in)
	IsBuiltin bool   // True if bundled into binary
	Family    string // Grouped family name
}

// DiscoverFonts scans standard system and user directories for .flf fonts,
// deduplicates them by name, and returns an alphabetically sorted list.
func DiscoverFonts(extraDir string) ([]FontEntry, error) {
	home, _ := os.UserHomeDir()

	candidateDirs := []string{}
	if extraDir != "" {
		candidateDirs = append(candidateDirs, extraDir)
	}

	if home != "" {
		candidateDirs = append(candidateDirs,
			filepath.Join(home, ".local", "share", "figlet-fonts"),
			filepath.Join(home, ".figlet-fonts"),
		)
	}

	candidateDirs = append(candidateDirs,
		"/usr/share/figlet/fonts",
		"/usr/share/figlet",
		"/usr/share/fonts/figlet",
		"/usr/local/share/figlet",
		".",
	)

	seen := make(map[string]bool)
	var result []FontEntry

	// Add built-in defaults first so they are always guaranteed to exist
	builtins := []string{"standard", "larry3d"}
	for _, b := range builtins {
		lower := strings.ToLower(b)
		seen[lower] = true
		result = append(result, FontEntry{
			Name:      b,
			Path:      "",
			IsBuiltin: true,
			Family:    b,
		})
	}

	for _, dir := range candidateDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue // Gracefully skip unreadable or non-existent directories
		}

		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			fname := e.Name()
			if !strings.HasSuffix(strings.ToLower(fname), ".flf") {
				continue // Ignore non-.flf files (.flc, .tlf, .md, etc.)
			}

			fontName := fname[:len(fname)-4]
			lower := strings.ToLower(fontName)
			if seen[lower] {
				continue
			}
			seen[lower] = true

			fullPath := filepath.Join(dir, fname)
			result = append(result, FontEntry{
				Name:      fontName,
				Path:      fullPath,
				IsBuiltin: false,
			})
		}
	}

	// Assign family names using family detection
	AssignFamilies(result)

	// Sort alphabetically case-insensitively
	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})

	return result, nil
}
