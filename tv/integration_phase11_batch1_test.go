package tv

// integration_phase11_batch1_test.go — Integration tests for Phase 11 Tasks 1–4:
// InputLine Keyboard Enhancements Checkpoint.
//
// Verifies cross-component behavior of the keyboard enhancements added in Tasks 1–4:
//   Task 1: Word boundary helpers (wordLeft / wordRight)
//   Task 2: Word movement — Ctrl+Left / Ctrl+Right
//   Task 3: Word deletion — Ctrl+Backspace / Ctrl+Delete
//   Task 4: Insert/overwrite mode toggle + Ctrl+Y clear
//
// Each test exercises a full scenario from SetText through keyboard events to final
// state inspection, verifying that the composed behaviors work correctly together.
//
// Test naming: TestIntegrationPhase11<DescriptiveSuffix>

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// Requirement 1: Ctrl+Left moves cursor to start of "world" (position 6)
// ---------------------------------------------------------------------------

// TestIntegrationPhase11CtrlLeftFromEndLandsAtWordStart verifies that typing
// "hello world" and pressing Ctrl+Left moves the cursor to position 6, which is
// the start of the word "world".
//
// Scenario: SetText("hello world") → cursor=11 → Ctrl+Left → cursor=6.
func TestIntegrationPhase11CtrlLeftFromEndLandsAtWordStart(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world") // cursor at 11

	event := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft, Modifiers: tcell.ModCtrl}}
	il.HandleEvent(event)

	if pos := il.CursorPos(); pos != 6 {
		t.Errorf("Ctrl+Left from end of \"hello world\": CursorPos() = %d, want 6 (start of \"world\")", pos)
	}
	if got := il.Text(); got != "hello world" {
		t.Errorf("Ctrl+Left must not modify text: Text() = %q, want %q", got, "hello world")
	}
}

// TestIntegrationPhase11CtrlLeftFromEndConsumedEvent verifies that the Ctrl+Left event
// is consumed as part of the word-movement operation.
func TestIntegrationPhase11CtrlLeftFromEndConsumedEvent(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	event := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft, Modifiers: tcell.ModCtrl}}
	il.HandleEvent(event)

	if !event.IsCleared() {
		t.Error("Ctrl+Left event was not consumed")
	}
}

// ---------------------------------------------------------------------------
// Requirement 2: Ctrl+Left twice from end moves cursor to start of "hello" (position 0)
// ---------------------------------------------------------------------------

// TestIntegrationPhase11CtrlLeftTwiceReachesStart verifies that two successive Ctrl+Left
// presses from the end of "hello world" move the cursor all the way to position 0.
//
// Scenario: cursor=11 → Ctrl+Left → cursor=6 → Ctrl+Left → cursor=0.
func TestIntegrationPhase11CtrlLeftTwiceReachesStart(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world") // cursor at 11

	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft, Modifiers: tcell.ModCtrl}})
	if pos := il.CursorPos(); pos != 6 {
		t.Fatalf("after first Ctrl+Left: CursorPos() = %d, want 6 (pre-condition for second step)", pos)
	}

	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft, Modifiers: tcell.ModCtrl}})
	if pos := il.CursorPos(); pos != 0 {
		t.Errorf("after second Ctrl+Left: CursorPos() = %d, want 0 (start of \"hello\")", pos)
	}
}

// TestIntegrationPhase11CtrlLeftTwiceTextUnchanged verifies the text is unchanged after
// two Ctrl+Left presses.
func TestIntegrationPhase11CtrlLeftTwiceTextUnchanged(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft, Modifiers: tcell.ModCtrl}})
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft, Modifiers: tcell.ModCtrl}})

	if got := il.Text(); got != "hello world" {
		t.Errorf("text after two Ctrl+Left: %q, want %q (must not modify text)", got, "hello world")
	}
}

// ---------------------------------------------------------------------------
// Requirement 3: Ctrl+Right from position 0 moves to position 6 (start of "world")
// ---------------------------------------------------------------------------

// TestIntegrationPhase11CtrlRightFromStartLandsAtWordBoundary verifies that with cursor at
// position 0, pressing Ctrl+Right moves to position 6, the start of "world".
//
// Scenario: SetText("hello world") → Home → cursor=0 → Ctrl+Right → cursor=6.
func TestIntegrationPhase11CtrlRightFromStartLandsAtWordBoundary(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyHome}})

	if pre := il.CursorPos(); pre != 0 {
		t.Fatalf("pre-condition: cursor after Home = %d, want 0", pre)
	}

	event := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight, Modifiers: tcell.ModCtrl}}
	il.HandleEvent(event)

	if pos := il.CursorPos(); pos != 6 {
		t.Errorf("Ctrl+Right from pos 0 in \"hello world\": CursorPos() = %d, want 6 (start of \"world\")", pos)
	}
}

// TestIntegrationPhase11CtrlRightFromStartConsumedEvent verifies the event is consumed.
func TestIntegrationPhase11CtrlRightFromStartConsumedEvent(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyHome}})

	event := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight, Modifiers: tcell.ModCtrl}}
	il.HandleEvent(event)

	if !event.IsCleared() {
		t.Error("Ctrl+Right event was not consumed")
	}
}

// TestIntegrationPhase11CtrlRightFromStartClearsSelection verifies Ctrl+Right clears any
// existing selection while moving the cursor word-right.
func TestIntegrationPhase11CtrlRightFromStartClearsSelection(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")
	// Create a selection covering the whole text.
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlA}})
	selStart, selEnd := il.Selection()
	if selStart == selEnd {
		t.Fatalf("pre-condition: Ctrl+A should create a selection, got selStart=%d selEnd=%d", selStart, selEnd)
	}

	// Ctrl+Right should clear the selection.
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight, Modifiers: tcell.ModCtrl}})

	start, end := il.Selection()
	if start != end {
		t.Errorf("after Ctrl+Right, Selection() = (%d, %d); expected no selection (start==end)", start, end)
	}
}

// ---------------------------------------------------------------------------
// Requirement 4: Ctrl+Shift+Left selects previous word
// ---------------------------------------------------------------------------

// TestIntegrationPhase11CtrlShiftLeftSelectsPreviousWord verifies that with cursor at the
// end of "hello world" (position 11), Ctrl+Shift+Left creates a selection from 11 to 6,
// covering the word "world".
//
// Scenario: cursor=11, no selection → Ctrl+Shift+Left → selStart=11, selEnd=6.
func TestIntegrationPhase11CtrlShiftLeftSelectsPreviousWord(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world") // cursor at 11, no selection

	event := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft, Modifiers: tcell.ModCtrl | tcell.ModShift}}
	il.HandleEvent(event)

	selStart, selEnd := il.Selection()
	if selStart != 11 || selEnd != 6 {
		t.Errorf("Ctrl+Shift+Left from pos 11: Selection() = (%d, %d), want (11, 6)", selStart, selEnd)
	}
}

// TestIntegrationPhase11CtrlShiftLeftCursorAtWordStart verifies the cursor lands at the
// start of the word after Ctrl+Shift+Left.
func TestIntegrationPhase11CtrlShiftLeftCursorAtWordStart(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world") // cursor at 11

	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft, Modifiers: tcell.ModCtrl | tcell.ModShift}})

	if pos := il.CursorPos(); pos != 6 {
		t.Errorf("Ctrl+Shift+Left from 11: CursorPos() = %d, want 6", pos)
	}
}

// TestIntegrationPhase11CtrlShiftLeftSelectionIsNonEmpty verifies the result is a
// non-empty selection (falsification guard against treating it like plain Ctrl+Left).
func TestIntegrationPhase11CtrlShiftLeftSelectionIsNonEmpty(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft, Modifiers: tcell.ModCtrl | tcell.ModShift}})

	start, end := il.Selection()
	if start == end {
		t.Errorf("Ctrl+Shift+Left: Selection() = (%d, %d) — empty selection; Ctrl+Shift+Left must create a selection", start, end)
	}
}

// TestIntegrationPhase11CtrlShiftLeftEventConsumed verifies the event is consumed.
func TestIntegrationPhase11CtrlShiftLeftEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	event := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft, Modifiers: tcell.ModCtrl | tcell.ModShift}}
	il.HandleEvent(event)

	if !event.IsCleared() {
		t.Error("Ctrl+Shift+Left event was not consumed")
	}
}

// ---------------------------------------------------------------------------
// Requirement 5: Ctrl+Backspace from end of "hello world" deletes "world" leaving "hello "
// ---------------------------------------------------------------------------

// TestIntegrationPhase11CtrlBackspaceFromEndDeletesLastWord verifies that with cursor at
// the end of "hello world", pressing Ctrl+Backspace deletes "world" leaving "hello ".
//
// Scenario: SetText("hello world") → cursor=11 → Ctrl+Backspace → Text()="hello ".
func TestIntegrationPhase11CtrlBackspaceFromEndDeletesLastWord(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world") // cursor at 11

	event := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyBackspace2, Modifiers: tcell.ModCtrl}}
	il.HandleEvent(event)

	if got := il.Text(); got != "hello " {
		t.Errorf("Ctrl+Backspace from end of \"hello world\": Text() = %q, want %q", got, "hello ")
	}
}

// TestIntegrationPhase11CtrlBackspaceFromEndCursorPosition verifies the cursor lands at
// position 6 (the former wordLeft position) after deleting "world".
func TestIntegrationPhase11CtrlBackspaceFromEndCursorPosition(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world") // cursor at 11

	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyBackspace2, Modifiers: tcell.ModCtrl}})

	if pos := il.CursorPos(); pos != 6 {
		t.Errorf("after Ctrl+Backspace from 11: CursorPos() = %d, want 6", pos)
	}
}

// TestIntegrationPhase11CtrlBackspaceFromEndEventConsumed verifies the event is consumed.
func TestIntegrationPhase11CtrlBackspaceFromEndEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	event := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyBackspace2, Modifiers: tcell.ModCtrl}}
	il.HandleEvent(event)

	if !event.IsCleared() {
		t.Error("Ctrl+Backspace event was not consumed")
	}
}

// ---------------------------------------------------------------------------
// Requirement 6: Ctrl+Delete from start of "hello world" deletes "hello " leaving "world"
// ---------------------------------------------------------------------------

// TestIntegrationPhase11CtrlDeleteFromStartDeletesFirstWord verifies that with cursor at
// position 0, pressing Ctrl+Delete deletes "hello " leaving "world".
//
// Scenario: SetText("hello world") → Home → cursor=0 → Ctrl+Delete → Text()="world".
func TestIntegrationPhase11CtrlDeleteFromStartDeletesFirstWord(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyHome}})

	if pre := il.CursorPos(); pre != 0 {
		t.Fatalf("pre-condition: cursor after Home = %d, want 0", pre)
	}

	event := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDelete, Modifiers: tcell.ModCtrl}}
	il.HandleEvent(event)

	if got := il.Text(); got != "world" {
		t.Errorf("Ctrl+Delete from pos 0 in \"hello world\": Text() = %q, want %q", got, "world")
	}
}

// TestIntegrationPhase11CtrlDeleteFromStartCursorStaysAtZero verifies the cursor stays at
// position 0 after Ctrl+Delete (it deletes forward, cursor does not move).
func TestIntegrationPhase11CtrlDeleteFromStartCursorStaysAtZero(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyHome}})

	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDelete, Modifiers: tcell.ModCtrl}})

	if pos := il.CursorPos(); pos != 0 {
		t.Errorf("after Ctrl+Delete from pos 0: CursorPos() = %d, want 0 (cursor must not move)", pos)
	}
}

// TestIntegrationPhase11CtrlDeleteFromStartEventConsumed verifies the event is consumed.
func TestIntegrationPhase11CtrlDeleteFromStartEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyHome}})

	event := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDelete, Modifiers: tcell.ModCtrl}}
	il.HandleEvent(event)

	if !event.IsCleared() {
		t.Error("Ctrl+Delete event was not consumed")
	}
}

// ---------------------------------------------------------------------------
// Requirement 7: Ctrl+Backspace with active selection deletes the selection (not the word)
// ---------------------------------------------------------------------------

// TestIntegrationPhase11CtrlBackspaceWithSelectionDeletesSelection verifies that when a
// selection is active, Ctrl+Backspace deletes the selected text rather than a word.
//
// Scenario: SetText("hello world") → select "world" (pos 6..11) via Shift+End from pos 6
// → Ctrl+Backspace → Text()="hello ".
func TestIntegrationPhase11CtrlBackspaceWithSelectionDeletesSelection(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	// Move cursor to pos 6 (start of "world") using Ctrl+Left from end.
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft, Modifiers: tcell.ModCtrl}})
	if pos := il.CursorPos(); pos != 6 {
		t.Fatalf("pre-condition: cursor after Ctrl+Left = %d, want 6", pos)
	}

	// Shift+End selects from 6 to 11 ("world").
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnd, Modifiers: tcell.ModShift}})
	selStart, selEnd := il.Selection()
	if selStart == selEnd {
		t.Fatalf("pre-condition: Shift+End should create selection, got selStart=%d selEnd=%d", selStart, selEnd)
	}

	event := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyBackspace2, Modifiers: tcell.ModCtrl}}
	il.HandleEvent(event)

	if got := il.Text(); got != "hello " {
		t.Errorf("Ctrl+Backspace with selection \"world\": Text() = %q, want %q", got, "hello ")
	}
}

// TestIntegrationPhase11CtrlBackspaceWithSelectionCursorAtSelectionStart verifies the
// cursor lands at the start of the former selection (not the word boundary).
func TestIntegrationPhase11CtrlBackspaceWithSelectionCursorAtSelectionStart(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	// Move to pos 6 and select "world" via Shift+End.
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft, Modifiers: tcell.ModCtrl}})
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnd, Modifiers: tcell.ModShift}})

	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyBackspace2, Modifiers: tcell.ModCtrl}})

	if pos := il.CursorPos(); pos != 6 {
		t.Errorf("cursor after Ctrl+Backspace on selection: CursorPos() = %d, want 6 (start of former selection)", pos)
	}
}

// TestIntegrationPhase11CtrlBackspaceWithSelectionNoSelectionAfter verifies the selection
// is cleared after the deletion.
func TestIntegrationPhase11CtrlBackspaceWithSelectionNoSelectionAfter(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlA}})

	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyBackspace2, Modifiers: tcell.ModCtrl}})

	start, end := il.Selection()
	if start != end {
		t.Errorf("after Ctrl+Backspace with selection, Selection() = (%d, %d); expected no selection", start, end)
	}
}

// ---------------------------------------------------------------------------
// Requirement 8: Insert key toggles overwrite mode; type "X" at pos 0 → "Xbc"
// ---------------------------------------------------------------------------

// TestIntegrationPhase11InsertKeyTogglesOverwriteAndReplaces verifies the full flow:
// type "abc", move to Home, press Insert to enter overwrite mode, type "X" → text becomes "Xbc".
//
// Scenario: type 'a','b','c' → cursor=3 → Home → cursor=0 → Insert → Overwrite()=true
// → type 'X' → Text()="Xbc".
func TestIntegrationPhase11InsertKeyTogglesOverwriteAndReplaces(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)

	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'a'}})
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'b'}})
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'c'}})

	if got := il.Text(); got != "abc" {
		t.Fatalf("pre-condition: after typing \"abc\", Text() = %q, want %q", got, "abc")
	}

	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyHome}})
	if pos := il.CursorPos(); pos != 0 {
		t.Fatalf("pre-condition: cursor after Home = %d, want 0", pos)
	}

	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyInsert}})
	if !il.Overwrite() {
		t.Fatalf("pre-condition: after Insert key, Overwrite() = false; want true")
	}

	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'X'}})

	if got := il.Text(); got != "Xbc" {
		t.Errorf("overwrite at pos 0 in \"abc\" with 'X': Text() = %q, want %q", got, "Xbc")
	}
}

// TestIntegrationPhase11InsertKeyOverwritePreservesTextLength verifies that overwriting
// a character in the middle of the text keeps the text length unchanged.
func TestIntegrationPhase11InsertKeyOverwritePreservesTextLength(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("abc")
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyHome}})
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyInsert}})
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'X'}})

	if n := len([]rune(il.Text())); n != 3 {
		t.Errorf("after overwrite at pos 0 in \"abc\": len(text) = %d, want 3 (replaced, not inserted)", n)
	}
}

// TestIntegrationPhase11InsertKeyTogglesOverwriteModeState verifies Overwrite() is true
// after pressing Insert and false after pressing Insert a second time.
func TestIntegrationPhase11InsertKeyTogglesOverwriteModeState(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)

	if il.Overwrite() {
		t.Fatal("pre-condition: Overwrite() should be false initially")
	}

	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyInsert}})
	if !il.Overwrite() {
		t.Error("after first Insert: Overwrite() = false, want true")
	}

	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyInsert}})
	if il.Overwrite() {
		t.Error("after second Insert: Overwrite() = true, want false (toggled back)")
	}
}

// ---------------------------------------------------------------------------
// Requirement 9: Overwrite at end of text appends
// ---------------------------------------------------------------------------

// TestIntegrationPhase11OverwriteAtEndAppendsCharacter verifies that when in overwrite
// mode and the cursor is at the end of "ab", typing 'c' appends to produce "abc".
//
// Scenario: SetText("ab") → cursor=2 (end) → Insert → Overwrite()=true → type 'c' → Text()="abc".
func TestIntegrationPhase11OverwriteAtEndAppendsCharacter(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("ab") // cursor at 2 (end)

	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyInsert}})
	if !il.Overwrite() {
		t.Fatalf("pre-condition: Overwrite() = false after Insert; want true")
	}

	if pos := il.CursorPos(); pos != 2 {
		t.Fatalf("pre-condition: cursor = %d, want 2 (end of \"ab\")", pos)
	}

	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'c'}})

	if got := il.Text(); got != "abc" {
		t.Errorf("overwrite at end of \"ab\" with 'c': Text() = %q, want %q", got, "abc")
	}
}

// TestIntegrationPhase11OverwriteAtEndIncreasesLength verifies that appending in overwrite
// mode at the end increases the text length by 1.
func TestIntegrationPhase11OverwriteAtEndIncreasesLength(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("ab") // cursor at 2 (end)
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyInsert}})
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'c'}})

	if n := len([]rune(il.Text())); n != 3 {
		t.Errorf("overwrite append at end: len(text) = %d, want 3", n)
	}
}

// TestIntegrationPhase11OverwriteAtEndCursorAdvances verifies the cursor advances after
// appending in overwrite mode at the end.
func TestIntegrationPhase11OverwriteAtEndCursorAdvances(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("ab")
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyInsert}})
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: 'c'}})

	if pos := il.CursorPos(); pos != 3 {
		t.Errorf("cursor after overwrite-append 'c' at end of \"ab\": CursorPos() = %d, want 3", pos)
	}
}

// ---------------------------------------------------------------------------
// Requirement 10: Ctrl+Y clears all text, cursor at 0, no selection
// ---------------------------------------------------------------------------

// TestIntegrationPhase11CtrlYClearsTextAndResetsCursor verifies that Ctrl+Y clears all
// text, moves the cursor to position 0, and leaves no selection.
//
// Scenario: SetText("hello world") → cursor=11 → Ctrl+Y → Text()="" + CursorPos()=0.
func TestIntegrationPhase11CtrlYClearsTextAndResetsCursor(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world") // cursor at 11

	event := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlY}}
	il.HandleEvent(event)

	if got := il.Text(); got != "" {
		t.Errorf("Ctrl+Y: Text() = %q, want empty string", got)
	}
	if pos := il.CursorPos(); pos != 0 {
		t.Errorf("Ctrl+Y: CursorPos() = %d, want 0", pos)
	}
	start, end := il.Selection()
	if start != end {
		t.Errorf("Ctrl+Y: Selection() = (%d, %d); expected no selection (start==end)", start, end)
	}
}

// TestIntegrationPhase11CtrlYEventConsumed verifies Ctrl+Y is consumed.
func TestIntegrationPhase11CtrlYEventConsumed(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	event := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlY}}
	il.HandleEvent(event)

	if !event.IsCleared() {
		t.Error("Ctrl+Y event was not consumed")
	}
}

// TestIntegrationPhase11CtrlYAfterTypingClearsInteractiveText verifies Ctrl+Y also
// clears text that was entered via keyboard events (not just SetText), covering the
// interaction between typing and clearing.
func TestIntegrationPhase11CtrlYAfterTypingClearsInteractiveText(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	for _, r := range "hello world" {
		il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRune, Rune: r}})
	}
	if got := il.Text(); got != "hello world" {
		t.Fatalf("pre-condition: typed text = %q, want %q", got, "hello world")
	}

	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlY}})

	if got := il.Text(); got != "" {
		t.Errorf("Ctrl+Y after typing \"hello world\": Text() = %q, want empty", got)
	}
	if pos := il.CursorPos(); pos != 0 {
		t.Errorf("Ctrl+Y after typing: CursorPos() = %d, want 0", pos)
	}
}

// ---------------------------------------------------------------------------
// Requirement 11: Ctrl+Y on non-empty text with selection clears everything
// ---------------------------------------------------------------------------

// TestIntegrationPhase11CtrlYWithSelectionClearsEverything verifies that when a selection
// is active, Ctrl+Y clears all text (not just the selected portion), resets the cursor
// to 0, and leaves no selection.
//
// Scenario: SetText("hello world") → Ctrl+A (select all) → Ctrl+Y → all cleared.
func TestIntegrationPhase11CtrlYWithSelectionClearsEverything(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	// Select all text.
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlA}})
	selStart, selEnd := il.Selection()
	if selStart == selEnd {
		t.Fatalf("pre-condition: Ctrl+A should create selection, got selStart=%d selEnd=%d", selStart, selEnd)
	}

	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlY}})

	if got := il.Text(); got != "" {
		t.Errorf("Ctrl+Y with selection active: Text() = %q, want empty (all text cleared)", got)
	}
	if pos := il.CursorPos(); pos != 0 {
		t.Errorf("Ctrl+Y with selection: CursorPos() = %d, want 0", pos)
	}
	start, end := il.Selection()
	if start != end {
		t.Errorf("Ctrl+Y with selection: Selection() = (%d, %d); expected no selection after clear", start, end)
	}
}

// TestIntegrationPhase11CtrlYWithPartialSelectionClearsAll verifies that even with only
// a partial selection active (e.g. "world" selected), Ctrl+Y clears ALL text, not just
// the selection.
//
// This distinguishes Ctrl+Y (clear all) from Delete/Backspace (delete selection).
func TestIntegrationPhase11CtrlYWithPartialSelectionClearsAll(t *testing.T) {
	il := NewInputLine(NewRect(0, 0, 40, 1), 0)
	il.SetText("hello world")

	// Move to pos 6 and select "world" via Shift+End.
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft, Modifiers: tcell.ModCtrl}})
	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnd, Modifiers: tcell.ModShift}})

	selStart, selEnd := il.Selection()
	if selStart == selEnd {
		t.Fatalf("pre-condition: Shift+End should select \"world\", got selStart=%d selEnd=%d", selStart, selEnd)
	}

	il.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlY}})

	if got := il.Text(); got != "" {
		t.Errorf("Ctrl+Y with partial selection: Text() = %q, want empty (Ctrl+Y must clear ALL text, not just selection)", got)
	}
	if pos := il.CursorPos(); pos != 0 {
		t.Errorf("Ctrl+Y with partial selection: CursorPos() = %d, want 0", pos)
	}
}

// TestIntegrationPhase11CtrlYNotSameAsDeleteSelection verifies Ctrl+Y and Delete behave
// differently: Ctrl+Y always clears all text while Delete with selection removes only the
// selected portion.
//
// This is a falsification guard — it confirms that Ctrl+Y does not accidentally route
// to the same code path as Delete.
func TestIntegrationPhase11CtrlYNotSameAsDeleteSelection(t *testing.T) {
	// Widget A: Ctrl+Y with partial selection → empty.
	ilA := NewInputLine(NewRect(0, 0, 40, 1), 0)
	ilA.SetText("hello world")
	ilA.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft, Modifiers: tcell.ModCtrl}})
	ilA.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnd, Modifiers: tcell.ModShift}})
	ilA.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyCtrlY}})
	textAfterCtrlY := ilA.Text()

	// Widget B: Delete with same partial selection → partial deletion.
	ilB := NewInputLine(NewRect(0, 0, 40, 1), 0)
	ilB.SetText("hello world")
	ilB.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft, Modifiers: tcell.ModCtrl}})
	ilB.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyEnd, Modifiers: tcell.ModShift}})
	ilB.HandleEvent(&Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDelete}})
	textAfterDelete := ilB.Text()

	if textAfterCtrlY == textAfterDelete {
		t.Errorf("Ctrl+Y and Delete produced the same result %q; they must behave differently (Ctrl+Y clears all, Delete removes selection)", textAfterCtrlY)
	}
	if textAfterCtrlY != "" {
		t.Errorf("Ctrl+Y result = %q, want empty string", textAfterCtrlY)
	}
}
