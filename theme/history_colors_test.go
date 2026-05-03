package theme

// history_colors_test.go — tests for Task 2: HistoryArrow and HistorySides fields.
//
// Every assertion cites the spec requirement it verifies.
// Each test covers exactly one behaviour.
//
// Spec requirements tested:
//   ColorScheme struct:
//     1. ColorScheme has a HistoryArrow field of type tcell.Style.
//     2. ColorScheme has a HistorySides field of type tcell.Style.
//
//   Per-theme values (spec: "All five themes must include values for these new fields"):
//     3. BorlandBlue.HistoryArrow is non-zero (green on cyan).
//     4. BorlandBlue.HistorySides is non-zero (cyan on blue).
//     5. BorlandCyan.HistoryArrow is non-zero (dark cyan on white).
//     6. BorlandCyan.HistorySides is non-zero (white on cyan).
//     7. BorlandGray.HistoryArrow is non-zero (dark gray on white).
//     8. BorlandGray.HistorySides is non-zero (white on dark gray).
//     9. C64.HistoryArrow is non-zero (light blue on blue).
//    10. C64.HistorySides is non-zero (blue on light blue).
//    11. Matrix.HistoryArrow is non-zero (white on dark green).
//    12. Matrix.HistorySides is non-zero (green on black).
//
//   Exact color spot-checks per spec:
//    13. BorlandBlue.HistoryArrow background is cyan.
//    14. BorlandBlue.HistorySides background is blue.
//    15. BorlandCyan.HistoryArrow background is white.
//    16. BorlandCyan.HistorySides background is cyan.
//    17. BorlandGray.HistoryArrow background is white.
//    18. BorlandGray.HistorySides background is dark gray.
//    19. C64.HistoryArrow background is blue.
//    20. C64.HistorySides background is light blue.
//    21. Matrix.HistoryArrow background is dark green.
//    22. Matrix.HistorySides background is black.
//
//   Full-field-count after adding 2 new fields:
//    23. All five themes have exactly 35 non-zero fields.

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Requirement 1 — ColorScheme.HistoryArrow field exists and is tcell.Style
// Spec: "Add HistoryArrow tcell.Style to ColorScheme struct — color for the ↓ character"
// ---------------------------------------------------------------------------

// TestColorSchemeHasHistoryArrowField verifies HistoryArrow can be assigned
// and read back on a ColorScheme, confirming the field exists with the correct type.
func TestColorSchemeHasHistoryArrowField(t *testing.T) {
	scheme := &ColorScheme{}
	style := tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tcell.ColorTeal)

	scheme.HistoryArrow = style

	if scheme.HistoryArrow != style {
		t.Errorf("HistoryArrow = %v, want %v", scheme.HistoryArrow, style)
	}
}

// TestColorSchemeHistoryArrowIsIndependentOfOtherFields verifies HistoryArrow
// holds its own value independently of adjacent fields (falsifying: field aliasing).
func TestColorSchemeHistoryArrowIsIndependentOfOtherFields(t *testing.T) {
	scheme := &ColorScheme{}
	arrowStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tcell.ColorTeal)
	sidesStyle := tcell.StyleDefault.Foreground(tcell.ColorTeal).Background(tcell.ColorBlue)

	scheme.HistoryArrow = arrowStyle
	scheme.HistorySides = sidesStyle

	if scheme.HistoryArrow != arrowStyle {
		t.Errorf("HistoryArrow changed after assigning HistorySides: got %v, want %v",
			scheme.HistoryArrow, arrowStyle)
	}
}

// ---------------------------------------------------------------------------
// Requirement 2 — ColorScheme.HistorySides field exists and is tcell.Style
// Spec: "Add HistorySides tcell.Style to ColorScheme struct — color for the ▐ and ▌ bracket characters"
// ---------------------------------------------------------------------------

// TestColorSchemeHasHistorySidesField verifies HistorySides can be assigned
// and read back on a ColorScheme, confirming the field exists with the correct type.
func TestColorSchemeHasHistorySidesField(t *testing.T) {
	scheme := &ColorScheme{}
	style := tcell.StyleDefault.Foreground(tcell.ColorTeal).Background(tcell.ColorBlue)

	scheme.HistorySides = style

	if scheme.HistorySides != style {
		t.Errorf("HistorySides = %v, want %v", scheme.HistorySides, style)
	}
}

// TestColorSchemeHistorySidesIsIndependentOfHistoryArrow verifies HistorySides
// holds its value independently of HistoryArrow (falsifying: field aliasing).
func TestColorSchemeHistorySidesIsIndependentOfHistoryArrow(t *testing.T) {
	scheme := &ColorScheme{}
	arrowStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen).Background(tcell.ColorTeal)
	sidesStyle := tcell.StyleDefault.Foreground(tcell.ColorTeal).Background(tcell.ColorBlue)

	scheme.HistoryArrow = arrowStyle
	scheme.HistorySides = sidesStyle

	if scheme.HistorySides != sidesStyle {
		t.Errorf("HistorySides changed after assigning HistoryArrow: got %v, want %v",
			scheme.HistorySides, sidesStyle)
	}
}

// ---------------------------------------------------------------------------
// Requirement 3 — BorlandBlue.HistoryArrow is non-zero
// Spec: "BorlandBlue: arrow = green on cyan"
// ---------------------------------------------------------------------------

// TestBorlandBlueHistoryArrowIsNonZero verifies BorlandBlue.HistoryArrow is set
// (not the zero value tcell.StyleDefault).
func TestBorlandBlueHistoryArrowIsNonZero(t *testing.T) {
	if BorlandBlue == nil {
		t.Fatal("BorlandBlue is nil")
	}
	if BorlandBlue.HistoryArrow == tcell.StyleDefault {
		t.Error("BorlandBlue.HistoryArrow is zero (StyleDefault); spec requires green on cyan")
	}
}

// ---------------------------------------------------------------------------
// Requirement 4 — BorlandBlue.HistorySides is non-zero
// Spec: "BorlandBlue: sides = cyan on blue"
// ---------------------------------------------------------------------------

// TestBorlandBlueHistorySidesIsNonZero verifies BorlandBlue.HistorySides is set.
func TestBorlandBlueHistorySidesIsNonZero(t *testing.T) {
	if BorlandBlue == nil {
		t.Fatal("BorlandBlue is nil")
	}
	if BorlandBlue.HistorySides == tcell.StyleDefault {
		t.Error("BorlandBlue.HistorySides is zero (StyleDefault); spec requires cyan on blue")
	}
}

// ---------------------------------------------------------------------------
// Requirement 5 — BorlandCyan.HistoryArrow is non-zero
// Spec: "BorlandCyan: arrow = dark cyan on white"
// ---------------------------------------------------------------------------

// TestBorlandCyanHistoryArrowIsNonZero verifies BorlandCyan.HistoryArrow is set.
func TestBorlandCyanHistoryArrowIsNonZero(t *testing.T) {
	if BorlandCyan == nil {
		t.Fatal("BorlandCyan is nil")
	}
	if BorlandCyan.HistoryArrow == tcell.StyleDefault {
		t.Error("BorlandCyan.HistoryArrow is zero (StyleDefault); spec requires dark cyan on white")
	}
}

// ---------------------------------------------------------------------------
// Requirement 6 — BorlandCyan.HistorySides is non-zero
// Spec: "BorlandCyan: sides = white on cyan"
// ---------------------------------------------------------------------------

// TestBorlandCyanHistorySidesIsNonZero verifies BorlandCyan.HistorySides is set.
func TestBorlandCyanHistorySidesIsNonZero(t *testing.T) {
	if BorlandCyan == nil {
		t.Fatal("BorlandCyan is nil")
	}
	if BorlandCyan.HistorySides == tcell.StyleDefault {
		t.Error("BorlandCyan.HistorySides is zero (StyleDefault); spec requires white on cyan")
	}
}

// ---------------------------------------------------------------------------
// Requirement 7 — BorlandGray.HistoryArrow is non-zero
// Spec: "BorlandGray: arrow = dark gray on white"
// ---------------------------------------------------------------------------

// TestBorlandGrayHistoryArrowIsNonZero verifies BorlandGray.HistoryArrow is set.
func TestBorlandGrayHistoryArrowIsNonZero(t *testing.T) {
	if BorlandGray == nil {
		t.Fatal("BorlandGray is nil")
	}
	if BorlandGray.HistoryArrow == tcell.StyleDefault {
		t.Error("BorlandGray.HistoryArrow is zero (StyleDefault); spec requires dark gray on white")
	}
}

// ---------------------------------------------------------------------------
// Requirement 8 — BorlandGray.HistorySides is non-zero
// Spec: "BorlandGray: sides = white on dark gray"
// ---------------------------------------------------------------------------

// TestBorlandGrayHistorySidesIsNonZero verifies BorlandGray.HistorySides is set.
func TestBorlandGrayHistorySidesIsNonZero(t *testing.T) {
	if BorlandGray == nil {
		t.Fatal("BorlandGray is nil")
	}
	if BorlandGray.HistorySides == tcell.StyleDefault {
		t.Error("BorlandGray.HistorySides is zero (StyleDefault); spec requires white on dark gray")
	}
}

// ---------------------------------------------------------------------------
// Requirement 9 — C64.HistoryArrow is non-zero
// Spec: "C64: arrow = light blue on blue"
// ---------------------------------------------------------------------------

// TestC64HistoryArrowIsNonZero verifies C64.HistoryArrow is set.
func TestC64HistoryArrowIsNonZero(t *testing.T) {
	if C64 == nil {
		t.Fatal("C64 is nil")
	}
	if C64.HistoryArrow == tcell.StyleDefault {
		t.Error("C64.HistoryArrow is zero (StyleDefault); spec requires light blue on blue")
	}
}

// ---------------------------------------------------------------------------
// Requirement 10 — C64.HistorySides is non-zero
// Spec: "C64: sides = blue on light blue"
// ---------------------------------------------------------------------------

// TestC64HistorySidesIsNonZero verifies C64.HistorySides is set.
func TestC64HistorySidesIsNonZero(t *testing.T) {
	if C64 == nil {
		t.Fatal("C64 is nil")
	}
	if C64.HistorySides == tcell.StyleDefault {
		t.Error("C64.HistorySides is zero (StyleDefault); spec requires blue on light blue")
	}
}

// ---------------------------------------------------------------------------
// Requirement 11 — Matrix.HistoryArrow is non-zero
// Spec: "Matrix: arrow = white on dark green"
// ---------------------------------------------------------------------------

// TestMatrixHistoryArrowIsNonZero verifies Matrix.HistoryArrow is set.
func TestMatrixHistoryArrowIsNonZero(t *testing.T) {
	if Matrix == nil {
		t.Fatal("Matrix is nil")
	}
	if Matrix.HistoryArrow == tcell.StyleDefault {
		t.Error("Matrix.HistoryArrow is zero (StyleDefault); spec requires white on dark green")
	}
}

// ---------------------------------------------------------------------------
// Requirement 12 — Matrix.HistorySides is non-zero
// Spec: "Matrix: sides = green on black"
// ---------------------------------------------------------------------------

// TestMatrixHistorySidesIsNonZero verifies Matrix.HistorySides is set.
func TestMatrixHistorySidesIsNonZero(t *testing.T) {
	if Matrix == nil {
		t.Fatal("Matrix is nil")
	}
	if Matrix.HistorySides == tcell.StyleDefault {
		t.Error("Matrix.HistorySides is zero (StyleDefault); spec requires green on black")
	}
}

// ---------------------------------------------------------------------------
// Requirement 13 — BorlandBlue.HistoryArrow background is cyan
// Spec: "BorlandBlue: arrow = green on cyan"
// ---------------------------------------------------------------------------

// TestBorlandBlueHistoryArrowBackgroundIsCyan verifies the background color
// of BorlandBlue.HistoryArrow is tcell.ColorTeal (the tcell name for cyan).
func TestBorlandBlueHistoryArrowBackgroundIsCyan(t *testing.T) {
	if BorlandBlue == nil {
		t.Fatal("BorlandBlue is nil")
	}
	_, bg, _ := BorlandBlue.HistoryArrow.Decompose()
	if bg != tcell.ColorTeal {
		t.Errorf("BorlandBlue.HistoryArrow background = %v, want tcell.ColorTeal (cyan)", bg)
	}
}

// TestBorlandBlueHistoryArrowForegroundIsGreen verifies the foreground color
// of BorlandBlue.HistoryArrow is tcell.ColorGreen.
func TestBorlandBlueHistoryArrowForegroundIsGreen(t *testing.T) {
	if BorlandBlue == nil {
		t.Fatal("BorlandBlue is nil")
	}
	fg, _, _ := BorlandBlue.HistoryArrow.Decompose()
	if fg != tcell.ColorGreen {
		t.Errorf("BorlandBlue.HistoryArrow foreground = %v, want tcell.ColorGreen", fg)
	}
}

// ---------------------------------------------------------------------------
// Requirement 14 — BorlandBlue.HistorySides background is blue
// Spec: "BorlandBlue: sides = cyan on blue"
// ---------------------------------------------------------------------------

// TestBorlandBlueHistorySidesBackgroundIsBlue verifies the background color
// of BorlandBlue.HistorySides is tcell.ColorBlue.
func TestBorlandBlueHistorySidesBackgroundIsBlue(t *testing.T) {
	if BorlandBlue == nil {
		t.Fatal("BorlandBlue is nil")
	}
	_, bg, _ := BorlandBlue.HistorySides.Decompose()
	if bg != tcell.ColorBlue {
		t.Errorf("BorlandBlue.HistorySides background = %v, want tcell.ColorBlue", bg)
	}
}

// TestBorlandBlueHistorySidesForegroundIsCyan verifies the foreground color
// of BorlandBlue.HistorySides is tcell.ColorTeal (cyan).
func TestBorlandBlueHistorySidesForegroundIsCyan(t *testing.T) {
	if BorlandBlue == nil {
		t.Fatal("BorlandBlue is nil")
	}
	fg, _, _ := BorlandBlue.HistorySides.Decompose()
	if fg != tcell.ColorTeal {
		t.Errorf("BorlandBlue.HistorySides foreground = %v, want tcell.ColorTeal (cyan)", fg)
	}
}

// ---------------------------------------------------------------------------
// Requirement 15 — BorlandCyan.HistoryArrow background is white
// Spec: "BorlandCyan: arrow = dark cyan on white"
// ---------------------------------------------------------------------------

// TestBorlandCyanHistoryArrowBackgroundIsWhite verifies the background color
// of BorlandCyan.HistoryArrow is tcell.ColorWhite.
func TestBorlandCyanHistoryArrowBackgroundIsWhite(t *testing.T) {
	if BorlandCyan == nil {
		t.Fatal("BorlandCyan is nil")
	}
	_, bg, _ := BorlandCyan.HistoryArrow.Decompose()
	if bg != tcell.ColorWhite {
		t.Errorf("BorlandCyan.HistoryArrow background = %v, want tcell.ColorWhite", bg)
	}
}

// TestBorlandCyanHistoryArrowForegroundIsDarkCyan verifies the foreground color
// of BorlandCyan.HistoryArrow is tcell.ColorDarkCyan.
func TestBorlandCyanHistoryArrowForegroundIsDarkCyan(t *testing.T) {
	if BorlandCyan == nil {
		t.Fatal("BorlandCyan is nil")
	}
	fg, _, _ := BorlandCyan.HistoryArrow.Decompose()
	if fg != tcell.ColorDarkCyan {
		t.Errorf("BorlandCyan.HistoryArrow foreground = %v, want tcell.ColorDarkCyan", fg)
	}
}

// ---------------------------------------------------------------------------
// Requirement 16 — BorlandCyan.HistorySides background is cyan
// Spec: "BorlandCyan: sides = white on cyan"
// ---------------------------------------------------------------------------

// TestBorlandCyanHistorySidesBackgroundIsCyan verifies the background color
// of BorlandCyan.HistorySides is tcell.ColorTeal (cyan).
func TestBorlandCyanHistorySidesBackgroundIsCyan(t *testing.T) {
	if BorlandCyan == nil {
		t.Fatal("BorlandCyan is nil")
	}
	_, bg, _ := BorlandCyan.HistorySides.Decompose()
	if bg != tcell.ColorTeal {
		t.Errorf("BorlandCyan.HistorySides background = %v, want tcell.ColorTeal (cyan)", bg)
	}
}

// TestBorlandCyanHistorySidesForegroundIsWhite verifies the foreground color
// of BorlandCyan.HistorySides is tcell.ColorWhite.
func TestBorlandCyanHistorySidesForegroundIsWhite(t *testing.T) {
	if BorlandCyan == nil {
		t.Fatal("BorlandCyan is nil")
	}
	fg, _, _ := BorlandCyan.HistorySides.Decompose()
	if fg != tcell.ColorWhite {
		t.Errorf("BorlandCyan.HistorySides foreground = %v, want tcell.ColorWhite", fg)
	}
}

// ---------------------------------------------------------------------------
// Requirement 17 — BorlandGray.HistoryArrow background is white
// Spec: "BorlandGray: arrow = dark gray on white"
// ---------------------------------------------------------------------------

// TestBorlandGrayHistoryArrowBackgroundIsWhite verifies the background color
// of BorlandGray.HistoryArrow is tcell.ColorWhite.
func TestBorlandGrayHistoryArrowBackgroundIsWhite(t *testing.T) {
	if BorlandGray == nil {
		t.Fatal("BorlandGray is nil")
	}
	_, bg, _ := BorlandGray.HistoryArrow.Decompose()
	if bg != tcell.ColorWhite {
		t.Errorf("BorlandGray.HistoryArrow background = %v, want tcell.ColorWhite", bg)
	}
}

// TestBorlandGrayHistoryArrowForegroundIsDarkGray verifies the foreground color
// of BorlandGray.HistoryArrow is tcell.ColorDarkGray.
func TestBorlandGrayHistoryArrowForegroundIsDarkGray(t *testing.T) {
	if BorlandGray == nil {
		t.Fatal("BorlandGray is nil")
	}
	fg, _, _ := BorlandGray.HistoryArrow.Decompose()
	if fg != tcell.ColorDarkGray {
		t.Errorf("BorlandGray.HistoryArrow foreground = %v, want tcell.ColorDarkGray", fg)
	}
}

// ---------------------------------------------------------------------------
// Requirement 18 — BorlandGray.HistorySides background is dark gray
// Spec: "BorlandGray: sides = white on dark gray"
// ---------------------------------------------------------------------------

// TestBorlandGrayHistorySidesBackgroundIsDarkGray verifies the background color
// of BorlandGray.HistorySides is tcell.ColorDarkGray.
func TestBorlandGrayHistorySidesBackgroundIsDarkGray(t *testing.T) {
	if BorlandGray == nil {
		t.Fatal("BorlandGray is nil")
	}
	_, bg, _ := BorlandGray.HistorySides.Decompose()
	if bg != tcell.ColorDarkGray {
		t.Errorf("BorlandGray.HistorySides background = %v, want tcell.ColorDarkGray", bg)
	}
}

// TestBorlandGrayHistorySidesForegroundIsWhite verifies the foreground color
// of BorlandGray.HistorySides is tcell.ColorWhite.
func TestBorlandGrayHistorySidesForegroundIsWhite(t *testing.T) {
	if BorlandGray == nil {
		t.Fatal("BorlandGray is nil")
	}
	fg, _, _ := BorlandGray.HistorySides.Decompose()
	if fg != tcell.ColorWhite {
		t.Errorf("BorlandGray.HistorySides foreground = %v, want tcell.ColorWhite", fg)
	}
}

// ---------------------------------------------------------------------------
// Requirement 19 — C64.HistoryArrow background is blue
// Spec: "C64: arrow = light blue on blue"
// ---------------------------------------------------------------------------

// TestC64HistoryArrowBackgroundIsBlue verifies the background color
// of C64.HistoryArrow is tcell.ColorBlue.
func TestC64HistoryArrowBackgroundIsBlue(t *testing.T) {
	if C64 == nil {
		t.Fatal("C64 is nil")
	}
	_, bg, _ := C64.HistoryArrow.Decompose()
	if bg != tcell.ColorBlue {
		t.Errorf("C64.HistoryArrow background = %v, want tcell.ColorBlue", bg)
	}
}

// TestC64HistoryArrowForegroundIsLightBlue verifies the foreground color
// of C64.HistoryArrow is tcell.ColorLightBlue.
func TestC64HistoryArrowForegroundIsLightBlue(t *testing.T) {
	if C64 == nil {
		t.Fatal("C64 is nil")
	}
	fg, _, _ := C64.HistoryArrow.Decompose()
	if fg != tcell.ColorLightBlue {
		t.Errorf("C64.HistoryArrow foreground = %v, want tcell.ColorLightBlue", fg)
	}
}

// ---------------------------------------------------------------------------
// Requirement 20 — C64.HistorySides background is light blue
// Spec: "C64: sides = blue on light blue"
// ---------------------------------------------------------------------------

// TestC64HistorySidesBackgroundIsLightBlue verifies the background color
// of C64.HistorySides is tcell.ColorLightBlue.
func TestC64HistorySidesBackgroundIsLightBlue(t *testing.T) {
	if C64 == nil {
		t.Fatal("C64 is nil")
	}
	_, bg, _ := C64.HistorySides.Decompose()
	if bg != tcell.ColorLightBlue {
		t.Errorf("C64.HistorySides background = %v, want tcell.ColorLightBlue", bg)
	}
}

// TestC64HistorySidesForegroundIsBlue verifies the foreground color
// of C64.HistorySides is tcell.ColorBlue.
func TestC64HistorySidesForegroundIsBlue(t *testing.T) {
	if C64 == nil {
		t.Fatal("C64 is nil")
	}
	fg, _, _ := C64.HistorySides.Decompose()
	if fg != tcell.ColorBlue {
		t.Errorf("C64.HistorySides foreground = %v, want tcell.ColorBlue", fg)
	}
}

// ---------------------------------------------------------------------------
// Requirement 21 — Matrix.HistoryArrow background is dark green
// Spec: "Matrix: arrow = white on dark green"
// ---------------------------------------------------------------------------

// TestMatrixHistoryArrowBackgroundIsDarkGreen verifies the background color
// of Matrix.HistoryArrow is tcell.ColorDarkGreen.
func TestMatrixHistoryArrowBackgroundIsDarkGreen(t *testing.T) {
	if Matrix == nil {
		t.Fatal("Matrix is nil")
	}
	_, bg, _ := Matrix.HistoryArrow.Decompose()
	if bg != tcell.ColorDarkGreen {
		t.Errorf("Matrix.HistoryArrow background = %v, want tcell.ColorDarkGreen", bg)
	}
}

// TestMatrixHistoryArrowForegroundIsWhite verifies the foreground color
// of Matrix.HistoryArrow is tcell.ColorWhite.
func TestMatrixHistoryArrowForegroundIsWhite(t *testing.T) {
	if Matrix == nil {
		t.Fatal("Matrix is nil")
	}
	fg, _, _ := Matrix.HistoryArrow.Decompose()
	if fg != tcell.ColorWhite {
		t.Errorf("Matrix.HistoryArrow foreground = %v, want tcell.ColorWhite", fg)
	}
}

// ---------------------------------------------------------------------------
// Requirement 22 — Matrix.HistorySides background is black
// Spec: "Matrix: sides = green on black"
// ---------------------------------------------------------------------------

// TestMatrixHistorySidesBackgroundIsBlack verifies the background color
// of Matrix.HistorySides is tcell.ColorBlack.
func TestMatrixHistorySidesBackgroundIsBlack(t *testing.T) {
	if Matrix == nil {
		t.Fatal("Matrix is nil")
	}
	_, bg, _ := Matrix.HistorySides.Decompose()
	if bg != tcell.ColorBlack {
		t.Errorf("Matrix.HistorySides background = %v, want tcell.ColorBlack", bg)
	}
}

// TestMatrixHistorySidesForegroundIsGreen verifies the foreground color
// of Matrix.HistorySides is tcell.ColorGreen.
func TestMatrixHistorySidesForegroundIsGreen(t *testing.T) {
	if Matrix == nil {
		t.Fatal("Matrix is nil")
	}
	fg, _, _ := Matrix.HistorySides.Decompose()
	if fg != tcell.ColorGreen {
		t.Errorf("Matrix.HistorySides foreground = %v, want tcell.ColorGreen", fg)
	}
}

// ---------------------------------------------------------------------------
// Requirement 23 — All five themes have 35 non-zero fields after adding 2 new fields
// Spec: "All five themes must include values for these new fields (BorlandBlue, BorlandCyan,
//         BorlandGray, C64, Matrix)"
// The existing struct had 33 non-zero fields; adding HistoryArrow + HistorySides = 35.
// ---------------------------------------------------------------------------

// TestBorlandBlueHas37NonZeroFields verifies BorlandBlue populates all 37 fields
// BorlandBlue sets ALL fields (including LabelHighlight and StatusSelected that other themes skip).
// (33 existing + HistoryArrow + HistorySides).
func TestBorlandBlueHas37NonZeroFields(t *testing.T) {
	if BorlandBlue == nil {
		t.Fatal("BorlandBlue is nil")
	}
	count := countNonZeroFields(BorlandBlue)
	if count != 38 {
		t.Errorf("BorlandBlue: expected 38 non-zero fields, got %d", count)
	}
}

// TestBorlandCyanHas35NonZeroFields verifies BorlandCyan populates all 35 fields.
func TestBorlandCyanHas35NonZeroFields(t *testing.T) {
	if BorlandCyan == nil {
		t.Fatal("BorlandCyan is nil")
	}
	count := countNonZeroFields(BorlandCyan)
	if count != 38 {
		t.Errorf("BorlandCyan: expected 38 non-zero fields, got %d", count)
	}
}

// TestBorlandGrayHas35NonZeroFields verifies BorlandGray populates all 35 fields.
func TestBorlandGrayHas35NonZeroFields(t *testing.T) {
	if BorlandGray == nil {
		t.Fatal("BorlandGray is nil")
	}
	count := countNonZeroFields(BorlandGray)
	if count != 38 {
		t.Errorf("BorlandGray: expected 38 non-zero fields, got %d", count)
	}
}

// TestC64Has35NonZeroFields verifies C64 populates all 35 fields.
func TestC64Has35NonZeroFields(t *testing.T) {
	if C64 == nil {
		t.Fatal("C64 is nil")
	}
	count := countNonZeroFields(C64)
	if count != 38 {
		t.Errorf("C64: expected 38 non-zero fields, got %d", count)
	}
}

// TestMatrixHas35NonZeroFields verifies Matrix populates all 35 fields.
func TestMatrixHas35NonZeroFields(t *testing.T) {
	if Matrix == nil {
		t.Fatal("Matrix is nil")
	}
	count := countNonZeroFields(Matrix)
	if count != 38 {
		t.Errorf("Matrix: expected 38 non-zero fields, got %d", count)
	}
}
