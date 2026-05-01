package tv

// memo_edit_test.go — Tests for Task 5: Memo Text Editing.
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test cites the exact spec sentence it verifies.
//
// Test organisation:
//   Section 1  — Insert character (KeyRune)
//   Section 2  — Enter (line split)
//   Section 3  — Enter with auto-indent
//   Section 4  — Backspace within line
//   Section 5  — Backspace at start of line (join with previous)
//   Section 6  — Backspace at (0,0) (no-op)
//   Section 7  — Delete within line
//   Section 8  — Delete at end of line (join with next)
//   Section 9  — Delete at end of document (no-op)
//   Section 10 — Ctrl+Y (delete line)
//   Section 11 — Event consumption
//   Section 12 — Ctrl+Backspace and Ctrl+Delete (deferred, not consumed)

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

// memoRuneEv creates a keyboard event for a printable rune with no modifiers.
func memoRuneEv(r rune) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: r}}
}

// lineCount returns the number of lines in a Memo's text content.
func lineCount(m *Memo) int {
	return len(strings.Split(m.Text(), "\n"))
}

// ---------------------------------------------------------------------------
// Section 1 — Insert character (KeyRune)
// ---------------------------------------------------------------------------

// TestInsertRuneAtStartOfLine verifies inserting a character at column 0
// prepends it to the line and advances the cursor by 1.
// Spec: "Printable character (KeyRune): insert rune at cursor position, advance cursor by 1"
func TestInsertRuneAtStartOfLine(t *testing.T) {
	m := newTestMemo()
	m.SetText("ello")
	// Cursor is at (0,0) after SetText.

	m.HandleEvent(memoRuneEv('h'))

	if got := m.Text(); got != "hello" {
		t.Errorf("Insert 'h' at start: Text() = %q, want %q", got, "hello")
	}
	row, col := m.CursorPos()
	if row != 0 || col != 1 {
		t.Errorf("Insert 'h' at start: CursorPos() = (%d, %d), want (0, 1)", row, col)
	}
}

// TestInsertRuneInMiddleOfLine verifies inserting a character in the middle of
// a line splits correctly and advances the cursor.
// Spec: "insert rune at cursor position, advance cursor by 1"
func TestInsertRuneInMiddleOfLine(t *testing.T) {
	m := newTestMemo()
	m.SetText("hllo")
	// Move cursor to col 1.
	m.HandleEvent(memoKeyEv(tcell.KeyRight))

	m.HandleEvent(memoRuneEv('e'))

	if got := m.Text(); got != "hello" {
		t.Errorf("Insert 'e' at col 1: Text() = %q, want %q", got, "hello")
	}
	row, col := m.CursorPos()
	if row != 0 || col != 2 {
		t.Errorf("Insert 'e' at col 1: CursorPos() = (%d, %d), want (0, 2)", row, col)
	}
}

// TestInsertRuneAtEndOfLine verifies inserting a character at the end of a
// line appends it and advances the cursor.
// Spec: "insert rune at cursor position, advance cursor by 1"
func TestInsertRuneAtEndOfLine(t *testing.T) {
	m := newTestMemo()
	m.SetText("hell")
	// Move to end of line (col 4).
	repeatKey(m, tcell.KeyRight, 4)

	m.HandleEvent(memoRuneEv('o'))

	if got := m.Text(); got != "hello" {
		t.Errorf("Insert 'o' at end: Text() = %q, want %q", got, "hello")
	}
	row, col := m.CursorPos()
	if row != 0 || col != 5 {
		t.Errorf("Insert 'o' at end: CursorPos() = (%d, %d), want (0, 5)", row, col)
	}
}

// TestInsertRuneOnSecondLine verifies inserting a character on a non-first row
// modifies the correct line.
// Spec: "insert rune at cursor position, advance cursor by 1"
func TestInsertRuneOnSecondLine(t *testing.T) {
	m := newTestMemo()
	m.SetText("first\nscnd")
	// Move to row 1.
	m.HandleEvent(memoKeyEv(tcell.KeyDown))
	// Move to col 1.
	m.HandleEvent(memoKeyEv(tcell.KeyRight))

	m.HandleEvent(memoRuneEv('e'))

	if got := m.Text(); got != "first\nsecnd" {
		t.Errorf("Insert 'e' on row 1: Text() = %q, want %q", got, "first\nsecnd")
	}
	row, col := m.CursorPos()
	if row != 1 || col != 2 {
		t.Errorf("Insert 'e' on row 1: CursorPos() = (%d, %d), want (1, 2)", row, col)
	}
}

// ---------------------------------------------------------------------------
// Section 2 — Enter (line split)
// ---------------------------------------------------------------------------

// TestEnterInMiddleOfLineSplitsLine verifies Enter in the middle of a line
// creates a new line with the text after the cursor.
// Spec: "Enter: split current line at cursor position, creating a new line"
func TestEnterInMiddleOfLineSplitsLine(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello world")
	// Move cursor to col 5 (between "hello" and " world").
	repeatKey(m, tcell.KeyRight, 5)

	m.HandleEvent(memoKeyEv(tcell.KeyEnter))

	if got := m.Text(); got != "hello\n world" {
		t.Errorf("Enter at col 5: Text() = %q, want %q", got, "hello\n world")
	}
}

// TestEnterInMiddleMovesTextAfterCursorToNewLine verifies that text after the
// cursor ends up at the start of the new line.
// Spec: "split current line at cursor position"
func TestEnterInMiddleMovesTextAfterCursorToNewLine(t *testing.T) {
	m := newTestMemo()
	m.SetText("abcdef")
	repeatKey(m, tcell.KeyRight, 3) // cursor at col 3

	m.HandleEvent(memoKeyEv(tcell.KeyEnter))

	lines := strings.Split(m.Text(), "\n")
	if len(lines) != 2 {
		t.Fatalf("Enter at col 3: want 2 lines, got %d", len(lines))
	}
	if lines[0] != "abc" {
		t.Errorf("Enter at col 3: line 0 = %q, want %q", lines[0], "abc")
	}
	if lines[1] != "def" {
		t.Errorf("Enter at col 3: line 1 = %q, want %q", lines[1], "def")
	}
}

// TestEnterAtStartOfLineCreatesEmptyLineBefore verifies Enter at column 0
// inserts an empty line before the current line.
// Spec: "split current line at cursor position, creating a new line"
func TestEnterAtStartOfLineCreatesEmptyLineBefore(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")
	// Cursor at (0,0).

	m.HandleEvent(memoKeyEv(tcell.KeyEnter))

	lines := strings.Split(m.Text(), "\n")
	if len(lines) != 2 {
		t.Fatalf("Enter at start: want 2 lines, got %d", len(lines))
	}
	if lines[0] != "" {
		t.Errorf("Enter at start: line 0 = %q, want empty", lines[0])
	}
	if lines[1] != "hello" {
		t.Errorf("Enter at start: line 1 = %q, want %q", lines[1], "hello")
	}
}

// TestEnterAtEndOfLineCreatesEmptyLineAfter verifies Enter at end of line
// appends an empty line.
// Spec: "split current line at cursor position, creating a new line"
func TestEnterAtEndOfLineCreatesEmptyLineAfter(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")
	repeatKey(m, tcell.KeyRight, 5) // end of line

	m.HandleEvent(memoKeyEv(tcell.KeyEnter))

	lines := strings.Split(m.Text(), "\n")
	if len(lines) != 2 {
		t.Fatalf("Enter at end: want 2 lines, got %d", len(lines))
	}
	if lines[0] != "hello" {
		t.Errorf("Enter at end: line 0 = %q, want %q", lines[0], "hello")
	}
	if lines[1] != "" {
		t.Errorf("Enter at end: line 1 = %q, want empty", lines[1])
	}
}

// TestEnterMoveCursorToNewLine verifies Enter places the cursor at the start
// of the new line (or after auto-indent prefix).
// Spec: "split current line at cursor position, creating a new line"
func TestEnterMovesCursorToNewLine(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello world")
	repeatKey(m, tcell.KeyRight, 5) // col 5

	m.HandleEvent(memoKeyEv(tcell.KeyEnter))

	row, col := m.CursorPos()
	if row != 1 {
		t.Errorf("After Enter: cursor row = %d, want 1", row)
	}
	if col != 0 {
		t.Errorf("After Enter (no auto-indent): cursor col = %d, want 0", col)
	}
}

// ---------------------------------------------------------------------------
// Section 3 — Enter with auto-indent
// ---------------------------------------------------------------------------

// TestEnterWithAutoIndentCopiesLeadingSpaces verifies that when auto-indent is
// enabled, Enter copies leading spaces from the current line to the new line.
// Spec: "If auto-indent is enabled, copy leading whitespace (spaces and tabs) from
//        the current line to the new line."
func TestEnterWithAutoIndentCopiesLeadingSpaces(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10), WithAutoIndent(true))
	m.SetText("   hello")
	// Move to end of line.
	repeatKey(m, tcell.KeyRight, 8)

	m.HandleEvent(memoKeyEv(tcell.KeyEnter))

	lines := strings.Split(m.Text(), "\n")
	if len(lines) != 2 {
		t.Fatalf("Enter with auto-indent: want 2 lines, got %d", len(lines))
	}
	if !strings.HasPrefix(lines[1], "   ") {
		t.Errorf("Enter with auto-indent: line 1 = %q, want leading 3 spaces", lines[1])
	}
}

// TestEnterWithAutoIndentCopiesLeadingTabs verifies that leading tabs are
// copied to the new line when auto-indent is enabled.
// Spec: "copy leading whitespace (spaces and tabs) from the current line"
func TestEnterWithAutoIndentCopiesLeadingTabs(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10), WithAutoIndent(true))
	m.SetText("\t\tcode")
	// Move to end of line.
	repeatKey(m, tcell.KeyRight, 6)

	m.HandleEvent(memoKeyEv(tcell.KeyEnter))

	lines := strings.Split(m.Text(), "\n")
	if len(lines) != 2 {
		t.Fatalf("Enter with auto-indent (tabs): want 2 lines, got %d", len(lines))
	}
	if !strings.HasPrefix(lines[1], "\t\t") {
		t.Errorf("Enter with auto-indent (tabs): line 1 = %q, want leading 2 tabs", lines[1])
	}
}

// TestEnterWithAutoIndentCursorAfterIndent verifies the cursor is placed after
// the auto-indented whitespace on the new line.
// Spec: "copy leading whitespace (spaces and tabs) from the current line to the new line"
func TestEnterWithAutoIndentCursorAfterIndent(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10), WithAutoIndent(true))
	m.SetText("  hello")
	repeatKey(m, tcell.KeyRight, 7) // end of line

	m.HandleEvent(memoKeyEv(tcell.KeyEnter))

	row, col := m.CursorPos()
	if row != 1 {
		t.Errorf("Enter with auto-indent: cursor row = %d, want 1", row)
	}
	// Leading whitespace is 2 spaces, cursor should be at col 2.
	if col != 2 {
		t.Errorf("Enter with auto-indent: cursor col = %d, want 2 (after indent)", col)
	}
}

// TestEnterWithAutoIndentDisabledNoIndent verifies no leading whitespace is
// copied when auto-indent is disabled.
// Spec: "If auto-indent is enabled, copy leading whitespace" (implicit: disabled → don't copy)
func TestEnterWithAutoIndentDisabledNoIndent(t *testing.T) {
	m := NewMemo(NewRect(0, 0, 40, 10), WithAutoIndent(false))
	m.SetText("   hello")
	repeatKey(m, tcell.KeyRight, 8)

	m.HandleEvent(memoKeyEv(tcell.KeyEnter))

	lines := strings.Split(m.Text(), "\n")
	if len(lines) != 2 {
		t.Fatalf("Enter with auto-indent disabled: want 2 lines, got %d", len(lines))
	}
	// The new line should be empty (no copied indent).
	if lines[1] != "" {
		t.Errorf("Enter with auto-indent disabled: line 1 = %q, want empty (no indent copied)", lines[1])
	}
}

// TestEnterWithAutoIndentSetAutoIndentDisabledNoIndent verifies SetAutoIndent(false)
// disables indent copying.
// Spec: "If auto-indent is enabled, copy leading whitespace"
func TestEnterSetAutoIndentFalseNoIndent(t *testing.T) {
	m := newTestMemo()
	m.SetAutoIndent(true)
	m.SetAutoIndent(false) // then disable
	m.SetText("  code")
	repeatKey(m, tcell.KeyRight, 6)

	m.HandleEvent(memoKeyEv(tcell.KeyEnter))

	lines := strings.Split(m.Text(), "\n")
	if len(lines) != 2 {
		t.Fatalf("Enter after SetAutoIndent(false): want 2 lines, got %d", len(lines))
	}
	if lines[1] != "" {
		t.Errorf("Enter after SetAutoIndent(false): line 1 = %q, want empty", lines[1])
	}
}

// ---------------------------------------------------------------------------
// Section 4 — Backspace within line
// ---------------------------------------------------------------------------

// TestBackspaceWithinLineDeletesCharBeforeCursor verifies Backspace deletes the
// character immediately before the cursor.
// Spec: "Within line → delete rune before cursor"
func TestBackspaceWithinLineDeletesCharBeforeCursor(t *testing.T) {
	m := newTestMemo()
	m.SetText("hxello")
	// Move to col 2 (after 'x').
	repeatKey(m, tcell.KeyRight, 2)

	m.HandleEvent(memoKeyEv(tcell.KeyBackspace2))

	if got := m.Text(); got != "hello" {
		t.Errorf("Backspace at col 2: Text() = %q, want %q", got, "hello")
	}
}

// TestBackspaceWithinLineMoveCursorBack verifies Backspace moves the cursor back
// one position after deleting.
// Spec: "Within line → delete rune before cursor"
func TestBackspaceWithinLineMoveCursorBack(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")
	repeatKey(m, tcell.KeyRight, 3) // col 3

	m.HandleEvent(memoKeyEv(tcell.KeyBackspace2))

	row, col := m.CursorPos()
	if row != 0 || col != 2 {
		t.Errorf("Backspace at col 3: CursorPos() = (%d, %d), want (0, 2)", row, col)
	}
}

// ---------------------------------------------------------------------------
// Section 5 — Backspace at start of line (join with previous)
// ---------------------------------------------------------------------------

// TestBackspaceAtStartOfLineJoinsWithPreviousLine verifies Backspace at column 0
// on a non-first row appends the current line's content to the previous line.
// Spec: "at start of line → join with previous line (append current line content
//        to previous line, remove current line, cursor at join point)"
func TestBackspaceAtStartOfLineJoinsWithPreviousLine(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello\nworld")
	// Move to row 1, col 0.
	m.HandleEvent(memoKeyEv(tcell.KeyDown))

	m.HandleEvent(memoKeyEv(tcell.KeyBackspace2))

	if got := m.Text(); got != "helloworld" {
		t.Errorf("Backspace at start of row 1: Text() = %q, want %q", got, "helloworld")
	}
}

// TestBackspaceAtStartOfLineReducesLineCount verifies joining removes the current line.
// Spec: "remove current line"
func TestBackspaceAtStartOfLineReducesLineCount(t *testing.T) {
	m := newTestMemo()
	m.SetText("first\nsecond")
	m.HandleEvent(memoKeyEv(tcell.KeyDown)) // row 1

	before := lineCount(m)
	m.HandleEvent(memoKeyEv(tcell.KeyBackspace2))
	after := lineCount(m)

	if after != before-1 {
		t.Errorf("Backspace at start of line: line count %d → %d, want %d", before, after, before-1)
	}
}

// TestBackspaceAtStartOfLineCursorAtJoinPoint verifies the cursor is placed at
// the join point (end of the original previous line content).
// Spec: "cursor at join point"
func TestBackspaceAtStartOfLineCursorAtJoinPoint(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello\nworld") // "hello" is 5 chars, join point is col 5
	m.HandleEvent(memoKeyEv(tcell.KeyDown)) // row 1

	m.HandleEvent(memoKeyEv(tcell.KeyBackspace2))

	row, col := m.CursorPos()
	if row != 0 {
		t.Errorf("Backspace join: cursor row = %d, want 0", row)
	}
	if col != 5 {
		t.Errorf("Backspace join: cursor col = %d, want 5 (join point)", col)
	}
}

// TestBackspaceAtStartOfLineWithEmptyCurrentLine verifies joining when current
// line is empty simply removes the empty line.
// Spec: "append current line content to previous line, remove current line"
func TestBackspaceAtStartOfLineWithEmptyCurrentLine(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello\n")
	m.HandleEvent(memoKeyEv(tcell.KeyDown)) // row 1 (empty)

	m.HandleEvent(memoKeyEv(tcell.KeyBackspace2))

	if got := m.Text(); got != "hello" {
		t.Errorf("Backspace on empty line: Text() = %q, want %q", got, "hello")
	}
}

// ---------------------------------------------------------------------------
// Section 6 — Backspace at (0,0) (no-op)
// ---------------------------------------------------------------------------

// TestBackspaceAtOriginDoesNothing verifies Backspace at (0,0) has no effect.
// Spec: "at start of line → join with previous line" — no previous line at row 0,
//        so the operation is a no-op.
func TestBackspaceAtOriginDoesNothing(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")
	// Cursor already at (0,0).

	m.HandleEvent(memoKeyEv(tcell.KeyBackspace2))

	if got := m.Text(); got != "hello" {
		t.Errorf("Backspace at (0,0): Text() = %q, want unchanged %q", got, "hello")
	}
	row, col := m.CursorPos()
	if row != 0 || col != 0 {
		t.Errorf("Backspace at (0,0): CursorPos() = (%d, %d), want (0, 0)", row, col)
	}
}

// TestBackspaceAtOriginOnEmptyDocumentDoesNothing verifies Backspace at (0,0)
// on an empty document leaves the document unchanged.
// Spec: no-op when there is no previous character and no previous line.
func TestBackspaceAtOriginOnEmptyDocumentDoesNothing(t *testing.T) {
	m := newTestMemo()
	m.SetText("")

	m.HandleEvent(memoKeyEv(tcell.KeyBackspace2))

	if got := m.Text(); got != "" {
		t.Errorf("Backspace on empty doc: Text() = %q, want empty", got)
	}
}

// ---------------------------------------------------------------------------
// Section 7 — Delete within line
// ---------------------------------------------------------------------------

// TestDeleteWithinLineDeletesCharAfterCursor verifies Delete removes the
// character immediately after the cursor.
// Spec: "Within line → delete rune after cursor"
func TestDeleteWithinLineDeletesCharAfterCursor(t *testing.T) {
	m := newTestMemo()
	m.SetText("hexllo")
	// Move to col 2 (before 'x').
	repeatKey(m, tcell.KeyRight, 2)

	m.HandleEvent(memoKeyEv(tcell.KeyDelete))

	if got := m.Text(); got != "hello" {
		t.Errorf("Delete at col 2: Text() = %q, want %q", got, "hello")
	}
}

// TestDeleteWithinLineDoesNotMoveCursor verifies Delete does not change the
// cursor position when within a line.
// Spec: "Within line → delete rune after cursor" (cursor stays)
func TestDeleteWithinLineDoesNotMoveCursor(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")
	repeatKey(m, tcell.KeyRight, 2) // col 2

	m.HandleEvent(memoKeyEv(tcell.KeyDelete))

	row, col := m.CursorPos()
	if row != 0 || col != 2 {
		t.Errorf("Delete at col 2: CursorPos() = (%d, %d), want (0, 2)", row, col)
	}
}

// ---------------------------------------------------------------------------
// Section 8 — Delete at end of line (join with next)
// ---------------------------------------------------------------------------

// TestDeleteAtEndOfLineJoinsWithNextLine verifies Delete at end of line
// appends the next line's content to the current line.
// Spec: "at end of line → join with next line (append next line content to
//        current line, remove next line)"
func TestDeleteAtEndOfLineJoinsWithNextLine(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello\nworld")
	// Move to end of line 0 (col 5).
	repeatKey(m, tcell.KeyRight, 5)

	m.HandleEvent(memoKeyEv(tcell.KeyDelete))

	if got := m.Text(); got != "helloworld" {
		t.Errorf("Delete at end of line 0: Text() = %q, want %q", got, "helloworld")
	}
}

// TestDeleteAtEndOfLineReducesLineCount verifies Delete at end of line removes
// the next line.
// Spec: "remove next line"
func TestDeleteAtEndOfLineReducesLineCount(t *testing.T) {
	m := newTestMemo()
	m.SetText("first\nsecond")
	repeatKey(m, tcell.KeyRight, 5) // end of line 0

	before := lineCount(m)
	m.HandleEvent(memoKeyEv(tcell.KeyDelete))
	after := lineCount(m)

	if after != before-1 {
		t.Errorf("Delete at end of line: line count %d → %d, want %d", before, after, before-1)
	}
}

// TestDeleteAtEndOfLineCursorStaysAtJoinPoint verifies the cursor stays at the
// end of the current line (join point) after the join.
// Spec: "at end of line → join with next line"
func TestDeleteAtEndOfLineCursorStaysAtJoinPoint(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello\nworld")
	repeatKey(m, tcell.KeyRight, 5) // col 5

	m.HandleEvent(memoKeyEv(tcell.KeyDelete))

	row, col := m.CursorPos()
	if row != 0 {
		t.Errorf("Delete join: cursor row = %d, want 0", row)
	}
	if col != 5 {
		t.Errorf("Delete join: cursor col = %d, want 5 (join point)", col)
	}
}

// ---------------------------------------------------------------------------
// Section 9 — Delete at end of document (no-op)
// ---------------------------------------------------------------------------

// TestDeleteAtEndOfDocumentDoesNothing verifies Delete at the last position of
// the document has no effect.
// Spec: "at end of line → join with next line" — no next line, so no-op.
func TestDeleteAtEndOfDocumentDoesNothing(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")
	repeatKey(m, tcell.KeyRight, 5) // end of document

	m.HandleEvent(memoKeyEv(tcell.KeyDelete))

	if got := m.Text(); got != "hello" {
		t.Errorf("Delete at end of document: Text() = %q, want unchanged %q", got, "hello")
	}
	row, col := m.CursorPos()
	if row != 0 || col != 5 {
		t.Errorf("Delete at end of document: CursorPos() = (%d, %d), want (0, 5)", row, col)
	}
}

// TestDeleteAtEndOfMultiLineDocumentDoesNothing verifies Delete at last position
// of a multi-line document has no effect.
// Spec: no next line, so no-op.
func TestDeleteAtEndOfMultiLineDocumentDoesNothing(t *testing.T) {
	m := newTestMemo()
	m.SetText("first\nlast")
	m.HandleEvent(memoKeyEv(tcell.KeyDown))
	repeatKey(m, tcell.KeyRight, 4) // end of "last"

	m.HandleEvent(memoKeyEv(tcell.KeyDelete))

	if got := m.Text(); got != "first\nlast" {
		t.Errorf("Delete at end of multi-line doc: Text() = %q, want unchanged", got)
	}
}

// ---------------------------------------------------------------------------
// Section 10 — Ctrl+Y (delete line)
// ---------------------------------------------------------------------------

// TestCtrlYDeletesEntireLine verifies Ctrl+Y removes the entire current line.
// Spec: "Ctrl+Y: delete entire current line"
func TestCtrlYDeletesEntireLine(t *testing.T) {
	m := newTestMemo()
	m.SetText("first\nsecond\nthird")
	m.HandleEvent(memoKeyEv(tcell.KeyDown)) // row 1

	m.HandleEvent(memoKeyEv(tcell.KeyCtrlY))

	if got := m.Text(); got != "first\nthird" {
		t.Errorf("Ctrl+Y on row 1: Text() = %q, want %q", got, "first\nthird")
	}
}

// TestCtrlYDecreasesLineCount verifies Ctrl+Y reduces the number of lines by 1.
// Spec: "delete entire current line"
func TestCtrlYDecreasesLineCount(t *testing.T) {
	m := newTestMemo()
	m.SetText("line0\nline1\nline2")

	before := lineCount(m)
	m.HandleEvent(memoKeyEv(tcell.KeyCtrlY))
	after := lineCount(m)

	if after != before-1 {
		t.Errorf("Ctrl+Y: line count %d → %d, want %d", before, after, before-1)
	}
}

// TestCtrlYCursorRowStaysTheSame verifies Ctrl+Y keeps the cursor row the same
// when not on the last line.
// Spec: "Cursor row stays the same (or decrements if at last line)"
func TestCtrlYCursorRowStaysTheSame(t *testing.T) {
	m := newTestMemo()
	m.SetText("first\nsecond\nthird")
	m.HandleEvent(memoKeyEv(tcell.KeyDown)) // row 1

	m.HandleEvent(memoKeyEv(tcell.KeyCtrlY))

	row, _ := m.CursorPos()
	if row != 1 {
		t.Errorf("Ctrl+Y on row 1 (non-last): cursor row = %d, want 1 (same)", row)
	}
}

// TestCtrlYOnOnlyLineClearsToEmpty verifies Ctrl+Y on the only line clears
// the content but keeps one empty line.
// Spec: "If it's the only line, clear it to empty"
func TestCtrlYOnOnlyLineClearsToEmpty(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")
	// Only one line.

	m.HandleEvent(memoKeyEv(tcell.KeyCtrlY))

	if got := m.Text(); got != "" {
		t.Errorf("Ctrl+Y on only line: Text() = %q, want empty string", got)
	}
}

// TestCtrlYOnOnlyLineDoesNotRemoveLine verifies Ctrl+Y on the only line keeps
// the document at one line (just clears content).
// Spec: "If it's the only line, clear it to empty" (line count stays 1)
func TestCtrlYOnOnlyLineKeepsOneLineInDocument(t *testing.T) {
	m := newTestMemo()
	m.SetText("content")

	m.HandleEvent(memoKeyEv(tcell.KeyCtrlY))

	if n := lineCount(m); n != 1 {
		t.Errorf("Ctrl+Y on only line: line count = %d, want 1 (still one empty line)", n)
	}
}

// TestCtrlYOnLastLineCursorRowDecrements verifies Ctrl+Y on the last line
// decrements the cursor row.
// Spec: "Cursor row stays the same (or decrements if at last line)"
func TestCtrlYOnLastLineCursorRowDecrements(t *testing.T) {
	m := newTestMemo()
	m.SetText("first\nsecond\nthird")
	repeatKey(m, tcell.KeyDown, 2) // row 2 (last)

	m.HandleEvent(memoKeyEv(tcell.KeyCtrlY))

	row, _ := m.CursorPos()
	// Last row was 2, after deleting it the new last row is 1, so cursor should decrement.
	if row != 1 {
		t.Errorf("Ctrl+Y on last row: cursor row = %d, want 1 (decremented)", row)
	}
}

// TestCtrlYDeletesCorrectLine verifies Ctrl+Y deletes only the current line and
// preserves all others.
// Spec: "delete entire current line"
func TestCtrlYDeletesCorrectLine(t *testing.T) {
	m := newTestMemo()
	m.SetText("alpha\nbeta\ngamma")
	m.HandleEvent(memoKeyEv(tcell.KeyDown)) // row 1 "beta"

	m.HandleEvent(memoKeyEv(tcell.KeyCtrlY))

	lines := strings.Split(m.Text(), "\n")
	if len(lines) != 2 {
		t.Fatalf("Ctrl+Y on row 1: want 2 lines, got %d: %v", len(lines), lines)
	}
	if lines[0] != "alpha" || lines[1] != "gamma" {
		t.Errorf("Ctrl+Y on row 1: lines = %v, want [alpha gamma]", lines)
	}
}

// ---------------------------------------------------------------------------
// Section 11 — Event consumption
// ---------------------------------------------------------------------------

// TestInsertRuneEventIsConsumed verifies inserting a printable character
// clears the event.
// Spec: "Editing operations consume the event (event.Clear())"
func TestInsertRuneEventIsConsumed(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")
	repeatKey(m, tcell.KeyRight, 5)

	ev := memoRuneEv('!')
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("KeyRune event: should be cleared after insert")
	}
}

// TestEnterEventIsConsumed verifies Enter clears the event.
// Spec: "Editing operations consume the event (event.Clear())"
func TestEnterEventIsConsumed(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")

	ev := memoKeyEv(tcell.KeyEnter)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("KeyEnter event: should be cleared after line split")
	}
}

// TestBackspaceEventIsConsumed verifies Backspace clears the event.
// Spec: "Editing operations consume the event (event.Clear())"
func TestBackspaceEventIsConsumed(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")
	m.HandleEvent(memoKeyEv(tcell.KeyRight)) // col 1

	ev := memoKeyEv(tcell.KeyBackspace2)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("KeyBackspace2 event: should be cleared after delete")
	}
}

// TestBackspaceAtOriginEventIsConsumed verifies Backspace at (0,0) still
// clears the event even though it is a no-op positionally.
// Spec: "Editing operations consume the event (event.Clear())"
func TestBackspaceAtOriginEventIsConsumed(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")
	// Cursor at (0,0).

	ev := memoKeyEv(tcell.KeyBackspace2)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("KeyBackspace2 at (0,0): should be cleared even as no-op")
	}
}

// TestDeleteEventIsConsumed verifies Delete clears the event.
// Spec: "Editing operations consume the event (event.Clear())"
func TestDeleteEventIsConsumed(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")

	ev := memoKeyEv(tcell.KeyDelete)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("KeyDelete event: should be cleared after delete")
	}
}

// TestDeleteAtEndOfDocumentEventIsConsumed verifies Delete at end of document
// still clears the event even though it is a no-op.
// Spec: "Editing operations consume the event (event.Clear())"
func TestDeleteAtEndOfDocumentEventIsConsumed(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello")
	repeatKey(m, tcell.KeyRight, 5) // end of document

	ev := memoKeyEv(tcell.KeyDelete)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("KeyDelete at end of document: should be cleared even as no-op")
	}
}

// TestCtrlYEventIsConsumed verifies Ctrl+Y clears the event.
// Spec: "Editing operations consume the event (event.Clear())"
func TestCtrlYEventIsConsumed(t *testing.T) {
	m := newTestMemo()
	m.SetText("hello\nworld")

	ev := memoKeyEv(tcell.KeyCtrlY)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("KeyCtrlY event: should be cleared after line delete")
	}
}

// Section 12 — Ctrl+Backspace and Ctrl+Delete are now implemented in Task 4.
// See memo_word_test.go for the full word-operation test suite.
