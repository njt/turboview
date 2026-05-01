package tv

// memo_selection_test.go — Tests for Task 1: Selection Model and Shift+Movement.
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test cites the exact spec sentence it verifies.
//
// Test organisation:
//   Section 1  — Selection() and HasSelection() baseline
//   Section 2  — SetText() resets selection
//   Section 3  — Ctrl+A selects all
//   Section 4  — Shift+Right extends selection
//   Section 5  — Shift+Left extends selection
//   Section 6  — Shift+Down extends selection
//   Section 7  — Shift+Up extends selection
//   Section 8  — Shift+Home extends selection (smart home)
//   Section 9  — Shift+End extends selection
//   Section 10 — Shift+Ctrl+Home extends selection to document start
//   Section 11 — Shift+Ctrl+End extends selection to document end
//   Section 12 — Shift+PgUp extends selection
//   Section 13 — Shift+PgDn extends selection
//   Section 14 — Non-Shift cursor keys collapse selection
//   Section 15 — Edge cases: Shift+Left at (0,0), Shift+Right at end of doc
//   Section 16 — Anchor is fixed during Shift+movement

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// selKeyEv creates a plain keyboard event with no modifiers.
func selKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key}}
}

// selShiftKeyEv creates a keyboard event with Shift held.
func selShiftKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key, Modifiers: tcell.ModShift}}
}

// selShiftCtrlKeyEv creates a keyboard event with Shift+Ctrl held.
func selShiftCtrlKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key, Modifiers: tcell.ModShift | tcell.ModCtrl}}
}

// selCtrlKeyEv creates a keyboard event with Ctrl held.
func selCtrlKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key, Modifiers: tcell.ModCtrl}}
}

// newSelMemo creates a Memo with generous bounds suitable for selection tests.
func newSelMemo() *Memo {
	return NewMemo(NewRect(0, 0, 40, 10))
}

// newSelSmallMemo creates a Memo with a specific height to exercise PgUp/PgDn.
func newSelSmallMemo(h int) *Memo {
	return NewMemo(NewRect(0, 0, 40, h))
}

// selRepeatKey sends the same key event n times to a Memo.
func selRepeatKey(m *Memo, key tcell.Key, n int) {
	for i := 0; i < n; i++ {
		m.HandleEvent(selKeyEv(key))
	}
}

// ---------------------------------------------------------------------------
// Section 1 — Selection() and HasSelection() baseline
// ---------------------------------------------------------------------------

// TestSelectionDefaultsToZero verifies that a freshly created Memo has all
// selection fields at zero.
// Spec: "Memo struct gains four new fields: selStartRow, selStartCol,
// selEndRow, selEndCol (all int, default 0)"
func TestSelectionDefaultsToZero(t *testing.T) {
	m := newSelMemo()

	sr, sc, er, ec := m.Selection()
	if sr != 0 || sc != 0 || er != 0 || ec != 0 {
		t.Errorf("Selection() = (%d,%d,%d,%d), want (0,0,0,0)", sr, sc, er, ec)
	}
}

// TestHasSelectionFalseByDefault verifies HasSelection returns false when
// start == end (collapsed selection).
// Spec: "HasSelection() bool returns true when selection start != selection end"
func TestHasSelectionFalseByDefault(t *testing.T) {
	m := newSelMemo()

	if m.HasSelection() {
		t.Error("HasSelection() = true on fresh Memo, want false")
	}
}

// TestHasSelectionFalseWhenCollapsed verifies HasSelection is false when all
// four fields are equal but non-zero.
// Falsifying: an implementation that always returns false would pass
// TestHasSelectionFalseByDefault; this forces it to evaluate the condition.
func TestHasSelectionFalseWhenCollapsed(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello\nworld")
	// Move cursor right to make a non-zero position, then do NOT extend selection.
	selRepeatKey(m, tcell.KeyRight, 3)

	// No Shift was held: selection should remain collapsed.
	if m.HasSelection() {
		t.Error("HasSelection() = true after plain cursor movement (no Shift), want false")
	}
}

// TestHasSelectionTrueAfterExtend verifies HasSelection returns true after a
// Shift+movement has extended the selection.
// Spec: "HasSelection() bool returns true when selection start != selection end"
func TestHasSelectionTrueAfterExtend(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello")

	m.HandleEvent(selShiftKeyEv(tcell.KeyRight))

	if !m.HasSelection() {
		t.Error("HasSelection() = false after Shift+Right, want true")
	}
}

// ---------------------------------------------------------------------------
// Section 2 — SetText() resets selection
// ---------------------------------------------------------------------------

// TestSetTextResetsSelection verifies that SetText resets all four selection
// fields to 0.
// Spec: "SetText() resets all four selection fields to 0"
func TestSetTextResetsSelection(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello")
	// Build a non-trivial selection.
	m.HandleEvent(selShiftKeyEv(tcell.KeyRight))
	m.HandleEvent(selShiftKeyEv(tcell.KeyRight))

	// Now reset with new text.
	m.SetText("different content")

	sr, sc, er, ec := m.Selection()
	if sr != 0 || sc != 0 || er != 0 || ec != 0 {
		t.Errorf("After SetText(), Selection() = (%d,%d,%d,%d), want (0,0,0,0)", sr, sc, er, ec)
	}
}

// TestSetTextResetsHasSelection verifies HasSelection returns false after SetText.
// Spec: "SetText() resets all four selection fields to 0"
func TestSetTextResetsHasSelection(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello")
	m.HandleEvent(selShiftKeyEv(tcell.KeyRight))

	m.SetText("new text")

	if m.HasSelection() {
		t.Error("HasSelection() = true after SetText(), want false")
	}
}

// ---------------------------------------------------------------------------
// Section 3 — Ctrl+A selects all
// ---------------------------------------------------------------------------

// TestCtrlASelectsAll verifies that Ctrl+A sets anchor at (0,0) and extent at
// the end of the last line.
// Spec: "Ctrl+A selects all text: anchor at (0,0), cursor at end of last line,
// extent follows cursor"
func TestCtrlASelectsAll(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello\nworld") // line 0: 5 chars, line 1: 5 chars

	m.HandleEvent(selCtrlKeyEv(tcell.KeyCtrlA))

	sr, sc, er, ec := m.Selection()
	if sr != 0 || sc != 0 {
		t.Errorf("After Ctrl+A, selection start = (%d,%d), want (0,0)", sr, sc)
	}
	if er != 1 || ec != 5 {
		t.Errorf("After Ctrl+A, selection end = (%d,%d), want (1,5) end of last line", er, ec)
	}
}

// TestCtrlAHasSelection verifies HasSelection returns true after Ctrl+A.
// Spec: "Ctrl+A selects all text"
func TestCtrlAHasSelection(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello")

	m.HandleEvent(selCtrlKeyEv(tcell.KeyCtrlA))

	if !m.HasSelection() {
		t.Error("HasSelection() = false after Ctrl+A, want true")
	}
}

// TestCtrlACursorAtEndOfDocument verifies that after Ctrl+A the cursor is at
// the end of the document.
// Spec: "After Ctrl+A, the cursor is at the end of the document (last row, last col)"
func TestCtrlACursorAtEndOfDocument(t *testing.T) {
	m := newSelMemo()
	m.SetText("abc\nde\nfghij") // last line: row 2, 5 chars → col 5

	m.HandleEvent(selCtrlKeyEv(tcell.KeyCtrlA))

	row, col := m.CursorPos()
	if row != 2 || col != 5 {
		t.Errorf("After Ctrl+A, CursorPos() = (%d,%d), want (2,5)", row, col)
	}
}

// TestCtrlAOnSingleLineText verifies Ctrl+A on single-line text.
// Falsifying: an implementation that only considers multi-line could fail here.
// Spec: "anchor at (0,0), cursor at end of last line"
func TestCtrlAOnSingleLineText(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello") // single line, 5 chars

	m.HandleEvent(selCtrlKeyEv(tcell.KeyCtrlA))

	sr, sc, er, ec := m.Selection()
	if sr != 0 || sc != 0 || er != 0 || ec != 5 {
		t.Errorf("After Ctrl+A on single line, Selection() = (%d,%d,%d,%d), want (0,0,0,5)", sr, sc, er, ec)
	}
}

// ---------------------------------------------------------------------------
// Section 4 — Shift+Right extends selection
// ---------------------------------------------------------------------------

// TestMemoShiftRightExtendsSelection verifies that Shift+Right from a collapsed
// cursor moves the extent one rune right while anchor stays fixed.
// Spec: "When Shift is held with any cursor movement key, the selection extends:
// the anchor (selStart) stays fixed, the extent (selEnd) follows the cursor"
func TestMemoShiftRightExtendsSelection(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello")
	// Cursor at (0,0).

	m.HandleEvent(selShiftKeyEv(tcell.KeyRight))

	sr, sc, er, ec := m.Selection()
	if sr != 0 || sc != 0 {
		t.Errorf("After Shift+Right, anchor = (%d,%d), want (0,0)", sr, sc)
	}
	if er != 0 || ec != 1 {
		t.Errorf("After Shift+Right, extent = (%d,%d), want (0,1)", er, ec)
	}
}

// TestShiftRightExtendsTwice verifies that two Shift+Right moves extend the
// extent while the anchor stays fixed.
// Spec: "the anchor (selStart) stays fixed, the extent (selEnd) follows the cursor"
func TestShiftRightExtendsTwice(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello")

	m.HandleEvent(selShiftKeyEv(tcell.KeyRight))
	m.HandleEvent(selShiftKeyEv(tcell.KeyRight))

	sr, sc, er, ec := m.Selection()
	if sr != 0 || sc != 0 {
		t.Errorf("After 2x Shift+Right, anchor = (%d,%d), want (0,0)", sr, sc)
	}
	if er != 0 || ec != 2 {
		t.Errorf("After 2x Shift+Right, extent = (%d,%d), want (0,2)", er, ec)
	}
}

// TestShiftRightSetsAnchorAtCurrentPos verifies that starting a Shift+movement
// from a non-zero cursor position sets the anchor there.
// Spec: "Starting a Shift+movement from a collapsed selection sets the anchor
// at the current cursor position, then moves the cursor and updates the extent"
func TestShiftRightSetsAnchorAtCurrentPos(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello")
	selRepeatKey(m, tcell.KeyRight, 2) // cursor at (0,2)

	m.HandleEvent(selShiftKeyEv(tcell.KeyRight))

	sr, sc, er, ec := m.Selection()
	if sr != 0 || sc != 2 {
		t.Errorf("Anchor after Shift+Right from col 2 = (%d,%d), want (0,2)", sr, sc)
	}
	if er != 0 || ec != 3 {
		t.Errorf("Extent after Shift+Right from col 2 = (%d,%d), want (0,3)", er, ec)
	}
}

// TestShiftRightAtEndOfDocumentDoesNothing verifies Shift+Right at end of
// document has no effect.
// Spec: "Shift+Right at end of document does nothing"
func TestShiftRightAtEndOfDocumentDoesNothing(t *testing.T) {
	m := newSelMemo()
	m.SetText("hi") // 2 chars, end at col 2
	selRepeatKey(m, tcell.KeyRight, 2) // move to end

	m.HandleEvent(selShiftKeyEv(tcell.KeyRight))

	row, col := m.CursorPos()
	if row != 0 || col != 2 {
		t.Errorf("Shift+Right at end of doc: CursorPos() = (%d,%d), want (0,2)", row, col)
	}
	// Selection should still be collapsed (no movement = no extension).
	if m.HasSelection() {
		t.Error("HasSelection() = true after Shift+Right at end of doc with no prior selection, want false")
	}
}

// ---------------------------------------------------------------------------
// Section 5 — Shift+Left extends selection
// ---------------------------------------------------------------------------

// TestShiftLeftExtendsSelectionBackward verifies Shift+Left from mid-line
// creates a backward selection.
// Spec: "the extent (selEnd) follows the cursor"
func TestShiftLeftExtendsSelectionBackward(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello")
	selRepeatKey(m, tcell.KeyRight, 3) // cursor at (0,3)

	m.HandleEvent(selShiftKeyEv(tcell.KeyLeft))

	sr, sc, er, ec := m.Selection()
	if sr != 0 || sc != 3 {
		t.Errorf("Anchor after Shift+Left from col 3 = (%d,%d), want (0,3)", sr, sc)
	}
	if er != 0 || ec != 2 {
		t.Errorf("Extent after Shift+Left from col 3 = (%d,%d), want (0,2)", er, ec)
	}
}

// TestShiftLeftAtOriginDoesNothing verifies Shift+Left at (0,0) does nothing.
// Spec: "Shift+Left at position (0,0) does nothing (no negative selection)"
func TestShiftLeftAtOriginDoesNothing(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello")
	// Cursor is at (0,0) by default after SetText.

	m.HandleEvent(selShiftKeyEv(tcell.KeyLeft))

	row, col := m.CursorPos()
	if row != 0 || col != 0 {
		t.Errorf("Shift+Left at (0,0): CursorPos() = (%d,%d), want (0,0)", row, col)
	}
	if m.HasSelection() {
		t.Error("HasSelection() = true after Shift+Left at (0,0), want false")
	}
}

// TestShiftLeftAtOriginSelectionUnchanged verifies Selection() is unchanged
// after Shift+Left at (0,0).
// Spec: "Shift+Left at position (0,0) does nothing (no negative selection)"
func TestShiftLeftAtOriginSelectionUnchanged(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello")

	m.HandleEvent(selShiftKeyEv(tcell.KeyLeft))

	sr, sc, er, ec := m.Selection()
	if sr != 0 || sc != 0 || er != 0 || ec != 0 {
		t.Errorf("Selection() after Shift+Left at (0,0) = (%d,%d,%d,%d), want (0,0,0,0)", sr, sc, er, ec)
	}
}

// ---------------------------------------------------------------------------
// Section 6 — Shift+Down extends selection
// ---------------------------------------------------------------------------

// TestShiftDownExtendsSelectionToNextRow verifies Shift+Down extends the
// selection to the next row.
// Spec: "the extent (selEnd) follows the cursor"
func TestShiftDownExtendsSelectionToNextRow(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello\nworld")
	// Cursor at (0,0).

	m.HandleEvent(selShiftKeyEv(tcell.KeyDown))

	sr, sc, er, _ := m.Selection()
	if sr != 0 || sc != 0 {
		t.Errorf("Anchor after Shift+Down = (%d,%d), want (0,0)", sr, sc)
	}
	if er != 1 {
		t.Errorf("Extent row after Shift+Down = %d, want 1", er)
	}
}

// TestShiftDownFromLastRowMovesToEndOfLine verifies Shift+Down from the last
// row moves the cursor to end of that line.
// Spec: "Shift+Down from last row moves cursor to end of line, extending selection"
func TestShiftDownFromLastRowMovesToEndOfLine(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello\nworld") // last row is 1
	m.HandleEvent(selKeyEv(tcell.KeyDown)) // move to row 1, col 0

	m.HandleEvent(selShiftKeyEv(tcell.KeyDown))

	row, col := m.CursorPos()
	if row != 1 || col != 5 {
		t.Errorf("Shift+Down from last row: CursorPos() = (%d,%d), want (1,5)", row, col)
	}
}

// TestShiftDownFromLastRowHasSelection verifies HasSelection is true after
// Shift+Down from last row (selection extends to end of line).
// Spec: "Shift+Down from last row moves cursor to end of line, extending selection"
func TestShiftDownFromLastRowHasSelection(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello\nworld") // last row is 1, 5 chars
	m.HandleEvent(selKeyEv(tcell.KeyDown)) // cursor at (1,0)

	m.HandleEvent(selShiftKeyEv(tcell.KeyDown))

	if !m.HasSelection() {
		t.Error("HasSelection() = false after Shift+Down from last row at col 0, want true")
	}
}

// ---------------------------------------------------------------------------
// Section 7 — Shift+Up extends selection
// ---------------------------------------------------------------------------

// TestShiftUpExtendsSelectionUpward verifies Shift+Up moves extent to previous row.
// Spec: "the extent (selEnd) follows the cursor"
func TestShiftUpExtendsSelectionUpward(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello\nworld")
	m.HandleEvent(selKeyEv(tcell.KeyDown)) // cursor at (1,0)

	m.HandleEvent(selShiftKeyEv(tcell.KeyUp))

	sr, sc, er, _ := m.Selection()
	if sr != 1 || sc != 0 {
		t.Errorf("Anchor after Shift+Up from (1,0) = (%d,%d), want (1,0)", sr, sc)
	}
	if er != 0 {
		t.Errorf("Extent row after Shift+Up = %d, want 0", er)
	}
}

// ---------------------------------------------------------------------------
// Section 8 — Shift+Home extends selection (smart home)
// ---------------------------------------------------------------------------

// TestShiftHomeExtendsSelectionToLineStart verifies Shift+Home extends
// selection applying smart home logic.
// Spec: "Shift+Home applies smart home logic while extending selection"
func TestShiftHomeExtendsSelectionToLineStart(t *testing.T) {
	m := newSelMemo()
	m.SetText("  hello") // leading spaces
	selRepeatKey(m, tcell.KeyRight, 5) // cursor at (0,5), past indent

	m.HandleEvent(selShiftKeyEv(tcell.KeyHome))

	sr, sc, er, ec := m.Selection()
	if sr != 0 || sc != 5 {
		t.Errorf("Anchor after Shift+Home = (%d,%d), want (0,5)", sr, sc)
	}
	// Smart home: first press goes to first non-space (col 2).
	if er != 0 || ec != 2 {
		t.Errorf("Extent after Shift+Home (smart home to col 2) = (%d,%d), want (0,2)", er, ec)
	}
}

// TestShiftHomeHasSelection verifies HasSelection after Shift+Home from mid-line.
// Spec: "Shift+Home applies smart home logic while extending selection"
func TestShiftHomeHasSelection(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello")
	selRepeatKey(m, tcell.KeyRight, 3)

	m.HandleEvent(selShiftKeyEv(tcell.KeyHome))

	if !m.HasSelection() {
		t.Error("HasSelection() = false after Shift+Home from col 3, want true")
	}
}

// ---------------------------------------------------------------------------
// Section 9 — Shift+End extends selection
// ---------------------------------------------------------------------------

// TestShiftEndExtendsSelectionToEndOfLine verifies Shift+End extends selection
// to the end of the current line.
// Spec: "Shift+End extends selection to end of current line"
func TestShiftEndExtendsSelectionToEndOfLine(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello\nworld")
	// Cursor at (0,0).

	m.HandleEvent(selShiftKeyEv(tcell.KeyEnd))

	sr, sc, er, ec := m.Selection()
	if sr != 0 || sc != 0 {
		t.Errorf("Anchor after Shift+End = (%d,%d), want (0,0)", sr, sc)
	}
	if er != 0 || ec != 5 {
		t.Errorf("Extent after Shift+End = (%d,%d), want (0,5)", er, ec)
	}
}

// TestShiftEndHasSelection verifies HasSelection is true after Shift+End from col 0.
// Spec: "Shift+End extends selection to end of current line"
func TestShiftEndHasSelection(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello")

	m.HandleEvent(selShiftKeyEv(tcell.KeyEnd))

	if !m.HasSelection() {
		t.Error("HasSelection() = false after Shift+End from col 0, want true")
	}
}

// TestShiftEndCursorAtEndOfLine verifies cursor position after Shift+End.
// Spec: "the extent (selEnd) follows the cursor"
func TestShiftEndCursorAtEndOfLine(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello") // 5 chars

	m.HandleEvent(selShiftKeyEv(tcell.KeyEnd))

	row, col := m.CursorPos()
	if row != 0 || col != 5 {
		t.Errorf("CursorPos() after Shift+End = (%d,%d), want (0,5)", row, col)
	}
}

// ---------------------------------------------------------------------------
// Section 10 — Shift+Ctrl+Home extends selection to document start
// ---------------------------------------------------------------------------

// TestShiftCtrlHomeExtendsToDocumentStart verifies Shift+Ctrl+Home extends
// selection so extent is at (0,0) while anchor stays at original position.
// Spec: "Shift+Ctrl+Home extends selection to document start"
func TestShiftCtrlHomeExtendsToDocumentStart(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello\nworld")
	m.HandleEvent(selKeyEv(tcell.KeyDown)) // cursor at (1,0)
	selRepeatKey(m, tcell.KeyRight, 3)      // cursor at (1,3)

	m.HandleEvent(selShiftCtrlKeyEv(tcell.KeyHome))

	sr, sc, er, ec := m.Selection()
	if sr != 1 || sc != 3 {
		t.Errorf("Anchor after Shift+Ctrl+Home = (%d,%d), want (1,3)", sr, sc)
	}
	if er != 0 || ec != 0 {
		t.Errorf("Extent after Shift+Ctrl+Home = (%d,%d), want (0,0)", er, ec)
	}
}

// TestShiftCtrlHomeCursorAtDocumentStart verifies cursor lands at (0,0).
// Spec: "Shift+Ctrl+Home extends selection to document start"
func TestShiftCtrlHomeCursorAtDocumentStart(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello\nworld")
	m.HandleEvent(selKeyEv(tcell.KeyDown))

	m.HandleEvent(selShiftCtrlKeyEv(tcell.KeyHome))

	row, col := m.CursorPos()
	if row != 0 || col != 0 {
		t.Errorf("CursorPos() after Shift+Ctrl+Home = (%d,%d), want (0,0)", row, col)
	}
}

// ---------------------------------------------------------------------------
// Section 11 — Shift+Ctrl+End extends selection to document end
// ---------------------------------------------------------------------------

// TestShiftCtrlEndExtendsToDocumentEnd verifies Shift+Ctrl+End extends
// selection so extent is at end of last line.
// Spec: "Shift+Ctrl+End extends selection to document end"
func TestShiftCtrlEndExtendsToDocumentEnd(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello\nworld") // last line row=1, 5 chars
	// Cursor at (0,0).

	m.HandleEvent(selShiftCtrlKeyEv(tcell.KeyEnd))

	sr, sc, er, ec := m.Selection()
	if sr != 0 || sc != 0 {
		t.Errorf("Anchor after Shift+Ctrl+End = (%d,%d), want (0,0)", sr, sc)
	}
	if er != 1 || ec != 5 {
		t.Errorf("Extent after Shift+Ctrl+End = (%d,%d), want (1,5)", er, ec)
	}
}

// TestShiftCtrlEndCursorAtDocumentEnd verifies cursor position after Shift+Ctrl+End.
// Spec: "Shift+Ctrl+End extends selection to document end"
func TestShiftCtrlEndCursorAtDocumentEnd(t *testing.T) {
	m := newSelMemo()
	m.SetText("abc\nde") // last line row=1, 2 chars

	m.HandleEvent(selShiftCtrlKeyEv(tcell.KeyEnd))

	row, col := m.CursorPos()
	if row != 1 || col != 2 {
		t.Errorf("CursorPos() after Shift+Ctrl+End = (%d,%d), want (1,2)", row, col)
	}
}

// ---------------------------------------------------------------------------
// Section 12 — Shift+PgUp extends selection
// ---------------------------------------------------------------------------

// TestShiftPgUpExtendsSelectionUpByPage verifies Shift+PgUp extends the
// selection upward by a page.
// Spec: "Shift+PgUp extends selection up by page"
func TestShiftPgUpExtendsSelectionUpByPage(t *testing.T) {
	// Height=5: page size = height-1 = 4 lines.
	m := newSelSmallMemo(5)
	m.SetText("a\nb\nc\nd\ne\nf\ng") // 7 lines (rows 0-6)
	// Move cursor to row 5.
	selRepeatKey(m, tcell.KeyDown, 5) // cursor at (5,0)

	m.HandleEvent(selShiftKeyEv(tcell.KeyPgUp))

	sr, sc, er, _ := m.Selection()
	if sr != 5 || sc != 0 {
		t.Errorf("Anchor after Shift+PgUp = (%d,%d), want (5,0)", sr, sc)
	}
	if er >= 5 {
		t.Errorf("Extent row after Shift+PgUp = %d, want < 5 (moved up by page)", er)
	}
}

// TestShiftPgUpHasSelection verifies HasSelection after Shift+PgUp.
// Spec: "Shift+PgUp extends selection up by page"
func TestShiftPgUpHasSelection(t *testing.T) {
	m := newSelSmallMemo(5)
	m.SetText("a\nb\nc\nd\ne")
	selRepeatKey(m, tcell.KeyDown, 4) // cursor at row 4

	m.HandleEvent(selShiftKeyEv(tcell.KeyPgUp))

	if !m.HasSelection() {
		t.Error("HasSelection() = false after Shift+PgUp, want true")
	}
}

// ---------------------------------------------------------------------------
// Section 13 — Shift+PgDn extends selection
// ---------------------------------------------------------------------------

// TestShiftPgDnExtendsSelectionDownByPage verifies Shift+PgDn extends the
// selection downward by a page.
// Spec: "Shift+PgDn extends selection down by page"
func TestShiftPgDnExtendsSelectionDownByPage(t *testing.T) {
	// Height=5: page size = height-1 = 4 lines.
	m := newSelSmallMemo(5)
	m.SetText("a\nb\nc\nd\ne\nf\ng") // 7 lines (rows 0-6)
	// Cursor at (0,0).

	m.HandleEvent(selShiftKeyEv(tcell.KeyPgDn))

	sr, sc, er, _ := m.Selection()
	if sr != 0 || sc != 0 {
		t.Errorf("Anchor after Shift+PgDn = (%d,%d), want (0,0)", sr, sc)
	}
	if er <= 0 {
		t.Errorf("Extent row after Shift+PgDn = %d, want > 0 (moved down by page)", er)
	}
}

// TestShiftPgDnHasSelection verifies HasSelection after Shift+PgDn.
// Spec: "Shift+PgDn extends selection down by page"
func TestShiftPgDnHasSelection(t *testing.T) {
	m := newSelSmallMemo(5)
	m.SetText("a\nb\nc\nd\ne\nf")

	m.HandleEvent(selShiftKeyEv(tcell.KeyPgDn))

	if !m.HasSelection() {
		t.Error("HasSelection() = false after Shift+PgDn, want true")
	}
}

// ---------------------------------------------------------------------------
// Section 14 — Non-Shift cursor keys collapse selection
// ---------------------------------------------------------------------------

// TestPlainLeftCollapsesSelection verifies that Left without Shift collapses
// any existing selection.
// Spec: "When no Shift is held, all cursor movement keys collapse the selection
// before moving"
func TestPlainLeftCollapsesSelection(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello")
	// Build a selection with Shift+Right.
	m.HandleEvent(selShiftKeyEv(tcell.KeyRight))
	m.HandleEvent(selShiftKeyEv(tcell.KeyRight))
	// Confirm selection exists.
	if !m.HasSelection() {
		t.Fatal("precondition failed: HasSelection() should be true before test")
	}

	m.HandleEvent(selKeyEv(tcell.KeyLeft))

	if m.HasSelection() {
		t.Error("HasSelection() = true after plain Left, want false")
	}
}

// TestPlainRightCollapsesSelection verifies Right without Shift collapses selection.
// Spec: "When no Shift is held, all cursor movement keys collapse the selection"
func TestPlainRightCollapsesSelection(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello")
	m.HandleEvent(selShiftKeyEv(tcell.KeyRight))

	m.HandleEvent(selKeyEv(tcell.KeyRight))

	if m.HasSelection() {
		t.Error("HasSelection() = true after plain Right, want false")
	}
}

// TestPlainUpCollapsesSelection verifies Up without Shift collapses selection.
// Spec: "When no Shift is held, all cursor movement keys collapse the selection"
func TestPlainUpCollapsesSelection(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello\nworld")
	m.HandleEvent(selKeyEv(tcell.KeyDown))
	m.HandleEvent(selShiftKeyEv(tcell.KeyUp))
	if !m.HasSelection() {
		t.Fatal("precondition failed")
	}

	m.HandleEvent(selKeyEv(tcell.KeyUp))

	if m.HasSelection() {
		t.Error("HasSelection() = true after plain Up, want false")
	}
}

// TestPlainDownCollapsesSelection verifies Down without Shift collapses selection.
// Spec: "When no Shift is held, all cursor movement keys collapse the selection"
func TestPlainDownCollapsesSelection(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello\nworld")
	m.HandleEvent(selShiftKeyEv(tcell.KeyDown))
	if !m.HasSelection() {
		t.Fatal("precondition failed")
	}

	m.HandleEvent(selKeyEv(tcell.KeyDown))

	if m.HasSelection() {
		t.Error("HasSelection() = true after plain Down, want false")
	}
}

// TestPlainHomeCollapsesSelection verifies Home without Shift collapses selection.
// Spec: "When no Shift is held, all cursor movement keys collapse the selection"
func TestPlainHomeCollapsesSelection(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello")
	m.HandleEvent(selShiftKeyEv(tcell.KeyRight))
	m.HandleEvent(selShiftKeyEv(tcell.KeyRight))

	m.HandleEvent(selKeyEv(tcell.KeyHome))

	if m.HasSelection() {
		t.Error("HasSelection() = true after plain Home, want false")
	}
}

// TestPlainEndCollapsesSelection verifies End without Shift collapses selection.
// Spec: "When no Shift is held, all cursor movement keys collapse the selection"
func TestPlainEndCollapsesSelection(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello")
	m.HandleEvent(selShiftKeyEv(tcell.KeyRight))

	m.HandleEvent(selKeyEv(tcell.KeyEnd))

	if m.HasSelection() {
		t.Error("HasSelection() = true after plain End, want false")
	}
}

// TestCtrlHomeCollapsesSelection verifies Ctrl+Home without Shift collapses selection.
// Spec: "When no Shift is held, all cursor movement keys collapse the selection"
func TestCtrlHomeCollapsesSelection(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello\nworld")
	m.HandleEvent(selShiftKeyEv(tcell.KeyRight))

	m.HandleEvent(selCtrlKeyEv(tcell.KeyHome))

	if m.HasSelection() {
		t.Error("HasSelection() = true after Ctrl+Home, want false")
	}
}

// TestCtrlEndCollapsesSelection verifies Ctrl+End without Shift collapses selection.
// Spec: "When no Shift is held, all cursor movement keys collapse the selection"
func TestCtrlEndCollapsesSelection(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello\nworld")
	m.HandleEvent(selShiftKeyEv(tcell.KeyRight))

	m.HandleEvent(selCtrlKeyEv(tcell.KeyEnd))

	if m.HasSelection() {
		t.Error("HasSelection() = true after Ctrl+End, want false")
	}
}

// TestPlainPgUpCollapsesSelection verifies PgUp without Shift collapses selection.
// Spec: "When no Shift is held, all cursor movement keys collapse the selection"
func TestPlainPgUpCollapsesSelection(t *testing.T) {
	m := newSelSmallMemo(5)
	m.SetText("a\nb\nc\nd\ne")
	selRepeatKey(m, tcell.KeyDown, 4)
	m.HandleEvent(selShiftKeyEv(tcell.KeyUp))
	if !m.HasSelection() {
		t.Fatal("precondition failed")
	}

	m.HandleEvent(selKeyEv(tcell.KeyPgUp))

	if m.HasSelection() {
		t.Error("HasSelection() = true after plain PgUp, want false")
	}
}

// TestPlainPgDnCollapsesSelection verifies PgDn without Shift collapses selection.
// Spec: "When no Shift is held, all cursor movement keys collapse the selection"
func TestPlainPgDnCollapsesSelection(t *testing.T) {
	m := newSelSmallMemo(5)
	m.SetText("a\nb\nc\nd\ne\nf")
	m.HandleEvent(selShiftKeyEv(tcell.KeyDown))
	if !m.HasSelection() {
		t.Fatal("precondition failed")
	}

	m.HandleEvent(selKeyEv(tcell.KeyPgDn))

	if m.HasSelection() {
		t.Error("HasSelection() = true after plain PgDn, want false")
	}
}

// TestAfterNonShiftCursorKeyHasSelectionFalse is a general catch: confirms
// HasSelection is false after each non-Shift cursor key listed in the spec.
// Spec: "After any non-Shift cursor key, HasSelection() returns false"
func TestAfterNonShiftCursorKeyHasSelectionFalse(t *testing.T) {
	keys := []tcell.Key{
		tcell.KeyLeft,
		tcell.KeyRight,
		tcell.KeyUp,
		tcell.KeyDown,
		tcell.KeyHome,
		tcell.KeyEnd,
	}
	for _, key := range keys {
		m := newSelMemo()
		m.SetText("hello\nworld")
		m.HandleEvent(selKeyEv(tcell.KeyDown)) // move off row 0 to give Up/Home room
		m.HandleEvent(selShiftKeyEv(tcell.KeyRight))

		m.HandleEvent(selKeyEv(key))

		if m.HasSelection() {
			t.Errorf("HasSelection() = true after plain %v, want false", key)
		}
	}
}

// ---------------------------------------------------------------------------
// Section 15 — Edge cases
// ---------------------------------------------------------------------------

// TestSelectionCollapsedAfterCollapsingMove verifies that after a non-Shift
// cursor key, Selection() returns four equal values (a collapsed point).
// Spec: "When no Shift is held, all cursor movement keys collapse the selection
// before moving"
func TestSelectionCollapsedAfterCollapsingMove(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello")
	m.HandleEvent(selShiftKeyEv(tcell.KeyRight))
	m.HandleEvent(selShiftKeyEv(tcell.KeyRight))

	m.HandleEvent(selKeyEv(tcell.KeyLeft))

	sr, sc, er, ec := m.Selection()
	row, col := m.CursorPos()
	// After collapse, all four must equal the cursor position.
	if sr != row || sc != col || er != row || ec != col {
		t.Errorf("After plain Left, Selection() = (%d,%d,%d,%d), cursor at (%d,%d); want all equal cursor pos",
			sr, sc, er, ec, row, col)
	}
}

// ---------------------------------------------------------------------------
// Section 16 — Anchor is fixed during Shift+movement (falsifying)
// ---------------------------------------------------------------------------

// TestAnchorDoesNotMoveOnSubsequentShiftKeys verifies that once an anchor is
// set, subsequent Shift+movements do not update the anchor.
// Spec: "the anchor (selStart) stays fixed, the extent (selEnd) follows the cursor"
func TestAnchorDoesNotMoveOnSubsequentShiftKeys(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello world")
	// Cursor at (0,0) — anchor will be set here.

	m.HandleEvent(selShiftKeyEv(tcell.KeyRight)) // extend: anchor=(0,0), end=(0,1)
	m.HandleEvent(selShiftKeyEv(tcell.KeyRight)) // extend more: anchor=(0,0), end=(0,2)
	m.HandleEvent(selShiftKeyEv(tcell.KeyRight)) // extend more: anchor=(0,0), end=(0,3)

	sr, sc, er, ec := m.Selection()
	if sr != 0 || sc != 0 {
		t.Errorf("Anchor moved after multiple Shift+Right: anchor = (%d,%d), want (0,0)", sr, sc)
	}
	if er != 0 || ec != 3 {
		t.Errorf("Extent after 3x Shift+Right = (%d,%d), want (0,3)", er, ec)
	}
}

// TestAnchorResetOnNewShiftSequence verifies that after a non-Shift key collapses
// the selection, starting a new Shift+movement sets a fresh anchor at the new
// cursor position.
// Spec: "Starting a Shift+movement from a collapsed selection sets the anchor
// at the current cursor position"
func TestAnchorResetOnNewShiftSequence(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello world")

	// First Shift sequence from (0,0).
	m.HandleEvent(selShiftKeyEv(tcell.KeyRight))
	m.HandleEvent(selShiftKeyEv(tcell.KeyRight))
	// Collapse by pressing Right (no Shift); cursor advances to (0,2).
	m.HandleEvent(selKeyEv(tcell.KeyRight))

	// New Shift sequence from current position (0,2).
	m.HandleEvent(selShiftKeyEv(tcell.KeyRight))

	sr, sc, er, ec := m.Selection()
	if sr != 0 || sc != 2 {
		t.Errorf("New anchor after reset = (%d,%d), want (0,2)", sr, sc)
	}
	if er != 0 || ec != 3 {
		t.Errorf("Extent after new Shift+Right from (0,2) = (%d,%d), want (0,3)", er, ec)
	}
}

// TestCtrlAOverwritesPriorSelection verifies Ctrl+A replaces any prior partial
// selection with a full-document selection.
// Spec: "Ctrl+A selects all text: anchor at (0,0), cursor at end of last line"
// Falsifying: an impl that only works from (0,0) could skip the reset step.
func TestCtrlAOverwritesPriorSelection(t *testing.T) {
	m := newSelMemo()
	m.SetText("hello\nworld")
	selRepeatKey(m, tcell.KeyRight, 2)          // cursor at (0,2)
	m.HandleEvent(selShiftKeyEv(tcell.KeyRight)) // partial selection (0,2)→(0,3)

	m.HandleEvent(selCtrlKeyEv(tcell.KeyCtrlA))

	sr, sc, er, ec := m.Selection()
	if sr != 0 || sc != 0 {
		t.Errorf("After Ctrl+A over prior selection, anchor = (%d,%d), want (0,0)", sr, sc)
	}
	if er != 1 || ec != 5 {
		t.Errorf("After Ctrl+A over prior selection, extent = (%d,%d), want (1,5)", er, ec)
	}
}
