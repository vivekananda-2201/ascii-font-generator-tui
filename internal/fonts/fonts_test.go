package fonts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscoverFonts(t *testing.T) {
	fonts, err := DiscoverFonts("")
	if err != nil {
		t.Fatalf("DiscoverFonts error: %v", err)
	}

	if len(fonts) == 0 {
		t.Fatalf("Expected to discover fonts, got 0")
	}

	// Verify standard and larry3d builtins are present
	hasStandard := false
	hasLarry := false
	for _, f := range fonts {
		if f.Name == "standard" {
			hasStandard = true
		}
		if f.Name == "larry3d" {
			hasLarry = true
		}
	}

	if !hasStandard {
		t.Errorf("Expected standard font in discovery")
	}
	if !hasLarry {
		t.Errorf("Expected larry3d font in discovery")
	}
}

func TestRenderText(t *testing.T) {
	fonts, err := DiscoverFonts("")
	if err != nil {
		t.Fatalf("DiscoverFonts error: %v", err)
	}

	renderer := NewRenderer(fonts)

	// Test default rendering with placeholder
	out, err := renderer.Render("standard", "", LayoutDefault)
	if err != nil {
		t.Fatalf("Render empty text error: %v", err)
	}
	if !strings.Contains(out, "AGEN") && len(out) == 0 {
		t.Errorf("Expected output for placeholder 'AGEN'")
	}

	// Test rendering custom text
	out2, err := renderer.Render("standard", "Hello", LayoutDefault)
	if err != nil {
		t.Fatalf("Render custom text error: %v", err)
	}
	if len(out2) == 0 {
		t.Errorf("Expected non-empty output for 'Hello'")
	}

	// Test Wide mode
	outWide, err := renderer.Render("standard", "HI", LayoutWide)
	if err != nil {
		t.Fatalf("Render wide error: %v", err)
	}
	if len(outWide) == 0 {
		t.Errorf("Expected non-empty output for wide layout")
	}

	// Test Compact mode
	outCompact, err := renderer.Render("standard", "HI", LayoutCompact)
	if err != nil {
		t.Fatalf("Render compact error: %v", err)
	}
	if len(outCompact) == 0 {
		t.Errorf("Expected non-empty output for compact layout")
	}
}

func TestDetectFamiliesAndVariants(t *testing.T) {
	allFonts := []FontEntry{
		{Name: "ANSI Shadow", Family: "Ansi"},
		{Name: "ANSI Regular", Family: "Ansi"},
		{Name: "Banner", Family: "Banner"},
	}

	variants := GetVariants(allFonts[0], allFonts)
	if len(variants) < 2 {
		t.Errorf("Expected at least 2 variants for ANSI Shadow family, got %d", len(variants))
	}

	hasRegular := false
	for _, v := range variants {
		if strings.Contains(v.Label, "ANSI Regular") {
			hasRegular = true
			break
		}
	}
	if !hasRegular {
		t.Errorf("Expected ANSI Regular to be among variants of ANSI Shadow")
	}
}

func TestDiscoverFontsWithExtraDir(t *testing.T) {
	// Create a temp dir with a dummy .flf file
	tmpDir, err := os.MkdirTemp("", "agen_fonts_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dummyFontPath := filepath.Join(tmpDir, "TestCustomFont.flf")
	// Header for minimal valid FIGlet font
	header := "flf2a$ 1 1 1 -1 0\n$\nTestCustomFont\n"
	if err := os.WriteFile(dummyFontPath, []byte(header), 0644); err != nil {
		t.Fatalf("Failed to write dummy font: %v", err)
	}

	fonts, err := DiscoverFonts(tmpDir)
	if err != nil {
		t.Fatalf("DiscoverFonts with extra dir error: %v", err)
	}

	found := false
	for _, f := range fonts {
		if f.Name == "TestCustomFont" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected TestCustomFont to be found in extra dir")
	}
}

func TestFontGlyphOrder(t *testing.T) {
	fonts, err := DiscoverFonts("")
	if err != nil {
		t.Fatalf("DiscoverFonts error: %v", err)
	}

	renderer := NewRenderer(fonts)

	// Test Alpha font: line 0 of 'V' should not be descenders
	outAlpha, err := renderer.Render("Alpha", "V", LayoutDefault)
	if err != nil {
		t.Fatalf("Render Alpha error: %v", err)
	}
	linesAlpha := strings.Split(strings.TrimRight(outAlpha, "\n"), "\n")
	if len(linesAlpha) < 10 {
		t.Fatalf("Expected Alpha to have at least 10 lines, got %d", len(linesAlpha))
	}
	// In Alpha, top of 'V' has _____ or empty space, not descenders \::::::/
	if strings.Contains(linesAlpha[0], `\::::::/`) {
		t.Errorf("Alpha has descenders on line 0: %q", linesAlpha[0])
	}

	// Test AMC Thin: top of 'V' should not be descenders
	outAMC, err := renderer.Render("AMC Thin", "V", LayoutDefault)
	if err != nil {
		t.Fatalf("Render AMC Thin error: %v", err)
	}
	linesAMC := strings.Split(strings.TrimRight(outAMC, "\n"), "\n")
	if len(linesAMC) < 5 {
		t.Fatalf("Expected AMC Thin to have at least 5 lines, got %d", len(linesAMC))
	}
	// In AMC Thin, top of V has .-. not |-| or `-'
	if strings.Contains(linesAMC[0], "`-'") {
		t.Errorf("AMC Thin has feet on line 0: %q", linesAMC[0])
	}
}
