package tv

// memo_cursor_test.go — Tests for Task 4: Memo Cursor Movement.
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test cites the exact spec sentence it verifies.
//
// Test organisation:
//   Section 1  — Left key
//   Section 2  — Right key
//   Section 3  — Up key
//   Section 4  — Down key
//   Section 5  — Home key (smart Home)
//   Section 6  — End key
//   Section 7  — Ctrl+Home
//   Section 8  — Ctrl+End
//   Section 9  — PgUp
//   Section 10 — PgDn
//   Section 11 — Event consumption (cursor keys clear the event)
//   Section 12 — Pass-through (Tab, Alt+x, F-keys do NOT clear event)

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// memoKeyEv creates a plain keyboard event for a named key with no modifiers.
func memoKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key}}
}

// memoCtrlKeyEv creates a keyboard event for a key with Ctrl modifier.
func memoCtrlKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key, Modifiers: tcell.ModCtrl}}
}

// memoAltRuneEv creates a keyboard event for an Alt+rune combination.
func memoAltRuneEv(r rune) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: r, Modifiers: tcell.ModAlt}}
}

// sendKeys sends a sequence of key events to a Memo.
func sendKeys(m *Memo, events ...*Event) {
	for _, ev := range events {
		m.HandleEvent(ev)
	}
}

// repeatKey sends the same key event n times to a Memo.
func repeatKey(m *Memo, key tcell.Key, n int) {
	for i := 0; i < n; i++ {
		m.HandleEvent(memoKeyEv(key))
	}
}

// newTestMemo creates a Memo with fixed bounds for cursor movement tests.
// Width 40, height 10 gives room for text without viewport complications.
func newTestMemo() *Memo {
	return NewMemo(NewRect(0, 0, 40, 10))
}

// newSmallMemo creates a Memo with a small height to exercise PgUp/PgDn.
// height=5 means PgUp/PgDn moves by 4 lines (height-1).
func newSmallMemo(h int) *Memo {
	return NewMemo(NewRect(0, 0, 40, h))
}

// ---------------------------------------------------------------------------
// Section 1 — Left key
// ---------------------------------------------------------------------------

// TestLeftMovesColumnBack verifies Left moves cursor one column to the left.
// Spec: "Left: move cursor one rune left"
func TestLeftMovesColumnBack(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")
	// Advance to col 2.
	repeatKey(m, tcell.KeyRight, 2)

	m.HandleEvent(memoKeyEv(tcell.KeyLeft))

	row, col := m.CursorPos()
	if row != 0 || col != 1 {
		t.Errorf("After Left from col 2: CursorPos() = (%d, %d), want (0, 1)", row, col)
	}
}

// TestLeftAtStartOfLineWrapsToEndOfPreviousLine verifies Left at column 0 on
// a non-first row wraps to end of the previous line.
// Spec: "at start of line wraps to end of previous line"
func TestLeftAtStartOfLineWrapsToEndOfPreviousLine(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello\nworld") // line 0: 5 chars, line 1: 5 chars
	// Move to start of line 1 (row=1, col=0).
	m.HandleEvent(memoKeyEv(tcell.KeyDown))

	m.HandleEvent(memoKeyEv(tcell.KeyLeft))

	row, col := m.CursorPos()
	if row != 0 || col != 5 {
		t.Errorf("Left at start of line 1: CursorPos() = (%d, %d), want (0, 5) (end of line 0)", row, col)
	}
}

// TestLeftAtOriginDoesNothing verifies Left at (0,0) has no effect.
// Spec: "at (0,0) does nothing"
func TestLeftAtOriginDoesNothing(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")
	// Cursor already at (0,0) after SetText.

	m.HandleEvent(memoKeyEv(tcell.KeyLeft))

	row, col := m.CursorPos()
	if row != 0 || col != 0 {
		t.Errorf("Left at (0,0): CursorPos() = (%d, %d), want (0, 0)", row, col)
	}
}

// TestLeftFromColOneReachesColZero verifies that Left from column 1 reaches
// column 0, not wrapping prematurely.
// Falsifying: an off-by-one error might cause wrap when col==1.
func TestLeftFromColOneReachesColZero(t *testing.T) {
	m := newTestMemo()
	m.SetText("abc")
	m.HandleEvent(memoKeyEv(tcell.KeyRight)) // col 1

	m.HandleEvent(memoKeyEv(tcell.KeyLeft))

	row, col := m.CursorPos()
	if row != 0 || col != 0 {
		t.Errorf("Left from col 1: CursorPos() = (%d, %d), want (0, 0)", row, col)
	}
}

// ---------------------------------------------------------------------------
// Section 2 — Right key
// ---------------------------------------------------------------------------

// TestRightMovesColumnForward verifies Right moves cursor one column to the right.
// Spec: "Right: move cursor one rune right"
func TestRightMovesColumnForward(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")

	m.HandleEvent(memoKeyEv(tcell.KeyRight))

	row, col := m.CursorPos()
	if row != 0 || col != 1 {
		t.Errorf("After Right from col 0: CursorPos() = (%d, %d), want (0, 1)", row, col)
	}
}

// TestRightAtEndOfLineWrapsToStartOfNextLine verifies Right at end of a non-last
// line wraps to start of the next line.
// Spec: "at end of line wraps to start of next line"
func TestRightAtEndOfLineWrapsToStartOfNextLine(t *testing.T) {
	m := newTestMemo()
	m.SetText("hi\nworld") // line 0: 2 chars, line 1: 5 chars
	// Move to end of line 0 (col 2).
	repeatKey(m, tcell.KeyRight, 2)

	m.HandleEvent(memoKeyEv(tcell.KeyRight))

	row, col := m.CursorPos()
	if row != 1 || col != 0 {
		t.Errorf("Right at end of line 0: CursorPos() = (%d, %d), want (1, 0)", row, col)
	}
}

// TestRightAtEndOfDocumentDoesNothing verifies Right at the last position has no effect.
// Spec: "at end of document does nothing"
func TestRightAtEndOfDocumentDoesNothing(t *testing.T) {
	m := newTestMemo()
	m.SetText("ab") // single line, 2 chars
	// Move to end of document (row=0, col=2).
	repeatKey(m, tcell.KeyRight, 2)

	m.HandleEvent(memoKeyEv(tcell.KeyRight))

	row, col := m.CursorPos()
	if row != 0 || col != 2 {
		t.Errorf("Right at end of document: CursorPos() = (%d, %d), want (0, 2)", row, col)
	}
}

// TestRightAtEndOfDocumentMultiLine verifies Right at last position of a
// multi-line document does nothing.
// Spec: "at end of document does nothing"
func TestRightAtEndOfDocumentMultiLine(t *testing.T) {
	m := newTestMemo()
	m.SetText("one\ntwo") // line 1 is "two", 3 chars
	// Navigate to end: row=1, col=3.
	m.HandleEvent(memoKeyEv(tcell.KeyDown))
	repeatKey(m, tcell.KeyRight, 3)

	m.HandleEvent(memoKeyEv(tcell.KeyRight))

	row, col := m.CursorPos()
	if row != 1 || col != 3 {
		t.Errorf("Right at end of multi-line doc: CursorPos() = (%d, %d), want (1, 3)", row, col)
	}
}

// ---------------------------------------------------------------------------
// Section 3 — Up key
// ---------------------------------------------------------------------------

// TestUpMovesRowUp verifies Up moves cursor one row up.
// Spec: "Up: move one line up"
func TestUpMovesRowUp(t *testing.T) {
	m := newTestMemo()
	m.SetText("first\nsecond")
	m.HandleEvent(memoKeyEv(tcell.KeyDown)) // row 1

	m.HandleEvent(memoKeyEv(tcell.KeyUp))

	row, col := m.CursorPos()
	if row != 0 {
		t.Errorf("After Up from row 1: CursorPos() row = %d, want 0", row)
	}
	_ = col
}

// TestUpAtRowZeroDoesNothing verifies Up at row 0 has no effect.
// Spec: "at row 0 does nothing"
func TestUpAtRowZeroDoesNothing(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello\nworld")
	// Cursor at (0,0).

	m.HandleEvent(memoKeyEv(tcell.KeyUp))

	row, col := m.CursorPos()
	if row != 0 || col != 0 {
		t.Errorf("Up at row 0: CursorPos() = (%d, %d), want (0, 0)", row, col)
	}
}

// TestUpClampsColumnToTargetLineLength verifies Up clamps column when the
// target line is shorter than the current column.
// Spec: "column clamped to target line length"
func TestUpClampsColumnToTargetLineLength(t *testing.T) {
	m := newTestMemo()
	m.SetText("hi\nlong line here") // line 0 is "hi" (2 chars), line 1 is longer
	// Go to row 1, col 5 (beyond "hi" length).
	m.HandleEvent(memoKeyEv(tcell.KeyDown))
	repeatKey(m, tcell.KeyRight, 5)

	m.HandleEvent(memoKeyEv(tcell.KeyUp))

	row, col := m.CursorPos()
	if row != 0 {
		t.Errorf("Up should have moved to row 0, got row %d", row)
	}
	// Line 0 is "hi", length 2. Column must clamp to 2.
	if col != 2 {
		t.Errorf("Up: column should clamp to 2 (length of \"hi\"), got %d", col)
	}
}

// TestUpPreservesColumnWhenTargetLineIsLongEnough verifies Up preserves
// column position when the target line is at least as long.
// Spec: "column clamped to target line length" — no clamping needed here.
func TestUpPreservesColumnWhenTargetLineIsLongEnough(t *testing.T) {
	m := newTestMemo()
	m.SetText("long enough\nshort") // line 0 is 11 chars, line 1 is 5 chars
	// Move to row 1, col 3.
	m.HandleEvent(memoKeyEv(tcell.KeyDown))
	repeatKey(m, tcell.KeyRight, 3)

	m.HandleEvent(memoKeyEv(tcell.KeyUp))

	row, col := m.CursorPos()
	if row != 0 || col != 3 {
		t.Errorf("Up to longer line: CursorPos() = (%d, %d), want (0, 3)", row, col)
	}
}

// ---------------------------------------------------------------------------
// Section 4 — Down key
// ---------------------------------------------------------------------------

// TestDownMovesRowDown verifies Down moves cursor one row down.
// Spec: "Down: move one line down"
func TestDownMovesRowDown(t *testing.T) {
	m := newTestMemo()
	m.SetText("first\nsecond")

	m.HandleEvent(memoKeyEv(tcell.KeyDown))

	row, _ := m.CursorPos()
	if row != 1 {
		t.Errorf("After Down from row 0: CursorPos() row = %d, want 1", row)
	}
}

// TestDownAtLastRowDoesNothing verifies Down at last row has no effect.
// Spec: "at last row does nothing"
func TestDownAtLastRowDoesNothing(t *testing.T) {
	m := newTestMemo()
	m.SetText("only\none") // 2 lines, last row is row 1
	m.HandleEvent(memoKeyEv(tcell.KeyDown)) // now at row 1

	m.HandleEvent(memoKeyEv(tcell.KeyDown))

	row, _ := m.CursorPos()
	if row != 1 {
		t.Errorf("Down at last row: CursorPos() row = %d, want 1 (stayed)", row)
	}
}

// TestDownClampsColumnToTargetLineLength verifies Down clamps column when the
// target line is shorter than the current column.
// Spec: "column clamped to target line length"
func TestDownClampsColumnToTargetLineLength(t *testing.T) {
	m := newTestMemo()
	m.SetText("long line here\nhi") // line 0 is long, line 1 is "hi" (2 chars)
	// Move to col 8 on row 0.
	repeatKey(m, tcell.KeyRight, 8)

	m.HandleEvent(memoKeyEv(tcell.KeyDown))

	row, col := m.CursorPos()
	if row != 1 {
		t.Errorf("Down should have moved to row 1, got row %d", row)
	}
	// Line 1 is "hi", length 2. Column must clamp to 2.
	if col != 2 {
		t.Errorf("Down: column should clamp to 2 (length of \"hi\"), got %d", col)
	}
}

// TestDownPreservesColumnWhenTargetLineIsLongEnough verifies Down preserves
// column when the target line is long enough.
// Spec: "column clamped to target line length" — no clamping when target is longer.
func TestDownPreservesColumnWhenTargetLineIsLongEnough(t *testing.T) {
	m := newTestMemo()
	m.SetText("short\nlong enough line") // line 0 is 5 chars, line 1 is 16 chars
	// Move to col 4 on row 0.
	repeatKey(m, tcell.KeyRight, 4)

	m.HandleEvent(memoKeyEv(tcell.KeyDown))

	row, col := m.CursorPos()
	if row != 1 || col != 4 {
		t.Errorf("Down to longer line: CursorPos() = (%d, %d), want (1, 4)", row, col)
	}
}

// ---------------------------------------------------------------------------
// Section 5 — Home key (smart Home)
// ---------------------------------------------------------------------------

// TestHomeMovesToFirstNonWhitespace verifies Home moves to first non-WS char
// when cursor is not already there.
// Spec: "Home: smart Home — move to first non-whitespace char"
func TestHomeMovesToFirstNonWhitespace(t *testing.T) {
	m := newTestMemo()
	m.SetText("   hello") // 3 leading spaces, first non-WS at col 3
	// Start at col 0 — Home should jump to col 3 (first non-WS).
	m.HandleEvent(memoKeyEv(tcell.KeyHome))

	row, col := m.CursorPos()
	if row != 0 || col != 3 {
		t.Errorf("Home (cursor at col 0): CursorPos() = (%d, %d), want (0, 3) (first non-WS)", row, col)
	}
}

// TestHomeWhenAlreadyAtFirstNonWhitespaceGoesToColZero verifies that a second
// Home (when already at first non-WS) moves to column 0.
// Spec: "if already there, move to column 0"
func TestHomeWhenAlreadyAtFirstNonWhitespaceGoesToColZero(t *testing.T) {
	m := newTestMemo()
	m.SetText("   hello") // first non-WS at col 3
	// First Home: go to col 3.
	m.HandleEvent(memoKeyEv(tcell.KeyHome))

	// Second Home: already at first non-WS (col 3), so go to col 0.
	m.HandleEvent(memoKeyEv(tcell.KeyHome))

	row, col := m.CursorPos()
	if row != 0 || col != 0 {
		t.Errorf("Home when at first non-WS: CursorPos() = (%d, %d), want (0, 0)", row, col)
	}
}

// TestHomeFromMiddleOfLineGoesToFirstNonWhitespace verifies Home from a
// mid-line position goes to first non-WS, not col 0.
// Spec: "move to first non-whitespace char"
func TestHomeFromMiddleOfLineGoesToFirstNonWhitespace(t *testing.T) {
	m := newTestMemo()
	m.SetText("  abc") // first non-WS at col 2
	// Position cursor in the middle of text.
	repeatKey(m, tcell.KeyRight, 4) // col 4

	m.HandleEvent(memoKeyEv(tcell.KeyHome))

	row, col := m.CursorPos()
	if row != 0 || col != 2 {
		t.Errorf("Home from col 4 on \"  abc\": CursorPos() = (%d, %d), want (0, 2)", row, col)
	}
}

// TestHomeOnLineWithNoLeadingWhitespace verifies Home on a line with no
// leading WS moves to col 0 directly.
// Spec: "on line with no leading WS" — first non-WS is col 0; smart Home stays at 0.
func TestHomeOnLineWithNoLeadingWhitespace(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello") // no leading whitespace, first non-WS is col 0
	// Move to col 3 first.
	repeatKey(m, tcell.KeyRight, 3)

	m.HandleEvent(memoKeyEv(tcell.KeyHome))

	row, col := m.CursorPos()
	if row != 0 || col != 0 {
		t.Errorf("Home on no-leading-WS line from col 3: CursorPos() = (%d, %d), want (0, 0)", row, col)
	}
}

// TestHomeOnEmptyLine verifies Home on an empty line moves to col 0.
// Spec: "move to first non-whitespace char" — empty line has no non-WS so col 0.
func TestHomeOnEmptyLine(t *testing.T) {
	m := newTestMemo()
	m.SetText("\nhello") // line 0 is empty
	// Cursor at (0,0) — Home should stay at col 0.
	m.HandleEvent(memoKeyEv(tcell.KeyHome))

	row, col := m.CursorPos()
	if row != 0 || col != 0 {
		t.Errorf("Home on empty line: CursorPos() = (%d, %d), want (0, 0)", row, col)
	}
}

// ---------------------------------------------------------------------------
// Section 6 — End key
// ---------------------------------------------------------------------------

// TestEndMovesToEndOfLine verifies End moves cursor to end of the current line.
// Spec: "End: move to end of current line"
func TestEndMovesToEndOfLine(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello") // 5 chars, end is col 5

	m.HandleEvent(memoKeyEv(tcell.KeyEnd))

	row, col := m.CursorPos()
	if row != 0 || col != 5 {
		t.Errorf("End on \"hello\": CursorPos() = (%d, %d), want (0, 5)", row, col)
	}
}

// TestEndOnSecondLine verifies End works on non-first rows.
// Spec: "End: move to end of current line"
func TestEndOnSecondLine(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello\nworld!") // line 1: "world!" = 6 chars
	m.HandleEvent(memoKeyEv(tcell.KeyDown)) // row 1

	m.HandleEvent(memoKeyEv(tcell.KeyEnd))

	row, col := m.CursorPos()
	if row != 1 || col != 6 {
		t.Errorf("End on row 1 \"world!\": CursorPos() = (%d, %d), want (1, 6)", row, col)
	}
}

// TestEndAlreadyAtEndStaysAtEnd verifies End when cursor is already at end
// of line does not move the cursor.
// Spec: "End: move to end of current line" — idempotent.
func TestEndAlreadyAtEndStaysAtEnd(t *testing.T) {
	m := newTestMemo()
	m.SetText("abc")
	repeatKey(m, tcell.KeyRight, 3) // col 3 (end)

	m.HandleEvent(memoKeyEv(tcell.KeyEnd))

	row, col := m.CursorPos()
	if row != 0 || col != 3 {
		t.Errorf("End when already at end: CursorPos() = (%d, %d), want (0, 3)", row, col)
	}
}

// TestEndOnEmptyLineStaysAtZero verifies End on an empty line keeps cursor at col 0.
// Spec: "End: move to end of current line" — empty line end is col 0.
func TestEndOnEmptyLineStaysAtZero(t *testing.T) {
	m := newTestMemo()
	m.SetText("") // single empty line

	m.HandleEvent(memoKeyEv(tcell.KeyEnd))

	row, col := m.CursorPos()
	if row != 0 || col != 0 {
		t.Errorf("End on empty line: CursorPos() = (%d, %d), want (0, 0)", row, col)
	}
}

// ---------------------------------------------------------------------------
// Section 7 — Ctrl+Home
// ---------------------------------------------------------------------------

// TestCtrlHomeMovesToOrigin verifies Ctrl+Home moves cursor to (0, 0).
// Spec: "Ctrl+Home: move to (0, 0)"
func TestCtrlHomeMovesToOrigin(t *testing.T) {
	m := newTestMemo()
	m.SetText("first\nsecond\nthird")
	// Navigate away from origin.
	m.HandleEvent(memoKeyEv(tcell.KeyDown))
	repeatKey(m, tcell.KeyRight, 3)

	m.HandleEvent(memoCtrlKeyEv(tcell.KeyHome))

	row, col := m.CursorPos()
	if row != 0 || col != 0 {
		t.Errorf("Ctrl+Home: CursorPos() = (%d, %d), want (0, 0)", row, col)
	}
}

// TestCtrlHomeFromLastRow verifies Ctrl+Home works from the last row.
// Spec: "Ctrl+Home: move to (0, 0)"
func TestCtrlHomeFromLastRow(t *testing.T) {
	m := newTestMemo()
	m.SetText("a\nb\nc")
	// Navigate to last row.
	repeatKey(m, tcell.KeyDown, 2)
	m.HandleEvent(memoKeyEv(tcell.KeyRight))

	m.HandleEvent(memoCtrlKeyEv(tcell.KeyHome))

	row, col := m.CursorPos()
	if row != 0 || col != 0 {
		t.Errorf("Ctrl+Home from last row: CursorPos() = (%d, %d), want (0, 0)", row, col)
	}
}

// TestCtrlHomeWhenAlreadyAtOriginStays verifies Ctrl+Home at (0,0) keeps cursor at (0,0).
// Spec: "Ctrl+Home: move to (0, 0)" — idempotent.
func TestCtrlHomeWhenAlreadyAtOriginStays(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")
	// Already at (0,0).

	m.HandleEvent(memoCtrlKeyEv(tcell.KeyHome))

	row, col := m.CursorPos()
	if row != 0 || col != 0 {
		t.Errorf("Ctrl+Home when at origin: CursorPos() = (%d, %d), want (0, 0)", row, col)
	}
}

// ---------------------------------------------------------------------------
// Section 8 — Ctrl+End
// ---------------------------------------------------------------------------

// TestCtrlEndMovesToEndOfDocument verifies Ctrl+End moves to (lastRow, len(lastLine)).
// Spec: "Ctrl+End: move to (lastRow, len(lastLine))"
func TestCtrlEndMovesToEndOfDocument(t *testing.T) {
	m := newTestMemo()
	m.SetText("first\nsecond\nthird!") // lastRow=2, lastLine="third!" len=6

	m.HandleEvent(memoCtrlKeyEv(tcell.KeyEnd))

	row, col := m.CursorPos()
	if row != 2 || col != 6 {
		t.Errorf("Ctrl+End: CursorPos() = (%d, %d), want (2, 6)", row, col)
	}
}

// TestCtrlEndSingleLine verifies Ctrl+End on a single-line document.
// Spec: "Ctrl+End: move to (lastRow, len(lastLine))" — lastRow=0 for single line.
func TestCtrlEndSingleLine(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello") // lastRow=0, lastLine="hello" len=5

	m.HandleEvent(memoCtrlKeyEv(tcell.KeyEnd))

	row, col := m.CursorPos()
	if row != 0 || col != 5 {
		t.Errorf("Ctrl+End on single line: CursorPos() = (%d, %d), want (0, 5)", row, col)
	}
}

// TestCtrlEndWhenAlreadyAtEndStays verifies Ctrl+End when already at end is idempotent.
// Spec: "Ctrl+End: move to (lastRow, len(lastLine))"
func TestCtrlEndWhenAlreadyAtEndStays(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")
	repeatKey(m, tcell.KeyRight, 5) // already at end

	m.HandleEvent(memoCtrlKeyEv(tcell.KeyEnd))

	row, col := m.CursorPos()
	if row != 0 || col != 5 {
		t.Errorf("Ctrl+End when already at end: CursorPos() = (%d, %d), want (0, 5)", row, col)
	}
}

// ---------------------------------------------------------------------------
// Section 9 — PgUp
// ---------------------------------------------------------------------------

// TestPgUpMovesUpByHeightMinusOne verifies PgUp moves up by (height-1) lines.
// Spec: "PgUp: move up by (height - 1) lines"
func TestPgUpMovesUpByHeightMinusOne(t *testing.T) {
	// height=5 → PgUp moves 4 lines.
	m := newSmallMemo(5)
	// Build a document with enough lines to move up from.
	m.SetText("line0\nline1\nline2\nline3\nline4\nline5\nline6\nline7")
	// Navigate to row 6.
	repeatKey(m, tcell.KeyDown, 6)

	m.HandleEvent(memoKeyEv(tcell.KeyPgUp))

	row, _ := m.CursorPos()
	if row != 2 {
		t.Errorf("PgUp (height=5): from row 6, CursorPos() row = %d, want 2 (6 - 4)", row)
	}
}

// TestPgUpClampsToRowZero verifies PgUp clamps to row 0 when fewer lines remain.
// Spec: "clamp row to 0"
func TestPgUpClampsToRowZero(t *testing.T) {
	// height=5 → PgUp moves 4 lines; starting from row 2 would go to -2, clamps to 0.
	m := newSmallMemo(5)
	m.SetText("a\nb\nc\nd\ne")
	// Navigate to row 2.
	repeatKey(m, tcell.KeyDown, 2)

	m.HandleEvent(memoKeyEv(tcell.KeyPgUp))

	row, _ := m.CursorPos()
	if row != 0 {
		t.Errorf("PgUp clamped to row 0: CursorPos() row = %d, want 0", row)
	}
}

// TestPgUpAtRowZeroDoesNotGoBelowZero verifies PgUp at row 0 stays at row 0.
// Spec: "clamp row to 0"
func TestPgUpAtRowZeroDoesNotGoBelowZero(t *testing.T) {
	m := newSmallMemo(5)
	m.SetText("line0\nline1\nline2")
	// Cursor at row 0.

	m.HandleEvent(memoKeyEv(tcell.KeyPgUp))

	row, _ := m.CursorPos()
	if row != 0 {
		t.Errorf("PgUp at row 0: CursorPos() row = %d, want 0", row)
	}
}

// ---------------------------------------------------------------------------
// Section 10 — PgDn
// ---------------------------------------------------------------------------

// TestPgDnMovesDownByHeightMinusOne verifies PgDn moves down by (height-1) lines.
// Spec: "PgDn: move down by (height - 1) lines"
func TestPgDnMovesDownByHeightMinusOne(t *testing.T) {
	// height=5 → PgDn moves 4 lines.
	m := newSmallMemo(5)
	m.SetText("line0\nline1\nline2\nline3\nline4\nline5\nline6\nline7")
	// Cursor at row 0.

	m.HandleEvent(memoKeyEv(tcell.KeyPgDn))

	row, _ := m.CursorPos()
	if row != 4 {
		t.Errorf("PgDn (height=5): from row 0, CursorPos() row = %d, want 4 (0 + 4)", row)
	}
}

// TestPgDnClampsToLastRow verifies PgDn clamps to last row when fewer lines remain.
// Spec: "clamp row to len(lines)-1"
func TestPgDnClampsToLastRow(t *testing.T) {
	// height=5 → PgDn moves 4 lines; doc has 3 lines (indices 0-2), last row = 2.
	m := newSmallMemo(5)
	m.SetText("a\nb\nc") // 3 lines, last row = 2
	// Cursor at row 0.

	m.HandleEvent(memoKeyEv(tcell.KeyPgDn))

	row, _ := m.CursorPos()
	if row != 2 {
		t.Errorf("PgDn clamped to last row: CursorPos() row = %d, want 2 (last row)", row)
	}
}

// TestPgDnAtLastRowDoesNotExceedLastRow verifies PgDn at last row stays there.
// Spec: "clamp row to len(lines)-1"
func TestPgDnAtLastRowDoesNotExceedLastRow(t *testing.T) {
	m := newSmallMemo(5)
	m.SetText("line0\nline1\nline2")
	// Navigate to last row (row 2).
	repeatKey(m, tcell.KeyDown, 2)

	m.HandleEvent(memoKeyEv(tcell.KeyPgDn))

	row, _ := m.CursorPos()
	if row != 2 {
		t.Errorf("PgDn at last row: CursorPos() row = %d, want 2 (no movement)", row)
	}
}

// ---------------------------------------------------------------------------
// Section 11 — Event consumption
// ---------------------------------------------------------------------------

// TestLeftEventIsConsumed verifies that after handling KeyLeft the event is cleared.
// Spec: "When cursor moves, the event is consumed (event.Clear())"
func TestLeftEventIsConsumed(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")
	repeatKey(m, tcell.KeyRight, 2) // cursor at col 2

	ev := memoKeyEv(tcell.KeyLeft)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Left key event: event should be cleared after cursor moves")
	}
}

// TestRightEventIsConsumed verifies that after handling KeyRight the event is cleared.
// Spec: "When cursor moves, the event is consumed (event.Clear())"
func TestRightEventIsConsumed(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")

	ev := memoKeyEv(tcell.KeyRight)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Right key event: event should be cleared after cursor moves")
	}
}

// TestUpEventIsConsumed verifies that after handling KeyUp the event is cleared.
// Spec: "When cursor moves, the event is consumed (event.Clear())"
func TestUpEventIsConsumed(t *testing.T) {
	m := newTestMemo()
	m.SetText("first\nsecond")
	m.HandleEvent(memoKeyEv(tcell.KeyDown))

	ev := memoKeyEv(tcell.KeyUp)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Up key event: event should be cleared after cursor moves")
	}
}

// TestDownEventIsConsumed verifies that after handling KeyDown the event is cleared.
// Spec: "When cursor moves, the event is consumed (event.Clear())"
func TestDownEventIsConsumed(t *testing.T) {
	m := newTestMemo()
	m.SetText("first\nsecond")

	ev := memoKeyEv(tcell.KeyDown)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Down key event: event should be cleared after cursor moves")
	}
}

// TestHomeEventIsConsumed verifies that after handling KeyHome the event is cleared.
// Spec: "When cursor moves, the event is consumed (event.Clear())"
func TestHomeEventIsConsumed(t *testing.T) {
	m := newTestMemo()
	m.SetText("  hello")
	repeatKey(m, tcell.KeyRight, 4) // cursor not at first non-WS

	ev := memoKeyEv(tcell.KeyHome)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Home key event: event should be cleared after cursor moves")
	}
}

// TestEndEventIsConsumed verifies that after handling KeyEnd the event is cleared.
// Spec: "When cursor moves, the event is consumed (event.Clear())"
func TestEndEventIsConsumed(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")

	ev := memoKeyEv(tcell.KeyEnd)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("End key event: event should be cleared after cursor moves")
	}
}

// TestCtrlHomeEventIsConsumed verifies that Ctrl+Home clears the event.
// Spec: "When cursor moves, the event is consumed (event.Clear())"
func TestCtrlHomeEventIsConsumed(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello\nworld")
	m.HandleEvent(memoKeyEv(tcell.KeyDown))

	ev := memoCtrlKeyEv(tcell.KeyHome)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Ctrl+Home event: event should be cleared after cursor moves")
	}
}

// TestCtrlEndEventIsConsumed verifies that Ctrl+End clears the event.
// Spec: "When cursor moves, the event is consumed (event.Clear())"
func TestCtrlEndEventIsConsumed(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")

	ev := memoCtrlKeyEv(tcell.KeyEnd)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Ctrl+End event: event should be cleared after cursor moves")
	}
}

// TestPgUpEventIsConsumed verifies that PgUp clears the event.
// Spec: "When cursor moves, the event is consumed (event.Clear())"
func TestPgUpEventIsConsumed(t *testing.T) {
	m := newSmallMemo(5)
	m.SetText("line0\nline1\nline2\nline3\nline4")
	repeatKey(m, tcell.KeyDown, 4)

	ev := memoKeyEv(tcell.KeyPgUp)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("PgUp event: event should be cleared after cursor moves")
	}
}

// TestPgDnEventIsConsumed verifies that PgDn clears the event.
// Spec: "When cursor moves, the event is consumed (event.Clear())"
func TestPgDnEventIsConsumed(t *testing.T) {
	m := newSmallMemo(5)
	m.SetText("line0\nline1\nline2\nline3\nline4")

	ev := memoKeyEv(tcell.KeyPgDn)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("PgDn event: event should be cleared after cursor moves")
	}
}

// TestLeftAtOriginEventIsStillConsumed verifies that KeyLeft at (0,0) (does nothing
// positionally) still consumes the event. The spec says movement events are consumed;
// the boundary "does nothing" refers to position, not event handling.
// Spec: "at (0,0) does nothing" + "When cursor moves, the event is consumed"
// Note: this test is intentionally included — if the impl only clears when position
// changes, this test will fail, revealing the discrepancy.
func TestLeftAtOriginEventIsStillConsumed(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")
	// Already at (0,0).

	ev := memoKeyEv(tcell.KeyLeft)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Left at (0,0): event should still be consumed even when position does not change")
	}
}

// ---------------------------------------------------------------------------
// Section 12 — Pass-through (Tab, Alt+x, F-keys do NOT clear event)
// ---------------------------------------------------------------------------

// TestTabKeyIsNotConsumed verifies Tab is not handled by Memo.
// Spec: "Tab key is NOT consumed (passes through for focus navigation)"
func TestTabKeyIsNotConsumed(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")

	ev := memoKeyEv(tcell.KeyTab)
	m.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("Tab key: event should NOT be cleared (Tab passes through for focus navigation)")
	}
}

// TestAltKeyIsNotConsumed verifies Alt+rune events are not handled by Memo.
// Spec: "Alt+anything is NOT consumed (passes through for shortcuts)"
func TestAltKeyIsNotConsumed(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")

	ev := memoAltRuneEv('x')
	m.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("Alt+x key: event should NOT be cleared (Alt+anything passes through)")
	}
}

// TestF1KeyIsNotConsumed verifies F1 is not handled by Memo.
// Spec: "F-keys are NOT consumed (window management)"
func TestF1KeyIsNotConsumed(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")

	ev := memoKeyEv(tcell.KeyF1)
	m.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("F1 key: event should NOT be cleared (F-keys pass through)")
	}
}

// TestF10KeyIsNotConsumed verifies F10 is not handled by Memo.
// Spec: "F-keys are NOT consumed (window management)"
func TestF10KeyIsNotConsumed(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyF10}}
	m.HandleEvent(ev)

	if ev.IsCleared() {
		t.Error("F10 key: event should NOT be cleared (F-keys pass through)")
	}
}

// TestAltWithDifferentRunes verifies multiple Alt+rune combos all pass through.
// Spec: "Alt+anything is NOT consumed (passes through for shortcuts)"
func TestAltWithDifferentRunes(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")

	runes := []rune{'a', 'z', 'n', 'q'}
	for _, r := range runes {
		ev := memoAltRuneEv(r)
		m.HandleEvent(ev)
		if ev.IsCleared() {
			t.Errorf("Alt+%c: event should NOT be cleared (Alt+anything passes through)", r)
		}
	}
}
