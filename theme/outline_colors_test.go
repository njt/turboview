package theme

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// outline_colors_test.go — tests for Batch 9 Task 1: Outline color scheme entries.
//
// Every assertion cites the spec requirement it verifies.
// Each test covers exactly one behaviour.
//
// Spec requirements tested:
//   ColorScheme struct gains 4 new fields:
//     1. OutlineNormal tcell.Style
//     2. OutlineFocused tcell.Style
//     3. OutlineSelected tcell.Style
//     4. OutlineCollapsed tcell.Style
//
//   All five themes must include values:
//     5. BorlandBlue.OutlineNormal is non-zero
//     6. BorlandBlue.OutlineFocused is non-zero
//     7. BorlandBlue.OutlineSelected is non-zero
//     8. BorlandBlue.OutlineCollapsed is non-zero
//     9. BorlandCyan.OutlineNormal is non-zero
//    10. BorlandCyan.OutlineFocused is non-zero
//    11. BorlandCyan.OutlineSelected is non-zero
//    12. BorlandCyan.OutlineCollapsed is non-zero
//    13. BorlandGray.OutlineNormal is non-zero
//    14. BorlandGray.OutlineFocused is non-zero
//    15. BorlandGray.OutlineSelected is non-zero
//    16. BorlandGray.OutlineCollapsed is non-zero
//    17. C64.OutlineNormal is non-zero
//    18. C64.OutlineFocused is non-zero
//    19. C64.OutlineSelected is non-zero
//    20. C64.OutlineCollapsed is non-zero
//    21. Matrix.OutlineNormal is non-zero
//    22. Matrix.OutlineFocused is non-zero
//    23. Matrix.OutlineSelected is non-zero
//    24. Matrix.OutlineCollapsed is non-zero
//
//   Full-field-count after adding 4 new fields:
//    25. BorlandBlue has 63 non-zero fields (45 base + 18 markdown); all other themes have 45 (41 existing + 4 outline).

// ---------------------------------------------------------------------------
// Requirement 1 — ColorScheme.OutlineNormal field exists and is tcell.Style
// Spec: "ColorScheme struct gains 4 new fields: OutlineNormal tcell.Style"
// ---------------------------------------------------------------------------

// TestColorSchemeHasOutlineNormalField verifies OutlineNormal can be assigned
// and read back on a ColorScheme, confirming the field exists with correct type.
func TestColorSchemeHasOutlineNormalField(t *testing.T) {
	scheme := &ColorScheme{}
	style := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlue)

	scheme.OutlineNormal = style

	if scheme.OutlineNormal != style {
		t.Errorf("OutlineNormal = %v, want %v", scheme.OutlineNormal, style)
	}
}

// TestColorSchemeOutlineNormalIsIndependent verifies OutlineNormal holds its
// value independently of adjacent outline fields.
func TestColorSchemeOutlineNormalIsIndependent(t *testing.T) {
	scheme := &ColorScheme{}
	normal := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlue)
	focused := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack)

	scheme.OutlineNormal = normal
	scheme.OutlineFocused = focused

	if scheme.OutlineNormal != normal {
		t.Error("OutlineNormal changed after assigning OutlineFocused")
	}
}

// ---------------------------------------------------------------------------
// Requirement 2 — ColorScheme.OutlineFocused field exists and is tcell.Style
// Spec: "ColorScheme struct gains 4 new fields: OutlineFocused tcell.Style"
// ---------------------------------------------------------------------------

// TestColorSchemeHasOutlineFocusedField verifies OutlineFocused can be assigned
// and read back.
func TestColorSchemeHasOutlineFocusedField(t *testing.T) {
	scheme := &ColorScheme{}
	style := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack)

	scheme.OutlineFocused = style

	if scheme.OutlineFocused != style {
		t.Errorf("OutlineFocused = %v, want %v", scheme.OutlineFocused, style)
	}
}

// TestColorSchemeOutlineFocusedIsIndependent verifies OutlineFocused is not
// aliased with other outline fields.
func TestColorSchemeOutlineFocusedIsIndependent(t *testing.T) {
	scheme := &ColorScheme{}
	focused := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(tcell.ColorBlack)
	selected := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)

	scheme.OutlineFocused = focused
	scheme.OutlineSelected = selected

	if scheme.OutlineFocused != focused {
		t.Error("OutlineFocused changed after assigning OutlineSelected")
	}
}

// ---------------------------------------------------------------------------
// Requirement 3 — ColorScheme.OutlineSelected field exists and is tcell.Style
// Spec: "ColorScheme struct gains 4 new fields: OutlineSelected tcell.Style"
// ---------------------------------------------------------------------------

// TestColorSchemeHasOutlineSelectedField verifies OutlineSelected can be assigned
// and read back.
func TestColorSchemeHasOutlineSelectedField(t *testing.T) {
	scheme := &ColorScheme{}
	style := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)

	scheme.OutlineSelected = style

	if scheme.OutlineSelected != style {
		t.Errorf("OutlineSelected = %v, want %v", scheme.OutlineSelected, style)
	}
}

// TestColorSchemeOutlineSelectedIsIndependent verifies OutlineSelected is not
// aliased with other outline fields.
func TestColorSchemeOutlineSelectedIsIndependent(t *testing.T) {
	scheme := &ColorScheme{}
	selected := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
	collapsed := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlue)

	scheme.OutlineSelected = selected
	scheme.OutlineCollapsed = collapsed

	if scheme.OutlineSelected != selected {
		t.Error("OutlineSelected changed after assigning OutlineCollapsed")
	}
}

// ---------------------------------------------------------------------------
// Requirement 4 — ColorScheme.OutlineCollapsed field exists and is tcell.Style
// Spec: "ColorScheme struct gains 4 new fields: OutlineCollapsed tcell.Style"
// ---------------------------------------------------------------------------

// TestColorSchemeHasOutlineCollapsedField verifies OutlineCollapsed can be assigned
// and read back.
func TestColorSchemeHasOutlineCollapsedField(t *testing.T) {
	scheme := &ColorScheme{}
	style := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlue)

	scheme.OutlineCollapsed = style

	if scheme.OutlineCollapsed != style {
		t.Errorf("OutlineCollapsed = %v, want %v", scheme.OutlineCollapsed, style)
	}
}

// TestColorSchemeOutlineCollapsedIsIndependent verifies OutlineCollapsed is not
// aliased with other outline fields.
func TestColorSchemeOutlineCollapsedIsIndependent(t *testing.T) {
	scheme := &ColorScheme{}
	collapsed := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlue)
	normal := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlue)

	scheme.OutlineCollapsed = collapsed
	scheme.OutlineNormal = normal

	if scheme.OutlineCollapsed != collapsed {
		t.Error("OutlineCollapsed changed after assigning OutlineNormal")
	}
}

// ---------------------------------------------------------------------------
// Requirements 5-8 — BorlandBlue outline fields
// Spec: "All 5 theme files must include values for these 4 new entries"
//        borland.go specifies BorlandBlue colors.
// ---------------------------------------------------------------------------

// TestBorlandBlueOutlineNormalIsNonZero verifies BorlandBlue.OutlineNormal is set.
func TestBorlandBlueOutlineNormalIsNonZero(t *testing.T) {
	if BorlandBlue == nil {
		t.Fatal("BorlandBlue is nil")
	}
	if BorlandBlue.OutlineNormal == tcell.StyleDefault {
		t.Error("BorlandBlue.OutlineNormal is zero (StyleDefault); spec requires a non-zero value")
	}
}

// TestBorlandBlueOutlineFocusedIsNonZero verifies BorlandBlue.OutlineFocused is set.
func TestBorlandBlueOutlineFocusedIsNonZero(t *testing.T) {
	if BorlandBlue == nil {
		t.Fatal("BorlandBlue is nil")
	}
	if BorlandBlue.OutlineFocused == tcell.StyleDefault {
		t.Error("BorlandBlue.OutlineFocused is zero (StyleDefault); spec requires a non-zero value")
	}
}

// TestBorlandBlueOutlineSelectedIsNonZero verifies BorlandBlue.OutlineSelected is set.
func TestBorlandBlueOutlineSelectedIsNonZero(t *testing.T) {
	if BorlandBlue == nil {
		t.Fatal("BorlandBlue is nil")
	}
	if BorlandBlue.OutlineSelected == tcell.StyleDefault {
		t.Error("BorlandBlue.OutlineSelected is zero (StyleDefault); spec requires a non-zero value")
	}
}

// TestBorlandBlueOutlineCollapsedIsNonZero verifies BorlandBlue.OutlineCollapsed is set.
func TestBorlandBlueOutlineCollapsedIsNonZero(t *testing.T) {
	if BorlandBlue == nil {
		t.Fatal("BorlandBlue is nil")
	}
	if BorlandBlue.OutlineCollapsed == tcell.StyleDefault {
		t.Error("BorlandBlue.OutlineCollapsed is zero (StyleDefault); spec requires a non-zero value")
	}
}

// ---------------------------------------------------------------------------
// Requirements 9-12 — BorlandCyan outline fields
// Spec: "All 5 theme files must include values for these 4 new entries"
//        borland_cyan.go specifies BorlandCyan colors.
// ---------------------------------------------------------------------------

// TestBorlandCyanOutlineNormalIsNonZero verifies BorlandCyan.OutlineNormal is set.
func TestBorlandCyanOutlineNormalIsNonZero(t *testing.T) {
	if BorlandCyan == nil {
		t.Fatal("BorlandCyan is nil")
	}
	if BorlandCyan.OutlineNormal == tcell.StyleDefault {
		t.Error("BorlandCyan.OutlineNormal is zero (StyleDefault); spec requires a non-zero value")
	}
}

// TestBorlandCyanOutlineFocusedIsNonZero verifies BorlandCyan.OutlineFocused is set.
func TestBorlandCyanOutlineFocusedIsNonZero(t *testing.T) {
	if BorlandCyan == nil {
		t.Fatal("BorlandCyan is nil")
	}
	if BorlandCyan.OutlineFocused == tcell.StyleDefault {
		t.Error("BorlandCyan.OutlineFocused is zero (StyleDefault); spec requires a non-zero value")
	}
}

// TestBorlandCyanOutlineSelectedIsNonZero verifies BorlandCyan.OutlineSelected is set.
func TestBorlandCyanOutlineSelectedIsNonZero(t *testing.T) {
	if BorlandCyan == nil {
		t.Fatal("BorlandCyan is nil")
	}
	if BorlandCyan.OutlineSelected == tcell.StyleDefault {
		t.Error("BorlandCyan.OutlineSelected is zero (StyleDefault); spec requires a non-zero value")
	}
}

// TestBorlandCyanOutlineCollapsedIsNonZero verifies BorlandCyan.OutlineCollapsed is set.
func TestBorlandCyanOutlineCollapsedIsNonZero(t *testing.T) {
	if BorlandCyan == nil {
		t.Fatal("BorlandCyan is nil")
	}
	if BorlandCyan.OutlineCollapsed == tcell.StyleDefault {
		t.Error("BorlandCyan.OutlineCollapsed is zero (StyleDefault); spec requires a non-zero value")
	}
}

// ---------------------------------------------------------------------------
// Requirements 13-16 — BorlandGray outline fields
// Spec: "All 5 theme files must include values for these 4 new entries"
//        borland_gray.go specifies BorlandGray colors.
// ---------------------------------------------------------------------------

// TestBorlandGrayOutlineNormalIsNonZero verifies BorlandGray.OutlineNormal is set.
func TestBorlandGrayOutlineNormalIsNonZero(t *testing.T) {
	if BorlandGray == nil {
		t.Fatal("BorlandGray is nil")
	}
	if BorlandGray.OutlineNormal == tcell.StyleDefault {
		t.Error("BorlandGray.OutlineNormal is zero (StyleDefault); spec requires a non-zero value")
	}
}

// TestBorlandGrayOutlineFocusedIsNonZero verifies BorlandGray.OutlineFocused is set.
func TestBorlandGrayOutlineFocusedIsNonZero(t *testing.T) {
	if BorlandGray == nil {
		t.Fatal("BorlandGray is nil")
	}
	if BorlandGray.OutlineFocused == tcell.StyleDefault {
		t.Error("BorlandGray.OutlineFocused is zero (StyleDefault); spec requires a non-zero value")
	}
}

// TestBorlandGrayOutlineSelectedIsNonZero verifies BorlandGray.OutlineSelected is set.
func TestBorlandGrayOutlineSelectedIsNonZero(t *testing.T) {
	if BorlandGray == nil {
		t.Fatal("BorlandGray is nil")
	}
	if BorlandGray.OutlineSelected == tcell.StyleDefault {
		t.Error("BorlandGray.OutlineSelected is zero (StyleDefault); spec requires a non-zero value")
	}
}

// TestBorlandGrayOutlineCollapsedIsNonZero verifies BorlandGray.OutlineCollapsed is set.
func TestBorlandGrayOutlineCollapsedIsNonZero(t *testing.T) {
	if BorlandGray == nil {
		t.Fatal("BorlandGray is nil")
	}
	if BorlandGray.OutlineCollapsed == tcell.StyleDefault {
		t.Error("BorlandGray.OutlineCollapsed is zero (StyleDefault); spec requires a non-zero value")
	}
}

// ---------------------------------------------------------------------------
// Requirements 17-20 — C64 outline fields
// Spec: "All 5 theme files must include values for these 4 new entries"
//        c64.go specifies C64-specific colors.
// ---------------------------------------------------------------------------

// TestC64OutlineNormalIsNonZero verifies C64.OutlineNormal is set.
func TestC64OutlineNormalIsNonZero(t *testing.T) {
	if C64 == nil {
		t.Fatal("C64 is nil")
	}
	if C64.OutlineNormal == tcell.StyleDefault {
		t.Error("C64.OutlineNormal is zero (StyleDefault); spec requires a non-zero value")
	}
}

// TestC64OutlineFocusedIsNonZero verifies C64.OutlineFocused is set.
func TestC64OutlineFocusedIsNonZero(t *testing.T) {
	if C64 == nil {
		t.Fatal("C64 is nil")
	}
	if C64.OutlineFocused == tcell.StyleDefault {
		t.Error("C64.OutlineFocused is zero (StyleDefault); spec requires a non-zero value")
	}
}

// TestC64OutlineSelectedIsNonZero verifies C64.OutlineSelected is set.
func TestC64OutlineSelectedIsNonZero(t *testing.T) {
	if C64 == nil {
		t.Fatal("C64 is nil")
	}
	if C64.OutlineSelected == tcell.StyleDefault {
		t.Error("C64.OutlineSelected is zero (StyleDefault); spec requires a non-zero value")
	}
}

// TestC64OutlineCollapsedIsNonZero verifies C64.OutlineCollapsed is set.
func TestC64OutlineCollapsedIsNonZero(t *testing.T) {
	if C64 == nil {
		t.Fatal("C64 is nil")
	}
	if C64.OutlineCollapsed == tcell.StyleDefault {
		t.Error("C64.OutlineCollapsed is zero (StyleDefault); spec requires a non-zero value")
	}
}

// ---------------------------------------------------------------------------
// Requirements 21-24 — Matrix outline fields
// Spec: "All 5 theme files must include values for these 4 new entries"
//        matrix.go specifies matrix-specific colors.
// ---------------------------------------------------------------------------

// TestMatrixOutlineNormalIsNonZero verifies Matrix.OutlineNormal is set.
func TestMatrixOutlineNormalIsNonZero(t *testing.T) {
	if Matrix == nil {
		t.Fatal("Matrix is nil")
	}
	if Matrix.OutlineNormal == tcell.StyleDefault {
		t.Error("Matrix.OutlineNormal is zero (StyleDefault); spec requires a non-zero value")
	}
}

// TestMatrixOutlineFocusedIsNonZero verifies Matrix.OutlineFocused is set.
func TestMatrixOutlineFocusedIsNonZero(t *testing.T) {
	if Matrix == nil {
		t.Fatal("Matrix is nil")
	}
	if Matrix.OutlineFocused == tcell.StyleDefault {
		t.Error("Matrix.OutlineFocused is zero (StyleDefault); spec requires a non-zero value")
	}
}

// TestMatrixOutlineSelectedIsNonZero verifies Matrix.OutlineSelected is set.
func TestMatrixOutlineSelectedIsNonZero(t *testing.T) {
	if Matrix == nil {
		t.Fatal("Matrix is nil")
	}
	if Matrix.OutlineSelected == tcell.StyleDefault {
		t.Error("Matrix.OutlineSelected is zero (StyleDefault); spec requires a non-zero value")
	}
}

// TestMatrixOutlineCollapsedIsNonZero verifies Matrix.OutlineCollapsed is set.
func TestMatrixOutlineCollapsedIsNonZero(t *testing.T) {
	if Matrix == nil {
		t.Fatal("Matrix is nil")
	}
	if Matrix.OutlineCollapsed == tcell.StyleDefault {
		t.Error("Matrix.OutlineCollapsed is zero (StyleDefault); spec requires a non-zero value")
	}
}

// ---------------------------------------------------------------------------
// TestOutlineColorEntriesAfterRegistration — verifies themes loaded via config
// (Register/Get) also have the 4 outline entries.
// Spec: "All 5 theme files must include values for these 4 new entries"
// ---------------------------------------------------------------------------

// TestOutlineColorEntriesAfterRegistration verifies all 5 registered themes
// (borland-blue, borland-cyan, borland-gray, c64, matrix) have non-zero
// values for all 4 outline color entries.
func TestOutlineColorEntriesAfterRegistration(t *testing.T) {
	themeNames := []string{
		"borland-blue",
		"borland-cyan",
		"borland-gray",
		"matrix",
		"c64",
	}

	for _, name := range themeNames {
		scheme := Get(name)
		if scheme == nil {
			t.Errorf("theme %q is not registered", name)
			continue
		}

		if scheme.OutlineNormal == tcell.StyleDefault {
			t.Errorf("theme %q: OutlineNormal is zero (StyleDefault)", name)
		}
		if scheme.OutlineFocused == tcell.StyleDefault {
			t.Errorf("theme %q: OutlineFocused is zero (StyleDefault)", name)
		}
		if scheme.OutlineSelected == tcell.StyleDefault {
			t.Errorf("theme %q: OutlineSelected is zero (StyleDefault)", name)
		}
		if scheme.OutlineCollapsed == tcell.StyleDefault {
			t.Errorf("theme %q: OutlineCollapsed is zero (StyleDefault)", name)
		}
	}
}
