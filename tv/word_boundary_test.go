package tv

// Tests for word boundary helpers: wordLeft and wordRight.
//
// wordLeft and wordRight are unexported methods on InputLine, so they are
// exercised indirectly via Ctrl+Left / Ctrl+Right keyboard events, which
// map directly to wordLeft / wordRight results.
//
// Each test cites the spec requirement it verifies.

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ctrlLeftEv builds a Ctrl+Left keyboard event.
func ctrlLeftEv() *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft, Modifiers: tcell.ModCtrl}}
}

// ctrlRightEv builds a Ctrl+Right keyboard event.
func ctrlRightEv() *Event {
	return &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight, Modifiers: tcell.ModCtrl}}
}

// ilAt creates an InputLine with the given text and cursor explicitly moved to
// pos using Home + Right presses, for tests that need a non-end starting position.
func ilAtPos(text string, pos int) *InputLine {
	il := NewInputLine(NewRect(0, 0, 80, 1), 0)
	il.SetText(text) // cursor now at len(text)
	// Move to 0 then advance to pos.
	il.HandleEvent(keyEv(tcell.KeyHome))
	for i := 0; i < pos; i++ {
		il.HandleEvent(keyEv(tcell.KeyRight))
	}
	return il
}

// ---------------------------------------------------------------------------
// wordLeft (via Ctrl+Left)
// ---------------------------------------------------------------------------

// TestWordLeftFromEndOfHelloWorld verifies wordLeft from end of "hello world" → 6.
// Spec: "from position pos, move left to the start of the previous word."
func TestWordLeftFromEndOfHelloWorld(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world") // cursor at 11 (end)
	il.HandleEvent(ctrlLeftEv())
	if pos := il.CursorPos(); pos != 6 {
		t.Errorf("wordLeft from 11 in \"hello world\" = %d, want 6 (start of \"world\")", pos)
	}
}

// TestWordLeftFromStartOfSecondWord verifies wordLeft from pos 6 → 0.
// Spec: "move left to the start of the previous word."
func TestWordLeftFromStartOfSecondWord(t *testing.T) {
	il := ilAtPos("hello world", 6)
	il.HandleEvent(ctrlLeftEv())
	if pos := il.CursorPos(); pos != 0 {
		t.Errorf("wordLeft from 6 in \"hello world\" = %d, want 0 (start of \"hello\")", pos)
	}
}

// TestWordLeftFromZeroReturnsZero verifies wordLeft from 0 → 0.
// Spec: "Returns 0 if no word boundary found."
func TestWordLeftFromZeroReturnsZero(t *testing.T) {
	il := ilAtPos("hello world", 0)
	il.HandleEvent(ctrlLeftEv())
	if pos := il.CursorPos(); pos != 0 {
		t.Errorf("wordLeft from 0 = %d, want 0 (already at start)", pos)
	}
}

// TestWordLeftFromMiddleOfFirstWord verifies wordLeft from pos 3 in "hello world" → 0.
// Spec: "move left to the start of the previous word."
func TestWordLeftFromMiddleOfFirstWord(t *testing.T) {
	il := ilAtPos("hello world", 3)
	il.HandleEvent(ctrlLeftEv())
	if pos := il.CursorPos(); pos != 0 {
		t.Errorf("wordLeft from 3 in \"hello world\" = %d, want 0 (start of \"hello\")", pos)
	}
}

// TestWordLeftMultipleSpaces verifies wordLeft handles multiple spaces correctly.
// Spec: "A word boundary is a transition from space to non-space when scanning leftward."
// "hello  world" (double space): wordLeft from end (12) → 7.
func TestWordLeftMultipleSpaces(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello  world") // cursor at 12
	il.HandleEvent(ctrlLeftEv())
	if pos := il.CursorPos(); pos != 7 {
		t.Errorf("wordLeft from 12 in \"hello  world\" = %d, want 7 (start of \"world\")", pos)
	}
}

// TestWordLeftOnLeadingSpaces verifies wordLeft on "  hello" from pos 2 → 0.
// Spec: "Returns 0 if no word boundary found."
func TestWordLeftOnLeadingSpaces(t *testing.T) {
	il := ilAtPos("  hello", 2)
	il.HandleEvent(ctrlLeftEv())
	if pos := il.CursorPos(); pos != 0 {
		t.Errorf("wordLeft from 2 in \"  hello\" = %d, want 0 (spaces then start)", pos)
	}
}

// TestWordLeftOnEmptyText verifies wordLeft on empty text → 0.
// Spec: "Returns 0 if no word boundary found."
func TestWordLeftOnEmptyText(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	// empty, cursor at 0
	il.HandleEvent(ctrlLeftEv())
	if pos := il.CursorPos(); pos != 0 {
		t.Errorf("wordLeft on empty text = %d, want 0", pos)
	}
}

// ---------------------------------------------------------------------------
// wordLeft — falsifying tests
// ---------------------------------------------------------------------------

// TestWordLeftNotCharLevel_FromEnd verifies wordLeft does NOT return pos-1 (char-level).
// Spec: falsification — wordLeft must not return pos-1.
func TestWordLeftNotCharLevel_FromEnd(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world") // cursor at 11
	il.HandleEvent(ctrlLeftEv())
	if pos := il.CursorPos(); pos == 10 {
		t.Errorf("wordLeft from 11 = 10 (pos-1); that is char-level movement, not word-level")
	}
}

// TestWordLeftNotCharLevel_FromMiddle verifies wordLeft does NOT return pos-1 from middle.
func TestWordLeftNotCharLevel_FromMiddle(t *testing.T) {
	il := ilAtPos("hello world", 8) // mid-word ("wor|ld")
	il.HandleEvent(ctrlLeftEv())
	if pos := il.CursorPos(); pos == 7 {
		t.Errorf("wordLeft from 8 = 7 (pos-1); that is char-level movement, not word-level")
	}
}

// ---------------------------------------------------------------------------
// wordRight (via Ctrl+Right)
// ---------------------------------------------------------------------------

// TestWordRightFromStartOfHelloWorld verifies wordRight from 0 in "hello world" → 6.
// Spec: "move right past the current word and any following spaces, arriving at start of next word."
func TestWordRightFromStartOfHelloWorld(t *testing.T) {
	il := ilAtPos("hello world", 0)
	il.HandleEvent(ctrlRightEv())
	if pos := il.CursorPos(); pos != 6 {
		t.Errorf("wordRight from 0 in \"hello world\" = %d, want 6 (start of \"world\")", pos)
	}
}

// TestWordRightFromStartOfSecondWord verifies wordRight from 6 in "hello world" → 11 (len).
// Spec: "Returns len(text) if no word boundary found."
func TestWordRightFromStartOfSecondWord(t *testing.T) {
	il := ilAtPos("hello world", 6)
	il.HandleEvent(ctrlRightEv())
	if pos := il.CursorPos(); pos != 11 {
		t.Errorf("wordRight from 6 in \"hello world\" = %d, want 11 (len)", pos)
	}
}

// TestWordRightFromEnd verifies wordRight from end → end (already at end).
// Spec: "Returns len(text) if no word boundary found."
func TestWordRightFromEnd(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world") // cursor at 11
	il.HandleEvent(ctrlRightEv())
	if pos := il.CursorPos(); pos != 11 {
		t.Errorf("wordRight from 11 (end) in \"hello world\" = %d, want 11", pos)
	}
}

// TestWordRightFromMiddleOfWord verifies wordRight from pos 3 in "hello world" → 6.
// Spec: "move right past the current word and any following spaces."
func TestWordRightFromMiddleOfWord(t *testing.T) {
	il := ilAtPos("hello world", 3)
	il.HandleEvent(ctrlRightEv())
	if pos := il.CursorPos(); pos != 6 {
		t.Errorf("wordRight from 3 in \"hello world\" = %d, want 6 (start of \"world\")", pos)
	}
}

// TestWordRightMultipleSpaces verifies wordRight handles multiple spaces correctly.
// Spec: "move right past the current word and any following spaces."
// "hello  world" (double space): wordRight from 0 → 7.
func TestWordRightMultipleSpaces(t *testing.T) {
	il := ilAtPos("hello  world", 0)
	il.HandleEvent(ctrlRightEv())
	if pos := il.CursorPos(); pos != 7 {
		t.Errorf("wordRight from 0 in \"hello  world\" = %d, want 7 (start of \"world\")", pos)
	}
}

// TestWordRightOnEmptyText verifies wordRight on empty text → 0.
// Spec: "Returns len(text) if no word boundary found."
func TestWordRightOnEmptyText(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	// empty, cursor at 0
	il.HandleEvent(ctrlRightEv())
	if pos := il.CursorPos(); pos != 0 {
		t.Errorf("wordRight on empty text = %d, want 0", pos)
	}
}

// ---------------------------------------------------------------------------
// wordRight — falsifying tests
// ---------------------------------------------------------------------------

// TestWordRightNotCharLevel_FromStart verifies wordRight does NOT return pos+1.
// Spec: falsification — wordRight must not return pos+1.
func TestWordRightNotCharLevel_FromStart(t *testing.T) {
	il := ilAtPos("hello world", 0)
	il.HandleEvent(ctrlRightEv())
	if pos := il.CursorPos(); pos == 1 {
		t.Errorf("wordRight from 0 = 1 (pos+1); that is char-level movement, not word-level")
	}
}

// TestWordRightNotCharLevel_FromMiddle verifies wordRight does NOT return pos+1 from middle.
func TestWordRightNotCharLevel_FromMiddle(t *testing.T) {
	il := ilAtPos("hello world", 3) // mid-word ("hel|lo")
	il.HandleEvent(ctrlRightEv())
	if pos := il.CursorPos(); pos == 4 {
		t.Errorf("wordRight from 3 = 4 (pos+1); that is char-level movement, not word-level")
	}
}
