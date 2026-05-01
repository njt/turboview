package tv

// memo_draw_selection_test.go — Tests for Task 2: Selection-Aware Draw.
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test cites the exact spec sentence it verifies.
//
// Test organisation:
//   Section 1  — No selection: all characters use MemoNormal
//   Section 2  — Single-line selection: selected range uses MemoSelected
//   Section 3  — Single-line selection: characters outside range use MemoNormal
//   Section 4  — Single-line selection: varying start/end positions
//   Section 5  — Backward selection on same line
//   Section 6  — Multi-line selection: first line
//   Section 7  — Multi-line selection: intermediate lines fully selected
//   Section 8  — Multi-line selection: trailing spaces on intermediate lines
//   Section 9  — Multi-line selection: last line
//   Section 10 — Multi-line selection: trailing spaces on last line

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/njt/turboview/theme"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// drawSelKeyEv creates a plain keyboard event with no modifiers.
func drawSelKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key}}
}

// drawSelShiftKeyEv creates a keyboard event with Shift held.
func drawSelShiftKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key, Modifiers: tcell.ModShift}}
}

// drawSelCtrlKeyEv creates a keyboard event with Ctrl held.
func drawSelCtrlKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key, Modifiers: tcell.ModCtrl}}
}

// newDrawSelMemo creates a Memo with BorlandBlue scheme, sized w×h.
func newDrawSelMemo(w, h int) *Memo {
	m := NewMemo(NewRect(0, 0, w, h))
	m.scheme = theme.BorlandBlue
	return m
}

// ---------------------------------------------------------------------------
// Section 1 — No selection: all characters use MemoNormal
// ---------------------------------------------------------------------------

// TestDrawNoSelectionTextUsesMemoNormal verifies that when no selection exists,
// all rendered text characters use MemoNormal style.
// Spec: "Characters outside the selection use MemoNormal style"
// (By extension, with no selection all characters are outside the selection.)
func TestDrawNoSelectionTextUsesMemoNormal(t *testing.T) {
	m := newDrawSelMemo(20, 3)
	m.SetText("hello")
	// No selection — HasSelection() is false.

	buf := NewDrawBuffer(20, 3)
	m.Draw(buf)

	scheme := theme.BorlandBlue
	for i := 0; i < 5; i++ {
		cell := buf.GetCell(i, 0)
		if cell.Style != scheme.MemoNormal {
			t.Errorf("No-selection Draw: cell(%d,0).Style = %v, want MemoNormal %v", i, cell.Style, scheme.MemoNormal)
		}
	}
}

// TestDrawNoSelectionFillSpacesUseMemoNormal verifies that trailing fill spaces
// with no selection also use MemoNormal style.
// Spec: "Characters outside the selection use MemoNormal style"
func TestDrawNoSelectionFillSpacesUseMemoNormal(t *testing.T) {
	m := newDrawSelMemo(10, 2)
	m.SetText("hi")

	buf := NewDrawBuffer(10, 2)
	m.Draw(buf)

	scheme := theme.BorlandBlue
	// Columns 2–9 are fill spaces; they must use MemoNormal.
	for col := 2; col < 10; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Style != scheme.MemoNormal {
			t.Errorf("No-selection fill space at cell(%d,0).Style = %v, want MemoNormal %v", col, cell.Style, scheme.MemoNormal)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 2 — Single-line selection: selected range uses MemoSelected
// ---------------------------------------------------------------------------

// TestDrawSingleLineSelectionSelectedCharsUseMemoSelected verifies that
// characters within the selection range use MemoSelected style.
// Spec: "When a selection exists, characters within the selection range use
// MemoSelected style"
func TestDrawSingleLineSelectionSelectedCharsUseMemoSelected(t *testing.T) {
	// Text "hello"; select chars 1..3 (columns 1 and 2).
	// Shift+Right from col 0 twice → selStart=(0,0), selEnd=(0,2).
	m := newDrawSelMemo(20, 3)
	m.SetText("hello")
	m.HandleEvent(drawSelShiftKeyEv(tcell.KeyRight))
	m.HandleEvent(drawSelShiftKeyEv(tcell.KeyRight))
	// selStart=(0,0), selEnd=(0,2) → columns 0 and 1 are selected.

	buf := NewDrawBuffer(20, 3)
	m.Draw(buf)

	scheme := theme.BorlandBlue
	// Columns 0 and 1 are within [selStartCol, selEndCol) = [0, 2).
	for col := 0; col < 2; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Style != scheme.MemoSelected {
			t.Errorf("Selected char at cell(%d,0).Style = %v, want MemoSelected %v", col, cell.Style, scheme.MemoSelected)
		}
	}
}

// TestDrawSingleLineSelectionSelectedCharsHaveCorrectRune verifies that selected
// characters still show the correct rune (selection style does not corrupt content).
// Spec: "characters within the selection range use MemoSelected style"
// (Style changes, content must not.)
func TestDrawSingleLineSelectionSelectedCharsHaveCorrectRune(t *testing.T) {
	m := newDrawSelMemo(20, 3)
	m.SetText("hello")
	m.HandleEvent(drawSelShiftKeyEv(tcell.KeyRight))
	m.HandleEvent(drawSelShiftKeyEv(tcell.KeyRight))
	// selStart=(0,0), selEnd=(0,2) → 'h' at col 0, 'e' at col 1 are selected.

	buf := NewDrawBuffer(20, 3)
	m.Draw(buf)

	if buf.GetCell(0, 0).Rune != 'h' {
		t.Errorf("Selected cell(0,0).Rune = %q, want 'h'", buf.GetCell(0, 0).Rune)
	}
	if buf.GetCell(1, 0).Rune != 'e' {
		t.Errorf("Selected cell(1,0).Rune = %q, want 'e'", buf.GetCell(1, 0).Rune)
	}
}

// ---------------------------------------------------------------------------
// Section 3 — Single-line selection: characters outside range use MemoNormal
// ---------------------------------------------------------------------------

// TestDrawSingleLineSelectionUnselectedCharsUseMemoNormal verifies that
// characters outside the selection range use MemoNormal style.
// Spec: "Characters outside the selection use MemoNormal style"
func TestDrawSingleLineSelectionUnselectedCharsUseMemoNormal(t *testing.T) {
	// Text "hello"; select only the first two chars (cols 0–1).
	m := newDrawSelMemo(20, 3)
	m.SetText("hello")
	m.HandleEvent(drawSelShiftKeyEv(tcell.KeyRight))
	m.HandleEvent(drawSelShiftKeyEv(tcell.KeyRight))
	// selStart=(0,0), selEnd=(0,2) → cols 2,3,4 are unselected.

	buf := NewDrawBuffer(20, 3)
	m.Draw(buf)

	scheme := theme.BorlandBlue
	for col := 2; col < 5; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Style != scheme.MemoNormal {
			t.Errorf("Unselected char at cell(%d,0).Style = %v, want MemoNormal %v", col, cell.Style, scheme.MemoNormal)
		}
	}
}

// TestDrawSingleLineSelectionTrailingFillUseMemoNormal verifies that trailing
// fill spaces after the line content on the selected line use MemoNormal style
// (the selection ends before them).
// Spec: "Characters outside the selection use MemoNormal style"
// Spec: "Trailing spaces after line content on the last selection line use MemoNormal style"
func TestDrawSingleLineSelectionTrailingFillUseMemoNormal(t *testing.T) {
	// Text "hi" (2 chars); select both chars. Trailing cols 2-9 are fill spaces
	// outside the selection → MemoNormal.
	m := newDrawSelMemo(10, 2)
	m.SetText("hi")
	m.HandleEvent(drawSelShiftKeyEv(tcell.KeyRight))
	m.HandleEvent(drawSelShiftKeyEv(tcell.KeyRight))
	// selStart=(0,0), selEnd=(0,2) → cols 0 and 1 are selected.

	buf := NewDrawBuffer(10, 2)
	m.Draw(buf)

	scheme := theme.BorlandBlue
	for col := 2; col < 10; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Style != scheme.MemoNormal {
			t.Errorf("Trailing fill at cell(%d,0).Style = %v, want MemoNormal %v (not selected)", col, cell.Style, scheme.MemoNormal)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 4 — Single-line selection: varying start/end positions
// ---------------------------------------------------------------------------

// TestDrawSingleLineSelectionMidLineStart verifies a selection that starts
// mid-line renders correctly: pre-selection chars use MemoNormal, selected
// chars use MemoSelected.
// Spec: "On the first line of the selection, only characters from selStartCol
// to end of line are selected"
// Spec: "If selection start and end are on the same line, only characters
// between the two columns are selected"
func TestDrawSingleLineSelectionMidLineStart(t *testing.T) {
	// Text "abcde"; move cursor to col 2, then Shift+Right twice to select cols 2-3.
	m := newDrawSelMemo(20, 3)
	m.SetText("abcde")
	m.HandleEvent(drawSelKeyEv(tcell.KeyRight))
	m.HandleEvent(drawSelKeyEv(tcell.KeyRight))
	// cursor at (0,2)
	m.HandleEvent(drawSelShiftKeyEv(tcell.KeyRight))
	m.HandleEvent(drawSelShiftKeyEv(tcell.KeyRight))
	// selStart=(0,2), selEnd=(0,4) → cols 2 and 3 are selected.

	buf := NewDrawBuffer(20, 3)
	m.Draw(buf)

	scheme := theme.BorlandBlue
	// Cols 0,1 are before the selection → MemoNormal.
	for col := 0; col < 2; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Style != scheme.MemoNormal {
			t.Errorf("Pre-selection cell(%d,0).Style = %v, want MemoNormal %v", col, cell.Style, scheme.MemoNormal)
		}
	}
	// Cols 2,3 are within the selection → MemoSelected.
	for col := 2; col < 4; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Style != scheme.MemoSelected {
			t.Errorf("Selected cell(%d,0).Style = %v, want MemoSelected %v", col, cell.Style, scheme.MemoSelected)
		}
	}
	// Col 4 is after the selection → MemoNormal.
	cell := buf.GetCell(4, 0)
	if cell.Style != scheme.MemoNormal {
		t.Errorf("Post-selection cell(4,0).Style = %v, want MemoNormal %v", cell.Style, scheme.MemoNormal)
	}
}

// TestDrawSingleLineSelectionSelectAll verifies Ctrl+A causes the entire line
// content to render with MemoSelected style.
// Spec: "When a selection exists, characters within the selection range use
// MemoSelected style"
func TestDrawSingleLineSelectionSelectAll(t *testing.T) {
	m := newDrawSelMemo(20, 3)
	m.SetText("hello")
	m.HandleEvent(drawSelCtrlKeyEv(tcell.KeyCtrlA))
	// selStart=(0,0), selEnd=(0,5) → all 5 chars selected.

	buf := NewDrawBuffer(20, 3)
	m.Draw(buf)

	scheme := theme.BorlandBlue
	for col := 0; col < 5; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Style != scheme.MemoSelected {
			t.Errorf("Ctrl+A selected cell(%d,0).Style = %v, want MemoSelected %v", col, cell.Style, scheme.MemoSelected)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 5 — Backward selection on same line
// ---------------------------------------------------------------------------

// TestDrawBackwardSelectionRendersCorrectly verifies that when selStartCol >
// selEndCol (backward selection on same line), the smaller column is treated
// as the visual start so the selection region renders correctly.
// Spec: "When selStartRow == selEndRow but selStartCol > selEndCol (backward
// selection on same line), the smaller column is treated as the visual start"
// Spec: "Selection rendering uses normalizedSelection() to determine display order"
func TestDrawBackwardSelectionRendersCorrectly(t *testing.T) {
	// Text "abcde"; start at col 3, Shift+Left twice → selStart=(0,3), selEnd=(0,1).
	// This is a backward selection: selStartCol(3) > selEndCol(1).
	// Visual range is [1, 3) → cols 1 and 2 should be MemoSelected.
	m := newDrawSelMemo(20, 3)
	m.SetText("abcde")
	m.HandleEvent(drawSelKeyEv(tcell.KeyRight))
	m.HandleEvent(drawSelKeyEv(tcell.KeyRight))
	m.HandleEvent(drawSelKeyEv(tcell.KeyRight))
	// cursor at (0,3)
	m.HandleEvent(drawSelShiftKeyEv(tcell.KeyLeft))
	m.HandleEvent(drawSelShiftKeyEv(tcell.KeyLeft))
	// selStart=(0,3), selEnd=(0,1) — backward selection.

	buf := NewDrawBuffer(20, 3)
	m.Draw(buf)

	scheme := theme.BorlandBlue
	// Col 0 is outside the visual selection → MemoNormal.
	if buf.GetCell(0, 0).Style != scheme.MemoNormal {
		t.Errorf("Backward sel: cell(0,0).Style = %v, want MemoNormal %v (before visual start col 1)", buf.GetCell(0, 0).Style, scheme.MemoNormal)
	}
	// Cols 1 and 2 are within the visual selection → MemoSelected.
	for col := 1; col < 3; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Style != scheme.MemoSelected {
			t.Errorf("Backward sel: cell(%d,0).Style = %v, want MemoSelected %v", col, cell.Style, scheme.MemoSelected)
		}
	}
	// Cols 3 and 4 are outside the visual selection → MemoNormal.
	for col := 3; col < 5; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Style != scheme.MemoNormal {
			t.Errorf("Backward sel: cell(%d,0).Style = %v, want MemoNormal %v (after visual end col 3)", col, cell.Style, scheme.MemoNormal)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 6 — Multi-line selection: first line
// ---------------------------------------------------------------------------

// TestDrawMultiLineSelectionFirstLineOnlyFromStartCol verifies that on the
// first line of a multi-line selection, only characters from selStartCol to
// the end of the line are selected.
// Spec: "On the first line of the selection, only characters from selStartCol
// to end of line are selected"
func TestDrawMultiLineSelectionFirstLineOnlyFromStartCol(t *testing.T) {
	// Text "abc\ndef"; move to col 1, then Shift+Down to extend to row 1.
	// selStart=(0,1), selEnd=(1,0).
	m := newDrawSelMemo(20, 5)
	m.SetText("abc\ndef")
	m.HandleEvent(drawSelKeyEv(tcell.KeyRight))
	// cursor at (0,1)
	m.HandleEvent(drawSelShiftKeyEv(tcell.KeyDown))
	// selStart=(0,1), selEnd=(1,1) — but exact selEndCol depends on cursor
	// column-memory. The test only checks the first line boundary.
	// Use Shift+End to get a definitive extent on row 1.

	buf := NewDrawBuffer(20, 5)
	m.Draw(buf)

	scheme := theme.BorlandBlue
	// Col 0 on row 0 is before selStartCol=1 → MemoNormal.
	cell := buf.GetCell(0, 0)
	if cell.Style != scheme.MemoNormal {
		t.Errorf("First-line pre-selection: cell(0,0).Style = %v, want MemoNormal %v (before selStartCol 1)", cell.Style, scheme.MemoNormal)
	}
	// Cols 1 and 2 on row 0 are from selStartCol to end of line → MemoSelected.
	for col := 1; col < 3; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Style != scheme.MemoSelected {
			t.Errorf("First-line in-selection: cell(%d,0).Style = %v, want MemoSelected %v", col, cell.Style, scheme.MemoSelected)
		}
	}
}

// TestDrawMultiLineSelectionFirstLineCharsBeforeStartAreNormal verifies
// characters before selStartCol on the first line use MemoNormal.
// Spec: "On the first line of the selection, only characters from selStartCol
// to end of line are selected"
func TestDrawMultiLineSelectionFirstLineCharsBeforeStartAreNormal(t *testing.T) {
	// Text "hello\nworld"; move to col 2, Shift+Down → selStart=(0,2), selEnd=(1,2).
	m := newDrawSelMemo(20, 5)
	m.SetText("hello\nworld")
	m.HandleEvent(drawSelKeyEv(tcell.KeyRight))
	m.HandleEvent(drawSelKeyEv(tcell.KeyRight))
	// cursor at (0,2)
	m.HandleEvent(drawSelShiftKeyEv(tcell.KeyDown))
	// selStart=(0,2), selEnd=(1,2).

	buf := NewDrawBuffer(20, 5)
	m.Draw(buf)

	scheme := theme.BorlandBlue
	// Cols 0 and 1 on row 0 are before selStartCol=2 → MemoNormal.
	for col := 0; col < 2; col++ {
		cell := buf.GetCell(col, 0)
		if cell.Style != scheme.MemoNormal {
			t.Errorf("First-line before-start: cell(%d,0).Style = %v, want MemoNormal %v", col, cell.Style, scheme.MemoNormal)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 7 — Multi-line selection: intermediate lines fully selected
// ---------------------------------------------------------------------------

// TestDrawMultiLineSelectionIntermediateLineFullySelected verifies that on
// intermediate lines of a multi-line selection, the full line content is
// rendered with MemoSelected style.
// Spec: "the selection on intermediate lines covers the full line content plus
// trailing spaces to the widget edge (representing the selected newline)"
func TestDrawMultiLineSelectionIntermediateLineFullySelected(t *testing.T) {
	// Text "line0\nline1\nline2"; select from row 0 to row 2.
	// Row 1 ("line1") is an intermediate line → all content columns MemoSelected.
	m := newDrawSelMemo(20, 5)
	m.SetText("line0\nline1\nline2")
	// Shift+Down twice from (0,0) → selStart=(0,0), selEnd=(2,0).
	m.HandleEvent(drawSelShiftKeyEv(tcell.KeyDown))
	m.HandleEvent(drawSelShiftKeyEv(tcell.KeyDown))

	buf := NewDrawBuffer(20, 5)
	m.Draw(buf)

	scheme := theme.BorlandBlue
	// Row 1 is an intermediate line: all 5 content chars ('l','i','n','e','1')
	// must use MemoSelected.
	wantRunes := []rune{'l', 'i', 'n', 'e', '1'}
	for col, want := range wantRunes {
		cell := buf.GetCell(col, 1)
		if cell.Rune != want {
			t.Errorf("Intermediate line content: cell(%d,1).Rune = %q, want %q", col, cell.Rune, want)
		}
		if cell.Style != scheme.MemoSelected {
			t.Errorf("Intermediate line content: cell(%d,1).Style = %v, want MemoSelected %v", col, cell.Style, scheme.MemoSelected)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 8 — Multi-line selection: trailing spaces on intermediate lines
// ---------------------------------------------------------------------------

// TestDrawMultiLineSelectionIntermediateLineTrailingSpacesUseMemoSelected
// verifies that trailing spaces after line content on an intermediate selected
// line use MemoSelected style (the newline is "selected").
// Spec: "Trailing spaces after line content on a selected intermediate line
// use MemoSelected style (the newline is 'selected')"
func TestDrawMultiLineSelectionIntermediateLineTrailingSpacesUseMemoSelected(t *testing.T) {
	// Text "hi\nworld\nend"; widget width=20.
	// Row 0 = "hi" (2 chars), row 1 = "world" (5 chars), row 2 = "end" (3 chars).
	// Select from row 0 to row 2 → row 1 is an intermediate line.
	// Cols 5–19 on row 1 are trailing spaces and must use MemoSelected.
	m := newDrawSelMemo(20, 5)
	m.SetText("hi\nworld\nend")
	m.HandleEvent(drawSelShiftKeyEv(tcell.KeyDown))
	m.HandleEvent(drawSelShiftKeyEv(tcell.KeyDown))
	// selStart=(0,0), selEnd=(2,0).

	buf := NewDrawBuffer(20, 5)
	m.Draw(buf)

	scheme := theme.BorlandBlue
	// Row 1 "world" is 5 chars wide. Cols 5–19 are trailing spaces on the
	// intermediate line → they represent the selected newline → MemoSelected.
	for col := 5; col < 20; col++ {
		cell := buf.GetCell(col, 1)
		if cell.Rune != ' ' {
			t.Errorf("Intermediate trailing: cell(%d,1).Rune = %q, want ' '", col, cell.Rune)
		}
		if cell.Style != scheme.MemoSelected {
			t.Errorf("Intermediate trailing: cell(%d,1).Style = %v, want MemoSelected %v (newline is selected)", col, cell.Style, scheme.MemoSelected)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 9 — Multi-line selection: last line
// ---------------------------------------------------------------------------

// TestDrawMultiLineSelectionLastLineOnlyToEndCol verifies that on the last
// line of a multi-line selection, only characters from the start of the line
// to selEndCol are selected.
// Spec: "On the last line of the selection, only characters from start of
// line to selEndCol are selected"
func TestDrawMultiLineSelectionLastLineOnlyToEndCol(t *testing.T) {
	// Text "abc\ndefgh"; cursor starts at (0,0), Shift+Down then back to row 1
	// with selEnd=(1,3) using Shift+Down from (0,0).
	// Simpler: cursor at (0,0), Shift+Down → selStart=(0,0), selEnd=(1,0).
	// On last line (row 1), only col 0 up to selEndCol=0 is selected — that is
	// an empty selection segment on the last row. Use a larger move to get a
	// non-trivial selEndCol.
	//
	// Strategy: text "abc\ndefgh"; cursor at (0,0), move to (0,3) with Shift+End
	// (selects to end of row 0), then advance the selection to row 1 col 2
	// using Shift+Down then Shift+Left. Too complex. Instead:
	//
	// Use Ctrl+A on two lines to select all, then the last line goes to its full end.
	// Text "ab\ncd"; Ctrl+A → selStart=(0,0), selEnd=(1,2). On row 1, cols 0,1
	// (selEndCol=2 means chars at 0 and 1) are selected; trailing fill (col 2+)
	// is MemoNormal.
	m := newDrawSelMemo(20, 5)
	m.SetText("ab\ncd")
	m.HandleEvent(drawSelCtrlKeyEv(tcell.KeyCtrlA))
	// selStart=(0,0), selEnd=(1,2); row 1 is last line, selEndCol=2.

	buf := NewDrawBuffer(20, 5)
	m.Draw(buf)

	scheme := theme.BorlandBlue
	// On the last line (row 1), cols 0 and 1 are within [0, selEndCol=2) → MemoSelected.
	for col := 0; col < 2; col++ {
		cell := buf.GetCell(col, 1)
		if cell.Style != scheme.MemoSelected {
			t.Errorf("Last-line selected: cell(%d,1).Style = %v, want MemoSelected %v", col, cell.Style, scheme.MemoSelected)
		}
	}
}

// TestDrawMultiLineSelectionLastLineCharsAfterEndColAreNormal verifies that
// characters after selEndCol on the last selected line use MemoNormal.
// Spec: "On the last line of the selection, only characters from start of
// line to selEndCol are selected"
func TestDrawMultiLineSelectionLastLineCharsAfterEndColAreNormal(t *testing.T) {
	// Text "abc\ndefgh"; select first line and partial second line.
	// Ctrl+A → selStart=(0,0), selEnd=(1,5) for "defgh" (5 chars).
	// All of row 1 is selected. Use a partial selection instead:
	// Move to (0,0), Shift+Down gets to row 1 col 0. That gives selEndCol=0
	// which means no chars on the last line are selected. That's valid per the spec
	// ("from start of line to selEndCol"), but the test would be trivial.
	//
	// Use: text "hello\nworld", cursor at (0,2), Shift+Down → selStart=(0,2),
	// selEnd=(1,2). On last line (row 1), cols 0 and 1 are selected; cols 2-4
	// use MemoNormal.
	m := newDrawSelMemo(20, 5)
	m.SetText("hello\nworld")
	m.HandleEvent(drawSelKeyEv(tcell.KeyRight))
	m.HandleEvent(drawSelKeyEv(tcell.KeyRight))
	// cursor at (0,2)
	m.HandleEvent(drawSelShiftKeyEv(tcell.KeyDown))
	// selStart=(0,2), selEnd=(1,2).

	buf := NewDrawBuffer(20, 5)
	m.Draw(buf)

	scheme := theme.BorlandBlue
	// On the last line (row 1, "world"), cols 2-4 ('r','l','d') are after
	// selEndCol=2 → MemoNormal.
	for col := 2; col < 5; col++ {
		cell := buf.GetCell(col, 1)
		if cell.Style != scheme.MemoNormal {
			t.Errorf("Last-line after selEndCol: cell(%d,1).Style = %v, want MemoNormal %v", col, cell.Style, scheme.MemoNormal)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 10 — Multi-line selection: trailing spaces on last line
// ---------------------------------------------------------------------------

// TestDrawMultiLineSelectionLastLineTrailingSpacesUseMemoNormal verifies that
// trailing fill spaces after the last line's content use MemoNormal style
// (the newline after the last selected line is NOT selected).
// Spec: "Trailing spaces after line content on the last selection line use
// MemoNormal style"
func TestDrawMultiLineSelectionLastLineTrailingSpacesUseMemoNormal(t *testing.T) {
	// Text "hi\nworld\nend"; widget width=20; select rows 0 and 1 (row 1 = last).
	// Row 1 "world" has 5 chars; cols 5–19 are trailing fill on the last line
	// → MemoNormal.
	m := newDrawSelMemo(20, 5)
	m.SetText("hi\nworld\nend")
	m.HandleEvent(drawSelShiftKeyEv(tcell.KeyDown))
	m.HandleEvent(drawSelShiftKeyEv(tcell.KeyDown))
	// selStart=(0,0), selEnd=(2,0); row 2 is the last line of the selection.
	// "end" has 3 chars; cols 3–19 on row 2 are trailing fill → MemoNormal.

	buf := NewDrawBuffer(20, 5)
	m.Draw(buf)

	scheme := theme.BorlandBlue
	// Row 2 ("end") is the last selection line; cols 3–19 are trailing fill.
	for col := 3; col < 20; col++ {
		cell := buf.GetCell(col, 2)
		if cell.Rune != ' ' {
			t.Errorf("Last-line trailing: cell(%d,2).Rune = %q, want ' '", col, cell.Rune)
		}
		if cell.Style != scheme.MemoNormal {
			t.Errorf("Last-line trailing: cell(%d,2).Style = %v, want MemoNormal %v (trailing fill on last selection line is not selected)", col, cell.Style, scheme.MemoNormal)
		}
	}
}

// TestDrawSelectionDoesNotAffectRowsBeyondLastLine verifies that rows beyond
// the text content (empty fill rows) always use MemoNormal even during selection.
// Spec: "Characters outside the selection use MemoNormal style"
// (Rows beyond the document have no content and are never selected.)
func TestDrawSelectionDoesNotAffectRowsBeyondLastLine(t *testing.T) {
	// Text "ab"; widget height=4; only 1 line, so rows 1-3 are beyond content.
	m := newDrawSelMemo(10, 4)
	m.SetText("ab")
	m.HandleEvent(drawSelCtrlKeyEv(tcell.KeyCtrlA))
	// Even with Ctrl+A, rows 1-3 are empty rows beyond last line.

	buf := NewDrawBuffer(10, 4)
	m.Draw(buf)

	scheme := theme.BorlandBlue
	for row := 1; row < 4; row++ {
		for col := 0; col < 10; col++ {
			cell := buf.GetCell(col, row)
			if cell.Style != scheme.MemoNormal {
				t.Errorf("Beyond-last-line row %d: cell(%d,%d).Style = %v, want MemoNormal %v", row, col, row, cell.Style, scheme.MemoNormal)
			}
		}
	}
}
