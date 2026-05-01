package tv

// memo_clipboard_test.go — Tests for Task 3: Clipboard Operations and
// Selection-Aware Editing.
//
// Written BEFORE any implementation exists; all tests drive the spec.
// Each test cites the exact spec sentence it verifies.
//
// Test organisation:
//   Section 1  — Ctrl+C (copy)
//   Section 2  — Ctrl+X (cut)
//   Section 3  — Ctrl+V (paste)
//   Section 4  — Ctrl+V with newlines
//   Section 5  — Paste post-conditions (cursor, selection)
//   Section 6  — Event consumption (Ctrl+C, Ctrl+X, Ctrl+V)
//   Section 7  — Selection-aware character typing
//   Section 8  — Selection-aware Enter
//   Section 9  — Selection-aware Backspace
//   Section 10 — Selection-aware Delete
//   Section 11 — Ctrl+Y (NOT selection-aware)
//   Section 12 — Falsifying tests

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// clipKeyEv creates a plain keyboard event with no modifiers.
func clipKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key}}
}

// clipRuneEv creates a keyboard event for a printable rune.
func clipRuneEv(r rune) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: r}}
}

// clipShiftKeyEv creates a keyboard event with Shift held.
func clipShiftKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key, Modifiers: tcell.ModShift}}
}

// clipCtrlKeyEv creates a keyboard event with Ctrl held.
func clipCtrlKeyEv(key tcell.Key) *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: key, Modifiers: tcell.ModCtrl}}
}

// newClipMemo creates a Memo with generous bounds for clipboard/edit tests.
func newClipMemo() *Memo {
	return NewMemo(NewRect(0, 0, 60, 20))
}

// clipRepeatKey sends the same plain-key event n times to a Memo.
func clipRepeatKey(m *Memo, key tcell.Key, n int) {
	for i := 0; i < n; i++ {
		m.HandleEvent(clipKeyEv(key))
	}
}

// selectChars selects n characters to the right from the current cursor using
// Shift+Right.
func selectChars(m *Memo, n int) {
	for i := 0; i < n; i++ {
		m.HandleEvent(clipShiftKeyEv(tcell.KeyRight))
	}
}

// ---------------------------------------------------------------------------
// Section 1 — Ctrl+C (copy)
// ---------------------------------------------------------------------------

// TestMemoCtrlCCopiesSelectionToClipboard verifies that Ctrl+C copies selected
// text to the package-level clipboard variable.
// Spec: "Ctrl+C copies the selected text to the package-level clipboard variable"
func TestMemoCtrlCCopiesSelectionToClipboard(t *testing.T) {
	clipboard = ""
	m := newClipMemo()
	m.SetText("hello world")
	selectChars(m, 5) // select "hello"

	m.HandleEvent(clipKeyEv(tcell.KeyCtrlC))

	if clipboard != "hello" {
		t.Errorf("clipboard = %q after Ctrl+C, want %q", clipboard, "hello")
	}
}

// TestCtrlCDoesNotModifyBuffer verifies that Ctrl+C leaves the buffer unchanged.
// Spec: Ctrl+C copies — it must not delete or alter the text.
// Falsifying: an impl that also deletes the selection (like cut) would fail.
func TestCtrlCDoesNotModifyBuffer(t *testing.T) {
	clipboard = ""
	m := newClipMemo()
	m.SetText("hello world")
	selectChars(m, 5) // select "hello"

	m.HandleEvent(clipKeyEv(tcell.KeyCtrlC))

	if got := m.Text(); got != "hello world" {
		t.Errorf("Ctrl+C modified buffer: Text() = %q, want %q", got, "hello world")
	}
}

// TestCtrlCWithNoSelectionDoesNotChangeClipboard verifies that Ctrl+C without a
// selection does not clear or modify the clipboard.
// Spec: "Ctrl+C with no selection does nothing (does not clear clipboard)"
func TestCtrlCWithNoSelectionDoesNotChangeClipboard(t *testing.T) {
	clipboard = "preserved"
	m := newClipMemo()
	m.SetText("hello world")
	// No selection made.

	m.HandleEvent(clipKeyEv(tcell.KeyCtrlC))

	if clipboard != "preserved" {
		t.Errorf("clipboard = %q after Ctrl+C with no selection, want %q (unchanged)", clipboard, "preserved")
	}
}

// TestCtrlCConsumesEvent verifies Ctrl+C consumes the event.
// Spec: "Ctrl+C, Ctrl+X, Ctrl+V all consume the event (event.Clear())"
func TestCtrlCConsumesEvent(t *testing.T) {
	clipboard = ""
	m := newClipMemo()
	m.SetText("hello")
	selectChars(m, 3)

	ev := clipKeyEv(tcell.KeyCtrlC)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Ctrl+C did not consume the event (IsCleared() = false), want true")
	}
}

// TestCtrlCConsumesEventEvenWithNoSelection verifies Ctrl+C consumes the event
// regardless of whether a selection exists.
// Spec: "Ctrl+C, Ctrl+X, Ctrl+V all consume the event (event.Clear())"
func TestCtrlCConsumesEventEvenWithNoSelection(t *testing.T) {
	clipboard = ""
	m := newClipMemo()
	m.SetText("hello")
	// No selection.

	ev := clipKeyEv(tcell.KeyCtrlC)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Ctrl+C with no selection did not consume the event, want IsCleared() = true")
	}
}

// ---------------------------------------------------------------------------
// Section 2 — Ctrl+X (cut)
// ---------------------------------------------------------------------------

// TestMemoCtrlXCutsSelectionToClipboard verifies Ctrl+X copies selected text to
// the clipboard.
// Spec: "Ctrl+X cuts: copies selection to clipboard, then deletes the selection"
func TestMemoCtrlXCutsSelectionToClipboard(t *testing.T) {
	clipboard = ""
	m := newClipMemo()
	m.SetText("hello world")
	selectChars(m, 5) // select "hello"

	m.HandleEvent(clipKeyEv(tcell.KeyCtrlX))

	if clipboard != "hello" {
		t.Errorf("clipboard = %q after Ctrl+X, want %q", clipboard, "hello")
	}
}

// TestCtrlXRemovesSelectionFromBuffer verifies Ctrl+X deletes the selected text
// from the buffer.
// Spec: "Ctrl+X cuts: copies selection to clipboard, then deletes the selection"
// Falsifying: an impl that only copies (no delete) would fail.
func TestCtrlXRemovesSelectionFromBuffer(t *testing.T) {
	clipboard = ""
	m := newClipMemo()
	m.SetText("hello world")
	selectChars(m, 5) // select "hello"

	m.HandleEvent(clipKeyEv(tcell.KeyCtrlX))

	if got := m.Text(); got != " world" {
		t.Errorf("Ctrl+X left buffer as %q, want %q", got, " world")
	}
}

// TestCtrlXCursorAtStartOfFormerSelection verifies cursor lands at the start of
// the former selection after Ctrl+X.
// Spec: "After deleting a selection, cursor is at the start (earlier position)
//
//	of the former selection"
func TestCtrlXCursorAtStartOfFormerSelection(t *testing.T) {
	clipboard = ""
	m := newClipMemo()
	m.SetText("hello world")
	clipRepeatKey(m, tcell.KeyRight, 6) // cursor at col 6
	selectChars(m, 5)                   // select "world" (cols 6–11)

	m.HandleEvent(clipKeyEv(tcell.KeyCtrlX))

	row, col := m.CursorPos()
	if row != 0 || col != 6 {
		t.Errorf("After Ctrl+X cursor at (%d,%d), want (0,6)", row, col)
	}
}

// TestCtrlXWithNoSelectionDoesNothing verifies Ctrl+X with no selection leaves
// clipboard and buffer unchanged.
// Spec: "Ctrl+X with no selection does nothing"
func TestCtrlXWithNoSelectionDoesNothing(t *testing.T) {
	clipboard = "safe"
	m := newClipMemo()
	m.SetText("hello")
	// No selection.

	m.HandleEvent(clipKeyEv(tcell.KeyCtrlX))

	if clipboard != "safe" {
		t.Errorf("clipboard = %q after Ctrl+X with no selection, want %q (unchanged)", clipboard, "safe")
	}
	if got := m.Text(); got != "hello" {
		t.Errorf("Ctrl+X with no selection modified buffer: Text() = %q, want %q", got, "hello")
	}
}

// TestCtrlXConsumesEvent verifies Ctrl+X consumes the event.
// Spec: "Ctrl+C, Ctrl+X, Ctrl+V all consume the event (event.Clear())"
func TestCtrlXConsumesEvent(t *testing.T) {
	clipboard = ""
	m := newClipMemo()
	m.SetText("hello")
	selectChars(m, 3)

	ev := clipKeyEv(tcell.KeyCtrlX)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Ctrl+X did not consume the event (IsCleared() = false), want true")
	}
}

// TestCtrlXConsumesEventEvenWithNoSelection verifies Ctrl+X consumes the event
// even when there is no selection.
// Spec: "Ctrl+C, Ctrl+X, Ctrl+V all consume the event (event.Clear())"
func TestCtrlXConsumesEventEvenWithNoSelection(t *testing.T) {
	clipboard = ""
	m := newClipMemo()
	m.SetText("hello")
	// No selection.

	ev := clipKeyEv(tcell.KeyCtrlX)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Ctrl+X with no selection did not consume the event, want IsCleared() = true")
	}
}

// ---------------------------------------------------------------------------
// Section 3 — Ctrl+V (paste)
// ---------------------------------------------------------------------------

// TestCtrlVPastesClipboardAtCursor verifies Ctrl+V inserts clipboard content at
// the cursor when no selection is active.
// Spec: "Ctrl+V pastes: if a selection exists, it is replaced by the clipboard
//
//	content; otherwise clipboard content is inserted at cursor"
func TestCtrlVPastesClipboardAtCursor(t *testing.T) {
	clipboard = "hello"
	m := newClipMemo()
	m.SetText("")
	// Cursor at (0,0), no selection.

	m.HandleEvent(clipKeyEv(tcell.KeyCtrlV))

	if got := m.Text(); got != "hello" {
		t.Errorf("Ctrl+V paste at empty cursor: Text() = %q, want %q", got, "hello")
	}
}

// TestCtrlVPastesAtMidLineCursor verifies Ctrl+V inserts clipboard at a
// non-zero cursor position.
// Spec: "otherwise clipboard content is inserted at cursor"
func TestCtrlVPastesAtMidLineCursor(t *testing.T) {
	clipboard = "XY"
	m := newClipMemo()
	m.SetText("ab cd")
	clipRepeatKey(m, tcell.KeyRight, 2) // cursor at col 2

	m.HandleEvent(clipKeyEv(tcell.KeyCtrlV))

	if got := m.Text(); got != "abXY cd" {
		t.Errorf("Ctrl+V at col 2: Text() = %q, want %q", got, "abXY cd")
	}
}

// TestCtrlVReplacesSelectionWithClipboard verifies Ctrl+V replaces the selection
// with clipboard content when a selection exists.
// Spec: "Ctrl+V pastes: if a selection exists, it is replaced by the clipboard content"
func TestCtrlVReplacesSelectionWithClipboard(t *testing.T) {
	clipboard = "NEW"
	m := newClipMemo()
	m.SetText("hello world")
	selectChars(m, 5) // select "hello"

	m.HandleEvent(clipKeyEv(tcell.KeyCtrlV))

	if got := m.Text(); got != "NEW world" {
		t.Errorf("Ctrl+V replace selection: Text() = %q, want %q", got, "NEW world")
	}
}

// TestCtrlVWithEmptyClipboardDoesNothing verifies Ctrl+V with an empty clipboard
// has no effect.
// Spec: "Ctrl+V with empty clipboard does nothing"
func TestCtrlVWithEmptyClipboardDoesNothing(t *testing.T) {
	clipboard = ""
	m := newClipMemo()
	m.SetText("hello")
	clipRepeatKey(m, tcell.KeyRight, 3) // cursor at col 3

	m.HandleEvent(clipKeyEv(tcell.KeyCtrlV))

	if got := m.Text(); got != "hello" {
		t.Errorf("Ctrl+V with empty clipboard changed buffer: Text() = %q, want %q", got, "hello")
	}
}

// TestCtrlVConsumesEvent verifies Ctrl+V consumes the event.
// Spec: "Ctrl+C, Ctrl+X, Ctrl+V all consume the event (event.Clear())"
func TestCtrlVConsumesEvent(t *testing.T) {
	clipboard = "data"
	m := newClipMemo()
	m.SetText("hello")

	ev := clipKeyEv(tcell.KeyCtrlV)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Ctrl+V did not consume the event (IsCleared() = false), want true")
	}
}

// TestCtrlVConsumesEventEvenWithEmptyClipboard verifies Ctrl+V consumes the
// event even when clipboard is empty (no-op case).
// Spec: "Ctrl+C, Ctrl+X, Ctrl+V all consume the event (event.Clear())"
func TestCtrlVConsumesEventEvenWithEmptyClipboard(t *testing.T) {
	clipboard = ""
	m := newClipMemo()
	m.SetText("hello")

	ev := clipKeyEv(tcell.KeyCtrlV)
	m.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Error("Ctrl+V with empty clipboard did not consume the event, want IsCleared() = true")
	}
}

// ---------------------------------------------------------------------------
// Section 4 — Ctrl+V with newlines
// ---------------------------------------------------------------------------

// TestCtrlVWithNewlineSplitsIntoMultipleLines verifies that pasting text
// containing newlines creates multiple lines in the buffer.
// Spec: "Pasted text may contain newlines — these are split into multiple lines
//
//	in the buffer"
func TestCtrlVWithNewlineSplitsIntoMultipleLines(t *testing.T) {
	clipboard = "foo\nbar"
	m := newClipMemo()
	m.SetText("")

	m.HandleEvent(clipKeyEv(tcell.KeyCtrlV))

	lines := strings.Split(m.Text(), "\n")
	if len(lines) != 2 {
		t.Errorf("Ctrl+V with newline: got %d lines, want 2: %v", len(lines), lines)
	}
	if lines[0] != "foo" {
		t.Errorf("Ctrl+V with newline: line 0 = %q, want %q", lines[0], "foo")
	}
	if lines[1] != "bar" {
		t.Errorf("Ctrl+V with newline: line 1 = %q, want %q", lines[1], "bar")
	}
}

// TestCtrlVWithMultipleNewlines verifies pasting content with multiple newlines
// results in the correct number of lines.
// Spec: "Pasted text may contain newlines — these are split into multiple lines"
func TestCtrlVWithMultipleNewlines(t *testing.T) {
	clipboard = "a\nb\nc"
	m := newClipMemo()
	m.SetText("")

	m.HandleEvent(clipKeyEv(tcell.KeyCtrlV))

	lines := strings.Split(m.Text(), "\n")
	if len(lines) != 3 {
		t.Errorf("Ctrl+V with 2 newlines: got %d lines, want 3: %v", len(lines), lines)
	}
}

// ---------------------------------------------------------------------------
// Section 5 — Paste post-conditions (cursor and selection)
// ---------------------------------------------------------------------------

// TestCtrlVCursorAtEndOfPastedText verifies that after paste the cursor is at
// the end of the pasted text.
// Spec: "After paste, cursor is at the end of the pasted text"
func TestCtrlVCursorAtEndOfPastedText(t *testing.T) {
	clipboard = "hello"
	m := newClipMemo()
	m.SetText("")

	m.HandleEvent(clipKeyEv(tcell.KeyCtrlV))

	row, col := m.CursorPos()
	// "hello" is 5 chars on one line; cursor should be at (0,5).
	if row != 0 || col != 5 {
		t.Errorf("After Ctrl+V, CursorPos() = (%d,%d), want (0,5)", row, col)
	}
}

// TestCtrlVCursorAtEndOfMultiLinePaste verifies cursor position after pasting
// text that spans multiple lines.
// Spec: "After paste, cursor is at the end of the pasted text"
func TestCtrlVCursorAtEndOfMultiLinePaste(t *testing.T) {
	clipboard = "foo\nbar"
	m := newClipMemo()
	m.SetText("")

	m.HandleEvent(clipKeyEv(tcell.KeyCtrlV))

	row, col := m.CursorPos()
	// "bar" ends at col 3 on row 1.
	if row != 1 || col != 3 {
		t.Errorf("After Ctrl+V multi-line, CursorPos() = (%d,%d), want (1,3)", row, col)
	}
}

// TestCtrlVClearsSelection verifies that after paste the selection is cleared
// (collapsed at the cursor).
// Spec: "After paste, selection is cleared (collapsed at cursor)"
func TestCtrlVClearsSelection(t *testing.T) {
	clipboard = "NEW"
	m := newClipMemo()
	m.SetText("hello world")
	selectChars(m, 5) // select "hello" — a selection exists before paste

	m.HandleEvent(clipKeyEv(tcell.KeyCtrlV))

	if m.HasSelection() {
		t.Error("HasSelection() = true after Ctrl+V, want false (selection should be cleared)")
	}
}

// TestCtrlVClearsSelectionOnSimplePaste verifies selection is cleared even when
// pasting into a collapsed cursor (no prior selection).
// Spec: "After paste, selection is cleared (collapsed at cursor)"
func TestCtrlVClearsSelectionOnSimplePaste(t *testing.T) {
	clipboard = "hi"
	m := newClipMemo()
	m.SetText("hello")
	// No selection — just verifying that paste never accidentally leaves a selection.

	m.HandleEvent(clipKeyEv(tcell.KeyCtrlV))

	if m.HasSelection() {
		t.Error("HasSelection() = true after Ctrl+V with no prior selection, want false")
	}
}

// ---------------------------------------------------------------------------
// Section 7 — Selection-aware character typing
// ---------------------------------------------------------------------------

// TestTypingCharacterReplacesSelection verifies that typing a printable character
// while a selection exists replaces the selection with the typed character.
// Spec: "Typing a printable character while a selection exists replaces the
//
//	selection with the typed character"
func TestTypingCharacterReplacesSelection(t *testing.T) {
	m := newClipMemo()
	m.SetText("hello world")
	selectChars(m, 5) // select "hello"

	m.HandleEvent(clipRuneEv('X'))

	if got := m.Text(); got != "X world" {
		t.Errorf("Type 'X' over selection: Text() = %q, want %q", got, "X world")
	}
}

// TestTypingCharacterOverSelectionDoesNotJustInsert verifies that typing over a
// selection removes the selected text, not just inserts ahead of it.
// Spec: "replaces the selection with the typed character"
// Falsifying: an impl that inserts without deleting the selection would fail.
func TestTypingCharacterOverSelectionDoesNotJustInsert(t *testing.T) {
	m := newClipMemo()
	m.SetText("hello")
	selectChars(m, 5) // select all "hello"

	m.HandleEvent(clipRuneEv('Z'))

	// The whole "hello" should be gone, replaced by just "Z".
	if got := m.Text(); got != "Z" {
		t.Errorf("Type 'Z' over full selection: Text() = %q, want %q", got, "Z")
	}
}

// TestTypingCharacterOverSelectionCursorAfterInsertedChar verifies cursor is
// placed after the inserted character.
// Spec: implied by "replaces the selection with the typed character" — cursor
// follows normal insert semantics (advance by 1).
func TestTypingCharacterOverSelectionCursorAfterInsertedChar(t *testing.T) {
	m := newClipMemo()
	m.SetText("hello world")
	selectChars(m, 5) // select "hello"

	m.HandleEvent(clipRuneEv('X'))

	row, col := m.CursorPos()
	// After replacing "hello" with "X", cursor should be at (0,1).
	if row != 0 || col != 1 {
		t.Errorf("After typing 'X' over selection, CursorPos() = (%d,%d), want (0,1)", row, col)
	}
}

// ---------------------------------------------------------------------------
// Section 8 — Selection-aware Enter
// ---------------------------------------------------------------------------

// TestEnterReplacesSelectionWithNewline verifies pressing Enter while a
// selection exists replaces the selection with a newline.
// Spec: "Enter while a selection exists replaces the selection with a newline
//
//	(with auto-indent from the line where the selection started)"
func TestEnterReplacesSelectionWithNewline(t *testing.T) {
	m := newClipMemo()
	m.SetText("hello world")
	selectChars(m, 5) // select "hello"

	m.HandleEvent(clipKeyEv(tcell.KeyEnter))

	lines := strings.Split(m.Text(), "\n")
	if len(lines) != 2 {
		t.Fatalf("Enter over selection: got %d lines, want 2; text=%q", len(lines), m.Text())
	}
	// Line 0 is empty (the selection "hello" was replaced by a newline at position 0)
	// and line 1 contains " world".
	if lines[1] != " world" {
		t.Errorf("Enter over selection: line 1 = %q, want %q", lines[1], " world")
	}
}

// TestEnterOverSelectionClearsSelection verifies that after Enter over a
// selection the selection is cleared.
// Spec: implied — replacing selection leaves cursor at new position with no
// active selection.
func TestEnterOverSelectionClearsSelection(t *testing.T) {
	m := newClipMemo()
	m.SetText("hello world")
	selectChars(m, 5)

	m.HandleEvent(clipKeyEv(tcell.KeyEnter))

	if m.HasSelection() {
		t.Error("HasSelection() = true after Enter over selection, want false")
	}
}

// TestEnterOverSelectionAutoIndentsFromStartLine verifies that Enter with a
// selection auto-indents the new line from the selection start line's leading
// whitespace, not the cursor's current line.
// Spec: "Enter while a selection exists replaces the selection with a newline
// (with auto-indent from the line where the selection started)"
func TestEnterOverSelectionAutoIndentsFromStartLine(t *testing.T) {
	m := newClipMemo()
	m.SetText("    hello world")
	// Select "hello" (cols 4-8): move to col 4, then Shift+Right 5 times
	clipRepeatKey(m, tcell.KeyRight, 4)
	selectChars(m, 5)

	m.HandleEvent(clipKeyEv(tcell.KeyEnter))

	text := m.Text()
	lines := strings.Split(text, "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2", len(lines))
	}
	if lines[0] != "    " {
		t.Errorf("line 0 = %q, want %q", lines[0], "    ")
	}
	// New line should have auto-indent (4 spaces) + remaining text
	if lines[1] != "     world" {
		t.Errorf("line 1 = %q, want %q (4-space auto-indent + ' world')", lines[1], "     world")
	}
}

// ---------------------------------------------------------------------------
// Section 9 — Selection-aware Backspace
// ---------------------------------------------------------------------------

// TestBackspaceWithSelectionDeletesSelection verifies Backspace with a selection
// deletes only the selected text.
// Spec: "Backspace while a selection exists deletes the selection (does not
//
//	delete an additional character)"
func TestBackspaceWithSelectionDeletesSelection(t *testing.T) {
	m := newClipMemo()
	m.SetText("hello world")
	clipRepeatKey(m, tcell.KeyRight, 6) // cursor at col 6
	selectChars(m, 5)                   // select "world"

	m.HandleEvent(clipKeyEv(tcell.KeyBackspace2))

	if got := m.Text(); got != "hello " {
		t.Errorf("Backspace over selection: Text() = %q, want %q", got, "hello ")
	}
}

// TestBackspaceWithSelectionDoesNotDeleteAdditionalChar verifies that Backspace
// with a selection does not delete any character beyond the selection.
// Spec: "Backspace while a selection exists deletes the selection (does not
//
//	delete an additional character)"
// Falsifying: an impl that deletes selection AND one more char would fail.
func TestBackspaceWithSelectionDoesNotDeleteAdditionalChar(t *testing.T) {
	m := newClipMemo()
	m.SetText("abcde")
	clipRepeatKey(m, tcell.KeyRight, 1) // cursor at col 1
	selectChars(m, 3)                   // select "bcd"

	m.HandleEvent(clipKeyEv(tcell.KeyBackspace2))

	// Only "bcd" should be deleted, leaving "ae".
	if got := m.Text(); got != "ae" {
		t.Errorf("Backspace over selection: Text() = %q, want %q (no extra char deleted)", got, "ae")
	}
}

// TestBackspaceWithSelectionCursorAtStartOfFormerSelection verifies cursor is
// at the start (earlier position) of the former selection after deleting.
// Spec: "After deleting a selection, cursor is at the start (earlier position)
//
//	of the former selection"
func TestBackspaceWithSelectionCursorAtStartOfFormerSelection(t *testing.T) {
	m := newClipMemo()
	m.SetText("hello world")
	clipRepeatKey(m, tcell.KeyRight, 6) // cursor at col 6
	selectChars(m, 5)                   // select "world" (cols 6–11)

	m.HandleEvent(clipKeyEv(tcell.KeyBackspace2))

	row, col := m.CursorPos()
	if row != 0 || col != 6 {
		t.Errorf("After Backspace over selection, CursorPos() = (%d,%d), want (0,6)", row, col)
	}
}

// TestBackspaceWithSelectionClearsSelection verifies that after deleting a
// selection via Backspace, HasSelection() returns false.
// Spec: "After deleting a selection, cursor is at the start (earlier position)
//
//	of the former selection" — selection is gone.
func TestBackspaceWithSelectionClearsSelection(t *testing.T) {
	m := newClipMemo()
	m.SetText("hello world")
	selectChars(m, 5)

	m.HandleEvent(clipKeyEv(tcell.KeyBackspace2))

	if m.HasSelection() {
		t.Error("HasSelection() = true after Backspace over selection, want false")
	}
}

// ---------------------------------------------------------------------------
// Section 10 — Selection-aware Delete
// ---------------------------------------------------------------------------

// TestDeleteWithSelectionDeletesSelection verifies Delete with a selection
// deletes only the selected text.
// Spec: "Delete while a selection exists deletes the selection (does not delete
//
//	an additional character)"
func TestDeleteWithSelectionDeletesSelection(t *testing.T) {
	m := newClipMemo()
	m.SetText("hello world")
	clipRepeatKey(m, tcell.KeyRight, 6) // cursor at col 6
	selectChars(m, 5)                   // select "world"

	m.HandleEvent(clipKeyEv(tcell.KeyDelete))

	if got := m.Text(); got != "hello " {
		t.Errorf("Delete over selection: Text() = %q, want %q", got, "hello ")
	}
}

// TestDeleteWithSelectionDoesNotDeleteAdditionalChar verifies that Delete with a
// selection does not delete any character beyond the selection.
// Spec: "Delete while a selection exists deletes the selection (does not delete
//
//	an additional character)"
// Falsifying: an impl that deletes selection AND one more char would fail.
func TestDeleteWithSelectionDoesNotDeleteAdditionalChar(t *testing.T) {
	m := newClipMemo()
	m.SetText("abcde")
	clipRepeatKey(m, tcell.KeyRight, 1) // cursor at col 1
	selectChars(m, 3)                   // select "bcd"

	m.HandleEvent(clipKeyEv(tcell.KeyDelete))

	// Only "bcd" deleted, leaving "ae".
	if got := m.Text(); got != "ae" {
		t.Errorf("Delete over selection: Text() = %q, want %q (no extra char deleted)", got, "ae")
	}
}

// TestDeleteWithSelectionCursorAtStartOfFormerSelection verifies cursor is at
// the start (earlier position) of the former selection after deleting.
// Spec: "After deleting a selection, cursor is at the start (earlier position)
//
//	of the former selection"
func TestDeleteWithSelectionCursorAtStartOfFormerSelection(t *testing.T) {
	m := newClipMemo()
	m.SetText("hello world")
	clipRepeatKey(m, tcell.KeyRight, 6) // cursor at col 6
	selectChars(m, 5)                   // select "world" (cols 6–11)

	m.HandleEvent(clipKeyEv(tcell.KeyDelete))

	row, col := m.CursorPos()
	if row != 0 || col != 6 {
		t.Errorf("After Delete over selection, CursorPos() = (%d,%d), want (0,6)", row, col)
	}
}

// TestDeleteWithSelectionClearsSelection verifies that after deleting a
// selection via Delete, HasSelection() returns false.
// Spec: "After deleting a selection, cursor is at the start (earlier position)"
// — selection is collapsed.
func TestDeleteWithSelectionClearsSelection(t *testing.T) {
	m := newClipMemo()
	m.SetText("hello world")
	selectChars(m, 5)

	m.HandleEvent(clipKeyEv(tcell.KeyDelete))

	if m.HasSelection() {
		t.Error("HasSelection() = true after Delete over selection, want false")
	}
}

// ---------------------------------------------------------------------------
// Section 11 — Ctrl+Y (NOT selection-aware)
// ---------------------------------------------------------------------------

// TestCtrlYWithSelectionDeletesCurrentLine verifies that Ctrl+Y with a selection
// still deletes the entire current line (not just the selection).
// Spec: "Ctrl+Y (delete line) is NOT selection-aware — it always deletes the
//
//	current line regardless of selection"
func TestCtrlYWithSelectionDeletesCurrentLine(t *testing.T) {
	m := newClipMemo()
	m.SetText("first\nsecond\nthird")
	m.HandleEvent(clipKeyEv(tcell.KeyDown)) // move to row 1 "second"
	selectChars(m, 3)                       // select "sec" (partial selection on row 1)

	m.HandleEvent(clipKeyEv(tcell.KeyCtrlY))

	// Entire "second" line deleted, not just "sec".
	if got := m.Text(); got != "first\nthird" {
		t.Errorf("Ctrl+Y with selection: Text() = %q, want %q", got, "first\nthird")
	}
}

// TestCtrlYWithSelectionClearsSelection verifies that Ctrl+Y clears the
// selection regardless of what it deletes.
// Spec: "Ctrl+Y (delete line) is NOT selection-aware ... and clears the selection"
func TestCtrlYWithSelectionClearsSelection(t *testing.T) {
	m := newClipMemo()
	m.SetText("first\nsecond\nthird")
	m.HandleEvent(clipKeyEv(tcell.KeyDown)) // row 1
	selectChars(m, 3)                       // partial selection on row 1
	if !m.HasSelection() {
		t.Fatal("precondition: HasSelection() should be true before Ctrl+Y")
	}

	m.HandleEvent(clipKeyEv(tcell.KeyCtrlY))

	if m.HasSelection() {
		t.Error("HasSelection() = true after Ctrl+Y, want false (Ctrl+Y clears selection)")
	}
}

// ---------------------------------------------------------------------------
// Section 12 — Falsifying tests (additional)
// ---------------------------------------------------------------------------

// TestCtrlXLeavesCursorAtSelectionStart verifies that after cutting, the cursor
// is at the start of the former selection, not at the end or some other position.
// Falsifying: an impl that places cursor at end of what remains (or doesn't move
// it) would fail.
// Spec: "After deleting a selection, cursor is at the start (earlier position)
//
//	of the former selection"
func TestCtrlXLeavesCursorAtSelectionStart(t *testing.T) {
	clipboard = ""
	m := newClipMemo()
	m.SetText("abcde")
	clipRepeatKey(m, tcell.KeyRight, 1) // cursor at col 1
	selectChars(m, 3)                   // select "bcd" (cols 1–4)

	m.HandleEvent(clipKeyEv(tcell.KeyCtrlX))

	row, col := m.CursorPos()
	if row != 0 || col != 1 {
		t.Errorf("After Ctrl+X, CursorPos() = (%d,%d), want (0,1) (start of former selection)", row, col)
	}
}

// TestCtrlXActuallyRemovesFromBuffer verifies the text is genuinely removed
// (not merely hidden) after Ctrl+X.
// Falsifying: a stub that sets clipboard but doesn't modify the buffer would fail.
// Spec: "Ctrl+X cuts: copies selection to clipboard, then deletes the selection"
func TestCtrlXActuallyRemovesFromBuffer(t *testing.T) {
	clipboard = ""
	m := newClipMemo()
	m.SetText("ABC")
	selectChars(m, 3) // select all

	m.HandleEvent(clipKeyEv(tcell.KeyCtrlX))

	if got := m.Text(); got != "" {
		t.Errorf("After Ctrl+X of entire content, Text() = %q, want empty", got)
	}
}

// TestTypingOverSelectionDoesNotLeaveSelectionResidual verifies that after
// typing over a selection, HasSelection() returns false.
// Spec: implied — the selection is consumed when it is replaced.
func TestTypingOverSelectionDoesNotLeaveSelectionResidual(t *testing.T) {
	m := newClipMemo()
	m.SetText("hello")
	selectChars(m, 5) // select all

	m.HandleEvent(clipRuneEv('A'))

	if m.HasSelection() {
		t.Error("HasSelection() = true after typing over selection, want false")
	}
}

// TestCtrlVReplacesSelectionCompletely verifies that when a selection is
// replaced by Ctrl+V, the previously selected text is fully gone.
// Falsifying: an impl that inserts clipboard before or after the selection
// without removing it would fail.
// Spec: "if a selection exists, it is replaced by the clipboard content"
func TestCtrlVReplacesSelectionCompletely(t *testing.T) {
	clipboard = "Z"
	m := newClipMemo()
	m.SetText("hello")
	selectChars(m, 5) // select all "hello"

	m.HandleEvent(clipKeyEv(tcell.KeyCtrlV))

	if got := m.Text(); got != "Z" {
		t.Errorf("Ctrl+V replacing full selection: Text() = %q, want %q", got, "Z")
	}
}
