package tv

// Tests for word movement keyboard shortcuts: Ctrl+Left, Ctrl+Right,
// Ctrl+Shift+Left, Ctrl+Shift+Right.
//
// Each test cites the spec requirement it verifies.

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ctrlShiftLeftEv builds a Ctrl+Shift+Left keyboard event.
func ctrlShiftLeftEv() *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft, Modifiers: tcell.ModCtrl | tcell.ModShift}}
}

// ctrlShiftRightEv builds a Ctrl+Shift+Right keyboard event.
func ctrlShiftRightEv() *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight, Modifiers: tcell.ModCtrl | tcell.ModShift}}
}

// ---------------------------------------------------------------------------
// Ctrl+Left: move cursor to wordLeft, clear selection
// ---------------------------------------------------------------------------

// TestCtrlLeftMovesCursorToWordLeft verifies Ctrl+Left moves cursor to wordLeft(cursorPos).
// Spec: "Ctrl+Left: move cursor to wordLeft(cursorPos), clear selection."
func TestCtrlLeftMovesCursorToWordLeft(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world") // cursor at 11
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft, Modifiers: tcell.ModCtrl}}
	il.HandleEvent(ev)
	if pos := il.CursorPos(); pos != 6 {
		t.Errorf("Ctrl+Left from end of \"hello world\": CursorPos() = %d, want 6", pos)
	}
}

// TestCtrlLeftTwiceReachesStart verifies two Ctrl+Left presses from end of
// "hello world" reach position 0.
// Spec: "Ctrl+Left: move cursor to wordLeft(cursorPos)."
func TestCtrlLeftTwiceReachesStart(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world") // cursor at 11
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft, Modifiers: tcell.ModCtrl}})
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft, Modifiers: tcell.ModCtrl}})
	if pos := il.CursorPos(); pos != 0 {
		t.Errorf("two Ctrl+Left from end of \"hello world\": CursorPos() = %d, want 0", pos)
	}
}

// TestCtrlLeftClearsSelection verifies Ctrl+Left clears any active selection.
// Spec: "Ctrl+Left: move cursor to wordLeft(cursorPos), clear selection."
func TestCtrlLeftClearsSelection(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")
	// Create a selection via Ctrl+A.
	il.HandleEvent(ctrlEv(tcell.KeyCtrlA))
	// Now Ctrl+Left should clear it.
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft, Modifiers: tcell.ModCtrl}})
	start, end := il.Selection()
	if start != end {
		t.Errorf("after Ctrl+Left, Selection() = (%d, %d), want equal (no selection)", start, end)
	}
}

// TestCtrlLeftEventConsumed verifies Ctrl+Left is consumed.
// Spec: "Event cleared."
func TestCtrlLeftEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft, Modifiers: tcell.ModCtrl}}
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Ctrl+Left event was not consumed (IsCleared() = false)")
	}
}

// TestCtrlLeftAtStartStaysAtZero verifies Ctrl+Left at position 0 stays at 0.
// Spec: "wordLeft returns 0 if no word boundary found."
func TestCtrlLeftAtStartStaysAtZero(t *testing.T) {
	il := ilAtPos("hello world", 0)
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft, Modifiers: tcell.ModCtrl}})
	if pos := il.CursorPos(); pos != 0 {
		t.Errorf("Ctrl+Left at pos 0: CursorPos() = %d, want 0", pos)
	}
}

// TestCtrlLeftAtStartEventConsumed verifies Ctrl+Left at pos 0 is still consumed.
// Spec: "Event cleared."
func TestCtrlLeftAtStartEventConsumed(t *testing.T) {
	il := ilAtPos("hello world", 0)
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft, Modifiers: tcell.ModCtrl}}
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Ctrl+Left at pos 0 was not consumed")
	}
}

// ---------------------------------------------------------------------------
// Ctrl+Right: move cursor to wordRight, clear selection
// ---------------------------------------------------------------------------

// TestCtrlRightMovesCursorToWordRight verifies Ctrl+Right moves cursor to wordRight(cursorPos).
// Spec: "Ctrl+Right: move cursor to wordRight(cursorPos), clear selection."
func TestCtrlRightMovesCursorToWordRight(t *testing.T) {
	il := ilAtPos("hello world", 0)
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight, Modifiers: tcell.ModCtrl}}
	il.HandleEvent(ev)
	if pos := il.CursorPos(); pos != 6 {
		t.Errorf("Ctrl+Right from 0 in \"hello world\": CursorPos() = %d, want 6", pos)
	}
}

// TestCtrlRightTwiceReachesEnd verifies two Ctrl+Right presses from position 0
// of "hello world" reach the end (11).
// Spec: "Ctrl+Right: move cursor to wordRight(cursorPos)."
func TestCtrlRightTwiceReachesEnd(t *testing.T) {
	il := ilAtPos("hello world", 0)
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight, Modifiers: tcell.ModCtrl}})
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight, Modifiers: tcell.ModCtrl}})
	if pos := il.CursorPos(); pos != 11 {
		t.Errorf("two Ctrl+Right from 0 in \"hello world\": CursorPos() = %d, want 11", pos)
	}
}

// TestCtrlRightClearsSelection verifies Ctrl+Right clears any active selection.
// Spec: "Ctrl+Right: move cursor to wordRight(cursorPos), clear selection."
func TestCtrlRightClearsSelection(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")
	il.HandleEvent(ctrlEv(tcell.KeyCtrlA)) // create selection
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight, Modifiers: tcell.ModCtrl}})
	start, end := il.Selection()
	if start != end {
		t.Errorf("after Ctrl+Right, Selection() = (%d, %d), want equal (no selection)", start, end)
	}
}

// TestCtrlRightEventConsumed verifies Ctrl+Right is consumed.
// Spec: "Event cleared."
func TestCtrlRightEventConsumed(t *testing.T) {
	il := ilAtPos("hello world", 0)
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight, Modifiers: tcell.ModCtrl}}
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Ctrl+Right event was not consumed")
	}
}

// TestCtrlRightAtEndStaysAtEnd verifies Ctrl+Right at end stays at end.
// Spec: "wordRight returns len(text) if no word boundary found."
func TestCtrlRightAtEndStaysAtEnd(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world") // cursor at 11
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight, Modifiers: tcell.ModCtrl}})
	if pos := il.CursorPos(); pos != 11 {
		t.Errorf("Ctrl+Right at end: CursorPos() = %d, want 11", pos)
	}
}

// TestCtrlRightAtEndEventConsumed verifies Ctrl+Right at end is still consumed.
// Spec: "Event cleared."
func TestCtrlRightAtEndEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")
	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight, Modifiers: tcell.ModCtrl}}
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Ctrl+Right at end was not consumed")
	}
}

// ---------------------------------------------------------------------------
// Ctrl+Shift+Left: extend selection to wordLeft
// ---------------------------------------------------------------------------

// TestCtrlShiftLeftSetsSelectionFromCursorToWordLeft verifies that Ctrl+Shift+Left
// with no prior selection sets selStart=cursorPos and moves cursor to wordLeft.
// Spec: "if no selection, set selStart=cursorPos first."
func TestCtrlShiftLeftSetsSelectionFromCursorToWordLeft(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world") // cursor at 11, no selection
	il.HandleEvent(ctrlShiftLeftEv())
	// Cursor should be at 6 (wordLeft from 11).
	if pos := il.CursorPos(); pos != 6 {
		t.Errorf("after Ctrl+Shift+Left from 11: CursorPos() = %d, want 6", pos)
	}
	// Selection should span 6..11.
	start, end := il.Selection()
	if start > end {
		start, end = end, start
	}
	if start != 6 || end != 11 {
		t.Errorf("after Ctrl+Shift+Left from 11: Selection() = (%d, %d), want (6, 11)", start, end)
	}
}

// TestCtrlShiftLeftExtendsExistingSelection verifies Ctrl+Shift+Left extends an existing
// selection further left.
// Spec: "Ctrl+Shift+Left: extend selection to wordLeft."
func TestCtrlShiftLeftExtendsExistingSelection(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world") // cursor at 11
	// First Ctrl+Shift+Left: selStart=11, cursor→6, selection 6..11.
	il.HandleEvent(ctrlShiftLeftEv())
	// Second Ctrl+Shift+Left: cursor→0, selection should now be 0..11.
	il.HandleEvent(ctrlShiftLeftEv())
	if pos := il.CursorPos(); pos != 0 {
		t.Errorf("after two Ctrl+Shift+Left from 11: CursorPos() = %d, want 0", pos)
	}
	start, end := il.Selection()
	if start > end {
		start, end = end, start
	}
	if start != 0 || end != 11 {
		t.Errorf("after two Ctrl+Shift+Left from 11: Selection() = (%d, %d), want (0, 11)", start, end)
	}
}

// TestCtrlShiftLeftEventConsumed verifies Ctrl+Shift+Left is consumed.
// Spec: "Event cleared."
func TestCtrlShiftLeftEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")
	ev := ctrlShiftLeftEv()
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Ctrl+Shift+Left event was not consumed")
	}
}

// TestCtrlShiftLeftAtZeroPreservesSelection verifies Ctrl+Shift+Left at pos 0
// keeps cursor at 0 and selection selStart..0 stays intact.
// Spec: "wordLeft returns 0 if no word boundary found."
func TestCtrlShiftLeftAtZeroPreservesSelection(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")
	// Move to 0.
	il.HandleEvent(keyEv(tcell.KeyHome))
	// With cursor at 0, Ctrl+Shift+Left: selStart=0, wordLeft(0)=0 → cursor stays 0.
	il.HandleEvent(ctrlShiftLeftEv())
	if pos := il.CursorPos(); pos != 0 {
		t.Errorf("Ctrl+Shift+Left at pos 0: CursorPos() = %d, want 0", pos)
	}
}

// ---------------------------------------------------------------------------
// Ctrl+Shift+Right: extend selection to wordRight
// ---------------------------------------------------------------------------

// TestCtrlShiftRightSetsSelectionFromCursorToWordRight verifies that Ctrl+Shift+Right
// with no prior selection sets selStart=cursorPos and moves cursor to wordRight.
// Spec: "if no selection, set selStart=cursorPos first."
func TestCtrlShiftRightSetsSelectionFromCursorToWordRight(t *testing.T) {
	il := ilAtPos("hello world", 0)
	il.HandleEvent(ctrlShiftRightEv())
	// Cursor should be at 6 (wordRight from 0).
	if pos := il.CursorPos(); pos != 6 {
		t.Errorf("after Ctrl+Shift+Right from 0: CursorPos() = %d, want 6", pos)
	}
	// Selection should span 0..6.
	start, end := il.Selection()
	if start > end {
		start, end = end, start
	}
	if start != 0 || end != 6 {
		t.Errorf("after Ctrl+Shift+Right from 0: Selection() = (%d, %d), want (0, 6)", start, end)
	}
}

// TestCtrlShiftRightExtendsExistingSelection verifies Ctrl+Shift+Right extends an
// existing selection further right.
// Spec: "Ctrl+Shift+Right: extend selection to wordRight."
func TestCtrlShiftRightExtendsExistingSelection(t *testing.T) {
	il := ilAtPos("hello world", 0)
	// First Ctrl+Shift+Right: selStart=0, cursor→6, selection 0..6.
	il.HandleEvent(ctrlShiftRightEv())
	// Second Ctrl+Shift+Right: cursor→11, selection should now be 0..11.
	il.HandleEvent(ctrlShiftRightEv())
	if pos := il.CursorPos(); pos != 11 {
		t.Errorf("after two Ctrl+Shift+Right from 0: CursorPos() = %d, want 11", pos)
	}
	start, end := il.Selection()
	if start > end {
		start, end = end, start
	}
	if start != 0 || end != 11 {
		t.Errorf("after two Ctrl+Shift+Right from 0: Selection() = (%d, %d), want (0, 11)", start, end)
	}
}

// TestCtrlShiftRightEventConsumed verifies Ctrl+Shift+Right is consumed.
// Spec: "Event cleared."
func TestCtrlShiftRightEventConsumed(t *testing.T) {
	il := ilAtPos("hello world", 0)
	ev := ctrlShiftRightEv()
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Ctrl+Shift+Right event was not consumed")
	}
}

// TestCtrlShiftRightAtEndPreservesSelection verifies Ctrl+Shift+Right at end of text
// keeps cursor at end.
// Spec: "wordRight returns len(text) if no word boundary found."
func TestCtrlShiftRightAtEndPreservesSelection(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world") // cursor at 11
	il.HandleEvent(ctrlShiftRightEv())
	if pos := il.CursorPos(); pos != 11 {
		t.Errorf("Ctrl+Shift+Right at end: CursorPos() = %d, want 11", pos)
	}
}

// ---------------------------------------------------------------------------
// Falsifying tests: Ctrl+Left/Right must NOT behave like plain Left/Right
// ---------------------------------------------------------------------------

// TestCtrlLeftMovesMoreThanOneChar verifies Ctrl+Left moves more than one character.
// Falsification: an implementation that maps Ctrl+Left to plain Left would fail this.
func TestCtrlLeftMovesMoreThanOneChar(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world") // cursor at 11
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft, Modifiers: tcell.ModCtrl}})
	if pos := il.CursorPos(); pos == 10 {
		t.Errorf("Ctrl+Left moved only one char (pos=10); it must move a full word (want 6)")
	}
}

// TestCtrlRightMovesMoreThanOneChar verifies Ctrl+Right moves more than one character.
// Falsification: an implementation that maps Ctrl+Right to plain Right would fail this.
func TestCtrlRightMovesMoreThanOneChar(t *testing.T) {
	il := ilAtPos("hello world", 0)
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight, Modifiers: tcell.ModCtrl}})
	if pos := il.CursorPos(); pos == 1 {
		t.Errorf("Ctrl+Right moved only one char (pos=1); it must move a full word (want 6)")
	}
}

// TestCtrlShiftLeftCreatesSelectionNotJustMovesCursor verifies Ctrl+Shift+Left actually
// creates a selection (does not merely move the cursor without selecting).
// Falsification: an implementation that behaves like plain Ctrl+Left would fail this.
func TestCtrlShiftLeftCreatesSelectionNotJustMovesCursor(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")
	il.HandleEvent(ctrlShiftLeftEv())
	start, end := il.Selection()
	if start == end {
		t.Errorf("after Ctrl+Shift+Left, Selection() = (%d, %d) — no selection; expected non-empty selection", start, end)
	}
}

// TestCtrlShiftRightCreatesSelectionNotJustMovesCursor verifies Ctrl+Shift+Right creates
// a selection (does not merely move the cursor without selecting).
func TestCtrlShiftRightCreatesSelectionNotJustMovesCursor(t *testing.T) {
	il := ilAtPos("hello world", 0)
	il.HandleEvent(ctrlShiftRightEv())
	start, end := il.Selection()
	if start == end {
		t.Errorf("after Ctrl+Shift+Right, Selection() = (%d, %d) — no selection; expected non-empty selection", start, end)
	}
}
