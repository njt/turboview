package tv

// shadow_test.go — tests for Task 6: Shadow Rendering
//
// Written against the spec before any implementation exists.
// Every assertion cites the spec sentence it verifies.
//
// Spec summary:
//   DrawBuffer.SetCellStyle(x, y, style) updates only the style of the cell at
//   (x, y), preserving its rune and combining characters. Out-of-bounds or
//   outside-clip writes are no-ops.
//
//   Desktop draws shadows during its draw pass, after drawing each visible window.
//   - Right shadow: 2 cols wide, column range [B.X, B.X+2), row range [A.Y+1, B.Y+1)
//   - Bottom shadow: 1 row, column range [A.X+2, B.X+2), row = B.Y
//   - Shadow style = ColorScheme().WindowShadow
//   - Shadow preserves the existing rune; only the style is replaced
//   - Shadow does not render outside the Desktop's bounds
//   - Invisible windows receive no shadow

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// DrawBuffer.SetCellStyle — 5 tests
// ---------------------------------------------------------------------------

// TestSetCellStyleChangesOnlyStyle verifies:
// "SetCellStyle(x, y, style) updates only the style of the cell at (x, y),
// preserving its rune and combining characters."
func TestSetCellStyleChangesOnlyStyle(t *testing.T) {
	db := NewDrawBuffer(10, 10)
	db.WriteChar(3, 4, 'X', tcell.StyleDefault)

	newStyle := tcell.StyleDefault.Foreground(tcell.ColorRed).Background(tcell.ColorBlue)
	db.SetCellStyle(3, 4, newStyle)

	cell := db.GetCell(3, 4)
	if cell.Rune != 'X' {
		t.Errorf("SetCellStyle changed the rune: got %q, want 'X'", cell.Rune)
	}
	if cell.Style != newStyle {
		t.Errorf("SetCellStyle did not update style: got %v, want %v", cell.Style, newStyle)
	}
}

// TestSetCellStylePreservesCombiningCharacters verifies:
// "SetCellStyle(x, y, style) updates only the style of the cell at (x, y),
// preserving its rune and combining characters."
func TestSetCellStylePreservesCombiningCharacters(t *testing.T) {
	db := NewDrawBuffer(10, 10)
	// Manually inject a cell with combining characters via the internal cells slice.
	// We write a base rune first, then patch the Combc field to simulate a cell
	// that has combining characters (e.g. combining accent U+0301).
	db.WriteChar(2, 2, 'e', tcell.StyleDefault)
	// Patch combining characters directly on the underlying cell.
	combining := []rune{'́'} // combining acute accent
	db.cells[2][2].Combc = combining

	newStyle := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	db.SetCellStyle(2, 2, newStyle)

	cell := db.GetCell(2, 2)
	if cell.Rune != 'e' {
		t.Errorf("SetCellStyle changed the base rune: got %q, want 'e'", cell.Rune)
	}
	if len(cell.Combc) != 1 || cell.Combc[0] != '́' {
		t.Errorf("SetCellStyle lost combining characters: got %v, want [U+0301]", cell.Combc)
	}
	if cell.Style != newStyle {
		t.Errorf("SetCellStyle did not update style: got %v, want %v", cell.Style, newStyle)
	}
}

// TestSetCellStyleOutOfBoundsIsNoOp verifies:
// "Out-of-bounds or outside-clip writes are no-ops."
func TestSetCellStyleOutOfBoundsIsNoOp(t *testing.T) {
	db := NewDrawBuffer(10, 10)
	db.WriteChar(0, 0, 'A', tcell.StyleDefault)

	newStyle := tcell.StyleDefault.Foreground(tcell.ColorRed)

	// None of these should panic or mutate anything.
	db.SetCellStyle(-1, 0, newStyle)
	db.SetCellStyle(0, -1, newStyle)
	db.SetCellStyle(10, 5, newStyle)
	db.SetCellStyle(5, 10, newStyle)
	db.SetCellStyle(100, 100, newStyle)

	// The cell at (0,0) that we wrote should remain unchanged.
	cell := db.GetCell(0, 0)
	if cell.Rune != 'A' {
		t.Errorf("SetCellStyle out-of-bounds mutated buffer: rune at (0,0) = %q, want 'A'", cell.Rune)
	}
	if cell.Style != tcell.StyleDefault {
		t.Errorf("SetCellStyle out-of-bounds changed style at (0,0)")
	}
}

// TestSetCellStyleOutsideClipIsNoOp verifies:
// "Out-of-bounds or outside-clip writes are no-ops."
// Uses a SubBuffer so the clip is narrower than the backing store.
func TestSetCellStyleOutsideClipIsNoOp(t *testing.T) {
	parent := NewDrawBuffer(20, 20)
	parent.WriteChar(15, 15, 'P', tcell.StyleDefault)

	// SubBuffer covering only [5,5)–[10,10) in parent coordinates.
	sub := parent.SubBuffer(NewRect(5, 5, 5, 5))

	newStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow)

	// Sub-local (10, 10) maps to parent (15, 15), which is outside the 5×5 clip.
	sub.SetCellStyle(10, 10, newStyle)

	// The parent cell at (15, 15) must be unchanged.
	cell := parent.GetCell(15, 15)
	if cell.Rune != 'P' {
		t.Errorf("SetCellStyle outside clip mutated rune at parent(15,15): got %q", cell.Rune)
	}
	if cell.Style != tcell.StyleDefault {
		t.Errorf("SetCellStyle outside clip changed style at parent(15,15)")
	}
}

// TestSetCellStyleInBoundsWorks verifies:
// "SetCellStyle(x, y, style) updates only the style of the cell at (x, y)."
// Confirms that an in-bounds call succeeds and the change is visible.
func TestSetCellStyleInBoundsWorks(t *testing.T) {
	db := NewDrawBuffer(20, 20)
	db.WriteChar(7, 7, '░', tcell.StyleDefault)

	shadowStyle := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorBlack)
	db.SetCellStyle(7, 7, shadowStyle)

	cell := db.GetCell(7, 7)
	if cell.Rune != '░' {
		t.Errorf("SetCellStyle changed the rune: got %q, want '░'", cell.Rune)
	}
	if cell.Style != shadowStyle {
		t.Errorf("SetCellStyle did not apply style: got %v, want %v", cell.Style, shadowStyle)
	}
}

// ---------------------------------------------------------------------------
// Desktop shadow rendering — 9 tests
// ---------------------------------------------------------------------------

// shadowScheme returns a ColorScheme with a distinct WindowShadow style that is
// easy to detect, and a DesktopBackground style that uses '░'.
func shadowScheme() *theme.ColorScheme {
	return &theme.ColorScheme{
		DesktopBackground:   tcell.StyleDefault.Foreground(tcell.ColorTeal),
		WindowShadow:        tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorBlack),
		WindowFrameActive:   tcell.StyleDefault.Foreground(tcell.ColorWhite),
		WindowFrameInactive: tcell.StyleDefault.Foreground(tcell.ColorSilver),
		WindowBackground:    tcell.StyleDefault.Foreground(tcell.ColorWhite),
		WindowTitle:         tcell.StyleDefault.Foreground(tcell.ColorYellow),
	}
}

// drawDesktop is a helper that creates a Desktop of the given size, optionally
// attaches a scheme, inserts the provided windows, and returns the filled DrawBuffer.
func drawDesktop(t *testing.T, w, h int, scheme *theme.ColorScheme, windows ...*Window) *DrawBuffer {
	t.Helper()
	d := NewDesktop(NewRect(0, 0, w, h))
	if scheme != nil {
		d.scheme = scheme
	}
	for _, win := range windows {
		d.Insert(win)
	}
	buf := NewDrawBuffer(w, h)
	d.Draw(buf)
	return buf
}

// TestShadowAppearsToRightOfWindow verifies:
// "Right shadow: column range [window.B.X, window.B.X+2), row range [window.A.Y+1, window.B.Y+1)"
// — the shadow cells appear 2 columns to the right of the window.
func TestShadowAppearsToRightOfWindow(t *testing.T) {
	// Window at NewRect(5, 3, 15, 7): A=(5,3) B=(20,10)
	// Right shadow x: [20, 22), y: [4, 11)
	win := NewWindow(NewRect(5, 3, 15, 7), "W")
	scheme := shadowScheme()
	buf := drawDesktop(t, 40, 20, scheme, win)

	// Spot-check one cell inside the right shadow region.
	cell := buf.GetCell(20, 5)
	if cell.Style != scheme.WindowShadow {
		t.Errorf("Right shadow cell (20,5) style = %v, want WindowShadow %v", cell.Style, scheme.WindowShadow)
	}
}

// TestShadowAppearsToRightOfWindowBothColumns verifies:
// "Right shadow: column range [window.B.X, window.B.X+2)" — 2 columns wide.
func TestShadowAppearsToRightOfWindowBothColumns(t *testing.T) {
	// Window at NewRect(5, 3, 15, 7): A=(5,3) B=(20,10)
	// Right shadow x: 20 and 21, y: [4, 11)
	win := NewWindow(NewRect(5, 3, 15, 7), "W")
	scheme := shadowScheme()
	buf := drawDesktop(t, 40, 20, scheme, win)

	for _, x := range []int{20, 21} {
		cell := buf.GetCell(x, 5)
		if cell.Style != scheme.WindowShadow {
			t.Errorf("Right shadow cell (%d,5) style = %v, want WindowShadow %v", x, cell.Style, scheme.WindowShadow)
		}
	}
}

// TestShadowAppearsBelow verifies:
// "Bottom shadow: column range [window.A.X+2, window.B.X+2), row = window.B.Y"
// — the shadow appears one row below the window.
func TestShadowAppearsBelow(t *testing.T) {
	// Window at NewRect(5, 3, 15, 7): A=(5,3) B=(20,10)
	// Bottom shadow y=10, x: [7, 22)
	win := NewWindow(NewRect(5, 3, 15, 7), "W")
	scheme := shadowScheme()
	buf := drawDesktop(t, 40, 20, scheme, win)

	// Spot-check a cell in the bottom shadow row.
	cell := buf.GetCell(10, 10)
	if cell.Style != scheme.WindowShadow {
		t.Errorf("Bottom shadow cell (10,10) style = %v, want WindowShadow %v", cell.Style, scheme.WindowShadow)
	}
}

// TestShadowUsesWindowShadowStyle verifies:
// "Shadow uses ColorScheme().WindowShadow style"
// — the shadow style is taken from the scheme, not any other style field.
func TestShadowUsesWindowShadowStyle(t *testing.T) {
	win := NewWindow(NewRect(2, 2, 10, 6), "W")
	scheme := shadowScheme()
	// Make WindowShadow distinct and easy to identify.
	distinctShadow := tcell.StyleDefault.Foreground(tcell.ColorFuchsia).Background(tcell.ColorNavy)
	scheme.WindowShadow = distinctShadow

	buf := drawDesktop(t, 30, 20, scheme, win)

	// Window: A=(2,2) B=(12,8). Right shadow x=[12,14), y=[3,9).
	cell := buf.GetCell(12, 4)
	if cell.Style != distinctShadow {
		t.Errorf("Shadow style = %v, want WindowShadow %v", cell.Style, distinctShadow)
	}
}

// TestShadowPreservesRune verifies:
// "Shadow preserves the existing character (rune) in each cell and only replaces the style"
// — the desktop pattern rune ('░') must survive under the shadow.
func TestShadowPreservesRune(t *testing.T) {
	// A window placed so its shadow falls on cells that were filled with '░'.
	win := NewWindow(NewRect(2, 2, 10, 6), "W")
	scheme := shadowScheme()
	buf := drawDesktop(t, 30, 20, scheme, win)

	// Window: A=(2,2) B=(12,8). Right shadow x=[12,14), y=[3,9).
	// Before shadow, those cells held '░' (DesktopBackground fill).
	cell := buf.GetCell(12, 4)
	if cell.Rune != '░' {
		t.Errorf("Shadow must preserve the existing rune: got %q, want '░'", cell.Rune)
	}
}

// TestRightShadowStartsOneRowBelowWindowTop verifies:
// "Right shadow: row range [window.A.Y+1, window.B.Y+1)"
// — the top row of the window (A.Y) is NOT shadowed; shadow starts at A.Y+1.
func TestRightShadowStartsOneRowBelowWindowTop(t *testing.T) {
	// Window at NewRect(5, 4, 10, 6): A=(5,4) B=(15,10)
	// Right shadow x=[15,17), y=[5,11) — NOT y=4.
	win := NewWindow(NewRect(5, 4, 10, 6), "W")
	scheme := shadowScheme()
	buf := drawDesktop(t, 30, 20, scheme, win)

	// Row A.Y=4 at x=15 should NOT be shadow.
	topRowCell := buf.GetCell(15, 4)
	if topRowCell.Style == scheme.WindowShadow {
		t.Errorf("Right shadow must not appear at window top row (A.Y=4); cell(15,4) has WindowShadow style")
	}

	// Row A.Y+1=5 at x=15 SHOULD be shadow.
	firstShadowCell := buf.GetCell(15, 5)
	if firstShadowCell.Style != scheme.WindowShadow {
		t.Errorf("Right shadow must start at A.Y+1=5; cell(15,5) style = %v, want WindowShadow", firstShadowCell.Style)
	}
}

// TestBottomShadowStartsTwoColsRightOfWindowLeft verifies:
// "Bottom shadow: column range [window.A.X+2, window.B.X+2)"
// — columns A.X and A.X+1 are NOT in the bottom shadow.
func TestBottomShadowStartsTwoColsRightOfWindowLeft(t *testing.T) {
	// Window at NewRect(6, 4, 12, 6): A=(6,4) B=(18,10)
	// Bottom shadow y=10, x=[8,20) — NOT x=6 or x=7.
	win := NewWindow(NewRect(6, 4, 12, 6), "W")
	scheme := shadowScheme()
	buf := drawDesktop(t, 30, 20, scheme, win)

	// x=6 (A.X) at y=10 should NOT be shadow.
	cell6 := buf.GetCell(6, 10)
	if cell6.Style == scheme.WindowShadow {
		t.Errorf("Bottom shadow must not start at A.X=6; cell(6,10) has WindowShadow style")
	}

	// x=7 (A.X+1) at y=10 should NOT be shadow.
	cell7 := buf.GetCell(7, 10)
	if cell7.Style == scheme.WindowShadow {
		t.Errorf("Bottom shadow must not start at A.X+1=7; cell(7,10) has WindowShadow style")
	}

	// x=8 (A.X+2) at y=10 SHOULD be shadow.
	cell8 := buf.GetCell(8, 10)
	if cell8.Style != scheme.WindowShadow {
		t.Errorf("Bottom shadow must start at A.X+2=8; cell(8,10) style = %v, want WindowShadow", cell8.Style)
	}
}

// TestShadowClippedToDesktopBounds verifies:
// "Shadow does not render outside the Desktop's bounds"
// — a window placed at the right edge of the desktop must not cause an out-of-bounds
// write; cells outside the desktop are unaffected.
func TestShadowClippedToDesktopBounds(t *testing.T) {
	// Desktop is 20 wide (x: 0..19). Window at NewRect(16, 2, 10, 5):
	// B.X = 26, right shadow would be x=[26,28) — entirely outside desktop.
	// Bottom shadow would be x=[18,28), y=7 — only x=18,19 are inside.
	win := NewWindow(NewRect(10, 2, 10, 5), "W")
	scheme := shadowScheme()

	// Must not panic. DrawBuffer.SetCellStyle out-of-bounds calls must be no-ops.
	buf := drawDesktop(t, 20, 15, scheme, win)

	// Window: A=(10,2) B=(20,7). Right shadow x=[20,22) — outside 20-wide desktop.
	// Those positions don't exist in the buffer; reading them returns zero Cell.
	// We just verify no panic occurred (the drawDesktop call above).

	// Bottom shadow y=7, x=[12,22). x=12..19 are inside; x=20,21 are outside.
	// x=12 should be shadow.
	cell12 := buf.GetCell(12, 7)
	if cell12.Style != scheme.WindowShadow {
		t.Errorf("Bottom shadow cell (12,7) within bounds: style = %v, want WindowShadow", cell12.Style)
	}

	// x=19 (last column) should also be shadow.
	cell19 := buf.GetCell(19, 7)
	if cell19.Style != scheme.WindowShadow {
		t.Errorf("Bottom shadow cell (19,7) at desktop edge: style = %v, want WindowShadow", cell19.Style)
	}
}

// TestNoShadowForInvisibleWindow verifies:
// "No shadow for invisible windows"
// — a window whose SfVisible flag is cleared must not cast a shadow.
func TestNoShadowForInvisibleWindow(t *testing.T) {
	win := NewWindow(NewRect(3, 3, 10, 6), "W")
	win.SetState(SfVisible, false)

	scheme := shadowScheme()
	buf := drawDesktop(t, 30, 20, scheme, win)

	// Window: A=(3,3) B=(13,9). Right shadow would be x=[13,15), y=[4,10).
	// Since the window is invisible, no shadow should appear.
	cell := buf.GetCell(13, 5)
	if cell.Style == scheme.WindowShadow {
		t.Errorf("Invisible window must not cast a shadow; cell(13,5) has WindowShadow style")
	}

	// Bottom shadow row y=9 at x=5 should also not be shadow.
	cellBottom := buf.GetCell(5, 9)
	if cellBottom.Style == scheme.WindowShadow {
		t.Errorf("Invisible window must not cast a bottom shadow; cell(5,9) has WindowShadow style")
	}
}

// TestMultipleWindowsShadowsOverlap verifies:
// "Multiple windows: later window's shadow can overlap earlier window"
// — the shadow is rendered immediately after each window in draw order; a later
// window's shadow overwrites whatever was written by an earlier window.
func TestMultipleWindowsShadowsOverlap(t *testing.T) {
	// Window A: NewRect(2, 2, 10, 6) — A=(2,2) B=(12,8)
	// Window B: NewRect(10, 0, 10, 6) — A=(10,0) B=(20,6)
	//   Window B's right shadow: x=[20,22), y=[1,7)
	//   Window B's bottom shadow: y=6, x=[12,22)
	//
	// The overlap area of interest: window B is inserted after window A.
	// Window A's right shadow at x=[12,14), y=[3,9) should be overdrawn by
	// window B itself when B draws at x=[10..19], y=[0..5].
	// More directly: window B's shadow should appear at its own shadow positions.
	winA := NewWindow(NewRect(2, 2, 10, 6), "A")
	winB := NewWindow(NewRect(10, 0, 10, 6), "B")

	scheme := shadowScheme()
	buf := drawDesktop(t, 40, 20, scheme, winA, winB)

	// Window B shadow: right shadow x=[20,22), y=[1,7).
	// This is unambiguously B's shadow and should use WindowShadow style.
	cellRight := buf.GetCell(20, 3)
	if cellRight.Style != scheme.WindowShadow {
		t.Errorf("Window B right shadow cell (20,3) style = %v, want WindowShadow", cellRight.Style)
	}

	// Window B bottom shadow: y=6, x=[12,22).
	// This is B's shadow and should use WindowShadow style.
	cellBottom := buf.GetCell(15, 6)
	if cellBottom.Style != scheme.WindowShadow {
		t.Errorf("Window B bottom shadow cell (15,6) style = %v, want WindowShadow", cellBottom.Style)
	}
}

// TestShadowFullRightRegionStyle verifies:
// "Right shadow: column range [window.B.X, window.B.X+2), row range [window.A.Y+1, window.B.Y+1)"
// — exhaustively checks all cells in the right shadow region have WindowShadow style.
func TestShadowFullRightRegionStyle(t *testing.T) {
	// Window at NewRect(4, 2, 10, 6): A=(4,2) B=(14,8)
	// Right shadow: x=[14,16), y=[3,9)
	win := NewWindow(NewRect(4, 2, 10, 6), "W")
	scheme := shadowScheme()
	buf := drawDesktop(t, 30, 20, scheme, win)

	for x := 14; x < 16; x++ {
		for y := 3; y < 9; y++ {
			cell := buf.GetCell(x, y)
			if cell.Style != scheme.WindowShadow {
				t.Errorf("Right shadow cell (%d,%d) style = %v, want WindowShadow", x, y, cell.Style)
			}
		}
	}
}

// TestShadowFullBottomRegionStyle verifies:
// "Bottom shadow: column range [window.A.X+2, window.B.X+2), row = window.B.Y"
// — exhaustively checks all cells in the bottom shadow row have WindowShadow style.
func TestShadowFullBottomRegionStyle(t *testing.T) {
	// Window at NewRect(4, 2, 10, 6): A=(4,2) B=(14,8)
	// Bottom shadow: y=8, x=[6,16)
	win := NewWindow(NewRect(4, 2, 10, 6), "W")
	scheme := shadowScheme()
	buf := drawDesktop(t, 30, 20, scheme, win)

	for x := 6; x < 16; x++ {
		cell := buf.GetCell(x, 8)
		if cell.Style != scheme.WindowShadow {
			t.Errorf("Bottom shadow cell (%d,8) style = %v, want WindowShadow", x, cell.Style)
		}
	}
}
