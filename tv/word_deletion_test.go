package tv

// Tests for word deletion keyboard shortcuts: Ctrl+Backspace and Ctrl+Delete.
//
// Each test cites the spec requirement it verifies.

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ctrlBackspaceEv builds a Ctrl+Backspace keyboard event.
// Spec note: "For Ctrl+Backspace: tcell.KeyBackspace2 with ModCtrl"
func ctrlBackspaceEv() *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyBackspace2, Modifiers: tcell.ModCtrl}}
}

// ctrlDeleteEv builds a Ctrl+Delete keyboard event.
func ctrlDeleteEv() *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDelete, Modifiers: tcell.ModCtrl}}
}

// ---------------------------------------------------------------------------
// Ctrl+Backspace: delete from wordLeft(cursorPos) to cursorPos
// ---------------------------------------------------------------------------

// TestCtrlBackspaceDeletesCurrentWord verifies Ctrl+Backspace with cursor at end of
// "hello world" deletes "world" (positions 6..11).
// Spec: "Ctrl+Backspace: delete from wordLeft(cursorPos) to cursorPos."
func TestCtrlBackspaceDeletesCurrentWord(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world") // cursor at 11
	il.HandleEvent(ctrlBackspaceEv())
	if got := il.Text(); got != "hello " {
		t.Errorf("Ctrl+Backspace from end of \"hello world\": Text() = %q, want %q", got, "hello ")
	}
}

// TestCtrlBackspaceCursorMovesToWordLeft verifies cursor moves to wordLeft(pos) after deletion.
// Spec: "Ctrl+Backspace: delete from wordLeft(cursorPos) to cursorPos."
func TestCtrlBackspaceCursorMovesToWordLeft(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world") // cursor at 11
	il.HandleEvent(ctrlBackspaceEv())
	if pos := il.CursorPos(); pos != 6 {
		t.Errorf("after Ctrl+Backspace from 11: CursorPos() = %d, want 6", pos)
	}
}

// TestCtrlBackspaceDeletesFirstWord verifies Ctrl+Backspace from mid-word deletes back
// to the start of that word.
// Spec: "delete from wordLeft(cursorPos) to cursorPos."
func TestCtrlBackspaceDeletesFirstWord(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")
	// Move cursor to 6 (start of "world") and then Ctrl+Backspace deletes "hello ".
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft, Modifiers: tcell.ModCtrl}})
	// cursor at 6
	il.HandleEvent(ctrlBackspaceEv())
	if got := il.Text(); got != "world" {
		t.Errorf("Ctrl+Backspace from pos 6 in \"hello world\": Text() = %q, want %q", got, "world")
	}
}

// TestCtrlBackspaceAtStartIsNoOp verifies Ctrl+Backspace at pos 0 is a no-op for text.
// Spec: "wordLeft returns 0 at position 0, so the range is 0..0 — nothing deleted."
func TestCtrlBackspaceAtStartIsNoOp(t *testing.T) {
	il := ilAtPos("hello world", 0)
	il.HandleEvent(ctrlBackspaceEv())
	if got := il.Text(); got != "hello world" {
		t.Errorf("Ctrl+Backspace at pos 0: Text() = %q, want %q (unchanged)", got, "hello world")
	}
}

// TestCtrlBackspaceAtStartCursorStaysAtZero verifies cursor stays at 0 when no-op.
// Spec: cursor at wordLeft(pos) = 0 when already at 0.
func TestCtrlBackspaceAtStartCursorStaysAtZero(t *testing.T) {
	il := ilAtPos("hello world", 0)
	il.HandleEvent(ctrlBackspaceEv())
	if pos := il.CursorPos(); pos != 0 {
		t.Errorf("Ctrl+Backspace at pos 0: CursorPos() = %d, want 0", pos)
	}
}

// TestCtrlBackspaceEventConsumed verifies Ctrl+Backspace is consumed.
// Spec: "Event cleared."
func TestCtrlBackspaceEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")
	ev := ctrlBackspaceEv()
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Ctrl+Backspace event was not consumed")
	}
}

// TestCtrlBackspaceEventConsumedAtStart verifies Ctrl+Backspace is consumed even at pos 0.
// Spec: "Event cleared."
func TestCtrlBackspaceEventConsumedAtStart(t *testing.T) {
	il := ilAtPos("hello world", 0)
	ev := ctrlBackspaceEv()
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Ctrl+Backspace at pos 0 was not consumed")
	}
}

// ---------------------------------------------------------------------------
// Ctrl+Backspace with active selection: delete selection
// ---------------------------------------------------------------------------

// TestCtrlBackspaceWithSelectionDeletesSelection verifies that if there is an active
// selection, Ctrl+Backspace deletes the selection (not a word).
// Spec: "Ctrl+Backspace: if selection, delete it."
func TestCtrlBackspaceWithSelectionDeletesSelection(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")
	il.HandleEvent(ctrlEv(tcell.KeyCtrlA)) // select all
	il.HandleEvent(ctrlBackspaceEv())
	if got := il.Text(); got != "" {
		t.Errorf("Ctrl+Backspace with selection: Text() = %q, want empty", got)
	}
}

// TestCtrlBackspaceWithSelectionDeletesOnlySelection verifies it deletes only the
// selected range, not extra characters.
// Spec: "if selection, delete it."
func TestCtrlBackspaceWithSelectionDeletesOnlySelection(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("abcde")
	il.HandleEvent(keyEv(tcell.KeyHome))
	il.HandleEvent(keyEv(tcell.KeyRight)) // cursor to 1
	// Shift+Right ×2 selects "bc" (1..3).
	il.HandleEvent(shiftKeyEv(tcell.KeyRight))
	il.HandleEvent(shiftKeyEv(tcell.KeyRight))
	il.HandleEvent(ctrlBackspaceEv())
	if got := il.Text(); got != "ade" {
		t.Errorf("Ctrl+Backspace deleting selection \"bc\" from \"abcde\": Text() = %q, want %q", got, "ade")
	}
}

// TestCtrlBackspaceWithSelectionEventConsumed verifies it is consumed when there is a selection.
func TestCtrlBackspaceWithSelectionEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello")
	il.HandleEvent(ctrlEv(tcell.KeyCtrlA))
	ev := ctrlBackspaceEv()
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Ctrl+Backspace with selection was not consumed")
	}
}

// ---------------------------------------------------------------------------
// Ctrl+Delete: delete from cursorPos to wordRight(cursorPos)
// ---------------------------------------------------------------------------

// TestCtrlDeleteDeletesToNextWordBoundary verifies Ctrl+Delete from pos 0 in
// "hello world" deletes "hello " (positions 0..6).
// Spec: "Ctrl+Delete: delete from cursorPos to wordRight(cursorPos)."
func TestCtrlDeleteDeletesToNextWordBoundary(t *testing.T) {
	il := ilAtPos("hello world", 0)
	il.HandleEvent(ctrlDeleteEv())
	if got := il.Text(); got != "world" {
		t.Errorf("Ctrl+Delete from pos 0 in \"hello world\": Text() = %q, want %q", got, "world")
	}
}

// TestCtrlDeleteCursorStaysAtSamePosition verifies cursor stays at original position.
// Spec: "delete from cursorPos to wordRight(cursorPos)" — cursor stays at cursorPos.
func TestCtrlDeleteCursorStaysAtSamePosition(t *testing.T) {
	il := ilAtPos("hello world", 0)
	il.HandleEvent(ctrlDeleteEv())
	if pos := il.CursorPos(); pos != 0 {
		t.Errorf("after Ctrl+Delete from pos 0: CursorPos() = %d, want 0", pos)
	}
}

// TestCtrlDeleteFromMiddleOfWord verifies Ctrl+Delete from mid-word deletes to next
// word start.
// Spec: "delete from cursorPos to wordRight(cursorPos)."
func TestCtrlDeleteFromMiddleOfWord(t *testing.T) {
	il := ilAtPos("hello world", 3) // cursor at 3 ("hel|lo")
	il.HandleEvent(ctrlDeleteEv())
	// wordRight(3) = 6, so "lo " (positions 3..6) is deleted → "helworld"
	if got := il.Text(); got != "helworld" {
		t.Errorf("Ctrl+Delete from pos 3 in \"hello world\": Text() = %q, want %q", got, "helworld")
	}
}

// TestCtrlDeleteAtEndIsNoOp verifies Ctrl+Delete at end of text does not change text.
// Spec: "wordRight returns len(text) at end, so range is end..end — nothing deleted."
func TestCtrlDeleteAtEndIsNoOp(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world") // cursor at 11
	il.HandleEvent(ctrlDeleteEv())
	if got := il.Text(); got != "hello world" {
		t.Errorf("Ctrl+Delete at end: Text() = %q, want %q (unchanged)", got, "hello world")
	}
}

// TestCtrlDeleteAtEndCursorStaysAtEnd verifies cursor stays at end when no-op.
func TestCtrlDeleteAtEndCursorStaysAtEnd(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world") // cursor at 11
	il.HandleEvent(ctrlDeleteEv())
	if pos := il.CursorPos(); pos != 11 {
		t.Errorf("Ctrl+Delete at end: CursorPos() = %d, want 11", pos)
	}
}

// TestCtrlDeleteEventConsumed verifies Ctrl+Delete is consumed.
// Spec: "Event cleared."
func TestCtrlDeleteEventConsumed(t *testing.T) {
	il := ilAtPos("hello world", 0)
	ev := ctrlDeleteEv()
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Ctrl+Delete event was not consumed")
	}
}

// TestCtrlDeleteEventConsumedAtEnd verifies Ctrl+Delete is consumed even at end.
// Spec: "Event cleared."
func TestCtrlDeleteEventConsumedAtEnd(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")
	ev := ctrlDeleteEv()
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Ctrl+Delete at end was not consumed")
	}
}

// ---------------------------------------------------------------------------
// Ctrl+Delete with active selection: delete selection
// ---------------------------------------------------------------------------

// TestCtrlDeleteWithSelectionDeletesSelection verifies that if there is an active
// selection, Ctrl+Delete deletes the selection (not a word range).
// Spec: "Ctrl+Delete: if selection, delete it."
func TestCtrlDeleteWithSelectionDeletesSelection(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")
	il.HandleEvent(ctrlEv(tcell.KeyCtrlA)) // select all
	il.HandleEvent(ctrlDeleteEv())
	if got := il.Text(); got != "" {
		t.Errorf("Ctrl+Delete with selection: Text() = %q, want empty", got)
	}
}

// TestCtrlDeleteWithSelectionDeletesOnlySelection verifies only the selected text is removed.
// Spec: "if selection, delete it."
func TestCtrlDeleteWithSelectionDeletesOnlySelection(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("abcde")
	il.HandleEvent(keyEv(tcell.KeyHome))
	il.HandleEvent(keyEv(tcell.KeyRight)) // cursor to 1
	// Shift+Right ×2 selects "bc" (1..3).
	il.HandleEvent(shiftKeyEv(tcell.KeyRight))
	il.HandleEvent(shiftKeyEv(tcell.KeyRight))
	il.HandleEvent(ctrlDeleteEv())
	if got := il.Text(); got != "ade" {
		t.Errorf("Ctrl+Delete deleting selection \"bc\" from \"abcde\": Text() = %q, want %q", got, "ade")
	}
}

// TestCtrlDeleteWithSelectionEventConsumed verifies it is consumed when there is a selection.
func TestCtrlDeleteWithSelectionEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello")
	il.HandleEvent(ctrlEv(tcell.KeyCtrlA))
	ev := ctrlDeleteEv()
	il.HandleEvent(ev)
	if !ev.IsCleared() {
		t.Error("Ctrl+Delete with selection was not consumed")
	}
}

// ---------------------------------------------------------------------------
// Plain Backspace/Delete unchanged
// ---------------------------------------------------------------------------

// TestPlainBackspaceStillWorksDeletesOneChar verifies plain Backspace still deletes
// exactly one character (not a word).
// Spec: "Plain Backspace/Delete unchanged."
func TestPlainBackspaceStillWorksDeletesOneChar(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world") // cursor at 11
	il.HandleEvent(keyEv(tcell.KeyBackspace2))
	if got := il.Text(); got != "hello worl" {
		t.Errorf("plain Backspace from end of \"hello world\": Text() = %q, want %q", got, "hello worl")
	}
}

// TestPlainDeleteStillWorksDeletesOneChar verifies plain Delete still deletes
// exactly one character (not a word).
// Spec: "Plain Backspace/Delete unchanged."
func TestPlainDeleteStillWorksDeletesOneChar(t *testing.T) {
	il := ilAtPos("hello world", 0)
	il.HandleEvent(keyEv(tcell.KeyDelete))
	if got := il.Text(); got != "ello world" {
		t.Errorf("plain Delete from pos 0 in \"hello world\": Text() = %q, want %q", got, "ello world")
	}
}

// ---------------------------------------------------------------------------
// Falsifying tests: Ctrl+Backspace/Delete must NOT behave like plain variants
// ---------------------------------------------------------------------------

// TestCtrlBackspaceDeletesMoreThanOneChar verifies Ctrl+Backspace deletes more than
// one character (it deletes a whole word segment).
// Falsification: an implementation that maps Ctrl+Backspace to plain Backspace would fail.
func TestCtrlBackspaceDeletesMoreThanOneChar(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world") // cursor at 11
	il.HandleEvent(ctrlBackspaceEv())
	// If it only deleted one char, text would be "hello worl". It must delete "world " or "world".
	if got := il.Text(); got == "hello worl" {
		t.Errorf("Ctrl+Backspace deleted only one char (result: %q); expected word-level deletion", got)
	}
}

// TestCtrlDeleteDeletesMoreThanOneChar verifies Ctrl+Delete deletes more than
// one character.
// Falsification: an implementation that maps Ctrl+Delete to plain Delete would fail.
func TestCtrlDeleteDeletesMoreThanOneChar(t *testing.T) {
	il := ilAtPos("hello world", 0)
	il.HandleEvent(ctrlDeleteEv())
	// If it only deleted one char, text would be "ello world". It must delete "hello " or similar.
	if got := il.Text(); got == "ello world" {
		t.Errorf("Ctrl+Delete deleted only one char (result: %q); expected word-level deletion", got)
	}
}
