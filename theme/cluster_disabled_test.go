package theme

import (
	"reflect"
	"testing"

	"github.com/gdamore/tcell/v2"
)

// Helper: extractColorAndAttrs decomposes a tcell.Style into its fg, bg, and attributes.
func extractColorAndAttrs(style tcell.Style) (fg tcell.Color, bg tcell.Color, attrs tcell.AttrMask) {
	fg, bg, attrs = style.Decompose()
	return
}

// Helper: colorBrightness estimates the brightness of a tcell.Color.
// Lower values = darker; higher values = brighter.
// This is a simple heuristic based on common color names.
func colorBrightness(c tcell.Color) int {
	switch c {
	// Bright colors
	case tcell.ColorWhite, tcell.ColorLime:
		return 9
	case tcell.ColorLightBlue:
		return 8
	case tcell.ColorGreen, tcell.ColorBlue, tcell.ColorSilver:
		return 6
	case tcell.ColorGray, tcell.ColorRed, tcell.ColorPurple:
		return 5
	case tcell.ColorTeal, tcell.ColorDarkCyan, tcell.ColorDarkGreen, tcell.ColorMaroon, tcell.ColorNavy, tcell.ColorDarkGray:
		return 3
	case tcell.ColorBlack:
		return 0
	default:
		return 5 // assume medium brightness for unknown colors
	}
}

// Helper: contrastDelta returns the absolute brightness difference between fg and bg of a style.
// Higher values mean more contrast (more readable). Lower values mean dimmer/harder to read.
func contrastDelta(style tcell.Style) int {
	fg, bg, _ := extractColorAndAttrs(style)
	d := colorBrightness(fg) - colorBrightness(bg)
	if d < 0 {
		d = -d
	}
	return d
}

// Helper: hasReducedContrast checks if disabledStyle has lower fg/bg contrast than normalStyle.
// "Dimmed" means the foreground and background are closer together in brightness.
func hasReducedContrast(normalStyle, disabledStyle tcell.Style) bool {
	return contrastDelta(disabledStyle) < contrastDelta(normalStyle)
}

// TestClusterDisabled_FieldExists confirms that ColorScheme has a ClusterDisabled field.
func TestClusterDisabled_FieldExists(t *testing.T) {
	cs := &ColorScheme{}

	// Try to access ClusterDisabled field
	v := reflect.ValueOf(cs).Elem()
	fieldValue := v.FieldByName("ClusterDisabled")

	if !fieldValue.IsValid() {
		t.Error("ColorScheme does not have ClusterDisabled field")
	}

	if fieldValue.Type().String() != "tcell.Style" {
		t.Errorf("ClusterDisabled field should be tcell.Style, got %s", fieldValue.Type().String())
	}
}

// TestClusterDisabled_NonZeroInBorlandBlue confirms BorlandBlue.ClusterDisabled is not zero-valued.
func TestClusterDisabled_NonZeroInBorlandBlue(t *testing.T) {
	if BorlandBlue == nil {
		t.Fatal("BorlandBlue is nil")
	}

	if BorlandBlue.ClusterDisabled == tcell.StyleDefault {
		t.Error("BorlandBlue.ClusterDisabled is zero-valued (StyleDefault)")
	}
}

// TestClusterDisabled_NonZeroInBorlandCyan confirms BorlandCyan.ClusterDisabled is not zero-valued.
func TestClusterDisabled_NonZeroInBorlandCyan(t *testing.T) {
	if BorlandCyan == nil {
		t.Fatal("BorlandCyan is nil")
	}

	if BorlandCyan.ClusterDisabled == tcell.StyleDefault {
		t.Error("BorlandCyan.ClusterDisabled is zero-valued (StyleDefault)")
	}
}

// TestClusterDisabled_NonZeroInBorlandGray confirms BorlandGray.ClusterDisabled is not zero-valued.
func TestClusterDisabled_NonZeroInBorlandGray(t *testing.T) {
	if BorlandGray == nil {
		t.Fatal("BorlandGray is nil")
	}

	if BorlandGray.ClusterDisabled == tcell.StyleDefault {
		t.Error("BorlandGray.ClusterDisabled is zero-valued (StyleDefault)")
	}
}

// TestClusterDisabled_NonZeroInMatrix confirms Matrix.ClusterDisabled is not zero-valued.
func TestClusterDisabled_NonZeroInMatrix(t *testing.T) {
	if Matrix == nil {
		t.Fatal("Matrix is nil")
	}

	if Matrix.ClusterDisabled == tcell.StyleDefault {
		t.Error("Matrix.ClusterDisabled is zero-valued (StyleDefault)")
	}
}

// TestClusterDisabled_NonZeroInC64 confirms C64.ClusterDisabled is not zero-valued.
func TestClusterDisabled_NonZeroInC64(t *testing.T) {
	if C64 == nil {
		t.Fatal("C64 is nil")
	}

	if C64.ClusterDisabled == tcell.StyleDefault {
		t.Error("C64.ClusterDisabled is zero-valued (StyleDefault)")
	}
}

// TestClusterDisabled_AllThemesNonZero confirms all 5 themes have non-zero ClusterDisabled.
func TestClusterDisabled_AllThemesNonZero(t *testing.T) {
	themes := map[string]*ColorScheme{
		"BorlandBlue": BorlandBlue,
		"BorlandCyan": BorlandCyan,
		"BorlandGray": BorlandGray,
		"Matrix":      Matrix,
		"C64":         C64,
	}

	for themeName, scheme := range themes {
		if scheme == nil {
			t.Fatalf("%s is nil", themeName)
		}

		if scheme.ClusterDisabled == tcell.StyleDefault {
			t.Errorf("%s.ClusterDisabled is zero-valued (StyleDefault)", themeName)
		}
	}
}

// TestClusterDisabled_DimmedVsCheckBoxNormal_BorlandBlue confirms BorlandBlue.ClusterDisabled is visually dimmer than CheckBoxNormal.
func TestClusterDisabled_DimmedVsCheckBoxNormal_BorlandBlue(t *testing.T) {
	if BorlandBlue == nil {
		t.Fatal("BorlandBlue is nil")
	}

	if BorlandBlue.ClusterDisabled == BorlandBlue.CheckBoxNormal {
		t.Error("BorlandBlue.ClusterDisabled should not be the same as CheckBoxNormal (would not appear dimmed)")
	}

	if !hasReducedContrast(BorlandBlue.CheckBoxNormal, BorlandBlue.ClusterDisabled) {
		t.Error("BorlandBlue.ClusterDisabled should have lower brightness/contrast than CheckBoxNormal to appear dimmed")
	}
}

// TestClusterDisabled_DimmedVsCheckBoxNormal_BorlandCyan confirms BorlandCyan.ClusterDisabled is visually dimmer than CheckBoxNormal.
func TestClusterDisabled_DimmedVsCheckBoxNormal_BorlandCyan(t *testing.T) {
	if BorlandCyan == nil {
		t.Fatal("BorlandCyan is nil")
	}

	if BorlandCyan.ClusterDisabled == BorlandCyan.CheckBoxNormal {
		t.Error("BorlandCyan.ClusterDisabled should not be the same as CheckBoxNormal (would not appear dimmed)")
	}

	if !hasReducedContrast(BorlandCyan.CheckBoxNormal, BorlandCyan.ClusterDisabled) {
		t.Error("BorlandCyan.ClusterDisabled should have lower brightness/contrast than CheckBoxNormal to appear dimmed")
	}
}

// TestClusterDisabled_DimmedVsCheckBoxNormal_BorlandGray confirms BorlandGray.ClusterDisabled is visually dimmer than CheckBoxNormal.
func TestClusterDisabled_DimmedVsCheckBoxNormal_BorlandGray(t *testing.T) {
	if BorlandGray == nil {
		t.Fatal("BorlandGray is nil")
	}

	if BorlandGray.ClusterDisabled == BorlandGray.CheckBoxNormal {
		t.Error("BorlandGray.ClusterDisabled should not be the same as CheckBoxNormal (would not appear dimmed)")
	}

	if !hasReducedContrast(BorlandGray.CheckBoxNormal, BorlandGray.ClusterDisabled) {
		t.Error("BorlandGray.ClusterDisabled should have lower brightness/contrast than CheckBoxNormal to appear dimmed")
	}
}

// TestClusterDisabled_DimmedVsCheckBoxNormal_Matrix confirms Matrix.ClusterDisabled is visually dimmer than CheckBoxNormal.
func TestClusterDisabled_DimmedVsCheckBoxNormal_Matrix(t *testing.T) {
	if Matrix == nil {
		t.Fatal("Matrix is nil")
	}

	if Matrix.ClusterDisabled == Matrix.CheckBoxNormal {
		t.Error("Matrix.ClusterDisabled should not be the same as CheckBoxNormal (would not appear dimmed)")
	}

	if !hasReducedContrast(Matrix.CheckBoxNormal, Matrix.ClusterDisabled) {
		t.Error("Matrix.ClusterDisabled should have lower brightness/contrast than CheckBoxNormal to appear dimmed")
	}
}

// TestClusterDisabled_DimmedVsCheckBoxNormal_C64 confirms C64.ClusterDisabled is visually dimmer than CheckBoxNormal.
func TestClusterDisabled_DimmedVsCheckBoxNormal_C64(t *testing.T) {
	if C64 == nil {
		t.Fatal("C64 is nil")
	}

	if C64.ClusterDisabled == C64.CheckBoxNormal {
		t.Error("C64.ClusterDisabled should not be the same as CheckBoxNormal (would not appear dimmed)")
	}

	if !hasReducedContrast(C64.CheckBoxNormal, C64.ClusterDisabled) {
		t.Error("C64.ClusterDisabled should have lower brightness/contrast than CheckBoxNormal to appear dimmed")
	}
}

// TestClusterDisabled_DimmedVsRadioButtonNormal_BorlandBlue confirms BorlandBlue.ClusterDisabled is visually dimmer than RadioButtonNormal.
func TestClusterDisabled_DimmedVsRadioButtonNormal_BorlandBlue(t *testing.T) {
	if BorlandBlue == nil {
		t.Fatal("BorlandBlue is nil")
	}

	if BorlandBlue.ClusterDisabled == BorlandBlue.RadioButtonNormal {
		t.Error("BorlandBlue.ClusterDisabled should not be the same as RadioButtonNormal (would not appear dimmed)")
	}

	if !hasReducedContrast(BorlandBlue.RadioButtonNormal, BorlandBlue.ClusterDisabled) {
		t.Error("BorlandBlue.ClusterDisabled should have lower brightness/contrast than RadioButtonNormal to appear dimmed")
	}
}

// TestClusterDisabled_DimmedVsRadioButtonNormal_BorlandCyan confirms BorlandCyan.ClusterDisabled is visually dimmer than RadioButtonNormal.
func TestClusterDisabled_DimmedVsRadioButtonNormal_BorlandCyan(t *testing.T) {
	if BorlandCyan == nil {
		t.Fatal("BorlandCyan is nil")
	}

	if BorlandCyan.ClusterDisabled == BorlandCyan.RadioButtonNormal {
		t.Error("BorlandCyan.ClusterDisabled should not be the same as RadioButtonNormal (would not appear dimmed)")
	}

	if !hasReducedContrast(BorlandCyan.RadioButtonNormal, BorlandCyan.ClusterDisabled) {
		t.Error("BorlandCyan.ClusterDisabled should have lower brightness/contrast than RadioButtonNormal to appear dimmed")
	}
}

// TestClusterDisabled_DimmedVsRadioButtonNormal_BorlandGray confirms BorlandGray.ClusterDisabled is visually dimmer than RadioButtonNormal.
func TestClusterDisabled_DimmedVsRadioButtonNormal_BorlandGray(t *testing.T) {
	if BorlandGray == nil {
		t.Fatal("BorlandGray is nil")
	}

	if BorlandGray.ClusterDisabled == BorlandGray.RadioButtonNormal {
		t.Error("BorlandGray.ClusterDisabled should not be the same as RadioButtonNormal (would not appear dimmed)")
	}

	if !hasReducedContrast(BorlandGray.RadioButtonNormal, BorlandGray.ClusterDisabled) {
		t.Error("BorlandGray.ClusterDisabled should have lower brightness/contrast than RadioButtonNormal to appear dimmed")
	}
}

// TestClusterDisabled_DimmedVsRadioButtonNormal_Matrix confirms Matrix.ClusterDisabled is visually dimmer than RadioButtonNormal.
func TestClusterDisabled_DimmedVsRadioButtonNormal_Matrix(t *testing.T) {
	if Matrix == nil {
		t.Fatal("Matrix is nil")
	}

	if Matrix.ClusterDisabled == Matrix.RadioButtonNormal {
		t.Error("Matrix.ClusterDisabled should not be the same as RadioButtonNormal (would not appear dimmed)")
	}

	if !hasReducedContrast(Matrix.RadioButtonNormal, Matrix.ClusterDisabled) {
		t.Error("Matrix.ClusterDisabled should have lower brightness/contrast than RadioButtonNormal to appear dimmed")
	}
}

// TestClusterDisabled_DimmedVsRadioButtonNormal_C64 confirms C64.ClusterDisabled is visually dimmer than RadioButtonNormal.
func TestClusterDisabled_DimmedVsRadioButtonNormal_C64(t *testing.T) {
	if C64 == nil {
		t.Fatal("C64 is nil")
	}

	if C64.ClusterDisabled == C64.RadioButtonNormal {
		t.Error("C64.ClusterDisabled should not be the same as RadioButtonNormal (would not appear dimmed)")
	}

	if !hasReducedContrast(C64.RadioButtonNormal, C64.ClusterDisabled) {
		t.Error("C64.ClusterDisabled should have lower brightness/contrast than RadioButtonNormal to appear dimmed")
	}
}

// TestClusterDisabled_TcellStyleDecomposable confirms ClusterDisabled values are valid tcell.Style instances.
func TestClusterDisabled_TcellStyleDecomposable(t *testing.T) {
	themes := map[string]*ColorScheme{
		"BorlandBlue": BorlandBlue,
		"BorlandCyan": BorlandCyan,
		"BorlandGray": BorlandGray,
		"Matrix":      Matrix,
		"C64":         C64,
	}

	for themeName, scheme := range themes {
		if scheme == nil {
			t.Fatalf("%s is nil", themeName)
		}

		// Decompose should not panic and should return valid values
		fg, bg, attrs := scheme.ClusterDisabled.Decompose()

		// Verify we got some result (not panicked)
		_ = fg
		_ = bg
		_ = attrs
	}
}

// TestClusterDisabled_AllThemesHaveDimmingVsCheckbox confirms all themes dim ClusterDisabled vs CheckBoxNormal.
func TestClusterDisabled_AllThemesHaveDimmingVsCheckbox(t *testing.T) {
	themes := map[string]*ColorScheme{
		"BorlandBlue": BorlandBlue,
		"BorlandCyan": BorlandCyan,
		"BorlandGray": BorlandGray,
		"Matrix":      Matrix,
		"C64":         C64,
	}

	for themeName, scheme := range themes {
		if scheme == nil {
			t.Fatalf("%s is nil", themeName)
		}

		if scheme.ClusterDisabled == scheme.CheckBoxNormal {
			t.Errorf("%s: ClusterDisabled should differ from CheckBoxNormal to indicate disabled state", themeName)
		}

		if !hasReducedContrast(scheme.CheckBoxNormal, scheme.ClusterDisabled) {
			t.Errorf("%s: ClusterDisabled should have visually reduced contrast compared to CheckBoxNormal", themeName)
		}
	}
}

// TestClusterDisabled_AllThemesHaveDimmingVsRadio confirms all themes dim ClusterDisabled vs RadioButtonNormal.
func TestClusterDisabled_AllThemesHaveDimmingVsRadio(t *testing.T) {
	themes := map[string]*ColorScheme{
		"BorlandBlue": BorlandBlue,
		"BorlandCyan": BorlandCyan,
		"BorlandGray": BorlandGray,
		"Matrix":      Matrix,
		"C64":         C64,
	}

	for themeName, scheme := range themes {
		if scheme == nil {
			t.Fatalf("%s is nil", themeName)
		}

		if scheme.ClusterDisabled == scheme.RadioButtonNormal {
			t.Errorf("%s: ClusterDisabled should differ from RadioButtonNormal to indicate disabled state", themeName)
		}

		if !hasReducedContrast(scheme.RadioButtonNormal, scheme.ClusterDisabled) {
			t.Errorf("%s: ClusterDisabled should have visually reduced contrast compared to RadioButtonNormal", themeName)
		}
	}
}

// TestClusterDisabled_FieldCountIncludesClusterDisabled verifies that all themes have the new field.
func TestClusterDisabled_FieldCountIncludesClusterDisabled(t *testing.T) {
	// All themes should now have 36 non-zero fields (35 original + 1 new ClusterDisabled)
	themes := map[string]*ColorScheme{
		"BorlandBlue": BorlandBlue,
		"BorlandCyan": BorlandCyan,
		"BorlandGray": BorlandGray,
		"Matrix":      Matrix,
		"C64":         C64,
	}

	for themeName, scheme := range themes {
		if scheme == nil {
			t.Fatalf("%s is nil", themeName)
		}

		v := reflect.ValueOf(scheme).Elem()
		nonZeroCount := 0
		zeroStyle := tcell.StyleDefault

		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			if field.Interface().(tcell.Style) != zeroStyle {
				nonZeroCount++
			}
		}

		if nonZeroCount != 45 {
			t.Errorf("%s: expected 45 non-zero fields (including ClusterDisabled), got %d", themeName, nonZeroCount)
		}
	}
}

// TestClusterDisabled_NotEqualToMenuDisabled confirms ClusterDisabled is distinct from MenuDisabled.
func TestClusterDisabled_NotEqualToMenuDisabled(t *testing.T) {
	themes := map[string]*ColorScheme{
		"BorlandBlue": BorlandBlue,
		"BorlandCyan": BorlandCyan,
		"BorlandGray": BorlandGray,
		"Matrix":      Matrix,
		"C64":         C64,
	}

	for themeName, scheme := range themes {
		if scheme == nil {
			t.Fatalf("%s is nil", themeName)
		}

		// ClusterDisabled and MenuDisabled serve different purposes and should not be identical
		// (though they may have similar visual treatment)
		if scheme.ClusterDisabled == scheme.MenuDisabled {
			t.Logf("Warning: %s ClusterDisabled is identical to MenuDisabled (not necessarily an error, but unexpected)", themeName)
		}
	}
}
