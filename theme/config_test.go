package theme

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// ParseStyleString tests
// ---------------------------------------------------------------------------

// TestParseStyleString_FgOnly confirms "#ff0000" sets foreground to RGB(255,0,0)
// with a default background.
func TestParseStyleString_FgOnly(t *testing.T) {
	style, err := ParseStyleString("#ff0000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fg, bg, _ := style.Decompose()

	wantFg := tcell.NewRGBColor(255, 0, 0)
	if fg != wantFg {
		t.Errorf("foreground: got %v, want %v", fg, wantFg)
	}

	if bg != tcell.ColorDefault {
		t.Errorf("background: got %v, want ColorDefault", bg)
	}
}

// TestParseStyleString_FgAndBg confirms "#ffffff:#000000" sets both foreground
// and background correctly.
func TestParseStyleString_FgAndBg(t *testing.T) {
	style, err := ParseStyleString("#ffffff:#000000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fg, bg, _ := style.Decompose()

	wantFg := tcell.NewRGBColor(255, 255, 255)
	wantBg := tcell.NewRGBColor(0, 0, 0)

	if fg != wantFg {
		t.Errorf("foreground: got %v, want %v", fg, wantFg)
	}
	if bg != wantBg {
		t.Errorf("background: got %v, want %v", bg, wantBg)
	}
}

// TestParseStyleString_FgBgBold confirms "#ffffff:#000000:bold" sets fg, bg,
// and the bold attribute.
func TestParseStyleString_FgBgBold(t *testing.T) {
	style, err := ParseStyleString("#ffffff:#000000:bold")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fg, bg, attrs := style.Decompose()

	wantFg := tcell.NewRGBColor(255, 255, 255)
	wantBg := tcell.NewRGBColor(0, 0, 0)

	if fg != wantFg {
		t.Errorf("foreground: got %v, want %v", fg, wantFg)
	}
	if bg != wantBg {
		t.Errorf("background: got %v, want %v", bg, wantBg)
	}
	if attrs&tcell.AttrBold == 0 {
		t.Error("expected bold attribute to be set")
	}
}

// TestParseStyleString_FgBgBoldUnderline confirms "#ffffff:#000000:bold,underline"
// sets fg, bg, bold, and underline.
func TestParseStyleString_FgBgBoldUnderline(t *testing.T) {
	style, err := ParseStyleString("#ffffff:#000000:bold,underline")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fg, bg, attrs := style.Decompose()

	wantFg := tcell.NewRGBColor(255, 255, 255)
	wantBg := tcell.NewRGBColor(0, 0, 0)

	if fg != wantFg {
		t.Errorf("foreground: got %v, want %v", fg, wantFg)
	}
	if bg != wantBg {
		t.Errorf("background: got %v, want %v", bg, wantBg)
	}
	if attrs&tcell.AttrBold == 0 {
		t.Error("expected bold attribute to be set")
	}
	if attrs&tcell.AttrUnderline == 0 {
		t.Error("expected underline attribute to be set")
	}
}

// TestParseStyleString_InvalidHex confirms that a malformed hex color returns
// an error.
func TestParseStyleString_InvalidHex(t *testing.T) {
	_, err := ParseStyleString("#xyz")
	if err == nil {
		t.Error("expected an error for invalid hex color, got nil")
	}
}

// TestParseStyleString_Empty confirms that an empty string returns a default
// style with no error.
func TestParseStyleString_Empty(t *testing.T) {
	style, err := ParseStyleString("")
	if err != nil {
		t.Fatalf("unexpected error for empty string: %v", err)
	}

	// An empty string should produce a default (no-override) style.
	// We check that no specific foreground/background was forced.
	fg, bg, _ := style.Decompose()
	if fg != tcell.ColorDefault {
		t.Errorf("foreground: got %v, want ColorDefault", fg)
	}
	if bg != tcell.ColorDefault {
		t.Errorf("background: got %v, want ColorDefault", bg)
	}
}

// TestParseStyleString_CaseInsensitive confirms that "#FF0000" and "#ff0000"
// produce identical colors.
func TestParseStyleString_CaseInsensitive(t *testing.T) {
	upper, err := ParseStyleString("#FF0000")
	if err != nil {
		t.Fatalf("unexpected error for uppercase hex: %v", err)
	}

	lower, err := ParseStyleString("#ff0000")
	if err != nil {
		t.Fatalf("unexpected error for lowercase hex: %v", err)
	}

	fgUpper, _, _ := upper.Decompose()
	fgLower, _, _ := lower.Decompose()

	if fgUpper != fgLower {
		t.Errorf("case-insensitive mismatch: upper=%v lower=%v", fgUpper, fgLower)
	}
}

// TestParseStyleString_BoldOnlyNoColorAttributes confirms that when only bold
// is requested, no spurious underline is present.
func TestParseStyleString_BoldOnlyNoUnderline(t *testing.T) {
	style, err := ParseStyleString("#ffffff:#000000:bold")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, _, attrs := style.Decompose()
	if attrs&tcell.AttrUnderline != 0 {
		t.Error("underline should not be set when only bold was requested")
	}
}

// TestParseStyleString_InvalidBgHex confirms that a bad background hex returns
// an error even when the foreground is valid.
func TestParseStyleString_InvalidBgHex(t *testing.T) {
	_, err := ParseStyleString("#ff0000:#zzzzzz")
	if err == nil {
		t.Error("expected an error for invalid background hex, got nil")
	}
}

// ---------------------------------------------------------------------------
// LoadConfig tests
// ---------------------------------------------------------------------------

// writeTempJSON creates a temporary JSON file containing content and returns
// its path. The file is removed when the test finishes.
func writeTempJSON(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "theme.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write temp JSON: %v", err)
	}
	return path
}

// TestLoadConfig_BaseWithFgOverride confirms that a valid config clones the
// named base theme and applies a foreground-only override.
func TestLoadConfig_BaseWithFgOverride(t *testing.T) {
	path := writeTempJSON(t, `{"base":"borland-blue","overrides":{"WindowBackground":"#00ff00"}}`)

	scheme, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if scheme == nil {
		t.Fatal("expected a non-nil ColorScheme")
	}

	fg, _, _ := scheme.WindowBackground.Decompose()
	wantFg := tcell.NewRGBColor(0, 255, 0)
	if fg != wantFg {
		t.Errorf("WindowBackground foreground: got %v, want %v", fg, wantFg)
	}
}

// TestLoadConfig_BaseWithFgAndBgOverride confirms that a fg:bg override sets
// both foreground and background on the target field.
func TestLoadConfig_BaseWithFgAndBgOverride(t *testing.T) {
	path := writeTempJSON(t, `{"base":"borland-blue","overrides":{"WindowBackground":"#00ff00:#000000"}}`)

	scheme, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if scheme == nil {
		t.Fatal("expected a non-nil ColorScheme")
	}

	fg, bg, _ := scheme.WindowBackground.Decompose()
	wantFg := tcell.NewRGBColor(0, 255, 0)
	wantBg := tcell.NewRGBColor(0, 0, 0)

	if fg != wantFg {
		t.Errorf("WindowBackground foreground: got %v, want %v", fg, wantFg)
	}
	if bg != wantBg {
		t.Errorf("WindowBackground background: got %v, want %v", bg, wantBg)
	}
}

// TestLoadConfig_UnknownBaseTheme confirms that a config referencing an
// unregistered base theme returns an error mentioning "unknown base theme".
func TestLoadConfig_UnknownBaseTheme(t *testing.T) {
	path := writeTempJSON(t, `{"base":"nonexistent"}`)

	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected an error for unknown base theme, got nil")
	}
	if !strings.Contains(err.Error(), "unknown base theme") {
		t.Errorf("error message %q does not mention \"unknown base theme\"", err.Error())
	}
}

// TestLoadConfig_UnknownOverrideField confirms that an override targeting a
// field name that does not exist in ColorScheme returns an error mentioning
// "unknown color scheme field".
func TestLoadConfig_UnknownOverrideField(t *testing.T) {
	path := writeTempJSON(t, `{"base":"borland-blue","overrides":{"FakeField":"#000000"}}`)

	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected an error for unknown field, got nil")
	}
	if !strings.Contains(err.Error(), "unknown color scheme field") {
		t.Errorf("error message %q does not mention \"unknown color scheme field\"", err.Error())
	}
}

// TestLoadConfig_InvalidJSON confirms that a file with malformed JSON returns
// an error.
func TestLoadConfig_InvalidJSON(t *testing.T) {
	path := writeTempJSON(t, `{this is not valid json}`)

	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected an error for invalid JSON, got nil")
	}
}

// TestLoadConfig_FileNotFound confirms that a non-existent path returns
// (nil, nil) — missing config is silently ignored.
func TestLoadConfig_FileNotFound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does_not_exist.json")

	scheme, err := LoadConfig(path)
	if err != nil {
		t.Errorf("expected nil error for missing file, got: %v", err)
	}
	if scheme != nil {
		t.Errorf("expected nil scheme for missing file, got non-nil")
	}
}

// TestLoadConfig_NonOverriddenFieldsRetainBaseValues confirms that fields not
// listed in overrides keep the base theme's original values after LoadConfig.
func TestLoadConfig_NonOverriddenFieldsRetainBaseValues(t *testing.T) {
	// Override only WindowBackground; every other field should still match
	// BorlandBlue's values.
	path := writeTempJSON(t, `{"base":"borland-blue","overrides":{"WindowBackground":"#00ff00"}}`)

	scheme, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if scheme == nil {
		t.Fatal("expected a non-nil ColorScheme")
	}

	base := BorlandBlue

	// Spot-check several unmodified fields.
	type check struct {
		name string
		got  tcell.Style
		want tcell.Style
	}
	checks := []check{
		{"WindowFrameActive", scheme.WindowFrameActive, base.WindowFrameActive},
		{"DialogBackground", scheme.DialogBackground, base.DialogBackground},
		{"ButtonNormal", scheme.ButtonNormal, base.ButtonNormal},
		{"MenuNormal", scheme.MenuNormal, base.MenuNormal},
		{"StatusShortcut", scheme.StatusShortcut, base.StatusShortcut},
	}

	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, c.got, c.want)
		}
	}
}

// TestLoadConfig_ReturnedSchemeIsAClone confirms that the returned ColorScheme
// is a distinct copy and that mutating it does not alter the base theme.
func TestLoadConfig_ReturnedSchemeIsAClone(t *testing.T) {
	path := writeTempJSON(t, `{"base":"borland-blue","overrides":{"WindowBackground":"#00ff00"}}`)

	scheme, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if scheme == nil {
		t.Fatal("expected a non-nil ColorScheme")
	}

	// Mutate the returned scheme.
	original := BorlandBlue.DialogBackground
	scheme.DialogBackground = tcell.StyleDefault.Foreground(tcell.ColorRed)

	// BorlandBlue must be unchanged.
	if BorlandBlue.DialogBackground != original {
		t.Error("mutating the returned scheme altered BorlandBlue (not a clone)")
	}
}

// TestLoadConfig_NoOverrides confirms a config with an empty overrides map
// produces a scheme identical to the base theme in all fields.
func TestLoadConfig_NoOverrides(t *testing.T) {
	path := writeTempJSON(t, `{"base":"borland-blue","overrides":{}}`)

	scheme, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if scheme == nil {
		t.Fatal("expected a non-nil ColorScheme")
	}

	base := BorlandBlue

	if scheme.WindowBackground != base.WindowBackground {
		t.Error("WindowBackground differs from base with no overrides")
	}
	if scheme.ButtonNormal != base.ButtonNormal {
		t.Error("ButtonNormal differs from base with no overrides")
	}
	if scheme.MenuNormal != base.MenuNormal {
		t.Error("MenuNormal differs from base with no overrides")
	}
}

// ---------------------------------------------------------------------------
// DefaultConfigPath tests
// ---------------------------------------------------------------------------

// TestDefaultConfigPath_ContainsHomeDir confirms the returned path is rooted
// under the user's home directory.
func TestDefaultConfigPath_ContainsHomeDir(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory on this platform")
	}

	got := DefaultConfigPath()
	if got == "" {
		t.Fatal("DefaultConfigPath returned empty string but home dir is available")
	}
	if !strings.HasPrefix(got, home) {
		t.Errorf("path %q does not start with home dir %q", got, home)
	}
}

// TestDefaultConfigPath_ExactSuffix confirms the path ends with the expected
// relative suffix.
func TestDefaultConfigPath_ExactSuffix(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory on this platform")
	}

	want := filepath.Join(home, ".config", "turboview", "theme.json")
	got := DefaultConfigPath()

	if got != want {
		t.Errorf("DefaultConfigPath: got %q, want %q", got, want)
	}
}

// TestDefaultConfigPath_Structure confirms the path components follow the
// XDG-like convention: $HOME/.config/turboview/theme.json.
func TestDefaultConfigPath_Structure(t *testing.T) {
	got := DefaultConfigPath()
	if got == "" {
		// Acceptable only if home dir is genuinely unavailable.
		_, err := os.UserHomeDir()
		if err == nil {
			t.Error("DefaultConfigPath returned empty string but home dir is available")
		}
		return
	}

	// The path must end with the correct filename.
	if filepath.Base(got) != "theme.json" {
		t.Errorf("filename: got %q, want \"theme.json\"", filepath.Base(got))
	}

	// The directory portion must end with "turboview".
	dir := filepath.Dir(got)
	if filepath.Base(dir) != "turboview" {
		t.Errorf("parent dir: got %q, want \"turboview\"", filepath.Base(dir))
	}

	// The grandparent must be ".config".
	grandDir := filepath.Dir(dir)
	if filepath.Base(grandDir) != ".config" {
		t.Errorf("grandparent dir: got %q, want \".config\"", filepath.Base(grandDir))
	}
}
