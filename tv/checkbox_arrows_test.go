package tv

// checkbox_arrows_test.go — Tests for Task 2: Arrow Key Navigation in CheckBoxes.
//
// Up arrow moves internal focus to the previous checkbox item (does NOT toggle).
// Down arrow moves internal focus to the next checkbox item (does NOT toggle).
// At the first item, Up does nothing (no wrap).
// At the last item, Down does nothing (no wrap).
// Arrow keys clear the event after handling.
//
// Note: CheckBoxes only handles arrow keys when SfSelected=true (it is the
// focused widget in its parent group). This prevents CheckBoxes from stealing
// Down/Up events from OfPostProcess siblings like History when it is not focused.

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// TestCheckBoxArrowsDownMovesFocusToNextItem verifies Down arrow moves focus from
// Item(0) to Item(1) without toggling the checked state.
// Spec: "Down arrow moves internal focus to the next checkbox item (does NOT toggle)."
func TestCheckBoxArrowsDownMovesFocusToNextItem(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	cbs.SetFocusedChild(cbs.Item(0))
	cbs.SetState(SfSelected, true) // CheckBoxes must be focused to handle arrow keys

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	cbs.HandleEvent(ev)

	if cbs.FocusedChild() != cbs.Item(1) {
		t.Errorf("Down arrow: FocusedChild() = %v, want Item(1)", cbs.FocusedChild())
	}
}

// TestCheckBoxArrowsDownDoesNotToggle verifies Down arrow does not change the
// checked state of any item.
// Spec: "Down arrow moves internal focus to the next checkbox item (does NOT toggle)."
func TestCheckBoxArrowsDownDoesNotToggle(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	cbs.SetFocusedChild(cbs.Item(0))
	cbs.SetState(SfSelected, true) // CheckBoxes must be focused to handle arrow keys

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	cbs.HandleEvent(ev)

	for i := 0; i < 3; i++ {
		if cbs.Item(i).Checked() {
			t.Errorf("Down arrow toggled Item(%d); arrow keys must not toggle", i)
		}
	}
}

// TestCheckBoxArrowsUpMovesFocusToPreviousItem verifies Up arrow moves focus from
// Item(2) to Item(1).
// Spec: "Up arrow moves internal focus to the previous checkbox item (does NOT toggle)."
func TestCheckBoxArrowsUpMovesFocusToPreviousItem(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	cbs.SetFocusedChild(cbs.Item(2))
	cbs.SetState(SfSelected, true) // CheckBoxes must be focused to handle arrow keys

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	cbs.HandleEvent(ev)

	if cbs.FocusedChild() != cbs.Item(1) {
		t.Errorf("Up arrow: FocusedChild() = %v, want Item(1)", cbs.FocusedChild())
	}
}

// TestCheckBoxArrowsUpAtFirstItemDoesNothing verifies Up at the first item is a
// no-op (no wrap).
// Spec: "At the first item, Up does nothing (no wrap)."
func TestCheckBoxArrowsUpAtFirstItemDoesNothing(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	cbs.SetFocusedChild(cbs.Item(0))
	cbs.SetState(SfSelected, true) // CheckBoxes must be focused to handle arrow keys

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	cbs.HandleEvent(ev)

	if cbs.FocusedChild() != cbs.Item(0) {
		t.Errorf("Up at first item: FocusedChild() changed to %v, want Item(0) (no wrap)", cbs.FocusedChild())
	}
}

// TestCheckBoxArrowsDownAtLastItemDoesNothing verifies Down at the last item is a
// no-op (no wrap).
// Spec: "At the last item, Down does nothing (no wrap)."
func TestCheckBoxArrowsDownAtLastItemDoesNothing(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	cbs.SetFocusedChild(cbs.Item(2))
	cbs.SetState(SfSelected, true) // CheckBoxes must be focused to handle arrow keys

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	cbs.HandleEvent(ev)

	if cbs.FocusedChild() != cbs.Item(2) {
		t.Errorf("Down at last item: FocusedChild() changed to %v, want Item(2) (no wrap)", cbs.FocusedChild())
	}
}

// TestCheckBoxArrowsDownClearsEvent verifies Down arrow clears (consumes) the event
// when CheckBoxes is focused.
// Spec: "Arrow keys clear the event after handling."
func TestCheckBoxArrowsDownClearsEvent(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	cbs.SetFocusedChild(cbs.Item(0))
	cbs.SetState(SfSelected, true) // CheckBoxes must be focused to handle arrow keys

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyDown}}
	cbs.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("Down arrow did not clear event; ev.What = %v, want EvNothing", ev.What)
	}
}

// TestCheckBoxArrowsUpClearsEvent verifies Up arrow clears (consumes) the event
// when CheckBoxes is focused.
// Spec: "Arrow keys clear the event after handling."
func TestCheckBoxArrowsUpClearsEvent(t *testing.T) {
	cbs := NewCheckBoxes(NewRect(0, 0, 20, 3), []string{"A", "B", "C"})
	cbs.SetFocusedChild(cbs.Item(2))
	cbs.SetState(SfSelected, true) // CheckBoxes must be focused to handle arrow keys

	ev := &Event{What: EvKeyboard, Key: &KeyEvent{Key: tcell.KeyUp}}
	cbs.HandleEvent(ev)

	if !ev.IsCleared() {
		t.Errorf("Up arrow did not clear event; ev.What = %v, want EvNothing", ev.What)
	}
}
