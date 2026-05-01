package tv

// memo_tab_test.go — Tests for Task 2: Tab Rendering in Memo Draw().
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test cites the exact spec sentence it verifies.
//
// Tab stop rule: a tab at visual column C expands to (8 - C%8) spaces,
// advancing the visual column to the next multiple of 8.
//
// Test organisation:
//   Section 1  — Tab at column 0 expands to 8 spaces
//   Section 2  — Tab at column 1 expands to 7 spaces
//   Section 3  — Tab at column 7 expands to 1 space
//   Section 4  — Tab at column 8 expands to 8 spaces (second tab stop)
//   Section 5  — Multiple tabs expand correctly
//   Section 6  — Tab clipped at widget right edge
//   Section 7  — Tab expansion uses MemoNormal style (no selection)
//   Section 8  — Tab in selection uses MemoSelected for all expanded spaces
//   Section 9  — Mixed content: text + tab + text renders correctly

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// tabKeyEv creates a plain keyboard event with no modifiers.
func tabKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key}}
}

// tabShiftKeyEv creates a keyboard event with Shift held.
func tabShiftKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key, Modifiers: tcell.ModShift}}
}

// tabCtrlKeyEv creates a keyboard event with Ctrl held.
func tabCtrlKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key, Modifiers: tcell.ModCtrl}}
}

// newTabMemo creates a Memo with BorlandBlue scheme, sized w×h.
func newTabMemo(w, h int) *Memo {
	m := NewMemo(NewRect(0, 0, w, h))
	m.scheme = theme.BorlandBlue
	return m
}

// ---------------------------------------------------------------------------
// Section 1 — Tab at column 0 expands to 8 spaces
// ---------------------------------------------------------------------------

// TestDrawTabAtColumn0ExpandsTo8Spaces verifies that a tab at visual column 0
// produces 8 spaces in the rendered output.
// Spec: "A tab at visual column 0 expands to 8 spaces"
// Spec: "Tab stops are at visual columns 0, 8, 16, 24, etc."
func TestDrawTabAtColumn0ExpandsTo8Spaces(t *testing.T) {
	// "\t" only: tab is at column 0, must expand to 8 space cells.
	m := newTabMemo(20, 2)
	m.SetText("\t")

	buf := NewDrawBuffer(20, 2)
	m.Draw(buf)

	for col := 0; col < 8; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Rune != ' ' {
			t.Errorf("Tab@col0: cell(%d,0).Rune = %q, want ' ' (tab expands to 8 spaces)", col, cell.Rune)
		}
	}
}

// TestDrawTabAtColumn0NextContentAt8 verifies that content immediately after
// a tab-at-column-0 appears at visual column 8.
// Spec: "A tab at visual column 0 expands to 8 spaces"
func TestDrawTabAtColumn0NextContentAt8(t *testing.T) {
	// "\tX": tab expands 0→8, then 'X' is at visual column 8.
	m := newTabMemo(20, 2)
	m.SetText("\tX")

	buf := NewDrawBuffer(20, 2)
	m.Draw(buf)

	cell := buf.GetCell(8, 0)
	if cell.Rune != 'X' {
		t.Errorf("After tab@col0: cell(8,0).Rune = %q, want 'X' (tab expands to 8 spaces)", cell.Rune)
	}
}

// ---------------------------------------------------------------------------
// Section 2 — Tab at column 1 expands to 7 spaces
// ---------------------------------------------------------------------------

// TestDrawTabAtColumn1ExpandsTo7Spaces verifies that a tab at visual column 1
// produces exactly 7 rendered spaces (advancing to the next tab stop at 8).
// Spec: "at column 1, expands to 7 spaces"
func TestDrawTabAtColumn1ExpandsTo7Spaces(t *testing.T) {
	// "A\t": 'A' at col 0, tab at visual col 1 → expands to 7 spaces (cols 1–7).
	m := newTabMemo(20, 2)
	m.SetText("A\t")

	buf := NewDrawBuffer(20, 2)
	m.Draw(buf)

	// 'A' is at col 0.
	if buf.GetCell(0, 0).Rune != 'A' {
		t.Errorf("Tab@col1: cell(0,0).Rune = %q, want 'A'", buf.GetCell(0, 0).Rune)
	}
	// Cols 1–7 must be spaces (7 spaces for tab from col 1 to stop at col 8).
	for col := 1; col < 8; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Rune != ' ' {
			t.Errorf("Tab@col1: cell(%d,0).Rune = %q, want ' ' (tab expands to 7 spaces)", col, cell.Rune)
		}
	}
}

// TestDrawTabAtColumn1NextContentAt8 verifies that content after a tab at
// column 1 appears at visual column 8.
// Spec: "at column 1, expands to 7 spaces"
func TestDrawTabAtColumn1NextContentAt8(t *testing.T) {
	// "A\tZ": 'A'@0, tab@1 expands 7 spaces, 'Z'@8.
	m := newTabMemo(20, 2)
	m.SetText("A\tZ")

	buf := NewDrawBuffer(20, 2)
	m.Draw(buf)

	cell := buf.GetCell(8, 0)
	if cell.Rune != 'Z' {
		t.Errorf("After tab@col1: cell(8,0).Rune = %q, want 'Z'", cell.Rune)
	}
}

// ---------------------------------------------------------------------------
// Section 3 — Tab at column 7 expands to 1 space
// ---------------------------------------------------------------------------

// TestDrawTabAtColumn7ExpandsTo1Space verifies that a tab at visual column 7
// expands to exactly 1 space (advancing to tab stop 8).
// Spec: "at column 7, expands to 1 space"
func TestDrawTabAtColumn7ExpandsTo1Space(t *testing.T) {
	// "ABCDEFG\t": 7 chars then tab at visual col 7 → 1 space at col 7, next at col 8.
	m := newTabMemo(20, 2)
	m.SetText("ABCDEFG\t")

	buf := NewDrawBuffer(20, 2)
	m.Draw(buf)

	// Col 7 must be a space (the single expanded space of the tab).
	cell := buf.GetCell(7, 0)
	if cell.Rune != ' ' {
		t.Errorf("Tab@col7: cell(7,0).Rune = %q, want ' ' (tab at col 7 expands to 1 space)", cell.Rune)
	}
}

// TestDrawTabAtColumn7NextContentAt8 verifies content after tab-at-col-7
// appears at visual column 8.
// Spec: "at column 7, expands to 1 space"
func TestDrawTabAtColumn7NextContentAt8(t *testing.T) {
	// "ABCDEFG\tZ": 7 chars, tab@7 (1 space), 'Z'@8.
	m := newTabMemo(20, 2)
	m.SetText("ABCDEFG\tZ")

	buf := NewDrawBuffer(20, 2)
	m.Draw(buf)

	cell := buf.GetCell(8, 0)
	if cell.Rune != 'Z' {
		t.Errorf("After tab@col7: cell(8,0).Rune = %q, want 'Z'", cell.Rune)
	}
}

// ---------------------------------------------------------------------------
// Section 4 — Tab at column 8 expands to 8 spaces (second tab stop)
// ---------------------------------------------------------------------------

// TestDrawTabAtColumn8ExpandsTo8Spaces verifies that a tab at visual column 8
// (already at a tab stop) expands to 8 spaces, reaching the next stop at 16.
// Spec: "at column 8, expands to 8 spaces"
// Spec: "Tab stops are at visual columns 0, 8, 16, 24, etc."
func TestDrawTabAtColumn8ExpandsTo8Spaces(t *testing.T) {
	// "\t\t": first tab@0 → 8 spaces (stops at col 8);
	// second tab@8 → 8 spaces (stops at col 16).
	m := newTabMemo(30, 2)
	m.SetText("\t\t")

	buf := NewDrawBuffer(30, 2)
	m.Draw(buf)

	// Cols 8–15 must all be spaces (8-space expansion of tab at col 8).
	for col := 8; col < 16; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Rune != ' ' {
			t.Errorf("Tab@col8: cell(%d,0).Rune = %q, want ' ' (tab at col 8 expands to 8 spaces)", col, cell.Rune)
		}
	}
}

// TestDrawTabAtColumn8NextContentAt16 verifies content after tab-at-col-8
// appears at visual column 16.
// Spec: "at column 8, expands to 8 spaces"
func TestDrawTabAtColumn8NextContentAt16(t *testing.T) {
	// "\t\tQ": first tab fills to col 8, second tab fills 8→16, 'Q'@16.
	m := newTabMemo(30, 2)
	m.SetText("\t\tQ")

	buf := NewDrawBuffer(30, 2)
	m.Draw(buf)

	cell := buf.GetCell(16, 0)
	if cell.Rune != 'Q' {
		t.Errorf("After tab@col8: cell(16,0).Rune = %q, want 'Q'", cell.Rune)
	}
}

// ---------------------------------------------------------------------------
// Section 5 — Multiple tabs expand correctly
// ---------------------------------------------------------------------------

// TestDrawThreeTabsExpandCorrectly verifies three consecutive tabs place the
// following character at visual column 24.
// Spec: "Tab stops are at visual columns 0, 8, 16, 24, etc."
func TestDrawThreeTabsExpandCorrectly(t *testing.T) {
	// "\t\t\tW": tabs at 0→8, 8→16, 16→24; 'W' at col 24.
	m := newTabMemo(40, 2)
	m.SetText("\t\t\tW")

	buf := NewDrawBuffer(40, 2)
	m.Draw(buf)

	cell := buf.GetCell(24, 0)
	if cell.Rune != 'W' {
		t.Errorf("Three tabs: cell(24,0).Rune = %q, want 'W' (tabs at 0,8,16 stop at 24)", cell.Rune)
	}
}

// TestDrawTabAfterContentAtNonTabStop verifies tab expansion from a column
// that is not a tab stop (e.g., col 5 → tab expands to 3 spaces, reaching col 8).
// Spec: "Tab stops are at visual columns 0, 8, 16, 24, etc."
func TestDrawTabAfterContentAtNonTabStop(t *testing.T) {
	// "ABCDE\tZ": 5 chars, tab@5 expands 3 spaces (to stop at 8), 'Z'@8.
	m := newTabMemo(20, 2)
	m.SetText("ABCDE\tZ")

	buf := NewDrawBuffer(20, 2)
	m.Draw(buf)

	// Cols 5, 6, 7 must be spaces (3-space expansion from col 5 to col 8).
	for col := 5; col < 8; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Rune != ' ' {
			t.Errorf("Tab@col5: cell(%d,0).Rune = %q, want ' '", col, cell.Rune)
		}
	}
	// 'Z' at col 8.
	cell := buf.GetCell(8, 0)
	if cell.Rune != 'Z' {
		t.Errorf("After tab@col5: cell(8,0).Rune = %q, want 'Z'", cell.Rune)
	}
}

// ---------------------------------------------------------------------------
// Section 6 — Tab clipped at widget right edge
// ---------------------------------------------------------------------------

// TestDrawTabClippedAtWidgetRightEdge verifies that tab expansion spaces that
// extend past the right edge of the widget are clipped (not written).
// Spec: "Tabs that extend past the right edge of the widget are clipped"
func TestDrawTabClippedAtWidgetRightEdge(t *testing.T) {
	// Widget is 5 wide. "\t" at col 0 would expand to 8 spaces, but only 5
	// columns exist, so only cols 0–4 are filled; no panic or out-of-bounds.
	m := newTabMemo(5, 2)
	m.SetText("\t")

	buf := NewDrawBuffer(5, 2)
	m.Draw(buf)

	// Cols 0–4 must be spaces (the visible portion of the tab expansion).
	for col := 0; col < 5; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Rune != ' ' {
			t.Errorf("Clipped tab: cell(%d,0).Rune = %q, want ' '", col, cell.Rune)
		}
	}
	// Test also verifies no panic occurs (Draw completes without crashing).
}

// TestDrawTabPartiallyClippedAtRightEdge verifies that when a tab starts
// within the widget but would extend past its right edge, the visible portion
// is still rendered correctly.
// Spec: "Tabs that extend past the right edge of the widget are clipped"
func TestDrawTabPartiallyClippedAtRightEdge(t *testing.T) {
	// Widget is 10 wide. "ABCDE\t" puts tab at col 5, expanding to col 8.
	// All 3 expansion spaces (cols 5,6,7) fit; col 8 and beyond don't exist.
	// Then "\tX": 8 chars wide widget, tab@0 → 8 spaces; 'X' would be at col 8
	// but widget is only 8 wide (cols 0-7), so 'X' is clipped.
	m := newTabMemo(8, 2)
	m.SetText("\tX")

	buf := NewDrawBuffer(8, 2)
	m.Draw(buf)

	// Tab fills all 8 cols (0-7) with spaces; 'X' at col 8 is outside and clipped.
	for col := 0; col < 8; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Rune != ' ' {
			t.Errorf("Tab fully fills width-8 widget: cell(%d,0).Rune = %q, want ' '", col, cell.Rune)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 7 — Tab expansion uses MemoNormal style (no selection)
// ---------------------------------------------------------------------------

// TestDrawTabExpansionUsesMemoNormalStyle verifies that each expanded space of
// a tab uses MemoNormal style when no selection is active.
// Spec: "Each space of a tab expansion uses the same style (normal or selected)
// as the tab character itself"
// (Without a selection, the tab character itself uses MemoNormal.)
func TestDrawTabExpansionUsesMemoNormalStyle(t *testing.T) {
	// "\t" at col 0 → 8 spaces; all must use MemoNormal.
	m := newTabMemo(20, 2)
	m.SetText("\t")

	buf := NewDrawBuffer(20, 2)
	m.Draw(buf)

	scheme := theme.BorlandBlue
	for col := 0; col < 8; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Style != scheme.MemoNormal {
			t.Errorf("Tab expansion style: cell(%d,0).Style = %v, want MemoNormal %v", col, cell.Style, scheme.MemoNormal)
		}
	}
}

// TestDrawTabExpansionAfterTextUsesMemoNormal verifies MemoNormal for tab
// expansion after ordinary text (mid-line tab, no selection).
// Spec: "Each space of a tab expansion uses the same style (normal or selected)
// as the tab character itself"
func TestDrawTabExpansionAfterTextUsesMemoNormal(t *testing.T) {
	// "AB\t": 'A'@0, 'B'@1, tab@2 expands to 6 spaces (cols 2–7).
	m := newTabMemo(20, 2)
	m.SetText("AB\t")

	buf := NewDrawBuffer(20, 2)
	m.Draw(buf)

	scheme := theme.BorlandBlue
	for col := 2; col < 8; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Style != scheme.MemoNormal {
			t.Errorf("Tab expansion after text: cell(%d,0).Style = %v, want MemoNormal %v", col, cell.Style, scheme.MemoNormal)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 8 — Tab in selection uses MemoSelected for all expanded spaces
// ---------------------------------------------------------------------------

// TestDrawTabInSelectionUsesMemoSelectedForAllSpaces verifies that when a tab
// character falls within a selection, all of its expanded spaces use
// MemoSelected style.
// Spec: "Each space of a tab expansion uses the same style (normal or selected)
// as the tab character itself"
func TestDrawTabInSelectionUsesMemoSelectedForAllSpaces(t *testing.T) {
	// "\tA": tab at col 0, 'A' at col 8.
	// Ctrl+A selects all → tab rune is within the selection.
	// All 8 expanded spaces must use MemoSelected.
	m := newTabMemo(20, 2)
	m.SetText("\tA")
	m.HandleEvent(tabCtrlKeyEv(tcell.KeyCtrlA))
	// selStart=(0,0), selEnd=(0,2) — tab at rune 0, 'A' at rune 1.

	buf := NewDrawBuffer(20, 2)
	m.Draw(buf)

	scheme := theme.BorlandBlue
	// The tab at rune index 0 expands to cols 0–7; all must be MemoSelected.
	for col := 0; col < 8; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Style != scheme.MemoSelected {
			t.Errorf("Tab in selection: cell(%d,0).Style = %v, want MemoSelected %v", col, cell.Style, scheme.MemoSelected)
		}
		if cell.Rune != ' ' {
			t.Errorf("Tab in selection: cell(%d,0).Rune = %q, want ' ' (tab expands to spaces)", col, cell.Rune)
		}
	}
}

// TestDrawTabPartiallyInSelectionBeforeTabNotSelected verifies that when a tab
// is not within the selection, its expansion uses MemoNormal.
// Spec: "Each space of a tab expansion uses the same style (normal or selected)
// as the tab character itself"
// (The tab character's style is determined by whether its rune index is within
// the selection range.)
func TestDrawTabPartiallyInSelectionBeforeTabNotSelected(t *testing.T) {
	// "AB\tCD": 'A'@rune0, 'B'@rune1, tab@rune2, 'C'@rune3, 'D'@rune4.
	// Select only 'A' and 'B': Shift+Right twice from (0,0).
	// selStart=(0,0), selEnd=(0,2). Tab at rune 2 is NOT in [0,2) → MemoNormal.
	m := newTabMemo(20, 2)
	m.SetText("AB\tCD")
	m.HandleEvent(tabShiftKeyEv(tcell.KeyRight))
	m.HandleEvent(tabShiftKeyEv(tcell.KeyRight))
	// selStart=(0,0), selEnd=(0,2): only rune indices 0 and 1 are selected.

	buf := NewDrawBuffer(20, 2)
	m.Draw(buf)

	scheme := theme.BorlandBlue
	// Tab expands at visual cols 2–7 (6 spaces from col 2 to stop at 8).
	// These spaces must be MemoNormal because the tab rune is outside the selection.
	for col := 2; col < 8; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Style != scheme.MemoNormal {
			t.Errorf("Tab outside selection: cell(%d,0).Style = %v, want MemoNormal %v", col, cell.Style, scheme.MemoNormal)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 9 — Mixed content: text + tab + text renders correctly
// ---------------------------------------------------------------------------

// TestDrawMixedContentTextTabTextRunesAndStyles verifies that in a line
// containing ordinary text, a tab, and more text, each segment renders its
// runes and styles correctly.
// Spec: "Tab characters (\t) expand visually to the next multiple of 8 columns"
// Spec: "Each space of a tab expansion uses the same style (normal or selected)
// as the tab character itself"
func TestDrawMixedContentTextTabTextRunesAndStyles(t *testing.T) {
	// "Hi\tBye": 'H'@0, 'i'@1, tab@2 expands to 6 spaces (cols 2-7), 'B'@8, 'y'@9, 'e'@10.
	m := newTabMemo(20, 2)
	m.SetText("Hi\tBye")

	buf := NewDrawBuffer(20, 2)
	m.Draw(buf)

	scheme := theme.BorlandBlue

	// 'H' at col 0.
	if buf.GetCell(0, 0).Rune != 'H' {
		t.Errorf("Mixed: cell(0,0).Rune = %q, want 'H'", buf.GetCell(0, 0).Rune)
	}
	if buf.GetCell(0, 0).Style != scheme.MemoNormal {
		t.Errorf("Mixed: cell(0,0).Style = %v, want MemoNormal %v", buf.GetCell(0, 0).Style, scheme.MemoNormal)
	}

	// 'i' at col 1.
	if buf.GetCell(1, 0).Rune != 'i' {
		t.Errorf("Mixed: cell(1,0).Rune = %q, want 'i'", buf.GetCell(1, 0).Rune)
	}

	// Tab expansion at cols 2–7: spaces with MemoNormal.
	for col := 2; col < 8; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Rune != ' ' {
			t.Errorf("Mixed tab expansion: cell(%d,0).Rune = %q, want ' '", col, cell.Rune)
		}
		if cell.Style != scheme.MemoNormal {
			t.Errorf("Mixed tab expansion: cell(%d,0).Style = %v, want MemoNormal %v", col, cell.Style, scheme.MemoNormal)
		}
	}

	// 'B' at col 8.
	if buf.GetCell(8, 0).Rune != 'B' {
		t.Errorf("Mixed: cell(8,0).Rune = %q, want 'B'", buf.GetCell(8, 0).Rune)
	}
	if buf.GetCell(8, 0).Style != scheme.MemoNormal {
		t.Errorf("Mixed: cell(8,0).Style = %v, want MemoNormal %v", buf.GetCell(8, 0).Style, scheme.MemoNormal)
	}

	// 'y' at col 9.
	if buf.GetCell(9, 0).Rune != 'y' {
		t.Errorf("Mixed: cell(9,0).Rune = %q, want 'y'", buf.GetCell(9, 0).Rune)
	}

	// 'e' at col 10.
	if buf.GetCell(10, 0).Rune != 'e' {
		t.Errorf("Mixed: cell(10,0).Rune = %q, want 'e'", buf.GetCell(10, 0).Rune)
	}
}

// TestDrawCursorColumnCountsRunesNotVisualColumns verifies that cursor column
// is rune-based, not visual-column-based.  After moving right past a tab, the
// cursor rune column increments by 1 (for the tab rune), not by 8.
// Spec: "deltaX remains rune-based (consistent with cursorCol,
// ensureCursorVisible(), and clampCursor()). Tab expansion is display-only"
// Spec: "Cursor column still counts in runes, not visual columns — tab
// rendering is display-only"
func TestDrawCursorColumnCountsRunesNotVisualColumns(t *testing.T) {
	// "\tA": rune 0 is tab, rune 1 is 'A'.
	// Press Right once → cursor should be at rune column 1, not column 8.
	m := newTabMemo(20, 2)
	m.SetText("\tA")
	m.HandleEvent(tabKeyEv(tcell.KeyRight))
	// Cursor moved one rune right, past the tab rune.

	row, col := m.CursorPos()
	if row != 0 || col != 1 {
		t.Errorf("After Right past tab: CursorPos() = (%d,%d), want (0,1); cursor counts runes, not visual columns", row, col)
	}
}
