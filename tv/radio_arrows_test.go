package tv

// radio_arrows_test.go — Tests for Task 3: Left/Right Arrow Keys in RadioButtons.
//
// Left arrow behaves like Up (moves selection to previous, same as existing Up).
// Right arrow behaves like Down (moves selection to next, same as existing Down).
// At boundaries, no wrap (same as existing Up/Down behavior).

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// TestRadioArrowsRightMovesToNextAndSelects verifies Right arrow moves selection to
// the next RadioButton (same behavior as Down arrow).
// Spec: "Right arrow behaves like Down (moves selection to next)."
func TestRadioArrowsRightMovesToNextAndSelects(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	// Item(0) selected by default.
	rbs.SetState(SfSelected, true) // RadioButtons must be focused to handle arrow keys

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
	rbs.HandleEvent(ev)

	if !rbs.Item(1).Selected() {
		t.Error("Right arrow did not select Item(1)")
	}
}

// TestRadioArrowsRightDeselectsPrevious verifies Right arrow deselects the previously
// selected button.
// Spec: "Right arrow behaves like Down (moves selection to next)."
func TestRadioArrowsRightDeselectsPrevious(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetState(SfSelected, true) // RadioButtons must be focused to handle arrow keys

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
	rbs.HandleEvent(ev)

	if rbs.Item(0).Selected() {
		t.Error("Item(0) should be deselected after Right arrow moved to Item(1)")
	}
}

// TestRadioArrowsLeftMovesToPreviousAndSelects verifies Left arrow moves selection to
// the previous RadioButton (same behavior as Up arrow).
// Spec: "Left arrow behaves like Up (moves selection to previous)."
func TestRadioArrowsLeftMovesToPreviousAndSelects(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetValue(2)
	rbs.SetFocusedChild(rbs.Item(2))
	rbs.SetState(SfSelected, true) // RadioButtons must be focused to handle arrow keys

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft}}
	rbs.HandleEvent(ev)

	if !rbs.Item(1).Selected() {
		t.Error("Left arrow did not select Item(1) after starting at Item(2)")
	}
}

// TestRadioArrowsRightAtLastItemIsNoOp verifies Right at the last button does not
// wrap and does not change selection.
// Spec: "At boundaries, no wrap (same as existing Up/Down behavior)."
func TestRadioArrowsRightAtLastItemIsNoOp(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetValue(2)
	rbs.SetFocusedChild(rbs.Item(2))

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
	rbs.HandleEvent(ev)

	if !rbs.Item(2).Selected() {
		t.Error("Right at last item should be a no-op; Item(2) is no longer selected")
	}
	if rbs.Item(0).Selected() {
		t.Error("Right at last item must not wrap to Item(0)")
	}
}

// TestRadioArrowsLeftAtFirstItemIsNoOp verifies Left at the first button does not
// wrap and does not change selection.
// Spec: "At boundaries, no wrap (same as existing Up/Down behavior)."
func TestRadioArrowsLeftAtFirstItemIsNoOp(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	// Item(0) is selected by default.

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft}}
	rbs.HandleEvent(ev)

	if !rbs.Item(0).Selected() {
		t.Error("Left at first item should be a no-op; Item(0) is no longer selected")
	}
	if rbs.Item(2).Selected() {
		t.Error("Left at first item must not wrap to Item(2)")
	}
}

// TestRadioArrowsRightClearsEvent verifies Right arrow clears (consumes) the event.
// Spec: "At boundaries, no wrap (same as existing Up/Down behavior)" — events consumed.
func TestRadioArrowsRightClearsEvent(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetState(SfSelected, true) // RadioButtons must be focused to handle arrow keys

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyRight}}
	rbs.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("Right arrow did not clear event; ev.What = %v, want EvNothing", ev.What)
	}
}

// TestRadioArrowsLeftClearsEvent verifies Left arrow clears (consumes) the event.
// Spec: "At boundaries, no wrap (same as existing Up/Down behavior)" — events consumed.
func TestRadioArrowsLeftClearsEvent(t *testing.T) {
	rbs := NewRadioButtons(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	rbs.SetValue(1)
	rbs.SetState(SfSelected, true) // RadioButtons must be focused to handle arrow keys

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyLeft}}
	rbs.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("Left arrow did not clear event; ev.What = %v, want EvNothing", ev.What)
	}
}
