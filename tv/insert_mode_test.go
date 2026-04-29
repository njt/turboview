package tv

// Tests for Insert/Overwrite mode toggle and Ctrl+Y (clear text).
//
// Each test cites the spec requirement it verifies.

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// insertKeyEv builds an Insert key event.
func insertKeyEv() *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyInsert}}
}

// ctrlYEv builds a Ctrl+Y keyboard event.
func ctrlYEv() *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlY}}
}

// ---------------------------------------------------------------------------
// Overwrite() getter — initial state
// ---------------------------------------------------------------------------

// TestOverwriteDefaultsFalse verifies a new InputLine is in insert mode (overwrite=false).
// Spec: "overwrite bool field, Overwrite() getter."
func TestOverwriteDefaultsFalse(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	if il.Overwrite() {
		t.Error("Overwrite() = true on new InputLine; expected false (insert mode by default)")
	}
}

// ---------------------------------------------------------------------------
// Insert key toggles overwrite
// ---------------------------------------------------------------------------

// TestInsertKeyTogglesOverwriteOn verifies Insert key switches to overwrite mode.
// Spec: "Insert key toggles overwrite."
func TestInsertKeyTogglesOverwriteOn(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.HandleEvent(insertKeyEv())
	if !il.Overwrite() {
		t.Error("Overwrite() = false after Insert key; expected true (overwrite mode toggled on)")
	}
}

// TestInsertKeyTogglesOverwriteOffAgain verifies a second Insert key switches back to insert mode.
// Spec: "Insert key toggles overwrite."
func TestInsertKeyTogglesOverwriteOffAgain(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.HandleEvent(insertKeyEv())
	il.HandleEvent(insertKeyEv())
	if il.Overwrite() {
		t.Error("Overwrite() = true after two Insert key presses; expected false (toggled back to insert)")
	}
}

// TestInsertKeyEventConsumed verifies the Insert key event is consumed.
// Spec: "Event cleared for both."
func TestInsertKeyEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	ev := insertKeyEv()
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Insert key event was not consumed")
	}
}

// TestInsertKeyEventConsumedOnToggleBack verifies Insert is consumed when toggling back.
func TestInsertKeyEventConsumedOnToggleBack(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.HandleEvent(insertKeyEv()) // first press
	ev := insertKeyEv()           // second press
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("second Insert key event was not consumed")
	}
}

// ---------------------------------------------------------------------------
// Overwrite mode: typing replaces char at cursor
// ---------------------------------------------------------------------------

// TestOverwriteModeReplacesCharAtCursor verifies that in overwrite mode, typing a rune
// replaces the character at cursorPos rather than inserting.
// Spec: "In overwrite mode: typing replaces char at cursorPos."
func TestOverwriteModeReplacesCharAtCursor(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("abc")
	il.HandleEvent(keyEv(tcell.KeyHome)) // cursor to 0
	il.HandleEvent(insertKeyEv())         // switch to overwrite
	il.HandleEvent(runeEv('X'))
	if got := il.Text(); got != "Xbc" {
		t.Errorf("overwrite at pos 0 in \"abc\" with 'X': Text() = %q, want %q", got, "Xbc")
	}
}

// TestOverwriteModeDoesNotInsert verifies overwrite mode does not lengthen the text
// (it replaces, not inserts).
// Spec: "typing replaces char at cursorPos" — text length unchanged.
func TestOverwriteModeDoesNotInsert(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("abc")
	il.HandleEvent(keyEv(tcell.KeyHome))
	il.HandleEvent(insertKeyEv())
	il.HandleEvent(runeEv('X'))
	if got := il.Text(); len([]rune(got)) != 3 {
		t.Errorf("overwrite at pos 0: Text() = %q has %d runes, want 3 (replace, not insert)", got, len([]rune(got)))
	}
}

// TestOverwriteModeAdvancesCursor verifies cursor advances after replacement.
// Spec: "typing replaces char at cursorPos" — cursor advances one position.
func TestOverwriteModeAdvancesCursor(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("abc")
	il.HandleEvent(keyEv(tcell.KeyHome)) // cursor to 0
	il.HandleEvent(insertKeyEv())
	il.HandleEvent(runeEv('X'))
	if pos := il.CursorPos(); pos != 1 {
		t.Errorf("overwrite at pos 0: CursorPos() = %d after typing, want 1", pos)
	}
}

// TestOverwriteModeReplacesMultipleChars verifies sequential overwrite replaces each char.
// Spec: "In overwrite mode: typing replaces char at cursorPos."
func TestOverwriteModeReplacesMultipleChars(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("abcde")
	il.HandleEvent(keyEv(tcell.KeyHome))
	il.HandleEvent(insertKeyEv())
	il.HandleEvent(runeEv('X'))
	il.HandleEvent(runeEv('Y'))
	il.HandleEvent(runeEv('Z'))
	if got := il.Text(); got != "XYZde" {
		t.Errorf("overwrite 3 chars in \"abcde\": Text() = %q, want %q", got, "XYZde")
	}
}

// TestOverwriteModeAtEndAppendsRune verifies that overwrite at end of text appends.
// Spec: "At end of text, appends."
func TestOverwriteModeAtEndAppendsRune(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("abc") // cursor at 3 (end)
	il.HandleEvent(insertKeyEv())
	il.HandleEvent(runeEv('X'))
	if got := il.Text(); got != "abcX" {
		t.Errorf("overwrite at end of \"abc\" with 'X': Text() = %q, want %q", got, "abcX")
	}
}

// TestOverwriteModeAtEndLengthensText verifies appending at end increases text length.
// Spec: "At end of text, appends."
func TestOverwriteModeAtEndLengthensText(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("abc") // cursor at 3 (end)
	il.HandleEvent(insertKeyEv())
	il.HandleEvent(runeEv('X'))
	if got := len([]rune(il.Text())); got != 4 {
		t.Errorf("overwrite at end: text length = %d, want 4 (appended)", got)
	}
}

// TestInsertModeStillInsertsAfterToggle verifies that after toggling back to insert mode,
// typing inserts rather than overwrites.
// Spec: "Insert key toggles overwrite."
func TestInsertModeStillInsertsAfterToggle(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("abc")
	il.HandleEvent(keyEv(tcell.KeyHome))
	il.HandleEvent(insertKeyEv()) // overwrite on
	il.HandleEvent(insertKeyEv()) // overwrite off (back to insert)
	il.HandleEvent(runeEv('X'))
	if got := il.Text(); got != "Xabc" {
		t.Errorf("after toggle back to insert, typing 'X' at pos 0 of \"abc\": Text() = %q, want %q", got, "Xabc")
	}
}

// ---------------------------------------------------------------------------
// Falsifying tests: overwrite must NOT behave like insert
// ---------------------------------------------------------------------------

// TestOverwriteModeDoesNotLengthenTextMidString verifies overwrite in the middle of a
// string does not increase its length (catches an implementation that inserts instead).
// Falsification: if overwrite inserts, the text would be longer.
func TestOverwriteModeDoesNotLengthenTextMidString(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello")
	il.HandleEvent(keyEv(tcell.KeyHome))
	il.HandleEvent(keyEv(tcell.KeyRight)) // cursor to 1
	il.HandleEvent(insertKeyEv())
	il.HandleEvent(runeEv('Z'))
	if got := len([]rune(il.Text())); got != 5 {
		t.Errorf("overwrite at pos 1 of \"hello\": text length = %d, want 5 (not inserted)", got)
	}
}

// TestInsertModeIncreasesTextLength verifies that in insert mode (default), typing a
// character at a non-end position increases the text length by 1.
// Falsification: catches an implementation where insert accidentally overwrites.
func TestInsertModeIncreasesTextLength(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello")
	il.HandleEvent(keyEv(tcell.KeyHome)) // cursor to 0
	// Overwrite is off by default.
	il.HandleEvent(runeEv('Z'))
	if got := len([]rune(il.Text())); got != 6 {
		t.Errorf("insert at pos 0 of \"hello\": text length = %d, want 6 (inserted)", got)
	}
}

// ---------------------------------------------------------------------------
// Ctrl+Y: clear text
// ---------------------------------------------------------------------------

// TestCtrlYClearsText verifies Ctrl+Y sets text to empty string.
// Spec: "Ctrl+Y: clears text, cursor to 0, clears selection, scrollOffset to 0."
func TestCtrlYClearsText(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")
	il.HandleEvent(ctrlYEv())
	if got := il.Text(); got != "" {
		t.Errorf("Ctrl+Y: Text() = %q, want empty string", got)
	}
}

// TestCtrlYSetsCursorToZero verifies Ctrl+Y moves cursor to position 0.
// Spec: "cursor to 0."
func TestCtrlYSetsCursorToZero(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world") // cursor at 11
	il.HandleEvent(ctrlYEv())
	if pos := il.CursorPos(); pos != 0 {
		t.Errorf("Ctrl+Y: CursorPos() = %d, want 0", pos)
	}
}

// TestCtrlYClearsSelection verifies Ctrl+Y clears any active selection.
// Spec: "clears selection."
func TestCtrlYClearsSelection(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")
	il.HandleEvent(ctrlEv(tcell.KeyCtrlA)) // select all
	il.HandleEvent(ctrlYEv())
	start, end := il.Selection()
	if start != end {
		t.Errorf("Ctrl+Y: Selection() = (%d, %d), want equal (selection cleared)", start, end)
	}
}

// TestCtrlYEventConsumed verifies Ctrl+Y is consumed.
// Spec: "Event cleared for both."
func TestCtrlYEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")
	ev := ctrlYEv()
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Ctrl+Y event was not consumed")
	}
}

// TestCtrlYOnEmptyTextIsNoOp verifies Ctrl+Y on already-empty text is safe.
// Spec: "Ctrl+Y: clears text."
func TestCtrlYOnEmptyTextIsNoOp(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	// text is empty, cursor at 0
	il.HandleEvent(ctrlYEv()) // must not panic
	if got := il.Text(); got != "" {
		t.Errorf("Ctrl+Y on empty text: Text() = %q, want empty", got)
	}
	if pos := il.CursorPos(); pos != 0 {
		t.Errorf("Ctrl+Y on empty text: CursorPos() = %d, want 0", pos)
	}
}

// TestCtrlYOnEmptyEventConsumed verifies Ctrl+Y on empty text is still consumed.
func TestCtrlYOnEmptyEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	ev := ctrlYEv()
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Ctrl+Y on empty text was not consumed")
	}
}

// TestCtrlYAfterTypingClearsText verifies Ctrl+Y clears text that was typed interactively.
// Spec: "Ctrl+Y: clears text."
func TestCtrlYAfterTypingClearsText(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.HandleEvent(runeEv('h'))
	il.HandleEvent(runeEv('i'))
	il.HandleEvent(ctrlYEv())
	if got := il.Text(); got != "" {
		t.Errorf("Ctrl+Y after typing \"hi\": Text() = %q, want empty", got)
	}
}

// ---------------------------------------------------------------------------
// Falsifying tests: Ctrl+Y must NOT merely move cursor or clear selection only
// ---------------------------------------------------------------------------

// TestCtrlYActuallyClearsTextNotJustCursor verifies Ctrl+Y removes text content, not
// just moving cursor to 0 while leaving text intact.
// Falsification: catches an implementation that only sets cursorPos=0 without clearing text.
func TestCtrlYActuallyClearsTextNotJustCursor(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello")
	il.HandleEvent(ctrlYEv())
	// If only the cursor moved (wrongly), Text() would still be "hello".
	if got := il.Text(); got == "hello" {
		t.Errorf("Ctrl+Y: Text() = %q (unchanged); expected empty (Ctrl+Y must clear text, not just move cursor)", got)
	}
}

// TestCtrlYActuallyClearsSelectionNotJustText verifies Ctrl+Y also clears the selection,
// not just the text.
// Falsification: catches an implementation that only clears text but forgets the selection.
func TestCtrlYActuallyClearsSelectionNotJustText(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")
	il.HandleEvent(ctrlEv(tcell.KeyCtrlA)) // create selection covering all text
	il.HandleEvent(ctrlYEv())
	start, end := il.Selection()
	if start != end {
		t.Errorf("Ctrl+Y: Selection() = (%d, %d) — selection not cleared (start != end)", start, end)
	}
}
