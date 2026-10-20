package fonts

import (
	"strings"
)

var variantKeywords = []string{
	"small", "big", "mini", "micro", "large", "medium", "shadow", "slant",
	"regular", "italic", "script", "bold", "thin", "thick", "3d", "isometric",
	"smushed", "sub", "super", "caps", "wide", "short", "tall", "ansi", "ascii",
	"mono", "chief", "rebel", "pepper", "money",
}

// NormalizeFamily extracts a common root family name from a font name.
func NormalizeFamily(name string) string {
	clean := strings.ToLower(name)
	clean = strings.ReplaceAll(clean, "_", " ")
	clean = strings.ReplaceAll(clean, "-", " ")
	words := strings.Fields(clean)

	var rootWords []string
	for _, w := range words {
		isVar := false
		for _, v := range variantKeywords {
			if w == v {
				isVar = true
				break
			}
		}
		if !isVar {
			rootWords = append(rootWords, w)
		}
	}

	if len(rootWords) == 0 {
		if len(words) > 0 {
			return strings.Title(words[0])
		}
		return name
	}

	for i := range rootWords {
		rootWords[i] = strings.Title(rootWords[i])
	}
	return strings.Join(rootWords, " ")
}

// AssignFamilies populates the Family field for each font in the slice.
func AssignFamilies(fontList []FontEntry) {
	for i := range fontList {
		fontList[i].Family = NormalizeFamily(fontList[i].Name)
	}
}

// LayoutMode represents character spacing layout modifications.
type LayoutMode int

const (
	LayoutDefault LayoutMode = iota
	LayoutWide
	LayoutExtraWide
	LayoutCompact
)

func (l LayoutMode) String() string {
	switch l {
	case LayoutWide:
		return "Wide (+1 Space)"
	case LayoutExtraWide:
		return "Extra Wide (+2 Spaces)"
	case LayoutCompact:
		return "Compact (Trimmed)"
	default:
		return "Default"
	}
}

// VariantOption describes a variant that can be selected in the modal.
type VariantOption struct {
	Label       string     // User-friendly display label (e.g. "ANSI Shadow", "Wide Spacing")
	FontName    string     // Target font name to render
	FontPath    string     // Target font path
	Layout      LayoutMode // Spacing / layout mode
	Description string     // Contextual description
}

// GetVariants returns the size/style variants for the selected font.
// If the font has related sibling fonts in the same family, those are included.
// Spacing/layout modes are also provided for fine-tuning.
func GetVariants(selected FontEntry, allFonts []FontEntry) []VariantOption {
	var siblings []FontEntry
	for _, f := range allFonts {
		if f.Family != "" && f.Family == selected.Family {
			siblings = append(siblings, f)
		}
	}

	var options []VariantOption

	if len(siblings) > 1 {
		// Include family sibling fonts as size/style variants
		for _, s := range siblings {
			desc := "Family Variant"
			if strings.EqualFold(s.Name, selected.Name) {
				desc = "Selected Font"
			}
			options = append(options, VariantOption{
				Label:       s.Name,
				FontName:    s.Name,
				FontPath:    s.Path,
				Layout:      LayoutDefault,
				Description: desc,
			})
		}
	}

	// Always offer layout/spacing adjustments for the active font
	options = append(options,
		VariantOption{
			Label:       selected.Name + " - Default",
			FontName:    selected.Name,
			FontPath:    selected.Path,
			Layout:      LayoutDefault,
			Description: "Standard FIGlet Layout",
		},
		VariantOption{
			Label:       selected.Name + " - Wide",
			FontName:    selected.Name,
			FontPath:    selected.Path,
			Layout:      LayoutWide,
			Description: "Spaced Characters (+1)",
		},
		VariantOption{
			Label:       selected.Name + " - Extra Wide",
			FontName:    selected.Name,
			FontPath:    selected.Path,
			Layout:      LayoutExtraWide,
			Description: "Spaced Characters (+2)",
		},
		VariantOption{
			Label:       selected.Name + " - Compact",
			FontName:    selected.Name,
			FontPath:    selected.Path,
			Layout:      LayoutCompact,
			Description: "Trimmed Margins",
		},
	)

	// Deduplicate if any exact label overlap occurs
	seen := make(map[string]bool)
	var finalOptions []VariantOption
	for _, opt := range options {
		if !seen[opt.Label] {
			seen[opt.Label] = true
			finalOptions = append(finalOptions, opt)
		}
	}

	return finalOptions
}
