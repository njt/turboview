package tv

// memo_word_test.go — Tests for Task 4: Word Movement and Word Deletion in Memo.
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test cites the exact spec sentence it verifies.
//
// Test organisation:
//   Section 1  — Ctrl+Left (previous word)
//   Section 2  — Ctrl+Right (next word)
//   Section 3  — Ctrl+Backspace (delete previous word)
//   Section 4  — Ctrl+Delete (delete next word)
//   Section 5  — Shift+Ctrl+Left (extend selection to previous word)
//   Section 6  — Shift+Ctrl+Right (extend selection to next word)
//   Section 7  — Falsifying tests

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Event helpers
// ---------------------------------------------------------------------------

func wordKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key}}
}

func wordCtrlKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key, Modifiers: tcell.ModCtrl}}
}

func wordShiftCtrlKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key, Modifiers: tcell.ModShift | tcell.ModCtrl}}
}

func wordShiftKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key, Modifiers: tcell.ModShift}}
}

// ---------------------------------------------------------------------------
// Memo constructor for word tests
// ---------------------------------------------------------------------------

// newWordMemo creates a Memo with 60-wide, 20-tall bounds as specified.
func newWordMemo() *Memo {
	return NewMemo(NewRect(0, 0, 60, 20))
}

// wordMoveTo moves the Memo cursor to (row, col) starting from (0,0) using
// Down and Right key presses. It assumes the cursor starts at (0,0) after SetText.
func wordMoveTo(m *Memo, row, col int) {
	for i := 0; i < row; i++ {
		m.HandleEvent(wordKeyEv(tcell.KeyDown))
	}
	for i := 0; i < col; i++ {
		m.HandleEvent(wordKeyEv(tcell.KeyRight))
	}
}

// ---------------------------------------------------------------------------
// Section 1 — Ctrl+Left (previous word)
// ---------------------------------------------------------------------------

// TestMemoCtrlLeftFromMiddleOfWord verifies Ctrl+Left from mid-word moves to
// the start of the current word.
// Spec: "skip backward over characters of the current class"
func TestMemoCtrlLeftFromMiddleOfWord(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")
	// cursor starts at (0,0); move to col 8 (mid "world": "wor|ld")
	wordMoveTo(m, 0, 8)

	m.HandleEvent(wordCtrlKeyEv(tcell.KeyLeft))

	row, col := m.CursorPos()
	if row != 0 || col != 6 {
		t.Errorf("Ctrl+Left from (0,8) in \"hello world\": CursorPos() = (%d,%d), want (0,6) (start of \"world\")", row, col)
	}
}

// TestMemoCtrlLeftFromStartOfWordPrecededBySpace verifies Ctrl+Left from the
// start of a word (with a space before it) skips the space and lands at the
// start of the previous word.
// Spec: "skip backward over whitespace, then skip backward over characters of that class"
func TestMemoCtrlLeftFromStartOfWordPrecededBySpace(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")
	// cursor at col 6 (start of "world")
	wordMoveTo(m, 0, 6)

	m.HandleEvent(wordCtrlKeyEv(tcell.KeyLeft))

	row, col := m.CursorPos()
	if row != 0 || col != 0 {
		t.Errorf("Ctrl+Left from (0,6) in \"hello world\": CursorPos() = (%d,%d), want (0,0) (start of \"hello\")", row, col)
	}
}

// TestMemoCtrlLeftFromMiddleOfWhitespace verifies Ctrl+Left from within
// whitespace skips the remaining whitespace and lands at the start of the
// preceding word.
// Spec: "If the character immediately before the cursor is whitespace, skip whitespace first"
func TestMemoCtrlLeftFromMiddleOfWhitespace(t *testing.T) {
	m := newWordMemo()
	// "hello   world" — three spaces between words
	m.SetText("hello   world")
	// cursor at col 7 (second space, mid-whitespace)
	wordMoveTo(m, 0, 7)

	m.HandleEvent(wordCtrlKeyEv(tcell.KeyLeft))

	row, col := m.CursorPos()
	if row != 0 || col != 0 {
		t.Errorf("Ctrl+Left from (0,7) in \"hello   world\": CursorPos() = (%d,%d), want (0,0) (start of \"hello\")", row, col)
	}
}

// TestMemoCtrlLeftAtColumnZeroNonFirstRow verifies Ctrl+Left at column 0 on
// a non-first row moves to end of the previous line.
// Spec: "If at column 0 and row > 0: move to end of previous line (stop there)"
func TestMemoCtrlLeftAtColumnZeroNonFirstRow(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello\nworld")
	// cursor at (1,0): start of "world"
	wordMoveTo(m, 1, 0)

	m.HandleEvent(wordCtrlKeyEv(tcell.KeyLeft))

	row, col := m.CursorPos()
	if row != 0 || col != 5 {
		t.Errorf("Ctrl+Left at (1,0): CursorPos() = (%d,%d), want (0,5) (end of line 0)", row, col)
	}
}

// TestMemoCtrlLeftAtDocumentStart verifies Ctrl+Left at (0,0) has no effect.
// Spec: "At document start (0,0): no movement"
func TestMemoCtrlLeftAtDocumentStart(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")
	// cursor already at (0,0) after SetText

	m.HandleEvent(wordCtrlKeyEv(tcell.KeyLeft))

	row, col := m.CursorPos()
	if row != 0 || col != 0 {
		t.Errorf("Ctrl+Left at (0,0): CursorPos() = (%d,%d), want (0,0) (no movement)", row, col)
	}
}

// TestMemoCtrlLeftWithPunctuation verifies that punctuation is treated as its
// own character class, so Ctrl+Left from a word stops at the punctuation boundary.
// Spec: "punctuation, and word characters (everything else)... A word boundary is
// a transition between different character classes"
func TestMemoCtrlLeftWithPunctuation(t *testing.T) {
	m := newWordMemo()
	// "hello,world" — comma is punctuation; from col 11 (end) → back through "world"
	// should land at col 6 (start of "world"), not col 5 (the comma)
	m.SetText("hello,world")
	wordMoveTo(m, 0, 11) // end of "world"

	m.HandleEvent(wordCtrlKeyEv(tcell.KeyLeft))

	row, col := m.CursorPos()
	if row != 0 || col != 6 {
		t.Errorf("Ctrl+Left from (0,11) in \"hello,world\": CursorPos() = (%d,%d), want (0,6) (start of \"world\")", row, col)
	}
}

// TestMemoCtrlLeftEventConsumed verifies Ctrl+Left clears the event.
// Spec: "All word operations consume the event (event.Clear())"
func TestMemoCtrlLeftEventConsumed(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")
	wordMoveTo(m, 0, 8)

	ev := wordCtrlKeyEv(tcell.KeyLeft)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Ctrl+Left: event was not consumed (IsCleared() = false)")
	}
}

// ---------------------------------------------------------------------------
// Section 2 — Ctrl+Right (next word)
// ---------------------------------------------------------------------------

// TestMemoCtrlRightFromMiddleOfWord verifies Ctrl+Right from mid-word moves
// past the word and any following whitespace to the start of the next word.
// Spec: "skip forward over characters of the current class, then skip forward over whitespace"
func TestMemoCtrlRightFromMiddleOfWord(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")
	// cursor at col 3 (mid "hello": "hel|lo")
	wordMoveTo(m, 0, 3)

	m.HandleEvent(wordCtrlKeyEv(tcell.KeyRight))

	row, col := m.CursorPos()
	if row != 0 || col != 6 {
		t.Errorf("Ctrl+Right from (0,3) in \"hello world\": CursorPos() = (%d,%d), want (0,6) (start of \"world\")", row, col)
	}
}

// TestMemoCtrlRightFromStartOfWordFollowedBySpace verifies Ctrl+Right from the
// start of a word skips the word and the following space to land at next word.
// Spec: "skip forward over characters of the current class, then skip forward over whitespace"
func TestMemoCtrlRightFromStartOfWordFollowedBySpace(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")
	// cursor at col 0 (start of "hello")

	m.HandleEvent(wordCtrlKeyEv(tcell.KeyRight))

	row, col := m.CursorPos()
	if row != 0 || col != 6 {
		t.Errorf("Ctrl+Right from (0,0) in \"hello world\": CursorPos() = (%d,%d), want (0,6) (start of \"world\")", row, col)
	}
}

// TestMemoCtrlRightFromWhitespace verifies Ctrl+Right from whitespace skips
// the whitespace to land at the start of the next word.
// Spec: "If the character at the cursor is whitespace, skip the whitespace first, then stop"
func TestMemoCtrlRightFromWhitespace(t *testing.T) {
	m := newWordMemo()
	// "hello   world" — three spaces
	m.SetText("hello   world")
	// cursor at col 6 (first space after "hello")
	wordMoveTo(m, 0, 6)

	m.HandleEvent(wordCtrlKeyEv(tcell.KeyRight))

	row, col := m.CursorPos()
	if row != 0 || col != 8 {
		t.Errorf("Ctrl+Right from (0,6) in \"hello   world\": CursorPos() = (%d,%d), want (0,8) (start of \"world\")", row, col)
	}
}

// TestMemoCtrlRightAtEndOfLineNotLastRow verifies Ctrl+Right at end of a
// non-last line moves to start of the next line.
// Spec: "If at end of line and not on last row: move to start of next line (stop there)"
func TestMemoCtrlRightAtEndOfLineNotLastRow(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello\nworld")
	// Move cursor to end of line 0: col 5
	wordMoveTo(m, 0, 5)

	m.HandleEvent(wordCtrlKeyEv(tcell.KeyRight))

	row, col := m.CursorPos()
	if row != 1 || col != 0 {
		t.Errorf("Ctrl+Right at end of line 0: CursorPos() = (%d,%d), want (1,0) (start of next line)", row, col)
	}
}

// TestMemoCtrlRightAtDocumentEnd verifies Ctrl+Right at the document end has no effect.
// Spec: "At document end: no movement"
func TestMemoCtrlRightAtDocumentEnd(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello")
	// cursor at end: (0,5) after navigating there
	wordMoveTo(m, 0, 5)

	m.HandleEvent(wordCtrlKeyEv(tcell.KeyRight))

	row, col := m.CursorPos()
	if row != 0 || col != 5 {
		t.Errorf("Ctrl+Right at document end (0,5): CursorPos() = (%d,%d), want (0,5) (no movement)", row, col)
	}
}

// TestMemoCtrlRightWithPunctuation verifies punctuation is treated as its own class.
// Spec: "punctuation... A word boundary is a transition between different character classes"
func TestMemoCtrlRightWithPunctuation(t *testing.T) {
	m := newWordMemo()
	// "hello,world" — from col 0, Ctrl+Right should skip "hello" then stop at comma (punctuation boundary)
	// The comma is punctuation class, so stopping at col 5 (the comma).
	m.SetText("hello,world")
	// cursor at (0,0)

	m.HandleEvent(wordCtrlKeyEv(tcell.KeyRight))

	row, col := m.CursorPos()
	if row != 0 || col != 5 {
		t.Errorf("Ctrl+Right from (0,0) in \"hello,world\": CursorPos() = (%d,%d), want (0,5) (start of punctuation)", row, col)
	}
}

// TestMemoCtrlRightEventConsumed verifies Ctrl+Right clears the event.
// Spec: "All word operations consume the event (event.Clear())"
func TestMemoCtrlRightEventConsumed(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")

	ev := wordCtrlKeyEv(tcell.KeyRight)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Ctrl+Right: event was not consumed (IsCleared() = false)")
	}
}

// ---------------------------------------------------------------------------
// Section 3 — Ctrl+Backspace (delete previous word)
// ---------------------------------------------------------------------------

// TestMemoCtrlBackspaceDeletesPreviousWord verifies Ctrl+Backspace deletes from
// cursor back to where Ctrl+Left would move.
// Spec: "Deletes from cursor position back to where Ctrl+Left would have moved the cursor"
func TestMemoCtrlBackspaceDeletesPreviousWord(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")
	// cursor at end of "world" = col 11
	wordMoveTo(m, 0, 11)

	m.HandleEvent(wordCtrlKeyEv(tcell.KeyBackspace2))

	if got := m.Text(); got != "hello " {
		t.Errorf("Ctrl+Backspace from (0,11): Text() = %q, want %q", got, "hello ")
	}
}

// TestMemoCtrlBackspaceWithSelectionDeletesSelection verifies that when a
// selection exists, Ctrl+Backspace deletes the selection instead.
// Spec: "If a selection exists, deletes the selection instead (same as regular Backspace with selection)"
func TestMemoCtrlBackspaceWithSelectionDeletesSelection(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")
	// Select "world" using Shift+Right from col 6
	wordMoveTo(m, 0, 6)
	for i := 0; i < 5; i++ {
		m.HandleEvent(wordShiftKeyEv(tcell.KeyRight))
	}
	if !m.HasSelection() {
		t.Fatal("precondition failed: expected selection before Ctrl+Backspace")
	}

	m.HandleEvent(wordCtrlKeyEv(tcell.KeyBackspace2))

	if got := m.Text(); got != "hello " {
		t.Errorf("Ctrl+Backspace with \"world\" selected: Text() = %q, want %q", got, "hello ")
	}
}

// TestMemoCtrlBackspaceWithSelectionDoesNotDeleteAdditionalWord verifies that
// when a selection exists, Ctrl+Backspace deletes only the selection, not an
// additional word.
// Spec: "If a selection exists, deletes the selection instead"
func TestMemoCtrlBackspaceWithSelectionDoesNotDeleteAdditionalWord(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")
	// Select just the single char "w" at col 6
	wordMoveTo(m, 0, 6)
	m.HandleEvent(wordShiftKeyEv(tcell.KeyRight))
	if !m.HasSelection() {
		t.Fatal("precondition failed: expected selection before Ctrl+Backspace")
	}

	m.HandleEvent(wordCtrlKeyEv(tcell.KeyBackspace2))

	// Only "w" (1 char) should be removed, leaving "hello orld"
	if got := m.Text(); got != "hello orld" {
		t.Errorf("Ctrl+Backspace with single-char selection: Text() = %q, want %q", got, "hello orld")
	}
}

// TestMemoCtrlBackspaceAtDocumentStart verifies Ctrl+Backspace at (0,0) is a no-op.
// Spec: Ctrl+Left at (0,0) has no movement, so deletion range is zero-width.
func TestMemoCtrlBackspaceAtDocumentStart(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")
	// cursor already at (0,0)

	m.HandleEvent(wordCtrlKeyEv(tcell.KeyBackspace2))

	if got := m.Text(); got != "hello world" {
		t.Errorf("Ctrl+Backspace at (0,0): Text() = %q, want %q (no change)", got, "hello world")
	}
}

// TestMemoCtrlBackspaceAcrossPunctuationBoundary verifies Ctrl+Backspace
// respects punctuation character class when determining deletion range.
// Spec: "Character classification follows TEditor's getCharType: whitespace, punctuation, word characters"
func TestMemoCtrlBackspaceAcrossPunctuationBoundary(t *testing.T) {
	m := newWordMemo()
	// "hello,world" — from end, one Ctrl+Backspace should delete "world" (word class)
	m.SetText("hello,world")
	wordMoveTo(m, 0, 11) // end

	m.HandleEvent(wordCtrlKeyEv(tcell.KeyBackspace2))

	if got := m.Text(); got != "hello," {
		t.Errorf("Ctrl+Backspace from end of \"hello,world\": Text() = %q, want %q", got, "hello,")
	}
}

// TestMemoCtrlBackspaceEventConsumed verifies Ctrl+Backspace clears the event.
// Spec: "All word operations consume the event (event.Clear())"
func TestMemoCtrlBackspaceEventConsumed(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")
	wordMoveTo(m, 0, 11)

	ev := wordCtrlKeyEv(tcell.KeyBackspace2)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Ctrl+Backspace: event was not consumed (IsCleared() = false)")
	}
}

// ---------------------------------------------------------------------------
// Section 4 — Ctrl+Delete (delete next word)
// ---------------------------------------------------------------------------

// TestMemoCtrlDeleteDeletesNextWord verifies Ctrl+Delete deletes from cursor
// forward to where Ctrl+Right would move.
// Spec: "Deletes from cursor position forward to where Ctrl+Right would have moved the cursor"
func TestMemoCtrlDeleteDeletesNextWord(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")
	// cursor at (0,0)

	m.HandleEvent(wordCtrlKeyEv(tcell.KeyDelete))

	// Ctrl+Right from col 0 lands at col 6; so "hello " is deleted → "world"
	if got := m.Text(); got != "world" {
		t.Errorf("Ctrl+Delete from (0,0): Text() = %q, want %q", got, "world")
	}
}

// TestMemoCtrlDeleteWithSelectionDeletesSelection verifies that when a
// selection exists, Ctrl+Delete deletes the selection instead.
// Spec: "If a selection exists, deletes the selection instead (same as regular Delete with selection)"
func TestMemoCtrlDeleteWithSelectionDeletesSelection(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")
	// Select "hello" using Shift+Right from col 0
	for i := 0; i < 5; i++ {
		m.HandleEvent(wordShiftKeyEv(tcell.KeyRight))
	}
	if !m.HasSelection() {
		t.Fatal("precondition failed: expected selection before Ctrl+Delete")
	}

	m.HandleEvent(wordCtrlKeyEv(tcell.KeyDelete))

	if got := m.Text(); got != " world" {
		t.Errorf("Ctrl+Delete with \"hello\" selected: Text() = %q, want %q", got, " world")
	}
}

// TestMemoCtrlDeleteWithSelectionDoesNotDeleteAdditionalWord verifies that
// when a selection exists, Ctrl+Delete deletes only the selection.
// Spec: "If a selection exists, deletes the selection instead"
func TestMemoCtrlDeleteWithSelectionDoesNotDeleteAdditionalWord(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")
	// Select just the single char "h" at col 0
	m.HandleEvent(wordShiftKeyEv(tcell.KeyRight))
	if !m.HasSelection() {
		t.Fatal("precondition failed: expected selection before Ctrl+Delete")
	}

	m.HandleEvent(wordCtrlKeyEv(tcell.KeyDelete))

	// Only "h" (1 char) should be removed, leaving "ello world"
	if got := m.Text(); got != "ello world" {
		t.Errorf("Ctrl+Delete with single-char selection: Text() = %q, want %q", got, "ello world")
	}
}

// TestMemoCtrlDeleteAtDocumentEnd verifies Ctrl+Delete at the document end is a no-op.
// Spec: Ctrl+Right at document end has no movement, so deletion range is zero-width.
func TestMemoCtrlDeleteAtDocumentEnd(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")
	wordMoveTo(m, 0, 11) // end of document

	m.HandleEvent(wordCtrlKeyEv(tcell.KeyDelete))

	if got := m.Text(); got != "hello world" {
		t.Errorf("Ctrl+Delete at document end: Text() = %q, want %q (no change)", got, "hello world")
	}
}

// TestMemoCtrlDeleteEventConsumed verifies Ctrl+Delete clears the event.
// Spec: "All word operations consume the event (event.Clear())"
func TestMemoCtrlDeleteEventConsumed(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")

	ev := wordCtrlKeyEv(tcell.KeyDelete)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Ctrl+Delete: event was not consumed (IsCleared() = false)")
	}
}

// ---------------------------------------------------------------------------
// Section 5 — Shift+Ctrl+Left (extend selection to previous word)
// ---------------------------------------------------------------------------

// TestMemoShiftCtrlLeftExtendsSelectionToPreviousWordStart verifies
// Shift+Ctrl+Left from a non-zero cursor position creates a selection
// from the previous word start to the cursor's original position.
// Spec: "Extend the selection by word (same word movement logic, but extends selection instead of collapsing)"
func TestMemoShiftCtrlLeftExtendsSelectionToPreviousWordStart(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")
	// cursor at col 11 (end of "world")
	wordMoveTo(m, 0, 11)

	m.HandleEvent(wordShiftCtrlKeyEv(tcell.KeyLeft))

	row, col := m.CursorPos()
	if row != 0 || col != 6 {
		t.Errorf("Shift+Ctrl+Left from (0,11): CursorPos() = (%d,%d), want (0,6)", row, col)
	}
	sr, sc, er, ec := m.Selection()
	// Anchor should be at (0,11), extent at (0,6)
	if sr != 0 || sc != 11 {
		t.Errorf("Shift+Ctrl+Left from (0,11): anchor = (%d,%d), want (0,11)", sr, sc)
	}
	if er != 0 || ec != 6 {
		t.Errorf("Shift+Ctrl+Left from (0,11): extent = (%d,%d), want (0,6)", er, ec)
	}
}

// TestMemoShiftCtrlLeftExtendsExistingSelection verifies Shift+Ctrl+Left
// extends an existing selection further left.
// Spec: "extends selection instead of collapsing"
func TestMemoShiftCtrlLeftExtendsExistingSelection(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")
	wordMoveTo(m, 0, 11)

	// First Shift+Ctrl+Left: anchor=(0,11), extent=(0,6)
	m.HandleEvent(wordShiftCtrlKeyEv(tcell.KeyLeft))
	// Second Shift+Ctrl+Left: anchor stays at (0,11), extent moves to (0,0)
	m.HandleEvent(wordShiftCtrlKeyEv(tcell.KeyLeft))

	row, col := m.CursorPos()
	if row != 0 || col != 0 {
		t.Errorf("After two Shift+Ctrl+Left from (0,11): CursorPos() = (%d,%d), want (0,0)", row, col)
	}
	sr, sc, er, ec := m.Selection()
	if sr != 0 || sc != 11 {
		t.Errorf("After two Shift+Ctrl+Left: anchor = (%d,%d), want (0,11)", sr, sc)
	}
	if er != 0 || ec != 0 {
		t.Errorf("After two Shift+Ctrl+Left: extent = (%d,%d), want (0,0)", er, ec)
	}
}

// TestMemoShiftCtrlLeftAtDocumentStartPreservesSelection verifies
// Shift+Ctrl+Left at (0,0) with an existing selection keeps cursor at (0,0)
// and does not destroy the selection.
// Spec: "At document start (0,0): no movement"
func TestMemoShiftCtrlLeftAtDocumentStartPreservesSelection(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")
	// Build a selection: from (0,0) extend right 3 chars
	m.HandleEvent(wordShiftKeyEv(tcell.KeyRight))
	m.HandleEvent(wordShiftKeyEv(tcell.KeyRight))
	m.HandleEvent(wordShiftKeyEv(tcell.KeyRight))
	// cursor at (0,3), anchor at (0,0); now do Shift+Ctrl+Left repeatedly to reach (0,0)
	// First: Ctrl+Left from (0,3) goes to (0,0), extending selection
	m.HandleEvent(wordShiftCtrlKeyEv(tcell.KeyLeft))
	// Now cursor is at (0,0); do another Shift+Ctrl+Left — no movement expected
	m.HandleEvent(wordShiftCtrlKeyEv(tcell.KeyLeft))

	row, col := m.CursorPos()
	if row != 0 || col != 0 {
		t.Errorf("Shift+Ctrl+Left at (0,0): CursorPos() = (%d,%d), want (0,0)", row, col)
	}
}

// TestMemoShiftCtrlLeftEventConsumed verifies Shift+Ctrl+Left clears the event.
// Spec: "All word operations consume the event (event.Clear())"
func TestMemoShiftCtrlLeftEventConsumed(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")
	wordMoveTo(m, 0, 6)

	ev := wordShiftCtrlKeyEv(tcell.KeyLeft)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Shift+Ctrl+Left: event was not consumed (IsCleared() = false)")
	}
}

// ---------------------------------------------------------------------------
// Section 6 — Shift+Ctrl+Right (extend selection to next word)
// ---------------------------------------------------------------------------

// TestMemoShiftCtrlRightExtendsSelectionToNextWordStart verifies
// Shift+Ctrl+Right from col 0 creates a selection to the start of the next word.
// Spec: "Extend the selection by word (same word movement logic, but extends selection instead of collapsing)"
func TestMemoShiftCtrlRightExtendsSelectionToNextWordStart(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")
	// cursor at (0,0)

	m.HandleEvent(wordShiftCtrlKeyEv(tcell.KeyRight))

	row, col := m.CursorPos()
	if row != 0 || col != 6 {
		t.Errorf("Shift+Ctrl+Right from (0,0): CursorPos() = (%d,%d), want (0,6)", row, col)
	}
	sr, sc, er, ec := m.Selection()
	if sr != 0 || sc != 0 {
		t.Errorf("Shift+Ctrl+Right from (0,0): anchor = (%d,%d), want (0,0)", sr, sc)
	}
	if er != 0 || ec != 6 {
		t.Errorf("Shift+Ctrl+Right from (0,0): extent = (%d,%d), want (0,6)", er, ec)
	}
}

// TestMemoShiftCtrlRightExtendsExistingSelection verifies Shift+Ctrl+Right
// extends an existing selection further right.
// Spec: "extends selection instead of collapsing"
func TestMemoShiftCtrlRightExtendsExistingSelection(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")
	// cursor at (0,0)

	// First Shift+Ctrl+Right: anchor=(0,0), extent=(0,6)
	m.HandleEvent(wordShiftCtrlKeyEv(tcell.KeyRight))
	// Second Shift+Ctrl+Right: anchor stays (0,0), extent moves to (0,11)
	m.HandleEvent(wordShiftCtrlKeyEv(tcell.KeyRight))

	row, col := m.CursorPos()
	if row != 0 || col != 11 {
		t.Errorf("After two Shift+Ctrl+Right from (0,0): CursorPos() = (%d,%d), want (0,11)", row, col)
	}
	sr, sc, er, ec := m.Selection()
	if sr != 0 || sc != 0 {
		t.Errorf("After two Shift+Ctrl+Right: anchor = (%d,%d), want (0,0)", sr, sc)
	}
	if er != 0 || ec != 11 {
		t.Errorf("After two Shift+Ctrl+Right: extent = (%d,%d), want (0,11)", er, ec)
	}
}

// TestMemoShiftCtrlRightAtDocumentEndPreservesSelection verifies
// Shift+Ctrl+Right at document end keeps cursor there.
// Spec: "At document end: no movement"
func TestMemoShiftCtrlRightAtDocumentEndPreservesSelection(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello")
	wordMoveTo(m, 0, 5) // end of document

	// Build a selection first
	m.HandleEvent(wordShiftKeyEv(tcell.KeyLeft))
	if !m.HasSelection() {
		t.Fatal("precondition failed: expected selection before Shift+Ctrl+Right at end")
	}
	// Now move cursor back to end
	m.HandleEvent(wordKeyEv(tcell.KeyEnd))
	// cursor at end; try Shift+Ctrl+Right
	m.HandleEvent(wordShiftCtrlKeyEv(tcell.KeyRight))

	row, col := m.CursorPos()
	if row != 0 || col != 5 {
		t.Errorf("Shift+Ctrl+Right at document end: CursorPos() = (%d,%d), want (0,5)", row, col)
	}
}

// TestMemoShiftCtrlRightEventConsumed verifies Shift+Ctrl+Right clears the event.
// Spec: "All word operations consume the event (event.Clear())"
func TestMemoShiftCtrlRightEventConsumed(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")

	ev := wordShiftCtrlKeyEv(tcell.KeyRight)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Shift+Ctrl+Right: event was not consumed (IsCleared() = false)")
	}
}

// ---------------------------------------------------------------------------
// Section 7 — Falsifying tests
// ---------------------------------------------------------------------------

// TestMemoCtrlLeftDoesNotStopMidWord verifies Ctrl+Left reaches the start of
// the current word, not just any earlier character.
// Falsification: an implementation that moves backward one char at a time would fail this.
func TestMemoCtrlLeftDoesNotStopMidWord(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")
	wordMoveTo(m, 0, 8) // "wor|ld"

	m.HandleEvent(wordCtrlKeyEv(tcell.KeyLeft))

	_, col := m.CursorPos()
	if col == 7 {
		t.Errorf("Ctrl+Left from col 8 stopped at col 7 (one char back); must reach word start (col 6)")
	}
}

// TestMemoCtrlRightDoesNotStopAtEndOfCurrentWord verifies Ctrl+Right lands at
// the START of the NEXT word, not at the end of the current word.
// Falsification: an impl that stops at the end of the current word fails this.
func TestMemoCtrlRightDoesNotStopAtEndOfCurrentWord(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")
	// cursor at col 0

	m.HandleEvent(wordCtrlKeyEv(tcell.KeyRight))

	_, col := m.CursorPos()
	// col 5 would be end of "hello" — must go to col 6 (start of "world")
	if col == 5 {
		t.Errorf("Ctrl+Right from col 0 stopped at col 5 (end of \"hello\"); must reach start of \"world\" (col 6)")
	}
}

// TestMemoCtrlBackspaceDoesNotDeleteJustOneChar verifies Ctrl+Backspace deletes
// more than one character (whole word segment).
// Falsification: an implementation that maps Ctrl+Backspace to plain Backspace fails.
func TestMemoCtrlBackspaceDoesNotDeleteJustOneChar(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")
	wordMoveTo(m, 0, 11)
	original := m.Text()
	_ = original

	m.HandleEvent(wordCtrlKeyEv(tcell.KeyBackspace2))

	got := m.Text()
	// Plain Backspace would leave "hello worl" (deleted 'd' only)
	if got == "hello worl" {
		t.Errorf("Ctrl+Backspace deleted only one char (result: %q); expected word-level deletion", got)
	}
}

// TestMemoCtrlDeleteDoesNotDeleteJustOneChar verifies Ctrl+Delete deletes
// more than one character (whole word segment).
// Falsification: an implementation that maps Ctrl+Delete to plain Delete fails.
func TestMemoCtrlDeleteDoesNotDeleteJustOneChar(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")
	// cursor at (0,0)

	m.HandleEvent(wordCtrlKeyEv(tcell.KeyDelete))

	got := m.Text()
	// Plain Delete would leave "ello world" (deleted 'h' only)
	if got == "ello world" {
		t.Errorf("Ctrl+Delete deleted only one char (result: %q); expected word-level deletion", got)
	}
}

// TestMemoShiftCtrlLeftCreatesSelection verifies Shift+Ctrl+Left creates a
// selection — it does not merely move the cursor.
// Falsification: an implementation that behaves like plain Ctrl+Left fails.
func TestMemoShiftCtrlLeftCreatesSelection(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")
	wordMoveTo(m, 0, 11)

	m.HandleEvent(wordShiftCtrlKeyEv(tcell.KeyLeft))

	if !m.HasSelection() {
		t.Error("Shift+Ctrl+Left: HasSelection() = false; expected a non-empty selection")
	}
}

// TestMemoShiftCtrlRightCreatesSelection verifies Shift+Ctrl+Right creates a
// selection — it does not merely move the cursor.
// Falsification: an implementation that behaves like plain Ctrl+Right fails.
func TestMemoShiftCtrlRightCreatesSelection(t *testing.T) {
	m := newWordMemo()
	m.SetText("hello world")
	// cursor at (0,0)

	m.HandleEvent(wordShiftCtrlKeyEv(tcell.KeyRight))

	if !m.HasSelection() {
		t.Error("Shift+Ctrl+Right: HasSelection() = false; expected a non-empty selection")
	}
}
